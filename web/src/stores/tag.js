import { defineStore } from 'pinia'
import { ref } from 'vue'
import { tagApi } from '@/api'

export const useTagStore = defineStore('tag', () => {
  const tags = ref([])
  const loading = ref(false)

  async function fetchTags() {
    loading.value = true
    try {
      tags.value = await tagApi.list() || []
    } finally {
      loading.value = false
    }
  }

  async function createTag(name, color) {
    const tag = await tagApi.create(name, color)
    await fetchTags()
    return tag
  }

  async function updateTag(id, data) {
    await tagApi.update(id, data)
    await fetchTags()
  }

  async function deleteTag(id) {
    await tagApi.delete(id)
    await fetchTags()
  }

  return {
    tags,
    loading,
    fetchTags,
    createTag,
    updateTag,
    deleteTag
  }
})
