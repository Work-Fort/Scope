use scope_core::domain::session::FortTokens;
use scope_core::domain::Store;
use scope_core::infra::proxy::ProxyHandler;
use scope_core::infra::discovery::ServiceDiscovery;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use url::Url;

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
        let map = self.forts.lock().unwrap();
        let t = map.get(fort)?;
        // FortTokens doesn't derive Clone, so reconstruct
        Some(FortTokens {
            jwt: t.jwt.clone(),
            refresh_token: t.refresh_token.clone(),
            expiry: t.expiry,
            auth_url: t.auth_url.clone(),
        })
    }

    pub fn set(&self, fort: &str, tokens: FortTokens) {
        self.forts.lock().unwrap().insert(fort.to_string(), tokens);
    }

    pub fn remove(&self, fort: &str) {
        self.forts.lock().unwrap().remove(fort);
    }

    /// Returns a snapshot of all fort names and their tokens.
    pub fn all(&self) -> Vec<(String, FortTokens)> {
        self.forts
            .lock()
            .unwrap()
            .iter()
            .map(|(k, v)| {
                (
                    k.clone(),
                    FortTokens {
                        jwt: v.jwt.clone(),
                        refresh_token: v.refresh_token.clone(),
                        expiry: v.expiry,
                        auth_url: v.auth_url.clone(),
                    },
                )
            })
            .collect()
    }
}

/// State shared across the Tauri app.
/// Uses scope-core's ProxyHandler for HTTP forwarding and Store for persistence.
#[derive(Clone)]
pub struct AppState {
    pub client: reqwest::Client,
    pub proxy: Arc<ProxyHandler>,
    pub tokens: TokenStore,
    pub store: Arc<dyn Store>,
    pub discovery: Arc<ServiceDiscovery>,
    pub api_base: Url,
}

impl AppState {
    pub fn new(api_base_url: &str, store: Arc<dyn Store>) -> Self {
        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(10))
            .build()
            .expect("Failed to build HTTP client");

        Self {
            client,
            proxy: Arc::new(ProxyHandler::new()),
            tokens: TokenStore::new(),
            store,
            discovery: Arc::new(ServiceDiscovery::new()),
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

/// Proxies a request to the API backend using scope-core's ProxyHandler.
/// Attaches the per-fort JWT if available.
/// Returns the response body bytes, status code, and content-type.
pub async fn proxy_request(
    state: &AppState,
    method: &str,
    path: &str,
    query: Option<&str>,
    body: Option<Vec<u8>>,
    content_type: Option<&str>,
) -> Result<(Vec<u8>, u16, String), String> {
    let service_url = state.api_base.as_str().trim_end_matches('/');

    // Build headers to forward
    let mut headers = Vec::new();
    if let Some(ct) = content_type {
        headers.push(("Content-Type".to_string(), ct.to_string()));
    }

    // Get per-fort JWT if this is a fort-scoped request
    let token = extract_fort_name(path)
        .and_then(|fort_name| state.tokens.get(fort_name))
        .map(|t| t.jwt);

    let resp = state
        .proxy
        .forward_http(
            service_url,
            method,
            path,
            query,
            &headers,
            body,
            token.as_deref(),
        )
        .await
        .map_err(|e| format!("Proxy error: {e}"))?;

    let ct = resp
        .headers
        .iter()
        .find(|(k, _)| k.eq_ignore_ascii_case("content-type"))
        .map(|(_, v)| v.clone())
        .unwrap_or_else(|| "application/octet-stream".to_string());

    Ok((resp.body, resp.status, ct))
}

/// Attempts to refresh the JWT for a specific fort using its stored refresh token.
/// Returns true if refresh succeeded, false otherwise.
// TODO: unify with scope-core session management when available
pub async fn try_refresh(state: &AppState, fort_name: &str) -> bool {
    let tokens = match state.tokens.get(fort_name) {
        Some(t) => t,
        None => return false,
    };

    let target = format!("{}/v1/auth/refresh", tokens.auth_url);
    let resp = state
        .client
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
                    state.tokens.set(
                        fort_name,
                        FortTokens {
                            jwt: jwt.to_string(),
                            refresh_token: rt.to_string(),
                            expiry,
                            auth_url: tokens.auth_url.clone(),
                        },
                    );
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
