<template>
  <div class="terminal-toolbar">
    <!-- 左侧：会话信息 -->
    <div class="toolbar-left">
      <!-- 连接状态指示器 -->
      <div class="status-indicator" :class="connectionStatus">
        <span class="status-dot"></span>
        <span class="status-text">{{ statusText }}</span>
      </div>

      <!-- 会话标签 -->
      <div class="session-label" v-if="session?.label">
        {{ session.label }}
      </div>

      <!-- 命令计数 -->
      <div class="command-count" v-if="session?.commandCount > 0">
        {{ session.commandCount }} 命令
      </div>
    </div>

    <!-- 中间：操作按钮 -->
    <div class="toolbar-center">
      <!-- 录制按钮 -->
      <button
        class="toolbar-btn"
        :class="{ recording: isRecording, paused: isRecording && isPaused }"
        @click="toggleRecording"
        :title="isRecording ? '停止录制' : '开始录制'"
      >
        <svg v-if="!isRecording" width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
          <circle cx="12" cy="12" r="8"/>
        </svg>
        <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
          <rect x="6" y="6" width="12" height="12" rx="2"/>
        </svg>
        <span class="btn-text">{{ isRecording ? '停止' : '录制' }}</span>
      </button>

      <!-- 录制时长 -->
      <div class="recording-duration" v-if="isRecording">
        <span class="recording-dot"></span>
        {{ recordingDuration }}
      </div>

      <!-- 保存录制 -->
      <button
        v-if="isRecording || hasRecording"
        class="toolbar-btn"
        @click="showSaveMenu = !showSaveMenu"
        title="保存录制"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
      </button>

      <!-- 保存菜单 -->
      <div class="save-menu" v-if="showSaveMenu">
        <div class="save-option" @click="saveRecording('text')">保存为文本</div>
        <div class="save-option" @click="saveRecording('html')">保存为 HTML</div>
        <div class="save-option" @click="saveRecording('json')">保存为 JSON</div>
      </div>
    </div>

    <!-- 右侧：功能按钮 -->
    <div class="toolbar-right">
      <!-- 搜索 -->
      <button class="toolbar-btn" @click="$emit('search')" title="搜索">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
      </button>

      <!-- 命令历史 -->
      <button class="toolbar-btn" @click="$emit('open-command-history')" title="命令历史">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <polyline points="12 6 12 12 16 14"/>
        </svg>
      </button>

      <!-- 会话管理 -->
      <button class="toolbar-btn" @click="$emit('open-session-manager')" title="会话管理">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="3" width="7" height="7"/>
          <rect x="14" y="3" width="7" height="7"/>
          <rect x="3" y="14" width="7" height="7"/>
          <rect x="14" y="14" width="7" height="7"/>
        </svg>
      </button>

      <!-- 重连按钮（断线时显示） -->
      <button
        v-if="connectionStatus === 'disconnected'"
        class="toolbar-btn reconnect"
        @click="$emit('reconnect')"
        title="重新连接"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M23 4v6h-6"/>
          <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
        </svg>
        <span class="btn-text">重连</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  session: { type: Object, default: null },
  isRecording: { type: Boolean, default: false },
  recordingDuration: { type: String, default: '00:00' },
  connectionStatus: { type: String, default: 'idle' },
  mode: { type: String, default: 'auto' },
  isInteractive: { type: Boolean, default: false }
})

const emit = defineEmits([
  'toggle-recording',
  'save-recording',
  'open-session-manager',
  'open-command-history',
  'search',
  'reconnect',
  'toggle-mode'
])

const showSaveMenu = ref(false)
const isPaused = ref(false)
const hasRecording = ref(false)

// 状态文本
const statusText = computed(() => {
  const statusMap = {
    idle: '空闲',
    starting: '启动中',
    active: '已连接',
    disconnected: '已断开',
    error: '错误'
  }
  return statusMap[props.connectionStatus] || '未知'
})

// 切换录制
function toggleRecording() {
  emit('toggle-recording')
}

// 保存录制
function saveRecording(format) {
  emit('save-recording', format)
  showSaveMenu.value = false
}
</script>

<style scoped>
.terminal-toolbar {
  height: 32px; display: flex; align-items: center;
  padding: 0 10px; background: #222; border-bottom: 1px solid var(--surface-hover); flex-shrink: 0;
}
.toolbar-left, .toolbar-center, .toolbar-right { display: flex; align-items: center; gap: 6px; }
.toolbar-left { flex: 1; min-width: 0; }
.toolbar-center { flex-shrink: 0; }
.toolbar-right { flex: 1; min-width: 0; justify-content: flex-end; }
.status-indicator { display: flex; align-items: center; gap: 4px; padding: 2px 6px; border-radius: 3px; background: var(--border-default); }
.status-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.status-indicator.active .status-dot { background: var(--accent-success); box-shadow: 0 0 5px rgba(76,175,80,.4); }
.status-indicator.disconnected .status-dot { background: var(--accent-danger); }
.status-indicator.starting .status-dot { background: var(--accent-warning); animation: pulse 1.5s infinite; }
.status-indicator.error .status-dot { background: var(--accent-danger); }
.status-indicator.idle .status-dot { background: var(--text-muted, #666); }
.status-text { font-size: 12px; color: var(--text-primary, #ccc); }
.session-label { font-size: 11px; color: var(--text-muted, #888); padding: 1px 5px; background: var(--border-default); border-radius: 3px; }
.command-count { font-size: 11px; color: var(--text-muted, #666); }
.toolbar-btn {
  display: flex; align-items: center; gap: 3px; padding: 3px 6px;
  background: transparent; border: none; border-radius: 3px;
  color: var(--text-secondary, #aaa); cursor: pointer; font-size: 12px;
}
.toolbar-btn:hover { background: var(--surface-hover); color: var(--text-primary, #ddd); }
.toolbar-btn.recording { background: var(--danger-bg, rgba(244,67,54,.12)); color: var(--accent-danger); }
.toolbar-btn.reconnect { background: var(--warning-bg, rgba(255,152,0,.12)); color: var(--accent-warning); }
.btn-text { font-size: 12px; }
.recording-duration { display: flex; align-items: center; gap: 4px; font-size: 12px; color: var(--accent-danger); }
.recording-dot { width: 5px; height: 5px; border-radius: 50%; background: var(--accent-danger); animation: pulse 1s infinite; }
.save-menu {
  position: absolute; top: 100%; left: 50%; transform: translateX(-50%); margin-top: 2px;
  background: var(--bg-panel-solid); border: 1px solid var(--surface-hover); border-radius: 4px;
  padding: 2px; min-width: 120px; box-shadow: 0 8px 20px var(--shadow-lg); z-index: 100;
}
.save-option { padding: 4px 10px; font-size: 12px; color: var(--text-secondary, #aaa); cursor: pointer; border-radius: 3px; }
.save-option:hover { background: var(--surface-hover); color: var(--text-primary, #ddd); }
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.4} }
</style>
