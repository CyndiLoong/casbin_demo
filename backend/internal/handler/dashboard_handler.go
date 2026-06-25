// Package handler 实现 HTTP 请求处理层（Controller/Presentation Layer）。
package handler

import (
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

// GetStats 获取仪表盘统计信息接口（需 JWT + Casbin 权限）。
// GET /api/dashboard
//
// 返回欢迎信息、当前用户角色、系统统计和技术栈信息。
// 统计数据通过 DashboardService 获取，内置 Redis 缓存（TTL=2min+抖动）。
func (h *DashboardHandler) GetStats(c *gin.Context) {
	username, _ := c.Get("username")
	roles, _ := c.Get("roles")

	stats, err := h.dashboardService.GetStats()
	if err != nil {
		response.ServerError(c, "获取统计数据失败")
		return
	}

	var totalUsers, totalRoles, totalPerms int64
	if stats != nil {
		totalUsers = stats.TotalUsers
		totalRoles = stats.TotalRoles
		totalPerms = stats.TotalPermissions
	}

	response.Success(c, gin.H{
		"welcome":  "欢迎回来",
		"username": username,
		"roles":    roles,
		"stats": gin.H{
			"total_users":       totalUsers,
			"total_roles":       totalRoles,
			"total_permissions": totalPerms,
		},
		"system_info": gin.H{
			"framework":  "Gin + Gorm + Casbin",
			"database":   "PostgreSQL",
			"cache":      "Redis",
			"go_version": "1.26.4",
		},
	})
}
