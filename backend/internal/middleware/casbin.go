package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	casbinpkg "casbin-demo/pkg/casbin"
	"casbin-demo/pkg/response"
)

/*
CasbinAuth Casbin RBAC 权限校验中间件。

必须在 JWTAuth 之后使用，依赖 Context 中的 "username" 字段。

校验逻辑：
 1. 从 Context 获取用户名（sub）
 2. 获取请求路径（obj）和 HTTP 方法（act）
 3. 调用 Casbin Enforcer.Enforce(sub, obj, act) 进行权限校验
 4. 若 Enforcer 未初始化（如 Redis 不可用降级场景）则直接放行

权限不足时返回 HTTP 403 并终止请求链。
*/
func CasbinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			response.Forbidden(c, "未认证用户")
			c.Abort()
			return
		}

		obj := c.Request.URL.Path
		act := c.Request.Method

		if casbinpkg.Enforcer == nil {
			c.Next()
			return
		}

		allowed, err := casbinpkg.Enforcer.Enforce(username.(string), obj, act)
		if err != nil {
			slog.Error("casbin enforce error", "error", err, "user", username, "path", obj, "method", act)
			response.ServerError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !allowed {
			slog.Warn("permission denied", "user", username, "path", obj, "method", act)
			response.Forbidden(c, "没有访问权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CustomRecovery 自定义 Recovery 中间件，panic 时返回 JSON 格式的 500 错误。
func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "path", c.Request.URL.Path, "method", c.Request.Method)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code:    500,
					Message: "服务器内部错误",
				})
			}
		}()
		c.Next()
	}
}
