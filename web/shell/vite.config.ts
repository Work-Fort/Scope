import { defineConfig } from 'vite';
import solid from 'vite-plugin-solid';
import UnoCSS from 'unocss/vite';
import { federation } from '@module-federation/vite';

// https://v2.tauri.app/start/frontend/vite/
const host = process.env.TAURI_DEV_HOST;

export default defineConfig({
  plugins: [
    UnoCSS(),
    solid(),
    federation({
      name: 'shell',
      remotes: {},
      shared: {
        'solid-js': { singleton: true, eager: true },
        // @solidjs/router excluded from shared — MF's dev-mode virtual module
        // generator uses require() to detect named exports, which fails for
        // ESM-only packages. Remotes don't need router context (they receive
        // props from the host's router), so sharing is unnecessary.
        '@workfort/ui': { singleton: true, eager: true },
        '@workfort/ui-solid': { singleton: true, eager: true },
        '@workfort/auth': { singleton: true, eager: true },
      },
    }),
  ],
  build: {
    target: 'esnext',
    outDir: 'dist',
  },
  // Vite options tailored for Tauri development
  clearScreen: false,
  server: {
    // Listen on all interfaces for Tauri mobile dev
    host: host || false,
    port: 5173,
    strictPort: true,
    hmr: host ? { protocol: 'ws', host } : undefined,
    watch: {
      // Tell Vite to ignore watching `src-tauri`
      ignored: ['**/src-tauri/**'],
    },
  },
});
