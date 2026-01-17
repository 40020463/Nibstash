import { defineStore } from 'pinia'
import { ref } from 'vue'
import { bookmarkApi } from '@/api'

export const useBookmarkStore = defineStore('bookmark', () => {
  const bookmarks = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)
  const selectedIds = ref([])

  // 筛选条件
  const filters = ref({
    search: '',
    tagId: null,
    folderPath: '',
    filterFolder: false,
    sortBy: 'time_desc'
  })

  async function fetchBookmarks() {
    loading.value = true
    try {
      const res = await bookmarkApi.list({
        page: page.value,
        page_size: pageSize.value,
        search: filters.value.search,
        tag_id: filters.value.tagId,
        folder_path: filters.value.folderPath,
        filter_folder: filters.value.filterFolder,
        sort_by: filters.value.sortBy
      })
      bookmarks.value = res.bookmarks || []
      total.value = res.total
    } finally {
      loading.value = false
    }
  }

  async function createBookmark(data) {
    const bookmark = await bookmarkApi.create(data)
    await fetchBookmarks()
    return bookmark
  }

  async function updateBookmark(id, data) {
    const bookmark = await bookmarkApi.update(id, data)
    await fetchBookmarks()
    return bookmark
  }

  async function deleteBookmark(id) {
    await bookmarkApi.delete(id)
    await fetchBookmarks()
  }

  async function batchDelete() {
    if (selectedIds.value.length === 0) return
    await bookmarkApi.batch('delete', selectedIds.value)
    selectedIds.value = []
    await fetchBookmarks()
  }

  async function batchMove(targetFolder) {
    if (selectedIds.value.length === 0) return
    await bookmarkApi.batch('move', selectedIds.value, targetFolder)
    selectedIds.value = []
    await fetchBookmarks()
  }

  function setPage(p) {
    page.value = p
    fetchBookmarks()
  }

  function setPageSize(size) {
    pageSize.value = size
    page.value = 1
    fetchBookmarks()
  }

  function setFilters(newFilters) {
    filters.value = { ...filters.value, ...newFilters }
    page.value = 1
    fetchBookmarks()
  }

  function toggleSelect(id) {
    const idx = selectedIds.value.indexOf(id)
    if (idx === -1) {
      selectedIds.value.push(id)
    } else {
      selectedIds.value.splice(idx, 1)
    }
  }

  function selectAll() {
    selectedIds.value = bookmarks.value.map(b => b.id)
  }

  function clearSelection() {
    selectedIds.value = []
  }

  return {
    bookmarks,
    total,
    page,
    pageSize,
    loading,
    selectedIds,
    filters,
    fetchBookmarks,
    createBookmark,
    updateBookmark,
    deleteBookmark,
    batchDelete,
    batchMove,
    setPage,
    setPageSize,
    setFilters,
    toggleSelect,
    selectAll,
    clearSelection
  }
})
