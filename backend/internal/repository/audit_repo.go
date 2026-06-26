package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"casbin-demo/internal/model"
)

var (
	ErrNotOwner               = errors.New("不是申请的提交者")
	ErrNotPending             = errors.New("申请已被处理")
	ErrOptimisticLockConflict = errors.New("数据版本冲突，请重试")
	ErrWithdrawWindowExpired  = errors.New("已超过撤回时间窗口")
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&model.AuditApplication{}, &model.SysMessage{})
}

func (r *AuditRepository) BeginTx() *gorm.DB {
	return r.db.Begin()
}

func (r *AuditRepository) Create(tx *gorm.DB, app *model.AuditApplication) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(app).Error
}

func (r *AuditRepository) CreateMessage(tx *gorm.DB, msg *model.SysMessage) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(msg).Error
}

func (r *AuditRepository) CreateMessagesBatch(tx *gorm.DB, msgs []*model.SysMessage) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(&msgs).Error
}

func (r *AuditRepository) FindByID(id uint) (*model.AuditApplication, error) {
	var app model.AuditApplication
	if err := r.db.First(&app, id).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *AuditRepository) FindByUUID(uuid string) (*model.AuditApplication, error) {
	var app model.AuditApplication
	if err := r.db.Where("uuid = ?", uuid).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *AuditRepository) ListByApplicant(applicantID uint, page, pageSize int, status *int) ([]model.AuditApplication, int64, error) {
	var apps []model.AuditApplication
	var total int64

	query := r.db.Model(&model.AuditApplication{}).Where("applicant_id = ?", applicantID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

func (r *AuditRepository) ListAll(page, pageSize int, status *int, excludePending bool, applicant string) ([]model.AuditApplication, int64, error) {
	var apps []model.AuditApplication
	var total int64

	query := r.db.Model(&model.AuditApplication{})
	if status != nil {
		query = query.Where("status = ?", *status)
	} else if excludePending {
		query = query.Where("status != ?", model.AuditStatusPending)
	}
	if applicant != "" {
		query = query.Where("applicant_name LIKE ?", "%"+applicant+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortColumn := "created_at"
	if status != nil && *status != model.AuditStatusPending {
		sortColumn = "COALESCE(reviewed_at, created_at)"
	} else if excludePending {
		sortColumn = "COALESCE(reviewed_at, created_at)"
	}
	offset := (page - 1) * pageSize
	if err := query.Order(sortColumn + " DESC").Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

func (r *AuditRepository) CountPending() (int64, error) {
	var count int64
	err := r.db.Model(&model.AuditApplication{}).Where("status = ?", model.AuditStatusPending).Count(&count).Error
	return count, err
}

func (r *AuditRepository) UpdateStatusWithVersion(tx *gorm.DB, id uint, currentVersion int, status int, reviewerID *uint, reviewerName, comment string) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	now := time.Now()
	result := db.Model(&model.AuditApplication{}).
		Where("id = ? AND version = ? AND status = ?", id, currentVersion, model.AuditStatusPending).
		Updates(map[string]interface{}{
			"status":         status,
			"reviewer_id":    reviewerID,
			"reviewer_name":  reviewerName,
			"review_comment": comment,
			"reviewed_at":    now,
			"version":        currentVersion + 1,
			"updated_at":     now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var app model.AuditApplication
		if err := db.First(&app, id).Error; err != nil {
			return err
		}
		if app.Status != model.AuditStatusPending {
			return ErrNotPending
		}
		return ErrOptimisticLockConflict
	}
	return nil
}

func (r *AuditRepository) WithdrawApplication(tx *gorm.DB, id, userID uint, currentVersion int, reason string) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	var app model.AuditApplication
	if err := db.First(&app, id).Error; err != nil {
		return err
	}

	if app.ApplicantID != userID {
		return ErrNotOwner
	}
	if app.Status != model.AuditStatusPending {
		return ErrNotPending
	}

	now := time.Now()
	result := db.Model(&model.AuditApplication{}).
		Where("id = ? AND version = ? AND status = ?", id, currentVersion, model.AuditStatusPending).
		Updates(map[string]interface{}{
			"status":          model.AuditStatusWithdrawn,
			"withdraw_reason": reason,
			"withdrawn_at":    now,
			"version":         currentVersion + 1,
			"updated_at":      now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var current model.AuditApplication
		if err := db.First(&current, id).Error; err != nil {
			return err
		}
		if current.Status != model.AuditStatusPending {
			return ErrNotPending
		}
		return ErrOptimisticLockConflict
	}
	return nil
}

func (r *AuditRepository) FindMessageByID(messageID uint) (*model.SysMessage, error) {
	var msg model.SysMessage
	if err := r.db.First(&msg, messageID).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *AuditRepository) FindMessageByReceiver(messageID, receiverID uint) (*model.SysMessage, error) {
	var msg model.SysMessage
	if err := r.db.Where("id = ? AND receiver_id = ?", messageID, receiverID).First(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *AuditRepository) MarkMessageRead(messageID, receiverID uint) error {
	return r.db.Model(&model.SysMessage{}).
		Where("id = ? AND receiver_id = ? AND is_read = ?", messageID, receiverID, false).
		Updates(map[string]interface{}{"is_read": true, "updated_at": time.Now()}).Error
}

func (r *AuditRepository) MarkAllMessagesRead(userID uint) error {
	return r.db.Model(&model.SysMessage{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{"is_read": true, "updated_at": time.Now()}).Error
}

func (r *AuditRepository) CountUnreadMessages(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.SysMessage{}).Where("receiver_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *AuditRepository) MarkMessageMQDelivered(messageID uint) error {
	return r.db.Model(&model.SysMessage{}).Where("id = ?", messageID).Update("mq_delivered", true).Error
}

func (r *AuditRepository) MarkMessagesDelivered(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.SysMessage{}).Where("id IN ?", ids).Update("mq_delivered", true).Error
}

func (r *AuditRepository) FindUnDeliveredMessages(limit int) ([]model.SysMessage, error) {
	var msgs []model.SysMessage
	cutoff := time.Now().Add(-10 * time.Second)
	err := r.db.Where("mq_delivered = ? AND created_at < ?", false, cutoff).
		Order("created_at ASC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

func (r *AuditRepository) ListAdminMessages(page, pageSize int) ([]model.SysMessage, int64, error) {
	var msgs []model.SysMessage
	var total int64

	query := r.db.Model(&model.SysMessage{}).
		Where("type IN ?", []string{model.MsgTypeNewApplication, model.MsgTypeApplicationWithdrawn})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}

	return msgs, total, nil
}

func (r *AuditRepository) ListMessages(userID uint, page, pageSize int, unread *bool) ([]model.SysMessage, int64, error) {
	var msgs []model.SysMessage
	var total int64

	query := r.db.Model(&model.SysMessage{}).Where("receiver_id = ?", userID)
	if unread != nil {
		query = query.Where("is_read = ?", *unread)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}

	return msgs, total, nil
}
