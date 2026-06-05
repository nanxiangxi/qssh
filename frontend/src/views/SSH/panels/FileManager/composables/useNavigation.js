import { computed } from 'vue'
import { getParentPath } from '../utils/pathHelper'

/**
 * 导航和排序 Composable
 * @param {Ref} currentPath - 当前路径
 * @param {Ref} files - 文件列表
 * @param {Ref} sortKey - 排序键
 * @param {Ref} sortOrder - 排序顺序
 * @param {Function} loadFiles - 重新加载文件列表的函数
 * @param {Ref} selectedFiles - 选中的文件集合
 */
export function useNavigation(currentPath, files, sortKey, sortOrder, loadFiles, selectedFiles) {
  /**
   * 向上一级
   */
  const goUp = () => {
    if (currentPath.value === '/') return
    currentPath.value = getParentPath(currentPath.value)
    // 清空选中状态
    if (selectedFiles) {
      selectedFiles.value.clear()
    }
    loadFiles()
  }

  /**
   * 导航到路径
   */
  const navigateToPath = () => {
    // 清空选中状态
    if (selectedFiles) {
      selectedFiles.value.clear()
    }
    loadFiles()
  }

  /**
   * 打开文件/目录
   * @param {Object} file - 目标文件
   */
  const openFile = (file) => {
    if (file.isDir) {
      currentPath.value = file.path
      // 清空选中状态
      if (selectedFiles) {
        selectedFiles.value.clear()
      }
      loadFiles()
    } else {
      // 双击非文本、非图片文件时显示提示（由父组件处理）
      // 这里只返回目录导航逻辑
    }
  }

  /**
   * 排序
   * @param {string} key - 排序键
   */
  const sortBy = (key) => {
    if (sortKey.value === key) {
      sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    } else {
      sortKey.value = key
      sortOrder.value = 'asc'
    }
  }

  /**
   * 排序后的文件列表
   */
  const sortedFiles = computed(() => {
    const sorted = [...files.value]
    sorted.sort((a, b) => {
      let aVal = a[sortKey.value]
      let bVal = b[sortKey.value]
      
      if (sortKey.value === 'name') {
        aVal = aVal.toLowerCase()
        bVal = bVal.toLowerCase()
      }
      
      if (sortOrder.value === 'asc') {
        return aVal > bVal ? 1 : -1
      } else {
        return aVal < bVal ? 1 : -1
      }
    })
    
    // 目录始终在前
    return sorted.sort((a, b) => {
      if (a.isDir && !b.isDir) return -1
      if (!a.isDir && b.isDir) return 1
      return 0
    })
  })

  return {
    goUp,
    navigateToPath,
    openFile,
    sortBy,
    sortedFiles
  }
}
