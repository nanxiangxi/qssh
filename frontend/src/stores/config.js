/**
 * 全局配置管理 Store
 */
import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import * as ConfigService from '../../bindings/changeme/ssh/configservice.js'
import { Events } from '@wailsio/runtime'

export const useConfigStore = defineStore('config', () => {
  // 配置值
  const config = ref({
    terminal: {
      defaultType: 'classic',
      autoSwitchClassic: true,
      switchMode: 'prompt',
      fontSize: 14
    },
    ui: {
      theme: 'dark'
    }
  })

  // 是否已加载
  const isLoaded = ref(false)

  function applyTheme(theme) {
    const resolved = theme === 'light' ? 'light' : 'dark'
    document.documentElement.dataset.theme = resolved
    console.log('[Config] 主题已应用:', resolved, 'html=', document.documentElement.dataset.theme)
  }

  // 初始化配置
  async function init() {
    if (isLoaded.value) return

    try {
      const result = await ConfigService.GetConfig()
      if (result) {
        console.log('[Config] loaded from backend:', result)
        // 确保 ui 字段存在
        if (!result.ui) {
          result.ui = { theme: 'dark' }
        }
        if (!result.ui.theme) {
          result.ui.theme = 'dark'
        }
        config.value = result
      }
      applyTheme(config.value.ui.theme)
      isLoaded.value = true
      console.log('[Config] 配置已加载:', config.value)

      // 监听其他窗口的主题切换事件，保持多窗口同步
      Events.On('ui:theme-changed', (event) => {
        const t = event?.data?.theme
        console.log('[Config] 收到主题同步事件:', t, event)
        if (!t || t === config.value?.ui?.theme) return
        config.value.ui.theme = t
        applyTheme(t)
        console.log('[Config] 收到其他窗口主题同步:', t)
      })
    } catch (e) {
      console.error('[Config] 加载配置失败:', e)
      applyTheme(config.value.ui.theme)
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

  // 切换主题便捷方法
  async function setTheme(theme) {
    if (!config.value.ui) {
      config.value.ui = { theme: 'dark' }
    }
    config.value.ui.theme = theme
    applyTheme(theme)
    try {
      await ConfigService.Set('ui', 'theme', theme)
      console.log('[Config] 主题已保存:', theme)
      // 通知其他窗口同步主题
      Events.Emit('ui:theme-changed', { theme }).then((ok) => {
        console.log('[Config] 主题同步事件发送结果:', ok)
      }).catch((e) => {
        console.error('[Config] 主题同步事件发送失败:', e)
      })
    } catch (e) {
      console.error('[Config] 保存主题失败:', e)
    }
  }

  // 获取默认终端类型
  function getDefaultTerminalType() {
    console.log('[Config] getDefaultTerminalType raw:', get('terminal', 'defaultType'))
    const t = get('terminal', 'defaultType')
    return (t === 'structured' || t === 'classic') ? t : 'structured'
  }

  // 监听主题变化并自动应用
  watch(
    () => config.value?.ui?.theme,
    (theme) => {
      if (theme) applyTheme(theme)
    }
  )

  return {
    config,
    isLoaded,
    init,
    get,
    set,
    setTheme,
    applyTheme,
    getDefaultTerminalType
  }
})
