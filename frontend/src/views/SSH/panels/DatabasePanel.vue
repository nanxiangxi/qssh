<template>
  <div class="db-panel">
    <!-- 顶部连接栏 -->
    <div class="conn-bar">
      <div class="conn-fields">
        <select v-model="dbType" class="conn-select">
          <option value="mysql">MySQL</option>
          <option value="postgresql">PostgreSQL</option>
          <option value="redis">Redis</option>
        </select>
        <input v-model="dbHost" type="text" :placeholder="t('database.host')" class="conn-input w-100" />
        <input v-model.number="dbPort" type="number" :placeholder="t('database.port')" class="conn-input w-60" />
        <input v-model="dbUser" type="text" :placeholder="t('database.username')" class="conn-input w-100" />
        <input v-model="dbPassword" type="password" :placeholder="t('database.password')" class="conn-input w-100" />
        <button class="btn btn-primary btn-sm" @click="connect" :disabled="connecting">
          {{ connecting ? t('database.connecting') : t('database.connect') }}
        </button>
        <button v-if="connected" class="btn btn-sm" @click="disconnect">{{ t('database.disconnect') }}</button>
      </div>
      <div v-if="connected" class="conn-status">
        <span class="status-dot success"></span>
        <span>{{ dbHost }}:{{ dbPort }}</span>
      </div>
    </div>

    <div v-if="!connected" class="welcome">
      <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" opacity="0.2">
        <ellipse cx="12" cy="5" rx="9" ry="3"/>
        <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
      </svg>
      <p>{{ t('database.welcomeHint') }}</p>
    </div>

    <div v-else class="db-main">
      <!-- 左侧：数据库/表树 -->
      <div class="db-sidebar">
        <div class="sidebar-header">
          <span>{{ t('database.database') }}</span>
          <button class="icon-btn-xs" @click="refreshDatabases" :title="t('database.refresh')">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          </button>
        </div>

        <div v-if="loadingDbs" class="sidebar-loading">
          <div class="spinner-sm"></div>
        </div>

        <div v-else class="db-tree">
          <div
            v-for="db in databases"
            :key="db"
            class="tree-db"
            :class="{ active: selectedDb === db, expanded: expandedDbs[db] }"
          >
            <div class="tree-db-head" @click="toggleDb(db)">
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg>
              <span>{{ db }}</span>
            </div>
            <div v-if="expandedDbs[db]" class="tree-tables">
              <div v-if="loadingTables[db]" class="tree-loading">
                <div class="spinner-sm"></div>
              </div>
              <div
                v-for="table in dbTables[db]"
                :key="table"
                class="tree-table"
                :class="{ active: selectedTable === table && selectedDb === db }"
                @click="selectTable(db, table)"
              >
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/></svg>
                <span>{{ table }}</span>
              </div>
              <div v-if="dbTables[db] && dbTables[db].length === 0" class="tree-empty">{{ t('database.noTable') }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：内容区 -->
      <div class="db-content">
        <!-- 无选中表时显示数据库信息 -->
        <div v-if="!selectedTable" class="db-info-page">
          <div class="info-header">
            <h3>{{ selectedDb }}</h3>
            <button class="btn btn-sm danger" @click="dropDb" :title="t('database.deleteDb')">{{ t('database.deleteDb') }}</button>
          </div>
          <div v-if="dbSize" class="info-card">
            <div class="info-label">{{ t('database.dbSize') }}</div>
            <div class="info-value">{{ dbSize }}</div>
          </div>
          <div class="info-card">
            <div class="info-label">{{ t('database.tableCount') }}</div>
            <div class="info-value">{{ (dbTables[selectedDb] || []).length }}</div>
          </div>
        </div>

        <!-- 选中表后显示标签页 -->
        <div v-else class="table-area">
          <!-- 标签页导航 -->
          <div class="table-tabs">
            <button :class="{ active: activeTab === 'data' }" @click="activeTab = 'data'; loadTableData()">{{ t('database.data') }}</button>
            <button :class="{ active: activeTab === 'structure' }" @click="activeTab = 'structure'; loadTableStructure()">{{ t('database.structure') }}</button>
            <button :class="{ active: activeTab === 'sql' }" @click="activeTab = 'sql'">SQL</button>
            <button :class="{ active: activeTab === 'indexes' }" @click="activeTab = 'indexes'; loadTableIndexes()">{{ t('database.indexes') }}</button>
            <button :class="{ active: activeTab === 'status' }" @click="activeTab = 'status'; loadTableStatus()">{{ t('database.status') }}</button>
            <div class="tab-spacer"></div>
            <div class="table-actions">
              <button class="btn btn-sm" @click="loadTableData" :title="t('database.refresh')">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
              </button>
              <button class="btn btn-sm" @click="showCreateSQL" :title="t('database.createTableSQL')">{{ t('database.createTableSQL') }}</button>
              <button class="btn btn-sm danger" @click="truncateTable" :title="t('database.truncate')">{{ t('database.truncate') }}</button>
              <button class="btn btn-sm danger" @click="dropTable" :title="t('database.dropTable')">{{ t('database.dropTable') }}</button>
            </div>
          </div>

          <!-- 数据标签页 -->
          <div v-if="activeTab === 'data'" class="tab-content">
            <div v-if="loadingData" class="loading-state"><div class="spinner"></div><span>{{ t('database.loading') }}</span></div>
            <div v-else-if="tableData.length === 0" class="empty-state">{{ t('database.noData') }}</div>
            <div v-else class="data-table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th v-for="col in tableColumns" :key="col">{{ col }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in tableData" :key="idx">
                    <td v-for="col in tableColumns" :key="col" :title="row[col]">
                      <span class="cell-value">{{ formatCell(row[col]) }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <!-- 分页 -->
            <div v-if="tableData.length > 0" class="pagination">
              <div class="page-info">
                <select v-model="pageSize" class="page-select" @change="loadTableData">
                  <option :value="25">{{ t('database.perPage', { n: 25 }) }}</option>
                  <option :value="50">{{ t('database.perPage', { n: 50 }) }}</option>
                  <option :value="100">{{ t('database.perPage', { n: 100 }) }}</option>
                  <option :value="500">{{ t('database.perPage', { n: 500 }) }}</option>
                </select>
                <span>{{ t('database.totalRows', { count: totalRows }) }}</span>
              </div>
              <div class="page-nav">
                <button class="page-btn" :disabled="currentPage <= 1" @click="currentPage--; loadTableData()">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
                </button>
                <span class="page-current">{{ currentPage }} / {{ totalPages }}</span>
                <button class="page-btn" :disabled="currentPage >= totalPages" @click="currentPage++; loadTableData()">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
                </button>
              </div>
            </div>
          </div>

          <!-- 结构标签页 -->
          <div v-if="activeTab === 'structure'" class="tab-content">
            <div v-if="loadingStructure" class="loading-state"><div class="spinner"></div><span>{{ t('database.loading') }}</span></div>
            <div v-else-if="structureData.length === 0" class="empty-state">{{ t('database.noStructure') }}</div>
            <div v-else class="data-table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th v-for="col in structureColumns" :key="col">{{ col }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in structureData" :key="idx">
                    <td v-for="col in structureColumns" :key="col">{{ row[col] }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- SQL 标签页 -->
          <div v-if="activeTab === 'sql'" class="tab-content sql-tab">
            <div class="sql-editor">
              <textarea v-model="sqlQuery" :placeholder="t('database.enterSQL')" rows="6" @keydown.ctrl.enter="executeSQL"></textarea>
              <div class="sql-toolbar">
                <button class="btn btn-primary btn-sm" @click="executeSQL" :disabled="sqlLoading">
                  {{ sqlLoading ? t('database.executing') : t('database.executeSQL') }}
                </button>
                <div class="sql-actions">
                  <button class="btn btn-sm" @click="sqlQuery = `SELECT * FROM ${selectedTable} LIMIT 100`">{{ t('database.queryAll') }}</button>
                  <button class="btn btn-sm" @click="sqlQuery = `SELECT COUNT(*) AS total FROM ${selectedTable}`">{{ t('database.count') }}</button>
                  <button class="btn btn-sm" @click="sqlQuery = `SHOW CREATE TABLE ${selectedTable}`">{{ t('database.createTableStmt') }}</button>
                </div>
              </div>
            </div>
            <div v-if="sqlResult" class="sql-result">
              <div class="sql-result-header">
                <span>{{ t('database.queryResult') }}</span>
                <button class="btn btn-sm" @click="copySqlResult">{{ t('database.copy') }}</button>
              </div>
              <pre class="sql-result-content">{{ sqlResult }}</pre>
            </div>
          </div>

          <!-- 索引标签页 -->
          <div v-if="activeTab === 'indexes'" class="tab-content">
            <div v-if="loadingIndexes" class="loading-state"><div class="spinner"></div><span>{{ t('database.loading') }}</span></div>
            <pre v-else class="raw-output">{{ indexData || t('database.noIndex') }}</pre>
          </div>

          <!-- 状态标签页 -->
          <div v-if="activeTab === 'status'" class="tab-content">
            <div v-if="loadingStatus" class="loading-state"><div class="spinner"></div><span>{{ t('database.loading') }}</span></div>
            <pre v-else class="raw-output">{{ statusData || t('database.noStatus') }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { showMessage } from '../../../utils/message'
import { useConfirm } from '../../../utils/confirm'
import * as DatabaseService from '../../../../bindings/changeme/ssh/databaseservice.js'

const { t } = useI18n()
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

// 数据库树
const databases = ref([])
const expandedDbs = ref({})
const dbTables = ref({})
const loadingDbs = ref(false)
const loadingTables = ref({})

// 选中状态
const selectedDb = ref('')
const selectedTable = ref('')
const activeTab = ref('data')

// 数据
const tableData = ref([])
const tableColumns = ref([])
const totalRows = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)
const loadingData = ref(false)

// 结构
const structureData = ref([])
const structureColumns = ref([])
const loadingStructure = ref(false)

// SQL
const sqlQuery = ref('')
const sqlResult = ref('')
const sqlLoading = ref(false)

// 索引/状态
const indexData = ref('')
const statusData = ref('')
const loadingIndexes = ref(false)
const loadingStatus = ref(false)

// 数据库信息
const dbSize = ref('')

const totalPages = computed(() => Math.max(1, Math.ceil(totalRows.value / pageSize.value)))

// 连接
async function connect() {
  if (!dbHost.value || !dbUser.value) {
    showMessage(t('database.pleaseFillHostUser'), 'warning')
    return
  }
  connecting.value = true
  try {
    console.log('[DB] connect:', { connId, dbType: dbType.value, dbHost: dbHost.value, dbPort: Number(dbPort.value), dbUser: dbUser.value })
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

function disconnect() {
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
    console.log('[DB] refreshDatabases:', { connId, dbType: dbType.value, dbHost: dbHost.value, dbPort: Number(dbPort.value), dbUser: dbUser.value })
    databases.value = await DatabaseService.ListDatabases(
      connId, dbType.value, dbHost.value, Number(dbPort.value), dbUser.value, dbPassword.value
    ) || []
    console.log('[DB] databases:', databases.value)
    // 自动选中第一个数据库
    if (databases.value.length > 0 && !selectedDb.value) {
      const firstDb = databases.value[0]
      console.log('[DB] auto-select:', firstDb)
      if (firstDb && firstDb.trim()) {
        await toggleDb(firstDb)
      }
    }
  } catch (e) {
    console.error('[DB] refreshDatabases error:', e)
    showMessage(t('database.loadDbFailed') + ': ' + (e?.message || e), 'error')
  } finally {
    loadingDbs.value = false
  }
}

async function toggleDb(db) {
  if (!db || !db.trim()) {
    console.warn('[DB] toggleDb: db is empty')
    return
  }
  console.log('[DB] toggleDb:', db)
  if (expandedDbs.value[db]) {
    expandedDbs.value[db] = false
    return
  }
  expandedDbs.value[db] = true
  selectedDb.value = db
  selectedTable.value = ''
  activeTab.value = 'data'

  if (!dbTables.value[db]) {
    loadingTables.value[db] = true
    try {
      console.log('[DB] ListTables:', { connId, dbType: dbType.value, dbHost: dbHost.value, dbPort: Number(dbPort.value), dbUser: dbUser.value, db })
      dbTables.value[db] = await DatabaseService.ListTables(
        connId, dbType.value, dbHost.value, Number(dbPort.value), dbUser.value, dbPassword.value, db
      ) || []
    } catch (e) {
      dbTables.value[db] = []
    } finally {
      loadingTables.value[db] = false
    }
  }

  // 加载数据库大小
  try {
    dbSize.value = await DatabaseService.GetDatabaseSize(
      connId, dbType.value, dbHost.value, Number(dbPort.value), dbUser.value, dbPassword.value, db
    ) || ''
  } catch {}
}

function selectTable(db, table) {
  selectedDb.value = db
  selectedTable.value = table
  activeTab.value = 'data'
  currentPage.value = 1
  loadTableData()
}

// 数据加载
async function loadTableData() {
  if (!selectedDb.value || !selectedTable.value) return
  loadingData.value = true
  try {
    const result = await DatabaseService.GetTableData(
      connId, dbType.value, dbHost.value, Number(dbPort.value),
      dbUser.value, dbPassword.value, selectedDb.value, selectedTable.value,
      currentPage.value, pageSize.value
    )
    // Wails 绑定返回数组：[rows, columns]
    const rows = result?.[0] || []
    const cols = result?.[1] || []
    tableColumns.value = cols
    tableData.value = rows.map(row => {
      const obj = {}
      cols.forEach((col, idx) => {
        obj[col] = row[idx] !== undefined ? row[idx] : ''
      })
      return obj
    })

    // 加载总数
    try {
      totalRows.value = await DatabaseService.GetTableCount(
        connId, dbType.value, dbHost.value, Number(dbPort.value),
        dbUser.value, dbPassword.value, selectedDb.value, selectedTable.value
      ) || 0
    } catch {}
  } catch (e) {
    showMessage(t('database.loadDataFailed') + ': ' + (e?.message || e), 'error')
  } finally {
    loadingData.value = false
  }
}

function formatCell(val) {
  if (val === null || val === undefined || val === 'NULL') return 'NULL'
  if (typeof val === 'string' && val.length > 100) return val.substring(0, 100) + '...'
  return val
}

// 结构加载
async function loadTableStructure() {
  if (!selectedDb.value || !selectedTable.value) return
  loadingStructure.value = true
  try {
    const result = await DatabaseService.GetTableStructure(
      connId, dbType.value, dbHost.value, Number(dbPort.value),
      dbUser.value, dbPassword.value, selectedDb.value, selectedTable.value
    )
    const rows = result?.[0] || []
    const cols = result?.[1] || []
    structureColumns.value = cols
    structureData.value = rows.map(row => {
      const obj = {}
      cols.forEach((col, idx) => {
        obj[col] = row[idx] !== undefined ? row[idx] : ''
      })
      return obj
    })
  } catch (e) {
    showMessage(t('database.loadStructureFailed') + ': ' + (e?.message || e), 'error')
  } finally {
    loadingStructure.value = false
  }
}

// 索引/状态
async function loadTableIndexes() {
  if (!selectedDb.value || !selectedTable.value) return
  loadingIndexes.value = true
  try {
    const result = await DatabaseService.GetTableIndexes(
      connId, dbType.value, dbHost.value, Number(dbPort.value),
      dbUser.value, dbPassword.value, selectedDb.value, selectedTable.value
    )
    const rows = result?.[0] || []
    const cols = result?.[1] || []
    if (cols.length > 0) {
      const header = cols.join('\t')
      const data = rows.map(r => r.join('\t')).join('\n')
      indexData.value = header + '\n' + data
    } else {
      indexData.value = t('database.noIndex')
    }
  } catch (e) {
    indexData.value = t('database.loadFailed') + ': ' + (e?.message || e)
  } finally {
    loadingIndexes.value = false
  }
}

async function loadTableStatus() {
  if (!selectedDb.value) return
  loadingStatus.value = true
  try {
    const result = await DatabaseService.GetTableStatus(
      connId, dbType.value, dbHost.value, Number(dbPort.value),
      dbUser.value, dbPassword.value, selectedDb.value
    )
    const rows = result?.[0] || []
    const cols = result?.[1] || []
    if (cols.length > 0) {
      const header = cols.join('\t')
      const data = rows.map(r => r.join('\t')).join('\n')
      statusData.value = header + '\n' + data
    } else {
      statusData.value = t('database.noStatus')
    }
  } catch (e) {
    statusData.value = t('database.loadFailed') + ': ' + (e?.message || e)
  } finally {
    loadingStatus.value = false
  }
}

// SQL 执行
async function executeSQL() {
  if (!sqlQuery.value.trim()) return
  sqlLoading.value = true
  sqlResult.value = ''
  try {
    const result = await DatabaseService.ExecuteQuery(
      connId, dbType.value, dbHost.value, Number(dbPort.value),
      dbUser.value, dbPassword.value, selectedDb.value, sqlQuery.value
    )
    const rows = result?.[0] || []
    const cols = result?.[1] || []
    if (cols.length > 0) {
      const header = cols.join('\t')
      const data = rows.map(r => r.join('\t')).join('\n')
      sqlResult.value = header + '\n' + data
    } else {
      sqlResult.value = t('database.emptyResult')
    }
  } catch (e) {
    sqlResult.value = t('database.sqlError') + ': ' + (e?.message || e)
  } finally {
    sqlLoading.value = false
  }
}

function copySqlResult() {
  navigator.clipboard.writeText(sqlResult.value)
  showMessage(t('database.copied'), 'success')
}

// 表操作
async function showCreateSQL() {
  activeTab.value = 'sql'
  sqlQuery.value = `SHOW CREATE TABLE ${selectedTable.value}`
  await executeSQL()
}

async function truncateTable() {
  const ok = await confirm({ title: t('database.truncateTable'), message: t('database.truncateConfirm', { name: selectedTable.value }), danger: true })
  if (!ok) return
  try {
    await DatabaseService.TruncateTable(
      connId, dbType.value, dbHost.value, Number(dbPort.value),
      dbUser.value, dbPassword.value, selectedDb.value, selectedTable.value
    )
    showMessage(t('database.tableTruncated'), 'success')
    await loadTableData()
  } catch (e) {
    showMessage(t('database.truncateFailed') + ': ' + (e?.message || e), 'error')
  }
}

async function dropTable() {
  const ok = await confirm({ title: t('database.dropTable'), message: t('database.dropTableConfirm', { name: selectedTable.value }), danger: true })
  if (!ok) return
  try {
    await DatabaseService.DropTable(
      connId, dbType.value, dbHost.value, Number(dbPort.value),
      dbUser.value, dbPassword.value, selectedDb.value, selectedTable.value
    )
    showMessage(t('database.tableDropped'), 'success')
    selectedTable.value = ''
    delete dbTables.value[selectedDb.value]
    await toggleDb(selectedDb.value)
  } catch (e) {
    showMessage(t('database.dropFailed') + ': ' + (e?.message || e), 'error')
  }
}

async function dropDb() {
  const ok = await confirm({ title: t('database.deleteDb'), message: t('database.deleteDbConfirm2', { name: selectedDb.value }), danger: true })
  if (!ok) return
  try {
    await DatabaseService.DropDatabase(
      connId, dbType.value, dbHost.value, Number(dbPort.value),
      dbUser.value, dbPassword.value, selectedDb.value
    )
    showMessage(t('database.dbDeleted'), 'success')
    selectedDb.value = ''
    selectedTable.value = ''
    await refreshDatabases()
  } catch (e) {
    showMessage(t('database.dropFailed') + ': ' + (e?.message || e), 'error')
  }
}
</script>

<style scoped>
.db-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-panel);
  font-size: 13px;
}

/* 连接栏 */
.conn-bar {
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

.conn-select, .conn-input {
  padding: 5px 8px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 12px;
  outline: none;
}

.conn-select { min-width: 90px; }
.w-100 { width: 100px; }
.w-60 { width: 60px; }

.conn-input:focus, .conn-select:focus {
  border-color: var(--accent-primary);
}

.conn-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.status-dot.success { background: #10b981; }

/* 欢迎页 */
.welcome {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-muted);
}

/* 主布局 */
.db-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* 左侧树 */
.db-sidebar {
  width: 200px;
  border-right: 1px solid var(--surface-hover);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  flex-shrink: 0;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  border-bottom: 1px solid var(--surface-hover);
}

.sidebar-loading {
  display: flex;
  justify-content: center;
  padding: 20px;
}

.db-tree {
  flex: 1;
  overflow-y: auto;
}

.tree-db-head {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 12px;
  transition: background 0.15s;
}

.tree-db-head:hover {
  background: var(--surface-hover);
}

.tree-db.active > .tree-db-head {
  color: var(--accent-primary);
  background: var(--accent-primary-bg);
}

.tree-db-head svg:first-child {
  transition: transform 0.2s;
}

.tree-db.expanded > .tree-db-head svg:first-child {
  transform: rotate(90deg);
}

.tree-tables {
  padding-left: 16px;
}

.tree-table {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 12px;
  transition: background 0.15s;
}

.tree-table:hover {
  background: var(--surface-hover);
}

.tree-table.active {
  color: var(--accent-primary);
  background: var(--accent-primary-bg);
}

.tree-loading, .tree-empty {
  padding: 8px;
  text-align: center;
  font-size: 11px;
  color: var(--text-muted);
}

/* 右侧内容 */
.db-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 数据库信息页 */
.db-info-page {
  padding: 20px;
}

.info-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.info-header h3 {
  margin: 0;
  color: var(--text-primary);
}

.info-card {
  padding: 12px 16px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 8px;
  margin-bottom: 8px;
}

.info-label {
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.info-value {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

/* 表区域 */
.table-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 标签页 */
.table-tabs {
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--surface-hover);
  flex-shrink: 0;
  padding: 0 8px;
}

.table-tabs button {
  padding: 8px 14px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-muted);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}

.table-tabs button:hover {
  color: var(--text-primary);
}

.table-tabs button.active {
  color: var(--accent-primary);
  border-bottom-color: var(--accent-primary);
}

.tab-spacer { flex: 1; }

.table-actions {
  display: flex;
  gap: 4px;
  padding: 4px 0;
}

.tab-content {
  flex: 1;
  overflow: auto;
  padding: 8px;
}

/* 数据表格 */
.data-table-wrap {
  overflow: auto;
  border: 1px solid var(--surface-hover);
  border-radius: 6px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.data-table th {
  padding: 8px 10px;
  background: var(--surface-hover);
  color: var(--text-muted);
  font-weight: 600;
  text-align: left;
  border-bottom: 1px solid var(--border-strong);
  white-space: nowrap;
  position: sticky;
  top: 0;
  z-index: 1;
}

.data-table td {
  padding: 6px 10px;
  border-bottom: 1px solid var(--surface-hover);
  color: var(--text-secondary);
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.data-table tr:hover td {
  background: var(--bg-input);
}

.cell-value {
  font-family: 'Courier New', Consolas, monospace;
  font-size: 11px;
}

/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  flex-shrink: 0;
}

.page-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.page-select {
  padding: 3px 6px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 11px;
  outline: none;
}

.page-nav {
  display: flex;
  align-items: center;
  gap: 6px;
}

.page-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-hover);
  border: 1px solid var(--border-strong);
  border-radius: 4px;
  color: var(--text-secondary);
  cursor: pointer;
}

.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.page-current {
  font-size: 12px;
  color: var(--text-secondary);
  min-width: 50px;
  text-align: center;
}

/* SQL 标签页 */
.sql-tab {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sql-editor textarea {
  width: 100%;
  padding: 10px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 6px;
  color: var(--text-primary);
  font-family: 'Courier New', Consolas, monospace;
  font-size: 12px;
  resize: vertical;
  outline: none;
}

.sql-editor textarea:focus {
  border-color: var(--accent-primary);
}

.sql-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.sql-actions {
  display: flex;
  gap: 4px;
}

.sql-result {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.sql-result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  background: var(--surface-hover);
  font-size: 12px;
  color: var(--text-secondary);
}

.sql-result-content {
  flex: 1;
  padding: 10px;
  margin: 0;
  background: var(--bg-input);
  font-family: 'Courier New', Consolas, monospace;
  font-size: 11px;
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-break: break-all;
  overflow: auto;
}

.raw-output {
  padding: 10px;
  margin: 0;
  background: var(--bg-input);
  border-radius: 6px;
  font-family: 'Courier New', Consolas, monospace;
  font-size: 11px;
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-break: break-all;
  overflow: auto;
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
.btn-sm { padding: 3px 8px; font-size: 11px; }
.btn-primary { background: var(--accent-primary); border-color: var(--accent-primary); color: white; }
.btn-primary:hover { opacity: 0.9; }
.btn.danger { color: var(--accent-danger); border-color: var(--accent-danger); }
.btn.danger:hover { background: var(--accent-danger-bg); }

.icon-btn-xs {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-muted);
  cursor: pointer;
}

.icon-btn-xs:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

/* 加载/空状态 */
.loading-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: var(--text-muted);
  font-size: 13px;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--surface-hover);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid var(--surface-hover);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
