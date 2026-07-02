<template>
  <div class="dockview-container" :class="themeClass" @contextmenu.prevent="onContainerContextMenu">
    <DockviewVue
      ref="dockviewRef"
      :default-layout="defaultLayout"
      @ready="onReady"
      class="dockview-host"
    />
    <!-- 右键菜单 -->
    <Teleport to="body">
      <div v-if="ctxMenu.show" class="ctx-menu-mask" @mousedown="ctxMenu.show=false">
        <div class="ctx-menu" :style="{top:ctxMenu.y+'px', left:ctxMenu.x+'px'}" @mousedown.stop>
          <div v-if="ctxMenu.panelType === 'terminal'" class="ctx-item" @click="newPanelBeside(ctxMenu.panelId, ctxMenu.panelType)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            新建终端
          </div>
          <div v-if="ctxMenu.panelType === 'terminal'" class="ctx-sep"></div>
          <div class="ctx-item" @click="closePanel(ctxMenu.panelId)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
            关闭
          </div>
          <div class="ctx-item" @click="closeOtherPanels(ctxMenu.panelId)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18M15 3v18"/></svg>
            关闭其他
          </div>
          <div class="ctx-item" @click="closeAllByType(ctxMenu.panelType)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 9l6 6M15 9l-6 6"/></svg>
            关闭全部{{ getPanelTitle(ctxMenu.panelType) }}
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script>
import { defineComponent, ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { DockviewVue } from 'dockview-vue'
import 'dockview-vue/dist/styles/dockview.css'
import { PANEL_CONFIG } from '../../../stores/sshLayout'
import { useTerminalSessionStore } from '../../../stores/terminalSessions'
import { useConfigStore } from '../../../stores/config'
import { Events } from '@wailsio/runtime'
import { initAIToolExecutor, destroyAIToolExecutor } from '../../../utils/aiToolExecutor'

import StructuredTerminalPanel from '../panels/Terminal/StructuredTerminalPanel.vue'
import FileManagerPanel from '../panels/FileManager/FileManagerPanel.vue'
import MonitorPanel from '../panels/MonitorPanel.vue'
import AIChatPanel from '../panels/AIChatPanel.vue'
import LogsPanel from '../panels/LogsPanel.vue'
import PortForwardPanel from '../panels/PortForwardPanel.vue'
import FirewallPanel from '../panels/FirewallPanel.vue'
import ProcessGuardPanel from '../panels/ProcessGuardPanel.vue'
import BatchCommandPanel from '../panels/BatchCommandPanel.vue'

export default defineComponent({
  name: 'DockviewLayout',
  components: {
    DockviewVue,
    terminal: StructuredTerminalPanel,  // 结构化终端（默认）
    
    fileManager: FileManagerPanel,
    monitor: MonitorPanel,
    aiChat: AIChatPanel,
    logs: LogsPanel,
    portForward: PortForwardPanel,
    firewall: FirewallPanel,
    guardian: ProcessGuardPanel,
    batchCmd: BatchCommandPanel
  },
  props: {
    connId: { type: String, required: true }
  },
  emits: ['panel-changed'],
  setup(props, { emit }) {
    const dockviewRef = ref(null)
    let dockviewApi = null

    const ctxMenu = reactive({ show: false, x: 0, y: 0, panelId: '', panelType: '' })
    const configStore = useConfigStore()

    const themeClass = computed(() => {
      return configStore.config?.ui?.theme === 'light' ? 'dockview-theme-light' : 'dockview-theme-dark'
    })

    const defaultLayout = {
      orientation: 'HORIZONTAL',
      root: { type: 'branch', data: [] }
    }

    const getPanelTitle = (type) => PANEL_CONFIG[type]?.title || type
    const sessionStore = useTerminalSessionStore()

    // 获取终端组件名称
    const getTerminalComponent = () => {
      const type = configStore.getDefaultTerminalType()
      return 'terminal'
    }

    // --- 面板管理 ---

    /** 获取下一个可用的终端编号 */
    const getNextTerminalNum = (prefix) => {
      if (!dockviewApi) return 1
      const existing = dockviewApi.panels
        .filter(p => p.id.startsWith(prefix))
        .map(p => {
          const match = p.id.match(/(\d+)$/)
          return match ? parseInt(match[1]) : 0
        })
      return existing.length > 0 ? Math.max(...existing) + 1 : 1
    }

    const addPanel = async (panelType, opts = {}) => {
      if (!dockviewApi) return null
      const config = PANEL_CONFIG[panelType]
      if (!config) return null
      const { setActive = true, referencePanel = null, position = null, isAI = false, group = null } = opts

      let panelId = panelType
      let title = config.title
      let sessionId = null
      let component = panelType

      if (panelType === 'terminal') {
        if (isAI) {
          const num = getNextTerminalNum('terminal_ai_')
          panelId = `terminal_ai_${num}`
          title = `AI 终端-${num}`
        } else {
          const num = getNextTerminalNum('terminal_')
          panelId = `terminal_${num}`
          title = `终端 ${num}`
        }
        // 使用配置的终端类型
        component = getTerminalComponent()
        // 分配 sessionId（不启动 shell，shell 由 TerminalPanel 挂载后启动）
        sessionId = sessionStore.createSession(panelId, props.connId, isAI)
        if (!sessionId) {
          console.warn('[DockviewLayout] 会话创建失败（可能已达上限），取消添加面板')
          return null
        }
      } else {
        const existing = dockviewApi.getPanel(panelType)
        if (existing) {
          if (setActive) existing.group.setActive()
          return existing
        }
      }

      try {
        const addOpts = {
          id: panelId,
          component,
          title,
          params: { connId: props.connId, sessionId, isAI, panelId }
        }
        if (referencePanel && position) {
          // 分割模式：添加到指定面板旁边
          addOpts.position = { referencePanel, direction: position }
        } else if (group) {
          // 添加到指定分组（作为标签页）
          addOpts.group = group
        }

        const panel = dockviewApi.addPanel(addOpts)
        if (setActive && panel.group) panel.group.setActive()
        emitChange()
        return panel
      } catch (e) {
        console.error('[DockviewLayout] 添加面板失败:', e)
        return null
      }
    }

    const closePanel = (panelId) => {
      if (!dockviewApi) return
      const panel = dockviewApi.getPanel(panelId)
      if (panel) {
        dockviewApi.removePanel(panel)
        // emitChange 由 onDidRemovePanel 触发
      }
      ctxMenu.show = false
    }

    const closeOtherPanels = (keepPanelId) => {
      if (!dockviewApi) return
      ;[...dockviewApi.panels].forEach(p => {
        if (p.id !== keepPanelId && !p.id.startsWith('terminal_ai_')) dockviewApi.removePanel(p)
      })
      // emitChange 由 onDidRemovePanel 逐个触发
      ctxMenu.show = false
    }

    const closeAllByType = (panelType) => {
      if (!dockviewApi) return
      dockviewApi.panels
        .filter(p => panelType === 'terminal' ? p.id.startsWith('terminal_') && !p.id.startsWith('terminal_ai_') : p.id === panelType)
        .forEach(p => dockviewApi.removePanel(p))
      // emitChange 由 onDidRemovePanel 逐个触发
      ctxMenu.show = false
    }

    const togglePanel = (panelType) => {
      if (!dockviewApi) return
      const hasPanel = dockviewApi.panels.some(p =>
        panelType === 'terminal' ? p.id.startsWith('terminal_') && !p.id.startsWith('terminal_ai_') : p.id === panelType
      )
      if (hasPanel) {
        // 关闭该类型的所有面板
        dockviewApi.panels
          .filter(p => panelType === 'terminal' ? p.id.startsWith('terminal_') && !p.id.startsWith('terminal_ai_') : p.id === panelType)
          .forEach(p => dockviewApi.removePanel(p))
        // emitChange 由 onDidRemovePanel 逐个触发
      } else {
        addPanel(panelType)
      }
    }

    const newPanelBeside = (refPanelId, panelType) => {
      if (!dockviewApi || !refPanelId) return
      const refPanel = dockviewApi.getPanel(refPanelId)
      if (!refPanel) return
      // 在同一分组中添加同类型面板（不创建新分割，不干扰布局）
      // 对于 terminal 类型，允许多个实例；其他类型只允许一个
      if (panelType !== 'terminal') {
        const existing = dockviewApi.getPanel(panelType)
        if (existing) {
          existing.group.setActive()
          ctxMenu.show = false
          return
        }
      }
      addPanel(panelType, {
        setActive: true,
        group: refPanel.group
      })
      ctxMenu.show = false
    }

    // --- 事件 ---

    const emitChange = () => {
      if (!dockviewApi) return
      const types = new Set()
      const terminals = []
      dockviewApi.panels.forEach(p => {
        if (p.id.startsWith('terminal')) {
          types.add('terminal')
          const sess = sessionStore.getSession(p.id)
          terminals.push({
            id: p.id,
            title: p.title || p.id,
            sessionId: sess?.sessionId || null,
            isAI: sess?.isAI || false,
            connId: props.connId
          })
        } else {
          types.add(p.id)
        }
      })
      emit('panel-changed', Array.from(types))
      Events.Emit('dockview:terminals-changed', { connId: props.connId, terminals })
    }

    const onContainerContextMenu = (e) => {
      // 检查是否点击在面板 tab 上
      const tab = e.target.closest('.dv-tab')
      if (!tab) return

      // 获取面板 ID
      const panelId = tab.dataset?.panelId || tab.textContent?.trim()
      // 通过 tab 找到对应的面板
      const panel = dockviewApi?.panels.find(p => {
        const tabEl = document.querySelector(`[data-panel-id="${p.id}"]`)
        return tabEl === tab || p.title === tab.textContent?.trim()
      })
      if (!panel) return

      const panelType = panel.id.startsWith('terminal') ? 'terminal' : panel.id
      ctxMenu.panelId = panel.id
      ctxMenu.panelType = panelType
      ctxMenu.x = e.clientX
      ctxMenu.y = e.clientY
      ctxMenu.show = true
    }

    // --- 生命周期 ---

    const onReady = (event) => {
      dockviewApi = event.api

      // 面板关闭时关闭对应的 SSH 会话
      dockviewApi.onDidRemovePanel((panel) => {
        sessionStore.closeSession(panel.id)
        emitChange()
      })

      // 面板切换时通知对应面板（用于 AI 聊天滚动到底部、终端聚焦等）
      dockviewApi.onDidActivePanelChange((panel) => {
        if (panel) {
          // 使用 document 事件（同窗口内比 Wails Events 更可靠）
          document.dispatchEvent(new CustomEvent('dockview:panel-activated', { detail: { panelId: panel.id } }))
        }
      })

      // 监听 AI 创建终端请求
      Events.On('ai:create-terminal', (event) => {
        if (!event?.data || event.data.connId !== props.connId) return
        addPanel('terminal', { setActive: true, isAI: true })
      })

      // 监听 AI 关闭终端请求（通过 sessionId 查找面板，过滤当前连接）
      Events.On('ai:close-terminal', (event) => {
        if (!event?.data || !dockviewApi) return
        // 只处理属于当前连接的关闭请求
        if (event.data.connId && event.data.connId !== props.connId) return
        const sid = event.data.sessionId
        const panelId = sessionStore.findPanelBySessionId(sid)
        if (panelId) {
          const panel = dockviewApi.getPanel(panelId)
          if (panel) {
            dockviewApi.removePanel(panel)
            console.log('[DockviewLayout] AI 关闭终端:', panelId)
          }
        }
      })

      // 监听终端 shell 启动完成
      Events.On('terminal:session-ready', (e) => {
        const d = e?.data
        if (!d || d.connId !== props.connId) return
        // 标记 session 就绪
        sessionStore.markSessionReady(d.sessionId)
        if (d.isAI) {
          Events.Emit('ai:ai-terminal-ready', { connId: d.connId, sessionId: d.sessionId })
          console.log('[DockviewLayout] AI 终端就绪:', d.sessionId)
        }
        emitChange()
      })

      // 加载默认终端
      addPanel('terminal', { setActive: true })
    }

    onMounted(() => {
      initAIToolExecutor()
      console.log('[DockviewLayout] 组件挂载, connId:', props.connId)
    })

    onUnmounted(() => {
      destroyAIToolExecutor()
      Events.Off('ai:create-terminal')
      Events.Off('ai:close-terminal')
      Events.Off('terminal:session-ready')
      console.log('[DockviewLayout] 组件卸载')
    })

    return {
      dockviewRef,
      defaultLayout,
      ctxMenu,
      themeClass,
      getPanelTitle,
      addPanel,
      closePanel,
      closeOtherPanels,
      closeAllByType,
      togglePanel,
      newPanelBeside,
      onContainerContextMenu,
      onReady
    }
  }
})
</script>

<style scoped>
/* 覆盖 dockview 默认 CSS 变量 */
:host,
.dockview-container {
  --dv-group-view-background-color: var(--toolbar-4) !important;
  --dv-tabs-and-actions-container-background: var(--toolbar-2) !important;
  --dv-tabs-and-actions-container-height: 35px;
  --dv-tabs-and-actions-container-font-size: 13px;
}

.dockview-container {
  width: 100%;
  height: 100%;
  position: relative;
  background: var(--toolbar-4);
}

.dockview-host {
  width: 100%;
  height: 100%;
}

/* Dockview 主题覆盖 */
:deep(.dv-tabs-and-actions-container) {
  background: var(--toolbar-2) !important;
  border-bottom: 1px solid var(--border-default) !important;
}

:deep(.dv-scrollable) {
  flex: 1 !important;
  min-width: 0 !important;
}

:deep(.dv-scrollbar) {
  display: none !important;
}

:deep(.dv-tabs-container) {
  flex: 1 !important;
  min-width: 0 !important;
}

/* void 容器不占空间，让标签填满整行 */
:deep(.dv-void-container) {
  flex: 0 0 0 !important;
  overflow: hidden !important;
}

/* 外层 dv-tab 承担所有视觉样式 */
:deep(.dv-tab) {
  background: var(--toolbar-3) !important;
  color: var(--text-secondary) !important;
  border: 1px solid var(--border-subtle) !important;
  border-bottom: none !important;
  border-radius: 0 !important;
  padding: 6px 14px !important;
  margin-right: -1px !important;
  flex-shrink: 0 !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
  transition: background 0.15s, color 0.15s, border-color 0.15s !important;
  display: flex !important;
  align-items: center !important;
  gap: 6px !important;
}

:deep(.dv-tab:hover) {
  background: var(--toolbar-4) !important;
  color: var(--text-primary) !important;
}

:deep(.dv-active-tab) {
  background: var(--toolbar-4) !important;
  color: var(--text-primary) !important;
  border-color: var(--border-default) !important;
  border-bottom: 2px solid var(--accent-primary) !important;
  z-index: 1 !important;
}

:deep(.dv-inactive-tab) {
  background: var(--toolbar-3) !important;
  color: var(--text-secondary) !important;
  border-color: var(--border-subtle) !important;
}

/* 内层消失，布局穿透到外层 */
:deep(.dv-default-tab) {
  display: contents !important;
}

/* 关闭按钮：非激活时隐藏，hover 外层时显示 */
:deep(.dv-inactive-tab .dv-default-tab-action) {
  visibility: hidden !important;
}
:deep(.dv-tab:hover .dv-default-tab-action) {
  visibility: visible !important;
}
:deep(.dv-active-tab .dv-default-tab-action) {
  visibility: visible !important;
}

:deep(.dv-default-tab-content) {
  flex: 1 !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
}

:deep(.dv-group-view) {
  background: var(--toolbar-4) !important;
}

:deep(.dv-separator) {
  background: var(--border-default) !important;
}

:deep(.dv-watermark-container) {
  background: var(--toolbar-4) !important;
}

:deep(.dv-group-view > div) {
  height: 100% !important;
  width: 100% !important;
  overflow: hidden !important;
}

:deep(.file-manager),
:deep(.terminal-panel),
:deep(.monitor-panel),
:deep(.ai-chat-panel),
:deep(.logs-panel) {
  height: 100% !important;
  width: 100% !important;
  overflow: hidden !important;
}

/* 确保 dockview 面板内容区域不产生多余滚动条 */
:deep(.dv-view) {
  overflow: hidden !important;
}

:deep(.dv-pane) {
  overflow: hidden !important;
}

/* 右键菜单 */
.ctx-menu-mask {
  position: fixed;
  inset: 0;
  z-index: 10000;
}

.ctx-menu {
  position: fixed;
  background: var(--bg-panel);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 4px;
  min-width: 160px;
  box-shadow: 0 8px 32px var(--shadow-lg);
  z-index: 10001;
  animation: ctxIn 0.12s ease-out;
}

@keyframes ctxIn {
  from { opacity: 0; transform: scale(0.95); }
}

.ctx-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  color: var(--text-secondary);
  font-size: 12px;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.12s;
}

.ctx-item:hover {
  background: var(--primary-bg);
  color: var(--text-primary);
}

.ctx-sep {
  height: 1px;
  background: var(--border-subtle);
  margin: 4px 0;
}
</style>

<style>
/* 非 scoped 样式：覆盖 dockview 默认样式 */

/* === 覆盖 dockview CSS 变量（控制 hover 时的默认颜色） === */
.dockview-container {
  --dv-tabs-container-scrollbar-color: transparent;
  --dv-scrollbar-background-color: transparent;
}

/* === dockview 自定义滚动条（.dv-scrollbar 元素，非浏览器原生） === */
.dockview-container .dv-scrollbar {
  background-color: var(--primary-bg) !important;
  background-image: linear-gradient(135deg, var(--primary-bg), var(--border-accent)) !important;
  border-radius: 999px !important;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1) !important;
  cursor: pointer !important;
}

/* 水平滚动条 */
.dockview-container .dv-scrollbar-horizontal {
  height: 3px !important;
  bottom: 0 !important;
}

.dockview-container .dv-scrollable:hover .dv-scrollbar-horizontal,
.dockview-container .dv-scrollable.dv-scrollable-scrolling .dv-scrollbar-horizontal,
.dockview-container .dv-scrollable.dv-scrollable-resizing .dv-scrollbar-horizontal {
  height: 6px !important;
  background-color: var(--border-accent) !important;
  background-image: linear-gradient(135deg, var(--border-accent), rgba(66, 153, 225, 0.6)) !important;
  box-shadow: 0 0 10px var(--primary-bg-hover) !important;
}

/* 垂直滚动条 */
.dockview-container .dv-scrollbar-vertical {
  width: 3px !important;
  right: 0 !important;
}

.dockview-container .dv-scrollable:hover .dv-scrollbar-vertical,
.dockview-container .dv-scrollable.dv-scrollable-scrolling .dv-scrollbar-vertical,
.dockview-container .dv-scrollable.dv-scrollable-resizing .dv-scrollbar-vertical {
  width: 6px !important;
  background-color: var(--border-accent) !important;
  background-image: linear-gradient(135deg, var(--border-accent), rgba(66, 153, 225, 0.6)) !important;
  box-shadow: 0 0 10px var(--primary-bg-hover) !important;
}

/* === dockview 溢出下拉按钮（.dv-tabs-overflow-dropdown-default） === */
.dockview-container .dv-tabs-overflow-dropdown-default {
  background: var(--primary-bg) !important;
  border: 1px solid var(--primary-bg) !important;
  border-radius: 6px !important;
  color: var(--primary-light) !important;
  font-size: 11px !important;
  font-weight: 500 !important;
  padding: 2px 10px !important;
  cursor: pointer !important;
  display: flex !important;
  align-items: center !important;
  gap: 4px !important;
  transition: all 0.15s !important;
}

.dockview-container .dv-tabs-overflow-dropdown-default:hover {
  background: var(--primary-bg) !important;
  border-color: var(--border-accent) !important;
  color: var(--accent-info) !important;
}

.dockview-container .dv-tabs-overflow-dropdown-default .dv-svg {
  fill: var(--primary-light) !important;
}

.dockview-container .dv-tabs-overflow-dropdown-default:hover .dv-svg {
  fill: var(--accent-info) !important;
}

/* === dockview 溢出容器（下拉列表） === */
.dockview-container .dv-tabs-overflow-container {
  background: var(--bg-panel-solid) !important;
  border: 1px solid var(--border-default) !important;
  border-radius: 10px !important;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.6), 0 0 1px var(--surface-hover) !important;
  padding: 4px !important;
  backdrop-filter: blur(12px) !important;
}

.dockview-container .dv-tabs-overflow-container::-webkit-scrollbar {
  width: 3px !important;
}

.dockview-container .dv-tabs-overflow-container::-webkit-scrollbar-track {
  background: transparent !important;
}

.dockview-container .dv-tabs-overflow-container::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, var(--primary-bg), var(--border-accent)) !important;
  border-radius: 999px !important;
}

.dockview-container .dv-tabs-overflow-container .dv-tab {
  border-radius: 4px !important;
  margin: 1px 0 !important;
  color: var(--text-secondary) !important;
}

.dockview-container .dv-tabs-overflow-container .dv-tab:hover {
  background: var(--primary-bg) !important;
  color: var(--text-primary) !important;
}

.dockview-container .dv-tabs-overflow-container .dv-active-tab {
  background: var(--primary-bg) !important;
  color: var(--primary-light) !important;
}

/* === 面板内 select 下拉框 === */
.dockview-container select {
  background: var(--bg-panel) !important;
  border: 1px solid var(--surface-hover) !important;
  border-radius: 6px !important;
  color: var(--text-secondary) !important;
  font-size: 12px !important;
  padding: 4px 24px 4px 8px !important;
  outline: none !important;
  appearance: none !important;
  -webkit-appearance: none !important;
  background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1L5 5L9 1' stroke='%23718096' stroke-width='1.5' stroke-linecap='round'/%3E%3C/svg%3E") !important;
  background-repeat: no-repeat !important;
  background-position: right 8px center !important;
}

.dockview-container select:focus {
  border-color: var(--border-accent) !important;
}

.dockview-container select option {
  background: var(--bg-panel-solid) !important;
  color: var(--text-primary) !important;
}
</style>
