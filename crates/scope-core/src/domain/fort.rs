use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Fort {
    pub name: String,
    pub local: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub pylon: Option<String>,
    pub services: Vec<ServiceConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceConfig {
    pub url: String,
}

/// Discovered at runtime by probing /ui/health.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrackedService {
    pub name: String,
    pub label: String,
    pub route: String,
    /// The original base URL used to probe this service.
    pub base_url: String,
    pub ui: bool,
    pub connected: bool,
    #[serde(default)]
    pub setup_mode: bool,
    #[serde(default)]
    pub admin_only: bool,
    #[serde(default = "default_display")]
    pub display: String,
    #[serde(default)]
    pub ws_paths: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub notification_path: Option<String>,
}

fn default_display() -> String {
    "nav".to_string()
}
