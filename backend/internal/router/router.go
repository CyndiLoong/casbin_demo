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
	Resource   *handler.ResourceHandler
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

// RateLimiter 返回限流器实例，供路由层使用（如用户级限流）。
func (ew *EngineWrapper) RateLimiter() *middleware.RateLimiter {
	return ew.rateLimiter
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
// 路由权限层级（三层架构）：
//
//  1. 公开接口（无需认证）：
//     GET  /health              健康检查（Docker/K8s 探针使用）
//     GET  /ws                  WebSocket 连接端点（需 JWT query token）
//     POST /api/login          登录
//     POST /api/register       注册
//
//  2. 登录态基础接口（仅需 JWT 认证，所有登录用户可访问）：
//     GET  /api/userinfo               获取当前用户信息
//     GET  /api/dashboard              仪表盘数据（handler 内部根据角色返回不同内容）
//     POST /api/audit/applications     提交资源申请
//     POST /api/audit/applications/:id/withdraw  撤回申请
//     GET  /api/audit/my-applications  查看我的申请列表
//     GET  /api/audit/applications/:id 查看申请详情
//     GET  /api/messages/unread-count  获取未读消息数
//     GET  /api/messages               获取消息列表
//     PUT  /api/messages/:id/read      标记消息已读
//     PUT  /api/messages/read-all      全部标记已读
//     GET  /api/resources              获取资源列表（所有用户可浏览）
//     GET  /api/resources/active       获取可用资源列表
//     GET  /api/resources/:id          获取资源详情
//
//  3. 管理员接口（需 JWT + Casbin RBAC 权限校验）：
//     /api/users/*             用户管理
//     /api/roles/*             角色管理
//     /api/permissions/*       权限管理
//     GET  /api/audit/applications    查看所有申请
//     POST /api/audit/applications/:id/review  审核申请
//     GET  /api/audit/pending-count   获取待审核数量
//     POST /api/resources             创建资源
//     PUT  /api/resources/:id         更新资源
//     DELETE /api/resources/:id       删除资源
//
// 限流策略：
//   - 全局：IP级（30req/s，突发60）+ 接口级（100req/s，突发200）
//   - 认证后：用户级限流（20req/s，突发40），防止单用户滥用
func RegisterRoutes(ew *EngineWrapper, h *Handlers) {
	r := ew.Engine
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "casbin-demo",
			"version": "3.1.0",
			"features": []string{"RBAC", "WebSocket", "RabbitMQ", "MultiCache", "DistributedRateLimit"},
			"time":    time.Now().Unix(),
		})
	})

	// WebSocket 端点（独立于 /api 组，query 参数传递 token）
	r.GET("/ws", h.WS.Connect)

	api := r.Group("/api")
	{
		// 公开接口（无需认证）
		api.POST("/login", h.Auth.Login)
		api.POST("/register", h.Auth.Register)

		// 认证接口组（需要 JWT，同时应用用户级限流）
		auth := api.Group("")
		auth.Use(middleware.JWTAuth())
		auth.Use(middleware.UserRateLimit(ew.RateLimiter()))
		{
			// 基础登录态接口（所有登录用户可访问，无需 Casbin 权限）
			auth.GET("/userinfo", h.Auth.GetUserInfo)
			auth.GET("/dashboard", h.Dashboard.GetStats)

			// 审核申请 - 普通用户操作
			auth.POST("/audit/applications", h.Audit.Submit)
			auth.POST("/audit/applications/:id/withdraw", h.Audit.Withdraw)
			auth.GET("/audit/my-applications", h.Audit.ListMyApplications)
			auth.GET("/audit/applications/:id", h.Audit.GetDetail)

			// 消息通知（所有登录用户可访问）
			auth.GET("/messages/unread-count", h.Audit.GetUnreadCount)
			auth.GET("/messages", h.Audit.ListMessages)
			auth.PUT("/messages/:id/read", h.Audit.MarkMessageRead)
			auth.PUT("/messages/read-all", h.Audit.MarkAllRead)

			// 资源浏览（所有登录用户可查看资源清单）
			auth.GET("/resources", h.Resource.ListResources)
			auth.GET("/resources/active", h.Resource.ListActiveResources)
			auth.GET("/resources/:id", h.Resource.GetResource)

			// ========== 管理员专属接口（需要 Casbin RBAC 权限校验） ==========
			admin := auth.Group("")
			admin.Use(middleware.CasbinAuth())
			{
				// 用户管理
				admin.GET("/users", h.User.List)
				admin.POST("/users", h.User.Create)
				admin.PUT("/users/:id", h.User.Update)
				admin.DELETE("/users/:id", h.User.Delete)
				admin.POST("/users/assign-role", h.User.AssignRole)

				// 角色管理
				admin.GET("/roles", h.Role.List)
				admin.POST("/roles", h.Role.Create)
				admin.PUT("/roles/:id", h.Role.Update)
				admin.DELETE("/roles/:id", h.Role.Delete)
				admin.POST("/roles/assign-permission", h.Role.AssignPermission)

				// 权限管理
				admin.GET("/permissions", h.Permission.List)
				admin.POST("/permissions", h.Permission.Create)
				admin.PUT("/permissions/:id", h.Permission.Update)
				admin.DELETE("/permissions/:id", h.Permission.Delete)

				// 审核管理 - 管理员操作
				admin.GET("/audit/applications", h.Audit.ListAllApplications)
				admin.POST("/audit/applications/:id/review", h.Audit.Review)
				admin.GET("/audit/pending-count", h.Audit.GetPendingCount)

				// 资源管理 - 管理员操作
				admin.POST("/resources", h.Resource.CreateResource)
				admin.PUT("/resources/:id", h.Resource.UpdateResource)
				admin.DELETE("/resources/:id", h.Resource.DeleteResource)
			}
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
