import { defineConfig } from 'vite';
import solid from 'vite-plugin-solid';
import { federation } from '@module-federation/vite';

export default defineConfig({
  plugins: [
    solid(),
    federation({
      name: 'shell',
      remotes: {},
      shared: {
        'solid-js': { singleton: true, eager: true },
        '@solidjs/router': { singleton: true, eager: true },
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
});
