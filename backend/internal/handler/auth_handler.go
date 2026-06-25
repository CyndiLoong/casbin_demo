// Package handler 实现 HTTP 请求处理层（Controller/Presentation Layer）。
//
// 本层职责：
//   - 接收并解析 HTTP 请求（参数绑定、校验）
//   - 调用 Service 层处理业务逻辑
//   - 构造统一格式的 HTTP 响应
//   - 不包含业务逻辑，只做参数转换和响应封装
package handler

import (
	"github.com/gin-gonic/gin"

	"casbin-demo/internal/model"
	"casbin-demo/internal/service"
	"casbin-demo/pkg/response"
)

// AuthHandler 认证相关 HTTP 处理器。
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建 AuthHandler 实例。
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login 用户登录接口。
// POST /api/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.authService.Login(req)
	if err != nil {
		response.Fail(c, 401, err.Error())
		return
	}

	response.Success(c, result)
}

// Register 用户注册接口。
// POST /api/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	response.SuccessWithMessage(c, "注册成功", gin.H{
		"id":       user.ID,
		"username": user.Username,
		"uuid":     user.UUID,
	})
}

// GetUserInfo 获取当前用户信息接口（需 JWT 认证）。
// GET /api/userinfo
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, _ := c.Get("user_id")
	info, err := h.authService.GetUserInfo(userID.(uint))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}

	response.Success(c, info)
}
