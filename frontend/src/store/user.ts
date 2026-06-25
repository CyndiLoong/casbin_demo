import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi, getUserInfo, type LoginParams, type UserInfo } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const login = async (params: LoginParams) => {
    const res = await loginApi(params)
    setToken(res.data.token)
    userInfo.value = res.data.user
    return res.data
  }

  const fetchUserInfo = async () => {
    if (!token.value) return null
    try {
      const res = await getUserInfo()
      userInfo.value = res.data
      return res.data
    } catch {
      logout()
      return null
    }
  }

  const logout = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  const isLoggedIn = () => !!token.value

  const hasRole = (role: string) => {
    return userInfo.value?.roles.includes(role) ?? false
  }

  return {
    token,
    userInfo,
    setToken,
    login,
    fetchUserInfo,
    logout,
    isLoggedIn,
    hasRole
  }
})
