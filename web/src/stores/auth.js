import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)

  const isAuthenticated = computed(() => !!token.value)

  async function login(password) {
    const res = await authApi.login(password)
    token.value = res.token
    user.value = res.user
    localStorage.setItem('token', res.token)
    return res
  }

  async function fetchUser() {
    if (!token.value) return
    try {
      user.value = await authApi.getMe()
    } catch {
      logout()
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    // 清除 cookie（用于 bookmarklet）
    document.cookie = 'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;'
  }

  async function changePassword(oldPassword, newPassword) {
    return await authApi.changePassword(oldPassword, newPassword)
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    fetchUser,
    logout,
    changePassword
  }
})
