<template>
  <Transition name="fade">
    <div v-if="visible" class="dialog-mask" @mousedown="close">
      <div class="dialog" @mousedown.stop>
        <!-- 标题栏 -->
        <div class="dialog-header">
          <h3 class="dialog-title">会话管理</h3>
          <button class="close-btn" @click="close">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <!-- 搜索栏 -->
        <div class="search-bar">
          <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="searchKeyword"
            class="search-input"
            placeholder="搜索会话..."
          />
        </div>

        <!-- 会话列表 -->
        <div class="session-list">
          <div
            v-for="session in filteredSessions"
            :key="session.id"
            class="session-item"
            :class="{ active: session.id === currentSessionId }"
            @click="switchSession(session)"
          >
            <!-- 状态指示器 -->
            <div class="session-status" :class="session.status">
              <span class="status-dot"></span>
            </div>

            <!-- 会话信息 -->
            <div class="session-info">
              <div class="session-name">
                <span v-if="session.type === 'ai'" class="ai-badge">AI</span>
                {{ session.label || session.id }}
              </div>
              <div class="session-meta">
                <span class="meta-item">{{ formatTime(session.createdAt) }}</span>
                <span class="meta-item">{{ session.commandCount }} 命令</span>
                <span v-if="session.isRecording" class="recording-badge">录制中</span>
              </div>
            </div>

            <!-- 标签 -->
            <div class="session-tags" v-if="session.tags.length > 0">
              <span v-for="tag in session.tags" :key="tag" class="tag">{{ tag }}</span>
            </div>

            <!-- 操作 -->
            <div class="session-actions">
              <button
                v-if="session.status === 'disconnected'"
                class="action-btn reconnect"
                @click.stop="reconnectSession(session)"
                title="重连"
              >
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M23 4v6h-6"/>
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                </svg>
              </button>
              <button
                class="action-btn close"
                @click.stop="closeSession(session)"
                title="关闭"
              >
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-if="filteredSessions.length === 0" class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
              <rect x="3" y="3" width="7" height="7"/>
              <rect x="14" y="3" width="7" height="7"/>
              <rect x="3" y="14" width="7" height="7"/>
              <rect x="14" y="14" width="7" height="7"/>
            </svg>
            <p>{{ searchKeyword ? '未找到匹配的会话' : '暂无会话' }}</p>
          </div>
        </div>

        <!-- 底部统计 -->
        <div class="dialog-footer">
          <div class="stats">
            <span class="stat-item">共 {{ stats.total }} 个会话</span>
            <span class="stat-item active">{{ stats.active }} 个活跃</span>
            <span class="stat-item" v-if="stats.recording > 0">{{ stats.recording }} 个录制中</span>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getSessionManager } from '../utils/sessionManager'

const props = defineProps({
  visible: { type: Boolean, default: false },
  connId: { type: String, default: '' }
})

const emit = defineEmits(['update:visible', 'switch-session'])

const sessionManager = getSessionManager()
const searchKeyword = ref('')
const currentSessionId = ref(null)
const sessions = ref([])

// 监听会话变化
let cleanupListener = null

onMounted(() => {
  cleanupListener = sessionManager.addListener(() => {
    sessions.value = sessionManager.getAllSessions()
  })
  sessions.value = sessionManager.getAllSessions()
})

onUnmounted(() => {
  if (cleanupListener) cleanupListener()
})

// 过滤会话
const filteredSessions = computed(() => {
  if (!searchKeyword.value) return sessions.value
  return sessionManager.searchSessions(searchKeyword.value)
})

// 统计信息
const stats = computed(() => sessionManager.getStats())

// 切换会话
function switchSession(session) {
  currentSessionId.value = session.id
  emit('switch-session', session.id)
  close()
}

// 重连会话
function reconnectSession(session) {
  // TODO: 实现重连
  console.log('[SessionManager] 重连会话:', session.id)
}

// 关闭会话
function closeSession(session) {
  sessionManager.removeSession(session.id)
}

// 关闭对话框
function close() {
  emit('update:visible', false)
}

// 格式化时间
function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.dialog-mask {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.dialog {
  width: 560px;
  max-height: 70vh;
  background: #1a1b26;
  border: 1px solid var(--surface-hover);
  border-radius: 12px;
  box-shadow: 0 16px 48px var(--shadow-lg);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.dialog-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s;
}

.close-btn:hover {
  background: var(--border-default);
  color: var(--text-secondary);
}

.search-bar {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.search-icon {
  color: var(--text-muted);
  margin-right: 10px;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}

.session-item:hover {
  background: var(--surface-1);
}

.session-item.active {
  background: var(--primary-bg);
  border: 1px solid var(--primary-bg);
}

.session-status {
  flex-shrink: 0;
}

.status-dot {
  display: block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.session-status.active .status-dot {
  background: var(--accent-success);
  box-shadow: 0 0 6px var(--success-bg);
}

.session-status.disconnected .status-dot {
  background: var(--accent-danger);
}

.session-status.starting .status-dot {
  background: var(--accent-warning);
  animation: pulse 1.5s infinite;
}

.session-status.idle .status-dot {
  background: var(--text-muted);
}

.session-status.error .status-dot {
  background: var(--accent-danger);
}

.session-info {
  flex: 1;
  min-width: 0;
}

.session-name {
  font-size: 13px;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.ai-badge {
  font-size: 9px;
  padding: 1px 4px;
  background: var(--accent-purple);
  color: var(--accent-purple);
  border-radius: 3px;
  font-weight: 600;
}

.session-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.meta-item {
  font-size: 11px;
  color: var(--text-muted);
}

.recording-badge {
  font-size: 10px;
  padding: 1px 4px;
  background: var(--danger-bg);
  color: var(--accent-danger);
  border-radius: 3px;
}

.session-tags {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.tag {
  font-size: 10px;
  padding: 2px 6px;
  background: var(--surface-1);
  color: var(--text-secondary);
  border-radius: 4px;
}

.session-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.15s;
}

.session-item:hover .session-actions {
  opacity: 1;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:hover {
  background: var(--border-default);
  color: var(--text-secondary);
}

.action-btn.reconnect:hover {
  background: var(--warning-bg, rgba(224, 175, 104, 0.15));
  color: var(--accent-warning);
}

.action-btn.close:hover {
  background: var(--danger-bg);
  color: var(--accent-danger);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: var(--text-muted);
}

.empty-state p {
  margin: 12px 0 0;
  font-size: 13px;
}

.dialog-footer {
  padding: 12px 20px;
  border-top: 1px solid var(--border-subtle);
}

.stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  font-size: 11px;
  color: var(--text-muted);
}

.stat-item.active {
  color: var(--accent-success);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
