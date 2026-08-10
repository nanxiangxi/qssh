<template>
  <div class="term-app">
    <div class="tb">
      <div class="tb-l">
        <span class="led" :class="status"></span>
        <span class="st" :class="status">{{ statusLabel }}</span>
        <span v-if="stats.commands > 0" class="s2">{{ t('terminal.commandCount', { count: stats.commands }) }}</span>
      </div>
      <div class="tb-c">
        <button class="tbb" :class="{ recording: isRecording }" @click="toggleRecording" :title="isRecording ? t('terminal.stopRecording') : t('terminal.startRecording')">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
            <circle v-if="!isRecording" cx="12" cy="12" r="8"/>
            <rect v-else x="6" y="6" width="12" height="12" rx="2"/>
          </svg>
        </button>
        <button class="tbb" @click="showHistory=!showHistory" :title="t('terminal.commandHistory')">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
          </svg>
        </button>
        <button class="tbb" @click="clearAll" :title="t('common.clear')">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
        </button>
      </div>
      <div class="tb-r">
        <span class="sep"></span>
        <button class="tbb vw" :class="{a: view==='block'}" @click="switchView('block')" :title="t('terminal.structuredView')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/>
          </svg>
        </button>
        <button class="tbb vw" :class="{a: view==='classic'}" @click="switchView('classic')" :title="t('terminal.classicTerminal')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
          </svg>
        </button>
      </div>
    </div>

    <div v-show="view==='block'" class="view-block" @contextmenu.prevent="onBlockContextMenu">
      <div class="bp" ref="blocksRef" @scroll="onBlocksScroll">
        <div v-if="bm.blocks.value.length===0" class="empty">{{ t('terminal.inputCommandToStart') }}</div>
        <TerminalBlock v-for="b in bm.blocks.value" :key="b.id" :block="b"
          @toggle-collapse="b.collapsed=!b.collapsed"
          @copy="copyBlock(b.id)" @remove="bm.removeBlock(b.id)"
          @re-execute="reExec(b.id)" :plain-text="true" />
      </div>
      <div class="shortcut-bar">
        <span class="shortcut"><kbd>Tab</kbd> {{ t('terminal.shortcutComplete') }}</span>
        <span class="shortcut"><kbd>Ctrl+C</kbd> {{ t('terminal.shortcutInterrupt') }}</span>
        <span class="shortcut"><kbd>Ctrl+L</kbd> {{ t('terminal.shortcutClear') }}</span>
        <span class="shortcut"><kbd>Ctrl+R</kbd> {{ t('terminal.shortcutSearch') }}</span>
        <span class="shortcut"><kbd>Ctrl+Z</kbd> {{ t('terminal.shortcutPause') }}</span>
        <span class="shortcut"><kbd>↑↓</kbd> {{ t('terminal.shortcutHistory') }}</span>
        <span class="sep"></span>
        <span class="shortcut app"><kbd>Ctrl+O</kbd> {{ t('terminal.shortcutTempTerminal') }}</span>
        <button class="shortcut-more" @click="showShortcuts = true">{{ t('terminal.shortcutMore') }}</button>
      </div>
      <div class="inp" ref="inpAreaRef">
        <span class="ps">$</span>
        <div class="iw">
          <input v-if="commandSendMode !== 'button'" ref="inpRef" v-model="cmd" class="ci"
            :placeholder="t('terminal.inputCommand')" :disabled="status!=='active'"
            @keydown="onKey" @input="onInput" />
          <textarea v-else ref="inpRef" v-model="cmd" class="ci ci-textarea"
            :placeholder="t('terminal.inputCommandMultiline')" :disabled="status!=='active'"
            @keydown="onKeyButton" @input="onInputButton" rows="1"></textarea>
        </div>
        <button v-if="commandSendMode === 'button'" class="send-btn" @click="exec" :disabled="!cmd.trim() || status !== 'active'" :title="t('terminal.sendCommand')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/>
          </svg>
        </button>
      </div>
    </div>

    <div v-show="view==='classic'" class="view-classic" ref="xtRef" @contextmenu.prevent="onClassicContextMenu"></div>

    <TerminalContextMenu
      v-model:visible="showContextMenu"
      :position="contextMenuPos"
      :has-selection="hasSelection"
      @copy="ctxCopy" @paste="ctxPaste" @select-all="ctxSelectAll"
      @clear="clearAll" @search="openSearch"
    />

    <Transition name="slide-down">
      <div v-if="showSearch" class="search-bar">
        <div class="search-box">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input ref="searchInputRef" v-model="searchQuery" class="search-input" :placeholder="t('terminal.searchPlaceholder')"
            @keydown.enter="searchNext" @keydown.escape="showSearch = false" @input="onSearchInput" />
          <span class="search-count" v-if="searchResults.length > 0">{{ searchIndex + 1 }}/{{ searchResults.length }}</span>
          <button class="search-btn" @click="searchPrev" :title="t('terminal.searchPrev')"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="18 15 12 9 6 15"/></svg></button>
          <button class="search-btn" @click="searchNext" :title="t('terminal.searchNext')"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg></button>
          <button class="search-btn close" @click="showSearch = false" :title="t('terminal.searchClose')"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
        </div>
      </div>
    </Transition>

    <CompletionPopup ref="completionRef" :visible="showCompletion" :suggestions="completionSuggestions"
      :position="completionPos" @apply="(i) => view==='classic' ? applyCompletionClassic(i) : applyCompletion(i)" @close="showCompletion=false" />

    <ConfirmSwitch :visible="showConfirmSwitch" :title="confirmTitle" :description="confirmDesc"
      @confirm="handleConfirmSwitch" @cancel="showConfirmSwitch = false" />

    <CommandHistoryDialog v-model:visible="showHistory" :history="ch.history.value" :favorites="ch.favorites.value"
      @execute-command="h => { cmd = h; exec(); showHistory = false }"
      @select-command="h => { cmd = h; showHistory = false; inpRef?.focus() }"
      @add-favorite="h => ch.addFavorite(h)" @remove-favorite="id => ch.removeFavorite(id)" />

    <ShortcutsDialog :visible="showShortcuts" @close="showShortcuts = false" />

    <InteractivePrompt :visible="showMiniTerminal" :xterm-element="xtermEl" :initial-key="miniInitialKey"
      :on-fit="refitXterm" @close="closeMiniTerminal" />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, inject, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { SearchAddon } from 'xterm-addon-search'
import 'xterm/css/xterm.css'
import { Events, Browser } from '@wailsio/runtime'
import { SSHService } from '../../../../../bindings/changeme/ssh/index.js'
import { useBlockManager } from './composables/useBlockManager'
import { useCommandHistory } from './composables/useCommandHistory'
import { useCommandCompletion } from './composables/useCommandCompletion'
import { useCdSync } from './composables/useCdSync'
import { useTerminalClipboard } from './composables/useTerminalClipboard'
import { useConfigStore } from '../../../../stores/config'
import { getSessionManager, SESSION_TYPE } from './utils/sessionManager'
import { logTerminalCommand, logConnectionEvent } from '../../../../utils/logger'
import TerminalBlock from './components/TerminalBlock.vue'
import CompletionPopup from './components/CompletionPopup.vue'
import ConfirmSwitch from './components/ConfirmSwitch.vue'
import InteractivePrompt from './components/InteractivePrompt.vue'
import CommandHistoryDialog from './components/CommandHistoryDialog.vue'
import ShortcutsDialog from './components/ShortcutsDialog.vue'
import TerminalContextMenu from './components/TerminalContextMenu.vue'
import { highlightOutput } from './utils/highlightAddon'

const { t } = useI18n()
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
const cdSync = useCdSync(connId)
const clipboard = useTerminalClipboard()

const commandSendMode = computed(() => cfg.get('terminal', 'commandSendMode') || 'enter')
const highlightEnabled = computed(() => cfg.get('terminal', 'codeHighlight') || false)
const view = ref(cfg.getDefaultTerminalType() === 'classic' ? 'classic' : 'block')

const status = ref('idle')
const cmd = ref('')
const showHistory = ref(false)
const showShortcuts = ref(false)
const showSearch = ref(false)
const searchQuery = ref('')
const searchResults = ref([])
const searchIndex = ref(-1)
const searchInputRef = ref(null)
const showCompletion = ref(false)
const completionSuggestions = ref([])
const completionPos = ref({ top: 0, left: 0 })
const showConfirmSwitch = ref(false)
const confirmTitle = ref('')
const confirmDesc = ref('')
const pendingKey = ref('')
const showMiniTerminal = ref(false)
const miniInitialKey = ref('')
const showContextMenu = ref(false)
const contextMenuPos = ref({ x: 0, y: 0 })

const xtRef = ref(null)
const inpRef = ref(null)
const inpAreaRef = ref(null)
const completionRef = ref(null)
const blocksRef = ref(null)
const xtermEl = ref(null)

const statusLabel = computed(() => ({ idle:t('terminal.statusIdle'), starting:t('terminal.statusConnecting'), active:t('terminal.statusConnected'), disconnected:t('terminal.statusDisconnected'), error:t('terminal.statusError') }[status.value]||''))
const stats = computed(() => bm.getStats())
const hasSelection = computed(() => clipboard.hasSelection.value || (view.value === 'classic' && xterm?.hasSelection()))

let xterm = null
let fitAddon = null
let resizeObs = null
let xtermBuf = ''
let yankBuf = ''
let isUserScrolling = false
let scrollTimer = null
let disposed = false
const isRecording = ref(false)
let recordingStartTime = null
const recordingEntries = ref([])

function scrollToBottom() { nextTick(() => { if (blocksRef.value && !isUserScrolling) blocksRef.value.scrollTop = blocksRef.value.scrollHeight }) }
watch(() => bm.blocks.value.length, (n) => { scrollToBottom(); if (n > 10) { const b = bm.blocks.value; for (let i = 0; i < n - 10; i++) if (!b[i].collapsed) b[i].collapsed = true } })
watch(showContextMenu, (v) => { if (!v && view.value === 'classic') nextTick(() => xterm?.focus()) })
watch(() => bm.rawOutput.value, () => scrollToBottom())

function onBlocksScroll() {
  if (!blocksRef.value) return
  const { scrollTop, scrollHeight, clientHeight } = blocksRef.value
  isUserScrolling = scrollHeight - scrollTop - clientHeight >= 100
  if (scrollTimer) clearTimeout(scrollTimer)
  scrollTimer = setTimeout(() => { isUserScrolling = false }, 5000)
}

function onOutput(event) {
  if (disposed) return
  const d = event?.data
  if (!d || d.sessionID !== sessionId) return
  let text = typeof d === 'string' ? d : (d.data || '')
  if (!text) return
  if (highlightEnabled.value) text = highlightOutput(text)
  if (xterm) xterm.write(text)
  bm.appendOutput(text)
  bm.endBlockIfPrompt()
  if (isRecording.value) recordingEntries.value.push({ type: 'output', data: text, timestamp: Date.now() - recordingStartTime })
}

function onDisconnected(e) { if (disposed || e?.data?.connID !== connId) return; status.value = 'disconnected'; bm.addSystemMessage(t('terminal.disconnectedMessage')); logConnectionEvent(connId, 'disconnect') }
function onReconnected(e) { if (disposed || e?.data?.connID !== connId) return; status.value = 'active'; SSHService.StartShellSessionWithID(connId, sessionId).catch(() => {}); logConnectionEvent(connId, 'reconnect') }

onMounted(async () => {
  sm.createSession(sessionId, connId, { type: isAI ? SESSION_TYPE.AI : SESSION_TYPE.NORMAL })
  initXterm()
  if (xtRef.value) clipboard.trackSelection(xtRef.value, () => xterm)
  if (!connId || connId === 'default-connection') return
  status.value = 'starting'
  try {
    await SSHService.StartShellSessionWithID(connId, sessionId)
    await SSHService.ResizeTerminalByID(connId, sessionId, xterm?.cols || 120, xterm?.rows || 40).catch(() => {})
    status.value = 'active'
    logConnectionEvent(connId, 'connect')
    await cdSync.initHomeDir()
    Events.Emit('terminal:session-ready', { connId, sessionId, panelId: sessionId, isAI })
  } catch (e) { status.value = 'error'; logConnectionEvent(connId, 'error', { error: String(e) }); return }
  Events.On('ssh:terminal-output', onOutput)
  Events.On('ssh:connection-disconnected', onDisconnected)
  Events.On('ssh:connection-reconnected', onReconnected)
  document.addEventListener('keydown', onGlobalKey)
  Events.On('ai:terminal-exec-start', (e) => { if (disposed) return; const d = e?.data; if (!d || d.sessionID !== sessionId) return; if (!bm.activeBlock.value) bm.startCommand(`[AI] ${d.command}`) })
})

onUnmounted(() => {
  if (disposed) return; disposed = true
  document.removeEventListener('keydown', onGlobalKey)
  SSHService.CloseShellSessionByID(connId, sessionId).catch(() => {})
  resizeObs?.disconnect(); xterm?.dispose(); sm.removeSession(sessionId)
  Events.Emit('terminal:session-closed', { connId, sessionId, isAI })
  // 清理事件监听器
  Events.Off('ssh:terminal-output')
  Events.Off('ssh:connection-disconnected')
  Events.Off('ssh:connection-reconnected')
  Events.Off('ai:terminal-exec-start')
})

function getCSSVar(n) { return getComputedStyle(document.documentElement).getPropertyValue(n).trim() }
function getXtermTheme() {
  const isLight = document.documentElement.dataset.theme === 'light'
  const base = { background: getCSSVar('--bg-terminal') || '#121212', foreground: getCSSVar('--text-primary') || '#d4d4d4', cursor: getCSSVar('--text-primary') || '#d4d4d4', cursorAccent: getCSSVar('--bg-terminal') || '#121212', selectionBackground: getCSSVar('--surface-hover') || 'rgba(255,255,255,0.1)', findMatch: getCSSVar('--accent-warning') || '#ff9800', findMatchSelected: getCSSVar('--accent-warning') || '#ff9800', findMatchHighlight: 'rgba(255,152,0,0.4)', findMatchHighlightSelected: 'rgba(255,152,0,0.6)' }
  if (isLight) return { ...base, black:'#000', red:'#d32f2f', green:'#388e3c', yellow:'#e65100', blue:'#1976d2', magenta:'#00838f', cyan:'#00695c', white:'#616161', brightBlack:'#9e9e9e', brightRed:'#f44336', brightGreen:'#4caf50', brightYellow:'#ff9800', brightBlue:'#2196f3', brightMagenta:'#00bcd4', brightCyan:'#009688', brightWhite:'#212121' }
  return base
}

function initXterm() {
  if (xterm) return
  if (!xtRef.value) { setTimeout(() => initXterm(), 50); return }
  const fontSize = cfg.get('terminal', 'fontSize') || 14
  xterm = new Terminal({ fontSize, fontFamily: '"Cascadia Code","Fira Code",Consolas,monospace', theme: getXtermTheme(), cursorBlink: true, scrollback: 10000 })
  fitAddon = new FitAddon(); xterm._searchAddon = new SearchAddon()
  xterm.loadAddon(fitAddon); xterm.loadAddon(new WebLinksAddon((e, uri) => { e.preventDefault(); Browser.OpenURL(uri) })); xterm.loadAddon(xterm._searchAddon)
  xterm.open(xtRef.value); fitAddon.fit()
  xterm.attachCustomKeyEventHandler((e) => { if (e.ctrlKey && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) return false; return true })
  xtermEl.value = xtRef.value.querySelector('.xterm') || xtRef.value
  resizeObs = new ResizeObserver(() => { fitAddon?.fit(); SSHService.ResizeTerminalByID(connId, sessionId, xterm.cols, xterm.rows).catch(() => {}) })
  resizeObs.observe(xtRef.value)

  xterm.onData((data) => {
    const cc = data.charCodeAt(0)
    if (cc === 6) { showSearch.value = true; nextTick(() => searchInputRef.value?.focus()); return }
    if (cc === 27) { showCompletion.value = false; if (showSearch.value) { showSearch.value = false; return } send(data); return }
    if (cc === 9) {
      if (showCompletion.value && completionSuggestions.value.length > 0) { applyCompletionClassic(completionRef.value?.selectedIndex || 0) }
      else if (xtermBuf) {
        const parts = xtermBuf.split(/\s+/)
        const lastWord = parts[parts.length - 1] || ''
        const isPath = lastWord.startsWith('/') || lastWord.startsWith('~') || lastWord.startsWith('.')
        if (isPath && connId) {
          // 路径补全：直接走异步远程补全
          completion.getSuggestionsAsync(xtermBuf, connId).then(asyncSugs => {
            if (asyncSugs.length === 1) { const c = completion.applyCompletion(xtermBuf, asyncSugs[0]); send('\b'.repeat(xtermBuf.length) + c); xtermBuf = c; showCompletion.value = false }
            else if (asyncSugs.length > 1) { completionSuggestions.value = asyncSugs; showCompletion.value = true }
          })
        } else {
          // 命令/选项补全：同步
          const sugs = completion.getSuggestions(xtermBuf)
          if (sugs.length === 1) { const c = completion.applyCompletion(xtermBuf, sugs[0]); send('\b'.repeat(xtermBuf.length) + c); xtermBuf = c }
          else if (sugs.length > 1) { completionSuggestions.value = sugs; showCompletion.value = true }
        }
      }
      return
    }
    if (data === '\r') {
      showCompletion.value = false
      let actualCmd = xtermBuf.trim()
      if (xterm) { try { const line = xterm.buffer.active.getLine(xterm.buffer.active.cursorY); if (line) { const lt = line.translateToString(true).trim(); const pi = Math.max(lt.lastIndexOf('$ '), lt.lastIndexOf('# ')); if (pi >= 0) actualCmd = lt.substring(pi + 2).trim(); else if (lt.length > 0) actualCmd = lt } } catch {} }
      if (actualCmd && !bm.activeBlock.value) { bm.startCommand(actualCmd); ch.addCommand(actualCmd); logTerminalCommand(connId, actualCmd); cdSync.detectCd(actualCmd) }
      xtermBuf = ''; send('\r'); return
    }
    if (cc === 3) { bm.cancelCommand(); xtermBuf = ''; showCompletion.value = false; send('\x03'); return }
    if (cc === 127 || cc === 8) { xtermBuf = xtermBuf.slice(0, -1); send(data); updateCompletionClassic(); return }
    if (cc >= 32) { xtermBuf += data; send(data); updateCompletionClassic(); return }
    send(data)
  })
}

function switchView(v) { view.value = v; xtermBuf = ''; showCompletion.value = false; nextTick(() => { if (v === 'classic') { if (!xterm) initXterm(); setTimeout(() => { fitAddon?.fit(); syncBlocksToXterm(); xterm?.focus() }, 50) } }) }
function syncBlocksToXterm() { if (!xterm || !bm.blocks.value.length) return; if (xterm.buffer.active.length > 1) return; for (const b of bm.blocks.value) { if (b.type === 'command') { xterm.writeln(`\x1b[32m$ ${b.command}\x1b[0m`); if (b.content) xterm.write(b.content) } else if (b.type === 'system') xterm.writeln(`\x1b[2m--- ${b.content} ---\x1b[0m`) } }

let completionHideTimer = null
function onInput() {
  if (!cmd.value) { showCompletion.value = false; return }
  if (completionHideTimer) { clearTimeout(completionHideTimer); completionHideTimer = null }
  const sugs = completion.getSuggestions(cmd.value)
  if (sugs.length > 0) { completionSuggestions.value = sugs; showCompletion.value = true; if (inpAreaRef.value) { const r = inpAreaRef.value.getBoundingClientRect(); const y = r.top - 210; completionPos.value = { top: Math.max(8, y < 8 ? r.bottom + 4 : y), left: r.left } } }
  else completionHideTimer = setTimeout(() => { showCompletion.value = false }, 300)
}
function applyCompletion(i) { const s = completionSuggestions.value[i]; if (!s) return; cmd.value = completion.applyCompletion(cmd.value, s); showCompletion.value = false; inpRef.value?.focus() }
function updateCompletionClassic() {
  if (!xtermBuf) { showCompletion.value = false; return }
  const sugs = completion.getSuggestions(xtermBuf)
  if (sugs.length > 0) { completionSuggestions.value = sugs; showCompletion.value = true; if (xterm && xtRef.value) { const r = xtRef.value.getBoundingClientRect(); const cx = xterm.buffer.active.cursorX; const cy = xterm.buffer.active.cursorY; const cw = r.width / xterm.cols; const ch = r.height / xterm.rows; let top = r.top + (cy + 1) * ch + 4; let left = r.left + cx * cw; if (top + 200 > window.innerHeight) top = r.top + cy * ch - 204; if (left + 250 > window.innerWidth) left = window.innerWidth - 258; completionPos.value = { top: Math.max(8, top), left: Math.max(8, left) } } }
  else showCompletion.value = false
}
function applyCompletionClassic(i) { const s = completionSuggestions.value[i]; if (!s) return; const c = completion.applyCompletion(xtermBuf, s); send('\b'.repeat(xtermBuf.length) + c); xtermBuf = c; showCompletion.value = false }

// 共享键盘逻辑（Tab 补全、Ctrl+C、历史导航、Esc）
function handleCommonKeys(e) {
  if (showCompletion.value && completionRef.value?.handleKey(e.key)) { e.preventDefault(); return true }
  if (e.ctrlKey && e.key === 'c') { e.preventDefault(); bm.cancelCommand(); send('\x03'); cmd.value = ''; showCompletion.value = false; return true }
  if (e.key === 'Escape') { showCompletion.value = false; return true }
  if (e.key === 'ArrowUp' && !showCompletion.value) { e.preventDefault(); const c = ch.getPreviousCommand(); if (c !== null) cmd.value = c; return true }
  if (e.key === 'ArrowDown' && !showCompletion.value) { e.preventDefault(); const c = ch.getNextCommand(); if (c !== null) cmd.value = c; return true }
  if (e.key === 'Tab') {
    e.preventDefault()
    if (showCompletion.value && completionSuggestions.value.length > 0) { applyCompletion(completionRef.value?.selectedIndex || 0) }
    else if (cmd.value) {
      const parts = cmd.value.split(/\s+/)
      const lastWord = parts[parts.length - 1] || ''
      const isPath = lastWord.startsWith('/') || lastWord.startsWith('~') || lastWord.startsWith('.')
      if (isPath && connId) {
        completion.getSuggestionsAsync(cmd.value, connId).then(asyncSugs => {
          if (asyncSugs.length === 1) { cmd.value = completion.applyCompletion(cmd.value, asyncSugs[0]); showCompletion.value = false }
          else if (asyncSugs.length > 1) { completionSuggestions.value = asyncSugs; showCompletion.value = true }
        })
      } else {
        const sugs = completion.getSuggestions(cmd.value)
        if (sugs.length === 1) cmd.value = completion.applyCompletion(cmd.value, sugs[0])
        else if (sugs.length > 1) { completionSuggestions.value = sugs; showCompletion.value = true }
      }
    }
    return true
  }
  return false
}

function onKey(e) {
  if (handleCommonKeys(e)) return
  const ik = { 'r': { code: '\x12', name: 'Ctrl+R (' + t('terminal.shortcutReverseSearch') + ')' }, 's': { code: '\x13', name: 'Ctrl+S (' + t('terminal.shortcutForwardSearch') + ')' } }
  if (e.ctrlKey && ik[e.key]) { e.preventDefault(); showCompletion.value = false; const info = ik[e.key]; const sm2 = cfg.get('terminal', 'switchMode') || 'prompt'; if (sm2 === 'inline') openMiniTerminal(info.code); else { pendingKey.value = e.key; confirmTitle.value = t('terminal.switchToClassic'); confirmDesc.value = t('terminal.switchDesc', { name: info.name }); showConfirmSwitch.value = true } return }
  if (e.ctrlKey && e.key === 'z') { e.preventDefault(); showCompletion.value = false; const sm2 = cfg.get('terminal', 'switchMode') || 'prompt'; if (sm2 === 'inline') openMiniTerminal('\x1a'); else { pendingKey.value = 'z'; confirmTitle.value = t('terminal.switchToClassic'); confirmDesc.value = t('terminal.switchDesc', { name: 'Ctrl+Z (' + t('terminal.shortcutPause') + ')' }); showConfirmSwitch.value = true } return }
  if (e.ctrlKey && e.key === '\\') { e.preventDefault(); send('\x1c'); return }
  if (e.ctrlKey && e.key === 's') { e.preventDefault(); send('\x13'); return }
  if (e.ctrlKey && e.key === 'q') { e.preventDefault(); send('\x11'); return }
  if (e.ctrlKey && e.key === 'd') { e.preventDefault(); send('\x04'); return }
  if (e.ctrlKey && e.key === 'g') { e.preventDefault(); send('\x07'); return }
  if (e.ctrlKey && e.key === 'f') { e.preventDefault(); openSearch(); return }
  if (e.ctrlKey && e.key === 'a') { send('\x01'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'e') { send('\x05'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'b') { send('\x02'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'f') { send('\x06'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'k') { const s = inpRef.value?.selectionStart||0; yankBuf = cmd.value.slice(s); cmd.value = cmd.value.slice(0,s); send('\x0b'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'u') { yankBuf = cmd.value; cmd.value = ''; send('\x15'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'w') { send('\x17'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'h') { send('\x08'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'y') { const p = inpRef.value?.selectionStart||cmd.value.length; cmd.value = cmd.value.slice(0,p)+yankBuf+cmd.value.slice(p); send('\x19'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 't') { send('\x14'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'p') { send('\x10'); e.preventDefault(); return }
  if (e.ctrlKey && e.key === 'n') { send('\x0e'); e.preventDefault(); return }
  if (e.altKey) { const k = e.key.toLowerCase(); const m = { b:'\x1bb', f:'\x1bf', d:'\x1bd', c:'\x1bc', u:'\x1bu', l:'\x1bl', '.':'\x1b.', y:'\x1by', backspace:'\x1b\x7f' }; if (m[k]) { send(m[k]); e.preventDefault(); return } return }
  if (e.ctrlKey && e.key === 'o') { e.preventDefault(); showCompletion.value = false; openMiniTerminal(''); return }
  if (e.key === 'Enter') { e.preventDefault(); showCompletion.value = false; exec(); return }
}

function onKeyButton(e) {
  if (handleCommonKeys(e)) return
  if (e.ctrlKey && e.key === 'Enter') { e.preventDefault(); exec(); return }
  if (e.key === 'Enter' && !e.ctrlKey) { nextTick(() => autoResizeTextarea()); return }
}
function onInputButton() { onInput(); nextTick(() => autoResizeTextarea()) }
function autoResizeTextarea() { const t = inpRef.value; if (!t) return; t.style.height = 'auto'; t.style.height = Math.min(t.scrollHeight, 120) + 'px' }

function exec() {
  const c = cmd.value.trim(); send((c || '') + '\r'); cmd.value = ''; xtermBuf = ''; showCompletion.value = false
  nextTick(() => { if (inpRef.value) inpRef.value.style.height = 'auto' })
  if (c) { logTerminalCommand(connId, c); if (!bm.activeBlock.value) { bm.startCommand(c); ch.addCommand(c); if (isRecording.value) recordingEntries.value.push({ type: 'command', data: c, timestamp: Date.now() - recordingStartTime }) } }
  cdSync.detectCd(c)
}
async function send(data) { if (status.value !== 'active') return; try { await SSHService.WriteToTerminalByID(connId, sessionId, data) } catch {} }

function onGlobalKey(e) { if (!disposed && e.ctrlKey && e.key === 'f') { e.preventDefault(); openSearch() } }
function openSearch() { showSearch.value = true; nextTick(() => searchInputRef.value?.focus()) }

// 搜索核心逻辑（经典模式用 SearchAddon，结构化模式用 bm.searchBlocks）
function doSearch(direction) {
  if (view.value === 'classic') {
    if (!xterm || !xterm._searchAddon || !searchQuery.value) return
    if (direction === 'next') xterm._searchAddon.findNext(searchQuery.value)
    else xterm._searchAddon.findPrevious(searchQuery.value)
    xterm.focus()
  } else {
    if (searchResults.value.length === 0) return
    const len = searchResults.value.length
    searchIndex.value = direction === 'next'
      ? (searchIndex.value + 1) % len
      : (searchIndex.value - 1 + len) % len
    scrollToBlock(searchResults.value[searchIndex.value].id)
  }
}

function onSearchInput() {
  if (!searchQuery.value) { searchResults.value = []; searchIndex.value = -1; if (xterm) xterm.clearSelection(); return }
  if (view.value === 'classic') { if (xterm && xterm._searchAddon) { xterm._searchAddon.findNext(searchQuery.value); xterm.focus() } }
  else { searchResults.value = bm.searchBlocks(searchQuery.value); searchIndex.value = searchResults.value.length > 0 ? 0 : -1; if (searchIndex.value >= 0) scrollToBlock(searchResults.value[searchIndex.value].id) }
}
function searchNext() { doSearch('next') }
function searchPrev() { doSearch('prev') }
function scrollToBlock(id) { nextTick(() => { const el = document.querySelector(`[data-block-id="${id}"]`); if (el) { el.scrollIntoView({ behavior: 'smooth', block: 'center' }); el.classList.add('highlight'); setTimeout(() => el.classList.remove('highlight'), 2000) } }) }

const interactiveKeyMap = { 'r': '\x12', 'z': '\x1a', 's': '\x13' }
function handleConfirmSwitch() { showConfirmSwitch.value = false; if (!pendingKey.value) return; view.value = 'classic'; nextTick(() => { if (!xterm) initXterm(); setTimeout(() => { fitAddon?.fit(); send(interactiveKeyMap[pendingKey.value] || '\x12'); xterm?.focus(); pendingKey.value = '' }, 50) }) }
function closeMiniTerminal() { showMiniTerminal.value = false; nextTick(() => { if (xtRef.value && xtermEl.value) { xtRef.value.appendChild(xtermEl.value); setTimeout(() => { fitAddon?.fit(); xterm?.focus() }, 50) } }) }
function openMiniTerminal(k = '') { miniInitialKey.value = k; showMiniTerminal.value = true }
function refitXterm() { nextTick(() => { fitAddon?.fit(); if (xterm) SSHService.ResizeTerminalByID(connId, sessionId, xterm.cols, xterm.rows).catch(() => {}); if (miniInitialKey.value) { setTimeout(() => { send(miniInitialKey.value); miniInitialKey.value = '' }, 50) } }) }

function calcMenuPos(x, y) { return { x: Math.min(x, window.innerWidth - 168), y: Math.min(y, window.innerHeight - 228) } }
function onClassicContextMenu(e) { e.preventDefault(); contextMenuPos.value = calcMenuPos(e.clientX, e.clientY); showContextMenu.value = true }
function onBlockContextMenu(e) { e.preventDefault(); contextMenuPos.value = calcMenuPos(e.clientX, e.clientY); showContextMenu.value = true }
function ctxCopy() { clipboard.copy({ value: () => xterm }, view); showContextMenu.value = false }
function ctxPaste() { clipboard.paste({ value: () => xterm }, view, send, cmd); showContextMenu.value = false }
function ctxSelectAll() { if (view.value === 'classic' && xterm) { xterm.selectAll(); xterm.focus() } else inpRef.value?.select(); showContextMenu.value = false }
function copyBlock(id) { const c = bm.getBlockContent(id); if (c) clipboard.doCopy(c.output || c.command || '') }
function reExec(id) { const c = bm.getBlockContent(id); if (c?.command) { cmd.value = c.command; inpRef.value?.focus() } }
function clearAll() { bm.clearBlocks(); if (xterm) xterm.clear() }
function toggleRecording() { if (isRecording.value) { isRecording.value = false; if (recordingEntries.value.length > 0) saveRecording() } else { isRecording.value = true; recordingStartTime = Date.now(); recordingEntries.value = [] } }
function saveRecording() { const lines = recordingEntries.value.map(e => { const s = Math.floor(e.timestamp/1000); const m = Math.floor(s/60); const h = Math.floor(m/60); const t = `${h}:${String(m%60).padStart(2,'0')}:${String(s%60).padStart(2,'0')}`; return e.type === 'command' ? `[$t] $ ${e.data}` : `[$t] ${e.data}` }); const blob = new Blob([`Session Recording\nStart: ${new Date(recordingStartTime).toLocaleString()}\nEntries: ${recordingEntries.value.length}\n---\n${lines.join('\n')}`], { type: 'text/plain' }); const a = document.createElement('a'); a.href = URL.createObjectURL(blob); a.download = `recording-${sessionId}-${Date.now()}.txt`; document.body.appendChild(a); a.click(); document.body.removeChild(a) }
</script>

<style scoped>
.term-app { height: 100%; display: flex; flex-direction: column; background: var(--bg-terminal); position: relative; }
.tb { height: 30px; display: flex; align-items: center; justify-content: space-between; padding: 0 8px; background: var(--bg-toolbar); border-bottom: 1px solid var(--border-default); flex-shrink: 0; }
.tb-l, .tb-c, .tb-r { display: flex; align-items: center; gap: 4px; }
.led { width: 6px; height: 6px; border-radius: 50%; }
.led.active { background: var(--accent-success); box-shadow: 0 0 4px rgba(76,175,80,.5); }
.led.disconnected, .led.error { background: var(--accent-danger); }
.led.starting { background: var(--accent-warning); animation: pulse 1s infinite; }
.st { font-size: 11px; color: var(--text-secondary); }
.st.active { color: var(--accent-success); }
.s2 { font-size: 10px; color: var(--text-muted); }
.tbb { display: flex; align-items: center; justify-content: center; width: 24px; height: 24px; border: none; border-radius: 4px; color: var(--text-muted); cursor: pointer; font-size: 12px; background: transparent; }
.tbb:hover { background: var(--surface-hover); color: var(--text-primary); }
.tbb.recording { color: var(--accent-danger); animation: pulse 1s infinite; }
.tbb.vw { font-size: 14px; width: 26px; }
.tbb.vw.a { background: var(--surface-hover); color: var(--accent-success); }
.sep { width: 1px; height: 14px; background: var(--surface-hover); margin: 0 4px; }
.view-block { flex: 1; display: flex; flex-direction: column; min-height: 0; overflow: hidden; }
.bp { flex: 1; min-height: 0; overflow-y: auto; overflow-x: hidden; padding: 0; scroll-behavior: smooth; }
.empty { display: flex; align-items: center; justify-content: center; height: 100%; color: var(--text-muted); font-size: 13px; }
.shortcut-bar { display: flex; align-items: center; gap: 12px; padding: 5px 10px; background: var(--bg-panel-solid); border-top: 1px solid var(--border-default); flex-shrink: 0; overflow-x: auto; }
.shortcut { display: flex; align-items: center; gap: 4px; font-size: 10px; color: var(--text-muted); white-space: nowrap; }
.shortcut kbd { display: inline-flex; align-items: center; min-width: 18px; height: 16px; padding: 0 4px; background: var(--border-default); border: 1px solid var(--border-default); border-radius: 3px; font-size: 9px; color: var(--text-secondary); font-family: inherit; }
.shortcut-more { background: transparent; border: none; color: var(--primary-light, #7aa2f7); font-size: 10px; cursor: pointer; padding: 2px 6px; border-radius: 3px; }
.shortcut-more:hover { background: var(--primary-bg); }
.shortcut.app { color: var(--accent-success); }
.shortcut.app kbd { background: var(--success-bg); border-color: var(--border-success); color: var(--accent-success); }
.view-classic { flex: 1; min-height: 0; overflow: hidden; }
.view-classic :deep(.xterm) { height: 100%; width: 100%; }
.view-classic :deep(.xterm-viewport) { overflow-y: auto !important; background-color: var(--bg-terminal) !important; }
.view-classic :deep(.xterm-screen) { padding: 0; }
.view-classic :deep(.xterm-rows) { color: var(--text-primary) !important; }
.view-classic :deep(.xterm-rows .xterm-cursor-block) { background-color: var(--text-primary) !important; color: var(--bg-terminal) !important; }
.view-classic :deep(.xterm-rows .xterm-cursor-outline) { outline-color: var(--text-primary) !important; }
.view-classic :deep(.xterm-rows .xterm-cursor-bar) { box-shadow: 1px 0 0 var(--text-primary) inset !important; }
.view-classic :deep(.xterm-rows .xterm-cursor-underline) { border-bottom-color: var(--text-primary) !important; }
.view-classic :deep(.focus .xterm-selection div) { background-color: var(--surface-hover) !important; }
.view-classic :deep(.xterm-selection div) { background-color: var(--surface-hover) !important; }
.inp { display: flex; align-items: center; padding: 6px 10px; background: var(--bg-toolbar); border-top: 1px solid var(--border-default); gap: 8px; flex-shrink: 0; }
.iw { flex: 1; position: relative; }
.ps { font-family: monospace; font-size: 13px; color: var(--accent-success); }
.ci { flex: 1; background: transparent; border: none; outline: none; color: var(--text-primary); font-family: 'Cascadia Code', monospace; font-size: 13px; width: 100%; }
.ci::placeholder { color: var(--text-muted); }
.ci:disabled { opacity: .4; }
.ci-textarea { resize: none; min-height: 20px; max-height: 120px; overflow-y: auto; line-height: 1.4; }
.send-btn { display: flex; align-items: center; justify-content: center; width: 28px; height: 28px; background: var(--primary-bg); border: 1px solid var(--border-accent); border-radius: 4px; color: var(--primary-light); cursor: pointer; transition: all 0.15s; flex-shrink: 0; }
.send-btn:hover:not(:disabled) { background: var(--primary-bg-hover); }
.send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.bp::-webkit-scrollbar { width: 6px; }
.bp::-webkit-scrollbar-track { background: transparent; }
.bp::-webkit-scrollbar-thumb { background: var(--surface-hover); border-radius: 3px; }
.bp::-webkit-scrollbar-thumb:hover { background: var(--text-muted); }
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.4} }
.view-classic :deep(.xterm-decoration-top), .view-classic :deep(.xterm-decoration-bottom), .view-classic :deep(.xterm-find-result-decoration) { background: rgba(255, 152, 0, 0.5) !important; }
.view-classic :deep(.xterm-decoration-over), .view-classic :deep(.xterm-find-result-selected-decoration) { background: rgba(255, 152, 0, 0.8) !important; }
.search-bar { position: absolute; top: 32px; right: 8px; z-index: 100; }
.search-box { display: flex; align-items: center; gap: 6px; padding: 6px 10px; background: var(--bg-toolbar); border: 1px solid var(--border-default); border-radius: 6px; }
.search-box svg { color: var(--text-muted); flex-shrink: 0; }
.search-input { width: 200px; background: transparent; border: none; outline: none; color: var(--text-primary); font-size: 12px; }
.search-input::placeholder { color: var(--text-muted); }
.search-count { font-size: 10px; color: var(--text-muted); white-space: nowrap; }
.search-btn { display: flex; align-items: center; justify-content: center; width: 20px; height: 20px; background: transparent; border: none; border-radius: 3px; color: var(--text-muted); cursor: pointer; }
.search-btn:hover { background: var(--surface-hover); color: var(--text-secondary); }
.search-btn.close:hover { background: rgba(244,67,54,.15); color: var(--accent-danger); }
:deep(.terminal-block.highlight) { border-color: var(--accent-warning) !important; box-shadow: 0 0 8px rgba(255,152,0,.3); }
.slide-down-enter-active { transition: all .2s ease; }
.slide-down-leave-active { transition: all .15s ease; }
.slide-down-enter-from, .slide-down-leave-to { opacity: 0; transform: translateY(-8px); }
</style>
