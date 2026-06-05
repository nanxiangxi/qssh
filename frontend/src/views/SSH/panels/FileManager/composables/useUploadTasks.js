import { ref, computed, watch, onUnmounted } from 'vue'
import { SSHService } from '@bindings/changeme/ssh/index.js'
import { Events } from '@wailsio/runtime'
import { showMessage } from '@/utils/message'
import { joinPath } from '../utils/pathHelper'
import { logFileOperation, logError } from '@/utils/logger'

/**
 * 将 ArrayBuffer 转换为 base64 字符串（高效方式，分块处理避免栈溢出）
 */
function arrayBufferToBase64(buffer) {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  const chunkSize = 0x8000 // 32KB 分块
  
  for (let i = 0; i < bytes.length; i += chunkSize) {
    const chunk = bytes.subarray(i, i + chunkSize)
    binary += String.fromCharCode.apply(null, chunk)
  }
  
  return btoa(binary)
}

/**
 * 上传任务状态
 */
export const UploadStatus = {
  PENDING: 'pending',      // 等待中
  UPLOADING: 'uploading',  // 上传中
  SUCCESS: 'success',      // 成功
  FAILED: 'failed',        // 失败
  CANCELLED: 'cancelled'   // 已取消
}

/**
 * 上传任务管理 Composable
 * @param {Ref} currentConnId - 当前连接ID
 * @param {Ref} currentPath - 当前路径
 * @param {Function} loadFiles - 重新加载文件列表的函数
 */
export function useUploadTasks(currentConnId, currentPath, loadFiles) {
  // 上传任务列表
  const uploadTasks = ref([])
  
  // 是否显示上传面板
  const showUploadPanel = ref(false)

  // 计算属性
  const activeUploads = computed(() => {
    return uploadTasks.value.filter(task => 
      task.status === UploadStatus.UPLOADING || task.status === UploadStatus.PENDING
    )
  })
  
  const completedUploads = computed(() => {
    return uploadTasks.value.filter(task => 
      task.status === UploadStatus.SUCCESS || task.status === UploadStatus.FAILED
    )
  })
  
  const hasActiveUploads = computed(() => {
    return activeUploads.value.length > 0
  })

  // ✅ 自动管理面板显示/隐藏：有任务时显示，无任务时隐藏
  watch(uploadTasks, (newTasks) => {
    if (newTasks.length > 0) {
      showUploadPanel.value = true
    } else {
      showUploadPanel.value = false
    }
  }, { deep: true })

  // 创建上传任务
  const createUploadTask = (file, customPath = null) => {
    const taskId = `upload_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    const task = {
      id: taskId,
      file: file,
      fileName: file.name,
      fileSize: file.size,
      uploadedSize: 0,
      progress: 0,
      status: UploadStatus.PENDING,
      errorMessage: null,
      startTime: Date.now(),
      endTime: null,
      customPath: customPath // 自定义目标路径（可选）
    }
    
    uploadTasks.value.unshift(task)
    // ✅ 不再需要手动设置 showUploadPanel，watch 会自动处理
    
    return task
  }

  // 通过数组索引更新任务（确保 Vue 响应式）
  const updateTask = (taskId, updates) => {
    const idx = uploadTasks.value.findIndex(t => t.id === taskId)
    if (idx !== -1) {
      uploadTasks.value[idx] = { ...uploadTasks.value[idx], ...updates }
    }
  }

  // 执行单个文件上传
  const executeUpload = async (task) => {
    const taskId = task.id
    try {
      updateTask(taskId, { status: UploadStatus.UPLOADING })
      console.log('[Upload] 开始上传:', task.fileName)

      const reader = new FileReader()

      return new Promise((resolve, reject) => {
        reader.onload = async (e) => {
          try {
            const arrayBuffer = e.target.result
            const remotePath = task.customPath || joinPath(currentPath.value, task.fileName)

            // 监听后端进度事件
            const progressHandler = (ev) => {
              const d = ev?.data || ev
              if (!d || d.connId !== currentConnId.value) return
              updateTask(taskId, {
                progress: d.progress || 0,
                uploadedSize: Math.round(task.fileSize * (d.progress || 0) / 100)
              })
            }
            Events.On('ssh:upload-progress', progressHandler)

            const base64Data = arrayBufferToBase64(arrayBuffer)

            try {
              await SSHService.UploadFile(currentConnId.value, remotePath, base64Data)
            } finally {
              Events.Off('ssh:upload-progress', progressHandler)
            }

            // 上传完成
            updateTask(taskId, {
              uploadedSize: task.fileSize,
              progress: 100,
              status: UploadStatus.SUCCESS,
              endTime: Date.now()
            })

            console.log('[Upload] 上传成功:', task.fileName)
            showMessage(`已上传: ${task.fileName}`, 'success')
            logFileOperation(currentConnId.value, 'upload', remotePath)
            resolve()
          } catch (error) {
            console.error('[Upload] 上传失败:', error)
            updateTask(taskId, {
              status: UploadStatus.FAILED,
              errorMessage: error.message || '上传失败',
              endTime: Date.now()
            })
            showMessage(`上传失败: ${task.fileName}`, 'error')
            logError(currentConnId.value, `上传文件失败: ${task.fileName}`, error)
            reject(error)
          }
        }

        reader.onerror = () => {
          updateTask(taskId, {
            status: UploadStatus.FAILED,
            errorMessage: '读取文件失败',
            endTime: Date.now()
          })
          showMessage(`读取文件失败: ${task.fileName}`, 'error')
          reject(new Error('读取文件失败'))
        }
        
        reader.readAsArrayBuffer(task.file)
      })
    } catch (error) {
      console.error('[Upload] 上传异常:', error)
      updateTask(task.id, {
        status: UploadStatus.FAILED,
        errorMessage: error.message,
        endTime: Date.now()
      })
      throw error
    }
  }

  // 批量上传文件
  const uploadFiles = async (files, customPaths = null) => {
    // 确保 files 是数组
    let filesArray
    if (Array.isArray(files)) {
      filesArray = files
    } else if (files instanceof FileList) {
      filesArray = Array.from(files)
    } else {
      console.warn('[Upload] 无效的文件参数:', files)
      return
    }
    
    if (!filesArray || filesArray.length === 0) {
      console.warn('[Upload] 没有文件需要上传')
      return
    }
    
    console.log(`[Upload] 准备上传 ${filesArray.length} 个文件`)
    
    // 创建所有任务（支持自定义路径）
    const tasks = filesArray.map((file, index) => {
      const customPath = customPaths && customPaths[index] ? customPaths[index] : null
      return createUploadTask(file, customPath)
    })
    
    // 串行上传（避免并发问题）
    for (const task of tasks) {
      try {
        await executeUpload(task)
        // 上传成功后刷新文件列表
        loadFiles()
      } catch (error) {
        console.error('[Upload] 任务执行失败:', task.fileName, error)
        // 继续处理下一个文件
      }
    }
    
    console.log('[Upload] 批量上传完成')
    
    // 2秒后自动移除所有成功任务
    setTimeout(() => {
      const successCount = uploadTasks.value.filter(t => t.status === UploadStatus.SUCCESS).length
      uploadTasks.value = uploadTasks.value.filter(t => t.status !== UploadStatus.SUCCESS)
      if (successCount > 0) {
        console.log(`[Upload] 已移除 ${successCount} 个成功任务`)
      }
      // 如果没有任务了，关闭面板
      if (uploadTasks.value.length === 0) {
        showUploadPanel.value = false
      }
    }, 2000)
  }

  // 取消上传任务
  const cancelTask = (taskId) => {
    const task = uploadTasks.value.find(t => t.id === taskId)
    if (task && task.status === UploadStatus.UPLOADING) {
      if (taskId.startsWith('dir_upload_')) {
        SSHService.CancelDirectoryUpload().catch(err => {
          console.error('[Upload] 调用后端取消失败:', err)
        })
      }
      updateTask(taskId, { status: UploadStatus.CANCELLED, endTime: Date.now() })
    }
  }

  // 重试失败的任务
  const retryTask = async (taskId) => {
    const task = uploadTasks.value.find(t => t.id === taskId)
    if (task && task.status === UploadStatus.FAILED) {
      updateTask(taskId, {
        status: UploadStatus.PENDING,
        errorMessage: null,
        progress: 0,
        uploadedSize: 0
      })
      try {
        await executeUpload(task)
        loadFiles()
      } catch (error) {
        console.error('[Upload] 重试失败:', task.fileName, error)
      }
    }
  }

  // 清除已完成的任务
  const clearCompleted = () => {
    uploadTasks.value = uploadTasks.value.filter(task => 
      task.status === UploadStatus.UPLOADING || task.status === UploadStatus.PENDING
    )
    console.log('[Upload] 已清除已完成的任务')
  }

  // 清除所有任务
  const clearAll = () => {
    uploadTasks.value = []
    // ✅ 不再需要手动设置 showUploadPanel，watch 会自动处理
    console.log('[Upload] 已清除所有任务')
  }

  // 格式化文件大小
  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
  }

  // 格式化时间
  const formatDuration = (startTime, endTime) => {
    const duration = endTime - startTime
    if (duration < 1000) return `${duration}ms`
    return `${(duration / 1000).toFixed(1)}s`
  }

  return {
    uploadTasks,
    showUploadPanel,
    activeUploads,
    completedUploads,
    hasActiveUploads,
    uploadFiles,
    cancelTask,
    retryTask,
    clearCompleted,
    clearAll,
    formatFileSize,
    formatDuration
  }
}
