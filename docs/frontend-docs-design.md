# Frontend Documentation — Design Spec

Documentation for building service frontends on the WorkFort platform. Lives at `docs/frontend/` as a self-contained developer docs site, separate from the historical design records in `docs/archive/` and `docs/plans/`.

## Audience

Open-source developers building service frontends. Initially internal (Passport, Sharkfin, Hive), eventually external teams deploying WorkFort on their own infrastructure. Assumes general web development experience but not familiarity with the WorkFort codebase.

## Structure

```
docs/frontend/
  README.md
  architecture.md
  service-contract.md
  shared-packages.md
  auth.md
  getting-started/
    solidjs.md
    react.md
    vue.md
    svelte.md
    web-components.md
  dev-workflow.md
```

## Document Specs

### README.md

Table of contents with one-line descriptions linking to each doc. Suggested reading order for new developers. Nothing else.

### architecture.md

How the shell loads service frontends at runtime:

- The shell is a Module Federation host. Service frontends are remotes discovered at runtime, not built into the shell.
- Request path: browser → `/forts/{fort}/api/{service}/ui/remoteEntry.js` → FortRouter strips fort prefix → ServiceProxy forwards to service backend → `pkg/frontend.Handler` serves the Vite build.
- Service discovery: two polling loops. The Go backend probes each service's `/ui/health` every 10 seconds, updating connectivity and `ui: true/false` state. The shell JS fetches `/forts/{fort}/api/services` every 30 seconds, picks up new services, and registers MF remotes. The 30-second interval is why a newly started service can take up to 30s to appear in the shell.
- Fort-scoped routing: each fort gets its own isolated instance (lazy-initialized, singleflight-guarded). Services are proxied per-fort. Cookies are scoped to `/forts/{fort}/`.

No framework-specific content. This is the platform-level view.

### service-contract.md

The spec a service frontend must satisfy. Two sides: TypeScript and Go.

**TypeScript (MF remote):**

- `ServiceModule` interface — the exact shape the shell expects from `loadRemote('{name}/index')`:
  - `default`: root component, receives `{ connected: boolean }`. For HTTP-only services, `connected` reflects health probe reachability. For WebSocket services, `connected` starts `false` and becomes `true` only after the first WS client connects — document this distinction.
  - `manifest`: `{ name, label, route, minWidth? }`
  - `SidebarContent?`: optional component injected into shell sidebar
  - `HeaderActions?`: optional component injected into shell header
- MF shared singletons: the shell currently shares `solid-js`, `@workfort/ui`, `@workfort/ui-solid`, and `@workfort/auth`. A remote should only declare `import: false` for singletons the shell actually provides. For non-SolidJS remotes: `solid-js` and `@workfort/ui-solid` are irrelevant — only mark `@workfort/ui` and `@workfort/auth` as shared. The framework-specific adapter (e.g. `@workfort/ui-react`) is bundled by the remote, not shared by the shell.
- `@solidjs/router` is NOT shared. Remotes that need routing must use their own instance.

**Go (`pkg/frontend`):**

- `Manifest` struct: `{ Name, Label, Route, WSPaths }`
- `Handler(fsys fs.FS, m Manifest) http.Handler` — mounts under `/ui/`
- Health probe: `GET /ui/health` returns 200 + manifest JSON if `remoteEntry.js` exists in the embedded FS, 503 otherwise.
- Cache behavior: `/ui/assets/*` is immutable (1 year), `remoteEntry.js` is no-cache, everything else is no-cache.
- The `fsys` must be rooted at the Vite build output directory.

### shared-packages.md

What's available to service frontends via MF singleton sharing.

**`@workfort/ui`** — Lit-based web components using light DOM. List each component with its purpose and key attributes/events. These work in any framework.

**Framework adapters** — `@workfort/ui-solid`, `@workfort/ui-react`, `@workfort/ui-vue`, `@workfort/ui-svelte`. Each exports `useAuth()` and `useTheme()`. Document the API for each.

**`@workfort/auth`** — The auth client from the Passport repo. Shared as singleton — never bundle your own copy. API surface must be verified against the Passport repo (`/home/kazw/Work/WorkFort/passport/lead/`) before writing the actual docs.

### auth.md

Authentication from a service frontend developer's perspective.

- The BFF pattern: the shell's Go backend converts session cookies into JWT Bearer tokens before proxying to services. Frontends never handle tokens directly.
- Per-fort cookie scoping: cookies are set with `Path: /forts/{fort}/`. Logging into one fort does not authenticate another.
- Using `useAuth()` (framework adapters) or subscribing to the auth client directly (web components). Reactive user state, logout handling.
- What happens when auth is down: the shell returns 502. What happens when the session expires: 401 + cookie cleared.

### getting-started/solidjs.md (and react.md, vue.md, svelte.md, web-components.md)

Step-by-step guide for each framework. Same logical steps, framework-specific code:

1. Scaffold a Vite project
2. Install dependencies (`@workfort/ui`, the relevant adapter, `@workfort/auth`, `@module-federation/vite`)
3. Configure the MF plugin in `vite.config` — name, exposes `./index`, shared singletons with `import: false`
4. Create the entry module: export `default` component + `manifest` object (+ optional `SidebarContent`, `HeaderActions`)
5. Wire `pkg/frontend.Handler` in the Go service — embed the Vite build, register the handler on the service's HTTP mux
6. Add the service URL to a fort's config
7. Start the shell + service, see it load in the browser

The web-components guide uses `@workfort/ui` directly with vanilla JS and the raw auth client — no framework adapter.

### dev-workflow.md

Running a service frontend in development:

- How `mise run dev:go` (Go on :16100) and `mise run dev:web` (Vite on :5173) work together — the `--dev` flag proxies SPA requests to Vite.
- Running a service frontend dev server: point the fort config's service URL at your local dev server. The Go backend proxies `/api/{service}/` to that URL, including `/ui/` paths. The shell's MF runtime loads `remoteEntry.js` through this proxy.
- HMR: Vite's HMR works for the service remote if the dev server is running. The shell's MF runtime fetches `remoteEntry.js` on each navigation (no-cache), so a rebuild is picked up on page refresh.
- Troubleshooting: common federation load failures (CORS, shared version mismatches, missing exports).

## Writing Style

- Concise. Every sentence earns its place but nothing is left ambiguous.
- Code examples are real, runnable, and minimal — not pseudocode.
- No marketing language, no filler, no "in this document we will..."
- Reference other docs by relative link rather than repeating content.
