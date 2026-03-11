# WorkFort Web UI — Design Spec

## Overview

A browser-based interface for the WorkFort CLI, served by `workfort web`. It provides a modern, approachable alternative to the terminal TUI for managing all WorkFort platform services: Sharkfin (chat), Nexus (VMs), and Hive (agent provisioning), with extensibility for future services.

## Constraints

- **Cross-platform**: Windows, macOS, Linux.
- **Fully static binary**: `CGO_ENABLED=0`, no dynamic linking. The `web` subcommand must have zero CGo dependencies.
- **Single binary distribution**: The SolidJS frontend is embedded via `go:embed`. No external files at runtime.
- **XDG compliance**: All configuration under `$XDG_CONFIG_HOME/workfort/config.yaml`, following existing conventions.

## Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Frontend framework | SolidJS | Highest runtime performance (near-vanilla JS), smallest runtime (~7KB). Fine-grained reactivity ideal for real-time chat. Validated by prior bundle-size research. |
| CSS | UnoCSS (Tailwind preset) | Tailwind syntax with better tree-shaking. Marginal bundle advantage over Tailwind v4 (19.61 vs 20.07 KB gzipped). |
| Router | @solidjs/router | Standard SolidJS routing. Client-side, SPA model. |
| State | nanostores | Framework-agnostic, lightweight. One store per service concern. |
| Build | Vite | Fast dev server, optimized production builds. Output to `web/dist/`. |
| Go server | net/http + httputil | Standard library only. No third-party HTTP frameworks. Static binary compatible. |
| Embedding | go:embed | `web/dist/` embedded into the Go binary at build time. |

## Visual Design

Inspired by [better-auth.com](https://better-auth.com/).

### Aesthetic

- Minimal, modern, generous spacing.
- No color accents beyond functional indicators (green for online, subtle badges for unread).
- Monospace `WORKFORT` branding in the top nav. All other text in Inter/system sans-serif.
- Subtle borders (`1px solid` at low opacity), no heavy shadows.

### Theme

- **Dark mode** (default): `#09090b` background, `#fafafa` text, zinc grays for borders and secondary text.
- **Light mode**: `#ffffff` background, `#09090b` text, zinc grays (`#e4e4e7` borders, `#71717a` secondary).
- Theme toggle in top nav, persisted to `localStorage`.
- UnoCSS dark mode variant for theme switching.

### Layout

- **Top nav bar**: WORKFORT branding (monospace) → fort name (static label in phase 1) → service tabs (Chat, Nexus, Hive, ...) → user info + theme toggle.
- **Sidebar**: Context-specific per active service. For chat: channels list, DM list. For Nexus: VM list. For Hive: teams/agents.
- **Main content area**: Full remaining width. For chat: channel header, message stream, input box.

## Domain Model

### Fort

A Fort is a named collection of services behind a single gateway. Users can belong to multiple forts (like GitHub organizations) and switch between them.

```go
// internal/domain/web.go

type Fort struct {
    Name     string
    Gateway  string    // Single origin URL for remote forts; empty for local
    Services []Service
}

type Service struct {
    Name     string   // "sharkfin", "nexus", "hive"
    PathBase string   // Derived: "/api/" + Name
    URL      string   // Direct backend URL (local fort only, e.g. "http://127.0.0.1:16000")
    WSPaths  []string // Paths that accept WebSocket upgrade (whitelist)
    Enabled  bool
}

type FortRegistry interface {
    Forts() []Fort
    Active() Fort
    SetActive(name string) error
}
```

`PathBase` is always derived as `"/api/" + service.Name`. It is not user-configurable.

`WSPaths` is a whitelist: only requests to these paths accept WebSocket upgrade. Upgrade requests to other paths for that service are rejected with `400 Bad Request`.

### Fort roles

- **Local fort**: The WorkFort CLI *is* the gateway. `workfort web` starts an HTTP server, serves the SPA, and proxies to local service daemons using each service's `URL` field. There is no separate gateway URL — the web server itself serves that role.
- **Remote fort**: The WorkFort CLI connects *to* an existing gateway at the configured `Gateway` URL. The Go server still proxies — it forwards `/api/*` requests to the remote gateway URL, preserving the `/api/<service>/` prefix (the remote gateway handles its own internal routing). This keeps the frontend identical for local and remote forts. Individual service `URL` fields are not used. Remote fort proxy details are a phase 2 concern.

## Configuration

All config lives in `$XDG_CONFIG_HOME/workfort/config.yaml`, managed by the existing Viper setup in `pkg/config/`.

```yaml
active-fort: local

forts:
  local:
    services:
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
    gateway: "https://fort.acme.com"
    services:
      sharkfin:
        enabled: true
        ws-paths: ["/ws", "/presence"]
      nexus:
        enabled: true
      hive:
        enabled: true
```

Note: the `local` fort has no `gateway` — the CLI itself is the gateway, proxying to each service's `url`. Remote forts like `acme-corp` have a `gateway` and no per-service `url` (the gateway handles routing internally).

### Defaults

Registered in `pkg/config/viper.go`:

```go
viper.SetDefault("active-fort", "local")
viper.SetDefault("forts.local.services.sharkfin.url", "http://127.0.0.1:16000")
viper.SetDefault("forts.local.services.sharkfin.enabled", true)
viper.SetDefault("forts.local.services.sharkfin.ws-paths", []string{"/ws", "/presence"})
viper.SetDefault("forts.local.services.nexus.url", "http://127.0.0.1:9600")
viper.SetDefault("forts.local.services.nexus.enabled", true)
viper.SetDefault("forts.local.services.hive.url", "http://127.0.0.1:17000")
viper.SetDefault("forts.local.services.hive.enabled", true)
```

Overridable via environment variables (`WORKFORT_ACTIVE_FORT`, `WORKFORT_FORTS_LOCAL_SERVICES_SHARKFIN_URL`, etc.).

Note: the existing flat `sharkfin-host` config key used by `workfort chat` is unaffected. The `workfort web` command reads from the `forts.*` namespace. The two config structures coexist — no migration needed. The `workfort chat` command may be updated to read from the fort config in a future phase.

### Adding a new service

Add a config block under `forts.<name>.services.<service-name>`. No Go code changes to the proxy infrastructure. The frontend discovers services dynamically via `GET /api/services`.

### Disabling a service

Set `enabled: false`. The Go proxy returns `503 Service Unavailable` with a JSON body (`{"error": "service disabled", "service": "hive"}`) for requests to that service's path. The frontend shows the service tab grayed out with a message explaining it's disabled.

## Architecture (Go Side — Hexagonal)

Follows the same hexagonal architecture used by Nexus and Hive. This is a new architectural pattern for the WorkFort CLI repository — the existing `internal/chat/` Bubble Tea model remains as-is. The two patterns coexist: the TUI uses Elm architecture via Bubble Tea, while the web subsystem uses hexagonal. No refactoring of `internal/chat/` is planned.

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
      embed.go          # go:embed for web/dist/

cmd/
  web/
    web.go              # Composition root: config → domain → app → infra → serve
```

### Cobra integration

`cmd/web/web.go` exports `NewWebCmd() *cobra.Command`, following the same pattern as `cmd/chat/chat.go`. It is registered in `cmd/root.go`:

```go
rootCmd.AddCommand(web.NewWebCmd())
```

### Dependency flow

```
cmd/web/web.go (composition root)
    │ constructs
    ▼
internal/app.WebService
    │ depends on interfaces from
    ▼
internal/domain (ports + pure types)
    ▲
    │ implemented by
internal/infra/httpapi
```

`cmd/web/web.go` is the only file importing both `app` and `infra`. Domain and app layers never import infra.

### Proxy behavior

- **REST**: `httputil.ReverseProxy` per enabled service. Request to `/api/nexus/v1/vms` strips `/api/nexus` prefix, forwards to the service's configured URL at `/v1/vms`.
- **WebSocket**: Only paths listed in a service's `ws-paths` accept WebSocket upgrade. The proxy detects the `Upgrade: websocket` header and performs a full-duplex connection hijack. Upgrade requests to paths not in `ws-paths` return `400 Bad Request`.
- **SPA fallback**: All non-`/api/*` routes serve `index.html` from the embedded filesystem, enabling client-side routing.
- **Disabled services**: Return `503 Service Unavailable` with a JSON error body.
- **Service metadata**: `GET /api/services` — see API Contracts below.

### Server lifecycle

- **Bind**: `127.0.0.1:8080` (configurable via `--bind`, `--port` flags or `WORKFORT_WEB_BIND`, `WORKFORT_WEB_PORT`).
- **Startup**: Auto-opens browser unless `--no-open` is passed.
- **Shutdown**: Handles `SIGINT` and `SIGTERM`. Initiates graceful shutdown with a 5-second timeout — stops accepting new connections, drains in-flight HTTP requests, closes proxied WebSocket connections.

## Chat WebSocket Integration

The Go proxy is a transparent passthrough — it does not interpret, translate, or mediate the Sharkfin WebSocket protocol. The browser speaks the raw Sharkfin envelope protocol directly (through the proxy).

### Protocol responsibilities

The frontend `ws.ts` implements the full Sharkfin envelope protocol:

1. **Connection**: Opens WebSocket to `/api/sharkfin/ws` (proxied to `ws://127.0.0.1:16000/ws`). Note: the `/presence` path in `ws-paths` is a separate Sharkfin endpoint used by other consumers (e.g., `mcp-bridge`). The web frontend uses only `/ws` — all chat messages and presence events arrive on this single connection.
2. **Hello handshake**: Reads the initial `hello` envelope from the server.
3. **Identity**: Sends `identify` (falling back to `register`) with the configured username. Username comes from the Go server via `GET /api/config` (returns `{"username": "...", "fort": "..."}`).
4. **Heartbeat**: Responds to `heartbeat` envelopes with a heartbeat reply.
5. **Message dispatch**: Parses inbound envelopes by type (`reply`, `message.new`, `presence`) and updates the appropriate nanostore.
6. **Ref tracking**: Maintains a pending refs map to correlate request/reply pairs, same as the Go `sharkfin.Client`.
7. **Reconnect**: On disconnect, reconnects with exponential backoff (1s initial, 30s max), same as the existing Go client.

This mirrors the existing `pkg/sharkfin/client.go` logic but implemented in TypeScript for the browser. The Go proxy adds no abstraction layer — it is a dumb WebSocket pipe.

### Why transparent proxy

- The Sharkfin protocol is simple (JSON envelopes with type/ref/data). No need for a translation layer.
- Keeps the Go server stateless — no session tracking, no message buffering.
- The frontend can evolve independently of the Go proxy when the Sharkfin protocol changes.
- Matches the pattern: the TUI's `pkg/sharkfin/` client speaks the protocol directly too.

## API Contracts

### `GET /api/services`

Returns the active fort's service list. The frontend uses this to render navigation tabs dynamically.

```json
{
  "fort": "local",
  "services": [
    {"name": "sharkfin", "label": "Chat", "route": "/chat", "enabled": true},
    {"name": "nexus", "label": "Nexus", "route": "/nexus", "enabled": true},
    {"name": "hive", "label": "Hive", "route": "/hive", "enabled": false}
  ]
}
```

- `label`: Display name for the nav tab. Derived from a static map in the Go server (service name → label). Unknown services use their name as the label.
- `route`: Client-side route prefix. Derived as `"/" + service.Name` (except `sharkfin` → `/chat`).
- All services are included regardless of `enabled` status. The frontend decides how to render disabled services.

### `GET /api/config`

Returns client configuration needed by the frontend at startup.

```json
{
  "username": "kazw",
  "fort": "local"
}
```

- `username`: From Viper config (`username` key), same source as the TUI.
- `fort`: Active fort name.

## Architecture (Frontend — SolidJS)

### Project structure

```
web/
  src/
    index.tsx               # Entry point
    app.tsx                 # Root component, router setup, theme provider
    components/
      shell/
        top-nav.tsx         # Branding, fort label, service tabs, user, theme toggle
        sidebar.tsx         # Context-specific sidebar per service
        layout.tsx          # Shell composition (nav + sidebar + content)
      chat/
        channel-list.tsx    # Channel sidebar items
        dm-list.tsx         # DM sidebar items
        message-list.tsx    # Scrollable message stream
        message-input.tsx   # Compose box
        channel-header.tsx  # Channel name, description, member count
      nexus/
        placeholder.tsx     # Placeholder for phase 1
      hive/
        placeholder.tsx     # Placeholder for phase 1
      shared/
        loading.tsx         # Loading spinner
        error-banner.tsx    # Dismissable error banner
        empty-state.tsx     # Empty state with icon + message
    stores/
      theme.ts              # Dark/light toggle, persisted to localStorage
      fort.ts               # Active fort, service list from /api/services
      chat.ts               # Channels, messages, presence, unread counts, WebSocket lifecycle
      connection.ts         # WebSocket connection state (connected/connecting/disconnected)
      nexus.ts              # VM state (future)
      hive.ts               # Agent/task state (future)
    lib/
      ws.ts                 # WebSocket client — full Sharkfin envelope protocol
      api.ts                # Typed fetch wrapper for /api/* endpoints
  index.html
  vite.config.ts
  uno.config.ts
  package.json
  tsconfig.json
```

### Routing

```
/                → redirect to /chat
/chat            → channel list + general channel
/chat/:channel   → specific channel view
/chat/dm/:user   → DM conversation
/nexus           → placeholder (phase 1)
/nexus/*         → reserved for VM views
/hive            → placeholder (phase 1)
/hive/*          → reserved for agent/task views
```

### State management

One nanostore per concern:

- **`theme`**: `atom<'dark' | 'light'>`, synced to `localStorage` and `<html>` class.
- **`fort`**: Active fort name + service list fetched from `/api/services`. Drives nav tab rendering.
- **`connection`**: `atom<'connected' | 'connecting' | 'disconnected'>`. Drives connection status UI.
- **`chat`**: Channels, messages, DMs, presence, unread counts. Manages WebSocket connection lifecycle (connect, reconnect with backoff, dispatch incoming messages to the correct sub-store).

### Frontend state patterns

- **Loading**: On initial app boot, a loading spinner renders until `/api/services` and `/api/config` resolve. Per-service views show a loading state until their initial data loads (e.g., channel list for chat).
- **Errors**: A dismissable error banner at the top of the content area for API failures. Non-blocking — the rest of the UI remains interactive.
- **WebSocket disconnection**: When the chat WebSocket disconnects, a persistent banner appears below the top nav: "Reconnecting..." with attempt count. Clears automatically on reconnect. Message input is disabled during disconnection.
- **Empty states**: Channels with no messages show a centered empty state. An empty channel list prompts to create or join a channel. An empty DM list prompts to start a conversation.

### Dynamic service navigation

The top nav service tabs are rendered from the `fort` store's service list, not hardcoded. Disabled services render as grayed-out, non-interactive tabs with a tooltip: "Not enabled. Configure in workfort config." Adding a new service to the config automatically surfaces it in the UI with no frontend code changes.

## Build & Development Workflow

### Toolchain

- **Node.js**: Managed via `mise`. Version specified in `web/.node-version` or `mise.toml`.
- **Package manager**: pnpm (fast, disk-efficient, workspace-friendly).
- **`web/dist/`**: Gitignored. Built on demand, never committed.

### Build commands

```bash
# Full production build (frontend + Go binary)
mise run build:web       # cd web && pnpm install && pnpm build
mise run build           # go build (embeds web/dist/)

# Or manually:
cd web && pnpm install && pnpm build
cd .. && CGO_ENABLED=0 go build -o build/workfort .
```

The `go:embed` directive in `internal/infra/httpapi/embed.go` references `web/dist/`. If `web/dist/` does not exist, the Go build fails. A stub `web/dist/index.html` placeholder is committed to the repo so that `go build` succeeds without a frontend build — useful for contributors working only on the Go side. The `mise run build` task ensures the real frontend is built before the production binary.

### Development mode

During development, the Vite dev server and Go server run concurrently:

```bash
# Terminal 1: Go server (serves API only in dev)
go run . web --dev

# Terminal 2: Vite dev server with proxy
cd web && pnpm dev
```

Vite's `server.proxy` in `vite.config.ts` forwards `/api/*` requests to the Go server (`http://localhost:8080`), avoiding CORS issues. The `--dev` flag on the Go server skips embedded file serving (since Vite handles it).

### Cross-compilation

```bash
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o build/workfort-linux-amd64 .
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -o build/workfort-darwin-arm64 .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o build/workfort-windows-amd64.exe .
```

Frontend assets are platform-agnostic — built once, embedded into all platform binaries.

## Phased Scope

### Phase 1 — Shell + Chat

- App shell: top nav with branding, active fort label (read-only), dynamic service tabs, theme toggle, user display.
- Chat service: channel list, DM list, message stream with scroll, message input, presence indicators, unread counts/badges.
- WebSocket connection through Go proxy with reconnect and status banner.
- Shared components: loading spinner, error banner, empty states.
- Nexus and Hive: routable placeholder views.

### Phase 2 — Expansion

- Interactive fort switcher dropdown in the top nav.
- Browser-based STT via Whisper WASM/WebGPU — replaces the TUI's CGo whisper.cpp dependency. Audio capture via Web Audio API, inference runs client-side. Voice input button in the message compose area.
- Nexus views: VM list, create/start/stop/delete, exec console.
- Hive views: teams, agents, roles, tasks, documents.
- Authentication/authorization for remote forts.

### Not in scope

- Browser push notifications.
- File uploads or rich media.
- Mobile-responsive layout (desktop-first).
