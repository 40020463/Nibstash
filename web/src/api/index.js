import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器
api.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      router.push('/login')
    }
    return Promise.reject(error.response?.data || error)
  }
)

export default api

// Auth API
export const authApi = {
  login: (password) => api.post('/auth/login', { password }),
  getMe: () => api.get('/auth/me'),
  changePassword: (oldPassword, newPassword) =>
    api.put('/auth/password', { old_password: oldPassword, new_password: newPassword })
}

// Bookmark API
export const bookmarkApi = {
  list: (params) => api.get('/bookmarks', { params }),
  get: (id) => api.get(`/bookmarks/${id}`),
  create: (data) => api.post('/bookmarks', data),
  update: (id, data) => api.put(`/bookmarks/${id}`, data),
  delete: (id) => api.delete(`/bookmarks/${id}`),
  batch: (action, ids, target) => api.post('/bookmarks/batch', { action, ids, target }),
  export: () => api.get('/bookmarks/export', { responseType: 'blob' }),
  import: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/bookmarks/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  clearAll: () => api.delete('/bookmarks/clear'),
  clearFolder: (folderPath) => api.post('/bookmarks/clear-folder', { folder_path: folderPath })
}

// Folder API
export const folderApi = {
  list: () => api.get('/folders'),
  create: (path) => api.post('/folders', { path }),
  move: (sourcePath, targetPath) => api.put('/folders/move', { source_path: sourcePath, target_path: targetPath }),
  merge: (sourcePath, targetPath) => api.put('/folders/merge', { source_path: sourcePath, target_path: targetPath }),
  delete: (path) => api.delete('/folders', { params: { path } })
}

// Tag API
export const tagApi = {
  list: () => api.get('/tags'),
  create: (name, color) => api.post('/tags', { name, color }),
  update: (id, data) => api.put(`/tags/${id}`, data),
  delete: (id) => api.delete(`/tags/${id}`)
}

// Domain API
export const domainApi = {
  list: () => api.get('/domains'),
  getBookmarks: (domain) => api.get(`/domains/${encodeURIComponent(domain)}/bookmarks`),
  delete: (domain) => api.delete(`/domains/${encodeURIComponent(domain)}`)
}

// Credential API
export const credentialApi = {
  list: () => api.get('/credentials'),
  get: (id) => api.get(`/credentials/${id}`),
  getByDomain: (domain) => api.get(`/credentials/domain/${encodeURIComponent(domain)}`),
  create: (data) => api.post('/credentials', data),
  update: (id, data) => api.put(`/credentials/${id}`, data),
  delete: (id) => api.delete(`/credentials/${id}`)
}

// Favicon API
export const faviconApi = {
  getPending: () => api.get('/favicons/pending'),
  update: (id, favicon) => api.put(`/favicons/${id}`, { favicon })
}
