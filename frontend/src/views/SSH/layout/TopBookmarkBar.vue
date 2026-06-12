<template>
  <header class="top-bar">
    <!-- 左侧：SSH 连接标签页 -->
    <div class="tab-container">
      <div
        v-for="tab in tabs"
        :key="tab.id"
        class="tab-item"
        :class="{ active: tab.id === activeTabId, disconnected: tab.status === 'disconnected' }"
        @click="openConnection(tab.id)"
        draggable="true"
        @dragstart="handleDragStart($event, tab)"
      >
        <span class="tab-status" :class="tab.status"></span>
        <span class="tab-name" :title="'ID: ' + tab.id">{{ getTabDisplayName(tab) }}</span>
        <!-- 重连按钮（仅断线时显示） -->
        <button
          v-if="tab.status === 'disconnected'"
          class="tab-reconnect"
          @click.stop="reconnectTab(tab.id)"
          title="重新连接"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
        </button>
        <button class="tab-close" @click.stop="closeTab(tab.id)">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
    </div>
    
    <!-- 右侧窗口控制按钮 -->
    <div class="window-controls">
      <button class="control-btn minimize" title="最小化" @click="Window.Minimise()">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <line x1="0" y1="6" x2="12" y2="6" stroke="currentColor" stroke-width="1.5"/>
        </svg>
      </button>
      <button class="control-btn maximize" :title="isMaximised ? '恢复' : '最大化'" @click="toggleMaximise">
        <svg v-if="!isMaximised" width="12" height="12" viewBox="0 0 12 12">
          <rect x="1" y="1" width="10" height="10" stroke="currentColor" stroke-width="1.5" fill="none"/>
        </svg>
        <svg v-else width="12" height="12" viewBox="0 0 12 12">
          <rect x="3" y="1" width="8" height="8" stroke="currentColor" stroke-width="1.5" fill="none"/>
          <rect x="1" y="3" width="8" height="8" stroke="currentColor" stroke-width="1.5" fill="white"/>
        </svg>
      </button>
      <button class="control-btn close" title="关闭本组" @click="handleClose">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <line x1="1" y1="1" x2="11" y2="11" stroke="currentColor" stroke-width="1.5"/>
          <line x1="11" y1="1" x2="1" y2="11" stroke="currentColor" stroke-width="1.5"/>
        </svg>
      </button>
    </div>
    
    <!-- 关闭标签确认对话框 -->
    <Modal
      :visible="showCloseTabModal"
      title="关闭连接"
      :content="closeTabContent"
      confirm-text="关闭"
      danger
      @close="showCloseTabModal = false"
      @confirm="confirmCloseTab"
    />

    <!-- 关闭窗口确认对话框 -->
    <Modal
      :visible="showCloseWindowModal"
      title="关闭窗口"
      content="确定关闭当前窗口？所有标签页将被关闭。"
      confirm-text="关闭窗口"
      danger
      @close="showCloseWindowModal = false"
      @confirm="confirmCloseWindow"
    />
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Window, Events } from '@wailsio/runtime'
import { useSSHTabsStore } from '../../../stores/sshTabs'
import { useSSHConnectionsStore } from '../../../stores/sshConnections'
import { useSSHLayoutStore } from '../../../stores/sshLayout'
import { storeToRefs } from 'pinia'
import Modal from '../../../components/Modal.vue'
import { Disconnect } from "../../../../bindings/changeme/ssh/sshservice.js"
import { listenToGroupUpdates } from '../../../utils/sshEvents'

const route = useRoute()
const router = useRouter()
const tabsStore = useSSHTabsStore()
const connectionsStore = useSSHConnectionsStore()
const sshLayoutStore = useSSHLayoutStore()
const { tabs, activeTabId } = storeToRefs(tabsStore)

// 从路由参数获取 groupID
const groupID = computed(() => route.query.group || '')
const isMaximised = ref(false)

// 断线状态映射
const disconnectedTabs = ref(new Set())

// 模态框状态
const showCloseTabModal = ref(false)
const tabToClose = ref(null)

// 事件监听器清理函数
let cleanupGroupListener = null

// 计算属性：关闭标签的内容
const closeTabContent = computed(() => {
  const name = tabToClose.value?.name || '该连接'
  return `确定要关闭 "${name}" 吗？`
})

// 获取带唯一标识的标签名称
const getTabDisplayName = (tab) => {
  const sameNameTabs = tabs.value.filter(t => t.name === tab.name)
  
  if (sameNameTabs.length > 1) {
    const shortId = tab.id.slice(-8)
    return `${tab.name} (${shortId})`
  }
  
  return tab.name
}

// 切换最大化/恢复
async function toggleMaximise() {
  if (isMaximised.value) {
    await Window.Restore()
  } else {
    await Window.Maximise()
  }
}

// 更新最大化按钮状态
async function updateMaximiseButton() {
  isMaximised.value = await Window.IsMaximised()
}

// 打开SSH连接（切换到对应标签）
const openConnection = (connId) => {
  console.log('[TopBookmarkBar] 🖱️ 点击标签:', connId)
  
  // 激活标签
  tabsStore.activateTab(connId)
  
  // 更新 URL 参数，记录当前激活的连接 ID
  router.replace({
    query: {
      ...route.query,
      activeConn: connId
    }
  })
  
  // 同时设置 SSH Layout 的当前连接
  sshLayoutStore.setCurrentConnection(connId)
  
  console.log('[TopBookmarkBar] ✅ 已激活标签:', connId)
}

// 关闭标签（显示确认对话框）
const closeTab = (connId) => {
  const tab = tabs.value.find(t => t.id === connId)
  tabToClose.value = tab
  showCloseTabModal.value = true
}

// 确认关闭标签
const confirmCloseTab = async () => {
  if (!tabToClose.value) return

  try {
    const tab = tabToClose.value
    const wasActive = tab.active
    const remainingTabs = tabs.value.filter(t => t.id !== tab.id)

    // ✅ 不再调用 Disconnect（断开整个SSH连接），而是关闭 DockviewLayout 窗口
    // DockviewLayout 内部的 onDidRemovePanel 会清理 session
    // 关闭标签
    tabsStore.closeTab(tab.id)

    // ✅ 如果关闭的是激活标签，切换到下一个标签
    if (wasActive && remainingTabs.length > 0) {
      const newActiveId = tabsStore.activeTabId
      if (newActiveId) {
        sshLayoutStore.setCurrentConnection(newActiveId)
      }
    }

    await connectionsStore.refresh()
    showCloseTabModal.value = false
    tabToClose.value = null
  } catch (error) {
    console.error('关闭连接失败:', error)
    alert('关闭连接失败: ' + error.message)
  }
}

// 处理关闭按钮点击 - 需要确认
const showCloseWindowModal = ref(false)

// 重连标签
const reconnectTab = async (connId) => {
  console.log('[TopBookmarkBar] 🔄 尝试重连:', connId)
  try {
    const { Reconnect } = await import('../../../../bindings/changeme/ssh/sshservice.js')
    await Reconnect(connId)
    console.log('[TopBookmarkBar] ✅ 重连请求已发送')
  } catch (error) {
    console.error('[TopBookmarkBar] ❌ 重连失败:', error)
  }
}

const handleClose = () => {
  showCloseWindowModal.value = true
}

const confirmCloseWindow = async () => {
  showCloseWindowModal.value = false
  try {
    await Window.Close()
  } catch (error) {
    console.error('[TopBookmarkBar] 关闭窗口失败:', error)
  }
}

// 拖拽开始
const handleDragStart = (event, tab) => {
  event.dataTransfer.setData('text/plain', JSON.stringify(tab))
  event.dataTransfer.effectAllowed = 'move'
}

onMounted(() => {
  updateMaximiseButton()
  setInterval(updateMaximiseButton, 500)

  // 监听 tabs 变化，如果全部关闭则自动关闭窗口
  watch(tabs, (newTabs) => {
    if (newTabs.length === 0) {
      console.log('[TopBookmarkBar] ⚠️ 所有标签已关闭，自动关闭窗口')
      setTimeout(() => {
        Window.Close()
      }, 100)
    }
  }, { deep: true })

  // 监听断线事件
  Events.On('ssh:connection-disconnected', (event) => {
    const data = event?.data
    if (!data) return
    console.log('[TopBookmarkBar] ⚠️ 连接断开:', data.connID)
    disconnectedTabs.value.add(data.connID)
    tabsStore.updateTabStatus(data.connID, 'disconnected')
  })

  // 监听重连成功事件
  Events.On('ssh:connection-reconnected', (event) => {
    const data = event?.data
    if (!data) return
    console.log('[TopBookmarkBar] ✅ 重连成功:', data.connID)
    disconnectedTabs.value.delete(data.connID)
    tabsStore.updateTabStatus(data.connID, 'connected')
  })

  // 监听重连中事件
  Events.On('ssh:connection-reconnecting', (event) => {
    const data = event?.data
    if (!data) return
    console.log('[TopBookmarkBar] 🔄 重连中:', data.connID)
    tabsStore.updateTabStatus(data.connID, 'reconnecting')
  })
  
  // 监听当前分组的更新事件
  if (groupID.value) {
    let lastConnectionsHash = '' // 用于检测数据是否真的变化
    
    cleanupGroupListener = listenToGroupUpdates(groupID.value, (connections) => {
      // 生成数据哈希，避免重复处理相同的数据
      const currentHash = JSON.stringify(connections.map(c => c.id))
      if (currentHash === lastConnectionsHash) {
        // 静默跳过，不打印日志
        return
      }
      
      console.log('[TopBookmarkBar] 📡 收到分组更新，连接数:', connections.length)
      lastConnectionsHash = currentHash
      
      // 如果所有连接都关闭了，自动关闭窗口
      if (connections.length === 0) {
        console.log('[TopBookmarkBar] ⚠️ 所有连接已关闭，自动关闭窗口')
        Window.Close()
        return
      }
      
      // ✅ 使用增量更新，而不是 closeAllTabs + addTab
      // 构建新的标签列表
      const newTabs = connections.map(conn => ({
        id: conn.id,
        name: conn.name || conn.host,
        status: conn.status || 'disconnected',
        active: false
      }))
      
      // 调用 store 的 refreshTabs（现在是增量更新）
      tabsStore.refreshTabs(newTabs)
      
      // 如果没有激活的标签，激活第一个并设置 currentConnectionId
      if (tabs.value.length > 0 && !activeTabId.value) {
        const firstConnId = tabs.value[0].id
        tabsStore.activateTab(firstConnId)
        sshLayoutStore.setCurrentConnection(firstConnId)
        console.log('[TopBookmarkBar] ✅ 已激活第一个连接:', firstConnId)
      }
    })
    
    console.log('[TopBookmarkBar] 👂 已注册分组监听器:', groupID.value)
  }
})

onUnmounted(() => {
  // 清理事件监听器
  if (cleanupGroupListener) {
    cleanupGroupListener()
    console.log('[TopBookmarkBar] 🔇 已移除分组监听器')
  }

  // 清理断线事件监听
  Events.Off('ssh:connection-disconnected')
  Events.Off('ssh:connection-reconnected')
  Events.Off('ssh:connection-reconnecting')
})
</script>

<style scoped>
.top-bar {
  height: 2.5rem;
  background: var(--bg-toolbar);
  border-bottom: 0.0625rem solid var(--border-default);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0.5rem;
  --wails-draggable: drag;
}

/* 标签页容器 */
.tab-container {
  display: flex;
  align-items: center;
  gap: 0.125rem;
  flex: 1;
  overflow-x: auto;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.75rem;
  background: var(--surface-2);
  border: 0.0625rem solid transparent;
  border-bottom: 2px solid transparent;
  border-radius: 0.25rem 0.25rem 0 0;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  min-width: 8rem;
  max-width: 15rem;
}

.tab-item:hover {
  background: var(--surface-hover);
}

.tab-item.disconnected {
  border-bottom-color: var(--accent-danger);
}

.tab-item.disconnected .tab-name {
  color: var(--accent-danger);
}

.tab-item.active {
  background: var(--bg-panel);
  border-color: var(--border-default);
  border-bottom-color: var(--accent-primary);
}

.tab-status {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.tab-status.connected {
  background: var(--accent-success);
  box-shadow: 0 0 0.375rem rgba(72, 187, 120, 0.5);
}

.tab-status.disconnected {
  background: var(--accent-danger);
  box-shadow: 0 0 0.375rem rgba(252, 129, 129, 0.5);
}

.tab-status.reconnecting {
  background: var(--accent-warning);
  box-shadow: 0 0 0.375rem rgba(251, 211, 141, 0.5);
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.tab-name {
  color: var(--text-primary);
  font-size: 0.8125rem;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  background: transparent;
  border: none;
  border-radius: 0.25rem;
  color: var(--text-secondary);
  cursor: pointer;
  opacity: 0;
  transition: all 0.2s;
  flex-shrink: 0;
}

.tab-reconnect {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  background: transparent;
  border: none;
  border-radius: 0.25rem;
  color: var(--accent-warning);
  cursor: pointer;
  opacity: 0;
  transition: all 0.2s;
  flex-shrink: 0;
}

.tab-item:hover .tab-reconnect {
  opacity: 1;
}

.tab-reconnect:hover {
  background: var(--warning-bg);
  color: var(--warning-light);
}

.tab-item:hover .tab-close {
  opacity: 1;
}

.tab-close:hover {
  background: var(--surface-hover);
  color: var(--accent-danger);
}

/* 窗口控制按钮 */
.window-controls {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-left: 0.5rem;
  padding-left: 0.5rem;
  border-left: 0.0625rem solid var(--border-default);
}

.control-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  background: transparent;
  border: none;
  border-radius: 0.5rem;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.control-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.control-btn.close:hover {
  background: var(--danger-bg);
  color: var(--accent-danger);
}

/* 滚动条样式 */
.tab-container::-webkit-scrollbar {
  height: 0.125rem;
}

.tab-container::-webkit-scrollbar-track {
  background: transparent;
}

.tab-container::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb);
  border-radius: 0.0625rem;
}

.tab-container::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover);
}
</style>
