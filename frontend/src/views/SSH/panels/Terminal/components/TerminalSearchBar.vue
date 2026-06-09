<template>
  <Transition name="slide">
    <div v-if="visible" class="search-bar">
      <div class="search-input-wrapper">
        <svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input
          ref="inputRef"
          v-model="keyword"
          class="search-input"
          placeholder="搜索终端内容..."
          @keydown.enter="searchNext"
          @keydown.escape="close"
          @keydown.enter.shift="searchPrevious"
        />
        <div class="search-actions">
          <button class="search-btn" @click="searchPrevious" title="上一个 (Shift+Enter)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="18 15 12 9 6 15"/>
            </svg>
          </button>
          <button class="search-btn" @click="searchNext" title="下一个 (Enter)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </button>
          <button class="search-btn close" @click="close" title="关闭 (Escape)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false }
})

const emit = defineEmits(['update:visible', 'search', 'close'])

const inputRef = ref(null)
const keyword = ref('')

// 自动聚焦
watch(() => props.visible, (visible) => {
  if (visible) {
    nextTick(() => {
      inputRef.value?.focus()
    })
  }
})

function searchNext() {
  if (keyword.value) {
    emit('search', keyword.value)
  }
}

function searchPrevious() {
  // TODO: 实现向上搜索
  if (keyword.value) {
    emit('search', keyword.value)
  }
}

function close() {
  keyword.value = ''
  emit('update:visible', false)
  emit('close')
}
</script>

<style scoped>
.search-bar {
  position: absolute;
  top: 40px;
  right: 16px;
  z-index: 100;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  background: rgba(26, 27, 38, 0.98);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 4px 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(12px);
}

.search-icon {
  color: #565f89;
  margin-right: 8px;
  flex-shrink: 0;
}

.search-input {
  width: 240px;
  background: transparent;
  border: none;
  outline: none;
  color: #c0caf5;
  font-size: 13px;
  font-family: inherit;
}

.search-input::placeholder {
  color: #565f89;
}

.search-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-left: 8px;
}

.search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: #a9b1d6;
  cursor: pointer;
  transition: all 0.15s;
}

.search-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #c0caf5;
}

.search-btn.close:hover {
  background: rgba(247, 118, 142, 0.15);
  color: #f7768e;
}

/* 动画 */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.2s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
