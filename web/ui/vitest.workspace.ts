import { defineWorkspace } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineWorkspace([
  {
    test: {
      name: 'core',
      include: ['tests/components/**/*.test.ts', 'tests/auth/**/*.test.ts'],
      environment: 'happy-dom',
    },
  },
  {
    test: {
      name: 'react',
      include: ['tests/react/**/*.test.tsx'],
      environment: 'happy-dom',
    },
    plugins: [react()],
  },
  {
    test: {
      name: 'vue',
      include: ['tests/vue/**/*.test.ts'],
      environment: 'happy-dom',
    },
  },
  {
    test: {
      name: 'svelte',
      include: ['tests/svelte/**/*.test.ts'],
      environment: 'happy-dom',
    },
  },
  {
    test: {
      name: 'solid',
      include: ['tests/solid/**/*.test.ts'],
      environment: 'happy-dom',
    },
  },
]);
