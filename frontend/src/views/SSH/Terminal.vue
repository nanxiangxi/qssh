<template>
  <div class="terminal-container">
    <!-- 加载中状态 -->
    <div v-if="loading" class="terminal-placeholder">
      <div class="placeholder-content">
        <div class="loading-spinner"></div>
        <h3>正在加载连接...</h3>
      </div>
    </div>
    
    <!-- 无标签状态 -->
    <div v-else-if="tabsStore.tabs.length === 0" class="terminal-placeholder">
      <div class="placeholder-content">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="icon">
          <polyline points="4 17 10 11 4 5"/>
          <line x1="12" y1="19" x2="20" y2="19"/>
        </svg>
        <h3>没有打开的连接</h3>
        <p>点击顶部标签或新建连接开始使用</p>
      </div>
    </div>
    
    <!-- 有标签但未激活 -->
    <div v-else-if="!tabsStore.activeTabId" class="terminal-placeholder">
      <div class="placeholder-content">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="icon">
          <polyline points="4 17 10 11 4 5"/>
          <line x1="12" y1="19" x2="20" y2="19"/>
        </svg>
        <h3>SSH 终端</h3>
        <p>选择一个连接开始使用终端</p>
      </div>
    </div>
    
    <!-- 连接成功状态 - TODO: 后续集成真实的终端模拟器 -->
    <div v-else class="terminal-ready">
      <div class="terminal-header">
        <span class="connection-badge">🟢 已连接</span>
        <span class="connection-name">{{ tabsStore.activeTab?.name }}</span>
      </div>
      <div class="terminal-body">
        <p>终端模拟器即将集成...</p>
        <p class="hint">提示：后续将集成 xterm.js 实现真正的SSH终端</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'

import { useSSHTabsStore } from '../../stores/sshTabs'
import {SSHService} from "../../../bindings/changeme/ssh/index.js"
import {Events} from '@wailsio/runtime'

const route = useRoute()
const tabsStore = useSSHTabsStore()

const groupID = ref('')
const groupName = ref('')
const loading = ref(false)

console.log('[Terminal] 🚀 Terminal 组件已挂载')
console.log('[Terminal] 📍 当前路由:', route.path, route.query)

// 加载分组的连接
const loadGroupConnections = async () => {
  if (!groupID.value) return
  
  loading.value = true
  console.log('[Terminal] 🔄 加载分组连接:', groupID.value)
  
  try {
    // 调用后端获取该分组的完整连接信息（包含名称）
    const connections = await SSHService.GetGroupConnectionInfos(groupID.value)
    console.log('[Terminal] ✅ 获取到连接列表:', connections)
    
    // 构建新的标签列表，使用连接名称
    const newTabs = connections.map(conn => ({
      id: conn.id,
      name: conn.name || `连接 ${conn.id.substring(0, 8)}...`,
      status: conn.status || 'connected',
      active: false
    }))
    
    // 使用 store 的方法刷新标签（保持响应式）
    tabsStore.refreshTabs(newTabs)
    
    console.log('[Terminal] ✅ 加载完成，共', tabsStore.tabs.length, '个标签')
  } catch (error) {
    console.error('[Terminal] ❌ 加载失败:', error)
  } finally {
    loading.value = false
  }
}

// 监听分组更新事件
const handleGroupUpdate = (event) => {
  const { groupID: updatedGroupID, action, connections } = event.data
  
  // 只处理当前窗口的分组
  if (updatedGroupID === groupID.value && connections && Array.isArray(connections)) {
    // 比较连接列表是否变化
    const currentConnIDs = tabsStore.tabs.map(tab => tab.id)
    const newConnIDs = connections.map(conn => conn.id)
    
    // 如果连接数量相同且所有ID都相同，则不更新
    if (currentConnIDs.length === newConnIDs.length && 
        currentConnIDs.every((id, index) => id === newConnIDs[index])) {
      // 数据没有变化，跳过更新
      return
    }
    
    console.log('[Terminal] 📨 收到分组状态同步:', connections.length, '个连接（已变化）')
    refreshTabsFromConnections(connections)
  }
}

// 从连接信息列表刷新标签
const refreshTabsFromConnections = (connections) => {
  // 构建新的标签列表，使用连接名称
  const newTabs = connections.map(conn => ({
    id: conn.id,
    name: conn.name || `连接 ${conn.id.substring(0, 8)}...`,  // 优先使用名称
    status: conn.status || 'connected',
    active: false
  }))
  
  // 使用 store 的方法刷新标签（保持响应式）
  tabsStore.refreshTabs(newTabs)
  
  console.log('[Terminal] ✅ 标签已更新，共', tabsStore.tabs.length, '个标签')
}

// 从路由参数获取 groupID
if (route.query.group) {
  groupID.value = route.query.group
  groupName.value = route.query.name || 'SSH终端'
  
  console.log('[Terminal] 🔑 从路由参数获取 groupID:', groupID.value)
  
  // 在下一个tick加载连接，确保组件完全初始化
  setTimeout(() => {
    loadGroupConnections()
  }, 0)
} else {
  console.log('[Terminal] ⚠️ 没有 group 参数')
}

onMounted(() => {
  console.log('[Terminal] 📦 onMounted 钩子执行')
  
  // 注册事件监听器
  Events.On('ssh:group-updated', handleGroupUpdate)
  console.log('[Terminal] 👂 已注册事件监听器: ssh:group-updated')
})

// 组件卸载时移除事件监听器并通知后端
onUnmounted(() => {
  Events.Off('ssh:group-updated', handleGroupUpdate)
  console.log('[Terminal] 🔇 已移除事件监听器: ssh:group-updated')
  
  // 通过事件通知后端窗口关闭
  if (groupID.value) {
    console.log('[Terminal] 📤 发送窗口关闭事件:', groupID.value)
    Events.Emit('ssh:window-closed', { groupID: groupID.value })
  }
})
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(30, 30, 30, 0.95);
}

.terminal-placeholder {
  text-align: center;
  color: #a0aec0;
}

.placeholder-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.icon {
  opacity: 0.3;
}

h3 {
  color: #e2e8f0;
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0;
}

p {
  font-size: 0.875rem;
  margin: 0;
}

/* 加载动画 */
.loading-spinner {
  width: 3rem;
  height: 3rem;
  border: 0.25rem solid rgba(66, 153, 225, 0.2);
  border-top-color: #4299e1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* 终端就绪状态 */
.terminal-ready {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.terminal-header {
  padding: 1rem 1.5rem;
  background: rgba(40, 40, 40, 0.9);
  border-bottom: 0.0625rem solid rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  gap: 1rem;
}

.connection-badge {
  color: #48bb78;
  font-size: 0.875rem;
  font-weight: 500;
}

.connection-name {
  color: #e2e8f0;
  font-size: 0.9375rem;
  font-weight: 600;
}

.terminal-body {
  flex: 1;
  padding: 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  color: #a0aec0;
}

.terminal-body p {
  font-size: 1rem;
}

.hint {
  font-size: 0.875rem !important;
  color: #718096;
  font-style: italic;
}
</style>
