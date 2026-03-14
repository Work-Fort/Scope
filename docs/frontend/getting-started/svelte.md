# Getting Started: Svelte Service Frontend

This guide walks you through building a service frontend in Svelte. The result is a Module Federation remote that runs alongside the shell and other services.

By the end, you'll have a working service with shared auth, UI components, and hot-reload development.

---

## 1. Scaffold the project

Create a new Vite + Svelte project:

```bash
pnpm create vite my-service --template svelte-ts
cd my-service
```

This creates a standard Svelte TypeScript skeleton with Vite.

---

## 2. Install dependencies

Add the required WorkFort packages and Module Federation:

```bash
pnpm add @workfort/ui @workfort/ui-svelte @workfort/auth @module-federation/vite
```

- `@workfort/ui` — Lit-based web components (light DOM)
- `@workfort/ui-svelte` — Svelte stores for auth and theme
- `@workfort/auth` — Singleton authentication client
- `@module-federation/vite` — Vite plugin for Module Federation remotes

---

## 3. Configure Vite for Module Federation

Replace `vite.config.ts` with:

```ts
import { defineConfig } from 'vite';
import { svelte } from 'vite-plugin-svelte';
import { federation } from '@module-federation/vite';

export default defineConfig({
  plugins: [
    svelte(),
    federation({
      name: 'my-service',
      filename: 'remoteEntry.js',
      exposes: {
        './index': './src/index.ts',
      },
      shared: {
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

- `name: 'my-service'` — must match the service name in your Go backend and fort config
- `exposes: { './index': './src/index.ts' }` — exports the entry module
- `shared` modules declare `@workfort/ui` and `@workfort/auth` with `singleton: true, import: false` to consume from the shell
- Do **not** share `svelte` — it is bundled locally
- `filename: 'remoteEntry.js'` — required for the shell's service discovery health check

---

## 4. Create the entry module

Create `src/App.svelte`:

```svelte
<script lang="ts">
  import { auth, theme } from '@workfort/ui-svelte';

  export let connected: boolean;
</script>

<wf-panel label="My Service">
  <div style="padding: 1rem;">
    {#if $auth.isAuthenticated}
      <p>Hello, <strong>{$auth.user?.displayName}</strong></p>
      <p>Service is {connected ? 'online' : 'offline'}</p>
      <p>Theme: {$theme}</p>
    {:else}
      <p>Not logged in. Please authenticate via the shell.</p>
    {/if}
  </div>
</wf-panel>

<style>
  :global(wf-panel) {
    display: block;
  }
</style>
```

Create `src/index.ts`:

```ts
import App from './App.svelte';

// Manifest describes this service to the shell.
// name, label, route must match the Go-side Manifest.
export const manifest = {
  name: 'my-service',
  label: 'My Service',
  route: '/my-service',
};

// Default export: a function that receives { connected: boolean } and returns a DOM element.
// Since ServiceModule.default expects a function, wrap the Svelte component.
export default function mount(props: { connected: boolean }) {
  const container = document.createElement('div');
  new App({ target: container, props });
  return container;
}

// Optional: render custom sidebar content
export function SidebarContent() {
  const container = document.createElement('div');
  container.textContent = 'My Service Sidebar';
  return container;
}

// Optional: render custom header actions
export function HeaderActions() {
  const container = document.createElement('div');
  const btn = document.createElement('wf-button');
  btn.setAttribute('variant', 'text');
  btn.textContent = 'Settings';
  container.appendChild(btn);
  return container;
}
```

Unlike SolidJS and React (which export component constructors), Svelte's entry exports a factory function that returns an `HTMLElement`. This satisfies the `ServiceModule` contract — the shell accepts either a framework component or a DOM element from `default(props)`.

`SidebarContent` and `HeaderActions` follow the same pattern: plain functions returning DOM elements.

See [Service Frontend Contract](../service-contract.md) for the full `ServiceModule` spec.

---

## 5. Wire up the Go backend

See [SolidJS Getting Started, section 5](./solidjs.md#5-wire-up-the-go-backend). The Go setup is identical. Build your Vite frontend to `web/dist` and embed it the same way.

---

## 6. Add to your fort config

See [SolidJS Getting Started, section 6](./solidjs.md#6-add-to-your-fort-config). The fort config is identical.

---

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

Build both the frontend and backend:

```bash
cd my-service
pnpm build

# From the root (or your service's root):
mise run build
```

The Go binary embeds the Vite build, so the executable contains the complete UI. No separate deployment needed.

---

## Components and Auth

### Using `@workfort/ui` Web Components

All `wf-*` components work natively in Svelte templates:

```svelte
<script lang="ts">
  import { auth } from '@workfort/ui-svelte';
</script>

<wf-panel label="Demo">
  <wf-button variant="filled">Click me</wf-button>
  <wf-badge count={5} />
  <wf-status-dot status="online" />
  <wf-divider />
  <wf-text-input placeholder="Type something" />
</wf-panel>
```

See [Shared Packages](../shared-packages.md) for the full component list and properties.

### Authentication

The `auth` store is reactive. Changes to the user (login, logout, session refresh) are automatically reflected:

```svelte
<script lang="ts">
  import { auth } from '@workfort/ui-svelte';
</script>

{#if $auth.isAuthenticated}
  <div>
    <h2>{$auth.user?.displayName}</h2>
    <p>Email: {$auth.user?.username}</p>
  </div>
{:else}
  <p>Please log in</p>
{/if}
```

`auth.user` and `auth.isAuthenticated` are readable stores. Use `$store` syntax for reactive access.

See [Authentication](../auth.md) for session handling, the BFF pattern, and per-fort cookie scoping.

---

## Troubleshooting

**Module not loading (blank page, no error in browser console):**

Check the shell's browser console for network errors. Verify:
1. The service is running on the correct port
2. `fort.yaml` has the correct `url`
3. Your Go backend returns 200 from `/ui/health` (or 503 if the build is missing)
4. `manifest.name`, `label`, and `route` match exactly between Go and TypeScript

**`connected` is always false:**

`connected` is set by the shell's `ServiceTracker`. If your service is HTTP-only, the tracker checks `/ui/health` every few seconds. If you have WebSocket paths, `connected` is driven by WS connection state instead. See [Service Frontend Contract](../service-contract.md#connected-semantics) for details.

**Auth state is null:**

The shell initializes `@workfort/auth` at startup. If your component loads before the shell finishes `auth.init()`, the user will be null. This is normal during early page load. Re-render when auth events fire.

---

## Next Steps

- Explore [Shared Packages](../shared-packages.md) for all available UI components and stores
- Read [Architecture](../architecture.md) to understand how services integrate with the shell
- Check [Service Frontend Contract](../service-contract.md) for the complete spec (optional `SidebarContent`, `HeaderActions`, WebSocket endpoints, etc.)
- See [Authentication](../auth.md) for the BFF pattern and token exchange details
