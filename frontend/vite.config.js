import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Base path - can be overridden via environment variable
// Usage: BASE_PATH=/schematics2 npm run build
const basePath = process.env.BASE_PATH ? `${process.env.BASE_PATH}/client/` : '/client/'

export default defineConfig({
  base: basePath,
  plugins: [vue()],
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
