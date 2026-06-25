package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"casbin-demo/internal/model"
	"casbin-demo/internal/repository"
	"casbin-demo/pkg/cache"
	"casbin-demo/pkg/mq"
	"casbin-demo/pkg/ws"
)

const (
	auditAppListCacheKey  = "cache:audit:list"
	auditPendingCountKey  = "cache:audit:pending_count"
	msgUnreadKeyPrefix    = "cache:msg:unread:"
	withdrawCachePrefix   = "cache:audit:withdraw:"
	withdrawWindowDuration = 2 * time.Minute
	retryInterval         = 30 * time.Second
	retryBatchSize        = 50
	mqPublishTimeout      = 3 * time.Second
)

type AuditService struct {
	auditRepo *repository.AuditRepository
	userRepo  *repository.UserRepository
	cache     *cache.Client
	rdb       *redis.Client
	mqClient  *mq.Client
	wsHub     *ws.Hub
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

func NewAuditService(
	auditRepo *repository.AuditRepository,
	userRepo *repository.UserRepository,
	cacheClient *cache.Client,
	rdb *redis.Client,
	mqClient *mq.Client,
	wsHub *ws.Hub,
) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
		userRepo:  userRepo,
		cache:     cacheClient,
		rdb:       rdb,
		mqClient:  mqClient,
		wsHub:     wsHub,
		stopCh:    make(chan struct{}),
	}
}

func getStatusText(status int) string {
	switch status {
	case model.AuditStatusPending:
		return "待审核"
	case model.AuditStatusApproved:
		return "已通过"
	case model.AuditStatusRejected:
		return "已驳回"
	case model.AuditStatusWithdrawn:
		return "已撤回"
	default:
		return "未知"
	}
}

func (s *AuditService) canWithdraw(app *model.AuditApplication) bool {
	if app.Status != model.AuditStatusPending {
		return false
	}
	if s.rdb != nil {
		key := fmt.Sprintf("%s%d", withdrawCachePrefix, app.ID)
		exists, err := s.rdb.Exists(context.Background(), key).Result()
		if err == nil && exists > 0 {
			return true
		}
	}
	return time.Since(app.CreatedAt) < withdrawWindowDuration
}

func (s *AuditService) setWithdrawTTL(appID uint) {
	if s.rdb == nil {
		return
	}
	key := fmt.Sprintf("%s%d", withdrawCachePrefix, appID)
	if err := s.rdb.Set(context.Background(), key, "1", withdrawWindowDuration).Err(); err != nil {
		slog.Warn("set withdraw ttl failed", "app_id", appID, "error", err)
	}
}

func (s *AuditService) delWithdrawTTL(appID uint) {
	if s.rdb == nil {
		return
	}
	key := fmt.Sprintf("%s%d", withdrawCachePrefix, appID)
	if err := s.rdb.Del(context.Background(), key).Err(); err != nil {
		slog.Warn("delete withdraw ttl failed", "app_id", appID, "error", err)
	}
}

func (s *AuditService) toAuditResponse(app *model.AuditApplication) model.AuditApplicationResponse {
	canWithdraw := s.canWithdraw(app)
	var withdrawRemain int64
	if canWithdraw {
		elapsed := time.Since(app.CreatedAt)
		remain := withdrawWindowDuration - elapsed
		if remain > 0 {
			withdrawRemain = remain.Milliseconds()
		}
	}
	resp := model.AuditApplicationResponse{
		ID:             app.ID,
		UUID:           app.UUID,
		ApplicantID:    app.ApplicantID,
		ApplicantName:  app.ApplicantName,
		ResourceName:   app.ResourceName,
		ResourceType:   app.ResourceType,
		APIName:        app.APIName,
		APIDescription: app.APIDescription,
		Purpose:        app.Purpose,
		ExpectedQPS:    app.ExpectedQPS,
		ContactInfo:    app.ContactInfo,
		Status:         app.Status,
		StatusText:     getStatusText(app.Status),
		CanWithdraw:    canWithdraw,
		WithdrawRemain: withdrawRemain,
		ReviewerName:   app.ReviewerName,
		ReviewComment:  app.ReviewComment,
		WithdrawReason: app.WithdrawReason,
		CreatedAt:      app.CreatedAt,
	}
	if app.ReviewerID != nil {
		resp.ReviewerID = app.ReviewerID
	}
	if app.ReviewedAt != nil {
		resp.ReviewedAt = app.ReviewedAt
	}
	if app.WithdrawnAt != nil {
		resp.WithdrawnAt = app.WithdrawnAt
	}
	return resp
}

func toMessageResponse(msg *model.SysMessage) model.SysMessageResponse {
	return model.SysMessageResponse{
		ID:           msg.ID,
		UUID:         msg.UUID,
		Type:         msg.Type,
		Title:        msg.Title,
		Content:      msg.Content,
		BusinessType: msg.BusinessType,
		BusinessID:   msg.BusinessID,
		IsRead:       msg.IsRead,
		CreatedAt:    msg.CreatedAt,
	}
}

func (s *AuditService) publishMQ(ctx context.Context, msg mq.NotificationMessage) error {
	if s.mqClient == nil || !s.mqClient.IsConnected() {
		return errors.New("mq not available")
	}
	ctx, cancel := context.WithTimeout(ctx, mqPublishTimeout)
	defer cancel()
	return s.mqClient.Publish(ctx, msg)
}

func (s *AuditService) mqNotifyAdmins(ctx context.Context, msg *model.SysMessage) {
	mqMsg := mq.NotificationMessage{
		MessageID:      msg.ID,
		TargetType:     model.TargetTypeAdmins,
		Type:           msg.Type,
		BusinessID:     msg.BusinessID,
		BusinessType:   msg.BusinessType,
		CreatedAt:      msg.CreatedAt,
		IdempotencyKey: strconv.FormatUint(uint64(msg.ID), 10),
	}
	if err := s.publishMQ(ctx, mqMsg); err != nil {
		slog.Warn("mq publish admin notification failed, will retry", "msg_id", msg.ID, "error", err)
	}
}

func (s *AuditService) mqNotifyUser(ctx context.Context, msg *model.SysMessage) {
	mqMsg := mq.NotificationMessage{
		MessageID:      msg.ID,
		TargetType:     model.TargetTypeUser,
		TargetID:       msg.ReceiverID,
		Type:           msg.Type,
		BusinessID:     msg.BusinessID,
		BusinessType:   msg.BusinessType,
		CreatedAt:      msg.CreatedAt,
		IdempotencyKey: strconv.FormatUint(uint64(msg.ID), 10),
	}
	if err := s.publishMQ(ctx, mqMsg); err != nil {
		slog.Warn("mq publish user notification failed, will retry", "msg_id", msg.ID, "error", err)
	}
}

func (s *AuditService) invalidateAuditCaches() {
	s.cache.Delete(auditPendingCountKey)
	s.cache.DeleteByPattern(auditAppListCacheKey + "*")
}

func (s *AuditService) SubmitApplication(ctx context.Context, userID uint, req *model.CreateAuditRequest) (*model.AuditApplicationResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	applicantName := user.Nickname
	if applicantName == "" {
		applicantName = user.Username
	}

	now := time.Now()
	app := &model.AuditApplication{
		UUID:           uuid.New().String(),
		ApplicantID:    userID,
		ApplicantName:  applicantName,
		ResourceName:   req.ResourceName,
		ResourceType:   req.ResourceType,
		APIName:        req.APIName,
		APIDescription: req.APIDescription,
		Purpose:        req.Purpose,
		ExpectedQPS:    req.ExpectedQPS,
		ContactInfo:    req.ContactInfo,
		Status:         model.AuditStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	adminIDs, err := s.userRepo.FindAdminUserIDs()
	if err != nil {
		slog.Warn("find admin ids failed", "error", err)
		adminIDs = nil
	}

	tx := s.auditRepo.BeginTx()
	if tx.Error != nil {
		return nil, errors.New("开启事务失败")
	}
	if err := s.auditRepo.Create(tx, app); err != nil {
		tx.Rollback()
		slog.Error("create audit application failed", "error", err)
		return nil, errors.New("提交申请失败")
	}

	adminMsgContent := fmt.Sprintf("%s 提交了「%s」资源的API使用申请，请及时审核。", applicantName, req.ResourceName)
	var adminMsgs []*model.SysMessage
	for _, aid := range adminIDs {
		adminMsgs = append(adminMsgs, &model.SysMessage{
			UUID:         uuid.New().String(),
			ReceiverID:   aid,
			Type:         model.MsgTypeNewApplication,
			Title:        "新的大模型API审核申请",
			Content:      adminMsgContent,
			BusinessType: model.BusinessTypeAudit,
			BusinessID:   app.ID,
			IsRead:       false,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if len(adminMsgs) > 0 {
		if err := s.auditRepo.CreateMessagesBatch(tx, adminMsgs); err != nil {
			tx.Rollback()
			slog.Error("create admin messages failed", "error", err)
			return nil, errors.New("创建消息失败")
		}
	}
	if err := tx.Commit().Error; err != nil {
		slog.Error("commit transaction failed", "error", err)
		return nil, errors.New("提交事务失败")
	}

	s.setWithdrawTTL(app.ID)
	for _, m := range adminMsgs {
		s.mqNotifyAdmins(ctx, m)
	}

	if s.wsHub != nil {
		notif := model.WsNotification{
			Type:         model.MsgTypeNewApplication,
			Title:        "新的审核申请",
			Content:      adminMsgContent,
			BusinessType: model.BusinessTypeAudit,
			BusinessID:   app.ID,
			Timestamp:    time.Now(),
			ID:           fmt.Sprintf("new-app-%d", app.ID),
			Data: map[string]interface{}{
				"application_id": app.ID,
				"applicant":      applicantName,
				"resource":       req.ResourceName,
			},
		}
		s.wsHub.SendToAdminsLocal(ws.WsMessage{
			Type:       "notification",
			TargetType: "room",
			Room:       ws.AdminRoom,
			Data:       notif,
			Timestamp:  time.Now(),
		})
	}

	s.invalidateAuditCaches()
	s.cache.DeleteByPattern(msgUnreadKeyPrefix + "*")

	slog.Info("audit submitted", "app_id", app.ID, "user_id", userID)
	resp := s.toAuditResponse(app)
	return &resp, nil
}

func (s *AuditService) ReviewApplication(ctx context.Context, reviewerID uint, appID uint, req *model.ReviewAuditRequest) error {
	app, err := s.auditRepo.FindByID(appID)
	if err != nil {
		return errors.New("申请不存在")
	}
	if app.Status != model.AuditStatusPending {
		return errors.New("该申请已被处理")
	}
	reviewer, err := s.userRepo.FindByID(reviewerID)
	if err != nil {
		return errors.New("审核人信息不存在")
	}
	reviewerName := reviewer.Nickname
	if reviewerName == "" {
		reviewerName = reviewer.Username
	}

	status := model.AuditStatusRejected
	statusText := "驳回"
	if req.Approved {
		status = model.AuditStatusApproved
		statusText = "通过"
	}
	title := "审核结果通知"
	content := fmt.Sprintf("您提交的「%s」资源API申请已%s。", app.ResourceName, statusText)
	if req.Comment != "" {
		content += fmt.Sprintf(" 审核意见：%s", req.Comment)
	}

	now := time.Now()
	rid := reviewerID
	tx := s.auditRepo.BeginTx()
	if tx.Error != nil {
		return errors.New("开启事务失败")
	}
	if err := s.auditRepo.UpdateStatusWithVersion(tx, appID, app.Version, status, &rid, reviewerName, req.Comment); err != nil {
		tx.Rollback()
		if errors.Is(err, repository.ErrOptimisticLockConflict) {
			return err
		}
		if errors.Is(err, repository.ErrNotPending) {
			return errors.New("该申请已被处理")
		}
		slog.Error("update status failed", "app_id", appID, "error", err)
		return errors.New("更新审核状态失败")
	}

	userMsg := &model.SysMessage{
		UUID:         uuid.New().String(),
		ReceiverID:   app.ApplicantID,
		Type:         model.MsgTypeReviewResult,
		Title:        title,
		Content:      content,
		BusinessType: model.BusinessTypeAudit,
		BusinessID:   appID,
		IsRead:       false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.auditRepo.CreateMessage(tx, userMsg); err != nil {
		tx.Rollback()
		slog.Error("create user message failed", "error", err)
		return errors.New("创建消息失败")
	}
	if err := tx.Commit().Error; err != nil {
		slog.Error("commit review transaction failed", "error", err)
		return errors.New("提交事务失败")
	}

	s.delWithdrawTTL(appID)
	s.mqNotifyUser(ctx, userMsg)

	if s.wsHub != nil {
		notif := model.WsNotification{
			Type:         model.MsgTypeReviewResult,
			Title:        title,
			Content:      content,
			BusinessType: model.BusinessTypeAudit,
			BusinessID:   appID,
			Timestamp:    time.Now(),
			ID:           fmt.Sprintf("review-%d", appID),
			Data: map[string]interface{}{
				"application_id": appID,
				"status":         status,
				"approved":       req.Approved,
			},
		}
		s.wsHub.SendToUserLocal(app.ApplicantID, ws.WsMessage{
			Type:       "notification",
			TargetType: "user",
			TargetID:   app.ApplicantID,
			Data:       notif,
			Timestamp:  time.Now(),
		})
	}

	s.invalidateAuditCaches()
	s.cache.Delete(fmt.Sprintf("%s%d", msgUnreadKeyPrefix, app.ApplicantID))
	slog.Info("audit reviewed", "app_id", appID, "approved", req.Approved)
	return nil
}

func (s *AuditService) WithdrawApplication(ctx context.Context, userID uint, appID uint, req *model.WithdrawAuditRequest) error {
	app, err := s.auditRepo.FindByID(appID)
	if err != nil {
		return errors.New("申请不存在")
	}
	if app.ApplicantID != userID {
		return errors.New("无权撤回此申请")
	}
	if app.Status != model.AuditStatusPending {
		return errors.New("该申请已被处理，无法撤回")
	}
	if !s.canWithdraw(app) {
		return repository.ErrWithdrawWindowExpired
	}

	now := time.Now()
	tx := s.auditRepo.BeginTx()
	if tx.Error != nil {
		return errors.New("开启事务失败")
	}
	if err := s.auditRepo.WithdrawApplication(tx, appID, userID, app.Version, req.Reason); err != nil {
		tx.Rollback()
		switch {
		case errors.Is(err, repository.ErrNotOwner):
			return errors.New("无权撤回此申请")
		case errors.Is(err, repository.ErrNotPending):
			return errors.New("该申请已被处理，无法撤回")
		case errors.Is(err, repository.ErrOptimisticLockConflict):
			return errors.New("申请状态已变更，请刷新后重试")
		case errors.Is(err, repository.ErrWithdrawWindowExpired):
			return repository.ErrWithdrawWindowExpired
		default:
			slog.Error("withdraw failed", "app_id", appID, "error", err)
			return errors.New("撤回失败，请稍后重试")
		}
	}

	adminIDs, _ := s.userRepo.FindAdminUserIDs()
	var wdMsgs []*model.SysMessage
	wdContent := fmt.Sprintf("%s 撤回了「%s」资源的API使用申请。", app.ApplicantName, app.ResourceName)
	for _, aid := range adminIDs {
		wdMsgs = append(wdMsgs, &model.SysMessage{
			UUID:         uuid.New().String(),
			ReceiverID:   aid,
			Type:         model.MsgTypeApplicationWithdrawn,
			Title:        "申请已撤回",
			Content:      wdContent,
			BusinessType: model.BusinessTypeAudit,
			BusinessID:   appID,
			IsRead:       false,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if len(wdMsgs) > 0 {
		if err := s.auditRepo.CreateMessagesBatch(tx, wdMsgs); err != nil {
			tx.Rollback()
			slog.Error("create withdraw messages failed", "error", err)
			return errors.New("撤回失败，请稍后重试")
		}
	}
	if err := tx.Commit().Error; err != nil {
		slog.Error("commit withdraw transaction failed", "error", err)
		return errors.New("撤回失败，请稍后重试")
	}

	s.delWithdrawTTL(appID)
	for _, m := range wdMsgs {
		s.mqNotifyAdmins(ctx, m)
	}

	if s.wsHub != nil && len(wdMsgs) > 0 {
		notif := model.WsNotification{
			Type:         model.MsgTypeApplicationWithdrawn,
			Title:        "申请已撤回",
			Content:      wdContent,
			BusinessType: model.BusinessTypeAudit,
			BusinessID:   appID,
			Timestamp:    time.Now(),
			ID:           fmt.Sprintf("wd-%d", appID),
			Data: map[string]interface{}{
				"application_id": appID,
				"applicant":      app.ApplicantName,
				"resource":       app.ResourceName,
			},
		}
		s.wsHub.SendToAdminsLocal(ws.WsMessage{
			Type:       "notification",
			TargetType: "room",
			Room:       ws.AdminRoom,
			Data:       notif,
			Timestamp:  time.Now(),
		})
	}

	if s.wsHub != nil {
		confirm := model.WsNotification{
			Type:         "withdraw_confirmed",
			Title:        "撤回成功",
			Content:      fmt.Sprintf("您已成功撤回「%s」的API申请", app.ResourceName),
			BusinessType: model.BusinessTypeAudit,
			BusinessID:   appID,
			Timestamp:    time.Now(),
			ID:           fmt.Sprintf("wd-ok-%d", appID),
		}
		s.wsHub.SendToUserLocal(userID, ws.WsMessage{
			Type:       "notification",
			TargetType: "user",
			TargetID:   userID,
			Data:       confirm,
			Timestamp:  time.Now(),
		})
	}

	s.invalidateAuditCaches()
	s.cache.DeleteByPattern(msgUnreadKeyPrefix + "*")
	slog.Info("audit withdrawn", "app_id", appID, "user_id", userID)
	return nil
}

func (s *AuditService) GetApplication(id uint) (*model.AuditApplicationResponse, error) {
	app, err := s.auditRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("申请不存在")
	}
	resp := s.toAuditResponse(app)
	return &resp, nil
}

func (s *AuditService) ListMyApplications(userID uint, page, pageSize int, status *int) ([]model.AuditApplicationResponse, int64, error) {
	apps, total, err := s.auditRepo.ListByApplicant(userID, page, pageSize, status)
	if err != nil {
		return nil, 0, errors.New("查询申请列表失败")
	}
	result := make([]model.AuditApplicationResponse, 0, len(apps))
	for i := range apps {
		result = append(result, s.toAuditResponse(&apps[i]))
	}
	return result, total, nil
}

func (s *AuditService) ListAllApplications(page, pageSize int, status *int, applicant string) ([]model.AuditApplicationResponse, int64, error) {
	apps, total, err := s.auditRepo.ListAll(page, pageSize, status, applicant)
	if err != nil {
		return nil, 0, errors.New("查询申请列表失败")
	}
	result := make([]model.AuditApplicationResponse, 0, len(apps))
	for i := range apps {
		result = append(result, s.toAuditResponse(&apps[i]))
	}
	return result, total, nil
}

func (s *AuditService) GetUnreadCount(userID uint, isAdmin bool) (*model.UnreadCountResponse, error) {
	key := fmt.Sprintf("%s%d", msgUnreadKeyPrefix, userID)
	var unread int64
	opt := cache.DefaultFetchOptions(cache.TTLHot)
	opt.UseLocalCache = true
	opt.LocalTTL = 30 * time.Second
	v, _, err := s.cache.Fetch(key, opt, &unread, func() (interface{}, error) {
		return s.auditRepo.CountUnreadMessages(userID)
	})
	if err != nil {
		return nil, err
	}
	if c, ok := v.(int64); ok {
		unread = c
	}
	var pending int64
	if isAdmin {
		pending, _ = s.GetPendingCount()
	}
	return &model.UnreadCountResponse{UnreadCount: unread, PendingCount: pending}, nil
}

func (s *AuditService) GetPendingCount() (int64, error) {
	var count int64
	opt := cache.DefaultFetchOptions(cache.TTLDashboard)
	opt.UseLocalCache = true
	opt.LocalTTL = 30 * time.Second
	v, _, err := s.cache.Fetch(auditPendingCountKey, opt, &count, func() (interface{}, error) {
		return s.auditRepo.CountPending()
	})
	if err != nil {
		return 0, err
	}
	if c, ok := v.(int64); ok {
		return c, nil
	}
	return count, nil
}

func (s *AuditService) ListMessages(userID uint, page, pageSize int, unread *bool) ([]model.SysMessageResponse, int64, error) {
	msgs, total, err := s.auditRepo.ListMessages(userID, page, pageSize, unread)
	if err != nil {
		return nil, 0, errors.New("查询消息列表失败")
	}
	result := make([]model.SysMessageResponse, 0, len(msgs))
	for i := range msgs {
		result = append(result, toMessageResponse(&msgs[i]))
	}
	return result, total, nil
}

func (s *AuditService) ListAdminMessages(page, pageSize int) ([]model.SysMessageResponse, int64, error) {
	msgs, total, err := s.auditRepo.ListAdminMessages(page, pageSize)
	if err != nil {
		return nil, 0, errors.New("查询管理员消息列表失败")
	}
	result := make([]model.SysMessageResponse, 0, len(msgs))
	for i := range msgs {
		result = append(result, toMessageResponse(&msgs[i]))
	}
	return result, total, nil
}

func (s *AuditService) MarkMessageRead(userID, messageID uint, isAdmin bool) error {
	_, err := s.auditRepo.FindMessageByReceiver(messageID, userID)
	if err != nil {
		return errors.New("消息不存在或无权操作")
	}
	if err := s.auditRepo.MarkMessageRead(messageID, userID); err != nil {
		return errors.New("标记已读失败")
	}
	s.cache.Delete(fmt.Sprintf("%s%d", msgUnreadKeyPrefix, userID))
	if isAdmin {
		s.cache.DeleteByPattern(msgUnreadKeyPrefix + "*")
	}
	return nil
}

func (s *AuditService) MarkAllMessagesRead(userID uint, isAdmin bool) error {
	if err := s.auditRepo.MarkAllMessagesRead(userID); err != nil {
		return errors.New("标记已读失败")
	}
	s.cache.Delete(fmt.Sprintf("%s%d", msgUnreadKeyPrefix, userID))
	if isAdmin {
		s.cache.DeleteByPattern(msgUnreadKeyPrefix + "*")
	}
	return nil
}

func (s *AuditService) HandleMQMessage(msg mq.NotificationMessage) error {
	var sysMsg *model.SysMessage
	var err error
	if msg.TargetType == model.TargetTypeAdmins {
		sysMsg, err = s.auditRepo.FindMessageByID(msg.MessageID)
	} else {
		sysMsg, err = s.auditRepo.FindMessageByReceiver(msg.MessageID, msg.TargetID)
	}
	if err != nil {
		slog.Warn("mq: message not found, ack", "msg_id", msg.MessageID, "error", err)
		return nil
	}

	notif := model.WsNotification{
		Type:         sysMsg.Type,
		Title:        sysMsg.Title,
		Content:      sysMsg.Content,
		BusinessType: sysMsg.BusinessType,
		BusinessID:   sysMsg.BusinessID,
		Timestamp:    time.Now(),
		ID:           msg.IdempotencyKey,
	}
	if s.wsHub != nil {
		wsMsg := ws.WsMessage{Type: "notification", Timestamp: time.Now(), Data: notif}
		switch msg.TargetType {
		case model.TargetTypeAdmins:
			wsMsg.TargetType = "room"
			wsMsg.Room = ws.AdminRoom
			s.wsHub.SendToAdminsLocal(wsMsg)
		case model.TargetTypeUser:
			wsMsg.TargetType = "user"
			wsMsg.TargetID = sysMsg.ReceiverID
			s.wsHub.SendToUserLocal(sysMsg.ReceiverID, wsMsg)
		}
	}

	if err := s.auditRepo.MarkMessageMQDelivered(msg.MessageID); err != nil {
		slog.Warn("mq: mark delivered failed", "msg_id", msg.MessageID, "error", err)
	}

	if msg.TargetType == model.TargetTypeAdmins {
		s.cache.Delete(auditPendingCountKey)
		s.cache.DeleteByPattern(msgUnreadKeyPrefix + "*")
	} else {
		s.cache.Delete(fmt.Sprintf("%s%d", msgUnreadKeyPrefix, sysMsg.ReceiverID))
	}
	return nil
}

func (s *AuditService) SetMQHandler() {
	if s.mqClient == nil {
		slog.Warn("mq not available, running in PG-only mode")
		return
	}
	s.mqClient.SetHandler(s.HandleMQMessage)
	slog.Info("mq consumer handler registered")
}

func (s *AuditService) StartRetryScheduler() {
	s.wgGo(func() {
		ticker := time.NewTicker(retryInterval)
		defer ticker.Stop()
		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.retryUndelivered()
			}
		}
	})
	slog.Info("mq retry scheduler started", "interval", retryInterval)
}

func (s *AuditService) wgGo(fn func()) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				slog.Error("audit service panic", "panic", r)
			}
		}()
		fn()
	}()
}

func (s *AuditService) retryUndelivered() {
	if s.mqClient == nil || !s.mqClient.IsConnected() {
		return
	}
	msgs, err := s.auditRepo.FindUnDeliveredMessages(retryBatchSize)
	if err != nil || len(msgs) == 0 {
		return
	}
	slog.Info("retry undelivered messages", "count", len(msgs))
	ctx := context.Background()
	delivered := make([]uint, 0, len(msgs))
	for i := range msgs {
		m := &msgs[i]
		ttype := model.TargetTypeUser
		tid := m.ReceiverID
		if m.Type == model.MsgTypeNewApplication || m.Type == model.MsgTypeApplicationWithdrawn {
			ttype = model.TargetTypeAdmins
			tid = 0
		}
		mqMsg := mq.NotificationMessage{
			MessageID:      m.ID,
			TargetType:     ttype,
			TargetID:       tid,
			Type:           m.Type,
			BusinessID:     m.BusinessID,
			BusinessType:   m.BusinessType,
			CreatedAt:      m.CreatedAt,
			IdempotencyKey: strconv.FormatUint(uint64(m.ID), 10),
		}
		if err := s.publishMQ(ctx, mqMsg); err != nil {
			slog.Warn("retry publish failed", "msg_id", m.ID, "error", err)
			continue
		}
		delivered = append(delivered, m.ID)
	}
	if len(delivered) > 0 {
		_ = s.auditRepo.MarkMessagesDelivered(delivered)
	}
}

func (s *AuditService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	slog.Info("audit service stopped")
}
