<template>
  <div class="sql-panel">
    <!-- 编辑器 -->
    <div class="sql-editor">
      <div class="sql-editor-head">
        <span>{{ t('database.sqlEditor') }}</span>
        <div class="sql-editor-actions">
          <button class="btn btn-sm" @click="loadTemplate('select')">SELECT</button>
          <button class="btn btn-sm" @click="loadTemplate('insert')">INSERT</button>
          <button class="btn btn-sm" @click="loadTemplate('update')">UPDATE</button>
          <button class="btn btn-sm" @click="loadTemplate('delete')">DELETE</button>
          <button class="btn btn-sm" @click="loadTemplate('create')">CREATE</button>
        </div>
      </div>
      <textarea
        v-model="query"
        class="sql-textarea"
        :placeholder="t('database.enterSQL')"
        rows="8"
        @keydown.ctrl.enter="execute"
        @keydown.meta.enter="execute"
      ></textarea>
      <div class="sql-toolbar">
        <button class="btn btn-primary" @click="execute" :disabled="executing">
          {{ executing ? t('database.executing') : t('database.executeSQL') }}
        </button>
        <div class="sql-meta">
          <span v-if="resultRows !== null">{{ t('database.resultInfo', { count: resultRows, time: resultTime }) }}</span>
        </div>
      </div>
    </div>

    <!-- 结果 -->
    <div v-if="resultColumns.length > 0" class="sql-result">
      <div class="sql-result-head">
        <span>{{ t('database.queryResult') }}</span>
        <button class="btn btn-sm" @click="copyResult">{{ t('database.copyResult') }}</button>
      </div>
      <div class="sql-result-table-wrap">
        <table class="sql-result-table">
          <thead>
            <tr>
              <th v-for="col in resultColumns" :key="col">{{ col }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, idx) in resultRows_data" :key="idx">
              <td v-for="(val, ci) in row" :key="ci" :class="{ null: val === 'NULL' }">
                {{ val }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 错误 -->
    <div v-if="error" class="sql-error">
      <pre>{{ error }}</pre>
    </div>

    <!-- 历史 -->
    <div v-if="history.length > 0" class="sql-history">
      <div class="sql-history-head">{{ t('database.queryHistory') }}</div>
      <div class="sql-history-list">
        <div v-for="(item, idx) in history" :key="idx" class="sql-history-item" @click="query = item.query">
          <span class="sql-history-time">{{ item.time }}</span>
          <span class="sql-history-query">{{ item.query }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { showMessage } from '../../../../../utils/message'

const { t } = useI18n()
import * as DatabaseService from '../../../../../../bindings/changeme/ssh/databaseservice.js'

const props = defineProps({
  connId: String, dbType: String, dbHost: String, dbPort: Number,
  dbUser: String, dbPassword: String, dbDatabase: String, tableName: String,
})

const query = ref('')
const executing = ref(false)
const resultColumns = ref([])
const resultRows_data = ref([])
const resultRows = ref(null)
const resultTime = ref(0)
const error = ref('')
const history = ref([])

const templates = {
  select: `SELECT * FROM \`${props.tableName || 'table_name'}\` LIMIT 100`,
  insert: `INSERT INTO \`${props.tableName || 'table_name'}\` (column1, column2) VALUES ('value1', 'value2')`,
  update: `UPDATE \`${props.tableName || 'table_name'}\` SET column1 = 'new_value' WHERE id = 1`,
  delete: `DELETE FROM \`${props.tableName || 'table_name'}\` WHERE id = 1`,
  create: `CREATE TABLE \`${props.tableName || 'new_table'}\` (\n  id INT PRIMARY KEY AUTO_INCREMENT,\n  name VARCHAR(255) NOT NULL,\n  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP\n)`,
}

function loadTemplate(type) {
  query.value = templates[type] || ''
}

async function execute() {
  if (!query.value.trim()) return
  executing.value = true
  error.value = ''
  resultColumns.value = []
  resultRows_data.value = []
  resultRows.value = null

  const start = Date.now()

  try {
    const result = await DatabaseService.ExecuteQuery(
      props.connId, props.dbType, props.dbHost, Number(props.dbPort),
      props.dbUser, props.dbPassword, props.dbDatabase, query.value
    )
    const rows = result?.[0] || []
    const cols = result?.[1] || []
    resultColumns.value = cols
    resultRows_data.value = rows
    resultRows.value = rows.length
    resultTime.value = Date.now() - start

    // 添加到历史
    history.value.unshift({
      query: query.value,
      time: new Date().toLocaleTimeString(),
    })
    if (history.value.length > 20) history.value.pop()
  } catch (e) {
    error.value = e?.message || String(e)
    resultTime.value = Date.now() - start
  } finally {
    executing.value = false
  }
}

function copyResult() {
  const header = resultColumns.value.join('\t')
  const rows = resultRows_data.value.map(r => r.join('\t')).join('\n')
  navigator.clipboard.writeText(header + '\n' + rows)
  showMessage(t('database.resultCopied'), 'success')
}
</script>

<style scoped>
.sql-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sql-editor {
  border-bottom: 1px solid var(--surface-hover);
  flex-shrink: 0;
}

.sql-editor-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: var(--surface-hover);
  font-size: 12px;
  color: var(--text-secondary);
}

.sql-editor-actions {
  display: flex;
  gap: 4px;
}

.sql-textarea {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg-input);
  border: none;
  color: var(--text-primary);
  font-family: 'Courier New', Consolas, monospace;
  font-size: 13px;
  resize: vertical;
  outline: none;
  min-height: 80px;
}

.sql-textarea:focus {
  background: var(--bg-panel);
}

.sql-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  border-top: 1px solid var(--surface-hover);
}

.sql-meta {
  font-size: 11px;
  color: var(--text-muted);
}

/* 结果 */
.sql-result {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.sql-result-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: var(--surface-hover);
  font-size: 12px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.sql-result-table-wrap {
  flex: 1;
  overflow: auto;
}

.sql-result-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.sql-result-table th {
  padding: 6px 10px;
  background: var(--surface-hover);
  color: var(--text-muted);
  font-weight: 600;
  text-align: left;
  border-bottom: 1px solid var(--border-strong);
  position: sticky;
  top: 0;
}

.sql-result-table td {
  padding: 4px 10px;
  border-bottom: 1px solid var(--surface-hover);
  color: var(--text-secondary);
  font-family: 'Courier New', Consolas, monospace;
  font-size: 11px;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sql-result-table td.null {
  color: var(--text-muted);
  font-style: italic;
}

/* 错误 */
.sql-error {
  padding: 12px;
  background: var(--accent-danger-bg);
  border-top: 1px solid var(--accent-danger);
  flex-shrink: 0;
}

.sql-error pre {
  margin: 0;
  font-size: 12px;
  color: var(--accent-danger);
  white-space: pre-wrap;
  word-break: break-all;
}

/* 历史 */
.sql-history {
  border-top: 1px solid var(--surface-hover);
  flex-shrink: 0;
  max-height: 120px;
  overflow-y: auto;
}

.sql-history-head {
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  background: var(--surface-hover);
}

.sql-history-item {
  display: flex;
  gap: 8px;
  padding: 4px 12px;
  cursor: pointer;
  font-size: 11px;
  transition: background 0.15s;
}

.sql-history-item:hover {
  background: var(--surface-hover);
}

.sql-history-time {
  color: var(--text-muted);
  white-space: nowrap;
  font-family: 'Courier New', monospace;
}

.sql-history-query {
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: 'Courier New', monospace;
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
  background: var(--surface-hover);
  color: var(--text-secondary);
  white-space: nowrap;
}

.btn:hover { background: var(--bg-toolbar); color: var(--text-primary); }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-sm { padding: 3px 8px; font-size: 11px; }
.btn-primary { background: var(--accent-primary); border-color: var(--accent-primary); color: white; }
.btn-primary:hover { opacity: 0.9; }
</style>
