package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"casbin-demo/internal/model"
	"casbin-demo/internal/service"
	"casbin-demo/pkg/response"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

func getCurrentUserID(c *gin.Context) (uint, bool) {
	uid, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := uid.(uint)
	return id, ok
}

func getCurrentUserRoles(c *gin.Context) []string {
	roles, exists := c.Get("roles")
	if !exists {
		return nil
	}
	if r, ok := roles.([]string); ok {
		return r
	}
	return nil
}

func isAdmin(c *gin.Context) bool {
	roles := getCurrentUserRoles(c)
	for _, r := range roles {
		if r == "admin" {
			return true
		}
	}
	return false
}

// Submit 提交审核申请（普通用户）。
// POST /api/audit/applications
func (h *AuditHandler) Submit(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	var req model.CreateAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	app, err := h.auditService.SubmitApplication(c.Request.Context(), userID, &req)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, app)
}

// Review 审核申请（管理员）。
// POST /api/audit/applications/:id/review
func (h *AuditHandler) Review(c *gin.Context) {
	if !isAdmin(c) {
		response.Forbidden(c, "无权限执行审核操作")
		return
	}
	reviewerID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的申请ID")
		return
	}
	var req model.ReviewAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	if err := h.auditService.ReviewApplication(c.Request.Context(), reviewerID, uint(id), &req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.SuccessWithMessage(c, "审核完成", nil)
}

// Withdraw 撤回申请（普通用户，2分钟窗口内）。
// POST /api/audit/applications/:id/withdraw
func (h *AuditHandler) Withdraw(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的申请ID")
		return
	}
	var req model.WithdrawAuditRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.auditService.WithdrawApplication(c.Request.Context(), userID, uint(id), &req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.SuccessWithMessage(c, "撤回成功", nil)
}

// GetDetail 获取申请详情。
// GET /api/audit/applications/:id
func (h *AuditHandler) GetDetail(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的申请ID")
		return
	}
	app, err := h.auditService.GetApplication(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	if !isAdmin(c) && app.ApplicantID != userID {
		response.Forbidden(c, "无权查看此申请")
		return
	}
	response.Success(c, app)
}

// ListMyApplications 查询我的申请列表（普通用户）。
// GET /api/audit/my-applications
func (h *AuditHandler) ListMyApplications(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	var status *int
	if s := c.Query("status"); s != "" {
		if sv, err := strconv.Atoi(s); err == nil {
			status = &sv
		}
	}
	list, total, err := h.auditService.ListMyApplications(userID, page, pageSize, status)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListAllApplications 查询所有申请列表（管理员）。
// GET /api/audit/applications
func (h *AuditHandler) ListAllApplications(c *gin.Context) {
	if !isAdmin(c) {
		response.Forbidden(c, "无权限查看所有申请")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	var status *int
	if s := c.Query("status"); s != "" {
		if sv, err := strconv.Atoi(s); err == nil {
			status = &sv
		}
	}
	applicant := c.Query("applicant")
	list, total, err := h.auditService.ListAllApplications(page, pageSize, status, applicant)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetPendingCount 获取待审核数量（管理员）。
// GET /api/audit/pending-count
func (h *AuditHandler) GetPendingCount(c *gin.Context) {
	if !isAdmin(c) {
		response.Forbidden(c, "无权限")
		return
	}
	count, err := h.auditService.GetPendingCount()
	if err != nil {
		response.Fail(c, 500, "获取待审核数量失败")
		return
	}
	response.Success(c, gin.H{"count": count})
}

// GetUnreadCount 获取当前用户未读消息数（管理员额外返回待审数）。
// GET /api/messages/unread-count
func (h *AuditHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	admin := isAdmin(c)
	result, err := h.auditService.GetUnreadCount(userID, admin)
	if err != nil {
		response.Fail(c, 500, "获取未读数量失败")
		return
	}
	response.Success(c, result)
}

// ListMessages 获取当前用户的消息列表。
// GET /api/messages
func (h *AuditHandler) ListMessages(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	var unread *bool
	if u := c.Query("unread"); u == "true" {
		t := true
		unread = &t
	} else if u == "false" {
		f := false
		unread = &f
	}
	list, total, err := h.auditService.ListMessages(userID, page, pageSize, unread)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// MarkMessageRead 标记单条消息已读。
// PUT /api/messages/:id/read
func (h *AuditHandler) MarkMessageRead(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的消息ID")
		return
	}
	admin := isAdmin(c)
	if err := h.auditService.MarkMessageRead(userID, uint(id), admin); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// MarkAllRead 标记所有消息已读。
// PUT /api/messages/read-all
func (h *AuditHandler) MarkAllRead(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	admin := isAdmin(c)
	if err := h.auditService.MarkAllMessagesRead(userID, admin); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, nil)
}
