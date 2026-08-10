<template>
  <div class="dbm-sidebar">
    <!-- 头部 -->
    <div class="sidebar-head">
      <span class="sidebar-title">{{ t('database.dbList') }}</span>
      <button class="icon-btn" @click="$emit('refresh')" :title="t('database.refresh')">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
      </button>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="sidebar-loading">
      <div class="spinner-sm"></div>
    </div>

    <!-- 数据库树 -->
    <div v-else class="db-tree">
      <div
        v-for="db in databases"
        :key="db"
        class="tree-db"
        :class="{ active: selectedDb === db, expanded: expandedDbs[db] }"
      >
        <div class="tree-db-head" @click="toggleDb(db)">
          <svg class="tree-arrow" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg>
          <span class="tree-db-name">{{ db }}</span>
          <button class="tree-db-menu" @click.stop="$emit('drop-db', db)" :title="t('database.deleteDb')">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
        </div>

        <!-- 表列表 -->
        <div v-if="expandedDbs[db]" class="tree-tables">
          <div v-if="!tables[db]" class="tree-loading">
            <div class="spinner-sm"></div>
          </div>
          <div v-else-if="tables[db].length === 0" class="tree-empty">{{ t('database.noTables') }}</div>
          <div
            v-for="table in tables[db]"
            :key="table"
            class="tree-table"
            :class="{ active: selectedTable === table && selectedDb === db }"
            @click="$emit('select-table', db, table)"
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/></svg>
            <span>{{ table }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  databases: { type: Array, default: () => [] },
  tables: { type: Object, default: () => ({}) },
  selectedDb: { type: String, default: '' },
  selectedTable: { type: String, default: '' },
  loading: { type: Boolean, default: false },
})

const emit = defineEmits(['select-db', 'select-table', 'refresh', 'drop-db'])

const expandedDbs = ref({})

function toggleDb(db) {
  if (expandedDbs.value[db]) {
    expandedDbs.value[db] = false
    return
  }
  expandedDbs.value[db] = true
  emit('select-db', db)
}
</script>

<style scoped>
.dbm-sidebar {
  width: 220px;
  border-right: 1px solid var(--surface-hover);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  flex-shrink: 0;
}

.sidebar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  border-bottom: 1px solid var(--surface-hover);
  flex-shrink: 0;
}

.sidebar-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.icon-btn {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-muted);
  cursor: pointer;
}

.icon-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.sidebar-loading {
  display: flex;
  justify-content: center;
  padding: 20px;
}

.db-tree {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

/* 数据库节点 */
.tree-db-head {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: background 0.15s;
}

.tree-db-head:hover {
  background: var(--surface-hover);
}

.tree-db.active > .tree-db-head {
  color: var(--accent-primary);
  background: var(--accent-primary-bg);
}

.tree-arrow {
  transition: transform 0.2s;
  flex-shrink: 0;
}

.tree-db.expanded > .tree-db-head .tree-arrow {
  transform: rotate(90deg);
}

.tree-db-name {
  flex: 1;
  font-size: 12px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-db-menu {
  opacity: 0;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.tree-db-head:hover .tree-db-menu {
  opacity: 1;
}

.tree-db-menu:hover {
  background: var(--accent-danger-bg);
  color: var(--accent-danger);
}

/* 表列表 */
.tree-tables {
  padding-left: 20px;
}

.tree-loading, .tree-empty {
  padding: 8px 12px;
  font-size: 11px;
  color: var(--text-muted);
}

.tree-table {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 8px;
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
