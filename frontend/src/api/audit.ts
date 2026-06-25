import request from './request'
import type { ApiResponse } from './request'

export interface AuditApplication {
  id: number
  uuid: string
  applicant_id: number
  applicant_name: string
  resource_name: string
  resource_type: string
  api_name: string
  api_description: string
  purpose: string
  expected_qps: number
  contact_info: string
  status: number
  status_text: string
  reviewer_id?: number
  reviewer_name?: string
  review_comment?: string
  reviewed_at?: string
  created_at: string
  can_withdraw: boolean
  withdraw_reason?: string
  withdrawn_at?: string
}

export interface SysMessage {
  id: number
  uuid: string
  type: string
  title: string
  content: string
  business_type: string
  business_id: number
  is_read: boolean
  created_at: string
}

export interface CreateAuditRequest {
  resource_name: string
  resource_type: string
  api_name: string
  api_description?: string
  purpose: string
  expected_qps: number
  contact_info?: string
}

export interface ReviewAuditRequest {
  approved: boolean
  comment?: string
}

export interface PaginationParams {
  page?: number
  page_size?: number
}

export interface ListResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface UnreadCountResponse {
  unread_count: number
  pending_count: number
}

export function submitAudit(data: CreateAuditRequest): Promise<ApiResponse<AuditApplication>> {
  return request.post('/audit/applications', data)
}

export function getMyApplications(params?: PaginationParams & { status?: number }): Promise<ApiResponse<ListResponse<AuditApplication>>> {
  return request.get('/audit/my-applications', { params })
}

export function getAllApplications(params?: PaginationParams & { status?: number; applicant?: string }): Promise<ApiResponse<ListResponse<AuditApplication>>> {
  return request.get('/audit/applications', { params })
}

export function getAuditDetail(id: number): Promise<ApiResponse<AuditApplication>> {
  return request.get(`/audit/applications/${id}`)
}

export function reviewAudit(id: number, data: ReviewAuditRequest): Promise<ApiResponse<null>> {
  return request.post(`/audit/applications/${id}/review`, data)
}

export function getPendingCount(): Promise<ApiResponse<{ count: number }>> {
  return request.get('/audit/pending-count')
}

export function getUnreadCount(): Promise<ApiResponse<UnreadCountResponse>> {
  return request.get('/messages/unread-count')
}

export function getMessages(params?: PaginationParams & { unread?: boolean }): Promise<ApiResponse<ListResponse<SysMessage>>> {
  return request.get('/messages', { params })
}

export function markMessageRead(id: number): Promise<ApiResponse<null>> {
  return request.put(`/messages/${id}/read`)
}

export function markAllRead(): Promise<ApiResponse<null>> {
  return request.put('/messages/read-all')
}
