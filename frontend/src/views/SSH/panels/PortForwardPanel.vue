<template>
  <div class="pf-panel">
    <!-- 顶部工具栏 -->
    <div class="pf-toolbar">
      <div class="pf-toolbar-left">
        <svg class="pf-icon-accent" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/>
          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
          <line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/>
        </svg>
        <span class="pf-title">{{ t('portForward.title') }}</span>
      </div>
      <div class="pf-toolbar-right">
        <button class="pf-btn pf-btn-primary" @click="closeDialog(); showAdd = true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span class="pf-btn-text">{{ t('common.add') }}</span>
        </button>
      </div>
    </div>

    <!-- 转发列表 -->
    <div class="pf-list">
      <div v-if="forwards.length === 0" class="pf-empty">
        <svg class="pf-icon-muted" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
        <p>{{ t('portForward.noRules') }}</p>
        <button class="pf-btn pf-btn-primary pf-btn-sm" @click="closeDialog(); showAdd = true">{{ t('portForward.add') }}</button>
      </div>

      <div v-else class="pf-table-wrap">
        <table class="pf-table">
          <thead>
            <tr>
              <th>{{ t('portForward.name') }}</th>
              <th>{{ t('common.type') }}</th>
              <th>{{ t('portForward.listenAddress') }}</th>
              <th>{{ t('portForward.targetAddress') }}</th>
              <th>{{ t('portForward.connection') }}</th>
              <th>{{ t('portForward.traffic') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="fwd in forwards" :key="fwd.id" :class="'pf-row-' + fwd.status">
              <td class="pf-name-cell" :title="getForwardName(fwd)">
                {{ getForwardName(fwd) || '—' }}
              </td>
              <td>
                <span class="pf-type-tag" :class="'pf-type-' + fwd.type">
                  {{ fwd.type === 'local' ? t('portForward.remoteToLocal') : t('portForward.localToRemote') }}
                </span>
              </td>
              <td class="pf-mono">{{ fwd.bindAddr }}:{{ fwd.bindPort }}</td>
              <td class="pf-mono">{{ fwd.remoteHost }}:{{ fwd.remotePort }}</td>
              <td class="pf-stat">
                <span class="pf-conn-count" :class="{ 'pf-conn-active': fwd.activeConns > 0 }">{{ fwd.activeConns || 0 }}</span>
                <span class="pf-conn-total">/{{ fwd.totalConns || 0 }}</span>
              </td>
              <td class="pf-stat pf-traffic">
                <span class="pf-traffic-up">{{ formatBytes(fwd.bytesSent || 0) }}</span>
                <span class="pf-traffic-sep">/</span>
                <span class="pf-traffic-down">{{ formatBytes(fwd.bytesRecv || 0) }}</span>
              </td>
              <td>
                <span class="pf-status" :class="'pf-status-' + fwd.status">
                  {{ statusLabel(fwd.status) }}
                </span>
              </td>
              <td>
                <div class="pf-actions">
                  <button v-if="fwd.status !== 'running'" class="pf-btn pf-btn-sm pf-btn-success" @click="startForward(fwd.id)" :title="t('common.start')">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                  </button>
                  <button v-else class="pf-btn pf-btn-sm pf-btn-warn" @click="stopForward(fwd.id)" :title="t('common.stop')">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
                  </button>
                  <button v-if="fwd.status !== 'running'" class="pf-btn pf-btn-sm" @click="editForward(fwd)" :title="t('common.edit')">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </button>
                  <button class="pf-btn pf-btn-sm pf-btn-danger" @click="removeForward(fwd.id)" :title="t('common.delete')">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 底部状态栏 -->
    <div class="pf-footer">
      <span>{{ t('portForward.rules', { count: forwards.length }) }}</span>
      <span>{{ t('portForward.running', { count: runningCount }) }}</span>
    </div>

    <!-- 添加对话框 -->
    <Teleport to="body">
      <div v-if="showAdd" class="pf-modal-mask" @click.self="showAdd = false">
        <div class="pf-modal">
          <div class="pf-modal-head">
            <h3>{{ editingId ? t('portForward.editForward') : t('portForward.addForward') }}</h3>
            <button @click="closeDialog">&times;</button>
          </div>
          <div class="pf-modal-body">
            <div class="pf-field">
              <label>{{ t('portForward.name') }} <span class="pf-hint-inline">（{{ t('portForward.nameHint') }}）</span></label>
              <input v-model="form.name" :placeholder="t('portForward.namePlaceholder')" class="pf-input pf-input-full" />
            </div>
            <div class="pf-field">
              <label>{{ t('portForward.forwardType') }}</label>
              <div class="pf-radio-group">
                <label class="pf-radio" :class="{ active: form.type === 'local' }">
                  <input type="radio" v-model="form.type" value="local" />
                  <div class="pf-radio-content">
                    <span class="pf-radio-title">{{ t('portForward.remoteToLocal') }}</span>
                    <span class="pf-radio-desc">{{ t('portForward.remoteToLocalDesc') }}</span>
                  </div>
                </label>
                <label class="pf-radio" :class="{ active: form.type === 'remote' }">
                  <input type="radio" v-model="form.type" value="remote" />
                  <div class="pf-radio-content">
                    <span class="pf-radio-title">{{ t('portForward.localToRemote') }}</span>
                    <span class="pf-radio-desc">{{ t('portForward.localToRemoteDesc') }}</span>
                  </div>
                </label>
              </div>
            </div>
            <div class="pf-field">
              <label>{{ form.type === 'local' ? t('portForward.localListenPort') : t('portForward.remoteOpenPort') }}</label>
              <div class="pf-hint">{{ form.type === 'local' ? t('portForward.localListenPortHint') : t('portForward.remoteOpenPortHint') }}</div>
              <div class="pf-input-row">
                <input v-model="form.bindAddr" :placeholder="form.type === 'local' ? '127.0.0.1' : '0.0.0.0'" class="pf-input pf-input-addr" />
                <input v-model.number="form.bindPort" type="number" :placeholder="t('portForward.port')" class="pf-input pf-input-port" />
              </div>
            </div>
            <div class="pf-field">
              <label>{{ form.type === 'local' ? t('portForward.remoteTargetAddress') : t('portForward.localServiceAddress') }}</label>
              <div class="pf-hint">{{ form.type === 'local' ? t('portForward.remoteTargetAddressHint') : t('portForward.localServiceAddressHint') }}</div>
              <div class="pf-input-row">
                <input v-model="form.remoteHost" placeholder="127.0.0.1" class="pf-input pf-input-addr" />
                <input v-model.number="form.remotePort" type="number" :placeholder="t('portForward.port')" class="pf-input pf-input-port" />
              </div>
            </div>
            <div v-if="formError" class="pf-error">{{ formError }}</div>
          </div>
          <div class="pf-modal-foot">
            <button class="pf-btn" @click="closeDialog">{{ t('common.cancel') }}</button>
            <button class="pf-btn pf-btn-primary" @click="addForward" :disabled="submitting">
              {{ submitting ? (editingId ? t('portForward.saving') : t('portForward.adding')) : (editingId ? t('common.save') : t('common.add')) }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { Events } from '@wailsio/runtime'
import * as PortForwardService from '../../../../bindings/changeme/ssh/portforwardservice.js'
import { showMessage } from '../../../utils/message'
import { addLog, LogType, LogLevel } from '../../../utils/logger'

const { t } = useI18n()
const connId = inject('connId')

const forwards = ref([])
const showAdd = ref(false)
const submitting = ref(false)
const formError = ref('')
const editingId = ref(null)

const form = ref({
  name: '',
  type: 'local',
  bindAddr: '127.0.0.1',
  bindPort: null,
  remoteHost: '127.0.0.1',
  remotePort: null
})

// 名称存储（localStorage）
const NAMES_KEY = 'pf_names'
function loadNames() {
  try { return JSON.parse(localStorage.getItem(NAMES_KEY) || '{}') } catch { return {} }
}
function saveName(fwd, name) {
  if (!name) return
  const names = loadNames()
  names[fwd.id] = name
  localStorage.setItem(NAMES_KEY, JSON.stringify(names))
}
function getForwardName(fwd) {
  const names = loadNames()
  return names[fwd.id] || ''
}

const runningCount = computed(() => forwards.value.filter(f => f.status === 'running').length)

// 切换类型时自动设置默认绑定地址
watch(() => form.value.type, (type) => {
  form.value.bindAddr = type === 'local' ? '127.0.0.1' : '0.0.0.0'
})

const statusLabel = (s) => ({ running: t('common.active'), stopped: t('common.disconnected'), error: t('common.error') }[s] || s)

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(i > 0 ? 1 : 0) + ' ' + sizes[i]
}

const loadForwards = async () => {
  if (!connId) return
  try {
    const list = await PortForwardService.GetForwards(connId)
    forwards.value = list || []
  } catch (e) {
    console.error('[PortForwardPanel] 加载失败:', e)
  }
}

const editForward = (fwd) => {
  editingId.value = fwd.id
  form.value = {
    name: getForwardName(fwd),
    type: fwd.type,
    bindAddr: fwd.bindAddr,
    bindPort: fwd.bindPort,
    remoteHost: fwd.remoteHost,
    remotePort: fwd.remotePort
  }
  formError.value = ''
  showAdd.value = true
}

const addForward = async () => {
  formError.value = ''
  if (!form.value.bindPort || !form.value.remotePort) {
    formError.value = t('portForward.pleaseFillPort')
    return
  }
  submitting.value = true
  try {
    // 编辑模式：先删除旧规则
    if (editingId.value) {
      try {
        await PortForwardService.RemoveForward(editingId.value)
        // 清除旧名称
        const names = loadNames()
        delete names[editingId.value]
        localStorage.setItem(NAMES_KEY, JSON.stringify(names))
      } catch (e) {}
    }

    const fn = form.value.type === 'local' ? 'AddLocalForward' : 'AddRemoteForward'
    const result = await PortForwardService[fn](connId, form.value.bindAddr, form.value.bindPort, form.value.remoteHost, form.value.remotePort)

    // 保存名称
    if (form.value.name && result) {
      saveName(result, form.value.name)
    }

    const typeLabel = form.value.type === 'local' ? '远程→本地' : '本地→远程'
    addLog(connId, 'portForward', LogLevel.SUCCESS,
      `添加端口转发: ${typeLabel} ${form.value.bindAddr}:${form.value.bindPort} → ${form.value.remoteHost}:${form.value.remotePort}`)

    closeDialog()
    await loadForwards()
  } catch (e) {
    formError.value = String(e?.message || e)
  } finally {
    submitting.value = false
  }
}

const closeDialog = () => {
  showAdd.value = false
  editingId.value = null
  form.value.name = ''
  form.value.bindPort = null
  form.value.remotePort = null
  form.value.bindAddr = '127.0.0.1'
  form.value.remoteHost = '127.0.0.1'
  form.value.type = 'local'
  formError.value = ''
}

const startForward = async (id) => {
  try {
    await PortForwardService.StartForward(id)
    addLog(connId, 'portForward', LogLevel.INFO, '启动端口转发: ' + id)
    showMessage(t('portForward.forwardStarted'), 'success')
    await loadForwards()
  } catch (e) {
    addLog(connId, 'portForward', LogLevel.ERROR, '启动端口转发失败: ' + (e?.message || e))
    showMessage(t('portForward.startFailed') + ': ' + (e?.message || e), 'error')
  }
}

const stopForward = async (id) => {
  try {
    await PortForwardService.StopForward(id)
    addLog(connId, 'portForward', LogLevel.INFO, '停止端口转发: ' + id)
    showMessage(t('portForward.forwardStopped'), 'success')
    await loadForwards()
  } catch (e) {
    addLog(connId, 'portForward', LogLevel.ERROR, '停止端口转发失败: ' + (e?.message || e))
    showMessage(t('portForward.stopFailed') + ': ' + (e?.message || e), 'error')
  }
}

const removeForward = async (id) => {
  try {
    await PortForwardService.RemoveForward(id)
    // 清除名称
    const names = loadNames()
    delete names[id]
    localStorage.setItem(NAMES_KEY, JSON.stringify(names))
    addLog(connId, 'portForward', LogLevel.INFO, '删除端口转发: ' + id)
    showMessage(t('portForward.forwardDeleted'), 'success')
    await loadForwards()
  } catch (e) {
    addLog(connId, 'portForward', LogLevel.ERROR, '删除端口转发失败: ' + (e?.message || e))
    showMessage(t('portForward.deleteFailed') + ': ' + (e?.message || e), 'error')
  }
}

const onStatusEvent = (e) => {
  const d = e?.data
  if (!d || d.connId !== connId) return
  forwards.value = d.forwards || []
}

onMounted(async () => {
  await loadForwards()
  Events.On('port-forward:status', onStatusEvent)
})

onUnmounted(() => {
  Events.Off('port-forward:status', onStatusEvent)
})
</script>

<style scoped>
.pf-panel {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-panel);
  overflow: hidden;
}

/* 工具栏 */
.pf-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--surface-hover);
  background: var(--toolbar-3);
  flex-shrink: 0;
}

.pf-toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pf-icon-accent {
  color: var(--primary-light);
}

.pf-icon-muted {
  color: var(--text-disabled);
}

.pf-title {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 0.875rem;
}

/* 列表 */
.pf-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem;
  min-height: 0;
}

.pf-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 0.75rem;
  color: var(--text-muted);
}

.pf-empty p {
  margin: 0;
  font-size: 0.875rem;
}

/* 表格 */
.pf-table-wrap {
  overflow-x: auto;
}

.pf-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.75rem;
}

.pf-table th {
  padding: 0.5rem 0.75rem;
  text-align: left;
  color: var(--text-secondary);
  font-weight: 600;
  border-bottom: 2px solid var(--surface-hover);
  white-space: nowrap;
}

.pf-table td {
  padding: 0.5rem 0.75rem;
  color: var(--text-primary);
  border-bottom: 1px solid var(--surface-1);
  white-space: nowrap;
}

.pf-mono {
  font-family: 'Courier New', monospace;
  font-size: 0.6875rem;
}

.pf-row-running td {
  background: var(--success-bg);
}

.pf-row-error td {
  background: var(--danger-bg);
}

/* 连接和流量统计 */
.pf-stat {
  font-size: 0.6875rem;
  font-family: 'Courier New', monospace;
}

.pf-conn-count {
  color: var(--text-primary);
  font-weight: 600;
}

.pf-conn-count.pf-conn-active {
  color: var(--success-light);
}

.pf-conn-total {
  color: var(--text-disabled);
  font-size: 0.5625rem;
}

.pf-traffic {
  font-size: 0.625rem;
}

.pf-traffic-up {
  color: var(--warning-light);
}

.pf-traffic-down {
  color: var(--primary-light);
}

.pf-traffic-sep {
  color: var(--text-disabled);
  margin: 0 0.125rem;
}

/* 类型标签 */
.pf-type-tag {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.625rem;
  font-weight: 600;
}

.pf-type-local {
  background: var(--primary-bg);
  color: var(--primary-light);
}

.pf-type-remote {
  background: var(--accent-purple-bg);
  color: var(--accent-purple);
}

/* 状态 */
.pf-status {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.625rem;
  font-weight: 600;
}

.pf-status-running {
  background: var(--success-bg);
  color: var(--success-light);
}

.pf-status-stopped {
  background: var(--surface-2);
  color: var(--text-secondary);
}

.pf-status-error {
  background: var(--danger-bg);
  color: var(--accent-danger);
}

/* 操作按钮 */
.pf-actions {
  display: flex;
  gap: 0.25rem;
}

/* 按钮 */
.pf-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background: var(--border-subtle);
  border: 1px solid var(--border-strong);
  border-radius: 0.375rem;
  color: var(--text-secondary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
}

.pf-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.pf-btn-sm {
  padding: 0.25rem 0.5rem;
}

.pf-btn-primary {
  background: var(--primary-bg);
  border-color: var(--border-accent);
  color: var(--primary-light);
}

.pf-btn-primary:hover {
  background: var(--primary-bg-hover);
}

.pf-btn-success {
  background: var(--success-bg);
  border-color: var(--border-success);
  color: var(--success-light);
}

.pf-btn-success:hover {
  background: var(--success-bg);
}

.pf-btn-warn {
  background: var(--warning-bg);
  border-color: var(--border-warning);
  color: var(--warning-light);
}

.pf-btn-warn:hover {
  background: var(--warning-bg);
}

.pf-btn-danger {
  background: var(--danger-bg);
  border-color: var(--border-danger);
  color: var(--accent-danger);
}

.pf-btn-danger:hover {
  background: var(--danger-bg);
}

.pf-btn-text {
  display: inline;
}

/* 底部 */
.pf-footer {
  display: flex;
  justify-content: space-between;
  padding: 0.375rem 1rem;
  border-top: 1px solid var(--surface-hover);
  background: var(--toolbar-3);
  color: var(--text-muted);
  font-size: 0.6875rem;
  flex-shrink: 0;
}

/* 对话框 */
.pf-modal-mask {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  backdrop-filter: blur(4px);
}

.pf-modal {
  background: var(--bg-panel-solid);
  border: 1px solid var(--surface-hover);
  border-radius: 0.75rem;
  width: 400px;
  max-width: 90vw;
  box-shadow: var(--shadow-lg);
}

.pf-modal-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border-subtle);
}

.pf-modal-head h3 {
  margin: 0;
  color: var(--text-primary);
  font-size: 0.9375rem;
}

.pf-modal-head button {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.25rem;
  cursor: pointer;
}

.pf-modal-head button:hover {
  color: var(--text-primary);
}

.pf-modal-body {
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
}

.pf-field label {
  display: block;
  color: var(--text-secondary);
  font-size: 0.75rem;
  margin-bottom: 0.375rem;
}

.pf-input-row {
  display: flex;
  gap: 0.5rem;
}

.pf-input {
  background: var(--toolbar-4);
  border: 1px solid var(--surface-hover);
  border-radius: 0.375rem;
  color: var(--text-primary);
  font-size: 0.8125rem;
  padding: 0.5rem 0.75rem;
  outline: none;
  transition: border-color 0.15s;
}

.pf-input:focus {
  border-color: var(--border-accent);
}

.pf-input-addr {
  flex: 1;
}

.pf-input-port {
  width: 80px;
}

.pf-input-full {
  width: 100%;
}

.pf-hint-inline {
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--text-muted);
}

.pf-name-cell {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-primary);
  font-size: 0.8125rem;
}

.pf-radio-group {
  display: flex;
  gap: 0.5rem;
}

.pf-radio {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: var(--surface-1);
  border: 1px solid var(--surface-hover);
  border-radius: 0.375rem;
  color: var(--text-secondary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
  flex: 1;
}

.pf-radio:hover {
  background: var(--border-subtle);
}

.pf-radio.active {
  background: var(--primary-bg);
  border-color: var(--border-accent);
  color: var(--text-primary);
}

.pf-radio input {
  display: none;
}

.pf-radio-content {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.pf-radio-title {
  font-weight: 600;
  font-size: 0.8125rem;
}

.pf-radio-desc {
  font-size: 0.625rem;
  opacity: 0.7;
}

.pf-hint {
  color: var(--text-disabled);
  font-size: 0.625rem;
  margin-bottom: 0.375rem;
}

.pf-error {
  color: var(--accent-danger);
  font-size: 0.75rem;
}

.pf-modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.875rem 1.25rem;
  border-top: 1px solid var(--border-subtle);
}

/* 滚动条 */
.pf-list::-webkit-scrollbar {
  width: 4px;
}

.pf-list::-webkit-scrollbar-track {
  background: transparent;
}

.pf-list::-webkit-scrollbar-thumb {
  background: var(--border-default);
  border-radius: 999px;
}
</style>

<style>
/* 响应式 */
@media (max-width: 600px) {
  .pf-toolbar {
    padding: 0.375rem 0.5rem !important;
  }

  .pf-title {
    font-size: 0.75rem !important;
  }

  .pf-btn-text {
    display: none !important;
  }

  .pf-table {
    font-size: 0.625rem !important;
  }

  .pf-table th,
  .pf-table td {
    padding: 0.375rem 0.5rem !important;
  }

  .pf-mono {
    font-size: 0.5625rem !important;
  }

  .pf-footer {
    padding: 0.25rem 0.5rem !important;
    font-size: 0.5625rem !important;
  }

  .pf-modal {
    width: 95vw !important;
  }
}

@media (max-height: 500px) {
  .pf-toolbar {
    padding: 0.25rem 0.5rem !important;
  }

  .pf-list {
    padding: 0.25rem !important;
  }

  .pf-table th,
  .pf-table td {
    padding: 0.25rem 0.5rem !important;
  }

  .pf-footer {
    padding: 0.125rem 0.5rem !important;
  }

  .pf-modal-body {
    padding: 0.75rem 1rem !important;
  }
}
</style>
