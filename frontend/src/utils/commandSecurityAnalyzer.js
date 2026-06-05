/**
 * SSH 命令安全分析器
 * 使用多层检测策略识别危险命令
 */

export class CommandSecurityAnalyzer {
  constructor() {
    // 危险命令模式库
    this.dangerousPatterns = [
      // Fork Bomb 变体
      {
        name: 'Fork Bomb',
        patterns: [
          /:\(\)\s*\{\s*:\|\s*:\s*&\s*\}\s*;/,  // :(){ :|:& };
          /:\(\)\s*\{\s*:\s*\|\s*:\s*&\s*\}\s*;/,  // 带空格变体
          /bomb\s*\(/i,  // bomb(
          /fork\s*bomb/i,  // fork bomb
        ],
        severity: 'critical',
        description: '进程炸弹 - 会耗尽系统资源导致崩溃'
      },
      
      // 删除根目录
      {
        name: '删除根目录',
        patterns: [
          /rm\s+(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*)\s+(\/\s*$|\/\*)/,  // rm -rf /
          /rm\s+(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*)\s+\/$/,  // rm -rf /
          /rm\s+-rf\s+\//,  // rm -rf /
          /rm\s+--no-preserve-root\s+\//,  // rm --no-preserve-root /
        ],
        severity: 'critical',
        description: '删除整个文件系统'
      },
      
      // 磁盘操作
      {
        name: '危险磁盘操作',
        patterns: [
          /dd\s+.*of=\/dev\/(sd|hd|xvd|vd)[a-z]/i,  // dd 写入磁盘
          /mkfs\.(ext[2-4]|xfs|btrfs|vfat)\s+\/dev\//,  // 格式化磁盘
          /fdisk\s+\/dev\//,  // 分区操作
          /parted\s+\/dev\//,  // 分区工具
          />\s*\/dev\/(sd|hd)[a-z]/,  // 清空磁盘设备
        ],
        severity: 'critical',
        description: '可能破坏磁盘数据或分区表'
      },
      
      // 系统文件操作
      {
        name: '系统文件破坏',
        patterns: [
          />\s*\/etc\/(passwd|shadow|fstab|ssh|nginx|apache|mysql)/i,  // 清空配置文件
          /chmod\s+777\s+(\/etc|\/usr|\/var|\/bin|\/sbin|\/)$/,  // 过度开放权限
          /chown\s+-R\s+\S+:\S+\s+\/$/,  // 递归修改根目录所有权
          /mv\s+\/etc\s/,  // 移动系统目录
        ],
        severity: 'high',
        description: '可能破坏系统配置或权限'
      },
      
      // 网络攻击
      {
        name: '网络攻击命令',
        patterns: [
          /iptables\s+-F/i,  // 清空防火墙规则
          /ufw\s+disable/i,  // 禁用防火墙
          /systemctl\s+stop\s+(firewalld|ufw|iptables)/i,  // 停止防火墙
          /nmap\s+-sS\s+-p-\s+/i,  // 全端口扫描
        ],
        severity: 'high',
        description: '可能削弱系统安全防护'
      },
      
      // 进程管理
      {
        name: '危险进程操作',
        patterns: [
          /kill\s+-9\s+-?\d+/,  // 强制杀死所有进程
          /killall\s+-9\s+\w+/,  // 强制杀死所有同名进程
          /pkill\s+-9\s+-f\s+.*$/,  // 强制杀死匹配进程
          /:> \/proc\/sysrq-trigger/,  // 触发内核魔术键
        ],
        severity: 'high',
        description: '可能导致系统不稳定或服务中断'
      },
      
      // 数据泄露
      {
        name: '数据泄露风险',
        patterns: [
          /cat\s+\/etc\/(passwd|shadow)/,  // 读取敏感文件
          /mysqldump\s+.*--all-databases/i,  // 导出所有数据库
          /tar\s+.*\/etc\s+.*\.tar/i,  // 打包系统目录
          /scp\s+-r\s+\/\s*/,  // 远程复制根目录
        ],
        severity: 'medium',
        description: '可能泄露敏感数据'
      },
      
      // 权限提升
      {
        name: '可疑权限操作',
        patterns: [
          /sudo\s+su\s*$/i,  // 切换到 root
          /sudo\s+bash\s*$/i,  // 以 root 启动 bash
          /sudo\s+\/bin\/sh\s*$/i,  // 以 root 启动 sh
          /passwd\s+\w+/,  // 修改用户密码
        ],
        severity: 'medium',
        description: '可能提升权限或修改认证'
      }
    ]
  }

  /**
   * 分析命令安全性
   * @param {string} command - 要分析的命令
   * @returns {Object} 分析结果
   */
  analyzeCommand(command) {
    if (!command || typeof command !== 'string') {
      return {
        isDangerous: false,
        riskLevel: 'safe',
        category: 'unknown',
        warnings: [],
        suggestions: []
      }
    }

    const trimmedCmd = command.trim()
    
    // 空命令
    if (!trimmedCmd) {
      return {
        isDangerous: false,
        riskLevel: 'safe',
        category: 'empty',
        warnings: [],
        suggestions: []
      }
    }

    // 逐层检测
    const results = []
    
    for (const pattern of this.dangerousPatterns) {
      for (const regex of pattern.patterns) {
        if (regex.test(trimmedCmd)) {
          results.push({
            name: pattern.name,
            severity: pattern.severity,
            description: pattern.description,
            matchedPattern: regex.source
          })
          break  // 每个类别只报告一次
        }
      }
    }

    // 如果没有匹配到任何危险模式
    if (results.length === 0) {
      return {
        isDangerous: false,
        riskLevel: 'safe',
        category: 'normal',
        warnings: [],
        suggestions: ['命令看起来是安全的']
      }
    }

    // 确定最高风险级别
    const severityOrder = { critical: 4, high: 3, medium: 2, low: 1 }
    const highestSeverity = results.reduce((max, curr) => 
      severityOrder[curr.severity] > severityOrder[max.severity] ? curr : max
    )

    // 生成警告和建议
    const warnings = results.map(r => `⚠️ ${r.name}: ${r.description}`)
    const suggestions = this.generateSuggestions(results, trimmedCmd)

    return {
      isDangerous: true,
      riskLevel: highestSeverity.severity,
      category: highestSeverity.name,
      warnings,
      suggestions,
      details: results
    }
  }

  /**
   * 生成安全建议
   */
  generateSuggestions(matches, command) {
    const suggestions = []
    
    for (const match of matches) {
      switch (match.name) {
        case 'Fork Bomb':
          suggestions.push('❌ 绝对不要执行此命令！会导致系统立即崩溃')
          suggestions.push('💡 如需测试系统负载，使用 stress 或 stress-ng 工具')
          break
          
        case '删除根目录':
          suggestions.push('❌ 这会删除整个文件系统，不可恢复！')
          suggestions.push('💡 使用 rm 时始终确认路径，建议使用 rm -i 交互模式')
          suggestions.push('💡 考虑使用 trash-cli 代替 rm，可以恢复删除的文件')
          break
          
        case '危险磁盘操作':
          suggestions.push('❌ 这会破坏磁盘数据，请先备份重要数据')
          suggestions.push('💡 使用 lsblk 或 fdisk -l 确认设备名称')
          suggestions.push('💡 在执行前使用 --dry-run 参数测试（如果支持）')
          break
          
        case '系统文件破坏':
          suggestions.push('⚠️ 修改系统文件可能导致系统不稳定')
          suggestions.push('💡 先备份原文件：cp /etc/xxx /etc/xxx.bak')
          suggestions.push('💡 使用版本控制工具管理配置文件')
          break
          
        case '网络攻击命令':
          suggestions.push('⚠️ 这可能削弱系统网络安全')
          suggestions.push('💡 确保有替代的安全措施后再执行')
          break
          
        default:
          suggestions.push('⚠️ 请谨慎执行此命令，确认了解其影响')
      }
    }
    
    return suggestions
  }

  /**
   * 检测命令是否包含管道、重定向等复杂结构
   */
  hasComplexStructure(command) {
    const complexPatterns = [
      /\|/,           // 管道
      />>?/,          // 重定向
      /&&/,           // 逻辑与
      /\|\|/,         // 逻辑或
      /;/,            // 命令分隔
      /`[^`]+`/,      // 命令替换
      /\$\([^)]+\)/,  // 命令替换
      /\{[^}]+\}/,    // 代码块
    ]
    
    return complexPatterns.some(pattern => pattern.test(command))
  }

  /**
   * 获取风险等级颜色
   */
  getRiskColor(riskLevel) {
    const colors = {
      safe: '#48bb78',      // 绿色
      low: '#ecc94b',       // 黄色
      medium: '#ed8936',    // 橙色
      high: '#f56565',      // 红色
      critical: '#e53e3e'   // 深红色
    }
    return colors[riskLevel] || '#718096'
  }

  /**
   * 获取风险等级图标
   */
  getRiskIcon(riskLevel) {
    const icons = {
      safe: '✅',
      low: '⚠️',
      medium: '🟠',
      high: '🔴',
      critical: '💥'
    }
    return icons[riskLevel] || '❓'
  }
}

// 导出单例
export const commandSecurityAnalyzer = new CommandSecurityAnalyzer()
