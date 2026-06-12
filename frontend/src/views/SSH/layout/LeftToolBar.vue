<template>
  <aside class="left-sidebar">
    <div class="tool-group">
      <button
        v-for="tool in tools"
        :key="tool.id"
        class="tool-btn"
        :class="{ active: isActive(tool.panelType) }"
        :title="tool.title"
        @click="$emit('toggle-panel', tool.panelType)"
      >
        <svg v-if="tool.panelType === 'terminal'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
        </svg>
        <svg v-else-if="tool.panelType === 'fileManager'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
        </svg>
        <svg v-else-if="tool.panelType === 'monitor'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
        </svg>
        <svg v-else-if="tool.panelType === 'aiChat'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
          <line x1="9" y1="9" x2="15" y2="15"/><line x1="15" y1="9" x2="9" y2="15"/>
        </svg>
        <svg v-else-if="tool.panelType === 'logs'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>
          <line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
        </svg>
        <svg v-else-if="tool.panelType === 'portForward'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="7 17 17 7"/><polyline points="7 7 17 7 17 17"/>
          <circle cx="6" cy="18" r="2"/><circle cx="18" cy="6" r="2"/>
        </svg>
        <svg v-else-if="tool.panelType === 'firewall'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
        </svg>
        <svg v-else-if="tool.panelType === 'guardian'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"/>
          <path d="M12 6v6l4 2"/>
        </svg>
      </button>
    </div>
  </aside>
</template>

<script setup>
const props = defineProps({
  activePanels: { type: Array, required: true }
})

defineEmits(['toggle-panel'])

const tools = [
  { id: 'terminal', panelType: 'terminal', title: '终端' },
  { id: 'files', panelType: 'fileManager', title: '文件管理' },
  { id: 'performance', panelType: 'monitor', title: '性能监控' },
  { id: 'portForward', panelType: 'portForward', title: '端口转发' },
  { id: 'firewall', panelType: 'firewall', title: '防火墙' },
  { id: 'guardian', panelType: 'guardian', title: '进程守护' }
]

const isActive = (panelType) => {
  return props.activePanels.includes(panelType)
}
</script>

<style scoped>
.left-sidebar {
  width: 3rem;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.tool-group {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.5rem 0;
  gap: 0.25rem;
}

.tool-btn {
  width: 2.5rem;
  height: 2.5rem;
  background: transparent;
  border: none;
  border-radius: 0.5rem;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  position: relative;
}

.tool-btn:hover {
  background: var(--surface-hover);
  color: var(--text-secondary);
}

.tool-btn.active {
  color: var(--primary-light);
  background: var(--primary-bg);
}

.tool-btn.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 1rem;
  background: var(--accent-primary);
  border-radius: 0 2px 2px 0;
}
</style>
