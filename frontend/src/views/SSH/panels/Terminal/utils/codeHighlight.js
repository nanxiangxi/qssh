/**
 * 终端输出代码高亮工具
 * 自动检测语言并高亮显示
 */
import hljs from 'highlight.js/lib/core'
import 'highlight.js/styles/atom-one-dark.css'
import javascript from 'highlight.js/lib/languages/javascript'
import python from 'highlight.js/lib/languages/python'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import sql from 'highlight.js/lib/languages/sql'
import dockerfile from 'highlight.js/lib/languages/dockerfile'
import nginx from 'highlight.js/lib/languages/nginx'
import go from 'highlight.js/lib/languages/go'
import rust from 'highlight.js/lib/languages/rust'
import java from 'highlight.js/lib/languages/java'
import cpp from 'highlight.js/lib/languages/cpp'
import typescript from 'highlight.js/lib/languages/typescript'

// 注册常用语言
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('dockerfile', dockerfile)
hljs.registerLanguage('nginx', nginx)
hljs.registerLanguage('go', go)
hljs.registerLanguage('rust', rust)
hljs.registerLanguage('java', java)
hljs.registerLanguage('cpp', cpp)
hljs.registerLanguage('typescript', typescript)

// 代码块检测正则
const CODE_BLOCK_PATTERNS = [
  // JSON
  { pattern: /^\s*[\{\[][\s\S]*[\}\]]\s*$/, lang: 'json' },
  // HTML/XML
  { pattern: /^\s*<[a-zA-Z][\s\S]*>/, lang: 'html' },
  // CSS
  { pattern: /^[a-zA-Z.#][\s\S]*\{[\s\S]*\}/, lang: 'css' },
  // SQL
  { pattern: /^\s*(SELECT|INSERT|UPDATE|DELETE|CREATE|ALTER|DROP)\s/i, lang: 'sql' },
  // Python
  { pattern: /^\s*(def |class |import |from |print\(|if __name__)/, lang: 'python' },
  // JavaScript/TypeScript
  { pattern: /^\s*(const |let |var |function |class |import |export |=>)/, lang: 'javascript' },
  // Bash/Shell
  { pattern: /^#!/, lang: 'bash' },
  // Go
  { pattern: /^\s*(package |func |import |type |struct |interface )/, lang: 'go' },
  // Dockerfile
  { pattern: /^\s*(FROM |RUN |CMD |EXPOSE |ENV |ADD |COPY |ENTRYPOINT )/, lang: 'dockerfile' },
]

/**
 * 检测代码语言
 */
export function detectLanguage(code) {
  if (!code || code.length < 10) return null

  for (const { pattern, lang } of CODE_BLOCK_PATTERNS) {
    if (pattern.test(code)) return lang
  }

  // 尝试自动检测
  try {
    const result = hljs.highlightAuto(code, [
      'javascript', 'python', 'bash', 'json', 'yaml', 'html', 'css', 'sql'
    ])
    if (result.relevance > 5) return result.language
  } catch (e) {
    // 忽略检测错误
  }

  return null
}

/**
 * 高亮代码
 */
export function highlightCode(code, language) {
  if (!code) return ''

  try {
    if (language && hljs.getLanguage(language)) {
      return hljs.highlight(code, { language }).value
    }
    // 自动检测
    return hljs.highlightAuto(code).value
  } catch (e) {
    // 高亮失败，返回纯文本
    return escapeHtml(code)
  }
}

/**
 * HTML 转义
 */
export function escapeHtml(str) {
  if (!str) return ''
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

/**
 * 高亮终端输出中的特殊部分
 * - 文件路径高亮
 * - 数字高亮
 * - 错误/成功标记高亮
 */
export function highlightTerminalOutput(text) {
  if (!text) return ''

  let html = escapeHtml(text)

  // 高亮文件路径
  html = html.replace(
    /(\/[\w\-./]+)/g,
    '<span class="hl-path">$1</span>'
  )

  // 高亮 IP 地址
  html = html.replace(
    /(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/g,
    '<span class="hl-ip">$1</span>'
  )

  // 高亮数字
  html = html.replace(
    /\b(\d+\.?\d*)\s*(bytes|KB|MB|GB|ms|s)\b/gi,
    '<span class="hl-number">$1</span> <span class="hl-unit">$2</span>'
  )

  // 高亮成功/失败标记
  html = html.replace(
    /\b(success|ok|done|complete|passed)\b/gi,
    '<span class="hl-success">$1</span>'
  )
  html = html.replace(
    /\b(error|fail|failed|denied|refused|timeout)\b/gi,
    '<span class="hl-error">$1</span>'
  )
  html = html.replace(
    /\b(warning|warn|caution)\b/gi,
    '<span class="hl-warning">$1</span>'
  )

  // 高亮权限标记
  html = html.replace(
    /\b([drwx-]{10})\b/g,
    '<span class="hl-perms">$1</span>'
  )

  return html
}
