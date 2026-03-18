mod auth;
mod proxy;

use proxy::{AppState, should_proxy, proxy_with_refresh};
use scope_core::config;
use scope_core::domain::NotificationLevel;
use tauri::UriSchemeResponder;
use tauri::http::{Request, Response};
use tauri::State;
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

// ── Notification commands ───────────────────────────────────────────

#[tauri::command]
async fn get_notifications(
    limit: Option<i64>,
    before_id: Option<i64>,
    state: State<'_, AppState>,
) -> Result<serde_json::Value, String> {
    let notifications = state
        .store
        .list_notifications(limit.unwrap_or(20), before_id)
        .await
        .map_err(|e| e.to_string())?;
    let unread = state.store.unread_count().await.map_err(|e| e.to_string())?;
    Ok(serde_json::json!({ "notifications": notifications, "unread": unread }))
}

#[tauri::command]
async fn mark_notifications_read(
    up_to_id: i64,
    state: State<'_, AppState>,
) -> Result<(), String> {
    state
        .store
        .mark_read(up_to_id)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn get_preference(
    service: String,
    state: State<'_, AppState>,
) -> Result<serde_json::Value, String> {
    let level = state
        .store
        .get_preference(&service)
        .await
        .map_err(|e| e.to_string())?;
    Ok(serde_json::json!({ "service": service, "level": level }))
}

#[tauri::command]
async fn set_preference(
    service: String,
    level: String,
    state: State<'_, AppState>,
) -> Result<(), String> {
    let level = match level.as_str() {
        "mute" => NotificationLevel::Mute,
        "passive_only" => NotificationLevel::PassiveOnly,
        _ => NotificationLevel::AllowUrgent,
    };
    state
        .store
        .set_preference(&service, level)
        .await
        .map_err(|e| e.to_string())
}

// ── App entry point ─────────────────────────────────────────────────

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    // Open SQLite store at XDG data dir (always SQLite for desktop).
    let db_path = config::data_dir().join("scope.db");
    let db_url = format!("sqlite:{}", db_path.display());

    // Ensure data directory exists before opening DB.
    if let Some(parent) = db_path.parent() {
        std::fs::create_dir_all(parent).ok();
    }

    let store = tauri::async_runtime::block_on(async {
        scope_core::open_store(&db_url)
            .await
            .expect("Failed to open SQLite store")
    });

    let state = AppState::new(&api_base_url(), store);
    let proxy_state = state.clone();
    let refresh_state = state.clone();

    tauri::Builder::default()
        .manage(state)
        .plugin(tauri_plugin_notification::init())
        .invoke_handler(tauri::generate_handler![
            auth::login,
            auth::logout,
            auth::get_user,
            get_notifications,
            mark_notifications_read,
            get_preference,
            set_preference,
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
// TODO: unify with scope-core session management when available
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
