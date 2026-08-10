<template>
  <div class="ts-panel">
    <div class="ts-toolbar">
      <button class="btn btn-sm" @click="loadStructure">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
        {{ t('database.refresh') }}
      </button>
      <span class="ts-info">{{ tableName }} · {{ t('database.fieldCount', { count: columns.length }) }}</span>
    </div>

    <div v-if="loading" class="ts-loading"><div class="spinner"></div><span>{{ t('database.loading') }}</span></div>

    <div v-else class="ts-table-wrap">
      <table class="ts-table">
        <thead>
          <tr>
            <th>{{ t('database.fieldName') }}</th>
            <th>{{ t('database.dataType') }}</th>
            <th>{{ t('database.nullable') }}</th>
            <th>{{ t('database.defaultValue') }}</th>
            <th>{{ t('database.primaryKey') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="col in columns" :key="col.name">
            <td class="ts-col-name">{{ col.name }}</td>
            <td class="ts-col-type">{{ col.type }}</td>
            <td>
              <span class="ts-tag" :class="col.nullable === 'YES' ? 'ts-tag-ok' : 'ts-tag-no'">
                {{ col.nullable === 'YES' ? 'YES' : 'NO' }}
              </span>
            </td>
            <td class="ts-col-default">{{ col.default || '-' }}</td>
            <td>
              <span v-if="col.isPk" class="ts-tag ts-tag-pk">PK</span>
              <span v-else>-</span>
            </td>
          </tr>
        </tbody>
      </table>
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
  dbUser: String, dbPassword: String, dbDatabase: String, tableName: String,
})

const loading = ref(false)
const columns = ref([])

onMounted(() => loadStructure())

async function loadStructure() {
  if (!props.tableName) return
  loading.value = true
  try {
    const result = await DatabaseService.GetTableStructure(
      props.connId, props.dbType, props.dbHost, Number(props.dbPort),
      props.dbUser, props.dbPassword, props.dbDatabase, props.tableName
    )
    const rows = result?.[0] || []
    const cols = result?.[1] || []
    columns.value = rows.map(row => ({
      name: row[0] || '',
      type: row[1] || '',
      nullable: row[2] || '',
      default: row[3] || '',
      isPk: false,
    }))
  } catch (e) {
    columns.value = []
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.ts-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ts-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--surface-hover);
  flex-shrink: 0;
}

.ts-info {
  font-size: 12px;
  color: var(--text-muted);
}

.ts-table-wrap {
  flex: 1;
  overflow: auto;
}

.ts-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.ts-table th {
  padding: 8px 12px;
  background: var(--surface-hover);
  color: var(--text-muted);
  font-weight: 600;
  text-align: left;
  border-bottom: 1px solid var(--border-strong);
  position: sticky;
  top: 0;
  z-index: 1;
}

.ts-table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--surface-hover);
  color: var(--text-secondary);
}

.ts-col-name {
  font-weight: 500;
  color: var(--text-primary);
}

.ts-col-type {
  font-family: 'Courier New', Consolas, monospace;
  color: var(--accent-primary);
}

.ts-col-default {
  font-family: 'Courier New', Consolas, monospace;
  font-size: 11px;
}

.ts-tag {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
}

.ts-tag-ok { background: #10b98120; color: #10b981; }
.ts-tag-no { background: #ef444420; color: #ef4444; }
.ts-tag-pk { background: #f59e0b20; color: #f59e0b; }

.ts-loading {
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
