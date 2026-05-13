import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { version } from './package.json'

// Base path - can be overridden via environment variable
// Usage: BASE_PATH=/schematics2 npm run build
const basePath = process.env.BASE_PATH ? `${process.env.BASE_PATH}/client/` : '/client/'
// API base is the root of the base path (e.g. /schematics2/ or /)
const apiBase = process.env.BASE_PATH ? `${process.env.BASE_PATH}/` : '/'

export default defineConfig({
  base: basePath,
  plugins: [vue()],
  define: {
    __API_BASE__: JSON.stringify(apiBase),
    __APP_VERSION__: JSON.stringify(version),
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'https://localhost:8443',
        changeOrigin: true,
        secure: false,
      },
      '/health': {
        target: 'https://localhost:8443',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
