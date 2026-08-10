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
        :placeholder="t('common.search') + '...'"
        class="search-input"
      />
    </div>
    
    <!-- 连接列表 -->
    <div class="connection-list">
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <span>{{ t('common.loading') }}</span>
      </div>

      <div v-else-if="savedConnections.length === 0 && cachedConnections.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 16 16" fill="currentColor" opacity="0.3">
          <path d="M8 1a2 2 0 0 1 2 2v4H6V3a2 2 0 0 1 2-2zm3 6V3a3 3 0 0 0-6 0v4a2 2 0 0 0-2 2v5a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"/>
        </svg>
        <p>{{ t('common.noData') }}</p>
        <button class="import-btn" @click="importSSHConfig">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M.5 9.9a.5.5 0 0 1 .5.5v2.5a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-2.5a.5.5 0 0 1 1 0v2.5a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2v-2.5a.5.5 0 0 1 .5-.5z"/><path d="M7.646 11.854a.5.5 0 0 0 .708 0l3-3a.5.5 0 0 0-.708-.708L8.5 10.293V1.5a.5.5 0 0 0-1 0v8.793L5.354 8.146a.5.5 0 1 0-.708.708l3 3z"/></svg>
          {{ t('connection.importSSHConfig') }}
        </button>
      </div>

      <template v-else>
        <!-- 已保存的连接 -->
        <div v-if="savedConnections.length > 0" class="connection-section">
          <div class="section-header">
            <svg class="icon-success" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
            <span class="section-title">{{ t('connection.saved') }}</span>
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
                  <button class="more-btn" @click.stop="toggleMenu($event, conn.id)" :title="t('sidebar.moreActions')">
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor"><path d="M3 9.5a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3z"/></svg>
                  </button>
                </div>
                <div class="connection-bottom">
                  <div class="connection-info">
                    <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor" class="address-icon"><path d="M8.051 1.999h.089c.822.003 4.987.033 6.11.335a2.01 2.01 0 0 1 1.415 1.42c.101.38.172.883.22 1.402l.01.104.022.26.008.104c.065.914.073 1.77.074 1.957v.075c-.001.194-.01 1.108-.082 2.06l-.008.105-.009.104c-.05.572-.124 1.14-.235 1.558a2.007 2.007 0 0 1-1.415 1.42c-1.16.312-5.569.334-6.18.335h-.142c-.309 0-1.587-.006-2.927-.052l-.17-.006-.087-.004-.171-.007-.171-.007c-1.11-.049-2.167-.128-2.654-.26a2.007 2.007 0 0 1-1.415-1.419c-.111-.417-.185-.986-.235-1.558L.09 9.82l-.008-.104A31.4 31.4 0 0 1 0 7.68v-.123c.002-.215.01-.958.064-1.778l.007-.103.003-.052.008-.104.022-.26.01-.104c.048-.519.119-1.023.22-1.402a2.007 2.007 0 0 1 1.415-1.42c.487-.13 1.544-.21 2.654-.26l.17-.007.172-.006.086-.003.171-.007A99.788 99.788 0 0 1 7.858 2h.193zM6.4 5.209v4.818l4.157-2.408L6.4 5.209z"/></svg>
                    <span>{{ conn.host }}:{{ conn.port }}</span>
                  </div>
                  <div class="status-badge" :class="conn.status">
                    {{ conn.status === 'connected' ? t('common.connected') : t('common.disconnected') }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 缓存的连接 -->
        <div v-if="cachedConnections.length > 0" class="connection-section">
          <div class="section-header">
            <svg class="icon-muted" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            <span class="section-title">{{ t('connection.cached') }}</span>
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
                  <button class="more-btn" @click.stop="toggleMenu($event, conn.id)" :title="t('sidebar.moreActions')">
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor"><path d="M3 9.5a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm5 0a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3z"/></svg>
                  </button>
                </div>
                <div class="connection-bottom">
                  <div class="connection-info">
                    <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor" class="address-icon"><path d="M8.051 1.999h.089c.822.003 4.987.033 6.11.335a2.01 2.01 0 0 1 1.415 1.42c.101.38.172.883.22 1.402l.01.104.022.26.008.104c.065.914.073 1.77.074 1.957v.075c-.001.194-.01 1.108-.082 2.06l-.008.105-.009.104c-.05.572-.124 1.14-.235 1.558a2.007 2.007 0 0 1-1.415 1.42c-1.16.312-5.569.334-6.18.335h-.142c-.309 0-1.587-.006-2.927-.052l-.17-.006-.087-.004-.171-.007-.171-.007c-1.11-.049-2.167-.128-2.654-.26a2.007 2.007 0 0 1-1.415-1.419c-.111-.417-.185-.986-.235-1.558L.09 9.82l-.008-.104A31.4 31.4 0 0 1 0 7.68v-.123c.002-.215.01-.958.064-1.778l.007-.103.003-.052.008-.104.022-.26.01-.104c.048-.519.119-1.023.22-1.402a2.007 2.007 0 0 1 1.415-1.42c.487-.13 1.544-.21 2.654-.26l.17-.007.172-.006.086-.003.171-.007A99.788 99.788 0 0 1 7.858 2h.193zM6.4 5.209v4.818l4.157-2.408L6.4 5.209z"/></svg>
                    <span>{{ conn.host }}:{{ conn.port }}</span>
                  </div>
                  <div class="status-badge" :class="conn.status">
                    {{ conn.status === 'connected' ? t('common.connected') : t('common.disconnected') }}
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
      <button class="icon-btn" @click="toggleSort" :title="sortBy === 'name' ? t('sidebar.sortByName') : t('sidebar.sortByStatus')">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
          <path d="M4 10.5a.5.5 0 0 0 .5.5h5a.5.5 0 0 0 0-1h-5a.5.5 0 0 0-.5.5zm0-3a.5.5 0 0 0 .5.5h7a.5.5 0 0 0 0-1h-7a.5.5 0 0 0-.5.5zm0-3a.5.5 0 0 0 .5.5h9a.5.5 0 0 0 0-1h-9a.5.5 0 0 0-.5.5z"/>
        </svg>
      </button>
      <button class="icon-btn" @click="openSettings" :title="t('sidebar.settings')">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="3"/>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
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
          <span>{{ t('sidebar.quickConnect') }}</span>
        </div>
        <div v-if="menuConn.status !== 'connected'" class="menu-item" @click.stop="editConnection(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M12.146.146a.5.5 0 0 1 .708 0l3 3a.5.5 0 0 1 0 .708l-10 10a.5.5 0 0 1-.168.11l-5 2a.5.5 0 0 1-.65-.65l2-5a.5.5 0 0 1 .11-.168l10-10zM11.207 2.5L13.5 4.793 14.793 3.5 12.5 1.207 11.207 2.5zm1.586 3L10.5 3.207 4 9.707V10h.5a.5.5 0 0 1 .5.5v.5h.5a.5.5 0 0 1 .5.5v.5h.293l6.5-6.5zm-9.761 5.175l-.106.106-1.528 3.821 3.821-1.528.106-.106A.5.5 0 0 1 13 15.5V15a.5.5 0 0 0-.5-.5H13a.5.5 0 0 1-.5-.5V14a.5.5 0 0 0-.5-.5h-.5a.5.5 0 0 1-.5-.5v-.5a.5.5 0 0 0-.5-.5h-.5a.5.5 0 0 1-.5-.5v-.5a.5.5 0 0 0-.5-.5h-.293z"/>
          </svg>
          <span>{{ t('sidebar.edit') }}</span>
        </div>
        <div class="menu-sep"></div>
        <div v-if="menuConn.saved" class="menu-item" @click.stop="deleteSaved(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M5.5 5.5A.5.5 0 0 1 6 6v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm2.5 0a.5.5 0 0 1 .5.5v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm3 .5a.5.5 0 0 0-1 0v6a.5.5 0 0 0 1 0V6z"/>
            <path fill-rule="evenodd" d="M14.5 3a1 1 0 0 1-1 1H13v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4h-.5a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1H6a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1h3.5a1 1 0 0 1 1 1v1zM4.118 4 4 4.059V13a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V4.059L11.882 4H4.118zM2.5 3V2h11v1h-11z"/>
          </svg>
          <span>{{ t('sidebar.deleteLocal') }}</span>
        </div>
        <div v-if="menuConn.status === 'connected'" class="menu-item" @click.stop="disconnectConnection(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M11.251.068a.5.5 0 0 1 .227.58L9.677 6.5H13a.5.5 0 0 1 .364.843l-8 8.5a.5.5 0 0 1-.864-.5L6.323 9.5H3a.5.5 0 0 1-.364-.843l8-8.5a.5.5 0 0 1 .615-.09z"/>
          </svg>
          <span>{{ t('sidebar.closeConnection') }}</span>
        </div>
        <div v-if="!menuConn.saved" class="menu-item" @click.stop="saveConnection(menuConn)">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <path d="M2 2a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V2zm10-1H4a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V2a1 1 0 0 0-1-1z"/>
            <path d="M10.854 5.146a.5.5 0 0 1 0 .708l-3 3a.5.5 0 0 1-.708 0l-1.5-1.5a.5.5 0 1 1 .708-.708L7.5 7.793l2.646-2.647a.5.5 0 0 1 .708 0z"/>
          </svg>
          <span>{{ t('sidebar.saveLocal') }}</span>
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
          <h3>{{ t('sidebar.editConnection') }}</h3>
          <button @click="showEditDialog = false">&times;</button>
        </div>
        <div class="edit-modal-body">
          <div class="edit-field">
            <label>{{ t('connection.name') }}</label>
            <input v-model="editForm.name" class="edit-input" :placeholder="t('connection.name')" />
          </div>
          <div class="edit-field">
            <label>{{ t('connection.host') }}</label>
            <input v-model="editForm.host" class="edit-input" placeholder="IP / Domain" />
          </div>
          <div class="edit-field">
            <label>{{ t('connection.port') }}</label>
            <input v-model.number="editForm.port" type="number" class="edit-input" placeholder="22" />
          </div>
          <div class="edit-field">
            <label>{{ t('connection.username') }}</label>
            <input v-model="editForm.username" class="edit-input" placeholder="root" />
          </div>
          <div class="edit-field">
            <label>{{ t('connection.password') }}</label>
            <input v-model="editForm.password" type="password" class="edit-input" :placeholder="t('common.cancel')" />
          </div>
          <div v-if="editError" class="edit-error">{{ editError }}</div>
        </div>
        <div class="edit-modal-foot">
          <button class="edit-btn" @click="showEditDialog = false">{{ t('common.cancel') }}</button>
          <button class="edit-btn edit-btn-primary" @click="saveEdit" :disabled="editSaving">
            {{ editSaving ? t('common.loading') : t('common.save') }}
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
    :confirm-text="t('common.confirm')"
    :cancel-text="t('common.cancel')"
    :danger="true"
    width="360px"
    @confirm="handleConfirmOk"
    @cancel="handleConfirmCancel"
    @close="handleConfirmCancel"
  />

  <!-- SSH 配置导入对话框 -->
  <Modal
    v-model:visible="importDialogVisible"
    :title="t('sidebar.importSSHConfig')"
    :cancel-text="t('common.cancel')"
    @close="importDialogVisible = false"
    @cancel="importDialogVisible = false"
    @confirm="confirmImport"
    width="500px"
  >
    <div class="import-dialog">
      <div v-if="importLoading" class="import-loading">
        <div class="spinner"></div>
        <span>{{ t('sidebar.importParsing') }}</span>
      </div>
      <div v-else-if="importConfigs.length === 0" class="import-empty">
        <p>{{ t('sidebar.importNoConfigs') }}</p>
        <p class="import-hint">{{ t('sidebar.importNoConfigsHint') }}</p>
      </div>
      <div v-else class="import-list">
        <p class="import-hint">{{ t('sidebar.importFound', { count: importConfigs.length }) }}</p>
        <div class="import-select-all">
          <label>
            <input type="checkbox" v-model="importSelectAll" @change="toggleSelectAll" />
            {{ t('sidebar.importSelectAll') }}
          </label>
        </div>
        <div v-for="(config, index) in importConfigs" :key="index" class="import-item">
          <label>
            <input type="checkbox" v-model="config.selected" />
            <span class="import-name">{{ config.name }}</span>
            <span class="import-info">{{ config.username }}@{{ config.host }}:{{ config.port }}</span>
          </label>
        </div>
      </div>
    </div>
  </Modal>

  <!-- 全局消息提示 -->
  <Message ref="messageRef" />
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useSSHConnectionsStore } from '../../../stores/sshConnections'
import { useConfigStore } from '../../../stores/config'
import { DeleteConnection, CreateAndConnectWithGroup, GetAllGroups, GetDefaultGroupID, OpenSSHWindow } from '../../../../bindings/changeme/ssh/sshservice.js'

import { Events } from '@wailsio/runtime'
import ConnectionOptionsDialog from '../ConnectionOptionsDialog.vue'
import Modal from '../../../components/Modal.vue'
import Message from '../../../components/Message.vue'
import {SSHConfig, SSHService} from "@bindings/changeme/ssh/index.js";

const { t } = useI18n()
const router = useRouter()
const connectionsStore = useSSHConnectionsStore()

const searchQuery = ref('')
const activeConnectionId = ref(null)
const sortBy = ref('name')
const activeMenu = ref(null)
const menuPos = ref({ top: 0, left: 0 })

// SSH 配置导入状态
const importDialogVisible = ref(false)
const importLoading = ref(false)
const importConfigs = ref([])
const importSelectAll = ref(true)

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

// SSH 配置导入
const importSSHConfig = async () => {
  importDialogVisible.value = true
  importLoading.value = true
  importConfigs.value = []

  try {
    const configs = await SSHService.ImportSSHConfig()
    importConfigs.value = (configs || []).map(c => ({
      name: c.name || c.host,
      host: c.host,
      port: c.port || 22,
      username: c.username || 'root',
      keyPath: c.keyPath || '',
      selected: true
    }))
  } catch (error) {
    console.error('[SidebarPanel] 导入 SSH 配置失败:', error)
    messageRef.value?.error(t('sidebar.importFailed', { message: error?.message || error }))
  } finally {
    importLoading.value = false
  }
}

const toggleSelectAll = () => {
  importConfigs.value.forEach(c => c.selected = importSelectAll.value)
}

const confirmImport = async () => {
  const selected = importConfigs.value.filter(c => c.selected)
  if (selected.length === 0) {
    messageRef.value?.warning(t('sidebar.importSelectMin'))
    return
  }

  let successCount = 0
  for (const config of selected) {
    try {
      const sshConfig = new SSHConfig()
      sshConfig.name = config.name
      sshConfig.host = config.host
      sshConfig.port = config.port
      sshConfig.username = config.username
      sshConfig.keyPath = config.keyPath
      await SSHService.SyncImportConnection({
        id: '',
        name: config.name,
        host: config.host,
        port: config.port,
        username: config.username,
        keyPath: config.keyPath,
        status: 'disconnected',
        saved: true
      })
      successCount++
    } catch (error) {
      console.error('[SidebarPanel] 导入连接失败:', config.name, error)
    }
  }

  importDialogVisible.value = false
  connectionsStore.refresh()
  messageRef.value?.success(t('sidebar.importSuccess', { count: successCount }))
}

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
    editError.value = t('sidebar.editFillRequired')
    return
  }
  editSaving.value = true
  editError.value = ''
  try {
    const connInfo = await SSHService.GetConnection(editForm.id)
    if (!connInfo) {
      editError.value = t('sidebar.editNotFound')
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
    messageRef.value?.success(t('sidebar.connectionUpdated'))
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
      if (group && group.conn_ids && group.conn_ids.length > 0) {
        hasExistingWindow = true
        break
      }
    }
  } catch (e) {}

  if (!hasExistingWindow) {
    // 没有窗口，直接用默认分组连接
    await executeQuickConnect('default-group')
  } else {
    // 已有窗口，根据配置决定行为
    const cfgStore = useConfigStore()
    await cfgStore.init()
    const groupBehavior = cfgStore.get('advanced', 'groupBehavior') || 'prompt'
    console.log('[SidebarPanel] 📋 分组行为配置:', groupBehavior)

    if (groupBehavior === 'join_default') {
      console.log('[SidebarPanel] 🎯 自动加入默认分组')
      await executeQuickConnect('default-group')
    } else if (groupBehavior === 'new_window') {
      console.log('[SidebarPanel] 🪟 自动打开新窗口')
      await executeQuickConnect('new-window')
    } else {
      console.log('[SidebarPanel] ✅ 显示选择对话框')
      showConnDialog.value = true
    }
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
    messageRef.value?.success(t('sidebar.connected', { name: config.name }))

    // 打开 SSH 窗口
    await OpenSSHWindow(result.groupID || '', config.name || '', result.connID || '')

    // 如果开启了自动托盘，连接成功后隐藏主窗口
    const cfgStore = useConfigStore()
    await cfgStore.init()
    if (cfgStore.get('ui', 'autoTray')) {
      Events.Emit('ssh:tray-hide')
    }
  } catch (e) {
    messageRef.value?.error(t('sidebar.connectFailed', { message: e?.message || e }))
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
  showConfirm(t('sidebar.deleteLocal'), t('sidebar.deleteConfirm', { name: conn.name }), async () => {
    try {
      await DeleteConnection(conn.id)
      await connectionsStore.refresh()
      messageRef.value?.success(t('sidebar.deleted'))
    } catch (error) {
      messageRef.value?.error(t('sidebar.deleteFailed', { message: error.message }))
    }
  })
}

const disconnectConnection = (conn) => {
  closeMenu()
  showConfirm(t('sidebar.closeConnection'), t('sidebar.closeConfirm', { name: conn.name }), async () => {
    try {
      const { Disconnect } = await import('../../../../bindings/changeme/ssh/sshservice.js')
      await Disconnect(conn.id)
      await connectionsStore.refresh()
      messageRef.value?.success(t('sidebar.disconnected'))
    } catch (error) {
      messageRef.value?.error(t('sidebar.disconnectFailed', { message: error.message }))
    }
  })
}

const saveConnection = async (conn) => {
  closeMenu()
  try {
    const { SaveConnection } = await import('../../../../bindings/changeme/ssh/sshservice.js')
    await SaveConnection(conn.id)
    await connectionsStore.refresh()
    messageRef.value?.success(t('sidebar.saved'))
  } catch (error) {
    messageRef.value?.error(t('sidebar.saveFailed', { message: error.message }))
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
  console.log('[SidebarPanel] 📡 收到 ssh:connections-updated 事件:', {
    timestamp,
    connectionsCount: connections?.length,
    hasConnections: !!connections,
    isArray: Array.isArray(connections)
  })

  if (connections && Array.isArray(connections)) {
    // 直接更新 store 中的数据（store内部会检测变化）
    console.log('[SidebarPanel] 📤 调用 updateConnections，连接数:', connections.length)
    connectionsStore.updateConnections(connections)
  } else {
    console.log('[SidebarPanel] ⚠️ 事件数据无效，跳过更新')
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

const openSettings = () => {
  router.push('/home/settings')
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
  background: var(--bg-panel);
  backdrop-filter: blur(20px);
  border: 0.0625rem solid var(--border-strong);
  border-radius: 0.5rem;
  padding: 0.25rem;
  min-width: 10rem;
  box-shadow: 0 0.5rem 1.5rem var(--bg-overlay);
  z-index: 100000;
}

.custom-menu .menu-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  color: var(--text-primary);
  font-size: 0.8125rem;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
}

.custom-menu .menu-item:hover {
  background: var(--surface-hover);
}

.custom-menu .menu-item svg {
  flex-shrink: 0;
  opacity: 0.7;
}

.custom-menu .menu-item-primary {
  color: var(--primary-light);
}

.custom-menu .menu-item-primary svg {
  opacity: 1;
  fill: var(--primary-light);
}

.custom-menu .menu-item-primary:hover {
  background: var(--primary-bg);
}

.custom-menu .menu-sep {
  height: 1px;
  background: var(--surface-3);
  margin: 0.25rem 0.5rem;
}

.icon-success {
  color: var(--success-light, #68d391);
}

.icon-muted {
  color: var(--text-muted, #718096);
}

/* 编辑连接弹窗 */
.edit-modal-mask {
  position: fixed; inset: 0;
  background: var(--bg-overlay); display: flex;
  align-items: center; justify-content: center;
  z-index: 10000; backdrop-filter: blur(4px);
}
.edit-modal {
  background: var(--bg-panel); border: 1px solid var(--surface-hover);
  border-radius: 0.75rem; width: 380px; max-width: 90vw;
  box-shadow: 0 16px 48px var(--shadow-lg, rgba(0, 0, 0, 0.5));
}
.edit-modal-head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--border-subtle);
}
.edit-modal-head h3 { margin: 0; color: var(--text-primary); font-size: 0.9375rem; }
.edit-modal-head button { background: none; border: none; color: var(--text-muted); font-size: 1.25rem; cursor: pointer; }
.edit-modal-head button:hover { color: var(--text-primary); }
.edit-modal-body { padding: 1rem 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; }
.edit-modal-foot {
  display: flex; justify-content: flex-end; gap: 0.5rem;
  padding: 0.875rem 1.25rem; border-top: 1px solid var(--border-subtle);
}
.edit-field label { display: block; color: var(--text-secondary); font-size: 0.75rem; margin-bottom: 0.25rem; }
.edit-input {
  width: 100%; background: var(--bg-panel);
  border: 1px solid var(--surface-hover); border-radius: 0.375rem;
  color: var(--text-primary); font-size: 0.8125rem; padding: 0.5rem 0.75rem;
  outline: none; box-sizing: border-box;
}
.edit-input:focus { border-color: var(--border-accent, rgba(66, 153, 225, 0.4)); }
.edit-error { color: var(--accent-danger); font-size: 0.75rem; }
.edit-btn {
  display: inline-flex; align-items: center; gap: 0.375rem;
  padding: 0.5rem 1rem;
  background: var(--border-subtle);
  border: 1px solid var(--surface-3);
  border-radius: 0.375rem;
  color: var(--text-secondary); font-size: 0.8125rem;
  cursor: pointer; transition: all 0.15s;
}
.edit-btn:hover:not(:disabled) { background: var(--surface-hover); color: var(--text-primary); }
.edit-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.edit-btn-primary { background: var(--primary-bg, rgba(66, 153, 225, 0.2)); border-color: var(--border-accent, rgba(66, 153, 225, 0.4)); color: var(--primary-light); }
.edit-btn-primary:hover:not(:disabled) { background: var(--primary-bg-hover, rgba(66, 153, 225, 0.35)); }
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
  color: var(--text-muted);
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 0.5rem 0.625rem 0.5rem 2rem;
  background: var(--surface-3);
  backdrop-filter: blur(10px);
  border: 0.0625rem solid var(--border-strong);
  border-radius: 0.375rem;
  color: var(--text-primary);
  font-size: 0.8125rem;
  outline: none;
  transition: all 0.2s;
  box-sizing: border-box;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-input:focus {
  background: var(--surface-3);
  border-color: var(--border-accent, rgba(66, 153, 225, 0.5));
  box-shadow: 0 0 0 0.125rem var(--primary-bg);
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
  color: var(--text-muted);
  text-align: center;
}

.spinner {
  width: 1.5rem;
  height: 1.5rem;
  border: 0.125rem solid var(--surface-hover);
  border-top-color: var(--accent-primary);
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
  color: var(--text-secondary);
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03125rem;
}

.group-count {
  color: var(--text-muted);
  font-size: 0.6875rem;
  background: var(--surface-1);
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
  color: var(--text-muted);
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.section-count {
  color: var(--text-disabled);
  font-size: 0.625rem;
  background: var(--surface-1);
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
  background: var(--surface-1);
  backdrop-filter: blur(10px);
  border: 0.0625rem solid var(--border-default);
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
}

.connection-item:hover {
  background: var(--surface-3);
  border-color: var(--border-strong);
}

.connection-item.active {
  background: var(--primary-bg, rgba(66, 153, 225, 0.12));
  border-color: var(--primary-bg-hover, rgba(66, 153, 225, 0.3));
}

.connection-item.connected {
  border-left: 3px solid var(--accent-success);
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
  color: var(--text-primary);
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
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.more-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
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
  background: var(--success-bg, rgba(72, 187, 120, 0.15));
  color: var(--accent-success);
  border: 0.0625rem solid var(--border-success, rgba(72, 187, 120, 0.3));
}

.status-badge.disconnected {
  background: var(--surface-2, rgba(160, 174, 192, 0.1));
  color: var(--text-secondary);
  border: 0.0625rem solid var(--surface-3, rgba(160, 174, 192, 0.2));
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
  color: var(--text-muted);
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
  border-top: 0.0625rem solid var(--surface-hover);
}

.icon-btn {
  flex: 1;
  padding: 0.5rem;
  background: var(--surface-1);
  border: 0.0625rem solid var(--surface-hover);
  border-radius: 0.375rem;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: var(--surface-hover);
  border-color: var(--scrollbar-thumb);
  color: var(--text-primary);
}

/* 滚动条样式 */
.connection-list::-webkit-scrollbar {
  width: 0.25rem;
}

.connection-list::-webkit-scrollbar-track {
  background: transparent;
}

.connection-list::-webkit-scrollbar-thumb {
  background: var(--border-strong);
  border-radius: 0.125rem;
}

.connection-list::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover);
}

/* 导入按钮 */
.import-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
  padding: 0.5rem 1rem;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-size: 0.8125rem;
  cursor: pointer;
  transition: opacity 0.2s;
}

.import-btn:hover {
  opacity: 0.9;
}

/* 导入对话框 */
.import-dialog {
  min-height: 200px;
}

.import-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 2rem;
  color: var(--text-secondary);
}

.import-empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.import-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin-bottom: 0.75rem;
}

.import-select-all {
  padding: 0.5rem 0;
  border-bottom: 0.0625rem solid var(--surface-hover);
  margin-bottom: 0.5rem;
}

.import-list {
  max-height: 300px;
  overflow-y: auto;
}

.import-item {
  padding: 0.375rem 0;
}

.import-item label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.import-name {
  font-weight: 500;
  color: var(--text-primary);
}

.import-info {
  font-size: 0.75rem;
  color: var(--text-muted);
}
</style>
