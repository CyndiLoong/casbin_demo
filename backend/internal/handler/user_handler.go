// user_handler.go 实现用户管理相关的 HTTP 接口处理器。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"casbin-demo/internal/model"
	"casbin-demo/internal/service"
	"casbin-demo/pkg/response"
)

// UserHandler 用户管理 HTTP 处理器。
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建 UserHandler 实例。
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List 分页查询用户列表接口（需 JWT + Casbin 权限）。
// GET /api/users?page=1&page_size=10
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.userService.List(page, pageSize)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create 创建用户接口（需 JWT + Casbin 权限）。
// POST /api/users
func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	user, err := h.userService.Create(req.Username, req.Password, req.Nickname, req.Email)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, gin.H{"id": user.ID, "username": user.Username})
}

// Update 更新用户信息接口（需 JWT + Casbin 权限）。
// PUT /api/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Status   int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.userService.Update(uint(id), req.Nickname, req.Email, req.Status); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除用户接口（需 JWT + Casbin 权限）。
// DELETE /api/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.userService.Delete(uint(id)); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// AssignRole 为用户分配角色接口（需 JWT + Casbin 权限）。
// POST /api/users/assign-role
func (h *UserHandler) AssignRole(c *gin.Context) {
	var req model.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.userService.AssignRole(req.UserID, req.RoleID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}
