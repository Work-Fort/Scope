use axum::{
    body::Body,
    extract::{ws::WebSocketUpgrade, Path, Request, State},
    http::StatusCode,
    response::{IntoResponse, Response},
};
use std::sync::Arc;

use crate::state::AppState;

/// ANY /forts/{fort}/api/{*rest}
pub async fn proxy_handler(
    State(state): State<Arc<AppState>>,
    Path((fort, rest)): Path<(String, String)>,
    req: Request,
) -> impl IntoResponse {
    // Extract service name from rest path (first segment)
    let parts: Vec<&str> = rest.splitn(2, '/').collect();
    let service_name = parts[0];
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

    // Get token for this fort
    let token = {
        let tokens = state.tokens.lock().await;
        tokens.get(&fort).map(|t| t.jwt.clone())
    };

    // Extract request details
    let method = req.method().to_string();
    let query = req.uri().query().map(|q| q.to_string());
    let headers: Vec<(String, String)> = req
        .headers()
        .iter()
        .map(|(k, v)| (k.to_string(), v.to_str().unwrap_or("").to_string()))
        .collect();
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

/// ANY /forts/{fort}/ws/{*rest}
pub async fn ws_proxy_handler(
    State(state): State<Arc<AppState>>,
    Path((fort, rest)): Path<(String, String)>,
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

    // Get token for this fort
    let token = {
        let tokens = state.tokens.lock().await;
        tokens.get(&fort).map(|t| t.jwt.clone())
    };

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
