import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// 初始化配置
import { useConfigStore } from './stores/config'
const configStore = useConfigStore()
configStore.init()

app.mount('#app')
