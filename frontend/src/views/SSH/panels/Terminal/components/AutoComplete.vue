<template>
  <Transition name="popup">
    <div v-if="visible && suggestions.length > 0" class="autocomplete-popup" :style="popupStyle">
      <div class="suggestions-list" ref="listRef">
        <div
          v-for="(suggestion, index) in suggestions"
          :key="index"
          class="suggestion-item"
          :class="{ selected: index === selectedIndex }"
          @click="applySuggestion(index)"
          @mouseenter="selectedIndex = index"
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
          <span class="suggestion-desc">{{ suggestion.description }}</span>
        </div>
      </div>
      <div class="popup-footer">
        <span class="hint">Tab 补全 · Esc 关闭</span>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, watch, nextTick, computed } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false },
  suggestions: { type: Array, default: () => [] },
  position: { type: Object, default: () => ({ top: 0, left: 0 }) }
})

const emit = defineEmits(['apply', 'close'])

const listRef = ref(null)
const selectedIndex = ref(0)

const popupStyle = computed(() => ({
  top: props.position.top + 'px',
  left: props.position.left + 'px'
}))

// 重置选择
watch(() => props.suggestions, () => {
  selectedIndex.value = 0
})

// 滚动到选中项
watch(selectedIndex, (index) => {
  if (index >= 0 && listRef.value) {
    nextTick(() => {
      const item = listRef.value.children[index]
      if (item) item.scrollIntoView({ block: 'nearest' })
    })
  }
})

function applySuggestion(index) {
  const suggestion = props.suggestions[index]
  if (suggestion) emit('apply', suggestion)
}

// 键盘导航
function handleKey(key) {
  if (key === 'ArrowUp') {
    selectedIndex.value = Math.max(0, selectedIndex.value - 1)
    return true
  }
  if (key === 'ArrowDown') {
    selectedIndex.value = Math.min(props.suggestions.length - 1, selectedIndex.value + 1)
    return true
  }
  if (key === 'Tab' || key === 'Enter') {
    applySuggestion(selectedIndex.value)
    return true
  }
  if (key === 'Escape') {
    emit('close')
    return true
  }
  return false
}

defineExpose({ handleKey })
</script>

<style scoped>
.autocomplete-popup {
  position: absolute;
  bottom: 100%;
  left: 40px;
  margin-bottom: 4px;
  background: var(--bg-toolbar);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  box-shadow: 0 4px 16px var(--shadow-lg);
  z-index: 100;
  min-width: 280px;
  max-width: 450px;
  max-height: 300px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.suggestions-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}

.suggestion-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background .1s;
}

.suggestion-item:hover,
.suggestion-item.selected {
  background: var(--surface-hover);
}

.suggestion-icon {
  flex-shrink: 0;
  color: var(--text-muted);
}

.suggestion-icon.command { color: var(--accent-success); }
.suggestion-icon.subcommand { color: var(--text-muted); }
.suggestion-icon.option { color: var(--accent-warning); }

.suggestion-value {
  font-family: 'Cascadia Code', monospace;
  font-size: 13px;
  color: var(--text-primary);
  white-space: nowrap;
}

.suggestion-desc {
  font-size: 11px;
  color: var(--text-muted);
  margin-left: auto;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 150px;
}

.popup-footer {
  padding: 4px 8px;
  border-top: 1px solid var(--surface-hover);
}

.hint {
  font-size: 10px;
  color: var(--text-muted);
}

.popup-enter-active { transition: all .15s ease; }
.popup-leave-active { transition: all .1s ease; }
.popup-enter-from, .popup-leave-to { opacity: 0; transform: translateY(4px); }

.suggestions-list::-webkit-scrollbar { width: 4px; }
.suggestions-list::-webkit-scrollbar-track { background: transparent; }
.suggestions-list::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 2px; }
</style>
