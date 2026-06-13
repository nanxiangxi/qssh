<template>
  <div class="cloud-panel">
    <div class="page-header">
      <div class="header-top">
        <button class="back-btn" @click="goBack" title="返回">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 12H5M12 19l-7-7 7-7"/>
          </svg>
          <span>返回</span>
        </button>
      </div>
      <h1 class="page-title">私有云端</h1>
      <p class="page-description">搭建个人私有云端，实现跨设备数据同步和协同</p>
    </div>

    <!-- 服务状态 -->
    <div class="status-card" :class="{ running: status.running }">
      <div class="status-header">
        <div class="status-icon">
          <svg v-if="status.running" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" class="icon-success" stroke-width="2">
            <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
          </svg>
          <svg v-else width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" class="icon-muted" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
        </div>
        <div class="status-info">
          <div class="status-title">{{ status.running ? '服务运行中' : '服务未启动' }}</div>
          <div class="status-desc" v-if="status.running">
            端口: {{ status.port }} · 设备: {{ status.deviceName }}
          </div>
        </div>
        <button class="toggle-btn" :class="{ active: status.running }" @click="toggleService">
          {{ status.running ? '停止' : '启动' }}
        </button>
      </div>
    </div>

    <!-- 配置 -->
    <div class="config-section">
      <div class="section-title">配置</div>
      <div class="config-list">
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">监听端口</div>
            <div class="config-desc">云端服务监听的端口号</div>
          </div>
          <div class="config-control">
            <input type="number" v-model.number="config.port" class="number-input" min="1024" max="65535">
          </div>
        </div>

        <div class="config-item" style="margin-top:12px;">
          <div class="config-info">
            <div class="config-label">认证令牌</div>
            <div class="config-desc">用于设备间认证的令牌</div>
          </div>
          <div class="config-control">
            <div class="token-display">
              <span class="token-text">{{ config.token }}</span>
              <button class="copy-btn" @click="copyToken" title="复制">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                </svg>
              </button>
            </div>
          </div>
        </div>

        <div class="config-item" style="margin-top:12px;">
          <div class="config-info">
            <div class="config-label">设备 ID</div>
            <div class="config-desc">本设备的唯一标识</div>
          </div>
          <div class="config-control">
            <span class="device-id">{{ config.deviceId }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 已连接设备 -->
    <div class="devices-section">
      <div class="section-title">已连接设备 ({{ devices.length }})</div>
      <div v-if="devices.length === 0" class="empty-devices">
        暂无已连接设备
      </div>
      <div v-else class="device-list">
        <div v-for="device in devices" :key="device.id" class="device-card">
          <div class="device-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
              <line x1="8" y1="21" x2="16" y2="21"/>
              <line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
          </div>
          <div class="device-info">
            <div class="device-name">{{ device.name }}</div>
            <div class="device-meta">{{ device.host }} · {{ device.os }}</div>
          </div>
          <div class="device-status" :class="device.status">
            {{ device.status === 'online' ? '在线' : '离线' }}
          </div>
        </div>
      </div>
    </div>

    <!-- 消息提示 -->
    <Message ref="messageRef" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Message from '../../../components/Message.vue'

const router = useRouter()
const messageRef = ref(null)

const status = ref({
  running: false,
  port: 9527,
  deviceName: '',
  peers: 0
})

const config = ref({
  enabled: false,
  port: 9527,
  token: '',
  deviceId: ''
})

const devices = ref([])

function goBack() {
  router.back()
}

async function toggleService() {
  try {
    if (status.value.running) {
      // TODO: 调用后端停止服务
      status.value.running = false
      messageRef.value?.success('云端服务已停止')
    } else {
      // TODO: 调用后端启动服务
      status.value.running = true
      messageRef.value?.success('云端服务已启动')
    }
  } catch (e) {
    messageRef.value?.error('操作失败: ' + (e.message || e))
  }
}

function copyToken() {
  navigator.clipboard.writeText(config.value.token)
  messageRef.value?.success('令牌已复制')
}

onMounted(async () => {
  // TODO: 从后端加载配置和状态
  config.value.token = 'demo-token-' + Math.random().toString(36).slice(2, 10)
  config.value.deviceId = 'device-' + Math.random().toString(36).slice(2, 10)
})
</script>

<style scoped>
.cloud-panel {
  height: 100%;
  padding: 24px;
  overflow-y: auto;
}

.page-header { margin-bottom: 24px; }
.header-top { margin-bottom: 16px; }

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--surface-1);
  border: 1px solid var(--surface-hover);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.back-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.page-title {
  margin: 0 0 8px;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

.page-description {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted);
}

.status-card {
  padding: 16px;
  background: var(--surface-1, rgba(255, 255, 255, 0.02));
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  margin-bottom: 24px;
}

.status-card.running {
  border-color: var(--border-success, rgba(76, 175, 80, 0.3));
  background: var(--success-bg, rgba(76, 175, 80, 0.05));
}

.status-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-1);
  border-radius: 8px;
}

.status-info { flex: 1; }

.status-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.status-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.toggle-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
  background: var(--success-bg, rgba(76, 175, 80, 0.2));
  color: var(--accent-success);
}

.toggle-btn:hover {
  background: var(--success-bg, rgba(76, 175, 80, 0.3));
}

.toggle-btn.active {
  background: var(--danger-bg, rgba(244, 67, 54, 0.2));
  color: var(--accent-danger);
}

.toggle-btn.active:hover {
  background: var(--danger-bg, rgba(244, 67, 54, 0.3));
}

.config-section, .devices-section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-subtle);
}

.config-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.config-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: var(--surface-1, rgba(255, 255, 255, 0.02));
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
}

.config-info {
  flex: 1;
  min-width: 0;
  margin-right: 24px;
}

.config-label {
  font-size: 14px;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.config-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
}

.config-control {
  flex-shrink: 0;
}

.number-input {
  width: 100px;
  padding: 6px 12px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 13px;
}

.number-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.token-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 6px;
}

.token-text {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-secondary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.copy-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 4px;
}

.copy-btn:hover {
  background: var(--surface-hover);
  color: var(--text-secondary);
}

.device-id {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-muted);
}

.empty-devices {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.device-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.device-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--surface-1, rgba(255, 255, 255, 0.02));
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
}

.device-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-1);
  border-radius: 6px;
  color: var(--text-secondary);
}

.device-info { flex: 1; }

.device-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.device-meta {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}

.device-status {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}

.device-status.online {
  background: var(--success-bg, rgba(76, 175, 80, 0.15));
  color: var(--accent-success);
}

.device-status.offline {
  background: var(--surface-1);
  color: var(--text-muted);
}

.icon-success {
  color: var(--accent-success, #4caf50);
}

.icon-muted {
  color: var(--text-muted, #666);
}
</style>
