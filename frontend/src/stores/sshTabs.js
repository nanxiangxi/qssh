import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useSSHTabsStore = defineStore('sshTabs', () => {
  // 标签页列表 - 初始为空，由 NewConnection 添加
  const tabs = ref([])

  // 当前激活的标签ID
  const activeTabId = ref(null)

  console.log('[SSHTabs] 🚀 SSHTabs Store 已初始化')
  console.log('[SSHTabs] 📊 初始 tabs:', tabs.value)
  console.log('[SSHTabs] 📊 初始 activeTabId:', activeTabId.value)

  // 获取当前激活的标签
  const activeTab = computed(() => {
    return tabs.value.find(tab => tab.id === activeTabId.value)
  })

  // 获取带唯一标识的标签名称（用于显示）
  const getTabDisplayName = (tab) => {
    // 检查是否有同名标签
    const sameNameTabs = tabs.value.filter(t => t.name === tab.name)
    
    if (sameNameTabs.length > 1) {
      // 如果有同名标签，添加 ID 后8位作为标识
      const shortId = tab.id.slice(-8)
      return `${tab.name} (${shortId})`
    }
    
    return tab.name
  }

  // 添加标签
  const addTab = (tab) => {
    console.log('[SSHTabs] ➕ 添加标签:', tab.id, tab.name)
    // 检查是否已存在
    const existing = tabs.value.find(t => t.id === tab.id)
    if (existing) {
      console.log('[SSHTabs] ⚠️ 标签已存在，激活它')
      // 已存在则激活
      activateTab(tab.id)
      return
    }
    
    // 取消其他标签的激活状态
    tabs.value.forEach(t => t.active = false)
    
    // 添加新标签并激活
    tab.active = true
    tabs.value.push(tab)
    activeTabId.value = tab.id
    console.log('[SSHTabs] ✅ 标签已添加，当前标签数:', tabs.value.length)
  }

  // 关闭标签
  const closeTab = (tabId) => {
    console.log('[SSHTabs] ❌ 关闭标签:', tabId)
    const index = tabs.value.findIndex(t => t.id === tabId)
    if (index === -1) {
      console.log('[SSHTabs] ⚠️ 标签不存在:', tabId)
      return
    }

    // 如果关闭的是当前激活的标签，需要激活另一个
    if (tabs.value[index].active) {
      if (tabs.value.length > 1) {
        // 激活前一个或后一个标签
        const nextIndex = index > 0 ? index - 1 : index + 1
        tabs.value[nextIndex].active = true
        activeTabId.value = tabs.value[nextIndex].id
        console.log('[SSHTabs] 🔄 激活下一个标签:', tabs.value[nextIndex].id)
      } else {
        activeTabId.value = null
        console.log('[SSHTabs] ⚠️ 没有更多标签')
      }
    }

    // 删除标签
    tabs.value.splice(index, 1)
    console.log('[SSHTabs] ✅ 标签已关闭，剩余:', tabs.value.length)
  }

  // 关闭所有标签
  const closeAllTabs = () => {
    console.log('[SSHTabs] ❌ 关闭所有标签')
    tabs.value = []
    activeTabId.value = null
  }

  // 激活标签
  const activateTab = (tabId) => {
    console.log('[SSHTabs] 🔵 激活标签:', tabId)
    tabs.value.forEach(t => t.active = false)
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) {
      tab.active = true
      activeTabId.value = tabId
      console.log('[SSHTabs] ✅ 标签已激活')
    } else {
      console.log('[SSHTabs] ⚠️ 标签不存在:', tabId)
    }
  }

  // 更新标签状态
  const updateTabStatus = (tabId, status) => {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) {
      tab.status = status
    }
  }

  // 更新标签名称
  const updateTabName = (tabId, name) => {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab) {
      tab.name = name
    }
  }

  // 刷新标签列表（用于事件同步）- 纯增量更新，不破坏现有状态
  const refreshTabs = (newTabs) => {
    console.log('[SSHTabs] 🔄 收到刷新请求，新标签数:', newTabs.length, '当前标签数:', tabs.value.length)
    
    // 如果数据完全相同，跳过更新
    if (tabs.value.length === newTabs.length && tabs.value.length > 0) {
      const isSame = tabs.value.every((tab, index) => {
        return tab.id === newTabs[index].id && 
               tab.name === newTabs[index].name &&
               tab.status === newTabs[index].status
      })
      
      if (isSame) {
        console.log('[SSHTabs] ⏭️ 数据未变化，跳过更新')
        return
      }
    }
    
    // ✅ 增量更新策略：只添加缺失的标签，不删除任何标签
    const currentIds = new Set(tabs.value.map(t => t.id))
    
    // 1. 添加新标签（后端有但前端没有的）
    let addedCount = 0
    for (const newTab of newTabs) {
      if (!currentIds.has(newTab.id)) {
        console.log('[SSHTabs] ➕ 添加新标签:', newTab.id)
        tabs.value.push({
          ...newTab,
          active: false
        })
        addedCount++
      }
    }
    
    // 2. 更新现有标签的状态和名称
    let updatedCount = 0
    for (const tab of tabs.value) {
      const newTab = newTabs.find(t => t.id === tab.id)
      if (newTab) {
        // 只更新名称和状态，保持用户的激活状态
        if (tab.name !== newTab.name) {
          tab.name = newTab.name
          updatedCount++
        }
        if (tab.status !== newTab.status) {
          tab.status = newTab.status
          updatedCount++
        }
      }
    }
    
    // 3. ❗ 不删除任何标签！用户手动关闭的标签由 closeTab() 处理
    // 如果后端列表比前端少，说明用户已经关闭了某些标签，不需要再处理
    
    console.log('[SSHTabs] ✅ 刷新完成，添加了', addedCount, '个，更新了', updatedCount, '个，当前标签数:', tabs.value.length)
  }

  return {
    tabs,
    activeTabId,
    activeTab,
    getTabDisplayName,
    addTab,
    closeTab,
    closeAllTabs,
    activateTab,
    updateTabStatus,
    updateTabName,
    refreshTabs
  }
})
