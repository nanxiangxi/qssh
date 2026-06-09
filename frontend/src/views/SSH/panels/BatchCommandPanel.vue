<template>
  <div class="bc-panel">
    <!-- 工具栏 -->
    <div class="bc-toolbar">
      <div class="bc-toolbar-left">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#ed8936" stroke-width="2">
          <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
          <line x1="12" y1="5" x2="20" y2="5"/>
        </svg>
        <span class="bc-title">批量命令</span>
      </div>
      <div class="bc-toolbar-right">
        <button class="bc-btn" @click="refreshAll" :disabled="refreshing">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          刷新
        </button>
      </div>
    </div>

    <!-- 目标列表 -->
    <div class="bc-targets">
      <div class="bc-targets-header">
        <span class="bc-targets-title">目标连接</span>
        <div class="bc-targets-actions">
          <button class="bc-link-btn" @click="selectAll">全选</button>
          <button class="bc-link-btn" @click="selectNone">取消</button>
        </div>
      </div>

      <div v-if="connections.length === 0" class="bc-empty-hint">
        当前组没有已连接的服务器
      </div>

      <div v-else class="bc-conn-grid">
        <div
          v-for="conn in connections"
          :key="conn.id"
          class="bc-conn-card"
          :class="{ selected: isSelected(conn.id), offline: conn.status !== 'connected' }"
          @click="toggleConn(conn.id)"
        >
          <div class="bc-conn-head">
            <input
              type="checkbox"
              :checked="isSelected(conn.id)"
              :disabled="conn.status !== 'connected'"
              class="bc-check"
              @click.stop
              @change="toggleConn(conn.id)"
            />
            <span class="bc-conn-name">{{ conn.name }}</span>
            <span class="bc-conn-badge" :class="conn.status === 'connected' ? 'bc-on' : 'bc-off'">
              {{ conn.status === 'connected' ? '在线' : '离线' }}
            </span>
          </div>
          <!-- 终端选择 -->
          <div v-if="isSelected(conn.id) && conn.status === 'connected'" class="bc-term-select" @click.stop>
            <label class="bc-term-label">目标终端:</label>
            <select
              :value="getSelectedTerminal(conn.id)"
              @change="setSelectedTerminal(conn.id, $event.target.value)"
              @click.stop
              class="bc-term-dropdown"
            >
              <option value="auto">自动选择</option>
              <option v-for="t in getConnTerminals(conn.id)" :key="t.sessionId" :value="t.sessionId">
                {{ getTerminalTitle(t) }}
              </option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- 命令输入 -->
    <div class="bc-input-area">
      <textarea
        v-model="command"
        class="bc-input"
        placeholder="输入要批量执行的命令..."
        rows="2"
      ></textarea>
      <button
        class="bc-btn bc-btn-primary bc-send-btn"
        @click="executeCommand"
        :disabled="!command.trim() || selectedIds.size === 0 || executing"
      >
        {{ executing ? '发送中...' : `发送 (${selectedIds.size})` }}
      </button>
    </div>

    <!-- 执行结果 -->
    <div v-if="results.length > 0" class="bc-results">
      <div class="bc-results-header">
        <span>执行结果</span>
        <button class="bc-link-btn" @click="results = []">清空</button>
      </div>
      <div class="bc-result-scroll">
        <div v-for="(r, i) in results" :key="i" class="bc-result-row">
          <span class="bc-result-icon">
            <svg v-if="r.success" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#68d391" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#fc8181" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
          </span>
          <span class="bc-result-name">{{ r.name }}</span>
          <span class="bc-result-msg">{{ r.success ? '已发送' : r.error }}</span>
        </div>
      </div>
    </div>

    <!-- 底部 -->
    <div class="bc-footer">
      <span>{{ connections.filter(c => c.status === 'connected').length }}/{{ connections.length }} 在线</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, inject } from 'vue'
import { useSSHTabsStore } from '../../../stores/sshTabs'
import { useTerminalSessionStore } from '../../../stores/terminalSessions'
import { Events } from '@wailsio/runtime'
import * as SSHService from '../../../../bindings/changeme/ssh/sshservice.js'
import { showMessage } from '../../../utils/message'
import { addLog, LogLevel } from '../../../utils/logger'

const connId = inject('connId')
const tabsStore = useSSHTabsStore()
const sessionStore = useTerminalSessionStore()

const command = ref('')
const executing = ref(false)
const results = ref([])
const refreshing = ref(false)

// 选中的连接 ID 集合
const selectedIds = ref(new Set())
// 每个连接选择的终端 { connId: sessionId | 'auto' }
const terminalSelections = ref({})
// 每个连接的可用终端列表 { connId: [{ sessionId, title, isAI, panelId }, ...] }
const connTerminals = ref({})

// 当前组的所有连接（来自 tabs store）
const connections = computed(() => tabsStore.tabs || [])

const isSelected = (id) => selectedIds.value.has(id)

const toggleConn = (id) => {
  const conn = connections.value.find(c => c.id === id)
  if (!conn || conn.status !== 'connected') return
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id)
  } else {
    selectedIds.value.add(id)
  }
  selectedIds.value = new Set(selectedIds.value)
}

const selectAll = () => {
  selectedIds.value = new Set(
    connections.value.filter(c => c.status === 'connected').map(c => c.id)
  )
}

const selectNone = () => {
  selectedIds.value = new Set()
}

const getSelectedTerminal = (cid) => terminalSelections.value[cid] || 'auto'
const setSelectedTerminal = (cid, val) => { terminalSelections.value[cid] = val }
const getConnTerminals = (cid) => getReadyTerminals(cid)
const getTerminalTitle = (t) => t.isAI ? `[AI] ${t.title}` : t.title

// 刷新（请求 DockviewLayout 重发终端列表）
const refreshAll = () => {
  refreshing.value = true
  // 终端列表由 dockview:terminals-changed 事件自动更新
  setTimeout(() => { refreshing.value = false }, 500)
}

// 监听终端列表变化（来自 DockviewLayout 的 emitChange）
const onTerminalsChanged = (e) => {
  const d = e?.data
  if (!d || !d.connId || !Array.isArray(d.terminals)) return
  connTerminals.value = {
    ...connTerminals.value,
    [d.connId]: d.terminals.filter(t => t.sessionId)
  }
}

// 获取可用终端列表（只包含 shell 已启动的）
const getReadyTerminals = (cid) => {
  return (connTerminals.value[cid] || []).filter(t => {
    const sess = sessionStore.getSession(t.id)
    return sess && sess.ready
  })
}

const executeCommand = async () => {
  if (!command.value.trim() || selectedIds.value.size === 0) return
  executing.value = true
  results.value = []
  const cmd = command.value.trim()
  const targets = connections.value.filter(c => selectedIds.value.has(c.id))

  addLog(connId, 'terminal', LogLevel.INFO, `批量命令: "${cmd}" → ${targets.length} 个连接`)

  const promises = targets.map(async (conn) => {
    try {
      let sid = terminalSelections.value[conn.id]
      if (!sid || sid === 'auto') {
        const ready = getReadyTerminals(conn.id)
        if (ready.length === 0) {
          return { name: conn.name, success: false, error: '无就绪终端' }
        }
        sid = ready[0].sessionId
      } else {
        // 检查指定终端是否就绪
        const ready = getReadyTerminals(conn.id)
        if (!ready.find(t => t.sessionId === sid)) {
          return { name: conn.name, success: false, error: '终端未就绪' }
        }
      }
      // 通知终端创建结构化块
      Events.Emit('ai:terminal-exec-start', { connId: conn.id, sessionID: sid, command: cmd })
      await SSHService.WriteToTerminalByID(conn.id, sid, cmd + '\n')
      return { name: conn.name, success: true }
    } catch (e) {
      return { name: conn.name, success: false, error: e?.message || String(e) }
    }
  })

  results.value = await Promise.all(promises)
  executing.value = false

  const ok = results.value.filter(r => r.success).length
  showMessage(`命令已发送到 ${ok}/${targets.length} 个连接`, ok > 0 ? 'success' : 'error')
}

onMounted(() => {
  Events.On('dockview:terminals-changed', onTerminalsChanged)
})
onUnmounted(() => {
  Events.Off('dockview:terminals-changed', onTerminalsChanged)
})
</script>

<style scoped>
.bc-panel {
  width: 100%; height: 100%;
  display: flex; flex-direction: column;
  background: rgba(30, 30, 30, 0.95);
  overflow: hidden; font-size: 13px;
}

/* 工具栏 */
.bc-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 14px; border-bottom: 1px solid rgba(255,255,255,0.08);
  background: rgba(40,40,40,0.5); flex-shrink: 0;
}
.bc-toolbar-left { display: flex; align-items: center; gap: 8px; }
.bc-toolbar-right { display: flex; gap: 6px; }
.bc-title { color: #e2e8f0; font-weight: 600; font-size: 14px; }

.bc-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 12px; background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.1); border-radius: 6px;
  color: #a0aec0; font-size: 12px; cursor: pointer; transition: all 0.15s;
}
.bc-btn:hover:not(:disabled) { background: rgba(255,255,255,0.1); color: #e2e8f0; }
.bc-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.bc-btn-primary { background: rgba(237,137,54,0.2); border-color: rgba(237,137,54,0.4); color: #f6ad55; }
.bc-btn-primary:hover:not(:disabled) { background: rgba(237,137,54,0.35); }

/* 目标区域 */
.bc-targets {
  flex: 1; min-height: 0; overflow-y: auto; display: flex; flex-direction: column;
}
.bc-targets-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 14px; flex-shrink: 0;
}
.bc-targets-title { color: #a0aec0; font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; }
.bc-targets-actions { display: flex; gap: 8px; }
.bc-link-btn { background: none; border: none; color: #63b3ed; font-size: 11px; cursor: pointer; padding: 0; }
.bc-link-btn:hover { text-decoration: underline; }

.bc-empty-hint { padding: 20px; text-align: center; color: #4a5568; font-size: 12px; }

/* 连接卡片网格 */
.bc-conn-grid {
  display: grid; grid-template-columns: 1fr 1fr; gap: 8px;
  padding: 0 14px 10px; flex: 1; align-content: start; overflow-y: auto;
}

.bc-conn-card {
  background: rgba(45,45,45,0.6); border: 1px solid rgba(255,255,255,0.06);
  border-radius: 8px; padding: 10px; cursor: pointer; transition: all 0.15s;
}
.bc-conn-card:hover { background: rgba(55,55,55,0.8); border-color: rgba(255,255,255,0.12); }
.bc-conn-card.selected { border-color: rgba(237,137,54,0.5); background: rgba(237,137,54,0.06); }
.bc-conn-card.offline { opacity: 0.5; cursor: not-allowed; }

.bc-conn-head {
  display: flex; align-items: center; gap: 8px;
}
.bc-check { accent-color: #ed8936; }
.bc-conn-name { flex: 1; color: #e2e8f0; font-size: 12px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bc-conn-badge {
  font-size: 10px; padding: 1px 6px; border-radius: 4px; font-weight: 600;
}
.bc-on { background: rgba(72,187,120,0.15); color: #68d391; }
.bc-off { background: rgba(160,174,192,0.1); color: #718096; }

/* 终端选择 */
.bc-term-select {
  display: flex; align-items: center; gap: 6px; margin-top: 8px;
}
.bc-term-label { font-size: 10px; color: #718096; white-space: nowrap; }
.bc-term-dropdown {
  flex: 1; background: rgba(30,30,30,0.8); border: 1px solid rgba(255,255,255,0.08);
  border-radius: 4px; color: #e2e8f0; font-size: 11px; padding: 3px 6px;
  font-family: monospace; outline: none; cursor: pointer;
}
.bc-term-dropdown:focus { border-color: rgba(237,137,54,0.4); }
.bc-term-dropdown option { background: #1a1a1a; color: #e2e8f0; }

/* 命令输入 */
.bc-input-area {
  padding: 10px 14px; border-top: 1px solid rgba(255,255,255,0.06);
  display: flex; gap: 8px; align-items: stretch; flex-shrink: 0;
}
.bc-input {
  flex: 1; background: rgba(30,30,30,0.8);
  border: 1px solid rgba(255,255,255,0.1); border-radius: 6px;
  color: #e2e8f0; font-size: 13px; font-family: 'Courier New', monospace;
  padding: 8px 10px; outline: none; resize: none;
}
.bc-input:focus { border-color: rgba(237,137,54,0.4); }
.bc-input::placeholder { color: #4a5568; }
.bc-send-btn { align-self: stretch; white-space: nowrap; }

/* 结果 */
.bc-results {
  border-top: 1px solid rgba(255,255,255,0.06); flex-shrink: 0;
  max-height: 35%; display: flex; flex-direction: column;
}
.bc-results-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 14px; color: #a0aec0; font-size: 12px; font-weight: 600;
  border-bottom: 1px solid rgba(255,255,255,0.04); flex-shrink: 0;
}
.bc-result-scroll { overflow-y: auto; padding: 6px 14px; }
.bc-result-row {
  display: flex; align-items: center; gap: 8px;
  padding: 5px 0; border-bottom: 1px solid rgba(255,255,255,0.03);
}
.bc-result-icon { flex-shrink: 0; display: flex; }
.bc-result-name { color: #e2e8f0; font-size: 12px; font-weight: 500; min-width: 80px; }
.bc-result-msg { color: #718096; font-size: 11px; flex: 1; }

/* 底部 */
.bc-footer {
  display: flex; align-items: center; padding: 4px 14px;
  border-top: 1px solid rgba(255,255,255,0.08); background: rgba(40,40,40,0.5);
  font-size: 11px; color: #718096; flex-shrink: 0;
}
</style>
