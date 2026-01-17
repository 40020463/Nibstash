<template>
  <el-dialog
    v-model="visible"
    :title="domain"
    width="600px"
    @close="handleClose"
  >
    <div class="domain-modal">
      <!-- 域名信息 -->
      <div class="domain-info">
        <span class="bookmark-count">{{ bookmarks.length }} 个相关书签</span>
      </div>

      <!-- 标签切换 -->
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="账号密码" name="credentials">
          <!-- 凭证列表 -->
          <div v-if="!showCredentialForm" class="credentials-list">
            <div v-if="credentials.length === 0" class="empty-tip">
              暂无保存的账号信息
            </div>
            <div
              v-for="(cred, index) in credentials"
              :key="cred.id"
              class="credential-card"
            >
              <div class="credential-header">
                <span class="credential-title">{{ cred.title || '账号 ' + (index + 1) }}</span>
                <div class="credential-actions">
                  <el-button link size="small" @click="editCredential(cred)">编辑</el-button>
                  <el-button link size="small" type="danger" @click="deleteCredential(cred.id)">删除</el-button>
                </div>
              </div>
              <div class="credential-body">
                <div class="credential-field">
                  <span class="field-label">用户名:</span>
                  <span class="field-value">{{ cred.username || '-' }}</span>
                  <el-button link size="small" @click="copyText(cred.username)">复制</el-button>
                </div>
                <div class="credential-field">
                  <span class="field-label">密码:</span>
                  <span class="field-value">{{ cred.showPassword ? cred.password : '••••••••' }}</span>
                  <el-button link size="small" @click="cred.showPassword = !cred.showPassword">
                    {{ cred.showPassword ? '隐藏' : '显示' }}
                  </el-button>
                  <el-button link size="small" @click="copyText(cred.password)">复制</el-button>
                </div>
                <div v-if="cred.notes" class="credential-field">
                  <span class="field-label">备注:</span>
                  <span class="field-value notes">{{ cred.notes }}</span>
                </div>
              </div>
            </div>
            <el-button style="width: 100%; margin-top: 12px" @click="addCredential">
              + 添加新账号
            </el-button>
          </div>

          <!-- 凭证编辑表单 -->
          <div v-else class="credential-form">
            <el-form :model="credentialForm" label-width="80px">
              <el-form-item label="备注标题">
                <el-input v-model="credentialForm.title" placeholder="例如：工作账号" />
              </el-form-item>
              <el-form-item label="用户名">
                <el-input v-model="credentialForm.username" placeholder="输入用户名或邮箱" />
              </el-form-item>
              <el-form-item label="密码">
                <el-input v-model="credentialForm.password" placeholder="输入密码" />
              </el-form-item>
              <el-form-item label="备注">
                <el-input v-model="credentialForm.notes" type="textarea" :rows="3" placeholder="其他备注信息" />
              </el-form-item>
            </el-form>
            <div class="form-actions">
              <el-button @click="cancelCredentialForm">取消</el-button>
              <el-button type="primary" :loading="saving" @click="saveCredential">保存</el-button>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="相关书签" name="bookmarks">
          <div v-if="bookmarks.length === 0" class="empty-tip">
            暂无相关书签
          </div>
          <div v-for="bm in bookmarks" :key="bm.id" class="bookmark-item">
            <img
              v-if="bm.favicon"
              :src="bm.favicon"
              class="bookmark-favicon"
              @error="e => e.target.style.display = 'none'"
            />
            <el-icon v-else class="bookmark-favicon-placeholder"><Link /></el-icon>
            <div class="bookmark-info">
              <a :href="bm.url" target="_blank" class="bookmark-title">{{ bm.title || '无标题' }}</a>
              <div class="bookmark-url">{{ truncateUrl(bm.url, 50) }}</div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <template #footer>
      <el-button type="danger" @click="handleDeleteDomain">删除此域名</el-button>
      <el-button @click="visible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { credentialApi, domainApi } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'

const props = defineProps({
  modelValue: Boolean,
  domain: String
})

const emit = defineEmits(['update:modelValue', 'deleted', 'refresh'])

const visible = ref(false)
const activeTab = ref('credentials')
const credentials = ref([])
const bookmarks = ref([])
const showCredentialForm = ref(false)
const editingCredentialId = ref(null)
const saving = ref(false)
const credentialForm = ref({
  title: '',
  username: '',
  password: '',
  notes: ''
})

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val && props.domain) {
    loadCredentials()
    loadBookmarks()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

async function loadCredentials() {
  try {
    const data = await credentialApi.getByDomain(props.domain)
    credentials.value = (data || []).map(c => ({ ...c, showPassword: false }))
  } catch (err) {
    credentials.value = []
  }
}

async function loadBookmarks() {
  try {
    const data = await domainApi.getBookmarks(props.domain)
    bookmarks.value = data.bookmarks || []
  } catch (err) {
    bookmarks.value = []
  }
}

function handleTabChange(tab) {
  if (tab === 'bookmarks' && bookmarks.value.length === 0) {
    loadBookmarks()
  }
}

function addCredential() {
  editingCredentialId.value = null
  credentialForm.value = { title: '', username: '', password: '', notes: '' }
  showCredentialForm.value = true
}

function editCredential(cred) {
  editingCredentialId.value = cred.id
  credentialForm.value = {
    title: cred.title || '',
    username: cred.username || '',
    password: cred.password || '',
    notes: cred.notes || ''
  }
  showCredentialForm.value = true
}

function cancelCredentialForm() {
  showCredentialForm.value = false
}

async function saveCredential() {
  saving.value = true
  try {
    if (editingCredentialId.value) {
      await credentialApi.update(editingCredentialId.value, credentialForm.value)
    } else {
      await credentialApi.create({
        ...credentialForm.value,
        domain: props.domain
      })
    }
    ElMessage.success('保存成功')
    showCredentialForm.value = false
    loadCredentials()
  } catch (err) {
    ElMessage.error(err.error || '保存失败')
  } finally {
    saving.value = false
  }
}

async function deleteCredential(id) {
  try {
    await ElMessageBox.confirm('确定要删除这个账号信息吗？', '提示', { type: 'warning' })
    await credentialApi.delete(id)
    ElMessage.success('删除成功')
    loadCredentials()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.error || '删除失败')
    }
  }
}

async function handleDeleteDomain() {
  try {
    await ElMessageBox.confirm(`确定要删除域名 "${props.domain}" 的所有凭证数据吗？（不会删除相关书签）`, '警告', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消'
    })
    await domainApi.delete(props.domain)
    ElMessage.success('删除成功')
    visible.value = false
    emit('deleted')
    emit('refresh')
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.error || '删除失败')
    }
  }
}

function handleClose() {
  showCredentialForm.value = false
  activeTab.value = 'credentials'
}

function copyText(text) {
  if (!text) return
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

function truncateUrl(url, maxLen) {
  if (!url || url.length <= maxLen) return url
  const shortUrl = url.replace(/^https?:\/\//, '')
  if (shortUrl.length <= maxLen) return shortUrl
  return shortUrl.substring(0, maxLen - 3) + '...'
}
</script>

<style lang="scss" scoped>
.domain-modal {
  .domain-info {
    margin-bottom: 16px;
    color: #909399;
    font-size: 13px;
  }
}

.credentials-list {
  .empty-tip {
    text-align: center;
    color: #909399;
    padding: 20px;
  }
}

.credential-card {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;

  .credential-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;

    .credential-title {
      font-weight: 500;
    }
  }

  .credential-body {
    .credential-field {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 6px;
      font-size: 13px;

      .field-label {
        color: #909399;
        width: 50px;
        flex-shrink: 0;
      }

      .field-value {
        flex: 1;
        word-break: break-all;

        &.notes {
          white-space: pre-wrap;
        }
      }
    }
  }
}

.credential-form {
  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 16px;
  }
}

.bookmark-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #ebeef5;

  &:last-child {
    border-bottom: none;
  }

  .bookmark-favicon {
    width: 20px;
    height: 20px;
    border-radius: 4px;
    flex-shrink: 0;
  }

  .bookmark-favicon-placeholder {
    width: 20px;
    height: 20px;
    color: #909399;
    flex-shrink: 0;
  }

  .bookmark-info {
    flex: 1;
    overflow: hidden;

    .bookmark-title {
      display: block;
      color: #303133;
      text-decoration: none;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;

      &:hover {
        color: var(--el-color-primary);
      }
    }

    .bookmark-url {
      font-size: 12px;
      color: #909399;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
}

.empty-tip {
  text-align: center;
  color: #909399;
  padding: 20px;
}
</style>
