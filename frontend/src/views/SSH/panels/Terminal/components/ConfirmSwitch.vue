<template>
  <Transition name="fade">
    <div v-if="visible" class="mask" @mousedown.self="$emit('cancel')">
      <div class="dialog">
        <div class="dialog-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--accent-warning)" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <div class="dialog-body">
          <h3 class="dialog-title">{{ title }}</h3>
          <p class="dialog-desc">{{ description }}</p>
        </div>
        <div class="dialog-actions">
          <button class="btn btn-cancel" @click="$emit('cancel')">取消</button>
          <button class="btn btn-confirm" @click="$emit('confirm')">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="9 18 15 12 9 6"/>
            </svg>
            切换并执行
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
defineProps({
  visible: { type: Boolean, default: false },
  title: { type: String, default: '需要切换终端模式' },
  description: { type: String, default: '' }
})

defineEmits(['confirm', 'cancel'])
</script>

<style scoped>
.mask {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 99999;
  backdrop-filter: blur(2px);
}

.dialog {
  background: var(--bg-panel);
  border: 1px solid var(--border-strong);
  border-radius: 12px;
  padding: 24px;
  max-width: 380px;
  width: 90%;
  box-shadow: var(--shadow-lg);
  text-align: center;
}

.dialog-icon {
  margin-bottom: 16px;
}

.dialog-title {
  margin: 0 0 8px;
  font-size: 16px;
  color: var(--text-primary);
}

.dialog-desc {
  margin: 0 0 24px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.dialog-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 20px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all .15s;
}

.btn-cancel {
  background: var(--surface-2);
  color: var(--text-secondary);
}

.btn-cancel:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.btn-confirm {
  background: var(--success-bg);
  color: var(--accent-success);
  border: 1px solid var(--success-bg);
}

.btn-confirm:hover {
  background: color-mix(in srgb, var(--accent-success) 30%, transparent);
}

.fade-enter-active { transition: all .2s ease; }
.fade-leave-active { transition: all .15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
