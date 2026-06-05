<template>
  <div class="pf-panel">
    <!-- 顶部工具栏 -->
    <div class="pf-toolbar">
      <div class="pf-toolbar-left">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#63b3ed" stroke-width="2" stroke-linecap="round">
          <circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/>
          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
          <line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/>
        </svg>
        <span class="pf-title">端口转发</span>
      </div>
      <div class="pf-toolbar-right">
        <button class="pf-btn pf-btn-primary" @click="closeDialog(); showAdd = true">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span class="pf-btn-text">添加</span>
        </button>
      </div>
    </div>

    <!-- 转发列表 -->
    <div class="pf-list">
      <div v-if="forwards.length === 0" class="pf-empty">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#4a5568" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
        <p>暂无端口转发规则</p>
        <button class="pf-btn pf-btn-primary pf-btn-sm" @click="closeDialog(); showAdd = true">添加转发</button>
      </div>

      <div v-else class="pf-table-wrap">
        <table class="pf-table">
          <thead>
            <tr>
              <th>类型</th>
              <th>监听地址</th>
              <th>目标地址</th>
              <th>连接</th>
              <th>流量</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="fwd in forwards" :key="fwd.id" :class="'pf-row-' + fwd.status">
              <td>
                <span class="pf-type-tag" :class="'pf-type-' + fwd.type">
                  {{ fwd.type === 'local' ? '远程→本地' : '本地→远程' }}
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
                  <button v-if="fwd.status !== 'running'" class="pf-btn pf-btn-sm pf-btn-success" @click="startForward(fwd.id)" title="启动">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                  </button>
                  <button v-else class="pf-btn pf-btn-sm pf-btn-warn" @click="stopForward(fwd.id)" title="停止">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
                  </button>
                  <button v-if="fwd.status !== 'running'" class="pf-btn pf-btn-sm" @click="editForward(fwd)" title="编辑">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </button>
                  <button class="pf-btn pf-btn-sm pf-btn-danger" @click="removeForward(fwd.id)" title="删除">
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
      <span>{{ forwards.length }} 条规则</span>
      <span>{{ runningCount }} 运行中</span>
    </div>

    <!-- 添加对话框 -->
    <Teleport to="body">
      <div v-if="showAdd" class="pf-modal-mask" @click.self="showAdd = false">
        <div class="pf-modal">
          <div class="pf-modal-head">
            <h3>{{ editingId ? '编辑端口转发' : '添加端口转发' }}</h3>
            <button @click="closeDialog">&times;</button>
          </div>
          <div class="pf-modal-body">
            <div class="pf-field">
              <label>转发类型</label>
              <div class="pf-radio-group">
                <label class="pf-radio" :class="{ active: form.type === 'local' }">
                  <input type="radio" v-model="form.type" value="local" />
                  <div class="pf-radio-content">
                    <span class="pf-radio-title">远程 → 本地</span>
                    <span class="pf-radio-desc">访问远程内网服务</span>
                  </div>
                </label>
                <label class="pf-radio" :class="{ active: form.type === 'remote' }">
                  <input type="radio" v-model="form.type" value="remote" />
                  <div class="pf-radio-content">
                    <span class="pf-radio-title">本地 → 远程</span>
                    <span class="pf-radio-desc">内网穿透</span>
                  </div>
                </label>
              </div>
            </div>
            <div class="pf-field">
              <label>{{ form.type === 'local' ? '本地监听端口' : '远程开放端口' }}</label>
              <div class="pf-hint">{{ form.type === 'local' ? '你的电脑上通过哪个端口访问' : '远程服务器上开放哪个端口' }}</div>
              <div class="pf-input-row">
                <input v-model="form.bindAddr" :placeholder="form.type === 'local' ? '127.0.0.1' : '0.0.0.0'" class="pf-input pf-input-addr" />
                <input v-model.number="form.bindPort" type="number" placeholder="端口" class="pf-input pf-input-port" />
              </div>
            </div>
            <div class="pf-field">
              <label>{{ form.type === 'local' ? '远程目标地址' : '本地服务地址' }}</label>
              <div class="pf-hint">{{ form.type === 'local' ? 'SSH 服务器能访问到的内网地址' : '你电脑上运行的服务地址' }}</div>
              <div class="pf-input-row">
                <input v-model="form.remoteHost" placeholder="127.0.0.1" class="pf-input pf-input-addr" />
                <input v-model.number="form.remotePort" type="number" placeholder="端口" class="pf-input pf-input-port" />
              </div>
            </div>
            <div v-if="formError" class="pf-error">{{ formError }}</div>
          </div>
          <div class="pf-modal-foot">
            <button class="pf-btn" @click="closeDialog">取消</button>
            <button class="pf-btn pf-btn-primary" @click="addForward" :disabled="submitting">
              {{ submitting ? (editingId ? '保存中...' : '添加中...') : (editingId ? '保存' : '添加') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, inject } from 'vue'
import { Events } from '@wailsio/runtime'
import * as PortForwardService from '../../../../bindings/changeme/ssh/portforwardservice.js'
import { showMessage } from '../../../utils/message'
import { addLog, LogType, LogLevel } from '../../../utils/logger'

const connId = inject('connId')

const forwards = ref([])
const showAdd = ref(false)
const submitting = ref(false)
const formError = ref('')
const editingId = ref(null)

const form = ref({
  type: 'local',
  bindAddr: '127.0.0.1',
  bindPort: null,
  remoteHost: '127.0.0.1',
  remotePort: null
})

const runningCount = computed(() => forwards.value.filter(f => f.status === 'running').length)

// 切换类型时自动设置默认绑定地址
watch(() => form.value.type, (type) => {
  form.value.bindAddr = type === 'local' ? '127.0.0.1' : '0.0.0.0'
})

const statusLabel = (s) => ({ running: '运行中', stopped: '已停止', error: '错误' }[s] || s)

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
    formError.value = '请填写端口号'
    return
  }
  submitting.value = true
  try {
    // 编辑模式：先删除旧规则
    if (editingId.value) {
      try {
        await PortForwardService.RemoveForward(editingId.value)
      } catch (e) {}
    }

    const fn = form.value.type === 'local' ? 'AddLocalForward' : 'AddRemoteForward'
    await PortForwardService[fn](connId, form.value.bindAddr, form.value.bindPort, form.value.remoteHost, form.value.remotePort)

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
    showMessage('转发已启动', 'success')
    await loadForwards()
  } catch (e) {
    addLog(connId, 'portForward', LogLevel.ERROR, '启动端口转发失败: ' + (e?.message || e))
    showMessage('启动失败: ' + (e?.message || e), 'error')
  }
}

const stopForward = async (id) => {
  try {
    await PortForwardService.StopForward(id)
    addLog(connId, 'portForward', LogLevel.INFO, '停止端口转发: ' + id)
    showMessage('转发已停止', 'success')
    await loadForwards()
  } catch (e) {
    addLog(connId, 'portForward', LogLevel.ERROR, '停止端口转发失败: ' + (e?.message || e))
    showMessage('停止失败: ' + (e?.message || e), 'error')
  }
}

const removeForward = async (id) => {
  try {
    await PortForwardService.RemoveForward(id)
    addLog(connId, 'portForward', LogLevel.INFO, '删除端口转发: ' + id)
    showMessage('转发已删除', 'success')
    await loadForwards()
  } catch (e) {
    addLog(connId, 'portForward', LogLevel.ERROR, '删除端口转发失败: ' + (e?.message || e))
    showMessage('删除失败: ' + (e?.message || e), 'error')
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
  background: rgba(30, 30, 30, 0.95);
  overflow: hidden;
}

/* 工具栏 */
.pf-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(40, 40, 40, 0.5);
  flex-shrink: 0;
}

.pf-toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pf-title {
  color: #e2e8f0;
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
  color: #718096;
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
  color: #a0aec0;
  font-weight: 600;
  border-bottom: 2px solid rgba(255, 255, 255, 0.1);
  white-space: nowrap;
}

.pf-table td {
  padding: 0.5rem 0.75rem;
  color: #e2e8f0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  white-space: nowrap;
}

.pf-mono {
  font-family: 'Courier New', monospace;
  font-size: 0.6875rem;
}

.pf-row-running td {
  background: rgba(72, 187, 120, 0.03);
}

.pf-row-error td {
  background: rgba(245, 101, 101, 0.03);
}

/* 连接和流量统计 */
.pf-stat {
  font-size: 0.6875rem;
  font-family: 'Courier New', monospace;
}

.pf-conn-count {
  color: #e2e8f0;
  font-weight: 600;
}

.pf-conn-count.pf-conn-active {
  color: #68d391;
}

.pf-conn-total {
  color: #4a5568;
  font-size: 0.5625rem;
}

.pf-traffic {
  font-size: 0.625rem;
}

.pf-traffic-up {
  color: #f6ad55;
}

.pf-traffic-down {
  color: #63b3ed;
}

.pf-traffic-sep {
  color: #4a5568;
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
  background: rgba(66, 153, 225, 0.2);
  color: #63b3ed;
}

.pf-type-remote {
  background: rgba(159, 122, 234, 0.2);
  color: #b794f4;
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
  background: rgba(72, 187, 120, 0.2);
  color: #68d391;
}

.pf-status-stopped {
  background: rgba(160, 174, 192, 0.2);
  color: #a0aec0;
}

.pf-status-error {
  background: rgba(245, 101, 101, 0.2);
  color: #fc8181;
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
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 0.375rem;
  color: #a0aec0;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
}

.pf-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

.pf-btn-sm {
  padding: 0.25rem 0.5rem;
}

.pf-btn-primary {
  background: rgba(66, 153, 225, 0.2);
  border-color: rgba(66, 153, 225, 0.4);
  color: #63b3ed;
}

.pf-btn-primary:hover {
  background: rgba(66, 153, 225, 0.35);
}

.pf-btn-success {
  background: rgba(72, 187, 120, 0.2);
  border-color: rgba(72, 187, 120, 0.4);
  color: #68d391;
}

.pf-btn-success:hover {
  background: rgba(72, 187, 120, 0.35);
}

.pf-btn-warn {
  background: rgba(237, 137, 54, 0.2);
  border-color: rgba(237, 137, 54, 0.4);
  color: #f6ad55;
}

.pf-btn-warn:hover {
  background: rgba(237, 137, 54, 0.35);
}

.pf-btn-danger {
  background: rgba(245, 101, 101, 0.2);
  border-color: rgba(245, 101, 101, 0.4);
  color: #fc8181;
}

.pf-btn-danger:hover {
  background: rgba(245, 101, 101, 0.35);
}

.pf-btn-text {
  display: inline;
}

/* 底部 */
.pf-footer {
  display: flex;
  justify-content: space-between;
  padding: 0.375rem 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(40, 40, 40, 0.5);
  color: #718096;
  font-size: 0.6875rem;
  flex-shrink: 0;
}

/* 对话框 */
.pf-modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  backdrop-filter: blur(4px);
}

.pf-modal {
  background: #1a1a1a;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.75rem;
  width: 400px;
  max-width: 90vw;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
}

.pf-modal-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.pf-modal-head h3 {
  margin: 0;
  color: #e2e8f0;
  font-size: 0.9375rem;
}

.pf-modal-head button {
  background: none;
  border: none;
  color: #718096;
  font-size: 1.25rem;
  cursor: pointer;
}

.pf-modal-head button:hover {
  color: #e2e8f0;
}

.pf-modal-body {
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
}

.pf-field label {
  display: block;
  color: #a0aec0;
  font-size: 0.75rem;
  margin-bottom: 0.375rem;
}

.pf-input-row {
  display: flex;
  gap: 0.5rem;
}

.pf-input {
  background: rgba(30, 30, 30, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.375rem;
  color: #e2e8f0;
  font-size: 0.8125rem;
  padding: 0.5rem 0.75rem;
  outline: none;
  transition: border-color 0.15s;
}

.pf-input:focus {
  border-color: rgba(66, 153, 225, 0.5);
}

.pf-input-addr {
  flex: 1;
}

.pf-input-port {
  width: 80px;
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
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.375rem;
  color: #a0aec0;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
  flex: 1;
}

.pf-radio:hover {
  background: rgba(255, 255, 255, 0.06);
}

.pf-radio.active {
  background: rgba(66, 153, 225, 0.15);
  border-color: rgba(66, 153, 225, 0.4);
  color: #e2e8f0;
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
  color: #4a5568;
  font-size: 0.625rem;
  margin-bottom: 0.375rem;
}

.pf-error {
  color: #fc8181;
  font-size: 0.75rem;
}

.pf-modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.875rem 1.25rem;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

/* 滚动条 */
.pf-list::-webkit-scrollbar {
  width: 4px;
}

.pf-list::-webkit-scrollbar-track {
  background: transparent;
}

.pf-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
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
