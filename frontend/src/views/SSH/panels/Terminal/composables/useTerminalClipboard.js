/**
 * 终端复制粘贴 composable
 * 统一经典模式（xterm）和结构化模式（浏览器原生）的复制粘贴逻辑
 * 解决 xterm 右键清除选区的问题
 */
import { ref, computed } from 'vue'

export function useTerminalClipboard() {
  const savedSelection = ref('')

  // 在 mouseup 时保存选区（xterm 右键 mousedown 会清除选区，mouseup 更早）
  // getXterm 是一个返回 xterm 实例的函数（因为 xterm 是 let 变量，不是 ref）
  function trackSelection(containerEl, getXterm) {
    if (!containerEl) return
    containerEl.addEventListener('mouseup', () => {
      const xt = typeof getXterm === 'function' ? getXterm() : getXterm?.value
      if (xt && xt.hasSelection()) {
        savedSelection.value = xt.getSelection()
      }
    })
  }

  // 是否有选中文本
  const hasSelection = computed(() => {
    if (savedSelection.value) return true
    const sel = window.getSelection()
    return sel ? sel.toString().length > 0 : false
  })

  // 复制（统一处理）
  function copy(xtermRef, viewRef) {
    let text = savedSelection.value
    // 经典模式：尝试 xterm 选区
    if (!text && viewRef?.value === 'classic' && xtermRef?.value) {
      text = xtermRef.value.hasSelection() ? xtermRef.value.getSelection() : ''
    }
    // 结构化模式：浏览器原生选区
    if (!text) {
      const sel = window.getSelection()
      text = sel ? sel.toString() : ''
    }
    if (text) {
      doCopy(text)
    }
    savedSelection.value = ''
    return !!text
  }

  // 执行复制（Clipboard API + fallback）
  function doCopy(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).catch(() => fallbackCopy(text))
    } else {
      fallbackCopy(text)
    }
  }

  function fallbackCopy(text) {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.cssText = 'position:fixed;left:-9999px;opacity:0'
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
  }

  // 粘贴
  async function paste(xtermRef, viewRef, sendFn, cmdRef) {
    try {
      const text = await navigator.clipboard.readText()
      if (!text) return
      if (viewRef?.value === 'classic') {
        sendFn(text)
      } else {
        cmdRef.value += text
      }
    } catch {}
  }

  return { savedSelection, hasSelection, trackSelection, copy, paste }
}
