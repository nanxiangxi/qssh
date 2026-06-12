<template>
  <Transition name="fade">
    <div v-if="visible" class="mask" @mousedown.self="$emit('close')">
      <div class="dialog">
        <div class="dialog-header">
          <h3>快捷键参考</h3>
          <button class="close-btn" @click="$emit('close')">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="dialog-body">
          <div class="section" v-for="s in sections" :key="s.title">
            <h4 class="section-title">{{ s.title }}</h4>
            <div class="shortcut-grid">
              <div class="shortcut-item" v-for="item in s.items" :key="item.key">
                <kbd>{{ item.key }}</kbd>
                <span class="desc">{{ item.desc }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { getKeybindingCategories } from '../utils/keybindings'

defineProps({ visible: { type: Boolean, default: false } })
defineEmits(['close'])

const sections = getKeybindingCategories()
</script>

<style scoped>
.mask {
  position: fixed; inset: 0; background: var(--bg-overlay);
  display: flex; align-items: center; justify-content: center;
  z-index: 99999; backdrop-filter: blur(2px);
}
.dialog {
  width: 650px; max-width: 90vw; max-height: 80vh;
  background: var(--bg-panel); border: 1px solid var(--border-default); border-radius: 12px;
  box-shadow: var(--shadow-lg);
  display: flex; flex-direction: column; overflow: hidden;
}
.dialog-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 18px; border-bottom: 1px solid var(--border-default);
}
.dialog-header h3 { margin: 0; font-size: 15px; color: var(--text-primary); }
.close-btn {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; background: transparent; border: none;
  border-radius: 6px; color: var(--text-muted); cursor: pointer;
}
.close-btn:hover { background: var(--surface-hover); color: var(--text-secondary); }
.dialog-body { flex: 1; overflow-y: auto; padding: 14px 18px; }
.section { margin-bottom: 18px; }
.section-title {
  margin: 0 0 8px; font-size: 11px; font-weight: 600;
  color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px;
}
.shortcut-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 4px; }
.shortcut-item {
  display: flex; align-items: center; gap: 8px;
  padding: 5px 8px; border-radius: 4px;
}
.shortcut-item:hover { background: var(--surface-hover); }
kbd {
  display: inline-flex; align-items: center; justify-content: center;
  min-width: 55px; height: 22px; padding: 0 6px;
  background: var(--surface-2); border: 1px solid var(--border-strong); border-radius: 4px;
  font-size: 10px; color: var(--text-secondary); font-family: 'Cascadia Code', monospace;
  white-space: nowrap;
}
.desc { font-size: 11px; color: var(--text-secondary); }
.dialog-body::-webkit-scrollbar { width: 4px; }
.dialog-body::-webkit-scrollbar-track { background: transparent; }
.dialog-body::-webkit-scrollbar-thumb { background: var(--scrollbar-thumb); border-radius: 2px; }
.fade-enter-active { transition: all .2s ease; }
.fade-leave-active { transition: all .15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
