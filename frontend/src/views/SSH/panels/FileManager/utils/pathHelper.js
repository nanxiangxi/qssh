/**
 * 路径处理工具函数
 */

/**
 * 规范化路径（避免双重斜杠等问题）
 * @param {string} path - 原始路径
 * @returns {string} 规范化后的路径
 */
export const normalizePath = (path) => {
  if (!path) return '/'
  // 替换多个连续斜杠为单个斜杠
  let normalized = path.replace(/\/+/g, '/')
  // 如果不是根目录且以斜杠结尾，去掉结尾斜杠
  if (normalized !== '/' && normalized.endsWith('/')) {
    normalized = normalized.slice(0, -1)
  }
  return normalized
}

/**
 * 拼接路径
 * @param {string} basePath - 基础路径
 * @param {string} fileName - 文件名
 * @returns {string} 拼接后的路径
 */
export const joinPath = (basePath, fileName) => {
  return normalizePath(`${basePath}/${fileName}`)
}

/**
 * 获取父目录路径
 * @param {string} path - 当前路径
 * @returns {string} 父目录路径
 */
export const getParentPath = (path) => {
  if (path === '/') return '/'
  const parts = path.split('/').filter(p => p)
  parts.pop()
  return parts.length === 0 ? '/' : '/' + parts.join('/')
}

/**
 * 格式化文件大小
 * @param {number} bytes - 字节数
 * @returns {string} 格式化后的大小
 */
export const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

/**
 * 格式化时间
 * @param {string|number} timestamp - 时间戳
 * @returns {string} 格式化后的时间
 */
export const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN')
}

/**
 * 将权限字符串转换为数字格式（如 drwxr-xr-x -> 755）
 * @param {string} modeStr - 权限字符串
 * @returns {string} 数字权限
 */
export const getNumericMode = (modeStr) => {
  if (!modeStr || modeStr.length < 10) return '---'
  
  const perms = modeStr.slice(1) // 去掉第一个字符（文件类型）
  let result = ''
  
  for (let i = 0; i < 9; i += 3) {
    let val = 0
    if (perms[i] !== '-') val += 4     // r
    if (perms[i + 1] !== '-') val += 2 // w
    if (perms[i + 2] !== '-') val += 1 // x
    result += val
  }
  
  return result
}
