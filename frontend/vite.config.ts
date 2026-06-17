import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
    server: {
    allowedHosts: ['a6b7267fbd6344b9a9e3f1aa62642178-27339565.eu-north-1.elb.amazonaws.com']
  }
})


