# Per-Fort Authentication — Design Spec

**Goal:** The Tauri Rust proxy manages per-fort JWTs with background refresh, matching the Go BFF's per-fort token model. Also standardizes the refresh buffer to 5 minutes across both platforms.

**Key Principle:** Every fort has its own passport instance, its own auth service URL, and its own token lifecycle. The proxy keeps all fort tokens alive in the background. Switching forts is instant — no re-auth. The webview never touches tokens.

---

## Token Store

```rust
struct FortTokens {
    jwt: String,
    refresh_token: String,
    expiry: Instant,
    auth_url: String,       // that fort's passport instance URL
}

struct TokenStore {
    forts: HashMap<String, FortTokens>,
}
```

Wrapped in `Arc<Mutex<TokenStore>>` on `AppState`, replacing the current single-token `Arc<Mutex<Option<String>>>`.

---

## Proxy Behavior

| Route | Auth | Behavior |
|-------|------|----------|
| `/api/forts` | None | Pass through, no token attached |
| `/forts/{fort}/api/*` | Per-fort JWT | Look up fort's token, attach `Authorization: Bearer`. If no token → return 401. If upstream returns 401 → refresh and retry once. |
| WebSocket `/forts/{fort}/ws/*` | Per-fort JWT | Same token lookup. Only active fort has open WS connections. |

---

## Auth Flow (per fort)

```
User picks fort "acme"
  → Shell requests /forts/acme/api/services
  → Proxy checks TokenStore for "acme"
  → No token found → returns 401 to shell
  → Shell shows login screen for "acme"
  → User enters credentials
  → Shell calls invoke("login", { fort: "acme", email, password })
  → Rust POSTs to acme's passport: {auth_url}/v1/auth/login
  → Receives JWT + refresh token + expiry
  → Stores in TokenStore under "acme"
  → Returns user info to shell
  → Shell retries /forts/acme/api/services → succeeds
```

---

## Background Token Refresh

A Tokio background task runs every 60 seconds:

```
for each (fort_name, tokens) in store.forts:
    if tokens.expiry - now < 5 minutes:
        POST {tokens.auth_url}/v1/token/refresh
        if success:
            update jwt, refresh_token, expiry
        if failure:
            log warning, retry on next cycle
            if token is actually expired:
                remove from store (user will need to re-auth)
```

**Refresh buffer: 5 minutes before expiry.** This applies to both the Rust proxy and the Go BFF (currently 1 minute, needs updating).

---

## Fort Auth Service Discovery

When the user logs into a fort, the proxy needs to know that fort's auth service URL. Two approaches:

**Option A (recommended):** The `/api/forts` response includes the pylon URL for each fort. The auth service is at `{pylon}/auth`. The proxy stores this when the fort list is fetched.

**Option B:** After login, the proxy calls the fort's service tracker to discover the auth service URL (like the Go BFF does with `tracker.ServiceByName("auth")`).

Option A is simpler — the pylon URL is already in the fort list response.

---

## Tauri Commands (updated)

```rust
#[tauri::command]
async fn login(fort: String, email: String, password: String, state: State<'_, AppState>) -> Result<UserInfo, String>

#[tauri::command]
async fn logout(fort: String, state: State<'_, AppState>) -> Result<(), String>

#[tauri::command]
async fn get_user(fort: String, state: State<'_, AppState>) -> Result<Option<UserInfo>, String>
```

All commands are now fort-scoped. `login` stores the tokens under the fort name. `logout` removes that fort's tokens. `get_user` checks if we have a valid token for that fort and returns the user info.

---

## Active Fort & WebSocket Management

- Only one fort is "active" at a time (the one the user is viewing)
- When switching forts, the proxy closes WebSocket connections for the previous fort
- Token refresh runs for ALL forts regardless of which is active
- Future: notification service maintains separate long-lived connections to each fort

---

## Go BFF Change

Update `internal/infra/httpapi/bff.go`:

```go
const (
    refreshBefore = 5 * time.Minute  // was 1 * time.Minute
)
```

One line change. Makes both platforms consistent.

---

## Changes Required

### Rust (`src-tauri/`)

| File | Change |
|------|--------|
| `src/proxy.rs` | Replace single `TokenStore` with `HashMap<String, FortTokens>`. Update `proxy_request()` to extract fort from path. Update `try_refresh()` to use per-fort auth URL. |
| `src/auth.rs` | Add `fort` parameter to `login`, `logout`, `get_user` commands. Store/retrieve tokens per fort. |
| `src/lib.rs` | Add background refresh task via `tokio::spawn`. Register updated commands. |

### Go (`internal/`)

| File | Change |
|------|--------|
| `internal/infra/httpapi/bff.go` | Change `refreshBefore` from 1 minute to 5 minutes |

### Shell (`web/shell/`)

| File | Change |
|------|--------|
| `src/boot/tauri.ts` | `authenticate()` becomes fort-aware — checks token for the current fort |
| `src/app.tsx` | Login screen passes fort name to `invoke("login", { fort, ... })` |
