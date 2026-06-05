import { ref, computed } from 'vue'
import { SSHService } from '@bindings/changeme/ssh/index.js'
import { showMessage } from '@/utils/message'
import { normalizePath, joinPath } from '../utils/pathHelper'
import { logFileOperation, logError, logSystemEvent } from '@/utils/logger'

/**
 * 文件操作 Composable
 * @param {Ref} currentConnId - 当前连接ID
 * @param {Ref} currentPath - 当前路径
 * @param {Function} loadFiles - 重新加载文件列表的函数
 */
export function useFileOperations(currentConnId, currentPath, loadFiles) {
  const loading = ref(false)
  
  // ✅ 使用 ref 追踪剪切状态，确保响应式更新
  const hasCutFile = ref(sessionStorage.getItem('cutFile') !== null)
  
  // 上传文件
  const uploadFile = async (file) => {
    try {
      console.log('[FileManager] 开始上传文件:', file.name)
      const reader = new FileReader()
      return new Promise((resolve, reject) => {
        reader.onload = async (e) => {
          try {
            const arrayBuffer = e.target.result
            const data = new Uint8Array(arrayBuffer)
            const remotePath = joinPath(currentPath.value, file.name)
            
            console.log('[FileManager] 上传参数:', {
              connId: currentConnId.value,
              remotePath: remotePath,
              dataSize: data.length
            })
            
            // 将 Uint8Array 转换为 base64 字符串
            const base64Data = arrayBufferToBase64(arrayBuffer)
            
            await SSHService.UploadFile(currentConnId.value, remotePath, base64Data)
            showMessage(`已上传: ${file.name}`, 'success')
            logFileOperation(currentConnId.value, 'upload', remotePath)
            loadFiles()
            resolve()
          } catch (error) {
            console.error('[FileManager] 上传失败:', error)
            showMessage(`上传失败: ${file.name}`, 'error')
            logError(currentConnId.value, `上传文件失败: ${file.name}`, error)
            reject(error)
          }
        }
        reader.onerror = reject
        reader.readAsArrayBuffer(file)
      })
    } catch (error) {
      console.error('[FileManager] 上传失败:', error)
      throw error
    }
  }

  // 将 ArrayBuffer 转换为 base64 字符串（高效方式）
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

  // 批量上传文件
  const uploadFiles = async (files) => {
    if (!files || files.length === 0) return
    
    for (const file of files) {
      await uploadFile(file)
    }
  }

  // 上传目录（递归）
  const uploadDirectory = async (directoryHandle, parentPath = '') => {
    try {
      console.log('[FileManager] 开始上传目录:', directoryHandle.name)
      
      // 创建远程目录
      const remoteDirPath = parentPath ? joinPath(currentPath.value, parentPath, directoryHandle.name) : joinPath(currentPath.value, directoryHandle.name)
      await SSHService.CreateDirectory(currentConnId.value, remoteDirPath)
      showMessage(`已创建目录: ${directoryHandle.name}`, 'success')
      
      // 遍历目录内容
      for await (const entry of directoryHandle.values()) {
        if (entry.kind === 'file') {
          // 上传文件
          const file = await entry.getFile()
          const reader = new FileReader()
          
          await new Promise((resolve, reject) => {
            reader.onload = async (e) => {
              try {
                const arrayBuffer = e.target.result
                const base64Data = arrayBufferToBase64(arrayBuffer)
                const remoteFilePath = joinPath(remoteDirPath, entry.name)
                
                await SSHService.UploadFile(currentConnId.value, remoteFilePath, base64Data)
                console.log('[FileManager] 已上传:', entry.name)
                resolve()
              } catch (error) {
                console.error('[FileManager] 上传文件失败:', entry.name, error)
                reject(error)
              }
            }
            reader.onerror = reject
            reader.readAsArrayBuffer(file)
          })
        } else if (entry.kind === 'directory') {
          // 递归上传子目录
          await uploadDirectory(entry, joinPath(parentPath, directoryHandle.name))
        }
      }
      
      showMessage(`目录上传完成: ${directoryHandle.name}`, 'success')
      loadFiles()
    } catch (error) {
      console.error('[FileManager] 上传目录失败:', error)
      showMessage('上传目录失败', 'error')
      throw error
    }
  }

  // 下载文件
  const downloadFile = async (file) => {
    if (file.isDir) {
      // 目录需要先压缩再下载
      await downloadDirectory(file)
      return
    }
    
    try {
      const normalizedPath = normalizePath(file.path)
      console.log('[FileManager] 下载文件:', { connId: currentConnId.value, path: normalizedPath })
      const base64Data = await SSHService.DownloadFile(currentConnId.value, normalizedPath)
      
      // 将 base64 转换为 Blob
      const binaryString = atob(base64Data)
      const bytes = new Uint8Array(binaryString.length)
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i)
      }
      
      const blob = new Blob([bytes])
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = file.name
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      
      showMessage(`已下载: ${file.name}`, 'success')
      logFileOperation(currentConnId.value, 'download', normalizedPath)
    } catch (error) {
      console.error('[FileManager] 下载失败:', error)
      showMessage(`下载失败: ${error.message || '未知错误'}`, 'error')
      logError(currentConnId.value, `下载文件失败: ${file.name}`, error)
    }
  }

  // 下载目录（先压缩）
  const downloadDirectory = async (file) => {
    showMessage('正在压缩目录...', 'info')
    try {
      const tarName = `${file.name}.tar.gz`
      const tarPath = joinPath(currentPath.value, tarName)
      
      // 使用 tar 命令压缩目录（tar 比 zip 更通用）
      const cmd = `cd "${currentPath.value}" && tar -czf "${tarName}" "${file.name}"`
      await SSHService.ExecuteCommand(currentConnId.value, cmd)
      
      // 下载压缩文件
      const base64Data = await SSHService.DownloadFile(currentConnId.value, tarPath)
      
      // 将 base64 转换为 Blob
      const binaryString = atob(base64Data)
      const bytes = new Uint8Array(binaryString.length)
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i)
      }
      
      const blob = new Blob([bytes])
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = tarName
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      
      // 删除临时压缩文件
      await SSHService.DeleteFile(currentConnId.value, tarPath)
      
      showMessage(`已下载: ${tarName}`, 'success')
      logFileOperation(currentConnId.value, 'download', tarPath)
    } catch (error) {
      console.error('[FileManager] 下载目录失败:', error)
      showMessage('下载目录失败: ' + error.message, 'error')
      logError(currentConnId.value, `下载目录失败: ${file.name}`, error)
    }
  }

  // 删除文件
  const deleteFile = async (file) => {
    try {
      const normalizedPath = normalizePath(file.path)
      console.log('[FileManager] 删除文件:', { connId: currentConnId.value, path: normalizedPath })
      await SSHService.DeleteFile(currentConnId.value, normalizedPath)
      showMessage(`已删除: ${file.name}`, 'success')
      logFileOperation(currentConnId.value, 'delete', normalizedPath)
      loadFiles()
    } catch (error) {
      console.error('[FileManager] 删除失败:', error)
      showMessage('删除失败', 'error')
      logError(currentConnId.value, `删除文件失败: ${file.name}`, error)
      throw error
    }
  }

  // 重命名文件
  const renameFile = async (file, newName) => {
    if (!newName || newName === file.name) return
    
    const newPath = joinPath(currentPath.value, newName)
    
    try {
      await SSHService.RenameFile(currentConnId.value, file.path, newPath)
      showMessage('重命名成功', 'success')
      logFileOperation(currentConnId.value, 'rename', newPath)
      loadFiles()
    } catch (error) {
      console.error('[FileManager] 重命名失败:', error)
      showMessage('重命名失败', 'error')
      logError(currentConnId.value, `重命名文件失败: ${file.name}`, error)
      throw error
    }
  }

  // 创建副本
  const duplicateFile = async (file) => {
    try {
      let newName
      if (file.isDir) {
        newName = `${file.name} - 副本`
      } else {
        const lastDot = file.name.lastIndexOf('.')
        if (lastDot > 0) {
          const name = file.name.substring(0, lastDot)
          const ext = file.name.substring(lastDot)
          newName = `${name} - 副本${ext}`
        } else {
          newName = `${file.name} - 副本`
        }
      }
      
      const newPath = joinPath(currentPath.value, newName)
      
      // 使用 cp 命令复制
      const cmd = `cp -r "${file.path}" "${newPath}"`
      await SSHService.ExecuteCommand(currentConnId.value, cmd)
      showMessage('创建副本成功', 'success')
      logFileOperation(currentConnId.value, 'create', newPath)
      loadFiles()
    } catch (error) {
      console.error('[FileManager] 创建副本失败:', error)
      showMessage('创建副本失败', 'error')
      logError(currentConnId.value, `创建副本失败: ${file.name}`, error)
      throw error
    }
  }

  // 压缩为 tar.gz
  const compressFile = async (file) => {
    const tarName = `${file.name}.tar.gz`
    
    try {
      showMessage('正在压缩...', 'info')
      // 使用 tar 命令压缩（tar 比 zip 更通用）
      const cmd = `cd "${currentPath.value}" && tar -czf "${tarName}" "${file.name}"`
      await SSHService.ExecuteCommand(currentConnId.value, cmd)
      showMessage(`已压缩: ${tarName}`, 'success')
      logFileOperation(currentConnId.value, 'compress', tarPath)
      loadFiles()
    } catch (error) {
      console.error('[FileManager] 压缩失败:', error)
      showMessage('压缩失败: ' + error.message, 'error')
      logError(currentConnId.value, `压缩文件失败: ${file.name}`, error)
      throw error
    }
  }

  // 剪切文件
  const cutFile = (file) => {
    // 存储到 sessionStorage
    sessionStorage.setItem('cutFile', JSON.stringify({
      path: file.path,
      name: file.name,
      isDir: file.isDir,
      fromPath: currentPath.value
    }))
      
    // ✅ 更新响应式状态
    hasCutFile.value = true
      
    showMessage(`已剪切: ${file.name}，导航到目标目录后点击“粘贴”`, 'info')
  }
  
  // 粘贴文件
  const pasteFile = async () => {
    const cutData = sessionStorage.getItem('cutFile')
    if (!cutData) {
      showMessage('没有要粘贴的文件', 'warning')
      return
    }
      
    try {
      const { path, name } = JSON.parse(cutData)
      const destPath = joinPath(currentPath.value, name)
        
      // 使用 mv 命令移动文件（剪切）
      const cmd = `mv "${path}" "${destPath}"`
      await SSHService.ExecuteCommand(currentConnId.value, cmd)
        
      // 清除剪贴板
      sessionStorage.removeItem('cutFile')
        
      // ✅ 更新响应式状态
      hasCutFile.value = false
        
      showMessage(`已粘贴: ${name}`, 'success')
      logFileOperation(currentConnId.value, 'rename', destPath)
      loadFiles()
    } catch (error) {
      console.error('[FileManager] 粘贴失败:', error)
      showMessage('粘贴失败', 'error')
      throw error
    }
  }

  return {
    loading,
    uploadFiles,
    uploadDirectory,
    downloadFile,
    deleteFile,
    renameFile,
    duplicateFile,
    compressFile,
    cutFile,
    pasteFile,
    hasCutFile
  }
}

