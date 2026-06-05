import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { commandSecurityAnalyzer } from '../utils/commandSecurityAnalyzer'

export const useCommandHistoryStore = defineStore('commandHistory', () => {
  // 每个连接的命令历史 { connId: [commands] }
  const commandHistories = ref({})
  
  // 获取指定连接的命令历史
  const getCommands = (connId) => {
    return commandHistories.value[connId] || []
  }
  
  // 获取指定连接的命令历史（别名，用于 BottomStatusBar）
  const getHistory = (connId) => {
    return getCommands(connId)
  }
  
  // 添加命令到历史
  const addCommand = (connId, command) => {
    if (!connId || connId === 'default-connection') return
    
    if (!commandHistories.value[connId]) {
      commandHistories.value[connId] = []
    }
    
    // 分析命令安全性
    const analysis = commandSecurityAnalyzer.analyzeCommand(command)
    
    const logEntry = {
      timestamp: new Date().toISOString(),
      type: 'command',
      level: analysis.riskLevel,
      content: command,
      riskLevel: analysis.riskLevel,
      category: analysis.category,
      warnings: analysis.warnings,
      suggestions: analysis.suggestions,
      isDangerous: analysis.isDangerous
    }
    
    commandHistories.value[connId].push(logEntry)
    
    // 保持最近 200 条
    if (commandHistories.value[connId].length > 200) {
      commandHistories.value[connId] = commandHistories.value[connId].slice(-200)
    }
    
    console.log('[CommandHistory] ✅ 命令已记录:', {
      command,
      riskLevel: analysis.riskLevel,
      isDangerous: analysis.isDangerous
    })
    
    return logEntry
  }
  
  // 清除指定连接的命令历史
  const clearHistory = (connId) => {
    if (commandHistories.value[connId]) {
      delete commandHistories.value[connId]
    }
  }
  
  // 分析指定连接的日志
  const analyzeCommands = (connId) => {
    const commands = getCommands(connId)
    
    const analysis = {
      totalCommands: 0,
      dangerousOps: [],
      warningOps: [],
      errorCount: 0,
      successRate: 100,
      topCommands: [],
      securityWarnings: []
    }
    
    const cmdCount = {}
    
    commands.forEach(entry => {
      if (entry.type === 'command') {
        analysis.totalCommands++
        
        // 统计命令频率
        const cmd = entry.content.trim()
        cmdCount[cmd] = (cmdCount[cmd] || 0) + 1
        
        // 识别危险操作
        if (entry.isDangerous || entry.riskLevel === 'critical' || entry.riskLevel === 'high') {
          analysis.dangerousOps.push(
            `${entry.content} [${formatTime(entry.timestamp)}]`
          )
          
          // 添加安全警告
          entry.warnings.forEach(warning => {
            analysis.securityWarnings.push(warning)
          })
        } else if (entry.riskLevel === 'medium' || entry.riskLevel === 'low') {
          analysis.warningOps.push(
            `${entry.content} [${formatTime(entry.timestamp)}]`
          )
        }
      } else if (entry.type === 'error') {
        analysis.errorCount++
      }
    })
    
    // 计算成功率
    const totalOps = analysis.totalCommands + analysis.errorCount
    if (totalOps > 0) {
      analysis.successRate = (analysis.totalCommands / totalOps) * 100
    }
    
    // 找出最常用的命令（Top 5）
    const freqList = Object.entries(cmdCount)
      .map(([cmd, count]) => ({ cmd, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 5)
    
    analysis.topCommands = freqList.map(
      item => `${item.cmd} (${item.count}次)`
    )
    
    return analysis
  }
  
  // 导出日志为 JSON 格式
  const exportLogs = (connId) => {
    const commands = getCommands(connId)
    
    const exportData = {
      connectionId: connId,
      exportTime: new Date().toISOString(),
      totalRecords: commands.length,
      logs: commands.map(entry => ({
        timestamp: entry.timestamp,
        type: entry.type,
        riskLevel: entry.riskLevel,
        content: entry.content,
        category: entry.category,
        warnings: entry.warnings || [],
        suggestions: entry.suggestions || [],
        isDangerous: entry.isDangerous || false
      }))
    }
    
    return JSON.stringify(exportData, null, 2)
  }
  
  // 格式化时间
  const formatTime = (timestamp) => {
    if (!timestamp) return ''
    const date = new Date(timestamp)
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false
    })
  }
  
  return {
    commandHistories,
    getCommands,
    getHistory,
    addCommand,
    clearHistory,
    analyzeCommands,
    exportLogs
  }
})
