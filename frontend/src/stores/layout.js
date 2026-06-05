import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

// 默认布局配置
const defaultLayout = {
  // 顶部标签栏高度
  topBarHeight: 40,
  
  // 左侧工具栏宽度
  leftToolBarWidth: 56,
  
  // 右侧AI工具栏宽度
  rightToolBarWidth: 56,
  
  // 底部状态栏高度
  bottomBarHeight: 40,
  
  // 主容器分割比例（左-中-右）
  mainSplit: [20, 60, 20], // 百分比
  
  // 是否显示各区域
  visibility: {
    leftToolBar: true,
    rightToolBar: true,
    bottomBar: true
  }
}

export const useLayoutStore = defineStore('layout', () => {
  // 当前布局配置
  const layout = ref({ ...defaultLayout })
  
  // 加载布局配置（从localStorage）
  const loadLayout = async () => {
    try {
      const saved = localStorage.getItem('ssh-layout')
      if (saved) {
        layout.value = { ...defaultLayout, ...JSON.parse(saved) }
      }
    } catch (error) {
      console.error('加载布局配置失败:', error)
    }
  }
  
  // 保存布局配置（到localStorage）
  const saveLayout = async () => {
    try {
      localStorage.setItem('ssh-layout', JSON.stringify(layout.value))
    } catch (error) {
      console.error('保存布局配置失败:', error)
    }
  }
  
  // 重置为默认布局
  const resetLayout = () => {
    layout.value = { ...defaultLayout }
    saveLayout()
  }
  
  // 更新顶部栏高度
  const updateTopBarHeight = (height) => {
    layout.value.topBarHeight = height
    saveLayout()
  }
  
  // 更新左侧工具栏宽度
  const updateLeftToolBarWidth = (width) => {
    layout.value.leftToolBarWidth = width
    saveLayout()
  }
  
  // 更新右侧工具栏宽度
  const updateRightToolBarWidth = (width) => {
    layout.value.rightToolBarWidth = width
    saveLayout()
  }
  
  // 更新底部栏高度
  const updateBottomBarHeight = (height) => {
    layout.value.bottomBarHeight = height
    saveLayout()
  }
  
  // 更新主容器分割比例
  const updateMainSplit = (split) => {
    layout.value.mainSplit = split
    saveLayout()
  }
  
  // 切换区域可见性
  const toggleVisibility = (area) => {
    layout.value.visibility[area] = !layout.value.visibility[area]
    saveLayout()
  }
  
  // 自动保存（监听变化）
  watch(layout, () => {
    // 防抖保存，避免频繁写入
    clearTimeout(window.layoutSaveTimer)
    window.layoutSaveTimer = setTimeout(() => {
      saveLayout()
    }, 1000)
  }, { deep: true })
  
  return {
    layout,
    loadLayout,
    saveLayout,
    resetLayout,
    updateTopBarHeight,
    updateLeftToolBarWidth,
    updateRightToolBarWidth,
    updateBottomBarHeight,
    updateMainSplit,
    toggleVisibility
  }
})
