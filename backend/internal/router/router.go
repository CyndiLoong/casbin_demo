// Package router 负责 HTTP 路由注册与中间件装配。
//
// 职责边界：
//   - 创建 Gin 引擎实例
//   - 注册全局中间件（Recovery、CORS、日志、限流）
//   - 按模块分组注册 API 路由
//   - 将 Handler 绑定到对应路由
//
// 不负责：
//   - 资源初始化（DB/Redis/Casbin/Cache/MQ/WS 等在 main.go 中完成）
//   - 依赖注入构造（Handler 在 main.go 中构造后传入）
package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"casbin-demo/internal/handler"
	"casbin-demo/internal/middleware"
	"casbin-demo/pkg/response"
)

// Handlers 聚合所有 Handler，用于路由注册。
type Handlers struct {
	Auth       *handler.AuthHandler
	User       *handler.UserHandler
	Role       *handler.RoleHandler
	Permission *handler.PermissionHandler
	Dashboard  *handler.DashboardHandler
	Audit      *handler.AuditHandler
	WS         *handler.WsHandler
}

// EngineWrapper 包装 Gin 引擎和需要优雅关闭的资源（如限流器）。
type EngineWrapper struct {
	*gin.Engine
	rateLimiter *middleware.RateLimiter
}

// Stop 优雅停止引擎相关资源（停止限流器后台 goroutine）。
func (ew *EngineWrapper) Stop() {
	if ew.rateLimiter != nil {
		ew.rateLimiter.Stop()
	}
}

// NewEngine 创建并配置 Gin 引擎。
//
// 中间件执行顺序（栈式，先进后执行 Next 之后的逻辑）：
//  1. CustomRecovery  — 兜底 panic 恢复，防止进程崩溃
//  2. CORS           — 跨域资源共享
//  3. RateLimit      — IP 维度限流（令牌桶，20req/s，突发50），防缓存穿透
//  4. RequestLogger  — 结构化请求日志
func NewEngine(mode string) *EngineWrapper {
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	limiter := middleware.DefaultAPIRateLimiter()

	r.Use(middleware.CustomRecovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))
	r.Use(middleware.RateLimit(limiter))
	r.Use(requestLogger())

	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "请求的资源不存在")
	})
	r.NoMethod(func(c *gin.Context) {
		response.MethodNotAllowed(c, "请求方法不被允许")
	})

	return &EngineWrapper{
		Engine:      r,
		rateLimiter: limiter,
	}
}

// RegisterRoutes 注册所有 HTTP 路由。
//
// 路由权限层级：
//
//	GET  /health              健康检查（公开，Docker/K8s 探针使用）
//	GET  /ws                  WebSocket 连接端点（需 JWT query token）
//	POST /api/login          登录（公开）
//	POST /api/register       注册（公开）
//	GET  /api/userinfo       获取当前用户信息（需 JWT）
//	GET  /api/dashboard      仪表盘统计（需 JWT + Casbin RBAC）
//	/api/users/*             用户管理（需 JWT + Casbin RBAC）
//	/api/roles/*             角色管理（需 JWT + Casbin RBAC）
//	/api/permissions/*       权限管理（需 JWT + Casbin RBAC）
//	/api/audit/*             审核申请管理（需 JWT，部分需 admin）
//	/api/messages/*          消息通知（需 JWT）
func RegisterRoutes(ew *EngineWrapper, h *Handlers) {
	r := ew.Engine
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "casbin-demo",
			"version": "3.0.0",
			"features": []string{"RBAC", "WebSocket", "RabbitMQ", "MultiCache"},
			"time":    time.Now().Unix(),
		})
	})

	// WebSocket 端点（独立于 /api 组，不需要经过 Casbin 中间件）
	r.GET("/ws", h.WS.Connect)

	api := r.Group("/api")
	{
		api.POST("/login", h.Auth.Login)
		api.POST("/register", h.Auth.Register)

		auth := api.Group("")
		auth.Use(middleware.JWTAuth())
		{
			auth.GET("/userinfo", h.Auth.GetUserInfo)
			auth.GET("/dashboard", middleware.CasbinAuth(), h.Dashboard.GetStats)

			auth.GET("/users", middleware.CasbinAuth(), h.User.List)
			auth.POST("/users", middleware.CasbinAuth(), h.User.Create)
			auth.PUT("/users/:id", middleware.CasbinAuth(), h.User.Update)
			auth.DELETE("/users/:id", middleware.CasbinAuth(), h.User.Delete)
			auth.POST("/users/assign-role", middleware.CasbinAuth(), h.User.AssignRole)

			auth.GET("/roles", middleware.CasbinAuth(), h.Role.List)
			auth.POST("/roles", middleware.CasbinAuth(), h.Role.Create)
			auth.PUT("/roles/:id", middleware.CasbinAuth(), h.Role.Update)
			auth.DELETE("/roles/:id", middleware.CasbinAuth(), h.Role.Delete)
			auth.POST("/roles/assign-permission", middleware.CasbinAuth(), h.Role.AssignPermission)

			auth.GET("/permissions", middleware.CasbinAuth(), h.Permission.List)
			auth.POST("/permissions", middleware.CasbinAuth(), h.Permission.Create)
			auth.PUT("/permissions/:id", middleware.CasbinAuth(), h.Permission.Update)
			auth.DELETE("/permissions/:id", middleware.CasbinAuth(), h.Permission.Delete)

			// 审核申请相关路由
			auth.POST("/audit/applications", h.Audit.Submit)
			auth.POST("/audit/applications/:id/withdraw", h.Audit.Withdraw)
			auth.GET("/audit/my-applications", h.Audit.ListMyApplications)
			auth.GET("/audit/applications", h.Audit.ListAllApplications)
			auth.GET("/audit/applications/:id", h.Audit.GetDetail)
			auth.POST("/audit/applications/:id/review", h.Audit.Review)
			auth.GET("/audit/pending-count", h.Audit.GetPendingCount)

			// 消息通知路由
			auth.GET("/messages/unread-count", h.Audit.GetUnreadCount)
			auth.GET("/messages", h.Audit.ListMessages)
			auth.PUT("/messages/:id/read", h.Audit.MarkMessageRead)
			auth.PUT("/messages/read-all", h.Audit.MarkAllRead)
		}
	}
}

// requestLogger 使用 slog 按状态码分级记录 HTTP 请求。
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		attrs := []any{
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		}
		switch {
		case status >= 500:
			slog.Error("server error", attrs...)
		case status >= 400:
			slog.Warn("client error", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	}
}
