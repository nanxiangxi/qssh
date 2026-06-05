<template>
  <footer class="bottom-bar">
    <div class="bottom-section left">
      <div class="command-history">
        <span class="label">最近命令:</span>
        <span class="command">{{ lastCommand }}</span>
      </div>
    </div>
    <div class="bottom-section center">
      <div class="connection-info">
        <span class="info-item" :class="connectionStatus">
          {{ connectionStatus === 'connected' ? '🟢 已连接' : '⚪ 未连接' }}
        </span>
        <span class="info-item">延迟: {{ latency }}ms</span>
      </div>
    </div>
    <div class="bottom-section right">
      <button class="action-btn" title="保存连接" @click="saveConnection">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/>
          <polyline points="17 21 17 13 7 13 7 21"/>
          <polyline points="7 3 7 8 15 8"/>
        </svg>
      </button>
      <button class="action-btn" title="断开连接" @click="disconnect">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
          <polyline points="16 17 21 12 16 7"/>
          <line x1="21" y1="12" x2="9" y2="12"/>
        </svg>
      </button>
    </div>
  </footer>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useSSHLayoutStore } from '../../../stores/sshLayout'
import { useCommandHistoryStore } from '../../../stores/commandHistory'

import { showMessage } from '../../../utils/message'
import { useConfirm } from '../../../utils/confirm'
import {SSHService} from "../../../../bindings/changeme/ssh/index.js";

const { confirm } = useConfirm()

const route = useRoute()
const sshLayoutStore = useSSHLayoutStore()
const commandHistoryStore = useCommandHistoryStore()

// 状态数据
const lastCommand = ref('无')
const connectionStatus = ref('disconnected')
const latency = ref(0)

// 获取当前连接ID
const currentConnId = computed(() => sshLayoutStore.currentConnectionId)

// 定时更新状态
let updateInterval = null

// 更新连接状态
const updateConnectionStatus = async () => {
  if (!currentConnId.value || currentConnId.value === 'default-connection') {
    connectionStatus.value = 'disconnected'
    latency.value = 0
    lastCommand.value = ''
    return
  }

  try {
    // 检查连接是否活跃
    const isConnected = await SSHService.IsConnected(currentConnId.value)
    connectionStatus.value = isConnected ? 'connected' : 'disconnected'
    
    if (isConnected) {
      // 获取实时延迟（后端会执行 ping 测量）
      const lat = await SSHService.GetLatency(currentConnId.value)
      latency.value = lat || 0
      
      // 从前端 Store 获取最近命令
      const history = commandHistoryStore.getHistory(currentConnId.value)
      lastCommand.value = history.length > 0 ? history[history.length - 1].content : '无'
    } else {
      latency.value = 0
      lastCommand.value = '无'
    }
  } catch (error) {
    console.error('[BottomStatusBar] 更新状态失败:', error)
  }
}

// 保存连接（仅保存当前标签的连接配置）
const saveConnection = async () => {
  if (!currentConnId.value || currentConnId.value === 'default-connection') {
    showMessage('没有可保存的连接', 'warning')
    return
  }

  try {
    const connInfo = await SSHService.GetConnection(currentConnId.value)
    if (!connInfo) {
      showMessage('连接不存在', 'error')
      return
    }

    const authType = connInfo.privateKey ? '密钥认证' : '密码认证'
    const ok = await confirm({
      title: '保存连接',
      message: `保存当前连接到本地？\n\n名称: ${connInfo.name}\n地址: ${connInfo.host}:${connInfo.port}\n用户: ${connInfo.username}\n认证: ${authType}`
    })
    if (!ok) return

    await SSHService.SaveConnection(currentConnId.value)
    showMessage('连接已保存', 'success')
  } catch (error) {
    console.error('[BottomStatusBar] 保存连接失败:', error)
    showMessage('保存连接失败: ' + error.message, 'error')
  }
}

// 断开连接
const disconnect = async () => {
  if (!currentConnId.value || currentConnId.value === 'default-connection') {
    showMessage('没有可断开的连接', 'warning')
    return
  }

  const ok = await confirm({
    title: '断开连接',
    message: '确定断开当前 SSH 连接？',
    danger: true
  })
  if (!ok) return

  try {
    await SSHService.Disconnect(currentConnId.value)
    showMessage('连接已断开', 'success')

    connectionStatus.value = 'disconnected'
    latency.value = 0
    lastCommand.value = '无'
  } catch (error) {
    console.error('[BottomStatusBar] 断开连接失败:', error)
    showMessage('断开连接失败: ' + error.message, 'error')
  }
}

// 组件挂载时开始定时更新
onMounted(() => {
  updateConnectionStatus() // 立即更新一次
  updateInterval = setInterval(updateConnectionStatus, 3000) // 每3秒更新一次
})

// 组件卸载时清理定时器
onUnmounted(() => {
  if (updateInterval) {
    clearInterval(updateInterval)
  }
})
</script>

<style scoped>
.bottom-bar {
  height: 2.5rem;
  background: rgba(40, 40, 40, 0.95);
  border-top: 0.0625rem solid rgba(255, 255, 255, 0.1);
  padding: 0 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.bottom-section {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.bottom-section.left {
  flex: 1;
  min-width: 0;
}

.bottom-section.center {
  flex-shrink: 0;
}

.bottom-section.right {
  flex-shrink: 0;
}

.command-history {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  overflow: hidden;
}

.command-history .label {
  color: #a0aec0;
  font-size: 0.75rem;
  white-space: nowrap;
}

.command-history .command {
  color: #e2e8f0;
  font-size: 0.75rem;
  font-family: 'Courier New', monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.connection-info {
  display: flex;
  gap: 1rem;
}

.info-item {
  color: #a0aec0;
  font-size: 0.75rem;
  white-space: nowrap;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  background: transparent;
  border: none;
  border-radius: 0.25rem;
  color: #a0aec0;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}
</style>
