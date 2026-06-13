<template>
  <span style="display:none"></span>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'

// 读取 CSS 变量
function getCSSVar(name) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

function getTypeConfig() {
  return {
    success: {
      color: getCSSVar('--accent-success') || '#68d391',
      borderColor: getCSSVar('--border-success') || 'rgba(72, 187, 120, 0.4)',
      bgGrad: `linear-gradient(135deg, ${getCSSVar('--success-bg') || 'rgba(72,187,120,0.08)'}, ${getCSSVar('--bg-panel') || 'rgba(30,30,30,0.65)'})`,
      iconBg: getCSSVar('--success-bg') || 'rgba(72, 187, 120, 0.15)',
      icon: '<polyline points="20 6 9 17 4 12"/>'
    },
    error: {
      color: getCSSVar('--accent-danger') || '#fc8181',
      borderColor: getCSSVar('--border-danger') || 'rgba(245, 101, 101, 0.4)',
      bgGrad: `linear-gradient(135deg, ${getCSSVar('--danger-bg') || 'rgba(245,101,101,0.08)'}, ${getCSSVar('--bg-panel') || 'rgba(30,30,30,0.65)'})`,
      iconBg: getCSSVar('--danger-bg') || 'rgba(245, 101, 101, 0.15)',
      icon: '<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>'
    },
    warning: {
      color: getCSSVar('--accent-warning') || '#fbd38d',
      borderColor: getCSSVar('--border-warning') || 'rgba(237, 137, 54, 0.4)',
      bgGrad: `linear-gradient(135deg, ${getCSSVar('--warning-bg') || 'rgba(237,137,54,0.08)'}, ${getCSSVar('--bg-panel') || 'rgba(30,30,30,0.65)'})`,
      iconBg: getCSSVar('--warning-bg') || 'rgba(237, 137, 54, 0.15)',
      icon: '<path d="M12 9v4"/><path d="M12 17h.01"/><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>'
    },
    info: {
      color: getCSSVar('--accent-info') || '#90cdf4',
      borderColor: getCSSVar('--border-accent') || 'rgba(66, 153, 225, 0.4)',
      bgGrad: `linear-gradient(135deg, ${getCSSVar('--primary-bg') || 'rgba(66,153,225,0.08)'}, ${getCSSVar('--bg-panel') || 'rgba(30,30,30,0.65)'})`,
      iconBg: getCSSVar('--primary-bg') || 'rgba(66, 153, 225, 0.15)',
      icon: '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>'
    }
  }
}

let container = null
let idCounter = 0

function ensureContainer() {
  if (container && document.body.contains(container)) return
  container = document.createElement('div')
  container.id = '__message_container__'
  Object.assign(container.style, {
    position: 'fixed',
    top: '1rem',
    right: '1rem',
    zIndex: '10000',
    display: 'flex',
    flexDirection: 'column',
    gap: '0.625rem',
    maxWidth: '22rem',
    pointerEvents: 'none'
  })
  document.body.appendChild(container)
}

function showMessage(content, type = 'info', duration = 3000) {
  ensureContainer()
  const typeConfig = getTypeConfig()
  const cfg = typeConfig[type] || typeConfig.info
  const id = ++idCounter

  const el = document.createElement('div')
  el.dataset.msgId = id
  el.style.pointerEvents = 'auto'
  Object.assign(el.style, {
    display: 'flex',
    alignItems: 'center',
    gap: '0.625rem',
    padding: '0.75rem 1rem',
    background: cfg.bgGrad,
    backdropFilter: 'blur(16px) saturate(1.4)',
    border: `1px solid ${cfg.borderColor}`,
    borderRadius: '0.625rem',
    boxShadow: `0 0.25rem 0.75rem ${getCSSVar('--shadow-sm') || 'rgba(0,0,0,0.3)'}, inset 0 1px 0 ${getCSSVar('--surface-1') || 'rgba(255,255,255,0.04)'}`,
    animation: '__msg_slideIn__ 0.25s cubic-bezier(0.16,1,0.3,1)',
    transition: 'opacity 0.2s, transform 0.2s',
    fontFamily: 'inherit'
  })

  const textColor = getCSSVar('--text-primary') || '#e2e8f0'
  const closeColor = getCSSVar('--text-muted') || 'rgba(255,255,255,0.3)'
  const closeHoverColor = getCSSVar('--text-primary') || 'rgba(255,255,255,0.7)'
  const closeHoverBg = getCSSVar('--surface-hover') || 'rgba(255,255,255,0.08)'

  el.innerHTML = `
    <div style="display:flex;align-items:center;justify-content:center;width:1.5rem;height:1.5rem;border-radius:50%;flex-shrink:0;color:${cfg.color};background:${cfg.iconBg}">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">${cfg.icon}</svg>
    </div>
    <span style="color:${textColor};font-size:0.8125rem;flex:1;line-height:1.5">${content}</span>
    <button style="display:flex;align-items:center;justify-content:center;width:1.375rem;height:1.375rem;background:transparent;border:none;border-radius:0.25rem;color:${closeColor};cursor:pointer;flex-shrink:0;transition:all 0.15s">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
    </button>
  `

  const closeBtn = el.querySelector('button')
  closeBtn.onmouseenter = () => { closeBtn.style.color = closeHoverColor; closeBtn.style.background = closeHoverBg }
  closeBtn.onmouseleave = () => { closeBtn.style.color = closeColor; closeBtn.style.background = 'transparent' }
  closeBtn.onclick = () => removeEl(el)

  container.appendChild(el)

  if (duration > 0) {
    setTimeout(() => removeEl(el), duration)
  }
  return id
}

function removeEl(el) {
  if (!el || !el.parentNode) return
  el.style.opacity = '0'
  el.style.transform = 'translateX(1rem) scale(0.98)'
  setTimeout(() => el.remove(), 250)
}

function success(content, duration) { return showMessage(content, 'success', duration) }
function error(content, duration) { return showMessage(content, 'error', duration) }
function warning(content, duration) { return showMessage(content, 'warning', duration) }
function info(content, duration) { return showMessage(content, 'info', duration) }

// 注入动画样式
onMounted(() => {
  if (!document.getElementById('__msg_anim__')) {
    const style = document.createElement('style')
    style.id = '__msg_anim__'
    style.textContent = `@keyframes __msg_slideIn__{from{opacity:0;transform:translateX(2rem) scale(0.95)}to{opacity:1;transform:translateX(0) scale(1)}}`
    document.head.appendChild(style)
  }
})

onUnmounted(() => {
  // 不移除容器，其他组件可能还在用
})

defineExpose({ success, error, warning, info, showMessage })
</script>
