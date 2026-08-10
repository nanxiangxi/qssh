<template>
  <div class="fw-panel">
    <!-- 工具栏 -->
    <div class="fw-toolbar">
      <div class="fw-toolbar-left">
        <svg class="icon-warning" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        <span class="fw-title">{{ t('firewall.title') }}</span>
        <span v-if="fwInfo.type !== 'unknown'" class="fw-badge" :class="'fw-badge-' + fwInfo.status">
          {{ fwTypeName }} · {{ fwInfo.status === 'active' ? t('firewall.running') : t('firewall.stopped') }}
        </span>
      </div>
      <div class="fw-toolbar-right">
        <button v-if="fwInfo.type !== 'unknown'" class="fw-btn" :class="fwInfo.status === 'active' ? 'fw-btn-warn' : 'fw-btn-success'" @click="toggleFirewall">
          {{ fwInfo.status === 'active' ? t('firewall.disableFirewall') : t('firewall.enableFirewall') }}
        </button>
        <button class="fw-btn" @click="showAdd = true" :disabled="fwInfo.type === 'unknown'">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span class="fw-btn-text">{{ t('firewall.addRule') }}</span>
        </button>
        <button class="fw-btn" @click="loadInfo" :disabled="loading">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          <span class="fw-btn-text">{{ t('common.refresh') }}</span>
        </button>
      </div>
    </div>

    <!-- 内容区 -->
    <div class="fw-content">
      <!-- 加载中 -->
      <div v-if="loading && !fwInfo.type" class="fw-loading">
        <div class="fw-spinner"></div>
        <p>{{ t('common.loading') }}...</p>
      </div>

      <!-- 未知防火墙 -->
      <div v-else-if="fwInfo.type === 'unknown'" class="fw-empty">
        <svg class="icon-disabled" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        <p>{{ t('firewall.notDetected') }}</p>
        <p class="fw-hint">{{ t('firewall.supportedTypes') }}</p>
      </div>

      <!-- 规则表格 -->
      <div v-else class="fw-body">
        <div v-if="fwInfo.rules.length === 0" class="fw-empty-rules">
          <p>{{ t('firewall.noRules') }}</p>
        </div>
        <div v-else class="fw-table-wrap">
          <table class="fw-table">
            <thead>
              <tr>
                <th>#</th>
                <th>{{ fwInfo.type === 'firewalld' ? t('firewall.zone') : t('firewall.direction') }}</th>
                <th>{{ t('firewall.action') }}</th>
                <th>{{ t('firewall.protocol') }}</th>
                <th>{{ t('firewall.portService') }}</th>
                <th>{{ t('firewall.source') }}</th>
                <th>{{ t('common.actions') }}</th>
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
                  <button class="fw-btn fw-btn-sm fw-btn-danger" @click="deleteRule(rule)" :title="t('firewall.deleteRule')">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 原始输出 -->
        <details class="fw-raw-section">
          <summary class="fw-raw-toggle">{{ t('firewall.viewRawOutput') }}</summary>
          <pre class="fw-raw">{{ fwInfo.rawOutput || t('firewall.none') }}</pre>
        </details>
      </div>
    </div>

    <!-- 自定义命令 -->
    <div class="fw-cmd-bar">
      <input v-model="customCmd" @keyup.enter="runCustom" class="fw-cmd-input" :placeholder="t('firewall.cmdPlaceholder')" :disabled="fwInfo.type === 'unknown'" />
      <button class="fw-btn fw-btn-sm" @click="runCustom" :disabled="!customCmd.trim() || running">
        {{ running ? t('firewall.executing') : t('firewall.execute') }}
      </button>
    </div>

    <!-- 底部 -->
    <div class="fw-footer">
      <span v-if="cmdResult" class="fw-cmd-result" :class="cmdError ? 'fw-cmd-error' : ''">{{ cmdResult }}</span>
      <span v-else class="fw-footer-info">{{ fwInfo.type !== 'unknown' ? t('firewall.rulesCount', { count: fwInfo.rules.length }) : '' }}</span>
    </div>

    <!-- 添加规则弹窗 -->
    <Teleport to="body">
      <div v-if="showAdd" class="fw-modal-mask" @click.self="showAdd = false">
        <div class="fw-modal">
          <div class="fw-modal-head">
            <h3>{{ t('firewall.addRule') }}</h3>
            <button @click="showAdd = false">&times;</button>
          </div>
          <div class="fw-modal-body">
            <!-- iptables 字段 -->
            <template v-if="fwInfo.type === 'iptables'">
              <div class="fw-field">
                <label>{{ t('firewall.direction') }}</label>
                <select v-model="addForm.chain" class="fw-input">
                  <option value="INPUT">{{ t('firewall.input') }} (INPUT)</option>
                  <option value="OUTPUT">{{ t('firewall.output') }} (OUTPUT)</option>
                  <option value="FORWARD">{{ t('firewall.forward') }} (FORWARD)</option>
                </select>
              </div>
              <div class="fw-field">
                <label>{{ t('firewall.action') }}</label>
                <select v-model="addForm.target" class="fw-input">
                  <option value="ACCEPT">{{ t('firewall.accept') }}</option>
                  <option value="DROP">{{ t('firewall.drop') }}</option>
                  <option value="REJECT">{{ t('firewall.reject') }}</option>
                </select>
              </div>
            </template>

            <!-- ufw 字段 -->
            <template v-if="fwInfo.type === 'ufw'">
              <div class="fw-field">
                <label>{{ t('firewall.action') }}</label>
                <select v-model="addForm.target" class="fw-input">
                  <option value="allow">{{ t('firewall.allow') }}</option>
                  <option value="deny">{{ t('firewall.deny') }}</option>
                </select>
              </div>
            </template>

            <!-- firewalld 字段 -->
            <template v-if="fwInfo.type === 'firewalld'">
              <div class="fw-field">
                <label>{{ t('firewall.zone') }}</label>
                <select v-model="addForm.chain" class="fw-input">
                  <option v-for="z in fwInfo.chains" :key="z" :value="z">{{ zoneLabel(z) }}</option>
                </select>
              </div>
            </template>

            <!-- 通用字段 -->
            <div class="fw-field">
              <label>{{ t('firewall.protocol') }}</label>
              <select v-model="addForm.protocol" class="fw-input">
                <option value="tcp">TCP</option>
                <option value="udp">UDP</option>
                <option value="tcp+udp">TCP + UDP</option>
              </select>
            </div>
            <div class="fw-field">
              <label>{{ t('firewall.port') }}</label>
              <input v-model="addForm.port" class="fw-input" :placeholder="t('firewall.portPlaceholder')" />
            </div>
            <div class="fw-field">
              <label>{{ t('firewall.sourceAddr') }}</label>
              <input v-model="addForm.source" class="fw-input" :placeholder="t('firewall.sourcePlaceholder')" />
            </div>
            <div v-if="fwInfo.type === 'iptables'" class="fw-field">
              <label>{{ t('firewall.comment') }}</label>
              <input v-model="addForm.comment" class="fw-input" :placeholder="t('firewall.commentPlaceholder')" />
            </div>
            <div v-if="addError" class="fw-error">{{ addError }}</div>
          </div>
          <div class="fw-modal-foot">
            <button class="fw-btn" @click="showAdd = false">{{ t('common.cancel') }}</button>
            <button class="fw-btn fw-btn-primary" @click="addRule" :disabled="addSubmitting">
              {{ addSubmitting ? t('firewall.adding') : t('common.add') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import * as FirewallService from '../../../../bindings/changeme/ssh/firewallservice.js'
import { showMessage } from '../../../utils/message'
import { useConfirm } from '../../../utils/confirm'
import { addLog, LogType, LogLevel } from '../../../utils/logger'

const { t } = useI18n()
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

const targetLabel = (val) => {
  const map = { ACCEPT: t('firewall.targetLabels.ACCEPT'), allow: t('firewall.allow'), DROP: t('firewall.targetLabels.DROP'), deny: t('firewall.deny'), REJECT: t('firewall.targetLabels.REJECT') }
  return map[val] || val
}

const chainLabel = (c) => {
  const map = { INPUT: t('firewall.chainLabels.INPUT'), OUTPUT: t('firewall.chainLabels.OUTPUT'), FORWARD: t('firewall.chainLabels.FORWARD'), IN: t('firewall.chainLabels.IN'), OUT: t('firewall.chainLabels.OUT'), public: t('firewall.chainLabels.public'), trusted: t('firewall.chainLabels.trusted'), drop: t('firewall.chainLabels.drop'), block: t('firewall.chainLabels.block'), dmz: t('firewall.chainLabels.dmz') }
  return map[c] || c
}

const protocolLabel = (p) => {
  const map = { tcp: 'TCP', udp: 'UDP', icmp: 'ICMP', all: t('firewall.protocolLabels.all'), service: t('firewall.protocolLabels.service') }
  return map[p] || (p || '-')
}

const zoneLabel = (z) => {
  const map = { public: t('firewall.zoneLabels.public'), trusted: t('firewall.zoneLabels.trusted'), drop: t('firewall.zoneLabels.drop'), block: t('firewall.zoneLabels.block'), dmz: t('firewall.zoneLabels.dmz'), home: t('firewall.zoneLabels.home'), work: t('firewall.zoneLabels.work'), internal: t('firewall.zoneLabels.internal'), external: t('firewall.zoneLabels.external') }
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
    addError.value = t('firewall.pleaseFillPort')
    return
  }
  addSubmitting.value = true
  addError.value = ''
  try {
    await FirewallService.AddRule(connId, addForm.chain, addForm.target, addForm.protocol, addForm.port, addForm.source, addForm.comment)
    addLog(connId, 'firewall', LogLevel.SUCCESS,
      `添加防火墙规则: ${addForm.chain} ${addForm.target} ${addForm.protocol} ${addForm.port}`)
    showMessage(t('firewall.ruleAdded'), 'success')
    showAdd.value = false
    addForm.port = ''
    addForm.source = ''
    addForm.comment = ''
    await loadInfo()
  } catch (e) {
    addError.value = String(e?.message || e)
    addLog(connId, 'firewall', LogLevel.ERROR, t('firewall.addFailed') + ': ' + (e?.message || e))
  } finally {
    addSubmitting.value = false
  }
}

const deleteRule = async (rule) => {
  const ok = await confirm({
    title: t('firewall.deleteRule'),
    message: t('firewall.deleteConfirm', { index: rule.index, info: `${chainLabel(rule.chain)} · ${targetLabel(rule.target)} · ${rule.port || '-'}` }),
    danger: true,
  })
  if (!ok) return
  try {
    await FirewallService.DeleteRule(connId, rule.chain, rule.index, rule.port, rule.protocol)
    addLog(connId, 'firewall', LogLevel.INFO, `删除防火墙规则: #${rule.index} ${chainLabel(rule.chain)} ${rule.port || ''}`)
    showMessage(t('firewall.ruleDeleted'), 'success')
    await loadInfo()
  } catch (e) {
    addLog(connId, 'firewall', LogLevel.ERROR, '删除防火墙规则失败: ' + (e?.message || e))
    showMessage(t('firewall.deleteFailed') + ': ' + (e?.message || e), 'error')
  }
}

const toggleFirewall = async () => {
  const enable = fwInfo.value.status !== 'active'
  const ok = await confirm({
    title: enable ? t('firewall.enableFirewall') : t('firewall.disableFirewall'),
    message: enable ? t('firewall.enableConfirm') : t('firewall.disableConfirm'),
    danger: !enable,
  })
  if (!ok) return
  try {
    await FirewallService.ToggleFirewall(connId, enable)
    const msg = enable ? t('firewall.enabled') : t('firewall.disabled')
    addLog(connId, 'firewall', enable ? LogLevel.SUCCESS : LogLevel.WARNING, msg)
    showMessage(msg, 'success')
    await loadInfo()
  } catch (e) {
    const failMsg = enable ? t('firewall.enableFailed') : t('firewall.disableFailed')
    addLog(connId, 'firewall', LogLevel.ERROR, failMsg + ': ' + (e?.message || e))
    showMessage(failMsg + ': ' + (e?.message || e), 'error')
  }
}

const runCustom = async () => {
  if (!customCmd.value.trim()) return
  running.value = true
  cmdResult.value = ''
  cmdError.value = false
  try {
    const result = await FirewallService.RunCustomCommand(connId, customCmd.value)
    cmdResult.value = result || t('firewall.cmdResultEmpty')
    addLog(connId, 'firewall', LogLevel.INFO, `执行防火墙命令: ${customCmd.value}`, { output: result })
    customCmd.value = ''
    showMessage(t('firewall.cmdSuccess'), 'success')
    await loadInfo()
  } catch (e) {
    cmdResult.value = String(e?.message || e)
    cmdError.value = true
    addLog(connId, 'firewall', LogLevel.ERROR, `防火墙命令失败: ${customCmd.value}`, { error: String(e?.message || e) })
    showMessage(t('firewall.cmdFailed'), 'error')
  } finally {
    running.value = false
  }
}

onMounted(() => loadInfo())
</script>

<style scoped>
.icon-warning { color: var(--warning-light); }
.icon-disabled { color: var(--text-disabled); }

.fw-panel {
  width: 100%; height: 100%;
  display: flex; flex-direction: column;
  background: var(--bg-panel);
  overflow: hidden;
}

.fw-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--surface-hover);
  background: var(--toolbar-3);
  flex-shrink: 0;
}
.fw-toolbar-left { display: flex; align-items: center; gap: 0.5rem; }
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
  background: var(--border-subtle);
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
.fw-btn-success { background: var(--success-bg); border-color: var(--border-success); color: var(--success-light); }
.fw-btn-warn { background: var(--warning-bg); border-color: var(--border-warning); color: var(--warning-light); }
.fw-btn-danger { background: var(--danger-bg); border-color: var(--border-danger); color: var(--accent-danger); }

.fw-content { flex: 1; overflow-y: auto; min-height: 0; }

.fw-loading, .fw-empty, .fw-empty-rules {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 100%; gap: 0.75rem; color: var(--text-muted);
}
.fw-loading p, .fw-empty p, .fw-empty-rules p { margin: 0; font-size: 0.875rem; }
.fw-hint { font-size: 0.75rem !important; color: var(--text-disabled) !important; }
.fw-empty-rules { height: auto; padding: 2rem; }

.fw-spinner {
  width: 24px; height: 24px;
  border: 2px solid var(--warning-bg); border-top-color: var(--warning-light);
  border-radius: 50%; animation: fw-spin 0.8s linear infinite;
}
@keyframes fw-spin { to { transform: rotate(360deg); } }

.fw-body { padding: 0.75rem; }

.fw-table-wrap { overflow-x: auto; margin-bottom: 0.75rem; }
.fw-table { width: 100%; border-collapse: collapse; font-size: 0.75rem; }
.fw-table th {
  padding: 0.5rem 0.625rem; text-align: left;
  color: var(--text-secondary); font-weight: 600; white-space: nowrap;
  border-bottom: 2px solid var(--surface-hover);
  background: var(--toolbar-3);
  position: sticky; top: 0; z-index: 1;
}
.fw-table td {
  padding: 0.375rem 0.625rem;
  color: var(--text-primary); border-bottom: 1px solid var(--surface-1);
  white-space: nowrap;
}
.fw-table tr:hover td { background: var(--bg-hover); }
.fw-td-num { color: var(--text-muted); font-family: monospace; }
.fw-mono { font-family: 'Courier New', monospace; font-size: 0.6875rem; }

.fw-chain-tag {
  display: inline-block; padding: 0.0625rem 0.375rem;
  background: var(--primary-bg); border-radius: 0.25rem;
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
.fw-raw-toggle:hover { background: var(--surface-1); }
.fw-raw {
  margin: 0; padding: 0.75rem;
  color: var(--text-secondary); font-size: 0.6875rem;
  font-family: 'Courier New', monospace; line-height: 1.5;
  white-space: pre; overflow-x: auto;
  max-height: 30vh; background: var(--bg-input);
}

.fw-cmd-bar {
  display: flex; gap: 0.5rem; padding: 0.5rem 1rem;
  border-top: 1px solid var(--border-subtle);
  background: var(--toolbar-3); flex-shrink: 0;
}
.fw-cmd-input {
  flex: 1; background: var(--bg-panel);
  border: 1px solid var(--surface-hover); border-radius: 0.375rem;
  color: var(--text-primary); font-size: 0.75rem; font-family: 'Courier New', monospace;
  padding: 0.375rem 0.625rem; outline: none;
}
.fw-cmd-input:focus { border-color: var(--border-accent); }

.fw-footer {
  display: flex; align-items: center;
  padding: 0.25rem 1rem;
  border-top: 1px solid var(--surface-hover);
  background: var(--toolbar-3);
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
  background: var(--bg-panel-solid); border: 1px solid var(--surface-hover);
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
  border: 1px solid var(--surface-hover); border-radius: 0.375rem;
  color: var(--text-primary); font-size: 0.8125rem; padding: 0.5rem 0.75rem;
  outline: none; box-sizing: border-box;
}
.fw-input:focus { border-color: var(--border-accent); }
.fw-error { color: var(--accent-danger); font-size: 0.75rem; }
</style>
