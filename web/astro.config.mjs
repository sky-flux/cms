// @ts-check
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwindcss from '@tailwindcss/vite';
import node from '@astrojs/node';

// https://astro.build/config
export default defineConfig({
  output: 'server',
  adapter: node({
    mode: 'standalone'
  }),
  server: {
    host: true, // 监听所有接口 (0.0.0.0)
    port: 3000  // 生产环境端口
  },
  integrations: [react()],
  vite: {
    plugins: [tailwindcss()],
    build: {
      rollupOptions: {
        output: {
          manualChunks: (id) => {
            // Split all BlockNote and ProseMirror dependencies into separate chunk
            if (id.includes('@blocknote/') || id.includes('prosemirror-')) {
              return 'blocknote';
            }
            // Split React Query separately
            if (id.includes('@tanstack/react-query')) {
              return 'tanstack';
            }
          }
        }
      },
      chunkSizeWarningLimit: 1600 // BlockNote editor is ~1.5MB (expected)
    },
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true
        }
      }
    }
  }
});
