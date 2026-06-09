/**
 * AI 工具执行器
 *
 * AI 通过终端 shell 执行命令（像用户输入一样）。
 * 复用结构化终端的提示符检测判断命令完成。
 *
 * 单命令：确保有 AI 终端（或使用用户指定终端），通过终端执行。
 * 并列任务：为每个命令创建/分配独立 AI 终端，通过终端并行执行。
 */

import { Events } from '@wailsio/runtime'
import * as SSHService from '../../bindings/changeme/ssh/sshservice.js'

let initialized = false

// AI 终端管理
const aiTerminals = new Map() // sessionId → { connId, ready }
const pendingCreates = new Map() // requestId → { resolve, timeout, connId }
let requestCounter = 0

// 用户在 AIChatPanel 选择的目标终端
let userSelectedTerminal = null

/** 由 AIChatPanel 调用，设置用户选择的终端 */
export function setTargetTerminal(sessionId) {
  userSelectedTerminal = sessionId || null
  console.log(`[AIToolExecutor] 用户选择终端: ${userSelectedTerminal || '自动'}`)
}

// 提示符检测（与 useBlockManager 一致）
const PROMPT_RE = /[\$#]\s*$/
const ANSI_RE = /\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g

export function initAIToolExecutor() {
  if (initialized) return
  initialized = true
  Events.On('ai:execute-tool', handleToolRequest)
  Events.On('ai:ai-terminal-ready', handleAITerminalReady)
  Events.On('terminal:session-closed', handleTerminalClosed)
  console.log('[AIToolExecutor] 已初始化')
}

export function destroyAIToolExecutor() {
  Events.Off('ai:execute-tool', handleToolRequest)
  Events.Off('ai:ai-terminal-ready', handleAITerminalReady)
  Events.Off('terminal:session-closed', handleTerminalClosed)
  initialized = false
  aiTerminals.clear()
  pendingCreates.clear()
}

// ==================== AI 终端事件 ====================

function handleAITerminalReady(event) {
  const data = event?.data
  if (!data) return
  console.log(`[AIToolExecutor] AI 终端就绪: ${data.sessionId}`)
  aiTerminals.set(data.sessionId, { connId: data.connId, ready: true })

  for (const [reqId, req] of pendingCreates) {
    if (req.connId === data.connId) {
      clearTimeout(req.timeout)
      pendingCreates.delete(reqId)
      req.resolve(data.sessionId)
      break
    }
  }
}

function handleTerminalClosed(event) {
  const data = event?.data
  if (!data || !data.isAI) return
  aiTerminals.delete(data.sessionId)
  console.log(`[AIToolExecutor] AI 终端已关闭: ${data.sessionId}`)
}

// ==================== AI 终端管理 ====================

function getExistingAITerminals(connId) {
  const result = []
  for (const [sessionId, info] of aiTerminals.entries()) {
    if (info.ready && info.connId === connId) result.push(sessionId)
  }
  return result
}

function requestCreateAITerminal(connId) {
  return new Promise((resolve, reject) => {
    const requestId = `req_${++requestCounter}`
    const timeout = setTimeout(() => {
      pendingCreates.delete(requestId)
      reject(new Error('创建 AI 终端超时'))
    }, 15000)
    pendingCreates.set(requestId, { connId, resolve, timeout })
    Events.Emit('ai:create-terminal', { connId, requestId })
    console.log(`[AIToolExecutor] 请求创建 AI 终端`)
  })
}

/** 确保有 N 个可用的 AI 终端，返回 sessionId 列表 */
async function ensureAITerminals(connId, count) {
  const existing = getExistingAITerminals(connId)
  const need = count - existing.length
  if (need <= 0) return existing.slice(0, count)

  console.log(`[AIToolExecutor] 需要创建 ${need} 个 AI 终端`)
  const created = []
  for (let i = 0; i < need; i++) {
    try {
      const sid = await requestCreateAITerminal(connId)
      created.push(sid)
    } catch (e) {
      console.warn('[AIToolExecutor] 创建 AI 终端失败:', e.message)
    }
  }
  return [...existing, ...created]
}

// ==================== 工具请求处理 ====================

async function handleToolRequest(event) {
  const data = event?.data
  if (!data) return

  const { connId, callId, tool, command, args } = data
  console.log(`[AIToolExecutor] 工具请求: ${tool} → ${command || ''}`)

  // 解析参数
  let parsedArgs = {}
  try {
    parsedArgs = typeof args === 'string' ? JSON.parse(args) : (args || {})
  } catch (e) {}

  let result = ''
  try {
    switch (tool) {
      case 'execute_command':
      case 'get_server_info':
      case 'get_system_status':
        result = await handleSingleCommand(connId, tool, command, parsedArgs)
        break
      case 'multi_execute':
        result = await handleMultiExecute(connId, parsedArgs)
        break
      default:
        result = `未知工具: ${tool}`
    }
  } catch (e) {
    result = `执行失败: ${e.message}`
  }

  Events.Emit('ai:submit-tool-result', { callId: String(callId), result: String(result) })
}

// ==================== 单命令 ====================

// 终端名称到 sessionId 的缓存
let terminalNameMap = {} // "终端 1" → sessionId, "AI 终端-1" → sessionId

// 监听终端列表变化，更新名称映射
Events.On('dockview:terminals-changed', (e) => {
  const terminals = e?.data?.terminals
  if (!Array.isArray(terminals)) return
  terminalNameMap = {}
  terminals.forEach(t => {
    if (t.title && t.sessionId) {
      terminalNameMap[t.title] = t.sessionId
    }
  })
})

/** 将用户/AI 指定的终端名解析为 sessionId */
function resolveTerminal(target) {
  if (!target) return null
  // 直接是 sessionId
  if (target.startsWith('term_')) return target
  // 按名称查找
  return terminalNameMap[target] || null
}

async function handleSingleCommand(connId, tool, command, args) {
  let actualCommand = command
  if (tool === 'get_server_info') {
    actualCommand = 'echo "=== 主机名 ===" && hostname && echo "=== 系统信息 ===" && uname -a && echo "=== 运行时间 ===" && uptime && echo "=== 内核版本 ===" && cat /proc/version 2>/dev/null || uname -r'
  } else if (tool === 'get_system_status') {
    actualCommand = 'echo "=== CPU ===" && top -bn1 | head -5 2>/dev/null && echo "" && echo "=== 内存 ===" && free -h 2>/dev/null && echo "" && echo "=== 磁盘 ===" && df -h 2>/dev/null && echo "" && echo "=== 负载 ===" && cat /proc/loadavg 2>/dev/null'
  }

  // 优先级：AI 参数指定 > 用户在面板选择 > 自动 AI 终端
  let sessionId = resolveTerminal(args.targetTerminal) || userSelectedTerminal
  if (!sessionId) {
    try {
      const terminals = await ensureAITerminals(connId, 1)
      sessionId = terminals[0]
    } catch (e) {
      console.warn('[AIToolExecutor] AI 终端创建失败:', e.message)
    }
  }

  console.log(`[AIToolExecutor] 使用终端: ${sessionId} (AI指定: ${args.targetTerminal || '无'}, 用户选择: ${userSelectedTerminal || '无'})`)

  return await executeInTerminal(connId, sessionId, actualCommand)
}

// ==================== 并列任务 ====================

async function handleMultiExecute(connId, args) {
  let commands = []
  try {
    commands = args.commands || []
  } catch (e) {
    return '错误：无法解析命令列表'
  }

  if (!Array.isArray(commands) || commands.length === 0) {
    return '错误：没有要执行的命令'
  }

  if (commands.length > 5) {
    return '错误：最多支持 5 个并行命令'
  }

  console.log(`[AIToolExecutor] 并列执行 ${commands.length} 个命令`)

  // 为每个命令分配一个 AI 终端
  const terminals = await ensureAITerminals(connId, commands.length)
  console.log(`[AIToolExecutor] 分配 ${terminals.length} 个终端`)

  // 并行执行：每个命令在自己的终端中执行
  const results = await Promise.all(commands.map(async (cmd, index) => {
    const label = cmd.label || `命令 ${index + 1}`
    const sessionId = terminals[index] || terminals[terminals.length - 1]
    try {
      const output = await executeInTerminal(connId, sessionId, cmd.command)
      return `=== ${label} ===\n${output}`
    } catch (e) {
      return `=== ${label} ===\n执行失败: ${e.message}`
    }
  }))

  return results.join('\n\n')
}

// ==================== 在终端中执行命令 ====================

function executeInTerminal(connId, sessionId, command) {
  if (!sessionId) {
    console.log(`[AIToolExecutor] 无终端，回退到 ExecuteCommand`)
    return SSHService.ExecuteCommand(connId, command).then(res =>
      res?.success ? (res.stdout || '(无输出)') : (res?.stderr || res?.stdout || '(执行失败)')
    ).catch(e => `执行失败: ${e.message}`)
  }

  console.log(`[AIToolExecutor] 在终端 ${sessionId} 执行: ${command}`)

  return new Promise((resolve) => {
    let output = ''
    let resolved = false
    let timeoutId = null
    let commandSent = false
    let flushDone = false

    const handler = (event) => {
      if (resolved) return
      const d = event?.data
      if (!d || d.sessionID !== sessionId) return
      const text = typeof d === 'string' ? d : (d.data || '')
      if (typeof text !== 'string' || !text) return

      // 命令发送前的输出是初始输出（登录横幅等），丢弃
      if (!commandSent) return

      output += text

      // 复用结构化终端的提示符检测
      const clean = output.replace(ANSI_RE, '')
      const lines = clean.split('\n')
      for (let i = Math.max(0, lines.length - 3); i < lines.length; i++) {
        if (PROMPT_RE.test(lines[i].trim())) {
          cleanup()
          resolved = true
          resolve(cleanOutput(output, command))
          return
        }
      }
    }

    const cleanup = () => {
      if (timeoutId) clearTimeout(timeoutId)
      Events.Off('ssh:terminal-output', handler)
    }

    Events.On('ssh:terminal-output', handler)

    // 等待初始输出（登录横幅、提示符）清空后再发送命令
    setTimeout(() => {
      if (resolved) return
      commandSent = true
      flushDone = true

      // 通知终端创建块
      Events.Emit('ai:terminal-exec-start', { connId, sessionID: sessionId, command })

      SSHService.WriteToTerminalByID(connId, sessionId, command + '\n').catch(e => {
        cleanup()
        if (!resolved) {
          resolved = true
          resolve(`写入失败: ${e.message}`)
        }
      })
    }, 500)

    timeoutId = setTimeout(() => {
      if (!resolved) {
        cleanup()
        resolved = true
        resolve(cleanOutput(output, command) || '(超时，无输出)')
      }
    }, 30000)
  })
}

/** 清理输出：去 ANSI、去命令回显、去提示符 */
function cleanOutput(raw, command) {
  let s = raw.replace(ANSI_RE, '')
  // 去命令回显
  if (command) {
    s = s.replace(command + '\r\n', '')
    s = s.replace(command + '\n', '')
  }
  // 去提示符行
  const lines = s.split('\n')
  while (lines.length > 0 && PROMPT_RE.test(lines[lines.length - 1].trim())) {
    lines.pop()
  }
  return lines.join('\n').trim()
}
