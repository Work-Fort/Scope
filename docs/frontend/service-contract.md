# Service Frontend Contract

A service frontend is a Module Federation remote. It must satisfy both a TypeScript export contract (consumed by the shell) and a Go HTTP contract (consumed by the shell's service tracker).

---

## TypeScript: `ServiceModule` interface

The shell loads each service via `loadRemote('{name}/index')` and validates the result. The remote's `index` entrypoint must export:

```ts
// web/shell/src/lib/remotes.ts
export interface ServiceModule {
  default: (props: { connected: boolean }) => any;
  manifest: { name: string; label: string; route: string; minWidth?: number };
  SidebarContent?: () => any;
  HeaderActions?: () => any;
}
```

`default` and `manifest` are required. `SidebarContent` and `HeaderActions` are optional slots rendered by the shell's layout.

### `connected` semantics

`connected` is derived from the shell's `ServiceTracker`, not from the remote itself.

- **HTTP-only service** (`WSPaths` empty): `connected` is `true` whenever the health probe returns 200, `false` when the service is unreachable.
- **WebSocket service** (`WSPaths` non-empty): `connected` starts `false` on discovery and becomes `true` only while at least one WS client is connected. The health probe does not affect `connected` for WS services. The tracker maintains a ref count via `OnConnect`/`OnDisconnect`.

### `manifest` must match the Go-side `Manifest`

`name`, `label`, and `route` must be identical to the values the Go binary declares in its `frontend.Manifest`. The shell reads those fields from the health probe; `ServiceModule.manifest` is used by the shell's routing and layout. A mismatch causes undefined behavior.

`minWidth` is TypeScript-only and has no Go equivalent.

---

## MF shared singletons

The shell declares the following shared modules in its `vite.config.ts`:

```ts
// web/shell/vite.config.ts
shared: {
  'solid-js': { singleton: true, eager: true },
  '@workfort/ui': { singleton: true, eager: true },
  '@workfort/ui-solid': { singleton: true, eager: true },
  '@workfort/auth': { singleton: true, eager: true },
}
```

Remotes that use SolidJS must declare `solid-js`, `@workfort/ui`, `@workfort/ui-solid`, and `@workfort/auth` as shared with `import: false` (consume from shell, do not bundle their own copy).

Remotes that do not use SolidJS should only declare `@workfort/ui` and `@workfort/auth` as shared. Framework adapter libraries are bundled by the remote.

`@solidjs/router` is **not** shared. MF's dev-mode virtual module generator uses `require()` to detect named exports, which fails for ESM-only packages. Remotes receive routing context through props, not through a shared router instance.

---

## Go: `frontend.Manifest` and `frontend.Handler`

### Manifest struct

```go
// go/frontend/frontend.go
type Manifest struct {
    Name    string   `json:"name"`
    Label   string   `json:"label"`
    Route   string   `json:"route"`
    WSPaths []string `json:"ws_paths,omitempty"`
}
```

`WSPaths` controls the `connected` semantics described above. An empty or omitted `WSPaths` marks the service as HTTP-only.

### `Handler(fsys, manifest)`

```go
func Handler(fsys fs.FS, m Manifest) http.Handler
```

Mounts the following routes under `/ui/`:

| Route | Behavior |
|---|---|
| `GET /ui/health` | 200 + manifest JSON if `remoteEntry.js` exists; 503 + manifest JSON otherwise |
| `/ui/assets/*` | `Cache-Control: public, max-age=31536000, immutable` |
| `/ui/*` (everything else) | `Cache-Control: no-cache` |

`fsys` must be rooted at the Vite build output directory. Pass the result of `fs.Sub(embedFS, "web/dist")` or equivalent — `Handler` opens files relative to that root.

The `remoteEntry.js` existence check runs once when `Handler` is called at server startup. It does not re-check at request time. The health status is immutable for the lifetime of the process.

Both 200 and 503 responses include the full manifest JSON body. The tracker uses both status codes: 200 means the UI is built and available (`hasUI = true`), 503 means the service is reachable but the UI is not built (`hasUI = false`). The presence of `ws_paths` in either response determines whether `connected` is WS-driven.

---

## Shell rendering: `ServiceMount`

The shell renders each service through `ServiceMount`, which layers four behaviors:

1. **`ErrorBoundary`** — catches any error from `loadRemote` or the remote component itself and renders `<Unavailable>`.
2. **`Suspense`** — renders `<wf-skeleton width="100%" height="200px">` while the remote module is loading.
3. **Warning banner** — if the module has loaded but `connected` is `false`, renders a `<wf-banner variant="warning">` instead of the component. The banner message tells the user the service is starting up or temporarily unavailable and will update automatically.
4. **`<Dynamic>`** — once the module is loaded and `connected` is `true`, renders `mod().default` with `{ connected }` as props.

The condition for the warning banner is `!(mod() || props.connected)`: the banner shows only when both the module is absent and `connected` is false. Once the module has loaded, it is rendered regardless of `connected` — the component receives the live `connected` value as a prop and handles it internally.
