<template>
  <div class="monitor-panel">
    <!-- 顶部概览卡片 -->
    <div class="overview-cards">
      <!-- CPU -->
      <div class="stat-card cpu">
        <div class="card-header">
          <svg class="card-icon icon-cpu" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><line x1="9" y1="1" x2="9" y2="4"/><line x1="15" y1="1" x2="15" y2="4"/><line x1="9" y1="20" x2="9" y2="23"/><line x1="15" y1="20" x2="15" y2="23"/><line x1="20" y1="9" x2="23" y2="9"/><line x1="20" y1="14" x2="23" y2="14"/><line x1="1" y1="9" x2="4" y2="9"/><line x1="1" y1="14" x2="4" y2="14"/></svg>
          <span class="card-title">CPU</span>
        </div>
        <div class="card-body">
          <div class="usage-circle" :style="{ background: getCpuGradient() }">
            <div class="usage-text">{{ systemStats?.cpu?.usagePercent?.toFixed(1) || 0 }}%</div>
          </div>
          <div class="card-details">
            <div class="detail-item">
              <span class="label">核心数:</span>
              <span class="value">{{ systemStats?.cpu?.cores || 0 }}</span>
            </div>
            <div class="detail-item">
              <span class="label">负载:</span>
              <span class="value">{{ systemStats?.cpu?.loadAvg?.load1?.toFixed(2) || '0.00' }}</span>
            </div>
            <div class="detail-item">
              <span class="label">运行时间:</span>
              <span class="value uptime-value">{{ systemStats?.uptime || '加载中...' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 内存 -->
      <div class="stat-card memory">
        <div class="card-header">
          <svg class="card-icon icon-memory" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><rect x="2" y="6" width="20" height="12" rx="2"/><line x1="6" y1="10" x2="6" y2="14"/><line x1="10" y1="10" x2="10" y2="14"/><line x1="14" y1="10" x2="14" y2="14"/><line x1="18" y1="10" x2="18" y2="14"/><line x1="2" y1="10" x2="2" y2="6"/><line x1="22" y1="10" x2="22" y2="6"/></svg>
          <span class="card-title">内存</span>
        </div>
        <div class="card-body">
          <div class="usage-bar">
            <div 
              class="usage-fill" 
              :style="{ width: (systemStats?.memory?.usedPercent || 0) + '%' }"
            ></div>
          </div>
          <div class="card-details">
            <div class="detail-item">
              <span class="label">已用:</span>
              <span class="value">{{ formatBytes(systemStats?.memory?.used || 0) }}</span>
            </div>
            <div class="detail-item">
              <span class="label">总计:</span>
              <span class="value">{{ formatBytes(systemStats?.memory?.total || 0) }}</span>
            </div>
            <div class="detail-item">
              <span class="label">使用率:</span>
              <span class="value">{{ (systemStats?.memory?.usedPercent || 0).toFixed(1) }}%</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 磁盘 -->
      <div class="stat-card disk">
        <div class="card-header">
          <svg class="card-icon icon-disk" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="3"/><line x1="12" y1="2" x2="12" y2="6"/></svg>
          <span class="card-title">磁盘</span>
        </div>
        <div class="card-body">
          <div class="disk-list">
            <div v-for="(partition, idx) in systemStats?.disk?.partitions?.slice(0, 3)" :key="idx" class="disk-item">
              <div class="disk-info">
                <div class="disk-name">{{ partition.mountpoint }}</div>
                <div class="disk-size">{{ formatBytes(partition.total) }}</div>
              </div>
              <div class="usage-bar small">
                <div 
                  class="usage-fill" 
                  :style="{ width: partition.usedPercent + '%' }"
                ></div>
              </div>
              <div class="disk-percent">{{ partition.usedPercent.toFixed(1) }}%</div>
            </div>
          </div>
          <!-- 磁盘 I/O 图表 -->
          <div class="chart-container">
            <svg class="sparkline-chart" viewBox="0 0 100 30" preserveAspectRatio="none">
              <polyline
                :points="getDiskIoPoints()"
                fill="none"
                class="chart-line-primary"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
            <div class="chart-value">{{ formatBytes(getCurrentDiskIoRate()) }}/s</div>
          </div>
        </div>
      </div>

      <!-- 网络 -->
      <div class="stat-card network">
        <div class="card-header">
          <svg class="card-icon icon-network" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
          <span class="card-title">网络</span>
        </div>
        <div class="card-body">
          <div class="network-stats">
            <div class="net-item">
              <span class="net-label">↓ 接收:</span>
              <span class="net-value">{{ formatBytes(systemStats?.network?.totalRx || 0) }}</span>
            </div>
            <div class="net-item">
              <span class="net-label">↑ 发送:</span>
              <span class="net-value">{{ formatBytes(systemStats?.network?.totalTx || 0) }}</span>
            </div>
          </div>
          <!-- 网络流量图表 -->
          <div class="chart-container network-chart">
            <svg class="sparkline-chart" viewBox="0 0 100 30" preserveAspectRatio="none">
              <!-- 下载流量 -->
              <polyline
                :points="getNetworkRxPoints()"
                fill="none"
                class="chart-line-success"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <!-- 上传流量 -->
              <polyline
                :points="getNetworkTxPoints()"
                fill="none"
                class="chart-line-warning"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
            <div class="network-rates">
              <span class="rate-item rx">↓ {{ formatBytes(getCurrentNetworkRxRate()) }}/s</span>
              <span class="rate-item tx">↑ {{ formatBytes(getCurrentNetworkTxRate()) }}/s</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 进程列表 -->
    <div class="process-section">
      <div class="section-header">
        <h3>进程列表</h3>
        <div class="header-actions">
          <button @click="refreshProcesses" class="refresh-btn" :disabled="loadingProcesses">
            {{ loadingProcesses ? '刷新中...' : '刷新' }}
          </button>
        </div>
      </div>

      <div class="process-table-container">
        <table class="process-table">
          <thead>
            <tr>
              <th @click="sortBy('name')" class="sortable">
                名称
                <span v-if="sortKey === 'name'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('pid')" class="sortable">
                PID
                <span v-if="sortKey === 'pid'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('cpuPercent')" class="sortable">
                CPU %
                <span v-if="sortKey === 'cpuPercent'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('memPercent')" class="sortable">
                内存 %
                <span v-if="sortKey === 'memPercent'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('memRss')" class="sortable">
                内存使用
                <span v-if="sortKey === 'memRss'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('startTime')" class="sortable">
                启动时间
                <span v-if="sortKey === 'startTime'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('elapsedTime')" class="sortable">
                运行时长
                <span v-if="sortKey === 'elapsedTime'" class="sort-indicator">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th>用户</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="proc in sortedProcesses" :key="proc.pid" class="process-row">
              <td class="process-name" :title="proc.cmdline">{{ proc.name }}</td>
              <td>{{ proc.pid }}</td>
              <td>
                <span :class="getUsageClass(proc.cpuPercent)">
                  {{ proc.cpuPercent.toFixed(1) }}%
                </span>
              </td>
              <td>
                <span :class="getUsageClass(proc.memPercent)">
                  {{ proc.memPercent.toFixed(1) }}%
                </span>
              </td>
              <td>{{ formatBytes(proc.memRss) }}</td>
              <td>{{ proc.startTime }}</td>
              <td>{{ proc.elapsedTime }}</td>
              <td>{{ proc.username }}</td>
              <td>{{ proc.status }}</td>
              <td>
                <div class="action-buttons">
                  <button 
                    @click="showSignalMenu(proc.pid, $event)" 
                    class="signal-btn"
                    title="发送信号"
                  >
                    信号
                  </button>
                  <button 
                    @click="killProcess(proc.pid)" 
                    class="kill-btn"
                    title="终止进程"
                  >
                    终止
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <!-- 确认对话框 -->
  <Modal
    v-model:visible="showKillConfirm"
    title="终止进程"
    :content="killConfirmMessage"
    danger
    confirm-text="终止"
    cancel-text="取消"
    @confirm="confirmKillProcess"
    @close="showKillConfirm = false"
    @cancel="showKillConfirm = false"
  />

  <!-- 信号选择菜单 -->
  <Teleport to="body">
    <div v-if="showSignalDropdown" class="signal-dropdown" :style="{ top: signalMenuPosition.y + 'px', left: signalMenuPosition.x + 'px' }" @click.stop>
      <div class="dropdown-header">发送信号到进程 {{ selectedPid }}</div>
      <button @click="sendSignal('TERM')" class="dropdown-item">SIGTERM (正常终止)</button>
      <button @click="sendSignal('HUP')" class="dropdown-item">SIGHUP (挂起)</button>
      <button @click="sendSignal('INT')" class="dropdown-item">SIGINT (中断)</button>
      <button @click="sendSignal('QUIT')" class="dropdown-item">SIGQUIT (退出)</button>
      <button @click="sendSignal('KILL')" class="dropdown-item danger">SIGKILL (强制终止)</button>
      <button @click="sendSignal('STOP')" class="dropdown-item">SIGSTOP (停止)</button>
      <button @click="sendSignal('CONT')" class="dropdown-item">SIGCONT (继续)</button>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useSSHLayoutStore } from '../../../stores/sshLayout'
import { SSHService } from '../../../../bindings/changeme/ssh/index.js'
import { showMessage } from '../../../utils/message'
import Modal from '../../../components/Modal.vue'
import { logFileOperation, logSystemEvent } from '@/utils/logger'

// 定义接受的 props（Dockview 传递的 params）
defineProps({
  params: {
    type: Object,
    default: () => ({})
  }
})

const sshLayoutStore = useSSHLayoutStore()

// 状态
const systemStats = ref(null)
const processList = ref([])
const loadingProcesses = ref(false)
const sortKey = ref('cpuPercent')
const sortOrder = ref('desc')

// 历史数据追踪（最多保留20个数据点）
const diskIoHistory = ref([]) // 磁盘 I/O 历史（速率，bytes/s）
const networkRxHistory = ref([]) // 网络接收历史（速率，bytes/s）
const networkTxHistory = ref([]) // 网络发送历史（速率，bytes/s）
const maxHistoryPoints = 20

// 上一次的数据（用于计算速率）
const lastDiskRead = ref(0)
const lastDiskWrite = ref(0)
const lastNetworkRx = ref(0)
const lastNetworkTx = ref(0)
const lastStatsTime = ref(0)

// 终止进程确认对话框
const showKillConfirm = ref(false)
const killConfirmMessage = ref('')
const pendingKillPid = ref(null)

// 信号菜单
const showSignalDropdown = ref(false)
const selectedPid = ref(null)
const signalMenuPosition = ref({ x: 0, y: 0 })

// 定时器
let statsTimer = null
let processesTimer = null

// 获取当前连接ID
const currentConnId = computed(() => sshLayoutStore.currentConnectionId)

// 获取系统统计
const fetchSystemStats = async () => {
  if (!currentConnId.value || currentConnId.value === 'default-connection') return
  
  try {
    const stats = await SSHService.GetSystemStats(currentConnId.value)
    systemStats.value = stats
    
    const currentTime = Date.now()
    const timeDiff = lastStatsTime.value > 0 ? (currentTime - lastStatsTime.value) / 1000 : 1 // 秒
    
    // 计算磁盘 I/O 速率（bytes/s）
    const currentDiskRead = stats?.disk?.ioStats?.readBytes || 0
    const currentDiskWrite = stats?.disk?.ioStats?.writeBytes || 0
    
    let diskIoRate = 0
    if (lastStatsTime.value > 0) {
      const readRate = (currentDiskRead - lastDiskRead.value) / timeDiff
      const writeRate = (currentDiskWrite - lastDiskWrite.value) / timeDiff
      diskIoRate = Math.max(0, readRate + writeRate) // 确保非负
    }
    
    diskIoHistory.value.push(diskIoRate)
    if (diskIoHistory.value.length > maxHistoryPoints) {
      diskIoHistory.value.shift()
    }
    
    // 更新上一次的值
    lastDiskRead.value = currentDiskRead
    lastDiskWrite.value = currentDiskWrite
    
    // 计算网络流量速率（bytes/s）
    const currentNetworkRx = stats?.network?.totalRx || 0
    const currentNetworkTx = stats?.network?.totalTx || 0
    
    let networkRxRate = 0
    let networkTxRate = 0
    if (lastStatsTime.value > 0) {
      networkRxRate = Math.max(0, (currentNetworkRx - lastNetworkRx.value) / timeDiff)
      networkTxRate = Math.max(0, (currentNetworkTx - lastNetworkTx.value) / timeDiff)
    }
    
    networkRxHistory.value.push(networkRxRate)
    networkTxHistory.value.push(networkTxRate)
    if (networkRxHistory.value.length > maxHistoryPoints) {
      networkRxHistory.value.shift()
      networkTxHistory.value.shift()
    }
    
    // 更新上一次的值
    lastNetworkRx.value = currentNetworkRx
    lastNetworkTx.value = currentNetworkTx
    lastStatsTime.value = currentTime
    
  } catch (error) {
    console.error('[MonitorPanel] 获取系统统计失败:', error)
  }
}

// 获取进程列表
const fetchProcessList = async () => {
  if (!currentConnId.value || currentConnId.value === 'default-connection') return
  
  loadingProcesses.value = true
  try {
    const processes = await SSHService.GetProcessList(currentConnId.value)
    processList.value = processes || []
  } catch (error) {
    console.error('[MonitorPanel] 获取进程列表失败:', error)
  } finally {
    loadingProcesses.value = false
  }
}

// 刷新进程
const refreshProcesses = () => {
  fetchProcessList()
}

// 终止进程
const killProcess = (pid) => {
  if (!currentConnId.value || currentConnId.value === 'default-connection') return
  
  // 显示自定义确认对话框
  pendingKillPid.value = pid
  killConfirmMessage.value = `确定要终止进程 ${pid} 吗？此操作不可恢复。`
  showKillConfirm.value = true
}

// 确认终止进程
const confirmKillProcess = async () => {
  if (!pendingKillPid.value) return
  
  try {
    await SSHService.KillProcess(currentConnId.value, pendingKillPid.value)
    showMessage('进程已终止', 'success')
    
    // 记录日志
    logFileOperation(currentConnId.value, 'kill_process', `PID: ${pendingKillPid.value}`)
    
    // 关闭对话框
    showKillConfirm.value = false
    pendingKillPid.value = null
    // 刷新进程列表
    setTimeout(() => fetchProcessList(), 500)
  } catch (error) {
    console.error('[MonitorPanel] 终止进程失败:', error)
    showMessage('终止进程失败', 'error')
  }
}

// 显示信号菜单
const showSignalMenu = (pid, event) => {
  event.stopPropagation()
  selectedPid.value = pid
  const rect = event.target.getBoundingClientRect()
  
  // 计算菜单位置，避免超出屏幕
  const menuWidth = 200
  const menuHeight = 280 // 估算菜单高度
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  
  let x = rect.left
  let y = rect.bottom + 5
  
  // 如果菜单会超出右边界，向左调整
  if (x + menuWidth > viewportWidth) {
    x = viewportWidth - menuWidth - 10
  }
  
  // 如果菜单会超出下边界，向上显示
  if (y + menuHeight > viewportHeight) {
    y = rect.top - menuHeight - 5
  }
  
  signalMenuPosition.value = { x, y }
  showSignalDropdown.value = true
  
  // 添加全局点击监听，点击外部关闭
  setTimeout(() => {
    document.addEventListener('click', closeSignalMenu)
  }, 0)
}

// 关闭信号菜单
const closeSignalMenu = () => {
  showSignalDropdown.value = false
  selectedPid.value = null
  document.removeEventListener('click', closeSignalMenu)
}

// 发送信号
const sendSignal = async (signal) => {
  if (!selectedPid.value) return
  
  try {
    await SSHService.SendSignal(currentConnId.value, selectedPid.value, signal)
    showMessage(`已发送 ${signal} 信号`, 'success')
    
    // 记录日志
    logSystemEvent(currentConnId.value, `发送信号 ${signal} 到进程 ${selectedPid.value}`, 'info')
    
    // 先关闭菜单
    closeSignalMenu()
    // 刷新进程列表
    setTimeout(() => fetchProcessList(), 500)
  } catch (error) {
    console.error('[MonitorPanel] 发送信号失败:', error)
    showMessage('发送信号失败', 'error')
  }
}

// 排序
const sortBy = (key) => {
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortOrder.value = 'desc'
  }
}

// 排序后的进程列表
const sortedProcesses = computed(() => {
  const processes = [...processList.value]
  processes.sort((a, b) => {
    let aVal = a[sortKey.value] || 0
    let bVal = b[sortKey.value] || 0
    
    if (typeof aVal === 'string') {
      aVal = aVal.toLowerCase()
      bVal = bVal.toLowerCase()
    }
    
    if (sortOrder.value === 'asc') {
      return aVal > bVal ? 1 : -1
    } else {
      return aVal < bVal ? 1 : -1
    }
  })
  return processes
})

// 格式化字节
const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

// 获取 CPU 渐变色
const getCpuGradient = () => {
  const percent = systemStats.value?.cpu?.usagePercent || 0
  const cs = getComputedStyle(document.documentElement)
  const bg = cs.getPropertyValue('--surface-3').trim() || '#2d3748'
  const green = cs.getPropertyValue('--accent-success').trim() || '#48bb78'
  const yellow = cs.getPropertyValue('--accent-warning').trim() || '#ecc94b'
  const red = cs.getPropertyValue('--danger-light').trim() || '#f56565'
  if (percent < 50) {
    return `conic-gradient(${green} ${percent}%, ${bg} ${percent}%)`
  } else if (percent < 80) {
    return `conic-gradient(${yellow} ${percent}%, ${bg} ${percent}%)`
  } else {
    return `conic-gradient(${red} ${percent}%, ${bg} ${percent}%)`
  }
}

// 获取使用率颜色类
const getUsageClass = (percent) => {
  if (percent > 80) return 'usage-high'
  if (percent > 50) return 'usage-medium'
  return 'usage-low'
}

// 生成 SVG 折线图数据点
const generateSparklinePoints = (data, width = 100, height = 30) => {
  if (!data || data.length === 0) return ''
  
  // 如果只有一个数据点，返回空
  if (data.length === 1) return `${width / 2},${height / 2}`
  
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1 // 避免除以零
  
  // 归一化并生成点坐标
  const points = data.map((value, index) => {
    const x = (index / (data.length - 1)) * width
    const normalizedValue = (value - min) / range
    const y = height - (normalizedValue * height * 0.8) - (height * 0.1) // 留上下边距
    return `${x.toFixed(1)},${y.toFixed(1)}`
  })
  
  return points.join(' ')
}

// 获取磁盘 I/O 图表数据点
const getDiskIoPoints = () => {
  return generateSparklinePoints(diskIoHistory.value)
}

// 获取网络接收图表数据点
const getNetworkRxPoints = () => {
  return generateSparklinePoints(networkRxHistory.value)
}

// 获取网络发送图表数据点
const getNetworkTxPoints = () => {
  return generateSparklinePoints(networkTxHistory.value)
}

// 获取当前磁盘 I/O 速率
const getCurrentDiskIoRate = () => {
  if (diskIoHistory.value.length === 0) return 0
  return diskIoHistory.value[diskIoHistory.value.length - 1]
}

// 获取当前网络接收速率
const getCurrentNetworkRxRate = () => {
  if (networkRxHistory.value.length === 0) return 0
  return networkRxHistory.value[networkRxHistory.value.length - 1]
}

// 获取当前网络发送速率
const getCurrentNetworkTxRate = () => {
  if (networkTxHistory.value.length === 0) return 0
  return networkTxHistory.value[networkTxHistory.value.length - 1]
}

// 生命周期
onMounted(() => {
  fetchSystemStats()
  fetchProcessList()
  
  // 每 2 秒更新系统统计
  statsTimer = setInterval(fetchSystemStats, 2000)
  
  // 每 5 秒更新进程列表
  processesTimer = setInterval(fetchProcessList, 5000)
})

onUnmounted(() => {
  if (statsTimer) clearInterval(statsTimer)
  if (processesTimer) clearInterval(processesTimer)
  // 清理事件监听器
  document.removeEventListener('click', closeSignalMenu)
})
</script>

<style scoped>
.monitor-panel {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-panel);
  overflow: hidden;
}

/* 概览卡片 - 限制最大高度，确保进程列表可见 */
.overview-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
  padding: 1rem;
  border-bottom: 1px solid var(--surface-hover);
  max-height: 45%;
  overflow-y: auto;
  flex-shrink: 0;
}

.stat-card {
  background: var(--surface-2, rgba(45, 45, 45, 0.6));
  border-radius: 0.5rem;
  padding: 1rem;
  border: 1px solid var(--border-default);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.card-icon {
  flex-shrink: 0;
}

.icon-cpu { color: var(--primary-light, #63b3ed); }
.icon-memory { color: var(--accent-success, #48bb78); }
.icon-disk { color: var(--accent-warning, #ecc94b); }
.icon-network { color: var(--warning-light, #f6ad55); }

.chart-line-primary { stroke: var(--primary-light, #63b3ed); }
.chart-line-success { stroke: var(--accent-success, #48bb78); }
.chart-line-warning { stroke: var(--warning-light, #f6ad55); }

.card-title {
  color: var(--text-primary, #e2e8f0);
  font-weight: 600;
  font-size: 0.875rem;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

/* CPU 圆形进度条 */
.usage-circle {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  position: relative;
}

.usage-circle::before {
  content: '';
  position: absolute;
  width: 60px;
  height: 60px;
  background: var(--bg-panel);
  border-radius: 50%;
}

.usage-text {
  position: relative;
  z-index: 1;
  color: var(--text-primary, #e2e8f0);
  font-size: 1rem;
  font-weight: 700;
}

/* 使用率进度条 */
.usage-bar {
  width: 100%;
  height: 8px;
  background: var(--surface-3, rgba(45, 45, 45, 0.8));
  border-radius: 4px;
  overflow: hidden;
}

.usage-bar.small {
  height: 6px;
}

.usage-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent-success, #48bb78), var(--success-light, #68d391));
  transition: width 0.3s ease;
}

.usage-high .usage-fill {
  background: linear-gradient(90deg, var(--danger-light, #f56565), var(--accent-danger, #fc8181));
}

.usage-medium .usage-fill {
  background: linear-gradient(90deg, var(--accent-warning, #ecc94b), var(--warning-light, #f6e05e));
}

/* 卡片详情 */
.card-details {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
}

.detail-item .label {
  color: var(--text-secondary, #a0aec0);
}

.detail-item .value {
  color: var(--text-primary, #e2e8f0);
  font-weight: 500;
}

.uptime-value {
  color: var(--primary-light, #63b3ed);
  font-weight: 600;
}

/* 磁盘列表 */
.disk-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.disk-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.disk-info {
  flex: 0 0 80px;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.disk-name {
  color: var(--text-primary, #e2e8f0);
  font-size: 0.75rem;
  font-family: monospace;
}

.disk-size {
  color: var(--text-secondary, #a0aec0);
  font-size: 0.625rem;
}

.disk-percent {
  flex: 0 0 40px;
  text-align: right;
  color: var(--text-primary, #e2e8f0);
  font-size: 0.75rem;
  font-weight: 600;
}

/* 网络统计 */
.network-stats {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.net-item {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
}

.net-label {
  color: var(--text-secondary, #a0aec0);
}

.net-value {
  color: var(--text-primary, #e2e8f0);
  font-weight: 500;
  font-family: monospace;
}

/* 图表容器 */
.chart-container {
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--surface-1);
  position: relative;
}

.sparkline-chart {
  width: 100%;
  height: 30px;
  display: block;
}

.chart-value {
  position: absolute;
  top: 0.25rem;
  right: 0.25rem;
  font-size: 0.6875rem;
  color: var(--primary-light, #63b3ed);
  font-weight: 600;
  font-family: monospace;
}

.network-rates {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.375rem;
  padding: 0 0.25rem;
}

.rate-item {
  font-size: 0.6875rem;
  font-weight: 600;
  font-family: monospace;
}

.rate-item.rx {
  color: var(--accent-success, #48bb78);
}

.rate-item.tx {
  color: var(--warning-light, #f6ad55);
}

/* 进程区域 */
.process-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--surface-hover);
}

.section-header h3 {
  margin: 0;
  color: var(--text-primary, #e2e8f0);
  font-size: 0.875rem;
  font-weight: 600;
}

.refresh-btn {
  padding: 0.375rem 0.75rem;
  background: var(--primary-bg, rgba(66, 153, 225, 0.2));
  border: 1px solid var(--border-accent, rgba(66, 153, 225, 0.5));
  border-radius: 0.25rem;
  color: var(--primary-light, #63b3ed);
  cursor: pointer;
  font-size: 0.75rem;
  transition: all 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--primary-bg-hover, rgba(66, 153, 225, 0.3));
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 进程表格 */
.process-table-container {
  flex: 1;
  overflow-y: auto;
}

.process-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.75rem;
}

.process-table thead {
  position: sticky;
  top: 0;
  background: var(--bg-toolbar);
  z-index: 1;
}

.process-table th {
  padding: 0.625rem 0.75rem;
  text-align: left;
  color: var(--text-secondary, #a0aec0);
  font-weight: 600;
  border-bottom: 2px solid var(--surface-hover);
  white-space: nowrap;
}

.process-table th.sortable {
  cursor: pointer;
  user-select: none;
}

.process-table th.sortable:hover {
  color: var(--text-primary, #e2e8f0);
}

.sort-indicator {
  margin-left: 0.25rem;
  color: var(--primary-light, #63b3ed);
}

.process-table td {
  padding: 0.5rem 0.75rem;
  color: var(--text-primary, #e2e8f0);
  border-bottom: 1px solid var(--surface-1);
}

.process-row:hover {
  background: var(--bg-hover, rgba(66, 153, 225, 0.1));
}

.process-name {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: monospace;
}

/* 使用率颜色 */
.usage-low {
  color: var(--success-light, #68d391);
}

.usage-medium {
  color: var(--accent-warning, #f6e05e);
}

.usage-high {
  color: var(--accent-danger, #fc8181);
  font-weight: 600;
}

/* 终止进程按钮 */
.kill-btn {
  padding: 0.25rem 0.5rem;
  background: var(--danger-bg, rgba(245, 101, 101, 0.2));
  border: 1px solid var(--border-danger, rgba(245, 101, 101, 0.5));
  border-radius: 0.25rem;
  color: var(--accent-danger, #fc8181);
  cursor: pointer;
  font-size: 0.6875rem;
  transition: all 0.2s;
}

.kill-btn:hover {
  background: var(--danger-bg, rgba(245, 101, 101, 0.3));
  border-color: var(--border-danger, rgba(245, 101, 101, 0.7));
}

/* 操作按钮组 */
.action-buttons {
  display: flex;
  gap: 0.375rem;
}

.signal-btn {
  padding: 0.25rem 0.5rem;
  background: var(--primary-bg, rgba(99, 179, 237, 0.2));
  border: 1px solid var(--border-accent, rgba(99, 179, 237, 0.5));
  border-radius: 0.25rem;
  color: var(--primary-light, #63b3ed);
  cursor: pointer;
  font-size: 0.6875rem;
  transition: all 0.2s;
}

.signal-btn:hover {
  background: var(--primary-bg-hover, rgba(99, 179, 237, 0.3));
  border-color: var(--border-accent, rgba(99, 179, 237, 0.7));
}

/* 信号下拉菜单 */
.signal-dropdown {
  position: fixed;
  background: var(--bg-tooltip, rgba(45, 45, 45, 0.98));
  border: 1px solid var(--border-strong);
  border-radius: 0.5rem;
  box-shadow: var(--shadow-lg, 0 10px 30px rgba(0, 0, 0, 0.5));
  padding: 0.5rem;
  min-width: 200px;
  z-index: 10000;
  animation: dropdownFadeIn 0.15s ease-out;
}

@keyframes dropdownFadeIn {
  from {
    opacity: 0;
    transform: translateY(-5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.dropdown-header {
  padding: 0.5rem;
  color: var(--text-secondary, #a0aec0);
  font-size: 0.75rem;
  border-bottom: 1px solid var(--surface-hover);
  margin-bottom: 0.375rem;
}

.dropdown-item {
  display: block;
  width: 100%;
  padding: 0.5rem 0.75rem;
  background: transparent;
  border: none;
  border-radius: 0.25rem;
  color: var(--text-primary, #e2e8f0);
  cursor: pointer;
  font-size: 0.8125rem;
  text-align: left;
  transition: all 0.15s;
}

.dropdown-item:hover {
  background: var(--primary-bg, rgba(66, 153, 225, 0.2));
}

.dropdown-item.danger {
  color: var(--accent-danger, #fc8181);
}

.dropdown-item.danger:hover {
  background: var(--danger-bg, rgba(245, 101, 101, 0.2));
}

/* 滚动条 */
.process-table-container::-webkit-scrollbar {
  width: 6px;
}

.process-table-container::-webkit-scrollbar-track {
  background: transparent;
}

.process-table-container::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb);
  border-radius: 3px;
}

.process-table-container::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover, rgba(255, 255, 255, 0.3));
}

/* === 响应式布局 === */

/* 中等宽度：2列卡片 */
@media (max-width: 900px) {
  .overview-cards {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
    padding: 0.75rem;
  }

  .stat-card {
    padding: 0.75rem;
  }

  .usage-circle {
    width: 60px;
    height: 60px;
  }

  .usage-text {
    font-size: 0.875rem;
  }

  .process-table-container {
    font-size: 0.7rem;
  }
}

/* 小宽度：1列卡片 */
@media (max-width: 500px) {
  .overview-cards {
    grid-template-columns: 1fr;
    gap: 0.5rem;
    padding: 0.5rem;
  }

  .stat-card {
    padding: 0.5rem;
  }

  .card-header {
    margin-bottom: 0.5rem;
  }

  .card-body {
    gap: 0.5rem;
  }

  .detail-item {
    font-size: 0.625rem;
  }

  .disk-info {
    flex: 0 0 60px;
  }

  .process-section {
    font-size: 0.625rem;
  }

  .section-header {
    padding: 0.5rem;
  }

  .process-table-container {
    max-height: 200px;
  }
}

/* 小高度：压缩卡片和进程区域 */
@media (max-height: 600px) {
  .overview-cards {
    gap: 0.375rem;
    padding: 0.375rem;
    max-height: 40%;
  }

  .stat-card {
    padding: 0.375rem;
  }

  .card-header {
    margin-bottom: 0.25rem;
    gap: 0.25rem;
  }

  .card-title {
    font-size: 0.75rem;
  }

  .card-body {
    gap: 0.375rem;
  }

  .usage-circle {
    width: 48px;
    height: 48px;
  }

  .usage-text {
    font-size: 0.625rem;
  }

  .usage-bar {
    height: 4px;
  }

  .usage-bar.small {
    height: 3px;
  }

  .detail-item {
    font-size: 0.625rem;
    gap: 0.125rem;
  }

  .disk-item {
    gap: 0.25rem;
  }

  .disk-info {
    flex: 0 0 50px;
  }

  .chart-container {
    height: 20px;
  }

  .chart-value {
    font-size: 0.625rem;
  }

  .process-section {
    min-height: 0;
  }

  .section-header {
    padding: 0.25rem 0.5rem;
  }

  .section-header h3 {
    font-size: 0.75rem;
  }

  .process-table-container {
    max-height: 150px;
    font-size: 0.625rem;
  }
}

/* 超小高度：极简模式 */
@media (max-height: 400px) {
  .overview-cards {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.25rem;
    padding: 0.25rem;
    max-height: 35%;
  }

  .stat-card {
    padding: 0.25rem;
  }

  .card-header {
    margin-bottom: 0.125rem;
  }

  .card-icon {
    width: 12px;
    height: 12px;
  }

  .card-title {
    font-size: 0.625rem;
  }

  .card-details {
    gap: 0.125rem;
  }

  .usage-circle {
    width: 36px;
    height: 36px;
  }

  .usage-text {
    font-size: 0.5rem;
  }

  .chart-container {
    display: none;
  }

  .process-table-container {
    max-height: 100px;
  }
}
</style>
