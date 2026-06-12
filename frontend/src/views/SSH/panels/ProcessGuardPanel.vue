<template>
  <div class="pg-panel">
    <!-- 工具栏 -->
    <div class="pg-toolbar">
      <div class="pg-toolbar-left">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        <span class="pg-title">进程守护</span>
        <span v-if="processes.length" class="pg-badge">{{ processes.length }}</span>
      </div>
      <div class="pg-toolbar-right">
        <button class="pg-btn" :class="{ 'pg-btn-active': autoRefresh }" @click="toggleAutoRefresh" title="自动刷新">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
        </button>
        <button class="pg-btn" @click="showAdd = true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span class="pg-btn-text">创建守护</span>
        </button>
        <button class="pg-btn" @click="loadProcesses" :disabled="loading">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          <span class="pg-btn-text">刷新</span>
        </button>
      </div>
    </div>

    <!-- 内容区 -->
    <div class="pg-content">
      <div v-if="loading && !processes.length" class="pg-loading">
        <div class="pg-spinner"></div>
        <p>正在检测守护进程...</p>
      </div>

      <div v-else-if="processes.length === 0" class="pg-empty">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        <p>暂无守护进程</p>
        <p class="pg-hint">创建守护进程可确保关键服务持续运行</p>
      </div>

      <div v-else class="pg-list">
        <div v-for="proc in processes" :key="proc.id" class="pg-card" :class="'pg-status-' + proc.status">
          <div class="pg-card-header">
            <div class="pg-card-info">
              <span class="pg-card-name">{{ proc.name }}</span>
              <span class="pg-status-tag" :class="'pg-st-' + proc.status">{{ statusLabel(proc.status) }}</span>
            </div>
            <div class="pg-card-actions">
              <button v-if="proc.status === 'running'" class="pg-btn pg-btn-sm pg-btn-warn" @click="stopProcess(proc)" title="停止">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
              </button>
              <button v-else class="pg-btn pg-btn-sm pg-btn-success" @click="startProcess(proc)" title="启动">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              </button>
              <button class="pg-btn pg-btn-sm" @click="restartProcess(proc)" title="重启">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
              </button>
              <button class="pg-btn pg-btn-sm" @click="viewLogs(proc)" title="日志">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              </button>
              <button class="pg-btn pg-btn-sm pg-btn-danger" @click="deleteProcess(proc)" title="删除">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          </div>
          <div class="pg-card-cmd">{{ proc.command }}</div>
          <div class="pg-card-meta">
            <span v-if="proc.pid" class="pg-meta-item">PID {{ proc.pid }}</span>
            <span v-if="proc.restarts > 0" class="pg-meta-item pg-restart">重启 {{ proc.restarts }} 次</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部 -->
    <div class="pg-footer">
      <span v-if="autoRefresh" class="pg-auto-badge">自动刷新</span>
      <span class="pg-footer-info">{{ runningCount }} 运行 / {{ processes.length }} 总计</span>
    </div>

    <!-- 创建守护弹窗 -->
    <Teleport to="body">
      <div v-if="showAdd" class="pg-modal-mask" @click.self="showAdd = false">
        <div class="pg-modal">
          <div class="pg-modal-head">
            <h3>创建守护进程</h3>
            <button class="pg-modal-close" @click="showAdd = false">&times;</button>
          </div>
          <div class="pg-modal-body">
            <div class="pg-field">
              <label>名称</label>
              <input v-model="addForm.name" class="pg-input" placeholder="如 my-app、nginx-monitor" />
            </div>
            <div class="pg-field">
              <label>启动命令</label>
              <input v-model="addForm.command" class="pg-input pg-mono" placeholder="如 python3 /opt/app/main.py" />
            </div>
            <div class="pg-field">
              <label>工作目录</label>
              <input v-model="addForm.workDir" class="pg-input pg-mono" placeholder="留空默认 /tmp" />
            </div>
            <div class="pg-field">
              <label class="pg-checkbox-label">
                <input type="checkbox" v-model="addForm.autoRestart" />
                <span>异常退出自动重启（3秒后重试）</span>
              </label>
            </div>
            <div v-if="addError" class="pg-error">{{ addError }}</div>
          </div>
          <div class="pg-modal-foot">
            <button class="pg-btn" @click="showAdd = false">取消</button>
            <button class="pg-btn pg-btn-primary" @click="createProcess" :disabled="addSubmitting">
              {{ addSubmitting ? '创建中...' : '创建' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 日志弹窗 -->
    <Teleport to="body">
      <div v-if="showLogs" class="pg-modal-mask" @click.self="showLogs = false">
        <div class="pg-modal pg-modal-wide">
          <div class="pg-modal-head">
            <h3>{{ logName }} - 运行日志</h3>
            <button class="pg-modal-close" @click="showLogs = false">&times;</button>
          </div>
          <div class="pg-modal-body">
            <pre ref="logRef" class="pg-log-content">{{ logContent || '暂无日志' }}</pre>
          </div>
          <div class="pg-modal-foot">
            <button class="pg-btn" @click="showLogs = false">关闭</button>
            <button class="pg-btn pg-btn-primary" @click="refreshLogs">刷新</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, inject, nextTick } from 'vue'
import * as GuardianService from '../../../../bindings/changeme/ssh/processguardianservice.js'
import { showMessage } from '../../../utils/message'
import { useConfirm } from '../../../utils/confirm'
import { addLog, LogLevel } from '../../../utils/logger'

const connId = inject('connId')
const { confirm } = useConfirm()

const loading = ref(false)
const processes = ref([])
const autoRefresh = ref(false)
let refreshTimer = null

const showAdd = ref(false)
const addSubmitting = ref(false)
const addError = ref('')
const addForm = reactive({ name: '', command: '', workDir: '', autoRestart: true })

const showLogs = ref(false)
const logName = ref('')
const logContent = ref('')
const logRef = ref(null)

const runningCount = computed(() => processes.value.filter(p => p.status === 'running').length)

const statusLabel = (s) => ({
  running: '运行中', stopped: '已停止', failed: '失败', unknown: '未知'
}[s] || s)

const loadProcesses = async () => {
  if (!connId) return
  loading.value = true
  try {
    processes.value = await GuardianService.GetGuardians(connId) || []
  } catch (e) {
    console.error('[ProcessGuardPanel] 加载失败:', e)
  } finally {
    loading.value = false
  }
}

const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    refreshTimer = setInterval(loadProcesses, 5000)
    showMessage('已开启自动刷新（5秒）', 'info')
  } else {
    clearInterval(refreshTimer)
    refreshTimer = null
    showMessage('已关闭自动刷新', 'info')
  }
}

const createProcess = async () => {
  if (!addForm.name.trim() || !addForm.command.trim()) {
    addError.value = '请填写名称和命令'
    return
  }
  addSubmitting.value = true
  addError.value = ''
  try {
    await GuardianService.CreateGuardian(connId, addForm.name, addForm.command, addForm.workDir, addForm.autoRestart)
    addLog(connId, 'guardian', LogLevel.SUCCESS, `创建守护进程: ${addForm.name} → ${addForm.command}`)
    showMessage('守护进程已创建', 'success')
    showAdd.value = false
    addForm.name = ''
    addForm.command = ''
    addForm.workDir = ''
    addForm.autoRestart = true
    await loadProcesses()
  } catch (e) {
    addError.value = String(e?.message || e)
    addLog(connId, 'guardian', LogLevel.ERROR, '创建守护进程失败: ' + (e?.message || e))
  } finally {
    addSubmitting.value = false
  }
}

const startProcess = async (proc) => {
  try {
    await GuardianService.StartGuardian(connId, proc.name)
    addLog(connId, 'guardian', LogLevel.INFO, `启动守护进程: ${proc.name}`)
    showMessage('已启动', 'success')
    await loadProcesses()
  } catch (e) {
    addLog(connId, 'guardian', LogLevel.ERROR, `启动失败: ${proc.name}`)
    showMessage('启动失败: ' + (e?.message || e), 'error')
  }
}

const stopProcess = async (proc) => {
  const ok = await confirm({ title: '停止进程', message: `确定停止 ${proc.name}？` })
  if (!ok) return
  try {
    await GuardianService.StopGuardian(connId, proc.name)
    addLog(connId, 'guardian', LogLevel.WARNING, `停止守护进程: ${proc.name}`)
    showMessage('已停止', 'success')
    await loadProcesses()
  } catch (e) {
    showMessage('停止失败: ' + (e?.message || e), 'error')
  }
}

const restartProcess = async (proc) => {
  try {
    await GuardianService.RestartGuardian(connId, proc.name)
    addLog(connId, 'guardian', LogLevel.INFO, `重启守护进程: ${proc.name}`)
    showMessage('已重启', 'success')
    await loadProcesses()
  } catch (e) {
    showMessage('重启失败: ' + (e?.message || e), 'error')
  }
}

const deleteProcess = async (proc) => {
  const ok = await confirm({ title: '删除守护进程', message: `确定删除 ${proc.name}？\n将停止服务并删除所有配置和日志。`, danger: true })
  if (!ok) return
  try {
    await GuardianService.DeleteGuardian(connId, proc.name)
    addLog(connId, 'guardian', LogLevel.WARNING, `删除守护进程: ${proc.name}`)
    showMessage('已删除', 'success')
    await loadProcesses()
  } catch (e) {
    showMessage('删除失败: ' + (e?.message || e), 'error')
  }
}

const viewLogs = async (proc) => {
  logName.value = proc.name
  logContent.value = ''
  showLogs.value = true
  try {
    logContent.value = await GuardianService.GetGuardianLogs(connId, proc.name, 500) || '暂无日志'
    await nextTick()
    if (logRef.value) logRef.value.scrollTop = logRef.value.scrollHeight
  } catch (e) {
    logContent.value = '获取日志失败: ' + (e?.message || e)
  }
}

const refreshLogs = async () => {
  if (!logName.value) return
  try {
    logContent.value = await GuardianService.GetGuardianLogs(connId, logName.value, 500) || '暂无日志'
    await nextTick()
    if (logRef.value) logRef.value.scrollTop = logRef.value.scrollHeight
  } catch (e) {}
}

onMounted(() => loadProcesses())
onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<style scoped>
.pg-panel {
  width: 100%; height: 100%;
  display: flex; flex-direction: column;
  background: var(--bg-panel);
  overflow: hidden;
}

.pg-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-default);
  background: var(--surface-2);
  flex-shrink: 0;
}
.pg-toolbar-left { display: flex; align-items: center; gap: 0.5rem; color: var(--accent-purple); }
.pg-toolbar-right { display: flex; gap: 0.375rem; }
.pg-title { color: var(--text-primary); font-weight: 600; font-size: 0.875rem; }
.pg-badge {
  padding: 0.125rem 0.5rem; border-radius: 0.25rem;
  font-size: 0.625rem; font-weight: 600;
  background: color-mix(in srgb, var(--accent-purple), transparent 80%);
  color: var(--accent-purple);
}

.pg-btn {
  display: inline-flex; align-items: center; gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 0.375rem;
  color: var(--text-secondary); font-size: 0.75rem;
  cursor: pointer; transition: all 0.15s;
}
.pg-btn:hover:not(:disabled) { background: var(--surface-hover); color: var(--text-primary); }
.pg-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.pg-btn-sm { padding: 0.25rem 0.5rem; }
.pg-btn-active { background: color-mix(in srgb, var(--accent-purple), transparent 80%); border-color: color-mix(in srgb, var(--accent-purple), transparent 60%); color: var(--accent-purple); }
.pg-btn-primary { background: color-mix(in srgb, var(--accent-purple), transparent 80%); border-color: color-mix(in srgb, var(--accent-purple), transparent 60%); color: var(--accent-purple); }
.pg-btn-primary:hover:not(:disabled) { background: color-mix(in srgb, var(--accent-purple), transparent 65%); }
.pg-btn-success { background: var(--success-bg); border-color: color-mix(in srgb, var(--accent-success), transparent 60%); color: var(--success-light); }
.pg-btn-warn { background: var(--warning-bg); border-color: color-mix(in srgb, var(--warning-light), transparent 60%); color: var(--warning-light); }
.pg-btn-danger { background: var(--danger-bg); border-color: color-mix(in srgb, var(--accent-danger), transparent 60%); color: var(--accent-danger); }

.pg-content { flex: 1; overflow-y: auto; min-height: 0; }

.pg-loading, .pg-empty {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 100%; gap: 0.75rem; color: var(--text-muted);
}
.pg-loading p, .pg-empty p { margin: 0; font-size: 0.875rem; }
.pg-hint { font-size: 0.75rem !important; color: var(--text-muted) !important; }

.pg-spinner {
  width: 24px; height: 24px;
  border: 2px solid color-mix(in srgb, var(--accent-purple), transparent 80%);
  border-top-color: var(--accent-purple);
  border-radius: 50%; animation: pg-spin 0.8s linear infinite;
}
@keyframes pg-spin { to { transform: rotate(360deg); } }

.pg-list { padding: 0.75rem; display: flex; flex-direction: column; gap: 0.5rem; }

.pg-card {
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: 0.5rem;
  padding: 0.75rem 1rem;
  transition: all 0.15s;
}
.pg-card:hover { background: var(--surface-2); }
.pg-card.pg-status-running { border-left: 3px solid var(--success-light); }
.pg-card.pg-status-stopped { border-left: 3px solid var(--text-muted); }
.pg-card.pg-status-failed { border-left: 3px solid var(--accent-danger); }

.pg-card-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 0.25rem;
}
.pg-card-info { display: flex; align-items: center; gap: 0.5rem; }
.pg-card-name { color: var(--text-primary); font-weight: 600; font-size: 0.875rem; }
.pg-card-actions { display: flex; gap: 0.25rem; }

.pg-status-tag {
  display: inline-block; padding: 0.0625rem 0.375rem;
  border-radius: 0.25rem; font-size: 0.625rem; font-weight: 600;
}
.pg-st-running { background: var(--success-bg); color: var(--success-light); }
.pg-st-stopped { background: var(--surface-2); color: var(--text-secondary); }
.pg-st-failed { background: var(--danger-bg); color: var(--accent-danger); }

.pg-card-cmd {
  color: var(--text-muted); font-size: 0.75rem;
  font-family: 'Courier New', monospace;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  margin-bottom: 0.25rem;
}

.pg-card-meta {
  display: flex; gap: 0.5rem;
}
.pg-meta-item {
  font-size: 0.625rem; color: var(--text-muted);
  font-family: monospace;
}
.pg-restart { color: var(--warning-light); }

.pg-footer {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.25rem 1rem;
  border-top: 1px solid var(--border-default);
  background: var(--surface-2);
  font-size: 0.625rem; flex-shrink: 0;
}
.pg-footer-info { color: var(--text-muted); }
.pg-auto-badge {
  padding: 0.0625rem 0.375rem;
  background: color-mix(in srgb, var(--accent-purple), transparent 85%);
  border-radius: 0.25rem;
  color: var(--accent-purple); font-size: 0.5625rem;
}

/* 弹窗 */
.pg-modal-mask {
  position: fixed; inset: 0;
  background: var(--bg-overlay); display: flex;
  align-items: center; justify-content: center;
  z-index: 10000; backdrop-filter: blur(4px);
}
.pg-modal {
  background: var(--bg-panel-solid); border: 1px solid var(--border-default);
  border-radius: 0.75rem; width: 400px; max-width: 90vw;
  box-shadow: var(--shadow-lg);
}
.pg-modal-wide { width: 600px; max-width: 90vw; }
.pg-modal-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--border-subtle);
}
.pg-modal-head h3 { margin: 0; color: var(--text-primary); font-size: 0.9375rem; }
.pg-modal-close {
  background: none; border: none; color: var(--text-muted); font-size: 1.5rem;
  cursor: pointer; padding: 0; line-height: 1; transition: color 0.15s;
}
.pg-modal-close:hover { color: var(--text-primary); }
.pg-modal-head-actions { display: flex; align-items: center; gap: 0.5rem; }
.pg-modal-head-actions button:last-child {
  background: none; border: none; color: var(--text-muted); font-size: 1.25rem; cursor: pointer;
}
.pg-modal-head-actions button:last-child:hover { color: var(--text-primary); }
.pg-modal-body { padding: 1rem 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; }
.pg-modal-foot {
  display: flex; justify-content: flex-end; gap: 0.5rem;
  padding: 0.875rem 1.25rem; border-top: 1px solid var(--border-subtle);
}
.pg-field label { display: block; color: var(--text-secondary); font-size: 0.75rem; margin-bottom: 0.25rem; }
.pg-input {
  width: 100%; background: var(--bg-input);
  border: 1px solid var(--border-default); border-radius: 0.375rem;
  color: var(--text-primary); font-size: 0.8125rem; padding: 0.5rem 0.75rem;
  outline: none; box-sizing: border-box;
}
.pg-input:focus { border-color: color-mix(in srgb, var(--accent-purple), transparent 60%); }
.pg-mono { font-family: 'Courier New', monospace; }
.pg-checkbox-label {
  display: flex; align-items: center; gap: 0.5rem;
  color: var(--text-secondary); font-size: 0.8125rem; cursor: pointer;
}
.pg-checkbox-label input { accent-color: var(--accent-purple); }
.pg-error { color: var(--accent-danger); font-size: 0.75rem; }

.pg-log-content {
  margin: 0; padding: 0.75rem;
  background: var(--surface-1);
  border-radius: 0.375rem;
  color: var(--text-secondary); font-size: 0.6875rem;
  font-family: 'Courier New', monospace;
  line-height: 1.5; white-space: pre-wrap;
  max-height: 50vh; overflow-y: auto;
}
</style>
