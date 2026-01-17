<template>
  <div class="tags-page">
    <div class="page-header">
      <h2>标签管理</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon> 新建标签
      </el-button>
    </div>

    <div v-loading="tagStore.loading" class="tag-grid">
      <div v-for="tag in tagStore.tags" :key="tag.id" class="tag-card">
        <div class="tag-color" :style="{ backgroundColor: tag.color }"></div>
        <div class="tag-info">
          <div class="tag-name">{{ tag.name }}</div>
          <div class="tag-count">{{ tag.count }} 个书签</div>
        </div>
        <div class="tag-actions">
          <el-button link @click="editTag(tag)">
            <el-icon><Edit /></el-icon>
          </el-button>
          <el-button link type="danger" @click="deleteTag(tag)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>

      <el-empty v-if="!tagStore.loading && tagStore.tags.length === 0" description="暂无标签" />
    </div>

    <!-- 创建/编辑弹窗 -->
    <el-dialog v-model="showDialog" :title="editingTag ? '编辑标签' : '新建标签'" width="400px">
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="标签名称" />
        </el-form-item>
        <el-form-item label="颜色">
          <el-color-picker v-model="form.color" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useTagStore } from '@/stores/tag'
import { ElMessage, ElMessageBox } from 'element-plus'

const tagStore = useTagStore()

const showCreateDialog = ref(false)
const editingTag = ref(null)
const form = ref({
  name: '',
  color: '#3b82f6'
})

const showDialog = computed({
  get: () => showCreateDialog.value || !!editingTag.value,
  set: (val) => {
    if (!val) {
      showCreateDialog.value = false
      editingTag.value = null
      form.value = { name: '', color: '#3b82f6' }
    }
  }
})

function editTag(tag) {
  editingTag.value = tag
  form.value = { name: tag.name, color: tag.color }
}

async function deleteTag(tag) {
  try {
    await ElMessageBox.confirm(`确定要删除标签"${tag.name}"吗？`, '提示', {
      type: 'warning'
    })
    await tagStore.deleteTag(tag.id)
    ElMessage.success('删除成功')
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.error || '删除失败')
    }
  }
}

async function handleSubmit() {
  if (!form.value.name) {
    ElMessage.warning('请输入标签名称')
    return
  }

  try {
    if (editingTag.value) {
      await tagStore.updateTag(editingTag.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await tagStore.createTag(form.value.name, form.value.color)
      ElMessage.success('创建成功')
    }
    showDialog.value = false
  } catch (err) {
    ElMessage.error(err.error || '操作失败')
  }
}
</script>

<style lang="scss" scoped>
.tags-page {
  height: 100%;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;

  h2 {
    margin: 0;
  }
}

.tag-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.tag-card {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);

  .tag-color {
    width: 40px;
    height: 40px;
    border-radius: 8px;
  }

  .tag-info {
    flex: 1;
  }

  .tag-name {
    font-size: 15px;
    font-weight: 500;
    color: #303133;
  }

  .tag-count {
    font-size: 13px;
    color: #909399;
  }

  .tag-actions {
    display: flex;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.2s;
  }

  &:hover .tag-actions {
    opacity: 1;
  }
}
</style>
