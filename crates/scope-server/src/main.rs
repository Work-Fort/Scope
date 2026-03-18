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
        .route("/ws/shell", get(routes::shell_ws::shell_ws_handler))
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
        .with_state(Arc::clone(&state));

    // Start discovery polling for the first fort (if any)
    {
        let forts = state.store.list_forts().await.unwrap_or_default();
        if let Some(fort) = forts.first() {
            let discovery = Arc::clone(&state.discovery);
            let fort_clone = fort.clone();
            tokio::spawn(async move {
                loop {
                    discovery.probe_all(&fort_clone).await;
                    tokio::time::sleep(std::time::Duration::from_secs(10)).await;
                }
            });
        }
    }

    // Start notification subscriber for services that support it
    {
        let state_clone = Arc::clone(&state);
        tokio::spawn(async move {
            // Wait for initial discovery probe to complete
            tokio::time::sleep(std::time::Duration::from_secs(2)).await;
            let services = state_clone.discovery.services().await;
            for svc in services {
                if let Some(notif_path) = &svc.notification_path {
                    let ws_url = format!(
                        "{}{}",
                        svc.base_url
                            .replace("http://", "ws://")
                            .replace("https://", "wss://"),
                        notif_path
                    );
                    let subscriber =
                        scope_core::infra::discovery::notifications::NotificationSubscriber::new(
                            Arc::clone(&state_clone.store),
                        );
                    let tx = state_clone.notify_tx.clone();
                    let svc_name = svc.name.clone();
                    tokio::spawn(async move {
                        log::info!(
                            "subscribing to notifications from {svc_name} at {ws_url}"
                        );
                        if let Err(e) = subscriber
                            .subscribe(
                                &svc_name,
                                &ws_url,
                                None, // TODO: attach fort token
                                move |notif| {
                                    let _ = tx.send(notif);
                                },
                            )
                            .await
                        {
                            log::warn!(
                                "notification subscriber for {svc_name} failed: {e}"
                            );
                        }
                    });
                }
            }
        });
    }

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
