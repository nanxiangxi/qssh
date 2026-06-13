import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/theme.css'
import { useConfigStore } from './stores/config'
import { Events } from '@wailsio/runtime'
import * as CloudService from '@bindings/changeme/ssh/cloudservice.js'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// 初始化配置
const configStore = useConfigStore()
configStore.init().then(() => {
  console.log('[Main] 配置加载完成')
  const serverUrl = configStore.get('cloud', 'serverUrl')
  const token = configStore.get('cloud', 'token')

  if (serverUrl && token) {
    const addr = serverUrl.replace(/^https?:\/\//, '')
    console.log('[Main] 自动连接云端:', addr)
    CloudService.Connect(addr, token).then(ok => {
      if (ok) {
        console.log('[Main] ✓ 云端自动连接成功')
      } else {
        console.warn('[Main] ✗ 云端自动连接失败')
      }
    })
  } else {
    console.log('[Main] 云端未配置，跳过自动连接')
  }
}).catch(e => {
  console.error('[Main] 配置加载失败:', e)
})

app.mount('#app')
