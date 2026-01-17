import { defineStore } from 'pinia'
import { ref } from 'vue'
import { folderApi } from '@/api'

export const useFolderStore = defineStore('folder', () => {
  const folders = ref([])
  const hasUncategorized = ref(false)
  const loading = ref(false)
  const currentPath = ref('')

  async function fetchFolders() {
    loading.value = true
    try {
      const res = await folderApi.list()
      folders.value = res.folders || []
      hasUncategorized.value = res.has_uncategorized
    } finally {
      loading.value = false
    }
  }

  async function createFolder(path) {
    await folderApi.create(path)
    await fetchFolders()
  }

  async function moveFolder(sourcePath, targetPath) {
    await folderApi.move(sourcePath, targetPath)
    await fetchFolders()
  }

  async function mergeFolder(sourcePath, targetPath) {
    await folderApi.merge(sourcePath, targetPath)
    await fetchFolders()
  }

  async function deleteFolder(path) {
    await folderApi.delete(path)
    await fetchFolders()
  }

  function setCurrentPath(path) {
    currentPath.value = path
  }

  return {
    folders,
    hasUncategorized,
    loading,
    currentPath,
    fetchFolders,
    createFolder,
    moveFolder,
    mergeFolder,
    deleteFolder,
    setCurrentPath
  }
})
