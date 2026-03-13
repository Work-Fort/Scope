# Go Web Shell — Design Spec

## Overview

The Go-side infrastructure for serving the WorkFort web UI. Three concerns:

1. **`pkg/frontend/`** — shared package that any Go service imports to embed and serve its Vite-built Module Federation remote
2. **Shell proxy** — BFF proxy in the CLI that routes browser traffic to services, converting session cookies to JWTs
3. **`cmd/web/`** — the `workfort web` command that wires everything together

This spec covers the Go implementation only. The frontend SPA and `@workfort/ui` component library are covered in separate specs.

**Related specs:**
- `docs/2026-03-11-web-ui-design.md` — full web UI architecture (frontend + Go)
- `docs/2026-03-11-service-auth-design.md` — auth system design (BFF pattern, JWT flow)
- `docs/2026-03-11-better-auth-setup.md` — better-auth server setup requirements

## Constraints

- **Standard library HTTP** — `net/http` and `httputil` for HTTP. `gorilla/websocket` (already in `go.mod`) for WebSocket proxying. No third-party HTTP frameworks.
- **Static binary** — `CGO_ENABLED=0` produces a working binary. No CGo dependencies exist in the codebase.
- **Single binary** — the shell SPA is embedded via `go:embed`. No external files at runtime.
- **BFF pattern** — the browser never sees a JWT. Session cookies are converted server-side.
- **Shell-only BFF** — the cookie-to-JWT conversion lives in the shell's proxy code, not in a shared package. Only the CLI does BFF.

---

## `pkg/frontend/` — Shared Frontend Serving

A small, focused package that any Go service imports to serve its Module Federation remote. Services embed their Vite build output and call one function.

### API

```go
package frontend

import "io/fs"

// Handler returns an http.Handler that serves an embedded Vite build
// as a Module Federation remote.
//
// Routes registered:
//   /ui/health           — 200 if remoteEntry.js exists, 503 if not
//   /ui/assets/*         — immutable content-hashed chunks (1yr cache)
//   /ui/remoteEntry.js   — federation entry point (no-cache)
//   /ui/*                — everything else (no-cache)
//
// The fsys must be rooted at the Vite build output directory
// (e.g., the result of fs.Sub(embedFS, "web/dist")).
func Handler(fsys fs.FS) http.Handler
```

### Usage

```go
//go:embed web/dist
var webFS embed.FS

func main() {
    distFS, _ := fs.Sub(webFS, "web/dist")
    mux.Handle("/ui/", frontend.Handler(distFS))
}
```

### Cache Headers

| Path | `Cache-Control` | Rationale |
|------|-----------------|-----------|
| `/ui/assets/*` | `public, max-age=31536000, immutable` | Vite content-hashes filenames — safe to cache forever |
| `/ui/remoteEntry.js` | `no-cache` | Federation entry point — must revalidate to pick up new deployments |
| `/ui/*` (other) | `no-cache` | Default safe policy for non-hashed files |

### Health Probe

`GET /ui/health` checks whether `remoteEntry.js` exists in the embedded filesystem:

- Exists: `200 {"status":"ok"}`
- Missing: `503 {"status":"unavailable"}`

The shell's `/api/services` endpoint probes each service's health at startup to set the `ui` availability flag.

### Content-Type

The handler sets `Content-Type` based on file extension. It uses `http.FileServer` semantics (Go's `mime` package) — no manual mapping needed.

---

## Domain Layer

Pure types and a port interface. No dependencies.

```go
// internal/domain/web.go

// Fort is a named collection of services.
type Fort struct {
    Name     string
    Local    bool       // true = proxy direct to each service URL
                        // false = proxy through Gateway
    Gateway  string     // single origin URL (only when Local is false)
    Services []Service
}

// Service is a backend service in a fort.
type Service struct {
    Name     string     // "auth", "sharkfin", "nexus", "hive"
    URL      string     // direct backend URL (only when fort is local)
    WSPaths  []string   // paths that accept WebSocket upgrade (whitelist)
    Enabled  bool
}

// FortRegistry reads fort configuration.
type FortRegistry interface {
    Forts() []Fort
    Active() Fort
    SetActive(name string) error
}
```

### Design decisions

- **No `PathBase` field.** The web UI design spec includes `PathBase` as a derived field (`"/api/" + Name`). This spec drops it — since it's always derived and never configurable, the proxy constructs it at route registration time. This is a deliberate simplification from the web UI spec.
- **No `internal/app/` layer.** The web UI design spec includes `internal/app/web_service.go`. This spec drops it — the proxy has no application logic worth abstracting; `cmd/web/` wires domain and infra directly. If application logic emerges later (e.g., fort switching with side effects), this layer can be introduced then.
- **`URL` only meaningful when `Fort.Local` is true.** When false, the gateway handles internal routing — individual service URLs are irrelevant.
- **`FortRegistry` is read-heavy.** `SetActive` is the only mutation (for future fort switching in the UI). Write-through to Viper config.

---

## Fort Config — Viper-backed Registry

```go
// internal/infra/fortconfig/registry.go

type Registry struct{}

func New() *Registry

func (r *Registry) Forts() []Fort
func (r *Registry) Active() Fort
func (r *Registry) SetActive(name string) error
```

Reads from Viper directly — Viper is already initialized in `cmd/root.go`. Walks `forts.<name>.services.<svc>` keys and builds `Fort`/`Service` structs.

Compile-time interface check:

```go
var _ domain.FortRegistry = (*Registry)(nil)
```

### Default configuration

Added in `cmd/root.go`'s init:

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

### Config file format

```yaml
active-fort: local

forts:
  local:
    local: true
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
    local: false
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

---

## HTTP Handler — Proxy, BFF Auth, SPA Serving

### File structure

```
internal/infra/httpapi/
  handler.go    — top-level mux wiring, /api/services, /api/config
  proxy.go      — reverse proxy construction, path stripping, WebSocket upgrade
  bff.go        — cookie-to-JWT conversion, token caching
  embed.go      — embedded SPA serving with fallback to index.html
```

### Route behavior

| Route pattern | Behavior |
|---------------|----------|
| `/api/auth/*` | **Strip-prefix** — strips `/api/auth`, forwards to auth service with cookies intact, no JWT conversion |
| `/api/{service}/*` | **BFF conversion** — reads session cookie, converts to JWT, forwards with `Authorization: Bearer` |
| `/api/{service}/ui/*` | **BFF conversion** — same as above, proxied to service's `/ui/*` |
| `/api/services` | **Shell endpoint** — returns enabled services with metadata |
| `/api/config` | **Shell endpoint** — returns active fort name |
| Non-`/api/*` | **SPA** — serves embedded files, falls back to `index.html` |

### BFF token conversion

Lives in `bff.go`. Shell-only — not a shared package.

```go
type tokenCache struct {
    mu      sync.RWMutex
    tokens  map[string]cachedToken  // session cookie value → JWT
}

type cachedToken struct {
    jwt    string
    expiry time.Time
}
```

Behavior:
1. Extract session cookie from incoming request
2. Check cache — if token exists and has >1 minute remaining, use it
3. Cache miss or near-expiry: forward session cookie to auth service's `GET /v1/token`, receive JWT
4. Cache the JWT, keyed by session cookie value
5. Attach `Authorization: Bearer <JWT>` to the forwarded request

Error handling:
- **Auth service unreachable** (can't convert cookie): `502 Bad Gateway` with JSON error body
- **Session expired** (auth returns 401): `401 Unauthorized`, evict cache entry for that cookie value, clear session cookie to force re-login
- **No session cookie**: `401 Unauthorized`

Cache eviction on 401 ensures stale entries never survive a server-side session invalidation. Cache key collisions are a non-risk — better-auth generates cryptographically random session tokens.

JWT lifetime is 15 minutes (per auth design spec). Cache refreshes at 14 minutes.

### Auth route handling

`/api/auth/*` requests are handled the same as other services: the proxy strips the `/api/auth` prefix and forwards to the auth service (e.g., `/api/auth/v1/session` → `/v1/session`). Cookies are included, no JWT conversion. This is how login, registration, session checks, and JWKS endpoints work.

### Reverse proxy

Per-service `httputil.ReverseProxy` instances, constructed at startup:

```go
// proxy.go

func newServiceProxy(service domain.Service) *httputil.ReverseProxy
```

- Strips the `/api/{service}` prefix before forwarding (e.g., `/api/nexus/v1/vms` → `/v1/vms`)
- For local forts: forwards to `service.URL`
- For gateway forts: forwards to `fort.Gateway` with the `/api/{service}` prefix preserved (e.g., `/api/nexus/v1/vms` → `https://fort.acme.com/api/nexus/v1/vms`). The gateway is namespace-aware and routes internally based on the service prefix.
- Disabled services: return `503 Service Unavailable` with JSON error body

### WebSocket proxy

- Only paths in the service's `WSPaths` whitelist accept upgrade
- Non-whitelisted upgrade requests → `400 Bad Request`
- BFF conversion happens during the upgrade handshake — JWT attached to the forwarded upgrade request
- After upgrade, frames are proxied bidirectionally using `gorilla/websocket`
- Path matching: after stripping `/api/{service}`, the remaining path is checked against `WSPaths` (e.g., `/api/sharkfin/presence` → checks `/presence` against whitelist)

### SPA serving

Two modes controlled by `--dev` flag:

**Production** (default): serves from `go:embed web/dist` filesystem
- Known files are served directly with correct `Content-Type`
- Unknown paths fall back to `index.html` (SPA client-side routing)
- `Cache-Control: no-cache` for HTML, long-lived cache for hashed assets

**Dev mode** (`--dev`): reverse proxy to Vite dev server
- All non-`/api/*` requests forwarded to `--dev-url` (default `http://localhost:5173`)
- Vite handles HMR, module resolution, etc.

### Shell endpoints

**`GET /api/services`:**

```json
{
  "fort": "local",
  "services": [
    {"name": "auth", "label": "Auth", "route": "/auth", "enabled": true, "ui": true},
    {"name": "sharkfin", "label": "Chat", "route": "/chat", "enabled": true, "ui": true},
    {"name": "nexus", "label": "Nexus", "route": "/nexus", "enabled": true, "ui": false},
    {"name": "hive", "label": "Hive", "route": "/hive", "enabled": false, "ui": false}
  ]
}
```

- All enabled services appear in the list, including auth
- `ui` flag set by probing each service's `/ui/health` at startup (cached, not per-request)
- **Known limitation:** a service that starts after the shell will show `ui: false` until the shell restarts. Acceptable for now — services typically start before the shell
- `label` and `route` from a static metadata map in the handler:

```go
var serviceMetadata = map[string]struct{ Label, Route string }{
    "auth":     {"Auth", "/auth"},
    "sharkfin": {"Chat", "/chat"},
    "nexus":    {"Nexus", "/nexus"},
    "hive":     {"Hive", "/hive"},
}
```

**`GET /api/config`:**

```json
{
  "fort": "local"
}
```

---

## `cmd/web/` — Composition Root

```go
// cmd/web/web.go

func New() *cobra.Command
```

The command wires all layers together:

1. Reads fort config via `fortconfig.New()`
2. Sets up BFF token converter pointing at the auth service URL
3. Builds reverse proxies for each enabled service
4. Registers shell endpoints (`/api/services`, `/api/config`)
5. Registers SPA handler (embedded or dev proxy based on `--dev`)
6. Starts `net/http.Server`
7. Opens browser unless `--no-open`
8. Graceful shutdown on SIGINT/SIGTERM (5-second drain)

### Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--bind` | `127.0.0.1` | Listen address |
| `--port` | `8080` | Listen port |
| `--dev` | `false` | Proxy SPA to Vite dev server instead of embedded files |
| `--dev-url` | `http://localhost:5173` | Vite dev server URL (only used with `--dev`) |
| `--no-open` | `false` | Skip auto-opening browser on startup |

### Server lifecycle

- **Bind**: `{bind}:{port}`, default `127.0.0.1:8080`
- **Startup**: logs the URL, auto-opens browser unless `--no-open`
- **Shutdown**: catches SIGINT/SIGTERM, calls `server.Shutdown(ctx)` with 5-second timeout

---

## Command Registration

> **Note:** CGo dependencies (whisper.cpp, pkg/stt) have been removed. No build gating is needed — both `chat` and `web` commands are registered unconditionally.

Register the web command alongside chat in `cmd/root.go`. Rename `NewChatCmd()` → `New()` for consistency.

---

## Package Structure Summary

```
pkg/
  frontend/
    frontend.go          # Handler(fsys) — embed + serve + health probe
    frontend_test.go

  auth/                  # (already implemented)
    ...

internal/
  domain/
    web.go               # Fort, Service, FortRegistry

  infra/
    fortconfig/
      registry.go        # Viper-backed FortRegistry
      registry_test.go
    httpapi/
      handler.go         # Mux wiring, /api/services, /api/config
      proxy.go           # Reverse proxy per service, path stripping
      bff.go             # Cookie → JWT conversion, token cache
      embed.go           # SPA serving (embedded or dev proxy)
      handler_test.go
      proxy_test.go
      bff_test.go

  chat/                  # (existing, unchanged)
    ...

cmd/
  web/
    web.go               # Composition root: New() *cobra.Command
  chat/
    chat.go              # Renamed: New() (was NewChatCmd)
  root.go                # Both chat and web commands registered here
  main.go
```

---

## Testing Strategy

### `pkg/frontend/`
- Unit tests with an in-memory `fs.FS` containing mock `remoteEntry.js` and `assets/` files
- Verify cache headers per path pattern
- Verify health probe returns correct status based on `remoteEntry.js` presence

### `internal/infra/fortconfig/`
- Set Viper values programmatically, verify `Registry` returns correct `Fort`/`Service` structs
- Test `SetActive` with valid and invalid fort names

### `internal/infra/httpapi/`
- **proxy_test.go**: `httptest.Server` backends, verify path stripping, disabled service 503, WebSocket upgrade whitelist
- **bff_test.go**: mock auth service, verify cookie-to-JWT conversion, caching, error cases (auth down, expired session)
- **handler_test.go**: integration test of the full mux — verify routing to correct handler per path pattern

### Command registration
- Verified by CI: `go build` succeeds and the binary runs both `workfort chat` and `workfort web`
