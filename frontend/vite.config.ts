import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'node:path';

export default defineConfig({
  plugins: [react()],
  base: '/app/',
  build: {
    outDir: resolve(__dirname, '../static/app'),
    sourcemap: true
  },
  server: {
    port: 4173
  }
});
