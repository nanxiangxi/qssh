<template>
  <Transition name="fade">
    <div v-if="visible" class="mask" @mousedown.self="$emit('cancel')">
      <div class="dialog">
        <div class="dialog-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#ff9800" stroke-width="2">
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
  background: rgba(0,0,0,.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 99999;
  backdrop-filter: blur(2px);
}

.dialog {
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 12px;
  padding: 24px;
  max-width: 380px;
  width: 90%;
  box-shadow: 0 16px 48px rgba(0,0,0,.5);
  text-align: center;
}

.dialog-icon {
  margin-bottom: 16px;
}

.dialog-title {
  margin: 0 0 8px;
  font-size: 16px;
  color: #d4d4d4;
}

.dialog-desc {
  margin: 0 0 24px;
  font-size: 13px;
  color: #888;
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
  background: #333;
  color: #aaa;
}

.btn-cancel:hover {
  background: #444;
  color: #d4d4d4;
}

.btn-confirm {
  background: rgba(76,175,80,.2);
  color: #4caf50;
  border: 1px solid rgba(76,175,80,.3);
}

.btn-confirm:hover {
  background: rgba(76,175,80,.3);
}

.fade-enter-active { transition: all .2s ease; }
.fade-leave-active { transition: all .15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
