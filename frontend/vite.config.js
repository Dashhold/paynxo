import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Dev server proxy: the SPA calls the API under the relative path `/api/*`
// (see src/data/apiClient.js). In development the Go API_Server listens on
// http://localhost:8080, so we proxy `/api` there to avoid CORS and keep the
// client code environment-agnostic. Override the target with the API_PROXY_TARGET
// env var if the backend runs elsewhere.
const API_TARGET = process.env.API_PROXY_TARGET || 'http://localhost:8080'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    open: true,
    proxy: {
      '/api': {
        target: API_TARGET,
        changeOrigin: true,
      },
    },
  },
  // Mirror the proxy for `vite preview` so a production-style preview build also
  // reaches the API at /api without CORS.
  preview: {
    port: 4173,
    proxy: {
      '/api': {
        target: API_TARGET,
        changeOrigin: true,
      },
    },
  },
})
