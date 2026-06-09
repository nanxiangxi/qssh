/**
 * Shell 会话管理组合式函数
 * 管理 SSH Shell 的生命周期、输入输出
 */
import { ref, computed, onUnmounted } from 'vue'
import { Events } from '@wailsio/runtime'
import { SSHService } from '../../../../../../bindings/changeme/ssh/index.js'

export function useShellSession(connId, sessionId, options = {}) {
  const { autoReconnect = true, maxReconnectAttempts = 3 } = options

  // 状态
  const status = ref('idle') // idle | starting | active | disconnected | error
  const error = ref(null)
  const isStarting = ref(false)

  // 重连状态
  const reconnectAttempts = ref(0)
  const isReconnecting = ref(false)

  // 输出回调
  let outputCallback = null
  let statusCallback = null

  // 事件清理函数
  let cleanupOutputListener = null
  let cleanupDisconnectListener = null

  // 计算属性
  const isActive = computed(() => status.value === 'active')
  const isDisconnected = computed(() => status.value === 'disconnected')

  // 启动 Shell
  async function start() {
    if (isStarting.value) {
      console.log('[useShellSession] Shell 正在启动中，跳过')
      return false
    }

    if (!connId || connId === 'default-connection') {
      error.value = '无效的连接 ID'
      status.value = 'error'
      return false
    }

    isStarting.value = true
    status.value = 'starting'
    error.value = null

    try {
      await SSHService.StartShellSessionWithID(connId, sessionId)
      status.value = 'active'
      reconnectAttempts.value = 0
      console.log('[useShellSession] Shell 启动成功:', sessionId)

      // 设置输出监听
      setupOutputListener()

      // 通知状态变化
      if (statusCallback) {
        statusCallback('active')
      }

      Events.Emit('terminal:session-ready', {
        connId,
        sessionId,
        isAI: options.isAI || false
      })

      return true
    } catch (e) {
      const errStr = String(e?.message || e)

      if (errStr.includes('连接不存在')) {
        error.value = '连接不存在，等待连接建立...'
        status.value = 'starting'
        // 延迟重试
        setTimeout(() => {
          isStarting.value = false
          start()
        }, 500)
        return false
      } else if (errStr.includes('administratively prohibited') || errStr.includes('open failed')) {
        error.value = 'SSH 服务器会话数已达上限'
      } else {
        error.value = `启动失败: ${errStr}`
      }

      status.value = 'error'
      console.error('[useShellSession] Shell 启动失败:', e)
      return false
    } finally {
      isStarting.value = false
    }
  }

  // 设置输出监听
  function setupOutputListener() {
    if (cleanupOutputListener) {
      cleanupOutputListener()
    }

    const handler = (event) => {
      const data = event?.data
      if (!data) return

      // 过滤：只处理当前会话的输出
      if (data.sessionID && data.sessionID !== sessionId) return
      if (data.connID && data.connID !== connId) return

      const text = typeof data === 'string' ? data : (data.data || '')
      if (typeof text === 'string' && text && outputCallback) {
        outputCallback(text)
      }
    }

    Events.On('ssh:terminal-output', handler)
    cleanupOutputListener = () => Events.Off('ssh:terminal-output', handler)
  }

  // 设置断线监听
  function setupDisconnectListener() {
    if (cleanupDisconnectListener) {
      cleanupDisconnectListener()
    }

    // 断线事件
    const disconnectHandler = (event) => {
      const data = event?.data
      if (!data || data.connID !== connId) return

      console.log('[useShellSession] 连接断开:', connId)
      status.value = 'disconnected'
      isReconnecting.value = false

      if (statusCallback) {
        statusCallback('disconnected')
      }

      // 自动重连
      if (autoReconnect) {
        attemptReconnect()
      }
    }

    // 重连中事件
    const reconnectingHandler = (event) => {
      const data = event?.data
      if (!data || data.connID !== connId) return

      isReconnecting.value = true
      if (statusCallback) {
        statusCallback('reconnecting')
      }
    }

    // 重连成功事件
    const reconnectedHandler = (event) => {
      const data = event?.data
      if (!data || data.connID !== connId) return

      console.log('[useShellSession] 重连成功:', connId)
      status.value = 'active'
      isReconnecting.value = false
      reconnectAttempts.value = 0

      if (statusCallback) {
        statusCallback('active')
      }

      // 重新启动 Shell
      start()
    }

    // 重连失败事件
    const reconnectFailedHandler = (event) => {
      const data = event?.data
      if (!data || data.connID !== connId) return

      isReconnecting.value = false
      reconnectAttempts.value++

      if (reconnectAttempts.value >= maxReconnectAttempts) {
        status.value = 'error'
        error.value = '自动重连失败，请手动重连'

        if (statusCallback) {
          statusCallback('error')
        }
      }
    }

    Events.On('ssh:connection-disconnected', disconnectHandler)
    Events.On('ssh:connection-reconnecting', reconnectingHandler)
    Events.On('ssh:connection-reconnected', reconnectedHandler)
    Events.On('ssh:connection-reconnect-failed', reconnectFailedHandler)

    cleanupDisconnectListener = () => {
      Events.Off('ssh:connection-disconnected', disconnectHandler)
      Events.Off('ssh:connection-reconnecting', reconnectingHandler)
      Events.Off('ssh:connection-reconnected', reconnectedHandler)
      Events.Off('ssh:connection-reconnect-failed', reconnectFailedHandler)
    }
  }

  // 尝试重连
  async function attemptReconnect() {
    if (isReconnecting.value) return

    isReconnecting.value = true
    reconnectAttempts.value++

    console.log(`[useShellSession] 尝试重连 (${reconnectAttempts.value}/${maxReconnectAttempts})`)

    try {
      await SSHService.Reconnect(connId)
    } catch (e) {
      console.error('[useShellSession] 重连失败:', e)
      isReconnecting.value = false

      if (reconnectAttempts.value < maxReconnectAttempts) {
        // 指数退避
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.value), 10000)
        setTimeout(attemptReconnect, delay)
      } else {
        status.value = 'error'
        error.value = '自动重连失败，请手动重连'
      }
    }
  }

  // 手动重连
  async function reconnect() {
    reconnectAttempts.value = 0
    await attemptReconnect()
  }

  // 写入数据到 Shell
  async function write(data) {
    if (status.value !== 'active') {
      console.warn('[useShellSession] Shell 未激活，无法写入')
      return false
    }

    try {
      await SSHService.WriteToTerminalByID(connId, sessionId, data)
      return true
    } catch (e) {
      console.error('[useShellSession] 写入失败:', e)
      return false
    }
  }

  // 调整终端大小
  async function resize(cols, rows) {
    if (status.value !== 'active') return false

    try {
      await SSHService.ResizeTerminalByID(connId, sessionId, cols, rows)
      return true
    } catch (e) {
      console.error('[useShellSession] 调整大小失败:', e)
      return false
    }
  }

  // 关闭 Shell
  async function close() {
    try {
      await SSHService.CloseShellSessionByID(connId, sessionId)
      status.value = 'idle'
      console.log('[useShellSession] Shell 已关闭:', sessionId)
    } catch (e) {
      console.error('[useShellSession] 关闭失败:', e)
    }
  }

  // 设置输出回调
  function onOutput(callback) {
    outputCallback = callback
  }

  // 设置状态回调
  function onStatusChange(callback) {
    statusCallback = callback
  }

  // 初始化
  function init() {
    setupDisconnectListener()
  }

  // 清理
  function cleanup() {
    if (cleanupOutputListener) {
      cleanupOutputListener()
      cleanupOutputListener = null
    }
    if (cleanupDisconnectListener) {
      cleanupDisconnectListener()
      cleanupDisconnectListener = null
    }
    outputCallback = null
    statusCallback = null
  }

  // 自动清理
  onUnmounted(() => {
    cleanup()
  })

  return {
    // 状态
    status,
    error,
    isStarting,
    isActive,
    isDisconnected,
    isReconnecting,
    reconnectAttempts,

    // 方法
    init,
    start,
    write,
    resize,
    close,
    reconnect,
    onOutput,
    onStatusChange,
    cleanup
  }
}
