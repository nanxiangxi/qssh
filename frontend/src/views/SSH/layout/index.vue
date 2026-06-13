<template>
  <div class="ssh-layout">
    <!-- 顶部书签栏 - SSH连接列表 -->
    <TopBookmarkBar />

    <!-- 内容区容器 -->
    <div class="content-container">
      <!-- ✅ 使用 tabsStore.tabs 而不是 connections，避免后端广播导致组件被移除 -->
      <TabContent
        v-for="tab in tabs"
        :key="tab.id"
        :conn-id="tab.id"
        :is-active="tab.id === currentConnectionId"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import TopBookmarkBar from './TopBookmarkBar.vue'
import TabContent from './TabContent.vue'
import { useLayoutStore } from '../../../stores/layout'
import { useSSHLayoutStore } from '../../../stores/sshLayout'
import { useSSHTabsStore } from '../../../stores/sshTabs'
import { listenToGroupUpdates } from '../../../utils/sshEvents'

const route = useRoute()
const layoutStore = useLayoutStore()
const sshLayoutStore = useSSHLayoutStore()
const tabsStore = useSSHTabsStore()

// ✅ 直接使用 tabsStore.tabs，不依赖后端广播的 connections
const tabs = computed(() => tabsStore.tabs)

let cleanupListener = null
let lastConnectionsHash = ''

// 监听分组更新事件（只用于同步状态，不用于渲染）
onMounted(() => {
  layoutStore.loadLayout()
  const groupID = route.query.group
  console.log('[SSH Layout] Group ID from URL:', groupID)
  
  if (groupID) {
    // 监听分组更新（带防抖）- 仅用于内部状态同步，不再更新 connections
    cleanupListener = listenToGroupUpdates(groupID, (conns) => {
      // 生成数据哈希，避免重复处理相同的数据
      const currentHash = JSON.stringify(conns.map(c => c.id))
      if (currentHash === lastConnectionsHash) {
        // 静默跳过，不打印日志
        return
      }
      
      console.log('[SSH Layout] 📡 收到分组更新，连接数:', conns.length)
      lastConnectionsHash = currentHash
      // ❗ 不再更新 connections.value，由 tabsStore 管理
    })
    
    // 优先从 URL 参数中获取 activeConn（用户上次激活的连接）
    const urlActiveConn = route.query.activeConn
    
    if (urlActiveConn && typeof urlActiveConn === 'string') {
      // URL 中有明确的激活连接 ID，直接使用
      console.log('[SSH Layout] 从 URL 获取激活连接:', urlActiveConn)
      sshLayoutStore.setCurrentConnection(urlActiveConn)
    } else {
      // 没有 activeConn，等待 TopBookmarkBar 初始化标签列表
      console.log('[SSH Layout] 等待 TopBookmarkBar 初始化标签列表')
    }
  } else if (!sshLayoutStore.currentConnectionId) {
    // 如果没有 groupID 且没有当前连接，设置一个默认连接
    sshLayoutStore.setCurrentConnection('default-connection')
  }
})

onUnmounted(() => {
  if (cleanupListener) {
    cleanupListener()
  }
})

const currentConnectionId = computed(() => sshLayoutStore.currentConnectionId)
</script>

<style scoped>
.ssh-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  background: var(--bg-panel);
}

/* 内容区容器 */
.content-container {
  flex: 1;
  position: relative;
  overflow: hidden;
}
</style>
