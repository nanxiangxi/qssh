/**
 * AI 工具执行器
 * 命令通过终端 shell session 执行
 *
 * 核心原则：
 * 1. AI 不负责关闭终端，终端生命周期完全由用户管理
 * 2. 单命令：复用已有 AI 终端，没有则创建
 * 3. 多终端：复用已有终端 + 只补创缺少的，不关闭任何终端
 */

import { Events } from '@wailsio/runtime'
import { AIService } from '../../bindings/changeme/ai/index.js'
import * as SSHService from '../../bindings/changeme/ssh/sshservice.js'

let initialized = false
let creatingTerminal = false

// sessionId → connId（每个终端绑定到创建它的连接）
const terminalConnMap = new Map()

// AI 终端 sessionId 列表（按创建顺序）
let aiTerminals = []

// 用户选择的目标终端
let userTarget = null

export function initAIToolExecutor() {
  if (initialized) return
  initialized = true

  Events.On('ai:execute-tool', handleToolRequest)

  // AI 终端就绪
  Events.On('terminal:ai-session-ready', (e) => {
    const d = e?.data
    if (!d || !d.sessionId) return
    if (!aiTerminals.includes(d.sessionId)) {
      aiTerminals.push(d.sessionId)
    }
    if (d.connId) {
      terminalConnMap.set(d.sessionId, d.connId)
    }
    creatingTerminal = false
    console.log(`[AIToolExecutor] AI 终端就绪: ${d.sessionId} (共 ${aiTerminals.length} 个)`)
  })

  // 用户手动关闭终端时清理记录
  Events.On('terminal:session-closed', (e) => {
    const d = e?.data
    if (!d) return
    aiTerminals = aiTerminals.filter(id => id !== d.sessionId)
    terminalConnMap.delete(d.sessionId)
    if (userTarget === d.sessionId) userTarget = null
    console.log(`[AIToolExecutor] 终端关闭: ${d.sessionId}, AI终端剩余: ${aiTerminals.length}`)
  })

  console.log('[AIToolExecutor] 已初始化')
}

export function destroyAIToolExecutor() {
  Events.Off('ai:execute-tool', handleToolRequest)
  Events.Off('terminal:ai-session-ready')
  Events.Off('terminal:session-closed')
  initialized = false
}

export function setTargetTerminal(sessionId) {
  userTarget = sessionId || null
}

export function getTerminalList() {
  return aiTerminals.map(id => ({ sessionId: id, connId: terminalConnMap.get(id), isAI: true }))
}

// ==================== 终端管理 ====================

/** 获取指定连接的所有可用 AI 终端 */
async function getAvailableAITerminals(targetConnId) {
  try {
    const available = await SSHService.GetSessionIDs(targetConnId)
    // 清理已失效的
    aiTerminals = aiTerminals.filter(id => available.includes(id))
    // 返回属于该连接的
    return aiTerminals.filter(id => {
      return available.includes(id) && terminalConnMap.get(id) === targetConnId
    })
  } catch (e) {
    return aiTerminals.filter(id => terminalConnMap.get(id) === targetConnId)
  }
}

/** 获取一个可用的 AI 终端（复用优先） */
async function getOneAITerminal(targetConnId) {
  const terminals = await getAvailableAITerminals(targetConnId)
  return terminals.length > 0 ? terminals[terminals.length - 1] : null
}

/** 创建新的 AI 终端，返回 sessionId */
async function createAITerminal(targetConnId) {
  if (creatingTerminal) {
    for (let i = 0; i < 50; i++) {
      await new Promise(r => setTimeout(r, 100))
      if (!creatingTerminal) {
        const connTerminals = aiTerminals.filter(id => terminalConnMap.get(id) === targetConnId)
        return connTerminals.length > 0 ? connTerminals[connTerminals.length - 1] : null
      }
    }
    return null
  }

  creatingTerminal = true
  const prevCount = aiTerminals.length
  Events.Emit('ai:create-terminal', { connId: targetConnId })

  for (let i = 0; i < 50; i++) {
    await new Promise(r => setTimeout(r, 100))
    if (aiTerminals.length > prevCount) {
      const newTerminal = aiTerminals[aiTerminals.length - 1]
      if (!terminalConnMap.has(newTerminal)) {
        terminalConnMap.set(newTerminal, targetConnId)
      }
      return newTerminal
    }
  }

  creatingTerminal = false
  return null
}

/** 验证 session 是否存在 */
async function verifySession(connId, sid) {
  try {
    return await SSHService.IsShellSessionActive(connId, sid)
  } catch (e) {
    return false
  }
}

/** 获取或创建一个 AI 终端（优先复用） */
async function getOrCreateOneTerminal(connId) {
  const existing = await getOneAITerminal(connId)
  if (existing) return existing
  return await createAITerminal(connId)
}

// ==================== 命令执行 ====================

/** 向终端发送命令并等待结果（marker 方式） */
function executeOnTerminal(connId, sid, command) {
  return new Promise((resolve) => {
    const marker = `__AI_${Date.now()}_${Math.random().toString(36).slice(2, 6)}__`
    const markerRe = new RegExp(marker.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
    let output = ''
    let resolved = false
    let timeoutId = null

    const handler = (ev) => {
      if (resolved) return
      const d = ev?.data
      if (!d || d.sessionID !== sid) return
      const text = typeof d === 'string' ? d : (d.data || '')
      if (typeof text !== 'string' || !text) return
      output += text
      const clean = output.replace(/\x1b\[[0-9;]*[a-zA-Z]/g, '').replace(/\x1b\].*?\x07/g, '')
      if (markerRe.test(clean)) {
        cleanup()
        resolved = true
        resolve(cleanOutput(output))
      }
    }

    const cleanup = () => {
      if (timeoutId) clearTimeout(timeoutId)
      Events.Off('ssh:terminal-output', handler)
    }

    Events.On('ssh:terminal-output', handler)

    SSHService.WriteToTerminalByID(connId, sid, command + '\n').then(() => {
      setTimeout(() => {
        if (!resolved) {
          SSHService.WriteToTerminalByID(connId, sid, `echo "${marker}"\n`).catch(() => {})
        }
      }, 200)
    }).catch(e => {
      cleanup()
      if (!resolved) {
        resolved = true
        resolve(`写入终端失败: ${e.message}`)
      }
    })

    timeoutId = setTimeout(() => {
      if (!resolved) {
        cleanup()
        resolved = true
        resolve(cleanOutput(output) || '(超时)')
      }
    }, 30000)
  })
}

// ==================== 工具请求处理 ====================

async function handleToolRequest(event) {
  const data = event?.data
  if (!data) return

  const { connId, callId, tool, command, args } = data
  console.log(`[AIToolExecutor] 工具请求: ${tool} → ${command || ''} (connId: ${connId})`)

  let result = ''
  try {
    switch (tool) {
      case 'execute_command':
      case 'get_server_info':
      case 'get_system_status':
        result = await execCommand(connId, tool, command)
        break
      case 'open_terminal':
        const sid = await createAITerminal(connId)
        result = sid ? 'AI 终端已创建' : '创建终端失败'
        break
      case 'close_terminal':
        result = '终端关闭由用户操作，AI 不再自动关闭终端'
        break
      case 'multi_execute':
        result = await handleMultiExecute(connId, args)
        break
      default:
        result = `未知工具: ${tool}`
    }
  } catch (e) {
    result = `执行失败: ${e.message}`
  }

  try {
    await AIService.SubmitToolResult(callId, result)
  } catch (e) {
    console.error('[AIToolExecutor] 提交结果失败:', e)
  }
}

/**
 * 单命令执行
 * 逻辑：用户指定终端 → 直接用；AI自动 → 复用已有AI终端 → 没有则创建
 */
async function execCommand(connId, tool, command) {
  let actualCommand = command
  if (tool === 'get_server_info') {
    actualCommand = 'echo "=== 主机名 ===" && hostname && echo "=== 系统信息 ===" && uname -a && echo "=== 运行时间 ===" && uptime && echo "=== 内核版本 ===" && cat /proc/version 2>/dev/null || uname -r'
  } else if (tool === 'get_system_status') {
    actualCommand = 'echo "=== CPU ===" && top -bn1 | head -5 2>/dev/null && echo "" && echo "=== 内存 ===" && free -h 2>/dev/null && echo "" && echo "=== 磁盘 ===" && df -h 2>/dev/null && echo "" && echo "=== 负载 ===" && cat /proc/loadavg 2>/dev/null'
  }

  let sid = null

  // 1. 用户指定了终端 → 直接使用
  if (userTarget) {
    const valid = await verifySession(connId, userTarget)
    if (valid) {
      sid = userTarget
    } else {
      userTarget = null
    }
  }

  // 2. AI自动 → 复用已有AI终端
  if (!sid) {
    sid = await getOneAITerminal(connId)
  }

  // 3. 没有可用终端 → 创建新的AI终端
  if (!sid) {
    sid = await createAITerminal(connId)
    if (!sid) return '错误：无法创建终端（可能已达到SSH会话上限，请关闭一些终端后重试）'
  }

  console.log(`[AIToolExecutor] 执行: ${actualCommand} (session: ${sid})`)

  return executeOnTerminal(connId, sid, actualCommand)
}

/**
 * 多终端并行执行
 * 逻辑：复用已有终端 + 只补创缺少的 → 并行执行 → 不关闭终端
 * 终端生命周期完全由用户管理
 */
async function handleMultiExecute(connId, argsJSON) {
  let commands = []
  try {
    const args = typeof argsJSON === 'string' ? JSON.parse(argsJSON) : argsJSON
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

  console.log(`[AIToolExecutor] 多终端并行执行: ${commands.length} 个命令`)

  // 1. 获取已有 AI 终端
  const existingTerminals = await getAvailableAITerminals(connId)
  console.log(`[AIToolExecutor] 已有 ${existingTerminals.length} 个 AI 终端`)

  // 2. 计算需要补创的数量
  const needCreate = Math.max(0, commands.length - existingTerminals.length)

  // 3. 补创缺少的终端
  const newTerminals = []
  for (let i = 0; i < needCreate; i++) {
    const sid = await createAITerminal(connId)
    if (!sid) {
      return `错误：无法创建第 ${i + 1} 个终端`
    }
    newTerminals.push(sid)
    await new Promise(r => setTimeout(r, 200))
  }

  // 4. 合并：已有终端 + 新终端
  const allTerminals = [...existingTerminals, ...newTerminals]

  // 5. 分配命令到终端
  const assignments = commands.map((cmd, i) => ({
    sessionId: allTerminals[i],
    label: cmd.label || `命令${i + 1}`,
    command: cmd.command
  }))

  // 6. 并行执行
  const results = await Promise.all(
    assignments.map(t => executeOnTerminal(connId, t.sessionId, t.command))
  )

  // 7. 汇总结果（不关闭终端）
  const summary = assignments.map((t, i) => {
    return `=== ${t.label} ===\n${results[i] || '(无输出)'}`
  }).join('\n\n')

  console.log(`[AIToolExecutor] 多终端执行完成，终端保留供用户使用`)
  return summary
}

function cleanOutput(text) {
  return text.replace(/\x1b\[[0-9;]*[a-zA-Z]/g, '').replace(/\x1b\].*?\x07/g, '').trim()
}
