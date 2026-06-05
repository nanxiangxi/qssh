import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { GetAllConnections } from '../../bindings/changeme/ssh/sshservice.js'

export const useSSHConnectionsStore = defineStore('sshConnections', () => {
  // 连接列表
  const connections = ref([])
  
  // 加载状态
  const loading = ref(false)
  
  // 最后更新时间
  const lastUpdated = ref(null)

  // 按状态分组的连接
  const groupedConnections = computed(() => {
    const groups = {
      connected: [],
      disconnected: []
    }
    
    connections.value.forEach(conn => {
      if (conn.status === 'connected') {
        groups.connected.push(conn)
      } else {
        groups.disconnected.push(conn)
      }
    })
    
    return groups
  })

  // 已连接的连接数
  const connectedCount = computed(() => {
    return connections.value.filter(c => c.status === 'connected').length
  })

  // 加载所有连接
  const loadConnections = async () => {
    loading.value = true
    try {
      const data = await GetAllConnections()
      connections.value = data || []
      lastUpdated.value = new Date()
      console.log('[SSHConnections] 加载了', connections.value.length, '个连接')
    } catch (error) {
      console.error('[SSHConnections] 加载连接列表失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 更新连接状态
  const updateConnectionStatus = (connId, status) => {
    const conn = connections.value.find(c => c.id === connId)
    if (conn) {
      conn.status = status
      console.log(`[SSHConnections] 更新连接 ${connId} 状态为: ${status}`)
    }
  }

  // 添加新连接
  const addConnection = (conn) => {
    connections.value.push(conn)
    console.log('[SSHConnections] 添加新连接:', conn.id)
  }

  // 删除连接
  const removeConnection = (connId) => {
    const index = connections.value.findIndex(c => c.id === connId)
    if (index > -1) {
      connections.value.splice(index, 1)
      console.log('[SSHConnections] 删除连接:', connId)
    }
  }

  // 刷新连接列表
  const refresh = async () => {
    await loadConnections()
  }
  
  // 直接更新连接列表（用于事件同步）
  let _updateTimer = null
  let _lastKey = ''
  const updateConnections = (newConnections) => {
    if (_updateTimer) clearTimeout(_updateTimer)
    _updateTimer = setTimeout(() => {
      _updateTimer = null
      // 用关键字段生成稳定指纹，忽略后端附加的动态字段
      const newKey = newConnections.map(c =>
        `${c.id}:${c.status}:${c.name}:${c.host}:${c.port}:${c.saved}:${c.group_id}`
      ).join('|')
      if (newKey === _lastKey) return
      _lastKey = newKey
      console.log('[SSHConnections] 🔄 更新连接列表，新连接数:', newConnections.length)
      connections.value = newConnections
      lastUpdated.value = new Date()
    }, 300)
  }

  return {
    connections,
    loading,
    lastUpdated,
    groupedConnections,
    connectedCount,
    loadConnections,
    updateConnectionStatus,
    addConnection,
    removeConnection,
    refresh,
    updateConnections  // 新增
  }
})
