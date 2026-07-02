<template>
  <div ref="toolbarRef" class="toolbar" :class="{ compact }">
    <div class="toolbar-left">
      <button @click="$emit('go-up')" class="tool-btn" :disabled="currentPath === '/'" title="上一级">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 15l-6-6-6 6"/>
        </svg>
      </button>
      <button @click="$emit('refresh')" class="tool-btn" title="刷新">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"></polyline>
          <polyline points="1 4 1 10 7 10"></polyline>
          <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"></path>
        </svg>
      </button>
    </div>
    
    <div class="path-bar">
      <span v-if="!compact" class="path-label">路径:</span>
      <input
        v-model="localPath"
        @keyup.enter="confirmPath"
        class="path-input"
        placeholder="/"
      />
      <button @click="confirmPath" class="path-go-btn" title="跳转">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </button>
    </div>
    
    <!-- 搜索框 -->
    <div class="search-bar" :class="{ 'searching': isSearching }">
      <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"></circle>
        <path d="m21 21-4.35-4.35"></path>
      </svg>
      <input 
        :value="searchKeyword"
        @input="$emit('search', $event.target.value)"
        @keyup.enter="$emit('search-enter')"
        class="search-input"
        placeholder="搜索文件（回车全目录搜索）..."
      />
      <button v-if="isSearching" @click="$emit('cancel-search')" class="cancel-search" title="取消搜索">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
      <button v-else-if="searchKeyword" @click="$emit('search', '')" class="clear-search" title="清除搜索">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>
    
    <div class="toolbar-right">
      <!-- 批量操作按钮 -->
      <button 
        v-if="selectedCount > 0" 
        @click="$emit('batch-download')" 
        class="action-btn batch-btn"
        :title="`下载选中的 ${selectedCount} 个文件`"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
          <polyline points="7 10 12 15 17 10"></polyline>
          <line x1="12" y1="15" x2="12" y2="3"></line>
        </svg>
        <span v-if="!compact">下载({{ selectedCount }})</span>
      </button>
      
      <button 
        v-if="selectedCount > 0" 
        @click="$emit('batch-delete')" 
        class="action-btn batch-btn danger"
        :title="`删除选中的 ${selectedCount} 个文件`"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6"></polyline>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
        </svg>
        <span v-if="!compact">删除({{ selectedCount }})</span>
      </button>
      
      <button 
        v-if="selectedCount > 0" 
        @click="$emit('batch-chmod')" 
        class="action-btn batch-btn"
        :title="`修改选中 ${selectedCount} 个文件的权限`"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
        </svg>
        <span v-if="!compact">权限({{ selectedCount }})</span>
      </button>
      
      <button v-if="hasCutFile" @click="$emit('paste')" class="action-btn" title="粘贴">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
          <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
        </svg>
        <span v-if="!compact">粘贴</span>
      </button>
      <button @click="$emit('new-folder')" class="action-btn" title="新建文件夹">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
          <line x1="12" y1="11" x2="12" y2="17"></line>
          <line x1="9" y1="14" x2="15" y2="14"></line>
        </svg>
        <span v-if="!compact">新建文件夹</span>
      </button>
      <button @click="$emit('upload')" class="action-btn" title="上传文件">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
          <polyline points="17 8 12 3 7 8"></polyline>
          <line x1="12" y1="3" x2="12" y2="15"></line>
        </svg>
        <span v-if="!compact">上传</span>
      </button>
      <!-- 上传目录按钮（前端自动压缩为 zip） -->
      <button @click="$emit('upload-directory')" class="action-btn" title="上传目录">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
          <polyline points="17 8 12 3 7 8"></polyline>
          <line x1="12" y1="3" x2="12" y2="15"></line>
        </svg>
        <span v-if="!compact">上传目录</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  currentPath: {
    type: String,
    required: true
  },
  hasCutFile: {
    type: Boolean,
    default: false
  },
  searchKeyword: {
    type: String,
    default: ''
  },
  selectedCount: {
    type: Number,
    default: 0
  },
  isSearching: {
    type: Boolean,
    default: false
  }
})

// 路径输入框：本地编辑，回车/点击→才跳转
const localPath = ref(props.currentPath)

// 当外部 currentPath 变化时（如终端 cd 同步），同步到输入框
watch(() => props.currentPath, (newPath) => {
  localPath.value = newPath
})

function confirmPath() {
  const p = localPath.value.trim()
  if (p) {
    emit('navigate', p)
  }
}

const emit = defineEmits([
  'go-up',
  'refresh',
  'navigate',
  'paste',
  'new-folder',
  'upload',
  'upload-directory',
  'search',
  'search-enter',
  'cancel-search',
  'batch-download',
  'batch-delete',
  'batch-chmod'
])
const toolbarRef = ref(null)
const compact = ref(false)
let resizeObs = null

onMounted(() => {
  const target = toolbarRef.value?.parentElement || toolbarRef.value
  if (target) {
    console.log('[FileToolbar] 观测目标:', target.className, '初始宽度:', target.clientWidth)
    resizeObs = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const w = entry.contentRect.width
        const prev = compact.value
        compact.value = w < 900
        if (prev !== compact.value) {
          console.log(`[FileToolbar] 宽度=${w}, compact=${compact.value}`)
        }
      }
    })
    resizeObs.observe(target)
  } else {
    console.warn('[FileToolbar] 未找到观测目标')
  }
})

onUnmounted(() => {
  if (resizeObs) resizeObs.disconnect()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--surface-hover);
  background: var(--surface-2);
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 0.5rem;
}

.tool-btn {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-hover);
  border: 1px solid var(--border-strong);
  border-radius: 0.375rem;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s;
}

.tool-btn:hover:not(:disabled) {
  background: var(--border-strong);
}

.tool-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.path-bar {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.path-label {
  color: var(--text-secondary);
  font-size: 0.8125rem;
}

.path-input {
  flex: 1;
  padding: 0.375rem 0.75rem;
  background: var(--bg-input);
  border: 1px solid var(--border-strong);
  border-radius: 0.375rem;
  color: var(--text-primary);
  font-size: 0.8125rem;
  font-family: monospace;
}

.path-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.path-go-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  background: var(--primary-bg);
  border: 1px solid var(--border-accent);
  border-radius: 0.375rem;
  color: var(--primary-light);
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.path-go-btn:hover {
  background: var(--primary-bg-hover);
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.75rem;
  background: var(--bg-input);
  border: 1px solid var(--border-strong);
  border-radius: 0.375rem;
  min-width: 200px;
  max-width: 300px;
}

.search-icon {
  color: var(--text-secondary);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 0.8125rem;
  outline: none;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.clear-search {
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  transition: all 0.2s;
}

.clear-search:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.cancel-search {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--danger-bg);
  border: none;
  color: var(--accent-danger);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  transition: all 0.2s;
}

.cancel-search:hover {
  background: var(--danger-bg);
  color: var(--danger-light);
}

.search-bar.searching {
  border-color: var(--accent-primary);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background: var(--primary-bg);
  border: 1px solid var(--border-accent);
  border-radius: 0.375rem;
  color: var(--primary-light);
  cursor: pointer;
  font-size: 0.8125rem;
  transition: all 0.2s;
}

.action-btn:hover {
  background: var(--primary-bg-hover);
}

.batch-btn {
  background: var(--success-bg);
  border-color: var(--border-success);
  color: var(--success-light);
}

.batch-btn:hover {
  background: var(--success-bg);
}

.batch-btn.danger {
  background: var(--danger-bg);
  border-color: var(--border-danger);
  color: var(--accent-danger);
}

.batch-btn.danger:hover {
  background: var(--danger-bg);
}
</style>

<style>
/* 工具栏基础：允许 flex 收缩（解决 min-width: auto 导致无法滚动的问题） */
.toolbar {
  min-width: 0 !important;
}

/* compact 模式：< 900px 时只显示图标，隐藏文字 */
.toolbar.compact {
  flex-wrap: nowrap !important;
  overflow-x: auto !important;
  overflow-y: hidden !important;
  gap: 0.25rem !important;
  padding: 0.375rem 0.5rem !important;
  scrollbar-width: thin !important;
  scrollbar-color: var(--primary-bg) transparent !important;
}

.toolbar.compact::-webkit-scrollbar {
  height: 3px !important;
}

.toolbar.compact::-webkit-scrollbar-track {
  background: transparent !important;
}

.toolbar.compact::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb) !important;
  border-radius: 999px !important;
}

.toolbar.compact::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover) !important;
}

.toolbar.compact .toolbar-left,
.toolbar.compact .toolbar-right {
  flex-shrink: 0 !important;
  gap: 0.25rem !important;
}

.toolbar.compact .tool-btn {
  width: 1.75rem !important;
  height: 1.75rem !important;
  flex-shrink: 0 !important;
}

.toolbar.compact .path-bar {
  flex: 0 0 120px !important;
  width: 120px !important;
  min-width: 120px !important;
  flex-shrink: 0 !important;
}

.toolbar.compact .path-input {
  font-size: 0.75rem !important;
  padding: 0.25rem 0.5rem !important;
  width: 100% !important;
  box-sizing: border-box !important;
}

.toolbar.compact .search-bar {
  flex: 0 0 120px !important;
  width: 120px !important;
  min-width: 120px !important;
  max-width: 120px !important;
  flex-shrink: 0 !important;
  padding: 0.25rem 0.5rem !important;
}

.toolbar.compact .search-input {
  font-size: 0.75rem !important;
  width: 100% !important;
}

.toolbar.compact .action-btn {
  flex-shrink: 0 !important;
  padding: 0.25rem !important;
  gap: 0 !important;
}

.toolbar.compact .batch-btn {
  flex-shrink: 0 !important;
  padding: 0.25rem !important;
  gap: 0 !important;
}
</style>
