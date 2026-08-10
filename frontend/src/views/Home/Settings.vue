<template>
  <div class="settings-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack" :title="t('settings.back')">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
        <span>{{ t('settings.back') }}</span>
      </button>
      <h1 class="page-title">{{ t('settings.title') }}</h1>
    </div>

    <!-- 语言设置 -->
    <div class="config-section">
      <div class="section-title">{{ t('settings.language') }}</div>
      <div class="config-list">
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.languageLabel') }}</div>
            <div class="config-desc">{{ t('settings.languageDesc') }}</div>
          </div>
          <div class="config-control">
            <select
              :value="currentLocale"
              @change="handleLocaleChange($event)"
              class="select-input"
            >
              <option value="zh-CN">中文</option>
              <option value="en-US">English</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <div class="config-section">
      <div class="section-title">{{ t('settings.terminal') }}</div>

      <div class="config-list">
        <!-- 默认终端类型 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.defaultTerminalType') }}</div>
            <div class="config-desc">{{ t('settings.defaultTerminalDesc') }}</div>
          </div>
          <div class="config-control">
            <select
              :value="defaultType"
              @change="handleChange($event)"
              class="select-input"
            >
              <option value="structured">{{ t('settings.structuredTerminal') }}</option>
              <option value="classic">{{ t('settings.classicTerminal') }}</option>
            </select>
          </div>
        </div>

        <!-- 交互式操作模式 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.interactiveMode') }}</div>
            <div class="config-desc">{{ t('settings.interactiveModeDesc') }}</div>
          </div>
          <div class="config-control">
            <select
              :value="switchMode"
              @change="handleSwitchModeChange($event)"
              class="select-input"
            >
              <option value="prompt">{{ t('settings.switchToClassic') }}</option>
              <option value="inline">{{ t('settings.inlineTerminal') }}</option>
            </select>
          </div>
        </div>

        <!-- 终端字体大小 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.fontSize') }}</div>
            <div class="config-desc">{{ t('settings.fontSizeDesc') }}</div>
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
            <div class="config-label">{{ t('settings.commandSendMode') }}</div>
            <div class="config-desc">{{ t('settings.commandSendModeDesc') }}</div>
          </div>
          <div class="config-control">
            <select
              :value="commandSendMode"
              @change="handleCommandSendModeChange($event)"
              class="select-input"
            >
              <option value="enter">{{ t('settings.enterSend') }}</option>
              <option value="button">{{ t('settings.buttonSend') }}</option>
            </select>
          </div>
        </div>

        <!-- 经典终端代码高亮 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.codeHighlight') }}</div>
            <div class="config-desc">{{ t('settings.codeHighlightDesc') }}</div>
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
      <div class="section-title">{{ t('settings.shortcuts') }}</div>
      <div class="config-list">
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.enableGlobalShortcuts') }}</div>
            <div class="config-desc">{{ t('settings.enableGlobalShortcutsDesc') }}</div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="shortcutsEnabled" @change="handleShortcutsChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">Ctrl+← / Ctrl+→</div>
            <div class="config-desc">{{ t('settings.switchTab') }}</div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="shortcutSwitchTab" @change="handleShortcutSwitchTabChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">Ctrl+Shift+S</div>
            <div class="config-desc">{{ t('settings.saveGroup') }}</div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="shortcutSaveGroup" @change="handleShortcutSaveGroupChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">Ctrl+Shift+U</div>
            <div class="config-desc">{{ t('settings.cloudUpload') }}</div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="shortcutCloudUpload" @change="handleShortcutCloudUploadChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">Ctrl+Shift+D</div>
            <div class="config-desc">{{ t('settings.cloudDownload') }}</div>
          </div>
          <div class="config-control">
            <label class="toggle">
              <input type="checkbox" :checked="shortcutCloudDownload" @change="handleShortcutCloudDownloadChange($event)">
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
      </div>
    </div>

    <div class="config-section">
      <div class="section-title">{{ t('settings.advanced') }}</div>
      <div class="config-list">
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.groupBehavior') }}</div>
            <div class="config-desc">{{ t('settings.groupBehaviorDesc') }}</div>
          </div>
          <div class="config-control">
            <select
              :value="groupBehavior"
              @change="handleGroupBehaviorChange($event)"
              class="select-input"
            >
              <option value="prompt">{{ t('settings.promptSelect') }}</option>
              <option value="join_default">{{ t('settings.joinDefault') }}</option>
              <option value="new_window">{{ t('settings.newWindow') }}</option>
            </select>
          </div>
        </div>
      </div>
    </div>
    <div class="config-section">
      <div class="section-title">{{ t('settings.window') }}</div>

      <div class="config-list">
        <!-- 主题 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.theme') }}</div>
            <div class="config-desc">{{ t('settings.themeDesc') }}</div>
          </div>
          <div class="config-control">
            <select
              :value="theme"
              @change="handleThemeChange($event)"
              class="select-input"
            >
              <option value="dark">{{ t('settings.dark') }}</option>
              <option value="light">{{ t('settings.light') }}</option>
            </select>
          </div>
        </div>

        <!-- 连接后自动最小化到托盘 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.autoTray') }}</div>
            <div class="config-desc">{{ t('settings.autoTrayDesc') }}</div>
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
            <div class="config-label">{{ t('settings.rememberPosition') }}</div>
            <div class="config-desc">{{ t('settings.rememberPositionDesc') }}</div>
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
            <div class="config-label">{{ t('settings.autoShowHome') }}</div>
            <div class="config-desc">{{ t('settings.autoShowHomeDesc') }}</div>
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
            <div class="config-label">{{ t('settings.clearWindowPositions') }}</div>
            <div class="config-desc">{{ t('settings.clearWindowPositionsDesc') }}</div>
          </div>
          <div class="config-control">
            <button class="action-btn-sm" @click="clearWindowPositions">{{ t('settings.clear') }}</button>
          </div>
        </div>
      </div>
    </div>

    <div class="config-section">
      <div class="section-title">{{ t('settings.cloud') }}</div>

      <div class="config-list">
        <!-- 启用云端 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.cloudEnabled') }}</div>
            <div class="config-desc">{{ t('settings.cloudEnabledDesc') }}</div>
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
            <div class="config-label">{{ t('settings.serverAddress') }}</div>
            <div class="config-desc">{{ t('settings.serverAddressDesc') }}</div>
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
            <div class="config-label">{{ t('settings.authToken') }}</div>
            <div class="config-desc">{{ t('settings.authTokenDesc') }}</div>
          </div>
          <div class="config-control">
            <input
              type="password"
              v-model="cloudToken"
              class="text-input"
              :placeholder="t('settings.inputToken')"
              @change="handleCloudConfigChange"
            >
          </div>
        </div>

        <!-- 同步状态 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.connectionStatus') }}</div>
            <div class="config-desc">
              {{ cloudStatusText }}
            </div>
          </div>
          <div class="config-control">
            <button class="action-btn-sm" @click="testCloudConnection">
              {{ t('settings.testConnection') }}
            </button>
          </div>
        </div>

        <!-- 自动同步到云端 -->
        <div class="config-item">
          <div class="config-info">
            <div class="config-label">{{ t('settings.autoSyncTo') }}</div>
            <div class="config-desc">{{ t('settings.autoSyncToDesc') }}</div>
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
            <div class="config-label">{{ t('settings.autoSyncFrom') }}</div>
            <div class="config-desc">{{ t('settings.autoSyncFromDesc') }}</div>
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
            <div class="config-label">{{ t('settings.manualSync') }}</div>
            <div class="config-desc">{{ t('settings.manualSyncDesc') }}</div>
          </div>
          <div class="config-control sync-btns">
            <button class="action-btn-sm" @click="syncUpload" :disabled="!cloudEnabled">{{ t('settings.upload') }}</button>
            <button class="action-btn-sm" @click="syncDownload" :disabled="!cloudEnabled">{{ t('settings.download') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- SSH 密钥管理 -->
    <div class="config-section">
      <div class="section-header">
        <div class="section-title">{{ t('settings.sshKeys') }}</div>
        <button class="section-action-btn" @click="loadSSHKeys">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"/>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
          </svg>
          {{ t('settings.refresh') }}
        </button>
      </div>

      <div class="config-list">
        <!-- 密钥列表 -->
        <div v-if="sshKeys.length > 0" class="key-grid">
          <div v-for="key in sshKeys" :key="key.name" class="key-card">
            <div class="key-card-header">
              <div class="key-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
                </svg>
              </div>
              <div class="key-card-title">{{ key.name }}</div>
              <div class="key-card-actions">
                <button class="icon-btn-xs" @click="showPublicKey(key)" :title="t('settings.copyPublicKey')">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                  </svg>
                </button>
                <button class="icon-btn-xs danger" @click="deleteKey(key)" :title="t('settings.deleteKey')">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"/>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                  </svg>
                </button>
              </div>
            </div>
            <div class="key-card-body">
              <div class="key-tag" :class="key.type">{{ key.type }}</div>
              <span v-if="key.bits" class="key-bits">{{ key.bits }} bit</span>
              <span v-if="key.comment" class="key-comment">{{ key.comment }}</span>
            </div>
            <div v-if="key.createdAt" class="key-card-footer">
              {{ key.createdAt }}
            </div>
          </div>
        </div>
        <div v-else class="empty-state-card">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.4">
            <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
          </svg>
          <span>{{ t('settings.noKeys') }}</span>
          <span class="empty-hint">{{ t('settings.noKeysHint') }}</span>
        </div>

        <!-- 生成新密钥 -->
        <div class="gen-section">
          <div class="gen-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
            </svg>
            {{ t('settings.generateNewKey') }}
          </div>
          <div class="gen-form">
            <div class="gen-row">
              <div class="gen-field">
                <label>{{ t('settings.keyName') }}</label>
                <input v-model="newKey.name" type="text" :placeholder="t('settings.keyNamePlaceholder')" class="gen-input" />
              </div>
              <div class="gen-field">
                <label>{{ t('settings.keyType') }}</label>
                <select v-model="newKey.type" class="gen-input">
                  <option value="ed25519">Ed25519 ({{ t('common.recommended') }})</option>
                  <option value="rsa">RSA (4096)</option>
                </select>
              </div>
              <div class="gen-field">
                <label>{{ t('settings.keyComment') }}</label>
                <input v-model="newKey.comment" type="text" :placeholder="t('settings.keyCommentPlaceholder')" class="gen-input" />
              </div>
              <button class="gen-btn" @click="generateKey" :disabled="generatingKey">
                {{ generatingKey ? t('settings.generating') : t('settings.generate') }}
              </button>
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
        {{ t('settings.export') }}
      </button>
      <button class="action-btn" @click="showImport = true">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        {{ t('settings.import') }}
      </button>
    </div>

    <!-- 赞助开发者 -->
    <div class="config-section">
      <div class="section-title">{{ t('settings.sponsor') }}</div>
      <div class="sponsor-card">
        <div class="sponsor-info">
          <p class="sponsor-text">{{ t('settings.sponsorText') }}</p>
          <p class="sponsor-hint">{{ t('settings.sponsorHint') }}</p>
        </div>
        <div class="sponsor-qr">
          <img src="/sponsor-qr.png" :alt="t('settings.sponsor')" class="qr-image" @error="qrError = true" />
          <div v-if="qrError" class="qr-placeholder">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3">
              <rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/>
              <rect x="14" y="14" width="3" height="3"/><rect x="18" y="14" width="3" height="3"/><rect x="14" y="18" width="3" height="3"/>
              <rect x="18" y="18" width="3" height="3"/>
            </svg>
            <span>{{ t('settings.qrPlaceholder') }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 导入对话框 -->
    <Modal
      v-model:visible="showImport"
      :title="t('settings.importConfig')"
      :cancel-text="t('common.cancel')"
      @confirm="handleImport"
    >
      <template #content>
        <textarea v-model="importJson" class="import-textarea" :placeholder="t('settings.pasteConfig')" rows="10"></textarea>
      </template>
    </Modal>

    <!-- 消息提示 -->
    <Message ref="messageRef" />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useConfigStore } from '../../stores/config'
import { Events } from '@wailsio/runtime'
import * as SSHService from '../../../bindings/changeme/ssh/sshservice.js'
import * as CloudService from '../../../bindings/changeme/ssh/cloudservice.js'
import Modal from '../../components/Modal.vue'
import Message from '../../components/Message.vue'

const { locale, t } = useI18n()
const router = useRouter()
const configStore = useConfigStore()
const messageRef = ref(null)

const currentLocale = ref(locale.value)
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
const qrError = ref(false)
const unsubCloud = ref(null)

// 云端配置
const cloudEnabled = ref(false)
const cloudHost = ref('')
const cloudPort = ref(9527)
const cloudToken = ref('')
const cloudAutoSyncTo = ref(false)
const cloudAutoSyncFrom = ref(false)
const cloudStatusText = ref(t('settings.disconnected'))
const shortcutsEnabled = ref(true)
const shortcutSwitchTab = ref(true)
const shortcutSaveGroup = ref(true)
const shortcutCloudUpload = ref(true)
const shortcutCloudDownload = ref(true)
const groupBehavior = ref('prompt')

// 语言切换
const handleLocaleChange = async (e) => {
  const newLocale = e.target.value
  locale.value = newLocale
  currentLocale.value = newLocale
  localStorage.setItem('locale', newLocale)
  // 保存到后端配置
  try {
    await configStore.set('locale', 'value', newLocale)
  } catch {}
  messageRef.value?.warning(t('settings.languageRestartRequired'))
}

// SSH 密钥管理
const sshKeys = ref([])
const newKey = ref({ name: '', type: 'ed25519', comment: '' })
const generatingKey = ref(false)

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
  shortcutsEnabled.value = configStore.get('shortcuts', 'enabled') !== false
  shortcutSwitchTab.value = configStore.get('shortcuts', 'switchTab') !== false
  shortcutSaveGroup.value = configStore.get('shortcuts', 'saveGroup') !== false
  shortcutCloudUpload.value = configStore.get('shortcuts', 'cloudUpload') !== false
  shortcutCloudDownload.value = configStore.get('shortcuts', 'cloudDownload') !== false
  groupBehavior.value = configStore.get('advanced', 'groupBehavior') || 'prompt'
  console.log('[Settings] shortcutsEnabled:', shortcutsEnabled.value, 'groupBehavior:', groupBehavior.value)

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
    cloudStatusText.value = connected ? t('settings.connected') : t('settings.disconnected')
  } catch {
    cloudStatusText.value = t('settings.disconnected')
  }

  // 监听云端状态变化
  unsubCloud.value = Events.On('cloud:status', (e) => {
    const d = e?.data
    if (d) {
      cloudStatusText.value = d.connected ? t('settings.connected') : t('settings.disconnected')
    }
  })

  // 加载 SSH 密钥
  loadSSHKeys()
})

onUnmounted(() => {
  if (unsubCloud.value) unsubCloud.value()
})

async function handleThemeChange(e) {
  theme.value = e.target.value
  await configStore.setTheme(e.target.value)
  showToast(t('settings.themeSwitched'))
}

async function handleChange(e) {
  defaultType.value = e.target.value
  await configStore.set('terminal', 'defaultType', e.target.value)
  showToast(t('settings.settingsSaved'))
}

async function handleSwitchModeChange(e) {
  switchMode.value = e.target.value
  await configStore.set('terminal', 'switchMode', e.target.value)
  showToast(t('settings.settingsSaved'))
}

async function changeFontSize(delta) {
  const newSize = Math.min(24, Math.max(10, fontSize.value + delta))
  if (newSize !== fontSize.value) {
    fontSize.value = newSize
    await configStore.set('terminal', 'fontSize', newSize)
    showToast(t('settings.settingsSaved'))
  }
}

async function handleCommandSendModeChange(e) {
  commandSendMode.value = e.target.value
  await configStore.set('terminal', 'commandSendMode', e.target.value)
  showToast(t('settings.settingsSaved'))
}

async function handleCodeHighlightChange(e) {
  codeHighlight.value = e.target.checked
  await configStore.set('terminal', 'codeHighlight', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleShortcutsChange(e) {
  shortcutsEnabled.value = e.target.checked
  await configStore.set('shortcuts', 'enabled', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleShortcutSwitchTabChange(e) {
  shortcutSwitchTab.value = e.target.checked
  await configStore.set('shortcuts', 'switchTab', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleShortcutSaveGroupChange(e) {
  shortcutSaveGroup.value = e.target.checked
  await configStore.set('shortcuts', 'saveGroup', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleShortcutCloudUploadChange(e) {
  shortcutCloudUpload.value = e.target.checked
  await configStore.set('shortcuts', 'cloudUpload', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleShortcutCloudDownloadChange(e) {
  shortcutCloudDownload.value = e.target.checked
  await configStore.set('shortcuts', 'cloudDownload', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleGroupBehaviorChange(e) {
  groupBehavior.value = e.target.value
  await configStore.set('advanced', 'groupBehavior', e.target.value)
  showToast(t('settings.settingsSaved'))
}

async function handleAutoTrayChange(e) {
  autoTray.value = e.target.checked
  await configStore.set('ui', 'autoTray', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleRememberPositionChange(e) {
  rememberPosition.value = e.target.checked
  await configStore.set('ui', 'rememberPosition', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleAutoShowHomeChange(e) {
  autoShowHome.value = e.target.checked
  await configStore.set('ui', 'autoShowHome', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function clearWindowPositions() {
  try {
    await SSHService.ClearWindowPositions()
    showToast(t('settings.windowPositionsCleared'))
  } catch (e) {
    showToast(t('settings.clearFailed'))
  }
}

async function handleCloudEnabledChange(e) {
  cloudEnabled.value = e.target.checked
  await configStore.set('cloud', 'enabled', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleAutoSyncToChange(e) {
  cloudAutoSyncTo.value = e.target.checked
  await configStore.set('cloud', 'autoSyncTo', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleAutoSyncFromChange(e) {
  cloudAutoSyncFrom.value = e.target.checked
  await configStore.set('cloud', 'autoSyncFrom', e.target.checked)
  showToast(t('settings.settingsSaved'))
}

async function handleCloudConfigChange() {
  await configStore.set('cloud', 'serverUrl', `${cloudHost.value}:${cloudPort.value || 9527}`)
  await configStore.set('cloud', 'token', cloudToken.value)
}

async function testCloudConnection() {
  if (!cloudHost.value || !cloudToken.value) {
    showToast(t('settings.pleaseFillAddress'))
    return
  }
  cloudStatusText.value = t('settings.connecting')
  const addr = `${cloudHost.value}:${cloudPort.value || 9527}`
  console.log('[Settings] 测试连接:', addr)
  const ok = await CloudService.Connect(addr, cloudToken.value)
  if (ok) {
    cloudStatusText.value = t('settings.connected')
    showToast(t('settings.connectionSuccess'))
    await CloudService.Disconnect()
  } else {
    cloudStatusText.value = t('settings.connectionFailed')
    showToast(t('settings.connectionFailed'))
  }
}

async function ensureCloudConnected() {
  const connected = await CloudService.IsConnected()
  if (connected) return true
  if (!cloudHost.value || !cloudToken.value) {
    showToast(t('settings.pleaseFillServerAddress'))
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
    showToast(t('settings.uploadSuccess', { count: merged.size }))
  } catch (e) {
    showToast(t('settings.uploadFailed') + ': ' + (e?.message || e))
  }
}

async function syncDownload() {
  try {
    if (!await ensureCloudConnected()) return
    const conns = await CloudService.PullSync()
    if (!conns || conns.length === 0) {
      showToast(t('settings.noCloudConfig'))
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
    showToast(t('settings.syncComplete', { count: count }))
  } catch (e) {
    showToast(t('settings.downloadFailed') + ': ' + (e?.message || e))
  }
}

function showToast(msg) {
  messageRef.value?.success(msg)
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  // 只显示时间部分（HH:MM:SS）
  const parts = timestamp.split(' ')
  return parts.length > 1 ? parts[1] : timestamp
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

// ========== SSH 密钥管理 ==========

async function loadSSHKeys() {
  try {
    const keys = await SSHService.ListSSHKeys()
    sshKeys.value = keys || []
  } catch (e) {
    console.error('[Settings] 加载 SSH 密钥失败:', e)
    sshKeys.value = []
  }
}

async function generateKey() {
  if (!newKey.value.name) {
    showToast(t('settings.pleaseFillKeyName'))
    return
  }
  generatingKey.value = true
  try {
    await SSHService.GenerateSSHKey(newKey.value.name, newKey.value.type, newKey.value.comment, 0)
    showToast(t('settings.keyGenerated'))
    newKey.value = { name: '', type: 'ed25519', comment: '' }
    await loadSSHKeys()
  } catch (e) {
    showToast(t('settings.keyGenerateFailed') + ': ' + (e?.message || e))
  } finally {
    generatingKey.value = false
  }
}

async function showPublicKey(key) {
  try {
    const pubKey = await SSHService.GetSSHPublicKey(key.name)
    // 复制到剪贴板
    await navigator.clipboard.writeText(pubKey)
    showToast(t('settings.publicKeyCopied'))
  } catch (e) {
    showToast(t('settings.getPublicKeyFailed') + ': ' + (e?.message || e))
  }
}

async function deleteKey(key) {
  if (!confirm(t('settings.deleteKeyConfirm', { name: key.name }))) return
  try {
    await SSHService.DeleteSSHKey(key.name)
    showToast(t('settings.keyDeleted'))
    await loadSSHKeys()
  } catch (e) {
    showToast(t('settings.keyDeleteFailed') + ': ' + (e?.message || e))
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

/* Section Header */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-actions {
  display: flex;
  gap: 8px;
}

.section-action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  background: var(--bg-panel);
  border: 1px solid var(--surface-hover);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.section-action-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}

/* SSH 密钥管理 - 卡片式 */
.key-grid {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.key-card {
  background: var(--bg-panel);
  border: 1px solid var(--surface-hover);
  border-radius: 10px;
  overflow: hidden;
  transition: all 0.2s;
}

.key-card:hover {
  border-color: var(--border-strong);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.key-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px 8px;
}

.key-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--accent-primary-bg);
  border-radius: 8px;
  color: var(--accent-primary);
}

.key-card-title {
  flex: 1;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.key-card-actions {
  display: flex;
  gap: 4px;
}

.icon-btn-xs {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn-xs:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.icon-btn-xs.danger:hover {
  background: var(--accent-danger-bg);
  color: var(--accent-danger);
}

.key-card-body {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 14px 10px;
  flex-wrap: wrap;
}

.key-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
}

.key-tag.ed25519 {
  background: #10b98120;
  color: #10b981;
}

.key-tag.rsa {
  background: #6366f120;
  color: #6366f1;
}

.key-bits {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.key-comment {
  font-size: 0.75rem;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.key-card-footer {
  padding: 8px 14px;
  border-top: 1px solid var(--surface-hover);
  font-size: 0.6875rem;
  color: var(--text-disabled);
}

/* 空状态卡片 */
.empty-state-card {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 32px 20px;
  background: var(--bg-panel);
  border: 1px dashed var(--surface-hover);
  border-radius: 10px;
  color: var(--text-muted);
  font-size: 0.875rem;
}

.empty-hint {
  font-size: 0.75rem;
  color: var(--text-disabled);
}

/* 生成密钥表单 */
.gen-section {
  grid-column: 1 / -1;
  background: var(--bg-panel);
  border: 1px solid var(--surface-hover);
  border-radius: 10px;
  padding: 14px;
}

.gen-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.gen-form {
  width: 100%;
}

.gen-row {
  display: flex;
  gap: 10px;
  align-items: flex-end;
}

.gen-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.gen-field label {
  font-size: 0.6875rem;
  color: var(--text-muted);
  font-weight: 500;
}

.gen-input {
  padding: 8px 10px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 0.8125rem;
  outline: none;
  transition: border-color 0.2s;
}

.gen-input:focus {
  border-color: var(--accent-primary);
}

.gen-btn {
  padding: 8px 20px;
  background: var(--accent-primary);
  border: none;
  border-radius: 6px;
  color: white;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.gen-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.gen-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.danger {
  color: var(--accent-danger) !important;
}

.danger:hover {
  background: var(--accent-danger-bg) !important;
}

/* 赞助开发者 */
.sponsor-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: 10px;
}

.sponsor-info {
  flex: 1;
}

.sponsor-text {
  font-size: 13px;
  color: var(--text-primary);
  margin: 0 0 4px;
  line-height: 1.5;
}

.sponsor-hint {
  font-size: 12px;
  color: var(--text-muted);
  margin: 0;
}

.sponsor-qr {
  flex-shrink: 0;
  width: 120px;
  height: 120px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--border-subtle);
}

.qr-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #fff;
}

.qr-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: var(--surface-2);
  color: var(--text-muted);
  font-size: 11px;
}
</style>
