<template>
  <div v-if="visible" class="dialog-overlay" @click.self="handleCancel">
    <div class="dialog-container">
      <h3 class="dialog-title">选择连接方式</h3>
      
      <div class="options-list">
        <!-- 选项A：打开新窗口 -->
        <div class="option-item" @click="handleSelect('new-window')">
          <div class="option-icon">🪟</div>
          <div class="option-content">
            <div class="option-name">打开新窗口</div>
            <div class="option-desc">为此连接创建独立的窗口和分组</div>
          </div>
        </div>
        
        <!-- 选项B：加入默认分组 -->
        <div class="option-item" @click="handleSelect('default-group')">
          <div class="option-icon">📑</div>
          <div class="option-content">
            <div class="option-name">加入默认分组</div>
            <div class="option-desc">将连接添加到现有标签页中</div>
          </div>
        </div>
        
        <!-- 选项C：取消 -->
        <div class="option-item cancel" @click="handleCancel">
          <div class="option-icon">❌</div>
          <div class="option-content">
            <div class="option-name">取消</div>
            <div class="option-desc">不进行SSH连接</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['select', 'cancel'])

const handleSelect = (type) => {
  emit('select', type)
}

const handleCancel = () => {
  emit('cancel')
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(4px);
}

.dialog-container {
  background: var(--bg-panel-solid);
  border: 1px solid var(--border-default);
  border-radius: 12px;
  padding: 24px;
  min-width: 400px;
  max-width: 500px;
  box-shadow: var(--shadow-lg);
}

.dialog-title {
  color: var(--text-primary);
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 20px 0;
  text-align: center;
}

.options-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--surface-2);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.option-item:hover {
  background: var(--bg-hover);
  border-color: color-mix(in srgb, var(--accent-primary), transparent 50%);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.option-item.cancel:hover {
  border-color: color-mix(in srgb, var(--accent-danger), transparent 50%);
}

.option-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.option-content {
  flex: 1;
}

.option-name {
  color: var(--text-primary);
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 4px;
}

.option-desc {
  color: var(--text-secondary);
  font-size: 0.875rem;
}
</style>
