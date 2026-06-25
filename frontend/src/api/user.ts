import request from './request'
import type { ApiResponse } from './request'

export interface User {
  id: number
  uuid: string
  username: string
  nickname: string
  email: string
  avatar: string
  status: number
  roles: string[]
  created_at: string
}

export interface UserListResult {
  list: User[]
  total: number
  page: number
  page_size: number
}

export const getUserList = (page = 1, pageSize = 10) =>
  request.get<any, ApiResponse<UserListResult>>('/users', { params: { page, page_size: pageSize } })

export const createUser = (data: { username: string; password: string; nickname?: string; email?: string }) =>
  request.post<any, ApiResponse>('/users', data)

export const updateUser = (id: number, data: { nickname?: string; email?: string; status?: number }) =>
  request.put<any, ApiResponse>(`/users/${id}`, data)

export const deleteUser = (id: number) =>
  request.delete<any, ApiResponse>(`/users/${id}`)

export const assignRole = (userId: number, roleId: number) =>
  request.post<any, ApiResponse>('/users/assign-role', { user_id: userId, role_id: roleId })
