<template>
  <header class="top-bar">
    <!-- 左侧：SSH 连接标签页 -->
    <div class="tab-container">
      <div 
        v-for="tab in tabs" 
        :key="tab.id"
        class="tab-item"
        :class="{ active: tab.id === activeTabId }"
        @click="openConnection(tab.id)"
        draggable="true"
        @dragstart="handleDragStart($event, tab)"
      >
        <span class="tab-status" :class="tab.status"></span>
        <span class="tab-name" :title="'ID: ' + tab.id">{{ getTabDisplayName(tab) }}</span>
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
import { Window } from '@wailsio/runtime'
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
    await Disconnect(tabToClose.value.id)
    
    // ✅ 先记录关闭前的状态
    const wasActive = tabToClose.value.active
    const remainingTabs = tabs.value.filter(t => t.id !== tabToClose.value.id)
    
    // 关闭标签（内部会激活下一个标签）
    tabsStore.closeTab(tabToClose.value.id)
    
    // ✅ 如果关闭的是激活标签，需要同步更新 currentConnectionId
    if (wasActive && remainingTabs.length > 0) {
      // closeTab 已经激活了下一个标签，获取新的 activeTabId
      const newActiveId = tabsStore.activeTabId
      if (newActiveId) {
        sshLayoutStore.setCurrentConnection(newActiveId)
        console.log('[TopBookmarkBar] ✅ 已切换到连接:', newActiveId)
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
})
</script>

<style scoped>
.top-bar {
  height: 2.5rem;
  background: rgba(45, 45, 45, 0.95);
  border-bottom: 0.0625rem solid rgba(255, 255, 255, 0.1);
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
  background: rgba(50, 50, 50, 0.4);
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
  background: rgba(60, 60, 60, 0.6);
}

.tab-item.active {
  background: rgba(45, 45, 45, 0.95);
  border-color: rgba(255, 255, 255, 0.1);
  border-bottom-color: #4299e1;
}

.tab-status {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.tab-status.connected {
  background: #48bb78;
  box-shadow: 0 0 0.375rem rgba(72, 187, 120, 0.5);
}

.tab-status.disconnected {
  background: #a0aec0;
}

.tab-name {
  color: #e2e8f0;
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
  color: #a0aec0;
  cursor: pointer;
  opacity: 0;
  transition: all 0.2s;
  flex-shrink: 0;
}

.tab-item:hover .tab-close {
  opacity: 1;
}

.tab-close:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #fc8181;
}

/* 窗口控制按钮 */
.window-controls {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-left: 0.5rem;
  padding-left: 0.5rem;
  border-left: 0.0625rem solid rgba(255, 255, 255, 0.1);
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
  color: #a0aec0;
  cursor: pointer;
  transition: all 0.2s;
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

.control-btn.close:hover {
  background: rgba(229, 62, 62, 0.2);
  color: #fc8181;
}

/* 滚动条样式 */
.tab-container::-webkit-scrollbar {
  height: 0.125rem;
}

.tab-container::-webkit-scrollbar-track {
  background: transparent;
}

.tab-container::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 0.0625rem;
}

.tab-container::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
