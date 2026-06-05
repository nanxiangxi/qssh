/**
 * 统一日志服务
 * 记录所有操作日志：终端、文件管理、连接管理等
 */

// 日志类型枚举
export const LogType = {
  TERMINAL: 'terminal',           // 终端命令
  FILE_MANAGER: 'fileManager',    // 文件管理操作
  CONNECTION: 'connection',       // 连接操作
  SYSTEM: 'system',               // 系统事件
  ERROR: 'error',                 // 错误信息
  SECURITY: 'security',           // 安全相关
  AI: 'ai',                       // AI 工具操作
  PORT_FORWARD: 'portForward',    // 端口转发
  FIREWALL: 'firewall',           // 防火墙操作
  GUARDIAN: 'guardian',           // 进程守护
}

// 日志级别枚举
export const LogLevel = {
  INFO: 'info',
  SUCCESS: 'success',
  WARNING: 'warning',
  ERROR: 'error',
}

// 日志条目接口
export class LogEntry {
  constructor(type, level, message, details = null) {
    this.id = `log_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    this.type = type
    this.level = level
    this.message = message
    this.details = details
    this.timestamp = new Date().toISOString()
  }
}

// 日志存储（按连接ID分组）
const logsByConnection = new Map()

/**
 * 添加日志
 */
export function addLog(connId, type, level, message, details = null) {
  if (!connId || connId === 'default-connection') {
    console.warn('[Logger] 无效的连接ID，忽略日志:', message)
    return
  }

  // 获取或创建该连接的日志数组
  if (!logsByConnection.has(connId)) {
    logsByConnection.set(connId, [])
  }

  const logs = logsByConnection.get(connId)
  const entry = new LogEntry(type, level, message, details)
  
  // 添加到开头（最新的在前）
  logs.unshift(entry)
  
  // 限制日志数量（最多保留1000条）
  if (logs.length > 1000) {
    logs.pop()
  }

  console.log(`[Logger] [${type}] ${message}`, entry)
  
  return entry
}

/**
 * 获取指定连接的所有日志
 */
export function getLogs(connId) {
  if (!connId || !logsByConnection.has(connId)) {
    return []
  }
  return logsByConnection.get(connId)
}

/**
 * 清空指定连接的日志
 */
export function clearLogs(connId) {
  if (connId && logsByConnection.has(connId)) {
    logsByConnection.set(connId, [])
  }
}

/**
 * 便捷方法：记录终端命令
 */
export function logTerminalCommand(connId, command, output = null) {
  return addLog(connId, LogType.TERMINAL, LogLevel.INFO, `执行命令: ${command}`, {
    command,
    output
  })
}

/**
 * 便捷方法：记录文件操作
 */
export function logFileOperation(connId, operation, filePath, details = null) {
  const messages = {
    upload: `上传文件: ${filePath}`,
    download: `下载文件: ${filePath}`,
    delete: `删除文件: ${filePath}`,
    rename: `重命名文件: ${filePath}`,
    create: `创建文件: ${filePath}`,
    chmod: `修改权限: ${filePath}`,
    compress: `压缩文件: ${filePath}`,
    extract: `解压文件: ${filePath}`,
    browse: `浏览目录: ${filePath}`,
    kill_process: `终止进程: ${filePath}`,
  }
  
  const message = messages[operation] || `${operation}: ${filePath}`
  
  // 浏览目录使用 INFO 级别，其他操作使用 SUCCESS 级别
  const level = operation === 'browse' ? LogLevel.INFO : LogLevel.SUCCESS
  
  return addLog(connId, LogType.FILE_MANAGER, level, message, {
    operation,
    filePath,
    ...details
  })
}

/**
 * 便捷方法：记录连接事件
 */
export function logConnectionEvent(connId, event, details = null) {
  const messages = {
    connect: 'SSH 连接建立',
    disconnect: 'SSH 连接断开',
    reconnect: 'SSH 重新连接',
    timeout: '连接超时',
    error: '连接错误',
  }
  
  const message = messages[event] || event
  const level = event === 'error' || event === 'timeout' ? LogLevel.ERROR : LogLevel.INFO
  
  return addLog(connId, LogType.CONNECTION, level, message, {
    event,
    ...details
  })
}

/**
 * 便捷方法：记录系统事件
 */
export function logSystemEvent(connId, message, level = LogLevel.INFO, details = null) {
  return addLog(connId, LogType.SYSTEM, level, message, details)
}

/**
 * 便捷方法：记录错误
 */
export function logError(connId, message, error = null) {
  return addLog(connId, LogType.ERROR, LogLevel.ERROR, message, {
    error: error?.message || error,
    stack: error?.stack
  })
}

/**
 * 便捷方法：记录 AI 工具操作
 */
export function logAITool(connId, action, command, details = null) {
  const levelMap = {
    call: LogLevel.INFO,
    approve: LogLevel.SUCCESS,
    deny: LogLevel.WARNING,
    execute: LogLevel.INFO,
    result: LogLevel.SUCCESS,
    error: LogLevel.ERROR,
  }
  return addLog(connId, LogType.AI, levelMap[action] || LogLevel.INFO, `[AI] ${action}: ${command}`, {
    action,
    command,
    ...details
  })
}

/**
 * 导出日志为 JSON
 */
export function exportLogs(connId) {
  const logs = getLogs(connId)
  
  if (!logs || logs.length === 0) {
    return null
  }

  const exportData = {
    connectionId: connId,
    exportTime: new Date().toISOString(),
    totalRecords: logs.length,
    logs: logs
  }

  return JSON.stringify(exportData, null, 2)
}
