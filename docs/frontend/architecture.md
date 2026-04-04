# Frontend Architecture

The shell is a Module Federation host. Service frontends are MF remotes — they are not built into the shell. The shell discovers them at runtime via a polling loop and registers them with the MF runtime on demand.

## How it works

Each service that ships a frontend exposes a `remoteEntry.js` via its embedded `go/frontend.Handler`. The shell's JS-side polling loop fetches `/forts/{fort}/api/services` every 30 seconds. For each service returned with `ui: true`, `registerNewRemotes` calls `@module-federation/runtime`'s `registerRemotes` to make the remote available. Once registered, `loadRemote("{name}/index")` loads the module lazily.

The MF runtime is initialized once at shell startup with no remotes:

```ts
// web/shell/src/lib/remotes.ts
init({ name: 'shell', remotes: [] });
```

Remotes are added incrementally as services appear. A `registeredNames` set prevents double-registration across poll cycles.

## Request lifecycle

A `remoteEntry.js` request travels this path:

1. **Browser** requests `GET /forts/{fort}/api/{service}/ui/remoteEntry.js`.

2. **`FortRouter.fortDispatch`** (`internal/infra/httpapi/fort_router.go`) validates the fort name via `domain.ValidFortName`. If invalid, returns 404. Strips the `/forts/{fort}` prefix from `r.URL.Path` and hands off to the per-fort handler.

3. **Per-fort mux** (`internal/infra/httpapi/handler.go`) was built by `NewHandler`. It registered a route for `/api/{service}/` during `initInstance`. The mux matches and dispatches to `bffMiddleware`, which wraps `NewServiceProxy`.

4. **`NewServiceProxy`** (`internal/infra/httpapi/proxy.go`) rewrites the path based on fort type:
   - **Local fort** (`local=true`): strips `/api/{serviceName}` prefix and proxies to `targetURL`. Example: `/api/nexus/ui/remoteEntry.js` becomes `/ui/remoteEntry.js` at `http://target`.
   - **Pylon fort** (`local=false`): discovers services via Pylon, proxies to each service's `base_url` directly.

5. **`go/frontend.Handler`** (`go/frontend/frontend.go`) serves the file from the embedded FS. It registers `/ui/remoteEntry.js` under the catch-all `/ui/` handler with `Cache-Control: no-cache`. Content-hashed assets under `/ui/assets/` get `Cache-Control: public, max-age=31536000, immutable`.

## Service discovery

Two polling loops run in parallel.

**Go-side (`ServiceTracker`)** — `internal/infra/httpapi/tracker.go`

`StartPolling` runs every 10 seconds (passed as `interval` by `initInstance`: `tracker.StartPolling(pollCtx, 10*time.Second)`). Each cycle calls `probeOne` for every configured service URL, which issues `GET {serviceURL}/ui/health`.

- A 200 response with a valid JSON manifest sets `ui: true`. The manifest also carries `name`, `label`, `route`, and optional `ws_paths`.
- If the service is unreachable (HTTP error), `Connected` is set to `false` for non-WS services. WS services derive `Connected` from their WebSocket reference count instead.
- Services newly discovered after the initial probe trigger `OnServiceDiscovered`, which calls `registerOneServiceRoute` to add the route to the live mux.

**JS-side** — `web/shell/src/stores/services.ts`

`startPolling` fires immediately then repeats every `POLL_INTERVAL = 30_000` ms. Each result passes through `handlePollResult`, which calls `registerNewRemotes(fort, services)`. Only services with `enabled: true` and `ui: true` that have not been registered before are passed to `registerRemotes`. New services therefore take up to 30 seconds to appear in the shell after the Go-side tracker first sees them.

## Fort isolation

Each fort gets a lazily initialized `FortInstance` holding its own `ServiceTracker`, `TokenConverter`, and `http.Handler`. Initialization is deduplicated with a `singleflight.Group` keyed by fort name.

`StartIdleCleanup` runs a ticker every 5 minutes. Any fort whose `lastReq` timestamp is older than `maxIdle` (30 minutes by default) has its polling context cancelled (`stopPolling`). The `FortInstance` remains in the `sync.Map` but is marked idle (`cancel == nil`). The next request to that fort sees `isIdle() == true`, triggers `initInstance` again — re-running `InitialProbe`, rebuilding the handler, and restarting the 10-second polling loop.

Cookies are scoped to `/forts/{fort}/` by `NewAuthProxy`, which rewrites `Set-Cookie` paths in `ModifyResponse`.
