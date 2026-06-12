<template>
  <div class="home-layout">
    <!-- 顶部标题栏 -->
    <AppLayout />
    
    <!-- 主体区域 -->
    <div class="app-body">
      <!-- 左侧边栏 -->
      <SidebarPanel />
      
      <!-- 右侧内容区 -->
      <div class="content-area">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup>
import AppLayout from './AppLayout.vue'
import SidebarPanel from './SidebarPanel.vue'
</script>

<style scoped>
.home-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background: var(--bg-app);
}

/* 主体区域 */
.app-body {
  display: flex;
  flex: 1;
  overflow: hidden;
  position: relative;
}

/* 右侧内容区 */
.content-area {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 1;
}

.content-area > * {
  width: 100%;
  height: 100%;
  background: var(--bg-panel);
  backdrop-filter: blur(20px);
  border: 0.0625rem solid var(--border-default);
  border-top-left-radius: 1rem;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0.9375rem;
  box-sizing: border-box;
}

/* 滚动条样式 */
.content-area > *::-webkit-scrollbar {
  width: 0.375rem;
}

.content-area > *::-webkit-scrollbar-track {
  background: transparent;
}

.content-area > *::-webkit-scrollbar-thumb {
  background: var(--scrollbar-thumb);
  border-radius: 0.1875rem;
}

.content-area > *::-webkit-scrollbar-thumb:hover {
  background: var(--scrollbar-thumb-hover);
}
</style>
