use std::sync::Arc;

use scope_core::domain::{Notification, Store};
use scope_core::infra::discovery::ServiceDiscovery;
use tokio::sync::broadcast;

pub struct AppState {
    pub store: Arc<dyn Store>,
    pub discovery: Arc<ServiceDiscovery>,
    pub notify_tx: broadcast::Sender<Notification>,
}
