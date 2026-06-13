<template>
  <div class="terminal-status-bar">
    <!-- 左侧 -->
    <div class="status-left">
      <!-- 连接状态 -->
      <div class="status-item" :class="connectionStatus">
        <span class="status-icon">
          <svg v-if="connectionStatus === 'active'" width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
            <circle cx="12" cy="12" r="10"/>
          </svg>
          <svg v-else-if="connectionStatus === 'disconnected'" width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
            <circle cx="12" cy="12" r="10"/>
            <line x1="4" y1="4" x2="20" y2="20" stroke="currentColor" stroke-width="2"/>
          </svg>
          <svg v-else width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
            <circle cx="12" cy="12" r="10" opacity="0.5"/>
          </svg>
        </span>
        <span>{{ statusText }}</span>
      </div>

      <!-- 会话类型 -->
      <div class="status-item" v-if="sessionType === 'ai'">
        <span class="ai-badge">AI</span>
      </div>
    </div>

    <!-- 中间 -->
    <div class="status-center">
      <!-- 模式指示 -->
      <div class="status-item mode" :class="{ interactive: isInteractive }">
        <svg v-if="isInteractive" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
          <line x1="8" y1="21" x2="16" y2="21"/>
          <line x1="12" y1="17" x2="12" y2="21"/>
        </svg>
        <svg v-else-if="mode === 'structured'" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="3" width="7" height="7"/>
          <rect x="14" y="3" width="7" height="7"/>
          <rect x="3" y="14" width="7" height="7"/>
          <rect x="14" y="14" width="7" height="7"/>
        </svg>
        <svg v-else width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"/>
          <line x1="12" y1="19" x2="20" y2="19"/>
        </svg>
        <span>{{ isInteractive ? '交互式' : (mode === 'structured' ? '结构化' : '经典') }}</span>
      </div>

      <!-- 录制状态 -->
      <div class="status-item recording" v-if="isRecording">
        <span class="recording-dot"></span>
        <span>录制中</span>
      </div>
    </div>

    <!-- 右侧 -->
    <div class="status-right">
      <!-- 命令计数 -->
      <div class="status-item" v-if="commandCount > 0">
        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"/>
          <line x1="12" y1="19" x2="20" y2="19"/>
        </svg>
        <span>{{ commandCount }}</span>
      </div>

      <!-- 终端尺寸 -->
      <div class="status-item">
        <span>{{ cols }}×{{ rows }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  cols: { type: Number, default: 0 },
  rows: { type: Number, default: 0 },
  connectionStatus: { type: String, default: 'idle' },
  commandCount: { type: Number, default: 0 },
  isRecording: { type: Boolean, default: false },
  sessionType: { type: String, default: 'normal' },
  mode: { type: String, default: 'auto' },
  isInteractive: { type: Boolean, default: false }
})

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
</script>

<style scoped>
.terminal-status-bar {
  height: 22px; display: flex; align-items: center; justify-content: space-between;
  padding: 0 10px; background: var(--bg-toolbar); border-top: 1px solid var(--border-default); flex-shrink: 0;
}
.status-left, .status-center, .status-right { display: flex; align-items: center; gap: 10px; }
.status-left { flex: 1; min-width: 0; }
.status-center { flex-shrink: 0; }
.status-right { flex: 1; min-width: 0; justify-content: flex-end; }
.status-item { display: flex; align-items: center; gap: 3px; font-size: 11px; color: var(--text-disabled); }
.status-item.active { color: var(--accent-success); }
.status-item.disconnected { color: var(--accent-danger); }
.status-item.starting { color: var(--accent-warning); }
.status-item.error { color: var(--accent-danger); }
.status-item.recording { color: var(--accent-danger); }
.status-item.mode { color: var(--text-muted); }
.status-item.mode.interactive { color: var(--accent-warning); }
.status-icon { display: flex; align-items: center; }
.ai-badge { font-size: 9px; padding: 1px 3px; background: var(--surface-2); color: var(--text-secondary); border-radius: 2px; font-weight: 600; }
.recording-dot { width: 5px; height: 5px; border-radius: 50%; background: var(--accent-danger); animation: pulse 1s infinite; }
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.4} }
</style>
