/**
 * 终端会话 Store
 * 管理面板 ↔ SSH session 的映射关系
 * 面板创建 = 会话创建，面板关闭 = 会话关闭
 *
 * 职责：仅管理 session 分配和生命周期，不负责 shell 启动
 * shell 由 TerminalPanel 挂载后自行启动
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as SSHService from '../../bindings/changeme/ssh/sshservice.js'

// SSH 服务器通常限制 MaxSessions=10，留 2 个余量
const MAX_SESSIONS = 8

export const useTerminalSessionStore = defineStore('terminalSessions', () => {
  // panelId → { sessionId, connId, ready, isAI, claimed }
  const sessions = ref({})

  // 分配终端会话 ID（不启动 shell）
  function createSession(panelId, connId, isAI = false) {
    // 检查当前连接的 session 数量
    const currentCount = Object.values(sessions.value).filter(s => s.connId === connId).length
    if (currentCount >= MAX_SESSIONS) {
      console.warn(`[TerminalSessions] 已达最大会话数 (${MAX_SESSIONS})，无法创建新会话`)
      return null
    }
    const sessionId = `term_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`
    sessions.value[panelId] = { sessionId, connId, ready: false, isAI, claimed: false }
    console.log(`[TerminalSessions] 会话分配: ${panelId} → ${sessionId} (AI: ${isAI}) [${currentCount + 1}/${MAX_SESSIONS}]`)
    return sessionId
  }

  // 认领一个未认领的终端
  function claimNextSession() {
    for (const [panelId, sess] of Object.entries(sessions.value)) {
      if (!sess.claimed) {
        sess.claimed = true
        return { panelId, ...sess }
      }
    }
    return null
  }

  // 标记会话就绪（shell 启动成功后由 TerminalPanel 调用）
  function markSessionReady(sessionId) {
    for (const [, sess] of Object.entries(sessions.value)) {
      if (sess.sessionId === sessionId) {
        sess.ready = true
        console.log(`[TerminalSessions] 会话就绪: ${sessionId}`)
        return
      }
    }
  }

  // 关闭终端会话（面板关闭时调用）
  async function closeSession(panelId) {
    const sess = sessions.value[panelId]
    if (!sess) return
    try {
      await SSHService.CloseShellSessionByID(sess.connId, sess.sessionId)
    } catch (e) {}
    delete sessions.value[panelId]
    console.log(`[TerminalSessions] 会话关闭: ${panelId}`)
  }

  function getSession(panelId) {
    return sessions.value[panelId] || null
  }

  function getReadySessions() {
    return Object.entries(sessions.value)
      .filter(([, s]) => s.ready)
      .map(([panelId, s]) => ({ panelId, ...s }))
  }

  function getAITerminals() {
    return Object.entries(sessions.value)
      .filter(([, s]) => s.ready && s.isAI)
      .map(([panelId, s]) => ({ panelId, ...s }))
  }

  function findPanelBySessionId(sessionId) {
    for (const [panelId, sess] of Object.entries(sessions.value)) {
      if (sess.sessionId === sessionId) return panelId
    }
    return null
  }

  // 通过 sessionId 查找会话对象
  function getSessionBySessionId(sessionId) {
    for (const sess of Object.values(sessions.value)) {
      if (sess.sessionId === sessionId) return sess
    }
    return null
  }

  return {
    sessions,
    createSession,
    closeSession,
    getSession,
    getReadySessions,
    getAITerminals,
    findPanelBySessionId,
    claimNextSession,
    markSessionReady,
    getSessionBySessionId
  }
})
