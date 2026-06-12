<template>
  <div class="file-list-container">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>
    
    <div v-else-if="files.length === 0" class="empty-state">
      <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
      </svg>
      <p>空目录</p>
    </div>
    
    <table v-else class="file-table">
      <thead>
        <tr>
          <th style="width: 40px;">
            <input 
              type="checkbox" 
              :checked="files.length > 0 && files.every(f => selectedFiles.has(f.name))"
              @change="$emit('select-all', $event.target.checked)"
              class="select-all-checkbox"
            />
          </th>
          <th @click="$emit('sort', 'name')" class="sortable">
            名称
            <span v-if="sortKey === 'name'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th @click="$emit('sort', 'size')" class="sortable">
            大小
            <span v-if="sortKey === 'size'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th @click="$emit('sort', 'modTime')" class="sortable">
            修改时间
            <span v-if="sortKey === 'modTime'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
          </th>
          <th>所有者</th>
          <th>权限</th>
          <th style="width: 60px;">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="file in files" :key="file.path" class="file-row" @dblclick="$emit('open-file', file)">
          <td>
            <input 
              type="checkbox" 
              :checked="selectedFiles.has(file.name)"
              @change="$emit('select-file', file.name)"
              @click.stop
              class="file-checkbox"
            />
          </td>
          <td class="file-name">
            <svg v-if="file.isDir" class="file-icon folder" width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
            </svg>
            <svg v-else class="file-icon file" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
              <polyline points="13 2 13 9 20 9"></polyline>
            </svg>
            <span>{{ file.displayName || file.name }}</span>
            <span v-if="file.relativePath" class="file-path">{{ file.relativePath }}</span>
          </td>
          <td>{{ formatSize(file.size) }}</td>
          <td>{{ formatTime(file.modTime) }}</td>
          <td>{{ file.owner }}</td>
          <td class="mode-text">{{ getNumericMode(file.mode) }}</td>
          <td>
            <div class="action-menu-wrapper" @click.stop>
              <button @click="$emit('show-menu', file, $event)" class="menu-btn" title="更多操作">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                  <circle cx="12" cy="5" r="2"></circle>
                  <circle cx="12" cy="12" r="2"></circle>
                  <circle cx="12" cy="19" r="2"></circle>
                </svg>
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { formatSize, formatTime, getNumericMode } from '../utils/pathHelper'

defineProps({
  files: {
    type: Array,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  sortKey: {
    type: String,
    required: true
  },
  sortOrder: {
    type: String,
    required: true
  },
  selectedFiles: {
    type: Set,
    default: () => new Set()
  }
})

defineEmits(['open-file', 'sort', 'show-menu', 'select-file', 'select-all'])
</script>

<style scoped>
.file-list-container {
  height: 100%;
  overflow-y: auto;
  position: relative;
}

.loading-state,
.empty-state {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  color: var(--text-secondary);
  background: var(--bg-panel);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--surface-2);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.file-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.file-table thead {
  position: sticky;
  top: 0;
  background: var(--bg-toolbar);
  z-index: 1;
}

.file-table th {
  padding: 0.625rem 0.75rem;
  text-align: left;
  color: var(--text-secondary);
  font-weight: 600;
  border-bottom: 2px solid var(--border-default);
  white-space: nowrap;
}

.file-table th.sortable {
  cursor: pointer;
  user-select: none;
}

.file-table th.sortable:hover {
  color: var(--text-primary);
}

.sort-indicator {
  margin-left: 0.25rem;
  color: var(--primary-light);
}

.file-table td {
  padding: 0.5rem 0.75rem;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-subtle);
}

.file-row {
  cursor: pointer;
  transition: background 0.15s;
}

.file-row:hover {
  background: var(--primary-bg);
}

.file-name {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.file-path {
  color: var(--text-muted);
  font-size: 0.75rem;
  margin-left: 0.25rem;
}

.file-icon {
  flex-shrink: 0;
}

.file-icon.folder {
  color: var(--accent-warning);
}

.file-icon.file {
  color: var(--primary-light);
}

.mode-text {
  font-family: monospace;
  color: var(--text-secondary);
  font-size: 0.75rem;
}

.action-menu-wrapper {
  position: relative;
}

.file-checkbox,
.select-all-checkbox {
  cursor: pointer;
  width: 16px;
  height: 16px;
  accent-color: var(--accent-primary);
}

.file-checkbox:hover,
.select-all-checkbox:hover {
  accent-color: var(--primary-light);
}

.menu-btn {
  width: 1.75rem;
  height: 1.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 0.25rem;
  color: var(--primary-light);
  cursor: pointer;
  transition: all 0.2s;
}

.menu-btn:hover {
  background: var(--primary-bg);
  border-color: var(--border-accent);
}

.file-list-container::-webkit-scrollbar {
  width: 6px;
}

.file-list-container::-webkit-scrollbar-track {
  background: transparent;
}

.file-list-container::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb);
  border-radius: 3px;
}

.file-list-container::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover);
}

/* === 响应式布局 === */

/* 小宽度：隐藏部分列 */
@media (max-width: 600px) {
  .file-table {
    font-size: 0.625rem;
  }

  .file-table th,
  .file-table td {
    padding: 0.25rem 0.375rem;
  }

  /* 隐藏权限和用户列 */
  .file-table th:nth-child(5),
  .file-table td:nth-child(5),
  .file-table th:nth-child(6),
  .file-table td:nth-child(6) {
    display: none;
  }

  .file-path {
    display: none;
  }

  .file-name {
    gap: 0.25rem;
  }

  .mode-text {
    font-size: 0.625rem;
  }
}

/* 小高度：压缩行高 */
@media (max-height: 500px) {
  .file-table th,
  .file-table td {
    padding: 0.125rem 0.375rem;
  }

  .file-table {
    font-size: 0.625rem;
  }

  .file-checkbox,
  .select-all-checkbox {
    width: 14px;
    height: 14px;
  }

  .menu-btn {
    width: 1.25rem;
    height: 1.25rem;
  }
}
</style>
