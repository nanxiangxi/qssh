<template>
  <aside class="right-sidebar">
    <div class="tool-group">
      <button 
        v-for="tool in aiTools" 
        :key="tool.id"
        class="tool-btn"
        :class="{ active: isPanelActive(tool.panelType) }"
        :title="tool.title"
        @click="togglePanel(tool.panelType)"
      >
        <!-- SVG图标 -->
        <svg v-if="tool.panelType === 'aiChat'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L9.5 8.5 3 9.5l4.5 5L6 21l6-3.5L18 21l-1.5-6.5L21 9.5l-6.5-1z"/>
        </svg>
        <svg v-else-if="tool.panelType === 'batchCmd'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
          <line x1="12" y1="5" x2="20" y2="5"/>
        </svg>
        <svg v-else-if="tool.panelType === 'logs'" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>
          <line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
        </svg>
      </button>
    </div>
    
    <!-- 分隔线 -->
    <div class="divider"></div>
  </aside>
</template>

<script setup>
const props = defineProps({
  activePanels: {
    type: Array,
    required: true
  }
})

const emit = defineEmits(['toggle-panel'])

// AI工具列表（映射到面板类型）
const aiTools = [
  { id: 'ai', panelType: 'aiChat', title: 'AI 助手' },
  { id: 'batchCmd', panelType: 'batchCmd', title: '批量命令' },
  { id: 'logs', panelType: 'logs', title: '操作日志' }
]

// 检查面板是否激活
const isPanelActive = (panelType) => {
  return props.activePanels.includes(panelType)
}

// 切换面板
const togglePanel = (panelType) => {
  emit('toggle-panel', panelType)
}
</script>

<style scoped>
.right-sidebar {
  width: 3.5rem;
  height: 100%;
  background: var(--bg-sidebar);
  border-left: 1px solid var(--border-subtle);
  padding: 0.75rem 0;
  display: flex;
  flex-direction: column;
  flex-shrink: 0; /* 防止被压缩 */
}

.tool-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

/* 分隔线 */
.divider {
  height: 1px;
  margin: 0.5rem 0.75rem;
  background: var(--border-default);
}

.tool-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  margin: 0 auto;
  background: transparent;
  border: none;
  border-radius: 0.5rem;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}

.tool-btn:hover {
  background: var(--surface-hover);
  color: var(--text-secondary);
}

.tool-btn.active {
  background: var(--primary-bg);
  color: var(--accent-primary);
}

/* SVG图标 */
.tool-btn svg {
  width: 24px;
  height: 24px;
}
</style>
