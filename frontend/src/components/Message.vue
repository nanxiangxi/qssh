<template>
  <span style="display:none"></span>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'

const TYPE_CONFIG = {
  success: {
    color: 'var(--accent-success)',
    borderColor: 'var(--border-default)',
    iconBg: 'var(--success-bg)',
    icon: '<polyline points="20 6 9 17 4 12"/>'
  },
  error: {
    color: 'var(--accent-danger)',
    borderColor: 'var(--border-default)',
    iconBg: 'var(--danger-bg)',
    icon: '<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>'
  },
  warning: {
    color: 'var(--accent-warning)',
    borderColor: 'var(--border-default)',
    iconBg: 'var(--warning-bg)',
    icon: '<path d="M12 9v4"/><path d="M12 17h.01"/><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>'
  },
  info: {
    color: 'var(--accent-primary)',
    borderColor: 'var(--border-default)',
    iconBg: 'var(--primary-bg)',
    icon: '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>'
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

function getThemeVars() {
  const root = getComputedStyle(document.documentElement)
  const isLight = document.documentElement.dataset.theme === 'light'
  return {
    isLight,
    accentSuccess: root.getPropertyValue('--accent-success').trim() || '#48bb78',
    accentDanger: root.getPropertyValue('--accent-danger').trim() || '#fc8181',
    accentWarning: root.getPropertyValue('--accent-warning').trim() || '#fbd38d',
    accentInfo: root.getPropertyValue('--accent-primary').trim() || '#4299e1',
    textPrimary: root.getPropertyValue('--text-primary').trim() || '#e2e8f0',
    bgPanel: root.getPropertyValue('--bg-panel').trim() || 'rgba(45,45,45,0.95)',
    borderDefault: root.getPropertyValue('--border-default').trim() || 'rgba(255,255,255,0.1)',
    shadowMd: root.getPropertyValue('--shadow-md').trim() || '0 8px 32px rgba(0,0,0,0.4)',
    surface2: root.getPropertyValue('--surface-2').trim() || 'rgba(255,255,255,0.05)'
  }
}

function showMessage(content, type = 'info', duration = 3000) {
  ensureContainer()
  const vars = getThemeVars()
  const cfg = TYPE_CONFIG[type] || TYPE_CONFIG.info
  const id = ++idCounter

  const el = document.createElement('div')
  el.dataset.msgId = id
  el.style.pointerEvents = 'auto'
  Object.assign(el.style, {
    display: 'flex',
    alignItems: 'center',
    gap: '0.625rem',
    padding: '0.75rem 1rem',
    background: vars.bgPanel,
    backdropFilter: 'blur(16px) saturate(1.4)',
    border: `1px solid ${cfg.borderColor}`,
    borderRadius: '0.625rem',
    boxShadow: vars.shadowMd,
    animation: '__msg_slideIn__ 0.25s cubic-bezier(0.16,1,0.3,1)',
    transition: 'opacity 0.2s, transform 0.2s',
    fontFamily: 'inherit'
  })

  el.innerHTML = `
    <div style="display:flex;align-items:center;justify-content:center;width:1.5rem;height:1.5rem;border-radius:50%;flex-shrink:0;color:${cfg.color};background:${cfg.iconBg}">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">${cfg.icon}</svg>
    </div>
    <span style="color:${vars.textPrimary};font-size:0.8125rem;flex:1;line-height:1.5">${content}</span>
    <button style="display:flex;align-items:center;justify-content:center;width:1.375rem;height:1.375rem;background:transparent;border:none;border-radius:0.25rem;color:var(--text-muted);cursor:pointer;flex-shrink:0;transition:all 0.15s">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
    </button>
  `

  const closeBtn = el.querySelector('button')
  closeBtn.onmouseenter = () => { closeBtn.style.color = vars.textPrimary; closeBtn.style.background = vars.surface2 }
  closeBtn.onmouseleave = () => { closeBtn.style.color = ''; closeBtn.style.background = 'transparent' }
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
