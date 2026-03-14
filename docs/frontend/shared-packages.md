# Shared Packages

These packages are shared across service frontends via Module Federation singleton sharing. Each service imports from them without bundling its own copy.

---

## `@workfort/ui`

Lit-based web components using light DOM (no shadow DOM). All components extend `WfElement` and register as standard custom elements.

Import the stylesheet once in the shell:

```js
import '@workfort/ui/style.css';
```

### Components

| Tag | Properties | Events |
|-----|-----------|--------|
| `wf-panel` | `label: string` | — |
| `wf-button` | `variant: 'text'\|'filled'`, `disabled: boolean` | `wf-click` |
| `wf-badge` | `count: number` (hidden when 0) | — |
| `wf-status-dot` | `status: 'online'\|'offline'\|'away'` | — |
| `wf-skeleton` | `width: string` (default `'100%'`), `height: string` (default `'1em'`) | — |
| `wf-divider` | — | — |
| `wf-text-input` | `placeholder: string`, `value: string`, `disabled: boolean` | `wf-input` (detail: `{ value }`), `wf-change` (detail: `{ value }`) |
| `wf-list` | — | — |
| `wf-list-item` | `active: boolean` | `wf-select` |
| `wf-scroll-area` | — | — |
| `wf-error-fallback` | `title: string`, `message: string` | — |
| `wf-banner` | `variant: 'error'\|'warning'\|'info'`, `dismissible: boolean`, `headline: string`, `details: string` | `wf-dismiss` |
| `wf-toast` | `variant: 'error'\|'warning'\|'info'\|'success'`, `sticky: boolean`, `duration: number` (ms, default 5000) | `wf-dismiss` |
| `wf-toast-container` | `position: 'top-right'\|'top-left'\|'bottom-right'\|'bottom-left'` | — |

Notes:

- `wf-banner` — `details` is expandable via a toggle button; close button only rendered when `dismissible` is true.
- `wf-toast` — auto-dismisses after `duration` ms unless `sticky` is set; always dispatches `wf-dismiss` and removes itself on close.
- `wf-list-item` — children with `data-wf="trailing"` are automatically wrapped in a trailing slot container.

---

## Framework Adapters

Each adapter provides auth and theme integration against the `@workfort/auth` singleton. All read from `data-theme` on `document.documentElement` for theme state.

### `@workfort/ui-solid`

```ts
import { useAuth, useTheme } from '@workfort/ui-solid';

const { user, isAuthenticated } = useAuth();
// user: Accessor<User | null>
// isAuthenticated: () => boolean

const theme = useTheme();
// theme: Accessor<'dark' | 'light'>
```

SolidJS handles `wf-*` custom elements natively in JSX. No wrappers needed.

### `@workfort/ui-react`

```ts
import { useAuth, useTheme } from '@workfort/ui-react';
import { Panel, Button, Badge, StatusDot, Skeleton, Divider,
         TextInput, List, ListItem, ScrollArea, ErrorFallback } from '@workfort/ui-react';

const { user, isAuthenticated } = useAuth();
// user: User | null
// isAuthenticated: boolean  (plain value, not a hook)

const theme = useTheme();
// theme: 'dark' | 'light'
```

Both hooks use `useSyncExternalStore` internally. React 18 does not forward `onX` props to custom element `addEventListener` calls, so the adapter provides typed React wrapper components for all `wf-*` elements (except `wf-banner`, `wf-toast`, and `wf-toast-container`, which have no wrappers — use the HTML tags directly). Event props use camelCase conversion: `onWfClick` → `wf-click`.

### `@workfort/ui-vue`

```ts
import { useAuth, useTheme } from '@workfort/ui-vue';

const { user, isAuthenticated } = useAuth();
// user: Readonly<Ref<User | null>>
// isAuthenticated: Readonly<Ref<boolean>>

const theme = useTheme();
// theme: Readonly<Ref<'dark' | 'light'>>
```

Vue handles `wf-*` custom elements natively. Add to your Vue config:

```js
compilerOptions: {
  isCustomElement: (tag) => tag.startsWith('wf-'),
}
```

### `@workfort/ui-svelte`

```ts
import { auth, theme } from '@workfort/ui-svelte';

// auth.user: Readable<User | null>
// auth.isAuthenticated: Readable<boolean>  (derived store)
// theme: Readable<'dark' | 'light'>
```

Svelte uses the store pattern rather than hooks. `auth` is an object with two readable stores. Svelte handles `wf-*` custom elements natively in templates.

---

## Framework CE Support Summary

| Framework | Native `wf-*` support | Needs wrappers |
|-----------|----------------------|----------------|
| Solid | Yes | No |
| Vue | Yes (requires `isCustomElement` config) | No |
| Svelte | Yes | No |
| React | No (React 18) | Yes — `Panel`, `Button`, `Badge`, `StatusDot`, `Skeleton`, `Divider`, `TextInput`, `List`, `ListItem`, `ScrollArea`, `ErrorFallback` |
