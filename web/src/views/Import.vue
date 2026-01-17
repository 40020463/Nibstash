<template>
  <div class="import-page">
    <div class="page-header">
      <h2>导入书签</h2>
    </div>

    <div class="import-card">
      <el-upload
        ref="uploadRef"
        drag
        :auto-upload="false"
        :limit="1"
        accept=".html,.htm"
        :on-change="handleFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将书签文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            支持从 Chrome、Firefox、Edge 等浏览器导出的 HTML 书签文件
          </div>
        </template>
      </el-upload>

      <div v-if="selectedFile" class="file-info">
        <el-icon><Document /></el-icon>
        <span>{{ selectedFile.name }}</span>
        <el-button link type="danger" @click="clearFile">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>

      <el-button
        type="primary"
        size="large"
        :loading="importing"
        :disabled="!selectedFile"
        style="width: 100%; margin-top: 20px"
        @click="handleImport"
      >
        开始导入
      </el-button>
    </div>

    <div v-if="result" class="result-card">
      <el-result
        :icon="result.imported > 0 ? 'success' : 'warning'"
        :title="result.imported > 0 ? '导入完成' : '导入完成（无新增）'"
      >
        <template #sub-title>
          <p>共处理 {{ result.total }} 个书签</p>
          <p>成功导入 {{ result.imported }} 个</p>
          <p>跳过重复 {{ result.skipped }} 个</p>
        </template>
        <template #extra>
          <el-button type="primary" @click="$router.push('/')">查看书签</el-button>
        </template>
      </el-result>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { bookmarkApi } from '@/api'
import { useBookmarkStore } from '@/stores/bookmark'
import { useFolderStore } from '@/stores/folder'
import { useDomainStore } from '@/stores/domain'
import { ElMessage } from 'element-plus'

const bookmarkStore = useBookmarkStore()
const folderStore = useFolderStore()
const domainStore = useDomainStore()

const uploadRef = ref(null)
const selectedFile = ref(null)
const importing = ref(false)
const result = ref(null)

function handleFileChange(file) {
  selectedFile.value = file.raw
}

function clearFile() {
  selectedFile.value = null
  uploadRef.value?.clearFiles()
  result.value = null
}

async function handleImport() {
  if (!selectedFile.value) return

  importing.value = true
  result.value = null

  try {
    const res = await bookmarkApi.import(selectedFile.value)
    result.value = res
    bookmarkStore.fetchBookmarks()
    folderStore.fetchFolders()
    domainStore.fetchDomains()
    ElMessage.success('导入完成')
  } catch (err) {
    ElMessage.error(err.error || '导入失败')
  } finally {
    importing.value = false
  }
}
</script>

<style lang="scss" scoped>
.import-page {
  max-width: 600px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;

  h2 {
    margin: 0;
  }
}

.import-card {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
  margin-top: 16px;

  span {
    flex: 1;
  }
}

.result-card {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  margin-top: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}
</style>
