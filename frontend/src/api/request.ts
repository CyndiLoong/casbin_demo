import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const service: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: {
    'X-Requested-With': 'XMLHttpRequest'
  }
})

service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

function handleError(status: number, message?: string) {
  const msg = message || '请求失败'
  switch (status) {
    case 401:
      localStorage.removeItem('token')
      ElMessage.error(msg || '登录已过期，请重新登录')
      router.push('/login')
      break
    case 403:
      ElMessage.error(msg || '没有访问权限')
      router.push('/403')
      break
    case 404:
      ElMessage.error(msg || '请求的资源不存在')
      router.push('/404')
      break
    case 400:
      ElMessage.error(msg || '请求参数错误')
      break
    default:
      if (status >= 500) {
        ElMessage.error(msg || '服务器内部错误')
        router.push('/500')
      } else {
        ElMessage.error(msg)
      }
  }
}

service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data
    if (res.code !== 200) {
      handleError(res.code, res.message)
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  (error) => {
    if (error.response) {
      const { status } = error.response
      const message = error.response.data?.message
      handleError(status, message)
    } else if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时，请稍后重试')
    } else {
      ElMessage.error('网络连接失败，请检查网络或后端服务')
    }
    return Promise.reject(error)
  }
)

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export default service
