import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Arahkan semua request /api ke backend Go Anda
      '/api': {
        target: 'http://localhost:8080', // Backend Go berjalan di port 8080
        changeOrigin: true,
        secure: false,
      },
    },
  },
})