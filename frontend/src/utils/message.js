import { ref } from 'vue'

// 全局消息实例
let messageInstance = null

// 设置消息实例
export function setMessageInstance(instance) {
  messageInstance = instance
}

// 显示消息
export function showMessage(content, type = 'info', duration = 3000) {
  if (messageInstance) {
    return messageInstance.showMessage(content, type, duration)
  }
  console.warn('Message instance not set')
}

// 便捷方法
export const success = (content, duration) => showMessage(content, 'success', duration)
export const error = (content, duration) => showMessage(content, 'error', duration)
export const warning = (content, duration) => showMessage(content, 'warning', duration)
export const info = (content, duration) => showMessage(content, 'info', duration)
