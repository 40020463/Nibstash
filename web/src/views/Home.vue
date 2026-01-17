<template>
  <div class="home-page">
    <!-- 批量操作栏 -->
    <div v-if="bookmarkStore.selectedIds.length > 0" class="batch-bar">
      <span>已选择 {{ bookmarkStore.selectedIds.length }} 项</span>
      <el-button size="small" @click="bookmarkStore.selectAll">全选</el-button>
      <el-button size="small" @click="bookmarkStore.clearSelection">取消</el-button>
      <el-button size="small" type="primary" @click="showMoveDialog = true">移动</el-button>
      <el-button size="small" type="danger" @click="handleBatchDelete">删除</el-button>
    </div>

    <!-- 书签列表 -->
    <div v-loading="bookmarkStore.loading" class="bookmark-list">
      <div
        v-for="bookmark in bookmarkStore.bookmarks"
        :key="bookmark.id"
        class="bookmark-card"
        :class="{ selected: bookmarkStore.selectedIds.includes(bookmark.id) }"
      >
        <div class="card-checkbox">
          <el-checkbox
            :model-value="bookmarkStore.selectedIds.includes(bookmark.id)"
            @change="bookmarkStore.toggleSelect(bookmark.id)"
          />
        </div>
        <div class="card-content" @click="openBookmark(bookmark.url)">
          <div class="card-header">
            <img
              v-if="bookmark.favicon"
              :src="bookmark.favicon"
              class="favicon"
              @error="e => e.target.style.display = 'none'"
            />
            <el-icon v-else class="favicon-placeholder"><Link /></el-icon>
            <span class="title">{{ bookmark.title }}</span>
          </div>
          <div class="url">{{ bookmark.url }}</div>
          <div v-if="bookmark.description" class="description">{{ bookmark.description }}</div>
          <div class="meta">
            <span v-if="bookmark.folder_path" class="folder">
              <el-icon><Folder /></el-icon> {{ bookmark.folder_path }}
            </span>
          </div>
        </div>
        <div class="card-actions">
          <el-button link @click.stop="editBookmark(bookmark)">
            <el-icon><Edit /></el-icon>
          </el-button>
          <el-button link type="danger" @click.stop="deleteBookmark(bookmark.id)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>

      <el-empty v-if="!bookmarkStore.loading && bookmarkStore.bookmarks.length === 0" description="暂无书签" />
    </div>

    <!-- 分页 -->
    <div class="pagination-bar">
      <div class="page-size-selector">
        <span>每页显示</span>
        <el-select v-model="pageSize" size="small" style="width: 80px" @change="handlePageSizeChange">
          <el-option :value="10" label="10" />
          <el-option :value="20" label="20" />
          <el-option :value="50" label="50" />
          <el-option :value="100" label="100" />
        </el-select>
        <span>条</span>
      </div>
      <el-pagination
        v-if="bookmarkStore.total > 0"
        v-model:current-page="bookmarkStore.page"
        :page-size="bookmarkStore.pageSize"
        :total="bookmarkStore.total"
        layout="total, prev, pager, next, jumper"
        @current-change="bookmarkStore.setPage"
      />
    </div>

    <!-- 编辑弹窗 -->
    <BookmarkFormDialog
      v-model="showEditDialog"
      :bookmark="editingBookmark"
      @success="handleEditSuccess"
    />

    <!-- 移动弹窗 -->
    <el-dialog v-model="showMoveDialog" title="移动到文件夹" width="400px">
      <el-select v-model="moveTarget" placeholder="选择目标文件夹" style="width: 100%">
        <el-option label="根目录" value="" />
        <el-option
          v-for="path in folderPaths"
          :key="path"
          :label="path"
          :value="path"
        />
      </el-select>
      <template #footer>
        <el-button @click="showMoveDialog = false">取消</el-button>
        <el-button type="primary" @click="handleBatchMove">移动</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useBookmarkStore } from '@/stores/bookmark'
import { useFolderStore } from '@/stores/folder'
import { useDomainStore } from '@/stores/domain'
import { ElMessage, ElMessageBox } from 'element-plus'
import BookmarkFormDialog from '@/components/BookmarkFormDialog.vue'

const bookmarkStore = useBookmarkStore()
const folderStore = useFolderStore()
const domainStore = useDomainStore()

const showEditDialog = ref(false)
const editingBookmark = ref(null)
const showMoveDialog = ref(false)
const moveTarget = ref('')
const pageSize = ref(bookmarkStore.pageSize)

const folderPaths = computed(() => {
  const paths = []
  function collect(nodes, prefix = '') {
    for (const node of nodes) {
      paths.push(node.path)
      if (node.children?.length) {
        collect(node.children, node.path + '/')
      }
    }
  }
  collect(folderStore.folders)
  return paths
})

onMounted(() => {
  bookmarkStore.fetchBookmarks()
})

function openBookmark(url) {
  window.open(url, '_blank')
}

function editBookmark(bookmark) {
  editingBookmark.value = bookmark
  showEditDialog.value = true
}

async function deleteBookmark(id) {
  try {
    await ElMessageBox.confirm('确定要删除这个书签吗？', '提示', {
      type: 'warning'
    })
    await bookmarkStore.deleteBookmark(id)
    domainStore.fetchDomains()
    ElMessage.success('删除成功')
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.error || '删除失败')
    }
  }
}

async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${bookmarkStore.selectedIds.length} 个书签吗？`, '提示', {
      type: 'warning'
    })
    await bookmarkStore.batchDelete()
    domainStore.fetchDomains()
    ElMessage.success('删除成功')
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.error || '删除失败')
    }
  }
}

async function handleBatchMove() {
  try {
    await bookmarkStore.batchMove(moveTarget.value)
    showMoveDialog.value = false
    moveTarget.value = ''
    folderStore.fetchFolders()
    ElMessage.success('移动成功')
  } catch (err) {
    ElMessage.error(err.error || '移动失败')
  }
}

function handleEditSuccess() {
  domainStore.fetchDomains()
}

function handlePageSizeChange(size) {
  bookmarkStore.setPageSize(size)
}
</script>

<style lang="scss" scoped>
.home-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #ecf5ff;
  border-radius: 8px;
  margin-bottom: 16px;
}

.bookmark-list {
  flex: 1;
  overflow: auto;
}

.bookmark-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  transition: box-shadow 0.3s;

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  &.selected {
    background: #ecf5ff;
  }

  .card-checkbox {
    padding-top: 2px;
  }

  .card-content {
    flex: 1;
    cursor: pointer;
    overflow: hidden;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
  }

  .favicon {
    width: 18px;
    height: 18px;
    border-radius: 4px;
  }

  .favicon-placeholder {
    width: 18px;
    height: 18px;
    color: #909399;
  }

  .title {
    font-size: 15px;
    font-weight: 500;
    color: #303133;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .url {
    font-size: 13px;
    color: #909399;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 6px;
  }

  .description {
    font-size: 13px;
    color: #606266;
    margin-bottom: 6px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .folder {
    font-size: 12px;
    color: #909399;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .card-actions {
    display: flex;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.2s;
  }

  &:hover .card-actions {
    opacity: 1;
  }
}

.pagination-bar {
  padding: 16px 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  border-radius: 8px;
  padding: 12px 16px;
  margin-top: 16px;

  .page-size-selector {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: #606266;
  }
}
</style>
