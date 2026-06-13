<template>
  <div class="file-manager">
    <!-- 工具栏 -->
    <FileToolbar
      :current-path="currentPath"
      :has-cut-file="hasCutFile"
      :search-keyword="searchKeyword"
      :selected-count="selectedCount"
      :is-searching="isSearching"
      @go-up="goUp"
      @refresh="refreshFiles"
      @navigate="(path) => { currentPath = path; navigateToPath() }"
      @paste="handlePaste"
      @new-folder="showNewFolderDialog = true"
      @upload="triggerUpload"
      @upload-directory="triggerDirectoryUpload"
      @search="handleSearch"
      @search-enter="executeRecursiveSearch"
      @cancel-search="cancelSearch"
      @batch-download="batchDownload"
      @batch-delete="batchDelete"
      @batch-chmod="batchChmod"
    />

    <!-- 文件列表 -->
    <div class="file-table-wrapper">
      <!-- 批量操作加载遮罩 -->
      <div v-if="operationLocks.batchDelete || operationLocks.batchDownload" class="operation-overlay">
        <div class="loading-spinner"></div>
        <div class="loading-text">
          {{ operationLocks.batchDelete ? '正在删除...' : '正在下载...' }}
        </div>
      </div>
      
      <FileTable
        :files="filteredFiles"
        :loading="loading"
        :sort-key="sortKey"
        :sort-order="sortOrder"
        :selected-files="selectedFiles"
        @open-file="openFile"
        @sort="sortBy"
        @show-menu="showActionMenu"
        @select-file="toggleFileSelection"
        @select-all="selectAllFiles"
      />
    </div>

    <!-- 状态栏 -->
    <div class="status-bar">
      <span>{{ files.length }} 个项目</span>
      <span v-if="selectedFile">| 已选择: {{ selectedFile.name }}</span>
    </div>

    <!-- 三点操作菜单 -->
    <ActionMenu
      v-model:visible="showContextMenu"
      :position="menuPosition"
      :file="selectedContextMenuFile"
      @preview="previewFile"
      @download="downloadFile"
      @compress="compressFile"
      @rename="renameFile"
      @duplicate="duplicateFile"
      @cut="cutFile"
      @chmod="chmodFile"
      @delete="deleteFile"
    />

    <!-- 文件预览/编辑对话框 -->
    <FilePreview
      v-model:visible="showPreviewDialog"
      :title="previewTitle"
      v-model:content="previewContent"
      :type="previewType"
      :editable="previewEditable"
      :file-name="previewFileName"
      @close="closePreview"
      @enable-edit="enableEdit"
      @save="saveEditedFile"
    />

    <!-- 权限编辑器 -->
    <PermissionEditor
      v-model:show-chmod-dialog="showChmodDialog"
      :chmod-target="chmodTarget"
      :permissions="chmodPermissions"
      :numeric-value="chmodValue"
      :apply-to-subdirs="applyToSubdirs"
      @confirm="handleChmodConfirm"
      @update:permission="handlePermissionUpdate"
      @update:apply-to-subdirs="applyToSubdirs = $event"
    />

    <!-- 确认对话框 -->
    <Modal
      v-model:visible="showConfirmDialog"
      :title="confirmTitle"
      @confirm="handleConfirm"
      @close="showConfirmDialog = false"
      @cancel="showConfirmDialog = false"
    >
      <p>{{ confirmMessage }}</p>
    </Modal>

    <!-- 输入对话框 -->
    <Modal
      v-model:visible="showInputDialog"
      :title="inputTitle"
      @confirm="handleInputConfirm"
      @close="showInputDialog = false"
      @cancel="showInputDialog = false"
    >
      <div class="input-dialog-content">
        <label>{{ inputLabel }}</label>
        <input 
          v-model="inputValue" 
          @keyup.enter="handleInputConfirm"
          class="dialog-input"
          :placeholder="inputPlaceholder"
          autofocus
        />
      </div>
    </Modal>

    <!-- 隐藏的文件上传输入 -->
    <input type="file" ref="fileInput" @change="handleFileUpload" style="display: none" multiple />
    
    <!-- 上传任务面板 -->
    <div v-if="showUploadPanel" class="upload-panel">
      <div class="upload-panel-header">
        <h3>上传任务 ({{ activeUploads.length }}/{{ uploadTasks.length }})</h3>
        <div class="upload-panel-actions">
          <button v-if="completedUploads.length > 0" @click="clearCompleted" class="btn-clear">清除已完成</button>
          <button @click="clearAll" class="btn-close">关闭</button>
        </div>
      </div>
      
      <div class="upload-panel-body">
        <div v-for="task in uploadTasks" :key="task.id" class="upload-task">
          <div class="task-info">
            <span class="task-name">{{ task.fileName }}</span>
            <span class="task-size">{{ formatFileSize(task.fileSize) }}</span>
          </div>
          
          <div class="task-progress">
            <div class="progress-bar">
              <div 
                class="progress-fill" 
                :class="task.status"
                :style="{ width: task.progress + '%' }"
              ></div>
            </div>
            <span class="progress-text">{{ task.progress }}%</span>
          </div>
          
          <div class="task-status">
            <span v-if="task.status === 'pending'" class="status-pending">等待中...</span>
            <span v-else-if="task.status === 'uploading'" class="status-uploading">上传中...</span>
            <span v-else-if="task.status === 'success'" class="status-success">✓ 成功</span>
            <span v-else-if="task.status === 'failed'" class="status-failed">✗ {{ task.errorMessage }}</span>
            <span v-else-if="task.status === 'cancelled'" class="status-cancelled">已取消</span>
            
            <div class="task-actions">
              <button 
                v-if="task.status === 'uploading'" 
                @click="cancelTask(task.id)" 
                class="btn-cancel"
              >
                取消
              </button>
              <button 
                v-if="task.status === 'failed'" 
                @click="retryTask(task.id)" 
                class="btn-retry"
              >
                重试
              </button>
            </div>
          </div>
        </div>
        
        <div v-if="uploadTasks.length === 0" class="empty-tasks">
          暂无上传任务
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, inject, nextTick } from 'vue'
import { SSHService } from '@bindings/changeme/ssh/index.js'
import { Events } from '@wailsio/runtime'
import { showMessage } from '@/utils/message'
import Modal from '@/components/Modal.vue'
import { logFileOperation, logError, logSystemEvent } from '@/utils/logger'

// 明确声明不接受任何 props，避免 Dockview params 干扰
const props = defineProps({})

// 从父组件注入 connId
const injectedConnId = inject('connId')
console.log('[FileManager] ========== 初始化 ==========')
console.log('[FileManager] injectedConnId:', injectedConnId)
const currentConnId = ref(injectedConnId)
console.log('[FileManager] currentConnId.value:', currentConnId.value)

// 导入子组件
import FileToolbar from './components/FileToolbar.vue'
import FileTable from './components/FileTable.vue'
import ActionMenu from './components/ActionMenu.vue'
import FilePreview from './components/FilePreview.vue'
import PermissionEditor from './components/PermissionEditor.vue'

// 导入 composables
import { useFileOperations } from './composables/useFileOperations'
import { usePermissions } from './composables/usePermissions'
import { useNavigation } from './composables/useNavigation'
import { useUploadTasks } from './composables/useUploadTasks'

// 导入工具函数
import { joinPath, getNumericMode } from './utils/pathHelper'

// 将 ArrayBuffer 转换为 base64 字符串（高效方式）
const arrayBufferToBase64 = (buffer) => {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  const chunkSize = 0x8000 // 32KB 分块
  
  for (let i = 0; i < bytes.length; i += chunkSize) {
    const chunk = bytes.subarray(i, i + chunkSize)
    binary += String.fromCharCode.apply(null, chunk)
  }
  
  return btoa(binary)
}

const fileInput = ref(null)

// 状态
const currentPath = ref('/')
const files = ref([])
const loading = ref(false)
const selectedFile = ref(null)
const sortKey = ref('name')
const sortOrder = ref('asc')
const showNewFolderDialog = ref(false)

// 搜索功能
const searchKeyword = ref('')

// 批量选择
const selectedFiles = ref(new Set())

// 搜索结果（用于全目录搜索）- 必须在 filteredFiles 之前定义
const searchResults = ref([])
const isSearching = ref(false)
const currentSearchID = ref('') // 当前搜索 ID
let searchResultUnsubscribe = null
let searchCompleteUnsubscribe = null

// 操作锁定状态（防止重复点击）
const operationLocks = ref({
  preview: false,
  download: false,
  delete: false,
  rename: false,
  duplicate: false,
  compress: false,
  chmod: false,
  batchDownload: false,  // 批量下载
  batchDelete: false     // 批量删除
})

// 自定义对话框状态
const showConfirmDialog = ref(false)
const confirmTitle = ref('')
const confirmMessage = ref('')
let confirmCallback = null

const showInputDialog = ref(false)
const inputTitle = ref('')
const inputLabel = ref('')
const inputValue = ref('')
const inputPlaceholder = ref('')
let inputCallback = null

// 预览/编辑对话框
const showPreviewDialog = ref(false)
const previewTitle = ref('')
const previewContent = ref('')
const previewType = ref('text')
const previewFileName = ref('')
const previewEditable = ref(false)
const previewTargetFile = ref(null)

// 权限修改对话框（直接使用 ref，不使用 composable）
const showChmodDialog = ref(false)
const chmodValue = ref('755')
const chmodTarget = ref(null)
const applyToSubdirs = ref(false)
const isBatchChmod = ref(false) // 是否是批量修改权限

// 权限复选框状态
const chmodPermissions = ref({
  owner: { r: false, w: false, x: false },
  group: { r: false, w: false, x: false },
  other: { r: false, w: false, x: false }
})

// 加载文件列表
async function loadFiles() {
  console.log('[FileManager] loadFiles 被调用')
  console.log('[FileManager] currentConnId.value:', currentConnId.value)
  console.log('[FileManager] currentPath.value:', currentPath.value)
  
  if (!currentConnId.value || currentConnId.value === 'default-connection') {
    console.warn('[FileManager] ⚠️ connId 无效，跳过加载')
    return
  }
  
  loading.value = true
  try {
    console.log('[FileManager] 调用 SSHService.ListFiles...')
    const fileList = await SSHService.ListFiles(currentConnId.value, currentPath.value)
    console.log('[FileManager] 收到文件列表:', fileList?.length || 0, '个文件')
    
    // 打印每个文件的详细信息
    if (fileList && fileList.length > 0) {
      fileList.forEach((file, index) => {
        console.log(`[FileManager] 文件[${index}]:`, {
          name: file.name,
          path: file.path,
          isDir: file.isDir,
          mode: file.mode,
          size: file.size
        })
      })
    }
    
    files.value = fileList || []
    
    // 记录浏览目录的日志
    logFileOperation(currentConnId.value, 'browse', currentPath.value)
  } catch (error) {
    console.error('[FileManager] 加载文件列表失败:', error)
    showMessage('加载文件列表失败', 'error')
    logError(currentConnId.value, '加载文件列表失败', error)
  } finally {
    loading.value = false
  }
}

// 规范化路径（避免双重斜杠等问题）
const normalizePath = (path) => {
  if (!path) return '/'
  // 替换多个连续斜杠为单个斜杠
  let normalized = path.replace(/\/+/g, '/')
  // 如果不是根目录且以斜杠结尾，去掉结尾斜杠
  if (normalized !== '/' && normalized.endsWith('/')) {
    normalized = normalized.slice(0, -1)
  }
  return normalized
}

// 使用 composables
const {
  uploadFiles,
  downloadFile,
  deleteFile: deleteFileOp,
  renameFile: renameFileOp,
  duplicateFile: duplicateFileOp,
  compressFile: compressFileOp,
  cutFile: cutFileOp,
  pasteFile: pasteFileOp,
  hasCutFile
} = useFileOperations(currentConnId, currentPath, loadFiles)

// 上传任务管理
const {
  uploadTasks,
  showUploadPanel,
  activeUploads,
  completedUploads,
  hasActiveUploads,
  uploadFiles: uploadFilesWithTasks,
  cancelTask,
  retryTask,
  clearCompleted,
  clearAll,
  formatFileSize,
  formatDuration
} = useUploadTasks(currentConnId, currentPath, loadFiles)

const {
  goUp,
  navigateToPath,
  openFile: openFileNav,
  sortBy,
  sortedFiles
} = useNavigation(currentPath, files, sortKey, sortOrder, loadFiles, selectedFiles)

// 计算属性 - 过滤后的文件列表
const filteredFiles = computed(() => {
  // 如果有搜索结果，显示搜索结果
  if (searchResults.value.length > 0) {
    return searchResults.value
  }
  
  // 否则显示当前目录的文件
  let result = sortedFiles.value
  
  // 应用搜索过滤（仅当前目录，实时响应）
  if (searchKeyword.value && searchResults.value.length === 0) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(file => 
      file.name.toLowerCase().includes(keyword)
    )
  }
  
  return result
})

// 处理搜索（输入时仅本地过滤）
const handleSearch = (keyword) => {
  searchKeyword.value = keyword
  
  // 如果清空搜索，清除结果并停止动画
  if (!keyword) {
    searchResults.value = []
    isSearching.value = false
    currentSearchID.value = ''
    
    // 取消事件监听
    if (searchResultUnsubscribe) {
      searchResultUnsubscribe()
      searchResultUnsubscribe = null
    }
    if (searchCompleteUnsubscribe) {
      searchCompleteUnsubscribe()
      searchCompleteUnsubscribe = null
    }
  }
  // 注意：输入时不触发递归搜索，只通过 computed 进行本地过滤
}

// 取消搜索
const cancelSearch = async () => {
  if (isSearching.value && currentSearchID.value) {
    console.log('[FileManager] 取消搜索:', currentSearchID.value)
    
    // 通知后端取消搜索
    try {
      await SSHService.CancelSearch(currentConnId.value, currentSearchID.value)
    } catch (error) {
      console.error('[FileManager] 取消搜索失败:', error)
    }
    
    isSearching.value = false
    searchResults.value = []
    currentSearchID.value = ''
    
    // 取消事件监听
    if (searchResultUnsubscribe) {
      searchResultUnsubscribe()
      searchResultUnsubscribe = null
    }
    if (searchCompleteUnsubscribe) {
      searchCompleteUnsubscribe()
      searchCompleteUnsubscribe = null
    }
    
    showMessage('已取消搜索', 'info')
  }
}

// 执行递归搜索（按回车时调用）- 流式版本
const executeRecursiveSearch = async () => {
  if (!searchKeyword.value) {
    searchResults.value = []
    return
  }
  
  // 生成唯一搜索 ID
  const searchID = `search_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  currentSearchID.value = searchID
  searchResults.value = []
  isSearching.value = true
  
  console.log('[FileManager] ========== 开始流式搜索 ==========')
  console.log('[FileManager] 搜索 ID:', searchID)
  console.log('[FileManager] 关键词:', searchKeyword.value)
  console.log('[FileManager] 路径:', currentPath.value)
  
  showMessage('正在全目录搜索...（点击 X 可取消，双击文件可停止）', 'info')
  
  // 设置事件监听器
  let resultCount = 0
  searchResultUnsubscribe = Events.On('search-result', (data) => {
    console.log('[FileManager] 收到 search-result 事件, data:', data)
    
    // Wails v3 事件数据可能在 data.data 中
    const eventData = data.data || data
    
    // 只处理当前搜索的结果
    if (eventData.searchID !== searchID) {
      console.log('[FileManager] 忽略非当前搜索结果:', eventData.searchID, '期望:', searchID)
      return
    }
    
    // 实时添加结果 - 使用新数组替换触发响应式更新
    resultCount++
    console.log(`[FileManager] 收到搜索结果 #${resultCount}:`, eventData.file.name)
    
    // 为搜索结果添加相对路径显示
    const fileWithDisplay = {
      ...eventData.file,
      displayName: eventData.file.name,
      relativePath: eventData.relativePath ? `(${eventData.relativePath})` : ''
    }
    
    // 创建新数组以触发 Vue 响应式更新
    searchResults.value = [...searchResults.value, fileWithDisplay]
    console.log('[FileManager] 当前搜索结果数量:', searchResults.value.length)
  })
  
  searchCompleteUnsubscribe = Events.On('search-complete', (data) => {
    console.log('[FileManager] 收到 search-complete 事件, data:', data)
    
    // Wails v3 事件数据可能在 data.data 中
    const eventData = data.data || data
    
    // 只处理当前搜索的完成事件
    if (eventData.searchID !== searchID) {
      console.log('[FileManager] 忽略非当前搜索完成事件:', eventData.searchID, '期望:', searchID)
      return
    }
    
    console.log('[FileManager] 搜索完成事件:', eventData)
    
    isSearching.value = false
    currentSearchID.value = ''
    
    // 取消事件监听
    if (searchResultUnsubscribe) {
      searchResultUnsubscribe()
      searchResultUnsubscribe = null
    }
    if (searchCompleteUnsubscribe) {
      searchCompleteUnsubscribe()
      searchCompleteUnsubscribe = null
    }
    
    if (eventData.error) {
      showMessage('搜索失败', 'error')
    } else if (eventData.totalFound > 0) {
      showMessage(`找到 ${eventData.totalFound} 个匹配的文件`, 'success')
    } else {
      showMessage('未找到匹配的文件', 'warning')
    }
  })
  
  try {
    console.log('[FileManager] 调用后端 SearchFiles API...')
    // 启动后端搜索（异步，不等待返回）
    await SSHService.SearchFiles(
      currentConnId.value,
      currentPath.value,
      searchKeyword.value,
      searchID
    )
    console.log('[FileManager] SearchFiles API 调用成功')
  } catch (error) {
    console.error('[FileManager] 启动搜索失败:', error)
    isSearching.value = false
    currentSearchID.value = ''
    
    // 取消事件监听
    if (searchResultUnsubscribe) {
      searchResultUnsubscribe()
      searchResultUnsubscribe = null
    }
    if (searchCompleteUnsubscribe) {
      searchCompleteUnsubscribe()
      searchCompleteUnsubscribe = null
    }
    
    showMessage('搜索启动失败', 'error')
  }
}

// 选中文件数量
const selectedCount = computed(() => selectedFiles.value.size)

// 三点菜单相关
const showContextMenu = ref(false)
const menuPosition = ref({ x: 0, y: 0 })
const selectedContextMenuFile = ref(null)

// 刷新文件
const refreshFiles = () => {
  loadFiles()
}

// 显示三点操作菜单
const showActionMenu = (file, event) => {
  event.stopPropagation()
  selectedContextMenuFile.value = file
  
  const rect = event.target.getBoundingClientRect()
  const menuWidth = 200
  const menuHeight = 400
  
  let x = rect.left
  let y = rect.bottom + 5
  
  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 10
  }
  
  if (y + menuHeight > window.innerHeight) {
    y = rect.top - menuHeight - 5
  }
  
  menuPosition.value = { x, y }
  showContextMenu.value = true
  
  setTimeout(() => {
    document.addEventListener('click', closeContextMenu)
  }, 0)
}

const closeContextMenu = () => {
  showContextMenu.value = false
  selectedContextMenuFile.value = null
  document.removeEventListener('click', closeContextMenu)
}

// 预览/编辑文件
const previewFile = async (file) => {
  closeContextMenu()
  
  // 检查是否正在操作
  if (operationLocks.value.preview) {
    showMessage('正在加载中，请稍候...', 'warning')
    return
  }
  
  if (file.isDir) {
    showMessage('无法预览目录', 'warning')
    return
  }
  
  previewFileName.value = file.name
  previewTargetFile.value = file
  
  const ext = file.name.split('.').pop().toLowerCase()
  
  // ✅ 图片格式
  const imageExts = ['png', 'jpg', 'jpeg', 'gif', 'bmp', 'svg', 'webp', 'ico', 'tiff', 'tif']
  
  // ✅ 视频格式
  const videoExts = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv', 'webm', 'm4v', '3gp']
  
  // ✅ 音频格式
  const audioExts = ['mp3', 'wav', 'ogg', 'flac', 'aac', 'wma', 'm4a']
  
  // ✅ 所有文本格式（支持编辑）
  const textExts = [
    // 编程语言
    'js', 'jsx', 'ts', 'tsx', 'py', 'java', 'c', 'cpp', 'h', 'hpp', 'cs', 'go', 'rs', 'rb', 'php', 'swift', 'kt', 'scala', 'r', 'm', 'mm',
    // Web 技术
    'html', 'htm', 'css', 'scss', 'less', 'vue', 'svelte', 'astro',
    // 配置/数据格式
    'json', 'xml', 'yaml', 'yml', 'toml', 'ini', 'conf', 'cfg', 'env', 'properties', 'sql',
    // 文档/标记
    'md', 'markdown', 'txt', 'log', 'csv', 'tsv',
    // Shell/脚本
    'sh', 'bash', 'zsh', 'fish', 'ps1', 'bat', 'cmd',
    // 其他
    'dockerfile', 'makefile', 'gradle', 'pom', 'gitignore', 'editorconfig'
  ]
  
  if (imageExts.includes(ext)) {
    previewType.value = 'image'
    previewTitle.value = `图片预览: ${file.name}`
    previewEditable.value = false
    
    // ✅ 检查文件大小，超过 10MB 提示用户
    const fileSizeMB = file.size / (1024 * 1024)
    if (fileSizeMB > 10) {
      showMessage(`图片文件较大 (${fileSizeMB.toFixed(1)} MB)，加载可能需要一些时间`, 'warning')
    }
    
    operationLocks.value.preview = true
    showMessage('正在加载图片...', 'info')
    
    try {
      const base64Data = await SSHService.DownloadFile(currentConnId.value, file.path)
      // 将 base64 转换为 Blob
      const binaryString = atob(base64Data)
      const bytes = new Uint8Array(binaryString.length)
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i)
      }
      const blob = new Blob([bytes])
      previewContent.value = URL.createObjectURL(blob)
      showPreviewDialog.value = true
    } catch (error) {
      console.error('[FileManager] 加载图片失败:', error)
      showMessage('加载图片失败', 'error')
    } finally {
      operationLocks.value.preview = false
    }
  } else if (videoExts.includes(ext)) {
    previewType.value = 'video'
    previewTitle.value = `视频预览: ${file.name}`
    previewEditable.value = false
    
    // ✅ 检查文件大小，超过 50MB 提示用户
    const fileSizeMB = file.size / (1024 * 1024)
    if (fileSizeMB > 50) {
      showMessage(`视频文件较大 (${fileSizeMB.toFixed(1)} MB)，建议下载到本地观看`, 'warning')
      return
    }
    
    operationLocks.value.preview = true
    showMessage('正在加载视频...', 'info')
    
    try {
      const base64Data = await SSHService.DownloadFile(currentConnId.value, file.path)
      // 将 base64 转换为 Blob
      const binaryString = atob(base64Data)
      const bytes = new Uint8Array(binaryString.length)
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i)
      }
      const blob = new Blob([bytes])
      previewContent.value = URL.createObjectURL(blob)
      showPreviewDialog.value = true
    } catch (error) {
      console.error('[FileManager] 加载视频失败:', error)
      showMessage('加载视频失败', 'error')
    } finally {
      operationLocks.value.preview = false
    }
  } else if (audioExts.includes(ext)) {
    previewType.value = 'audio'
    previewTitle.value = `音频预览: ${file.name}`
    previewEditable.value = false
    
    // ✅ 检查文件大小，超过 20MB 提示用户
    const fileSizeMB = file.size / (1024 * 1024)
    if (fileSizeMB > 20) {
      showMessage(`音频文件较大 (${fileSizeMB.toFixed(1)} MB)，建议下载到本地收听`, 'warning')
      return
    }
    
    operationLocks.value.preview = true
    showMessage('正在加载音频...', 'info')
    
    try {
      const base64Data = await SSHService.DownloadFile(currentConnId.value, file.path)
      // 将 base64 转换为 Blob
      const binaryString = atob(base64Data)
      const bytes = new Uint8Array(binaryString.length)
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i)
      }
      const blob = new Blob([bytes])
      previewContent.value = URL.createObjectURL(blob)
      showPreviewDialog.value = true
    } catch (error) {
      console.error('[FileManager] 加载音频失败:', error)
      showMessage('加载音频失败', 'error')
    } finally {
      operationLocks.value.preview = false
    }
  } else if (textExts.includes(ext)) {
    previewType.value = 'text'
    previewTitle.value = `查看文件: ${file.name}`
    previewEditable.value = false
    
    operationLocks.value.preview = true
    showMessage('正在加载文件...', 'info')
    
    try {
      const base64Data = await SSHService.DownloadFile(currentConnId.value, file.path)
      // 将 base64 转换为 ArrayBuffer
      const binaryString = atob(base64Data)
      const bytes = new Uint8Array(binaryString.length)
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i)
      }
      const decoder = new TextDecoder('utf-8')
      previewContent.value = decoder.decode(bytes)
      showPreviewDialog.value = true
    } catch (error) {
      console.error('[FileManager] 加载文件失败:', error)
      showMessage('加载文件失败', 'error')
    } finally {
      operationLocks.value.preview = false
    }
  } else {
    showMessage(`该文件格式暂不支持预览: .${ext}`, 'info')
  }
}

const enableEdit = () => {
  previewEditable.value = true
  previewTitle.value = `编辑文件: ${previewFileName.value}`
}

const saveEditedFile = async () => {
  if (!previewTargetFile.value) return
  
  try {
    // ✅ 从 FilePreview 组件获取最新内容（通过 codeEditorRef）
    // 注意：这里 previewContent 已经被 FilePreview 的 @update:content 更新了
    const encoder = new TextEncoder()
    const data = encoder.encode(previewContent.value)
    // 将 Uint8Array 转换为 base64 字符串
    const base64Data = btoa(String.fromCharCode(...data))
    await SSHService.UploadFile(currentConnId.value, previewTargetFile.value.path, base64Data)
    showMessage('保存成功', 'success')
    
    // ✅ 刷新文件列表
    loadFiles()
    
    // ✅ 关闭预览对话框
    closePreview()
  } catch (error) {
    console.error('[FileManager] 保存失败:', error)
    showMessage('保存失败', 'error')
  }
}

const closePreview = () => {
  showPreviewDialog.value = false
  // ✅ 清理所有类型的 URL（图片、视频、音频）
  if ((previewType.value === 'image' || previewType.value === 'video' || previewType.value === 'audio') && previewContent.value) {
    URL.revokeObjectURL(previewContent.value)
  }
  previewContent.value = ''
  previewEditable.value = false
  previewTargetFile.value = null
}

// 打开文件/目录
const openFile = (file) => {
  // 如果正在搜索，停止搜索
  if (isSearching.value) {
    cancelSearch()
  }
  
  openFileNav(file)
}

// 删除文件
const deleteFile = (file) => {
  closeContextMenu()
  
  console.log('[FileManager] 准备删除文件:', {
    name: file.name,
    path: file.path,
    isDir: file.isDir,
    mode: file.mode,
    size: file.size
  })
  
  showConfirmDialog.value = true
  confirmTitle.value = '确认删除'
  confirmMessage.value = `确定要删除 ${file.isDir ? '目录' : '文件'} "${file.name}" 吗？此操作不可恢复！`
  confirmCallback = async () => {
    try {
      await deleteFileOp(file)
      
      // 如果处于搜索状态，从搜索结果中移除该文件
      if (searchResults.value.length > 0) {
        searchResults.value = searchResults.value.filter(f => f.path !== file.path)
        console.log('[FileManager] 已从搜索结果中移除:', file.path)
        
        // 如果搜索结果变为空，清除搜索状态
        if (searchResults.value.length === 0) {
          showMessage('已删除，搜索结果已清空', 'success')
          searchResults.value = []
          isSearching.value = false
          currentSearchID.value = ''
        } else {
          showMessage(`已删除，剩余 ${searchResults.value.length} 个结果`, 'success')
        }
      }
    } catch (error) {
      // 错误已在 composable 中处理
    }
  }
}

const handleConfirm = () => {
  if (confirmCallback) {
    confirmCallback()
  }
  showConfirmDialog.value = false
}

// 重命名文件
const renameFile = (file) => {
  closeContextMenu()
  
  showInputDialog.value = true
  inputTitle.value = '重命名'
  inputLabel.value = '请输入新名称:'
  inputValue.value = file.name
  inputPlaceholder.value = '文件名'
  inputCallback = async (newName) => {
    try {
      await renameFileOp(file, newName)
    } catch (error) {
      // 错误已在 composable 中处理
    }
  }
}

const handleInputConfirm = () => {
  if (inputCallback) {
    inputCallback(inputValue.value)
  }
  showInputDialog.value = false
}

// 创建副本
const duplicateFile = async (file) => {
  closeContextMenu()
  try {
    await duplicateFileOp(file)
  } catch (error) {
    // 错误已在 composable 中处理
  }
}

// 剪切
const cutFile = (file) => {
  closeContextMenu()
  cutFileOp(file)
}

// 粘贴
const handlePaste = async () => {
  try {
    await pasteFileOp()
  } catch (error) {
    // 错误已在 composable 中处理
  }
}

// 压缩
const compressFile = async (file) => {
  closeContextMenu()
  try {
    await compressFileOp(file)
  } catch (error) {
    // 错误已在 composable 中处理
  }
}

// 修改权限
const chmodFile = (file) => {
  closeContextMenu()
  
  // 检查是否是符号链接
  if (file.mode && file.mode.startsWith('L')) {
    showMessage('符号链接的权限不可修改', 'warning')
    return
  }
  
  chmodTarget.value = file
  const numericMode = getNumericMode(file.mode)
  chmodValue.value = numericMode
  parsePermissions(numericMode)
  applyToSubdirs.value = false
  showChmodDialog.value = true
}

// 从复选框状态计算数字权限
const calculateNumericMode = () => {
  const calc = (perm) => {
    let val = 0
    if (perm.r) val += 4
    if (perm.w) val += 2
    if (perm.x) val += 1
    return val
  }
  
  const owner = calc(chmodPermissions.value.owner)
  const group = calc(chmodPermissions.value.group)
  const other = calc(chmodPermissions.value.other)
  
  chmodValue.value = `${owner}${group}${other}`
}

// 从数字权限转换为复选框状态
const parsePermissions = (numericMode) => {
  if (!numericMode || numericMode.length !== 3) return
  
  const owner = parseInt(numericMode[0])
  const group = parseInt(numericMode[1])
  const other = parseInt(numericMode[2])
  
  chmodPermissions.value = {
    owner: {
      r: (owner & 4) !== 0,
      w: (owner & 2) !== 0,
      x: (owner & 1) !== 0
    },
    group: {
      r: (group & 4) !== 0,
      w: (group & 2) !== 0,
      x: (group & 1) !== 0
    },
    other: {
      r: (other & 4) !== 0,
      w: (other & 2) !== 0,
      x: (other & 1) !== 0
    }
  }
}

const handleChmodConfirm = async () => {
  if (!chmodTarget.value) return
  
  try {
    // 如果是批量操作
    if (isBatchChmod.value) {
      const selectedFileList = filteredFiles.value.filter(f => selectedFiles.value.has(f.name))
      
      if (selectedFileList.length === 0) {
        showMessage('没有选中的文件', 'warning')
        return
      }
      
      showConfirmDialog.value = true
      confirmTitle.value = '确认批量修改权限'
      confirmMessage.value = `确定要修改选中的 ${selectedFileList.length} 个${selectedFileList[0].isDir ? '目录' : '文件'}的权限为 ${chmodValue.value} 吗？`
      confirmCallback = async () => {
        let successCount = 0
        let failCount = 0
        
        for (const file of selectedFileList) {
          try {
            const normalizedPath = normalizePath(file.path)
            let command = `chmod ${chmodValue.value} "${normalizedPath}"`
            
            // 如果勾选了应用到子目录且是目录
            if (applyToSubdirs.value && file.isDir) {
              command = `chmod -R ${chmodValue.value} "${normalizedPath}"`
            }
            
            console.log('[FileManager] 执行权限修改命令:', command)
            const result = await SSHService.ExecuteCommand(currentConnId.value, command)
            
            const success = result.Success || result.success || false
            if (success) {
              successCount++
            } else {
              failCount++
              console.error(`[FileManager] 修改失败: ${file.name}`, result.Stderr || result.stderr)
            }
          } catch (error) {
            console.error(`[FileManager] 修改失败: ${file.name}`, error)
            failCount++
          }
        }
        
        selectedFiles.value.clear()
        await loadFiles()
        
        if (failCount === 0) {
          showMessage(`成功修改 ${successCount} 个文件的权限`, 'success')
        } else {
          showMessage(`修改完成: ${successCount} 成功, ${failCount} 失败`, 'warning')
        }
        
        isBatchChmod.value = false
        showChmodDialog.value = false
      }
      return
    }
    
    // 单个文件权限修改
    const normalizedPath = normalizePath(chmodTarget.value.path)
    let command = `chmod ${chmodValue.value} "${normalizedPath}"`
    
    // 如果勾选了应用到子目录
    if (applyToSubdirs.value && chmodTarget.value.isDir) {
      command = `chmod -R ${chmodValue.value} "${normalizedPath}"`
      console.log('[FileManager] ✓ 递归修改权限（包括子目录）')
    } else {
      console.log('[FileManager] 仅修改当前项权限')
    }
    
    console.log('[FileManager] 执行权限修改命令:', command)
    const result = await SSHService.ExecuteCommand(currentConnId.value, command)
    console.log('[FileManager] ExecuteCommand 返回结果:', result)
    
    // 检查返回结果
    if (result === null || result === undefined) {
      // ExecuteCommand 在连接不存在时返回 null
      throw new Error('连接不存在或命令执行失败')
    }
    
    // Wails v3 的 ExecuteCommand 返回的是 CommandResult 对象
    // 尝试多种字段名（大写和小写）
    const success = result.Success || result.success || false
    const stdout = result.Stdout || result.stdout || ''
    const stderr = result.Stderr || result.stderr || ''
    
    console.log('[FileManager] success:', success)
    console.log('[FileManager] stdout:', stdout)
    console.log('[FileManager] stderr:', stderr)
    
    if (success) {
      console.log('[FileManager] ✓ 权限修改成功，刷新文件列表...')
      
      // 如果有 stderr 输出，显示警告
      if (stderr && stderr.trim()) {
        console.warn('[FileManager] ⚠️ 权限修改有警告:', stderr)
        showMessage(`权限修改成功，但有警告: ${stderr}`, 'warning')
      } else {
        showMessage('权限修改成功', 'success')
      }
      
      // 强制刷新文件列表
      await loadFiles()
      console.log('[FileManager] 文件列表已刷新')
      isBatchChmod.value = false // 重置标志
      showChmodDialog.value = false
    } else {
      throw new Error(stderr || '权限修改失败')
    }
  } catch (error) {
    console.error('[FileManager] 权限修改失败:', error)
    showMessage(`权限修改失败: ${error.message}`, 'error')
  }
}

// 处理权限更新
const handlePermissionUpdate = ({ group, perm, value }) => {
  chmodPermissions.value[group][perm] = value
}

// 监听权限复选框变化，自动计算数字权限
watch(chmodPermissions, () => {
  calculateNumericMode()
}, { deep: true })

// 触发上传
const triggerUpload = () => {
  fileInput.value?.click()
}

// 处理文件上传
const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files || files.length === 0) return
  
  // 将 FileList 转换为数组
  const filesArray = Array.from(files)
  
  console.log('[FileManager] 选择文件:', filesArray.length, '个')
  
  try {
    // 检查是否为压缩包（单个 zip/tar.gz/tar 文件）
    if (filesArray.length === 1) {
      const file = filesArray[0]
      const fileName = file.name.toLowerCase()
      
      if (fileName.endsWith('.zip') || fileName.endsWith('.tar.gz') || 
          fileName.endsWith('.tgz') || fileName.endsWith('.tar')) {
        // 压缩包上传，自动解压
        console.log('[FileManager] 检测到压缩包:', fileName)
        showMessage('正在上传压缩包...', 'info')
        
        // 1. 先上传压缩包
        await uploadFilesWithTasks(filesArray)
        
        // 2. 等待上传完成
        showMessage('正在解压...', 'info')
        
        // 3. 解压到当前目录
        const remoteArchivePath = normalizePath(joinPath(currentPath.value, file.name))
        await SSHService.ExtractArchive(currentConnId.value, remoteArchivePath, currentPath.value)
        
        // 4. 删除临时压缩包
        await SSHService.DeleteTempFile(currentConnId.value, remoteArchivePath)
        
        showMessage('解压成功', 'success')
        logFileOperation(currentConnId.value, 'extract', file.name)
        loadFiles()
        event.target.value = ''
        return
      }
    }
    
    // 普通文件上传
    await uploadFilesWithTasks(filesArray)
  } catch (error) {
    console.error('[FileManager] 上传失败:', error)
    showMessage('上传失败: ' + error.message, 'error')
  }
  
  event.target.value = ''
}

// 监听新建文件夹对话框
watch(showNewFolderDialog, (newValue) => {
  if (newValue) {
    inputTitle.value = '新建文件夹'
    inputLabel.value = '请输入文件夹名称:'
    inputValue.value = ''
    inputPlaceholder.value = '文件夹名称'
    inputCallback = async (folderName) => {
      if (!folderName) return
      
      const folderPath = `${currentPath.value}/${folderName}`
      
      try {
        await SSHService.CreateDirectory(currentConnId.value, folderPath)
        showMessage('创建成功', 'success')
        logFileOperation(currentConnId.value, 'create', folderPath)
        loadFiles()
      } catch (error) {
        console.error('[FileManager] 创建失败:', error)
        showMessage('创建失败', 'error')
      }
    }
    showInputDialog.value = true
    showNewFolderDialog.value = false
  }
})

// 监听权限对话框关闭，重置批量标志
watch(showChmodDialog, (newValue) => {
  if (!newValue) {
    // 对话框关闭时重置标志
    isBatchChmod.value = false
  }
})

// 生命周期
let resizeTimeout = null
const handleResize = () => {
  // 防抖处理
  if (resizeTimeout) {
    clearTimeout(resizeTimeout)
  }
  
  resizeTimeout = setTimeout(async () => {
    console.log('[FileManager] 📐 窗口 resize，强制刷新布局')
    await nextTick()
    // 触发一个自定义事件让 FileTable 重新计算高度
    window.dispatchEvent(new CustomEvent('filemanager-resize'))
    resizeTimeout = null
  }, 150)
}

onMounted(() => {
  console.log('[FileManager] 🎬 组件挂载, connId:', currentConnId.value)
  loadFiles()
  
  // 注册 resize 监听器
  window.addEventListener('resize', handleResize)
  console.log('[FileManager] 👂 已注册 resize 监听器')
  
  // 注册目录上传进度事件监听器
  Events.On('directory-upload-progress', (event) => {
    const data = event.data || event
    console.log('[FileManager] 收到目录上传进度:', data)
    
    // 如果有正在进行的目录上传任务，更新它
    if (directoryUploadTask && directoryUploadTask.status === 'uploading') {
      directoryUploadTask.fileName = data.message || '处理中...'
      directoryUploadTask.progress = data.progress || 0
      
      // 如果收到了文件大小，更新任务
      if (data.fileSize) {
        directoryUploadTask.fileSize = data.fileSize
      }
    }
  })
})

onUnmounted(() => {
  console.log('[FileManager] 🔇 组件卸载')
  
  // 清理 resize 监听器
  if (resizeTimeout) {
    clearTimeout(resizeTimeout)
  }
  window.removeEventListener('resize', handleResize)
  
  document.removeEventListener('click', closeContextMenu)
  
  // 清理搜索事件监听器
  if (searchResultUnsubscribe) {
    searchResultUnsubscribe()
    searchResultUnsubscribe = null
  }
  if (searchCompleteUnsubscribe) {
    searchCompleteUnsubscribe()
    searchCompleteUnsubscribe = null
  }
})

// 切换文件选择状态
const toggleFileSelection = (fileName) => {
  if (selectedFiles.value.has(fileName)) {
    selectedFiles.value.delete(fileName)
  } else {
    selectedFiles.value.add(fileName)
  }
  // 触发响应式更新
  selectedFiles.value = new Set(selectedFiles.value)
}

// 全选/取消全选
const selectAllFiles = (checked) => {
  if (checked) {
    // 全选当前显示的文件（可能是搜索结果）
    selectedFiles.value = new Set(filteredFiles.value.map(f => f.name))
  } else {
    // 取消全选
    selectedFiles.value.clear()
  }
}

// 触发目录上传（使用 Wails 原生对话框）
let directoryUploadTask = null // 当前目录上传任务

const triggerDirectoryUpload = async () => {
  try {
    console.log('[FileManager] 开始目录上传')
    
    // 创建一个虚拟任务来显示进度
    const taskId = `dir_upload_${Date.now()}`
    directoryUploadTask = {
      id: taskId,
      fileName: '正在选择目录...',
      fileSize: 0,
      uploadedSize: 0,
      progress: 0,
      status: 'uploading',
      errorMessage: null,
      startTime: Date.now(),
      endTime: null,
      customPath: null
    }
    
    // 添加到任务列表
    uploadTasks.value.unshift(directoryUploadTask)
    // ✅ 不再需要手动设置 showUploadPanel，watch 会自动处理
    
    // 调用后端方法，使用原生对话框选择目录并自动压缩上传
    showMessage('正在选择目录...', 'info')
    directoryUploadTask.fileName = '正在选择目录...'
    directoryUploadTask.progress = 5
    
    await SSHService.SelectLocalDirectoryAndUpload(currentConnId.value, currentPath.value)
    
    // 更新进度
    directoryUploadTask.fileName = '目录上传成功'
    directoryUploadTask.progress = 100
    directoryUploadTask.status = 'success'
    directoryUploadTask.endTime = Date.now()
    
    showMessage('目录上传成功', 'success')
    logFileOperation(currentConnId.value, 'upload', currentPath.value)
    loadFiles()
    
    // 2秒后自动清除任务并关闭面板
    setTimeout(() => {
      clearAll()
    }, 2000)
    
  } catch (error) {
    if (error && error.message && error.message.includes('取消')) {
      console.log('[FileManager] 用户取消选择')
      // 移除任务
      const taskIndex = uploadTasks.value.findIndex(t => t.id === directoryUploadTask?.id)
      if (taskIndex !== -1) {
        uploadTasks.value.splice(taskIndex, 1)
      }
      return
    }
    console.error('[FileManager] 目录上传失败:', error)
    
    // 更新任务状态为失败
    if (directoryUploadTask) {
      directoryUploadTask.status = 'failed'
      directoryUploadTask.errorMessage = error.message
      directoryUploadTask.endTime = Date.now()
    }
    
    showMessage('目录上传失败: ' + error.message, 'error')
  } finally {
    directoryUploadTask = null
  }
}

// 批量下载
const batchDownload = async () => {
  if (selectedFiles.value.size === 0) return
  
  // 检查是否正在操作
  if (operationLocks.value.batchDownload) {
    showMessage('正在下载中，请稍候...', 'warning')
    return
  }
  
  operationLocks.value.batchDownload = true
  
  try {
    const selectedFileList = filteredFiles.value.filter(f => selectedFiles.value.has(f.name))
    
    if (selectedFileList.length === 0) {
      showMessage('没有选中的文件', 'warning')
      return
    }
    
    // 生成压缩包名称（使用时间戳）
    const timestamp = new Date().getTime()
    const archiveName = `batch_download_${timestamp}`
    
    // 收集所有文件路径
    const filePaths = selectedFileList.map(f => f.path)
    
    console.log('[FileManager] 开始批量下载，文件数:', filePaths.length)
    console.log('[FileManager] 文件列表:', filePaths)
    
    showMessage(`正在创建压缩包 (${filePaths.length} 个文件)...`, 'info')
    
    // 1. 后端创建压缩包
    const tempArchivePath = await SSHService.CreateArchive(currentConnId.value, filePaths, archiveName)
    
    if (!tempArchivePath) {
      throw new Error('创建压缩包失败')
    }
    
    console.log('[FileManager] 压缩包创建成功:', tempArchivePath)
    showMessage('正在下载压缩包...', 'info')
    
    // 2. 下载压缩包
    const base64Data = await SSHService.DownloadFile(currentConnId.value, tempArchivePath)
    
    // 3. 转换为 Blob 并触发下载
    const binaryString = atob(base64Data)
    const bytes = new Uint8Array(binaryString.length)
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i)
    }
    const blob = new Blob([bytes], { type: 'application/gzip' })
    const url = URL.createObjectURL(blob)
    
    const a = document.createElement('a')
    a.href = url
    a.download = `${archiveName}.tar.gz`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    
    console.log('[FileManager] 下载完成')
    showMessage('下载成功', 'success')
    
    // 4. 删除临时文件
    console.log('[FileManager] 删除临时文件:', tempArchivePath)
    await SSHService.DeleteTempFile(currentConnId.value, tempArchivePath)
    
    // 记录批量下载日志
    selectedFileList.forEach(file => {
      logFileOperation(currentConnId.value, 'download', file.path)
    })
    
    // 清空选择
    selectedFiles.value.clear()
    
  } catch (error) {
    console.error('[FileManager] 批量下载失败:', error)
    showMessage('批量下载失败: ' + error.message, 'error')
  } finally {
    operationLocks.value.batchDownload = false
  }
}

// 批量删除
const batchDelete = async () => {
  if (selectedFiles.value.size === 0) return
  
  // 检查是否正在操作
  if (operationLocks.value.batchDelete) {
    showMessage('正在删除中，请稍候...', 'warning')
    return
  }
  
  showConfirmDialog.value = true
  confirmTitle.value = '确认批量删除'
  confirmMessage.value = `确定要删除选中的 ${selectedFiles.value.size} 个文件/目录吗？此操作不可恢复！`
  confirmCallback = async () => {
    operationLocks.value.batchDelete = true
    let successCount = 0
    let failCount = 0
    const totalCount = selectedFiles.value.size
    
    try {
      // 显示开始消息
      showMessage(`正在删除 ${totalCount} 个文件...`, 'info')
      
      // 批量删除时不立即刷新，最后统一刷新
      for (const fileName of selectedFiles.value) {
        const file = filteredFiles.value.find(f => f.name === fileName)
        if (file) {
          try {
            const normalizedPath = normalizePath(file.path)
            await SSHService.DeleteFile(currentConnId.value, normalizedPath)
            
            // 记录删除日志
            logFileOperation(currentConnId.value, 'delete', normalizedPath)
            
            // 如果处于搜索状态，从搜索结果中移除该文件
            if (searchResults.value.length > 0) {
              searchResults.value = searchResults.value.filter(f => f.path !== file.path)
            }
            
            successCount++
            
            // 实时更新进度提示
            const remaining = totalCount - successCount - failCount
            if (remaining > 0) {
              showMessage(`正在删除... (${successCount + failCount}/${totalCount})`, 'info')
            }
          } catch (error) {
            console.error(`[FileManager] 删除失败: ${fileName}`, error)
            failCount++
          }
        }
      }
      
      selectedFiles.value.clear()
      
      // 统一刷新一次
      loadFiles()
      
      if (failCount === 0) {
        showMessage(`成功删除 ${successCount} 个文件`, 'success')
      } else {
        showMessage(`删除完成: ${successCount} 成功, ${failCount} 失败`, 'warning')
      }
    } finally {
      operationLocks.value.batchDelete = false
    }
  }
}

// 批量修改权限
const batchChmod = () => {
  if (selectedFiles.value.size === 0) return
  
  // 检查是否都是文件或目录（不能混合）
  const selectedFileList = filteredFiles.value.filter(f => selectedFiles.value.has(f.name))
  const hasDir = selectedFileList.some(f => f.isDir)
  const hasFile = selectedFileList.some(f => !f.isDir)
  
  if (hasDir && hasFile) {
    showMessage('不能同时修改文件和目录的权限', 'warning')
    return
  }
  
  // 使用第一个选中文件作为目标
  const firstFile = selectedFileList[0]
  if (!firstFile) return
  
  chmodTarget.value = firstFile
  const numericMode = getNumericMode(firstFile.mode)
  chmodValue.value = numericMode
  parsePermissions(numericMode)
  applyToSubdirs.value = false
  isBatchChmod.value = true // 标记为批量操作
  showChmodDialog.value = true
}
</script>

<style scoped>
.file-manager {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-panel);
  overflow: hidden;
}

/* ✅ 文件表格容器 - 占据剩余空间 */
.file-table-wrapper {
  flex: 1;
  min-height: 0; /* ✅ 关键：允许 flex 子项缩小 */
  position: relative;
  overflow: hidden; /* ✅ 防止内容溢出 */
}

.status-bar {
  padding: 0.5rem 1rem;
  border-top: 1px solid var(--surface-hover);
  background: var(--surface-2);
  color: var(--text-secondary);
  font-size: 0.75rem;
}

.input-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.input-dialog-content label {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.dialog-input {
  padding: 0.5rem 0.75rem;
  background: var(--bg-input);
  border: 1px solid var(--border-strong);
  border-radius: 0.375rem;
  color: var(--text-primary);
  font-size: 0.875rem;
}

.dialog-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

/* 上传任务面板 */
.upload-panel {
  position: absolute;
  bottom: 40px;
  right: 20px;
  width: 400px;
  max-height: 500px;
  background: var(--bg-tooltip);
  border: 1px solid var(--surface-hover);
  border-radius: 8px;
  box-shadow: var(--shadow-md);
  display: flex;
  flex-direction: column;
  z-index: 1000;
}

.upload-panel-header {
  padding: 1rem;
  border-bottom: 1px solid var(--surface-hover);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.upload-panel-header h3 {
  margin: 0;
  color: var(--text-primary);
  font-size: 0.875rem;
}

.upload-panel-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-clear,
.btn-close {
  padding: 0.25rem 0.75rem;
  background: var(--primary-bg);
  border: 1px solid var(--border-accent);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-clear:hover,
.btn-close:hover {
  background: var(--primary-bg-hover);
}

.upload-panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem;
}

.upload-task {
  padding: 0.75rem;
  background: var(--surface-2);
  border-radius: 6px;
  margin-bottom: 0.5rem;
}

.task-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.task-name {
  color: var(--text-primary);
  font-size: 0.875rem;
  font-weight: 500;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-size {
  color: var(--text-secondary);
  font-size: 0.75rem;
  margin-left: 1rem;
}

.task-progress {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: var(--surface-hover);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  transition: width 0.3s ease;
}

.progress-fill.pending {
  background: var(--text-secondary);
}

.progress-fill.uploading {
  background: linear-gradient(90deg, var(--accent-primary), var(--primary-light));
}

.progress-fill.success {
  background: var(--accent-success);
}

.progress-fill.failed {
  background: var(--danger-light);
}

.progress-fill.cancelled {
  background: var(--text-secondary);
}

.progress-text {
  color: var(--text-secondary);
  font-size: 0.75rem;
  min-width: 40px;
  text-align: right;
}

.task-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-pending,
.status-uploading,
.status-success,
.status-failed,
.status-cancelled {
  font-size: 0.75rem;
}

.status-pending {
  color: var(--text-secondary);
}

.status-uploading {
  color: var(--accent-primary);
}

.status-success {
  color: var(--accent-success);
}

.status-failed {
  color: var(--danger-light);
}

.status-cancelled {
  color: var(--text-secondary);
}

.task-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-cancel,
.btn-retry {
  padding: 0.25rem 0.5rem;
  background: transparent;
  border: 1px solid var(--scrollbar-thumb);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: var(--danger-bg);
  border-color: var(--danger-light);
}

.btn-retry:hover {
  background: var(--primary-bg);
  border-color: var(--accent-primary);
}

.empty-tasks {
  text-align: center;
  color: var(--text-secondary);
  padding: 2rem;
  font-size: 0.875rem;
}

/* 批量操作加载遮罩 */
.operation-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-overlay);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 100;
  backdrop-filter: blur(4px);
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 4px solid var(--primary-bg);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  margin-top: 1rem;
  color: var(--text-primary);
  font-size: 1rem;
  font-weight: 500;
}

/* === 响应式布局 === */

/* 小宽度 */
@media (max-width: 600px) {
  .status-bar {
    padding: 0.25rem 0.5rem;
    font-size: 0.625rem;
  }

  .upload-panel {
    left: 10px;
    right: 10px;
    width: auto;
    bottom: 30px;
  }

  .upload-panel-header {
    padding: 0.5rem;
  }

  .upload-panel-header h3 {
    font-size: 0.75rem;
  }

  .upload-task {
    padding: 0.5rem;
  }

  .task-name {
    font-size: 0.75rem;
  }

  .task-size {
    font-size: 0.625rem;
  }

  .dialog-input {
    font-size: 0.75rem;
    padding: 0.375rem 0.5rem;
  }
}

/* 小高度 */
@media (max-height: 500px) {
  .status-bar {
    padding: 0.125rem 0.5rem;
    font-size: 0.625rem;
  }

  .upload-panel {
    max-height: 250px;
    bottom: 25px;
  }

  .upload-panel-header {
    padding: 0.375rem 0.5rem;
  }

  .upload-panel-body {
    padding: 0.375rem;
  }

  .upload-task {
    padding: 0.375rem;
    margin-bottom: 0.25rem;
  }

  .task-info {
    margin-bottom: 0.25rem;
  }

  .task-progress {
    margin-bottom: 0.25rem;
  }

  .loading-spinner {
    width: 30px;
    height: 30px;
    border-width: 3px;
  }

  .loading-text {
    font-size: 0.75rem;
    margin-top: 0.5rem;
  }

  .empty-tasks {
    padding: 1rem;
    font-size: 0.75rem;
  }
}
</style>
