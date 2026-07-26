/**
 * cd 命令同步 composable
 * 检测终端中的 cd 命令，通过后端独立 SSH session 解析真实路径
 * 不向终端注入任何命令，零副作用
 */
import { ref } from 'vue'
import { Events } from '@wailsio/runtime'
import { SSHService } from '../../../../../../bindings/changeme/ssh/index.js'

export function useCdSync(connId) {
  const terminalCwd = ref('')

  // 检测 cd 命令并通过后端解析真实路径
  async function detectCd(command) {
    const trimmed = command.trim()
    const isBareCd = trimmed === 'cd'
    const cdMatch = trimmed.match(/^cd\s+(.+)$/)
    if (!isBareCd && !cdMatch) return

    let targetPath
    if (isBareCd) {
      targetPath = '~'
    } else {
      targetPath = cdMatch[1].trim()
      // 去掉引号
      if ((targetPath.startsWith('"') && targetPath.endsWith('"')) ||
          (targetPath.startsWith("'") && targetPath.endsWith("'"))) {
        targetPath = targetPath.slice(1, -1)
      }
      // 忽略 cd - 和纯选项
      if (targetPath === '-' || (targetPath.startsWith('-') && !targetPath.startsWith('/'))) return
    }

    try {
      const currentDir = terminalCwd.value || '~'
      const realPath = await SSHService.ResolveCdPath(connId, currentDir, targetPath)
      const p = (realPath || '').trim()
      if (p && p.startsWith('/') && p !== terminalCwd.value) {
        terminalCwd.value = p
        Events.Emit('terminal:cd', { connId, path: p })
      }
    } catch (e) {
      console.warn('[CdSync] ResolveCdPath 失败:', e)
    }
  }

  // 初始化 home 目录（连接成功后调用）
  async function initHomeDir() {
    try {
      const homeResult = await SSHService.RunCommand(connId, 'pwd')
      terminalCwd.value = homeResult.trim()
    } catch (e) {
      terminalCwd.value = '/root'
    }
  }

  return { terminalCwd, detectCd, initHomeDir }
}
