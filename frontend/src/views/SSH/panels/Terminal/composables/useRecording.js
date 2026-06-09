/**
 * 终端录制组合式函数
 * 支持录制终端会话、导出为文本或 HTML
 */
import { ref, computed } from 'vue'

export function useRecording(sessionId) {
  // 状态
  const isRecording = ref(false)
  const isPaused = ref(false)
  const startTime = ref(null)
  const duration = ref(0)

  // 录制数据
  const entries = ref([])
  let durationTimer = null

  // 计算属性
  const entryCount = computed(() => entries.value.length)
  const formattedDuration = computed(() => {
    const seconds = Math.floor(duration.value / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)

    if (hours > 0) {
      return `${hours}:${String(minutes % 60).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}`
    }
    return `${minutes}:${String(seconds % 60).padStart(2, '0')}`
  })

  // 开始录制
  function startRecording() {
    if (isRecording.value) return

    entries.value = []
    startTime.value = Date.now()
    duration.value = 0
    isRecording.value = true
    isPaused.value = false

    // 计时器
    durationTimer = setInterval(() => {
      if (!isPaused.value) {
        duration.value = Date.now() - startTime.value
      }
    }, 100)

    console.log('[useRecording] 开始录制:', sessionId)
  }

  // 停止录制
  function stopRecording() {
    if (!isRecording.value) return

    isRecording.value = false
    isPaused.value = false

    if (durationTimer) {
      clearInterval(durationTimer)
      durationTimer = null
    }

    console.log('[useRecording] 停止录制:', sessionId, '条目数:', entries.value.length)
    return entries.value
  }

  // 暂停录制
  function pauseRecording() {
    if (!isRecording.value || isPaused.value) return
    isPaused.value = true
  }

  // 恢复录制
  function resumeRecording() {
    if (!isRecording.value || !isPaused.value) return
    isPaused.value = false
  }

  // 记录输入
  function recordInput(data) {
    if (!isRecording.value || isPaused.value) return

    entries.value.push({
      type: 'input',
      data,
      timestamp: Date.now() - startTime.value
    })
  }

  // 记录输出
  function recordOutput(data) {
    if (!isRecording.value || isPaused.value) return

    entries.value.push({
      type: 'output',
      data,
      timestamp: Date.now() - startTime.value
    })
  }

  // 记录命令
  function recordCommand(command) {
    if (!isRecording.value || isPaused.value) return

    entries.value.push({
      type: 'command',
      data: command,
      timestamp: Date.now() - startTime.value
    })
  }

  // 导出为纯文本
  function exportAsText() {
    const lines = []
    lines.push(`Session Recording: ${sessionId}`)
    lines.push(`Start Time: ${new Date(startTime.value).toLocaleString()}`)
    lines.push(`Duration: ${formattedDuration.value}`)
    lines.push(`Entries: ${entries.value.length}`)
    lines.push('---')

    entries.value.forEach(entry => {
      const time = formatTimestamp(entry.timestamp)
      if (entry.type === 'command') {
        lines.push(`[${time}] $ ${entry.data}`)
      } else if (entry.type === 'input') {
        lines.push(`[${time}] > ${entry.data}`)
      } else {
        lines.push(`[${time}] ${entry.data}`)
      }
    })

    return lines.join('\n')
  }

  // 导出为 HTML
  function exportAsHTML() {
    const entriesHTML = entries.value.map(entry => {
      const time = formatTimestamp(entry.timestamp)
      const escapedData = escapeHTML(entry.data)

      if (entry.type === 'command') {
        return `<div class="entry command"><span class="time">[${time}]</span> <span class="prompt">$</span> ${escapedData}</div>`
      } else if (entry.type === 'input') {
        return `<div class="entry input"><span class="time">[${time}]</span> <span class="arrow">></span> ${escapedData}</div>`
      } else {
        return `<div class="entry output"><span class="time">[${time}]</span> <pre>${escapedData}</pre></div>`
      }
    }).join('\n')

    return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Terminal Recording - ${sessionId}</title>
  <style>
    body { background: #1a1b26; color: #c0caf5; font-family: 'Cascadia Code', monospace; padding: 20px; }
    .header { margin-bottom: 20px; padding-bottom: 10px; border-bottom: 1px solid #33467c; }
    .entry { margin: 4px 0; line-height: 1.5; }
    .time { color: #565f89; font-size: 12px; }
    .command { color: #9ece6a; }
    .input { color: #7aa2f7; }
    .output { color: #a9b1d6; }
    .prompt { color: #f7768e; }
    .arrow { color: #7dcfff; }
    pre { margin: 0; white-space: pre-wrap; }
  </style>
</head>
<body>
  <div class="header">
    <h1>Terminal Recording</h1>
    <p>Session: ${sessionId}</p>
    <p>Start: ${new Date(startTime.value).toLocaleString()}</p>
    <p>Duration: ${formattedDuration.value}</p>
    <p>Entries: ${entries.value.length}</p>
  </div>
  <div class="content">
    ${entriesHTML}
  </div>
</body>
</html>`
  }

  // 导出为 JSON
  function exportAsJSON() {
    return JSON.stringify({
      sessionId,
      startTime: startTime.value,
      duration: duration.value,
      entries: entries.value
    }, null, 2)
  }

  // 保存录制文件
  function saveRecording(format = 'text') {
    let content, filename, mimeType

    switch (format) {
      case 'html':
        content = exportAsHTML()
        filename = `terminal-recording-${sessionId}-${Date.now()}.html`
        mimeType = 'text/html'
        break
      case 'json':
        content = exportAsJSON()
        filename = `terminal-recording-${sessionId}-${Date.now()}.json`
        mimeType = 'application/json'
        break
      default:
        content = exportAsText()
        filename = `terminal-recording-${sessionId}-${Date.now()}.txt`
        mimeType = 'text/plain'
    }

    // 创建下载链接
    const blob = new Blob([content], { type: mimeType })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)

    console.log('[useRecording] 保存录制:', filename)
  }

  // 清除录制数据
  function clearRecording() {
    entries.value = []
    startTime.value = null
    duration.value = 0
  }

  // 工具函数
  function formatTimestamp(ms) {
    const seconds = Math.floor(ms / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)

    return `${hours}:${String(minutes % 60).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}.${String(ms % 1000).padStart(3, '0')}`
  }

  function escapeHTML(str) {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;')
  }

  return {
    // 状态
    isRecording,
    isPaused,
    startTime,
    duration,
    entryCount,
    formattedDuration,

    // 方法
    startRecording,
    stopRecording,
    pauseRecording,
    resumeRecording,
    recordInput,
    recordOutput,
    recordCommand,
    exportAsText,
    exportAsHTML,
    exportAsJSON,
    saveRecording,
    clearRecording
  }
}
