<template>
  <div class="terminal-block" :class="[block.type, block.status, { collapsed: block.collapsed }]" :data-block-id="block.id">
    <!-- 块头部 -->
    <div class="block-header" @click="toggleCollapse">
        <!-- 类型图标 -->
        <div class="block-icon">
        <!-- 命令图标 -->
        <svg v-if="block.type === 'command'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"/>
          <line x1="12" y1="19" x2="20" y2="19"/>
        </svg>
        <!-- 输出图标 -->
        <svg v-else-if="block.type === 'output'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
        </svg>
        <!-- 错误图标 -->
        <svg v-else-if="block.type === 'error'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="15" y1="9" x2="9" y2="15"/>
          <line x1="9" y1="9" x2="15" y2="15"/>
        </svg>
        <!-- 系统图标 -->
        <svg v-else-if="block.type === 'system'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="16" x2="12" y2="12"/>
          <line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
        <!-- 输入图标 -->
        <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
        </svg>
      </div>

      <!-- 命令文本 -->
      <div class="block-title" v-if="block.type === 'command'">
        <span class="command-text">{{ block.command }}</span>
      </div>

      <!-- 系统消息 -->
      <div class="block-title" v-else-if="block.type === 'system'">
        <span class="system-text">{{ block.content }}</span>
      </div>

      <!-- 其他类型 -->
      <div class="block-title" v-else>
        <span class="type-label">{{ typeLabel }}</span>
      </div>

      <!-- 状态和元数据 -->
      <div class="block-meta">
        <!-- 运行状态 -->
        <span v-if="block.status === 'running'" class="status-badge running">
          <span class="spinner"></span>
          运行中
        </span>
        <span v-else-if="block.status === 'success'" class="status-badge success">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          {{ block.exitCode === 0 ? '成功' : `退出码: ${block.exitCode}` }}
        </span>
        <span v-else-if="block.status === 'failed'" class="status-badge failed">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
          失败
        </span>

        <!-- 耗时 -->
        <span v-if="block.duration" class="duration">
          {{ formatDuration(block.duration) }}
        </span>

        <!-- 时间戳 -->
        <span class="timestamp">
          {{ formatTime(block.timestamp) }}
        </span>
      </div>

      <!-- 折叠按钮 -->
      <button class="collapse-btn" v-if="hasContent">
        <svg
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          :class="{ rotated: !block.collapsed }"
        >
          <polyline points="6 9 12 15 18 9"/>
        </svg>
      </button>
    </div>

    <!-- 块内容 -->
    <Transition name="content">
      <div v-if="!block.collapsed && hasContent" class="block-content">
        <!-- 命令输出 -->
        <div v-if="block.type === 'command'" class="output-content">
          <pre v-if="displayContent" class="output-text" v-html="highlightedContent"></pre>
          <pre v-if="displayError" class="error-text" v-html="highlightedError"></pre>
          <div v-if="!displayContent && !displayError && block.status === 'running'" class="waiting-output">
            <span class="dot-pulse"></span>
          </div>
        </div>

        <!-- 输出块 -->
        <div v-else-if="block.type === 'output'" class="output-content">
          <pre class="output-text" v-html="highlightedContent"></pre>
        </div>

        <!-- 错误块 -->
        <div v-else-if="block.type === 'error'" class="error-content">
          <pre class="error-text" v-html="highlightedError"></pre>
        </div>

        <!-- 输入块 -->
        <div v-else-if="block.type === 'input'" class="input-content">
          <span class="input-text">{{ displayContent }}</span>
        </div>
      </div>
    </Transition>

    <!-- 块操作栏 -->
    <div class="block-actions" v-if="!block.collapsed">
      <button class="action-btn" @click.stop="$emit('copy', block.id)" title="复制">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="9" y="9" width="13" height="13" rx="2"/>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
        </svg>
      </button>
      <button
        v-if="block.type === 'command'"
        class="action-btn"
        @click.stop="$emit('re-execute', block.id)"
        title="重新执行"
      >
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M23 4v6h-6"/>
          <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
        </svg>
      </button>
      <button class="action-btn delete" @click.stop="$emit('remove', block.id)" title="删除">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6"/>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { highlightTerminalOutput, detectLanguage, highlightCode } from '../utils/codeHighlight'

const props = defineProps({
  block: { type: Object, required: true },
  plainText: { type: Boolean, default: false }
})

const emit = defineEmits(['toggle-collapse', 'copy', 'remove', 're-execute'])

// 去除 ANSI 转义序列
function stripAnsi(s) {
  if (!s) return s
  return s.replace(/\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g, '')
}

// 显示内容（plainText=true 时剥离 ANSI）
const displayContent = computed(() => {
  if (!props.plainText) return props.block.content
  return stripAnsi(props.block.content)
})

const displayError = computed(() => {
  if (!props.plainText || !props.block.errorOutput) return props.block.errorOutput
  return stripAnsi(props.block.errorOutput)
})

// 高亮内容（用于 HTML 渲染）
const highlightedContent = computed(() => {
  const text = displayContent.value
  if (!text) return ''

  // 检测是否为代码块
  const lang = detectLanguage(text)
  if (lang) {
    return highlightCode(text, lang)
  }

  // 终端输出高亮
  return highlightTerminalOutput(text)
})

const highlightedError = computed(() => {
  const text = displayError.value
  if (!text) return ''
  return highlightTerminalOutput(text)
})

// 类型标签
const typeLabel = computed(() => {
  const labels = {
    command: '命令',
    output: '输出',
    error: '错误',
    system: '系统',
    input: '输入'
  }
  return labels[props.block.type] || '未知'
})

// 是否有内容
const hasContent = computed(() => {
  const block = props.block
  if (block.type === 'command') {
    return block.content || block.errorOutput || block.status === 'running'
  }
  return block.content
})

// 切换折叠
function toggleCollapse() {
  emit('toggle-collapse', props.block.id)
}

// 格式化时间
function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 格式化耗时
function formatDuration(ms) {
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`
}
</script>

<style scoped>
.terminal-block {
  background: var(--bg-panel-solid);
  border: 1px solid var(--border-default);
  border-radius: 0;
  overflow: hidden;
  transition: border-color .15s;
}

.terminal-block:hover { border-color: var(--border-default); }

/* 头部（含左侧竖线） */
.block-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  user-select: none;
  background: var(--bg-toolbar);
  border-bottom: 1px solid var(--border-default);
  border-left: 3px solid var(--surface-hover);
}

.command .block-header { border-left-color: var(--accent-success); }
.error .block-header { border-left-color: var(--accent-danger); }
.system .block-header { border-left-color: var(--text-muted); }
.input .block-header { border-left-color: var(--accent-warning); }
.running .block-header { border-left-color: var(--accent-warning); animation: pulse 1s infinite; }
.failed .block-header { border-left-color: var(--accent-danger); }

.terminal-block.collapsed .block-header {
  border-bottom: none;
}

.block-icon { flex-shrink: 0; color: var(--text-muted); }
.command .block-icon { color: var(--accent-success); }
.error .block-icon { color: var(--accent-danger); }
.system .block-icon { color: var(--text-muted); }

.block-title { flex: 1; min-width: 0; }

.command-text {
  font-family: 'Cascadia Code', monospace;
  font-size: 13px;
  color: var(--text-primary);
  word-break: break-all;
}

.system-text { font-size: 12px; color: var(--text-muted); }
.type-label { font-size: 12px; color: var(--text-muted); }

.block-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 3px;
}

.status-badge.running { background: rgba(255,152,0,.12); color: var(--accent-warning); }
.status-badge.success { background: rgba(76,175,80,.12); color: var(--accent-success); }
.status-badge.failed { background: rgba(244,67,54,.12); color: var(--accent-danger); }

.spinner {
  width: 10px; height: 10px;
  border: 2px solid rgba(255,152,0,.3); border-top-color: var(--accent-warning);
  border-radius: 50%; animation: spin 1s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.4} }

.duration { font-size: 11px; color: var(--text-muted); font-variant-numeric: tabular-nums; }
.timestamp { font-size: 11px; color: var(--text-muted); }

.collapse-btn {
  display: flex; align-items: center; justify-content: center;
  width: 20px; height: 20px; background: transparent; border: none;
  border-radius: 3px; color: var(--text-muted); cursor: pointer;
}

.collapse-btn:hover { background: var(--surface-hover); color: var(--text-secondary); }
.collapse-btn svg { transition: transform .2s; }
.collapse-btn svg.rotated { transform: rotate(0deg); }
.collapse-btn svg:not(.rotated) { transform: rotate(-90deg); }

/* 内容区（无左侧竖线） */
.block-content {
  padding: 10px 12px;
}

.output-content, .error-content, .input-content {
  max-height: 400px;
  overflow-y: auto;
}

.output-text, .error-text {
  margin: 0;
  font-family: 'Cascadia Code', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.output-text { color: var(--text-primary); }
.error-text { color: var(--accent-danger); }
.input-text { font-family: 'Cascadia Code', monospace; font-size: 12px; color: var(--accent-warning); }

.waiting-output {
  display: flex; align-items: center; justify-content: center; padding: 20px;
}

.dot-pulse { display: flex; gap: 4px; }
.dot-pulse::before, .dot-pulse::after, .dot-pulse {
  width: 6px; height: 6px; border-radius: 50%; background: var(--text-muted);
  animation: dotPulse 1.4s infinite;
}
.dot-pulse::before { content: ''; animation-delay: 0s; }
.dot-pulse::after { content: ''; animation-delay: .2s; }

@keyframes dotPulse { 0%,80%,100%{opacity:.3;transform:scale(.8)} 40%{opacity:1;transform:scale(1)} }

/* 操作栏 */
.block-actions {
  display: flex; gap: 4px; padding: 4px 12px 8px;
  opacity: 0; transition: opacity .15s;
}

.terminal-block:hover .block-actions { opacity: 1; }

.action-btn {
  display: flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; background: transparent; border: none;
  border-radius: 3px; color: var(--text-muted); cursor: pointer;
}

.action-btn:hover { background: var(--surface-hover); color: var(--text-secondary); }
.action-btn.delete:hover { background: rgba(244,67,54,.15); color: var(--accent-danger); }

/* 折叠动画 */
.content-enter-active { transition: all .2s ease; }
.content-leave-active { transition: all .15s ease; }
.content-enter-from, .content-leave-to { opacity: 0; max-height: 0; }

/* 滚动条 */
.output-content::-webkit-scrollbar, .error-content::-webkit-scrollbar { width: 4px; }
.output-content::-webkit-scrollbar-track, .error-content::-webkit-scrollbar-track { background: transparent; }
.output-content::-webkit-scrollbar-thumb, .error-content::-webkit-scrollbar-thumb { background: var(--surface-hover); border-radius: 2px; }

/* 代码高亮 */
.output-text :deep(.hljs-keyword),
.output-text :deep(.hljs-selector-tag),
.output-text :deep(.hljs-built_in) { color: #c678dd; }
.output-text :deep(.hljs-string),
.output-text :deep(.hljs-attr) { color: #98c379; }
.output-text :deep(.hljs-number),
.output-text :deep(.hljs-literal) { color: #d19a66; }
.output-text :deep(.hljs-comment) { color: #5c6370; font-style: italic; }
.output-text :deep(.hljs-function),
.output-text :deep(.hljs-title) { color: #61afef; }
.output-text :deep(.hljs-variable),
.output-text :deep(.hljs-params) { color: #e06c75; }
.output-text :deep(.hljs-type),
.output-text :deep(.hljs-class) { color: #e5c07b; }

/* 终端输出高亮 */
.output-text :deep(.hl-path) { color: #61afef; }
.output-text :deep(.hl-ip) { color: #c678dd; }
.output-text :deep(.hl-number) { color: #d19a66; }
.output-text :deep(.hl-unit) { color: #5c6370; }
.output-text :deep(.hl-success) { color: var(--accent-success); font-weight: 600; }
.output-text :deep(.hl-error) { color: var(--accent-danger); font-weight: 600; }
.output-text :deep(.hl-warning) { color: var(--accent-warning); font-weight: 600; }
.output-text :deep(.hl-perms) { color: #c678dd; font-family: monospace; }
</style>
