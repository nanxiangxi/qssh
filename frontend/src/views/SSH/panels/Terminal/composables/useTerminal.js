/**
 * 终端核心组合式函数
 * 管理 xterm.js 实例、插件加载、尺寸调整
 */
import { ref, nextTick } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { SearchAddon } from 'xterm-addon-search'
import 'xterm/css/xterm.css'

// 默认终端主题
const DEFAULT_THEME = {
  background: '#1a1b26',
  foreground: '#c0caf5',
  cursor: '#c0caf5',
  cursorAccent: '#1a1b26',
  selectionBackground: '#33467c',
  selectionForeground: '#c0caf5',
  black: '#15161e',
  red: '#f7768e',
  green: '#9ece6a',
  yellow: '#e0af68',
  blue: '#7aa2f7',
  magenta: '#bb9af7',
  cyan: '#7dcfff',
  white: '#a9b1d6',
  brightBlack: '#414868',
  brightRed: '#f7768e',
  brightGreen: '#9ece6a',
  brightYellow: '#e0af68',
  brightBlue: '#7aa2f7',
  brightMagenta: '#bb9af7',
  brightCyan: '#7dcfff',
  brightWhite: '#c0caf5'
}

export function useTerminal(containerRef, options = {}) {
  const {
    fontSize = 14,
    fontFamily = '"Cascadia Code", "Fira Code", JetBrains Mono, Consolas, monospace',
    theme = DEFAULT_THEME,
    cursorBlink = true,
    scrollback = 10000
  } = options

  // 终端实例
  const terminal = ref(null)
  const fitAddon = ref(null)
  const searchAddon = ref(null)

  // 状态
  const isReady = ref(false)
  const cols = ref(0)
  const rows = ref(0)

  // 初始化终端
  function init() {
    if (!containerRef.value) {
      console.error('[useTerminal] 容器不存在')
      return false
    }

    // 创建终端实例
    terminal.value = new Terminal({
      fontSize,
      fontFamily,
      theme,
      cursorBlink,
      scrollback,
      allowProposedApi: true,
      convertEol: true,
      disableStdin: false
    })

    // 加载插件
    fitAddon.value = new FitAddon()
    searchAddon.value = new SearchAddon()
    const webLinksAddon = new WebLinksAddon()

    terminal.value.loadAddon(fitAddon.value)
    terminal.value.loadAddon(searchAddon.value)
    terminal.value.loadAddon(webLinksAddon)

    // 打开终端
    terminal.value.open(containerRef.value)

    // 初始适配
    fit()

    isReady.value = true
    cols.value = terminal.value.cols
    rows.value = terminal.value.rows

    console.log('[useTerminal] 终端初始化完成', cols.value, 'x', rows.value)
    return true
  }

  // 适配容器尺寸
  function fit() {
    if (!fitAddon.value || !terminal.value) return

    try {
      const { clientWidth, clientHeight } = containerRef.value
      if (clientWidth > 0 && clientHeight > 0) {
        fitAddon.value.fit()
        cols.value = terminal.value.cols
        rows.value = terminal.value.rows
      }
    } catch (e) {
      console.warn('[useTerminal] fit 失败:', e)
    }
  }

  // 写入数据
  function write(data) {
    if (terminal.value) {
      terminal.value.write(data)
    }
  }

  // 写入带换行的文本
  function writeln(text) {
    if (terminal.value) {
      terminal.value.writeln(text)
    }
  }

  // 清屏
  function clear() {
    if (terminal.value) {
      terminal.value.clear()
    }
  }

  // 全选
  function selectAll() {
    if (terminal.value) {
      terminal.value.selectAll()
    }
  }

  // 获取选中文本
  function getSelection() {
    return terminal.value?.getSelection() || ''
  }

  // 搜索
  function search(keyword, options = {}) {
    if (searchAddon.value && keyword) {
      return searchAddon.value.findNext(keyword, options)
    }
    return false
  }

  function searchPrevious(keyword) {
    if (searchAddon.value && keyword) {
      return searchAddon.value.findPrevious(keyword)
    }
    return false
  }

  function clearSearch() {
    if (searchAddon.value) {
      searchAddon.value.clearDecorations()
    }
  }

  // 注册输入回调
  function onData(callback) {
    if (terminal.value) {
      return terminal.value.onData(callback)
    }
  }

  // 注册二进制输入回调
  function onBinary(callback) {
    if (terminal.value) {
      return terminal.value.onBinary(callback)
    }
  }

  // 注册调整大小回调
  function onResize(callback) {
    if (terminal.value) {
      return terminal.value.onResize(callback)
    }
  }

  // 销毁终端
  function dispose() {
    if (terminal.value) {
      terminal.value.dispose()
      terminal.value = null
    }
    isReady.value = false
  }

  return {
    // 状态
    terminal,
    isReady,
    cols,
    rows,

    // 方法
    init,
    fit,
    write,
    writeln,
    clear,
    selectAll,
    getSelection,
    search,
    searchPrevious,
    clearSearch,
    onData,
    onBinary,
    onResize,
    dispose
  }
}
