/**
 * 命令历史管理组合式函数
 * 支持命令记录、搜索、收藏、分类
 */
import { ref, computed } from 'vue'
import { commandSecurityAnalyzer } from '../../../../../utils/commandSecurityAnalyzer'

export function useCommandHistory(connId, options = {}) {
  const { maxHistory = 500, persistKey = 'cmd_history_global' } = options

  // 命令历史
  const history = ref([])
  const favorites = ref([])

  // 当前输入缓冲
  let currentInput = ''
  let historyIndex = -1

  // 计算属性
  const recentCommands = computed(() => {
    return history.value.slice(-20).reverse()
  })

  const frequentCommands = computed(() => {
    const freq = {}
    history.value.forEach(entry => {
      const cmd = entry.command.trim()
      if (cmd) {
        freq[cmd] = (freq[cmd] || 0) + 1
      }
    })

    return Object.entries(freq)
      .map(([cmd, count]) => ({ command: cmd, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10)
  })

  // 加载历史
  function loadHistory() {
    try {
      const saved = localStorage.getItem(persistKey)
      if (saved) {
        const data = JSON.parse(saved)
        history.value = data.history || []
        favorites.value = data.favorites || []
      }
    } catch (e) {
      console.warn('[useCommandHistory] 加载历史失败:', e)
    }
  }

  // 保存历史
  function saveHistory() {
    try {
      localStorage.setItem(persistKey, JSON.stringify({
        history: history.value.slice(-maxHistory),
        favorites: favorites.value
      }))
    } catch (e) {
      console.warn('[useCommandHistory] 保存历史失败:', e)
    }
  }

  // 添加命令
  function addCommand(command) {
    if (!command || !command.trim()) return

    const trimmed = command.trim()

    // 分析安全性
    const analysis = commandSecurityAnalyzer.analyzeCommand(trimmed)

    const entry = {
      id: Date.now().toString(36) + Math.random().toString(36).slice(2, 6),
      command: trimmed,
      timestamp: Date.now(),
      riskLevel: analysis.riskLevel,
      category: analysis.category,
      isDangerous: analysis.isDangerous
    }

    history.value.push(entry)

    // 限制历史长度
    if (history.value.length > maxHistory) {
      history.value = history.value.slice(-maxHistory)
    }

    saveHistory()
    return entry
  }

  // 搜索命令
  function searchCommands(keyword) {
    if (!keyword) return history.value.slice(-50)

    const lower = keyword.toLowerCase()
    return history.value.filter(entry =>
      entry.command.toLowerCase().includes(lower)
    ).slice(-50)
  }

  // 获取上一条命令（用于上箭头）
  function getPreviousCommand() {
    if (history.value.length === 0) return null

    if (historyIndex === -1) {
      historyIndex = history.value.length - 1
    } else if (historyIndex > 0) {
      historyIndex--
    }

    return history.value[historyIndex]?.command || null
  }

  // 获取下一条命令（用于下箭头）
  function getNextCommand() {
    if (historyIndex === -1) return null

    if (historyIndex < history.value.length - 1) {
      historyIndex++
      return history.value[historyIndex]?.command || null
    }

    historyIndex = -1
    return ''
  }

  // 重置历史导航
  function resetNavigation() {
    historyIndex = -1
  }

  // 添加到收藏
  function addFavorite(command, label = '') {
    if (!command || !command.trim()) return

    const trimmed = command.trim()

    // 检查是否已收藏
    if (favorites.value.some(f => f.command === trimmed)) return

    favorites.value.push({
      id: Date.now().toString(36),
      command: trimmed,
      label: label || trimmed,
      timestamp: Date.now()
    })

    saveHistory()
  }

  // 移除收藏
  function removeFavorite(id) {
    favorites.value = favorites.value.filter(f => f.id !== id)
    saveHistory()
  }

  // 检查是否已收藏
  function isFavorite(command) {
    return favorites.value.some(f => f.command === command.trim())
  }

  // 清除历史
  function clearHistory() {
    history.value = []
    saveHistory()
  }

  // 清除收藏
  function clearFavorites() {
    favorites.value = []
    saveHistory()
  }

  // 统计信息
  function getStats() {
    const total = history.value.length
    const dangerous = history.value.filter(e => e.isDangerous).length
    const categories = {}

    history.value.forEach(entry => {
      if (entry.category) {
        categories[entry.category] = (categories[entry.category] || 0) + 1
      }
    })

    return {
      total,
      dangerous,
      categories,
      favoriteCount: favorites.value.length
    }
  }

  // 导出历史
  function exportHistory() {
    return JSON.stringify({
      connId,
      exportTime: new Date().toISOString(),
      history: history.value,
      favorites: favorites.value,
      stats: getStats()
    }, null, 2)
  }

  // 初始化
  loadHistory()

  return {
    // 状态
    history,
    favorites,
    recentCommands,
    frequentCommands,

    // 命令管理
    addCommand,
    searchCommands,

    // 历史导航
    getPreviousCommand,
    getNextCommand,
    resetNavigation,

    // 收藏管理
    addFavorite,
    removeFavorite,
    isFavorite,

    // 其他
    clearHistory,
    clearFavorites,
    getStats,
    exportHistory
  }
}
