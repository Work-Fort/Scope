# Getting Started: Vue Service Frontend

This guide walks you through building a service frontend in Vue. The result is a Module Federation remote that runs alongside the shell and other services.

By the end, you'll have a working service with shared auth, UI components, and hot-reload development.

## 1. Scaffold the project

```bash
pnpm create vite my-service --template vue-ts
cd my-service
pnpm install
```

## 2. Install dependencies

Add the required WorkFort packages and Module Federation:

```bash
pnpm add @workfort/ui @workfort/ui-vue @workfort/auth @module-federation/vite
```

- `@workfort/ui` — Lit-based web components (light DOM)
- `@workfort/ui-vue` — Vue composables for auth and theme
- `@workfort/auth` — Singleton authentication client
- `@module-federation/vite` — Vite plugin for Module Federation remotes

## 3. Configure Vite for Module Federation

Replace `vite.config.ts` with:

```ts
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { federation } from '@module-federation/vite';

export default defineConfig({
  plugins: [
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag.startsWith('wf-'),
        },
      },
    }),
    federation({
      name: 'my-service',
      filename: 'remoteEntry.js',
      exposes: {
        './index': './src/index.ts',
      },
      shared: {
        'vue': { singleton: true, import: false },
        '@workfort/ui': { singleton: true, import: false },
        '@workfort/auth': { singleton: true, import: false },
      },
    }),
  ],
  build: {
    target: 'esnext',
    outDir: 'dist',
  },
});
```

Key points:

- `isCustomElement: tag => tag.startsWith('wf-')` — tells Vue to treat `wf-*` tags as native custom elements, avoiding "Unknown custom element" warnings
- `name: 'my-service'` — must match the service name in your Go backend and fort config
- `exposes: { './index': './src/index.ts' }` — exports the entry module
- `shared` modules declare WorkFort packages with `singleton: true, import: false` to consume from the shell instead of bundling copies
- Vue itself is bundled by the remote, not shared by the shell
- `filename: 'remoteEntry.js'` — required for the shell's service discovery health check

## 4. Create the entry module

Create `src/index.ts`:

```ts
import { defineComponent, h } from 'vue';
import { useAuth, useTheme } from '@workfort/ui-vue';
import MyService from './MyService.vue';

export const manifest = {
  name: 'my-service',
  label: 'My Service',
  route: '/my-service',
  minWidth: 320,
};

export default defineComponent({
  name: 'MyServiceRemote',
  props: {
    connected: {
      type: Boolean,
      required: true,
    },
  },
  setup(props) {
    return () => h(MyService, { connected: props.connected });
  },
});

export function SidebarContent() {
  return h('div', 'My Service Sidebar');
}

export function HeaderActions() {
  return h('button', h('wf-button', { variant: 'text' }, 'Settings'));
}
```

Create `src/MyService.vue`:

```vue
<template>
  <wf-panel :label="manifest.label">
    <div style="padding: 1rem">
      <template v-if="isAuthenticated">
        <p>Hello, <strong>{{ user?.displayName }}</strong></p>
        <p>Service is {{ connected ? 'online' : 'offline' }}</p>
        <p>Theme: {{ theme }}</p>
      </template>
      <template v-else>
        <p>Not logged in. Please authenticate via the shell.</p>
      </template>
    </div>
  </wf-panel>
</template>

<script setup lang="ts">
import { useAuth, useTheme } from '@workfort/ui-vue';

const { manifest } = defineProps({
  connected: {
    type: Boolean,
    required: true,
  },
  manifest: {
    type: Object,
  },
});

const { user, isAuthenticated } = useAuth();
const theme = useTheme();
</script>
```

The shell validates that your module exports both `default` and `manifest`. It calls `default(props)` with `connected` state and reads `manifest` for routing and layout. `SidebarContent` and `HeaderActions` are optional.

See [Service Frontend Contract](../service-contract.md) for the full `ServiceModule` spec.

## 5. Wire up the Go backend

See [Getting Started: SolidJS](./solidjs.md#5-wire-up-the-go-backend) — the Go setup is identical across frameworks.

## 6. Add to your fort config

See [Getting Started: SolidJS](./solidjs.md#6-add-to-your-fort-config) — the fort config is identical across frameworks.

## 7. Build and run

### Development

Terminal 1 — build and watch the Go backend:

```bash
mise run dev:go
```

This starts the shell on `:16100` and reloads on Go file changes.

Terminal 2 — run Vite in watch mode:

```bash
cd my-service
pnpm dev
```

Vite watches `src/` and rebuilds the Module Federation remote on changes. The Go backend serves your build from the embedded filesystem at startup, but during development you point the shell to the local Vite dev server. Check the shell's dev config for how it discovers and rewrites service URLs.

Open `http://localhost:16100` in your browser. The shell discovers your service from `fort.yaml`, loads `/ui/remoteEntry.js`, and renders your component.

### Production

```bash
pnpm build
```

The build outputs to `dist/`. Embed this directory in your Go binary as shown in step 5.

## Components and Auth

### Using `@workfort/ui` Web Components

All `wf-*` components work natively in Vue templates thanks to the `isCustomElement` config. Pass props as attributes:

```vue
<template>
  <wf-button variant="primary">Click me</wf-button>
  <wf-input placeholder="Enter text" />
  <wf-panel label="Title">Content here</wf-panel>
</template>
```

See [Shared Packages](../shared-packages.md) for the full component API.

### Authentication

Use the `useAuth()` composable from `@workfort/ui-vue`:

```ts
const { user, isAuthenticated } = useAuth();
// user: Readonly<Ref<User | null>>
// isAuthenticated: Readonly<Ref<boolean>>
```

Both are reactive refs. Wrap conditionals in `computed()` or templates in `v-if` to stay in sync with shell auth state.

See [Authentication](../auth.md) for the BFF pattern and how auth flows through Module Federation.

## Troubleshooting

**"Unknown custom element 'wf-*'" warnings:** Add `isCustomElement` to your Vue compiler options in `vite.config.ts`.

**Module not loading:** Verify your `manifest` values (name, route, label) match exactly between `src/index.ts` and your Go `frontend.Manifest`.

**Auth not working:** Ensure `@workfort/auth` is in the `shared` list with `singleton: true, import: false`. The shell injects the authenticated user — your service reads it via `useAuth()`.

**Vite not finding `src/index.ts`:** Check that the `exposes` entry in federation config points to the correct file path.

## Next Steps

- Explore the [Shared Packages](../shared-packages.md) docs for component APIs
- Add API endpoints in your Go backend and call them from Vue
- Check the [Service Frontend Contract](../service-contract.md) for advanced options like `WSPaths` and custom header actions
