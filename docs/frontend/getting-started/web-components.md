# Getting Started: Web Components Service Frontend

This guide walks you through building a service frontend with vanilla TypeScript and Web Components. No framework, no adapter — just the raw `@workfort/auth` client and standard DOM APIs.

The result is a Module Federation remote that works for any framework without a dedicated adapter (Preact, Alpine, htmx, etc.).

By the end, you'll have a working service with shared auth, UI components, and hot-reload development.

---

## 1. Scaffold the project

Create a new Vite + vanilla TypeScript project:

```bash
pnpm create vite my-service --template vanilla-ts
cd my-service
```

This creates a standard TypeScript + Vite skeleton.

---

## 2. Install dependencies

Add the required WorkFort packages and Module Federation:

```bash
pnpm add @workfort/ui @workfort/auth @module-federation/vite
```

- `@workfort/ui` — Lit-based web components (light DOM)
- `@workfort/auth` — Singleton authentication client
- `@module-federation/vite` — Vite plugin for Module Federation remotes

No framework adapter needed. `@workfort/ui` components are standard custom elements that work in plain HTML and vanilla JS.

---

## 3. Configure Vite for Module Federation

Replace `vite.config.ts` with:

```ts
import { defineConfig } from 'vite';
import { federation } from '@module-federation/vite';

export default defineConfig({
  plugins: [
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
- `shared` modules declare all WorkFort packages with `singleton: true, import: false` to consume from the shell instead of bundling copies
- `filename: 'remoteEntry.js'` — required for the shell's service discovery health check

---

## 4. Create the entry module

Create `src/index.ts`:

```ts
import '@workfort/ui/style.css';
import { getAuthClient } from '@workfort/auth';

// Manifest describes this service to the shell.
// name, label, route must match the Go-side Manifest.
// minWidth is optional and TypeScript-only.
export const manifest = {
  name: 'my-service',
  label: 'My Service',
  route: '/my-service',
  minWidth: 320,
};

// Default export: a function that receives { connected: boolean } and returns a DOM element.
// This is called once by the shell and the element is mounted into the service panel.
export default function mount(props: { connected: boolean }) {
  // Create the root container
  const el = document.createElement('div');

  // Get the shared auth client
  const client = getAuthClient();

  // Render function updates the DOM when auth state changes
  function render() {
    const user = client.getUser();
    const isAuthenticated = client.isAuthenticated;

    el.innerHTML = '';

    const panel = document.createElement('wf-panel');
    panel.setAttribute('label', manifest.label);

    const content = document.createElement('div');
    content.style.padding = '1rem';

    if (isAuthenticated) {
      const greeting = document.createElement('p');
      greeting.innerHTML = `Hello, <strong>${user?.displayName || 'User'}</strong>`;
      content.appendChild(greeting);

      const status = document.createElement('p');
      status.textContent = `Service is ${props.connected ? 'online' : 'offline'}`;
      content.appendChild(status);
    } else {
      const fallback = document.createElement('p');
      fallback.textContent = 'Not logged in. Please authenticate via the shell.';
      content.appendChild(fallback);
    }

    panel.appendChild(content);
    el.appendChild(panel);
  }

  // Subscribe to auth changes and re-render
  client.on('change', render);
  client.on('logout', render);

  // Initial render
  render();

  return el;
}

// Optional: sidebar content
export function SidebarContent() {
  const el = document.createElement('div');
  el.textContent = 'My Service Sidebar';
  return el;
}

// Optional: header actions
export function HeaderActions() {
  const el = document.createElement('button');
  const btn = document.createElement('wf-button');
  btn.setAttribute('variant', 'text');
  btn.textContent = 'Settings';
  el.appendChild(btn);
  return el;
}
```

The shell validates that your module exports both `default` and `manifest`. It calls `default(props)` with `connected` state and reads `manifest` for routing and layout. `SidebarContent` and `HeaderActions` are optional.

See [Service Frontend Contract](../service-contract.md) for the full `ServiceModule` spec.

---

## 5. Auth client direct usage

Unlike framework adapters, you use `@workfort/auth` directly:

```ts
import { getAuthClient } from '@workfort/auth';

const client = getAuthClient();

// Get current state
const user = client.getUser();        // User | null
const session = client.getSession();  // Session | null
const isAuth = client.isAuthenticated; // boolean (getter)

// Listen for changes
client.on('change', (user) => {
  // User logged in or session refreshed
  // Re-render your component
});

client.on('logout', () => {
  // User logged out
  // Clear local state and re-render
});

// Logout
await client.logout();
```

There is no reactive system — you manually subscribe to events and update the DOM. For a clean pattern, call a `render()` function on every auth event (see example above).

---

## 6. Using Web Components

All `wf-*` components work natively as custom elements:

```ts
const panel = document.createElement('wf-panel');
panel.setAttribute('label', 'My Panel');

const btn = document.createElement('wf-button');
btn.setAttribute('variant', 'filled');
btn.textContent = 'Click me';
panel.appendChild(btn);

btn.addEventListener('click', () => {
  console.log('Clicked!');
});

document.body.appendChild(panel);
```

Always import `@workfort/ui/style.css` in your entry module for component styles.

See [Shared Packages](../shared-packages.md) for the full component list and properties.

---

## 7. Wire up the Go backend

See [SolidJS Guide > Step 5](./solidjs.md#5-wire-up-the-go-backend) for Go wiring. The process is identical — embed the Vite build and call `frontend.Handler(fsys, manifest)`.

---

## 8. Add to your fort config

See [SolidJS Guide > Step 6](./solidjs.md#6-add-to-your-fort-config) for fort config setup. The process is identical.

---

## 9. Build and run

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

Terminal 3 — start your service:

```bash
cd my-service
pnpm dev  # if it has its own dev server
# or just rely on the embedded static server
```

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

The shell initializes `@workfort/auth` at startup. If your component loads before the shell finishes `auth.init()`, the user will be null. This is normal during early page load. Subscribe to `change` events to re-render when auth initializes.

---

## Next Steps

- Explore [Shared Packages](../shared-packages.md) for all available UI components and properties
- Read [Architecture](../architecture.md) to understand how services integrate with the shell
- Check [Service Frontend Contract](../service-contract.md) for the complete spec (optional `SidebarContent`, `HeaderActions`, WebSocket endpoints, etc.)
- See [Authentication](../auth.md) for the BFF pattern and token exchange details
- This same pattern works for any framework without a dedicated adapter (Preact, Alpine, htmx, etc.) — just use `getAuthClient()` and listen to `change`/`logout` events
