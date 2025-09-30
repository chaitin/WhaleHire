import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  // 加载环境变量 - 正确加载 VITE_ 前缀的变量
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  
  // 获取环境变量，设置默认值
  const backendHost = env.VITE_BACKEND_HOST || 'http://localhost:8888'
  const devServerPort = parseInt(env.VITE_DEV_SERVER_PORT) || 5175
  const devServerHost = env.VITE_DEV_SERVER_HOST || 'localhost'
  const enableSourcemap = env.VITE_BUILD_SOURCEMAP !== 'false'
  const enableMinify = env.VITE_BUILD_MINIFY !== 'false'

  return {
    plugins: [react()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      port: devServerPort,
      host: devServerHost === 'localhost' ? true : devServerHost,
      allowedHosts: ['hire.chaitin.net'],
      proxy: {
        '/api': {
          target: backendHost,
          changeOrigin: true,
          secure: false,
          rewrite: (path) => path.replace(/^\/api/, '/api'),
        },
      },
    },
    build: {
      outDir: 'dist',
      sourcemap: enableSourcemap,
      minify: enableMinify,
      rollupOptions: {
        output: {
          manualChunks: {
            vendor: ['react', 'react-dom'],
            router: ['react-router-dom'],
          },
        },
      },
    },
    preview: {
      allowedHosts: ['hire.chaitin.net'],
    },
    // 定义全局常量，可在代码中使用
    define: {
      __APP_ENV__: JSON.stringify(env.VITE_APP_ENV || mode),
      __APP_VERSION__: JSON.stringify(env.VITE_APP_VERSION || '1.0.0'),
    },
  }
})
