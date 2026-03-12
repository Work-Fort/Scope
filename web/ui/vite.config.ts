import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import dts from 'vite-plugin-dts';

export default defineConfig({
  plugins: [
    react({ include: /\.tsx$/ }),
    dts({ tsconfigPath: './tsconfig.build.json' }),
  ],
  build: {
    lib: {
      entry: {
        index: 'src/index.ts',
        'auth/index': 'src/auth/index.ts',
        'react/index': 'src/react/index.tsx',
        'vue/index': 'src/vue/index.ts',
        'svelte/index': 'src/svelte/index.ts',
        'solid/index': 'src/solid/index.ts',
      },
      formats: ['es'],
    },
    rollupOptions: {
      external: [
        'lit',
        /^lit\//,
        'react',
        'react/jsx-runtime',
        'react-dom',
        'vue',
        /^svelte/,
        'solid-js',
        /^solid-js\//,
      ],
    },
    cssCodeSplit: false,
  },
});
