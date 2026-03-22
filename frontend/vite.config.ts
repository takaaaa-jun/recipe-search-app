import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

/**
 * Vite設定ファイル
 *
 * プロキシ設定:
 *   /api へのリクエストはバックエンド（ポート8005）に転送されます。
 *   これにより、フロントエンドとバックエンドが異なるポートで起動していても
 *   CORSエラーなくAPIを呼び出せます。
 */
export default defineConfig({
  base: '/recipe-search-app/',
  plugins: [react()],
  server: {
    proxy: {
      // '/api' から始まるリクエストをバックエンドにプロキシ
      '/api': {
        target: 'http://backend:8005',
        changeOrigin: true,
        // Docker外（ローカル開発）では http://localhost:8005 に変更
      },
    },
  },
})
