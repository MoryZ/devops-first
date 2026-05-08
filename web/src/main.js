import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersist from 'pinia-plugin-persistedstate'
import Antd from 'ant-design-vue'
import router from './router'
import 'ant-design-vue/dist/reset.css'
import './style.css'
import App from './App.vue'

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersist)
app.use(pinia)
app.use(Antd)
app.use(router)
app.mount('#app')
