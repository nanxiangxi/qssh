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
  background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 50%, #1f1f1f 100%);
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
  background: rgba(45, 45, 45, 0.9);
  backdrop-filter: blur(20px);
  border: 0.0625rem solid rgba(255, 255, 255, 0.1);
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
  background: rgba(255, 255, 255, 0.2);
  border-radius: 0.1875rem;
}

.content-area > *::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
