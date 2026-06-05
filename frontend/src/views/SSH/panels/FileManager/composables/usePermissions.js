import { ref, watch } from 'vue'
import { SSHService } from '@bindings/changeme/ssh/index.js'
import { showMessage } from '@/utils/message'
import { normalizePath, getNumericMode } from '../utils/pathHelper'

/**
 * 权限管理 Composable
 * @param {Function} loadFiles - 重新加载文件列表的函数
 */
export function usePermissions(loadFiles) {
  const showChmodDialog = ref(false)
  const chmodValue = ref('755')
  const chmodTarget = ref(null)
  const applyToSubdirs = ref(false)

  console.log('[usePermissions] 初始化, showChmodDialog:', showChmodDialog.value)

  // 权限复选框状态
  const chmodPermissions = ref({
    owner: { r: false, w: false, x: false },
    group: { r: false, w: false, x: false },
    other: { r: false, w: false, x: false }
  })

  /**
   * 从数字权限转换为复选框状态
   * @param {string} numericMode - 数字权限（如 "755"）
   */
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

  /**
   * 从复选框状态计算数字权限
   */
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

  /**
   * 设置预设权限
   * @param {string} value - 预设权限值（如 "755"）
   */
  const setPreset = (value) => {
    chmodValue.value = value
    parsePermissions(value)
  }

  /**
   * 打开权限修改对话框
   * @param {Object} file - 目标文件
   */
  const openChmodDialog = (file) => {
    console.log('[usePermissions] openChmodDialog 被调用, file:', file?.name)
    chmodTarget.value = file
    const numericMode = getNumericMode(file.mode)
    chmodValue.value = numericMode
    parsePermissions(numericMode)
    applyToSubdirs.value = false
    showChmodDialog.value = true
    console.log('[usePermissions] showChmodDialog 设置为:', showChmodDialog.value)
  }

  /**
   * 确认修改权限
   * @param {string} connId - 连接ID
   */
  const confirmChmod = async (connId) => {
    if (!chmodTarget.value) return
    
    try {
      const normalizedPath = normalizePath(chmodTarget.value.path)
      let command = `chmod ${chmodValue.value} "${normalizedPath}"`
      
      // 如果勾选了应用到子目录
      if (applyToSubdirs.value && chmodTarget.value.isDir) {
        command = `chmod -R ${chmodValue.value} "${normalizedPath}"`
      }
      
      const result = await SSHService.ExecuteCommand(connId, command)
      
      if (result && result.Success) {
        showMessage('权限修改成功', 'success')
        loadFiles()
        showChmodDialog.value = false
      } else {
        throw new Error(result?.Stderr || '未知错误')
      }
    } catch (error) {
      console.error('[FileManager] 权限修改失败:', error)
      showMessage(`权限修改失败: ${error.message}`, 'error')
      throw error
    }
  }

  /**
   * 关闭权限对话框
   */
  const closeChmodDialog = () => {
    showChmodDialog.value = false
  }

  // 监听权限复选框变化，自动计算数字权限
  watch(chmodPermissions, () => {
    calculateNumericMode()
  }, { deep: true })

  return {
    showChmodDialog,
    chmodValue,
    chmodTarget,
    applyToSubdirs,
    chmodPermissions,
    openChmodDialog,
    confirmChmod,
    closeChmodDialog,
    setPreset
  }
}
