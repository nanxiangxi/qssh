<template>
  <!-- 顶部标题栏 -->
  <header class="app-header">
    <div class="header-left">
      <div class="logo-container">
        <img src="../../../零启.svg" alt="启SSH Logo" class="logo" />
        <span class="app-name">启SSH</span>
        <span class="version">v{{ version }}</span>
      </div>
    </div>
    
    <div class="header-right">
      <div class="window-controls">
        <button class="control-btn minimize" title="最小化" @click="Window.Minimise()">
          <svg width="12" height="12" viewBox="0 0 12 12">
            <line x1="0" y1="6" x2="12" y2="6" stroke="currentColor" stroke-width="1.5"/>
          </svg>
        </button>
        <button class="control-btn maximize" :title="isMaximised ? '恢复' : '最大化'" @click="toggleMaximise">
          <svg v-if="!isMaximised" width="12" height="12" viewBox="0 0 12 12">
            <rect x="1" y="1" width="10" height="10" stroke="currentColor" stroke-width="1.5" fill="none"/>
          </svg>
          <svg v-else width="12" height="12" viewBox="0 0 12 12">
            <rect x="3" y="1" width="8" height="8" stroke="currentColor" stroke-width="1.5" fill="none"/>
            <rect x="1" y="3" width="8" height="8" stroke="currentColor" stroke-width="1.5" fill="white"/>
          </svg>
        </button>
        <button class="control-btn close" title="关闭" @click="Window.Close()">
          <svg width="12" height="12" viewBox="0 0 12 12">
            <line x1="1" y1="1" x2="11" y2="11" stroke="currentColor" stroke-width="1.5"/>
            <line x1="11" y1="1" x2="1" y2="11" stroke="currentColor" stroke-width="1.5"/>
          </svg>
        </button>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Window } from '@wailsio/runtime'
import { GetVersion } from '../../../../bindings/changeme/ssh/greetservice.js'

const isMaximised = ref(false)
const version = ref('0.2.0')

onMounted(async () => {
  try {
    version.value = await GetVersion()
  } catch (e) {
    console.warn('获取版本失败:', e)
  }
})

// 切换最大化/恢复
async function toggleMaximise() {
  if (isMaximised.value) {
    await Window.Restore()
  } else {
    await Window.Maximise()
  }
}

// 更新最大化按钮状态
async function updateMaximiseButton() {
  isMaximised.value = await Window.IsMaximised()
}

// 监听窗口大小变化
onMounted(() => {
  updateMaximiseButton()
  
  // 定期检查窗口状态
  setInterval(updateMaximiseButton, 500)
})
</script>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 3.5rem;
  padding: 0 1.25rem;
  background: transparent;
  border-bottom: none;
  --wails-draggable: drag;
}

.header-left {
  display: flex;
  align-items: center;
}

.logo-container {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo {
  width: 2rem;
  height: 2rem;
  object-fit: contain;
  filter: invert(1);
}

:global([data-theme="light"]) .logo {
  filter: none;
}

.app-name {
  color: var(--text-primary);
  font-size: 1.3em;
  font-weight: 700;
  letter-spacing: 0.03125rem;
}

.version {
  color: var(--text-muted);
  font-size: 0.7em;
  font-weight: 400;
  padding: 2px 6px;
  background: var(--surface-1);
  border-radius: 4px;
}

.header-right {
  display: flex;
  align-items: center;
}

.window-controls {
  display: flex;
  gap: 0.5rem;
  -webkit-app-region: no-drag;
}

.control-btn {
  width: 2.25rem;
  height: 2.25rem;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.control-btn:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.control-btn.close:hover {
  background: var(--danger-bg);
  color: var(--accent-danger);
}
</style>
