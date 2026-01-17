<template>
  <div class="app-layout">
    <!-- 顶部导航栏 -->
    <header class="top-navbar">
      <div class="navbar-left" @click="goHome" style="cursor: pointer;">
        <span class="logo">🐿️ 囤囤鼠</span>
      </div>
      <div class="navbar-center">
        <el-input
          v-model="searchText"
          placeholder="搜索标题或链接..."
          clearable
          style="width: 400px"
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
      <div class="navbar-right">
        <el-button type="primary" @click="showAddBookmark = true">
          <el-icon><Plus /></el-icon> 添加
        </el-button>
        <el-button @click="$router.push('/import')">
          <el-icon><Upload /></el-icon> 导入
        </el-button>
        <el-button @click="handleExport">
          <el-icon><Download /></el-icon> 导出
        </el-button>
        <el-button @click="startLoadFavicons">
          <el-icon><Picture /></el-icon> 加载图标
        </el-button>
        <el-dropdown trigger="click">
          <el-button>
            <el-icon><MoreFilled /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="$router.push('/bookmarklet')">
                <el-icon><Link /></el-icon> Bookmarklet
              </el-dropdown-item>
              <el-dropdown-item divided @click="handleClearFolder">
                <el-icon><Delete /></el-icon> 清空当前文件夹
              </el-dropdown-item>
              <el-dropdown-item @click="handleClearAll">
                <el-icon><DeleteFilled /></el-icon> 清空全部书签
              </el-dropdown-item>
              <el-dropdown-item divided @click="handleLogout">
                <el-icon><SwitchButton /></el-icon> 退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <div class="main-container">
      <!-- 左侧边栏 -->
      <aside class="sidebar">
        <!-- 排序选项 -->
        <div class="sidebar-section">
          <div class="section-title">排序方式</div>
          <el-select v-model="sortBy" size="small" style="width: 100%" @change="handleSortChange">
            <el-option label="时间倒序" value="time_desc" />
            <el-option label="时间正序" value="time_asc" />
            <el-option label="标题 A-Z" value="title_asc" />
            <el-option label="标题 Z-A" value="title_desc" />
            <el-option label="URL A-Z" value="url_asc" />
            <el-option label="URL Z-A" value="url_desc" />
          </el-select>
        </div>

        <!-- 文件夹树 -->
        <div class="sidebar-section folder-section">
          <div class="section-title">
            <span>文件夹</span>
            <el-button link size="small" @click="showCreateFolder = true">
              <el-icon><Plus /></el-icon>
            </el-button>
          </div>
          <div class="folder-tree">
            <div
              class="folder-item"
              :class="{ active: currentFolder === '' && !filterFolder }"
              @click="selectFolder('', false)"
              @dragover="handleRootDragOver"
              @dragleave="handleRootDragLeave"
              @drop="handleRootDrop"
            >
              <el-icon class="folder-icon"><Folder /></el-icon>
              <span class="folder-name">全部书签</span>
            </div>
            <div
              v-if="folderStore.hasUncategorized"
              class="folder-item"
              :class="{ active: currentFolder === '' && filterFolder }"
              @click="selectFolder('', true)"
            >
              <el-icon class="folder-icon"><FolderOpened /></el-icon>
              <span class="folder-name">未分类</span>
            </div>
            <FolderTreeNode
              v-for="folder in folderStore.folders"
              :key="folder.path"
              :folder="folder"
              :current-folder="currentFolder"
              :filter-folder="filterFolder"
              @select="selectFolder"
              @folder-drop="handleFolderDrop"
            />
          </div>
        </div>
      </aside>

      <!-- 主内容区 -->
      <main class="main-content">
        <div class="content-area">
          <router-view />
        </div>
      </main>

      <!-- 右侧域名管理栏 -->
      <aside class="domain-sidebar">
        <div class="sidebar-header">
          <h3>域名管理</h3>
          <el-button link size="small" @click="domainStore.fetchDomains">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
        <div class="domain-search">
          <el-input
            v-model="domainSearchText"
            placeholder="搜索域名..."
            size="small"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        <div class="domain-list">
          <template v-for="(group, index) in filteredDomains" :key="group.top_domain">
            <div class="domain-group">
              <div
                class="domain-group-header"
                @click="toggleDomainGroup(index)"
              >
                <span class="domain-caret">
                  {{ group.sub_domains?.length > 0 ? (expandedDomains.includes(index) ? '▼' : '▶') : '—' }}
                </span>
                <span class="domain-name">{{ group.top_domain }}</span>
                <el-icon v-if="group.has_credentials" class="credential-icon" title="有凭证"><Key /></el-icon>
                <span class="domain-count">
                  ({{ group.bookmark_count }}{{ group.sub_domains?.length > 0 ? '/' + group.sub_domains.length : '' }})
                </span>
                <el-button
                  link
                  size="small"
                  class="domain-manage-btn"
                  @click.stop="openDomainModal(group.top_domain)"
                >
                  管理
                </el-button>
              </div>
              <div v-if="expandedDomains.includes(index) && group.sub_domains?.length > 0" class="domain-sublist">
                <div
                  v-for="sub in group.sub_domains"
                  :key="sub"
                  class="domain-subitem"
                >
                  <span class="domain-subitem-name">{{ sub }}</span>
                  <el-button
                    link
                    size="small"
                    class="domain-manage-btn"
                    @click.stop="openDomainModal(sub)"
                  >
                    管理
                  </el-button>
                </div>
              </div>
            </div>
          </template>
          <div v-if="filteredDomains.length === 0" class="domain-empty">
            暂无域名
          </div>
        </div>
      </aside>
    </div>

    <!-- 添加书签弹窗 -->
    <BookmarkFormDialog
      v-model="showAddBookmark"
      @success="handleBookmarkAdded"
    />

    <!-- 创建文件夹弹窗 -->
    <el-dialog v-model="showCreateFolder" title="创建文件夹" width="400px">
      <el-input v-model="newFolderPath" placeholder="文件夹路径，如：工作/项目" />
      <template #footer>
        <el-button @click="showCreateFolder = false">取消</el-button>
        <el-button type="primary" @click="handleCreateFolder">创建</el-button>
      </template>
    </el-dialog>

    <!-- 文件夹移动/合并弹窗 -->
    <el-dialog v-model="showFolderMoveDialog" title="移动文件夹" width="400px">
      <p>{{ folderMoveMessage }}</p>
      <p class="folder-move-hint">
        <strong>移动</strong>：保留文件夹结构<br>
        <strong>合并</strong>：将内容直接放入目标文件夹
      </p>
      <template #footer>
        <el-button @click="showFolderMoveDialog = false">取消</el-button>
        <el-button @click="handleMergeFolder">合并</el-button>
        <el-button type="primary" @click="handleMoveFolder">移动</el-button>
      </template>
    </el-dialog>

    <!-- Favicon 加载弹窗 -->
    <el-dialog
      v-model="showFaviconDialog"
      title="加载网站图标"
      width="450px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <div class="favicon-progress">
        <div class="progress-text">
          <span>{{ faviconCurrent }}</span> / <span>{{ faviconTotal }}</span>
        </div>
        <el-progress
          :percentage="faviconPercent"
          :stroke-width="12"
          :show-text="false"
        />
        <p class="status-text">{{ faviconStatus }}</p>
      </div>
      <template #footer>
        <el-button @click="minimizeFavicon">后台运行</el-button>
        <el-button type="danger" @click="cancelFavicon">取消</el-button>
      </template>
    </el-dialog>

    <!-- 后台加载浮动条 -->
    <div v-if="faviconMinimized" class="favicon-float">
      <div class="favicon-float-header">
        <span>加载图标</span>
        <el-button link size="small" @click="expandFavicon">
          <el-icon><FullScreen /></el-icon>
        </el-button>
      </div>
      <div class="favicon-float-progress">
        <span>{{ faviconCurrent }}/{{ faviconTotal }}</span>
        <el-progress
          :percentage="faviconPercent"
          :stroke-width="6"
          :show-text="false"
        />
      </div>
    </div>

    <!-- 域名管理弹窗 -->
    <DomainModal
      v-model="showDomainModal"
      :domain="currentDomainForModal"
      @refresh="domainStore.fetchDomains"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useBookmarkStore } from '@/stores/bookmark'
import { useFolderStore } from '@/stores/folder'
import { useDomainStore } from '@/stores/domain'
import { bookmarkApi, faviconApi } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import FolderTreeNode from '@/components/FolderTreeNode.vue'
import BookmarkFormDialog from '@/components/BookmarkFormDialog.vue'
import DomainModal from '@/components/DomainModal.vue'

const router = useRouter()
const authStore = useAuthStore()
const bookmarkStore = useBookmarkStore()
const folderStore = useFolderStore()
const domainStore = useDomainStore()

const searchText = ref('')
const sortBy = ref('time_desc')
const currentFolder = ref('')
const filterFolder = ref(false)
const showAddBookmark = ref(false)
const showCreateFolder = ref(false)
const newFolderPath = ref('')

// 文件夹拖放相关
const showFolderMoveDialog = ref(false)
const folderMoveMessage = ref('')
const dragSourceFolder = ref('')
const dragTargetFolder = ref('')

// Favicon 加载相关
const showFaviconDialog = ref(false)
const faviconMinimized = ref(false)
const faviconCancelled = ref(false)
const faviconCurrent = ref(0)
const faviconTotal = ref(0)
const faviconStatus = ref('')
const faviconPercent = computed(() => {
  if (faviconTotal.value === 0) return 0
  return Math.round((faviconCurrent.value / faviconTotal.value) * 100)
})

// 域名管理弹窗
const showDomainModal = ref(false)
const currentDomainForModal = ref('')
const domainSearchText = ref('')
const expandedDomains = ref([])

const filteredDomains = computed(() => {
  const allDomains = domainStore.domains || []
  if (!domainSearchText.value) return allDomains
  const kw = domainSearchText.value.toLowerCase().trim()
  if (!kw) return allDomains
  const result = []
  for (const g of allDomains) {
    if (g.top_domain && g.top_domain.toLowerCase().includes(kw)) {
      result.push(g)
    } else if (g.sub_domains && g.sub_domains.length > 0) {
      const matched = g.sub_domains.filter(s => s && s.toLowerCase().includes(kw))
      if (matched.length > 0) {
        result.push({ ...g, sub_domains: matched })
      }
    }
  }
  return result
})

onMounted(() => {
  folderStore.fetchFolders()
  domainStore.fetchDomains()
})

function handleSearch() {
  bookmarkStore.setFilters({ search: searchText.value })
}

function handleSortChange() {
  bookmarkStore.setFilters({ sortBy: sortBy.value })
}

function selectFolder(path, filter) {
  currentFolder.value = path
  filterFolder.value = filter
  bookmarkStore.setFilters({
    folderPath: path,
    filterFolder: filter
  })
}

function selectDomain(domain) {
  currentDomainForModal.value = domain
  showDomainModal.value = true
}

function openDomainModal(domain) {
  currentDomainForModal.value = domain
  showDomainModal.value = true
}

function toggleDomainGroup(index) {
  const idx = expandedDomains.value.indexOf(index)
  if (idx === -1) {
    expandedDomains.value.push(index)
  } else {
    expandedDomains.value.splice(idx, 1)
  }
}

async function handleCreateFolder() {
  if (!newFolderPath.value) {
    ElMessage.warning('请输入文件夹路径')
    return
  }
  try {
    await folderStore.createFolder(newFolderPath.value)
    ElMessage.success('创建成功')
    showCreateFolder.value = false
    newFolderPath.value = ''
  } catch (err) {
    ElMessage.error(err.error || '创建失败')
  }
}

function handleBookmarkAdded() {
  bookmarkStore.fetchBookmarks()
  domainStore.fetchDomains()
  folderStore.fetchFolders()
}

async function handleExport() {
  try {
    const blob = await bookmarkApi.export()
    const url = window.URL.createObjectURL(new Blob([blob]))
    const link = document.createElement('a')
    link.href = url
    link.download = 'bookmarks.html'
    link.click()
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (err) {
    ElMessage.error('导出失败')
  }
}

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

function goHome() {
  router.push('/')
  currentFolder.value = ''
  filterFolder.value = false
  searchText.value = ''
  bookmarkStore.setFilters({
    search: '',
    folderPath: '',
    filterFolder: false
  })
}

async function handleClearAll() {
  try {
    await ElMessageBox.confirm('确定要清空所有书签吗？此操作不可恢复！', '警告', {
      type: 'warning',
      confirmButtonText: '确定清空',
      cancelButtonText: '取消'
    })
    await bookmarkApi.clearAll()
    ElMessage.success('清空成功')
    bookmarkStore.fetchBookmarks()
    folderStore.fetchFolders()
    domainStore.fetchDomains()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.error || '清空失败')
    }
  }
}

async function handleClearFolder() {
  const folderName = filterFolder.value
    ? (currentFolder.value || '未分类')
    : '全部书签'

  try {
    await ElMessageBox.confirm(`确定要清空"${folderName}"中的书签吗？此操作不可恢复！`, '警告', {
      type: 'warning',
      confirmButtonText: '确定清空',
      cancelButtonText: '取消'
    })

    if (filterFolder.value) {
      await bookmarkApi.clearFolder(currentFolder.value)
    } else {
      await bookmarkApi.clearAll()
    }

    ElMessage.success('清空成功')
    bookmarkStore.fetchBookmarks()
    folderStore.fetchFolders()
    domainStore.fetchDomains()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.error || '清空失败')
    }
  }
}

// 文件夹拖放处理
function handleRootDragOver(e) {
  e.preventDefault()
  e.currentTarget.classList.add('drag-over')
}

function handleRootDragLeave(e) {
  e.currentTarget.classList.remove('drag-over')
}

function handleRootDrop(e) {
  e.preventDefault()
  e.currentTarget.classList.remove('drag-over')
  const sourcePath = e.dataTransfer.getData('text/plain')
  if (sourcePath) {
    dragSourceFolder.value = sourcePath
    dragTargetFolder.value = ''
    const sourceName = sourcePath.split('/').pop()
    folderMoveMessage.value = `确定要将文件夹 "${sourceName}" 移动到根目录吗？`
    showFolderMoveDialog.value = true
  }
}

function handleFolderDrop({ sourcePath, targetPath }) {
  dragSourceFolder.value = sourcePath
  dragTargetFolder.value = targetPath
  const sourceName = sourcePath.split('/').pop()
  const targetName = targetPath.split('/').pop()
  folderMoveMessage.value = `确定要将文件夹 "${sourceName}" 移动到 "${targetName}" 下吗？`
  showFolderMoveDialog.value = true
}

async function handleMoveFolder() {
  try {
    await folderStore.moveFolder(dragSourceFolder.value, dragTargetFolder.value)
    ElMessage.success('移动成功')
    showFolderMoveDialog.value = false
    bookmarkStore.fetchBookmarks()
  } catch (err) {
    ElMessage.error(err.error || '移动失败')
  }
}

async function handleMergeFolder() {
  try {
    await folderStore.mergeFolder(dragSourceFolder.value, dragTargetFolder.value)
    ElMessage.success('合并成功')
    showFolderMoveDialog.value = false
    bookmarkStore.fetchBookmarks()
  } catch (err) {
    ElMessage.error(err.error || '合并失败')
  }
}

// Favicon 加载功能
async function startLoadFavicons() {
  faviconCancelled.value = false
  faviconMinimized.value = false
  faviconCurrent.value = 0
  faviconTotal.value = 0
  faviconStatus.value = '获取待处理书签...'
  showFaviconDialog.value = true

  try {
    const data = await faviconApi.getPending()
    const bookmarks = data.bookmarks || []
    faviconTotal.value = bookmarks.length

    if (bookmarks.length === 0) {
      faviconStatus.value = '所有书签已有图标'
      setTimeout(() => {
        showFaviconDialog.value = false
      }, 1500)
      return
    }

    faviconStatus.value = '准备开始...'
    await processFavicons(bookmarks, 0)
  } catch (err) {
    faviconStatus.value = '获取失败: ' + (err.error || err.message || '未知错误')
  }
}

async function processFavicons(bookmarks, index) {
  if (faviconCancelled.value || index >= bookmarks.length) {
    if (!faviconCancelled.value) {
      faviconStatus.value = '完成！'
      faviconCurrent.value = bookmarks.length
      setTimeout(() => {
        showFaviconDialog.value = false
        faviconMinimized.value = false
        bookmarkStore.fetchBookmarks()
      }, 1000)
    }
    return
  }

  const bm = bookmarks[index]
  const urlDisplay = bm.url.length > 40 ? bm.url.substring(0, 40) + '...' : bm.url
  faviconStatus.value = '处理: ' + urlDisplay
  faviconCurrent.value = index + 1

  // 获取 favicon URL
  const favicon = getFaviconURL(bm.url)

  try {
    await faviconApi.update(bm.id, favicon)
  } catch (err) {
    // 失败时忽略，继续下一个
  }

  // 延迟处理下一个
  setTimeout(() => {
    processFavicons(bookmarks, index + 1)
  }, 50)
}

function getFaviconURL(url) {
  try {
    const domain = new URL(url).hostname
    return 'https://www.google.com/s2/favicons?domain=' + domain + '&sz=32'
  } catch (e) {
    return ''
  }
}

function minimizeFavicon() {
  faviconMinimized.value = true
  showFaviconDialog.value = false
}

function expandFavicon() {
  faviconMinimized.value = false
  showFaviconDialog.value = true
}

function cancelFavicon() {
  faviconCancelled.value = true
  showFaviconDialog.value = false
  faviconMinimized.value = false
}
</script>

<style lang="scss" scoped>
.app-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.top-navbar {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  padding: 0 20px;
  flex-shrink: 0;

  .navbar-left {
    display: flex;
    align-items: center;

    .logo {
      font-size: 20px;
      font-weight: 600;
      color: #303133;
    }
  }

  .navbar-center {
    flex: 1;
    display: flex;
    justify-content: center;
    padding: 0 40px;
  }

  .navbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.main-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: 260px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;

  h3 {
    margin: 0;
    font-size: 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
}

.sidebar-section {
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;

  .section-title {
    font-size: 13px;
    color: #909399;
    margin-bottom: 12px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
}

.folder-section {
  flex: 1;
  overflow: auto;
}

.folder-tree {
  .folder-item {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    cursor: pointer;
    border-radius: 6px;
    margin: 2px 0;
    transition: background 0.2s;

    &:hover {
      background: #f5f7fa;
    }

    &.active {
      background: #ecf5ff;
      color: var(--el-color-primary);
    }

    &.drag-over {
      background: #e6f7ff;
      border: 2px dashed #409eff;
    }

    .folder-icon {
      margin-right: 8px;
    }

    .folder-name {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f5f7fa;
}

.content-area {
  flex: 1;
  overflow: auto;
  padding: 20px;
}

.domain-sidebar {
  width: 280px;
  background: #fff;
  border-left: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.domain-search {
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
}

.domain-list {
  flex: 1;
  overflow: auto;
  padding: 8px 0;

  .domain-group {
    .domain-group-header {
      display: flex;
      align-items: center;
      padding: 8px 16px;
      cursor: pointer;
      transition: background 0.2s;

      &:hover {
        background: #f5f7fa;
      }

      .domain-caret {
        width: 16px;
        font-size: 10px;
        color: #909399;
        flex-shrink: 0;
      }

      .domain-name {
        flex: 1;
        font-size: 14px;
        color: #303133;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .credential-icon {
        color: #e6a23c;
        font-size: 12px;
        margin-left: 4px;
        flex-shrink: 0;
      }

      .domain-count {
        font-size: 12px;
        color: #909399;
        margin-left: 4px;
        flex-shrink: 0;
      }

      .domain-manage-btn {
        margin-left: 8px;
        opacity: 0;
        transition: opacity 0.2s;
      }

      &:hover .domain-manage-btn {
        opacity: 1;
      }
    }

    .domain-sublist {
      background: #fafafa;

      .domain-subitem {
        display: flex;
        align-items: center;
        padding: 6px 16px 6px 32px;
        cursor: pointer;
        transition: background 0.2s;

        &:hover {
          background: #f0f0f0;
        }

        .domain-subitem-name {
          flex: 1;
          font-size: 13px;
          color: #606266;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .domain-manage-btn {
          opacity: 0;
          transition: opacity 0.2s;
        }

        &:hover .domain-manage-btn {
          opacity: 1;
        }
      }
    }
  }

  .domain-empty {
    text-align: center;
    color: #909399;
    padding: 20px;
    font-size: 14px;
  }
}

.folder-move-hint {
  font-size: 13px;
  color: #909399;
  margin-top: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
}

// Favicon 加载弹窗样式
.favicon-progress {
  text-align: center;

  .progress-text {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 16px;
    color: #303133;
  }

  .status-text {
    margin-top: 16px;
    font-size: 13px;
    color: #909399;
    word-break: break-all;
  }
}

// 后台加载浮动条
.favicon-float {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 240px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: 12px;
  z-index: 1000;

  .favicon-float-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
    font-size: 14px;
    font-weight: 500;
  }

  .favicon-float-progress {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: #909399;

    .el-progress {
      flex: 1;
    }
  }
}

@media (max-width: 1200px) {
  .domain-sidebar {
    display: none;
  }
}

@media (max-width: 768px) {
  .sidebar {
    display: none;
  }

  .navbar-center {
    padding: 0 10px;
  }
}
</style>
