<template>
  <div class="settings-page">
    <div class="page-header">
      <div class="header-top">
        <button class="back-btn" @click="goBack" title="返回">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 12H5M12 19l-7-7 7-7"/>
          </svg>
          <span>返回</span>
        </button>
      </div>
      <h1 class="page-title">设置</h1>
      <p class="page-description">配置应用程序的行为和外观</p>
      <div class="version-info">
        <span class="version-label">版本</span>
        <span class="version-value">v{{ version }}</span>
      </div>
    </div>

    <div class="config-section">
      <div class="section-title">界面</div>

      <div class="config-list">
        <!-- 主题 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">主题</div>
            <div class="config-desc">
              切换应用程序的整体外观。浅色主题下终端也会同步切换。
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
      </div>
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
        <div class="config-item" style="margin-top:12px;">
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
        <div class="config-item" style="margin-top:12px;">
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
import { GetVersion } from '../../../bindings/changeme/ssh/greetservice.js'
import Modal from '../../components/Modal.vue'
import Message from '../../components/Message.vue'

const router = useRouter()
const configStore = useConfigStore()
const messageRef = ref(null)
const version = ref('0.2.0')

const defaultType = ref('classic')
const switchMode = ref('prompt')
const fontSize = ref(14)
const theme = ref('dark')
const showImport = ref(false)
const importJson = ref('')

function goBack() {
  router.back()
}

onMounted(async () => {
  try {
    version.value = await GetVersion()
  } catch (e) {
    console.warn('获取版本失败:', e)
  }
  await configStore.init()
  defaultType.value = configStore.getDefaultTerminalType()
  switchMode.value = configStore.get('terminal', 'switchMode') || 'prompt'
  fontSize.value = configStore.get('terminal', 'fontSize') || 14
  theme.value = configStore.get('ui', 'theme') || 'dark'
})

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

async function handleThemeChange(e) {
  theme.value = e.target.value
  await configStore.setTheme(e.target.value)
  showToast('主题已切换')
}

async function changeFontSize(delta) {
  const newSize = Math.min(24, Math.max(10, fontSize.value + delta))
  if (newSize !== fontSize.value) {
    fontSize.value = newSize
    await configStore.set('terminal', 'fontSize', newSize)
    showToast('设置已保存')
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
  margin-bottom: 24px;
}

.header-top {
  margin-bottom: 16px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--surface-2);
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
}

.page-title {
  margin: 0 0 8px;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

.page-description {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted);
}

.version-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}

.version-label {
  font-size: 12px;
  color: var(--text-muted);
}

.version-value {
  font-size: 12px;
  color: var(--primary-light);
  padding: 2px 8px;
  background: var(--primary-bg);
  border-radius: 4px;
}

.config-section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-subtle);
}

.config-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.config-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  transition: border-color 0.15s;
}

.config-item:hover {
  border-color: var(--border-default);
}

.config-info {
  flex: 1;
  min-width: 0;
  margin-right: 24px;
}

.config-label {
  font-size: 14px;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.config-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
}

.config-control {
  flex-shrink: 0;
}

.select-input {
  padding: 6px 32px 6px 12px;
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1L5 5L9 1' stroke='%2394a3b8' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
}

.select-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.select-input option {
  background: var(--bg-input);
  color: var(--text-primary);
}

.section-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding: 10px 14px;
  background: var(--primary-bg);
  border: 1px solid var(--border-accent);
  border-radius: 8px;
  color: var(--primary-light);
  font-size: 12px;
}

.global-actions {
  display: flex;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: var(--surface-2);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.import-textarea {
  width: 100%;
  min-height: 200px;
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
.toggle{position:relative;display:inline-block;width:40px;height:22px;cursor:pointer}
.toggle input{opacity:0;width:0;height:0}
.toggle-slider{position:absolute;inset:0;background:var(--border-default);border-radius:11px;transition:.2s}
.toggle-slider::before{content:'';position:absolute;height:16px;width:16px;left:3px;bottom:3px;background:var(--text-muted);border-radius:50%;transition:.2s}
.toggle input:checked+.toggle-slider{background:var(--accent-success)}
.toggle input:checked+.toggle-slider::before{transform:translateX(18px);background:var(--text-on-accent)}

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
</style>
