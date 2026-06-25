import request from './request'
import type { ApiResponse } from './request'

export interface Role {
  id: number
  name: string
  label: string
  description: string
  status: number
  permissions?: Permission[]
  created_at: string
}

export interface Permission {
  id: number
  name: string
  label: string
  description: string
  path: string
  method: string
  created_at: string
}

export const getRoleList = () =>
  request.get<any, ApiResponse<Role[]>>('/roles')

export const createRole = (data: { name: string; label: string; description?: string; status?: number }) =>
  request.post<any, ApiResponse<Role>>('/roles', data)

export const updateRole = (id: number, data: { name: string; label: string; description?: string; status?: number }) =>
  request.put<any, ApiResponse>(`/roles/${id}`, data)

export const deleteRole = (id: number) =>
  request.delete<any, ApiResponse>(`/roles/${id}`)

export const assignPermission = (roleId: number, permissionId: number) =>
  request.post<any, ApiResponse>('/roles/assign-permission', { role_id: roleId, permission_id: permissionId })

export const getPermissionList = () =>
  request.get<any, ApiResponse<Permission[]>>('/permissions')

export const createPermission = (data: { name: string; label: string; description?: string; path: string; method: string }) =>
  request.post<any, ApiResponse<Permission>>('/permissions', data)

export const updatePermission = (id: number, data: { name: string; label: string; description?: string; path: string; method: string }) =>
  request.put<any, ApiResponse>(`/permissions/${id}`, data)

export const deletePermission = (id: number) =>
  request.delete<any, ApiResponse>(`/permissions/${id}`)
