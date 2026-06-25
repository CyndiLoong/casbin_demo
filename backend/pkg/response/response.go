// Package response 提供统一的 HTTP 响应格式封装。
//
// 设计原则：HTTP 状态码与响应体中的 code 字段保持一致。
//   - 成功响应：HTTP 200，code=200
//   - 错误响应：HTTP 状态码与 code 字段相同（如 400/401/403/404/500）
//
// 统一响应结构：
//
//	{
//	  "code": 200,        // 业务码，与 HTTP 状态码一致：200=成功，非200=失败
//	  "message": "success", // 消息
//	  "data": {}          // 数据（成功时返回，失败时省略）
//	}
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一 API 响应结构体。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 返回成功响应（HTTP 200，code=200）。
// 请求成功处理并能够返回期望结果时使用。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 返回带自定义消息的成功响应（HTTP 200，code=200）。
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// Fail 返回失败响应。
// HTTP 状态码与 code 参数一致，响应体中的 code 字段也使用该值。
//
// 参数 code 应为标准 HTTP 状态码（如 400/401/403/404/500），
// HTTP 传输层状态码和响应体业务码保持一致，便于客户端处理。
func Fail(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// FailWithStatus 返回带自定义 HTTP 状态码的失败响应。
// httpStatus 用于 HTTP 传输层状态码，code 用于响应体业务码。
// 适用于业务码与 HTTP 状态码需要分离的场景。
func FailWithStatus(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// Unauthorized 返回未授权响应（HTTP 401，code=401）。
// 用于认证失败：Token缺失/无效/过期等场景。
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    http.StatusUnauthorized,
		Message: message,
	})
}

// Forbidden 返回禁止访问响应（HTTP 403，code=403）。
// 用于已认证但无权限访问资源的场景。
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    http.StatusForbidden,
		Message: message,
	})
}

// BadRequest 返回请求参数错误响应（HTTP 400，code=400）。
// 用于请求参数校验失败、格式错误等场景。
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}

// ServerError 返回服务器内部错误响应（HTTP 500，code=500）。
// 用于服务器内部异常、数据库操作失败等非预期错误。
func ServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: message,
	})
}

// NotFound 返回资源未找到响应（HTTP 404，code=404）。
// 用于请求的资源不存在。
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    http.StatusNotFound,
		Message: message,
	})
}

// MethodNotAllowed 返回方法不允许响应（HTTP 405，code=405）。
// 用于请求方法不被允许（如用 POST 访问只支持 GET 的接口）。
func MethodNotAllowed(c *gin.Context, message string) {
	c.JSON(http.StatusMethodNotAllowed, Response{
		Code:    http.StatusMethodNotAllowed,
		Message: message,
	})
}
