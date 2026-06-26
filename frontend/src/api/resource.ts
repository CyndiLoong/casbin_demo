import request from './request'
import type { ApiResponse } from './request'

export interface Resource {
  id: number
  uuid: string
  name: string
  type: string
  api_name: string
  description: string
  provider: string
  version: string
  default_qps: number
  max_qps: number
  status: number
  docs_url: string
  tags: string
  created_at: string
  updated_at: string
}

export interface ResourceListParams {
  page?: number
  page_size?: number
  type?: string
  status?: number
  keyword?: string
}

export interface ResourceListResponse {
  list: Resource[]
  total: number
  page: number
  page_size: number
}

export const getResources = (params?: ResourceListParams) =>
  request.get<any, ApiResponse<ResourceListResponse>>('/resources', { params })

export const getActiveResources = () =>
  request.get<any, ApiResponse<{ list: Resource[]; total: number }>>('/resources/active')

export const getResource = (id: number) =>
  request.get<any, ApiResponse<Resource>>(`/resources/${id}`)

export const createResource = (data: Partial<Resource>) =>
  request.post<any, ApiResponse<Resource>>('/resources', data)

export const updateResource = (id: number, data: Partial<Resource>) =>
  request.put<any, ApiResponse<Resource>>(`/resources/${id}`, data)

export const deleteResource = (id: number) =>
  request.delete<any, ApiResponse<void>>(`/resources/${id}`)
