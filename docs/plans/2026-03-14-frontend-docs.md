# Frontend Documentation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Write the `docs/frontend/` developer documentation site for building service frontends on WorkFort.

**Architecture:** 11 markdown files across `docs/frontend/` and `docs/frontend/getting-started/`. Each doc is self-contained with cross-references via relative links. Content is derived from the actual codebase — every claim must be verifiable against source files.

**Tech Stack:** Markdown. Source of truth: `pkg/frontend/frontend.go`, `web/shell/src/lib/remotes.ts`, `web/shell/vite.config.ts`, `web/packages/`, `internal/infra/httpapi/`.

**Design spec:** `docs/frontend-docs-design.md`

---

## Chunk 1: Platform Fundamentals

### Task 1: architecture.md

**Files:**
- Create: `docs/frontend/architecture.md`

**Source files to reference:**
- `internal/infra/httpapi/fort_router.go` — fort dispatch, prefix stripping
- `internal/infra/httpapi/handler.go` — per-fort mux, service route registration
- `internal/infra/httpapi/proxy.go` — NewServiceProxy path rewriting
- `internal/infra/httpapi/tracker.go` — service discovery, health probing, TrackedService
- `web/shell/src/lib/remotes.ts` — MF runtime registration
- `web/shell/src/stores/services.ts` — JS-side polling
- `pkg/frontend/frontend.go` — /ui/ handler

- [ ] **Step 1: Write the doc**

Sections and key points:

**How it works** — The shell is an MF host. Service frontends are remotes loaded at runtime. No build-time coupling.

**Request lifecycle** — Trace a `remoteEntry.js` request:
1. Browser requests `/forts/{fort}/api/{service}/ui/remoteEntry.js`
2. `FortRouter.fortDispatch` validates fort name, strips `/forts/{fort}` prefix
3. Per-fort `NewHandler` mux matches `/api/{service}/`
4. `NewServiceProxy` forwards to service backend — for local forts, strips `/api/{service}` prefix and sends to the service URL directly; for gateway forts, preserves the prefix and forwards to the gateway URL
5. `pkg/frontend.Handler` serves the file from the embedded FS

**Service discovery** — Two polling loops:
- Go-side: `ServiceTracker` probes each service's `/ui/health` every 10s. Sets `ui: true` if 200, `connected` based on HTTP reachability (or WS ref-count for WebSocket services).
- JS-side: Shell fetches `/forts/{fort}/api/services` every 30s. Registers new MF remotes for services with `ui: true`. New services take up to 30s to appear.

**Fort isolation** — Each fort gets a lazy-initialized `FortInstance` (singleflight-guarded). Separate service tracker, token converter, and handler per fort. Cookies scoped to `/forts/{fort}/`. Idle forts stop polling after 30 minutes; the next request to an idle fort re-runs the initial probe, recreates the handler, and restarts polling.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/architecture.md
git commit -m "docs(frontend): architecture overview"
```

---

### Task 2: service-contract.md

**Files:**
- Create: `docs/frontend/service-contract.md`

**Source files to reference:**
- `web/shell/src/lib/remotes.ts` — ServiceModule interface
- `web/shell/src/components/service-mount.tsx` — how the shell renders remotes
- `web/shell/vite.config.ts` — shared singletons
- `pkg/frontend/frontend.go` — Manifest struct, Handler function, health probe
- `internal/infra/httpapi/tracker.go` — how `connected` is set (HTTP health vs WS ref-count)

- [ ] **Step 1: Write the doc**

Sections and key points:

**TypeScript side** — The `ServiceModule` interface (copy from `remotes.ts`):
- `default` component: receives `{ connected: boolean }`. Explain connected semantics — HTTP services: reflects health probe. WS services: starts false, true after first client connects.
- `manifest` object: `{ name, label, route, minWidth? }`. Must match Go-side Manifest.
- Optional exports: `SidebarContent`, `HeaderActions`.
- The shell loads via `loadRemote('{name}/index')`.

**MF shared singletons** — What the shell shares and what remotes should declare:
- SolidJS remotes: `solid-js`, `@workfort/ui`, `@workfort/ui-solid`, `@workfort/auth` — all `import: false`
- Non-SolidJS remotes: only `@workfort/ui` and `@workfort/auth` as shared. Framework adapters are bundled by the remote.
- `@solidjs/router` is NOT shared (MF dev-mode `require()` breaks ESM-only packages).

**Go side** — `pkg/frontend.Manifest` struct (fields: Name, Label, Route, WSPaths). `frontend.Handler(fsys, manifest)` mounts under `/ui/`:
- Health probe: `GET /ui/health` returns 200 + manifest if `remoteEntry.js` exists in the embedded FS, 503 otherwise. The file-existence check runs once when `Handler()` is called at server init — the result is immutable at runtime.
- Cache headers: `/ui/assets/*` immutable 1yr, everything else no-cache.
- `fsys` must be rooted at the Vite build output dir (`fs.Sub`).

**How the shell renders** — ServiceMount behavior: loading skeleton → error boundary → connected check with warning banner → `<Dynamic component={mod.default}>`.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/service-contract.md
git commit -m "docs(frontend): service module contract reference"
```

---

## Chunk 2: Ecosystem Reference

### Task 3: shared-packages.md

**Files:**
- Create: `docs/frontend/shared-packages.md`

**Source files to reference:**
- `web/packages/ui/src/` — all component source files
- `web/packages/solid/src/index.ts`
- `web/packages/react/src/index.tsx`
- `web/packages/vue/src/index.ts`
- `web/packages/svelte/src/index.ts`

- [ ] **Step 1: Write the doc**

Sections and key points:

**@workfort/ui** — Lit web components, light DOM. Table of all components:

| Tag | Key properties | Key events |
|-----|---------------|------------|
| `wf-panel` | `label` | — |
| `wf-button` | `variant`, `disabled` | `wf-click` |
| `wf-badge` | `count` (hides at 0) | — |
| `wf-status-dot` | `status` (online/offline/away) | — |
| `wf-skeleton` | `width`, `height` | — |
| `wf-divider` | — | — |
| `wf-text-input` | `placeholder`, `value`, `disabled` | `wf-input`, `wf-change` |
| `wf-list` | — | — |
| `wf-list-item` | `active` | `wf-select` |
| `wf-scroll-area` | — | — |
| `wf-error-fallback` | `title`, `message` | — |
| `wf-banner` | `variant`, `dismissible`, `headline`, `details` | `wf-dismiss` |
| `wf-toast` | `variant`, `sticky`, `duration` | `wf-dismiss` |
| `wf-toast-container` | `position` | — |

Also exports `@workfort/ui/style.css`.

**Framework adapters** — Each adapter provides auth and theme integration for its framework:

- `@workfort/ui-solid`: `useAuth()` → `{ user: Accessor<User|null>, isAuthenticated: () => boolean }`. `useTheme()` → `Accessor<'dark'|'light'>`.
- `@workfort/ui-react`: `useAuth()` → `{ user: User|null, isAuthenticated: boolean }` (via `useSyncExternalStore`). `useTheme()` → `'dark'|'light'`. Also provides React wrapper components for all `wf-*` elements (needed for React 18 CE event compat).
- `@workfort/ui-vue`: `useAuth()` → `{ user: Readonly<Ref<User|null>>, isAuthenticated: Readonly<Ref<boolean>> }`. `useTheme()` → `Readonly<Ref<'dark'|'light'>>`. Requires `compilerOptions.isCustomElement: tag => tag.startsWith('wf-')`.
- `@workfort/ui-svelte`: exports `auth` store (`{ user: Readable<User|null>, isAuthenticated: Derived<boolean> }`) and `theme` store (`Readable<'dark'|'light'>`). Not hooks — Svelte store pattern.

Note which frameworks handle `wf-*` elements natively (Solid, Vue, Svelte) vs. need wrappers (React).

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/shared-packages.md
git commit -m "docs(frontend): shared packages reference"
```

---

### Task 4: auth.md

**Files:**
- Create: `docs/frontend/auth.md`

**Source files to reference:**
- `/home/kazw/Work/WorkFort/passport/lead/packages/auth/src/` — AuthClient, types
- `web/packages/solid/src/index.ts` — useAuth implementation
- `internal/infra/httpapi/handler.go` — bffMiddleware, writeAuthError
- `internal/infra/httpapi/proxy.go` — NewAuthProxy, rewriteCookiePaths

- [ ] **Step 1: Write the doc**

Sections and key points:

**BFF pattern** — The shell's Go backend handles auth tokens. Session cookies are converted to JWT Bearer tokens by `bffMiddleware` before proxying to services. Frontends never see or manage tokens.

**@workfort/auth client API** — Verify against Passport repo, then document:
- `getAuthClient()` → singleton `AuthClient`
- `client.getUser()` → `User | null`
- `client.getSession()` → `Session | null`
- `client.isAuthenticated` (getter)
- `client.init()` — fetches session from `GET /api/auth/v1/session`, sets up visibility-based auto-refresh
- `client.logout()` — POSTs `/api/auth/v1/sign-out`, clears state, emits events
- Events: `change` (User | null), `logout` (void)
- `User` type: `{ id, username, name, displayName, type: 'user'|'agent'|'service' }`
- `Session` type: `{ id, expiresAt, refreshedAt }`

**Using auth in each framework** — Brief example per adapter showing reactive user state and logout handling. Reference shared-packages.md for full adapter API.

**Per-fort cookie scoping** — Cookies set with `Path: /forts/{fort}/`. Logging into fort A does not authenticate fort B. Session expiry returns 401 + clears the cookie. Auth service down returns 502.

**Shared singleton** — `@workfort/auth` MUST be shared (`import: false`). Never bundle your own copy.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/auth.md
git commit -m "docs(frontend): authentication guide"
```

---

## Chunk 3: Getting Started — SolidJS & React

### Task 5: getting-started/solidjs.md

**Files:**
- Create: `docs/frontend/getting-started/solidjs.md`

**Source files to reference:**
- `web/shell/vite.config.ts` — MF shared config to mirror
- `web/shell/src/lib/remotes.ts` — ServiceModule interface
- `pkg/frontend/frontend.go` — Go handler wiring

- [ ] **Step 1: Write the doc**

Steps to cover (with minimal code examples — real, runnable, not pseudocode):

1. **Scaffold** — `pnpm create vite my-service --template solid-ts`
2. **Install deps** — `@workfort/ui`, `@workfort/ui-solid`, `@workfort/auth`, `@module-federation/vite`
3. **Vite config** — MF plugin: `name`, `exposes: { './index': './src/index.tsx' }`, shared singletons (`solid-js`, `@workfort/ui`, `@workfort/ui-solid`, `@workfort/auth` — all `singleton: true, import: false`)
4. **Entry module** — Export `default` component (receives `{ connected }`), export `manifest` object, optionally export `SidebarContent`
5. **Go wiring** — Embed the `dist/` dir, call `frontend.Handler(fsys, manifest)`, mount on the service's HTTP mux
6. **Fort config** — Add the service URL to a fort's config YAML
7. **Run** — Start shell (`mise run dev:go` + `mise run dev:web`), start service, navigate to it

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/getting-started/solidjs.md
git commit -m "docs(frontend): SolidJS getting started guide"
```

---

### Task 6: getting-started/react.md

**Files:**
- Create: `docs/frontend/getting-started/react.md`

- [ ] **Step 1: Write the doc**

Same logical steps as SolidJS with React-specific differences:

1. **Scaffold** — `pnpm create vite my-service --template react-ts`
2. **Install deps** — `@workfort/ui`, `@workfort/ui-react`, `@workfort/auth`, `@module-federation/vite`
3. **Vite config** — Shared: only `@workfort/ui` and `@workfort/auth` (`import: false`). `solid-js` and `@workfort/ui-solid` are NOT relevant. React itself is bundled by the remote (not shared by the shell).
4. **Entry module** — Same ServiceModule shape. Use React wrapper components from `@workfort/ui-react` (needed for event handling in React 18).
5. **Go wiring** — Identical to SolidJS guide (link to it).
6. **Fort config** — Same.
7. **Run** — Same.

Note the React-specific gotcha: `wf-*` custom element events don't work with React's synthetic event system — use the wrapper components from `@workfort/ui-react`.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/getting-started/react.md
git commit -m "docs(frontend): React getting started guide"
```

---

## Chunk 4: Getting Started — Vue, Svelte, Web Components

### Task 7: getting-started/vue.md

**Files:**
- Create: `docs/frontend/getting-started/vue.md`

- [ ] **Step 1: Write the doc**

Same steps with Vue-specific differences:

1. **Scaffold** — `pnpm create vite my-service --template vue-ts`
2. **Install deps** — `@workfort/ui`, `@workfort/ui-vue`, `@workfort/auth`, `@module-federation/vite`
3. **Vite config** — Shared: `@workfort/ui`, `@workfort/auth`. Add `compilerOptions.isCustomElement: tag => tag.startsWith('wf-')` to the Vue plugin config.
4. **Entry module** — Same ServiceModule shape. Vue handles `wf-*` natively with the custom element compiler option.
5. **Go wiring** — Link to SolidJS guide.
6. **Run** — Same.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/getting-started/vue.md
git commit -m "docs(frontend): Vue getting started guide"
```

---

### Task 8: getting-started/svelte.md

**Files:**
- Create: `docs/frontend/getting-started/svelte.md`

- [ ] **Step 1: Write the doc**

Svelte-specific differences:

1. **Scaffold** — `pnpm create vite my-service --template svelte-ts`
2. **Install deps** — `@workfort/ui`, `@workfort/ui-svelte`, `@workfort/auth`, `@module-federation/vite`
3. **Vite config** — Shared: `@workfort/ui`, `@workfort/auth`.
4. **Entry module** — Same ServiceModule shape. Svelte handles `wf-*` natively. Auth/theme use Svelte stores (not hooks) — import from `@workfort/ui-svelte`.
5. **Go wiring** — Link to SolidJS guide.
6. **Run** — Same.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/getting-started/svelte.md
git commit -m "docs(frontend): Svelte getting started guide"
```

---

### Task 9: getting-started/web-components.md

**Files:**
- Create: `docs/frontend/getting-started/web-components.md`

- [ ] **Step 1: Write the doc**

The lowest-level path — no framework, no adapter:

1. **Scaffold** — `pnpm create vite my-service --template vanilla-ts`
2. **Install deps** — `@workfort/ui`, `@workfort/auth`, `@module-federation/vite` (no adapter package)
3. **Vite config** — Shared: `@workfort/ui`, `@workfort/auth`.
4. **Entry module** — Same ServiceModule shape. The `default` export is a function that receives `{ connected }` and returns a DOM element (or mounts into a container). Import `@workfort/ui/style.css`. Use `getAuthClient()` directly for auth — subscribe to `change`/`logout` events manually.
5. **Go wiring** — Link to SolidJS guide.
6. **Run** — Same.

Emphasize: this is the path for any framework without a dedicated adapter.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/getting-started/web-components.md
git commit -m "docs(frontend): web components getting started guide"
```

---

## Chunk 5: Workflow & Index

### Task 10: dev-workflow.md

**Files:**
- Create: `docs/frontend/dev-workflow.md`

**Source files to reference:**
- `.mise/tasks/dev/go` and `.mise/tasks/dev/web`
- `cmd/web/web.go` — `--dev` flag, SPA proxy
- `internal/infra/httpapi/spa.go` — NewSPADevProxy implementation
- `internal/infra/httpapi/fort_router.go` — proxy chain

- [ ] **Step 1: Write the doc**

Sections and key points:

**Shell dev servers** — `mise run dev:go` (Go on :16100, `--dev` proxies SPA to Vite) + `mise run dev:web` (Vite on :5173). Both must be running.

**Running a service frontend** — Point the fort config's service URL at your local dev server (e.g. `http://localhost:3001`). The Go backend proxies `/api/{service}/` to that URL. The shell's MF runtime loads `remoteEntry.js` through this proxy chain. Your service's Vite dev server must serve the MF build output including `remoteEntry.js` at the `/ui/` prefix.

**HMR** — Vite HMR works within the service remote. Full MF remote reload requires a page refresh (remoteEntry.js is no-cache).

**Troubleshooting** — Common issues:
- Service not appearing: check `/api/forts` and `/forts/{fort}/api/services` responses. Up to 30s delay (JS poll interval).
- MF load failure: check browser console for shared version mismatches. Ensure `import: false` on shared deps.
- CORS errors: all requests go through the Go proxy — no direct cross-origin calls needed.
- `ui: false`: the Go backend couldn't reach `/ui/health` or didn't get a 200. Check the service is running and serving `remoteEntry.js`.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/dev-workflow.md
git commit -m "docs(frontend): development workflow guide"
```

---

### Task 11: README.md

**Files:**
- Create: `docs/frontend/README.md`

- [ ] **Step 1: Write the doc**

Table of contents with one-line descriptions and relative links. Suggested reading order:

1. architecture.md — How the shell loads service frontends
2. service-contract.md — What a service frontend must export (TS + Go)
3. shared-packages.md — Available UI components and framework adapters
4. auth.md — Authentication from the frontend perspective
5. getting-started/ — Step-by-step per framework (SolidJS, React, Vue, Svelte, Web Components)
6. dev-workflow.md — Running in development, troubleshooting

Nothing else in this file.

- [ ] **Step 2: Commit**

```bash
git add docs/frontend/README.md
git commit -m "docs(frontend): add README index"
```

---

## Verification

After all tasks:

- [ ] Every relative link between docs resolves correctly
- [ ] Every code reference matches the actual source file
- [ ] `@workfort/auth` API verified against Passport repo
- [ ] No doc references files or APIs that don't exist
