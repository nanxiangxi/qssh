/**
 * 全局配置管理 Store
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as ConfigService from '../../bindings/changeme/ssh/configservice.js'

export const useConfigStore = defineStore('config', () => {
  // 配置值
  const config = ref({
    terminal: {
      defaultType: 'classic',
      autoSwitchClassic: true,
      switchMode: 'prompt', // prompt | auto | inline
      fontSize: 14
    }
  })

  // 是否已加载
  const isLoaded = ref(false)

  // 初始化配置
  async function init() {
    if (isLoaded.value) return

    try {
      const result = await ConfigService.GetConfig()
      if (result) {
        console.log("[Config] loaded from backend:", result)
        config.value = result
      }
      isLoaded.value = true
      console.log('[Config] 配置已加载:', config.value)
    } catch (e) {
      console.error('[Config] 加载配置失败:', e)
      isLoaded.value = true
    }
  }

  // 获取配置值
  function get(category, key) {
    if (!config.value[category]) return undefined
    return config.value[category][key]
  }

  // 设置配置值
  async function set(category, key, value) {
    if (!config.value[category]) {
      config.value[category] = {}
    }
    config.value[category][key] = value

    try {
      await ConfigService.Set(category, key, value)
      console.log('[Config] 配置已保存:', category, key, value)
    } catch (e) {
      console.error('[Config] 保存配置失败:', e)
    }
  }

  // 获取默认终端类型
  function getDefaultTerminalType() {
    console.log("[Config] getDefaultTerminalType raw:", get("terminal", "defaultType"))
    const t = get("terminal", "defaultType")
    return (t === 'structured' || t === 'classic') ? t : 'structured'
  }

  return {
    config,
    isLoaded,
    init,
    get,
    set,
    getDefaultTerminalType
  }
})
