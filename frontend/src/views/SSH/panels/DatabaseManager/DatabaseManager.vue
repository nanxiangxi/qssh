<template>
  <div class="dbm-panel">
    <!-- 顶部连接栏 -->
    <div class="dbm-conn-bar">
      <div class="conn-fields">
        <select v-model="dbType" class="conn-sel">
          <option value="mysql">MySQL</option>
          <option value="postgresql">PostgreSQL</option>
        </select>
        <input v-model="dbHost" type="text" :placeholder="t('database.hostPlaceholder')" class="conn-inp w-120" />
        <input v-model.number="dbPort" type="number" :placeholder="t('database.portPlaceholder')" class="conn-inp w-70" />
        <input v-model="dbUser" type="text" :placeholder="t('database.usernamePlaceholder')" class="conn-inp w-100" />
        <input v-model="dbPassword" type="password" :placeholder="t('database.passwordPlaceholder')" class="conn-inp w-100" />
        <button class="btn btn-primary" @click="doConnect" :disabled="connecting">
          {{ connecting ? t('database.connecting') : t('database.connect') }}
        </button>
        <button v-if="connected" class="btn" @click="doDisconnect">{{ t('database.disconnect') }}</button>
      </div>
      <div v-if="connected" class="conn-status">
        <span class="dot green"></span>
        <span>{{ dbHost }}:{{ dbPort }}</span>
        <span class="sep">|</span>
        <span>{{ dbUser }}</span>
      </div>
    </div>

    <!-- 未连接 -->
    <div v-if="!connected" class="dbm-welcome">
      <div class="welcome-icon">
        <svg width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" opacity="0.15">
          <ellipse cx="12" cy="5" rx="9" ry="3"/>
          <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
          <path d="M3 12c0 1.66 4 3 9 3s9-1.34 9-3"/>
        </svg>
      </div>
      <p class="welcome-title">{{ t('database.databaseManager') }}</p>
      <p class="welcome-desc">{{ t('database.enterConnectionInfo') }}</p>
    </div>

    <!-- 主界面 -->
    <div v-else class="dbm-main">
      <!-- 左侧：数据库/表树 -->
      <DbSidebar
        :databases="databases"
        :tables="dbTables"
        :selected-db="selectedDb"
        :selected-table="selectedTable"
        :loading="loadingDbs"
        @select-db="onSelectDb"
        @select-table="onSelectTable"
        @refresh="refreshDatabases"
        @drop-db="dropDatabase"
      />

      <!-- 右侧：内容区 -->
      <div class="dbm-content">
        <!-- 无选中表 -->
        <div v-if="!selectedTable" class="dbm-empty">
          <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" opacity="0.15">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <line x1="3" y1="9" x2="21" y2="9"/>
            <line x1="9" y1="21" x2="9" y2="9"/>
          </svg>
          <p>{{ t('database.selectTableHint') }}</p>
        </div>

        <!-- 选中表后 -->
        <div v-else class="dbm-table-area">
          <!-- 标签页 -->
          <div class="dbm-tabs">
            <button v-for="tab in tabs" :key="tab.id" :class="{ active: activeTab === tab.id }" @click="switchTab(tab.id)">
              {{ tab.label }}
            </button>
          </div>

          <!-- 标签页内容 -->
          <div class="dbm-tab-content">
            <TableData
              v-if="activeTab === 'data'"
              ref="tableDataRef"
              :conn-id="connId"
              :db-type="dbType"
              :db-host="dbHost"
              :db-port="dbPort"
              :db-user="dbUser"
              :db-password="dbPassword"
              :db-database="selectedDb"
              :table-name="selectedTable"
            />
            <TableStructure
              v-if="activeTab === 'structure'"
              :conn-id="connId"
              :db-type="dbType"
              :db-host="dbHost"
              :db-port="dbPort"
              :db-user="dbUser"
              :db-password="dbPassword"
              :db-database="selectedDb"
              :table-name="selectedTable"
            />
            <SqlEditor
              v-if="activeTab === 'sql'"
              :conn-id="connId"
              :db-type="dbType"
              :db-host="dbHost"
              :db-port="dbPort"
              :db-user="dbUser"
              :db-password="dbPassword"
              :db-database="selectedDb"
              :table-name="selectedTable"
            />
            <TableStatus
              v-if="activeTab === 'status'"
              :conn-id="connId"
              :db-type="dbType"
              :db-host="dbHost"
              :db-port="dbPort"
              :db-user="dbUser"
              :db-password="dbPassword"
              :db-database="selectedDb"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { showMessage } from '../../../../utils/message'
import { useConfirm } from '../../../../utils/confirm'
import * as DatabaseService from '../../../../../bindings/changeme/ssh/databaseservice.js'

const { t } = useI18n()
import DbSidebar from './components/DbSidebar.vue'
import TableData from './components/TableData.vue'
import TableStructure from './components/TableStructure.vue'
import SqlEditor from './components/SqlEditor.vue'
import TableStatus from './components/TableStatus.vue'

const connId = inject('connId')
const { confirm } = useConfirm()

// 连接配置
const dbType = ref('mysql')
const dbHost = ref('127.0.0.1')
const dbPort = ref(3306)
const dbUser = ref('root')
const dbPassword = ref('')
const connecting = ref(false)
const connected = ref(false)

// 数据库状态
const databases = ref([])
const dbTables = ref({})
const selectedDb = ref('')
const selectedTable = ref('')
const loadingDbs = ref(false)
const activeTab = ref('data')

const tabs = [
  { id: 'data', label: t('database.data') },
  { id: 'structure', label: t('database.structure') },
  { id: 'sql', label: t('database.sql') },
  { id: 'status', label: t('database.status') },
]

// 连接
async function doConnect() {
  if (!dbHost.value || !dbUser.value) {
    showMessage(t('database.connectionFailed'), 'warning')
    return
  }
  connecting.value = true
  try {
    await DatabaseService.TestConnection(connId, dbType.value, dbHost.value, Number(dbPort.value), dbUser.value, dbPassword.value)
    connected.value = true
    showMessage(t('database.connectionSuccess'), 'success')
    await refreshDatabases()
  } catch (e) {
    showMessage(t('database.connectionFailed') + ': ' + (e?.message || e), 'error')
  } finally {
    connecting.value = false
  }
}

function doDisconnect() {
  connected.value = false
  databases.value = []
  dbTables.value = {}
  selectedDb.value = ''
  selectedTable.value = ''
}

// 数据库操作
async function refreshDatabases() {
  loadingDbs.value = true
  try {
    databases.value = await DatabaseService.ListDatabases(
      connId, dbType.value, dbHost.value, Number(dbPort.value), dbUser.value, dbPassword.value
    ) || []
  } catch (e) {
    showMessage(t('database.loadDbFailed') + ': ' + (e?.message || e), 'error')
  } finally {
    loadingDbs.value = false
  }
}

async function onSelectDb(db) {
  selectedDb.value = db
  selectedTable.value = ''
  if (!dbTables.value[db]) {
    try {
      dbTables.value[db] = await DatabaseService.ListTables(
        connId, dbType.value, dbHost.value, Number(dbPort.value), dbUser.value, dbPassword.value, db
      ) || []
    } catch (e) {
      dbTables.value[db] = []
    }
  }
}

function onSelectTable(db, table) {
  selectedDb.value = db
  selectedTable.value = table
  activeTab.value = 'data'
}

function switchTab(tab) {
  activeTab.value = tab
}

async function dropDatabase(db) {
  const ok = await confirm({ title: t('database.deleteDb'), message: t('database.deleteDbConfirm', { name: db }), danger: true })
  if (!ok) return
  try {
    await DatabaseService.DropDatabase(connId, dbType.value, dbHost.value, Number(dbPort.value), dbUser.value, dbPassword.value, db)
    showMessage(t('database.dbDeleted'), 'success')
    if (selectedDb.value === db) {
      selectedDb.value = ''
      selectedTable.value = ''
    }
    await refreshDatabases()
  } catch (e) {
    showMessage(t('database.dropFailed') + ': ' + (e?.message || e), 'error')
  }
}

onMounted(() => {
  // 自动填充默认端口
  if (dbType.value === 'mysql') dbPort.value = 3306
  else if (dbType.value === 'postgresql') dbPort.value = 5432
})
</script>

<style scoped>
.dbm-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-panel);
  font-size: 13px;
}

/* 连接栏 */
.dbm-conn-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid var(--surface-hover);
  background: var(--toolbar-3);
  flex-shrink: 0;
  gap: 6px;
  flex-wrap: wrap;
}

.conn-fields {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.conn-sel, .conn-inp {
  padding: 5px 8px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 12px;
  outline: none;
}

.conn-sel { min-width: 100px; }
.w-120 { width: 120px; }
.w-100 { width: 100px; }
.w-70 { width: 70px; }

.conn-inp:focus, .conn-sel:focus {
  border-color: var(--accent-primary);
}

.conn-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.dot.green { background: #10b981; }

.sep {
  color: var(--text-muted);
}

/* 欢迎页 */
.dbm-welcome {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.welcome-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.welcome-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin: 0;
}

/* 主布局 */
.dbm-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* 内容区 */
.dbm-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dbm-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-muted);
}

/* 表区域 */
.dbm-table-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 标签页 */
.dbm-tabs {
  display: flex;
  border-bottom: 1px solid var(--surface-hover);
  flex-shrink: 0;
  padding: 0 8px;
}

.dbm-tabs button {
  padding: 8px 16px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-muted);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}

.dbm-tabs button:hover {
  color: var(--text-primary);
}

.dbm-tabs button.active {
  color: var(--accent-primary);
  border-bottom-color: var(--accent-primary);
}

.dbm-tab-content {
  flex: 1;
  overflow: hidden;
}

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border: 1px solid var(--border-strong);
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
  background: var(--surface-hover);
  color: var(--text-secondary);
  white-space: nowrap;
}

.btn:hover { background: var(--bg-toolbar); color: var(--text-primary); }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary { background: var(--accent-primary); border-color: var(--accent-primary); color: white; }
.btn-primary:hover { opacity: 0.9; }
</style>
