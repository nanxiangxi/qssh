import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import wails from "@wailsio/runtime/plugins/vite";
import path from "path";

export default defineConfig({
  plugins: [vue(), wails("./bindings")],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
      "@bindings": path.resolve(__dirname, "bindings"),
    },
  },
  server: {
    host: '127.0.0.1',
    port: 9245,
    strictPort: true,
    watch: {
      ignored: ['**/node_modules/**', '**/dist/**', '**/.git/**']
    }
  },
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', 'axios'],
    exclude: ['monaco-editor']
  },
  build: {
    minify: 'esbuild',
    sourcemap: false,
    reportCompressedSize: false,
    rollupOptions: {
      output: {
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
        manualChunks(id) {
          // 1. 彻底隔离 Monaco 及其所有相关依赖（通常是体积最大的元凶）
          if (id.includes('monaco-editor') || id.includes('monaco')) {
            return 'monaco';
          }
          // 2. 隔离其他可能存在的巨型库（如 xterm, echarts, lodash 等，按需添加）
          if (id.includes('xterm') || id.includes('echarts') || id.includes('lodash')) {
            return 'heavy-libs';
          }
          // 3. Vue 核心生态
          if (id.includes('node_modules/vue/') || id.includes('node_modules/vue-router/') || id.includes('node_modules/pinia/')) {
            return 'vue-vendor';
          }
          // 4. Wails 绑定文件
          if (id.includes('bindings')) {
            return 'wails-bindings';
          }
          // 5. 剩余的普通第三方库
          if (id.includes('node_modules')) {
            return 'vendor';
          }
        }
      }
    }
  }
});