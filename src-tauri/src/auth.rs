use serde::{Deserialize, Serialize};
use tauri::State;

use crate::proxy::AppState;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct UserInfo {
    pub id: String,
    pub email: String,
    pub name: String,
}

#[derive(Debug, Deserialize)]
struct AuthResponse {
    token: String,
    refresh_token: String,
    user: UserInfo,
}

/// Tauri command: login with email/password.
/// Posts to the API auth endpoint, stores JWT + refresh token in memory.
#[tauri::command]
pub async fn login(
    state: State<'_, AppState>,
    email: String,
    password: String,
) -> Result<UserInfo, String> {
    let target = state.api_base.join("/api/auth/login").unwrap();

    let resp = state.client
        .post(target)
        .json(&serde_json::json!({
            "email": email,
            "password": password,
        }))
        .send()
        .await
        .map_err(|e| format!("Login request failed: {e}"))?;

    if !resp.status().is_success() {
        let status = resp.status().as_u16();
        let body = resp.text().await.unwrap_or_default();
        return Err(format!("Login failed ({status}): {body}"));
    }

    let auth: AuthResponse = resp.json().await
        .map_err(|e| format!("Invalid auth response: {e}"))?;

    state.tokens.set_jwt(Some(auth.token));
    state.tokens.set_refresh(Some(auth.refresh_token));

    Ok(auth.user)
}

/// Tauri command: logout. Clears JWT + refresh token from memory.
#[tauri::command]
pub async fn logout(state: State<'_, AppState>) -> Result<(), String> {
    state.tokens.clear();
    Ok(())
}

/// Tauri command: get current user info.
/// Calls the API's user endpoint using the stored JWT.
#[tauri::command]
pub async fn get_user(state: State<'_, AppState>) -> Result<Option<UserInfo>, String> {
    let jwt = match state.tokens.get_jwt() {
        Some(j) => j,
        None => return Ok(None),
    };

    let target = state.api_base.join("/api/auth/me").unwrap();

    let resp = state.client
        .get(target)
        .header("Authorization", format!("Bearer {jwt}"))
        .send()
        .await
        .map_err(|e| format!("Get user failed: {e}"))?;

    if resp.status().as_u16() == 401 {
        // Token expired, try refresh
        if crate::proxy::try_refresh(&state).await {
            // Retry with new JWT
            let new_jwt = state.tokens.get_jwt().unwrap();
            let target = state.api_base.join("/api/auth/me").unwrap();
            let resp = state.client
                .get(target)
                .header("Authorization", format!("Bearer {new_jwt}"))
                .send()
                .await
                .map_err(|e| format!("Get user retry failed: {e}"))?;

            if resp.status().is_success() {
                let user: UserInfo = resp.json().await
                    .map_err(|e| format!("Invalid user response: {e}"))?;
                return Ok(Some(user));
            }
        }
        // Refresh failed or retry failed — user is not authenticated
        state.tokens.clear();
        return Ok(None);
    }

    if !resp.status().is_success() {
        return Err(format!("Get user failed: {}", resp.status()));
    }

    let user: UserInfo = resp.json().await
        .map_err(|e| format!("Invalid user response: {e}"))?;
    Ok(Some(user))
}
