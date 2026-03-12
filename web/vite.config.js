import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
const backendTarget = process.env.VITE_BACKEND_TARGET || 'http://localhost:8081'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/auth': {
        target: backendTarget,
      },
      '/ws': {
        target: backendTarget,
        ws: true,
      },
      '/api': {
        target: backendTarget,
      },
    },
  },
})
