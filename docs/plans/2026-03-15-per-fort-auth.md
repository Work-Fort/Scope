# Per-Fort Authentication — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the single-token proxy with per-fort token management in the Tauri Rust proxy. Each fort gets its own JWT, refresh token, and auth URL. A background task keeps all tokens alive. Also standardize the Go BFF refresh buffer to 5 minutes.

**Tech Stack:** Rust (Tauri 2, reqwest, tokio), Go

**Spec:** `docs/per-fort-auth-design.md`

---

## Chunk 1: Rust Token Store Refactor

### Task 1: Replace single token store with per-fort HashMap

**Why:** The current `TokenStore` holds one JWT for the entire app. We need a `HashMap<String, FortTokens>` so each fort has its own token lifecycle.

**Files:**
- Edit: `src-tauri/src/proxy.rs`

- [ ] **Step 1: Add tokio and log dependencies to Cargo.toml**

We need `tokio` (for `Instant` and the background task) and `log` (for refresh logging).

```toml
# src-tauri/Cargo.toml — add to [dependencies]
tokio = { version = "1", features = ["time", "rt"] }
log = "0.4"
```

```bash
cd /home/kazw/Work/WorkFort/scope/lead/src-tauri && cargo check 2>&1 | tail -5
```

Expected: compiles clean.

- [ ] **Step 2: Replace TokenStore and AppState in proxy.rs**

Replace the entire `TokenStore` struct and its impl with the new per-fort version. Replace `AppState` to use the new store.

```rust
// src-tauri/src/proxy.rs — full file replacement

use reqwest::Client;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use url::Url;

/// Per-fort token data. Each fort has its own passport instance.
#[derive(Clone, Debug)]
pub struct FortTokens {
    pub jwt: String,
    pub refresh_token: String,
    pub expiry: Instant,
    pub auth_url: String, // e.g. "https://acme.example.com/auth"
}

/// In-memory token store. Keyed by fort name. Cleared when app is killed.
/// JWTs and refresh tokens live here — never in the webview.
#[derive(Clone)]
pub struct TokenStore {
    pub forts: Arc<Mutex<HashMap<String, FortTokens>>>,
}

impl TokenStore {
    pub fn new() -> Self {
        Self {
            forts: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn get(&self, fort: &str) -> Option<FortTokens> {
        self.forts.lock().unwrap().get(fort).cloned()
    }

    pub fn set(&self, fort: &str, tokens: FortTokens) {
        self.forts.lock().unwrap().insert(fort.to_string(), tokens);
    }

    pub fn remove(&self, fort: &str) {
        self.forts.lock().unwrap().remove(fort);
    }

    /// Returns a snapshot of all fort names and their tokens.
    pub fn all(&self) -> Vec<(String, FortTokens)> {
        self.forts.lock().unwrap().iter().map(|(k, v)| (k.clone(), v.clone())).collect()
    }
}

/// State shared across the Tauri app: HTTP client, token store, API base URL.
#[derive(Clone)]
pub struct AppState {
    pub client: Client,
    pub tokens: TokenStore,
    pub api_base: Url,
}

impl AppState {
    pub fn new(api_base_url: &str) -> Self {
        let client = Client::builder()
            .timeout(Duration::from_secs(10))
            .build()
            .expect("Failed to build HTTP client");

        Self {
            client,
            tokens: TokenStore::new(),
            api_base: Url::parse(api_base_url).expect("Invalid API base URL"),
        }
    }
}

/// Determines whether a request path should be proxied to the API backend.
pub fn should_proxy(path: &str) -> bool {
    path.starts_with("/api/") || path.starts_with("/forts/")
}

/// Extracts the fort name from a path like `/forts/{fort}/api/...`.
/// Returns None for paths that don't match the pattern (e.g. `/api/forts`).
pub fn extract_fort_name(path: &str) -> Option<&str> {
    let path = path.strip_prefix("/forts/")?;
    path.split('/').next().filter(|s| !s.is_empty())
}

/// Proxies a request to the API backend, attaching the per-fort JWT if available.
/// Returns the response body bytes, status code, and content-type.
pub async fn proxy_request(
    state: &AppState,
    method: &str,
    path: &str,
    query: Option<&str>,
    body: Option<Vec<u8>>,
    content_type: Option<&str>,
) -> Result<(Vec<u8>, u16, String), String> {
    // Build target URL
    let mut target = state.api_base.clone();
    target.set_path(path);
    if let Some(q) = query {
        target.set_query(Some(q));
    }

    // Build request
    let reqwest_method = method.parse::<reqwest::Method>()
        .map_err(|e| format!("Invalid method: {e}"))?;
    let mut req = state.client.request(reqwest_method, target);

    // Attach per-fort JWT if this is a fort-scoped request
    if let Some(fort_name) = extract_fort_name(path) {
        if let Some(tokens) = state.tokens.get(fort_name) {
            req = req.header("Authorization", format!("Bearer {}", tokens.jwt));
        }
    }
    // /api/forts and other /api/* paths pass through without auth

    // Attach body and content-type
    if let Some(b) = body {
        if let Some(ct) = content_type {
            req = req.header("Content-Type", ct);
        }
        req = req.body(b);
    }

    // Execute
    let resp = req.send().await.map_err(|e| format!("Proxy error: {e}"))?;
    let status = resp.status().as_u16();
    let ct = resp.headers()
        .get("content-type")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("application/octet-stream")
        .to_string();
    let bytes = resp.bytes().await.map_err(|e| format!("Read body: {e}"))?;

    Ok((bytes.to_vec(), status, ct))
}

/// Attempts to refresh the JWT for a specific fort using its stored refresh token.
/// Returns true if refresh succeeded, false otherwise.
pub async fn try_refresh(state: &AppState, fort_name: &str) -> bool {
    let tokens = match state.tokens.get(fort_name) {
        Some(t) => t,
        None => return false,
    };

    let target = format!("{}/v1/auth/refresh", tokens.auth_url);
    let resp = state.client
        .post(&target)
        .json(&serde_json::json!({ "refresh_token": tokens.refresh_token }))
        .send()
        .await;

    match resp {
        Ok(r) if r.status().is_success() => {
            if let Ok(body) = r.json::<serde_json::Value>().await {
                let jwt = body.get("token").and_then(|v| v.as_str());
                let rt = body.get("refresh_token").and_then(|v| v.as_str());
                let exp = body.get("expires_in").and_then(|v| v.as_u64());

                if let (Some(jwt), Some(rt)) = (jwt, rt) {
                    let expiry = Instant::now() + Duration::from_secs(exp.unwrap_or(900));
                    state.tokens.set(fort_name, FortTokens {
                        jwt: jwt.to_string(),
                        refresh_token: rt.to_string(),
                        expiry,
                        auth_url: tokens.auth_url.clone(),
                    });
                    return true;
                }
            }
            false
        }
        _ => {
            // Refresh failed — remove this fort's tokens, force re-login
            state.tokens.remove(fort_name);
            false
        }
    }
}

/// Proxy with automatic 401 retry: if the first request returns 401,
/// attempt a per-fort token refresh and retry once.
pub async fn proxy_with_refresh(
    state: &AppState,
    method: &str,
    path: &str,
    query: Option<&str>,
    body: Option<Vec<u8>>,
    content_type: Option<&str>,
) -> Result<(Vec<u8>, u16, String), String> {
    let result = proxy_request(state, method, path, query, body.clone(), content_type).await?;

    if result.1 == 401 {
        if let Some(fort_name) = extract_fort_name(path) {
            if try_refresh(state, fort_name).await {
                // Retry with new JWT
                return proxy_request(state, method, path, query, body, content_type).await;
            }
        }
    }

    Ok(result)
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd /home/kazw/Work/WorkFort/scope/lead/src-tauri && cargo check 2>&1 | tail -5
```

Expected: compiles clean (auth.rs will have errors — we fix those in Chunk 2).

- [ ] **Step 4: Commit**

```bash
git add src-tauri/Cargo.toml src-tauri/src/proxy.rs
git commit -m "refactor(proxy): replace single token store with per-fort HashMap<String, FortTokens>"
```

---

## Chunk 2: Fort-Scoped Tauri Commands

### Task 2: Update auth commands to accept fort parameter

**Why:** Every auth command needs to know which fort it's operating on. The shell passes `fort` as the first argument.

**Files:**
- Edit: `src-tauri/src/auth.rs`

- [ ] **Step 1: Rewrite auth.rs with fort-scoped commands**

```rust
// src-tauri/src/auth.rs — full file replacement

use serde::{Deserialize, Serialize};
use std::time::{Duration, Instant};
use tauri::State;

use crate::proxy::{AppState, FortTokens};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct UserInfo {
    pub id: String,
    pub email: String,
    pub name: String,
}

#[derive(Debug, Deserialize)]
struct AuthResponse {
    token: String,
    refresh_token: String,
    #[serde(default)]
    expires_in: Option<u64>,
    user: UserInfo,
}

/// Tauri command: login to a specific fort with email/password.
/// Posts to that fort's auth service, stores JWT + refresh token under the fort name.
#[tauri::command]
pub async fn login(
    state: State<'_, AppState>,
    fort: String,
    auth_url: String,
    email: String,
    password: String,
) -> Result<UserInfo, String> {
    let target = format!("{}/v1/auth/login", auth_url);

    let resp = state.client
        .post(&target)
        .json(&serde_json::json!({
            "email": email,
            "password": password,
        }))
        .send()
        .await
        .map_err(|e| format!("Login request failed: {e}"))?;

    if !resp.status().is_success() {
        let status = resp.status().as_u16();
        let body = resp.text().await.unwrap_or_default();
        return Err(format!("Login failed ({status}): {body}"));
    }

    let auth: AuthResponse = resp.json().await
        .map_err(|e| format!("Invalid auth response: {e}"))?;

    let expiry = Instant::now() + Duration::from_secs(auth.expires_in.unwrap_or(900));
    state.tokens.set(&fort, FortTokens {
        jwt: auth.token,
        refresh_token: auth.refresh_token,
        expiry,
        auth_url,
    });

    Ok(auth.user)
}

/// Tauri command: logout from a specific fort. Removes that fort's tokens.
#[tauri::command]
pub async fn logout(
    state: State<'_, AppState>,
    fort: String,
) -> Result<(), String> {
    state.tokens.remove(&fort);
    Ok(())
}

/// Tauri command: get current user info for a specific fort.
/// Calls that fort's auth service using the stored JWT.
#[tauri::command]
pub async fn get_user(
    state: State<'_, AppState>,
    fort: String,
) -> Result<Option<UserInfo>, String> {
    let tokens = match state.tokens.get(&fort) {
        Some(t) => t,
        None => return Ok(None),
    };

    let target = format!("{}/v1/auth/me", tokens.auth_url);

    let resp = state.client
        .get(&target)
        .header("Authorization", format!("Bearer {}", tokens.jwt))
        .send()
        .await
        .map_err(|e| format!("Get user failed: {e}"))?;

    if resp.status().as_u16() == 401 {
        // Token expired, try refresh
        if crate::proxy::try_refresh(&state, &fort).await {
            // Retry with new JWT
            let new_tokens = state.tokens.get(&fort).unwrap();
            let target = format!("{}/v1/auth/me", new_tokens.auth_url);
            let resp = state.client
                .get(&target)
                .header("Authorization", format!("Bearer {}", new_tokens.jwt))
                .send()
                .await
                .map_err(|e| format!("Get user retry failed: {e}"))?;

            if resp.status().is_success() {
                let user: UserInfo = resp.json().await
                    .map_err(|e| format!("Invalid user response: {e}"))?;
                return Ok(Some(user));
            }
        }
        // Refresh failed or retry failed — user is not authenticated for this fort
        state.tokens.remove(&fort);
        return Ok(None);
    }

    if !resp.status().is_success() {
        return Err(format!("Get user failed: {}", resp.status()));
    }

    let user: UserInfo = resp.json().await
        .map_err(|e| format!("Invalid user response: {e}"))?;
    Ok(Some(user))
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/kazw/Work/WorkFort/scope/lead/src-tauri && cargo check 2>&1 | tail -5
```

Expected: compiles clean. The command registration in `lib.rs` doesn't change — same function names.

- [ ] **Step 3: Commit**

```bash
git add src-tauri/src/auth.rs
git commit -m "feat(auth): add fort parameter to login, logout, get_user commands"
```

---

## Chunk 3: Background Token Refresh

### Task 3: Add background refresh task in lib.rs

**Why:** Tokens expire. A background task checks every 60 seconds and refreshes any token within 5 minutes of expiry. This keeps all forts alive without user interaction.

**Files:**
- Edit: `src-tauri/src/lib.rs`

- [ ] **Step 1: Add background refresh task to lib.rs setup**

Add a `setup` hook that spawns the background task. The existing `handle_request` and command registration stay the same.

```rust
// src-tauri/src/lib.rs — full file replacement

mod auth;
mod proxy;

use proxy::{AppState, should_proxy, proxy_with_refresh};
use tauri::UriSchemeResponder;
use tauri::http::{Request, Response};
use std::time::{Duration, Instant};

/// API base URL — read from WORKFORT_API_URL env var, defaulting to localhost.
fn api_base_url() -> String {
    std::env::var("WORKFORT_API_URL")
        .unwrap_or_else(|_| "http://localhost:16100".to_string())
}

/// How often the background task checks token expiry.
const REFRESH_INTERVAL: Duration = Duration::from_secs(60);

/// Refresh tokens this long before they expire.
const REFRESH_BUFFER: Duration = Duration::from_secs(5 * 60);

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let state = AppState::new(&api_base_url());
    let proxy_state = state.clone();
    let refresh_state = state.clone();

    tauri::Builder::default()
        .manage(state)
        .invoke_handler(tauri::generate_handler![
            auth::login,
            auth::logout,
            auth::get_user,
        ])
        .register_asynchronous_uri_scheme_protocol("https", move |_ctx, request, responder| {
            let state = proxy_state.clone();
            tauri::async_runtime::spawn(async move {
                handle_request(state, request, responder).await;
            });
        })
        .setup(move |_app| {
            // Spawn background token refresh task
            let state = refresh_state.clone();
            tauri::async_runtime::spawn(async move {
                background_refresh(state).await;
            });
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

/// Background task: checks all fort tokens every REFRESH_INTERVAL.
/// Refreshes any token within REFRESH_BUFFER of expiry.
async fn background_refresh(state: AppState) {
    loop {
        tokio::time::sleep(REFRESH_INTERVAL).await;

        let fort_tokens = state.tokens.all();
        for (fort_name, tokens) in fort_tokens {
            let remaining = tokens.expiry.saturating_duration_since(Instant::now());
            if remaining < REFRESH_BUFFER {
                log::info!("Refreshing token for fort '{}' (expires in {:?})", fort_name, remaining);
                if proxy::try_refresh(&state, &fort_name).await {
                    log::info!("Token refreshed for fort '{}'", fort_name);
                } else {
                    log::warn!("Token refresh failed for fort '{}'", fort_name);
                    // If token is already expired, remove it
                    if remaining.is_zero() {
                        log::warn!("Token expired for fort '{}', removing", fort_name);
                        state.tokens.remove(&fort_name);
                    }
                    // Otherwise, retry on next cycle
                }
            }
        }
    }
}

async fn handle_request(
    state: AppState,
    request: Request<Vec<u8>>,
    responder: UriSchemeResponder,
) {
    let uri = request.uri().clone();
    let path = uri.path();

    if !should_proxy(path) {
        // Let Tauri handle non-API requests (serve frontend assets).
        // Return a 404 so the default handler takes over.
        let resp = Response::builder()
            .status(404)
            .body(Vec::new())
            .unwrap();
        responder.respond(resp);
        return;
    }

    let method = request.method().as_str();
    let query = uri.query();
    let content_type = request.headers()
        .get("content-type")
        .and_then(|v| v.to_str().ok());
    let body = if request.body().is_empty() {
        None
    } else {
        Some(request.body().clone())
    };

    match proxy_with_refresh(
        &state,
        method,
        path,
        query,
        body,
        content_type,
    ).await {
        Ok((bytes, status, ct)) => {
            let resp = Response::builder()
                .status(status)
                .header("Content-Type", ct)
                .header("Access-Control-Allow-Origin", "*")
                .body(bytes)
                .unwrap();
            responder.respond(resp);
        }
        Err(e) => {
            let resp = Response::builder()
                .status(502)
                .header("Content-Type", "application/json")
                .body(format!(r#"{{"error":"{}"}}"#, e).into_bytes())
                .unwrap();
            responder.respond(resp);
        }
    }
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/kazw/Work/WorkFort/scope/lead/src-tauri && cargo check 2>&1 | tail -5
```

Expected: compiles clean.

- [ ] **Step 3: Commit**

```bash
git add src-tauri/src/lib.rs
git commit -m "feat(proxy): add background token refresh task (60s interval, 5min buffer)"
```

---

## Chunk 4: Go BFF Update

### Task 4: Change refreshBefore from 1 minute to 5 minutes

**Why:** Standardize refresh buffer across both platforms. The Go BFF currently uses 1 minute which is too tight.

**Files:**
- Edit: `internal/infra/httpapi/bff.go`

- [ ] **Step 1: Update the constant**

In `internal/infra/httpapi/bff.go`, change line 15:

```go
// Before:
refreshBefore     = 1 * time.Minute

// After:
refreshBefore     = 5 * time.Minute
```

- [ ] **Step 2: Run Go tests**

```bash
cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/httpapi/... -v -run TestTokenConverter 2>&1
```

Expected: all `TestTokenConverter_*` tests pass. The `CacheHit` test uses the default `NewTokenConverter` with 15-minute lifetime, so a 5-minute refresh buffer still keeps the cache valid. The `ExpiredSession` test uses `NewTokenConverterForTest` with custom timing so it's unaffected.

- [ ] **Step 3: Run full Go test suite**

```bash
cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/... 2>&1 | tail -10
```

Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/infra/httpapi/bff.go
git commit -m "fix(bff): increase refreshBefore from 1min to 5min to match Tauri proxy"
```

---

## Chunk 5: Verification

### Task 5: Full compilation and test check

**Why:** Confirm everything compiles and all tests pass before declaring done.

- [ ] **Step 1: Rust compilation check**

```bash
cd /home/kazw/Work/WorkFort/scope/lead/src-tauri && cargo check 2>&1 | tail -5
```

Expected: `Finished` with no errors.

- [ ] **Step 2: Go test suite**

```bash
cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/... 2>&1 | tail -10
```

Expected: all `ok`.

- [ ] **Step 3: Verify extract_fort_name logic manually**

Quick sanity check — these are the key path patterns:

| Path | `extract_fort_name` result |
|------|---------------------------|
| `/api/forts` | `None` (no auth) |
| `/forts/acme/api/services` | `Some("acme")` |
| `/forts/dev-local/api/chat/v1/channels` | `Some("dev-local")` |
| `/forts/acme/ws/events` | `Some("acme")` |

---

## Summary of changes

| File | Change |
|------|--------|
| `src-tauri/Cargo.toml` | Add `tokio` and `log` dependencies |
| `src-tauri/src/proxy.rs` | `FortTokens` struct, `TokenStore` with `HashMap<String, FortTokens>`, `extract_fort_name()`, per-fort `proxy_request()` and `try_refresh()` |
| `src-tauri/src/auth.rs` | `fort` + `auth_url` params on `login()`, `fort` param on `logout()` and `get_user()` |
| `src-tauri/src/lib.rs` | `background_refresh()` task via `tokio::spawn` in setup hook |
| `internal/infra/httpapi/bff.go` | `refreshBefore` from `1 * time.Minute` to `5 * time.Minute` |

**DO NOT PUSH** — all changes are local commits only.
