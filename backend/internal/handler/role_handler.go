// role_handler.go 实现角色管理相关的 HTTP 接口处理器。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"casbin-demo/internal/model"
	"casbin-demo/internal/service"
	"casbin-demo/pkg/response"
)

// RoleHandler 角色管理 HTTP 处理器。
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler 创建 RoleHandler 实例。
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// List 查询所有角色接口（需 JWT + Casbin 权限）。
// GET /api/roles
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.roleService.List()
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, roles)
}

// Create 创建角色接口（需 JWT + Casbin 权限）。
// POST /api/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	role, err := h.roleService.Create(req)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, role)
}

// Update 更新角色信息接口（需 JWT + Casbin 权限）。
// PUT /api/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.roleService.Update(uint(id), req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除角色接口（需 JWT + Casbin 权限）。
// DELETE /api/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	if err := h.roleService.Delete(uint(id)); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// AssignPermission 为角色分配权限接口（需 JWT + Casbin 权限）。
// POST /api/roles/assign-permission
func (h *RoleHandler) AssignPermission(c *gin.Context) {
	var req model.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.roleService.AssignPermission(req.RoleID, req.PermissionID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}
