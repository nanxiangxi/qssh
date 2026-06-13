<template>
  <Teleport to="body">
    <div v-if="visible" class="context-menu-mask" @mousedown="close">
      <div
        class="context-menu"
        :style="{ top: position.y + 'px', left: position.x + 'px' }"
        @mousedown.stop
      >
        <!-- 复制 -->
        <div class="menu-item" :class="{ disabled: !hasSelection }" @click="$emit('copy')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="9" y="9" width="13" height="13" rx="2"/>
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
          </svg>
          <span>复制</span>
          <span class="shortcut">Ctrl+Shift+C</span>
        </div>

        <!-- 粘贴 -->
        <div class="menu-item" @click="$emit('paste')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/>
            <rect x="8" y="2" width="8" height="4" rx="1"/>
          </svg>
          <span>粘贴</span>
          <span class="shortcut">Ctrl+Shift+V</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 全选 -->
        <div class="menu-item" @click="$emit('select-all')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 11V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h6"/>
            <path d="M12 12H3M12 12V3"/>
          </svg>
          <span>全选</span>
          <span class="shortcut">Ctrl+Shift+A</span>
        </div>

        <!-- 清屏 -->
        <div class="menu-item" @click="$emit('clear')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
          <span>清屏</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 搜索 -->
        <div class="menu-item" @click="$emit('search')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <span>搜索</span>
          <span class="shortcut">Ctrl+Shift+F</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 新建终端 -->
        <div class="menu-item" @click="$emit('new-terminal')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          <span>新建终端</span>
        </div>

        <!-- 水平分屏 -->
        <div class="menu-item" @click="$emit('split-horizontal')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <line x1="3" y1="12" x2="21" y2="12"/>
          </svg>
          <span>水平分屏</span>
        </div>

        <!-- 垂直分屏 -->
        <div class="menu-item" @click="$emit('split-vertical')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <line x1="12" y1="3" x2="12" y2="21"/>
          </svg>
          <span>垂直分屏</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
const props = defineProps({
  visible: { type: Boolean, default: false },
  position: { type: Object, default: () => ({ x: 0, y: 0 }) },
  hasSelection: { type: Boolean, default: false }
})

const emit = defineEmits([
  'update:visible',
  'copy',
  'paste',
  'select-all',
  'clear',
  'search',
  'new-terminal',
  'split-horizontal',
  'split-vertical'
])

function close() {
  emit('update:visible', false)
}
</script>

<style scoped>
.context-menu-mask {
  position: fixed;
  inset: 0;
  z-index: 10000;
}

.context-menu {
  position: fixed;
  background: var(--bg-panel);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 4px;
  min-width: 200px;
  box-shadow: var(--shadow-md);
  z-index: 10001;
  animation: contextMenuIn 0.1s ease-out;
}

@keyframes contextMenuIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  color: var(--text-secondary);
  font-size: 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.1s;
}

.menu-item:hover {
  background: var(--primary-bg);
  color: var(--text-primary);
}

.menu-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.menu-item.disabled:hover {
  background: transparent;
  color: var(--text-secondary);
}

.shortcut {
  margin-left: auto;
  color: var(--text-muted);
  font-size: 10px;
  font-family: monospace;
}

.menu-separator {
  height: 1px;
  background: var(--border-subtle);
  margin: 4px 0;
}
</style>
