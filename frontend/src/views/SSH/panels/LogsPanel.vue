<template>
  <div class="logs-panel">
    <!-- 顶部工具栏 -->
    <div class="logs-toolbar">
      <div class="toolbar-left">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
        <h3>操作日志</h3>
        <span class="connection-badge" v-if="currentConnId">
          {{ getConnectionName() }}
        </span>
      </div>
      <div class="toolbar-right">
        <button class="tool-btn" @click="exportLogs" title="导出日志">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
            <polyline points="7 10 12 15 17 10"></polyline>
            <line x1="12" y1="15" x2="12" y2="3"></line>
          </svg>
        </button>
        <button class="tool-btn" @click="toggleAutoRefresh" :class="{ active: autoRefresh }" title="自动刷新">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"></polyline>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
          </svg>
        </button>
        <button class="tool-btn" @click="loadLogs" title="手动刷新">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"></polyline>
            <polyline points="1 4 1 10 7 10"></polyline>
            <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"></path>
          </svg>
        </button>
      </div>
    </div>

    <!-- 搜索和过滤栏 -->
    <div class="logs-filter-bar">
      <div class="search-box">
        <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"></circle>
          <path d="m21 21-4.35-4.35"></path>
        </svg>
        <input 
          v-model="searchKeyword" 
          @input="handleSearch"
          type="text" 
          placeholder="搜索..."
          class="search-input"
        />
      </div>
      
      <div class="filter-buttons">
        <select v-model="filterType" @change="applyFilters" class="filter-select">
          <option value="all">全部类型</option>
          <option value="terminal">终端命令</option>
          <option value="fileManager">文件管理</option>
          <option value="connection">连接事件</option>
          <option value="system">系统事件</option>
          <option value="error">错误信息</option>
          <option value="security">安全相关</option>
          <option value="ai">AI 工具</option>
          <option value="portForward">端口转发</option>
          <option value="firewall">防火墙</option>
          <option value="guardian">进程守护</option>
        </select>

        <select v-model="filterLevel" @change="applyFilters" class="filter-select">
          <option value="all">全部级别</option>
          <option value="info">信息</option>
          <option value="success">成功</option>
          <option value="warning">警告</option>
          <option value="error">错误</option>
        </select>
      </div>
    </div>

    <!-- 日志列表容器 -->
    <div class="logs-container" ref="logsContainer">
      <div v-if="filteredLogs.length === 0" class="empty-state">
        <svg class="empty-icon" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
        <p v-if="searchKeyword || filterType !== 'all' || filterLevel !== 'all'">
          没有找到匹配的日志
        </p>
        <p v-else>暂无日志记录</p>
        <button v-if="!autoRefresh" @click="loadLogs" class="retry-btn">
          重新加载
        </button>
      </div>
      
      <div v-else class="log-list">
        <div 
          v-for="(log, index) in filteredLogs" 
          :key="index" 
          class="log-card"
          @click="toggleLogDetail(index)"
        >
          <div class="log-card-header">
            <div class="log-meta-group">
              <span class="log-time">{{ formatTime(log.timestamp) }}</span>
              <span v-if="getTypeLabel(log.type)" class="type-tag" :class="'type-' + log.type">
                {{ getTypeLabel(log.type) }}
              </span>
              <span v-if="log.level && log.level !== 'info'" class="log-level-badge" :class="'badge-' + log.level">
                {{ getLevelLabel(log.level) }}
              </span>
            </div>
            <svg v-if="log.details" @click.stop="toggleLogDetail(index)" class="expand-icon" :class="{ 'expanded': expandedLogs[index] }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="18 15 12 9 6 15"></polyline>
            </svg>
          </div>
          
          <div class="log-content">
            <span class="message-text">{{ log.message }}</span>
          </div>
          
          <!-- 只显示危险命令的警告，安全命令不显示任何提示 -->
          <div v-if="(log.riskLevel === 'critical' || log.riskLevel === 'high') && log.warnings && log.warnings.length > 0" class="log-security-info">
            <div v-for="(warning, idx) in log.warnings" :key="'w-' + idx" class="warning-text">
              {{ warning }}
            </div>
          </div>
          
          <div v-if="log.details && expandedLogs[index]" class="log-details">
            <div class="details-label">详细信息:</div>
            <pre class="details-content">{{ log.details }}</pre>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部状态栏 -->
    <div class="logs-footer">
      <div class="footer-left">
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" @change="toggleAutoRefresh" />
          <span>自动刷新 ({{ refreshInterval / 1000 }}s)</span>
        </label>
      </div>
      <div class="footer-right">
        <span class="last-update">最后更新: {{ formatTime(lastUpdateTime) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useSSHLayoutStore } from '../../../stores/sshLayout'
import { useSSHTabsStore } from '../../../stores/sshTabs'
import { SSHService } from '../../../../bindings/changeme/ssh/index.js'
import { showMessage } from '../../../utils/message'
import { getLogs, exportLogs as exportLogsUtil } from '../../../utils/logger'
import { useAIToolLogStore } from '../../../stores/aiToolLog'

const sshLayoutStore = useSSHLayoutStore()
const sshTabsStore = useSSHTabsStore()
const aiToolLog = useAIToolLogStore()
const logsContainer = ref(null)

// 状态
const allLogs = ref([])
const filteredLogs = ref([])
const searchKeyword = ref('')
const filterType = ref('all')
const filterLevel = ref('all')
const autoRefresh = ref(true)
const lastUpdateTime = ref(null)
const expandedLogs = ref({})

// 刷新间隔（毫秒）
const refreshInterval = 3000
let refreshTimer = null

// 获取当前连接ID
const currentConnId = computed(() => sshLayoutStore.currentConnectionId)

// 获取连接名称
const getConnectionName = () => {
  const tabs = sshTabsStore.tabs || []
  const currentTab = tabs.find(tab => tab.id === currentConnId.value)
  return currentTab ? currentTab.name : '未知连接'
}

// 加载日志 - 合并统一日志服务 + AI 工具日志
const loadLogs = async () => {
  if (!currentConnId.value || currentConnId.value === 'default-connection') {
    allLogs.value = []
    filteredLogs.value = []
    return
  }

  try {
    // 从统一日志服务获取
    const entries = getLogs(currentConnId.value)

    // 合并 AI 工具日志
    const aiLogs = aiToolLog.getLogs(currentConnId.value)
    const aiEntries = aiLogs.map(log => ({
      id: log.id,
      type: 'ai',
      level: log.status === 'denied' ? 'warning' : log.status === 'completed' ? 'success' : 'info',
      message: `[AI] ${log.command || log.tool || ''}`,
      details: log.result ? `结果 (${log.result.length} 字符):\n${log.result.slice(0, 500)}` : null,
      timestamp: log.timestamp
    }))

    // 合并并按时间排序（最新的在前）
    const merged = [...entries, ...aiEntries].sort((a, b) =>
      new Date(b.timestamp) - new Date(a.timestamp)
    )

    allLogs.value = merged
    lastUpdateTime.value = new Date()
    applyFilters()
    console.log('[LogsPanel] ✅ 加载日志:', entries.length, '条系统 +', aiEntries.length, '条AI')
  } catch (error) {
    console.error('[LogsPanel] 加载日志失败:', error)
    showMessage('加载日志失败', 'error')
  }
}

// 应用过滤器
const applyFilters = () => {
  let result = [...allLogs.value]
  
  // 按类型过滤
  if (filterType.value !== 'all') {
    result = result.filter(log => log.type === filterType.value)
  }
  
  // 按级别过滤
  if (filterLevel.value !== 'all') {
    result = result.filter(log => log.level === filterLevel.value)
  }
  
  // 关键词搜索
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(log => {
      const msg = (log.message || log.content || '').toLowerCase()
      const det = typeof log.details === 'string' ? log.details.toLowerCase() : ''
      return msg.includes(keyword) || det.includes(keyword)
    })
  }
  
  filteredLogs.value = result
}

// 处理搜索输入（带 debounce）
let searchTimeout = null
const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    applyFilters()
  }, 300)
}

// 切换日志详情展开/收起
const toggleLogDetail = (index) => {
  expandedLogs.value[index] = !expandedLogs.value[index]
}

// 导出日志 - 🎯 使用统一日志服务
const exportLogs = async () => {
  if (!currentConnId.value) {
    showMessage('没有可导出的日志', 'warning')
    return
  }

  try {
    // 从日志服务获取数据
    const jsonContent = exportLogsUtil(currentConnId.value)
    
    if (!jsonContent) {
      showMessage('没有可导出的日志', 'warning')
      return
    }
    
    // 创建下载链接
    const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `ssh-logs-${currentConnId.value}-${Date.now()}.log`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    
    const logs = getLogs(currentConnId.value)
    showMessage(`已导出 ${logs.length} 条日志`, 'success')
    console.log('[LogsPanel] ✅ 日志已导出为 JSON 格式 (.log)')
  } catch (error) {
    console.error('[LogsPanel] 导出日志失败:', error)
    showMessage('导出日志失败', 'error')
  }
}

// 切换自动刷新
const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

// 启动自动刷新
const startAutoRefresh = () => {
  stopAutoRefresh()
  refreshTimer = setInterval(loadLogs, refreshInterval)
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 格式化时间 - 年月日时分秒
const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 获取类型标签
const getTypeLabel = (type) => {
  const labels = {
    terminal: '终端',
    fileManager: '文件',
    connection: '连接',
    system: '系统',
    error: '错误',
    security: '安全',
    ai: 'AI',
    portForward: '端口转发',
    firewall: '防火墙',
    guardian: '进程守护'
  }
  return labels[type] || type
}

// 获取级别标签
const getLevelLabel = (level) => {
  const labels = {
    info: '信息',
    success: '成功',
    warning: '警告',
    error: '错误'
  }
  return labels[level] || level
}

// 生命周期
onMounted(() => {
  loadLogs()
  if (autoRefresh.value) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  stopAutoRefresh()
  clearTimeout(searchTimeout)
})
</script>

<style scoped>
.logs-panel {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-panel);
  overflow: hidden;
}

/* 工具栏 */
.logs-toolbar {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-default);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-toolbar);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.toolbar-left svg {
  color: var(--primary-light);
}

.toolbar-left h3 {
  margin: 0;
  color: var(--text-primary);
  font-size: 0.875rem;
  font-weight: 600;
}

.connection-badge {
  padding: 0.25rem 0.5rem;
  background: var(--primary-bg);
  color: var(--primary-light);
  border-radius: 0.25rem;
  font-size: 0.6875rem;
  font-weight: 500;
}

.toolbar-right {
  display: flex;
  gap: 0.375rem;
}

.tool-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  background: transparent;
  border: 1px solid var(--border-default);
  border-radius: 0.25rem;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.tool-btn:hover {
  background: var(--bg-hover);
  border-color: var(--border-strong);
  color: var(--text-primary);
}

.tool-btn.active {
  background: var(--primary-bg);
  border-color: var(--border-accent);
  color: var(--primary-light);
}

/* 过滤栏 */
.logs-filter-bar {
  padding: 0.5rem 1rem;
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  background: var(--bg-toolbar);
}

.search-box {
  flex: 0 0 240px;
  position: relative;
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
  padding: 0.375rem 0.625rem 0.375rem 2rem;
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: 0.375rem;
  color: var(--text-primary);
  font-size: 0.75rem;
  outline: none;
  transition: all 0.2s;
}

.search-input:focus {
  border-color: var(--border-accent);
  background: var(--bg-input-focus);
  box-shadow: 0 0 0 2px var(--primary-bg);
}

.search-input::placeholder {
  color: var(--text-muted);
}

.filter-buttons {
  display: flex;
  gap: 0.5rem;
}

.filter-select {
  min-width: 110px;
  padding: 0.375rem 0.625rem;
  background: var(--bg-input);
  border: 1px solid var(--border-default);
  border-radius: 0.375rem;
  color: var(--text-primary);
  font-size: 0.75rem;
  outline: none;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-select:hover {
  border-color: var(--border-strong);
}

.filter-select:focus {
  border-color: var(--border-accent);
  box-shadow: 0 0 0 2px var(--primary-bg);
}

/* 日志容器 */
.logs-container {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem;
}

.logs-container::-webkit-scrollbar {
  width: 6px;
}

.logs-container::-webkit-scrollbar-track {
  background: transparent;
}

.logs-container::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb);
  border-radius: 3px;
}

.logs-container::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover);
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 1rem;
  color: var(--text-muted);
}

.empty-icon {
  opacity: 0.4;
  color: var(--text-muted);
}

.empty-state p {
  margin: 0;
  font-size: 0.875rem;
}

.retry-btn {
  padding: 0.5rem 1rem;
  background: var(--primary-bg);
  border: 1px solid var(--border-accent);
  border-radius: 0.25rem;
  color: var(--primary-light);
  cursor: pointer;
  font-size: 0.75rem;
  transition: all 0.2s;
}

.retry-btn:hover {
  background: var(--primary-bg-hover);
}

/* 日志卡片 - 简洁现代化设计 */
.log-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.log-card {
  padding: 0.75rem 1rem;
  background: var(--surface-1);
  border-radius: 0.375rem;
  transition: all 0.2s ease;
  cursor: pointer;
}

.log-card:hover {
  background: var(--surface-hover);
  transform: translateX(2px);
}

.log-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.375rem;
}

.log-meta-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.log-time {
  font-size: 0.6875rem;
  color: var(--text-muted);
  font-family: 'Courier New', monospace;
}

.type-tag {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.625rem;
  font-weight: 600;
}

.type-terminal {
  background: var(--primary-bg);
  color: var(--primary-light);
}
.type-fileManager {
  background: var(--success-bg);
  color: var(--success-light);
}
.type-connection {
  background: color-mix(in srgb, var(--accent-purple), transparent 80%);
  color: var(--accent-purple);
}
.type-system {
  background: var(--warning-bg);
  color: var(--warning-light);
}
.type-error {
  background: var(--danger-bg);
  color: var(--danger-light);
}
.type-security {
  background: color-mix(in srgb, var(--accent-warning), transparent 80%);
  color: var(--warning-light);
}
.type-ai {
  background: color-mix(in srgb, var(--accent-purple), transparent 80%);
  color: var(--accent-purple);
}
.type-portForward {
  background: var(--primary-bg);
  color: var(--primary-light);
}
.type-firewall {
  background: var(--warning-bg);
  color: var(--warning-light);
}
.type-guardian {
  background: color-mix(in srgb, var(--accent-purple), transparent 80%);
  color: var(--accent-purple);
}

.log-level-badge {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.625rem;
  font-weight: 600;
}

.badge-info {
  background: var(--primary-bg);
  color: var(--primary-light);
}
.badge-success {
  background: var(--success-bg);
  color: var(--success-light);
}
.badge-warning {
  background: color-mix(in srgb, var(--accent-warning), transparent 80%);
  color: var(--warning-light);
}
.badge-error {
  background: var(--danger-bg);
  color: var(--danger-light);
}
.badge-low {
  background: color-mix(in srgb, var(--accent-warning), transparent 80%);
  color: var(--warning-light);
}
.badge-medium {
  background: color-mix(in srgb, var(--accent-warning), transparent 70%);
  color: var(--warning-light);
}
.badge-high {
  background: color-mix(in srgb, var(--accent-danger), transparent 75%);
  color: var(--danger-light);
  font-weight: 700;
}
.badge-critical {
  background: color-mix(in srgb, var(--accent-danger), transparent 70%);
  color: var(--danger-light);
  font-weight: 700;
}

.expand-icon {
  color: var(--text-muted);
  transition: transform 0.2s;
  cursor: pointer;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.log-content {
  font-size: 0.8125rem;
  line-height: 1.5;
  word-break: break-word;
}

.command-text {
  color: var(--text-primary);
  font-family: 'Courier New', monospace;
  font-weight: 500;
}

.message-text {
  color: var(--text-primary);
  line-height: 1.5;
}

/* 安全信息区域 - 仅显示危险命令警告 */
.log-security-info {
  margin-top: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: var(--danger-bg);
  border-radius: 0.25rem;
  border-left: 2px solid var(--accent-danger);
}

.security-warnings {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.warning-text {
  font-size: 0.75rem;
  color: var(--danger-light);
  line-height: 1.4;
}

.log-details {
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--border-default);
}

.details-label {
  font-size: 0.6875rem;
  color: var(--text-muted);
  margin-bottom: 0.375rem;
  font-weight: 600;
}

.details-content {
  font-size: 0.75rem;
  color: var(--text-secondary);
  background: var(--surface-1);
  padding: 0.625rem;
  border-radius: 0.25rem;
  overflow-x: auto;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Courier New', monospace;
}

/* 底部状态栏 */
.logs-footer {
  padding: 0.5rem 1rem;
  border-top: 1px solid var(--border-default);
  background: var(--bg-toolbar);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.footer-left,
.footer-right {
  display: flex;
  align-items: center;
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.6875rem;
  color: var(--text-secondary);
  cursor: pointer;
}

.auto-refresh-toggle input[type="checkbox"] {
  cursor: pointer;
}

.last-update {
  font-size: 0.6875rem;
  color: var(--text-muted);
}

/* === 响应式布局 === */

/* 小宽度 */
@media (max-width: 600px) {
  .logs-toolbar {
    padding: 0.375rem 0.5rem;
    gap: 0.25rem;
  }

  .toolbar-left h3 {
    font-size: 0.75rem;
  }

  .connection-badge {
    display: none;
  }

  .logs-filter-bar {
    padding: 0.375rem 0.5rem;
    flex-wrap: wrap;
    gap: 0.375rem;
  }

  .search-box {
    flex: 1 1 100%;
    min-width: 0;
  }

  .search-input {
    font-size: 0.625rem;
    padding: 0.25rem 0.5rem 0.25rem 1.75rem;
  }

  .filter-buttons {
    flex: 1 1 100%;
    gap: 0.375rem;
  }

  .filter-select {
    flex: 1;
    min-width: 0;
    font-size: 0.625rem;
    padding: 0.25rem 0.375rem;
  }

  .logs-container {
    padding: 0.375rem;
  }

  .log-card {
    padding: 0.5rem 0.625rem;
  }

  .log-time {
    font-size: 0.5625rem;
  }

  .type-tag,
  .log-level-badge {
    font-size: 0.5rem;
    padding: 0.0625rem 0.375rem;
  }

  .log-content {
    font-size: 0.6875rem;
  }

  .details-content {
    font-size: 0.625rem;
    padding: 0.375rem;
  }

  .logs-footer {
    padding: 0.25rem 0.5rem;
    font-size: 0.5625rem;
  }
}

/* 小高度 */
@media (max-height: 500px) {
  .logs-toolbar {
    padding: 0.25rem 0.5rem;
  }

  .toolbar-left h3 {
    font-size: 0.75rem;
  }

  .toolbar-left svg {
    width: 14px;
    height: 14px;
  }

  .tool-btn {
    width: 1.5rem;
    height: 1.5rem;
  }

  .logs-filter-bar {
    padding: 0.25rem 0.5rem;
  }

  .search-input {
    font-size: 0.625rem;
    padding: 0.1875rem 0.375rem 0.1875rem 1.5rem;
  }

  .filter-select {
    font-size: 0.625rem;
    padding: 0.1875rem 0.375rem;
  }

  .logs-container {
    padding: 0.25rem;
  }

  .log-card {
    padding: 0.375rem 0.5rem;
  }

  .log-card-header {
    margin-bottom: 0.125rem;
  }

  .log-meta-group {
    gap: 0.25rem;
  }

  .log-content {
    font-size: 0.6875rem;
    line-height: 1.3;
  }

  .log-security-info {
    margin-top: 0.25rem;
    padding: 0.25rem 0.5rem;
  }

  .log-details {
    margin-top: 0.375rem;
    padding-top: 0.375rem;
  }

  .details-content {
    font-size: 0.625rem;
    padding: 0.25rem;
    max-height: 100px;
  }

  .logs-footer {
    padding: 0.125rem 0.5rem;
  }

  .empty-icon {
    width: 32px;
    height: 32px;
  }

  .empty-state p {
    font-size: 0.75rem;
  }
}
</style>
