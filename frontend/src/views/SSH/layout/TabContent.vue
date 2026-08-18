<template>
  <div class="tab-content" :class="{ active: isActive }">
    <div class="main-container">
      <!-- 左侧工具栏 -->
      <LeftToolBar
        :active-panels="activePanels"
        @toggle-panel="handleTogglePanel"
      />

      <!-- 中心内容区：v-show 隐藏非活跃标签，但保持组件存活 -->
      <div class="center-content" ref="contentRef">
        <DockviewLayout
          v-show="isActive"
          ref="dockviewRef"
          :conn-id="connId"
          @panel-changed="onPanelChanged"
        />
      </div>

      <!-- 右侧AI工具栏 -->
      <RightAIToolbar
        :active-panels="activePanels"
        @toggle-panel="handleTogglePanel"
      />
    </div>

    <BottomStatusBar :conn-id="connId" />
  </div>
</template>

<script setup>
import { ref, computed, provide, watch, onMounted, onUnmounted } from 'vue'
import { Events } from '@wailsio/runtime'
import LeftToolBar from './LeftToolBar.vue'
import RightAIToolbar from './RightAIToolbar.vue'
import BottomStatusBar from './BottomStatusBar.vue'
import DockviewLayout from './DockviewLayout.vue'

const props = defineProps({
  connId: { type: String, required: true },
  isActive: { type: Boolean, default: false }
})

provide('connId', props.connId)

const dockviewRef = ref(null)
const contentRef = ref(null)
const activePanels = ref(['terminal'])

// DockviewLayout 面板变化回调
const onPanelChanged = (panelTypes) => {
  activePanels.value = panelTypes
}

// 工具栏点击 → 切换面板
const handleTogglePanel = (panelType) => {
  if (dockviewRef.value) {
    dockviewRef.value.togglePanel(panelType)
  }
}

// 标签激活时触发 dockview layout 重算和焦点恢复
let resizeTimer = null
watch(() => props.isActive, (active) => {
  console.log(`[TabContent] isActive 变化: connId=${props.connId}, active=${active}`)
  if (active) {
    clearTimeout(resizeTimer)
    resizeTimer = setTimeout(() => {
      console.log(`[TabContent] 触发 resize 和焦点恢复: connId=${props.connId}`)
      triggerDockviewResize()
      // 触发焦点恢复到活动面板
      if (dockviewRef.value?.api) {
        const activePanel = dockviewRef.value.api.activePanel
        console.log(`[TabContent] activePanel:`, activePanel?.id || 'null')
        if (activePanel) {
          activePanel.group.setActive()
        }
      }
      // 通知终端组件恢复焦点
      console.log(`[TabContent] 发送 terminal:focus 事件: connId=${props.connId}`)
      Events.Emit('terminal:focus', { connId: props.connId })
    }, 150)
  }
})

function triggerDockviewResize() {
  const el = contentRef.value
  if (el) {
    const { clientWidth: w, clientHeight: h } = el
    if (w > 0 && h > 0 && dockviewRef.value?.api) {
      dockviewRef.value.api.layout(w, h)
    }
  }
}

// ResizeObserver 兜底：容器尺寸变化时自动 resize dockview
let resizeObs = null
onMounted(() => {
  if (contentRef.value) {
    resizeObs = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect
        if (width > 0 && height > 0 && dockviewRef.value?.api) {
          dockviewRef.value.api.layout(width, height)
        }
      }
    })
    resizeObs.observe(contentRef.value)
  }
})
onUnmounted(() => {
  clearTimeout(resizeTimer)
  resizeObs?.disconnect()
})
</script>

<style scoped>
.tab-content {
  width: 100%;
  height: 100%;
  display: none;
  flex-direction: column;
}

.tab-content.active {
  display: flex;
}

.main-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.center-content {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  position: relative;
}
</style>
