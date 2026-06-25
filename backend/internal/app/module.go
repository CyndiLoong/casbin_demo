/*
Package app 使用 uber-go/fx 实现依赖注入（DI），组装应用各层组件。

依赖注入链路：

	config.Config
	  ├─→ *gorm.DB (PostgreSQL)
	  │    ├─→ *repository.UserRepository
	  │    ├─→ *repository.RoleRepository
	  │    ├─→ *repository.PermissionRepository
	  │    └─→ *repository.AuditRepository
	  ├─→ *redis.Client (可选，失败时为 nil)
	  ├─→ *mq.Client (RabbitMQ，可选，失败时为 nil)
	  ├─→ *ws.Hub (WebSocket Hub，基于 Redis Pub/Sub 跨实例)
	  └─→ *cache.Client (L1+L2多级缓存)

	Repositories + CacheClient + MQ + WS Hub
	  ├─→ *service.AuthService
	  ├─→ *service.UserService
	  ├─→ *service.RoleService
	  ├─→ *service.PermissionService
	  ├─→ *service.DashboardService
	  └─→ *service.AuditService (审核业务 + MQ消费者 + 定时补发)

	Services
	  └─→ *router.Handlers (聚合所有 Handler，含 WsHandler)

	Handlers + Config.Server.Mode
	  └─→ *router.EngineWrapper (Gin引擎 + 限流器)

	EngineWrapper + Config.Server.Port
	  └─→ HTTP Server (通过 fx.Lifecycle 管理启动/优雅关闭)

生命周期钩子（Lifecycle Hooks）执行顺序：
 1. OnStart（按 Provide 注册顺序启动）：
    - WS Hub 启动事件循环
    - AuditService 启动 MQ 消费者和定时补发
    - HTTP Server 在 goroutine 中启动
 2. OnStop（按注册逆序关闭）：
    a. HTTP Server 优雅关闭（等待请求完成，10s超时）
    b. AuditService 停止（等待定时任务 goroutine 结束）
    c. WS Hub 停止（关闭所有连接）
    d. MQ Client 关闭（等待消费者 ACK）
    e. 限流器停止（后台清理 goroutine）
    f. 缓存客户端关闭（异步重建 worker + LocalCache GC）
    g. Redis 连接关闭
*/
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"casbin-demo/internal/config"
	"casbin-demo/internal/handler"
	"casbin-demo/internal/repository"
	"casbin-demo/internal/router"
	"casbin-demo/internal/service"
	"casbin-demo/pkg/cache"
	casbinpkg "casbin-demo/pkg/casbin"
	"casbin-demo/pkg/mq"
	"casbin-demo/pkg/ws"
)

// Module 返回 fx.Option，包含所有组件的 Provide 和 Invoke。
//
// 使用方式（main.go）：
//
//	fx.New(app.Module()).Run()
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			// 基础资源
			provideConfig,
			provideDB,
			provideRedis,
			provideMQ,
			provideWSHub,
			provideCacheClient,

			// Repository 层
			repository.NewUserRepository,
			repository.NewRoleRepository,
			repository.NewPermissionRepository,
			repository.NewAuditRepository,

			// Service 层
			service.NewAuthService,
			service.NewUserService,
			service.NewRoleService,
			service.NewPermissionService,
			service.NewDashboardService,
			service.NewAuditService,

			// Handler 层
			handler.NewAuthHandler,
			handler.NewUserHandler,
			handler.NewRoleHandler,
			handler.NewPermissionHandler,
			handler.NewDashboardHandler,
			handler.NewAuditHandler,
			handler.NewWsHandler,

			// Router 层
			provideHandlers,
			provideEngine,
		),
		fx.Invoke(
			setupLogger,
			autoMigrate,
			seedData,
			initCasbin,
			warmupCache,
			startWSHub,
			startMQConsumers,
			registerRoutes,
			startHTTPServer,
		),
	)
}

// setupLogger 初始化结构化日志（fx.Invoke 确保在应用启动时执行一次）。
func setupLogger() {
	setupSlog()
	slog.Info("=== Casbin RBAC Demo Starting ===",
		"go_version", "1.26.4",
		"di_framework", "uber-go/fx",
		"features", "RBAC+WebSocket+RabbitMQ+Cache",
		"time", time.Now().Format(time.RFC3339),
	)
}

// provideConfig 加载应用配置。
func provideConfig() *config.Config {
	return mustLoadConfig()
}

// provideDB 初始化 PostgreSQL 数据库连接。
func provideDB(cfg *config.Config) *gorm.DB {
	return mustInitDB(cfg)
}

// provideRedis 初始化 Redis 客户端（可选组件，失败返回 nil）。
func provideRedis(cfg *config.Config) *redis.Client {
	rdb, err := repository.InitRedis(cfg)
	if err != nil {
		slog.Warn("redis unavailable, running without cache", "error", err)
		return nil
	}
	slog.Info("redis connected", "host", cfg.Redis.Host, "port", cfg.Redis.Port)
	return rdb
}

// provideMQ 初始化 RabbitMQ 客户端（可选组件，失败返回 nil，降级为仅PG持久化+定时补发）。
func provideMQ(cfg *config.Config) *mq.Client {
	mqCfg := &mq.Config{
		Host:     cfg.RabbitMQ.Host,
		Port:     cfg.RabbitMQ.Port,
		User:     cfg.RabbitMQ.User,
		Password: cfg.RabbitMQ.Password,
		VHost:    cfg.RabbitMQ.VHost,
	}
	client, err := mq.NewClient(mqCfg)
	if err != nil {
		slog.Warn("rabbitmq unavailable, running without MQ (PG only mode with retry scheduler)", "error", err)
		return nil
	}
	return client
}

// provideWSHub 创建 WebSocket Hub 实例。
func provideWSHub(rdb *redis.Client) *ws.Hub {
	return ws.NewHub(rdb)
}

// provideCacheClient 创建缓存客户端，管理其生命周期（fx.Lifecycle）。
func provideCacheClient(lc fx.Lifecycle, rdb *redis.Client) *cache.Client {
	c := cache.NewClient(rdb)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.Debug("cache client closing...")
			c.Close()
			if rdb != nil {
				return rdb.Close()
			}
			return nil
		},
	})
	return c
}

// provideHandlers 聚合所有 Handler 实例。
func provideHandlers(
	auth *handler.AuthHandler,
	user *handler.UserHandler,
	role *handler.RoleHandler,
	perm *handler.PermissionHandler,
	dash *handler.DashboardHandler,
	audit *handler.AuditHandler,
	wsHandler *handler.WsHandler,
) *router.Handlers {
	return &router.Handlers{
		Auth:       auth,
		User:       user,
		Role:       role,
		Permission: perm,
		Dashboard:  dash,
		Audit:      audit,
		WS:         wsHandler,
	}
}

// provideEngine 创建 Gin 引擎包装器，管理限流器生命周期。
func provideEngine(cfg *config.Config, lc fx.Lifecycle, mqClient *mq.Client, wsHub *ws.Hub, auditSvc *service.AuditService) *router.EngineWrapper {
	engine := router.NewEngine(cfg.Server.Mode)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.Debug("shutting down components...")
			auditSvc.Stop()
			wsHub.Stop()
			if mqClient != nil {
				_ = mqClient.Close()
			}
			slog.Debug("rate limiter stopping...")
			engine.Stop()
			return nil
		},
	})
	return engine
}

// autoMigrate 自动迁移新增的数据表（审核申请、系统消息）。
func autoMigrate(auditRepo *repository.AuditRepository) {
	if err := auditRepo.AutoMigrate(); err != nil {
		slog.Error("auto migrate failed", "error", err)
		panic(fmt.Errorf("auto migrate failed: %w", err))
	}
	slog.Info("database migration completed")
}

// seedData 初始化种子数据（幂等）。
func seedData(db *gorm.DB) {
	mustSeedData(db)
}

// initCasbin 初始化 Casbin 权限引擎并加载策略。
func initCasbin(db *gorm.DB) {
	enforcer, err := casbinpkg.InitCasbin(db)
	if err != nil {
		slog.Error("casbin init failed", "error", err)
		panic(fmt.Errorf("casbin init failed: %w", err))
	}
	if err := casbinpkg.LoadCasbinPolicy(enforcer, db); err != nil {
		slog.Error("casbin policy load failed", "error", err)
		panic(fmt.Errorf("casbin policy load failed: %w", err))
	}
	slog.Info("casbin enforcer ready")
}

// warmupCache 缓存预热：角色/权限列表 + 布隆过滤器用户ID预热。
func warmupCache(
	cacheClient *cache.Client,
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	permRepo *repository.PermissionRepository,
) {
	warmupItems := []cache.WarmupItem{
		{
			Key:     cache.CacheKey("role", "list", "all"),
			TTL:     cache.TTLConfig,
			Logical: true,
			Loader: func() (interface{}, error) {
				return roleRepo.List()
			},
		},
		{
			Key:     cache.CacheKey("permission", "list", "all"),
			TTL:     cache.TTLConfig,
			Logical: true,
			Loader: func() (interface{}, error) {
				return permRepo.List()
			},
		},
	}

	users, _, err := userRepo.List(1, 100)
	if err == nil {
		for _, u := range users {
			cacheClient.BloomFilter().Add(cache.CacheKeyUint("user", "id", u.ID))
		}
	}

	cacheClient.Warmup(warmupItems...)
}

// startWSHub 启动 WebSocket Hub 事件循环。
func startWSHub(lc fx.Lifecycle, hub *ws.Hub) {
	hub.Run()
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// WS Hub 的停止在 provideEngine 的 OnStop 中统一处理（顺序控制）
			return nil
		},
	})
}

// startMQConsumers 设置MQ消息处理器并启动定时补发任务。
//
// 执行顺序说明：
//  1. SetMQHandler() — 将业务消息处理器注入 MQ 客户端（MQ 在 provideMQ 阶段已连接并启动消费者，
//     此时handler为nil消息会被ACK丢弃，设置handler后开始正常处理）
//  2. StartRetryScheduler() — 启动定时对账补发（每30秒扫描未投递消息重发）
//
// 注意：MQ客户端的连接和消费者在provideMQ中已启动，这里只注入handler和启动定时任务。
func startMQConsumers(auditSvc *service.AuditService) {
	auditSvc.SetMQHandler()
	auditSvc.StartRetryScheduler()
}

// registerRoutes 注册所有 HTTP 路由。
func registerRoutes(engine *router.EngineWrapper, h *router.Handlers) {
	router.RegisterRoutes(engine, h)
}

// startHTTPServer 启动 HTTP 服务器，通过 fx.Lifecycle 管理优雅关闭。
func startHTTPServer(lc fx.Lifecycle, cfg *config.Config, engine *router.EngineWrapper) {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				slog.Info("server listening", "addr", addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					slog.Error("server failed to start", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("shutdown signal received, gracefully stopping...")
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				slog.Error("server shutdown error", "error", err)
				return err
			}
			slog.Info("server stopped gracefully")
			return nil
		},
	})
}
