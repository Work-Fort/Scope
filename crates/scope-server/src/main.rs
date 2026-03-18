use std::sync::Arc;

use axum::{
    routing::{get, post},
    Router,
};
use tokio::sync::broadcast;

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

    let state = Arc::new(AppState {
        store,
        discovery,
        notify_tx,
    });

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
