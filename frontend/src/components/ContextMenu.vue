<template>
  <Teleport to="body">
    <div v-if="visible" class="ctx-mask" @mousedown.self="close" @contextmenu.prevent="close">
      <div
        ref="menuRef"
        class="ctx-menu"
        :class="{ 'ctx-submenu': isSubmenu }"
        :style="menuStyle"
      >
        <!-- 标题 -->
        <div v-if="header" class="ctx-header">{{ header }}</div>

        <template v-for="(item, idx) in items" :key="idx">
          <!-- 分隔线 -->
          <div v-if="item.divider" class="ctx-divider" />

          <!-- 子菜单 -->
          <div
            v-else-if="item.children?.length"
            class="ctx-item ctx-has-children"
            :class="{ disabled: item.disabled }"
            @mouseenter="openSubmenu(idx, $event)"
            @mouseleave="scheduleCloseSubmenu"
          >
            <span class="ctx-icon" v-if="item.icon" v-html="item.icon" />
            <span class="ctx-label">{{ item.label }}</span>
            <span class="ctx-arrow">
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </span>
            <span v-if="item.shortcut" class="ctx-shortcut">{{ item.shortcut }}</span>
          </div>

          <!-- 普通菜单项 -->
          <div
            v-else
            class="ctx-item"
            :class="{ danger: item.danger, disabled: item.disabled }"
            @click="handleClick(item)"
          >
            <span class="ctx-icon" v-if="item.icon" v-html="item.icon" />
            <span class="ctx-label">{{ item.label }}</span>
            <span v-if="item.shortcut" class="ctx-shortcut">{{ item.shortcut }}</span>
          </div>
        </template>
      </div>

      <!-- 子菜单 -->
      <ContextMenu
        v-if="activeSubmenu"
        :visible="true"
        :items="activeSubmenuItems"
        :position="submenuPos"
        :is-submenu="true"
        @select="(item) => { close(); $emit('select', item) }"
        @close="closeSubmenu"
      />
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false },
  header: { type: String, default: '' },
  items: { type: Array, default: () => [] },
  position: { type: Object, default: () => ({ x: 0, y: 0 }) },
  isSubmenu: { type: Boolean, default: false },
})

const emit = defineEmits(['select', 'close'])

const menuRef = ref(null)
const activeSubmenu = ref(null)
const submenuPos = ref({ x: 0, y: 0 })
let closeSubmenuTimer = null

const activeSubmenuItems = computed(() => {
  if (!activeSubmenu.value) return []
  return props.items[activeSubmenu.value]?.children || []
})

// 计算菜单位置（自动检测视口边界）
const menuStyle = computed(() => {
  const pos = { ...props.position }
  const menuW = 220
  const menuH = (props.header ? 1 : 0) * 36 + props.items.filter(i => !i.divider).length * 36
  const vw = window.innerWidth
  const vh = window.innerHeight

  if (pos.x + menuW > vw) pos.x = vw - menuW - 8
  if (pos.y + menuH > vh) pos.y = vh - menuH - 8
  if (pos.x < 8) pos.x = 8
  if (pos.y < 8) pos.y = 8

  return { top: pos.y + 'px', left: pos.x + 'px' }
})

function handleClick(item) {
  if (item.disabled) return
  emit('select', item)
  // 自动关闭（除非 item 指定 keepOpen）
  if (!item.keepOpen) {
    close()
  }
}

function openSubmenu(idx, event) {
  if (closeSubmenuTimer) {
    clearTimeout(closeSubmenuTimer)
    closeSubmenuTimer = null
  }
  activeSubmenu.value = idx
  const rect = event.currentTarget.getBoundingClientRect()
  submenuPos.value = { x: rect.right + 4, y: rect.top }
}

function scheduleCloseSubmenu() {
  closeSubmenuTimer = setTimeout(() => {
    activeSubmenu.value = null
  }, 200)
}

function closeSubmenu() {
  activeSubmenu.value = null
}

function close() {
  activeSubmenu.value = null
  emit('close')
}
</script>

<style>
.ctx-mask {
  position: fixed;
  inset: 0;
  z-index: 100000;
}

.ctx-menu {
  position: fixed;
  min-width: 180px;
  max-width: 320px;
  background: var(--bg-panel-solid, #1e1e2e);
  border: 1px solid var(--border-default, #333);
  border-radius: 8px;
  padding: 4px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  z-index: 100001;
  animation: ctxIn 0.12s ease-out;
  backdrop-filter: blur(12px);
}

@keyframes ctxIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

.ctx-header {
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted, #888);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-subtle, #333);
  margin-bottom: 4px;
}

.ctx-divider {
  height: 1px;
  background: var(--border-subtle, #333);
  margin: 4px 8px;
}

.ctx-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  color: var(--text-secondary, #ccc);
  font-size: 13px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.1s;
  user-select: none;
}

.ctx-item:hover {
  background: var(--primary-bg, rgba(66, 153, 225, 0.15));
  color: var(--text-primary, #fff);
}

.ctx-item.danger:hover {
  background: var(--danger-bg, rgba(239, 68, 68, 0.15));
  color: var(--accent-danger, #ef4444);
}

.ctx-item.disabled {
  opacity: 0.4;
  cursor: not-allowed;
  pointer-events: none;
}

.ctx-icon {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: inherit;
}

.ctx-icon svg {
  width: 14px;
  height: 14px;
}

.ctx-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ctx-shortcut {
  font-size: 11px;
  color: var(--text-muted, #666);
  margin-left: auto;
  font-family: monospace;
}

.ctx-arrow {
  margin-left: auto;
  color: var(--text-muted, #666);
}

.ctx-has-children {
  position: relative;
}
</style>
