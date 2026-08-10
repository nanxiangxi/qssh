<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { SSHService, SSHConfig } from "../../../bindings/changeme/ssh"
import { useSSHConnectionsStore } from '../../stores/sshConnections'
import { useConfigStore } from '../../stores/config'
import { Events, Dialogs } from '@wailsio/runtime'
import ConnectionOptionsDialog from './ConnectionOptionsDialog.vue'
import Message from '../../components/Message.vue'

const { t } = useI18n()

// 消息提示
const messageRef = ref<InstanceType<typeof Message> | null>(null)

// 对话框状态
const showDialog = ref(false)
const pendingConfig = ref<any>(null)

// 解析后端返回的错误对象，提取真正的错误消息
const extractErrorMessage = (error: any): string => {
  if (!error) return t('common.error')
  
  // 如果 error 是字符串，直接返回
  if (typeof error === 'string') return error
  
  // 尝试从不同位置提取消息
  // 1. 优先使用 cause.Message（后端 FriendlyError 的结构）
  if (error.cause && error.cause.Message) {
    return error.cause.Message
  }
  
  // 2. 其次使用 message 字段
  if (error.message) {
    return error.message
  }
  
  // 3. 最后尝试 Message 字段
  if (error.Message) {
    return error.Message
  }
  
  // 4. 兜底：转换为字符串
  return String(error)
}

// 表单数据
const formData = reactive({
  name: '',
  host: '',
  port: 22,
  username: '',
  authType: 'password', // password 或 key
  password: '',
  privateKey: '',
  timeout: 30,
  // 跳板机配置
  useJumpHost: false,
  jumpHost: '',
  jumpPort: 22,
  jumpUsername: '',
  jumpAuthType: 'password', // password 或 key
  jumpPassword: '',
  jumpPrivateKey: '',
})

// 加载状态
const loading = ref(false)

// 验证表单
const validateForm = () => {
  if (!formData.name.trim()) {
    messageRef.value?.error(t('connection.pleaseEnterName'))
    return false
  }
  if (!formData.host.trim()) {
    messageRef.value?.error(t('connection.pleaseEnterHost'))
    return false
  }
  if (formData.port < 1 || formData.port > 65535) {
    messageRef.value?.error(t('connection.invalidPort'))
    return false
  }
  if (!formData.username.trim()) {
    messageRef.value?.error(t('connection.pleaseEnterUsername'))
    return false
  }
  if (formData.authType === 'password' && !formData.password) {
    messageRef.value?.error(t('connection.pleaseEnterPassword'))
    return false
  }
  if (formData.authType === 'key' && !formData.privateKey.trim()) {
    messageRef.value?.error(t('connection.pleaseEnterPrivateKey'))
    return false
  }
  return true
}

// 测试连接
const testConnection = async () => {
  if (!validateForm()) return

  loading.value = true
  try {
    const config: any = {
      host: formData.host,
      port: formData.port,
      username: formData.username,
      password: formData.authType === 'password' ? formData.password : undefined,
      privateKey: formData.authType === 'key' ? formData.privateKey : undefined,
      timeout: formData.timeout
    }
    // 跳板机配置
    if (formData.useJumpHost && formData.jumpHost) {
      config.jumpHost = `${formData.jumpHost}:${formData.jumpPort}`
      config.jumpUsername = formData.jumpUsername || undefined
      config.jumpPassword = formData.jumpAuthType === 'password' ? (formData.jumpPassword || undefined) : undefined
      config.jumpKeyPath = formData.jumpAuthType === 'key' ? (formData.jumpPrivateKey || undefined) : undefined
    }
    await SSHService.TestConnection(new SSHConfig(config))

    messageRef.value?.success(t('connection.testSuccess'))
  } catch (error: any) {
    messageRef.value?.error(extractErrorMessage(error))
  } finally {
    loading.value = false
  }
}

// 立即连接
const saveAndConnect = async () => {
  console.log('[NewConnection] ========== 开始连接流程 ==========')

  if (!validateForm()) {
    console.log('[NewConnection] ❌ 表单验证失败')
    return
  }

  // 检查是否已经有SSH窗口打开
  const hasExistingWindow = await checkExistingSSHWindow()

  // 先保存配置
  const config: any = {
    name: formData.name || `${formData.username}@${formData.host}`,
    host: formData.host,
    port: formData.port,
    username: formData.username,
    password: formData.authType === 'password' ? formData.password : undefined,
    privateKey: formData.authType === 'key' ? formData.privateKey : undefined,
    timeout: formData.timeout
  }
  // 跳板机配置
  if (formData.useJumpHost && formData.jumpHost) {
    config.jumpHost = `${formData.jumpHost}:${formData.jumpPort}`
    config.jumpUsername = formData.jumpUsername || undefined
    config.jumpPassword = formData.jumpAuthType === 'password' ? (formData.jumpPassword || undefined) : undefined
    config.jumpKeyPath = formData.jumpAuthType === 'key' ? (formData.jumpPrivateKey || undefined) : undefined
  }
  pendingConfig.value = config

  if (!hasExistingWindow) {
    // 第一个连接，直接使用默认分组，不显示对话框
    console.log('[NewConnection] 🎯 第一个连接，直接使用默认分组')
    await handleDialogSelect('default-group')
  } else {
    // 已有窗口，根据配置决定行为
    const configStore = useConfigStore()
    await configStore.init()

    // 诊断日志：打印完整配置
    console.log('[NewConnection] 📋 完整配置对象:', JSON.stringify(configStore.config, null, 2))
    console.log('[NewConnection] 📋 advanced 分组:', configStore.config?.advanced)

    const groupBehavior = configStore.get('advanced', 'groupBehavior')
    console.log('[NewConnection] 📋 原始 groupBehavior 值:', groupBehavior, '类型:', typeof groupBehavior)

    const finalBehavior = groupBehavior || 'prompt'
    console.log('[NewConnection] 📋 最终分组行为:', finalBehavior)

    if (finalBehavior === 'join_default') {
      // 自动加入默认分组
      console.log('[NewConnection] 🎯 自动加入默认分组')
      await handleDialogSelect('default-group')
    } else if (finalBehavior === 'new_window') {
      // 自动打开新窗口
      console.log('[NewConnection] 🪟 自动打开新窗口')
      await handleDialogSelect('new-window')
    } else {
      // 弹出选择对话框（默认行为）
      console.log('[NewConnection] ✅ 显示选择对话框')
      showDialog.value = true
    }
  }
}

// 检查是否已有SSH窗口打开（检查实际在线的连接，而非分组中的旧ID）
const checkExistingSSHWindow = async (): Promise<boolean> => {
  try {
    const connections = await SSHService.GetAllConnections()
    const onlineCount = (connections || []).filter(c => c.status === 'connected').length
    return onlineCount > 0
  } catch (error) {
    return false
  }
}

// 处理对话框选择
const handleDialogSelect = async (type: string) => {
  showDialog.value = false
  
  if (type === 'cancel') {
    console.log('[NewConnection] ❌ 用户取消连接')
    pendingConfig.value = null
    return
  }
  
  loading.value = true
  
  try {
    const config = pendingConfig.value
    console.log('[NewConnection] 📝 连接配置:', JSON.stringify(config, null, 2))
    console.log('[NewConnection] 🎯 用户选择:', type)
    
    let groupID: string
    
    if (type === 'new-window') {
      // 创建新分组
      console.log('[NewConnection] 🪟 创建新分组...')
      groupID = await SSHService.CreateGroup(t('connection.temporaryGroup'))
      console.log('[NewConnection] ✅ 分组创建成功:', groupID)
    } else {
      // 使用默认分组
      console.log('[NewConnection] 📑 使用默认分组...')
      groupID = await SSHService.GetDefaultGroupID()
      console.log('[NewConnection] ✅ 默认分组ID:', groupID)
    }
    
    // 创建连接并指定分组
    console.log('[NewConnection] 🚀 调用 CreateAndConnectWithGroup...')
    const result = await SSHService.CreateAndConnectWithGroup(config, groupID)
    
    console.log('[NewConnection] ✅ 后端返回结果:', result)
    console.log('[NewConnection]    - connID:', result.connID)
    console.log('[NewConnection]    - groupID:', result.groupID)
    
    // 刷新连接列表
    console.log('[NewConnection] 🔄 刷新连接列表...')
    const connectionsStore = useSSHConnectionsStore()
    await connectionsStore.refresh()
    console.log('[NewConnection] ✅ 连接列表已刷新')
    
    // 重置表单
    resetForm()
    
    messageRef.value?.success(t('connection.connectSuccess'))
    
    // 打开SSH窗口（传递 connID 作为 activeConn）
    console.log('[NewConnection] 🪟 调用 OpenSSHWindow...')
    console.log('[NewConnection]    - groupID:', result.groupID)
    console.log('[NewConnection]    - groupName:', config.name)
    console.log('[NewConnection]    - activeConn:', result.connID)
    
    await SSHService.OpenSSHWindow(result.groupID!, config.name!, result.connID!)

    console.log('[NewConnection] 🪟 窗口操作完成')

    // 如果开启了自动托盘，连接成功后隐藏主窗口
    const configStore = useConfigStore()
    await configStore.init()
    if (configStore.get('ui', 'autoTray')) {
      Events.Emit('ssh:tray-hide')
    }
    
  } catch (error: any) {
    console.error('[NewConnection] ❌ 错误:', error)
    messageRef.value?.error(extractErrorMessage(error))
  } finally {
    loading.value = false
    pendingConfig.value = null
    console.log('[NewConnection] ========== 连接流程结束 ==========')
  }
}

// 处理对话框取消
const handleDialogCancel = () => {
  showDialog.value = false
  pendingConfig.value = null
  console.log('[NewConnection] ❌ 用户取消')
}

// 重置表单
const resetForm = () => {
  formData.name = ''
  formData.host = ''
  formData.port = 22
  formData.username = ''
  formData.authType = 'password'
  formData.password = ''
  formData.privateKey = ''
  formData.timeout = 30
  formData.useJumpHost = false
  formData.jumpHost = ''
  formData.jumpPort = 22
  formData.jumpUsername = ''
  formData.jumpAuthType = 'password'
  formData.jumpPassword = ''
  formData.jumpPrivateKey = ''
}
</script>

<template>
  <div class="new-connection-container">
    <!-- 全局消息提示 -->
    <Message ref="messageRef" />

    <!-- 连接选项对话框 -->
    <ConnectionOptionsDialog
      :visible="showDialog"
      @select="handleDialogSelect"
      @cancel="handleDialogCancel"
    />

    <!-- 上半部分：连接配置表单 -->
    <div class="form-section">
      <h2 class="section-title">{{ t('connection.newConnection') }}</h2>
      
      <div class="form-grid">
        <!-- 连接名称 -->
        <div class="form-group">
          <label for="name">{{ t('connection.name') }}</label>
          <input
            id="name"
            v-model="formData.name"
            type="text"
            :placeholder="t('connection.name')"
            class="form-input"
          />
        </div>

        <!-- 主机地址 -->
        <div class="form-group">
          <label for="host">{{ t('connection.host') }}</label>
          <input
            id="host"
            v-model="formData.host"
            type="text"
            placeholder="192.168.1.100"
            class="form-input"
          />
        </div>

        <!-- 端口 -->
        <div class="form-group">
          <label for="port">{{ t('connection.port') }}</label>
          <input
            id="port"
            v-model.number="formData.port"
            type="number"
            min="1"
            max="65535"
            class="form-input"
          />
        </div>

        <!-- 用户名 -->
        <div class="form-group">
          <label for="username">{{ t('connection.username') }}</label>
          <input
            id="username"
            v-model="formData.username"
            type="text"
            :placeholder="t('connection.username') + ' (root)'"
            class="form-input"
          />
        </div>

        <!-- 认证方式 -->
        <div class="form-group full-width">
          <label>{{ t('connection.authType') }}</label>
          <div class="auth-type-selector">
            <button
              type="button"
              class="auth-btn"
              :class="{ active: formData.authType === 'password' }"
              @click="formData.authType = 'password'"
            >
              {{ t('connection.passwordAuth') }}
            </button>
            <button
              type="button"
              class="auth-btn"
              :class="{ active: formData.authType === 'key' }"
              @click="formData.authType = 'key'"
            >
              {{ t('connection.keyAuth') }}
            </button>
          </div>
        </div>

        <!-- 密码输入 -->
        <div v-if="formData.authType === 'password'" class="form-group full-width">
          <label for="password">{{ t('connection.password') }}</label>
          <input
            id="password"
            v-model="formData.password"
            type="password"
            :placeholder="t('connection.passwordPlaceholder')"
            class="form-input"
          />
        </div>

        <!-- 密钥输入 -->
        <div v-if="formData.authType === 'key'" class="form-group full-width">
          <label for="privateKey">{{ t('connection.privateKey') }}</label>
          <textarea
            id="privateKey"
            v-model="formData.privateKey"
            :placeholder="t('connection.privateKeyPlaceholder')"
            class="form-textarea key-textarea"
            rows="6"
          ></textarea>
        </div>

        <!-- 超时时间 -->
        <div class="form-group">
          <label for="timeout">{{ t('connection.timeout') }} (s)</label>
          <input
            id="timeout"
            v-model.number="formData.timeout"
            type="number"
            min="5"
            max="300"
            class="form-input"
          />
        </div>

        <!-- 跳板机配置 -->
        <div class="form-group full-width">
          <div class="divider-toggle" @click="formData.useJumpHost = !formData.useJumpHost">
            <div class="divider-line"></div>
            <span class="divider-text" :class="{ active: formData.useJumpHost }">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M4 12h16m-8-8l8 8-8 8"/>
              </svg>
              {{ t('connection.jumpHost') }} (ProxyJump)
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ rotated: formData.useJumpHost }">
                <polyline points="6 9 12 15 18 9"/>
              </svg>
            </span>
            <div class="divider-line"></div>
          </div>
        </div>

        <template v-if="formData.useJumpHost">
          <!-- 跳板机主机信息 -->
          <div class="form-group">
            <label for="jumpHost">{{ t('connection.jumpHostAddr') }}</label>
            <input
              id="jumpHost"
              v-model="formData.jumpHost"
              type="text"
              placeholder="10.0.0.1"
              class="form-input"
            />
          </div>

          <div class="form-group">
            <label for="jumpPort">{{ t('connection.jumpHostPort') }}</label>
            <input
              id="jumpPort"
              v-model.number="formData.jumpPort"
              type="number"
              min="1"
              max="65535"
              class="form-input"
            />
          </div>

          <div class="form-group">
            <label for="jumpUsername">{{ t('connection.jumpHostUser') }}</label>
            <input
              id="jumpUsername"
              v-model="formData.jumpUsername"
              type="text"
              :placeholder="t('connection.jumpUsernamePlaceholder')"
              class="form-input"
            />
          </div>

          <!-- 跳板机认证方式 -->
          <div class="form-group full-width">
            <label>{{ t('connection.authType') }}</label>
            <div class="auth-type-selector">
              <button
                type="button"
                class="auth-btn"
                :class="{ active: formData.jumpAuthType === 'password' }"
                @click="formData.jumpAuthType = 'password'"
              >
                {{ t('connection.passwordAuth') }}
              </button>
              <button
                type="button"
                class="auth-btn"
                :class="{ active: formData.jumpAuthType === 'key' }"
                @click="formData.jumpAuthType = 'key'"
              >
                {{ t('connection.keyAuth') }}
              </button>
            </div>
          </div>

          <!-- 跳板机密码输入 -->
          <div v-if="formData.jumpAuthType === 'password'" class="form-group full-width">
            <label for="jumpPassword">{{ t('connection.jumpHostPassword') }}</label>
            <input
              id="jumpPassword"
              v-model="formData.jumpPassword"
              type="password"
              :placeholder="t('connection.jumpPasswordPlaceholder')"
              class="form-input"
            />
          </div>

          <!-- 跳板机密钥输入 -->
          <div v-if="formData.jumpAuthType === 'key'" class="form-group full-width">
            <label for="jumpPrivateKey">{{ t('connection.jumpPrivateKey') }}</label>
            <textarea
              id="jumpPrivateKey"
              v-model="formData.jumpPrivateKey"
              :placeholder="t('connection.jumpPrivateKeyPlaceholder')"
              class="form-textarea key-textarea"
              rows="4"
            ></textarea>
          </div>
        </template>
      </div>
    </div>

    <!-- 下半部分：操作按钮 -->
    <div class="action-section">
      <div class="button-group">
        <button
          type="button"
          class="btn btn-reset"
          @click="resetForm"
          :disabled="loading"
        >
          {{ t('common.reset') }}
        </button>
        <button
          type="button"
          class="btn btn-test"
          @click="testConnection"
          :disabled="loading"
        >
          {{ loading ? t('common.loading') : t('connection.test') }}
        </button>
        <button
          type="button"
          class="btn btn-connect"
          @click="saveAndConnect"
          :disabled="loading"
        >
          {{ loading ? t('common.connecting') : t('connection.connect') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.new-connection-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 1.5rem;
}

.form-section {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.section-title {
  color: var(--text-primary);
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 1.5rem 0;
  padding-bottom: 0.75rem;
  border-bottom: 0.0625rem solid var(--surface-hover);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-group label {
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
}

.form-input {
  padding: 0.625rem 0.875rem;
  background: var(--bg-panel);
  border: 0.0625rem solid var(--border-strong);
  border-radius: 0.5rem;
  color: var(--text-primary);
  font-size: 0.9375rem;
  transition: all 0.2s;
  outline: none;
}

.form-input:focus {
  border-color: var(--accent-primary);
  background: var(--bg-input-focus);
  box-shadow: 0 0 0 0.1875rem var(--primary-bg);
}

.form-input::placeholder {
  color: var(--text-muted);
}

.auth-type-selector {
  display: flex;
  gap: 0.75rem;
}

.auth-btn {
  flex: 1;
  padding: 0.625rem 1rem;
  background: var(--bg-panel);
  border: 0.0625rem solid var(--border-strong);
  border-radius: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.auth-btn:hover {
  background: var(--bg-toolbar);
  border-color: var(--scrollbar-thumb-hover);
  color: var(--text-primary);
}

.auth-btn.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: var(--text-on-accent);
  font-weight: 500;
}

.key-textarea {
  min-height: 140px;
  font-family: 'Courier New', Consolas, monospace;
  font-size: 0.75rem;
  line-height: 1.6;
  resize: vertical;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 0.5rem;
  padding: 0.75rem;
  color: var(--text-primary);
  white-space: pre;
  overflow-x: auto;
}

.key-textarea:focus {
  border-color: var(--border-accent);
  box-shadow: 0 0 0 2px var(--primary-bg);
}

.key-textarea::placeholder {
  color: var(--text-disabled);
}

.action-section {
  padding: 0.5rem;
}

.button-group {
  display: flex;
  justify-content: flex-end;
  gap: 0.875rem;
}

.btn {
  padding: 0.625rem 1.5rem;
  border-radius: 0.5rem;
  font-size: 0.9375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
  min-width: 6.25rem;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-reset {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.btn-reset:hover:not(:disabled) {
  background: var(--border-strong);
}

.btn-test {
  background: var(--accent-warning);
  color: var(--text-on-accent);
}

.btn-test:hover:not(:disabled) {
  background: var(--warning-light);
  transform: translateY(-0.0625rem);
  box-shadow: 0 0.25rem 0.75rem var(--warning-bg);
}

.btn-connect {
  background: var(--accent-success);
  color: var(--text-on-accent);
}

.btn-connect:hover:not(:disabled) {
  background: var(--success-light);
  transform: translateY(-0.0625rem);
  box-shadow: 0 0.25rem 0.75rem var(--success-bg);
}

/* 滚动条样式 */
.form-section::-webkit-scrollbar {
  width: 0.5rem;
}

.form-section::-webkit-scrollbar-track {
  background: var(--bg-input);
  border-radius: 0.25rem;
}

.form-section::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb);
  border-radius: 0.25rem;
}

.form-section::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover);
}

/* 分隔线切换器（跳板机） */
.divider-toggle {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  user-select: none;
}

.divider-line {
  flex: 1;
  height: 0.0625rem;
  background: var(--surface-hover);
}

.divider-text {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  color: var(--text-muted);
  white-space: nowrap;
  transition: color 0.2s;
  padding: 0.375rem 0.75rem;
  border-radius: 0.375rem;
}

.divider-text:hover {
  color: var(--text-secondary);
  background: var(--surface-hover);
}

.divider-text.active {
  color: var(--accent-primary);
}

.divider-text svg {
  transition: transform 0.2s;
}

.divider-text svg.rotated {
  transform: rotate(180deg);
}

</style>