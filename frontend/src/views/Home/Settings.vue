<template>
  <div class="settings-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack" title="返回">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
        <span>返回</span>
      </button>
      <h1 class="page-title">设置</h1>
    </div>

    <div class="config-section">
      <div class="section-title">终端</div>

      <div class="config-list">
        <!-- 默认终端类型 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">默认终端类型</div>
            <div class="config-desc">
              新建终端时使用的类型。结构化终端适合查看命令历史，经典终端兼容性最好。
            </div>
          </div>
          <div class="config-control">
            <select
              :value="defaultType"
              @change="handleChange($event)"
              class="select-input"
            >
              <option value="structured">结构化终端（推荐）</option>
              <option value="classic">经典终端</option>
            </select>
          </div>
        </div>

        <!-- 交互式操作模式 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">结构化终端交互式操作模式</div>
            <div class="config-desc">
              按 Ctrl+R/Z 等交互式快捷键时的处理方式。
            </div>
          </div>
          <div class="config-control">
            <select
              :value="switchMode"
              @change="handleSwitchModeChange($event)"
              class="select-input"
            >
              <option value="prompt">跳转至经典终端</option>
              <option value="inline">临时终端</option>
            </select>
          </div>
        </div>

        <!-- 终端字体大小 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">终端字体大小</div>
            <div class="config-desc">
              终端显示的字体大小，范围 10-24。
            </div>
          </div>
          <div class="config-control">
            <div class="number-control">
              <button class="number-btn" @click="changeFontSize(-1)" :disabled="fontSize <= 10">-</button>
              <span class="number-value">{{ fontSize }}</span>
              <button class="number-btn" @click="changeFontSize(1)" :disabled="fontSize >= 24">+</button>
            </div>
          </div>
        </div>

        <!-- 命令发送方式 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">结构化终端命令发送方式</div>
            <div class="config-desc">
              回车发送：按 Enter 直接发送命令。<br>
              按钮发送：按 Enter 换行，点击按钮发送，支持多行命令。
            </div>
          </div>
          <div class="config-control">
            <select
              :value="commandSendMode"
              @change="handleCommandSendModeChange($event)"
              class="select-input"
            >
              <option value="enter">回车发送</option>
              <option value="button">按钮发送</option>
            </select>
          </div>
        </div>

        <!-- 经典终端代码高亮 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">经典终端代码高亮</div>
            <div class="config-desc">
              开启后，经典终端输出会自动对代码关键字、字符串、数字进行语法高亮。
            </div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="codeHighlight" @change="handleCodeHighlightChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
      </div>
    </div>

    <div class="config-section">
      <div class="section-title">窗口</div>

      <div class="config-list">
        <!-- 主题 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">主题</div>
            <div class="config-desc">
              选择界面主题风格，切换后所有窗口同步生效。
            </div>
          </div>
          <div class="config-control">
            <select
              :value="theme"
              @change="handleThemeChange($event)"
              class="select-input"
            >
              <option value="dark">深色</option>
              <option value="light">浅色</option>
            </select>
          </div>
        </div>

        <!-- 连接后自动最小化到托盘 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">连接后自动最小化到托盘</div>
            <div class="config-desc">
              开启后，SSH 连接成功时自动将主窗口最小化到系统托盘。
            </div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="autoTray" @change="handleAutoTrayChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- 记忆窗口位置 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">记忆窗口位置和大小</div>
            <div class="config-desc">
              开启后，关闭窗口时自动保存窗口的位置和大小，下次打开时恢复到上次的状态。
            </div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="rememberPosition" @change="handleRememberPositionChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- SSH窗口关闭后自动显示首页 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">SSH窗口关闭后自动显示首页</div>
            <div class="config-desc">
              开启后，当所有SSH窗口关闭时，首页窗口会自动显示并获得焦点。
            </div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="autoShowHome" @change="handleAutoShowHomeChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- 清除窗口位置记忆 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">清除窗口位置记忆</div>
            <div class="config-desc">
              清除已保存的所有窗口位置和大小记录，下次打开窗口将使用默认位置。
            </div>
          </div>
          <div class="config-control">
            <button class="action-btn-sm" @click="clearWindowPositions">清除</button>
          </div>
        </div>
      </div>
    </div>

    <div class="config-section">
      <div class="section-title">私有云端</div>

      <div class="config-list">
        <!-- 启用云端 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">启用私有云端同步</div>
            <div class="config-desc">
              连接到私有云端服务器，实现跨设备连接配置同步。
            </div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="cloudEnabled" @change="handleCloudEnabledChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- 服务器地址 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">服务器地址</div>
            <div class="config-desc">
              私有云端服务器 IP 和端口
            </div>
          </div>
          <div class="config-control cloud-addr">
            <input
              type="text"
              v-model="cloudHost"
              class="text-input cloud-host"
              placeholder="192.168.1.100"
              @change="handleCloudConfigChange"
            >
            <span class="cloud-sep">:</span>
            <input
              type="number"
              v-model.number="cloudPort"
              class="text-input cloud-port"
              placeholder="9527"
              @change="handleCloudConfigChange"
            >
          </div>
        </div>

        <!-- 认证令牌 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">认证令牌</div>
            <div class="config-desc">
              云端服务器的认证令牌，与服务端启动时设置的令牌一致。
            </div>
          </div>
          <div class="config-control">
            <input
              type="password"
              v-model="cloudToken"
              class="text-input"
              placeholder="输入令牌"
              @change="handleCloudConfigChange"
            >
          </div>
        </div>

        <!-- 同步状态 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">连接状态</div>
            <div class="config-desc">
              {{ cloudStatusText }}
            </div>
          </div>
          <div class="config-control">
            <button class="action-btn-sm" @click="testCloudConnection">
              测试连接
            </button>
          </div>
        </div>

        <!-- 自动同步到云端 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">自动同步到云端</div>
            <div class="config-desc">
              开启后，本地连接配置变更时自动上传到云端，其他设备可同步。
            </div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="cloudAutoSyncTo" @change="handleAutoSyncToChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- 自动从云端同步 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">自动从云端同步</div>
            <div class="config-desc">
              开启后，自动从云端下载其他设备的连接配置，按 IP:端口 去重，最新覆盖旧的。
            </div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="cloudAutoSyncFrom" @change="handleAutoSyncFromChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- 手动同步 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">手动同步</div>
            <div class="config-desc">
              立即上传本地连接配置到云端，或从云端下载。
            </div>
          </div>
          <div class="config-control sync-btns">
            <button class="action-btn-sm" @click="syncUpload" :disabled="!cloudEnabled">上传</button>
            <button class="action-btn-sm" @click="syncDownload" :disabled="!cloudEnabled">下载</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 导入/导出 -->
    <div class="global-actions">
      <button class="action-btn" @click="exportConfig">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        导出
      </button>
      <button class="action-btn" @click="showImport = true">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        导入
      </button>
    </div>

    <!-- 导入对话框 -->
    <Modal
      v-model:visible="showImport"
      title="导入配置"
      @confirm="handleImport"
    >
      <template #content>
        <textarea v-model="importJson" class="import-textarea" placeholder="粘贴配置 JSON..." rows="10"></textarea>
      </template>
    </Modal>

    <!-- 消息提示 -->
    <Message ref="messageRef" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useConfigStore } from '../../stores/config'
import { Events } from '@wailsio/runtime'
import * as SSHService from '../../../bindings/changeme/ssh/sshservice.js'
import * as CloudService from '../../../bindings/changeme/ssh/cloudservice.js'
import Modal from '../../components/Modal.vue'
import Message from '../../components/Message.vue'

const router = useRouter()
const configStore = useConfigStore()
const messageRef = ref(null)

const theme = ref('dark')
const defaultType = ref('classic')
const switchMode = ref('prompt')
const fontSize = ref(14)
const commandSendMode = ref('enter')
const codeHighlight = ref(false)
const autoTray = ref(false)
const rememberPosition = ref(true)
const autoShowHome = ref(true)
const showImport = ref(false)
const importJson = ref('')

// 云端配置
const cloudEnabled = ref(false)
const cloudHost = ref('')
const cloudPort = ref(9527)
const cloudToken = ref('')
const cloudAutoSyncTo = ref(false)
const cloudAutoSyncFrom = ref(false)
const cloudStatusText = ref('未连接')

function goBack() {
  router.back()
}

onMounted(async () => {
  await configStore.init()
  theme.value = configStore.get('ui', 'theme') || 'dark'
  defaultType.value = configStore.getDefaultTerminalType()
  switchMode.value = configStore.get('terminal', 'switchMode') || 'prompt'
  fontSize.value = configStore.get('terminal', 'fontSize') || 14
  commandSendMode.value = configStore.get('terminal', 'commandSendMode') || 'enter'
  codeHighlight.value = configStore.get('terminal', 'codeHighlight') || false
  autoTray.value = configStore.get('ui', 'autoTray') || false
  rememberPosition.value = configStore.get('ui', 'rememberPosition') !== false
  autoShowHome.value = configStore.get('ui', 'autoShowHome') !== false

  // 加载云端配置
  cloudEnabled.value = configStore.get('cloud', 'enabled') || false
  cloudToken.value = configStore.get('cloud', 'token') || ''
  cloudAutoSyncTo.value = configStore.get('cloud', 'autoSyncTo') || false
  cloudAutoSyncFrom.value = configStore.get('cloud', 'autoSyncFrom') || false
  const serverUrl = configStore.get('cloud', 'serverUrl') || ''
  if (serverUrl) {
    const clean = serverUrl.replace(/^https?:\/\//, '')
    const parts = clean.split(':')
    cloudHost.value = parts[0] || ''
    cloudPort.value = parseInt(parts[1]) || 9527
  }

  // 检查实际连接状态
  try {
    const connected = await CloudService.IsConnected()
    cloudStatusText.value = connected ? '已连接' : '未连接'
  } catch {
    cloudStatusText.value = '未连接'
  }

  // 监听云端状态变化
  Events.On('cloud:status', (e) => {
    const d = e?.data
    if (d) {
      cloudStatusText.value = d.connected ? '已连接' : '未连接'
    }
  })
})

async function handleThemeChange(e) {
  theme.value = e.target.value
  await configStore.setTheme(e.target.value)
  showToast('主题已切换')
}

async function handleChange(e) {
  defaultType.value = e.target.value
  await configStore.set('terminal', 'defaultType', e.target.value)
  showToast('设置已保存')
}

async function handleSwitchModeChange(e) {
  switchMode.value = e.target.value
  await configStore.set('terminal', 'switchMode', e.target.value)
  showToast('设置已保存')
}

async function changeFontSize(delta) {
  const newSize = Math.min(24, Math.max(10, fontSize.value + delta))
  if (newSize !== fontSize.value) {
    fontSize.value = newSize
    await configStore.set('terminal', 'fontSize', newSize)
    showToast('设置已保存')
  }
}

async function handleCommandSendModeChange(e) {
  commandSendMode.value = e.target.value
  await configStore.set('terminal', 'commandSendMode', e.target.value)
  showToast('设置已保存')
}

async function handleCodeHighlightChange(e) {
  codeHighlight.value = e.target.checked
  await configStore.set('terminal', 'codeHighlight', e.target.checked)
  showToast('设置已保存')
}

async function handleAutoTrayChange(e) {
  autoTray.value = e.target.checked
  await configStore.set('ui', 'autoTray', e.target.checked)
  showToast('设置已保存')
}

async function handleRememberPositionChange(e) {
  rememberPosition.value = e.target.checked
  await configStore.set('ui', 'rememberPosition', e.target.checked)
  showToast('设置已保存')
}

async function handleAutoShowHomeChange(e) {
  autoShowHome.value = e.target.checked
  await configStore.set('ui', 'autoShowHome', e.target.checked)
  showToast('设置已保存')
}

async function clearWindowPositions() {
  try {
    await SSHService.ClearWindowPositions()
    showToast('窗口位置记忆已清除')
  } catch (e) {
    showToast('清除失败')
  }
}

async function handleCloudEnabledChange(e) {
  cloudEnabled.value = e.target.checked
  await configStore.set('cloud', 'enabled', e.target.checked)
  showToast('设置已保存')
}

async function handleAutoSyncToChange(e) {
  cloudAutoSyncTo.value = e.target.checked
  await configStore.set('cloud', 'autoSyncTo', e.target.checked)
  showToast('设置已保存')
}

async function handleAutoSyncFromChange(e) {
  cloudAutoSyncFrom.value = e.target.checked
  await configStore.set('cloud', 'autoSyncFrom', e.target.checked)
  showToast('设置已保存')
}

async function handleCloudConfigChange() {
  await configStore.set('cloud', 'serverUrl', `${cloudHost.value}:${cloudPort.value || 9527}`)
  await configStore.set('cloud', 'token', cloudToken.value)
}

async function testCloudConnection() {
  if (!cloudHost.value || !cloudToken.value) {
    showToast('请填写服务器地址和令牌')
    return
  }
  cloudStatusText.value = '连接中...'
  const addr = `${cloudHost.value}:${cloudPort.value || 9527}`
  console.log('[Settings] 测试连接:', addr)
  const ok = await CloudService.Connect(addr, cloudToken.value)
  if (ok) {
    cloudStatusText.value = '已连接'
    showToast('连接成功')
    await CloudService.Disconnect()
  } else {
    cloudStatusText.value = '连接失败'
    showToast('连接失败')
  }
}

async function ensureCloudConnected() {
  const connected = await CloudService.IsConnected()
  if (connected) return true
  if (!cloudHost.value || !cloudToken.value) {
    showToast('请先配置服务器地址和令牌')
    return false
  }
  const addr = `${cloudHost.value}:${cloudPort.value || 9527}`
  await CloudService.Connect(addr, cloudToken.value)
  return true
}

async function syncUpload() {
  try {
    if (!await ensureCloudConnected()) return

    // 1. 拉取云端已有数据
    let cloudConns = []
    try {
      cloudConns = await CloudService.PullSync() || []
    } catch {}

    // 2. 获取本地连接
    const localConns = await SSHService.GetAllConnections() || []
    const now = new Date().toISOString()

    // 3. 合并：本地连接覆盖同 host:port 的云端连接，保留云端独有的连接
    const merged = new Map()
    // 先放入云端连接（非本设备来源的）
    for (const c of cloudConns) {
      const key = `${c.host}:${c.port}`
      merged.set(key, c)
    }
    // 本地连接覆盖（标记为本设备来源）
    for (const c of localConns) {
      const key = `${c.host}:${c.port}`
      merged.set(key, {
        id: c.id, name: c.name, host: c.host, port: c.port || 22,
        username: c.username, password: c.password || '', keyPath: c.keyPath || '',
        source: 'local', updatedAt: now,
      })
    }

    // 4. 推送合并后的完整数据
    await CloudService.PushSync(Array.from(merged.values()))
    showToast(`上传成功，共 ${merged.size} 个连接`)
  } catch (e) {
    showToast('上传失败: ' + (e?.message || e))
  }
}

async function syncDownload() {
  try {
    if (!await ensureCloudConnected()) return
    const conns = await CloudService.PullSync()
    if (!conns || conns.length === 0) {
      showToast('云端无连接配置')
      return
    }

    let count = 0
    for (const c of conns) {
      try {
        await SSHService.SyncImportConnection({
          name: c.name || `${c.username}@${c.host}`,
          host: c.host,
          port: c.port || 22,
          username: c.username,
          password: c.password || '',
          keyPath: c.keyPath || '',
        })
        count++
      } catch (e) {
        console.warn('[Cloud] 同步导入失败:', c.name, e.message)
      }
    }
    showToast(`同步完成，处理了 ${count} 个连接`)
  } catch (e) {
    showToast('下载失败: ' + (e?.message || e))
  }
}

function showToast(msg) {
  messageRef.value?.success(msg)
}

async function exportConfig() {
  const json = await configStore.exportConfig?.() || JSON.stringify(configStore.config, null, 2)
  const blob = new Blob([json], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `pzssh-config-${new Date().toISOString().slice(0, 10)}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

async function handleImport() {
  if (importJson.value) {
    const ok = await configStore.importConfig?.(importJson.value)
    if (ok) {
      showImport.value = false
      importJson.value = ''
      defaultType.value = configStore.getDefaultTerminalType()
    }
  }
}
</script>

<style scoped>
.settings-page {
  height: 100%;
  padding: 24px;
  overflow-y: auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: transparent;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.back-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.config-section {
  margin-bottom: 28px;
}

.section-title {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  margin-bottom: 8px;
  padding: 0;
  border: none;
}

.config-list {
  display: flex;
  flex-direction: column;
  gap: 1px;
  background: var(--border-subtle);
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--border-subtle);
}

.config-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  background: var(--bg-panel-solid);
  transition: background 0.15s;
}

.config-item:hover {
  background: var(--surface-hover);
}

.config-info {
  flex: 1;
  min-width: 0;
}

.config-label {
  font-size: 13px;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.config-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.4;
}

.config-control {
  flex-shrink: 0;
}

.select-input {
  padding: 5px 28px 5px 10px;
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1L5 5L9 1' stroke='%2394a3b8' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
  transition: border-color 0.15s;
}

.select-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.select-input option {
  background: var(--bg-input);
  color: var(--text-primary);
}

.global-actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  padding: 14px 16px;
  background: var(--surface-1);
  border-radius: 10px;
  border: 1px solid var(--border-subtle);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: transparent;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}

.action-btn-sm {
  padding: 4px 12px;
  background: transparent;
  border: 1px solid var(--border-default);
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn-sm:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}

.text-input {
  width: 180px;
  padding: 5px 10px;
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 12px;
  transition: border-color 0.15s;
}

.text-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.text-input::placeholder {
  color: var(--text-muted);
}

.import-textarea {
  width: 100%;
  min-height: 180px;
  padding: 12px;
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  color: var(--text-primary);
  font-family: 'Cascadia Code', monospace;
  font-size: 12px;
  resize: vertical;
}

.import-textarea:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.toggle{position:relative;display:inline-block;width:36px;height:20px;cursor:pointer}
.toggle input{opacity:0;width:0;height:0}
.toggle-slider{position:absolute;inset:0;background:var(--border-default);border-radius:10px;transition:.2s}
.toggle-slider::before{content:'';position:absolute;height:14px;width:14px;left:3px;bottom:3px;background:var(--text-muted);border-radius:50%;transition:.2s}
.toggle input:checked+.toggle-slider{background:var(--accent-success)}
.toggle input:checked+.toggle-slider::before{transform:translateX(16px);background:var(--text-on-accent)}

.number-control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.number-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--surface-2);
  border: 1px solid var(--border-default);
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 16px;
  cursor: pointer;
  transition: all 0.15s;
}

.number-btn:hover:not(:disabled) {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.number-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.number-value {
  min-width: 24px;
  text-align: center;
  font-size: 14px;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}

.cloud-addr {
  display: flex;
  align-items: center;
  gap: 4px;
}

.cloud-host { width: 130px; }
.cloud-port { width: 64px; }
.cloud-sep { color: var(--text-muted); font-size: 14px; }

.sync-btns {
  display: flex;
  gap: 6px;
}
</style>
