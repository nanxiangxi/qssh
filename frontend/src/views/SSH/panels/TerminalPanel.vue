<template><div class="terminal-panel" ref="terminalContainer" @contextmenu.prevent="onContextMenu">
    <div ref="terminalElement" class="terminal-container"></div>
    <!-- 终端右键菜单 -->
    <Teleport to="body">
      <div v-if="ctxMenu.show" class="term-ctx-mask" @mousedown="ctxMenu.show=false">
        <div class="term-ctx-menu" :style="{top:ctxMenu.y+'px', left:ctxMenu.x+'px'}" @mousedown.stop>
          <div class="term-ctx-item" @click="doCopy">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            复制
            <span class="term-ctx-key">Ctrl+Shift+C</span>
          </div>
          <div class="term-ctx-item" @click="doPaste">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1"/></svg>
            粘贴
            <span class="term-ctx-key">Ctrl+Shift+V</span>
          </div>
          <div class="term-ctx-sep"></div>
          <div class="term-ctx-item" @click="doSelectAll">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 11V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h6"/><path d="M12 12H3M12 12V3"/></svg>
            全选
            <span class="term-ctx-key">Ctrl+Shift+A</span>
          </div>
          <div class="term-ctx-sep"></div>
          <div class="term-ctx-item" @click="doClear">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            清屏
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted, onUnmounted, inject, nextTick } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import 'xterm/css/xterm.css'
import { SSHService } from '../../../../bindings/changeme/ssh/index.js'
import { useTerminalSessionStore } from '../../../stores/terminalSessions'
import { Events } from '@wailsio/runtime'
import { useCommandHistoryStore } from '../../../stores/commandHistory'
import { commandSecurityAnalyzer } from '../../../utils/commandSecurityAnalyzer'
import { logTerminalCommand } from '@/utils/logger'

// 从父组件注入 connId
const connId = inject('connId')

// dockview 面板参数（包含 sessionId, isAI 等）
const props = defineProps({
  params: { type: Object, default: () => ({}) }
})

// 使用命令历史 Store
const commandHistoryStore = useCommandHistoryStore()

const terminalContainer = ref(null)

// 终端右键菜单
const ctxMenu = reactive({ show: false, x: 0, y: 0 })
const onContextMenu = (e) => {
  ctxMenu.x = e.clientX
  ctxMenu.y = e.clientY
  ctxMenu.show = true
}
const doCopy = () => {
  if (terminal?.hasSelection()) {
    navigator.clipboard.writeText(terminal.getSelection())
  }
  ctxMenu.show = false
}
const doPaste = async () => {
  try {
    const text = await navigator.clipboard.readText()
    if (text && connID && sessionID) {
      SSHService.WriteToTerminalByID(connID, sessionID, text)
    }
  } catch (e) {}
  ctxMenu.show = false
}
const doSelectAll = () => {
  terminal?.selectAll()
  ctxMenu.show = false
}
const doClear = () => {
  terminal?.clear()
  ctxMenu.show = false
}

// 🚨 显示危险命令警告对话框
const showDangerousCommandWarning = async (command, analysis) => {
  return new Promise((resolve) => {
    // 创建模态对话框
    const overlay = document.createElement('div')
    overlay.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(0, 0, 0, 0.85);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 99999;
      backdrop-filter: blur(4px);
    `
    
    const dialog = document.createElement('div')
    const riskColor = commandSecurityAnalyzer.getRiskColor(analysis.riskLevel)
    dialog.style.cssText = `
      background: #1e1e1e;
      border: 2px solid ${riskColor};
      border-radius: 12px;
      padding: 28px;
      max-width: 650px;
      width: 90%;
      box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
      animation: slideIn 0.3s ease-out;
    `
    
    const title = document.createElement('h3')
    title.style.cssText = `color: ${riskColor}; margin: 0 0 20px 0; font-size: 22px; display: flex; align-items: center; gap: 10px;`
    title.innerHTML = `${commandSecurityAnalyzer.getRiskIcon(analysis.riskLevel)} 安全警告 - ${analysis.category}`
    
    const cmdBox = document.createElement('div')
    cmdBox.style.cssText = `
      background: #2d2d2d;
      padding: 16px;
      border-radius: 6px;
      font-family: 'Courier New', monospace;
      color: #fc8181;
      margin-bottom: 20px;
      word-break: break-all;
      font-size: 14px;
      border-left: 4px solid ${riskColor};
    `
    cmdBox.textContent = command
    
    const warningsDiv = document.createElement('div')
    warningsDiv.style.cssText = 'margin-bottom: 20px;'
    
    analysis.warnings.forEach(warning => {
      const p = document.createElement('p')
      p.style.cssText = `color: #feb2b2; margin: 8px 0; font-size: 14px; line-height: 1.6;`
      p.textContent = warning
      warningsDiv.appendChild(p)
    })
    
    const suggestionsDiv = document.createElement('div')
    if (analysis.suggestions && analysis.suggestions.length > 0) {
      suggestionsDiv.style.cssText = 'background: #2d3748; padding: 16px; border-radius: 6px; margin-bottom: 24px;'
      const label = document.createElement('div')
      label.style.cssText = 'color: #90cdf4; font-weight: bold; margin-bottom: 10px; font-size: 14px;'
      label.textContent = '💡 安全建议：'
      suggestionsDiv.appendChild(label)
      
      analysis.suggestions.forEach(suggestion => {
        const p = document.createElement('p')
        p.style.cssText = 'color: #e2e8f0; margin: 6px 0; font-size: 13px; line-height: 1.5;'
        p.textContent = suggestion
        suggestionsDiv.appendChild(p)
      })
    }
    
    const buttonGroup = document.createElement('div')
    buttonGroup.style.cssText = 'display: flex; gap: 12px; justify-content: flex-end;'
    
    const cancelBtn = document.createElement('button')
    cancelBtn.textContent = '❌ 取消（推荐）'
    cancelBtn.style.cssText = `
      padding: 12px 24px;
      background: #4a5568;
      color: white;
      border: none;
      border-radius: 6px;
      cursor: pointer;
      font-size: 14px;
      font-weight: 500;
      transition: all 0.2s;
    `
    cancelBtn.onmouseover = () => cancelBtn.style.background = '#2d3748'
    cancelBtn.onmouseout = () => cancelBtn.style.background = '#4a5568'
    cancelBtn.onclick = () => {
      document.body.removeChild(overlay)
      resolve(false)
    }
    
    const confirmBtn = document.createElement('button')
    confirmBtn.textContent = '⚠️ 我确认要执行'
    confirmBtn.style.cssText = `
      padding: 12px 24px;
      background: ${riskColor};
      color: white;
      border: none;
      border-radius: 6px;
      cursor: pointer;
      font-size: 14px;
      font-weight: bold;
      transition: all 0.2s;
    `
    confirmBtn.onmouseover = () => confirmBtn.style.opacity = '0.9'
    confirmBtn.onmouseout = () => confirmBtn.style.opacity = '1'
    confirmBtn.onclick = () => {
      document.body.removeChild(overlay)
      resolve(true)
    }
    
    buttonGroup.appendChild(cancelBtn)
    buttonGroup.appendChild(confirmBtn)
    
    dialog.appendChild(title)
    dialog.appendChild(cmdBox)
    dialog.appendChild(warningsDiv)
    if (analysis.suggestions && analysis.suggestions.length > 0) {
      dialog.appendChild(suggestionsDiv)
    }
    dialog.appendChild(buttonGroup)
    
    overlay.appendChild(dialog)
    document.body.appendChild(overlay)
    
    // ESC 键取消
    const handleEsc = (e) => {
      if (e.key === 'Escape') {
        document.removeEventListener('keydown', handleEsc)
        document.body.removeChild(overlay)
        resolve(false)
      }
    }
    document.addEventListener('keydown', handleEsc)
  })
}

const terminalElement = ref(null)

let terminal = null
let fitAddon = null
let connID = null
let sessionID = null
let isAI = false
let shellStarted = false

// 从面板参数获取 session（DockviewLayout 创建面板时已创建 session 并传入）
// dockview-vue 的 VueRenderer.init() 会把 params 包装为 { params: { params: ..., api: ... } }
// 所以实际的 addPanel params 在 props.params.params 里
const sessionStore = useTerminalSessionStore()

// 同步获取 sessionId（在 setup 阶段立即执行）
const dockviewParams = props.params?.params || props.params
console.log('[TerminalPanel] dockviewParams:', JSON.stringify(dockviewParams))
if (dockviewParams?.sessionId) {
  sessionID = dockviewParams.sessionId
  isAI = dockviewParams.isAI || false
  console.log('[TerminalPanel] 从参数获取 session:', sessionID, 'isAI:', isAI)
} else {
  // 降级：从 store 认领
  const sess = sessionStore.claimNextSession()
  if (sess) {
    sessionID = sess.sessionId
    isAI = sess.isAI || false
    console.log('[TerminalPanel] 从 store 认领 session:', sessionID, 'isAI:', isAI)
  } else {
    console.warn('[TerminalPanel] 无法获取 session，params:', props.params)
  }
}
let isStarting = false
let cleanupOutputListener = null

// 设置终端输出监听器（按 sessionID 过滤）
const setupOutputListener = () => {
  if (cleanupOutputListener) {
    cleanupOutputListener()
    cleanupOutputListener = null
  }
  if (!connID || connID === 'default-connection') return

  const handler = (event) => {
    const data = event?.data
    if (!data) return
    // 按 sessionID 过滤，只接收当前终端的输出
    if (data.sessionID && data.sessionID !== sessionID) return
    if (data.connID && data.connID !== connID) return
    const text = typeof data === 'string' ? data : (data.data || '')
    if (typeof text === 'string' && text && terminal) {
      terminal.write(text)
    }
  }
  Events.On('ssh:terminal-output', handler)
  cleanupOutputListener = () => Events.Off('ssh:terminal-output', handler)
}

// 初始化终端
const initTerminal = () => {
  if (!terminalElement.value) return

  // 使用注入的 connId
  connID = connId
  console.log('[TerminalPanel] 🎬 开始初始化终端, connId:', connID, '(来源: inject)')
  console.log('[TerminalPanel] 📦 terminalElement:', terminalElement.value)
  console.log('[TerminalPanel] 📦 terminalContainer:', terminalContainer.value)

  if (!connID) {
    console.error('[TerminalPanel] ❌ 没有获取到 connId!')
    terminal.write('\r\n❌ 错误: 无法获取连接 ID\r\n')
    return
  }

  // 创建终端实例
  console.log('[TerminalPanel] 🔨 创建 Terminal 实例')
  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: '"Cascadia Code", "Fira Code", Consolas, "Courier New", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#d4d4d4',
      selectionBackground: '#264f78',
      black: '#000000',
      red: '#cd3131',
      green: '#0dbc79',
      yellow: '#e5e510',
      blue: '#2472c8',
      magenta: '#bc3fbc',
      cyan: '#11a8cd',
      white: '#e5e5e5',
      brightBlack: '#666666',
      brightRed: '#f14c4c',
      brightGreen: '#23d18b',
      brightYellow: '#f5f543',
      brightBlue: '#3b8eea',
      brightMagenta: '#d670d6',
      brightCyan: '#29b8db',
      brightWhite: '#ffffff'
    }
  })
  console.log('[TerminalPanel] ✅ Terminal 实例创建成功')

  // 添加插件
  fitAddon = new FitAddon()
  const webLinksAddon = new WebLinksAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(webLinksAddon)
  console.log('[TerminalPanel] ✅ 插件加载完成')

  // 打开终端
  console.log('[TerminalPanel] 📺 打开终端到 DOM 元素:', terminalElement.value)
  terminal.open(terminalElement.value)
  console.log('[TerminalPanel] ✅ 终端已打开')
  
  fitAddon.fit()
  console.log('[TerminalPanel] 📐 初始 fit 完成, 尺寸:', terminal.cols, 'x', terminal.rows)

  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)

  // 如果已经有有效的connID，立即启动Shell
  if (connID && connID !== 'default-connection') {
    startShellSession()
  } else {
    terminal.write('\r\n⏳ 等待连接...\r\n')
  }

  // 监听用户输入 - 捕获所有操作以确保安全审计
  let currentInput = ''
  
  // 🔍 额外监控：通过终端输出反向推断执行的命令
  let lastPromptPosition = 0
  
  terminal.onData(async data => {
    const charCode = data.charCodeAt(0)
    
    // 检测回车键 - 命令结束
    if (data === '\r' || data === '\n') {
      console.log('[TerminalPanel] ✅ 检测到回车，记录命令:', currentInput.trim())
      if (currentInput.trim()) {
        const command = currentInput.trim()
        
        // 🚨 使用专业安全分析器检查命令
        const analysis = commandSecurityAnalyzer.analyzeCommand(command)
        
        if (analysis.isDangerous) {
          console.warn('[TerminalPanel] 🚨 检测到危险命令:', analysis.category, '风险等级:', analysis.riskLevel)
          
          // 显示强警告对话框（需要用户确认）
          const confirmed = await showDangerousCommandWarning(command, analysis)
          
          if (!confirmed) {
            console.log('[TerminalPanel] ❌ 用户取消了危险命令执行')
            currentInput = ''
            return  // 阻止命令发送
          }
          
          console.log('[TerminalPanel] ⚠️ 用户确认执行危险命令')
        }
        
        // 🎯 使用前端 Store 记录命令（不依赖后端）
        commandHistoryStore.addCommand(connID, command)
        console.log('[TerminalPanel] ✅ 命令已记录到前端 Store')
        
        // 📝 记录命令执行日志
        logTerminalCommand(connID, command)
        
        // ✅ 只在回车时才发送到后端
        if (connID && connID !== 'default-connection') {
          SSHService.WriteToTerminalByID(connID, sessionID, data)
            .catch(err => console.error('[TerminalPanel] 发送数据失败:', err))
        }
        
        currentInput = ''
      } else {
        // 空命令也要发送回车
        if (connID && connID !== 'default-connection') {
          SSHService.WriteToTerminalByID(connID, sessionID, data)
            .catch(err => console.error('[TerminalPanel] 发送数据失败:', err))
        }
      }
    } else if (charCode === 27) {
      // ESC 键 - 可能是用户按下，也可能是 ANSI 转义序列的开始
      // 如果是单独的 ESC (data.length === 1)，则是用户输入
      // 如果后面跟着其他字符，则是 ANSI 序列（由终端输出产生）
      
      // 只处理单独的 ESC 键（用户主动按下）
      if (data.length === 1) {
        console.log('[TerminalPanel] 🎯 用户按下了 ESC 键')
        // 可以选择清空当前输入或其他操作
        // currentInput = ''
      }
      // ANSI 序列不打印日志，让它自然发送到后端
      
      // ESC 序列也要发送到后端
      if (connID && connID !== 'default-connection') {
        SSHService.WriteToTerminalByID(connID, sessionID, data)
          .catch(err => console.error('[TerminalPanel] 发送数据失败:', err))
      }
    } else if (charCode === 127 || charCode === 8) {
      // 退格键 (DEL=127, BS=8)，删除最后一个字符
      currentInput = currentInput.slice(0, -1)
      console.log('[TerminalPanel] ⬅️ 退格，当前输入:', currentInput)
      // 退格也要发送到后端
      if (connID && connID !== 'default-connection') {
        SSHService.WriteToTerminalByID(connID, sessionID, data)
          .catch(err => console.error('[TerminalPanel] 发送数据失败:', err))
      }
    } else if (charCode === 3) {
      // Ctrl+C - 中断命令，记录为特殊操作
      console.log('[TerminalPanel] 🛑 检测到 Ctrl+C')
      if (currentInput.trim()) {
        commandHistoryStore.addCommand(connID, currentInput.trim() + ' [Ctrl+C]')
        currentInput = ''
      }
      // Ctrl+C 也要发送到后端
      if (connID && connID !== 'default-connection') {
        SSHService.WriteToTerminalByID(connID, sessionID, data)
          .catch(err => console.error('[TerminalPanel] 发送数据失败:', err))
      }
    } else if (charCode === 4) {
      // Ctrl+D - EOF，可能退出 shell
      console.log('[TerminalPanel] 🚪 检测到 Ctrl+D')
      if (currentInput.trim()) {
        commandHistoryStore.addCommand(connID, currentInput.trim() + ' [Ctrl+D]')
        currentInput = ''
      }
      // Ctrl+D 也要发送到后端
      if (connID && connID !== 'default-connection') {
        SSHService.WriteToTerminalByID(connID, sessionID, data)
          .catch(err => console.error('[TerminalPanel] 发送数据失败:', err))
      }
    } else if (charCode >= 32 || data === '\t') {
      // 记录所有可见字符、制表符和特殊符号
      // 包括：:(){ }|&;<>$`"'!#~[]\ 等所有可能的命令字符
      currentInput += data
      
      console.log('[TerminalPanel] ➕ 添加字符:', JSON.stringify(data), '当前输入:', JSON.stringify(currentInput))
      
      // ✅ 普通字符也要发送到后端（实时显示）
      if (connID && connID !== 'default-connection') {
        SSHService.WriteToTerminalByID(connID, sessionID, data)
          .catch(err => console.error('[TerminalPanel] 发送数据失败:', err))
      }
    } else if (charCode < 32 && charCode !== 10 && charCode !== 13) {
      // 其他控制字符（除了 \n 和 \r）
      // 这些可能是快捷键或特殊操作，也应该被记录
      console.log('[TerminalPanel] ℹ️ 检测到控制字符:', charCode, data)
      // 控制字符也要发送到后端
      if (connID && connID !== 'default-connection') {
        SSHService.WriteToTerminalByID(connID, sessionID, data)
          .catch(err => console.error('[TerminalPanel] 发送数据失败:', err))
      }
    }
    // 注意：\n (10) 和 \r (13) 已在上面处理
  })

  // 注册终端输出监听器
  setupOutputListener()
}

// 启动Shell会话
const startShellSession = async () => {
  console.log('[TerminalPanel] 🚀 尝试启动 Shell 会话, connID:', connID, 'isStarting:', isStarting)
  
  // 防止重复调用
  if (isStarting) {
    console.log('[TerminalPanel] ⚠️ Shell正在启动中，跳过')
    return
  }
  
  if (!connID || connID === 'default-connection') {
    console.warn('[TerminalPanel] ❌ 没有有效的连接ID')
    terminal.write('\r\n⏳ 等待连接建立...\r\n')
    return
  }

  isStarting = true
  console.log('[TerminalPanel] 🔄 调用 SSHService.StartShellSession...')

  try {
    await SSHService.StartShellSessionWithID(connID, sessionID)
    console.log('[TerminalPanel] ✅ Shell会话已启动, sessionID:', sessionID, 'isAI:', isAI)
    isStarting = false
    shellStarted = true
    // 标记会话就绪
    sessionStore.markSessionReady(sessionID)
    Events.Emit('terminal:session-ready', { connId: connID, sessionId: sessionID, panelId: terminal.element?.id || '', isAI })
  } catch (error) {
    console.error('[TerminalPanel] ❌ 启动Shell会话失败:', error)
    isStarting = false

    const errStr = String(error?.message || error)
    if (errStr.includes('连接不存在')) {
      terminal.write('\r\n⏳ 正在建立连接...\r\n')
      setTimeout(async () => {
        if (!isStarting) {
          try {
            await SSHService.StartShellSessionWithID(connID, sessionID)
          } catch (e) {
            terminal.write(`\r\n❌ 连接失败: ${e}\r\n`)
          }
        }
      }, 500)
    } else if (errStr.includes('administratively prohibited') || errStr.includes('open failed')) {
      terminal.write('\r\n❌ SSH 服务器会话数已达上限，请关闭其他终端后重试\r\n')
    } else {
      terminal.write(`\r\n❌ 启动Shell失败: ${errStr}\r\n`)
    }
  }
}

// 处理窗口大小变化
let resizeTimeout = null
let resizeObserver = null

const handleResize = () => {
  if (resizeTimeout) clearTimeout(resizeTimeout)
  resizeTimeout = setTimeout(() => {
    if (fitAddon && terminal && terminalContainer.value) {
      const { clientWidth, clientHeight } = terminalContainer.value
      // 容器有实际尺寸时才 fit（避免面板移动过程中触发无效调整）
      if (clientWidth > 0 && clientHeight > 0) {
        fitAddon.fit()
        if (connID && sessionID) {
          SSHService.ResizeTerminalByID(connID, sessionID, terminal.cols, terminal.rows)
            .catch(() => {})
        }
      }
    }
    resizeTimeout = null
  }, 100)
}

onMounted(() => {
  initTerminal()

  // 使用 ResizeObserver 监听容器尺寸变化
  // 比 window.resize 更可靠：能捕获 dockview 面板移动、分割等布局变化
  if (terminalContainer.value) {
    resizeObserver = new ResizeObserver(() => handleResize())
    resizeObserver.observe(terminalContainer.value)
  }

  Events.On('ai:terminal-display', handleTerminalDisplay)
})

// 处理 AI 工具的终端显示（模拟用户输入，复刻终端样式）
const handleTerminalDisplay = (event) => {
  const data = event?.data
  if (!data || data.connId !== connId || !terminal) return

  const cmd = data.command || ''
  const output = data.output || ''

  // 新行 + 命令（模拟用户输入）
  terminal.write(`\r\n${cmd}\r\n`)
  // 输出
  if (output) {
    terminal.write(output.replace(/\n/g, '\r\n'))
  }
  // 空行分隔
  terminal.write('\r\n')
}

onUnmounted(() => {
  // 清理 resize 防抖定时器
  if (resizeTimeout) clearTimeout(resizeTimeout)
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }

  if (cleanupOutputListener) {
    cleanupOutputListener()
  }

  Events.Off('ai:terminal-display', handleTerminalDisplay)

  // 通知其他模块终端已关闭
  Events.Emit('terminal:session-closed', { connId: connID, sessionId: sessionID })

  if (terminal) {
    terminal.dispose()
  }
  isStarting = false
})
</script>

<style scoped>
.terminal-panel {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  overflow: hidden;
  box-sizing: border-box;
}

.terminal-container {
  flex: 1;
  min-height: 0;
  width: 100%;
  padding: 4px;
  box-sizing: border-box;
  overflow: hidden;
}

/* xterm.js 样式覆盖 */
:deep(.xterm) {
  height: 100%;
  width: 100%;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;

  &::-webkit-scrollbar {
    width: 3px;
    transition: width 0.25s;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: linear-gradient(180deg, rgba(99, 179, 237, 0.15), rgba(66, 153, 225, 0.3));
    border-radius: 999px;
    transition: all 0.25s;
  }

  &::-webkit-scrollbar-thumb:hover {
    width: 6px;
    background: linear-gradient(180deg, rgba(99, 179, 237, 0.3), rgba(66, 153, 225, 0.5));
    box-shadow: 0 0 8px rgba(66, 153, 225, 0.2);
  }
}

:deep(.xterm-screen) {
  padding: 0;
}

:deep(.xterm-helper-textarea) {
  position: fixed !important;
}
</style>

<style>
/* 终端右键菜单（非 scoped，因为 Teleport 到 body） */
.term-ctx-mask {
  position: fixed;
  inset: 0;
  z-index: 10000;
}

.term-ctx-menu {
  position: fixed;
  background: rgba(30, 30, 30, 0.98);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  padding: 4px;
  min-width: 200px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  z-index: 10001;
  animation: termCtxIn 0.1s ease-out;
}

@keyframes termCtxIn {
  from { opacity: 0; transform: scale(0.95); }
}

.term-ctx-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  color: #a0aec0;
  font-size: 12px;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.1s;
}

.term-ctx-item:hover {
  background: rgba(66, 153, 225, 0.15);
  color: #e2e8f0;
}

.term-ctx-key {
  margin-left: auto;
  color: #4a5568;
  font-size: 10px;
  font-family: monospace;
}

.term-ctx-sep {
  height: 1px;
  background: rgba(255, 255, 255, 0.06);
  margin: 4px 0;
}
</style>
