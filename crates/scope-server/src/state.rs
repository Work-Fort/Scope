use std::collections::HashMap;
use std::sync::Arc;

use scope_core::domain::session::FortTokens;
use scope_core::domain::{Notification, Store};
use scope_core::infra::discovery::ServiceDiscovery;
use scope_core::infra::proxy::ProxyHandler;
use tokio::sync::{broadcast, Mutex};

pub struct AppState {
    pub store: Arc<dyn Store>,
    pub discovery: Arc<ServiceDiscovery>,
    pub notify_tx: broadcast::Sender<Notification>,
    pub proxy: ProxyHandler,
    pub tokens: Mutex<HashMap<String, FortTokens>>,
}
