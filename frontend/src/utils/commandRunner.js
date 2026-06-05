/**
 * AI 命令执行器
 * AI 通过自己的终端 session 发送命令，监听该 session 的输出
 */

import * as SSHService from '../../bindings/changeme/ssh/sshservice.js'
import { Events } from '@wailsio/runtime'

// AI 终端 session ID
let aiSessionID = null

export function setAISessionID(id) {
  aiSessionID = id
}

/**
 * 通过 AI 终端 session 执行命令
 */
export async function runCommand(connId, command) {
  if (!aiSessionID) {
    return { success: false, output: 'AI 终端未初始化', exitCode: -1 }
  }

  console.log(`[CommandRunner] 执行: ${command} (session: ${aiSessionID})`)

  return new Promise((resolve) => {
    let output = ''
    let resolved = false
    let timeoutId = null

    const handler = (event) => {
      if (resolved) return
      const data = event?.data
      if (!data) return
      // 只接收 AI 终端 session 的输出
      if (data.sessionID !== aiSessionID) return
      const text = typeof data === 'string' ? data : (data.data || '')
      if (typeof text !== 'string' || !text) return
      output += text
    }

    const cleanup = () => {
      if (timeoutId) clearTimeout(timeoutId)
      Events.Off('ssh:terminal-output', handler)
    }

    Events.On('ssh:terminal-output', handler)

    // 通过 AI 终端 session 写入命令
    SSHService.WriteToTerminalByID(connId, aiSessionID, command + '\n').catch(e => {
      cleanup()
      if (!resolved) {
        resolved = true
        resolve({ success: false, output: `写入失败: ${e.message}`, exitCode: -1 })
      }
    })

    // 超时保护：等待足够时间让命令完成
    timeoutId = setTimeout(() => {
      if (!resolved) {
        cleanup()
        resolved = true
        const clean = output.replace(/\x1b\[[0-9;]*[a-zA-Z]/g, '').replace(/\x1b\].*?\x07/g, '').trim()
        resolve({ success: true, output: clean || '(超时，无输出)', exitCode: 0 })
      }
    }, 30000)
  })
}
