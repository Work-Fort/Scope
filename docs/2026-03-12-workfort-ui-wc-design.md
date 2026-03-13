# `@workfort/ui` — Framework-Agnostic Web Components Design

## Overview

`@workfort/ui` is a Web Components library that provides the shared UI primitives for WorkFort. Built with Lit, rendered in the light DOM, themed via `--wf-*` CSS custom properties. Any framework (React, Vue, Svelte, SolidJS, vanilla JS) can use the components directly as Custom Elements. Framework adapters provide idiomatic DX — typed props, hooks, composables — for each ecosystem.

**Supersedes:** The `@workfort/ui` section (lines 220-313) of `docs/2026-03-11-web-ui-design.md` and the implementation plan at `docs/plans/2026-03-11-workfort-ui.md`. Those described a SolidJS-only component library. This spec replaces the component library approach while leaving the rest of the web-ui design (shell, routing, proxy, Module Federation host) unchanged.

**Why the change:** WorkFort is designed around teams building custom software with their existing tech stacks. Requiring SolidJS for all service UIs forces every adopter to learn a new framework. Web Components are the platform-native solution — they work in every framework and require no framework at all.

**Related specs:**
- `docs/2026-03-11-web-ui-design.md` — full web UI architecture (shell, routing, Module Federation)
- `docs/2026-03-12-go-web-shell-design.md` — Go-side shell (proxy, BFF, embedded SPA)
- `docs/2026-03-11-service-auth-design.md` — auth system design

---

## Constraints

- **Zero framework dependency in the core** — `@workfort/ui` and `@workfort/ui/auth` must not import React, Vue, Svelte, SolidJS, or any other framework.
- **Light DOM** — No Shadow DOM. Components render into the regular DOM so CSS variables, global styles, and framework styling work naturally.
- **Custom Elements v1** — Standard `customElements.define()`. No polyfills required (baseline browser support since 2020).
- **CSS variable theming only** — Components consume `--wf-*` tokens. Never hardcode colors, spacing, or fonts.
- **Apache 2.0 licensed** — All packages in `@workfort/ui/*`.

---

## Package Structure

A single npm package with sub-path exports. Consumers install `@workfort/ui` and import from sub-paths:

```
import { WfPanel } from '@workfort/ui';              // Core Web Components
import { AuthClient } from '@workfort/ui/auth';       // Auth state
import { Panel, useAuth } from '@workfort/ui/react';  // React adapter
import { WfPanel, useAuth } from '@workfort/ui/vue';  // Vue adapter
import { auth } from '@workfort/ui/svelte';           // Svelte adapter
import { Panel, useAuth } from '@workfort/ui/solid';  // SolidJS adapter
```

All sub-paths are tree-shakeable — consumers only pay for what they import. Framework adapters list their framework as a `peerDependency` in `package.json`'s `peerDependenciesMeta` (optional), so installing `@workfort/ui` does not pull in React, Vue, Svelte, or SolidJS unless the consumer uses that sub-path.

### `@workfort/ui` (root) — Core Web Components

Lit-based Custom Elements that register globally. Once loaded, `<wf-panel>`, `<wf-button>`, etc. are available to any code on the page regardless of framework.

**Dependencies:** `lit` (runtime ~5KB gzipped). No other runtime dependencies.

**Distribution:**
- npm package — `npm install @workfort/ui`
- CDN script tag — `<script type="module" src="https://unpkg.com/@workfort/ui"></script>` (no bundler required)
- Module Federation is NOT used for sharing `@workfort/ui` — Custom Elements are global by definition. Services import the package or the shell loads it; either way, elements register once.

### `@workfort/ui/auth` — Auth State

Framework-agnostic TypeScript module. Manages auth state by talking to the better-auth session endpoint. Zero framework dependencies.

**Dependencies:** None (uses `fetch` and `EventTarget`).

### Framework Adapters

Thin wrappers providing idiomatic DX for each framework. Each adapter imports from `@workfort/ui` (core) and `@workfort/ui/auth` (auth state) — these are internal imports within the same package, not cross-package dependencies.

| Sub-path | Framework | Peer dependency | Provides |
|----------|-----------|-----------------|----------|
| `@workfort/ui/react` | React 18+ | `react` | Typed components, `useAuth()`, `useTheme()` hooks |
| `@workfort/ui/vue` | Vue 3+ | `vue` | Typed components, `useAuth()`, `useTheme()` composables |
| `@workfort/ui/svelte` | Svelte 4+/5+ | `svelte` | Typed components, auth/theme stores |
| `@workfort/ui/solid` | SolidJS 1.8+ | `solid-js` | Typed components, `useAuth()`, `useTheme()` primitives |

Adapter size: ~10-20 lines per component wrapper, ~300-400 lines total per adapter. Mechanical code.

---

## Core Web Components

### Component Inventory (Phase 1)

| Component | Element | Purpose |
|-----------|---------|---------|
| Panel | `<wf-panel>` | Container with optional label, min-width enforcement |
| List | `<wf-list>` | Scrollable list with selection state |
| ListItem | `<wf-list-item>` | List item with active state, leading/trailing slots |
| TextInput | `<wf-text-input>` | Single-line input with placeholder |
| Badge | `<wf-badge>` | Count indicator |
| Button | `<wf-button>` | Action trigger (text, icon, or both) |
| Divider | `<wf-divider>` | Horizontal separator |
| ScrollArea | `<wf-scroll-area>` | Scrollable container with themed scrollbar |
| Skeleton | `<wf-skeleton>` | Loading placeholder |
| StatusDot | `<wf-status-dot>` | Online/offline indicator |
| ErrorFallback | `<wf-error-fallback>` | Displays a styled error message (not a React-style error boundary — error catching is the framework's job; this is just the display component) |

### Light DOM Rendering

All components render in the light DOM. In Lit, this means overriding `createRenderRoot()`:

```ts
createRenderRoot() {
  return this;
}
```

Benefits:
- `--wf-*` CSS variables inherit naturally — no shadow boundary to cross
- Adopters can style internal elements with global CSS
- No `::part()` or `::slotted()` workarounds
- Framework CSS-in-JS solutions work without special configuration

Trade-off: No style encapsulation. Components must use namespaced CSS classes (e.g., `.wf-panel`, `.wf-button`) to avoid collisions with consumer styles. This is acceptable — the `wf-` prefix provides sufficient namespacing.

### Component Principles

Unchanged from the original spec:

- **Structural, not styled** — Define layout and behavior. Visual appearance comes from CSS custom properties.
- **Themed via CSS variables** — All colors, spacing, fonts reference `--wf-*` tokens. Never hardcode visual values.
- **Graceful degradation** — Components that can't render at minimum size show a clear error message.
- **Accessible** — Keyboard navigation, ARIA attributes, focus management built in.
- **Children-based composition** — Light DOM means children are rendered directly (no `<slot>` projection — that's a Shadow DOM feature). Components use `this.children` or Lit's `@queryAssignedElements` to interact with child content. Named content areas use a `data-wf` attribute convention (e.g., `<wf-list-item><span data-wf="trailing">...</span></wf-list-item>` where the component queries children by `[data-wf="trailing"]` as a CSS selector). We avoid reusing the `slot` attribute to prevent confusion with Shadow DOM semantics.

### Attribute/Property Convention

Lit reactive properties map to both HTML attributes and JS properties:

```ts
@property({ type: String }) label = '';
@property({ type: Boolean, reflect: true }) active = false;
@property({ type: Number }) count = 0;
```

- **Strings/numbers/booleans** are attributes (works in all frameworks and plain HTML)
- **Complex objects** (arrays, callbacks) are properties only (set via JS, not HTML attributes)
- Events use standard `CustomEvent` dispatch — frameworks handle these natively

### Theme Tokens

Same contract as the original spec. The shell sets these on `:root`:

```css
/* Colors */
--wf-bg, --wf-bg-secondary
--wf-text, --wf-text-secondary, --wf-text-muted
--wf-border, --wf-accent

/* Spacing */
--wf-space-xs, --wf-space-sm, --wf-space-md, --wf-space-lg, --wf-space-xl

/* Typography */
--wf-font-sans, --wf-font-mono
--wf-font-size-xs, --wf-font-size-sm, --wf-font-size-base, --wf-font-size-lg

/* Borders */
--wf-radius-sm, --wf-radius-md, --wf-radius-lg
```

Components reference these exclusively. Example:

```css
.wf-panel {
  background: var(--wf-bg);
  border: 1px solid var(--wf-border);
  border-radius: var(--wf-radius-md);
  padding: var(--wf-space-md);
}
```

---

## Auth Package

### `@workfort/ui/auth`

A plain TypeScript package with zero framework dependencies. Manages auth state by talking to the better-auth session endpoint through the BFF proxy.

### API

```ts
import { AuthClient } from '@workfort/ui/auth';

const auth = new AuthClient();

// Lifecycle
await auth.init();           // Fetches session from GET /api/auth/v1/session
await auth.refresh();        // Re-fetches session (e.g., after tab becomes visible)
await auth.logout();         // Clears session, redirects to login

// State (synchronous reads after init)
auth.getUser();              // User | null
auth.getSession();           // Session | null
auth.isAuthenticated;        // boolean (getter)

// Events
auth.on('change', (user: User | null) => { ... });
auth.on('logout', () => { ... });
auth.off('change', handler);

// Event type signatures:
// type AuthEvents = {
//   change: (user: User | null) => void;
//   logout: () => void;
// };
```

### Types

```ts
interface User {
  id: string;
  username: string;
  name: string;
  displayName: string;
  type: 'user' | 'agent' | 'service';
}

interface Session {
  id: string;
  expiresAt: string;   // ISO 8601 timestamp
  refreshedAt: string;  // When the session was last validated
}
```

The `User` interface maps the better-auth session endpoint response (camelCase fields). Note: JWT claims use snake_case (`display_name`), but `AuthClient` talks to the session endpoint, not the JWT — so the camelCase mapping is intentional. See `docs/2026-03-11-better-auth-setup.md` for both formats.

### How It Works

1. Shell creates a singleton `AuthClient` and calls `init()` on boot
2. `init()` calls `GET /api/auth/v1/session` (BFF proxy strips `/api/auth`, forwards `GET /v1/session` to the auth service with cookies)
3. If the session is valid, stores `User` and `Session` in memory, emits `change`
4. If unauthorized (401), sets state to `null` — the shell shows the login UI
5. On `logout()`, calls the better-auth signout endpoint, clears state, emits `logout`

### `init()` Error Handling

`init()` can fail in three ways:

- **401 Unauthorized** — Session expired or missing. Sets state to `null`, does not throw. The shell checks `isAuthenticated` and shows login UI. This is a normal flow, not an error.
- **Network error** (auth service unreachable) — Throws an `AuthInitError` with `cause` set to the original error. The shell catches this and shows a connection error screen. `init()` can be retried.
- **Unexpected response** (non-JSON, 5xx) — Same as network error: throws `AuthInitError`.

```ts
class AuthInitError extends Error {
  constructor(message: string, options?: { cause?: unknown }) {
    super(message, options);
    this.name = 'AuthInitError';
  }
}
```

### Session Refresh Strategy

The `AuthClient` does not proactively refresh sessions — better-auth extends the session server-side on activity (the BFF proxy's cookie-forwarding keeps the session alive). The client re-validates on demand:

- **`init()`** validates once at boot
- **`refresh()`** re-fetches the session (used after returning from a background tab or after network recovery)
- **Visibility change** — When the document becomes visible after being hidden for >5 minutes, `refresh()` is called automatically
- **If `refresh()` gets a 401** — Session expired server-side. Sets state to `null`, emits `logout`

No polling. No timers. The BFF proxy handles JWT refresh (14-minute cycle); the frontend only cares about session validity.

### Singleton Pattern

The shell owns the `AuthClient` instance. It's made available via a module-level singleton:

```ts
// @workfort/ui/auth
let instance: AuthClient | null = null;

export function getAuthClient(): AuthClient {
  if (!instance) {
    instance = new AuthClient();
  }
  return instance;
}
```

This works because `@workfort/ui` is a single npm package — all sub-paths (`/auth`, `/react`, `/solid`, etc.) resolve to the same package instance, so `getAuthClient()` always returns the same object. If the package were split into separate npm packages, this guarantee would break (each package could get its own module instance). The single-package design is load-bearing here.

Framework adapters call `getAuthClient()` internally — consumers never need to wire this up.

### Future: `@workfort/ui/services`

Follows the same pattern. A `ServiceClient` with `getServices()`, `isHealthy(name)`, and events on status changes. Framework adapters wrap it identically. Not in scope for this spec — noted here to confirm the architecture supports it.

---

## Framework Adapters

### What Each Adapter Provides

1. **Typed component wrappers** — Framework-native components that render the underlying `<wf-*>` element with TypeScript prop types and event types
2. **`useAuth()`** — Hook/composable/store that subscribes to `AuthClient` events and triggers framework-specific reactivity
3. **`useTheme()`** — Reads the current theme mode by observing `document.documentElement`'s `data-theme` attribute (set by the shell). Returns `'dark' | 'light'`. Uses a `MutationObserver` on the attribute — when the shell toggles theme, all adapter hooks re-render.

### React Adapter (`@workfort/ui/react`)

```tsx
// Component wrappers use React.forwardRef + prop forwarding
import { Panel, List, ListItem, Badge, useAuth } from '@workfort/ui/react';

function Sidebar() {
  const { user, isAuthenticated } = useAuth();
  return (
    <Panel label="Channels">
      <List>
        <ListItem trailing={<Badge count={3} />}>general</ListItem>
      </List>
    </Panel>
  );
}
```

React has historically had friction with Custom Elements (event handling, boolean attributes). The wrapper handles this:
- Maps `onX` React event props to `addEventListener` on the Custom Element
- Handles boolean attribute semantics correctly
- Provides `ref` forwarding to the underlying element

### Vue Adapter (`@workfort/ui/vue`)

```vue
<script setup>
import { WfPanel, WfList, WfListItem, useAuth } from '@workfort/ui/vue';
const { user } = useAuth();
</script>
<template>
  <WfPanel label="Channels">
    <WfList>
      <WfListItem>general</WfListItem>
    </WfList>
  </WfPanel>
</template>
```

Vue 3 has good native Web Component support. The adapter adds TypeScript types and wraps `useAuth()` as a composable returning reactive refs.

### Svelte Adapter (`@workfort/ui/svelte`)

```svelte
<script>
import { WfPanel, WfList, WfListItem } from '@workfort/ui/svelte';
import { auth } from '@workfort/ui/svelte';

const user = auth.user;  // Svelte store
</script>

<WfPanel label="Channels">
  <WfList>
    <WfListItem>general</WfListItem>
  </WfList>
</WfPanel>
```

Svelte has native Web Component support. The adapter provides Svelte stores for auth/theme.

### SolidJS Adapter (`@workfort/ui/solid`)

```tsx
import { Panel, List, ListItem, useAuth } from '@workfort/ui/solid';

function Sidebar() {
  const { user } = useAuth();
  return (
    <Panel label="Channels">
      <List>
        <ListItem>general</ListItem>
      </List>
    </Panel>
  );
}
```

SolidJS has excellent Web Component interop. The adapter wraps `useAuth()` as a SolidJS primitive (signal-based). The shell uses this adapter.

### Vanilla JS (No Adapter)

```html
<script type="module">
  import '@workfort/ui';
  import { getAuthClient } from '@workfort/ui/auth';

  const auth = getAuthClient();
  await auth.init();

  auth.on('change', (user) => {
    document.getElementById('username').textContent = user?.username ?? 'anonymous';
  });
</script>

<wf-panel label="Channels">
  <wf-list>
    <wf-list-item>general</wf-list-item>
  </wf-list>
</wf-panel>
```

No adapter needed. Custom Elements work natively in all browsers.

---

## Changes to Web UI Design Spec

This spec amends `docs/2026-03-11-web-ui-design.md`:

### Stack table (line 22)

**Before:** `Frontend framework | SolidJS (required for all service UIs)`
**After:** `Frontend framework | SolidJS (shell only). Service UIs can use any framework — @workfort/ui provides Web Components + framework adapters for React, Vue, Svelte, SolidJS.`

### Module Federation sharing (lines 226-237)

**Before:** Shell shares `@workfort/ui` and `solid-js` as singletons via federation.
**After:** Shell does NOT share `@workfort/ui` via federation. Custom Elements register globally — once loaded, they're available to all code on the page. Services import `@workfort/ui` (or the adapter for their framework) as a regular npm dependency. Module Federation is still used for loading service remotes at runtime; it's just not the distribution mechanism for the component library.

### Service remote requirements (line 80)

**Before:** "Uses `@workfort/ui` for structural components (inherits shell theme)"
**After:** "Uses `@workfort/ui` Web Components (directly or via framework adapter) for structural components. Inherits shell theme via CSS custom properties."

### Component inventory (lines 249-274)

Same components, now described as Custom Elements (`<wf-panel>`, `<wf-button>`, etc.) rather than SolidJS components.

### Usage example (lines 278-312)

Replace SolidJS example with multiple framework examples showing the same UI built with React, Vue, and vanilla JS.

---

## Testing Strategy

### Core Web Components (`@workfort/ui`)
- **Unit tests with `@open-wc/testing`** — the standard Web Component testing library
- Test each component: rendering, attribute changes, event dispatch, child composition, accessibility
- Run in a real browser environment (Playwright or `@web/test-runner`)

### Auth Package (`@workfort/ui/auth`)
- **Unit tests with Vitest** — mock `fetch` for session endpoints
- Test: init success/failure, event emission, singleton behavior, logout

### Framework Adapters
- **Unit tests with each framework's testing library** — React Testing Library, Vue Test Utils, Svelte Testing Library, Solid Testing Library
- Test: prop forwarding, event handling, hook reactivity, TypeScript types

### Integration
- **Storybook** with stories for each component in each framework
- Visual regression via Chromatic or similar (future)

---

## Build and Workspace

### Source Layout

Single package, sub-path exports. One `package.json` at the root:

```
package.json                  # name: @workfort/ui
tsconfig.json
vite.config.ts

src/
  index.ts                    # Registers all elements, re-exports
  components/
    panel.ts                  # WfPanel extends LitElement
    button.ts
    list.ts
    list-item.ts
    text-input.ts
    badge.ts
    divider.ts
    scroll-area.ts
    skeleton.ts
    status-dot.ts
    error-fallback.ts
  styles/
    tokens.css                # CSS for --wf-* fallback values

  auth/
    index.ts                  # Re-exports AuthClient, types, getAuthClient
    client.ts                 # AuthClient implementation
    types.ts                  # User, Session, AuthInitError

  react/
    index.ts                  # Re-exports all wrappers + hooks
    components.tsx            # React wrappers for each WC
    use-auth.ts               # useAuth() hook
    use-theme.ts              # useTheme() hook

  vue/
    index.ts
    components.ts             # Vue component wrappers
    use-auth.ts               # useAuth() composable
    use-theme.ts              # useTheme() composable

  svelte/
    index.ts
    components/               # Svelte component wrappers
    auth.ts                   # Auth store
    theme.ts                  # Theme store

  solid/
    index.ts
    components.tsx            # SolidJS wrappers
    use-auth.ts               # useAuth() primitive
    use-theme.ts              # useTheme() primitive

tests/
  components/                 # @web/test-runner tests for WCs
  auth/                       # Vitest tests for auth
  react/                      # React Testing Library
  vue/                        # Vue Test Utils
  svelte/                     # Svelte Testing Library
  solid/                      # Solid Testing Library
```

### package.json (key fields)

```json
{
  "name": "@workfort/ui",
  "type": "module",
  "exports": {
    ".": "./dist/index.js",
    "./auth": "./dist/auth/index.js",
    "./react": "./dist/react/index.js",
    "./vue": "./dist/vue/index.js",
    "./svelte": "./dist/svelte/index.js",
    "./solid": "./dist/solid/index.js"
  },
  "dependencies": {
    "lit": "^3.0.0"
  },
  "peerDependencies": {
    "react": "^18.0.0 || ^19.0.0",
    "vue": "^3.3.0",
    "svelte": "^4.0.0 || ^5.0.0",
    "solid-js": "^1.8.0"
  },
  "peerDependenciesMeta": {
    "react": { "optional": true },
    "vue": { "optional": true },
    "svelte": { "optional": true },
    "solid-js": { "optional": true }
  }
}
```

All framework peer dependencies are optional. Installing `@workfort/ui` only pulls in `lit`. Framework adapters work when the consumer has the relevant framework installed.

### Build Tools

| Tool | Purpose |
|------|---------|
| Lit | Web Component authoring (core) |
| Vite | Build (library mode, multiple entry points) |
| Vitest | Tests for auth and adapters |
| `@web/test-runner` | Tests for core Web Components (needs real DOM) |
| TypeScript | All source |

---

## Migration from SolidJS Plan

The existing SolidJS implementation plan (`docs/plans/2026-03-11-workfort-ui.md`) is superseded. Key differences:

| Aspect | SolidJS Plan | Web Components Plan |
|--------|-------------|-------------------|
| Component engine | SolidJS | Lit |
| Rendering | Virtual DOM (SolidJS reactive) | Light DOM (Lit templates) |
| Distribution | Module Federation singleton | npm package, Custom Elements register globally |
| Framework support | SolidJS only | Any framework + 4 adapters |
| Auth/theme | SolidJS context providers | Framework-agnostic core + adapter hooks |
| npm package | 1 | 1 (with sub-path exports for auth + 4 adapters) |
| Build output | SolidJS library (ESM) | Custom Elements (ESM) + adapter sub-paths |

---

## Amendment: Package Split (2026-03-12)

> This section amends the package structure described above. The original spec specified a single `@workfort/ui` package with sub-path exports. That decision has been reversed.

### What changed

**Auth moved to `@workfort/auth` in the passport repo.** The `@workfort/ui/auth` sub-path described in this spec no longer exists. All framework adapters now import from `@workfort/auth` as an external dependency.

### Why the single-package design no longer applies

The singleton justification (line 276 above) stated:

> "If the package were split into separate npm packages, this guarantee would break (each package could get its own module instance). The single-package design is load-bearing here."

This was correct when `AuthClient` lived inside `@workfort/ui`. Now that auth is `@workfort/auth` — a separate npm package — the singleton is already managed across package boundaries. npm deduplication ensures all packages that depend on `@workfort/auth@^0.0.1` resolve to the same module instance. The constraint that made single-package load-bearing is gone.

### New package structure

The monolithic `@workfort/ui` is replaced by 5 independent npm packages in a pnpm workspace under `web/`:

| Package | Contents | Dependencies | Peer Dependencies |
|---------|----------|-------------|-------------------|
| `@workfort/ui` | Core Lit Web Components + CSS tokens | `lit` | — |
| `@workfort/ui-react` | React component wrappers + `useAuth()` + `useTheme()` hooks | `@workfort/auth` | `@workfort/ui`, `react` |
| `@workfort/ui-vue` | `useAuth()` + `useTheme()` composables | `@workfort/auth` | `@workfort/ui`, `vue` |
| `@workfort/ui-svelte` | Auth + theme Svelte stores | `@workfort/auth` | `@workfort/ui`, `svelte` |
| `@workfort/ui-solid` | `useAuth()` + `useTheme()` Solid primitives | `@workfort/auth` | `@workfort/ui`, `solid-js` |

Consumer imports change:

```
// Before (sub-path exports):
import { WfPanel } from '@workfort/ui';
import { Panel, useAuth } from '@workfort/ui/react';
import { auth } from '@workfort/ui/svelte';

// After (separate packages):
import { WfPanel } from '@workfort/ui';
import { Panel, useAuth } from '@workfort/ui-react';
import { auth } from '@workfort/ui-svelte';
```

`@workfort/ui` remains the core — framework packages depend on it as a peer dependency. The Web Components, CSS tokens, light DOM rendering, and theme contract are all unchanged.

### Sections of this spec affected

- **Package Structure** (line 28): Single package → workspace with 5 packages
- **Auth Package** (line 169): `@workfort/ui/auth` no longer exists — see `@workfort/auth` in the passport repo
- **Singleton Pattern** (line 260): Constraint removed — auth singleton managed by `@workfort/auth`
- **Framework Adapters** (line 286): Import paths change from `@workfort/ui/{framework}` to `@workfort/ui-{framework}`
- **Build and Workspace** (line 449): Single Vite build → pnpm workspace with per-package builds

### Implementation plan

See `docs/plans/2026-03-12-ui-package-split.md`.
