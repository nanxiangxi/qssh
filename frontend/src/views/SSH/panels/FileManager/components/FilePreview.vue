<template>
  <Modal
    :visible="visible"
    :show-header="false"
    :show-footer="false"
    :show-close="false"
    :close-on-overlay="false"
    width="95%"
    height="90%"
    :border-radius="6"
    :body-padding="0"
    background-color="rgba(30, 30, 30, 0.98)"
    border-color="rgba(255, 255, 255, 0.1)"
    :overlay-opacity="0.85"
    :overlay-blur="6"
    @close="$emit('close')"
  >
    <div class="code-editor-window">
      <!-- 简化的顶部工具栏 -->
      <div class="editor-toolbar">
        <div class="toolbar-left">
          <!-- 文件信息 -->
          <div class="file-info">
            <span class="file-name">{{ fileName }}</span>
            <span class="language-badge" :class="detectedLanguage">
              {{ detectedLanguage.toUpperCase() }}
            </span>
          </div>
        </div>
        
        <div class="toolbar-right">
          <!-- 功能按钮组 - 仅文本文件显示编辑按钮 -->
          <button 
            v-if="!editable && type === 'text'" 
            @click="$emit('enable-edit')" 
            class="toolbar-btn"
            title="编辑 (Ctrl+E)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
            </svg>
            <span>编辑</span>
          </button>
          
          <button 
            v-if="editable" 
            @click="handleSave" 
            class="toolbar-btn primary"
            :disabled="saving"
            title="保存 (Ctrl+S)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path>
              <polyline points="17 21 17 13 7 13 7 21"></polyline>
              <polyline points="7 3 7 8 15 8"></polyline>
            </svg>
            <span>{{ saving ? '保存中...' : '保存' }}</span>
          </button>
          
          <!-- 关闭按钮 -->
          <button @click="$emit('close')" class="close-btn" title="关闭 (Esc)">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>
      
      <!-- 编辑器主体 -->
      <div class="editor-content">
        <!-- 图片预览 -->
        <div v-if="type === 'image'" class="media-preview">
          <img :src="content" alt="" />
        </div>
        
        <!-- 视频预览 -->
        <div v-else-if="type === 'video'" class="media-preview">
          <video :src="content" controls autoplay></video>
        </div>
        
        <!-- 音频预览 -->
        <div v-else-if="type === 'audio'" class="media-preview audio-player">
          <audio :src="content" controls autoplay></audio>
        </div>
        
        <!-- 代码编辑器 -->
        <CodeEditor
          v-else-if="type === 'text'"
          ref="codeEditorRef"
          :content="content"
          :language="detectedLanguage"
          :readonly="!editable"
          :word-wrap="wordWrap"
          @update:content="$emit('update:content', $event)"
        />
      </div>
      
      <!-- 底部状态栏 -->
      <div class="editor-statusbar">
        <div class="statusbar-left">
          <span v-if="editable" class="status-item modified" v-show="isModified">
            ● 已修改
          </span>
          <span class="status-item">
            行 {{ cursorPosition.line }}, 列 {{ cursorPosition.column }}
          </span>
        </div>
        <div class="statusbar-right">
          <span class="status-item">
            UTF-8
          </span>
          <span class="status-item">
            {{ detectedLanguage.toUpperCase() }}
          </span>
        </div>
      </div>
    </div>
  </Modal>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import Modal from '@/components/Modal.vue'
import CodeEditor from './CodeEditor.vue'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  title: {
    type: String,
    required: true
  },
  content: {
    type: String,
    required: true
  },
  type: {
    type: String,
    default: 'text' // 'text', 'image', 'video', 'audio'
  },
  editable: {
    type: Boolean,
    default: false
  },
  fileName: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'update:visible',
  'update:content',
  'close',
  'enable-edit',
  'save'
])

const codeEditorRef = ref(null)
const saving = ref(false)
const wordWrap = ref(true)
const isModified = ref(false)
const cursorPosition = ref({ line: 1, column: 1 })
let originalContent = ''

// 根据文件扩展名检测语言（常用语言）
const detectedLanguage = computed(() => {
  if (!props.fileName) return 'plaintext'
  
  const ext = props.fileName.split('.').pop().toLowerCase()
  
  // ✅ 常用语言映射
  const languageMap = {
    // JavaScript/TypeScript
    'js': 'javascript',
    'jsx': 'javascript',
    'ts': 'typescript',
    'tsx': 'typescript',
    'mjs': 'javascript',
    'cjs': 'javascript',
    // Python
    'py': 'python',
    'pyw': 'python',
    // Java
    'java': 'java',
    // C/C++
    'c': 'c',
    'cpp': 'cpp',
    'cc': 'cpp',
    'cxx': 'cpp',
    'h': 'c',
    'hpp': 'cpp',
    // Ruby
    'rb': 'ruby',
    // Web
    'html': 'html',
    'htm': 'html',
    'xhtml': 'html',
    'css': 'css',
    'scss': 'css',
    'sass': 'css',
    'less': 'css',
    'vue': 'html',
    'svelte': 'html',
    // Data formats
    'json': 'json',
    'json5': 'json',
    'xml': 'xml',
    'svg': 'xml',
    'yaml': 'yaml',
    'yml': 'yaml',
    // SQL
    'sql': 'sql',
    // Markdown
    'md': 'markdown',
    'markdown': 'markdown',
    // Plain text (所有其他文件)
    'txt': 'plaintext',
    'log': 'plaintext',
    'csv': 'plaintext',
    'tsv': 'plaintext',
    'ini': 'plaintext',
    'conf': 'plaintext',
    'cfg': 'plaintext',
    'env': 'plaintext',
    'properties': 'plaintext',
    'toml': 'plaintext',
    'editorconfig': 'plaintext',
    'gitignore': 'plaintext',
    'dockerfile': 'plaintext',
    'makefile': 'plaintext',
    'gradle': 'plaintext',
    'iml': 'xml',
    'sh': 'shell',
    'bash': 'shell',
    'zsh': 'shell',
    'ps1': 'plaintext',
    'bat': 'plaintext',
    'cmd': 'plaintext',
  }
  
  return languageMap[ext] || 'plaintext'
})

// 处理保存
const handleSave = async () => {
  if (saving.value) return
  
  saving.value = true
  try {
    // ✅ 从 CodeEditor 获取最新内容
    if (codeEditorRef.value && props.type === 'text') {
      const newContent = codeEditorRef.value.getValue()
      // 先更新父组件的 content
      emit('update:content', newContent)
    }
    // 然后触发保存事件
    await emit('save')
    isModified.value = false
    originalContent = props.content
  } finally {
    saving.value = false
  }
}

// 监听内容变化，检测是否修改
watch(() => props.content, (newContent) => {
  if (originalContent && newContent !== originalContent) {
    isModified.value = true
  }
})

// 监听编辑状态变化
watch(() => props.editable, (newEditable) => {
  if (newEditable) {
    originalContent = props.content
    isModified.value = false
  }
})

// 键盘快捷键
const handleKeyDown = (e) => {
  if (!props.visible) return
  
  // Ctrl+S 保存
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    if (props.editable) {
      handleSave()
    }
  }
  
  // Ctrl+E 编辑
  if ((e.ctrlKey || e.metaKey) && e.key === 'e') {
    e.preventDefault()
    if (!props.editable) {
      emit('enable-edit')
    }
  }
  
  // Esc 关闭
  if (e.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<style scoped>
/* ========== 编辑器窗口 ========== */
.code-editor-window {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  border-radius: 6px;
  overflow: hidden;
}

/* ========== 顶部工具栏 ========== */
.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  background: #2d2d2d;
  border-bottom: 1px solid #3e3e3e;
  flex-shrink: 0;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

/* 工具栏按钮 - VS Code 风格 */
.toolbar-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.625rem;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #a0aec0;
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.15s;
}

.toolbar-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #e2e8f0;
}

.toolbar-btn.primary {
  color: #63b3ed;
}

.toolbar-btn.primary:hover {
  background: rgba(66, 153, 225, 0.15);
  color: #90cdf4;
}

.toolbar-btn.active {
  background: rgba(66, 153, 225, 0.2);
  color: #63b3ed;
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 文件信息 */
.file-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0 0.75rem;
}

.file-name {
  color: #e2e8f0;
  font-size: 0.8125rem;
  font-weight: 500;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.language-badge {
  padding: 0.125rem 0.5rem;
  background: rgba(66, 153, 225, 0.2);
  border-radius: 3px;
  color: #63b3ed;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
}

/* 关闭按钮 */
.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #a0aec0;
  cursor: pointer;
  transition: all 0.15s;
}

.close-btn:hover {
  background: rgba(245, 101, 101, 0.15);
  color: #fc8181;
}

/* ========== 编辑器主体 ========== */
.editor-content {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.media-preview {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 2rem;
  background: rgba(25, 25, 25, 0.5);
}

.media-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.media-preview video {
  max-width: 100%;
  max-height: 100%;
  outline: none;
}

.audio-player {
  padding: 3rem;
}

.audio-player audio {
  width: 100%;
  max-width: 600px;
  outline: none;
}

/* ========== 底部状态栏 ========== */
.editor-statusbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.375rem 1rem;
  background: #2d2d2d;
  border-top: 1px solid #3e3e3e;
  font-size: 0.75rem;
  flex-shrink: 0;
}

.statusbar-left,
.statusbar-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.status-item {
  color: #a0aec0;
  font-size: 0.75rem;
}

.status-item.modified {
  color: #f6ad55;
  font-weight: 500;
}
</style>
