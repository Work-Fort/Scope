use axum::{
    extract::{
        ws::{Message, WebSocket, WebSocketUpgrade},
        State,
    },
    response::IntoResponse,
};
use std::sync::Arc;

use crate::state::AppState;

pub async fn shell_ws_handler(
    State(state): State<Arc<AppState>>,
    ws: WebSocketUpgrade,
) -> impl IntoResponse {
    ws.on_upgrade(move |socket| handle_shell_ws(socket, state))
}

async fn handle_shell_ws(mut socket: WebSocket, state: Arc<AppState>) {
    // Subscribe to the notification broadcast channel
    let mut notify_rx = state.notify_tx.subscribe();

    // Send initial service list
    let services = state.discovery.services().await;
    let initial = serde_json::json!({
        "type": "services_changed",
        "data": services,
    });
    if socket
        .send(Message::Text(initial.to_string().into()))
        .await
        .is_err()
    {
        return;
    }

    // Event loop: forward notifications and handle client messages
    loop {
        tokio::select! {
            // Notification from broadcast channel
            Ok(notification) = notify_rx.recv() => {
                let event = serde_json::json!({
                    "type": "notification",
                    "data": notification,
                });
                if socket.send(Message::Text(event.to_string().into())).await.is_err() {
                    break;
                }
            }
            // Message from client (ping/pong, or close)
            msg = socket.recv() => {
                match msg {
                    Some(Ok(Message::Ping(data))) => {
                        if socket.send(Message::Pong(data)).await.is_err() {
                            break;
                        }
                    }
                    Some(Ok(Message::Close(_))) | None => break,
                    _ => {} // Ignore other messages for now
                }
            }
        }
    }
}
