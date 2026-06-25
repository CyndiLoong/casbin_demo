// permission_handler.go 实现权限管理相关的 HTTP 接口处理器。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"casbin-demo/internal/model"
	"casbin-demo/internal/service"
	"casbin-demo/pkg/response"
)

// PermissionHandler 权限管理 HTTP 处理器。
type PermissionHandler struct {
	permService *service.PermissionService
}

// NewPermissionHandler 创建 PermissionHandler 实例。
func NewPermissionHandler(permService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permService: permService}
}

// List 查询所有权限接口（需 JWT + Casbin 权限）。
// GET /api/permissions
func (h *PermissionHandler) List(c *gin.Context) {
	perms, err := h.permService.List()
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, perms)
}

// Create 创建权限接口（需 JWT + Casbin 权限）。
// POST /api/permissions
func (h *PermissionHandler) Create(c *gin.Context) {
	var req model.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	perm, err := h.permService.Create(req)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, perm)
}

// Update 更新权限信息接口（需 JWT + Casbin 权限）。
// PUT /api/permissions/:id
func (h *PermissionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}

	var req model.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.permService.Update(uint(id), req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除权限接口（需 JWT + Casbin 权限）。
// DELETE /api/permissions/:id
func (h *PermissionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}

	if err := h.permService.Delete(uint(id)); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}
