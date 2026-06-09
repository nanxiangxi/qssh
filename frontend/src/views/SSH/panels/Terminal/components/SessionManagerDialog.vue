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
  background: rgba(0, 0, 0, 0.6);
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
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.dialog-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #c0caf5;
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
  color: #565f89;
  cursor: pointer;
  transition: all 0.15s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #a9b1d6;
}

.search-bar {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.search-icon {
  color: #565f89;
  margin-right: 10px;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: #c0caf5;
  font-size: 13px;
  font-family: inherit;
}

.search-input::placeholder {
  color: #565f89;
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
  background: rgba(255, 255, 255, 0.04);
}

.session-item.active {
  background: rgba(122, 162, 247, 0.1);
  border: 1px solid rgba(122, 162, 247, 0.2);
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
  background: #9ece6a;
  box-shadow: 0 0 6px rgba(158, 206, 106, 0.5);
}

.session-status.disconnected .status-dot {
  background: #f7768e;
}

.session-status.starting .status-dot {
  background: #e0af68;
  animation: pulse 1.5s infinite;
}

.session-status.idle .status-dot {
  background: #565f89;
}

.session-status.error .status-dot {
  background: #f7768e;
}

.session-info {
  flex: 1;
  min-width: 0;
}

.session-name {
  font-size: 13px;
  color: #c0caf5;
  display: flex;
  align-items: center;
  gap: 6px;
}

.ai-badge {
  font-size: 9px;
  padding: 1px 4px;
  background: rgba(187, 154, 247, 0.2);
  color: #bb9af7;
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
  color: #565f89;
}

.recording-badge {
  font-size: 10px;
  padding: 1px 4px;
  background: rgba(247, 118, 142, 0.15);
  color: #f7768e;
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
  background: rgba(255, 255, 255, 0.05);
  color: #a9b1d6;
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
  color: #565f89;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #a9b1d6;
}

.action-btn.reconnect:hover {
  background: rgba(224, 175, 104, 0.15);
  color: #e0af68;
}

.action-btn.close:hover {
  background: rgba(247, 118, 142, 0.15);
  color: #f7768e;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #565f89;
}

.empty-state p {
  margin: 12px 0 0;
  font-size: 13px;
}

.dialog-footer {
  padding: 12px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  font-size: 11px;
  color: #565f89;
}

.stat-item.active {
  color: #9ece6a;
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
