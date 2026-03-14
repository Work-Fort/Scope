# Fort-Scoped Routing

> **Status:** Design
> **Supersedes:** Single-fort active-state model from `2026-03-11-web-ui-design.md` and `2026-03-12-go-web-shell-design.md`

## Problem

The current architecture loads a single "active fort" at startup via `registry.Active()`. All API routes live under `/api/{service}/*` with no fort context in the path. This causes:

- **Mutable server-side state.** Switching forts requires a `SetActive()` mutation — not idempotent, races between tabs.
- **WebSocket breakage.** Switching the active fort invalidates existing WS connections in other tabs.
- **Cookie leakage.** A single session cookie is sent to all service proxies regardless of which fort they belong to. Forts are user-run instances that could be malicious.
- **No multi-tab support.** Two browser tabs cannot view different forts simultaneously.

## Solution

Put the fort name in every URL path. The server becomes stateless — it reads the fort from the request path and dispatches to the right handler. "Active fort" is purely a frontend concern (which URL you're on).

## Scale Target

Support at least 24 configured forts (contractor with many clients). Lazy initialization ensures only visited forts consume resources.

## Fort Name Validation

Fort names must match `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$` — lowercase alphanumeric + hyphens, no leading or trailing hyphens. Validated at config load time (registry rejects invalid names) and at URL extraction time (FortRouter returns 404 for invalid names). This prevents path traversal (`..`), URL encoding attacks (`%2F`), and collisions with reserved names (fort name `api` would conflict with `/api/forts`).

---

## URL Structure

### Go Server

```
Root (no fort scope):
  GET  /api/forts                              → list all configured forts

Fort-scoped:
  GET  /forts/{fort}/api/services              → discovered services for this fort
  GET  /forts/{fort}/api/config                → config for this fort
  *    /forts/{fort}/api/auth/*                → pass-through proxy (no BFF)
  *    /forts/{fort}/api/{service}/*           → bffMiddleware → reverse proxy
  GET  /forts/{fort}/api/{service}/ws          → bffMiddleware → WS proxy
  GET  /forts/{fort}/api/{service}/presence    → bffMiddleware → WS proxy

SPA:
  GET  /*                                      → SPA handler (fallback)
```

### Frontend (SolidJS Router)

```
/                                → Fort picker (or auto-redirect if 1 fort)
/forts/:fort/:service/*rest      → Service page
```

---

## Cookie Scoping

The Go auth proxy rewrites `Set-Cookie` response headers to enforce `Path=/forts/{fort}/`. The auth service (better-auth) has no knowledge of fort-scoped paths — it sets cookies with its own default path. The proxy's `ModifyResponse` intercepts these and rewrites:

```
Auth service responds:
  Set-Cookie: better-auth.session_token=abc; Path=/; HttpOnly; Secure; SameSite=Lax

Proxy rewrites to:
  Set-Cookie: better-auth.session_token=abc; Path=/forts/local/; HttpOnly; Secure; SameSite=Lax
```

The browser only sends a cookie when the request path matches. A malicious fort's service can never see another fort's session cookie. `GET /api/forts` (root-level) receives no cookies.

**Auth is per-fort.** Each fort has its own auth service URL. The `TokenConverter` for fort A calls fort A's auth, using fort A's cookie. Complete isolation.

**Cookie clearing** also uses the fort-scoped path. The `bffMiddleware` receives the fort name so `writeAuthError` can set `Path=/forts/{fort}/` when clearing expired session cookies.

---

## Go Server Architecture

### New: FortRouter

Top-level HTTP handler. Owns the `sync.Map` of `FortInstance`s and the `/api/forts` endpoint.

```go
type FortRouter struct {
    registry  domain.FortRegistry
    instances sync.Map  // map[string]*FortInstance
    spaHandler http.Handler  // dev proxy or embedded SPA
}
```

- `GET /api/forts` — reads `registry.Forts()`, returns `[{name, local, gateway}]`
- `/forts/{fort}/*` — extracts fort name, looks up or lazily creates `FortInstance`, strips prefix, dispatches
- `/*` — SPA fallback

### New: FortInstance

Per-fort isolation unit. Created on first request to that fort.

```go
type FortInstance struct {
    fort     domain.Fort
    tracker  *ServiceTracker
    tc       *TokenConverter
    handler  http.Handler      // the existing NewHandler() output
    lastReq  atomic.Int64      // unix timestamp of last request
    cancel   context.CancelFunc // stops polling
}
```

### Lazy Initialization

Uses `singleflight.Group` to prevent duplicate initialization when concurrent requests arrive for a new fort:

1. Request arrives: `GET /forts/acme-corp/api/services`
2. `FortRouter` extracts `"acme-corp"`, checks `sync.Map` → miss
3. Enters `singleflight.Do("acme-corp", initFn)` — concurrent requests block here
4. Looks up `registry.Fort("acme-corp")` → returns fort config (404 if not found)
5. Creates `ServiceTracker` from fort's service URLs
6. Runs `tracker.InitialProbe(ctx)` (synchronous)
7. Creates `TokenConverter` pointing at fort's auth service
8. Calls `NewHandler(fort, tracker, tc, nil)` — no SPA (SPA is top-level)
9. Starts `tracker.StartPolling(ctx, 10s)`
10. Stores `FortInstance` in `sync.Map`, returns
11. All concurrent requests unblock, serve from the single `FortInstance`
12. Subsequent requests: `sync.Map` hit, update `lastReq`, serve immediately

### Idle Cleanup

Background goroutine checks every 5 minutes. If a `FortInstance` has no requests in 30 minutes:
- Cancel context → stops polling goroutine
- `FortInstance` stays in `sync.Map` (cheap without goroutines, ~2KB)
- Next request re-initializes: re-runs `InitialProbe` + restarts polling
- The handler is recreated on re-initialization (old routes may reference dead services)

At 24 forts with ~4 services each, 3 active + 21 idle: 3 polling goroutines, ~12 HTTP health probes every 10 seconds.

### Dev Mode

`FortRouter.spaHandler` is set to `NewSPADevProxy(devURL)` in dev mode, or `NewSPAHandler(spaFS)` in production. No separate `topMux` is needed — `FortRouter` owns the SPA fallback in both modes. The existing `topMux` pattern in `cmd/web/web.go` is removed.

---

## Changes to Existing Code

### New Files

| File | Purpose |
|------|---------|
| `internal/infra/httpapi/fort_router.go` | `FortRouter`, `FortInstance`, lazy init, idle cleanup |
| `internal/infra/httpapi/fort_router_test.go` | Multi-fort routing tests |
| `web/shell/src/components/fort-picker.tsx` | Fort selection landing page |

### Modified Files

| File | Change |
|------|--------|
| `internal/infra/httpapi/handler.go` | Remove SPA handling (moved to FortRouter). Handler routes stay as `/api/*` — fort prefix already stripped before dispatch. Accept fort name parameter for cookie path scoping. |
| `internal/infra/httpapi/handler.go` | `bffMiddleware` and `writeAuthError` receive fort name for cookie path scoping (`Path=/forts/{fort}/`). Use `NewAuthProxy` for auth routes. SPA disabled by passing `nil` for `spaFS`. |
| `internal/infra/httpapi/proxy.go` | Auth proxy gets `ModifyResponse` to rewrite `Set-Cookie` headers with `Path=/forts/{fort}/`. |
| `internal/domain/web.go` | Add `Fort(name string) (Fort, bool)` to `FortRegistry` interface. Remove `Active()` and `SetActive()`. |
| `internal/infra/fortconfig/registry.go` | Implement `Fort(name)`. Remove `Active()`/`SetActive()`. |
| `cmd/web/web.go` | Create `FortRouter` with full registry instead of single fort. Remove single tracker/token converter creation. SPA handler wraps `FortRouter`. |
| `web/shell/src/lib/api.ts` | Add `fetchForts()`. All other functions take `fort` parameter: `fetchServices(fort)`, `fetchConfig(fort)`. |
| `web/shell/src/stores/services.ts` | Fort-scoped polling. Start/stop tied to current fort from URL param. |
| `web/shell/src/app.tsx` | Router: `/` → FortPicker, `/forts/:fort/:service/*rest` → ServicePage. |
| `web/shell/src/components/nav-bar.tsx` | Show fort name from URL param. |

### Removed

| Item | Reason |
|------|--------|
| `FortRegistry.Active()` | No server-side active fort |
| `FortRegistry.SetActive()` | No server-side fort switching |
| `active-fort` config key | Replaced by URL routing |

---

## Frontend

### Fort Picker (`/`)

- Fetches `GET /api/forts`
- If only 1 fort: `<Navigate>` redirects to `/forts/{fort}/{first-service}`
- If multiple: renders a list of fort names with click → navigate to `/forts/{fort}/{first-service}`
- If zero forts configured: shows "No forts configured" with guidance
- If fort has no healthy services: redirects to `/forts/{fort}` (empty state page)

### Fort-Scoped Services Store

The `services` store becomes fort-aware:
- Reads `:fort` from URL params
- Polls `GET /forts/{fort}/api/services`
- When `:fort` changes (navigation), stops old polling, starts new
- Each fort's service state is independent

### API Client

```typescript
function fetchForts(): Promise<FortsResponse>           // GET /api/forts
function fetchServices(fort: string): Promise<ServicesResponse>  // GET /forts/{fort}/api/services
function fetchConfig(fort: string): Promise<ConfigResponse>     // GET /forts/{fort}/api/config
```

---

## Request Flow

```
Browser: GET /forts/local/api/nexus/v1/vms
  Cookie: better-auth.session_token=abc (Path=/forts/local/)

  → FortRouter
    → extract fort = "local"
    → sync.Map lookup → FortInstance for "local"
    → strip /forts/local → /api/nexus/v1/vms

  → FortInstance.handler (existing NewHandler output)
    → bffMiddleware
      → TokenConverter.Token(r) → reads cookie → calls local's auth
      → GET http://127.0.0.1:3000/v1/token → JWT
      → sets Authorization: Bearer {jwt}

    → ServiceProxy
      → strips /api/nexus → /v1/vms
      → proxies to http://127.0.0.1:9600/v1/vms

  → Response flows back
```

---

## Future: Constellation

The lazy-init + tracker architecture is designed for the constellation migration. When constellation arrives:
- `ServiceTracker` switches from HTTP polling to WS subscription
- Idle cleanup becomes less important (WS connections are cheap when idle)
- `FortRouter` and `FortInstance` structure stays the same
- Only the tracker internals change
