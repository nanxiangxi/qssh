/**
 * xterm.js 代码高亮插件
 * 根据当前主题自动切换高亮颜色
 */

// 语言关键字映射
const KEYWORDS = {
  shell: ['if', 'then', 'else', 'elif', 'fi', 'for', 'while', 'do', 'done', 'case', 'esac', 'function', 'return', 'in', 'select', 'until', 'shift', 'export', 'source', 'alias', 'unalias', 'cd', 'ls', 'grep', 'sed', 'awk', 'find', 'cat', 'echo', 'printf', 'read', 'test', 'exit', 'set', 'unset', 'declare', 'local', 'readonly', 'trap', 'exec', 'eval', 'break', 'continue'],
  python: ['def', 'class', 'if', 'elif', 'else', 'for', 'while', 'return', 'import', 'from', 'as', 'try', 'except', 'finally', 'raise', 'with', 'yield', 'lambda', 'pass', 'break', 'continue', 'and', 'or', 'not', 'in', 'is', 'True', 'False', 'None', 'print', 'self'],
  javascript: ['const', 'let', 'var', 'function', 'return', 'if', 'else', 'for', 'while', 'do', 'switch', 'case', 'break', 'continue', 'new', 'this', 'class', 'extends', 'import', 'export', 'default', 'async', 'await', 'try', 'catch', 'finally', 'throw', 'typeof', 'instanceof', 'true', 'false', 'null', 'undefined'],
  go: ['func', 'package', 'import', 'return', 'if', 'else', 'for', 'range', 'switch', 'case', 'default', 'break', 'continue', 'go', 'chan', 'select', 'defer', 'map', 'struct', 'interface', 'type', 'var', 'const', 'true', 'false', 'nil'],
}

// 深色主题颜色
const DARK_COLORS = {
  keyword: '\x1b[35m',   // 紫色
  string: '\x1b[32m',    // 绿色
  number: '\x1b[36m',    // 青色
  comment: '\x1b[90m',   // 灰色
  reset: '\x1b[0m',
}

// 浅色主题颜色（高对比度，避免紫色/黑色混淆）
const LIGHT_COLORS = {
  keyword: '\x1b[34m',   // 蓝色
  string: '\x1b[32m',    // 绿色
  number: '\x1b[33m',    // 黄色/橙色
  comment: '\x1b[90m',   // 灰色
  reset: '\x1b[0m',
}

// 获取当前主题
function getCurrentTheme() {
  return document.documentElement.dataset.theme || 'dark'
}

// 获取当前主题的颜色
function getColors() {
  return getCurrentTheme() === 'light' ? LIGHT_COLORS : DARK_COLORS
}

/**
 * 为输出添加语法高亮
 */
// 剥离所有 ANSI/OSC 转义序列，只保留可见文本
function stripAnsiSequences(str) {
  // \x1b] ... \x07 (OSC)  \x1b] ... \x1b\\ (OSC ST)
  // \x1b[ ... letter (CSI)  \x1b[ ... space+letter (私有)
  // \x1b] ... \x07
  // \x1b[^\x1b\x07]*[\x07\x1b\\]
  return str
    .replace(/\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)/g, '')  // OSC 序列
    .replace(/\x1b\[[\x30-\x3f]*[\x20-\x2f]*[\x40-\x7e]/g, '')  // CSI 序列
    .replace(/\x1b[\x20-\x2f]*[\x30-\x3f]*[\x40-\x5f]/g, '')  // 其他转义
    .replace(/\x1b./g, '')  // fallback: 任何剩余的 ESC + 字符
}

export function highlightOutput(output) {
  if (!output || output.length > 50000) return output

  // 如果输出中包含转义序列，先剥离再判断是否有意义
  if (output.includes('\x1b') || output.includes('\x07')) {
    const stripped = stripAnsiSequences(output)
    // 如果剥离后无可显示内容，直接返回原始输出（让 xterm 处理转义）
    if (!stripped.trim()) return output
    // 对纯文本部分做高亮，再把原始转义序列拼回去会很复杂且易出错
    // 最安全的做法：已有转义序列的输出不做额外高亮，直接交给 xterm 渲染
    return output
  }

  const lang = detectLanguage(output)
  const keywords = KEYWORDS[lang] || KEYWORDS.shell
  const COLORS = getColors()

  let result = output

  // 高亮注释
  result = result.replace(/(\/\/.*$|#.*$)/gm, (match) => {
    if (match.includes('\x1b[')) return match
    return COLORS.comment + match + COLORS.reset
  })

  // 高亮字符串
  result = result.replace(/(["'])(?:(?!\1|\\).|\\.)*\1/g, (match) => {
    if (match.includes('\x1b[')) return match
    return COLORS.string + match + COLORS.reset
  })

  // 高亮数字
  result = result.replace(/\b(\d+\.?\d*)\b/g, (match) => {
    return COLORS.number + match + COLORS.reset
  })

  // 高亮关键字
  const keywordPattern = new RegExp(`\\b(${keywords.join('|')})\\b`, 'g')
  result = result.replace(keywordPattern, (match) => {
    return COLORS.keyword + match + COLORS.reset
  })

  return result
}

/**
 * 检测输出中使用的编程语言
 */
function detectLanguage(output) {
  if (output.includes('def ') || output.includes('import ') || output.includes('print(')) return 'python'
  if (output.includes('func ') || output.includes('package ') || output.includes('import (')) return 'go'
  if (output.includes('const ') || output.includes('let ') || output.includes('function ')) return 'javascript'
  return 'shell'
}
