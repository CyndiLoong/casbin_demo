// Package model 定义应用程序的数据模型层。
//
// 本层职责：
//   - 定义数据库表对应的 GORM 实体结构体（User/Role/Permission/AuditApplication/SysMessage）
//   - 定义 API 请求/响应的数据传输对象（DTO）
//   - 通过 GORM tag 指定字段约束、索引、关联关系
//   - 通过 binding tag 定义 Gin 参数校验规则
//
// 数据模型关系：
//
//	User ──< UserRole >── Role ──< RolePermission >── Permission
//	(多对多)              (多对多)
//
//	User ──< AuditApplication (申请人)
//	User ──< SysMessage (接收人)
//	AuditApplication ──< SysMessage (业务关联)
//
// 使用 GORM 的软删除（gorm.DeletedAt）支持数据恢复。
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表实体，对应数据库 users 表。
//
// 字段说明：
//   - ID: 自增主键
//   - UUID: 全局唯一标识符，用于对外暴露（避免主键泄露）
//   - Username: 登录用户名，唯一索引
//   - Password: bcrypt 加密后的密码哈希（json:"-" 确保不会序列化到响应）
//   - Status: 状态 1=启用 0=禁用
//   - Roles: 用户角色关联（多对多，通过 user_roles 中间表）
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UUID      string         `gorm:"uniqueIndex;size:36;not null" json:"uuid"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Nickname  string         `gorm:"size:50" json:"nickname"`
	Email     string         `gorm:"size:100" json:"email"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	Status    int            `gorm:"default:1" json:"status"`
	Roles     []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定 GORM 使用的表名。
func (User) TableName() string {
	return "users"
}

// Role 角色表实体，对应数据库 roles 表。
//
// 字段说明：
//   - Name: 角色标识名（如 admin/user），唯一索引，用于 Casbin 策略匹配
//   - Label: 角色显示名称（如 管理员/普通用户）
//   - Permissions: 角色关联的权限（多对多，通过 role_permissions 中间表）
//   - Users: 反向关联用户（json:"-" 不序列化）
type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Label       string         `gorm:"size:100" json:"label"`
	Description string         `gorm:"size:255" json:"description"`
	Status      int            `gorm:"default:1" json:"status"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	Users       []User         `gorm:"many2many:user_roles;" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定 GORM 使用的表名。
func (Role) TableName() string {
	return "roles"
}

// Permission 权限表实体，对应数据库 permissions 表。
//
// 字段说明：
//   - Name: 权限标识名（如 user:list），唯一索引
//   - Path: API 路径（如 /api/users），用于 Casbin 策略匹配 (obj)
//   - Method: HTTP 方法（如 GET/POST），用于 Casbin 策略匹配 (act)
//   - Roles: 反向关联角色（json:"-" 不序列化）
type Permission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Label       string         `gorm:"size:100" json:"label"`
	Description string         `gorm:"size:255" json:"description"`
	Path        string         `gorm:"size:255" json:"path"`
	Method      string         `gorm:"size:20" json:"method"`
	Roles       []Role         `gorm:"many2many:role_permissions;" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定 GORM 使用的表名。
func (Permission) TableName() string {
	return "permissions"
}

// LoginRequest 登录请求 DTO。
// binding:"required" 表示字段必填，Gin 会自动校验。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求 DTO。
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// UserResponse 用户信息响应 DTO（脱敏）。
type UserResponse struct {
	ID        uint      `json:"id"`
	UUID      string    `json:"uuid"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Status    int       `json:"status"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse 登录响应 DTO。
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// CreateRoleRequest 创建角色请求 DTO。
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Label       string `json:"label" binding:"required"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// CreatePermissionRequest 创建权限请求 DTO。
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Label       string `json:"label" binding:"required"`
	Description string `json:"description"`
	Path        string `json:"path" binding:"required"`
	Method      string `json:"method" binding:"required"`
}

// AssignRoleRequest 用户分配角色请求 DTO。
type AssignRoleRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	RoleID uint `json:"role_id" binding:"required"`
}

// AssignPermissionRequest 角色分配权限请求 DTO。
type AssignPermissionRequest struct {
	RoleID       uint `json:"role_id" binding:"required"`
	PermissionID uint `json:"permission_id" binding:"required"`
}

// AuditStatus 审核状态常量
const (
	AuditStatusPending   = 0 // 待审核
	AuditStatusApproved  = 1 // 已通过
	AuditStatusRejected  = 2 // 已驳回
	AuditStatusWithdrawn = 3 // 已撤回
)

// MessageType 系统消息类型常量
const (
	MsgTypeNewApplication       = "new_application"       // 新申请通知（管理员）
	MsgTypeReviewResult         = "review_result"         // 审核结果通知（用户）
	MsgTypeApplicationWithdrawn = "application_withdrawn" // 申请撤回通知（管理员）
)

// MessageTargetType MQ消息目标类型
const (
	TargetTypeAdmins = "admins" // 广播给所有管理员
	TargetTypeUser   = "user"   // 发送给指定用户
)

// WithdrawWindow 撤回窗口时长（服务端Redis TTL控制，前端不可篡改）
const WithdrawWindow = 2 * time.Minute

// BusinessType 业务类型常量
const (
	BusinessTypeAudit = "audit_application"
)

// Cache key 前缀常量
const (
	CacheKeyWithdrawWindow = "cache:audit:withdraw:"
)

// AuditApplication 大模型API资源审核申请表实体，对应 api_audit_applications 表。
//
// 索引设计：
//   - idx_applicant_status: (applicant_id, status) 复合索引，用户查询自己的申请列表
//   - idx_status_created: (status, created_at) 复合索引，管理员按状态筛选+时间排序
//   - idx_reviewer: reviewer_id 索引，审核人查询
//
// 乐观锁：
//   - Version 字段用于乐观锁，更新时 WHERE version = ? 防止并发修改导致状态错乱
//
// 审核流程：
//  1. 普通用户填写并提交申请 → status=pending，写入 sys_message 给管理员，Redis设置2分钟撤回TTL
//  2. MQ fanout广播 → 所有实例消费 → WS推送给在线管理员（Redis PubSub跨实例）
//  3. 2分钟内用户可撤回（Redis TTL优先校验 + PG时间戳兜底 + 乐观锁防并发）
//  4. 管理员查看详情并审核（通过/驳回）→ PG事务更新状态 + 乐观锁校验，写入sys_message给申请人
//  5. MQ广播 → WS推送给在线申请人
type AuditApplication struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UUID           string         `gorm:"uniqueIndex;size:36;not null" json:"uuid"`
	ApplicantID    uint           `gorm:"index:idx_applicant_status,priority:1;not null" json:"applicant_id"`
	ApplicantName  string         `gorm:"size:50" json:"applicant_name"`
	ResourceName   string         `gorm:"size:200;not null" json:"resource_name"`
	ResourceType   string         `gorm:"size:50;not null" json:"resource_type"`
	APIName        string         `gorm:"size:200;not null" json:"api_name"`
	APIDescription string         `gorm:"type:text" json:"api_description"`
	Purpose        string         `gorm:"type:text;not null" json:"purpose"`
	ExpectedQPS    int            `gorm:"default:0" json:"expected_qps"`
	ContactInfo    string         `gorm:"size:200" json:"contact_info"`
	Status         int            `gorm:"default:0;index:idx_applicant_status,priority:2;index:idx_status_created,priority:1" json:"status"`
	ReviewerID     *uint          `gorm:"index" json:"reviewer_id,omitempty"`
	ReviewerName   string         `gorm:"size:50" json:"reviewer_name,omitempty"`
	ReviewComment  string         `gorm:"type:text" json:"review_comment,omitempty"`
	WithdrawReason string         `gorm:"type:text" json:"withdraw_reason,omitempty"`
	WithdrawnAt    *time.Time     `json:"withdrawn_at,omitempty"`
	ReviewedAt     *time.Time     `json:"reviewed_at,omitempty"`
	Version        int            `gorm:"not null;default:0" json:"-"`
	CreatedAt      time.Time      `gorm:"index:idx_status_created,priority:2" json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定 GORM 使用的表名。
func (AuditApplication) TableName() string {
	return "api_audit_applications"
}

// SysMessage 系统消息表实体，对应 sys_messages 表。
//
// 消息持久化策略（PG唯一可信持久层）：
//   - 所有通知消息先落 PG，再发 MQ，MQ 发送失败不影响数据可靠性（定时对账补发兜底）
//   - 支持已读/未读状态、分页查询、多条件筛选
//
// 索引设计（GIN索引用于复杂查询场景，此处使用B-tree覆盖高频查询）：
//   - idx_receiver_read_created: (receiver_id, is_read, created_at) 覆盖未读数+消息列表查询
//   - idx_business: (business_type, business_id) 按业务反查消息
//   - idx_mq_delivered: (mq_delivered, created_at) 定时补发扫描未投递消息
type SysMessage struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UUID         string         `gorm:"uniqueIndex;size:36;not null" json:"uuid"`
	ReceiverID   uint           `gorm:"index:idx_receiver_read_created,priority:1;not null" json:"receiver_id"`
	Type         string         `gorm:"size:50;index;not null" json:"type"`
	Title        string         `gorm:"size:200;not null" json:"title"`
	Content      string         `gorm:"type:text" json:"content"`
	BusinessType string         `gorm:"size:50;index:idx_business,priority:1" json:"business_type"`
	BusinessID   uint           `gorm:"index:idx_business,priority:2" json:"business_id"`
	IsRead       bool           `gorm:"default:false;index:idx_receiver_read_created,priority:2" json:"is_read"`
	MQDelivered  bool           `gorm:"default:false;index:idx_mq_delivered,priority:1" json:"-"`
	CreatedAt    time.Time      `gorm:"index:idx_receiver_read_created,priority:3;index:idx_mq_delivered,priority:2" json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定 GORM 使用的表名。
func (SysMessage) TableName() string {
	return "sys_messages"
}

// CreateAuditRequest 创建审核申请请求 DTO。
type CreateAuditRequest struct {
	ResourceName   string `json:"resource_name" binding:"required,max=200"`
	ResourceType   string `json:"resource_type" binding:"required,max=50"`
	APIName        string `json:"api_name" binding:"required,max=200"`
	APIDescription string `json:"api_description"`
	Purpose        string `json:"purpose" binding:"required"`
	ExpectedQPS    int    `json:"expected_qps" binding:"min=0"`
	ContactInfo    string `json:"contact_info" binding:"max=200"`
}

// ReviewAuditRequest 审核申请请求 DTO。
type ReviewAuditRequest struct {
	Approved bool   `json:"approved"`
	Comment  string `json:"comment"`
}

// WithdrawAuditRequest 撤回申请请求 DTO。
type WithdrawAuditRequest struct {
	Reason string `json:"reason" binding:"max=500"`
}

// AuditListRequest 审核列表查询请求 DTO。
type AuditListRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	Status    *int   `form:"status"`
	Applicant string `form:"applicant"`
}

// MessageListRequest 消息列表查询请求 DTO。
type MessageListRequest struct {
	Page     int  `form:"page" binding:"min=1"`
	PageSize int  `form:"page_size" binding:"min=1,max=100"`
	Unread   *bool `form:"unread"`
}

// AuditApplicationResponse 审核申请响应 DTO（脱敏）。
//
// CanWithdraw 字段：服务端计算，true 表示申请处于 pending 状态且在2分钟撤回窗口内。
type AuditApplicationResponse struct {
	ID             uint      `json:"id"`
	UUID           string    `json:"uuid"`
	ApplicantID    uint      `json:"applicant_id"`
	ApplicantName  string    `json:"applicant_name"`
	ResourceName   string    `json:"resource_name"`
	ResourceType   string    `json:"resource_type"`
	APIName        string    `json:"api_name"`
	APIDescription string    `json:"api_description"`
	Purpose        string    `json:"purpose"`
	ExpectedQPS    int       `json:"expected_qps"`
	ContactInfo    string    `json:"contact_info"`
	Status         int       `json:"status"`
	StatusText     string    `json:"status_text"`
	CanWithdraw    bool      `json:"can_withdraw"`
	WithdrawRemain int64     `json:"withdraw_remain_ms,omitempty"`
	ReviewerID     *uint     `json:"reviewer_id,omitempty"`
	ReviewerName   string    `json:"reviewer_name,omitempty"`
	ReviewComment  string    `json:"review_comment,omitempty"`
	WithdrawReason string    `json:"withdraw_reason,omitempty"`
	WithdrawnAt    *time.Time `json:"withdrawn_at,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// UnreadCountResponse 未读消息数（含管理员待审数）响应 DTO。
type UnreadCountResponse struct {
	UnreadCount  int64 `json:"unread_count"`
	PendingCount int64 `json:"pending_count"`
}

// SysMessageResponse 系统消息响应 DTO。
type SysMessageResponse struct {
	ID           uint      `json:"id"`
	UUID         string    `json:"uuid"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	BusinessType string    `json:"business_type"`
	BusinessID   uint      `json:"business_id"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
}

// WsNotification WebSocket 推送通知结构。
type WsNotification struct {
	Type         string      `json:"type"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	BusinessType string      `json:"business_type"`
	BusinessID   uint        `json:"business_id"`
	Timestamp    time.Time   `json:"timestamp"`
	Data         interface{} `json:"data,omitempty"`
	ID           string      `json:"id,omitempty"`
}

// GetID 实现 ws.IDProvider 接口，用于幂等去重。
func (n WsNotification) GetID() string { return n.ID }

// MQNotificationMessage MQ 传输的轻量消息体（只传 message_id，详情查PG）。
type MQNotificationMessage struct {
	MessageID      uint      `json:"message_id"`
	TargetType     string    `json:"target_type"`
	TargetID       uint      `json:"target_id,omitempty"`
	Type           string    `json:"type"`
	BusinessID     uint      `json:"business_id"`
	BusinessType   string    `json:"business_type"`
	CreatedAt      time.Time `json:"created_at"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
}
