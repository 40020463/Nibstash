<template>
  <el-dialog
    :model-value="modelValue"
    :title="bookmark ? '编辑书签' : '添加'"
    width="500px"
    @update:model-value="$emit('update:modelValue', $event)"
    @close="resetForm"
  >
    <!-- 标签切换（仅新增时显示） -->
    <el-tabs v-if="!bookmark" v-model="activeTab">
      <el-tab-pane label="添加收藏" name="bookmark" />
      <el-tab-pane label="添加文件夹" name="folder" />
    </el-tabs>

    <!-- 添加收藏表单 -->
    <el-form v-if="activeTab === 'bookmark'" :model="form" label-width="80px">
      <el-form-item label="URL" required>
        <el-input v-model="form.url" placeholder="https://example.com" />
      </el-form-item>
      <el-form-item label="标题">
        <el-input v-model="form.title" placeholder="留空自动获取" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="2"
          placeholder="可选描述"
        />
      </el-form-item>
      <el-form-item label="文件夹">
        <el-select v-model="form.folder_path" placeholder="选择文件夹" clearable style="width: 100%">
          <el-option label="根目录" value="" />
          <el-option
            v-for="path in folderPaths"
            :key="path"
            :label="path"
            :value="path"
          />
        </el-select>
      </el-form-item>
    </el-form>

    <!-- 添加文件夹表单 -->
    <el-form v-else :model="folderForm" label-width="80px">
      <el-form-item label="文件夹名">
        <el-input v-model="folderForm.name" placeholder="新建文件夹" />
      </el-form-item>
      <el-form-item label="父文件夹">
        <el-select v-model="folderForm.parent" placeholder="选择父文件夹" clearable style="width: 100%">
          <el-option label="根目录" value="" />
          <el-option
            v-for="path in folderPaths"
            :key="path"
            :label="path"
            :value="path"
          />
        </el-select>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        {{ bookmark ? '保存' : '确定' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useBookmarkStore } from '@/stores/bookmark'
import { useFolderStore } from '@/stores/folder'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  bookmark: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const bookmarkStore = useBookmarkStore()
const folderStore = useFolderStore()

const loading = ref(false)
const activeTab = ref('bookmark')
const form = ref({
  url: '',
  title: '',
  description: '',
  folder_path: ''
})
const folderForm = ref({
  name: '',
  parent: ''
})

const folderPaths = computed(() => {
  const paths = []
  function collect(nodes) {
    for (const node of nodes) {
      paths.push(node.path)
      if (node.children?.length) {
        collect(node.children)
      }
    }
  }
  collect(folderStore.folders)
  return paths
})

watch(() => props.bookmark, (val) => {
  if (val) {
    form.value = {
      url: val.url,
      title: val.title,
      description: val.description || '',
      folder_path: val.folder_path || ''
    }
    activeTab.value = 'bookmark'
  }
}, { immediate: true })

function resetForm() {
  form.value = {
    url: '',
    title: '',
    description: '',
    folder_path: ''
  }
  folderForm.value = {
    name: '',
    parent: ''
  }
  activeTab.value = 'bookmark'
}

async function handleSubmit() {
  if (activeTab.value === 'bookmark') {
    await submitBookmark()
  } else {
    await submitFolder()
  }
}

async function submitBookmark() {
  if (!form.value.url) {
    ElMessage.warning('请输入 URL')
    return
  }

  // 自动添加协议
  let url = form.value.url.trim()
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    url = 'https://' + url
  }

  loading.value = true
  try {
    const data = {
      url,
      title: form.value.title || url,
      description: form.value.description,
      folder_path: form.value.folder_path
    }

    if (props.bookmark) {
      await bookmarkStore.updateBookmark(props.bookmark.id, data)
      ElMessage.success('更新成功')
    } else {
      await bookmarkStore.createBookmark(data)
      ElMessage.success('添加成功')
    }
    emit('update:modelValue', false)
    emit('success')
    resetForm()
  } catch (err) {
    ElMessage.error(err.error || '操作失败')
  } finally {
    loading.value = false
  }
}

async function submitFolder() {
  const name = folderForm.value.name.trim() || '新建文件夹'
  const parent = folderForm.value.parent

  const path = parent ? `${parent}/${name}` : name

  loading.value = true
  try {
    await folderStore.createFolder(path)
    ElMessage.success('创建成功')
    emit('update:modelValue', false)
    emit('success')
    resetForm()
  } catch (err) {
    ElMessage.error(err.error || '创建失败')
  } finally {
    loading.value = false
  }
}
</script>
