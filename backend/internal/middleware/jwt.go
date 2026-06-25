// Package middleware 提供 HTTP 中间件（JWT认证、Casbin权限、CORS、日志等）。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	jwtpkg "casbin-demo/pkg/jwt"
	"casbin-demo/pkg/response"
)

// JWTAuth JWT 认证中间件。
//
// 处理流程：
//  1. 从 Authorization 头提取 Bearer Token
//  2. 验证 Token 格式（必须是 "Bearer <token>"）
//  3. 解析并验证 JWT 签名和有效期
//  4. 将用户信息（user_id/uuid/username/roles）写入 Gin Context
//
// 认证失败时直接返回 401 并终止请求链。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		claims, err := jwtpkg.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "无效的认证令牌")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("uuid", claims.UUID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}
