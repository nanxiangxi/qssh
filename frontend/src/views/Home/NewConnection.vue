<script setup lang="ts">
import { ref, reactive } from 'vue'
import { SSHService } from "../../../bindings/changeme/ssh"
import { useSSHConnectionsStore } from '../../stores/sshConnections'
import { Dialogs } from '@wailsio/runtime'
import ConnectionOptionsDialog from './ConnectionOptionsDialog.vue'
import Message from '../../components/Message.vue'

// 消息提示
const messageRef = ref<InstanceType<typeof Message> | null>(null)

// 对话框状态
const showDialog = ref(false)
const pendingConfig = ref<any>(null)

// 解析后端返回的错误对象，提取真正的错误消息
const extractErrorMessage = (error: any): string => {
  if (!error) return '未知错误'
  
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
  timeout: 30
})

// 加载状态
const loading = ref(false)

// 验证表单
const validateForm = () => {
  if (!formData.name.trim()) {
    messageRef.value?.error('请输入连接名称')
    return false
  }
  if (!formData.host.trim()) {
    messageRef.value?.error('请输入主机地址')
    return false
  }
  if (formData.port < 1 || formData.port > 65535) {
    messageRef.value?.error('端口号必须在 1-65535 之间')
    return false
  }
  if (!formData.username.trim()) {
    messageRef.value?.error('请输入用户名')
    return false
  }
  if (formData.authType === 'password' && !formData.password) {
    messageRef.value?.error('请输入密码')
    return false
  }
  if (formData.authType === 'key' && !formData.privateKey.trim()) {
    messageRef.value?.error('请输入私钥内容')
    return false
  }
  return true
}

// 测试连接
const testConnection = async () => {
  if (!validateForm()) return

  loading.value = true
  try {
    // 直接传递配置对象给后端
    await SSHService.TestConnection({
      host: formData.host,
      port: formData.port,
      username: formData.username,
      password: formData.authType === 'password' ? formData.password : undefined,
      privateKey: formData.authType === 'key' ? formData.privateKey : undefined,
      timeout: formData.timeout
    })

    messageRef.value?.success('连接测试成功！')
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
  pendingConfig.value = {
    name: formData.name || `${formData.username}@${formData.host}`,
    host: formData.host,
    port: formData.port,
    username: formData.username,
    password: formData.authType === 'password' ? formData.password : undefined,
    privateKey: formData.authType === 'key' ? formData.privateKey : undefined,
    timeout: formData.timeout
  }
  
  if (!hasExistingWindow) {
    // 第一个连接，直接使用默认分组，不显示对话框
    console.log('[NewConnection] 🎯 第一个连接，直接使用默认分组')
    await handleDialogSelect('default-group')
  } else {
    // 已有窗口，显示选择对话框
    console.log('[NewConnection] ✅ 已有窗口，显示选择对话框')
    showDialog.value = true
  }
}

// 检查是否已有SSH窗口打开
const checkExistingSSHWindow = async (): Promise<boolean> => {
  try {
    // 获取所有分组
    const groups = await SSHService.GetAllGroups()
    console.log('[NewConnection] 🔍 当前分组列表:', groups)
    
    // 检查是否有非空分组（除了默认分组）
    for (const group of groups) {
      if (group.conn_ids && group.conn_ids.length > 0) {
        console.log('[NewConnection] ✅ 发现已有连接的分组:', group.id)
        return true
      }
    }
    
    console.log('[NewConnection] ❌ 没有已连接的分组')
    return false
  } catch (error) {
    console.error('[NewConnection] ⚠️ 检查窗口失败:', error)
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
      groupID = await SSHService.CreateGroup('临时分组')
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
    
    messageRef.value?.success('连接成功！正在打开窗口...')
    
    // 打开SSH窗口（传递 connID 作为 activeConn）
    console.log('[NewConnection] 🪟 调用 OpenSSHWindow...')
    console.log('[NewConnection]    - groupID:', result.groupID)
    console.log('[NewConnection]    - groupName:', config.name)
    console.log('[NewConnection]    - activeConn:', result.connID)
    
    await SSHService.OpenSSHWindow(result.groupID, config.name, result.connID)
    
    console.log('[NewConnection] 🪟 窗口操作完成')
    
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
      <h2 class="section-title">新建 SSH 连接</h2>
      
      <div class="form-grid">
        <!-- 连接名称 -->
        <div class="form-group">
          <label for="name">连接名称</label>
          <input
            id="name"
            v-model="formData.name"
            type="text"
            placeholder="例如：我的服务器"
            class="form-input"
          />
        </div>

        <!-- 主机地址 -->
        <div class="form-group">
          <label for="host">主机地址</label>
          <input
            id="host"
            v-model="formData.host"
            type="text"
            placeholder="例如：192.168.1.100 或 example.com"
            class="form-input"
          />
        </div>

        <!-- 端口 -->
        <div class="form-group">
          <label for="port">端口</label>
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
          <label for="username">用户名</label>
          <input
            id="username"
            v-model="formData.username"
            type="text"
            placeholder="例如：root"
            class="form-input"
          />
        </div>

        <!-- 认证方式 -->
        <div class="form-group full-width">
          <label>认证方式</label>
          <div class="auth-type-selector">
            <button
              type="button"
              class="auth-btn"
              :class="{ active: formData.authType === 'password' }"
              @click="formData.authType = 'password'"
            >
              密码认证
            </button>
            <button
              type="button"
              class="auth-btn"
              :class="{ active: formData.authType === 'key' }"
              @click="formData.authType = 'key'"
            >
              密钥认证
            </button>
          </div>
        </div>

        <!-- 密码输入 -->
        <div v-if="formData.authType === 'password'" class="form-group full-width">
          <label for="password">密码</label>
          <input
            id="password"
            v-model="formData.password"
            type="password"
            placeholder="输入SSH密码"
            class="form-input"
          />
        </div>

        <!-- 密钥输入 -->
        <div v-if="formData.authType === 'key'" class="form-group full-width">
          <label for="privateKey">私钥内容</label>
          <textarea
            id="privateKey"
            v-model="formData.privateKey"
            placeholder="粘贴 SSH 私钥内容（以 -----BEGIN 开头）"
            class="form-textarea key-textarea"
            rows="6"
          ></textarea>
        </div>

        <!-- 超时时间 -->
        <div class="form-group">
          <label for="timeout">超时时间（秒）</label>
          <input
            id="timeout"
            v-model.number="formData.timeout"
            type="number"
            min="5"
            max="300"
            class="form-input"
          />
        </div>
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
          重置
        </button>
        <button
          type="button"
          class="btn btn-test"
          @click="testConnection"
          :disabled="loading"
        >
          {{ loading ? '测试中...' : '测试连接' }}
        </button>
        <button
          type="button"
          class="btn btn-connect"
          @click="saveAndConnect"
          :disabled="loading"
        >
          {{ loading ? '连接中...' : '立即连接' }}
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
  color: #e2e8f0;
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 1.5rem 0;
  padding-bottom: 0.75rem;
  border-bottom: 0.0625rem solid rgba(255, 255, 255, 0.1);
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
  color: #a0aec0;
  font-size: 0.875rem;
  font-weight: 500;
}

.form-input {
  padding: 0.625rem 0.875rem;
  background: rgba(30, 30, 30, 0.8);
  border: 0.0625rem solid rgba(255, 255, 255, 0.15);
  border-radius: 0.5rem;
  color: #e2e8f0;
  font-size: 0.9375rem;
  transition: all 0.2s;
  outline: none;
}

.form-input:focus {
  border-color: #4299e1;
  background: rgba(30, 30, 30, 1);
  box-shadow: 0 0 0 0.1875rem rgba(66, 153, 225, 0.2);
}

.form-input::placeholder {
  color: #718096;
}

.auth-type-selector {
  display: flex;
  gap: 0.75rem;
}

.auth-btn {
  flex: 1;
  padding: 0.625rem 1rem;
  background: rgba(30, 30, 30, 0.8);
  border: 0.0625rem solid rgba(255, 255, 255, 0.15);
  border-radius: 0.5rem;
  color: #a0aec0;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.auth-btn:hover {
  background: rgba(40, 40, 40, 0.9);
  border-color: rgba(255, 255, 255, 0.25);
  color: #e2e8f0;
}

.auth-btn.active {
  background: linear-gradient(135deg, #4299e1, #3182ce);
  border-color: #4299e1;
  color: white;
  font-weight: 500;
}

.key-textarea {
  min-height: 140px;
  font-family: 'Courier New', Consolas, monospace;
  font-size: 0.75rem;
  line-height: 1.6;
  resize: vertical;
  background: rgba(20, 20, 20, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  padding: 0.75rem;
  color: #e2e8f0;
  white-space: pre;
  overflow-x: auto;
}

.key-textarea:focus {
  border-color: rgba(66, 153, 225, 0.5);
  box-shadow: 0 0 0 2px rgba(66, 153, 225, 0.1);
}

.key-textarea::placeholder {
  color: #4a5568;
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
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

.btn-reset:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.15);
}

.btn-test {
  background: linear-gradient(135deg, #ed8936, #dd6b20);
  color: white;
}

.btn-test:hover:not(:disabled) {
  background: linear-gradient(135deg, #f09546, #e07830);
  transform: translateY(-0.0625rem);
  box-shadow: 0 0.25rem 0.75rem rgba(237, 137, 54, 0.3);
}

.btn-connect {
  background: linear-gradient(135deg, #48bb78, #38a169);
  color: white;
}

.btn-connect:hover:not(:disabled) {
  background: linear-gradient(135deg, #58c688, #48b179);
  transform: translateY(-0.0625rem);
  box-shadow: 0 0.25rem 0.75rem rgba(72, 187, 120, 0.3);
}

/* 滚动条样式 */
.form-section::-webkit-scrollbar {
  width: 0.5rem;
}

.form-section::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 0.25rem;
}

.form-section::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 0.25rem;
}

.form-section::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

</style>