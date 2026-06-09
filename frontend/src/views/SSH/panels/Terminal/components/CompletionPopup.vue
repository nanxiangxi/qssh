<template>
  <div v-if="visible && suggestions.length > 0" class="completion-popup" :style="popupStyle">
    <div class="suggestions-list" ref="listRef">
      <div
        v-for="(suggestion, index) in suggestions"
        :key="index"
        class="suggestion-item"
        :class="{ selected: index === selectedIndex }"
        @click="$emit('apply', index)"
      >
        <span class="suggestion-icon" :class="suggestion.type">
          <svg v-if="suggestion.type === 'command'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="4 17 10 11 4 5"/>
            <line x1="12" y1="19" x2="20" y2="19"/>
          </svg>
          <svg v-else-if="suggestion.type === 'subcommand'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="9 18 15 12 9 6"/>
          </svg>
          <svg v-else width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="4" y1="21" x2="4" y2="14"/>
            <line x1="4" y1="10" x2="4" y2="3"/>
            <line x1="12" y1="21" x2="12" y2="12"/>
            <line x1="12" y1="8" x2="12" y2="3"/>
            <line x1="20" y1="21" x2="20" y2="16"/>
            <line x1="20" y1="12" x2="20" y2="3"/>
          </svg>
        </span>
        <span class="suggestion-value">{{ suggestion.value }}</span>
        <span v-if="suggestion.description" class="suggestion-desc">{{ suggestion.description }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false },
  suggestions: { type: Array, default: () => [] },
  position: { type: Object, default: () => ({ top: 0, left: 0 }) }
})

const emit = defineEmits(['apply', 'close'])

const listRef = ref(null)
const popupRef = ref(null)
const selectedIndex = ref(0)

// 计算弹窗位置，自动避让边缘
const popupStyle = computed(() => {
  let top = props.position.top
  let left = props.position.left

  // 如果靠近底部，改为向上弹出
  if (top > window.innerHeight - 250) {
    top = props.position.top - 220
  }
  // 如果靠近右侧，向左偏移
  if (left > window.innerWidth - 400) {
    left = window.innerWidth - 400
  }
  // 边界保护
  if (top < 0) top = 4
  if (left < 0) left = 4

  return { top: top + 'px', left: left + 'px', position: 'fixed' }
})

watch(() => props.suggestions, () => {
  selectedIndex.value = 0
})

watch(selectedIndex, (i) => {
  if (i >= 0 && listRef.value) {
    const item = listRef.value.children[i]
    if (item) item.scrollIntoView({ block: 'nearest' })
  }
})

function handleKey(key) {
  if (key === 'ArrowUp') {
    selectedIndex.value = Math.max(0, selectedIndex.value - 1)
    return true
  }
  if (key === 'ArrowDown') {
    selectedIndex.value = Math.min(props.suggestions.length - 1, selectedIndex.value + 1)
    return true
  }
  if (key === 'Tab') {
    emit('apply', selectedIndex.value)
    return true
  }
  if (key === 'Enter') {
    emit('close')
    return false  // 不拦截，让 Enter 正常执行命令
  }
  if (key === 'Escape') {
    emit('close')
    return true
  }
  return false
}

defineExpose({ handleKey, selectedIndex })
</script>

<style scoped>
.completion-popup {
  position: fixed;
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0,0,0,.5);
  max-height: 200px;
  overflow-y: auto;
  padding: 4px;
  min-width: 250px;
  z-index: 10000;
}

.suggestion-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
}

.suggestion-item:hover,
.suggestion-item.selected {
  background: #333;
}

.suggestion-icon {
  flex-shrink: 0;
  color: #666;
}

.suggestion-icon.command { color: #4caf50; }
.suggestion-icon.subcommand { color: #888; }
.suggestion-icon.option { color: #ff9800; }

.suggestion-value {
  font-family: 'Cascadia Code', monospace;
  font-size: 13px;
  color: #d4d4d4;
  white-space: nowrap;
}

.suggestion-desc {
  font-size: 11px;
  color: #666;
  margin-left: auto;
  white-space: nowrap;
}

::-webkit-scrollbar { width: 4px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #444; border-radius: 2px; }
</style>
