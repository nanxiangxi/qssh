<template>
  <div class="tab-content" :class="{ active: isActive }">
    <div class="main-container">
      <!-- 左侧工具栏 -->
      <LeftToolBar
        :active-panels="activePanels"
        @toggle-panel="handleTogglePanel"
      />

      <!-- 中心内容区 -->
      <div class="center-content">
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
import { ref, computed, onMounted, provide, watch, nextTick } from 'vue'
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

// 标签激活时触发 resize
watch(() => props.isActive, (active) => {
  if (active) {
    nextTick(() => window.dispatchEvent(new Event('resize')))
  }
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
}
</style>
