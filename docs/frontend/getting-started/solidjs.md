# Getting Started: SolidJS Service Frontend

This guide walks you through building a service frontend in SolidJS. The result is a Module Federation remote that runs alongside the shell and other services.

By the end, you'll have a working service with shared auth, UI components, and hot-reload development.

---

## 1. Scaffold the project

Create a new Vite + SolidJS project:

```bash
pnpm create vite my-service --template solid-ts
cd my-service
```

This creates a standard SolidJS TypeScript skeleton with Vite.

---

## 2. Install dependencies

Add the required WorkFort packages and Module Federation:

```bash
pnpm add @workfort/ui @workfort/ui-solid @workfort/auth @module-federation/vite
```

- `@workfort/ui` — Lit-based web components (light DOM)
- `@workfort/ui-solid` — SolidJS hooks for auth and theme
- `@workfort/auth` — Singleton authentication client
- `@module-federation/vite` — Vite plugin for Module Federation remotes

---

## 3. Configure Vite for Module Federation

Replace `vite.config.ts` with:

```ts
import { defineConfig } from 'vite';
import solid from 'vite-plugin-solid';
import { federation } from '@module-federation/vite';

export default defineConfig({
  plugins: [
    solid(),
    federation({
      name: 'my-service',
      filename: 'remoteEntry.js',
      exposes: {
        './index': './src/index.tsx',
      },
      shared: {
        'solid-js': { singleton: true, import: false },
        '@workfort/ui': { singleton: true, import: false },
        '@workfort/ui-solid': { singleton: true, import: false },
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
- `exposes: { './index': './src/index.tsx' }` — exports the entry module
- `shared` modules declare all WorkFort packages with `singleton: true, import: false` to consume from the shell instead of bundling copies
- `filename: 'remoteEntry.js'` — required for the shell's service discovery health check

---

## 4. Create the entry module

Create `src/index.tsx`:

```tsx
import { useAuth, useTheme } from '@workfort/ui-solid';

// Manifest describes this service to the shell.
// name, label, route must match the Go-side Manifest.
// minWidth is optional and TypeScript-only.
export const manifest = {
  name: 'my-service',
  label: 'My Service',
  route: '/my-service',
  minWidth: 320,
};

// Default export: a SolidJS component that receives { connected: boolean }.
// connected is true when the Go backend is reachable (HTTP) or at least one
// WebSocket is connected (if the service declares WSPaths).
export default function MyService(props: { connected: boolean }) {
  const { user, isAuthenticated } = useAuth();
  const theme = useTheme();

  return (
    <wf-panel label={manifest.label}>
      <div style={{ padding: '1rem' }}>
        {isAuthenticated() ? (
          <>
            <p>
              Hello, <strong>{user()?.displayName}</strong>
            </p>
            <p>Service is {props.connected ? 'online' : 'offline'}</p>
            <p>Theme: {theme()}</p>
          </>
        ) : (
          <p>Not logged in. Please authenticate via the shell.</p>
        )}
      </div>
    </wf-panel>
  );
}

// Optional: render custom sidebar content
export function SidebarContent() {
  return <div>My Service Sidebar</div>;
}

// Optional: render custom header actions
export function HeaderActions() {
  return (
    <button>
      <wf-button variant="text">Settings</wf-button>
    </button>
  );
}
```

The shell validates that your module exports both `default` and `manifest`. It calls `default(props)` with `connected` state and reads `manifest` for routing and layout. `SidebarContent` and `HeaderActions` are optional.

See [Service Frontend Contract](../service-contract.md) for the full `ServiceModule` spec.

---

## 5. Wire up the Go backend

Create a simple Go handler that embeds your Vite build and registers it with `pkg/frontend.Handler`.

In your service's main file (e.g., `cmd/my-service/main.go`):

```go
package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/Work-Fort/Scope/pkg/frontend"
)

//go:embed web/dist
var embedFS embed.FS

func main() {
	// Root the filesystem at the Vite build output.
	distFS, _ := fs.Sub(embedFS, "web/dist")

	// Define the manifest. These values must match your TypeScript manifest.
	manifest := frontend.Manifest{
		Name:   "my-service",
		Label:  "My Service",
		Route:  "/my-service",
		// WSPaths: []string{"/api/v1/ws"}, // if you have WebSocket endpoints
	}

	// Handler mounts the UI at /ui/ and serves health checks.
	// Cache headers are set automatically:
	// - /ui/assets/* → 1 year, immutable
	// - /ui/* → no-cache
	handler := frontend.Handler(distFS, manifest)

	mux := http.NewServeMux()

	// Mount the frontend on /ui/
	mux.Handle("/ui/", handler)

	// Mount your API on /api/
	// mux.Handle("POST /api/v1/greet", handleGreet)

	server := &http.Server{
		Addr:    ":16200",
		Handler: mux,
	}

	server.ListenAndServe()
}
```

Key details:

- `//go:embed web/dist` — embeds the Vite build output
- `fs.Sub(embedFS, "web/dist")` — creates a filesystem rooted at `dist/`
- `frontend.Handler(fsys, manifest)` — mounts routes under `/ui/`
- `manifest` values **must match** your TypeScript `src/index.tsx` exports exactly

See [Service Frontend Contract](../service-contract.md) for the Go `Manifest` and `Handler` API.

---

## 6. Add to your fort config

In your `fort.yaml`:

```yaml
forts:
  my-fort:
    services:
      - name: my-service
        url: http://localhost:16200
        # Optional: enable WebSocket connection tracking
        # wsUrl: ws://localhost:16200
```

The shell's service tracker polls `/ui/health` and loads the remote from `/ui/remoteEntry.js`. The `url` field is where the shell finds your service.

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

## Components and Auth

### Using `@workfort/ui` Web Components

All `wf-*` components work natively in SolidJS JSX:

```tsx
import { useAuth } from '@workfort/ui-solid';

export default function Demo() {
  const { user } = useAuth();

  return (
    <wf-panel label="Demo">
      <wf-button variant="filled">Click me</wf-button>
      <wf-badge count={5} />
      <wf-status-dot status="online" />
      <wf-divider />
      <wf-text-input placeholder="Type something" />
    </wf-panel>
  );
}
```

See [Shared Packages](../shared-packages.md) for the full component list and properties.

### Authentication

The `useAuth()` hook is reactive. Changes to the user (login, logout, session refresh) are automatically reflected:

```tsx
import { useAuth } from '@workfort/ui-solid';

export default function Profile() {
  const { user, isAuthenticated } = useAuth();

  return (
    <Show
      when={isAuthenticated()}
      fallback={<p>Please log in</p>}
    >
      <div>
        <h2>{user()?.displayName}</h2>
        <p>Email: {user()?.username}</p>
      </div>
    </Show>
  );
}
```

`user` is a SolidJS signal. `isAuthenticated` is a derived accessor.

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

The shell initializes `@workfort/auth` at startup. If your component loads before the shell finishes `auth.init()`, the user will be null. This is normal during early page load. Re-render when auth events fire or use error boundaries.

---

## Next Steps

- Explore [Shared Packages](../shared-packages.md) for all available UI components and hooks
- Read [Architecture](../architecture.md) to understand how services integrate with the shell
- Check [Service Frontend Contract](../service-contract.md) for the complete spec (optional `SidebarContent`, `HeaderActions`, WebSocket endpoints, etc.)
- See [Authentication](../auth.md) for the BFF pattern and token exchange details
