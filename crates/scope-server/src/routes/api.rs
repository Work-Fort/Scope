use axum::{
    extract::{Path, Query, State},
    http::{HeaderMap, StatusCode},
    response::IntoResponse,
    Json,
};
use std::sync::Arc;

use crate::state::AppState;

/// GET /api/forts
pub async fn list_forts(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    match state.store.list_forts().await {
        Ok(forts) => Json(forts).into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response(),
    }
}

/// Validate session by forwarding the cookie to the auth service's /v1/session endpoint.
async fn check_auth_session(state: &AppState, headers: &HeaderMap) -> serde_json::Value {
    // Find the auth service URL from discovery
    let services = state.discovery.services().await;
    let auth_svc = services.iter().find(|s| s.name == "auth");
    let auth_url = match auth_svc {
        Some(svc) => &svc.base_url,
        None => return serde_json::json!({ "authenticated": false }),
    };

    // Extract cookie header from the incoming request
    let cookie = match headers.get("cookie").and_then(|v| v.to_str().ok()) {
        Some(c) => {
            log::debug!("session check: got cookie header ({} bytes)", c.len());
            c.to_string()
        }
        None => {
            log::debug!("session check: no cookie header in request");
            return serde_json::json!({ "authenticated": false });
        }
    };

    // Forward to Passport's /v1/session
    let client = reqwest::Client::new();
    let resp = client
        .get(format!("{auth_url}/v1/get-session"))
        .header("cookie", &cookie)
        .send()
        .await;

    let r = match resp {
        Ok(r) => r,
        Err(e) => {
            log::debug!("session check: passport request failed: {e}");
            return serde_json::json!({ "authenticated": false });
        }
    };

    let status = r.status();
    let body_text = r.text().await.unwrap_or_default();
    log::debug!("session check: passport {} — {}", status, &body_text[..body_text.len().min(200)]);

    if !status.is_success() {
        return serde_json::json!({ "authenticated": false });
    }

    let body: serde_json::Value = match serde_json::from_str(&body_text) {
        Ok(v) => v,
        Err(_) => return serde_json::json!({ "authenticated": false }),
    };

    if body.get("session").is_some() && body.get("user").is_some() {
        let user = &body["user"];
        let role = user.get("role").and_then(|r| r.as_str()).unwrap_or("user");
        serde_json::json!({ "authenticated": true, "role": role })
    } else {
        serde_json::json!({ "authenticated": false })
    }
}

/// GET /api/session
pub async fn session(
    State(state): State<Arc<AppState>>,
    headers: HeaderMap,
) -> impl IntoResponse {
    Json(check_auth_session(&state, &headers).await)
}

/// GET /api/services or GET /forts/{fort}/api/services
pub async fn list_services(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let services = state.discovery.services().await;
    // TODO: filter admin_only based on session role (Task 2 from passport plan)
    Json(serde_json::json!({ "services": services }))
}

/// GET /forts/{fort}/api/services — fort-scoped wrapper
pub async fn fort_services(
    State(state): State<Arc<AppState>>,
    Path(fort): Path<String>,
) -> impl IntoResponse {
    let services = state.discovery.services().await;
    Json(serde_json::json!({
        "fort": fort,
        "services": services,
        "conflicts": [],
    }))
}

/// GET /forts/{fort}/api/session — fort-scoped session check
pub async fn fort_session(
    State(state): State<Arc<AppState>>,
    Path(_fort): Path<String>,
    headers: HeaderMap,
) -> impl IntoResponse {
    Json(check_auth_session(&state, &headers).await)
}

/// GET /api/notifications?limit=20&before_id=123
#[derive(serde::Deserialize)]
pub struct NotificationQuery {
    #[serde(default = "default_limit")]
    limit: i64,
    before_id: Option<i64>,
}

fn default_limit() -> i64 {
    20
}

pub async fn list_notifications(
    State(state): State<Arc<AppState>>,
    Query(params): Query<NotificationQuery>,
) -> impl IntoResponse {
    match state
        .store
        .list_notifications(params.limit, params.before_id)
        .await
    {
        Ok(notifications) => Json(serde_json::json!({
            "notifications": notifications,
            "unread": state.store.unread_count().await.unwrap_or(0),
        }))
        .into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response(),
    }
}

/// POST /api/notifications/read  { "up_to_id": 123 }
pub async fn mark_read(
    State(state): State<Arc<AppState>>,
    Json(body): Json<serde_json::Value>,
) -> impl IntoResponse {
    let up_to_id = body.get("up_to_id").and_then(|v| v.as_i64()).unwrap_or(0);
    match state.store.mark_read(up_to_id).await {
        Ok(()) => StatusCode::OK.into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response(),
    }
}

/// GET /api/preferences/:service
pub async fn get_preference(
    State(state): State<Arc<AppState>>,
    Path(service): Path<String>,
) -> impl IntoResponse {
    match state.store.get_preference(&service).await {
        Ok(level) => {
            Json(serde_json::json!({ "service": service, "level": level })).into_response()
        }
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response(),
    }
}

/// PUT /api/preferences/:service  { "level": "mute" }
pub async fn set_preference(
    State(state): State<Arc<AppState>>,
    Path(service): Path<String>,
    Json(body): Json<serde_json::Value>,
) -> impl IntoResponse {
    let level_str = body
        .get("level")
        .and_then(|v| v.as_str())
        .unwrap_or("allow_urgent");
    let level = match level_str {
        "mute" => scope_core::domain::NotificationLevel::Mute,
        "passive_only" => scope_core::domain::NotificationLevel::PassiveOnly,
        _ => scope_core::domain::NotificationLevel::AllowUrgent,
    };
    match state.store.set_preference(&service, level).await {
        Ok(()) => StatusCode::OK.into_response(),
        Err(e) => (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response(),
    }
}
