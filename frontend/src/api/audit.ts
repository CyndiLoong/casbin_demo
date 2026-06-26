// 审核系统API接口层
// 提供大模型API资源申请、审核、撤回、站内消息等功能的接口封装
import request from './request'
import type { ApiResponse } from './request'

// ==================== 类型定义 ====================

/**
 * 审核申请实体
 * 包含资源申请的完整信息，包括申请人、资源信息、审核状态、撤回信息等
 */
export interface AuditApplication {
  id: number
  uuid: string
  applicant_id: number          // 申请人用户ID
  applicant_name: string        // 申请人姓名/昵称
  resource_name: string         // 资源名称（如"GPT-4 API"）
  resource_type: string         // 资源类型：llm/embedding/image/audio
  api_name: string              // API标识名（如"gpt-4-turbo"）
  api_description: string       // API详细描述
  purpose: string               // 使用用途说明
  expected_qps: number          // 预估QPS
  contact_info: string          // 联系方式
  status: number                // 审核状态：0-待审核 1-已通过 2-已驳回 3-已撤回
  status_text: string           // 状态中文描述
  reviewer_id?: number          // 审核人用户ID
  reviewer_name?: string        // 审核人姓名
  review_comment?: string       // 审核意见
  reviewed_at?: string          // 审核时间
  created_at: string            // 申请提交时间
  can_withdraw: boolean         // 是否可撤回（状态=待审核 且 在2分钟窗口内）
  withdraw_remain_ms: number    // 剩余可撤回时间（毫秒）
  withdraw_reason?: string      // 撤回原因
  withdrawn_at?: string         // 撤回时间
}

/**
 * 站内消息实体
 * 用于存储系统通知、审核结果等消息，支持离线消息持久化
 */
export interface SysMessage {
  id: number
  uuid: string
  type: string                  // 消息类型：new_application/review_result/withdraw...
  title: string                 // 消息标题
  content: string               // 消息内容
  business_type: string         // 关联业务类型：audit_application
  business_id: number           // 关联业务ID
  is_read: boolean              // 是否已读
  created_at: string            // 创建时间
}

/** 提交审核申请请求体 */
export interface CreateAuditRequest {
  resource_name: string
  resource_type: string
  api_name: string
  api_description?: string
  purpose: string
  expected_qps: number
  contact_info?: string
}

/** 审核申请请求体（管理员操作） */
export interface ReviewAuditRequest {
  approved: boolean             // 审核结果：true-通过 false-驳回
  comment?: string              // 审核意见
}

/** 分页查询参数 */
export interface PaginationParams {
  page?: number
  page_size?: number
}

/** 分页响应结构 */
export interface ListResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

/** 未读消息数响应 */
export interface UnreadCountResponse {
  unread_count: number          // 未读消息数
  pending_count: number         // 待审核数量（管理员视角）
}

/** 撤回申请请求体 */
export interface WithdrawAuditRequest {
  reason?: string               // 撤回原因
}

// ==================== 申请相关接口 ====================

/**
 * 提交资源审核申请
 * @param data 申请信息
 * @returns 创建成功的申请详情
 */
export function submitAudit(data: CreateAuditRequest): Promise<ApiResponse<AuditApplication>> {
  return request.post('/audit/applications', data)
}

/**
 * 获取我的申请列表
 * @param params 分页参数 + 状态筛选
 * @returns 分页申请列表
 */
export function getMyApplications(params?: PaginationParams & { status?: number }): Promise<ApiResponse<ListResponse<AuditApplication>>> {
  return request.get('/audit/my-applications', { params })
}

/**
 * 获取全部申请列表（管理员）
 * @param params 分页参数 + 状态筛选 + 关键词 + 是否排除待审核
 * @returns 分页申请列表
 */
export function getAllApplications(params?: PaginationParams & { status?: number; applicant?: string; exclude_pending?: boolean }): Promise<ApiResponse<ListResponse<AuditApplication>>> {
  return request.get('/audit/applications', { params })
}

/**
 * 获取申请详情
 * @param id 申请ID
 * @returns 申请详情
 */
export function getAuditDetail(id: number): Promise<ApiResponse<AuditApplication>> {
  return request.get(`/audit/applications/${id}`)
}

/**
 * 审核申请（管理员）
 * @param id 申请ID
 * @param data 审核结果（通过/驳回 + 意见）
 */
export function reviewAudit(id: number, data: ReviewAuditRequest): Promise<ApiResponse<null>> {
  return request.post(`/audit/applications/${id}/review`, data)
}

/**
 * 撤回申请
 * @param id 申请ID
 * @param data 撤回原因（可选）
 * @description 仅在申请状态为"待审核"且提交后2分钟内可撤回
 */
export function withdrawAudit(id: number, data?: WithdrawAuditRequest): Promise<ApiResponse<null>> {
  return request.post(`/audit/applications/${id}/withdraw`, data || {})
}

// ==================== 统计接口 ====================

/**
 * 获取待审核申请数量（管理员）
 * @returns 待审核数量
 */
export function getPendingCount(): Promise<ApiResponse<{ count: number }>> {
  return request.get('/audit/pending-count')
}

/**
 * 获取未读消息数
 * @returns 未读消息数 + 待审核数
 */
export function getUnreadCount(): Promise<ApiResponse<UnreadCountResponse>> {
  return request.get('/messages/unread-count')
}

// ==================== 消息相关接口 ====================

/**
 * 获取站内消息列表
 * @param params 分页参数 + 未读筛选
 * @returns 分页消息列表
 */
export function getMessages(params?: PaginationParams & { unread?: boolean }): Promise<ApiResponse<ListResponse<SysMessage>>> {
  return request.get('/messages', { params })
}

/**
 * 标记单条消息为已读
 * @param id 消息ID
 */
export function markMessageRead(id: number): Promise<ApiResponse<null>> {
  return request.put(`/messages/${id}/read`)
}

/**
 * 标记所有消息为已读
 */
export function markAllRead(): Promise<ApiResponse<null>> {
  return request.put('/messages/read-all')
}
