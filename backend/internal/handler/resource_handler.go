// Package handler 实现 HTTP 请求处理层（Controller/Presentation Layer）。
//
// 本文件处理资源清单相关的 HTTP 请求。
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"casbin-demo/internal/model"
	"casbin-demo/internal/service"
	"casbin-demo/pkg/response"
)

// ResourceHandler 资源相关 HTTP 处理器。
type ResourceHandler struct {
	resourceService *service.ResourceService
}

// NewResourceHandler 创建 ResourceHandler 实例。
func NewResourceHandler(resourceService *service.ResourceService) *ResourceHandler {
	return &ResourceHandler{resourceService: resourceService}
}

// ListResources 分页查询资源列表接口（需 JWT 认证）。
// GET /api/resources
//
// 支持按类型、状态、关键词筛选，分页返回资源列表。
// 普通用户和管理员都可以访问，用于资源清单展示。
func (h *ResourceHandler) ListResources(c *gin.Context) {
	var req model.ResourceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	var status *int
	if c.Query("status") != "" {
		s, err := strconv.Atoi(c.Query("status"))
		if err == nil {
			status = &s
		}
	}

	list, total, err := h.resourceService.GetResourceList(
		req.Page, req.PageSize, req.Type, status, req.Keyword,
	)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetResource 获取资源详情接口（需 JWT 认证）。
// GET /api/resources/:id
//
// 根据资源ID获取资源详细信息。
func (h *ResourceHandler) GetResource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的资源ID")
		return
	}

	resource, err := h.resourceService.GetResourceByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, resource)
}

// ListActiveResources 获取可用资源列表接口（需 JWT 认证）。
// GET /api/resources/active
//
// 获取所有状态为可用的资源，用于资源清单展示。
func (h *ResourceHandler) ListActiveResources(c *gin.Context) {
	resources, err := h.resourceService.GetActiveResources()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  resources,
		"total": len(resources),
	})
}

// CreateResource 创建资源接口（需 JWT + Casbin 权限）。
// POST /api/resources
//
// 管理员创建新的大模型API资源。
func (h *ResourceHandler) CreateResource(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		ResourceType string `json:"resource_type" binding:"required"`
		APIName      string `json:"api_name" binding:"required"`
		Description  string `json:"description"`
		Provider     string `json:"provider"`
		Version      string `json:"version"`
		DefaultQPS   int    `json:"default_qps"`
		MaxQPS       int    `json:"max_qps"`
		DocsURL      string `json:"docs_url"`
		Tags         string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	resource, err := h.resourceService.CreateResource(
		req.Name, req.ResourceType, req.APIName, req.Description,
		req.Provider, req.Version, req.DefaultQPS, req.MaxQPS,
		req.DocsURL, req.Tags,
	)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, resource)
}

// UpdateResource 更新资源接口（需 JWT + Casbin 权限）。
// PUT /api/resources/:id
//
// 管理员更新资源信息。
func (h *ResourceHandler) UpdateResource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的资源ID")
		return
	}

	var req struct {
		Name         string `json:"name" binding:"required"`
		ResourceType string `json:"resource_type" binding:"required"`
		APIName      string `json:"api_name" binding:"required"`
		Description  string `json:"description"`
		Provider     string `json:"provider"`
		Version      string `json:"version"`
		DefaultQPS   int    `json:"default_qps"`
		MaxQPS       int    `json:"max_qps"`
		Status       int    `json:"status"`
		DocsURL      string `json:"docs_url"`
		Tags         string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	resource, err := h.resourceService.UpdateResource(
		uint(id), req.Name, req.ResourceType, req.APIName, req.Description,
		req.Provider, req.Version, req.DefaultQPS, req.MaxQPS, req.Status,
		req.DocsURL, req.Tags,
	)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, resource)
}

// DeleteResource 删除资源接口（需 JWT + Casbin 权限）。
// DELETE /api/resources/:id
//
// 管理员删除资源（软删除）。
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的资源ID")
		return
	}

	if err := h.resourceService.DeleteResource(uint(id)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
