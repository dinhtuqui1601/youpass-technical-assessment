import { fileURLToPath, URL } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      // NOTE: use "/s/" (with trailing slash) so short-link routes (/s/CODE)
      // are proxied without also matching Vite's /src/* source files.
      '/s/': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
