// Package handler 实现 HTTP 请求处理层（Controller/Presentation Layer）。
package handler

import (
	"log/slog"
	"slices"

	"github.com/gin-gonic/gin"

	"casbin-demo/internal/service"
	"casbin-demo/pkg/response"
)

// DashboardHandler 仪表盘相关 HTTP 处理器。
type DashboardHandler struct {
	dashboardService *service.DashboardService
}

// NewDashboardHandler 创建 DashboardHandler 实例。
func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetStats 获取仪表盘统计信息接口（需 JWT 认证，管理员返回系统统计，普通用户返回基础信息）。
// GET /api/dashboard
//
// 权限说明：
//   - 所有登录用户均可访问此接口（无需 Casbin 权限校验）
//   - 管理员（admin 角色）返回系统级统计数据（用户数、角色数、权限数）
//   - 普通用户仅返回欢迎信息和基础系统信息，不返回敏感统计数据
//
// 缓存策略：管理员统计数据通过 DashboardService 获取，内置 Redis 缓存（TTL=2min+抖动）。
func (h *DashboardHandler) GetStats(c *gin.Context) {
	username, _ := c.Get("username")
	rolesVal, _ := c.Get("roles")

	var isAdmin bool
	if roles, ok := rolesVal.([]string); ok {
		isAdmin = slices.Contains(roles, "admin")
	}

	resp := gin.H{
		"welcome":  "欢迎回来",
		"username": username,
		"roles":    rolesVal,
		"is_admin": isAdmin,
	}

	if isAdmin {
		stats, err := h.dashboardService.GetStats()
		if err != nil {
			slog.Error("get dashboard stats failed", "error", err, "user", username)
		} else if stats != nil {
			resp["stats"] = gin.H{
				"total_users":       stats.TotalUsers,
				"total_roles":       stats.TotalRoles,
				"total_permissions": stats.TotalPermissions,
			}
		}
	}

	resp["system_info"] = gin.H{
		"framework":  "Gin + Gorm + Casbin",
		"database":   "PostgreSQL",
		"cache":      "Redis",
		"go_version": "1.26.4",
	}

	response.Success(c, resp)
}
