mod auth;
mod proxy;

use proxy::{AppState, should_proxy, proxy_with_refresh};
use tauri::UriSchemeResponder;
use tauri::http::{Request, Response};

/// API base URL — read from WORKFORT_API_URL env var, defaulting to localhost.
fn api_base_url() -> String {
    std::env::var("WORKFORT_API_URL")
        .unwrap_or_else(|_| "http://localhost:16100".to_string())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let state = AppState::new(&api_base_url());
    let proxy_state = state.clone();

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
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
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
