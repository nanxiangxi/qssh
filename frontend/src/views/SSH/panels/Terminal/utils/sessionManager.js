/**
 * 终端会话管理器
 * 管理会话元数据、标签、状态
 */

// 会话状态常量
export const SESSION_STATUS = {
  IDLE: 'idle',
  STARTING: 'starting',
  ACTIVE: 'active',
  DISCONNECTED: 'disconnected',
  ERROR: 'error'
}

// 会话类型
export const SESSION_TYPE = {
  NORMAL: 'normal',
  AI: 'ai'
}

export class SessionManager {
  constructor() {
    // sessionId → sessionMeta
    this.sessions = new Map()
    // 监听器
    this.listeners = new Set()
  }

  // 创建会话
  createSession(sessionId, connId, options = {}) {
    const session = {
      id: sessionId,
      connId,
      type: options.type || SESSION_TYPE.NORMAL,
      label: options.label || '',
      status: SESSION_STATUS.IDLE,
      createdAt: Date.now(),
      lastActiveAt: Date.now(),
      commandCount: 0,
      isRecording: false,
      tags: options.tags || [],
      metadata: options.metadata || {}
    }

    this.sessions.set(sessionId, session)
    this.notifyListeners('created', session)
    return session
  }

  // 获取会话
  getSession(sessionId) {
    return this.sessions.get(sessionId)
  }

  // 获取所有会话
  getAllSessions() {
    return Array.from(this.sessions.values())
  }

  // 获取指定连接的会话
  getSessionsByConnId(connId) {
    return this.getAllSessions().filter(s => s.connId === connId)
  }

  // 获取活跃会话
  getActiveSessions() {
    return this.getAllSessions().filter(s => s.status === SESSION_STATUS.ACTIVE)
  }

  // 更新会话状态
  updateStatus(sessionId, status) {
    const session = this.sessions.get(sessionId)
    if (!session) return

    session.status = status
    session.lastActiveAt = Date.now()
    this.notifyListeners('statusChanged', session)
  }

  // 更新会话标签
  updateLabel(sessionId, label) {
    const session = this.sessions.get(sessionId)
    if (!session) return

    session.label = label
    this.notifyListeners('updated', session)
  }

  // 添加标签
  addTag(sessionId, tag) {
    const session = this.sessions.get(sessionId)
    if (!session || session.tags.includes(tag)) return

    session.tags.push(tag)
    this.notifyListeners('updated', session)
  }

  // 移除标签
  removeTag(sessionId, tag) {
    const session = this.sessions.get(sessionId)
    if (!session) return

    session.tags = session.tags.filter(t => t !== tag)
    this.notifyListeners('updated', session)
  }

  // 更新元数据
  updateMetadata(sessionId, key, value) {
    const session = this.sessions.get(sessionId)
    if (!session) return

    session.metadata[key] = value
    this.notifyListeners('updated', session)
  }

  // 增加命令计数
  incrementCommandCount(sessionId) {
    const session = this.sessions.get(sessionId)
    if (!session) return

    session.commandCount++
    session.lastActiveAt = Date.now()
  }

  // 设置录制状态
  setRecording(sessionId, isRecording) {
    const session = this.sessions.get(sessionId)
    if (!session) return

    session.isRecording = isRecording
    this.notifyListeners('updated', session)
  }

  // 删除会话
  removeSession(sessionId) {
    const session = this.sessions.get(sessionId)
    if (!session) return

    this.sessions.delete(sessionId)
    this.notifyListeners('removed', session)
  }

  // 搜索会话
  searchSessions(keyword) {
    if (!keyword) return this.getAllSessions()

    const lower = keyword.toLowerCase()
    return this.getAllSessions().filter(session =>
      session.label.toLowerCase().includes(lower) ||
      session.id.toLowerCase().includes(lower) ||
      session.tags.some(t => t.toLowerCase().includes(lower))
    )
  }

  // 按标签筛选
  filterByTag(tag) {
    return this.getAllSessions().filter(s => s.tags.includes(tag))
  }

  // 获取统计信息
  getStats() {
    const sessions = this.getAllSessions()
    return {
      total: sessions.length,
      active: sessions.filter(s => s.status === SESSION_STATUS.ACTIVE).length,
      idle: sessions.filter(s => s.status === SESSION_STATUS.IDLE).length,
      disconnected: sessions.filter(s => s.status === SESSION_STATUS.DISCONNECTED).length,
      error: sessions.filter(s => s.status === SESSION_STATUS.ERROR).length,
      recording: sessions.filter(s => s.isRecording).length,
      ai: sessions.filter(s => s.type === SESSION_TYPE.AI).length
    }
  }

  // 监听变化
  addListener(callback) {
    this.listeners.add(callback)
    return () => this.listeners.delete(callback)
  }

  // 通知监听器
  notifyListeners(event, session) {
    this.listeners.forEach(callback => {
      try {
        callback(event, session)
      } catch (e) {
        console.error('[SessionManager] 监听器错误:', e)
      }
    })
  }

  // 导出数据
  export() {
    return {
      sessions: this.getAllSessions(),
      stats: this.getStats(),
      exportTime: new Date().toISOString()
    }
  }
}

// 单例实例
let instance = null

export function getSessionManager() {
  if (!instance) {
    instance = new SessionManager()
  }
  return instance
}
