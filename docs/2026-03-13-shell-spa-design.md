# Shell SPA — Design Spec

**Date:** 2026-03-13
**Status:** Draft
**Spec for:** `web/shell/` — the WorkFort web shell SPA

---

## Overview

The shell is a thin SolidJS application that serves as the Module Federation host for the WorkFort platform. It owns chrome (nav bar, service tabs, sidebar frame, theme toggle), auth/theme context, and dynamic service mounting. It contains zero service-specific UI — all service UIs are loaded at runtime as Module Federation remotes.

The Go backend (`workfort web`) embeds the shell's build output and serves it as the SPA. In development, `--dev` proxies to Vite's dev server.

---

## Responsibilities

The shell does exactly three things:

1. **Chrome** — Top nav with service tabs, sidebar frame, theme toggle. Built with `<wf-*>` web components from `@workfort/ui`.
2. **Auth + Theme context** — `useAuth()` and `useTheme()` from `@workfort/ui-solid`. Shared to remotes via Module Federation `shared` config so all services use the same singleton instances.
3. **Service mounting** — Fetches `/api/services` to discover available services, dynamically loads `remoteEntry.js` for each enabled service with `ui: true`, and mounts the default export component into a content area.

The shell does NOT:
- Render chat, nexus, hive, or any service-specific UI
- Own WebSocket connections (services manage their own)
- Know what services exist at build time

---

## Architecture

### Dynamic Remote Loading

Remotes are discovered at runtime, not declared in the Vite config. This supports hot-adding services without rebuilding the shell.

**Flow:**

1. Shell starts, fetches `GET /api/services`
2. For each service where `enabled: true` and `ui: true` (note: `ui` is a handler-layer field set by probing `/ui/health` at startup — not part of the domain model. Currently hardcoded to `false`; the Go handler needs to implement health probing before remotes will load):
   - Call `registerRemotes()` from `@module-federation/runtime` with the service's remote entry URL
   - Remote entry URL: `/api/{service}/ui/remoteEntry.js` (proxied by Go to the service's `/ui/remoteEntry.js`)
3. Router creates a route for each service using `route` from the services response
4. When the user navigates to a service route, `loadRemote()` loads the remote's default export and mounts it

**Module Federation Shared Dependencies:**

The shell declares these as shared singletons so remotes don't bundle their own copies:

```
@workfort/ui       — singleton, eager (shell provides)
@workfort/ui-solid — singleton, eager (shell provides)
@workfort/auth     — singleton, eager (shell provides)
solid-js           — singleton, eager (shell provides)
@solidjs/router    — singleton, eager (shell provides, remotes participate in shell routing context)
```

### Remote Contract

Each service remote exposes:

| Export | Type | Required | Purpose |
|--------|------|----------|---------|
| `default` | `Component` | Yes | Main content component, mounted in the content area |
| `manifest` | `{ name, label, route, minWidth? }` | Yes | Metadata; overrides `/api/services` values if present |
| `SidebarContent` | `Component` | No | Content for the shell's sidebar frame |
| `HeaderActions` | `Component` | No | Extra buttons/controls in the nav bar |

The shell renders `SidebarContent` and `HeaderActions` when the service is active. Missing optional exports result in empty slots.

### Graceful Degradation

| Scenario | Behavior |
|----------|----------|
| Service `enabled: false` | No tab in nav, route returns 503 from Go |
| Service `ui: false` | Tab shown but grayed out, clicking shows "unavailable" |
| Remote fails to load | Caught by ErrorBoundary, shows `<unavailable>` component |
| No services available | Shell renders chrome with empty content area |

---

## File Structure

```
web/shell/
├── package.json
├── tsconfig.json
├── vite.config.ts
├── index.html
├── src/
│   ├── index.tsx
│   ├── app.tsx
│   ├── lib/
│   │   ├── api.ts
│   │   └── remotes.ts
│   ├── stores/
│   │   ├── services.ts
│   │   └── theme.ts
│   └── components/
│       ├── shell-layout.tsx
│       ├── nav-bar.tsx
│       ├── service-mount.tsx
│       └── unavailable.tsx
```

### File Responsibilities

**`package.json`** — SolidJS app, not a library. Dependencies: `solid-js`, `@solidjs/router`, `@workfort/ui`, `@workfort/ui-solid`, `@workfort/auth`, `@module-federation/runtime`. Dev dependencies: `vite`, `vite-plugin-solid`, `@module-federation/vite`, `typescript`.

**`vite.config.ts`** — Module Federation host config. No remotes declared — they're registered at runtime. Shared deps: `@workfort/ui`, `@workfort/ui-solid`, `@workfort/auth`, `solid-js`, `@solidjs/router` as eager singletons. No dev server proxy needed — in dev mode, the user hits the Go server which proxies to Vite.

**`index.html`** — Minimal HTML. Sets `data-theme="dark"` as a static attribute on `<html>` (avoids flash of unstyled content). Mounts into `<div id="app">`. Component styles and theme tokens are imported via JS in `src/index.tsx` (`import '@workfort/ui/style.css'`) so Vite can resolve the bare specifier.

**`src/index.tsx`** — Imports `@workfort/ui` (registers custom elements), calls `render(() => <App />, document.getElementById('app'))`.

**`src/app.tsx`** — Top-level component. Wraps everything in `<Router>`. Fetches services on mount via `createResource`. Registers remotes. Renders `<ShellLayout>` with dynamic child routes.

**`src/lib/api.ts`** — Typed fetch wrapper. `fetchServices(): Promise<ServicesResponse>` and `fetchConfig(): Promise<ConfigResponse>`. Handles errors, returns typed responses matching the Go endpoint shapes.

**`src/lib/remotes.ts`** — Wraps `@module-federation/runtime`. `initRemotes(services: ServiceInfo[])` calls `registerRemotes()` for each service with `ui: true`. `loadServiceModule(serviceName: string)` calls `loadRemote()` and returns the module's exports. Handles load failures gracefully.

**`src/stores/services.ts`** — SolidJS resource that fetches `/api/services` and exposes the service list as a reactive signal. Used by nav bar and router.

**`src/stores/theme.ts`** — Reads/writes `data-theme` attribute on `<html>` and persists to `localStorage`. Exposes `toggle()` function. The `useTheme()` hook from `@workfort/ui-solid` reads the attribute reactively — this store just manages the toggle.

**`src/components/shell-layout.tsx`** — Three-region layout: nav bar (top), sidebar (left), content area (center). Uses CSS grid. Sidebar renders the active service's `SidebarContent` export if available.

**`src/components/nav-bar.tsx`** — Horizontal bar with: WorkFort branding (left), service tabs built from the services signal (center), theme toggle button (right). Uses `<wf-list>` and `<wf-list-item>` for tabs. Active tab highlighted based on current route.

**`src/components/service-mount.tsx`** — Takes a service name prop. Calls `loadServiceModule()` inside `createResource`. Wraps in `<Suspense>` (shows `<wf-skeleton>`) and `<ErrorBoundary>` (shows `<Unavailable>`). Renders the loaded component.

**`src/components/unavailable.tsx`** — Simple centered message: "{Service Label} is unavailable." Uses `<wf-error-fallback>`.

---

## Routing

```
/                    → Redirect to first enabled service's route
/{service-route}/*   → ServiceMount for that service
```

Routes are generated dynamically from `/api/services`. Each service's `route` field (e.g., `/chat`, `/nexus`) becomes a top-level route. Services own their own sub-routing inside the mounted component.

The shell uses `@solidjs/router` with a catch-all that redirects unknown paths to `/`.

---

## Theme System

The shell sets `data-theme` on `<html>` (`"dark"` or `"light"`). Theme tokens are defined in `@workfort/ui/style.css` as CSS custom properties. Currently only dark mode tokens exist on `:root`. **Prerequisite:** Light mode tokens must be added to `tokens.css` under `[data-theme="light"]` before the theme toggle is functional. Until then, the toggle can exist in the UI but will be a no-op visually.

The shell's theme store manages toggle + localStorage persistence. It mutates the DOM directly via `document.documentElement.setAttribute('data-theme', theme)`. The `useTheme()` hook from `@workfort/ui-solid` observes the `data-theme` attribute via MutationObserver, returning a read-only accessor `() => 'dark' | 'light'` (not a writable signal). The store owns writes, the hook owns reads.

Default: dark.

---

## Auth

The shell imports `@workfort/auth` and initializes the auth client. The `useAuth()` hook from `@workfort/ui-solid` provides `{ user, isAuthenticated }` signals.

Auth state flows through Module Federation's shared config — remotes get the same `@workfort/auth` singleton, so `useAuth()` in a remote returns the same state as in the shell.

The BFF proxy in Go handles the actual cookie-to-JWT conversion. The shell's auth client calls `/api/auth/api/auth/get-session` (better-auth's session endpoint, proxied through Go) to check whether a session exists. The auth client is initialized in `src/index.tsx` before render. Login/logout UI is owned by the auth service's remote.

---

## Dev Workflow

1. Start Vite dev server: `cd web/shell && pnpm dev` (runs on `:5173`)
2. Start Go server: `mise run build && ./build/workfort web --dev --no-open` (runs on `:8080`)
3. Open `http://127.0.0.1:8080`
4. Go handles `/api/*` directly, proxies everything else to Vite at `:5173`

No reverse proxy from Vite back to Go is needed — the browser talks to Go, which owns all routing.

---

## Build + Embed

Production build: `cd web/shell && pnpm build` outputs to `web/shell/dist/`.

The Go embed in `cmd/web/embed.go` currently uses `//go:embed all:placeholder`. For production, the build pipeline copies `web/shell/dist/*` into `cmd/web/dist/`, then the embed directive changes to `//go:embed all:dist` with `fs.Sub(webFS, "dist")`. Go embed paths are relative to the file's directory, so the dist contents must be under `cmd/web/`.

Build sequence: `pnpm --filter ./web/shell build` → copy `web/shell/dist/` to `cmd/web/dist/` → `go build`.

A `mise` task (`build:web`) will chain these. The placeholder directory remains for `go build` without a frontend build.

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `solid-js` | UI framework |
| `@solidjs/router` | Client-side routing |
| `@workfort/ui` | Web components + CSS tokens |
| `@workfort/ui-solid` | `useAuth()`, `useTheme()` hooks |
| `@workfort/auth` | Auth client (better-auth) |
| `@module-federation/runtime` | Dynamic remote loading |
| `vite` | Build tool + dev server |
| `vite-plugin-solid` | SolidJS JSX transform |
| `@module-federation/vite` | MF host plugin for Vite |

---

## Future: Dynamic Service Discovery

The architecture supports hot-adding services without rebuilding the shell. A future enhancement would poll `/api/services` on an interval (or receive push updates) and call `registerRemotes()` for newly discovered services. The router would need to add routes dynamically. The current design doesn't implement polling but doesn't prevent it — `registerRemotes()` is additive.

---

## Constraints

- Shell source lives at `web/shell/`, not in `web/packages/` (it's an app, not a library)
- No service-specific code in the shell
- All styling via `--wf-*` CSS custom properties
- Module Federation remotes loaded at runtime, not build time
- Shell must render usable chrome even when all services are down
