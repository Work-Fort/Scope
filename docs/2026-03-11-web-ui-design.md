# WorkFort Web UI — Design Spec

## Overview

A browser-based micro-frontend shell for the WorkFort CLI, served by `workfort web`. The shell provides theming, authentication context, routing, and a headless component library (`@workfort/ui`). Each platform service (Sharkfin, Nexus, Hive, auth) owns and serves its own UI as a Module Federation remote. The shell loads service UIs at runtime and renders them in standardized slots.

The WorkFort CLI is purely a UI shell and proxy — it runs no backend logic. Authentication, chat, VM management, and all other business logic live in their respective services, which may run locally (in Nexus VMs) or remotely (behind a gateway).

## Constraints

- **Cross-platform**: Windows, macOS, Linux.
- **Fully static binary**: `CGO_ENABLED=0`, no dynamic linking. The shell has zero backend dependencies — it only proxies and serves the SPA.
- **Single binary distribution**: The shell SPA is embedded via `go:embed`. No external files at runtime.
- **XDG compliance**: All configuration under `$XDG_CONFIG_HOME/workfort/config.yaml`.
- **Service independence**: Each service owns its UI. The shell never contains service-specific UI code.
- **Apache 2.0 licensing**: The WorkFort CLI (including `pkg/auth/` and `@workfort/ui`) is Apache 2.0 licensed. Service remotes can be any license (proprietary, GPL, MIT).
- **BFF proxy**: The proxy is NOT transparent for auth — it converts session cookies to JWTs for service routes (`/api/{service}/*`). Auth routes (`/api/auth/*`) are pass-through. See the [service auth design spec](2026-03-11-service-auth-design.md) for details.

## Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Frontend framework | SolidJS (required for all service UIs) | Highest runtime performance, smallest runtime (~7KB). Shared across shell and all remotes via Module Federation. |
| CSS | UnoCSS (Tailwind preset) | Tailwind syntax with better tree-shaking. Shell controls all visual tokens via CSS custom properties. |
| Module Federation | @module-federation/vite | Runtime loading of service UI remotes. Dependency sharing (SolidJS, @workfort/ui) as singletons. |
| Router | @solidjs/router | Client-side SPA routing. Shell owns top-level routes, services own sub-routes. |
| State | nanostores | Framework-agnostic, lightweight. Shell stores for theme, fort config, auth. |
| Build | Vite | Fast dev server, optimized production builds. Both shell and service remotes use Vite. |
| Component library | @workfort/ui | Headless structural components. Published as npm package for types; resolved at runtime from shell via federation shared scope. |
| Go server | net/http + httputil | Standard library only. Proxy + embedded SPA serving. No business logic. |
| Embedding | go:embed | Shell SPA embedded into the Go binary at build time. |
| Auth | better-auth | TypeScript-native auth framework. Runs as a service (not in the CLI). Email + password + OAuth. SQLite storage. |

## Architecture

### Three layers

```
Browser
  │
  ▼
WorkFort CLI (Go binary)
  ├── Serves embedded shell SPA
  ├── Proxies /api/auth/*     → auth service
  ├── Proxies /api/sharkfin/* → Sharkfin (API + WS + UI bundle)
  ├── Proxies /api/nexus/*    → Nexus (API + UI bundle)
  ├── Proxies /api/hive/*     → Hive (API + UI bundle)
  └── Serves /api/services, /api/config (shell's own endpoints)

Service Remotes (loaded via Module Federation)
  ├── Sharkfin: chat UI (channels, messages, presence)
  ├── Nexus: VM management UI
  ├── Hive: agent provisioning UI
  └── Auth: login/register UI (special — loaded before other remotes)
```

### Shell responsibilities

The shell owns:
- **Theme system**: CSS custom properties on `:root`, dark/light toggle, `localStorage` persistence
- **`@workfort/ui`**: Headless component library — structural primitives themed by the shell's CSS variables
- **Module Federation host**: Discovers, loads, and mounts service remotes at runtime
- **Top nav**: Branding, fort label, service tabs (from `/api/services`), user info, theme toggle
- **Sidebar frame**: Chrome around service-provided sidebar content
- **Routing**: Top-level route per service; services own sub-routes within their prefix
- **Auth context**: Reads session from auth service, provides `useAuth()` to remotes
- **Proxy layer**: Transparent HTTP/WebSocket proxy to services

The shell does NOT own:
- Any service-specific UI (chat messages, VM lists, agent views)
- Any backend logic (auth, data storage, message handling)
- Any service-specific state management

### Service remote responsibilities

Each service:
- Builds a Vite Module Federation remote
- Serves its UI bundle from its own HTTP server (e.g., `/ui/remoteEntry.js`)
- Exports components and metadata for the shell to mount
- Uses `@workfort/ui` for structural components (inherits shell theme)
- Uses `useAuth()` from `@workfort/ui/auth` for authenticated user context
- Manages its own state (nanostores, local signals, WebSocket connections)
- Never sets colors, fonts, or spacing directly — only uses themed components

### Gateway abstraction

The shell's proxy layer abstracts the routing based on the fort's `local` flag:
- **`local: true`**: CLI proxies directly to each service's configured URL (e.g., `http://127.0.0.1:16000`)
- **`local: false`**: CLI proxies all traffic through the fort's `gateway` URL
- **Future**: An inlet/outlet gateway (VPN-like tunnel) replaces direct proxy for production

The frontend and service remotes are identical in both modes — they always talk to `/api/{service}/*` relative paths.

## Visual Design

Inspired by [better-auth.com](https://better-auth.com/).

### Aesthetic

- Minimal, modern, generous spacing.
- No color accents beyond functional indicators (green for online, subtle badges for unread).
- Monospace `WORKFORT` branding in the top nav. All other text in Inter/system sans-serif.
- Subtle borders (`1px solid` at low opacity), no heavy shadows.

### Theme system

The shell defines all visual tokens as CSS custom properties:

```css
:root {
  /* Set by shell based on dark/light mode */
  --wf-bg: #09090b;
  --wf-bg-secondary: #18181b;
  --wf-text: #fafafa;
  --wf-text-secondary: #a1a1aa;
  --wf-text-muted: #71717a;
  --wf-border: rgba(255, 255, 255, 0.06);
  --wf-accent: #22c55e;

  /* Spacing scale */
  --wf-space-xs: 4px;
  --wf-space-sm: 8px;
  --wf-space-md: 16px;
  --wf-space-lg: 24px;
  --wf-space-xl: 32px;

  /* Typography */
  --wf-font-sans: 'Inter', system-ui, sans-serif;
  --wf-font-mono: ui-monospace, monospace;
  --wf-font-size-xs: 11px;
  --wf-font-size-sm: 12px;
  --wf-font-size-base: 13px;
  --wf-font-size-lg: 14px;

  /* Radii */
  --wf-radius-sm: 4px;
  --wf-radius-md: 6px;
  --wf-radius-lg: 8px;
}
```

- **Dark mode** (default): `#09090b` background, `#fafafa` text, zinc grays.
- **Light mode**: Inverted zinc grays:
  ```css
  :root.light {
    --wf-bg: #ffffff;
    --wf-bg-secondary: #f4f4f5;
    --wf-text: #09090b;
    --wf-text-secondary: #52525b;
    --wf-text-muted: #a1a1aa;
    --wf-border: rgba(0, 0, 0, 0.08);
    --wf-accent: #16a34a;
  }
  ```
- Theme toggle in top nav, persisted to `localStorage`.
- Toggle swaps the CSS custom property values on `:root` — all service UIs update reactively.
- Services never reference colors directly — they use `@workfort/ui` components which consume these variables.

### Layout

- **Top nav bar**: WORKFORT branding (monospace) → fort name → service tabs (dynamic from `/api/services`) → user info + theme toggle.
- **Sidebar frame**: Shell-owned chrome; content provided by the active service's `SidebarContent` export.
- **Main content area**: Full remaining width. Filled by the active service's default export.

## Module Federation Contract

### Remote entry

Each service exposes a federation remote served from its HTTP server. The shell discovers the remote URL from the service config and loads it at runtime.

```
Service serves:  /ui/remoteEntry.js     (federation entry point)
                 /ui/assets/*           (chunks, CSS, etc.)
Shell proxies:   /api/{service}/ui/*  → service's /ui/*
```

### Required exports

```ts
// Every service remote must export:

// Required — main content component
export default function MainView(): JSX.Element;

// Required — metadata for the shell
export const manifest: {
  name: string;        // Service identifier (matches config key)
  label: string;       // Display name in nav tab
  route: string;       // Client-side route prefix (e.g., "/chat")
  minWidth?: number;   // Minimum main content width in px (default: 300)
};
```

### Optional exports

```ts
// Optional — sidebar content
export function SidebarContent(): JSX.Element;

// Optional — header action buttons
export function HeaderActions(): JSX.Element;
```

### Shell slots

The shell provides four mount points:

1. **Main content** — `default` export, fills the primary area right of the sidebar
2. **Sidebar** — `SidebarContent`, rendered inside the shell's sidebar frame
3. **Header actions** — `HeaderActions`, rendered in the top nav
4. **Manifest** — metadata for routing, nav tabs, and graceful degradation

### Graceful degradation

- **Service down**: If a remote can't be loaded (network error, service not running), the shell shows a placeholder: "[Service Name] is unavailable."
- **Min width violation**: If the viewport can't satisfy `minWidth`, the component renders an error message instead of breaking layout.
- **Missing optional exports**: If `SidebarContent` or `HeaderActions` aren't exported, those slots are simply empty.
- **Load errors**: Caught by `<Suspense>` + `<ErrorBoundary>`. Shell never crashes from a service failure.

## @workfort/ui — Headless Component Library

Published as an npm package for type safety and dev-time autocomplete. At runtime, Module Federation resolves it from the shell as a singleton — no duplication.

### Distribution

```ts
// Shell's Vite federation config:
shared: {
  '@workfort/ui': { singleton: true, eager: true },
  'solid-js': { singleton: true, eager: true },
}

// Service's Vite federation config:
shared: {
  '@workfort/ui': { singleton: true, import: false },
  'solid-js': { singleton: true, import: false },
}
```

Services install `@workfort/ui` as a dev dependency for types. At runtime, the shell provides the actual implementation.

### Component principles

- **Structural, not styled**: Components define layout and behavior. Visual appearance comes from CSS custom properties set by the shell.
- **Themed via CSS variables**: All colors, spacing, fonts reference `--wf-*` variables. Components never hardcode visual values.
- **Graceful degradation**: Components that can't render at their minimum size show a clear error message.
- **Accessible**: Keyboard navigation, ARIA attributes, focus management built in.

### Component inventory (phase 1)

```
@workfort/ui
  ├── Panel          — container with optional label, min-width enforcement
  ├── List           — scrollable list with selection state
  ├── List.Item      — list item with active state, leading/trailing slots
  ├── TextInput      — single-line input with placeholder
  ├── Badge          — small count indicator
  ├── Button         — action trigger (text, icon, or both)
  ├── Divider        — horizontal separator
  ├── ScrollArea     — scrollable container with themed scrollbar
  ├── Skeleton       — loading placeholder
  ├── StatusDot      — online/offline indicator
  └── ErrorBoundary  — catches render errors, shows fallback message

@workfort/ui/theme
  ├── ThemeProvider   — provides theme context to component tree
  ├── useTheme()      — reads current theme (dark/light) and tokens
  └── themeTokens     — typed token definitions

@workfort/ui/auth
  ├── AuthProvider    — provides auth context (session, user)
  ├── useAuth()       — reads authenticated user, session state
  └── RequireAuth     — wrapper that redirects to login if unauthenticated
```

### Usage example

```tsx
// In Sharkfin's UI code:
import { Panel, List, Badge, StatusDot, TextInput } from "@workfort/ui";
import { useAuth } from "@workfort/ui/auth";

export function SidebarContent() {
  const { user } = useAuth();
  // ... fetch channels, DMs

  return (
    <>
      <Panel label="Channels">
        <List items={channels()} renderItem={(ch) => (
          <List.Item
            active={ch.name === activeChannel()}
            onClick={() => selectChannel(ch.name)}
            trailing={ch.unread > 0 && <Badge count={ch.unread} />}
          >
            # {ch.name}
          </List.Item>
        )} />
      </Panel>
      <Panel label="Direct Messages">
        <List items={dms()} renderItem={(dm) => (
          <List.Item
            leading={<StatusDot online={dm.online} />}
            onClick={() => openDM(dm.username)}
          >
            {dm.username}
          </List.Item>
        )} />
      </Panel>
    </>
  );
}
```

## Authentication

### Auth as a service

Authentication is a service in the fort config, not shell infrastructure. The shell proxies `/api/auth/*` like any other service.

```yaml
forts:
  local:
    services:
      auth:
        url: "http://127.0.0.1:3000"
        enabled: true
      sharkfin:
        url: "http://127.0.0.1:16000"
        enabled: true
```

### better-auth

The auth service runs better-auth (TypeScript-native). It handles:
- Email + password sign-up/sign-in
- OAuth providers (GitHub, Google — configurable)
- Session management (cookie-based)
- SQLite storage

How the auth service is deployed (Nexus VM, container, etc.) is a Nexus concern, not a shell concern. The shell only needs the URL.

### Auth flow

1. Shell boots, fetches `/api/auth/session` to check for an existing session
2. No session → shell renders the auth service's login UI (loaded via Module Federation, or a minimal built-in fallback)
3. User authenticates → better-auth sets a session cookie
4. Shell re-reads session, populates `AuthProvider` context
5. Service remotes use `useAuth()` to access the authenticated user
6. All proxied API calls carry the session cookie automatically (same-origin)

### Auth remote as a special case

The auth service can ship a Module Federation remote like any other service (login form, account settings, OAuth callback pages). However, since auth must be available before other remotes load, the shell includes a minimal fallback login form as a built-in. If the auth service's UI remote loads successfully, it replaces the fallback. If the auth backend itself is unreachable (`/api/auth/session` fails), the shell shows an error page: "Authentication service unavailable. Check that the auth service is running."

## Domain Model

### Fort

A Fort is a named collection of services. Users can belong to multiple forts and switch between them. Whether the CLI proxies directly to service URLs or through a remote gateway is determined by the `Local` flag — not by the fort's name.

```go
// internal/domain/web.go

type Fort struct {
    Name     string
    Local    bool      // true = CLI proxies directly to each service URL
                       // false = CLI proxies through Gateway
    Gateway  string    // Single origin URL (only used when Local is false)
    Services []Service
}

type Service struct {
    Name     string   // "auth", "sharkfin", "nexus", "hive"
    PathBase string   // Derived: "/api/" + Name
    URL      string   // Direct backend URL (only used when fort is local)
    WSPaths  []string // Paths that accept WebSocket upgrade (whitelist)
    Enabled  bool
}

type FortRegistry interface {
    Forts() []Fort
    Active() Fort
    SetActive(name string) error
}
```

`PathBase` is always derived as `"/api/" + service.Name`. Not user-configurable.

`WSPaths` is a whitelist: only these paths accept WebSocket upgrade. Others return `400`. Paths are matched against the suffix after the `/api/{service}` prefix is stripped — e.g., `/api/sharkfin/presence` is matched against `"/presence"` in the whitelist.

### Fort modes

The `local` flag controls how the CLI routes traffic:

- **`local: true`**: The CLI proxies directly to each service's `url`. Each service must have a `url` configured. No gateway needed.
- **`local: false`**: The CLI proxies all traffic through the fort's `gateway` URL. Individual service `url` fields are ignored — the gateway handles internal routing.

## Configuration

All config lives in `$XDG_CONFIG_HOME/workfort/config.yaml`, managed by existing Viper setup.

```yaml
active-fort: local

forts:
  local:
    local: true       # CLI proxies directly to each service URL
    services:
      auth:
        url: "http://127.0.0.1:3000"
        enabled: true
      sharkfin:
        url: "http://127.0.0.1:16000"
        enabled: true
        ws-paths: ["/ws", "/presence"]
      nexus:
        url: "http://127.0.0.1:9600"
        enabled: true
      hive:
        url: "http://127.0.0.1:17000"
        enabled: false

  acme-corp:
    local: false      # CLI proxies through the gateway
    gateway: "https://fort.acme.com"
    services:
      auth:
        enabled: true
      sharkfin:
        enabled: true
        ws-paths: ["/ws", "/presence"]
      nexus:
        enabled: true
      hive:
        enabled: true
```

### Defaults

```go
viper.SetDefault("active-fort", "local")
viper.SetDefault("forts.local.local", true)
viper.SetDefault("forts.local.services.auth.url", "http://127.0.0.1:3000")
viper.SetDefault("forts.local.services.auth.enabled", true)
viper.SetDefault("forts.local.services.sharkfin.url", "http://127.0.0.1:16000")
viper.SetDefault("forts.local.services.sharkfin.enabled", true)
viper.SetDefault("forts.local.services.sharkfin.ws-paths", []string{"/ws", "/presence"})
viper.SetDefault("forts.local.services.nexus.url", "http://127.0.0.1:9600")
viper.SetDefault("forts.local.services.nexus.enabled", true)
viper.SetDefault("forts.local.services.hive.url", "http://127.0.0.1:17000")
viper.SetDefault("forts.local.services.hive.enabled", false)
```

### Service discovery for Module Federation

The shell needs to know where each service's UI remote lives. This is derived from the proxy path:

```
Remote entry URL: /api/{service}/ui/remoteEntry.js
```

The shell fetches `/api/services` to get the list of enabled services, then attempts to load each service's remote entry. If a service doesn't serve a UI remote (or is down), the shell shows a placeholder.

### Adding a new service

1. Add a config block under `forts.<name>.services.<service-name>`
2. The service must serve `/ui/remoteEntry.js` from its HTTP server
3. The shell discovers it via `/api/services` and loads its remote
4. No changes to the shell codebase

### Disabling a service

Set `enabled: false`. The proxy returns `503`. The nav tab renders grayed out.

## Architecture (Go Side — Hexagonal)

Same hexagonal architecture used by Nexus and Hive. The existing `internal/chat/` Bubble Tea model remains as-is.

### Layer structure

```
internal/
  domain/
    web.go              # Fort, Service, FortRegistry — port interfaces + pure types

  app/
    web_service.go      # Application service, holds domain interfaces only

  infra/
    httpapi/
      handler.go        # Primary adapter: serves SPA, registers proxy routes
      proxy.go          # Reverse proxy construction per service
      embed.go          # go:embed for shell SPA
    fortconfig/
      registry.go       # FortRegistry implementation backed by Viper

cmd/
  web/
    web.go              # Composition root: config → domain → app → infra → serve
```

### Cobra integration

`cmd/web/web.go` exports `NewWebCmd() *cobra.Command`, registered in `cmd/root.go`. As a prerequisite, the chat command import must be gated behind `//go:build cgo` (in a new `cmd/root_cgo.go`) so that `CGO_ENABLED=0` builds exclude it — `pkg/stt` has CGo dependencies that prevent static linking. This gating does not exist in the current codebase and must be added as part of the implementation plan.

### Proxy behavior

- **REST**: `httputil.ReverseProxy` per enabled service. `/api/nexus/v1/vms` strips `/api/nexus`, forwards to service at `/v1/vms`.
- **WebSocket**: Only paths in `ws-paths` accept upgrade. Others return `400`.
- **UI assets**: `/api/{service}/ui/*` proxied to the service's `/ui/*`.
- **SPA fallback**: All non-`/api/*` routes serve `index.html` from embedded filesystem.
- **Disabled services**: Return `503` with JSON error body.

### Server lifecycle

- **Bind**: `127.0.0.1:8080` (configurable via `--bind`, `--port`).
- **Startup**: Auto-opens browser unless `--no-open`.
- **Shutdown**: Graceful on SIGINT/SIGTERM, 5-second drain timeout.

## API Contracts (Shell)

### `GET /api/services`

```json
{
  "fort": "local",
  "services": [
    {"name": "auth", "label": "Auth", "route": "/auth", "enabled": true, "ui": true},
    {"name": "sharkfin", "label": "Chat", "route": "/chat", "enabled": true, "ui": true},
    {"name": "nexus", "label": "Nexus", "route": "/nexus", "enabled": true, "ui": true},
    {"name": "hive", "label": "Hive", "route": "/hive", "enabled": false, "ui": false}
  ]
}
```

- `ui`: Whether the service's Module Federation remote loaded successfully. The shell probes each service's `/ui/remoteEntry.js` at startup and caches the result.
- `label` and `route` are derived from a static map in Go:

```go
var serviceMetadata = map[string]struct{ Label, Route string }{
    "auth":     {"Auth", "/auth"},
    "sharkfin": {"Chat", "/chat"},
    "nexus":    {"Nexus", "/nexus"},
    "hive":     {"Hive", "/hive"},
}
```

Once a service's Module Federation remote loads, the manifest's `label` and `route` take precedence over this static map. The static map exists so that `/api/services` can return metadata before remotes are loaded (e.g., for rendering disabled service tabs). Overridable in future phases.

### `GET /api/config`

```json
{
  "fort": "local"
}
```

Note: `username` is no longer in the config response — it comes from the auth session.

## Frontend Architecture (Shell SPA)

### Project structure

```
web/
  src/
    index.tsx               # Entry point
    app.tsx                 # Root: auth gate → router → shell layout
    components/
      shell/
        top-nav.tsx         # Branding, fort label, service tabs, user, theme toggle
        sidebar.tsx         # Sidebar frame (chrome only, content from remote)
        layout.tsx          # Shell composition (nav + sidebar + content)
        service-loader.tsx  # Loads a Module Federation remote into shell slots
        login-fallback.tsx  # Minimal login form (used if auth remote unavailable)
    stores/
      theme.ts              # Dark/light toggle, CSS custom property management
      fort.ts               # Active fort, service list, remote loading state
      auth.ts               # Session state from better-auth client
    lib/
      api.ts                # Typed fetch wrapper for shell's /api/* endpoints
      federation.ts         # Dynamic remote loading utilities
  index.html
  vite.config.ts
  uno.config.ts
  package.json
  tsconfig.json
```

### Routing

```
/                → redirect to first enabled service route
/login           → auth service's login UI (or built-in fallback)
/chat/*          → Sharkfin remote's default export
/nexus/*         → Nexus remote's default export
/hive/*          → Hive remote's default export
/auth/*          → Auth remote (account settings, etc.)
```

Routes are registered dynamically from `/api/services`. The shell doesn't hardcode service routes.

### Auth gate

```tsx
// app.tsx (simplified)
function App() {
  const auth = useAuth();

  return (
    <Show when={auth.session()} fallback={<LoginPage />}>
      <Router>
        <Route path="/" component={() => <Navigate href={firstEnabledRoute()} />} />
        <Route path="/:service/*" component={ServiceView} />
      </Router>
    </Show>
  );
}
```

`ServiceView` reads the `:service` param, looks up the corresponding Module Federation remote, and mounts it into the shell layout.

### State management

- **`theme`**: `atom<'dark' | 'light'>`. Syncs to `localStorage` and swaps CSS custom properties on `:root`.
- **`fort`**: Service list from `/api/services`. Drives nav tabs and remote loading.
- **`auth`**: Session state from better-auth client (`createAuthClient` from `better-auth/solid`). Provides `useAuth()` via context.

Service-specific state (channels, messages, VMs, etc.) lives inside each service remote — the shell doesn't manage it.

## Build & Development Workflow

### Shell build

```bash
# Build shell SPA
cd web && pnpm install && pnpm build

# Build Go binary with embedded shell
CGO_ENABLED=0 go build -o build/workfort .
```

### Service remote build (example: Sharkfin)

Each service builds its own remote. This is documented in each service's repo, not here. The general pattern:

```bash
cd sharkfin/web && pnpm build
# Output: dist/remoteEntry.js + dist/assets/*
# Served by Sharkfin at /ui/*
```

### Development mode

```bash
# Terminal 1: Go proxy (proxies to Vite dev server instead of serving embedded files)
go run . web --dev    # Proxies non-/api/* routes to localhost:5173 (Vite default)

# Terminal 2: Shell dev server
cd web && pnpm dev

# Terminal 3+: Service dev servers (each with their own Vite + federation)
cd ../sharkfin/web && pnpm dev   # serves remote at localhost:5174
```

Vite's `server.proxy` forwards `/api/*` to the Go server. The shell's federation config in dev mode points to service dev servers directly.

### Cross-compilation

```bash
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o build/workfort-linux-amd64 .
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -o build/workfort-darwin-arm64 .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o build/workfort-windows-amd64.exe .
```

## Phased Scope

### Phase 1 — Shell + Federation Infrastructure

- Shell SPA: top nav, sidebar frame, theme system, CSS custom properties
- `@workfort/ui` component library (headless, themed): Panel, List, TextInput, Badge, Button, StatusDot, ScrollArea, Skeleton, ErrorBoundary
- Module Federation host: dynamic remote loading, graceful degradation
- Auth integration: better-auth client, session gate, fallback login form
- Go proxy: service routing, WebSocket upgrade, SPA fallback, embedded shell
- Fort config: service discovery, `/api/services`, `/api/config`
- Build pipeline: mise tasks for shell build + Go binary

Phase 1 does NOT include any service remote UIs — those are built in their respective repos. Phase 1 delivers the shell infrastructure that service remotes plug into.

### Phase 2 — Service Remotes + Expansion

- Sharkfin remote: chat UI (channels, DMs, messages, presence, WebSocket)
- Nexus remote: VM management (list, create, start/stop, exec console)
- Hive remote: agent provisioning (teams, agents, roles, tasks)
- Auth remote: account settings, OAuth provider management
- Interactive fort switcher dropdown
- Inlet/outlet gateway integration
- Browser-based STT via Whisper WASM/WebGPU

### Not in scope

- Browser push notifications
- File uploads or rich media
- Mobile-responsive layout (desktop-first)
