<template>
  <div class="st-panel">
    <div class="st-toolbar">
      <button class="btn btn-sm" @click="loadStatus">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
        {{ t('database.refresh') }}
      </button>
      <span class="st-info">{{ dbDatabase }} · {{ t('database.dataOverview') }}</span>
    </div>

    <div v-if="loading" class="st-loading"><div class="spinner"></div><span>{{ t('database.loading') }}</span></div>

    <div v-else class="st-content">
      <!-- 表状态表格 -->
      <div v-if="tableStatuses.length > 0" class="st-table-wrap">
        <table class="st-table">
          <thead>
            <tr>
              <th>{{ t('database.tableLabel') }}</th>
              <th>{{ t('database.tableEngine') }}</th>
              <th>{{ t('database.rowCount') }}</th>
              <th>{{ t('database.tableSize') }}</th>
              <th>{{ t('database.charset') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="ts in tableStatuses" :key="ts.name">
              <td class="st-name">{{ ts.name }}</td>
              <td class="st-engine">{{ ts.engine }}</td>
              <td>{{ ts.rows }}</td>
              <td>{{ ts.size }}</td>
              <td>{{ ts.charset }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 数据库变量 -->
      <div v-if="variables.length > 0" class="st-section">
        <div class="st-section-head">{{ t('database.serverParams') }}</div>
        <div class="st-vars">
          <div v-for="v in variables" :key="v.name" class="st-var">
            <span class="st-var-name">{{ v.name }}</span>
            <span class="st-var-value">{{ v.value }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import * as DatabaseService from '../../../../../../bindings/changeme/ssh/databaseservice.js'

const { t } = useI18n()

const props = defineProps({
  connId: String, dbType: String, dbHost: String, dbPort: Number,
  dbUser: String, dbPassword: String, dbDatabase: String,
})

const loading = ref(false)
const tableStatuses = ref([])
const variables = ref([])

onMounted(() => loadStatus())

async function loadStatus() {
  if (!props.dbDatabase) return
  loading.value = true
  try {
    const result = await DatabaseService.GetTableStatus(
      props.connId, props.dbType, props.dbHost, Number(props.dbPort),
      props.dbUser, props.dbPassword, props.dbDatabase
    )
    const rows = result?.[0] || []
    const cols = result?.[1] || []
    tableStatuses.value = rows.map(row => {
      const obj = {}
      cols.forEach((col, idx) => { obj[col.toLowerCase()] = row[idx] || '' })
      return {
        name: obj.name || obj.table_name || '',
        engine: obj.engine || '-',
        rows: obj.rows || obj.row_count || '-',
        size: obj.data_length ? formatSize(Number(obj.data_length)) : (obj.total_size || '-'),
        charset: obj.charset || obj.character_set || '-',
      }
    })
  } catch (e) {
    tableStatuses.value = []
  } finally {
    loading.value = false
  }
}

function formatSize(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}
</script>

<style scoped>
.st-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.st-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--surface-hover);
  flex-shrink: 0;
}

.st-info {
  font-size: 12px;
  color: var(--text-muted);
}

.st-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.st-table-wrap {
  overflow-x: auto;
  margin-bottom: 16px;
}

.st-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.st-table th {
  padding: 8px 12px;
  background: var(--surface-hover);
  color: var(--text-muted);
  font-weight: 600;
  text-align: left;
  border-bottom: 1px solid var(--border-strong);
}

.st-table td {
  padding: 6px 12px;
  border-bottom: 1px solid var(--surface-hover);
  color: var(--text-secondary);
}

.st-name {
  font-weight: 500;
  color: var(--text-primary);
}

.st-engine {
  font-family: 'Courier New', monospace;
  color: var(--accent-primary);
}

.st-section {
  margin-top: 16px;
}

.st-section-head {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
  padding-bottom: 4px;
  border-bottom: 1px solid var(--surface-hover);
}

.st-vars {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
}

.st-var {
  display: flex;
  justify-content: space-between;
  padding: 4px 8px;
  font-size: 11px;
  background: var(--bg-input);
  border-radius: 4px;
}

.st-var-name {
  color: var(--text-muted);
  font-family: 'Courier New', monospace;
}

.st-var-value {
  color: var(--text-primary);
  font-family: 'Courier New', monospace;
}

.st-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: var(--text-muted);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border: 1px solid var(--border-strong);
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  background: var(--surface-hover);
  color: var(--text-secondary);
}

.btn:hover { background: var(--bg-toolbar); color: var(--text-primary); }
.btn-sm { padding: 3px 8px; font-size: 11px; }

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--surface-hover);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }
</style>
