<template>
  <Teleport to="body">
    <Transition name="cd-fade">
      <div v-if="confirmState.show" class="cd-mask" @click.self="handleCancel">
        <div class="cd-modal" :class="{ 'cd-danger': confirmState.danger }">
          <div class="cd-icon">
            <svg v-if="confirmState.danger" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fc8181" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            <svg v-else width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#63b3ed" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          </div>
          <h3 class="cd-title">{{ confirmState.title }}</h3>
          <p class="cd-message">{{ confirmState.message }}</p>
          <div class="cd-actions">
            <button class="cd-btn cd-btn-cancel" @click="handleCancel">{{ confirmState.cancelText }}</button>
            <button class="cd-btn" :class="confirmState.danger ? 'cd-btn-danger' : 'cd-btn-primary'" @click="handleConfirm">
              {{ confirmState.confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { useConfirm } from '../utils/confirm'
const { confirmState, handleConfirm, handleCancel } = useConfirm()
</script>

<style scoped>
.cd-mask {
  position: fixed; inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex; align-items: center; justify-content: center;
  z-index: 20000;
  backdrop-filter: blur(2px);
}

.cd-modal {
  background: #1e1e1e;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
  width: 360px;
  max-width: 90vw;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
  text-align: center;
}

.cd-modal.cd-danger {
  border-color: rgba(245, 101, 101, 0.3);
}

.cd-icon {
  margin-bottom: 0.75rem;
}

.cd-title {
  margin: 0 0 0.5rem;
  color: #e2e8f0;
  font-size: 1rem;
  font-weight: 600;
}

.cd-message {
  margin: 0 0 1.25rem;
  color: #a0aec0;
  font-size: 0.8125rem;
  line-height: 1.5;
  white-space: pre-line;
}

.cd-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
}

.cd-btn {
  padding: 0.5rem 1.25rem;
  border-radius: 8px;
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.15s;
  border: 1px solid transparent;
}

.cd-btn-cancel {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
  color: #a0aec0;
}
.cd-btn-cancel:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
}

.cd-btn-primary {
  background: rgba(66, 153, 225, 0.2);
  border-color: rgba(66, 153, 225, 0.4);
  color: #63b3ed;
}
.cd-btn-primary:hover {
  background: rgba(66, 153, 225, 0.35);
}

.cd-btn-danger {
  background: rgba(245, 101, 101, 0.2);
  border-color: rgba(245, 101, 101, 0.4);
  color: #fc8181;
}
.cd-btn-danger:hover {
  background: rgba(245, 101, 101, 0.35);
}

.cd-fade-enter-active,
.cd-fade-leave-active {
  transition: opacity 0.15s ease;
}
.cd-fade-enter-from,
.cd-fade-leave-to {
  opacity: 0;
}
.cd-fade-enter-active .cd-modal {
  animation: cd-in 0.15s ease-out;
}
@keyframes cd-in {
  from { transform: scale(0.95); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
</style>
