// 前端命令安全分析器
// 基于 security/analyzer.go 的规则，在前端实现相同的分析逻辑

export class CommandSecurityAnalyzer {
  constructor() {
    this.dangerousPatterns = []
    this.warningPatterns = []
    this.safeCommands = new Set()
    this.initializeRules()
    this.initializeSafeCommands()
  }

  // 初始化危险规则
  initializeRules() {
    // ========== 极度危险规则 ==========
    
    // Fork Bomb:
    this.addDangerousRule(
      /^:\(\)\s*\{\s*:\|\s*:\s*&\s*\}\s*;\s*:/,
      'process',
      'Fork Bomb - 进程炸弹',
      [
        '⚠️ 检测到 Fork Bomb！这会导致系统资源耗尽并崩溃',
        '🔴 立即终止此命令，否则系统将无法响应'
      ],
      [
        '使用 ulimit 限制用户进程数',
        '监控系统负载，设置自动保护机制'
      ]
    )

    // rm -rf / 或 rm -rf /*
    this.addDangerousRule(
      /^rm\s+(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*)\s+(\/\s*|\/\*)/,
      'file',
      '删除根目录',
      [
        '🔴 极度危险：尝试删除根目录所有文件！',
        '这将导致系统完全崩溃且无法恢复'
      ],
      [
        '使用 rm -i 交互模式删除文件',
        '重要数据定期备份',
        '考虑使用 trash-cli 替代 rm'
      ]
    )

    // dd 写入关键设备
    this.addDangerousRule(
      /^dd\s+.*of=\/dev\/(sd|hd|xvd|vd)[a-z]/,
      'disk',
      '直接写入磁盘设备',
      [
        '🔴 高危：直接写入磁盘设备，可能导致数据丢失',
        '此操作会覆盖整个磁盘分区'
      ],
      [
        '确认设备名称是否正确',
        '操作前备份重要数据',
        '使用 dd if=/dev/zero of=... bs=1M count=1 先测试'
      ]
    )

    // 格式化磁盘
    this.addDangerousRule(
      /^(mkfs\.|fdisk\s|parted\s|format\s)/,
      'disk',
      '磁盘格式化/分区操作',
      [
        '🔴 高危：磁盘格式化操作，将清除所有数据',
        '分区表修改可能导致系统无法启动'
      ],
      [
        '确认目标设备是否正确',
        '备份分区表和重要数据',
        '使用 lsblk 确认设备信息'
      ]
    )

    // chmod 777 系统目录
    this.addDangerousRule(
      /^chmod\s+777\s+(\/etc|\/usr|\/var|\/bin|\/sbin|\/)/,
      'security',
      '系统目录权限过度开放',
      [
        '🔴 高危：系统关键目录权限设置为 777',
        '任何用户都可以读写执行，存在严重安全风险'
      ],
      [
        '使用最小权限原则',
        '根据实际需求设置权限（如 755, 644）',
        '定期审计文件权限'
      ]
    )

    // sudo + 危险操作
    this.addDangerousRule(
      /^sudo\s+(rm\s+-rf|dd\s+|mkfs\.|fdisk\s)/,
      'system',
      '以 root 权限执行危险操作',
      [
        '🔴 以 root 权限执行危险操作，极其危险',
        '一旦误操作将无法撤销'
      ],
      [
        '仔细检查命令参数',
        '考虑使用普通权限执行',
        '先在测试环境验证'
      ]
    )

    // 清空系统配置文件
    this.addDangerousRule(
      /^>\s*\/etc\/(passwd|shadow|fstab|ssh|nginx|apache)/,
      'system',
      '清空系统配置文件',
      [
        '🔴 尝试清空系统配置文件，可能导致系统无法启动',
        '服务配置丢失将导致服务中断'
      ],
      [
        '使用 cp 备份后再编辑',
        '使用文本编辑器而非重定向',
        '保持配置文件的版本控制'
      ]
    )

    // ========== 警告规则 ==========

    // rm 删除文件
    this.addWarningRule(
      /^rm\s+(-[a-zA-Z]*r[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*)?\s+\S/,
      'file',
      '删除文件',
      ['⚠️ 删除文件操作，请确认文件路径正确'],
      [
        '使用 rm -i 进行交互式删除',
        '删除前先用 ls 确认文件',
        '考虑移动到回收站而非直接删除'
      ]
    )

    // 移动/复制系统文件
    this.addWarningRule(
      /^(mv|cp)\s+. *\/(etc|usr|var|boot)\//,
      'system',
      '操作系统文件',
      ['⚠️ 移动或复制系统文件，可能影响系统稳定性'],
      [
        '操作前备份原文件',
        '确认目标路径正确',
        '避免在业务高峰期操作'
      ]
    )

    // 防火墙配置
    this.addWarningRule(
      /^(iptables|ufw|firewall-cmd)\s+/,
      'network',
      '防火墙配置修改',
      ['⚠️ 修改防火墙规则，可能影响网络连接'],
      [
        '先查看当前规则：iptables -L',
        '保留 SSH 端口访问权限',
        '测试后再应用永久规则'
      ]
    )

    // 停止服务
    this.addWarningRule(
      /^(systemctl|service)\s+(stop|disable)\s+/,
      'service',
      '停止或禁用服务',
      ['⚠️ 停止或禁用系统服务，可能影响业务运行'],
      [
        '确认服务是否可以安全停止',
        '检查是否有依赖此服务的其他应用',
        '记录操作时间以便快速恢复'
      ]
    )

    // 强制终止进程
    this.addWarningRule(
      /^(kill\s+-9|pkill\s|killall\s)/,
      'process',
      '强制终止进程',
      ['⚠️ 强制终止进程，可能导致数据丢失'],
      [
        '先尝试 SIGTERM (kill -15)',
        '确认进程 ID 是否正确',
        '检查进程是否有未保存的数据'
      ]
    )

    // 编辑关键配置文件
    this.addWarningRule(
      /^(vi|vim|nano|emacs)\s+\/etc\/(passwd|shadow|ssh|nginx|apache|mysql)/,
      'system',
      '编辑关键配置文件',
      ['⚠️ 编辑系统关键配置文件'],
      [
        '编辑前先备份：cp file file.bak',
        '使用 visudo 编辑 sudoers',
        '测试配置语法后再重启服务'
      ]
    )

    // 下载并执行脚本
    this.addWarningRule(
      /(curl|wget)\s+.*\|\s*(bash|sh)/,
      'security',
      '下载并执行远程脚本',
      ['⚠️ 从网络下载并直接执行脚本，存在安全风险'],
      [
        '先下载脚本审查内容',
        '验证脚本来源的可信度',
        '在沙箱环境中测试'
      ]
    )

    // 修改用户权限
    this.addWarningRule(
      /^(usermod|groupmod|passwd)\s+/,
      'user',
      '修改用户/组配置',
      ['⚠️ 修改用户或组配置，可能影响访问控制'],
      [
        '确认用户名是否正确',
        '保留至少一个 sudo 用户',
        '记录修改前后的配置'
      ]
    )

    // 安装包
    this.addWarningRule(
      /^(apt|yum|dnf|pip|npm)\s+(install|remove|upgrade)/,
      'package',
      '软件包管理操作',
      ['⚠️ 安装、卸载或升级软件包'],
      [
        '查看将要安装的依赖包',
        '生产环境先在测试服务器验证',
        '记录已安装的包版本'
      ]
    )

    // 危险的重定向操作
    this.addDangerousRule(
      /^\s*>\s*\/(dev|proc|sys)\//,
      'system',
      '重定向到系统虚拟文件系统',
      [
        '🔴 高危：向系统虚拟文件系统写入数据',
        '可能导致系统不稳定或崩溃'
      ],
      [
        '确认目标路径是否正确',
        '避免直接写入 /dev, /proc, /sys'
      ]
    )

    // 修改环境变量
    this.addWarningRule(
      /^export\s+(PATH|LD_LIBRARY_PATH|PYTHONPATH)=/,
      'system',
      '修改关键环境变量',
      [
        '⚠️ 修改系统关键环境变量',
        '可能影响其他程序的正常运行'
      ],
      [
        '使用局部变量而非 export',
        '在脚本中设置而非全局环境',
        '记录原始值以便恢复'
      ]
    )

    // crontab 操作
    this.addWarningRule(
      /^crontab\s+(-e|-r)/,
      'service',
      '编辑或删除定时任务',
      [
        '⚠️ 修改定时任务配置',
        '错误的 cron 表达式可能导致意外执行'
      ],
      [
        '先备份当前 crontab: crontab -l > backup.txt',
        '仔细检查 cron 表达式',
        '测试命令是否能正常执行'
      ]
    )
  }

  addDangerousRule(pattern, category, description, warnings, suggestions) {
    this.dangerousPatterns.push({
      pattern,
      riskLevel: 'error',
      category,
      description,
      warnings,
      suggestions,
      isDangerous: true
    })
  }

  addWarningRule(pattern, category, description, warnings, suggestions) {
    this.warningPatterns.push({
      pattern,
      riskLevel: 'warning',
      category,
      description,
      warnings,
      suggestions,
      isDangerous: false
    })
  }

  initializeSafeCommands() {
    const safeCmds = [
      'ls', 'll', 'la', 'pwd', 'cd', 'cat', 'less', 'more',
      'head', 'tail', 'grep', 'find', 'which', 'whereis',
      'man', 'help', 'echo', 'date', 'cal', 'uptime',
      'top', 'htop', 'ps', 'df', 'du', 'free',
      'ping', 'traceroute', 'netstat', 'ss', 'ip', 'ifconfig',
      'hostname', 'uname', 'whoami', 'id', 'groups',
      'history', 'alias', 'env', 'printenv',
      'mkdir', 'touch', 'cp', 'mv', 'ln',
      'tar', 'zip', 'unzip', 'gzip', 'gunzip',
      'journalctl', 'dmesg', 'lscpu', 'lsblk', 'lspci'
    ]
    
    safeCmds.forEach(cmd => this.safeCommands.add(cmd))
  }

  analyzeCommand(command) {
    const cmd = command.trim()
    
    if (!cmd) {
      return {
        riskLevel: 'info',
        category: 'other',
        description: '空命令',
        warnings: [],
        suggestions: [],
        isDangerous: false
      }
    }

    // 检测特殊操作标记
    if (cmd.includes('[Ctrl+C]')) {
      return {
        riskLevel: 'warning',
        category: 'process',
        description: '用户中断命令执行',
        warnings: ['⚠️ 用户手动中断了命令执行'],
        suggestions: [
          '检查命令是否需要更长时间执行',
          '确认是否有资源限制问题'
        ],
        isDangerous: false
      }
    }

    if (cmd.includes('[Ctrl+D]')) {
      return {
        riskLevel: 'warning',
        category: 'system',
        description: '用户发送 EOF 信号',
        warnings: ['⚠️ 用户可能正在退出 Shell 会话'],
        suggestions: [
          '确认是否要结束会话',
          '确保所有工作已保存'
        ],
        isDangerous: false
      }
    }

    // 提取基础命令
    const baseCmd = this.extractBaseCommand(cmd)
    if (this.safeCommands.has(baseCmd)) {
      return {
        riskLevel: 'info',
        category: 'other',
        description: '常用安全命令',
        warnings: [],
        suggestions: [],
        isDangerous: false
      }
    }

    // 检查危险规则
    for (const rule of this.dangerousPatterns) {
      if (rule.pattern.test(cmd)) {
        return {
          riskLevel: rule.riskLevel,
          category: rule.category,
          description: rule.description,
          warnings: rule.warnings,
          suggestions: rule.suggestions,
          isDangerous: rule.isDangerous
        }
      }
    }

    // 检查警告规则
    for (const rule of this.warningPatterns) {
      if (rule.pattern.test(cmd)) {
        return {
          riskLevel: rule.riskLevel,
          category: rule.category,
          description: rule.description,
          warnings: rule.warnings,
          suggestions: rule.suggestions,
          isDangerous: rule.isDangerous
        }
      }
    }

    // 默认返回安全
    return {
      riskLevel: 'info',
      category: 'other',
      description: '普通命令',
      warnings: [],
      suggestions: [],
      isDangerous: false
    }
  }

  extractBaseCommand(command) {
    const parts = command.trim().split(/\s+/)
    if (parts.length === 0) return ''
    
    let base = parts[0]
    
    // 处理 sudo, time 等前缀
    if (['sudo', 'time', 'nice'].includes(base)) {
      return parts[1] || ''
    }
    
    return base
  }

  getRiskLevelString(level) {
    const levels = {
      'error': '危险',
      'warning': '警告',
      'info': '安全'
    }
    return levels[level] || level
  }

  getCategoryString(category) {
    const categories = {
      'file': '文件操作',
      'system': '系统配置',
      'network': '网络操作',
      'process': '进程管理',
      'disk': '磁盘操作',
      'security': '安全相关',
      'package': '包管理',
      'user': '用户管理',
      'service': '服务管理',
      'other': '其他'
    }
    return categories[category] || category
  }
}

// 导出单例
export const commandAnalyzer = new CommandSecurityAnalyzer()
