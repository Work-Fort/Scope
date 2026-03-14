# Service Discovery & Frontend Status — Design Spec

## Overview

Dynamic service discovery for the WorkFort web shell. Services are self-describing — the shell only needs a URL to discover everything about a service: its name, label, route, whether it has a frontend, and whether it's connected. The shell surfaces all status and error states visually in the chrome, designed for both technical and non-technical users.

**Future:** The discovery source will migrate from fort config URLs + health probing to a constellation API. The architecture is designed so only the data source changes — the tracker, shell SPA, and status UX stay the same.

**Edge case — constellation outage:** When the constellation endpoint is the discovery source and it goes down, the tracker cannot probe any services. Without special handling, all non-WS services would flip to `connected: false` on the next poll cycle while WS services with open connections stay `connected: true` (ref-counting is local). This creates a misleading split state. The correct behavior: if the discovery source itself is unreachable, the tracker should **freeze the last-known service state** (don't mark everything disconnected) and surface a single system banner: "Cannot reach constellation — service status may be outdated." The tracker already knows the difference between "probed a service and got an error" vs "couldn't reach the discovery source at all." This distinction must be preserved when the constellation migration happens.

**Related specs:**
- `docs/2026-03-12-go-web-shell-design.md` — Go web shell architecture
- `docs/2026-03-11-web-ui-design.md` — full web UI architecture

---

## Self-Describing Services

Services declare their own identity via `pkg/frontend`. The shell never hardcodes metadata about services.

### `pkg/frontend` Manifest

`frontend.Handler()` takes a manifest at construction:

```go
frontend.Handler(fsys, frontend.Manifest{
    Name:    "sharkfin",
    Label:   "Chat",
    Route:   "/chat",
    WSPaths: []string{"/ws", "/presence"},
})
```

The `/ui/health` endpoint returns the manifest:

```json
GET /ui/health → 200
{
  "status": "ok",
  "name": "sharkfin",
  "label": "Chat",
  "route": "/chat",
  "ws_paths": ["/ws", "/presence"]
}
```

When the service has no frontend built (no `remoteEntry.js`):

```json
GET /ui/health → 503
{
  "status": "unavailable"
}
```

The service owns its identity. The shell just reads it.

### Fort Config Simplification

Fort config reduces to a list of URLs. No name, label, route, or ws-paths — those come from the service:

```yaml
forts:
  local:
    local: true
    services:
      - url: "http://127.0.0.1:16000"
      - url: "http://127.0.0.1:9600"
      - url: "http://127.0.0.1:3000"
```

The current config uses a named map (`services.sharkfin.url`). This changes to a list. `fortconfig/registry.go` switches from `viper.GetStringMap` to `viper.UnmarshalKey` to parse the list of service structs.

### Domain Changes

`domain.Service` simplifies. The config only provides the URL. Everything else is discovered:

```go
// ConfigService is what comes from the fort config file — just a URL.
// All configured services are considered enabled. To disable a service,
// remove it from the config.
type ConfigService struct {
    URL string
}

// Fort is a named collection of services.
type Fort struct {
    Name     string
    Local    bool
    Gateway  string
    Services []ConfigService
}
```

The full service state (name, label, route, ui, connected) lives in the tracker, not in the domain types. The domain only models what the config provides.

Delete the hardcoded `serviceMetadata` map from `handler.go`.

### Handler Bootstrap Sequence

The HTTP handler needs service names to register proxy routes, but names aren't known until `/ui/health` is probed. The startup sequence:

1. `ServiceTracker` is created with the list of service URLs
2. Tracker runs an **initial synchronous probe** — blocks until all services are probed once (with the 3s timeout per service)
3. After the initial probe, the tracker has discovered names for reachable services and can build the proxy route map
4. `NewHandler` receives the tracker and registers proxy routes for all discovered services
5. Background polling begins for ongoing health checks
6. Services that were unreachable during initial probe get picked up on subsequent poll cycles — the tracker calls a route registration callback to add new proxy routes dynamically

For services that are down at startup: the handler registers no route for them. When the tracker discovers them later, it registers the proxy route via the mux. Go's `http.ServeMux` supports adding routes after the server starts.

---

## Service Tracker (Go Backend)

A `ServiceTracker` in `internal/infra/httpapi/` that maintains the live state of all services by combining two signals: UI health probing and WebSocket connection state.

### UI Health Probing

- Background goroutine probes each enabled service's `/ui/health` on a fixed interval (default 10s)
- Probes immediately on startup (no waiting for first interval)
- Constructs probe URL directly from the service's configured URL: `service.URL + "/ui/health"`. This works in both local and gateway mode because the config URL always points to where the service is reachable from Scope
- Parses the manifest from the health response to get name, label, route, ws-paths
- Stores results in a `sync.RWMutex`-guarded map
- Respects context cancellation for clean shutdown
- HTTP client uses a short timeout (3s) so a hung service doesn't block the cycle

### Connected Status

Two sources depending on service type:

**WebSocket services** (services whose manifest declares `ws_paths`): The WS proxy reports connection state via callbacks. A single service can have many concurrent WebSocket connections. The tracker reference-counts them — `OnConnect` increments, `OnDisconnect` decrements. `connected = true` when count > 0, `false` when count reaches 0. This means the service stays "connected" as long as at least one WS connection is alive.

**Non-WebSocket services**: Connected mirrors the UI health probe result. If `/ui/health` returns 200, connected = true.

### Conflict Detection

When the prober collects manifests, it checks for collisions before updating the service map:

- **Duplicate name**: Two services at different URLs claiming the same name. First discovered wins; the second is excluded and added to the conflicts list.
- **Duplicate route**: Two services claiming the same route path. Same — first wins, second is excluded.

Conflicts are surfaced in the `/api/services` response, not just logged.

### `/api/services` Response

```json
{
  "fort": "local",
  "services": [
    {
      "name": "sharkfin",
      "label": "Chat",
      "route": "/chat",
      "enabled": true,
      "ui": true,
      "connected": true
    },
    {
      "name": "nexus",
      "label": "Nexus",
      "route": "/nexus",
      "enabled": true,
      "ui": true,
      "connected": false
    },
    {
      "name": "auth",
      "label": "Auth",
      "route": "/auth",
      "enabled": true,
      "ui": false,
      "connected": true
    }
  ],
  "conflicts": [
    {
      "url": "http://127.0.0.1:9700",
      "name": "sharkfin",
      "reason": "duplicate name (already registered from http://127.0.0.1:16000)"
    }
  ]
}
```

### WS Proxy Changes

`NewWSProxy` currently creates a stateless proxy handler. It needs to accept connection state callbacks:

```go
type ConnectionCallbacks struct {
    OnConnect    func(service string)
    OnDisconnect func(service string)
}

func NewWSProxy(url string, paths []string, service string, cb *ConnectionCallbacks) http.Handler
```

The proxy calls `OnConnect` after a successful upstream WebSocket handshake and `OnDisconnect` when the connection closes (for any reason — clean close, error, timeout). The callbacks are optional (nil-safe) for backward compatibility.

---

## Shell SPA Changes

### Services Store — Polling

The services store (`stores/services.ts`) changes from a one-shot `createResource` to a polling loop:

- Fetches `/api/services` on an interval (30s)
- Exposes reactive signals for `services()`, `conflicts()`, and `fortName()`
- When a new service appears with `ui: true` that wasn't previously known, calls `registerRemotes()` incrementally (Module Federation runtime supports this). The current one-shot `initialized` guard in `initRemotes()` is removed — registration happens on every poll that discovers new services, skipping already-registered names.
- Cleanup: clears interval on unmount

**Polling tradeoff:** The backend probes every 10s, the SPA polls every 30s. A service state change can take up to 30s to reach the browser. This is intentional — the SPA poll is lightweight and 30s is fast enough for status updates without being noisy.

### Remote Contract — Connected Prop

Mounted remotes receive their backend's `connected` status as a prop. The remote owns its degradation UX — the shell doesn't prescribe what "disconnected" looks like for each service.

```typescript
export interface ServiceModule {
  default: (props: { connected: boolean }) => any;
  manifest: { name: string; label: string; route: string };
  SidebarContent?: () => any;
  HeaderActions?: () => any;
}
```

`ServiceMount` passes `connected` from the services store to the mounted remote. When the polled value changes, the remote receives the update reactively.

### `ServiceInfo` Type Update

```typescript
export interface ServiceInfo {
  name: string;
  label: string;
  route: string;
  enabled: boolean;
  ui: boolean;
  connected: boolean;
}

export interface ServicesResponse {
  fort: string;
  services: ServiceInfo[];
  conflicts: Conflict[];
}

export interface Conflict {
  url: string;
  name: string;
  reason: string;
}
```

---

## Frontend Status UX

All error and status states surface in the shell chrome. Users should never need to check logs — the shell is the only interface. Designed for both technical and non-technical users.

### Design Tokens

Add semantic status colors to `@workfort/ui` tokens:

```css
:root {
  --wf-error: #ef4444;
  --wf-error-subtle: rgba(239, 68, 68, 0.12);
  --wf-warning: #f59e0b;
  --wf-warning-subtle: rgba(245, 158, 11, 0.12);
  --wf-success: #22c55e;
  --wf-success-subtle: rgba(34, 197, 94, 0.12);
}

[data-theme="light"] {
  --wf-error: #dc2626;
  --wf-error-subtle: rgba(220, 38, 38, 0.08);
  --wf-warning: #d97706;
  --wf-warning-subtle: rgba(217, 119, 6, 0.08);
  --wf-success: #16a34a;
  --wf-success-subtle: rgba(22, 163, 74, 0.08);
}
```

### Shell Layout Changes

The current grid is `"nav nav" / "sidebar content"`. Banners go **above** the nav bar, pushing the entire UI down:

```
grid-template-rows: auto auto 1fr;
grid-template-areas:
  "banners banners"
  "nav     nav"
  "sidebar content";
```

The `banners` row is `auto` — collapses to 0 when no banners are active. When banners appear, they push the nav and everything below it down by their height. Banners span full viewport width, same as the nav.

### `wf-banner` Component

A new component in `@workfort/ui` for surfacing persistent status messages. Renders in the `banners` grid area, above the nav bar.

**Anatomy:**

```
┌─────────────────────────────────────────────────────────┐
│ ● [icon]  Headline text for all users          [▾] [✕]  │
│                                                         │
│   Technical details (collapsed by default)              │
│   URL: http://127.0.0.1:9700                            │
│   Error: duplicate name "sharkfin"                      │
└─────────────────────────────────────────────────────────┘
```

**Variants:** `error`, `warning`, `info`. Styled with corresponding `--wf-error-subtle` / `--wf-warning-subtle` backgrounds and `--wf-error` / `--wf-warning` left border accent (4px solid).

**Behavior:**
- **Dismissible.** User can close the banner. It stays dismissed until the condition changes (e.g., a different conflict appears, or the same service goes down again after recovering). Each banner condition has a stable identity key (e.g., `conflict:{name}` or `disconnected:{name}`) used to track dismissal state.
- **Expandable details.** Headline is always visible — plain language, understandable by anyone. A disclosure toggle reveals technical details (URLs, error codes, conflict specifics) for users who want them.
- **Stacks.** Multiple banners stack vertically if there are multiple issues. Order: errors first, then warnings.
- **Auto-clears.** When the condition resolves (next poll shows the conflict gone or service recovered), the banner removes itself.

**API:**

```
<wf-banner variant="error" dismissible>
  <span slot="headline">Chat service conflict detected</span>
  <span slot="details">
    Two services at different URLs are using the name "sharkfin".
    http://127.0.0.1:9700 was excluded.
    Contact your administrator to resolve.
  </span>
</wf-banner>
```

**Typography:** Headline uses `--wf-font-sans` at `--wf-font-size-sm`, semi-bold. Details use `--wf-font-mono` at `--wf-font-size-xs` for technical content, `--wf-font-sans` for prose.

### `wf-toast` Component

A new component in `@workfort/ui` for transient event notifications. Toasts are for discrete events ("Nexus reconnected", "Config updated") as opposed to banners which are for persistent states ("Chat is down").

**Behavior:**
- **Sticky is configurable.** The `sticky` attribute controls whether the toast requires manual dismissal or auto-dismisses after a timeout. Developers choose per-toast based on their use case. Non-sticky toasts auto-dismiss after a configurable duration (default 5s).
- **Position:** Top-right of the viewport, user-configurable. Stacks downward from the top-right corner, offset from the edge by `--wf-space-lg`.
- **Dismissible.** Every toast has a close button regardless of sticky setting.

**Anatomy:**

```
┌──────────────────────────────────┐
│ ●  Nexus reconnected        [✕] │
└──────────────────────────────────┘
```

**Variants:** Same as banner — `error`, `warning`, `info`, `success`. Uses the same semantic token backgrounds (`--wf-error-subtle`, `--wf-success-subtle`, etc.) with matching left border accent.

**API:**

```
<wf-toast variant="success">Nexus reconnected</wf-toast>
<wf-toast variant="error" sticky>Chat connection lost</wf-toast>
<wf-toast variant="info" duration="8000">Config reloaded</wf-toast>
```

**Toast container:** A `wf-toast-container` element wraps the toast stack. Positioned `fixed` at the user's configured corner. The shell layout renders it once, and toasts are added/removed imperatively via a toast store.

**Shared components:** `wf-banner` and `wf-toast` are part of `@workfort/ui` and available to service remotes. A service frontend can use them inside its own content area for service-specific notifications.

**System banner bar access:** The banner bar above the nav is a shared resource, not exclusive to the shell. The shell uses it for system-level issues (conflicts, unreachable services). When the shell isn't using it, a mounted remote can push banners into it for app-level alerts that warrant top-level visibility. Access is via a banner store — remotes call `addBanner(key, variant, headline, details?)` and `removeBanner(key)`. Shell system banners take priority and render first; app banners render below them.

**Shell integration:** A toast store in `stores/toasts.ts` exposes `addToast(variant, message, options?)` and `dismissToast(id)`. Options include `sticky` and `duration`. The services store calls `addToast` on state transitions — e.g., when a service goes from connected to disconnected, or when a conflict is first detected.

### Nav Tab Status Indicators

Add `wf-status-dot` to each nav tab to reflect connected state:

- **Connected:** `wf-status-dot` with `online` state, recolored to use `--wf-success`
- **Disconnected:** `wf-status-dot` with `offline` state, recolored to use `--wf-error`
- **No UI:** Tab is grayed out (existing behavior), no status dot

The existing `wf-status-dot` states (`online`, `offline`, `away`) are remapped to use the new semantic tokens (`--wf-success`, `--wf-error`, `--wf-warning`) instead of their current hardcoded colors.

These indicators persist regardless of banner dismissal — the user always knows at a glance which services are healthy. The dot is the persistent indicator; the banner is the detailed, dismissible explanation.

### Content Area States

When a user navigates to a service route:

**Service has UI and is connected:** Remote loads normally via `ServiceMount`.

**Service has UI but is disconnected:** Remote still loads (it was loaded previously or can be loaded from cache). The remote receives `connected: false` and handles its own degradation UX.

**Service has UI, disconnected, never loaded:** `ServiceMount` shows a `wf-banner` with variant `warning`: *"Chat is starting up or temporarily unavailable. This page will update automatically when it's ready."*

**Service has no UI:** The nav tab for this service has no status dot and is grayed out. If the user somehow navigates to its route (e.g. direct URL), the existing `Unavailable` component renders.

**Service conflict:** The conflicted service doesn't appear in the nav tabs (it was excluded). The banner in the content area explains what happened.

---

## Summary of Changes

### New files
- `internal/infra/httpapi/tracker.go` — ServiceTracker with background probing and connection state
- `web/packages/ui/src/components/banner.ts` — `wf-banner` component
- `web/packages/ui/src/components/toast.ts` — `wf-toast` component
- `web/packages/ui/src/components/toast-container.ts` — `wf-toast-container` component
- `web/packages/ui/src/styles/banner.css` — banner styles
- `web/packages/ui/src/styles/toast.css` — toast styles
- `web/shell/src/stores/banners.ts` — banner store with add/remove, shared between shell and remotes
- `web/shell/src/stores/toasts.ts` — toast store with add/dismiss

### Modified files
- `pkg/frontend/frontend.go` — `Handler()` takes `Manifest`, `/ui/health` returns manifest JSON
- `internal/domain/web.go` — `Service` struct simplifies to URL + Enabled
- `internal/infra/fortconfig/registry.go` — reads simplified config (list of URLs)
- `internal/infra/httpapi/handler.go` — delete `serviceMetadata`, wire ServiceTracker into services handler
- `internal/infra/httpapi/ws.go` — `NewWSProxy` accepts `ConnectionCallbacks`
- `web/packages/ui/src/styles/tokens.css` — add `--wf-error`, `--wf-warning`, `--wf-success` tokens
- `web/packages/ui/src/index.ts` — register `wf-banner`, `wf-toast`, `wf-toast-container`
- `web/shell/src/lib/api.ts` — add `connected`, `conflicts` to types
- `web/shell/src/lib/remotes.ts` — `ServiceModule.default` takes `{ connected }` prop
- `web/shell/src/stores/services.ts` — polling loop, expose `conflicts()`
- `web/shell/src/components/service-mount.tsx` — pass `connected` prop to remote
- `web/shell/src/components/nav-bar.tsx` — status dots use `--wf-error`/`--wf-success`
- `web/shell/src/components/shell-layout.tsx` — render banners above nav bar, add toast container
- `web/shell/src/global.css` — update grid to include banners row above nav
