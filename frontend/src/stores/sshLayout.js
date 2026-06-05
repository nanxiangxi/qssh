import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 面板类型定义
export const PANEL_TYPES = {
  TERMINAL: 'terminal',           // 终端
  FILE_MANAGER: 'fileManager',    // 文件管理器
  MONITOR: 'monitor',             // 资源监控
  AI_CHAT: 'aiChat',              // AI 对话
  LOGS: 'logs',                   // 日志查看
  PORT_FORWARD: 'portForward',    // 端口转发
  FIREWALL: 'firewall',           // 防火墙
  GUARDIAN: 'guardian',           // 进程守护
  BATCH_CMD: 'batchCmd',         // 批量命令
}

// 面板配置
export const PANEL_CONFIG = {
  [PANEL_TYPES.TERMINAL]: {
    title: '终端',
    icon: 'command-line',
    component: 'TerminalPanel',
  },
  [PANEL_TYPES.FILE_MANAGER]: {
    title: '文件管理',
    icon: 'folder',
    component: 'FileManagerPanel',
  },
  [PANEL_TYPES.MONITOR]: {
    title: '资源监控',
    icon: 'chart',
    component: 'MonitorPanel',
  },
  [PANEL_TYPES.AI_CHAT]: {
    title: 'AI 助手',
    icon: 'robot',
    component: 'AIChatPanel',
  },
  [PANEL_TYPES.LOGS]: {
    title: '日志',
    icon: 'document-text',
    component: 'LogsPanel',
  },
  [PANEL_TYPES.PORT_FORWARD]: {
    title: '端口转发',
    icon: 'port-forward',
    component: 'PortForwardPanel',
  },
  [PANEL_TYPES.FIREWALL]: {
    title: '防火墙',
    icon: 'firewall',
    component: 'FirewallPanel',
  },
  [PANEL_TYPES.GUARDIAN]: {
    title: '进程守护',
    icon: 'guardian',
    component: 'ProcessGuardPanel',
  },
  [PANEL_TYPES.BATCH_CMD]: {
    title: '批量命令',
    icon: 'batch-cmd',
    component: 'BatchCommandPanel',
  },
}

// 布局类型
export const LAYOUT_TYPES = {
  VERTICAL: 'vertical',     // 垂直堆叠
  HORIZONTAL: 'horizontal', // 水平并排
  GRID: 'grid',            // 网格布局（默认）
}

export const useSSHLayoutStore = defineStore('sshLayout', () => {
  // 当前激活的连接 ID
  const currentConnectionId = ref(null)
  
  // 所有连接的布局配置（只在内存中，不持久化）
  // { connectionId: { activePanels, layoutType, panelSizes, gridConfig } }
  const connectionsLayout = ref({})
  
  // 获取默认网格配置
  function getDefaultGridConfig() {
    return {
      columns: 1,
      rows: 1,
      gap: 8,
      items: {}
    }
  }
  
  // 获取或创建连接的布局配置
  function getOrCreateLayout(connectionId) {
    if (!connectionsLayout.value[connectionId]) {
      connectionsLayout.value[connectionId] = {
        activePanels: ['terminal'],
        layoutType: LAYOUT_TYPES.GRID,
        panelSizes: {},
        gridConfig: getDefaultGridConfig(),
      }
    }
    return connectionsLayout.value[connectionId]
  }
  
  // 获取当前连接的布局配置
  function getCurrentLayout() {
    if (!currentConnectionId.value) {
      return {
        activePanels: ['terminal'],
        layoutType: LAYOUT_TYPES.GRID,
        panelSizes: {},
        gridConfig: getDefaultGridConfig(),
      }
    }
    return getOrCreateLayout(currentConnectionId.value)
  }
  
  // 计算属性：当前激活的面板列表
  const activePanels = computed(() => {
    const layout = getCurrentLayout()
    return layout.activePanels || []
  })
  
  // 计算属性：当前布局类型
  const layoutType = computed(() => {
    const layout = getCurrentLayout()
    return layout.layoutType || LAYOUT_TYPES.GRID
  })
  
  // 计算属性：当前面板尺寸配置
  const panelSizes = computed(() => {
    const layout = getCurrentLayout()
    return layout.panelSizes || {}
  })
  
  // 计算属性：当前网格配置
  const gridConfig = computed(() => {
    const layout = getCurrentLayout()
    return layout.gridConfig || getDefaultGridConfig()
  })
  
  // 计算属性：是否有多个面板
  const hasMultiplePanels = computed(() => {
    return activePanels.value.length > 1
  })
  
  // 计算属性：单个面板时占满
  const isSinglePanel = computed(() => {
    return activePanels.value.length === 1
  })
  
  /**
   * 设置当前激活的连接
   */
  function setCurrentConnection(connectionId) {
    currentConnectionId.value = connectionId
    // 确保该连接有布局配置
    getOrCreateLayout(connectionId)
  }
  
  /**
   * 添加面板
   */
  function addPanel(panelType) {
    const layout = getCurrentLayout()
    
    // 检查是否已存在
    if (layout.activePanels.includes(panelType)) {
      return
    }
    
    layout.activePanels.push(panelType)
  }
  
  /**
   * 移除面板
   */
  function removePanel(panelType) {
    const layout = getCurrentLayout()
    const index = layout.activePanels.indexOf(panelType)
    
    if (index > -1) {
      layout.activePanels.splice(index, 1)
      delete layout.panelSizes[panelType]
    }
  }
  
  /**
   * 切换面板（存在则移除，不存在则添加）
   */
  function togglePanel(panelType) {
    const layout = getCurrentLayout()
    
    console.log('[SSHLayout] 🔄 togglePanel 被调用:', panelType)
    console.log('[SSHLayout]    - 当前 activePanels:', layout.activePanels)
    
    if (layout.activePanels.includes(panelType)) {
      console.log('[SSHLayout]    - 移除面板:', panelType)
      removePanel(panelType)
    } else {
      console.log('[SSHLayout]    - 添加面板:', panelType)
      addPanel(panelType)
    }
    
    console.log('[SSHLayout]    - 更新后 activePanels:', layout.activePanels)
  }
  
  /**
   * 设置布局类型
   */
  function setLayoutType(type) {
    const layout = getCurrentLayout()
    layout.layoutType = type
    
    let newGridConfig = getDefaultGridConfig()
    
    switch (type) {
      case LAYOUT_TYPES.VERTICAL:
        newGridConfig.columns = 1
        newGridConfig.rows = Math.max(1, layout.activePanels.length)
        break
      case LAYOUT_TYPES.HORIZONTAL:
        newGridConfig.columns = Math.max(1, layout.activePanels.length)
        newGridConfig.rows = 1
        break
      case LAYOUT_TYPES.GRID:
      default:
        newGridConfig.columns = 'auto'
        newGridConfig.rows = 'auto'
        break
    }
    
    layout.gridConfig = newGridConfig
  }
  
  /**
   * 更新面板尺寸
   */
  function updatePanelSize(panelType, size) {
    const layout = getCurrentLayout()
    if (size >= 15 && size <= 85) {
      layout.panelSizes[panelType] = size
    }
  }
  
  /**
   * 更新网格配置
   */
  function updateGridConfig(config) {
    const layout = getCurrentLayout()
    layout.gridConfig = { ...getDefaultGridConfig(), ...config }
  }
  
  /**
   * 更新单个面板的网格配置
   */
  function updatePanelGridConfig(panelType, config) {
    const layout = getCurrentLayout()
    if (!layout.gridConfig.items) {
      layout.gridConfig.items = {}
    }
    layout.gridConfig.items[panelType] = { 
      ...layout.gridConfig.items[panelType], 
      ...config 
    }
  }
  
  /**
   * 重新排序面板
   */
  function reorderPanels(newOrder) {
    const layout = getCurrentLayout()
    layout.activePanels = newOrder
  }
  
  /**
   * 清空所有面板
   */
  function clearAllPanels() {
    const layout = getCurrentLayout()
    layout.activePanels = []
    layout.panelSizes = {}
    layout.gridConfig = getDefaultGridConfig()
  }
  
  return {
    // 状态
    currentConnectionId,
    activePanels,
    layoutType,
    panelSizes,
    gridConfig,
    
    // 计算属性
    hasMultiplePanels,
    isSinglePanel,
    
    // 方法
    setCurrentConnection,
    addPanel,
    removePanel,
    togglePanel,
    setLayoutType,
    updatePanelSize,
    updateGridConfig,
    updatePanelGridConfig,
    reorderPanels,
    clearAllPanels,
  }
})
