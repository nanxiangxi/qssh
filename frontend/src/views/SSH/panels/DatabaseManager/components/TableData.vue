<template>
  <div class="td-panel">
    <!-- 工具栏 -->
    <div class="td-toolbar">
      <div class="td-toolbar-left">
        <button class="btn btn-sm" @click="loadData" :title="t('database.refreshData')">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          {{ t('database.refresh') }}
        </button>
        <button class="btn btn-sm" @click="insertRow" :title="t('database.insertNewRow')">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          {{ t('database.insertNewRow') }}
        </button>
        <button class="btn btn-sm danger" @click="deleteSelected" :disabled="selectedRows.size === 0" :title="t('database.deleteSelected')">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          {{ t('database.delete') }} ({{ selectedRows.size }})
        </button>
        <div class="td-sep"></div>
        <span class="td-info">{{ tableName }} · {{ t('database.totalRows', { count: totalRows }) }}</span>
      </div>
      <div class="td-toolbar-right">
        <input v-model="filterText" type="text" :placeholder="t('database.searchFilter')" class="td-filter" @keyup.enter="applyFilter" />
        <select v-model="pageSize" class="td-pagesel" @change="currentPage = 1; loadData()">
          <option :value="25">{{ t('database.perPage', { n: 25 }) }}</option>
          <option :value="50">{{ t('database.perPage', { n: 50 }) }}</option>
          <option :value="100">{{ t('database.perPage', { n: 100 }) }}</option>
          <option :value="500">{{ t('database.perPage', { n: 500 }) }}</option>
        </select>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="td-table-wrap">
      <div v-if="loading" class="td-loading">
        <div class="spinner"></div>
        <span>{{ t('database.loading') }}</span>
      </div>

      <div v-else-if="columns.length === 0" class="td-empty">
        <p>{{ t('database.noDataInTable') }}</p>
      </div>

      <table v-else class="td-table">
        <thead>
          <tr>
            <th class="td-th-check">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" />
            </th>
            <th v-for="col in columns" :key="col" class="td-th" @click="toggleSort(col)">
              <span>{{ col }}</span>
              <svg v-if="sortCol === col" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline v-if="sortDir === 'asc'" points="18 15 12 9 6 15"/>
                <polyline v-else points="6 9 12 15 18 9"/>
              </svg>
            </th>
            <th class="td-th-action">{{ t('database.action') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, idx) in filteredRows" :key="idx" :class="{ selected: selectedRows.has(idx), editing: editingRow === idx }">
            <td class="td-td-check">
              <input type="checkbox" :checked="selectedRows.has(idx)" @change="toggleSelect(idx)" />
            </td>
            <td v-for="col in columns" :key="col" class="td-td" @dblclick="startEdit(idx, col, row[col])">
              <template v-if="editingCell.row === idx && editingCell.col === col">
                <input
                  v-model="editingCell.value"
                  class="td-edit-input"
                  @keydown.enter="saveEdit(idx, col)"
                  @keydown.escape="cancelEdit"
                  @blur="saveEdit(idx, col)"
                  ref="editInput"
                />
              </template>
              <template v-else>
                <span class="td-cell" :class="{ null: row[col] === 'NULL' }">{{ formatCell(row[col]) }}</span>
              </template>
            </td>
            <td class="td-td-action">
              <button class="td-btn-icon" @click="deleteRow(idx)" :title="t('database.delete')">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div v-if="totalRows > 0" class="td-pagination">
      <div class="page-info">
        <span>{{ t('database.totalRows', { count: totalRows }) }}</span>
      </div>
      <div class="page-nav">
        <button class="page-btn" :disabled="currentPage <= 1" @click="currentPage--; loadData()">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
        </button>
        <span class="page-current">{{ currentPage }} / {{ totalPages }}</span>
        <button class="page-btn" :disabled="currentPage >= totalPages" @click="currentPage++; loadData()">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
        </button>
      </div>
    </div>

    <!-- 插入行对话框 -->
    <div v-if="showInsert" class="modal-mask" @click.self="showInsert = false">
      <div class="modal-box">
        <div class="modal-header">{{ t('database.insertNewRecord') }}</div>
        <div class="modal-body">
          <div v-for="col in columns" :key="col" class="form-group">
            <label>{{ col }}</label>
            <input v-model="insertData[col]" type="text" :placeholder="col" class="form-input" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showInsert = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" @click="confirmInsert">{{ t('database.confirmInsert') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { showMessage } from '../../../../../utils/message'
import { useConfirm } from '../../../../../utils/confirm'
import * as DatabaseService from '../../../../../../bindings/changeme/ssh/databaseservice.js'

const props = defineProps({
  connId: String,
  dbType: String,
  dbHost: String,
  dbPort: Number,
  dbUser: String,
  dbPassword: String,
  dbDatabase: String,
  tableName: String,
})

const { t } = useI18n()
const { confirm } = useConfirm()

const loading = ref(false)
const columns = ref([])
const rows = ref([])
const totalRows = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)
const filterText = ref('')
const sortCol = ref('')
const sortDir = ref('asc')
const selectedRows = ref(new Set())
const editingRow = ref(-1)
const editingCell = ref({ row: -1, col: '', value: '' })
const showInsert = ref(false)
const insertData = ref({})

const totalPages = computed(() => Math.max(1, Math.ceil(totalRows.value / pageSize.value)))
const isAllSelected = computed(() => rows.value.length > 0 && selectedRows.value.size === rows.value.length)

const filteredRows = computed(() => {
  let data = [...rows.value]
  if (filterText.value) {
    const q = filterText.value.toLowerCase()
    data = data.filter(row => Object.values(row).some(v => String(v).toLowerCase().includes(q)))
  }
  if (sortCol.value) {
    data.sort((a, b) => {
      const va = a[sortCol.value] ?? ''
      const vb = b[sortCol.value] ?? ''
      const cmp = String(va).localeCompare(String(vb), undefined, { numeric: true })
      return sortDir.value === 'asc' ? cmp : -cmp
    })
  }
  return data
})

onMounted(() => loadData())

async function loadData() {
  if (!props.tableName) return
  loading.value = true
  selectedRows.value.clear()
  try {
    const result = await DatabaseService.GetTableData(
      props.connId, props.dbType, props.dbHost, Number(props.dbPort),
      props.dbUser, props.dbPassword, props.dbDatabase, props.tableName,
      currentPage.value, pageSize.value
    )
    const r = result?.[0] || []
    const c = result?.[1] || []
    columns.value = c
    rows.value = r.map(row => {
      const obj = {}
      c.forEach((col, idx) => { obj[col] = row[idx] !== undefined ? row[idx] : '' })
      return obj
    })

    try {
      totalRows.value = await DatabaseService.GetTableCount(
        props.connId, props.dbType, props.dbHost, Number(props.dbPort),
        props.dbUser, props.dbPassword, props.dbDatabase, props.tableName
      ) || 0
    } catch {}
  } catch (e) {
    showMessage(t('database.loadDataFailed') + ': ' + (e?.message || e), 'error')
  } finally {
    loading.value = false
  }
}

function formatCell(val) {
  if (val === null || val === undefined || val === 'NULL') return 'NULL'
  if (typeof val === 'string' && val.length > 80) return val.substring(0, 80) + '...'
  return val
}

function toggleSort(col) {
  if (sortCol.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortCol.value = col
    sortDir.value = 'asc'
  }
}

function toggleSelect(idx) {
  if (selectedRows.value.has(idx)) {
    selectedRows.value.delete(idx)
  } else {
    selectedRows.value.add(idx)
  }
  selectedRows.value = new Set(selectedRows.value)
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedRows.value.clear()
  } else {
    selectedRows.value = new Set(rows.value.map((_, i) => i))
  }
}

// 编辑单元格
function startEdit(row, col, value) {
  editingCell.value = { row, col, value: value === 'NULL' ? '' : String(value) }
  nextTick(() => {
    const input = document.querySelector('.td-edit-input')
    if (input) input.focus()
  })
}

function cancelEdit() {
  editingCell.value = { row: -1, col: '', value: '' }
}

async function saveEdit(row, col) {
  const newVal = editingCell.value.value
  const oldVal = rows.value[row][col]
  cancelEdit()

  if (newVal === oldVal || (newVal === '' && oldVal === 'NULL')) return

  // 构建 UPDATE 语句
  const pkCol = columns.value[0] // 假设第一列是主键
  const pkVal = rows.value[row][pkCol]
  const escapedVal = newVal === '' ? 'NULL' : `'${newVal.replace(/'/g, "''")}'`
  const query = `UPDATE \`${props.tableName}\` SET \`${col}\` = ${escapedVal} WHERE \`${pkCol}\` = '${String(pkVal).replace(/'/g, "''")}'`

  try {
    await DatabaseService.ExecuteQuery(
      props.connId, props.dbType, props.dbHost, Number(props.dbPort),
      props.dbUser, props.dbPassword, props.dbDatabase, query
    )
    rows.value[row][col] = newVal === '' ? 'NULL' : newVal
    showMessage(t('database.recordUpdated'), 'success')
  } catch (e) {
    showMessage(t('database.updateFailed') + ': ' + (e?.message || e), 'error')
  }
}

// 插入行
function insertRow() {
  insertData.value = {}
  columns.value.forEach(col => { insertData.value[col] = '' })
  showInsert.value = true
}

async function confirmInsert() {
  const cols = columns.value.filter(c => insertData.value[c] !== '' && insertData.value[c] !== undefined)
  const vals = cols.map(c => {
    const v = insertData.value[c]
    return v === 'NULL' || v === 'null' ? 'NULL' : `'${String(v).replace(/'/g, "''")}'`
  })
  const query = `INSERT INTO \`${props.tableName}\` (\`${cols.join('`, `')}\`) VALUES (${vals.join(', ')})`

  try {
    await DatabaseService.ExecuteQuery(
      props.connId, props.dbType, props.dbHost, Number(props.dbPort),
      props.dbUser, props.dbPassword, props.dbDatabase, query
    )
    showInsert.value = false
    showMessage(t('database.recordInserted'), 'success')
    await loadData()
  } catch (e) {
    showMessage(t('database.insertFailed') + ': ' + (e?.message || e), 'error')
  }
}

// 删除行
async function deleteRow(idx) {
  const pkCol = columns.value[0]
  const pkVal = rows.value[idx][pkCol]
  const ok = await confirm({ title: t('database.deleteRecord'), message: t('database.deleteRecordConfirm', { pk: pkCol, val: pkVal }), danger: true })
  if (!ok) return

  const query = `DELETE FROM \`${props.tableName}\` WHERE \`${pkCol}\` = '${String(pkVal).replace(/'/g, "''")}'`
  try {
    await DatabaseService.ExecuteQuery(
      props.connId, props.dbType, props.dbHost, Number(props.dbPort),
      props.dbUser, props.dbPassword, props.dbDatabase, query
    )
    showMessage(t('database.recordDeleted'), 'success')
    await loadData()
  } catch (e) {
    showMessage(t('database.deleteFailed') + ': ' + (e?.message || e), 'error')
  }
}

async function deleteSelected() {
  if (selectedRows.value.size === 0) return
  const ok = await confirm({ title: t('database.batchDelete'), message: t('database.batchDeleteConfirm', { count: selectedRows.value.size }), danger: true })
  if (!ok) return

  const pkCol = columns.value[0]
  const ids = [...selectedRows.value].map(idx => `'${String(rows.value[idx][pkCol]).replace(/'/g, "''")}'`)
  const query = `DELETE FROM \`${props.tableName}\` WHERE \`${pkCol}\` IN (${ids.join(', ')})`

  try {
    await DatabaseService.ExecuteQuery(
      props.connId, props.dbType, props.dbHost, Number(props.dbPort),
      props.dbUser, props.dbPassword, props.dbDatabase, query
    )
    showMessage(t('database.batchDeleted', { count: selectedRows.value.size }), 'success')
    selectedRows.value.clear()
    await loadData()
  } catch (e) {
    showMessage(t('database.deleteFailed') + ': ' + (e?.message || e), 'error')
  }
}

function applyFilter() {}
</script>

<style scoped>
.td-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 工具栏 */
.td-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid var(--surface-hover);
  flex-shrink: 0;
  gap: 8px;
  flex-wrap: wrap;
}

.td-toolbar-left, .td-toolbar-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.td-sep {
  width: 1px;
  height: 16px;
  background: var(--surface-hover);
}

.td-info {
  font-size: 12px;
  color: var(--text-muted);
}

.td-filter {
  padding: 4px 8px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 12px;
  width: 120px;
  outline: none;
}

.td-filter:focus {
  border-color: var(--accent-primary);
}

.td-pagesel {
  padding: 4px 6px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 11px;
  outline: none;
}

/* 表格 */
.td-table-wrap {
  flex: 1;
  overflow: auto;
}

.td-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.td-th {
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
  cursor: pointer;
  user-select: none;
}

.td-th:hover {
  background: var(--bg-toolbar);
}

.td-th-check, .td-td-check {
  width: 36px;
  text-align: center;
}

.td-th-action, .td-td-action {
  width: 60px;
  text-align: center;
}

.td-td {
  padding: 6px 10px;
  border-bottom: 1px solid var(--surface-hover);
  color: var(--text-secondary);
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.td-td:hover {
  background: var(--bg-input);
}

tr.selected .td-td {
  background: var(--accent-primary-bg);
}

.td-cell {
  font-family: 'Courier New', Consolas, monospace;
  font-size: 11px;
}

.td-cell.null {
  color: var(--text-muted);
  font-style: italic;
}

/* 编辑输入 */
.td-edit-input {
  width: 100%;
  padding: 2px 4px;
  background: var(--bg-input);
  border: 1px solid var(--accent-primary);
  border-radius: 3px;
  color: var(--text-primary);
  font-family: 'Courier New', Consolas, monospace;
  font-size: 12px;
  outline: none;
}

/* 操作按钮 */
.td-btn-icon {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-muted);
  cursor: pointer;
}

.td-btn-icon:hover {
  background: var(--accent-danger-bg);
  color: var(--accent-danger);
}

/* 分页 */
.td-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-top: 1px solid var(--surface-hover);
  flex-shrink: 0;
}

.page-info {
  font-size: 12px;
  color: var(--text-muted);
}

.page-nav {
  display: flex;
  align-items: center;
  gap: 8px;
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

/* 加载/空状态 */
.td-loading, .td-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: var(--text-muted);
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

/* 模态框 */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10001;
}

.modal-box {
  background: var(--bg-panel);
  border: 1px solid var(--surface-hover);
  border-radius: 12px;
  width: 500px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.modal-header {
  padding: 14px 18px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 1px solid var(--surface-hover);
}

.modal-body {
  padding: 18px;
  overflow-y: auto;
  max-height: 60vh;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 10px;
}

.form-group label {
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 500;
}

.form-input {
  padding: 6px 10px;
  background: var(--bg-input);
  border: 1px solid var(--surface-hover);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 12px;
  outline: none;
}

.form-input:focus {
  border-color: var(--accent-primary);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 18px;
  border-top: 1px solid var(--surface-hover);
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--surface-hover);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
