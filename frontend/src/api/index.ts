import axios from 'axios'
import { ElMessage } from 'element-plus'

// 动态获取 API 基础地址
// 开发环境：使用 Vite 代理，相对路径 /api
// 生产环境：使用相对路径 /api（后端服务前端静态文件）
const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截器，添加 token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.error || error.message || '请求失败'
    
    // 401 未登录或 token 过期，跳转到登录页
    if (status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      // 如果不在登录页，则跳转
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    } else {
      ElMessage.error(message)
    }
    
    return Promise.reject(error)
  }
)

export default api
