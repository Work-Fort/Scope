use std::collections::HashMap;
use std::sync::Arc;

use axum::{
    routing::{any, get, post},
    Router,
};
use tokio::sync::{broadcast, Mutex};
use tower_http::services::{ServeDir, ServeFile};

mod routes;
mod state;

use state::AppState;

#[tokio::main]
async fn main() {
    env_logger::init();

    // Load config
    let config = scope_core::config::Config::load().unwrap_or_else(|e| {
        log::error!("Failed to load config: {e}");
        std::process::exit(1);
    });

    let listen_addr = config.listen.clone();

    // Open store
    let store = scope_core::open_store(&config.database)
        .await
        .unwrap_or_else(|e| {
            log::error!("Failed to open database: {e}");
            std::process::exit(1);
        });

    // Seed forts from config into store
    for fort in config.into_forts() {
        if let Err(e) = store.upsert_fort(&fort).await {
            log::warn!("Failed to seed fort {}: {e}", fort.name);
        }
    }

    // Create discovery and notification broadcast
    let discovery = Arc::new(scope_core::infra::discovery::ServiceDiscovery::new());
    let (notify_tx, _) = broadcast::channel(256);

    let proxy = scope_core::infra::proxy::ProxyHandler::new();

    let state = Arc::new(AppState {
        store,
        discovery,
        notify_tx,
        proxy,
        tokens: Mutex::new(HashMap::new()),
    });

    // SPA fallback: serve static files, fall back to index.html for client-side routing
    let spa = ServeDir::new("web/shell/dist")
        .not_found_service(ServeFile::new("web/shell/dist/index.html"));

    // Build router
    let app = Router::new()
        .route("/api/forts", get(routes::api::list_forts))
        .route("/api/session", get(routes::api::session))
        .route("/api/services", get(routes::api::list_services))
        .route(
            "/api/notifications",
            get(routes::api::list_notifications),
        )
        .route(
            "/api/notifications/read",
            post(routes::api::mark_read),
        )
        .route(
            "/api/preferences/{service}",
            get(routes::api::get_preference).put(routes::api::set_preference),
        )
        // Proxy routes
        .route(
            "/forts/{fort}/api/{*rest}",
            any(routes::proxy::proxy_handler),
        )
        .route(
            "/forts/{fort}/ws/{*rest}",
            any(routes::proxy::ws_proxy_handler),
        )
        // SPA fallback (must be last)
        .fallback_service(spa)
        .with_state(state);

    // Start server
    log::info!("scope-server listening on {listen_addr}");
    let listener = tokio::net::TcpListener::bind(&listen_addr)
        .await
        .unwrap_or_else(|e| {
            log::error!("Failed to bind {listen_addr}: {e}");
            std::process::exit(1);
        });
    axum::serve(listener, app).await.unwrap();
}
