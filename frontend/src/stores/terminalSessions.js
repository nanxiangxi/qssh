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
  // 复用模式：如果 connId 下已有可用会话，直接复用（用于标签切换场景）
  function createSession(panelId, connId, isAI = false, reuse = false) {
    console.log(`[TerminalSessions] createSession: panelId=${panelId}, connId=${connId}, isAI=${isAI}, reuse=${reuse}`)
    console.log(`[TerminalSessions] 当前会话:`, Object.entries(sessions.value).map(([k, v]) => `${k} → ${v.sessionId}(${v.connId})`).join(', '))

    // 复用模式：查找同连接下已有的可用会话
    if (reuse && !isAI) {
      for (const [existingPanelId, sess] of Object.entries(sessions.value)) {
        if (sess.connId === connId && !sess.isAI && existingPanelId !== panelId) {
          // 复用已有会话
          sessions.value[panelId] = { ...sess, claimed: false }
          console.log(`[TerminalSessions] 复用会话: ${panelId} → ${sess.sessionId}`)
          return sess.sessionId
        }
      }
    }

    // 检查当前连接的 session 数量
    const currentCount = Object.values(sessions.value).filter(s => s.connId === connId).length
    if (currentCount >= MAX_SESSIONS) {
      console.warn(`[TerminalSessions] 已达最大会话数 (${MAX_SESSIONS})，无法创建新会话`)
      return null
    }
    const sessionId = `term_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`
    sessions.value[panelId] = { sessionId, connId, ready: false, isAI, claimed: false }
    console.log(`[TerminalSessions] 新会话分配: ${panelId} → ${sessionId} [${currentCount + 1}/${MAX_SESSIONS}]`)
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

  // 检查会话是否就绪
  function isSessionReady(sessionId) {
    for (const sess of Object.values(sessions.value)) {
      if (sess.sessionId === sessionId) {
        return sess.ready
      }
    }
    return false
  }

  // 关闭终端会话（面板关闭时调用）
  async function closeSession(panelId) {
    const sess = sessions.value[panelId]
    if (!sess) {
      console.log(`[TerminalSessions] closeSession: 面板 ${panelId} 无会话，跳过`)
      return
    }
    console.log(`[TerminalSessions] closeSession: 面板=${panelId}, sessionId=${sess.sessionId}, connId=${sess.connId}`)
    delete sessions.value[panelId]

    // 检查是否还有其他面板在使用同一个 sessionId
    const stillInUse = Object.values(sessions.value).filter(s => s.sessionId === sess.sessionId)
    console.log(`[TerminalSessions] 会话 ${sess.sessionId} 使用者数量: ${stillInUse.length}`, stillInUse.map(s => s.connId))

    if (stillInUse.length === 0) {
      // 没有其他面板使用此会话，才真正关闭 SSH 会话
      try {
        console.log(`[TerminalSessions] 关闭 SSH 会话: ${sess.sessionId}`)
        await SSHService.CloseShellSessionByID(sess.connId, sess.sessionId)
        console.log(`[TerminalSessions] SSH 会话已关闭: ${sess.sessionId}`)
      } catch (e) {
        console.error(`[TerminalSessions] 关闭 SSH 会话失败: ${sess.sessionId}`, e)
      }
    } else {
      console.log(`[TerminalSessions] 会话仍被 ${stillInUse.length} 个面板使用，跳过关闭: ${sess.sessionId}`)
    }

    // 打印当前所有会话状态
    console.log(`[TerminalSessions] 当前会话列表:`, Object.entries(sessions.value).map(([k, v]) => `${k} → ${v.sessionId}`).join(', '))
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
    isSessionReady,
    getSessionBySessionId
  }
})
