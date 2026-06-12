<template>
  <Teleport to="body">
    <Transition :name="transitionName">
      <div 
        v-if="visible" 
        class="modal-overlay" 
        :class="overlayClass"
        :style="overlayStyle"
        @click="handleOverlayClick"
      >
        <div 
          class="modal-container" 
          :class="containerClass"
          :style="computedStyle"
          @click.stop
        >
          <!-- 标题栏（可隐藏） -->
          <div v-if="showHeader" class="modal-header" :class="headerClass">
            <slot name="header">
              <h3 class="modal-title">{{ title }}</h3>
            </slot>
            <button v-if="showClose" class="modal-close-btn" @click="close" :title="closeTooltip">
              <slot name="close-icon">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </slot>
            </button>
          </div>
          
          <!-- 内容区 -->
          <div class="modal-body" :class="bodyClass" :style="{ padding: bodyPaddingValue }">
            <slot>{{ content }}</slot>
          </div>
          
          <!-- 底部按钮区（可自定义） -->
          <div v-if="showFooter" class="modal-footer" :class="footerClass">
            <slot name="footer">
              <!-- 默认按钮布局 -->
              <div class="footer-buttons" :class="buttonLayout">
                <button 
                  v-if="showCancel" 
                  class="btn btn-cancel" 
                  :class="cancelButtonClass"
                  @click="handleCancel"
                >
                  {{ cancelText }}
                </button>
                <button 
                  v-if="showConfirm" 
                  class="btn btn-confirm" 
                  :class="confirmButtonClass"
                  @click="handleConfirm"
                  :disabled="confirming || confirmDisabled"
                >
                  {{ confirming ? loadingText : confirmText }}
                </button>
              </div>
            </slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  // ========== 基础配置 ==========
  visible: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: '提示'
  },
  content: {
    type: String,
    default: ''
  },
  
  // ========== 尺寸配置 ==========
  width: {
    type: [String, Number],
    default: '400px'
  },
  height: {
    type: [String, Number],
    default: 'auto'
  },
  maxWidth: {
    type: [String, Number],
    default: '90vw'
  },
  maxHeight: {
    type: [String, Number],
    default: '90vh'
  },
  minWidth: {
    type: [String, Number],
    default: '300px'
  },
  minHeight: {
    type: [String, Number],
    default: 'auto'
  },
  
  // ========== 样式配置 ==========
  borderRadius: {
    type: [String, Number],
    default: '12px'
  },
  bodyPadding: {
    type: [String, Number],
    default: '1.5rem'
  },
  overlayOpacity: {
    type: Number,
    default: 0.7
  },
  overlayBlur: {
    type: Number,
    default: 4
  },
  overlayColor: {
    type: String,
    default: 'var(--bg-overlay)'
  },
  backgroundColor: {
    type: String,
    default: 'var(--bg-panel)'
  },
  borderColor: {
    type: String,
    default: 'var(--border-strong)'
  },
  
  // ========== 显示控制 ==========
  showHeader: {
    type: Boolean,
    default: true
  },
  showClose: {
    type: Boolean,
    default: true
  },
  showFooter: {
    type: Boolean,
    default: true
  },
  showCancel: {
    type: Boolean,
    default: true
  },
  showConfirm: {
    type: Boolean,
    default: true
  },
  
  // ========== 按钮配置 ==========
  cancelText: {
    type: String,
    default: '取消'
  },
  confirmText: {
    type: String,
    default: '确定'
  },
  loadingText: {
    type: String,
    default: '处理中...'
  },
  closeTooltip: {
    type: String,
    default: '关闭'
  },
  confirmDisabled: {
    type: Boolean,
    default: false
  },
  danger: {
    type: Boolean,
    default: false
  },
  
  // ========== 布局配置 ==========
  buttonLayout: {
    type: String,
    default: 'right' // 'left', 'center', 'right', 'space-between'
  },
  
  // ========== 交互配置 ==========
  closeOnOverlay: {
    type: Boolean,
    default: true
  },
  closeOnEscape: {
    type: Boolean,
    default: true
  },
  draggable: {
    type: Boolean,
    default: false
  },
  resizable: {
    type: Boolean,
    default: false
  },
  
  // ========== 动画配置 ==========
  transitionName: {
    type: String,
    default: 'modal'
  },
  animationDuration: {
    type: Number,
    default: 200
  },
  
  // ========== 自定义类名 ==========
  overlayClass: {
    type: [String, Array, Object],
    default: ''
  },
  containerClass: {
    type: [String, Array, Object],
    default: ''
  },
  headerClass: {
    type: [String, Array, Object],
    default: ''
  },
  bodyClass: {
    type: [String, Array, Object],
    default: ''
  },
  footerClass: {
    type: [String, Array, Object],
    default: ''
  },
  cancelButtonClass: {
    type: [String, Array, Object],
    default: ''
  },
  confirmButtonClass: {
    type: [String, Array, Object],
    default: ''
  }
})

const emit = defineEmits(['close', 'cancel', 'confirm', 'update:visible'])

const confirming = ref(false)

// 计算样式
const computedStyle = computed(() => ({
  width: typeof props.width === 'number' ? `${props.width}px` : props.width,
  height: typeof props.height === 'number' ? `${props.height}px` : props.height,
  maxWidth: typeof props.maxWidth === 'number' ? `${props.maxWidth}px` : props.maxWidth,
  maxHeight: typeof props.maxHeight === 'number' ? `${props.maxHeight}px` : props.maxHeight,
  minWidth: typeof props.minWidth === 'number' ? `${props.minWidth}px` : props.minWidth,
  minHeight: typeof props.minHeight === 'number' ? `${props.minHeight}px` : props.minHeight,
  borderRadius: typeof props.borderRadius === 'number' ? `${props.borderRadius}px` : props.borderRadius,
  backgroundColor: props.backgroundColor,
  borderColor: props.borderColor
}))

// 计算内容区内边距
const bodyPaddingValue = computed(() => {
  return typeof props.bodyPadding === 'number' ? `${props.bodyPadding}px` : props.bodyPadding
})

const overlayStyle = computed(() => ({
  backgroundColor: props.overlayColor.startsWith('var(')
    ? props.overlayColor
    : props.overlayColor.replace(/\d+\.\d+\)$/, `${props.overlayOpacity})`),
  backdropFilter: `blur(${props.overlayBlur}px)`
}))

// 关闭模态框
const close = () => {
  emit('close')
  emit('update:visible', false)
}

// 处理遮罩点击
const handleOverlayClick = () => {
  if (props.closeOnOverlay) {
    close()
  }
}

// 处理取消
const handleCancel = () => {
  emit('cancel')
  close()
}

// 处理确认
const handleConfirm = async () => {
  if (confirming.value) return
  
  confirming.value = true
  try {
    await emit('confirm')
  } finally {
    confirming.value = false
  }
}

// ESC 键关闭
const handleEscape = (e) => {
  if (props.closeOnEscape && e.key === 'Escape' && props.visible) {
    close()
  }
}

// 监听可见性变化，添加/移除 ESC 监听
watch(() => props.visible, (newVal) => {
  if (newVal && props.closeOnEscape) {
    document.addEventListener('keydown', handleEscape)
  } else {
    document.removeEventListener('keydown', handleEscape)
  }
}, { immediate: true })
</script>

<style scoped>
/* ========== 遮罩层 ========== */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

/* ========== 容器 ========== */
.modal-container {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  border: 1px solid v-bind('props.borderColor');
  animation: modalSlideIn v-bind('props.animationDuration + "ms"') ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* ========== 标题栏 ========== */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-default);
  flex-shrink: 0;
}

.modal-title {
  margin: 0;
  color: var(--text-primary);
  font-size: 1rem;
  font-weight: 600;
}

.modal-close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.modal-close-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

/* ========== 内容区 ========== */
.modal-body {
  flex: 1;
  color: var(--text-secondary);
  font-size: 0.9375rem;
  line-height: 1.6;
  overflow-y: auto;
  min-height: 0; /* ✅ 关键：允许 flex 子项缩小 */
}

/* ========== 底部按钮区 ========== */
.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border-default);
  flex-shrink: 0;
}

.footer-buttons {
  display: flex;
  gap: 0.75rem;
}

.footer-buttons.left {
  justify-content: flex-start;
}

.footer-buttons.center {
  justify-content: center;
}

.footer-buttons.right {
  justify-content: flex-end;
}

.footer-buttons.space-between {
  justify-content: space-between;
}

/* ========== 按钮 ========== */
.btn {
  padding: 0.5rem 1.25rem;
  border: none;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 80px;
}

.btn-cancel {
  background: var(--surface-2);
  color: var(--text-primary);
}

.btn-cancel:hover {
  background: var(--surface-hover);
}

.btn-confirm {
  background: var(--accent-primary);
  color: var(--text-on-accent);
}

.btn-confirm:hover {
  background: var(--primary-light);
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-danger {
  background: var(--accent-danger);
}

.btn-danger:hover {
  background: var(--danger-light);
}

/* ========== 动画过渡 ========== */
.modal-enter-active,
.modal-leave-active {
  transition: opacity v-bind('props.animationDuration + "ms"') ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

/* fade 动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity v-bind('props.animationDuration + "ms"') ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* slide-up 动画 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all v-bind('props.animationDuration + "ms"') ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

/* zoom 动画 */
.zoom-enter-active,
.zoom-leave-active {
  transition: all v-bind('props.animationDuration + "ms"') ease;
}

.zoom-enter-from,
.zoom-leave-to {
  opacity: 0;
  transform: scale(0.9);
}
</style>
