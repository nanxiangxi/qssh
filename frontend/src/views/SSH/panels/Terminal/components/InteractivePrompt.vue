<template>
  <Transition name="fade">
    <div v-if="visible" class="mini-terminal-mask" @mousedown.self="$emit('close')">
      <div class="mini-terminal">
        <div class="mini-header">
          <div class="mini-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="4 17 10 11 4 5"/>
              <line x1="12" y1="19" x2="20" y2="19"/>
            </svg>
            <span>{{ title }}</span>
          </div>
          <button class="mini-btn" @click="$emit('close')" title="关闭 (Esc)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <!-- xterm 容器（接收移入的 xterm DOM） -->
        <div class="mini-body" ref="miniBodyRef"></div>
        <div class="mini-status">
          <span>Esc 关闭 · 同一会话 · 输入直接发送</span>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false },
  xtermElement: { type: Object, default: null },  // 主终端的 xterm DOM 元素
  initialKey: { type: String, default: '' },
  onFit: { type: Function, default: null }  // 调整尺寸的回调
})

const emit = defineEmits(['close'])

const miniBodyRef = ref(null)

const title = computed(() => {
  if (props.initialKey === '\x12') return 'Ctrl+R 反向搜索'
  if (props.initialKey === '\x1a') return 'Ctrl+Z 暂停'
  return '临时终端'
})

// 监听可见性变化，移动 xterm DOM
watch(() => props.visible, (v) => {
  if (v) {
    nextTick(() => moveXterm(true))
  } else {
    moveXterm(false)
  }
})

function moveXterm(toPopup) {
  const el = props.xtermElement
  if (!el) return

  if (toPopup && miniBodyRef.value) {
    // 移动到弹窗
    miniBodyRef.value.appendChild(el)
    console.log('[MiniTerminal] xterm 已移动到弹窗')
    // 强制重新调整尺寸
    nextTick(() => {
      // 多次调整确保正确
      if (props.onFit) props.onFit()
      setTimeout(() => {
        if (props.onFit) props.onFit()
        // 聚焦输入
        const textarea = el.querySelector('.xterm-helper-textarea')
        if (textarea) textarea.focus()
      }, 100)
    })
  }
  // 移回操作由父组件处理
}
</script>

<style scoped>
.mini-terminal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100000;
  backdrop-filter: blur(2px);
}

.mini-terminal {
  width: 700px;
  max-width: 90vw;
  height: 450px;
  max-height: 70vh;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 12px;
  box-shadow: 0 16px 48px rgba(0,0,0,.6);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.mini-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: #252525;
  border-bottom: 1px solid #333;
  flex-shrink: 0;
}

.mini-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #aaa;
  font-size: 13px;
}

.mini-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: #666;
  cursor: pointer;
}

.mini-btn:hover {
  background: #333;
  color: #aaa;
}

.mini-body {
  flex: 1;
  min-height: 0;
  padding: 0;
  overflow: hidden;
  position: relative;
}

.mini-body :deep(.xterm) {
  position: absolute;
  inset: 0;
  width: 100% !important;
  height: 100% !important;
}

.mini-body :deep(.xterm-viewport) {
  overflow-y: auto !important;
}

.mini-body :deep(.xterm-screen) {
  width: 100% !important;
}

.mini-status {
  padding: 6px 14px;
  background: #1e1e1e;
  border-top: 1px solid #2a2a2a;
  flex-shrink: 0;
  font-size: 10px;
  color: #555;
}

.fade-enter-active { transition: all .2s ease; }
.fade-leave-active { transition: all .15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
