<template>
  <div class="folder-tree-node">
    <div
      class="folder-item"
      :class="{ active: currentFolder === folder.path && filterFolder, 'drag-over': isDragOver }"
      :style="{ paddingLeft: `${depth * 16 + 12}px` }"
      draggable="true"
      @click="$emit('select', folder.path, true)"
      @dragstart="handleDragStart"
      @dragend="handleDragEnd"
      @dragover="handleDragOver"
      @dragleave="handleDragLeave"
      @drop="handleDrop"
    >
      <el-icon
        v-if="folder.children?.length"
        class="expand-icon"
        @click.stop="expanded = !expanded"
      >
        <ArrowDown v-if="expanded" />
        <ArrowRight v-else />
      </el-icon>
      <span v-else class="expand-placeholder"></span>
      <el-icon class="folder-icon">
        <FolderOpened v-if="expanded" />
        <Folder v-else />
      </el-icon>
      <span class="folder-name">{{ folder.name }}</span>
      <span class="folder-count">{{ folder.count }}</span>
    </div>

    <div v-if="expanded && folder.children?.length" class="children">
      <FolderTreeNode
        v-for="child in folder.children"
        :key="child.path"
        :folder="child"
        :current-folder="currentFolder"
        :filter-folder="filterFolder"
        :depth="depth + 1"
        @select="(path, filter) => $emit('select', path, filter)"
        @folder-drop="(data) => $emit('folder-drop', data)"
      />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  folder: {
    type: Object,
    required: true
  },
  currentFolder: {
    type: String,
    default: ''
  },
  filterFolder: {
    type: Boolean,
    default: false
  },
  depth: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['select', 'folder-drop'])

const expanded = ref(false)
const isDragOver = ref(false)
const isDragging = ref(false)

function handleDragStart(e) {
  isDragging.value = true
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', props.folder.path)
}

function handleDragEnd() {
  isDragging.value = false
}

function handleDragOver(e) {
  e.preventDefault()
  const sourcePath = e.dataTransfer.getData('text/plain')
  // 不能拖到自己或自己的子文件夹
  if (sourcePath && sourcePath !== props.folder.path && !props.folder.path.startsWith(sourcePath + '/')) {
    isDragOver.value = true
    e.dataTransfer.dropEffect = 'move'
  }
}

function handleDragLeave() {
  isDragOver.value = false
}

function handleDrop(e) {
  e.preventDefault()
  isDragOver.value = false
  const sourcePath = e.dataTransfer.getData('text/plain')
  if (sourcePath && sourcePath !== props.folder.path && !props.folder.path.startsWith(sourcePath + '/')) {
    emit('folder-drop', {
      sourcePath,
      targetPath: props.folder.path
    })
  }
}
</script>

<style lang="scss" scoped>
.folder-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 6px;
  margin: 2px 0;
  user-select: none;
  transition: background 0.2s, border 0.2s;

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

  &[draggable="true"]:active {
    opacity: 0.6;
  }

  .expand-icon {
    font-size: 12px;
    color: #909399;
    margin-right: 4px;
    cursor: pointer;
  }

  .expand-placeholder {
    width: 12px;
    margin-right: 4px;
  }

  .folder-icon {
    margin-right: 8px;
    flex-shrink: 0;
  }

  .folder-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 14px;
  }

  .folder-count {
    font-size: 12px;
    color: #909399;
    margin-left: 8px;
  }
}
</style>
