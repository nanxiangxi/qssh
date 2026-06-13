/**
 * 私有云端客户端
 * 按 host:port 去重，最新覆盖旧的
 */

import { Events } from '@wailsio/runtime'

let config = {
  enabled: false,
  serverUrl: '',
  token: '',
  deviceId: '',
  autoSyncTo: false,    // 自动同步到云端
  autoSyncFrom: false,  // 自动从云端同步
  syncInterval: 60000,
}

let syncTimer = null
let connected = false

// ==================== 配置 ====================

export function getCloudConfig() {
  return { ...config }
}

export function setCloudConfig(cfg) {
  config = { ...config, ...cfg }
  if (config.enabled && config.serverUrl && config.token) {
    startSync()
  } else {
    stopSync()
  }
}

// ==================== API ====================

async function apiRequest(path, options = {}) {
  const { serverUrl, token } = config
  if (!serverUrl || !token) throw new Error('云端未配置')

  const url = `${serverUrl.replace(/\/$/, '')}${path}`
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
    ...options.headers,
  }

  const res = await fetch(url, { ...options, headers })
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'request_failed' }))
    throw new Error(error.error || `HTTP ${res.status}`)
  }
  return res.json()
}

// ==================== 设备 ====================

export async function registerDevice() {
  try {
    const device = await apiRequest('/api/register', {
      method: 'POST',
      body: JSON.stringify({
        name: getDeviceName(),
        host: getLocalHost(),
        port: 0,
        os: navigator.platform,
        version: '0.3.0',
        token: config.token,
        timestamp: new Date().toISOString(),
        nonce: generateNonce(),
      }),
    })
    config.deviceId = device.id
    connected = true
    Events.Emit('cloud:connected', { deviceId: device.id })
    console.log('[Cloud] 设备注册成功:', device.id)
    return device
  } catch (e) {
    connected = false
    Events.Emit('cloud:error', { error: e.message })
    console.error('[Cloud] 设备注册失败:', e.message)
    throw e
  }
}

export async function sendHeartbeat() {
  if (!connected || !config.deviceId) return
  try {
    await apiRequest('/api/heartbeat', {
      method: 'POST',
      body: JSON.stringify({ deviceId: config.deviceId }),
    })
  } catch (e) {
    console.warn('[Cloud] 心跳失败:', e.message)
  }
}

// ==================== 同步 ====================

/** 上传连接配置到云端 */
export async function uploadConnections(connections) {
  try {
    const syncData = {
      connections: connections.map(c => ({
        id: c.id,
        name: c.name,
        host: c.host,
        port: c.port || 22,
        username: c.username,
        password: c.password || '',
        keyPath: c.keyPath || '',
      })),
      deviceId: config.deviceId,
    }
    await apiRequest('/api/sync', {
      method: 'POST',
      body: JSON.stringify(syncData),
    })
    console.log('[Cloud] 已上传:', connections.length, '个连接')
    Events.Emit('cloud:synced', { direction: 'upload', count: connections.length })
  } catch (e) {
    console.error('[Cloud] 上传失败:', e.message)
    Events.Emit('cloud:error', { error: e.message })
  }
}

/** 从云端下载连接配置 */
export async function downloadConnections() {
  try {
    const data = await apiRequest('/api/sync')
    if (data && data.connections) {
      console.log('[Cloud] 已下载:', data.connections.length, '个连接')
      Events.Emit('cloud:synced', { direction: 'download', count: data.connections.length })
      return data.connections
    }
    return []
  } catch (e) {
    console.error('[Cloud] 下载失败:', e.message)
    Events.Emit('cloud:error', { error: e.message })
    return []
  }
}

// ==================== 自动同步 ====================

export function startSync() {
  stopSync()
  if (!config.enabled || !config.serverUrl || !config.token) return

  console.log('[Cloud] 启动同步，间隔:', config.syncInterval / 1000, '秒',
    '| 上传:', config.autoSyncTo, '| 下载:', config.autoSyncFrom)

  // 首次注册
  registerDevice().catch(() => {})

  syncTimer = setInterval(async () => {
    await sendHeartbeat()

    // 自动上传本地连接到云端
    if (config.autoSyncTo) {
      Events.Emit('cloud:request-upload')
    }

    // 自动从云端下载
    if (config.autoSyncFrom) {
      const remoteConns = await downloadConnections()
      if (remoteConns.length > 0) {
        Events.Emit('cloud:remote-connections', { connections: remoteConns })
      }
    }
  }, config.syncInterval)
}

export function stopSync() {
  if (syncTimer) {
    clearInterval(syncTimer)
    syncTimer = null
  }
  connected = false
}

export function isConnected() {
  return connected
}

// ==================== 工具 ====================

function getDeviceName() {
  return window.location.hostname || '启SSH客户端'
}

function getLocalHost() {
  return window.location.hostname || 'localhost'
}

function generateNonce() {
  const arr = new Uint8Array(16)
  crypto.getRandomValues(arr)
  return Array.from(arr, b => b.toString(16).padStart(2, '0')).join('')
}
