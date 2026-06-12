<template>
  <Teleport to="body">
    <Transition name="cd-fade">
      <div v-if="confirmState.show" class="cd-mask" @click.self="handleCancel">
        <div class="cd-modal" :class="{ 'cd-danger': confirmState.danger }">
          <div class="cd-icon">
            <svg v-if="confirmState.danger" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="var(--accent-danger)" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            <svg v-else width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="var(--primary-light)" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
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
  background: var(--bg-overlay);
  display: flex; align-items: center; justify-content: center;
  z-index: 20000;
  backdrop-filter: blur(2px);
}

.cd-modal {
  background: var(--bg-panel);
  border: 1px solid var(--border-default);
  border-radius: 12px;
  padding: 1.5rem;
  width: 360px;
  max-width: 90vw;
  box-shadow: var(--shadow-lg);
  text-align: center;
}

.cd-modal.cd-danger {
  border-color: var(--border-danger);
}

.cd-icon {
  margin-bottom: 0.75rem;
}

.cd-title {
  margin: 0 0 0.5rem;
  color: var(--text-primary);
  font-size: 1rem;
  font-weight: 600;
}

.cd-message {
  margin: 0 0 1.25rem;
  color: var(--text-secondary);
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
  background: var(--surface-2);
  border-color: var(--border-default);
  color: var(--text-secondary);
}
.cd-btn-cancel:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.cd-btn-primary {
  background: var(--primary-bg);
  border-color: var(--border-accent);
  color: var(--primary-light);
}
.cd-btn-primary:hover {
  background: var(--primary-bg-hover);
}

.cd-btn-danger {
  background: var(--danger-bg);
  border-color: var(--border-danger);
  color: var(--accent-danger);
}
.cd-btn-danger:hover {
  background: rgba(239, 68, 68, 0.25);
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
