pub mod notifications;

use std::sync::Arc;
use std::time::Duration;

use reqwest::Client;
use tokio::sync::RwLock;

use crate::domain::{Fort, TrackedService};

fn default_display() -> String {
    "nav".to_string()
}

pub struct ServiceDiscovery {
    client: Client,
    services: Arc<RwLock<Vec<TrackedService>>>,
}

impl ServiceDiscovery {
    pub fn new() -> Self {
        Self {
            client: Client::builder()
                .timeout(Duration::from_secs(5))
                .build()
                .expect("failed to build discovery client"),
            services: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Probe all service URLs in a fort, parse /ui/health manifests.
    pub async fn probe_all(&self, fort: &Fort) {
        if !fort.local && fort.pylon.is_some() {
            return;
        }

        let mut discovered = Vec::new();
        for svc_config in &fort.services {
            if let Some(tracked) = self.probe_one(&svc_config.url).await {
                discovered.push(tracked);
            }
        }
        *self.services.write().await = discovered;
    }

    /// Fetch service list from a Pylon server. Used for non-local forts.
    /// Returns Some(passport_url) if authentication is required.
    /// Returns Some("__expired__") if the token is invalid (401).
    /// Returns None on success (services updated) or on network error.
    pub async fn fetch_from_pylon(
        &self,
        fort: &Fort,
        token: Option<&str>,
    ) -> Option<String> {
        let pylon_url = fort.pylon.as_deref()?;
        let base = pylon_url.trim_end_matches('/');
        let url = format!("{base}/api/services");

        let mut req = self.client.get(&url);
        if let Some(tok) = token {
            req = req.header("Authorization", format!("Bearer {tok}"));
        }

        let resp = match req.send().await {
            Ok(r) => r,
            Err(e) => {
                log::warn!("pylon fetch failed for {}: {e}", fort.name);
                return None;
            }
        };

        if resp.status().as_u16() == 401 {
            log::info!("pylon token expired for {}", fort.name);
            return Some("__expired__".into());
        }

        let body: serde_json::Value = match resp.json().await {
            Ok(v) => v,
            Err(e) => {
                log::warn!("pylon response parse failed for {}: {e}", fort.name);
                return None;
            }
        };

        // If response has passport_url, authentication is needed
        if let Some(passport_url) = body.get("passport_url").and_then(|v| v.as_str()) {
            return Some(passport_url.to_string());
        }

        // Parse the services array
        if let Some(services_val) = body.get("services") {
            let services: Vec<TrackedService> =
                serde_json::from_value(services_val.clone()).unwrap_or_default();
            *self.services.write().await = services;
        }

        None
    }

    /// Probe a single service URL. Returns None if unreachable or unparseable.
    async fn probe_one(&self, service_url: &str) -> Option<TrackedService> {
        let base = service_url.trim_end_matches('/');
        let url = format!("{}/ui/health", base);
        let resp = self.client.get(&url).send().await.ok()?;

        #[derive(serde::Deserialize)]
        struct HealthManifest {
            #[serde(default)]
            name: String,
            #[serde(default)]
            label: String,
            #[serde(default)]
            route: String,
            #[serde(default)]
            setup_mode: bool,
            #[serde(default)]
            admin_only: bool,
            #[serde(default = "default_display")]
            display: String,
            #[serde(default)]
            ws_paths: Vec<String>,
            #[serde(default)]
            notification_path: Option<String>,
        }

        let status = resp.status();
        let manifest: HealthManifest = resp.json().await.ok()?;

        Some(TrackedService {
            name: manifest.name,
            label: manifest.label,
            route: manifest.route,
            base_url: base.to_string(),
            ui: status.as_u16() == 200,
            connected: true,
            setup_mode: manifest.setup_mode,
            admin_only: manifest.admin_only,
            display: manifest.display,
            ws_paths: manifest.ws_paths,
            notification_path: manifest.notification_path,
        })
    }

    /// Start background polling. Returns a JoinHandle that can be aborted to stop.
    pub fn start_polling(
        self: &Arc<Self>,
        fort: Fort,
        interval: Duration,
    ) -> tokio::task::JoinHandle<()> {
        let discovery = Arc::clone(self);
        tokio::spawn(async move {
            loop {
                discovery.probe_all(&fort).await;
                tokio::time::sleep(interval).await;
            }
        })
    }

    /// Get a snapshot of currently tracked services.
    pub async fn services(&self) -> Vec<TrackedService> {
        self.services.read().await.clone()
    }

    /// Get services that have notification_path set.
    /// Returns (service_name, base_url, notification_path) tuples.
    pub async fn notification_services(&self) -> Vec<(String, String, String)> {
        self.services
            .read()
            .await
            .iter()
            .filter_map(|s| {
                s.notification_path
                    .as_ref()
                    .map(|path| (s.name.clone(), s.base_url.clone(), path.clone()))
            })
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::response::IntoResponse;
    use axum::{routing::get, Json, Router};
    use crate::domain::ServiceConfig;

    #[tokio::test]
    async fn discovery_probes_health() {
        let manifest = serde_json::json!({
            "name": "sharkfin",
            "label": "Sharkfin Chat",
            "route": "/chat",
            "setup_mode": false,
            "admin_only": false,
            "ws_paths": ["/ws"],
            "notification_path": "/notifications/subscribe"
        });

        let app = Router::new().route(
            "/ui/health",
            get(move || {
                let m = manifest.clone();
                async move { Json(m) }
            }),
        );

        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });

        let fort = Fort {
            name: "test".into(),
            local: true,
            pylon: None,
            services: vec![ServiceConfig {
                url: format!("http://{}", addr),
            }],
        };

        let discovery = ServiceDiscovery::new();
        discovery.probe_all(&fort).await;

        let services = discovery.services().await;
        assert_eq!(services.len(), 1);

        let svc = &services[0];
        assert_eq!(svc.name, "sharkfin");
        assert_eq!(svc.label, "Sharkfin Chat");
        assert_eq!(svc.route, "/chat");
        assert!(svc.ui);
        assert!(svc.connected);
        assert!(!svc.setup_mode);
        assert!(!svc.admin_only);
        assert_eq!(svc.ws_paths, vec!["/ws"]);
        assert_eq!(
            svc.notification_path.as_deref(),
            Some("/notifications/subscribe")
        );
        assert_eq!(svc.base_url, format!("http://{}", addr));
    }

    #[tokio::test]
    async fn discovery_skips_unreachable() {
        let fort = Fort {
            name: "test".into(),
            local: true,
            pylon: None,
            services: vec![ServiceConfig {
                url: "http://127.0.0.1:1".into(), // nothing listening
            }],
        };

        let discovery = ServiceDiscovery::new();
        discovery.probe_all(&fort).await;
        assert!(discovery.services().await.is_empty());
    }

    #[tokio::test]
    async fn notification_services_filters() {
        let discovery = ServiceDiscovery::new();
        {
            let mut svcs = discovery.services.write().await;
            svcs.push(TrackedService {
                name: "chat".into(),
                label: "Chat".into(),
                route: "/chat".into(),
                base_url: "http://localhost:3000".into(),
                ui: true,
                connected: true,
                setup_mode: false,
                admin_only: false,
                display: "nav".into(),
                ws_paths: vec![],
                notification_path: Some("/notifications/subscribe".into()),
            });
            svcs.push(TrackedService {
                name: "wiki".into(),
                label: "Wiki".into(),
                route: "/wiki".into(),
                base_url: "http://localhost:3001".into(),
                ui: true,
                connected: true,
                setup_mode: false,
                admin_only: false,
                display: "nav".into(),
                ws_paths: vec![],
                notification_path: None,
            });
        }

        let notif_services = discovery.notification_services().await;
        assert_eq!(notif_services.len(), 1);
        assert_eq!(notif_services[0].0, "chat");
        assert_eq!(notif_services[0].1, "http://localhost:3000");
        assert_eq!(notif_services[0].2, "/notifications/subscribe");
    }

    #[tokio::test]
    async fn discovery_267_sets_ui_false() {
        let manifest = serde_json::json!({
            "name": "nexus",
            "label": "VMs",
            "route": "/vms",
        });

        let app = Router::new().route(
            "/ui/health",
            get(move || {
                let m = manifest.clone();
                async move {
                    (axum::http::StatusCode::from_u16(267).unwrap(), Json(m))
                }
            }),
        );

        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });

        let fort = Fort {
            name: "test".into(),
            local: true,
            pylon: None,
            services: vec![ServiceConfig {
                url: format!("http://{}", addr),
            }],
        };

        let discovery = ServiceDiscovery::new();
        discovery.probe_all(&fort).await;

        let services = discovery.services().await;
        assert_eq!(services.len(), 1);
        assert_eq!(services[0].name, "nexus");
        assert!(services[0].connected);
        assert!(!services[0].ui); // 267 means no UI
    }

    #[tokio::test]
    async fn discovery_fetches_from_pylon() {
        let app = Router::new().route(
            "/api/services",
            get(|headers: axum::http::HeaderMap| async move {
                let auth = headers.get("authorization").and_then(|v| v.to_str().ok());
                match auth {
                    Some(h) if h.starts_with("Bearer ") => {
                        Json(serde_json::json!({
                            "services": [{
                                "name": "sharkfin",
                                "label": "Chat",
                                "route": "/chat",
                                "base_url": "http://10.0.0.1:16000",
                                "ui": true,
                                "connected": true,
                                "display": "nav",
                                "ws_paths": ["/ws"],
                            }]
                        }))
                        .into_response()
                    }
                    _ => {
                        Json(serde_json::json!({
                            "passport_url": "https://passport.example.com"
                        }))
                        .into_response()
                    }
                }
            }),
        );

        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });

        let fort = Fort {
            name: "acme".into(),
            local: false,
            pylon: Some(format!("http://{}", addr)),
            services: vec![],
        };

        let discovery = ServiceDiscovery::new();
        let result = discovery.fetch_from_pylon(&fort, Some("test-jwt")).await;

        // No passport_url returned when token is valid
        assert!(result.is_none());

        let services = discovery.services().await;
        assert_eq!(services.len(), 1);
        assert_eq!(services[0].name, "sharkfin");
        assert_eq!(services[0].base_url, "http://10.0.0.1:16000");
        assert!(services[0].ui);
    }

    #[tokio::test]
    async fn discovery_pylon_returns_passport_url_when_no_token() {
        let app = Router::new().route(
            "/api/services",
            get(|| async {
                Json(serde_json::json!({
                    "passport_url": "https://passport.example.com"
                }))
            }),
        );

        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });

        let fort = Fort {
            name: "acme".into(),
            local: false,
            pylon: Some(format!("http://{}", addr)),
            services: vec![],
        };

        let discovery = ServiceDiscovery::new();
        let result = discovery.fetch_from_pylon(&fort, None).await;

        assert_eq!(result.as_deref(), Some("https://passport.example.com"));
        assert!(discovery.services().await.is_empty());
    }
}
