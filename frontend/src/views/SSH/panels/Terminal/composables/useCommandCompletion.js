/**
 * 命令补全组合式函数
 * 支持命令、子命令、选项、远程路径补全
 */
import { ref, computed } from 'vue'
import { SSHService } from '@bindings/changeme/ssh/index.js'

// 内置命令库
const BUILTIN_COMMANDS = {
  // 文件操作
  ls: {
    options: ['-l', '-a', '-la', '-lh', '-R', '-h', '--help', '--color'],
    description: '列出目录内容'
  },
  cd: {
    options: ['..', '~', '-', '/'],
    description: '切换目录'
  },
  cp: {
    options: ['-r', '-f', '-i', '-v', '-p', '--help'],
    description: '复制文件或目录'
  },
  mv: {
    options: ['-f', '-i', '-v', '-n', '--help'],
    description: '移动/重命名文件'
  },
  rm: {
    options: ['-r', '-f', '-rf', '-i', '-v', '--help'],
    description: '删除文件或目录'
  },
  mkdir: {
    options: ['-p', '-v', '--help'],
    description: '创建目录'
  },
  touch: {
    options: ['-a', '-m', '-t', '--help'],
    description: '创建空文件/更新时间戳'
  },
  cat: {
    options: ['-n', '-b', '-s', '--help'],
    description: '查看文件内容'
  },
  less: {
    options: ['-N', '-S', '-R', '--help'],
    description: '分页查看文件'
  },
  head: {
    options: ['-n', '-c', '-q', '-v', '--help'],
    description: '查看文件头部'
  },
  tail: {
    options: ['-n', '-f', '-F', '-q', '-v', '--help'],
    description: '查看文件尾部'
  },
  find: {
    options: ['-name', '-type', '-size', '-mtime', '-exec', '-delete', '--help'],
    description: '查找文件'
  },
  grep: {
    options: ['-r', '-i', '-v', '-n', '-l', '-c', '-E', '-A', '-B', '-C', '--help'],
    description: '搜索文本'
  },
  chmod: {
    options: ['-R', '-v', '--help'],
    description: '修改文件权限'
  },
  chown: {
    options: ['-R', '-v', '--help'],
    description: '修改文件所有者'
  },
  ln: {
    options: ['-s', '-f', '-v', '--help'],
    description: '创建链接'
  },
  tar: {
    options: ['-x', '-c', '-v', '-f', '-z', '-j', '-J', '-t', '--help'],
    description: '打包/解包文件'
  },
  zip: {
    options: ['-r', '-q', '-v', '--help'],
    description: '压缩文件'
  },
  unzip: {
    options: ['-l', '-o', '-q', '-v', '--help'],
    description: '解压文件'
  },

  // 系统信息
  uname: {
    options: ['-a', '-r', '-m', '-n', '-s', '--help'],
    description: '系统信息'
  },
  whoami: {
    options: ['--help'],
    description: '当前用户'
  },
  hostname: {
    options: ['-f', '-i', '-I', '--help'],
    description: '主机名'
  },
  date: {
    options: ['+%Y-%m-%d', '+%H:%M:%S', '+%Y%m%d', '-u', '--help'],
    description: '日期时间'
  },
  uptime: {
    options: ['-p', '-s', '--help'],
    description: '系统运行时间'
  },
  df: {
    options: ['-h', '-T', '-i', '--help'],
    description: '磁盘空间'
  },
  du: {
    options: ['-h', '-s', '-a', '--help'],
    description: '目录大小'
  },
  free: {
    options: ['-h', '-m', '-g', '-t', '--help'],
    description: '内存使用'
  },
  top: {
    options: ['-b', '-n', '-p', '-u', '--help'],
    description: '进程监控'
  },
  ps: {
    options: ['-aux', '-ef', '-u', '-p', '--help'],
    description: '进程列表'
  },
  kill: {
    options: ['-9', '-15', '-TERM', '-HUP', '--help'],
    description: '终止进程'
  },
  killall: {
    options: ['-9', '-u', '--help'],
    description: '按名称终止进程'
  },

  // 网络
  ping: {
    options: ['-c', '-i', '-W', '-t', '--help'],
    description: '测试网络连接'
  },
  curl: {
    options: ['-o', '-O', '-L', '-k', '-s', '-S', '-X', '-H', '-d', '-u', '--help'],
    description: 'HTTP 请求'
  },
  wget: {
    options: ['-O', '-P', '-c', '-q', '-r', '--help'],
    description: '下载文件'
  },
  ssh: {
    options: ['-p', '-i', '-L', '-R', '-D', '-N', '-v', '--help'],
    description: 'SSH 连接'
  },
  scp: {
    options: ['-r', '-P', '-i', '-q', '--help'],
    description: '远程复制'
  },
  rsync: {
    options: ['-a', '-v', '-z', '-r', '-P', '-e', '--help'],
    description: '同步文件'
  },
  netstat: {
    options: ['-tuln', '-an', '-p', '-r', '--help'],
    description: '网络状态'
  },
  ss: {
    options: ['-tuln', '-an', '-p', '-r', '--help'],
    description: '套接字统计'
  },
  ifconfig: {
    options: ['-a', '--help'],
    description: '网络接口配置'
  },
  ip: {
    options: ['addr', 'link', 'route', 'neigh', '--help'],
    subcommands: {
      addr: ['show', 'add', 'del', 'flush'],
      link: ['show', 'set', 'add', 'del'],
      route: ['show', 'add', 'del', 'get']
    },
    description: '网络配置'
  },

  // 包管理
  apt: {
    subcommands: ['update', 'upgrade', 'install', 'remove', 'search', 'list', 'autoclean', 'clean'],
    options: ['-y', '-q', '--help'],
    description: 'Debian/Ubuntu 包管理'
  },
  yum: {
    subcommands: ['install', 'update', 'remove', 'search', 'list', 'clean', 'info'],
    options: ['-y', '-q', '--help'],
    description: 'CentOS/RHEL 包管理'
  },
  dnf: {
    subcommands: ['install', 'update', 'remove', 'search', 'list', 'clean', 'info'],
    options: ['-y', '-q', '--help'],
    description: 'Fedora 包管理'
  },
  pacman: {
    subcommands: ['-S', '-R', '-Q', '-Syu', '-Ss', '-Si'],
    options: ['-y', '-u', '--noconfirm', '--help'],
    description: 'Arch Linux 包管理'
  },
  pip: {
    subcommands: ['install', 'uninstall', 'list', 'show', 'search', 'freeze'],
    options: ['-r', '-U', '-q', '--help'],
    description: 'Python 包管理'
  },


  // Docker
  docker: {
    subcommands: ['run', 'ps', 'images', 'pull', 'push', 'build', 'exec', 'logs', 'stop', 'rm', 'rmi', 'compose'],
    options: ['-d', '-it', '-p', '-v', '-e', '--name', '--help'],
    description: 'Docker 容器管理'
  },
  'docker-compose': {
    subcommands: ['up', 'down', 'ps', 'logs', 'build', 'pull', 'exec', 'restart'],
    options: ['-d', '-f', '--help'],
    description: 'Docker Compose 管理'
  },

  // Git
  git: {
    subcommands: ['init', 'clone', 'add', 'commit', 'push', 'pull', 'fetch', 'status', 'log', 'diff', 'branch', 'checkout', 'merge', 'rebase', 'stash', 'reset', 'remote', 'tag', 'cherry-pick'],
    options: ['--help', '--version'],
    description: 'Git 版本控制',
    subcommandOptions: {
      add: ['.', '-A', '-u', '-p', '--help'],
      commit: ['-m', '-a', '--amend', '-s', '--help'],
      push: ['-u', '-f', '--tags', '--help'],
      pull: ['--rebase', '--help'],
      branch: ['-a', '-d', '-D', '-r', '--help'],
      checkout: ['-b', '-', '--help'],
      log: ['--oneline', '--graph', '--all', '-n', '--help'],
      status: ['-s', '-b', '--help'],
      diff: ['--staged', '--stat', '--help'],
      reset: ['--soft', '--mixed', '--hard', 'HEAD~1', '--help'],
      stash: ['push', 'pop', 'list', 'drop', 'clear', '--help']
    }
  },

  // Systemd
  systemctl: {
    subcommands: ['start', 'stop', 'restart', 'status', 'enable', 'disable', 'reload', 'daemon-reload', 'list-units', 'list-timers'],
    options: ['--help', '--no-pager'],
    description: '系统服务管理'
  },
  journalctl: {
    options: ['-u', '-f', '-n', '--since', '--until', '--no-pager', '--help'],
    description: '系统日志'
  },

  // 其他常用
  echo: {
    options: ['-n', '-e', '--help'],
    description: '输出文本'
  },
  export: {
    options: [],
    description: '设置环境变量'
  },
  source: {
    options: [],
    description: '加载配置文件'
  },
  sudo: {
    options: ['-u', '-i', '-s', '--help'],
    description: '以管理员权限执行'
  },
  su: {
    options: ['-', '-l', '--help'],
    description: '切换用户'
  },
  passwd: {
    options: ['--help'],
    description: '修改密码'
  },
  alias: {
    options: [],
    description: '命令别名'
  },
  crontab: {
    options: ['-l', '-e', '-r', '--help'],
    description: '定时任务'
  },
  man: {
    options: [],
    description: '查看手册'
  },
  which: {
    options: ['-a', '--help'],
    description: '查找命令路径'
  },
  whereis: {
    options: ['-b', '-m', '-s', '--help'],
    description: '查找命令相关文件'
  },
  locate: {
    options: ['-i', '-r', '--help'],
    description: '快速查找文件'
  },
  awk: {
    options: ['-F', '-v', '--help'],
    description: '文本处理'
  },
  sed: {
    options: ['-i', '-e', '-n', '--help'],
    description: '流编辑器'
  },
  sort: {
    options: ['-r', '-n', '-k', '-u', '-t', '--help'],
    description: '排序'
  },
  uniq: {
    options: ['-c', '-d', '-u', '--help'],
    description: '去重'
  },
  wc: {
    options: ['-l', '-w', '-c', '--help'],
    description: '统计行数/字数'
  },
  tee: {
    options: ['-a', '--help'],
    description: '输出到文件和屏幕'
  },
  xargs: {
    options: ['-I', '-n', '-p', '--help'],
    description: '参数传递'
  },
  watch: {
    options: ['-n', '-d', '-t', '--help'],
    description: '定时执行命令'
  },
  screen: {
    options: ['-S', '-r', '-ls', '-d', '--help'],
    description: '终端复用'
  },
  tmux: {
    subcommands: ['new', 'attach', 'detach', 'ls', 'kill-session'],
    options: ['-s', '-t', '--help'],
    description: '终端复用'
  },
  vim: {
    options: ['+n', '-R', '-d', '--help'],
    description: '文本编辑器'
  },
  nano: {
    options: ['-w', '-m', '--help'],
    description: '文本编辑器'
  },

  // ========== 补充常用命令 ==========
  help: {
    options: [],
    description: '显示帮助信息'
  },
  history: {
    options: ['-c', '-d', '-a', '-n', '--help'],
    description: '命令历史'
  },
  halt: {
    options: ['-f', '-p', '--help'],
    description: '关机'
  },
  reboot: {
    options: ['-f', '--help'],
    description: '重启'
  },
  poweroff: {
    options: ['-f', '--help'],
    description: '关机'
  },
  shutdown: {
    options: ['-h', '-r', '-c', '-k', '--help'],
    description: '关机/重启'
  },
  mount: {
    options: ['-t', '-o', '-a', '--help'],
    description: '挂载文件系统'
  },
  umount: {
    options: ['-a', '-f', '-l', '--help'],
    description: '卸载文件系统'
  },
  fdisk: {
    options: ['-l', '-s', '--help'],
    description: '磁盘分区'
  },
  mkfs: {
    options: ['-t', '--help'],
    description: '格式化文件系统'
  },
  fsck: {
    options: ['-a', '-y', '-r', '--help'],
    description: '文件系统检查'
  },
  lsof: {
    options: ['-i', '-p', '-u', '-n', '--help'],
    description: '列出打开的文件'
  },
  strace: {
    options: ['-p', '-e', '-f', '--help'],
    description: '系统调用追踪'
  },
  ltrace: {
    options: ['-p', '-e', '-f', '--help'],
    description: '库调用追踪'
  },
  gdb: {
    options: ['-p', '-ex', '--help'],
    description: '调试器'
  },
  nm: {
    options: ['-D', '-S', '--help'],
    description: '查看符号表'
  },
  objdump: {
    options: ['-d', '-t', '-s', '--help'],
    description: '查看目标文件'
  },
  readelf: {
    options: ['-a', '-h', '-S', '--help'],
    description: '查看 ELF 文件'
  },
  make: {
    options: ['-f', '-j', '-C', '--help'],
    description: '构建工具'
  },
  cmake: {
    options: ['-S', '-B', '-G', '-D', '--help'],
    description: '构建系统生成器'
  },
  gcc: {
    options: ['-o', '-c', '-Wall', '-g', '-O2', '--help'],
    description: 'C 编译器'
  },
  'g++': {
    options: ['-o', '-c', '-Wall', '-g', '-O2', '-std=c++17', '--help'],
    description: 'C++ 编译器'
  },
  java: {
    options: ['-cp', '-jar', '--help'],
    description: 'Java 运行时'
  },
  javac: {
    options: ['-cp', '-d', '--help'],
    description: 'Java 编译器'
  },
  python: {
    options: ['-c', '-m', '-V', '--help'],
    description: 'Python 解释器'
  },
  python3: {
    options: ['-c', '-m', '-V', '--help'],
    description: 'Python3 解释器'
  },
  node: {
    options: ['-e', '-v', '--help'],
    description: 'Node.js 运行时'
  },
  npm: {
    subcommands: ['install', 'uninstall', 'update', 'list', 'run', 'init', 'publish', 'test', 'start', 'build'],
    options: ['-g', '-S', '-D', '-f', '--help'],
    description: 'Node.js 包管理'
  },
  yarn: {
    subcommands: ['add', 'remove', 'install', 'run', 'init', 'publish'],
    options: ['-g', '--help'],
    description: 'Yarn 包管理'
  },
  go: {
    subcommands: ['build', 'run', 'test', 'get', 'mod', 'fmt', 'vet', 'lint'],
    options: ['-v', '-x', '--help'],
    description: 'Go 语言工具'
  },
  cargo: {
    subcommands: ['build', 'run', 'test', 'new', 'init', 'add', 'publish'],
    options: ['--release', '--help'],
    description: 'Rust 包管理'
  },
  mvn: {
    subcommands: ['clean', 'compile', 'test', 'package', 'install', 'deploy'],
    options: ['-D', '-P', '--help'],
    description: 'Maven 构建工具'
  },
  gradle: {
    subcommands: ['build', 'test', 'run', 'clean', 'assemble'],
    options: ['--daemon', '--help'],
    description: 'Gradle 构建工具'
  },
  terraform: {
    subcommands: ['init', 'plan', 'apply', 'destroy', 'validate', 'fmt'],
    options: ['-var', '-var-file', '--help'],
    description: '基础设施即代码'
  },
  ansible: {
    subcommands: ['playbook', 'inventory', 'vault', 'galaxy'],
    options: ['-i', '-u', '--help'],
    description: '自动化工具'
  },
  kubectl: {
    subcommands: ['get', 'describe', 'create', 'apply', 'delete', 'logs', 'exec', 'port-forward'],
    options: ['-n', '-o', '-f', '--help'],
    description: 'Kubernetes CLI'
  },
  helm: {
    subcommands: ['install', 'upgrade', 'uninstall', 'list', 'search', 'repo'],
    options: ['-n', '--set', '--help'],
    description: 'Kubernetes 包管理'
  },
  vault: {
    subcommands: ['login', 'read', 'write', 'delete', 'list', 'status'],
    options: ['-address', '-token', '--help'],
    description: '密钥管理'
  }
}

// 常用路径
const COMMON_PATHS = [
  '~',
  '~/Desktop',
  '~/Documents',
  '~/Downloads',
  '/etc',
  '/var',
  '/var/log',
  '/var/www',
  '/tmp',
  '/opt',
  '/usr',
  '/usr/local',
  '/usr/bin',
  '/usr/sbin',
  '/home',
  '/root',
  '/tmp',
  '/dev',
  '/proc',
  '/sys'
]

export function useCommandCompletion() {
  // 补全状态
  const suggestions = ref([])
  const selectedIndex = ref(-1)
  const isVisible = ref(false)

  // 计算属性
  const hasSuggestions = computed(() => suggestions.value.length > 0)
  const selectedSuggestion = computed(() => suggestions.value[selectedIndex.value])

  // 获取补全建议（同步版本，不含远程路径补全）
  function getSuggestions(input, context = {}) {
    if (!input) {
      suggestions.value = []
      return []
    }

    const parts = input.split(/\s+/)
    const results = []

    // 第一个词：命令补全
    if (parts.length <= 1) {
      const prefix = parts[0] || ''
      const commandMatches = Object.keys(BUILTIN_COMMANDS)
        .filter(cmd => cmd.startsWith(prefix))
        .map(cmd => ({
          type: 'command',
          value: cmd,
          display: cmd,
          description: BUILTIN_COMMANDS[cmd].description
        }))

      results.push(...commandMatches)
    }
    // 后续词：选项/子命令/路径补全
    else {
      const command = parts[0]
      const currentWord = parts[parts.length - 1]
      const prevWord = parts[parts.length - 2]
      const cmdDef = BUILTIN_COMMANDS[command]

      if (cmdDef) {
        // 选项补全
        if (currentWord.startsWith('-')) {
          const options = cmdDef.options || []
          const optionMatches = options
            .filter(opt => opt.startsWith(currentWord))
            .map(opt => ({
              type: 'option',
              value: opt,
              display: opt,
              description: '选项'
            }))
          results.push(...optionMatches)
        }
        // 子命令补全
        else if (cmdDef.subcommands && !currentWord.startsWith('-')) {
          const subcommands = cmdDef.subcommands || []
          const subMatches = subcommands
            .filter(sub => sub.startsWith(currentWord))
            .map(sub => ({
              type: 'subcommand',
              value: sub,
              display: sub,
              description: `${command} 子命令`
            }))
          results.push(...subMatches)
        }
        // Git 子命令选项补全
        else if (cmdDef.subcommandOptions && prevWord) {
          const subOpts = cmdDef.subcommandOptions[prevWord] || []
          const subOptMatches = subOpts
            .filter(opt => opt.startsWith(currentWord))
            .map(opt => ({
              type: 'option',
              value: opt,
              display: opt,
              description: `${prevWord} 选项`
            }))
          results.push(...subOptMatches)
        }
      }

      // 路径补全（如果当前词看起来像路径）
      if (currentWord.startsWith('/') || currentWord.startsWith('~') || currentWord.startsWith('.')) {
        const pathMatches = getLocalPathCompletions(currentWord, context.cwd)
        results.push(...pathMatches)
      }
    }

    suggestions.value = results.slice(0, 20) // 限制最多 20 个建议
    selectedIndex.value = results.length > 0 ? 0 : -1
    isVisible.value = results.length > 0

    return results
  }

  // 获取补全建议（异步版本，包含远程路径补全，支持相对路径）
  async function getSuggestionsAsync(input, context = {}) {
    if (!input) {
      suggestions.value = []
      return []
    }

    // 先获取同步结果（命令、选项等）
    const syncResults = getSuggestions(input, context)

    // 如果有参数需要补全，尝试远程补全
    const parts = input.split(/\s+/)
    if (parts.length > 1 && context.connId) {
      const currentWord = parts[parts.length - 1]
      try {
        let remotePaths
        if (currentWord.startsWith('/') || currentWord.startsWith('~') || currentWord.startsWith('.')) {
          // 绝对路径或特殊路径
          remotePaths = await getRemotePathCompletions(currentWord, context.connId)
        } else {
          // 相对路径：列出当前工作目录
          remotePaths = await getRemotePathCompletionsCwd(currentWord, context.connId)
        }
        if (remotePaths.length > 0) {
          // 移除本地路径补全结果，替换为远程结果
          const nonPathResults = syncResults.filter(r => r.type !== 'path')
          const finalResults = [...nonPathResults, ...remotePaths].slice(0, 20)
          suggestions.value = finalResults
          selectedIndex.value = finalResults.length > 0 ? 0 : -1
          isVisible.value = finalResults.length > 0
          return finalResults
        }
      } catch (e) {
        // 远程补全失败，保留本地结果
      }
    }

    return syncResults
  }

  // 路径补全缓存（避免短时间内重复请求）
  let pathCache = { key: '', results: [], timestamp: 0 }
  const PATH_CACHE_TTL = 3000 // 3 秒缓存

  // 路径补全（支持远程服务器路径）
  function getPathCompletions(partial, context = {}) {
    // 先尝试远程补全
    if (context.connId) {
      return getRemotePathCompletions(partial, context.connId)
    }
    // 回退：本地静态路径
    return getLocalPathCompletions(partial, context.cwd)
  }

  // 远程路径补全
  async function getRemotePathCompletions(partial, connId) {
    try {
      // 解析目录部分和前缀部分
      // 例如: "/etc/pas" → dir="/etc/", prefix="pas"
      // 例如: "~/Doc" → dir="~/", prefix="Doc"
      let dir, prefix

      const lastSlash = partial.lastIndexOf('/')
      if (lastSlash >= 0) {
        dir = partial.substring(0, lastSlash + 1)
        prefix = partial.substring(lastSlash + 1)
      } else {
        dir = ''
        prefix = partial
      }

      // 处理 ~ 开头的路径
      if (dir.startsWith('~') || partial.startsWith('~')) {
        // 将 ~ 替换为 $HOME 后再列目录
        // 但先尝试直接列出（某些服务器支持）
      }

      // 检查缓存
      const cacheKey = `${connId}:${dir}`
      const now = Date.now()
      if (pathCache.key === cacheKey && now - pathCache.timestamp < PATH_CACHE_TTL) {
        return filterPathResults(pathCache.results, prefix)
      }

      // 调用后端列出目录内容
      const targetDir = dir || './'
      const files = await SSHService.ListDirectory(connId, targetDir)

      if (!files || !Array.isArray(files)) {
        return getLocalPathCompletions(partial)
      }

      // 缓存结果
      pathCache = { key: cacheKey, results: files, timestamp: now }

      return filterPathResults(files, prefix, dir)
    } catch (e) {
      console.warn('[Completion] 远程路径补全失败，回退本地:', e)
      return getLocalPathCompletions(partial)
    }
  }

  // 从文件列表中筛选匹配前缀的补全项
  function filterPathResults(files, prefix, dirPrefix = '') {
    return files
      .filter(f => f.name.toLowerCase().startsWith(prefix.toLowerCase()))
      .map(f => {
        const fullPath = dirPrefix + f.name
        return {
          type: 'path',
          value: fullPath + (f.isDir ? '/' : ''),
          display: f.name + (f.isDir ? '/' : ''),
          description: f.isDir ? '目录' : '文件'
        }
      })
  }

  // 远程路径补全（当前工作目录，用于相对路径）
  async function getRemotePathCompletionsCwd(partial, connId) {
    try {
      const cacheKey = `${connId}:./`
      const now = Date.now()
      let files
      if (pathCache.key === cacheKey && now - pathCache.timestamp < PATH_CACHE_TTL) {
        files = pathCache.results
      } else {
        files = await SSHService.ListDirectory(connId, './')
        if (!files || !Array.isArray(files)) {
          return []
        }
        pathCache = { key: cacheKey, results: files, timestamp: now }
      }
      return files
        .filter(f => f.name.toLowerCase().startsWith(partial.toLowerCase()))
        .map(f => ({
          type: 'path',
          value: f.name + (f.isDir ? '/' : ''),
          display: f.name + (f.isDir ? '/' : ''),
          description: f.isDir ? '目录' : '文件'
        }))
    } catch (e) {
      console.warn('[Completion] 远程路径补全(CWD)失败:', e)
      return []
    }
  }

  // 本地静态路径补全（回退方案）
  function getLocalPathCompletions(partial, cwd) {
    const paths = [...COMMON_PATHS]

    if (cwd) {
      paths.unshift(cwd)
      paths.unshift('.')
      paths.unshift('..')
    }

    return paths
      .filter(p => p.startsWith(partial))
      .map(p => ({
        type: 'path',
        value: p,
        display: p,
        description: '路径'
      }))
  }

  // 应用补全
  function applyCompletion(input, suggestion) {
    if (!suggestion) return input

    const parts = input.split(/\s+/)

    // 命令补全
    if (parts.length <= 1) {
      return suggestion.value + ' '
    }

    // 替换最后一个词
    parts[parts.length - 1] = suggestion.value

    // 如果是选项，加空格；如果是路径，可能需要继续补全
    if (suggestion.type === 'path' && !suggestion.value.endsWith('/')) {
      return parts.join(' ')
    }

    return parts.join(' ') + ' '
  }

  // 选择下一个建议
  function selectNext() {
    if (selectedIndex.value < suggestions.value.length - 1) {
      selectedIndex.value++
    } else {
      selectedIndex.value = 0
    }
  }

  // 选择上一个建议
  function selectPrevious() {
    if (selectedIndex.value > 0) {
      selectedIndex.value--
    } else {
      selectedIndex.value = suggestions.value.length - 1
    }
  }

  // 获取当前选中的建议
  function getSelected() {
    return suggestions.value[selectedIndex.value] || null
  }

  // 隐藏建议列表
  function hide() {
    isVisible.value = false
    selectedIndex.value = -1
  }

  // 显示建议列表
  function show() {
    if (suggestions.value.length > 0) {
      isVisible.value = true
    }
  }

  // 重置
  function reset() {
    suggestions.value = []
    selectedIndex.value = -1
    isVisible.value = false
  }

  // 获取命令帮助
  function getCommandHelp(command) {
    const cmdDef = BUILTIN_COMMANDS[command]
    if (!cmdDef) return null

    return {
      command,
      description: cmdDef.description,
      options: cmdDef.options || [],
      subcommands: cmdDef.subcommands || []
    }
  }

  // 获取所有命令列表
  function getAllCommands() {
    return Object.entries(BUILTIN_COMMANDS).map(([cmd, def]) => ({
      command: cmd,
      description: def.description
    }))
  }

  return {
    // 状态
    suggestions,
    selectedIndex,
    isVisible,
    hasSuggestions,
    selectedSuggestion,

    // 方法
    getSuggestions,
    getSuggestionsAsync,
    applyCompletion,
    selectNext,
    selectPrevious,
    getSelected,
    hide,
    show,
    reset,
    getCommandHelp,
    getAllCommands
  }
}
