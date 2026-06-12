<template>
  <div class="term-app">
    <!-- 工具栏 -->
    <div class="tb">
      <div class="tb-l">
        <span class="led" :class="status"></span>
        <span class="st" :class="status">{{ statusLabel }}</span>
        <span v-if="stats.commands > 0" class="s2">{{ stats.commands }} 条</span>
      </div>
      <div class="tb-c">
        <!-- 录制按钮 -->
        <button class="tbb" :class="{ recording: isRecording }" @click="toggleRecording" :title="isRecording ? '停止录制' : '开始录制'">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
            <circle v-if="!isRecording" cx="12" cy="12" r="8"/>
            <rect v-else x="6" y="6" width="12" height="12" rx="2"/>
          </svg>
        </button>
        <!-- 命令历史 -->
        <button class="tbb" @click="showHistory=!showHistory" title="命令历史">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </button>
        <!-- 清空 -->
        <button class="tbb" @click="clearAll" title="清空">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
        </button>
      </div>
      <div class="tb-r">
        <span class="sep"></span>
        <!-- 结构化模式 -->
        <button class="tbb vw" :class="{a: view==='block'}" @click="switchView('block')" title="结构化视图">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="7" height="7"/>
            <rect x="14" y="3" width="7" height="7"/>
            <rect x="3" y="14" width="7" height="7"/>
          </svg>
        </button>
        <!-- 经典模式 -->
        <button class="tbb vw" :class="{a: view==='classic'}" @click="switchView('classic')" title="经典终端">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="4 17 10 11 4 5"/>
            <line x1="12" y1="19" x2="20" y2="19"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- 块视图（读取 xterm 的输出缓冲） -->
    <div v-show="view==='block'" class="view-block">
      <div class="bp" ref="blocksRef" @scroll="onBlocksScroll">
        <div v-if="bm.blocks.value.length===0" class="empty">输入命令开始</div>
        <TerminalBlock v-for="b in bm.blocks.value" :key="b.id" :block="b"
          @toggle-collapse="b.collapsed=!b.collapsed"
          @copy="copyBlock(b.id)" @remove="bm.removeBlock(b.id)"
          @re-execute="reExec(b.id)" :plain-text="true" />
      </div>
      <div class="shortcut-bar">
        <span class="shortcut"><kbd>Tab</kbd> 补全</span>
        <span class="shortcut"><kbd>Ctrl+C</kbd> 中断</span>
        <span class="shortcut"><kbd>Ctrl+L</kbd> 清屏</span>
        <span class="shortcut"><kbd>Ctrl+R</kbd> 搜索</span>
        <span class="shortcut"><kbd>Ctrl+Z</kbd> 暂停</span>
        <span class="shortcut"><kbd>↑↓</kbd> 历史</span>
        <span class="sep"></span>
        <span class="shortcut app"><kbd>Ctrl+O</kbd> 临时终端</span>
        <button class="shortcut-more" @click="showShortcuts = true">更多...</button>
      </div>
      <div class="inp" ref="inpAreaRef">
        <span class="ps">$</span>
        <div class="iw">
          <input ref="inpRef" v-model="cmd" class="ci"
            placeholder="输入命令" :disabled="status!=='active'"
            @keydown="onKey" @input="onInput" />
        </div>
      </div>
    </div>

    <!-- 经典视图（xterm，始终挂载） -->
    <div v-show="view==='classic'" class="view-classic" ref="xtRef"></div>

    <!-- 搜索栏 -->
    <Transition name="slide-down">
      <div v-if="showSearch" class="search-bar">
        <div class="search-box">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            ref="searchInputRef"
            v-model="searchQuery"
            class="search-input"
            placeholder="搜索命令或输出..."
            @keydown.enter="searchNext"
            @keydown.escape="showSearch = false"
            @input="onSearchInput"
          />
          <span class="search-count" v-if="searchResults.length > 0">
            {{ searchIndex + 1 }}/{{ searchResults.length }}
          </span>
          <button class="search-btn" @click="searchPrev" title="上一个">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="18 15 12 9 6 15"/>
            </svg>
          </button>
          <button class="search-btn" @click="searchNext" title="下一个">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </button>
          <button class="search-btn close" @click="showSearch = false" title="关闭">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
      </div>
    </Transition>

    <!-- 补全弹窗（全局，两种模式都可用） -->
    <CompletionPopup
      ref="completionRef"
      :visible="showCompletion"
      :suggestions="completionSuggestions"
      :position="completionPos"
      @apply="(i) => view==='classic' ? applyCompletionClassic(i) : applyCompletion(i)"
      @close="showCompletion=false"
    />

    <!-- 确认切换对话框 -->
    <ConfirmSwitch
      :visible="showConfirmSwitch"
      :title="confirmTitle"
      :description="confirmDesc"
      @confirm="handleConfirmSwitch"
      @cancel="showConfirmSwitch = false"
    />

    <!-- 命令历史 -->
    <CommandHistoryDialog
      v-model:visible="showHistory"
      :history="ch.history.value"
      :favorites="ch.favorites.value"
      @execute-command="h => { cmd = h; exec(); showHistory = false }"
      @select-command="h => { cmd = h; showHistory = false; inpRef?.focus() }"
      @add-favorite="h => ch.addFavorite(h)"
      @remove-favorite="id => ch.removeFavorite(id)"
    />

    <!-- 快捷键帮助 -->
    <ShortcutsDialog :visible="showShortcuts" @close="showShortcuts = false" />

    <!-- 临时终端（复用主终端 xterm） -->
    <InteractivePrompt
      :visible="showMiniTerminal"
      :xterm-element="xtermEl"
      :initial-key="miniInitialKey"
      :on-fit="refitXterm"
      @close="closeMiniTerminal"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, inject, nextTick } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { SearchAddon } from 'xterm-addon-search'
import 'xterm/css/xterm.css'
import { Events } from '@wailsio/runtime'
import { SSHService } from '../../../../../bindings/changeme/ssh/index.js'
import { useBlockManager, BLOCK_STATUS } from './composables/useBlockManager'
import { useCommandHistory } from './composables/useCommandHistory'
import { useCommandCompletion } from './composables/useCommandCompletion'
import { useConfigStore } from '../../../../stores/config'
import { getSessionManager, SESSION_TYPE } from './utils/sessionManager'
import { logTerminalCommand, logConnectionEvent } from '../../../../utils/logger'
import TerminalBlock from './components/TerminalBlock.vue'
import CompletionPopup from './components/CompletionPopup.vue'
import ConfirmSwitch from './components/ConfirmSwitch.vue'
import InteractivePrompt from './components/InteractivePrompt.vue'
import CommandHistoryDialog from './components/CommandHistoryDialog.vue'
import ShortcutsDialog from './components/ShortcutsDialog.vue'
import { useRecording } from './composables/useRecording'
import { getThemeForUI } from './utils/terminalThemes.js'

const props = defineProps({ params: { type: Object, default: () => ({}) } })
const connId = inject('connId')
const dp = props.params?.params || props.params
const sessionId = dp?.sessionId || `term_${Date.now()}`
const isAI = dp?.isAI || false

const bm = useBlockManager()
const ch = useCommandHistory(connId)
const cfg = useConfigStore()
const sm = getSessionManager()
const completion = useCommandCompletion()

// 录制状态
const isRecording = ref(false)
const recordingStartTime = ref(null)
const recordingEntries = ref([])
let recordingTimer = null

// 命令历史弹窗
const showHistory = ref(false)

// 快捷键帮助
const showShortcuts = ref(false)

// 搜索
const showSearch = ref(false)
const searchQuery = ref('')
const searchResults = ref([])
const searchIndex = ref(-1)
const searchInputRef = ref(null)

const xtRef = ref(null)
const inpRef = ref(null)
const inpAreaRef = ref(null)
const completionRef = ref(null)
const blocksRef = ref(null)
const xtermEl = ref(null)  // xterm 实际 DOM 元素
const xtermElement = ref(null)  // xterm DOM 元素引用
const cmd = ref('')
const status = ref('idle')
const showCompletion = ref(false)
const completionSuggestions = ref([])
const completionPos = ref({ top: 0, left: 0 })
const showConfirmSwitch = ref(false)
const confirmTitle = ref('')
const confirmDesc = ref('')
const pendingKey = ref('')
const showMiniTerminal = ref(false)
const miniInitialKey = ref('')

// 视图模式：block=结构化, classic=经典
const view = ref(cfg.getDefaultTerminalType() === 'classic' ? 'classic' : 'block')

const statusLabel = computed(() => ({ idle:'空闲', starting:'连接中', active:'已连接', disconnected:'已断开', error:'错误' }[status.value]||''))
const stats = computed(() => bm.getStats())

// 同时监听 Pinia 配置和 DOM data-theme，确保跨窗口/跨方式都能同步
const uiTheme = ref(cfg.config?.ui?.theme || document.documentElement.dataset.theme || 'dark')
const terminalTheme = computed(() => getThemeForUI('default', uiTheme.value || 'dark'))
let themeObserver = null

// xterm 实例（始终初始化，始终接收输出）
let xterm = null
let fitAddon = null
let resizeObs = null
let xtermBuf = ''
let yankBuf = ''
let isUserScrolling = false
let scrollTimer = null

function applyTerminalTheme() {
  if (!xterm) return
  xterm.options.theme = { ...terminalTheme.value }
  xterm.refresh(0, xterm.rows - 1)
}

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (blocksRef.value && !isUserScrolling) {
      blocksRef.value.scrollTop = blocksRef.value.scrollHeight
    }
  })
}

// 监听块数量变化，自动滚动 + 自动折叠旧块
watch(() => bm.blocks.value.length, (newLen) => {
  scrollToBottom()
  // 超过10条时自动折叠前面的块
  if (newLen > 10) {
    const blocks = bm.blocks.value
    for (let i = 0; i < blocks.length - 10; i++) {
      if (!blocks[i].collapsed) blocks[i].collapsed = true
    }
  }
})

// 监听块内容变化（输出追加时）
watch(() => bm.rawOutput.value, () => {
  scrollToBottom()
})

// UI 主题变化时同步更新终端主题
watch(
  () => terminalTheme.value,
  () => {
    console.log('[Terminal] 终端主题变更:', uiTheme.value, terminalTheme.value.background)
    applyTerminalTheme()
  },
  { deep: false }
)

// 监听滚动位置
function onBlocksScroll() {
  if (!blocksRef.value) return
  const { scrollTop, scrollHeight, clientHeight } = blocksRef.value
  // 如果用户滚动到接近底部，重新启用自动滚动
  if (scrollHeight - scrollTop - clientHeight < 100) {
    isUserScrolling = false
  } else {
    isUserScrolling = true
  }
  // 清除之前的定时器
  if (scrollTimer) clearTimeout(scrollTimer)
  // 5秒后重新启用自动滚动
  scrollTimer = setTimeout(() => { isUserScrolling = false }, 5000)
}

// ========== 输出处理（统一数据源） ==========
function onOutput(event) {
  if (disposed) return
  const d = event?.data
  if (!d || d.sessionID !== sessionId) return
  const text = typeof d === 'string' ? d : (d.data || '')
  if (!text) return

  // 1. 写入 xterm（始终）
  if (xterm) xterm.write(text)

  // 2. 写入块管理器（AI 终端和普通终端一样处理）
  bm.appendOutput(text)
  const ended = bm.endBlockIfPrompt()
  if (ended) {
    console.log('[MainTerminal] 块已结束（检测到提示符）')
  }

  // 3. 记录到录制
  if (isRecording.value) {
    recordToRecording('output', text)
  }
}

function onDisconnected(e) {
  if (disposed) return
  if (e?.data?.connID !== connId) return
  status.value = 'disconnected'
  bm.addSystemMessage('连接已断开')
  logConnectionEvent(connId, 'disconnect')
}

function onReconnected(e) {
  if (disposed) return
  if (e?.data?.connID !== connId) return
  status.value = 'active'
  SSHService.StartShellSessionWithID(connId, sessionId).catch(() => {})
  logConnectionEvent(connId, 'reconnect')
}

// ========== 初始化 ==========
onMounted(async () => {
  // 监听 html 的 data-theme 属性变化，作为 Pinia 之外的兜底同步手段
  themeObserver = new MutationObserver((mutations) => {
    for (const m of mutations) {
      if (m.attributeName === 'data-theme') {
        uiTheme.value = document.documentElement.dataset.theme || 'dark'
        console.log('[Terminal] data-theme 变化:', uiTheme.value)
      }
    }
  })
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })

  sm.createSession(sessionId, connId, { type: isAI ? SESSION_TYPE.AI : SESSION_TYPE.NORMAL })

  // 始终初始化 xterm
  initXterm()

  if (!connId || connId === 'default-connection') return

  status.value = 'starting'
  try {
    await SSHService.StartShellSessionWithID(connId, sessionId)
    await SSHService.ResizeTerminalByID(connId, sessionId, xterm?.cols || 120, xterm?.rows || 40).catch(() => {})
    status.value = 'active'
    logConnectionEvent(connId, 'connect')

    // 通知 DockviewLayout 终端就绪
    // isAI=true 时 DockviewLayout 会转发为 terminal:ai-session-ready
    // AI executor 依赖此事件来知道 AI 终端可以接收命令
    Events.Emit('terminal:session-ready', {
      connId,
      sessionId,
      panelId: sessionId,
      isAI
    })
  } catch (e) {
    status.value = 'error'
    logConnectionEvent(connId, 'error', { error: String(e) })
    return
  }

  Events.On('ssh:terminal-output', onOutput)
  Events.On('ssh:connection-disconnected', onDisconnected)
  Events.On('ssh:connection-reconnected', onReconnected)

  // AI 通过终端执行命令时，创建块
  Events.On('ai:terminal-exec-start', (e) => {
    if (disposed) return
    const d = e?.data
    if (!d || d.sessionID !== sessionId) return
    if (!bm.activeBlock.value) {
      bm.startCommand(`[AI] ${d.command}`)
    }
  })
})

let disposed = false
onUnmounted(() => {
  if (disposed) return
  disposed = true
  // 不调用 Events.Off，避免误移除其他终端的监听器
  // handler 内部检查 disposed，不会处理已关闭终端的事件
  SSHService.CloseShellSessionByID(connId, sessionId).catch(() => {})
  resizeObs?.disconnect()
  themeObserver?.disconnect()
  xterm?.dispose()
  sm.removeSession(sessionId)
  Events.Emit('terminal:session-closed', { connId, sessionId, isAI })
})

// ========== xterm 初始化（始终执行） ==========
function initXterm() {
  if (xterm) return
  if (!xtRef.value) {
    setTimeout(() => initXterm(), 50)
    return
  }

  const fontSize = cfg.get('terminal', 'fontSize') || 14
  xterm = new Terminal({
    fontSize,
    fontFamily: '"Cascadia Code","Fira Code",Consolas,monospace',
    theme: { ...terminalTheme.value },
    cursorBlink: true,
    scrollback: 10000
  })

  fitAddon = new FitAddon()
  xterm._searchAddon = new SearchAddon()
  xterm.loadAddon(fitAddon)
  xterm.loadAddon(new WebLinksAddon())
  xterm.loadAddon(xterm._searchAddon)
  xterm.open(xtRef.value)
  fitAddon.fit()

  // 存储 xterm 实际 DOM 元素（.xterm 容器内的元素）
  xtermEl.value = xtRef.value.querySelector('.xterm') || xtRef.value

  resizeObs = new ResizeObserver(() => {
    fitAddon?.fit()
    SSHService.ResizeTerminalByID(connId, sessionId, xterm.cols, xterm.rows).catch(() => {})
  })
  resizeObs.observe(xtRef.value)

  // xterm 输入处理
  xterm.onData((data) => {
    const cc = data.charCodeAt(0)

    // Ctrl+F 搜索（经典模式）
    if (cc === 6) {
      showSearch.value = true
      nextTick(() => searchInputRef.value?.focus())
      return
    }

    // Esc 关闭补全/搜索
    if (cc === 27) {
      showCompletion.value = false
      if (showSearch.value) { showSearch.value = false; return }
      send(data); return
    }

    // Tab 补全
    if (cc === 9) {
      if (showCompletion.value && completionSuggestions.value.length > 0) {
        // 应用选中的补全
        const idx = completionRef.value?.selectedIndex || 0
        applyCompletionClassic(idx)
      } else if (xtermBuf) {
        // 触发补全
        const sugs = completion.getSuggestions(xtermBuf)
        if (sugs.length === 1) {
          // 直接补全
          const completed = completion.applyCompletion(xtermBuf, sugs[0])
          send('\b'.repeat(xtermBuf.length) + completed)
          xtermBuf = completed
        } else if (sugs.length > 1) {
          // 显示补全列表
          completionSuggestions.value = sugs
          showCompletion.value = true
        }
      }
      return
    }

    // 回车
    if (data === '\r') {
      showCompletion.value = false
      if (xtermBuf.trim() && !bm.activeBlock.value) {
        // 只有在没有活跃命令时才创建新块
        bm.startCommand(xtermBuf.trim())
        ch.addCommand(xtermBuf.trim())
        // 记录日志
        logTerminalCommand(connId, xtermBuf.trim())
      }
      xtermBuf = ''
      send('\r'); return
    }

    // Ctrl+C
    if (cc === 3) {
      bm.cancelCommand()
      xtermBuf = ''
      showCompletion.value = false
      send('\x03')
      return
    }

    // 退格
    if (cc === 127 || cc === 8) {
      xtermBuf = xtermBuf.slice(0, -1)
      send(data)
      updateCompletionClassic()
      return
    }

    // 可见字符
    if (cc >= 32) {
      xtermBuf += data
      send(data)
      updateCompletionClassic()
      return
    }

    send(data)
  })
}

// ========== 视图切换 ==========
function switchView(v) {
  view.value = v
  xtermBuf = ''  // 切换时重置输入缓冲
  showCompletion.value = false

  nextTick(() => {
    if (v === 'classic') {
      if (!xterm) initXterm()
      // 延迟 fit 确保 DOM 已更新
      setTimeout(() => {
        fitAddon?.fit()
        syncBlocksToXterm()
        xterm?.focus()
      }, 50)
    }
  })
}

// 将块内容同步到 xterm（确保经典模式有完整内容）
function syncBlocksToXterm() {
  if (!xterm || !bm.blocks.value.length) return
  // 如果 xterm 已经有内容，跳过
  if (xterm.buffer.active.length > 1) return
  // 重放所有块到 xterm
  for (const b of bm.blocks.value) {
    if (b.type === 'command') {
      xterm.writeln(`\x1b[32m$ ${b.command}\x1b[0m`)
      if (b.content) xterm.write(b.content)
    } else if (b.type === 'system') {
      xterm.writeln(`\x1b[2m--- ${b.content} ---\x1b[0m`)
    }
  }
}

// ========== 输入处理（结构化模式） ==========
function onInput() {
  if (!cmd.value) { showCompletion.value = false; return }
  const sugs = completion.getSuggestions(cmd.value)
  if (sugs.length > 0) {
    completionSuggestions.value = sugs
    showCompletion.value = true
    if (inpAreaRef.value) {
      const rect = inpAreaRef.value.getBoundingClientRect()
      completionPos.value = { top: rect.top - 210, left: rect.left }
    }
  } else {
    showCompletion.value = false
  }
}

function applyCompletion(index) {
  const s = completionSuggestions.value[index]
  if (!s) return
  cmd.value = completion.applyCompletion(cmd.value, s)
  showCompletion.value = false
  inpRef.value?.focus()
}

// 经典模式补全
function updateCompletionClassic() {
  if (!xtermBuf) { showCompletion.value = false; return }
  const sugs = completion.getSuggestions(xtermBuf)
  if (sugs.length > 0) {
    completionSuggestions.value = sugs
    showCompletion.value = true
    // 计算弹窗位置（基于屏幕坐标）
    if (xterm && xtRef.value) {
      const rect = xtRef.value.getBoundingClientRect()
      const cx = xterm.buffer.active.cursorX
      const cy = xterm.buffer.active.cursorY
      const cw = rect.width / xterm.cols
      const ch = rect.height / xterm.rows
      completionPos.value = {
        top: rect.top + (cy + 1) * ch + 4,
        left: rect.left + cx * cw
      }
    }
  } else {
    showCompletion.value = false
  }
}

function applyCompletionClassic(index) {
  const s = completionSuggestions.value[index]
  if (!s) return
  const completed = completion.applyCompletion(xtermBuf, s)
  send('\b'.repeat(xtermBuf.length) + completed)
  xtermBuf = completed
  showCompletion.value = false
}

// ========== 键盘处理 ==========
function onKey(e) {
  // 补全弹窗导航
  if (showCompletion.value && completionRef.value?.handleKey(e.key)) { e.preventDefault(); return }

  // ========== 交互式快捷键（需要 shell 交互） ==========
  const interactiveKeys = {
    'r': { code: '\x12', name: 'Ctrl+R (反向搜索)' },
    's': { code: '\x13', name: 'Ctrl+S (正向搜索)' },
  }
  if (e.ctrlKey && interactiveKeys[e.key]) {
    e.preventDefault(); showCompletion.value = false
    const info = interactiveKeys[e.key]
    const switchMode = cfg.get('terminal', 'switchMode') || 'prompt'
    if (switchMode === 'inline') {
      openMiniTerminal(info.code)
    } else {
      pendingKey.value = e.key
      confirmTitle.value = '需要切换到经典终端'
      confirmDesc.value = `${info.name} 需要在经典终端中执行。`
      showConfirmSwitch.value = true
    }
    return
  }

  // ========== 依附当前命令（中断/暂停） ==========
  // Ctrl+C: 中断
  if (e.ctrlKey && e.key === 'c') {
    e.preventDefault()
    console.log('[MainTerminal] onKey: Ctrl+C')
    bm.cancelCommand()
    send('\x03')
    cmd.value = ''
    showCompletion.value = false
    return
  }

  // Ctrl+Z: 暂停进程（交互式）
  if (e.ctrlKey && e.key === 'z') {
    e.preventDefault(); showCompletion.value = false
    const switchMode = cfg.get('terminal', 'switchMode') || 'prompt'
    if (switchMode === 'inline') {
      openMiniTerminal('\x1a')
    } else {
      pendingKey.value = 'z'
      confirmTitle.value = '需要切换到经典终端'
      confirmDesc.value = 'Ctrl+Z (暂停进程) 需要在经典终端中执行。'
      showConfirmSwitch.value = true
    }
    return
  }

  // ========== 立即执行（独立） ==========
  if (e.ctrlKey && e.key === '\\') { e.preventDefault(); send('\x1c'); return }
  if (e.ctrlKey && e.key === 's') { e.preventDefault(); send('\x13'); return }
  if (e.ctrlKey && e.key === 'q') { e.preventDefault(); send('\x11'); return }
  if (e.ctrlKey && e.key === 'd') { e.preventDefault(); send('\x04'); return }
  if (e.ctrlKey && e.key === 'g') { e.preventDefault(); send('\x07'); return }
  if (e.ctrlKey && e.key === 'f') { e.preventDefault(); openSearch(); return }

  // ========== Readline 编辑（本地 + 转发） ==========
  // 光标移动
  if (e.ctrlKey && e.key === 'a') { send('\x01'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'e') { send('\x05'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'b') { send('\x02'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'f') { send('\x06'); e.preventDefault(); return }

  // 剪切
  if (e.ctrlKey && e.key === 'k') {
    const s = inpRef.value?.selectionStart||0
    yankBuf = cmd.value.slice(s)
    cmd.value = cmd.value.slice(0,s)
    send('\x0b'); e.preventDefault(); return
  }
  if (e.ctrlKey && e.key === 'u') {
    yankBuf = cmd.value
    cmd.value = ''
    send('\x15'); e.preventDefault(); return
  }
  if (e.ctrlKey && e.key === 'w') { send('\x17'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'h') { send('\x08'); e.preventDefault(); return }

  // 粘贴
  if (e.ctrlKey && e.key === 'y') {
    const p = inpRef.value?.selectionStart||cmd.value.length
    cmd.value = cmd.value.slice(0,p)+yankBuf+cmd.value.slice(p)
    send('\x19'); e.preventDefault(); return
  }

  // 交换字符
  if (e.ctrlKey && e.key === 't') { send('\x14'); e.preventDefault(); return }

  // 历史
  if (e.ctrlKey && e.key === 'p') { send('\x10'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'n') { send('\x0e'); e.preventDefault(); return }

  // ========== Alt/Meta 组合键 ==========
  if (e.altKey) {
    const k = e.key.toLowerCase()
    if (k === 'b') { send('\x1bb'); e.preventDefault(); return }
    if (k === 'f') { send('\x1bf'); e.preventDefault(); return }
    if (k === 'd') { send('\x1bd'); e.preventDefault(); return }
    if (k === 'c') { send('\x1bc'); e.preventDefault(); return }
    if (k === 'u') { send('\x1bu'); e.preventDefault(); return }
    if (k === 'l') { send('\x1bl'); e.preventDefault(); return }
    if (k === '.') { send('\x1b.'); e.preventDefault(); return }
    if (k === 'y') { send('\x1by'); e.preventDefault(); return }
    if (k === 'backspace') { send('\x1b\x7f'); e.preventDefault(); return }
    return
  }

  // ========== Ctrl+O: 临时终端（不与终端快捷键冲突） ==========
  if (e.ctrlKey && e.key === 'o') {
    e.preventDefault(); showCompletion.value = false
    openMiniTerminal('')
    return
  }

  // ========== Enter: 执行命令（不补全） ==========
  if (e.key === 'Enter') {
    e.preventDefault()
    showCompletion.value = false
    exec()
    return
  }

  // ========== 历史导航 ==========
  if (e.key === 'ArrowUp' && !showCompletion.value) {
    e.preventDefault()
    const c = ch.getPreviousCommand()
    if (c !== null) cmd.value = c
    return
  }
  if (e.key === 'ArrowDown' && !showCompletion.value) {
    e.preventDefault()
    const c = ch.getNextCommand()
    if (c !== null) cmd.value = c
    return
  }

  // ========== Tab: 补全（仅 Tab 触发） ==========
  if (e.key === 'Tab') {
    e.preventDefault()
    if (showCompletion.value && completionSuggestions.value.length > 0) {
      applyCompletion(completionRef.value?.selectedIndex || 0)
    } else if (cmd.value) {
      const sugs = completion.getSuggestions(cmd.value)
      if (sugs.length === 1) {
        cmd.value = completion.applyCompletion(cmd.value, sugs[0])
      } else if (sugs.length > 1) {
        completionSuggestions.value = sugs
        showCompletion.value = true
      }
    }
    return
  }
}

function exec() {
  const c = cmd.value.trim()
  send((c || '') + '\r')
  cmd.value = ''
  xtermBuf = ''

  if (c) {
    // 记录日志
    logTerminalCommand(connId, c)

    // 只有在没有活跃命令时才创建新块
    if (!bm.activeBlock.value) {
      bm.startCommand(c)
      ch.addCommand(c)
      if (isRecording.value) {
        recordToRecording('command', c)
      }
    } else {
      // 有活跃命令时，输入属于当前命令（如交互式程序）
      console.log('[MainTerminal] 命令正在运行，输入属于当前块')
    }
  }
}

async function send(data) {
  if (status.value !== 'active') return
  try { await SSHService.WriteToTerminalByID(connId, sessionId, data) } catch {}
}

// ========== 搜索功能 ==========
function openSearch() {
  showSearch.value = true
  nextTick(() => searchInputRef.value?.focus())
}

function onSearchInput() {
  if (!searchQuery.value) {
    searchResults.value = []
    searchIndex.value = -1
    if (xterm) xterm.clearSelection()
    return
  }
  if (view.value === 'classic') {
    // 经典模式：使用 xterm SearchAddon
    if (xterm && xterm._searchAddon) {
      xterm._searchAddon.findNext(searchQuery.value)
      // 确保 xterm 有焦点以显示高亮
      xterm.focus()
    }
  } else {
    // 结构化模式：搜索块内容
    searchResults.value = bm.searchBlocks(searchQuery.value)
    searchIndex.value = searchResults.value.length > 0 ? 0 : -1
    if (searchIndex.value >= 0) scrollToBlock(searchResults.value[searchIndex.value].id)
  }
}

function searchNext() {
  if (view.value === 'classic') {
    if (xterm && xterm._searchAddon && searchQuery.value) {
      xterm._searchAddon.findNext(searchQuery.value)
      xterm.focus()
    }
  } else {
    if (searchResults.value.length === 0) return
    searchIndex.value = (searchIndex.value + 1) % searchResults.value.length
    scrollToBlock(searchResults.value[searchIndex.value].id)
  }
}

function searchPrev() {
  if (view.value === 'classic') {
    if (xterm && xterm._searchAddon && searchQuery.value) {
      xterm._searchAddon.findPrevious(searchQuery.value)
      xterm.focus()
    }
  } else {
    if (searchResults.value.length === 0) return
    searchIndex.value = (searchIndex.value - 1 + searchResults.value.length) % searchResults.value.length
    scrollToBlock(searchResults.value[searchIndex.value].id)
  }
}

function scrollToBlock(blockId) {
  nextTick(() => {
    const el = document.querySelector(`[data-block-id="${blockId}"]`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
      el.classList.add('highlight')
      setTimeout(() => el.classList.remove('highlight'), 2000)
    }
  })
}

// ========== 其他 ==========
const interactiveKeyMap = {
  'r': '\x12',
  'z': '\x1a',
  's': '\x13',
}

function handleConfirmSwitch() {
  showConfirmSwitch.value = false
  if (!pendingKey.value) return
  const ctrl = interactiveKeyMap[pendingKey.value] || '\x12'

  // 切换到经典模式
  view.value = 'classic'
  nextTick(() => {
    if (!xterm) initXterm()
    setTimeout(() => {
      fitAddon?.fit()
      // 发送控制码
      send(ctrl)
      // 立即聚焦 xterm
      xterm?.focus()
      pendingKey.value = ''
    }, 50)
  })
}

// 关闭临时终端（把 xterm 移回主终端）
function closeMiniTerminal() {
  showMiniTerminal.value = false

  // 把 xterm 移回主终端容器
  nextTick(() => {
    if (xtRef.value && xtermEl.value) {
      xtRef.value.appendChild(xtermEl.value)
      // 延迟调整尺寸，确保 DOM 已更新
      setTimeout(() => {
        fitAddon?.fit()
        xterm?.focus()
        console.log('[MainTerminal] xterm 已移回主终端')
      }, 50)
    }
  })
}

// 打开临时终端（发送初始按键到主 session）
function openMiniTerminal(initialKey = '') {
  miniInitialKey.value = initialKey
  showMiniTerminal.value = true
}

// 调整 xterm 尺寸 + 发送初始按键（由 InteractivePrompt 调用）
function refitXterm() {
  nextTick(() => {
    fitAddon?.fit()
    if (xterm) {
      SSHService.ResizeTerminalByID(connId, sessionId, xterm.cols, xterm.rows).catch(() => {})
    }
    // 发送初始按键（如 Ctrl+R）- 在 xterm 移动到弹窗后执行
    if (miniInitialKey.value) {
      setTimeout(() => {
        send(miniInitialKey.value)
        miniInitialKey.value = ''
      }, 50)
    }
  })
}

function copyBlock(id) { const c=bm.getBlockContent(id); if(c)navigator.clipboard.writeText(c.output||c.command||'') }
function reExec(id) { const c=bm.getBlockContent(id); if(c?.command){cmd.value=c.command;inpRef.value?.focus()} }

// 录制
function toggleRecording() {
  if (isRecording.value) {
    // 停止录制
    isRecording.value = false
    if (recordingTimer) {
      clearInterval(recordingTimer)
      recordingTimer = null
    }
    // 保存录制内容
    if (recordingEntries.value.length > 0) {
      saveRecordingContent()
    }
  } else {
    // 开始录制
    isRecording.value = true
    recordingStartTime.value = Date.now()
    recordingEntries.value = []
    recordingTimer = setInterval(() => {
      // 更新录制时长显示
    }, 1000)
  }
}

// 记录输出到录制
function recordToRecording(type, data) {
  if (!isRecording.value) return
  recordingEntries.value.push({
    type,
    data,
    timestamp: Date.now() - recordingStartTime.value
  })
}

// 保存录制内容
function saveRecordingContent() {
  const lines = recordingEntries.value.map(entry => {
    const time = formatDuration(entry.timestamp)
    if (entry.type === 'command') return `[${time}] $ ${entry.data}`
    if (entry.type === 'input') return `[${time}] > ${entry.data}`
    return `[${time}] ${entry.data}`
  })

  const content = `Session Recording\nStart: ${new Date(recordingStartTime.value).toLocaleString()}\nEntries: ${recordingEntries.value.length}\n---\n${lines.join('\n')}`

  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `recording-${sessionId}-${Date.now()}.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

function formatDuration(ms) {
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  const h = Math.floor(m / 60)
  return `${h}:${String(m % 60).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
}

// 清空
function clearAll() {
  bm.clearBlocks()
  if (xterm) xterm.clear()
}
</script>

<style scoped>
.term-app {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-terminal);
}

.tb {
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 8px;
  background: var(--bg-toolbar);
  border-bottom: 1px solid var(--border-default);
  flex-shrink: 0;
}

.tb-l, .tb-c, .tb-r {
  display: flex;
  align-items: center;
  gap: 4px;
}

.led { width: 6px; height: 6px; border-radius: 50%; }
.led.active { background: var(--accent-success); box-shadow: 0 0 4px color-mix(in srgb, var(--accent-success) 50%, transparent); }
.led.disconnected, .led.error { background: var(--accent-danger); }
.led.starting { background: var(--accent-warning); animation: pulse 1s infinite; }
.st { font-size: 11px; color: var(--text-secondary); }
.st.active { color: var(--accent-success); }
.s2 { font-size: 10px; color: var(--text-muted); }

.tbb {
  display: flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border: none; border-radius: 4px;
  color: var(--text-muted); cursor: pointer; font-size: 12px; background: transparent;
}
.tbb:hover { background: var(--surface-hover); color: var(--text-secondary); }
.tbb.recording { color: var(--accent-danger); animation: pulse 1s infinite; }
.tbb.vw { font-size: 14px; width: 26px; }
.tbb.vw.a { background: var(--surface-hover); color: var(--accent-success); }

.sep { width: 1px; height: 14px; background: var(--border-default); margin: 0 4px; }

/* 结构化视图 */
.view-block {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.bp {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px;
  scroll-behavior: smooth;
}

.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  font-size: 13px;
}

.shortcut-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 5px 10px;
  background: var(--bg-toolbar);
  border-top: 1px solid var(--border-default);
  flex-shrink: 0;
  overflow-x: auto;
}

.shortcut {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  color: var(--text-muted);
  white-space: nowrap;
}

.shortcut kbd {
  display: inline-flex;
  align-items: center;
  min-width: 18px;
  height: 16px;
  padding: 0 4px;
  background: var(--surface-2);
  border: 1px solid var(--border-strong);
  border-radius: 3px;
  font-size: 9px;
  color: var(--text-secondary);
  font-family: inherit;
}

.shortcut-more {
  background: transparent;
  border: none;
  color: var(--primary-light);
  font-size: 10px;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 3px;
}

.shortcut.app {
  color: var(--accent-success);
}

.shortcut.app kbd {
  background: var(--success-bg);
  border-color: var(--success-bg);
  color: var(--accent-success);
}

.sep {
  width: 1px;
  height: 12px;
  background: var(--border-default);
  margin: 0 4px;
}

.shortcut-more:hover {
  background: var(--primary-bg);
}

/* 经典视图 */
.view-classic {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.view-classic :deep(.xterm) {
  height: 100%;
  width: 100%;
}

.view-classic :deep(.xterm-viewport) {
  overflow-y: auto !important;
  background-color: transparent !important;
}

.view-classic :deep(.xterm-screen) {
  padding: 0;
}

/* 输入栏 */
.inp {
  display: flex;
  align-items: center;
  padding: 6px 10px;
  background: var(--bg-toolbar);
  border-top: 1px solid var(--border-default);
  gap: 8px;
  flex-shrink: 0;
}

.iw { flex: 1; position: relative; }

.ps {
  font-family: monospace;
  font-size: 13px;
  color: var(--accent-success);
}

.ci {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-family: 'Cascadia Code', monospace;
  font-size: 13px;
  width: 100%;
}

.ci::placeholder { color: var(--text-muted); }
.ci:disabled { opacity: .4; }

/* 滚动条 */
.bp::-webkit-scrollbar { width: 6px; }
.bp::-webkit-scrollbar-track { background: transparent; }
.bp::-webkit-scrollbar-thumb { background: var(--scrollbar-thumb); border-radius: 3px; }
.bp::-webkit-scrollbar-thumb:hover { background: var(--scrollbar-thumb-hover); }

@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.4} }

/* xterm 搜索高亮 */
.view-classic :deep(.xterm-decoration-top),
.view-classic :deep(.xterm-decoration-bottom),
.view-classic :deep(.xterm-find-result-decoration) {
  background: color-mix(in srgb, var(--accent-warning) 50%, transparent) !important;
}

.view-classic :deep(.xterm-decoration-over),
.view-classic :deep(.xterm-find-result-selected-decoration) {
  background: color-mix(in srgb, var(--accent-warning) 80%, transparent) !important;
}

/* 搜索栏 */
.search-bar {
  position: absolute;
  top: 32px;
  right: 8px;
  z-index: 100;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: var(--bg-toolbar);
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  box-shadow: var(--shadow-md);
}

.search-box svg { color: var(--text-muted); flex-shrink: 0; }

.search-input {
  width: 200px;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: 12px;
}

.search-input::placeholder { color: var(--text-muted); }

.search-count {
  font-size: 10px;
  color: var(--text-muted);
  white-space: nowrap;
}

.search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: transparent;
  border: none;
  border-radius: 3px;
  color: var(--text-muted);
  cursor: pointer;
}

.search-btn:hover { background: var(--surface-hover); color: var(--text-secondary); }
.search-btn.close:hover { background: var(--danger-bg); color: var(--accent-danger); }

/* 块高亮（搜索结果） */
:deep(.terminal-block.highlight) {
  border-color: var(--accent-warning) !important;
  box-shadow: 0 0 8px color-mix(in srgb, var(--accent-warning) 30%, transparent);
}

/* 搜索栏动画 */
.slide-down-enter-active { transition: all .2s ease; }
.slide-down-leave-active { transition: all .15s ease; }
.slide-down-enter-from, .slide-down-leave-to { opacity: 0; transform: translateY(-8px); }

.term-app { position: relative; }
</style>
