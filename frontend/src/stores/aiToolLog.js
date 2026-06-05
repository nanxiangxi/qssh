import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { logAITool } from '../utils/logger'

const MAX_LOGS_PER_CONN = 500

export const useAIToolLogStore = defineStore('aiToolLog', () => {
  const logs = ref({})

  const addLog = (connId, entry) => {
    if (!logs.value[connId]) {
      logs.value[connId] = []
    }
    const arr = logs.value[connId]
    if (arr.length >= MAX_LOGS_PER_CONN) {
      arr.splice(0, arr.length - MAX_LOGS_PER_CONN + 1)
    }
    arr.push({
      id: `log_${Date.now()}`,
      timestamp: new Date().toISOString(),
      ...entry
    })
  }

  const logToolCall = (connId, callId, tool, command) => {
    addLog(connId, { type: 'tool_call', callId, tool, command, status: 'pending' })
    logAITool(connId, 'call', command, { callId, tool })
  }

  const logToolApprove = (connId, callId) => {
    const log = getLogByCallId(connId, callId)
    if (log) {
      log.status = 'approved'
      logAITool(connId, 'approve', log.command, { callId })
    }
  }

  const logToolDeny = (connId, callId) => {
    const log = getLogByCallId(connId, callId)
    if (log) {
      log.status = 'denied'
      logAITool(connId, 'deny', log.command, { callId })
    }
  }

  const logToolResult = (connId, callId, result) => {
    const log = getLogByCallId(connId, callId)
    if (log) {
      log.status = 'completed'
      log.result = result
      log.completedAt = new Date().toISOString()
      logAITool(connId, 'result', log.command, { callId, resultLength: result?.length || 0 })
    }
  }

  // 获取连接的所有日志
  const getLogs = (connId) => {
    return logs.value[connId] || []
  }

  // 获取最近 N 条日志的摘要（用于 AI 上下文）
  const getSessionSummary = (connId, limit = 10) => {
    const connLogs = logs.value[connId] || []
    const recent = connLogs.slice(-limit)
    if (recent.length === 0) return ''

    const lines = recent.map(log => {
      const time = new Date(log.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
      const statusMap = { pending: '待执行', approved: '已批准', denied: '已拒绝', completed: '已完成' }
      const status = statusMap[log.status] || log.status
      return `- [${time}] ${log.command} → ${status}`
    })

    return `本次会话已执行的命令记录：\n${lines.join('\n')}`
  }

  // 获取指定 callId 的日志
  const getLogByCallId = (connId, callId) => {
    const connLogs = logs.value[connId] || []
    return connLogs.find(l => l.callId === callId)
  }

  // 清除连接的日志
  const clearLogs = (connId) => {
    delete logs.value[connId]
  }

  // 日志数量
  const logCount = computed(() => {
    let count = 0
    for (const connId in logs.value) {
      count += logs.value[connId].length
    }
    return count
  })

  return {
    logs,
    addLog,
    logToolCall,
    logToolApprove,
    logToolDeny,
    logToolResult,
    getLogs,
    getSessionSummary,
    clearLogs,
    logCount
  }
})
