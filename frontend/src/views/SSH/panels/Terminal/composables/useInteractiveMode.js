/**
 * 交互式模式管理
 * 检测交互式程序，管理输入转发策略
 */
import { ref, computed } from 'vue'

// 交互式程序列表
const INTERACTIVE_PROGRAMS = [
  // 编辑器
  'vim', 'vi', 'nvim', 'nano', 'emacs', 'ed',
  // 分页器
  'less', 'more', 'most',
  // 监控工具
  'top', 'htop', 'atop', 'btop', 'glances', 'iotop', 'iftop', 'nmon',
  // 交互式 Shell
  'bash', 'zsh', 'fish', 'sh', 'dash',
  // 数据库客户端
  'mysql', 'psql', 'mongo', 'redis-cli', 'sqlite3',
  // REPL
  'python', 'python3', 'node', 'irb', 'pry', 'ghci', 'scala', 'kotlin',
  // 其他交互式程序
  'tmux', 'screen', 'byobu',
  'mc', 'ranger', 'nnn', 'lf', 'vifm',
  'ssh', 'telnet', 'ftp', 'sftp',
  'aptitude', 'dialog', 'whiptail',
  'watch', 'tail -f',
  'ping', 'traceroute', 'mtr',
  'dd', 'pv',
  'gdb', 'lldb',
  'expect', 'unbuffer'
]

// 交互式程序模式
export const INTERACTIVE_MODE = {
  AUTO: 'auto',           // 自动检测
  STRUCTURED: 'structured', // 强制结构化
  CLASSIC: 'classic'      // 强制经典
}

export function useInteractiveMode() {
  // 当前模式
  const mode = ref(INTERACTIVE_MODE.AUTO)

  // 是否在交互式程序中
  const isInInteractive = ref(false)

  // 当前运行的程序
  const currentProgram = ref(null)

  // 输入缓冲（用于检测命令）
  let inputBuffer = ''
  let promptPattern = null

  // 计算属性
  const isStructuredMode = computed(() => {
    if (mode.value === INTERACTIVE_MODE.STRUCTURED) return true
    if (mode.value === INTERACTIVE_MODE.CLASSIC) return false
    return !isInInteractive.value
  })

  const isClassicMode = computed(() => !isStructuredMode.value)

  // 检测是否是交互式程序
  function isInteractiveProgram(command) {
    if (!command) return false

    const cmd = command.trim().split(/\s+/)[0]
    const baseCmd = cmd.split('/').pop() // 获取基础命令名

    return INTERACTIVE_PROGRAMS.some(prog => {
      // 精确匹配或前缀匹配
      return baseCmd === prog || baseCmd.startsWith(prog + ' ')
    })
  }

  // 分析命令并更新状态
  function analyzeCommand(command) {
    if (!command) return

    const cmd = command.trim().split(/\s+/)[0]
    const baseCmd = cmd.split('/').pop()

    // 检查是否是交互式程序
    if (isInteractiveProgram(command)) {
      enterInteractiveMode(baseCmd)
      return true
    }

    // 检查是否是退出命令
    if (isExitCommand(command)) {
      exitInteractiveMode()
      return true
    }

    return false
  }

  // 进入交互式模式
  function enterInteractiveMode(program) {
    isInInteractive.value = true
    currentProgram.value = program
    console.log('[InteractiveMode] 进入交互式模式:', program)
  }

  // 退出交互式模式
  function exitInteractiveMode() {
    isInInteractive.value = false
    currentProgram.value = null
    console.log('[InteractiveMode] 退出交互式模式')
  }

  // 检测是否是退出命令
  function isExitCommand(command) {
    if (!command) return false

    const cmd = command.trim().toLowerCase()
    const exitCommands = [
      'exit', 'quit', 'q', ':q', ':q!', ':wq', ':wq!',
      'zz', 'ctrl+c', 'ctrl+d', 'ctrl+z'
    ]

    return exitCommands.includes(cmd)
  }

  // 分析输出，检测程序退出
  function analyzeOutput(output) {
    if (!isInInteractive.value) return false

    // 检测常见的退出提示
    const exitPatterns = [
      /\[Process completed\]/,
      /\[exited\]/,
      /^logout/,
      /^exit$/,
      /\$ $/,  // Shell 提示符
      /^> $/,
      /^# $/
    ]

    for (const pattern of exitPatterns) {
      if (pattern.test(output)) {
        exitInteractiveMode()
        return true
      }
    }

    return false
  }

  // 设置模式
  function setMode(newMode) {
    mode.value = newMode
    console.log('[InteractiveMode] 模式切换:', newMode)
  }

  // 切换模式
  function toggleMode() {
    if (mode.value === INTERACTIVE_MODE.AUTO) {
      mode.value = isInInteractive.value ? INTERACTIVE_MODE.STRUCTURED : INTERACTIVE_MODE.CLASSIC
    } else if (mode.value === INTERACTIVE_MODE.STRUCTURED) {
      mode.value = INTERACTIVE_MODE.CLASSIC
    } else {
      mode.value = INTERACTIVE_MODE.AUTO
    }
    console.log('[InteractiveMode] 模式切换:', mode.value)
  }

  // 获取模式标签
  function getModeLabel() {
    const labels = {
      [INTERACTIVE_MODE.AUTO]: '自动',
      [INTERACTIVE_MODE.STRUCTURED]: '结构化',
      [INTERACTIVE_MODE.CLASSIC]: '经典'
    }
    return labels[mode.value] || '自动'
  }

  // 获取状态描述
  function getStatusDescription() {
    if (isInInteractive.value) {
      return `交互式程序: ${currentProgram.value}`
    }
    return `模式: ${getModeLabel()}`
  }

  // 添加交互式程序
  function addInteractiveProgram(program) {
    if (!INTERACTIVE_PROGRAMS.includes(program)) {
      INTERACTIVE_PROGRAMS.push(program)
    }
  }

  // 移除交互式程序
  function removeInteractiveProgram(program) {
    const index = INTERACTIVE_PROGRAMS.indexOf(program)
    if (index > -1) {
      INTERACTIVE_PROGRAMS.splice(index, 1)
    }
  }

  // 获取交互式程序列表
  function getInteractivePrograms() {
    return [...INTERACTIVE_PROGRAMS]
  }

  // 重置状态
  function reset() {
    mode.value = INTERACTIVE_MODE.AUTO
    isInInteractive.value = false
    currentProgram.value = null
    inputBuffer = ''
  }

  return {
    // 状态
    mode,
    isInInteractive,
    currentProgram,
    isStructuredMode,
    isClassicMode,

    // 方法
    isInteractiveProgram,
    analyzeCommand,
    analyzeOutput,
    enterInteractiveMode,
    exitInteractiveMode,
    setMode,
    toggleMode,
    getModeLabel,
    getStatusDescription,
    addInteractiveProgram,
    removeInteractiveProgram,
    getInteractivePrograms,
    reset
  }
}
