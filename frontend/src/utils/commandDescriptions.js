/**
 * 常见命令说明库
 * 提供命令的中文描述和分类
 */

export const commandDescriptions = {
  // 文件操作
  'ls': '列出目录内容',
  'll': '详细列出目录内容',
  'la': '列出所有文件（包括隐藏）',
  'cd': '切换目录',
  'pwd': '显示当前工作目录',
  'mkdir': '创建目录',
  'rmdir': '删除空目录',
  'touch': '创建空文件或更新时间戳',
  'cp': '复制文件或目录',
  'mv': '移动或重命名文件',
  'rm': '删除文件或目录',
  'ln': '创建链接',
  
  // 文件查看
  'cat': '查看文件内容',
  'less': '分页查看文件',
  'more': '分页查看文件',
  'head': '查看文件开头',
  'tail': '查看文件末尾',
  'grep': '搜索文本模式',
  'find': '查找文件',
  'which': '查找命令位置',
  'whereis': '查找命令、源码和手册页',
  
  // 系统信息
  'uname': '显示系统信息',
  'hostname': '显示主机名',
  'whoami': '显示当前用户',
  'id': '显示用户ID和组ID',
  'groups': '显示用户所属组',
  'date': '显示或设置日期时间',
  'cal': '显示日历',
  'uptime': '显示系统运行时间',
  
  // 进程管理
  'ps': '显示进程状态',
  'top': '实时显示进程',
  'htop': '交互式进程查看器',
  'kill': '终止进程',
  'killall': '按名称终止进程',
  'pkill': '按模式终止进程',
  'bg': '将作业放到后台',
  'fg': '将作业放到前台',
  'jobs': '列出后台作业',
  
  // 磁盘管理
  'df': '显示磁盘空间使用情况',
  'du': '显示目录空间使用',
  'free': '显示内存使用情况',
  'mount': '挂载文件系统',
  'umount': '卸载文件系统',
  'fdisk': '磁盘分区工具',
  'mkfs': '创建文件系统',
  'fsck': '检查和修复文件系统',
  
  // 网络
  'ping': '测试网络连接',
  'traceroute': '跟踪路由路径',
  'netstat': '显示网络连接',
  'ss': '显示套接字统计',
  'ip': 'IP 配置工具',
  'ifconfig': '网络接口配置',
  'curl': '传输数据',
  'wget': '下载文件',
  'scp': '安全复制文件',
  'ssh': '安全 shell 连接',
  
  // 权限管理
  'chmod': '修改文件权限',
  'chown': '修改文件所有者',
  'chgrp': '修改文件组',
  'sudo': '以超级用户权限执行',
  'su': '切换用户',
  'passwd': '修改密码',
  
  // 包管理
  'apt': 'Debian/Ubuntu 包管理',
  'yum': 'RHEL/CentOS 包管理',
  'dnf': 'Fedora 包管理',
  'pip': 'Python 包管理',
  'npm': 'Node.js 包管理',
  'dpkg': 'Debian 包管理器',
  'rpm': 'RPM 包管理器',
  
  // 压缩归档
  'tar': '打包和解包',
  'gzip': 'GNU 压缩',
  'gunzip': 'GNU 解压缩',
  'zip': 'ZIP 压缩',
  'unzip': 'ZIP 解压缩',
  '7z': '7-Zip 压缩',
  
  // 文本处理
  'echo': '输出文本',
  'printf': '格式化输出',
  'sed': '流编辑器',
  'awk': '文本处理工具',
  'cut': '剪切文本列',
  'sort': '排序文本行',
  'uniq': '去重',
  'wc': '统计行数、字数、字节数',
  
  // 系统服务
  'systemctl': '系统服务管理',
  'service': '服务管理',
  'journalctl': '查看系统日志',
  'dmesg': '显示内核消息',
  'crontab': '定时任务管理',
  
  // 安全相关
  'iptables': '防火墙规则管理',
  'ufw': '简单防火墙',
  'firewall-cmd': 'firewalld 管理',
  'ssh-keygen': 'SSH 密钥生成',
  'openssl': 'SSL/TLS 工具',
  
  // 开发工具
  'git': '版本控制',
  'gcc': 'C 编译器',
  'g++': 'C++ 编译器',
  'make': '构建自动化工具',
  'cmake': '跨平台构建系统',
  'python': 'Python 解释器',
  'node': 'Node.js 运行时',
  
  // 危险命令（特殊标记）
  'dd': '⚠️ 转换和复制文件（可破坏数据）',
  'mkfs': '⚠️ 创建文件系统（会格式化）',
  'fdisk': '⚠️ 磁盘分区（可能丢失数据）',
  'parted': '⚠️ 分区编辑工具',
  'format': '⚠️ 格式化磁盘',
}

/**
 * 获取命令的中文描述
 */
export function getCommandDescription(command) {
  if (!command) return ''
  
  // 提取基础命令（去掉参数和路径）
  const baseCmd = command.trim().split(/\s+/)[0].split('/').pop()
  
  return commandDescriptions[baseCmd] || ''
}

/**
 * 获取命令分类
 */
export function getCommandCategory(command) {
  if (!command) return 'other'
  
  const baseCmd = command.trim().split(/\s+/)[0].split('/').pop()
  
  // 根据命令前缀分类
  if (['ls', 'cd', 'pwd', 'mkdir', 'cp', 'mv', 'rm', 'touch', 'cat', 'less', 'head', 'tail', 'grep', 'find'].includes(baseCmd)) {
    return 'file'
  }
  if (['ps', 'top', 'kill', 'jobs', 'bg', 'fg'].includes(baseCmd)) {
    return 'process'
  }
  if (['df', 'du', 'free', 'mount', 'fdisk', 'mkfs'].includes(baseCmd)) {
    return 'disk'
  }
  if (['ping', 'curl', 'wget', 'ssh', 'scp', 'netstat', 'ip'].includes(baseCmd)) {
    return 'network'
  }
  if (['chmod', 'chown', 'sudo', 'su', 'passwd'].includes(baseCmd)) {
    return 'permission'
  }
  if (['apt', 'yum', 'pip', 'npm'].includes(baseCmd)) {
    return 'package'
  }
  if (['systemctl', 'service', 'crontab'].includes(baseCmd)) {
    return 'service'
  }
  
  return 'other'
}

/**
 * 获取分类的中文标签
 */
export function getCategoryLabel(category) {
  const labels = {
    'file': '文件操作',
    'process': '进程管理',
    'disk': '磁盘管理',
    'network': '网络操作',
    'permission': '权限管理',
    'package': '包管理',
    'service': '服务管理',
    'other': '命令'
  }
  return labels[category] || labels['other']
}

/**
 * 获取命令的动作标签（更具体的描述）
 */
export function getActionTag(command) {
  if (!command) return ''
  
  const baseCmd = command.trim().split(/\s+/)[0].split('/').pop()
  const fullCmd = command.trim().toLowerCase()
  
  // 根据具体命令和参数返回更精确的动作标签
  const actionTags = {
    'ls': '列出文件',
    'll': '详细列表',
    'la': '显示全部',
    'cd': '切换目录',
    'pwd': '显示路径',
    'mkdir': '创建目录',
    'rmdir': '删除目录',
    'touch': '创建文件',
    'cp': '复制文件',
    'mv': '移动文件',
    'rm': '删除文件',
    'ln': '创建链接',
    'cat': '查看内容',
    'less': '分页查看',
    'more': '分页浏览',
    'head': '查看开头',
    'tail': '查看末尾',
    'grep': '搜索文本',
    'find': '查找文件',
    'which': '定位命令',
    'whereis': '查找位置',
    'chmod': '修改权限',
    'chown': '更改所有者',
    'chgrp': '更改组别',
    'sudo': '提权执行',
    'su': '切换用户',
    'passwd': '修改密码',
    'ps': '查看进程',
    'top': '监控进程',
    'kill': '终止进程',
    'df': '磁盘空间',
    'du': '目录大小',
    'free': '内存使用',
    'mount': '挂载设备',
    'umount': '卸载设备',
    'ping': '测试连接',
    'curl': '传输数据',
    'wget': '下载文件',
    'ssh': '远程连接',
    'scp': '安全复制',
    'tar': '打包压缩',
    'gzip': '压缩文件',
    'unzip': '解压文件',
    'systemctl': '管理服务',
    'service': '控制系统',
    'crontab': '定时任务',
    'git': '版本控制',
    'python': '运行脚本',
    'node': '执行代码',
  }
  
  // 特殊危险命令的动作标签
  if (fullCmd.includes('rm -rf') || fullCmd.includes('rm -fr')) {
    return '强制删除'
  }
  if (fullCmd.includes('dd ')) {
    return '底层写入'
  }
  if (fullCmd.includes('mkfs')) {
    return '格式化'
  }
  if (fullCmd.includes('fdisk') || fullCmd.includes('parted')) {
    return '分区操作'
  }
  if (fullCmd.includes(':(){')) {
    return '进程炸弹'
  }
  if (fullCmd.includes('iptables -f') || fullCmd.includes('ufw disable')) {
    return '禁用防火墙'
  }
  if (fullCmd.includes('kill -9 -1')) {
    return '杀死所有进程'
  }
  
  return actionTags[baseCmd] || ''
}
