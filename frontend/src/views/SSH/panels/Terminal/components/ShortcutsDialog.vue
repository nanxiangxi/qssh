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
  position: fixed; inset: 0; background: rgba(0,0,0,.6);
  display: flex; align-items: center; justify-content: center;
  z-index: 99999; backdrop-filter: blur(2px);
}
.dialog {
  width: 650px; max-width: 90vw; max-height: 80vh;
  background: #1a1a1a; border: 1px solid #333; border-radius: 12px;
  box-shadow: 0 16px 48px rgba(0,0,0,.5);
  display: flex; flex-direction: column; overflow: hidden;
}
.dialog-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 18px; border-bottom: 1px solid #2a2a2a;
}
.dialog-header h3 { margin: 0; font-size: 15px; color: #d4d4d4; }
.close-btn {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; background: transparent; border: none;
  border-radius: 6px; color: #666; cursor: pointer;
}
.close-btn:hover { background: #333; color: #aaa; }
.dialog-body { flex: 1; overflow-y: auto; padding: 14px 18px; }
.section { margin-bottom: 18px; }
.section-title {
  margin: 0 0 8px; font-size: 11px; font-weight: 600;
  color: #666; text-transform: uppercase; letter-spacing: 0.5px;
}
.shortcut-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 4px; }
.shortcut-item {
  display: flex; align-items: center; gap: 8px;
  padding: 5px 8px; border-radius: 4px;
}
.shortcut-item:hover { background: #222; }
kbd {
  display: inline-flex; align-items: center; justify-content: center;
  min-width: 55px; height: 22px; padding: 0 6px;
  background: #2a2a2a; border: 1px solid #444; border-radius: 4px;
  font-size: 10px; color: #ccc; font-family: 'Cascadia Code', monospace;
  white-space: nowrap;
}
.desc { font-size: 11px; color: #888; }
.dialog-body::-webkit-scrollbar { width: 4px; }
.dialog-body::-webkit-scrollbar-track { background: transparent; }
.dialog-body::-webkit-scrollbar-thumb { background: #444; border-radius: 2px; }
.fade-enter-active { transition: all .2s ease; }
.fade-leave-active { transition: all .15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
