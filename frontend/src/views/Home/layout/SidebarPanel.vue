<template>
  <div class="sidebar-panel">
    <!-- 搜索框 -->
    <div class="search-box">
      <svg class="search-icon" width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
        <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z"/>
      </svg>
      <input 
        v-model="searchQuery" 
        type="text" 
        placeholder="搜索..." 
        class="search-input"
      />
    </div>
    
    <!-- 连接列表 -->
    <div class="connection-list">
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <span>加载中...</span>
      </div>
      
      <div v-else-if="savedConnections.length === 0 && cachedConnections.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 16 16" fill="currentColor" opacity="0.3">
          <path d="M8 1a2 2 0 0 1 2 2v4H6V3a2 2 0 0 1 2-2zm3 6V3a3 3 0 0 0-6 0v4a2 2 0 0 0-2 2v5a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"/>
        </svg>
        <p>暂无连接</p>
      </div>

      <template v-else>
        <!-- 已保存的连接 -->
        <div v-if="savedConnections.length > 0" class="connection-section">
          <div class="section-header">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#68d391" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
            <span class="section-title">已保存</span>
            <span class="section-count">{{ savedConnections.length }}</span>
          </div>
          <div class="connections">
            <div
              v-for="conn in savedConnections"
              :key="conn.id"
              class="connection-item"
              :class="{ active: conn.id === activeConnectionId, connected: conn.status === 'connected' }"
              @click="selectConnection(conn)"
              @dblclick="quickConnect(conn)"
            >
              <div class="connection-content">
                <div class="connection-top">
                  <div class="connection-name">{{ conn.name }}</div>
                  <button class="more-btn" @click.stop="toggleMenu($event, conn.id)" title="更多操作">
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor"><path d="M3 9.5a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3z"/></svg>
                  </button>
                </div>
                <div class="connection-bottom">
                  <div class="connection-info">
                    <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor" class="address-icon"><path d="M8.051 1.999h.089c.822.003 4.987.033 6.11.335a2.01 2.01 0 0 1 1.415 1.42c.101.38.172.883.22 1.402l.01.104.022.26.008.104c.065.914.073 1.77.074 1.957v.075c-.001.194-.01 1.108-.082 2.06l-.008.105-.009.104c-.05.572-.124 1.14-.235 1.558a2.007 2.007 0 0 1-1.415 1.42c-1.16.312-5.569.334-6.18.335h-.142c-.309 0-1.587-.006-2.927-.052l-.17-.006-.087-.004-.171-.007-.171-.007c-1.11-.049-2.167-.128-2.654-.26a2.007 2.007 0 0 1-1.415-1.419c-.111-.417-.185-.986-.235-1.558L.09 9.82l-.008-.104A31.4 31.4 0 0 1 0 7.68v-.123c.002-.215.01-.958.064-1.778l.007-.103.003-.052.008-.104.022-.26.01-.104c.048-.519.119-1.023.22-1.402a2.007 2.007 0 0 1 1.415-1.42c.487-.13 1.544-.21 2.654-.26l.17-.007.172-.006.086-.003.171-.007A99.788 99.788 0 0 1 7.858 2h.193zM6.4 5.209v4.818l4.157-2.408L6.4 5.209z"/></svg>
                    <span>{{ conn.host }}:{{ conn.port }}</span>
                  </div>
                  <div class="status-badge" :class="conn.status">
                    {{ conn.status === 'connected' ? '已连接' : '未连接' }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 缓存的连接 -->
        <div v-if="cachedConnections.length > 0" class="connection-section">
          <div class="section-header">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#718096" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            <span class="section-title">缓存</span>
            <span class="section-count">{{ cachedConnections.length }}</span>
          </div>
          <div class="connections">
            <div
              v-for="conn in cachedConnections"
              :key="conn.id"
              class="connection-item cached"
              :class="{ active: conn.id === activeConnectionId, connected: conn.status === 'connected' }"
              @click="selectConnection(conn)"
              @dblclick="quickConnect(conn)"
            >
              <div class="connection-content">
                <div class="connection-top">
                  <div class="connection-name">{{ conn.name }}</div>
                  <button class="more-btn" @click.stop="toggleMenu($event, conn.id)" title="更多操作">
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor"><path d="M3 9.5a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3z"/></svg>
                  </button>
                </div>
                <div class="connection-bottom">
                  <div class="connection-info">
                    <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor" class="address-icon"><path d="M8.051 1.999h.089c.822.003 4.987.033 6.11.335a2.01 2.01 0 0 1 1.415 1.42c.101.38.172.883.22 1.402l.01.104.022.26.008.104c.065.914.073 1.77.074 1.957v.075c-.001.194-.01 1.108-.082 2.06l-.008.105-.009.104c-.05.572-.124 1.14-.235 1.558a2.007 2.007 0 0 1-1.415 1.42c-1.16.312-5.569.334-6.18.335h-.142c-.309 0-1.587-.006-2.927-.052l-.17-.006-.087-.004-.171-.007-.171-.007c-1.11-.049-2.167-.128-2.654-.26a2.007 2.007 0 0 1-1.415-1.419c-.111-.417-.185-.986-.235-1.558L.09 9.82l-.008-.104A31.4 31.4 0 0 1 0 7.68v-.123c.002-.215.01-.958.064-1.778l.007-.103.003-.052.008-.104.022-.26.01-.104c.048-.519.119-1.023.22-1.402a2.007 2.007 0 0 1 1.415-1.42c.487-.13 1.544-.21 2.654-.26l.17-.007.172-.006.086-.003.171-.007A99.788 99.788 0 0 1 7.858 2h.193zM6.4 5.209v4.818l4.157-2.408L6.4 5.209z"/></svg>
                    <span>{{ conn.host }}:{{ conn.port }}</span>
                  </div>
                  <div class="status-badge" :class="conn.status">
                    {{ conn.status === 'connected' ? '已连接' : '未连接' }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
    
    <!-- 底部工具栏 -->
    <div class="bottom-toolbar">
      <button class="icon-btn" @click="toggleSort" :title="sortBy === 'name' ? '按名称排序' : '按状态排序'">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
          <path d="M4 10.5a.5.5 0 0 0 .5.5h5a.5.5 0 0 0 0-1h-5a.5.5 0 0 0-.5.5zm0-3a.5.5 0 0 0 .5.5h7a.5.5 0 0 0 0-1h-7a.5.5 0 0 0-.5.5zm0-3a.5.5 0 0 0 .5.5h9a.5.5 0 0 0 0-1h-9a.5.5 0 0 0-.5.5z"/>
        </svg>
      </button>
    </div>
  </div>

  <!-- 全局固定定位菜单，Teleport 到 body 避免被任何父级裁剪/遮挡 -->
  <Teleport to="body">
    <div
      v-if="activeMenu && menuConn"
      class="custom-menu-overlay"
      @mousedown.prevent="closeMenu"
    >
      <div
        class="custom-menu"
        :style="{ top: menuPos.top + 'px', left: menuPos.left + 'px' }"
        @mousedown.stop
      >
        <div v-if="menuConn.status !== 'connected'" class="menu-item menu-item-primary" @click.stop="quickConnect(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M11.596 8.697l-6.363 3.692c-.54.313-1.233-.066-1.233-.697V4.308c0-.63.692-1.01 1.233-.696l6.363 3.692a.802.802 0 0 1 0 1.393z"/>
          </svg>
          <span>快速连接</span>
        </div>
        <div v-if="menuConn.status !== 'connected'" class="menu-item" @click.stop="editConnection(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M12.146.146a.5.5 0 0 1 .708 0l3 3a.5.5 0 0 1 0 .708l-10 10a.5.5 0 0 1-.168.11l-5 2a.5.5 0 0 1-.65-.65l2-5a.5.5 0 0 1 .11-.168l10-10zM11.207 2.5L13.5 4.793 14.793 3.5 12.5 1.207 11.207 2.5zm1.586 3L10.5 3.207 4 9.707V10h.5a.5.5 0 0 1 .5.5v.5h.5a.5.5 0 0 1 .5.5v.5h.293l6.5-6.5zm-9.761 5.175l-.106.106-1.528 3.821 3.821-1.528.106-.106A.5.5 0 0 1 13 15.5V15a.5.5 0 0 0-.5-.5H13a.5.5 0 0 1-.5-.5V14a.5.5 0 0 0-.5-.5h-.5a.5.5 0 0 1-.5-.5v-.5a.5.5 0 0 0-.5-.5h-.5a.5.5 0 0 1-.5-.5v-.5a.5.5 0 0 0-.5-.5h-.293z"/>
          </svg>
          <span>编辑</span>
        </div>
        <div class="menu-sep"></div>
        <div v-if="menuConn.saved" class="menu-item" @click.stop="deleteSaved(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M5.5 5.5A.5.5 0 0 1 6 6v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm2.5 0a.5.5 0 0 1 .5.5v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm3 .5a.5.5 0 0 0-1 0v6a.5.5 0 0 0 1 0V6z"/>
            <path fill-rule="evenodd" d="M14.5 3a1 1 0 0 1-1 1H13v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4h-.5a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1H6a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1h3.5a1 1 0 0 1 1 1v1zM4.118 4 4 4.059V13a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V4.059L11.882 4H4.118zM2.5 3V2h11v1h-11z"/>
          </svg>
          <span>删除本地存储</span>
        </div>
        <div v-if="menuConn.status === 'connected'" class="menu-item" @click.stop="disconnectConnection(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M11.251.068a.5.5 0 0 1 .227.58L9.677 6.5H13a.5.5 0 0 1 .364.843l-8 8.5a.5.5 0 0 1-.864-.5L6.323 9.5H3a.5.5 0 0 1-.364-.843l8-8.5a.5.5 0 0 1 .615-.09z"/>
          </svg>
          <span>关闭连接</span>
        </div>
        <div v-if="!menuConn.saved" class="menu-item" @click.stop="saveConnection(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M2 2a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V2zm10-1H4a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V2a1 1 0 0 0-1-1z"/>
            <path d="M10.854 5.146a.5.5 0 0 1 0 .708l-3 3a.5.5 0 0 1-.708 0l-1.5-1.5a.5.5 0 1 1 .708-.708L7.5 7.793l2.646-2.647a.5.5 0 0 1 .708 0z"/>
          </svg>
          <span>保存到本地</span>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- 连接方式选择弹窗 -->
  <ConnectionOptionsDialog
    :visible="showConnDialog"
    @select="handleConnDialogSelect"
    @cancel="handleConnDialogCancel"
  />

  <!-- 编辑连接弹窗 -->
  <Teleport to="body">
    <div v-if="showEditDialog" class="edit-modal-mask" @click.self="showEditDialog = false">
      <div class="edit-modal">
        <div class="edit-modal-head">
          <h3>编辑连接</h3>
          <button @click="showEditDialog = false">&times;</button>
        </div>
        <div class="edit-modal-body">
          <div class="edit-field">
            <label>名称</label>
            <input v-model="editForm.name" class="edit-input" placeholder="连接名称" />
          </div>
          <div class="edit-field">
            <label>地址</label>
            <input v-model="editForm.host" class="edit-input" placeholder="IP 或域名" />
          </div>
          <div class="edit-field">
            <label>端口</label>
            <input v-model.number="editForm.port" type="number" class="edit-input" placeholder="22" />
          </div>
          <div class="edit-field">
            <label>用户名</label>
            <input v-model="editForm.username" class="edit-input" placeholder="root" />
          </div>
          <div class="edit-field">
            <label>密码</label>
            <input v-model="editForm.password" type="password" class="edit-input" placeholder="留空则不修改" />
          </div>
          <div v-if="editError" class="edit-error">{{ editError }}</div>
        </div>
        <div class="edit-modal-foot">
          <button class="edit-btn" @click="showEditDialog = false">取消</button>
          <button class="edit-btn edit-btn-primary" @click="saveEdit" :disabled="editSaving">
            {{ editSaving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- 自定义确认弹窗 -->
  <Modal
    v-model:visible="confirmVisible"
    :title="confirmTitle"
    :content="confirmContent"
    confirm-text="确定"
    cancel-text="取消"
    :danger="true"
    width="360px"
    @confirm="handleConfirmOk"
    @cancel="handleConfirmCancel"
    @close="handleConfirmCancel"
  />

  <!-- 全局消息提示 -->
  <Message ref="messageRef" />
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useSSHConnectionsStore } from '../../../stores/sshConnections'
import { DeleteConnection, CreateAndConnectWithGroup, GetAllGroups, GetDefaultGroupID, OpenSSHWindow } from '../../../../bindings/changeme/ssh/sshservice.js'

import { Events } from '@wailsio/runtime'
import ConnectionOptionsDialog from '../ConnectionOptionsDialog.vue'
import Modal from '../../../components/Modal.vue'
import Message from '../../../components/Message.vue'
import {SSHConfig, SSHService} from "@bindings/changeme/ssh/index.js";

const connectionsStore = useSSHConnectionsStore()

const searchQuery = ref('')
const activeConnectionId = ref(null)
const sortBy = ref('name')
const activeMenu = ref(null)
const menuPos = ref({ top: 0, left: 0 })

// 自定义确认弹窗状态
const confirmVisible = ref(false)
const confirmTitle = ref('')
const confirmContent = ref('')
const confirmAction = ref(null) // 确认后要执行的回调
const messageRef = ref(null)

// 使用 store 中的数据
const connections = computed(() => connectionsStore.connections)
const loading = computed(() => connectionsStore.loading)

// 当前菜单对应的连接对象
const menuConn = computed(() => {
  if (!activeMenu.value) return null
  return connections.value.find(c => c.id === activeMenu.value) || null
})

// 搜索过滤 + 排序
const filterAndSort = (list) => {
  let result = [...list]
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(conn =>
      conn.name.toLowerCase().includes(query) ||
      conn.host.toLowerCase().includes(query)
    )
  }
  result.sort((a, b) => {
    const sa = a.status === 'connected' ? 0 : 1
    const sb = b.status === 'connected' ? 0 : 1
    if (sa !== sb) return sa - sb
    return a.name.localeCompare(b.name) || a.id.localeCompare(b.id)
  })
  return result
}

// 已保存的连接（永久存储）
const savedConnections = computed(() => filterAndSort(connections.value.filter(c => c.saved)))

// 缓存的连接（临时）
const cachedConnections = computed(() => filterAndSort(connections.value.filter(c => !c.saved)))

const selectConnection = (conn) => {
  activeConnectionId.value = conn.id
  closeMenu()
}

// 编辑连接
const showEditDialog = ref(false)
const editSaving = ref(false)
const editError = ref('')
const editForm = reactive({
  id: '',
  name: '',
  host: '',
  port: 22,
  username: '',
  password: ''
})

const editConnection = (conn) => {
  closeMenu()
  editForm.id = conn.id
  editForm.name = conn.name || ''
  editForm.host = conn.host || ''
  editForm.port = conn.port || 22
  editForm.username = conn.username || ''
  editForm.password = '' // 不回显密码
  editError.value = ''
  showEditDialog.value = true
}

const saveEdit = async () => {
  if (!editForm.host.trim() || !editForm.username.trim()) {
    editError.value = '请填写地址和用户名'
    return
  }
  editSaving.value = true
  editError.value = ''
  try {
    const connInfo = await SSHService.GetConnection(editForm.id)
    if (!connInfo) {
      editError.value = '连接不存在'
      return
    }
    // 更新字段
    connInfo.name = editForm.name || `${editForm.username}@${editForm.host}`
    connInfo.host = editForm.host
    connInfo.port = editForm.port || 22
    connInfo.username = editForm.username
    if (editForm.password) {
      connInfo.password = editForm.password
    }
    await SSHService.UpdateConnection(connInfo)
    connectionsStore.refresh()
    showEditDialog.value = false
    messageRef.value?.success('连接已更新')
  } catch (e) {
    editError.value = String(e?.message || e)
  } finally {
    editSaving.value = false
  }
}

// 快速连接（复用 NewConnection.vue 的逻辑）
const showConnDialog = ref(false)
const pendingConn = ref(null)

const quickConnect = async (conn) => {
  closeMenu()
  if (conn.status === 'connected') return

  // 保存待连接配置
  pendingConn.value = {
    name: conn.name || `${conn.username}@${conn.host}`,
    host: conn.host,
    port: conn.port,
    username: conn.username,
    password: conn.password || undefined,
    privateKey: conn.privateKey || undefined,
    keyPath: conn.keyPath || undefined,
    timeout: 30
  }

  // 检查是否已有 SSH 窗口
  let hasExistingWindow = false
  try {
    const groups = await GetAllGroups()
    for (const group of groups) {
      if (group.conn_ids && group.conn_ids.length > 0) {
        hasExistingWindow = true
        break
      }
    }
  } catch (e) {}

  if (!hasExistingWindow) {
    // 没有窗口，直接用默认分组连接
    await executeQuickConnect('default-group')
  } else {
    // 已有窗口，弹出选择对话框
    showConnDialog.value = true
  }
}

const handleConnDialogSelect = async (type) => {
  showConnDialog.value = false
  if (type === 'cancel') {
    pendingConn.value = null
    return
  }
  await executeQuickConnect(type)
}

const handleConnDialogCancel = () => {
  showConnDialog.value = false
  pendingConn.value = null
}

const executeQuickConnect = async (type) => {
  const config = pendingConn.value
  if (!config) return

  try {
    let groupID
    if (type === 'new-window') {
      groupID = await (await import('../../../../bindings/changeme/ssh/sshservice.js')).CreateGroup(config.name)
    } else {
      groupID = await GetDefaultGroupID()
    }

    const result = await CreateAndConnectWithGroup(new SSHConfig(config), groupID)
    connectionsStore.refresh()
    messageRef.value?.success(`已连接 ${config.name}`)

    // 打开 SSH 窗口
    await OpenSSHWindow(result.groupID, config.name, result.connID)
  } catch (e) {
    messageRef.value?.error(`连接失败: ${e?.message || e}`)
  } finally {
    pendingConn.value = null
  }
}

// 菜单相关功能
const toggleMenu = (event, connId) => {
  if (activeMenu.value === connId) {
    closeMenu()
  } else {
    // 根据触发按钮的位置计算菜单的固定坐标
    const rect = event.currentTarget.getBoundingClientRect()
    menuPos.value = {
      top: rect.bottom + 4,
      left: rect.left
    }
    activeMenu.value = connId
  }
}

const closeMenu = () => {
  activeMenu.value = null
}

// 显示自定义确认弹窗
const showConfirm = (title, content, action) => {
  confirmTitle.value = title
  confirmContent.value = content
  confirmAction.value = action
  confirmVisible.value = true
}

const handleConfirmOk = async () => {
  if (confirmAction.value) {
    await confirmAction.value()
  }
  confirmVisible.value = false
  confirmAction.value = null
}

const handleConfirmCancel = () => {
  confirmVisible.value = false
  confirmAction.value = null
}

const deleteSaved = (conn) => {
  closeMenu()
  showConfirm('删除本地存储', `确定要删除 "${conn.name}" 的本地存储吗？`, async () => {
    try {
      await DeleteConnection(conn.id)
      await connectionsStore.refresh()
      messageRef.value?.success('已删除本地存储')
    } catch (error) {
      messageRef.value?.error('删除失败: ' + error.message)
    }
  })
}

const disconnectConnection = (conn) => {
  closeMenu()
  showConfirm('关闭连接', `确定要关闭 "${conn.name}" 的连接吗？`, async () => {
    try {
      const { Disconnect } = await import('../../../../bindings/changeme/ssh/sshservice.js')
      await Disconnect(conn.id)
      await connectionsStore.refresh()
      messageRef.value?.success('已关闭连接')
    } catch (error) {
      messageRef.value?.error('断开连接失败: ' + error.message)
    }
  })
}

const saveConnection = async (conn) => {
  closeMenu()
  try {
    const { SaveConnection } = await import('../../../../bindings/changeme/ssh/sshservice.js')
    await SaveConnection(conn.id)
    await connectionsStore.refresh()
    messageRef.value?.success('已保存到本地')
  } catch (error) {
    messageRef.value?.error('保存失败: ' + error.message)
  }
}

// 点击外部关闭菜单
const handleClickOutside = (event) => {
  if (!event.target.closest('.more-btn') && !event.target.closest('.custom-menu')) {
    closeMenu()
  }
}

// 监听后端广播的连接状态更新
const handleConnectionsUpdated = (event) => {
  const { connections, timestamp } = event.data || {}
  
  if (connections && Array.isArray(connections)) {
    // 直接更新 store 中的数据（store内部会检测变化）
    connectionsStore.updateConnections(connections)
  }
}

// 监听窗口关闭事件（用户通过系统方式关闭窗口，SSH保持连接）
const handleWindowClosed = async (event) => {
  const { groupID } = event.data || {}
  if (groupID) {
    console.log('[SidebarPanel] 🪟 检测到窗口关闭:', groupID)
    // 强制刷新连接列表，确保状态同步
    await connectionsStore.refresh()
    console.log('[SidebarPanel] ✅ 已刷新连接列表')
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  
  // 从 store 加载数据
  connectionsStore.loadConnections()
  
  // 注册事件监听器
  Events.On('ssh:connections-updated', handleConnectionsUpdated)
  Events.On('ssh:window-closed', handleWindowClosed)
  console.log('[SidebarPanel] 👂 已注册事件监听器: ssh:connections-updated, ssh:window-closed')
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  
  // 移除事件监听器
  Events.Off('ssh:connections-updated', handleConnectionsUpdated)
  Events.Off('ssh:window-closed', handleWindowClosed)
  console.log('[SidebarPanel] 🔇 已移除事件监听器: ssh:connections-updated, ssh:window-closed')
})

const toggleSort = () => {
  sortBy.value = sortBy.value === 'name' ? 'status' : 'name'
}
</script>

<style>
/* Teleport 到 body 的菜单样式，不能用 scoped */
.custom-menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99999;
}

.custom-menu {
  position: fixed;
  background: rgba(45, 45, 45, 0.98);
  backdrop-filter: blur(20px);
  border: 0.0625rem solid rgba(255, 255, 255, 0.15);
  border-radius: 0.5rem;
  padding: 0.25rem;
  min-width: 10rem;
  box-shadow: 0 0.5rem 1.5rem rgba(0, 0, 0, 0.4);
  z-index: 100000;
}

.custom-menu .menu-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  color: #e2e8f0;
  font-size: 0.8125rem;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
}

.custom-menu .menu-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.custom-menu .menu-item svg {
  flex-shrink: 0;
  opacity: 0.7;
}

.custom-menu .menu-item-primary {
  color: #63b3ed;
}

.custom-menu .menu-item-primary svg {
  opacity: 1;
  fill: #63b3ed;
}

.custom-menu .menu-item-primary:hover {
  background: rgba(66, 153, 225, 0.15);
}

.custom-menu .menu-sep {
  height: 1px;
  background: rgba(255, 255, 255, 0.08);
  margin: 0.25rem 0.5rem;
}

/* 编辑连接弹窗 */
.edit-modal-mask {
  position: fixed; inset: 0;
  background: rgba(0, 0, 0, 0.7); display: flex;
  align-items: center; justify-content: center;
  z-index: 10000; backdrop-filter: blur(4px);
}
.edit-modal {
  background: #1a1a1a; border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.75rem; width: 380px; max-width: 90vw;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
}
.edit-modal-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1rem 1.25rem; border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.edit-modal-head h3 { margin: 0; color: #e2e8f0; font-size: 0.9375rem; }
.edit-modal-head button { background: none; border: none; color: #718096; font-size: 1.25rem; cursor: pointer; }
.edit-modal-head button:hover { color: #e2e8f0; }
.edit-modal-body { padding: 1rem 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; }
.edit-modal-foot {
  display: flex; justify-content: flex-end; gap: 0.5rem;
  padding: 0.875rem 1.25rem; border-top: 1px solid rgba(255, 255, 255, 0.06);
}
.edit-field label { display: block; color: #a0aec0; font-size: 0.75rem; margin-bottom: 0.25rem; }
.edit-input {
  width: 100%; background: rgba(30, 30, 30, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 0.375rem;
  color: #e2e8f0; font-size: 0.8125rem; padding: 0.5rem 0.75rem;
  outline: none; box-sizing: border-box;
}
.edit-input:focus { border-color: rgba(66, 153, 225, 0.4); }
.edit-error { color: #fc8181; font-size: 0.75rem; }
.edit-btn {
  display: inline-flex; align-items: center; gap: 0.375rem;
  padding: 0.5rem 1rem;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 0.375rem;
  color: #a0aec0; font-size: 0.8125rem;
  cursor: pointer; transition: all 0.15s;
}
.edit-btn:hover:not(:disabled) { background: rgba(255, 255, 255, 0.1); color: #e2e8f0; }
.edit-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.edit-btn-primary { background: rgba(66, 153, 225, 0.2); border-color: rgba(66, 153, 225, 0.4); color: #63b3ed; }
.edit-btn-primary:hover:not(:disabled) { background: rgba(66, 153, 225, 0.35); }
</style>

<style scoped>
.sidebar-panel {
  width: 13.75rem;
  min-width: 13.75rem;
  background: transparent;
  border-right: none;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  z-index: 1;
  padding: 0.9375rem;
}

/* 搜索框 */
.search-box {
  position: relative;
  margin-bottom: 0.625rem;
  flex-shrink: 0;
}

.search-icon {
  position: absolute;
  left: 0.625rem;
  top: 50%;
  transform: translateY(-50%);
  color: #718096;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 0.5rem 0.625rem 0.5rem 2rem;
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px);
  border: 0.0625rem solid rgba(255, 255, 255, 0.15);
  border-radius: 0.375rem;
  color: #e2e8f0;
  font-size: 0.8125rem;
  outline: none;
  transition: all 0.2s;
  box-sizing: border-box;
}

.search-input::placeholder {
  color: #718096;
}

.search-input:focus {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(66, 153, 225, 0.5);
  box-shadow: 0 0 0 0.125rem rgba(66, 153, 225, 0.15);
}

/* 连接列表 */
.connection-list {
  flex: 1;
  overflow-y: auto;
  margin-bottom: 0.625rem;
}

.loading-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.875rem 0.9375rem;
  color: #718096;
  text-align: center;
}

.spinner {
  width: 1.5rem;
  height: 1.5rem;
  border: 0.125rem solid rgba(255, 255, 255, 0.1);
  border-top-color: #4299e1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 0.5rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state p {
  margin: 0.5rem 0 0;
  font-size: 0.8125rem;
}

/* 分组样式 */
.connection-group {
  margin-bottom: 0.9375rem;
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.375rem 0.5rem;
  margin-bottom: 0.375rem;
}

.group-title {
  color: #a0aec0;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03125rem;
}

.group-count {
  color: #718096;
  font-size: 0.6875rem;
  background: rgba(255, 255, 255, 0.05);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
}

/* 分段标题（已保存 / 缓存） */
.connection-section {
  margin-bottom: 0.25rem;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  user-select: none;
}

.section-title {
  color: #718096;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.section-count {
  color: #4a5568;
  font-size: 0.625rem;
  background: rgba(255, 255, 255, 0.04);
  padding: 0.0625rem 0.375rem;
  border-radius: 0.25rem;
}

.connections {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.connection-item {
  padding: 0.625rem;
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(10px);
  border: 0.0625rem solid rgba(255, 255, 255, 0.08);
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
}

.connection-item:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.15);
}

.connection-item.active {
  background: rgba(66, 153, 225, 0.12);
  border-color: rgba(66, 153, 225, 0.3);
}

.connection-item.connected {
  border-left: 3px solid #68d391;
}

.connection-item.cached {
  opacity: 0.65;
  border-style: dashed;
}

.connection-content {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.connection-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
}

.connection-name {
  color: #e2e8f0;
  font-size: 0.8125rem;
  font-weight: 600;
  text-align: left;
  flex: 1;
  margin-right: 0.375rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.more-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 0.25rem;
  color: #718096;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.more-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

.status-badge {
  padding: 0  0.5rem;
  border-radius: 0.25rem;
  font-size: 0.625rem;
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
}

.status-badge.connected {
  background: rgba(72, 187, 120, 0.15);
  color: #48bb78;
  border: 0.0625rem solid rgba(72, 187, 120, 0.3);
}

.status-badge.disconnected {
  background: rgba(160, 174, 192, 0.1);
  color: #a0aec0;
  border: 0.0625rem solid rgba(160, 174, 192, 0.2);
}

.connection-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.connection-info {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  color: #718096;
  font-size: 0.6875rem;
  font-family: 'Consolas', monospace;
}

.address-icon {
  flex-shrink: 0;
  opacity: 0.6;
}

/* 底部工具栏 */
.bottom-toolbar {
  display: flex;
  gap: 0.375rem;
  flex-shrink: 0;
  padding-top: 0.625rem;
  border-top: 0.0625rem solid rgba(255, 255, 255, 0.1);
}

.icon-btn {
  flex: 1;
  padding: 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  border: 0.0625rem solid rgba(255, 255, 255, 0.1);
  border-radius: 0.375rem;
  color: #a0aec0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
  color: #e2e8f0;
}

/* 滚动条样式 */
.connection-list::-webkit-scrollbar {
  width: 0.25rem;
}

.connection-list::-webkit-scrollbar-track {
  background: transparent;
}

.connection-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 0.125rem;
}

.connection-list::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.25);
}
</style>
