<template>
  <div class="fw-panel">
    <!-- 工具栏 -->
    <div class="fw-toolbar">
      <div class="fw-toolbar-left">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        <span class="fw-title">防火墙</span>
        <span v-if="fwInfo.type !== 'unknown'" class="fw-badge" :class="'fw-badge-' + fwInfo.status">
          {{ fwTypeName }} · {{ fwInfo.status === 'active' ? '运行中' : '已关闭' }}
        </span>
      </div>
      <div class="fw-toolbar-right">
        <button v-if="fwInfo.type !== 'unknown'" class="fw-btn" :class="fwInfo.status === 'active' ? 'fw-btn-warn' : 'fw-btn-success'" @click="toggleFirewall">
          {{ fwInfo.status === 'active' ? '关闭防火墙' : '开启防火墙' }}
        </button>
        <button class="fw-btn" @click="showAdd = true" :disabled="fwInfo.type === 'unknown'">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span class="fw-btn-text">添加规则</span>
        </button>
        <button class="fw-btn" @click="loadInfo" :disabled="loading">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          <span class="fw-btn-text">刷新</span>
        </button>
      </div>
    </div>

    <!-- 内容区 -->
    <div class="fw-content">
      <!-- 加载中 -->
      <div v-if="loading && !fwInfo.type" class="fw-loading">
        <div class="fw-spinner"></div>
        <p>正在检测防火墙...</p>
      </div>

      <!-- 未知防火墙 -->
      <div v-else-if="fwInfo.type === 'unknown'" class="fw-empty">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        <p>未检测到支持的防火墙</p>
        <p class="fw-hint">支持 iptables / firewalld / ufw</p>
      </div>

      <!-- 规则表格 -->
      <div v-else class="fw-body">
        <div v-if="fwInfo.rules.length === 0" class="fw-empty-rules">
          <p>暂无规则</p>
        </div>
        <div v-else class="fw-table-wrap">
          <table class="fw-table">
            <thead>
              <tr>
                <th>#</th>
                <th>{{ fwInfo.type === 'firewalld' ? '区域' : '方向' }}</th>
                <th>动作</th>
                <th>协议</th>
                <th>端口/服务</th>
                <th>来源</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="rule in fwInfo.rules" :key="rule.index + rule.chain">
                <td class="fw-td-num">{{ rule.index }}</td>
                <td><span class="fw-chain-tag">{{ chainLabel(rule.chain) }}</span></td>
                <td>
                  <span class="fw-target-tag" :class="'fw-target-' + rule.target.toLowerCase()">
                    {{ targetLabel(rule.target) }}
                  </span>
                </td>
                <td>{{ protocolLabel(rule.protocol) }}</td>
                <td class="fw-mono">{{ rule.port || '-' }}</td>
                <td class="fw-mono">{{ rule.source || '-' }}</td>
                <td>
                  <button class="fw-btn fw-btn-sm fw-btn-danger" @click="deleteRule(rule)" title="删除此规则">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 原始输出 -->
        <details class="fw-raw-section">
          <summary class="fw-raw-toggle">查看原始输出</summary>
          <pre class="fw-raw">{{ fwInfo.rawOutput || '无' }}</pre>
        </details>
      </div>
    </div>

    <!-- 自定义命令 -->
    <div class="fw-cmd-bar">
      <input v-model="customCmd" @keyup.enter="runCustom" class="fw-cmd-input" placeholder="输入自定义防火墙命令..." :disabled="fwInfo.type === 'unknown'" />
      <button class="fw-btn fw-btn-sm" @click="runCustom" :disabled="!customCmd.trim() || running">
        {{ running ? '执行中...' : '执行' }}
      </button>
    </div>

    <!-- 底部 -->
    <div class="fw-footer">
      <span v-if="cmdResult" class="fw-cmd-result" :class="cmdError ? 'fw-cmd-error' : ''">{{ cmdResult }}</span>
      <span v-else class="fw-footer-info">{{ fwInfo.type !== 'unknown' ? fwInfo.rules.length + ' 条规则' : '' }}</span>
    </div>

    <!-- 添加规则弹窗 -->
    <Teleport to="body">
      <div v-if="showAdd" class="fw-modal-mask" @click.self="showAdd = false">
        <div class="fw-modal">
          <div class="fw-modal-head">
            <h3>添加防火墙规则</h3>
            <button @click="showAdd = false">&times;</button>
          </div>
          <div class="fw-modal-body">
            <!-- iptables 字段 -->
            <template v-if="fwInfo.type === 'iptables'">
              <div class="fw-field">
                <label>方向</label>
                <select v-model="addForm.chain" class="fw-input">
                  <option value="INPUT">入站 (INPUT)</option>
                  <option value="OUTPUT">出站 (OUTPUT)</option>
                  <option value="FORWARD">转发 (FORWARD)</option>
                </select>
              </div>
              <div class="fw-field">
                <label>动作</label>
                <select v-model="addForm.target" class="fw-input">
                  <option value="ACCEPT">允许</option>
                  <option value="DROP">拒绝（静默丢弃）</option>
                  <option value="REJECT">拒绝（返回错误）</option>
                </select>
              </div>
            </template>

            <!-- ufw 字段 -->
            <template v-if="fwInfo.type === 'ufw'">
              <div class="fw-field">
                <label>动作</label>
                <select v-model="addForm.target" class="fw-input">
                  <option value="allow">允许</option>
                  <option value="deny">拒绝</option>
                </select>
              </div>
            </template>

            <!-- firewalld 字段 -->
            <template v-if="fwInfo.type === 'firewalld'">
              <div class="fw-field">
                <label>区域</label>
                <select v-model="addForm.chain" class="fw-input">
                  <option v-for="z in fwInfo.chains" :key="z" :value="z">{{ zoneLabel(z) }}</option>
                </select>
              </div>
            </template>

            <!-- 通用字段 -->
            <div class="fw-field">
              <label>协议</label>
              <select v-model="addForm.protocol" class="fw-input">
                <option value="tcp">TCP</option>
                <option value="udp">UDP</option>
                <option value="tcp+udp">TCP + UDP</option>
              </select>
            </div>
            <div class="fw-field">
              <label>端口</label>
              <input v-model="addForm.port" class="fw-input" placeholder="如 80、443、8000-9000" />
            </div>
            <div class="fw-field">
              <label>来源地址</label>
              <input v-model="addForm.source" class="fw-input" placeholder="留空表示允许所有来源" />
            </div>
            <div v-if="fwInfo.type === 'iptables'" class="fw-field">
              <label>备注</label>
              <input v-model="addForm.comment" class="fw-input" placeholder="可选，便于识别规则用途" />
            </div>
            <div v-if="addError" class="fw-error">{{ addError }}</div>
          </div>
          <div class="fw-modal-foot">
            <button class="fw-btn" @click="showAdd = false">取消</button>
            <button class="fw-btn fw-btn-primary" @click="addRule" :disabled="addSubmitting">
              {{ addSubmitting ? '添加中...' : '添加' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, inject } from 'vue'
import * as FirewallService from '../../../../bindings/changeme/ssh/firewallservice.js'
import { showMessage } from '../../../utils/message'
import { useConfirm } from '../../../utils/confirm'
import { addLog, LogType, LogLevel } from '../../../utils/logger'

const { confirm } = useConfirm()

const connId = inject('connId')

const loading = ref(false)
const fwInfo = ref({ type: 'unknown', status: 'unknown', rules: [], rawOutput: '', chains: [] })
const customCmd = ref('')
const cmdResult = ref('')
const cmdError = ref(false)
const running = ref(false)

const showAdd = ref(false)
const addSubmitting = ref(false)
const addError = ref('')
const addForm = reactive({
  chain: 'INPUT',
  target: 'ACCEPT',
  protocol: 'tcp',
  port: '',
  source: '',
  comment: ''
})

// 中文映射
const fwTypeName = computed(() => {
  const map = { iptables: 'iptables', firewalld: 'Firewalld', ufw: 'UFW' }
  return map[fwInfo.value.type] || fwInfo.value.type
})

const targetLabel = (t) => {
  const map = { ACCEPT: '允许', allow: '允许', DROP: '拒绝', deny: '拒绝', REJECT: '拒绝(响应)' }
  return map[t] || t
}

const chainLabel = (c) => {
  const map = { INPUT: '入站', OUTPUT: '出站', FORWARD: '转发', IN: '入站', OUT: '出站', public: '公共', trusted: '信任', drop: '丢弃', block: '阻塞', dmz: '隔离' }
  return map[c] || c
}

const protocolLabel = (p) => {
  const map = { tcp: 'TCP', udp: 'UDP', icmp: 'ICMP', all: '全部', service: '服务' }
  return map[p] || (p || '-')
}

const zoneLabel = (z) => {
  const map = { public: '公共', trusted: '信任', drop: '丢弃', block: '阻塞', dmz: '隔离', home: '家庭', work: '工作', internal: '内部', external: '外部' }
  return map[z] ? `${map[z]} (${z})` : z
}

const loadInfo = async () => {
  if (!connId) return
  loading.value = true
  try {
    const info = await FirewallService.GetFirewallInfo(connId)
    fwInfo.value = info || { type: 'unknown', status: 'unknown', rules: [], rawOutput: '', chains: [] }
    if (fwInfo.value.chains?.length) {
      addForm.chain = fwInfo.value.chains[0]
    }
  } catch (e) {
    console.error('[FirewallPanel] 加载失败:', e)
    fwInfo.value = { type: 'unknown', status: 'error', rules: [], rawOutput: String(e?.message || e), chains: [] }
  } finally {
    loading.value = false
  }
}

const addRule = async () => {
  if (!addForm.port.trim()) {
    addError.value = '请填写端口'
    return
  }
  addSubmitting.value = true
  addError.value = ''
  try {
    await FirewallService.AddRule(connId, addForm.chain, addForm.target, addForm.protocol, addForm.port, addForm.source, addForm.comment)
    addLog(connId, 'firewall', LogLevel.SUCCESS,
      `添加防火墙规则: ${addForm.chain} ${addForm.target} ${addForm.protocol} ${addForm.port}`)
    showMessage('规则已添加', 'success')
    showAdd.value = false
    addForm.port = ''
    addForm.source = ''
    addForm.comment = ''
    await loadInfo()
  } catch (e) {
    addError.value = String(e?.message || e)
    addLog(connId, 'firewall', LogLevel.ERROR, '添加防火墙规则失败: ' + (e?.message || e))
  } finally {
    addSubmitting.value = false
  }
}

const deleteRule = async (rule) => {
  const ok = await confirm({
    title: '删除规则',
    message: `确定删除第 ${rule.index} 条规则？\n${chainLabel(rule.chain)} · ${targetLabel(rule.target)} · ${rule.port || '-'}`,
    danger: true,
  })
  if (!ok) return
  try {
    await FirewallService.DeleteRule(connId, rule.chain, rule.index, rule.port, rule.protocol)
    addLog(connId, 'firewall', LogLevel.INFO, `删除防火墙规则: #${rule.index} ${chainLabel(rule.chain)} ${rule.port || ''}`)
    showMessage('规则已删除', 'success')
    await loadInfo()
  } catch (e) {
    addLog(connId, 'firewall', LogLevel.ERROR, '删除防火墙规则失败: ' + (e?.message || e))
    showMessage('删除失败: ' + (e?.message || e), 'error')
  }
}

const toggleFirewall = async () => {
  const enable = fwInfo.value.status !== 'active'
  const action = enable ? '开启' : '关闭'
  const ok = await confirm({
    title: `${action}防火墙`,
    message: enable ? '开启后将按规则过滤网络流量' : '关闭后所有网络流量将不受限制',
    danger: !enable,
  })
  if (!ok) return
  try {
    await FirewallService.ToggleFirewall(connId, enable)
    addLog(connId, 'firewall', enable ? LogLevel.SUCCESS : LogLevel.WARNING, `防火墙已${action}`)
    showMessage(`防火墙已${action}`, 'success')
    await loadInfo()
  } catch (e) {
    addLog(connId, 'firewall', LogLevel.ERROR, `防火墙${action}失败: ` + (e?.message || e))
    showMessage(`${action}失败: ` + (e?.message || e), 'error')
  }
}

const runCustom = async () => {
  if (!customCmd.value.trim()) return
  running.value = true
  cmdResult.value = ''
  cmdError.value = false
  try {
    const result = await FirewallService.RunCustomCommand(connId, customCmd.value)
    cmdResult.value = result || '执行成功（无输出）'
    addLog(connId, 'firewall', LogLevel.INFO, `执行防火墙命令: ${customCmd.value}`, { output: result })
    customCmd.value = ''
    showMessage('命令执行成功', 'success')
    await loadInfo()
  } catch (e) {
    cmdResult.value = String(e?.message || e)
    cmdError.value = true
    addLog(connId, 'firewall', LogLevel.ERROR, `防火墙命令失败: ${customCmd.value}`, { error: String(e?.message || e) })
    showMessage('命令执行失败', 'error')
  } finally {
    running.value = false
  }
}

onMounted(() => loadInfo())
</script>

<style scoped>
.fw-panel {
  width: 100%; height: 100%;
  display: flex; flex-direction: column;
  background: var(--bg-panel);
  overflow: hidden;
}

.fw-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-default);
  background: var(--surface-2);
  flex-shrink: 0;
}
.fw-toolbar-left { display: flex; align-items: center; gap: 0.5rem; color: var(--warning-light); }
.fw-toolbar-right { display: flex; gap: 0.375rem; }
.fw-title { color: var(--text-primary); font-weight: 600; font-size: 0.875rem; }

.fw-badge {
  padding: 0.125rem 0.5rem; border-radius: 0.25rem;
  font-size: 0.625rem; font-weight: 600;
}
.fw-badge-active { background: var(--success-bg); color: var(--success-light); }
.fw-badge-inactive { background: var(--surface-2); color: var(--text-secondary); }

.fw-btn {
  display: inline-flex; align-items: center; gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 0.375rem;
  color: var(--text-secondary); font-size: 0.75rem;
  cursor: pointer; transition: all 0.15s;
}
.fw-btn:hover:not(:disabled) { background: var(--surface-hover); color: var(--text-primary); }
.fw-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.fw-btn-sm { padding: 0.25rem 0.5rem; }
.fw-btn-primary { background: var(--primary-bg); border-color: var(--border-accent); color: var(--primary-light); }
.fw-btn-primary:hover:not(:disabled) { background: var(--primary-bg-hover); }
.fw-btn-success { background: var(--success-bg); border-color: color-mix(in srgb, var(--accent-success), transparent 60%); color: var(--success-light); }
.fw-btn-warn { background: var(--warning-bg); border-color: color-mix(in srgb, var(--warning-light), transparent 60%); color: var(--warning-light); }
.fw-btn-danger { background: var(--danger-bg); border-color: color-mix(in srgb, var(--accent-danger), transparent 60%); color: var(--accent-danger); }

.fw-content { flex: 1; overflow-y: auto; min-height: 0; }

.fw-loading, .fw-empty, .fw-empty-rules {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 100%; gap: 0.75rem; color: var(--text-muted);
}
.fw-loading p, .fw-empty p, .fw-empty-rules p { margin: 0; font-size: 0.875rem; }
.fw-hint { font-size: 0.75rem !important; color: var(--text-muted) !important; }
.fw-empty-rules { height: auto; padding: 2rem; }

.fw-spinner {
  width: 24px; height: 24px;
  border: 2px solid color-mix(in srgb, var(--warning-light), transparent 80%); border-top-color: var(--warning-light);
  border-radius: 50%; animation: fw-spin 0.8s linear infinite;
}
@keyframes fw-spin { to { transform: rotate(360deg); } }

.fw-body { padding: 0.75rem; }

.fw-table-wrap { overflow-x: auto; margin-bottom: 0.75rem; }
.fw-table { width: 100%; border-collapse: collapse; font-size: 0.75rem; }
.fw-table th {
  padding: 0.5rem 0.625rem; text-align: left;
  color: var(--text-secondary); font-weight: 600; white-space: nowrap;
  border-bottom: 2px solid var(--border-default);
  background: var(--surface-2);
  position: sticky; top: 0; z-index: 1;
}
.fw-table td {
  padding: 0.375rem 0.625rem;
  color: var(--text-primary); border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}
.fw-table tr:hover td { background: var(--bg-hover); }
.fw-td-num { color: var(--text-muted); font-family: monospace; }
.fw-mono { font-family: 'Courier New', monospace; font-size: 0.6875rem; }

.fw-chain-tag {
  display: inline-block; padding: 0.0625rem 0.375rem;
  background: color-mix(in srgb, var(--primary-light), transparent 90%); border-radius: 0.25rem;
  color: var(--primary-light); font-size: 0.625rem; font-weight: 600;
}
.fw-target-tag {
  display: inline-block; padding: 0.0625rem 0.375rem;
  border-radius: 0.25rem; font-size: 0.625rem; font-weight: 600;
}
.fw-target-accept, .fw-target-allow { background: var(--success-bg); color: var(--success-light); }
.fw-target-drop, .fw-target-deny { background: var(--danger-bg); color: var(--accent-danger); }
.fw-target-reject { background: var(--warning-bg); color: var(--warning-light); }

.fw-raw-section {
  border: 1px solid var(--border-subtle);
  border-radius: 0.375rem; overflow: hidden;
}
.fw-raw-toggle {
  padding: 0.5rem 0.75rem; color: var(--text-muted); font-size: 0.75rem;
  cursor: pointer; background: var(--surface-1);
}
.fw-raw-toggle:hover { background: var(--surface-2); }
.fw-raw {
  margin: 0; padding: 0.75rem;
  color: var(--text-secondary); font-size: 0.6875rem;
  font-family: 'Courier New', monospace; line-height: 1.5;
  white-space: pre; overflow-x: auto;
  max-height: 30vh; background: var(--surface-1);
}

.fw-cmd-bar {
  display: flex; gap: 0.5rem; padding: 0.5rem 1rem;
  border-top: 1px solid var(--border-subtle);
  background: var(--surface-2); flex-shrink: 0;
}
.fw-cmd-input {
  flex: 1; background: var(--bg-input);
  border: 1px solid var(--border-default); border-radius: 0.375rem;
  color: var(--text-primary); font-size: 0.75rem; font-family: 'Courier New', monospace;
  padding: 0.375rem 0.625rem; outline: none;
}
.fw-cmd-input:focus { border-color: var(--border-accent); }

.fw-footer {
  display: flex; align-items: center;
  padding: 0.25rem 1rem;
  border-top: 1px solid var(--border-default);
  background: var(--surface-2);
  font-size: 0.625rem; flex-shrink: 0;
}
.fw-footer-info { color: var(--text-muted); }
.fw-cmd-result { color: var(--success-light); font-family: monospace; white-space: pre-wrap; word-break: break-all; }
.fw-cmd-error { color: var(--accent-danger); }

.fw-modal-mask {
  position: fixed; inset: 0;
  background: var(--bg-overlay); display: flex;
  align-items: center; justify-content: center;
  z-index: 10000; backdrop-filter: blur(4px);
}
.fw-modal {
  background: var(--bg-panel-solid); border: 1px solid var(--border-default);
  border-radius: 0.75rem; width: 400px; max-width: 90vw;
  box-shadow: var(--shadow-lg);
}
.fw-modal-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--border-subtle);
}
.fw-modal-head h3 { margin: 0; color: var(--text-primary); font-size: 0.9375rem; }
.fw-modal-head button { background: none; border: none; color: var(--text-muted); font-size: 1.25rem; cursor: pointer; }
.fw-modal-head button:hover { color: var(--text-primary); }
.fw-modal-body { padding: 1rem 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; }
.fw-modal-foot {
  display: flex; justify-content: flex-end; gap: 0.5rem;
  padding: 0.875rem 1.25rem; border-top: 1px solid var(--border-subtle);
}
.fw-field label { display: block; color: var(--text-secondary); font-size: 0.75rem; margin-bottom: 0.25rem; }
.fw-input {
  width: 100%; background: var(--bg-input);
  border: 1px solid var(--border-default); border-radius: 0.375rem;
  color: var(--text-primary); font-size: 0.8125rem; padding: 0.5rem 0.75rem;
  outline: none; box-sizing: border-box;
}
.fw-input:focus { border-color: var(--border-accent); }
.fw-error { color: var(--accent-danger); font-size: 0.75rem; }
</style>
