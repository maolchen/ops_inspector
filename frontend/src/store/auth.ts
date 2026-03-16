import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'

export interface UserInfo {
  id: number
  username: string
  display_name: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  // 从 localStorage 恢复用户信息
  const savedUser = localStorage.getItem('user')
  if (savedUser) {
    try {
      user.value = JSON.parse(savedUser)
    } catch {
      user.value = null
    }
  }

  async function login(username: string, password: string) {
    const response: any = await api.post('/auth/login', { username, password })
    const data = response.data
    token.value = data.token
    user.value = data.user
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
    return data
  }

  async function logout() {
    try {
      await api.post('/auth/logout')
    } catch {
      // ignore
    }
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function fetchUser() {
    if (!token.value) return null
    try {
      const response: any = await api.get('/auth/user')
      user.value = response.data
      localStorage.setItem('user', JSON.stringify(response.data))
      return response.data
    } catch {
      logout()
      return null
    }
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    await api.put('/auth/password', {
      old_password: oldPassword,
      new_password: newPassword
    })
  }

  return {
    token,
    user,
    isLoggedIn,
    login,
    logout,
    fetchUser,
    changePassword
  }
})
