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
        let mut discovered = Vec::new();
        for svc_config in &fort.services {
            if let Some(tracked) = self.probe_one(&svc_config.url).await {
                discovered.push(tracked);
            }
        }
        *self.services.write().await = discovered;
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
            ui: status.is_success(),
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
}
