import { ref } from 'vue'

// 全局确认对话框状态
const confirmState = ref({
  show: false,
  title: '',
  message: '',
  confirmText: '确定',
  cancelText: '取消',
  danger: false,
  resolve: null,
})

export function useConfirm() {
  const confirm = (opts) => {
    return new Promise((resolve) => {
      if (typeof opts === 'string') {
        opts = { message: opts }
      }
      confirmState.value = {
        show: true,
        title: opts.title || '确认操作',
        message: opts.message || '',
        confirmText: opts.confirmText || '确定',
        cancelText: opts.cancelText || '取消',
        danger: opts.danger || false,
        resolve,
      }
    })
  }

  const handleConfirm = () => {
    confirmState.value.resolve?.(true)
    confirmState.value.show = false
  }

  const handleCancel = () => {
    confirmState.value.resolve?.(false)
    confirmState.value.show = false
  }

  return { confirmState, confirm, handleConfirm, handleCancel }
}
