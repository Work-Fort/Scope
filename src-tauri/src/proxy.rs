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
