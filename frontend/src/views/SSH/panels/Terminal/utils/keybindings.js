/**
 * 终端快捷键配置
 *
 * 分类：
 * 1. TTY 信号 - 直接转发到 SSH，不修改
 * 2. Readline 编辑 - 本地处理 + 转发到 SSH
 * 3. 交互式 - 需要临时终端或切换到经典模式
 * 4. 程序快捷键 - 应用自身功能，不转发
 */

// TTY 控制信号（直接转发到 SSH）
export const TTY_SIGNALS = {
  'c': { code: '\x03', name: 'Ctrl+C', desc: '中断当前进程' },
  '\\': { code: '\x1c', name: 'Ctrl+\\', desc: '强制终止 (SIGQUIT)' },
  's': { code: '\x13', name: 'Ctrl+S', desc: '暂停输出 (XOFF)' },
  'q': { code: '\x11', name: 'Ctrl+Q', desc: '恢复输出 (XON)' },
  'd': { code: '\x04', name: 'Ctrl+D', desc: 'EOF / 退出 Shell' },
  'g': { code: '\x07', name: 'Ctrl+G', desc: '中止当前操作' },
}

// Readline 编辑快捷键（本地处理 + 转发到 SSH）
export const READLINE_KEYS = {
  // 光标移动
  'a': { code: '\x01', name: 'Ctrl+A', desc: '移到行首' },
  'e': { code: '\x05', name: 'Ctrl+E', desc: '移到行尾' },
  'b': { code: '\x02', name: 'Ctrl+B', desc: '后退一个字符' },
  'f': { code: '\x06', name: 'Ctrl+F', desc: '前进一个字符' },

  // 剪切
  'k': { code: '\x0b', name: 'Ctrl+K', desc: '剪切到行尾' },
  'u': { code: '\x15', name: 'Ctrl+U', desc: '剪切到行首' },
  'w': { code: '\x17', name: 'Ctrl+W', desc: '剪切前一个单词' },
  'h': { code: '\x08', name: 'Ctrl+H', desc: '删除前一个字符' },

  // 粘贴
  'y': { code: '\x19', name: 'Ctrl+Y', desc: '粘贴剪切内容' },

  // 其他
  't': { code: '\x14', name: 'Ctrl+T', desc: '交换光标前后字符' },
  'p': { code: '\x10', name: 'Ctrl+P', desc: '上一条历史命令' },
  'n': { code: '\x0e', name: 'Ctrl+N', desc: '下一条历史命令' },
  'l': { code: '\x0c', name: 'Ctrl+L', desc: '清屏' },
}

// Alt/Meta 组合键（转发到 SSH）
export const ALT_KEYS = {
  'b': { code: '\x1bb', name: 'Alt+B', desc: '后退一个单词' },
  'f': { code: '\x1bf', name: 'Alt+F', desc: '前进一个单词' },
  'd': { code: '\x1bd', name: 'Alt+D', desc: '剪切后一个单词' },
  'c': { code: '\x1bc', name: 'Alt+C', desc: '首字母大写' },
  'u': { code: '\x1bu', name: 'Alt+U', desc: '全部大写' },
  'l': { code: '\x1bl', name: 'Alt+L', desc: '全部小写' },
  '.': { code: '\x1b.', name: 'Alt+.', desc: '插入上条命令最后参数' },
}

// 交互式快捷键（需要临时终端或切换到经典模式）
export const INTERACTIVE_KEYS = {
  'r': { code: '\x12', name: 'Ctrl+R', desc: '反向搜索历史' },
  'z': { code: '\x1a', name: 'Ctrl+Z', desc: '暂停进程' },
}

// 程序快捷键（不转发到 SSH）
export const APP_KEYS = {
  'o': { name: 'Ctrl+O', desc: '打开临时终端' },
  'f': { name: 'Ctrl+F', desc: '搜索' },
}

// 获取所有快捷键分类
export function getKeybindingCategories() {
  return [
    {
      title: 'TTY 控制信号',
      items: Object.values(TTY_SIGNALS).map(k => ({ key: k.name, desc: k.desc }))
    },
    {
      title: '光标移动',
      items: [
        { key: 'Ctrl+A', desc: '移到行首' },
        { key: 'Ctrl+E', desc: '移到行尾' },
        { key: 'Ctrl+B', desc: '后退一个字符' },
        { key: 'Ctrl+F', desc: '前进一个字符' },
        { key: 'Alt+B', desc: '后退一个单词' },
        { key: 'Alt+F', desc: '前进一个单词' },
      ]
    },
    {
      title: '剪切与粘贴',
      items: [
        { key: 'Ctrl+K', desc: '剪切到行尾' },
        { key: 'Ctrl+U', desc: '剪切到行首' },
        { key: 'Ctrl+W', desc: '剪切前一个单词' },
        { key: 'Ctrl+H', desc: '删除前一个字符' },
        { key: 'Ctrl+Y', desc: '粘贴剪切内容' },
        { key: 'Ctrl+T', desc: '交换光标前后字符' },
        { key: 'Alt+D', desc: '剪切后一个单词' },
      ]
    },
    {
      title: '历史命令',
      items: [
        { key: 'Ctrl+P', desc: '上一条命令' },
        { key: 'Ctrl+N', desc: '下一条命令' },
        { key: 'Ctrl+R', desc: '反向搜索历史（临时终端）' },
        { key: '↑', desc: '上一条历史' },
        { key: '↓', desc: '下一条历史' },
      ]
    },
    {
      title: '补全',
      items: [
        { key: 'Tab', desc: '自动补全' },
      ]
    },
    {
      title: '程序功能',
      items: [
        { key: 'Ctrl+O', desc: '打开临时终端' },
        { key: 'Ctrl+F', desc: '搜索' },
        { key: 'Enter', desc: '执行命令' },
        { key: 'Esc', desc: '关闭弹窗' },
      ]
    }
  ]
}
