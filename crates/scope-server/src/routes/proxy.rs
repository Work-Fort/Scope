use axum::{
    body::Body,
    extract::{ws::WebSocketUpgrade, FromRequest, Path, Request, State},
    http::{HeaderMap, StatusCode},
    response::{IntoResponse, Response},
};
use std::sync::Arc;
use std::time::{Duration, Instant};

use crate::state::AppState;

/// Exchange a session cookie for a JWT via Passport's /v1/token endpoint,
/// without checking the cache. Used by fort_session to bootstrap Pylon auth.
pub async fn exchange_cookie_for_token(state: &AppState, fort: &str, headers: &HeaderMap) -> Option<String> {
    let cookie = headers.get("cookie")?.to_str().ok()?;

    let services = state.discovery.services().await;
    let auth_url = match services.iter().find(|s| s.name == "auth") {
        Some(svc) => svc.base_url.clone(),
        None => state.passport_urls.lock().await.get(fort).cloned()?,
    };

    let client = reqwest::Client::new();
    let resp = client
        .get(format!("{auth_url}/v1/token"))
        .header("cookie", cookie)
        .send()
        .await
        .ok()?;

    if !resp.status().is_success() {
        log::warn!("token exchange failed: {}", resp.status());
        return None;
    }

    let body: serde_json::Value = resp.json().await.ok()?;
    let jwt = body.get("token").and_then(|t| t.as_str())?.to_string();

    {
        let mut tokens = state.tokens.lock().await;
        tokens.insert(
            fort.to_string(),
            scope_core::domain::session::FortTokens {
                jwt: jwt.clone(),
                refresh_token: String::new(),
                expiry: std::time::Instant::now() + std::time::Duration::from_secs(15 * 60),
                auth_url: auth_url.clone(),
            },
        );
    }

    Some(jwt)
}

/// Exchange a session cookie for a JWT via Passport's /v1/token endpoint.
/// Caches the JWT in the tokens map for subsequent requests.
async fn get_or_refresh_token(state: &AppState, fort: &str, headers: &HeaderMap) -> Option<String> {
    // Check cache first
    {
        let tokens = state.tokens.lock().await;
        if let Some(t) = tokens.get(fort) {
            // If token hasn't expired (with 1 minute buffer), use it
            if t.expiry > Instant::now() + Duration::from_secs(60) {
                return Some(t.jwt.clone());
            }
        }
    }

    // No cached token or expired — exchange session cookie for JWT
    let cookie = headers.get("cookie")?.to_str().ok()?;

    // Find auth service URL
    let services = state.discovery.services().await;
    let auth_svc = services.iter().find(|s| s.name == "auth")?;
    let auth_url = &auth_svc.base_url;

    // Call Passport's /v1/token (Better Auth JWT plugin)
    let client = reqwest::Client::new();
    let resp = client
        .get(format!("{auth_url}/v1/token"))
        .header("cookie", cookie)
        .send()
        .await
        .ok()?;

    if !resp.status().is_success() {
        log::warn!("token exchange failed: {}", resp.status());
        return None;
    }

    let body: serde_json::Value = resp.json().await.ok()?;
    let jwt = body.get("token").and_then(|t| t.as_str())?.to_string();

    // Cache with 15 minute expiry (matching Passport's JWT lifetime)
    {
        let mut tokens = state.tokens.lock().await;
        tokens.insert(
            fort.to_string(),
            scope_core::domain::session::FortTokens {
                jwt: jwt.clone(),
                refresh_token: String::new(),
                expiry: Instant::now() + Duration::from_secs(15 * 60),
                auth_url: auth_url.clone(),
            },
        );
    }

    Some(jwt)
}

/// ANY /forts/{fort}/api/{*rest}
/// Handles both HTTP requests and WebSocket upgrades.
pub async fn proxy_handler(
    State(state): State<Arc<AppState>>,
    Path((fort, rest)): Path<(String, String)>,
    req: Request,
) -> Response {
    // Check for WebSocket upgrade
    let is_ws_upgrade = req
        .headers()
        .get("upgrade")
        .and_then(|v| v.to_str().ok())
        .map(|v| v.eq_ignore_ascii_case("websocket"))
        .unwrap_or(false);
    // Extract service name from rest path (first segment)
    let parts: Vec<&str> = rest.splitn(2, '/').collect();
    let service_name = parts[0];
    let service_path = if parts.len() > 1 {
        format!("/{}", parts[1])
    } else {
        "/".into()
    };

    // Find service URL from discovery, falling back to passport_url for auth
    let services = state.discovery.services().await;
    let service = services.iter().find(|s| s.name == service_name);
    let base_url = match service {
        Some(s) => s.base_url.clone(),
        None if service_name == "auth" => {
            match state.passport_urls.lock().await.get(&fort).cloned() {
                Some(url) => url,
                None => return (StatusCode::BAD_GATEWAY, "auth service not available").into_response(),
            }
        }
        None => {
            return (
                StatusCode::BAD_GATEWAY,
                format!("service not found: {service_name}"),
            )
                .into_response()
        }
    };

    // Get or exchange token for this fort
    let token = get_or_refresh_token(&state, &fort, req.headers()).await;

    // If this is a WebSocket upgrade, handle it as WS proxy
    if is_ws_upgrade {
        let ws_base = base_url
            .replacen("https://", "wss://", 1)
            .replacen("http://", "ws://", 1);
        let ws_url = format!("{ws_base}{service_path}");
        let svc_name = service_name.to_string();

        log::info!("WS proxy: {svc_name}{service_path} → {ws_url}");

        // Extract WebSocketUpgrade from the request
        match WebSocketUpgrade::from_request(req, &state).await {
            Ok(ws) => {
                return ws
                    .on_upgrade(move |socket| async move {
                        if let Err(e) = handle_ws_pipe(socket, &ws_url, token.as_deref()).await {
                            log::error!("WS proxy error for {svc_name}: {e}");
                        }
                    })
                    .into_response();
            }
            Err(e) => {
                return (StatusCode::BAD_REQUEST, format!("WS upgrade failed: {e}")).into_response();
            }
        }
    }

    // Extract request details
    let method = req.method().to_string();
    let query = req.uri().query().map(|q| q.to_string());
    let headers: Vec<(String, String)> = req
        .headers()
        .iter()
        .map(|(k, v)| (k.to_string(), v.to_str().unwrap_or("").to_string()))
        .collect();
    // Debug: log cookie header for proxy requests
    for (k, v) in &headers {
        if k == "cookie" {
            log::debug!("proxy {service_name}{service_path}: cookie={}", &v[..v.len().min(60)]);
        }
    }

    let body = axum::body::to_bytes(req.into_body(), 10 * 1024 * 1024) // 10MB limit
        .await
        .ok()
        .map(|b| b.to_vec());

    // Forward
    match state
        .proxy
        .forward_http(
            &base_url,
            &method,
            &service_path,
            query.as_deref(),
            &headers,
            body,
            token.as_deref(),
        )
        .await
    {
        Ok(resp) => {
            let mut response = Response::builder().status(resp.status);
            for (key, value) in &resp.headers {
                response = response.header(key.as_str(), value.as_str());
            }
            response
                .body(Body::from(resp.body))
                .unwrap()
                .into_response()
        }
        Err(e) => (StatusCode::BAD_GATEWAY, e.to_string()).into_response(),
    }
}

/// GET /forts/{fort}/api/{service}/ws
pub async fn ws_proxy_handler_ws(
    state: State<Arc<AppState>>,
    path: Path<(String, String)>,
    headers: HeaderMap,
    ws: WebSocketUpgrade,
) -> impl IntoResponse {
    ws_proxy_handler_for(state, path, headers, ws, "/ws").await
}

/// GET /forts/{fort}/api/{service}/presence
pub async fn ws_proxy_handler_presence(
    state: State<Arc<AppState>>,
    path: Path<(String, String)>,
    headers: HeaderMap,
    ws: WebSocketUpgrade,
) -> impl IntoResponse {
    ws_proxy_handler_for(state, path, headers, ws, "/presence").await
}

/// Shared WS proxy logic for service-specific endpoints.
async fn ws_proxy_handler_for(
    State(state): State<Arc<AppState>>,
    Path((fort, service)): Path<(String, String)>,
    headers: HeaderMap,
    ws: WebSocketUpgrade,
    ws_path: &'static str,
) -> impl IntoResponse {

    // Find service URL from discovery
    let services = state.discovery.services().await;
    let svc = services.iter().find(|s| s.name == service);
    let base_url = match svc {
        Some(s) => s.base_url.clone(),
        None => {
            return (
                StatusCode::BAD_GATEWAY,
                format!("service not found: {service}"),
            )
                .into_response()
        }
    };

    // Get or exchange token
    let token = get_or_refresh_token(&state, &fort, &headers).await;

    let ws_base = base_url
        .replacen("https://", "wss://", 1)
        .replacen("http://", "ws://", 1);
    let ws_url = format!("{ws_base}{ws_path}");

    log::info!("WS proxy: {service}{ws_path} → {ws_url}");

    ws.on_upgrade(move |socket| async move {
        if let Err(e) = handle_ws_pipe(socket, &ws_url, token.as_deref()).await {
            log::error!("WS proxy error for {service}: {e}");
        }
    })
}

/// ANY /forts/{fort}/ws/{*rest}
pub async fn ws_proxy_handler(
    State(state): State<Arc<AppState>>,
    Path((fort, rest)): Path<(String, String)>,
    headers: HeaderMap,
    ws: WebSocketUpgrade,
) -> impl IntoResponse {
    // Extract service name from rest path (first segment)
    let parts: Vec<&str> = rest.splitn(2, '/').collect();
    let service_name = parts[0].to_string();
    let service_path = if parts.len() > 1 {
        format!("/{}", parts[1])
    } else {
        "/".into()
    };

    // Find service URL from discovery
    let services = state.discovery.services().await;
    let service = services.iter().find(|s| s.name == service_name);
    let base_url = match service {
        Some(s) => s.base_url.clone(),
        None => {
            return (
                StatusCode::BAD_GATEWAY,
                format!("service not found: {service_name}"),
            )
                .into_response()
        }
    };

    // Get or exchange token for this fort
    let token = get_or_refresh_token(&state, &fort, &headers).await;

    // Convert http:// base URL to ws://
    let ws_base = base_url
        .replacen("https://", "wss://", 1)
        .replacen("http://", "ws://", 1);
    let ws_url = format!("{ws_base}{service_path}");

    ws.on_upgrade(move |socket| async move {
        if let Err(e) = handle_ws_pipe(socket, &ws_url, token.as_deref()).await {
            log::error!("WS proxy error: {e}");
        }
    })
}

/// Bridge an axum WebSocket to a backend tokio-tungstenite WebSocket.
///
/// Axum and tokio-tungstenite have different Message types, so we convert
/// between them in each direction.
async fn handle_ws_pipe(
    socket: axum::extract::ws::WebSocket,
    ws_url: &str,
    token: Option<&str>,
) -> Result<(), scope_core::Error> {
    use axum::extract::ws::Message as AxumMsg;
    use futures_util::{SinkExt, StreamExt};
    use tokio_tungstenite::tungstenite::{
        client::IntoClientRequest, Message as TungsteniteMsg,
    };

    // Build the backend WS request with optional auth
    let mut request = ws_url
        .into_client_request()
        .map_err(|e| scope_core::Error::Other(format!("invalid WS URL: {e}")))?;

    if let Some(tok) = token {
        request.headers_mut().insert(
            "Authorization",
            format!("Bearer {tok}")
                .parse()
                .map_err(|e| scope_core::Error::Other(format!("invalid auth header: {e}")))?,
        );
    }

    let (backend_ws, _resp) = tokio_tungstenite::connect_async(request)
        .await
        .map_err(|e| scope_core::Error::Other(format!("backend WS connect failed: {e}")))?;

    let (mut backend_tx, mut backend_rx) = backend_ws.split();
    let (mut client_tx, mut client_rx) = socket.split();

    // client → backend
    let c2b = async {
        while let Some(Ok(msg)) = client_rx.next().await {
            let tung_msg = match msg {
                AxumMsg::Text(t) => TungsteniteMsg::Text(t.to_string().into()),
                AxumMsg::Binary(b) => TungsteniteMsg::Binary(b.to_vec().into()),
                AxumMsg::Ping(p) => TungsteniteMsg::Ping(p.to_vec().into()),
                AxumMsg::Pong(p) => TungsteniteMsg::Pong(p.to_vec().into()),
                AxumMsg::Close(_) => {
                    let _ = backend_tx.close().await;
                    break;
                }
            };
            if backend_tx.send(tung_msg).await.is_err() {
                break;
            }
        }
    };

    // backend → client
    let b2c = async {
        while let Some(Ok(msg)) = backend_rx.next().await {
            let axum_msg = match msg {
                TungsteniteMsg::Text(t) => AxumMsg::Text(t.to_string().into()),
                TungsteniteMsg::Binary(b) => AxumMsg::Binary(b.to_vec().into()),
                TungsteniteMsg::Ping(p) => AxumMsg::Ping(p.to_vec().into()),
                TungsteniteMsg::Pong(p) => AxumMsg::Pong(p.to_vec().into()),
                TungsteniteMsg::Close(_) => {
                    let _ = client_tx.close().await;
                    break;
                }
                TungsteniteMsg::Frame(_) => continue,
            };
            if client_tx.send(axum_msg).await.is_err() {
                break;
            }
        }
    };

    tokio::select! {
        _ = c2b => {},
        _ = b2c => {},
    }

    Ok(())
}
