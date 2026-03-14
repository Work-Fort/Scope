# Authentication — Service Frontend Developer Guide

## BFF Pattern

The shell's Go backend acts as a Backend-for-Frontend. Session cookies are never forwarded to downstream services. Instead, `bffMiddleware` (`internal/infra/httpapi/handler.go`) calls `tc.Token(r)` to exchange the incoming cookie for a short-lived JWT, then sets `Authorization: Bearer <token>` before proxying the request.

```
Browser ──[session cookie]──> shell (Go) ──[Bearer JWT]──> your service
```

Your service receives a standard Bearer JWT. It never sees cookies, and the frontend never sees tokens.

WebSocket upgrades follow the same path: the cookie is exchanged before the upgrade handshake.

## `@workfort/auth` Client API

All adapters use a singleton `AuthClient`. Obtain it with:

```ts
import { getAuthClient } from '@workfort/auth';

const client = getAuthClient();
```

`getAuthClient()` returns the same instance on every call within a module graph. See [Shared Singleton](#shared-singleton) for why this matters.

### Methods and Properties

| Member | Signature | Description |
|---|---|---|
| `getUser()` | `() => User \| null` | Current user, or `null` if unauthenticated |
| `getSession()` | `() => Session \| null` | Current session, or `null` if unauthenticated |
| `isAuthenticated` | `boolean` (getter) | `true` when a user is set |
| `init()` | `async () => void` | Fetches session from auth service, sets up auto-refresh on tab visibility |
| `logout()` | `async () => void` | POSTs sign-out, clears state, emits `logout` then `change` |
| `on(event, listener)` | — | Subscribe to events |
| `off(event, listener)` | — | Unsubscribe |

`init()` must be called once at app startup before rendering auth-dependent UI. It throws `AuthInitError` if the auth service is unreachable; a 401 response is treated as unauthenticated (no throw).

Auto-refresh: after the tab has been hidden for more than 5 minutes, the next visibility event triggers a silent `refresh()`. If the session has expired, the `logout` event fires.

`logout()` is best-effort on the network request — it clears local state and emits events regardless of fetch success.

### Events

| Event | Payload | When |
|---|---|---|
| `change` | `User \| null` | Session fetched or cleared |
| `logout` | `void` | `logout()` called, or session expired on background refresh |

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
  expiresAt: string;   // ISO 8601
  refreshedAt: string; // ISO 8601
}
```

## Using Auth Per Framework

Framework adapters wrap the singleton client and expose reactive state. See `docs/frontend/shared-packages.md` for the full adapter API.

### SolidJS (`@workfort/solid`)

`useAuth()` returns a reactive signal for the current user.

```ts
import { useAuth } from '@workfort/solid';

function MyComponent() {
  const { user, isAuthenticated } = useAuth();

  return (
    <Show when={isAuthenticated()} fallback={<p>Not logged in</p>}>
      <p>Hello, {user()?.displayName}</p>
    </Show>
  );
}
```

`user` is a SolidJS signal (`Accessor<User | null>`). `isAuthenticated` is a derived accessor (`() => user() !== null`). Both update reactively when auth state changes.

Cleanup is handled automatically via `onCleanup` — no manual unsubscription needed.

### Other Frameworks

Vue and React adapters follow the same pattern: call the framework-specific hook, receive reactive user state. The underlying singleton and event system are identical. Consult `docs/frontend/shared-packages.md`.

## Per-Fort Cookie Scoping

Session cookies are scoped to `Path: /forts/{fortName}/`. This is enforced by `NewAuthProxy` (`internal/infra/httpapi/proxy.go`), which rewrites every `Set-Cookie` path on responses from the auth service.

Consequences:
- Logging into fort `alpha` does not authenticate fort `beta`.
- The browser will not send fort `alpha`'s cookie on fort `beta`'s requests.

### Error Responses from `bffMiddleware`

| Condition | HTTP Status | Side Effect |
|---|---|---|
| No session cookie / not logged in | 401 | — |
| Session expired | 401 | Cookie cleared (`Max-Age: -1`) |
| Auth service unreachable | 502 | — |

On 401, redirect the user to the fort's login page. On 502, display a service-unavailable message; do not redirect.

## Shared Singleton

`@workfort/auth` maintains a module-level singleton. If your service frontend bundles its own copy of `@workfort/auth`, it gets a separate instance with no shared state, breaking cross-app auth coordination.

In your Vite module federation config, declare `@workfort/auth` as a shared singleton with `import: false`:

```ts
// vite.config.ts
federation({
  shared: {
    '@workfort/auth': { singleton: true, import: false },
  },
})
```

`import: false` means the host (shell) provides the module; your service does not include it in its own bundle. Never bundle your own copy.
