# Pylon Backend Integration — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable Scope's backend to fetch service listings from a Pylon server instead of probing individual services directly, supporting both local (direct probe) and Pylon (remote fetch) modes.

**Architecture:** `ServiceDiscovery` in scope-core gains a second code path: when `!fort.local && fort.pylon.is_some()`, it fetches `GET {pylon_url}/api/services` with a Bearer token (obtained via Pylon's auth dance) instead of probing each service's `/ui/health`. Local probes also get updated to handle HTTP 267 (no UI). scope-server is enforced as single-fort with appropriate polling intervals. The shell WebSocket broadcasts service changes when discovery detects differences.

**Tech Stack:** Rust (axum, reqwest, tokio, serde), SQLx, tokio-tungstenite

**References:**
- `~/Work/WorkFort/pylon/lead/docs/http-status-codes.md` — the status code contract
- `docs/pylon-integration-design.md` — the design doc

---

### Task 1: Handle HTTP 267 in local discovery

Local probes currently treat any `status.is_success()` as `ui: true`. The new contract uses 200 = UI available, 267 = no UI.

**Files:**
- Modify: `crates/scope-core/src/infra/discovery/mod.rs:68-83`
- Test: `crates/scope-core/src/infra/discovery/mod.rs` (tests module at bottom)

**Step 1: Write a failing test for 267 handling**

Add to the `tests` module in `crates/scope-core/src/infra/discovery/mod.rs`:

```rust
#[tokio::test]
async fn discovery_267_sets_ui_false() {
    let manifest = serde_json::json!({
        "name": "nexus",
        "label": "VMs",
        "route": "/vms",
    });

    let app = Router::new().route(
        "/ui/health",
        get(move || {
            let m = manifest.clone();
            async move {
                (axum::http::StatusCode::from_u16(267).unwrap(), Json(m))
            }
        }),
    );

    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let addr = listener.local_addr().unwrap();
    tokio::spawn(async move {
        axum::serve(listener, app).await.unwrap();
    });

    let fort = Fort {
        name: "test".into(),
        local: true,
        pylon: None,
        services: vec![ServiceConfig {
            url: format!("http://{}", addr),
        }],
    };

    let discovery = ServiceDiscovery::new();
    discovery.probe_all(&fort).await;

    let services = discovery.services().await;
    assert_eq!(services.len(), 1);
    assert_eq!(services[0].name, "nexus");
    assert!(services[0].connected);
    assert!(!services[0].ui); // 267 = no UI
}
```

**Step 2: Run test to verify it fails**

Run: `cargo test -p scope-core discovery_267 -- --nocapture`
Expected: FAIL — `services[0].ui` is `true` because `is_success()` returns true for 2xx

**Step 3: Update `probe_one` to distinguish 200 from 267**

In `crates/scope-core/src/infra/discovery/mod.rs`, change `probe_one`:

```rust
let status = resp.status();
let manifest: HealthManifest = resp.json().await.ok()?;

// 200 = UI available, 267 = connected but no UI, other 2xx = treat as UI available
let ui = status.as_u16() != 267 && status.is_success();

Some(TrackedService {
    name: manifest.name,
    label: manifest.label,
    route: manifest.route,
    base_url: base.to_string(),
    ui,
    connected: true,
    // ... rest unchanged
```

**Step 4: Run tests to verify they pass**

Run: `cargo test -p scope-core -- --nocapture`
Expected: All 18 tests PASS (17 existing + 1 new)

**Step 5: Commit**

```bash
git add crates/scope-core/src/infra/discovery/mod.rs
git commit -m "feat(discovery): handle HTTP 267 as connected-but-no-ui"
```

---

### Task 2: Add Pylon fetch path to discovery

When a fort has `pylon` set and `local: false`, discovery fetches the service list from Pylon's `GET /api/services` instead of probing each service individually.

**Files:**
- Modify: `crates/scope-core/src/infra/discovery/mod.rs`

**Step 1: Write a failing test for Pylon fetch**

```rust
#[tokio::test]
async fn discovery_fetches_from_pylon() {
    // Mock Pylon server that returns a service list when given a Bearer token
    let app = Router::new().route(
        "/api/services",
        get(|headers: axum::http::HeaderMap| async move {
            let auth = headers.get("authorization").and_then(|v| v.to_str().ok());
            match auth {
                Some(h) if h.starts_with("Bearer ") => {
                    Json(serde_json::json!({
                        "services": [{
                            "name": "sharkfin",
                            "label": "Chat",
                            "route": "/chat",
                            "base_url": "http://10.0.0.1:16000",
                            "ui": true,
                            "connected": true,
                            "display": "nav",
                            "ws_paths": ["/ws"],
                        }]
                    }))
                    .into_response()
                }
                _ => {
                    Json(serde_json::json!({
                        "passport_url": "https://passport.example.com"
                    }))
                    .into_response()
                }
            }
        }),
    );

    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let addr = listener.local_addr().unwrap();
    tokio::spawn(async move {
        axum::serve(listener, app).await.unwrap();
    });

    let fort = Fort {
        name: "acme".into(),
        local: false,
        pylon: Some(format!("http://{}", addr)),
        services: vec![], // no local services — Pylon provides them
    };

    let discovery = ServiceDiscovery::new();
    // For now, test with a pre-set token
    discovery.fetch_from_pylon(&fort, Some("test-jwt")).await;

    let services = discovery.services().await;
    assert_eq!(services.len(), 1);
    assert_eq!(services[0].name, "sharkfin");
    assert_eq!(services[0].base_url, "http://10.0.0.1:16000");
    assert!(services[0].ui);
}
```

**Step 2: Run test to verify it fails**

Run: `cargo test -p scope-core discovery_fetches_from_pylon -- --nocapture`
Expected: FAIL — `fetch_from_pylon` method doesn't exist

**Step 3: Implement `fetch_from_pylon`**

Add to `ServiceDiscovery` in `crates/scope-core/src/infra/discovery/mod.rs`:

```rust
/// Fetch service list from a Pylon server. Used for non-local forts.
/// Returns the passport_url if authentication is required (no token provided).
pub async fn fetch_from_pylon(
    &self,
    fort: &Fort,
    token: Option<&str>,
) -> Option<String> {
    let pylon_url = fort.pylon.as_deref()?;
    let base = pylon_url.trim_end_matches('/');
    let url = format!("{base}/api/services");

    let mut req = self.client.get(&url);
    if let Some(tok) = token {
        req = req.header("Authorization", format!("Bearer {tok}"));
    }

    let resp = match req.send().await {
        Ok(r) => r,
        Err(e) => {
            log::warn!("pylon fetch failed for {}: {e}", fort.name);
            return None;
        }
    };

    let status = resp.status();

    if status.as_u16() == 401 {
        log::info!("pylon token expired for {}", fort.name);
        return Some("__expired__".into());
    }

    let body: serde_json::Value = match resp.json().await {
        Ok(v) => v,
        Err(e) => {
            log::warn!("pylon response parse failed for {}: {e}", fort.name);
            return None;
        }
    };

    // If response has passport_url, authentication is needed
    if let Some(passport_url) = body.get("passport_url").and_then(|v| v.as_str()) {
        return Some(passport_url.to_string());
    }

    // Parse the services array
    if let Some(services_val) = body.get("services") {
        let services: Vec<TrackedService> =
            serde_json::from_value(services_val.clone()).unwrap_or_default();
        *self.services.write().await = services;
    }

    None // No auth needed, services updated
}
```

Also update `probe_all` to route to the right code path:

```rust
pub async fn probe_all(&self, fort: &Fort) {
    if !fort.local && fort.pylon.is_some() {
        // Pylon forts are handled by fetch_from_pylon, called externally
        // with token management. probe_all is for local forts only.
        return;
    }

    let mut discovered = Vec::new();
    for svc_config in &fort.services {
        if let Some(tracked) = self.probe_one(&svc_config.url).await {
            discovered.push(tracked);
        }
    }
    *self.services.write().await = discovered;
}
```

**Step 4: Run tests to verify they pass**

Run: `cargo test -p scope-core -- --nocapture`
Expected: All tests PASS

**Step 5: Commit**

```bash
git add crates/scope-core/src/infra/discovery/mod.rs
git commit -m "feat(discovery): add Pylon fetch path for non-local forts"
```

---

### Task 3: Add test for Pylon auth dance (passport_url response)

**Files:**
- Test: `crates/scope-core/src/infra/discovery/mod.rs`

**Step 1: Write test for passport_url detection**

```rust
#[tokio::test]
async fn discovery_pylon_returns_passport_url_when_no_token() {
    let app = Router::new().route(
        "/api/services",
        get(|| async {
            Json(serde_json::json!({
                "passport_url": "https://passport.example.com"
            }))
        }),
    );

    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let addr = listener.local_addr().unwrap();
    tokio::spawn(async move {
        axum::serve(listener, app).await.unwrap();
    });

    let fort = Fort {
        name: "acme".into(),
        local: false,
        pylon: Some(format!("http://{}", addr)),
        services: vec![],
    };

    let discovery = ServiceDiscovery::new();
    let result = discovery.fetch_from_pylon(&fort, None).await;

    assert_eq!(result.as_deref(), Some("https://passport.example.com"));
    assert!(discovery.services().await.is_empty());
}
```

**Step 2: Run test — should pass with the implementation from Task 2**

Run: `cargo test -p scope-core discovery_pylon_returns_passport_url -- --nocapture`
Expected: PASS

**Step 3: Commit**

```bash
git add crates/scope-core/src/infra/discovery/mod.rs
git commit -m "test(discovery): verify Pylon auth dance returns passport_url"
```

---

### Task 4: Service change detection + broadcast

Discovery should detect when the service list changes between polls and notify the shell WebSocket.

**Files:**
- Modify: `crates/scope-core/src/infra/discovery/mod.rs`
- Modify: `crates/scope-server/src/routes/shell_ws.rs`
- Modify: `crates/scope-server/src/state.rs`
- Modify: `crates/scope-server/src/main.rs`

**Step 1: Add a `services_changed` broadcast channel to AppState**

In `crates/scope-server/src/state.rs`:

```rust
pub struct AppState {
    pub store: Arc<dyn Store>,
    pub discovery: Arc<ServiceDiscovery>,
    pub notify_tx: broadcast::Sender<Notification>,
    pub services_tx: broadcast::Sender<Vec<TrackedService>>,
    pub proxy: ProxyHandler,
    pub tokens: Mutex<HashMap<String, FortTokens>>,
}
```

**Step 2: Add change detection to discovery**

Add a method to `ServiceDiscovery` that compares the current snapshot with a previous one:

```rust
/// Check if services have changed since the given snapshot.
pub async fn has_changed_since(&self, prev: &[TrackedService]) -> bool {
    let current = self.services.read().await;
    if current.len() != prev.len() {
        return true;
    }
    for (a, b) in current.iter().zip(prev.iter()) {
        if a.name != b.name || a.connected != b.connected || a.ui != b.ui
            || a.base_url != b.base_url || a.setup_mode != b.setup_mode
        {
            return true;
        }
    }
    false
}
```

**Step 3: Update main.rs polling loop to detect changes and broadcast**

In `crates/scope-server/src/main.rs`, update the polling spawn:

```rust
let discovery = Arc::clone(&state.discovery);
let services_tx = state.services_tx.clone();
let fort_clone = fort.clone();
let interval = if fort.local {
    Duration::from_secs(10)
} else {
    Duration::from_secs(120)
};
tokio::spawn(async move {
    loop {
        let prev = discovery.services().await;
        discovery.probe_all(&fort_clone).await;
        if discovery.has_changed_since(&prev).await {
            let current = discovery.services().await;
            let _ = services_tx.send(current);
        }
        tokio::time::sleep(interval).await;
    }
});
```

**Step 4: Update shell_ws to listen for service change broadcasts**

In `crates/scope-server/src/routes/shell_ws.rs`, add a third `select!` branch:

```rust
let mut services_rx = state.services_tx.subscribe();

// In the select! loop:
Ok(services) = services_rx.recv() => {
    let event = serde_json::json!({
        "type": "services_changed",
        "data": services,
    });
    if socket.send(Message::Text(event.to_string().into())).await.is_err() {
        break;
    }
}
```

**Step 5: Create the second broadcast channel in main.rs**

Where `notify_tx` is created, add:

```rust
let (services_tx, _) = broadcast::channel(16);
```

And pass it to `AppState`.

**Step 6: Run tests and check compilation**

Run: `cargo check --workspace && cargo test -p scope-core -- --nocapture`
Expected: Compiles and all tests pass

**Step 7: Commit**

```bash
git add crates/scope-core/src/infra/discovery/mod.rs crates/scope-server/src/
git commit -m "feat: broadcast service changes over shell WebSocket"
```

---

### Task 5: scope-server single-fort enforcement + polling interval

scope-server should enforce exactly one fort and use appropriate polling intervals.

**Files:**
- Modify: `crates/scope-server/src/main.rs:102-115`

**Step 1: Enforce single fort on startup**

Replace the current `if let Some(fort) = forts.first()` with:

```rust
let forts = state.store.list_forts().await.unwrap_or_default();
if forts.len() > 1 {
    log::warn!(
        "scope-server supports one fort, but {} are configured. Using '{}'.",
        forts.len(),
        forts[0].name
    );
}
let fort = match forts.into_iter().next() {
    Some(f) => f,
    None => {
        log::error!("no forts configured");
        std::process::exit(1);
    }
};
```

**Step 2: Use fort.local to decide poll interval**

Already covered in Task 4's polling loop change. Verify it's in place:
- Local fort: 10 seconds
- Pylon fort: 120 seconds

**Step 3: Run check**

Run: `cargo check -p scope-server`
Expected: Compiles

**Step 4: Commit**

```bash
git add crates/scope-server/src/main.rs
git commit -m "feat(server): enforce single fort, adjust poll interval for Pylon"
```

---

### Task 6: Wire Pylon discovery into scope-server polling

The main.rs polling loop needs to call `fetch_from_pylon` for non-local forts, handling the auth dance.

**Files:**
- Modify: `crates/scope-server/src/main.rs`

**Step 1: Update polling loop for Pylon forts**

For Pylon forts, the loop must:
1. Try `fetch_from_pylon` with cached token
2. If it returns `passport_url`, the BFF can't authenticate on its own (it needs the user's session). Log a warning — the frontend will handle auth.
3. If it returns `__expired__`, clear the cached token

```rust
if fort.local {
    tokio::spawn(async move {
        loop {
            let prev = discovery.services().await;
            discovery.probe_all(&fort_clone).await;
            if discovery.has_changed_since(&prev).await {
                let _ = services_tx.send(discovery.services().await);
            }
            tokio::time::sleep(Duration::from_secs(10)).await;
        }
    });
} else {
    let tokens = Arc::clone(&tokens_for_polling);
    tokio::spawn(async move {
        loop {
            let prev = discovery.services().await;
            let token = {
                let t = tokens.lock().await;
                t.get(&fort_clone.name).map(|ft| ft.jwt.clone())
            };
            let result = discovery
                .fetch_from_pylon(&fort_clone, token.as_deref())
                .await;
            if let Some(url) = result {
                if url == "__expired__" {
                    tokens.lock().await.remove(&fort_clone.name);
                } else {
                    log::info!("pylon requires auth via {url} for fort '{}'", fort_clone.name);
                }
            }
            if discovery.has_changed_since(&prev).await {
                let _ = services_tx.send(discovery.services().await);
            }
            tokio::time::sleep(Duration::from_secs(120)).await;
        }
    });
}
```

**Step 2: Run check**

Run: `cargo check -p scope-server`
Expected: Compiles

**Step 3: Commit**

```bash
git add crates/scope-server/src/main.rs
git commit -m "feat(server): wire Pylon fetch into polling loop with auth handling"
```

---

### Task 7: Update config.yaml

Replace `gateway` with `pylon` in the live config file.

**Files:**
- Modify: `~/.config/workfort/config.yaml`

**Step 1: Update the config**

The current config:
```yaml
listen: "0.0.0.0:16100"

forts:
  local:
    local: true
    services:
      - url: "http://passport.nexus:3000"
      - url: "http://127.0.0.1:16000"
```

This fort is `local: true` with no gateway/pylon, so no field rename is needed — the config is already correct. If there were a `gateway:` key, it would need to be renamed to `pylon:`. **No change required for this specific file.**

**Step 2: Verify config loads correctly**

Run: `cargo run -p scope-server` (then Ctrl+C)
Expected: Starts up, logs `scope-server listening on 0.0.0.0:16100`

**Step 3: Commit (skip if no changes)**

No commit needed — config already compatible.

---

### Task 8: Full integration test

Run the complete test suite to verify nothing is broken.

**Step 1: Run all tests**

Run: `cargo test --workspace -- --nocapture`
Expected: All tests pass

**Step 2: Run lint**

Run: `cargo clippy --workspace`
Expected: No new warnings

**Step 3: Commit any fixes, if needed**
