use std::sync::Arc;

use chrono::Utc;
use futures_util::StreamExt;
use tokio_tungstenite::tungstenite;

use crate::domain::{Notification, NotificationLevel, Store, Urgency};

pub struct NotificationSubscriber {
    store: Arc<dyn Store>,
}

/// Wire format for notifications arriving from services.
#[derive(serde::Deserialize)]
struct IncomingNotification {
    title: String,
    #[serde(default)]
    body: Option<String>,
    #[serde(default = "default_urgency")]
    urgency: Urgency,
    #[serde(default)]
    route: Option<String>,
}

fn default_urgency() -> Urgency {
    Urgency::Passive
}

impl NotificationSubscriber {
    pub fn new(store: Arc<dyn Store>) -> Self {
        Self { store }
    }

    /// Subscribe to a service's notification WebSocket.
    ///
    /// Calls `on_notification` for each notification that passes preference
    /// filtering. Returns when the connection is closed or an error occurs.
    pub async fn subscribe<F>(
        &self,
        service_name: &str,
        ws_url: &str,
        token: Option<&str>,
        on_notification: F,
    ) -> Result<(), crate::Error>
    where
        F: Fn(Notification) + Send + Sync + 'static,
    {
        // Build the WS request, optionally adding an auth header.
        let request = ws_url
            .parse::<tungstenite::http::Uri>()
            .map_err(|e| crate::Error::Other(format!("invalid ws url: {e}")))?;

        let ws_stream = if let Some(tok) = token {
            let req = tungstenite::http::Request::builder()
                .uri(&request)
                .header("Authorization", format!("Bearer {tok}"))
                .header("Connection", "Upgrade")
                .header("Upgrade", "websocket")
                .header("Sec-WebSocket-Version", "13")
                .header(
                    "Sec-WebSocket-Key",
                    tungstenite::handshake::client::generate_key(),
                )
                .header("Host", request.host().unwrap_or("localhost"))
                .body(())
                .map_err(|e| crate::Error::Other(format!("ws request build error: {e}")))?;
            let (stream, _resp) = tokio_tungstenite::connect_async(req)
                .await
                .map_err(|e| crate::Error::Other(format!("ws connect error: {e}")))?;
            stream
        } else {
            let (stream, _resp) = tokio_tungstenite::connect_async(ws_url)
                .await
                .map_err(|e| crate::Error::Other(format!("ws connect error: {e}")))?;
            stream
        };

        let (_write, mut read) = ws_stream.split();

        while let Some(msg_result) = read.next().await {
            let msg = match msg_result {
                Ok(m) => m,
                Err(e) => {
                    log::warn!("ws read error from {service_name}: {e}");
                    break;
                }
            };

            let text = match msg {
                tungstenite::Message::Text(t) => t,
                tungstenite::Message::Close(_) => break,
                _ => continue,
            };

            let incoming: IncomingNotification = match serde_json::from_str(&text) {
                Ok(n) => n,
                Err(e) => {
                    log::warn!("bad notification JSON from {service_name}: {e}");
                    continue;
                }
            };

            let notification = Notification {
                id: None,
                service: service_name.to_string(),
                title: incoming.title,
                body: incoming.body,
                urgency: incoming.urgency,
                route: incoming.route,
                read: false,
                created_at: Utc::now(),
            };

            // Check user preference for this service.
            let level = self
                .store
                .get_preference(service_name)
                .await
                .unwrap_or_default();

            let should_notify = match level {
                NotificationLevel::Mute => false,
                NotificationLevel::PassiveOnly => true,
                NotificationLevel::AllowUrgent => true,
            };

            // Always store for the inbox regardless of mute.
            if let Err(e) = self.store.insert_notification(&notification).await {
                log::warn!("failed to store notification from {service_name}: {e}");
            }

            if should_notify {
                on_notification(notification);
            }
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{
        extract::ws::{Message as AxumMessage, WebSocket, WebSocketUpgrade},
        response::IntoResponse,
        routing::get,
        Router,
    };
    use std::sync::Mutex;

    async fn ws_handler(ws: WebSocketUpgrade) -> impl IntoResponse {
        ws.on_upgrade(handle_socket)
    }

    async fn handle_socket(mut socket: WebSocket) {
        let payload = serde_json::json!({
            "title": "New message in #general",
            "body": "@admin mentioned you",
            "urgency": "active",
            "route": "/chat"
        });
        let _ = socket
            .send(AxumMessage::Text(payload.to_string().into()))
            .await;
        // Close after sending.
        let _ = socket
            .send(AxumMessage::Close(None))
            .await;
    }

    #[tokio::test]
    async fn subscriber_receives_notification() {
        let app = Router::new().route("/notifications/subscribe", get(ws_handler));

        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(async move {
            axum::serve(listener, app).await.unwrap();
        });

        // Open an in-memory SQLite store.
        let store = crate::open_store("sqlite::memory:").await.unwrap();

        let subscriber = NotificationSubscriber::new(store.clone());

        let received: Arc<Mutex<Vec<Notification>>> = Arc::new(Mutex::new(Vec::new()));
        let received_clone = Arc::clone(&received);

        let ws_url = format!("ws://{}/notifications/subscribe", addr);
        subscriber
            .subscribe("sharkfin", &ws_url, None, move |n| {
                received_clone.lock().unwrap().push(n);
            })
            .await
            .unwrap();

        // Verify callback was called.
        let captured = received.lock().unwrap();
        assert_eq!(captured.len(), 1);
        assert_eq!(captured[0].title, "New message in #general");
        assert_eq!(captured[0].body.as_deref(), Some("@admin mentioned you"));
        assert_eq!(captured[0].urgency, Urgency::Active);
        assert_eq!(captured[0].route.as_deref(), Some("/chat"));
        assert_eq!(captured[0].service, "sharkfin");

        // Verify stored in DB.
        let stored = store.list_notifications(10, None).await.unwrap();
        assert_eq!(stored.len(), 1);
        assert_eq!(stored[0].title, "New message in #general");
    }
}
