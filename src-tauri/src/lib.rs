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
