import request from './request'
import type { ApiResponse } from './request'

export interface LoginParams {
  username: string
  password: string
}

export interface RegisterParams {
  username: string
  password: string
  nickname?: string
  email?: string
}

export interface UserInfo {
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

export interface LoginResult {
  token: string
  user: UserInfo
}

export const login = (data: LoginParams) =>
  request.post<any, ApiResponse<LoginResult>>('/login', data)

export const register = (data: RegisterParams) =>
  request.post<any, ApiResponse>('/register', data)

export const getUserInfo = () =>
  request.get<any, ApiResponse<UserInfo>>('/userinfo')

export const getDashboard = () =>
  request.get<any, ApiResponse<any>>('/dashboard')
