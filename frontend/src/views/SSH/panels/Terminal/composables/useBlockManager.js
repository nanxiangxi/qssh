/**
 * 终端块管理器
 * 基于 Shell 提示符检测实现命令边界分割
 */
import { ref, computed } from 'vue'

export const BLOCK_TYPE = { COMMAND: 'command', SYSTEM: 'system' }
export const BLOCK_STATUS = { RUNNING: 'running', SUCCESS: 'success', FAILED: 'failed', CANCELLED: 'cancelled' }

// Shell 提示符正则（匹配行尾的 $ 或 # 提示符）
// 匹配格式: "user@host:path$", "user@host:path$ ", "[root@host ~]#", "[root@host ~]# "
// 兼容 ANSI 转义序列
const PROMPT_RE = /[\$#]\s*$/
const ANY_PROMPT_RE = /[\$#]\s*$/
// ANSI 转义序列正则
const ANSI_RE = /\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g

export function useBlockManager() {
  const blocks = ref([])
  const activeBlock = ref(null)

  // 当前累积的终端输出（未分割的原始流）
  const rawOutput = ref('')

  const blockCount = computed(() => blocks.value.length)
  const commandBlocks = computed(() => blocks.value.filter(b => b.type === BLOCK_TYPE.COMMAND))
  const runningBlocks = computed(() => blocks.value.filter(b => b.status === BLOCK_STATUS.RUNNING))

  // 创建系统消息块
  function addSystemMessage(content) {
    const block = makeBlock(BLOCK_TYPE.SYSTEM, { content })
    blocks.value.push(block)
    return block
  }

  function makeBlock(type, data = {}) {
    return {
      id: `b_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`,
      type,
      timestamp: Date.now(),
      content: '',
      status: null,
      exitCode: null,
      duration: null,
      collapsed: false,
      command: '',
      ...data
    }
  }

  // 开始新命令
  function startCommand(command) {
    console.log('[BlockManager] startCommand:', command)
    // 结束上一个活跃的命令（保留原有状态：取消的块保持 CANCELLED）
    if (activeBlock.value) {
      const prev = activeBlock.value.status
      console.log('[BlockManager] 结束上一个块:', activeBlock.value.command, '状态:', prev)
      endActiveBlock(prev === BLOCK_STATUS.CANCELLED ? BLOCK_STATUS.CANCELLED : BLOCK_STATUS.SUCCESS, prev === BLOCK_STATUS.CANCELLED ? null : 0)
    }
    const block = makeBlock(BLOCK_TYPE.COMMAND, {
      command,
      content: '',
      status: BLOCK_STATUS.RUNNING,
      startTime: Date.now()
    })
    blocks.value.push(block)
    activeBlock.value = block
    console.log('[BlockManager] 新块已创建，ID:', block.id)
    return block
  }

  // 追加输出到当前活跃块，同时检测提示符
  function appendOutput(text) {
    rawOutput.value += text

    if (activeBlock.value) {
      activeBlock.value.content += text
      console.log('[BlockManager] appendOutput:', text.length, '字节, 块:', activeBlock.value.command, '内容长度:', activeBlock.value.content.length)
    } else {
      console.log('[BlockManager] appendOutput: 无活跃块，输出到 rawOutput')
    }
  }

  // 检测字符串末尾是否以 Shell 提示符结尾
  function endsWithPrompt(str) {
    if (!str) return false
    // 去除 ANSI 转义序列后再检测
    const clean = str.replace(ANSI_RE, '')
    const lines = clean.split('\n')
    // 检查最后 3 行是否包含提示符
    for (let i = Math.max(0, lines.length - 3); i < lines.length; i++) {
      const line = lines[i].trim()
      if (ANY_PROMPT_RE.test(line)) {
        return true
      }
    }
    return false
  }

  // 从字符串末尾去除一行提示符
  function stripTrailingPrompt(str) {
    if (!str) return str
    const lines = str.split('\n')
    if (lines.length === 0) return str
    const last = lines[lines.length - 1]
    if (ANY_PROMPT_RE.test(last)) {
      lines.pop() // 去掉提示符行
    }
    // 去掉末尾空行
    while (lines.length > 0 && lines[lines.length - 1].trim() === '') {
      lines.pop()
    }
    return lines.join('\n')
  }

  // 如果当前块末尾出现提示符则自动结束（保持原有状态）
  function endBlockIfPrompt() {
    if (!activeBlock.value) return false
    if (endsWithPrompt(activeBlock.value.content)) {
      const prev = activeBlock.value.status
      const st = prev === BLOCK_STATUS.RUNNING ? BLOCK_STATUS.SUCCESS : prev
      console.log('[BlockManager] endBlockIfPrompt: 检测到提示符，结束块:', activeBlock.value.command, '状态:', prev, '->', st)
      endActiveBlock(st, prev === BLOCK_STATUS.RUNNING ? 0 : null)
      return true
    }
    return false
  }

  // 结束活跃块
  function endActiveBlock(status, exitCode) {
    if (!activeBlock.value) return
    console.log('[BlockManager] endActiveBlock:', activeBlock.value.command, '状态:', status)

    activeBlock.value.status = status
    activeBlock.value.exitCode = exitCode
    activeBlock.value.endTime = Date.now()
    activeBlock.value.duration = activeBlock.value.endTime - activeBlock.value.startTime

    // 剥离末尾的提示符行
    let content = stripTrailingPrompt(activeBlock.value.content)

    // 剥离开头的命令回显（shell echo 回来的命令本身）
    // 格式: "command\n" 或 "command\r\n" 开头
    if (activeBlock.value.command && content) {
      const cmd = activeBlock.value.command
      if (content.startsWith(cmd + '\r\n')) {
        content = content.slice(cmd.length + 2)
      } else if (content.startsWith(cmd + '\n')) {
        content = content.slice(cmd.length + 1)
      }
      // 去掉开头的空行
      content = content.replace(/^[\r\n]+/, '')
    }

    // 取消的块去掉 ^C 回显及残留空行
    if (status === BLOCK_STATUS.CANCELLED && content) {
      content = content.replace(/\^C[\r\n]*/g, '')
    }

    activeBlock.value.content = content
    console.log('[BlockManager] 块已结束，内容长度:', content.length)
    activeBlock.value = null
  }

  // 手动结束某个块
  function endCommand(blockId, status = BLOCK_STATUS.SUCCESS, exitCode = 0) {
    const block = blocks.value.find(b => b.id === blockId)
    if (!block) return
    block.status = status
    block.exitCode = exitCode
    block.endTime = Date.now()
    block.duration = block.endTime - block.startTime
    if (activeBlock.value?.id === blockId) activeBlock.value = null
  }

  // Ctrl+C 取消当前命令（不释放活跃块，继续捕获后续输出和提示符）
  function cancelCommand() {
    if (activeBlock.value && activeBlock.value.status === BLOCK_STATUS.RUNNING) {
      activeBlock.value.status = BLOCK_STATUS.CANCELLED
      const stack = new Error().stack
      console.log('[BlockManager] cancelCommand:', activeBlock.value.command, '\n调用栈:', stack)
    }
  }

  // 折叠/展开
  function toggleCollapse(blockId) {
    const b = blocks.value.find(b => b.id === blockId)
    if (b) b.collapsed = !b.collapsed
  }

  function collapseAll() { blocks.value.forEach(b => b.collapsed = true) }
  function expandAll() { blocks.value.forEach(b => b.collapsed = false) }

  function removeBlock(blockId) {
    const i = blocks.value.findIndex(b => b.id === blockId)
    if (i !== -1) blocks.value.splice(i, 1)
  }

  function clearBlocks() {
    blocks.value = []
    activeBlock.value = null
    rawOutput.value = ''
  }

  function getBlockContent(blockId) {
    const b = blocks.value.find(b => b.id === blockId)
    if (!b) return null
    return { command: b.command, output: b.content, exitCode: b.exitCode, duration: b.duration }
  }

  function searchBlocks(keyword) {
    if (!keyword) return blocks.value
    const lower = keyword.toLowerCase()
    return blocks.value.filter(b =>
      b.command?.toLowerCase().includes(lower) ||
      b.content?.toLowerCase().includes(lower)
    )
  }

  function getFailedCommands() {
    return blocks.value.filter(b =>
      b.type === BLOCK_TYPE.COMMAND &&
      b.status === BLOCK_STATUS.FAILED
    )
  }

  function getStats() {
    const total = blocks.value.length
    const commands = blocks.value.filter(b => b.type === BLOCK_TYPE.COMMAND).length
    const failed = getFailedCommands().length
    const running = blocks.value.filter(b => b.status === BLOCK_STATUS.RUNNING).length
    return { total, commands, failed, running }
  }

  function exportAsText() {
    return blocks.value.map(b => {
      if (b.type === 'command') return `$ ${b.command}\n${b.content}`
      return `${b.content}`
    }).join('\n')
  }

  return {
    blocks, activeBlock, blockCount, commandBlocks, runningBlocks, rawOutput,
    startCommand, appendOutput, endCommand, endActiveBlock, endBlockIfPrompt,
    endsWithPrompt, stripTrailingPrompt, cancelCommand, addSystemMessage,
    toggleCollapse, collapseAll, expandAll,
    removeBlock, clearBlocks, getBlockContent,
    searchBlocks, getFailedCommands, getStats, exportAsText
  }
}
