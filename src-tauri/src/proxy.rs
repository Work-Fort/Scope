use reqwest::Client;
use std::sync::{Arc, Mutex};
use url::Url;

/// In-memory token store. Cleared when app is killed.
/// JWT and refresh token live here — never in the webview.
#[derive(Clone)]
pub struct TokenStore {
    pub jwt: Arc<Mutex<Option<String>>>,
    pub refresh_token: Arc<Mutex<Option<String>>>,
}

impl TokenStore {
    pub fn new() -> Self {
        Self {
            jwt: Arc::new(Mutex::new(None)),
            refresh_token: Arc::new(Mutex::new(None)),
        }
    }

    pub fn get_jwt(&self) -> Option<String> {
        self.jwt.lock().unwrap().clone()
    }

    pub fn set_jwt(&self, token: Option<String>) {
        *self.jwt.lock().unwrap() = token;
    }

    pub fn get_refresh(&self) -> Option<String> {
        self.refresh_token.lock().unwrap().clone()
    }

    pub fn set_refresh(&self, token: Option<String>) {
        *self.refresh_token.lock().unwrap() = token;
    }

    pub fn clear(&self) {
        self.set_jwt(None);
        self.set_refresh(None);
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
            .timeout(std::time::Duration::from_secs(10))
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

/// Proxies a request to the API backend, attaching the JWT if available.
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

    // Attach JWT
    if let Some(jwt) = state.tokens.get_jwt() {
        req = req.header("Authorization", format!("Bearer {jwt}"));
    }

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

/// Attempts to refresh the JWT using the stored refresh token.
/// Returns true if refresh succeeded, false otherwise.
pub async fn try_refresh(state: &AppState) -> bool {
    let refresh = match state.tokens.get_refresh() {
        Some(r) => r,
        None => return false,
    };

    let target = state.api_base.join("/api/auth/refresh").unwrap();
    let resp = state.client
        .post(target)
        .json(&serde_json::json!({ "refresh_token": refresh }))
        .send()
        .await;

    match resp {
        Ok(r) if r.status().is_success() => {
            if let Ok(body) = r.json::<serde_json::Value>().await {
                if let Some(jwt) = body.get("token").and_then(|v| v.as_str()) {
                    state.tokens.set_jwt(Some(jwt.to_string()));
                }
                if let Some(rt) = body.get("refresh_token").and_then(|v| v.as_str()) {
                    state.tokens.set_refresh(Some(rt.to_string()));
                }
                return true;
            }
            false
        }
        _ => {
            // Refresh failed — clear all tokens, force re-login
            state.tokens.clear();
            false
        }
    }
}

/// Proxy with automatic 401 retry: if the first request returns 401,
/// attempt a token refresh and retry once.
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
        if try_refresh(state).await {
            // Retry with new JWT
            return proxy_request(state, method, path, query, body, content_type).await;
        }
    }

    Ok(result)
}
