<template>
  <Teleport to="body">
    <div v-if="visible" class="context-menu" :style="{ top: position.y + 'px', left: position.x + 'px' }" @click.stop>
      <div class="menu-item" @click="$emit('preview', file)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
          <circle cx="12" cy="12" r="3"></circle>
        </svg>
        <span>预览/编辑</span>
      </div>
      <div class="menu-item" @click="$emit('download', file)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
          <polyline points="7 10 12 15 17 10"></polyline>
          <line x1="12" y1="15" x2="12" y2="3"></line>
        </svg>
        <span>下载</span>
      </div>
      <div class="menu-item" @click="$emit('compress', file)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
          <polyline points="7 10 12 15 17 10"></polyline>
          <line x1="12" y1="15" x2="12" y2="3"></line>
        </svg>
        <span>压缩为 ZIP</span>
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="$emit('rename', file)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
        </svg>
        <span>重命名</span>
      </div>
      <div class="menu-item" @click="$emit('duplicate', file)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
        </svg>
        <span>创建副本</span>
      </div>
      <div class="menu-item" @click="$emit('cut', file)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="6" cy="6" r="3"></circle>
          <circle cx="6" cy="18" r="3"></circle>
          <line x1="20" y1="4" x2="8.12" y2="15.88"></line>
          <line x1="14.47" y1="14.48" x2="20" y2="20"></line>
          <line x1="8.12" y1="8.12" x2="12" y2="12"></line>
        </svg>
        <span>剪切</span>
      </div>
      <div class="menu-divider"></div>
      <div class="menu-item" @click="$emit('chmod', file)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
        </svg>
        <span>修改权限</span>
      </div>
      <div class="menu-item danger" @click="$emit('delete', file)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6"></polyline>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
        </svg>
        <span>删除</span>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  position: {
    type: Object,
    required: true
  },
  file: {
    type: Object,
    default: null
  }
})

defineEmits([
  'update:visible',
  'preview',
  'download',
  'compress',
  'rename',
  'duplicate',
  'cut',
  'chmod',
  'delete'
])
</script>

<style>
.context-menu {
  position: fixed;
  min-width: 180px;
  background: var(--bg-tooltip);
  border: 1px solid var(--border-strong);
  border-radius: 0.5rem;
  box-shadow: 0 4px 12px var(--shadow-lg);
  z-index: 9999;
  padding: 0.5rem 0;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 1rem;
  color: var(--text-primary);
  cursor: pointer;
  transition: background 0.15s;
  font-size: 0.875rem;
}

.menu-item:hover {
  background: var(--primary-bg);
}

.menu-item.danger {
  color: var(--accent-danger);
}

.menu-item.danger:hover {
  background: var(--danger-bg);
}

.menu-divider {
  height: 1px;
  background: var(--surface-hover);
  margin: 0.5rem 0;
}
</style>
