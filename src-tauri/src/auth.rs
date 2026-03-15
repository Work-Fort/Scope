use serde::{Deserialize, Serialize};
use std::time::{Duration, Instant};
use tauri::State;

use crate::proxy::{AppState, FortTokens};

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
    #[serde(default)]
    expires_in: Option<u64>,
    user: UserInfo,
}

/// Tauri command: login to a specific fort with email/password.
/// Posts to that fort's auth service, stores JWT + refresh token under the fort name.
#[tauri::command]
pub async fn login(
    state: State<'_, AppState>,
    fort: String,
    auth_url: String,
    email: String,
    password: String,
) -> Result<UserInfo, String> {
    let target = format!("{}/v1/auth/login", auth_url);

    let resp = state.client
        .post(&target)
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

    let expiry = Instant::now() + Duration::from_secs(auth.expires_in.unwrap_or(900));
    state.tokens.set(&fort, FortTokens {
        jwt: auth.token,
        refresh_token: auth.refresh_token,
        expiry,
        auth_url,
    });

    Ok(auth.user)
}

/// Tauri command: logout from a specific fort. Removes that fort's tokens.
#[tauri::command]
pub async fn logout(
    state: State<'_, AppState>,
    fort: String,
) -> Result<(), String> {
    state.tokens.remove(&fort);
    Ok(())
}

/// Tauri command: get current user info for a specific fort.
/// Calls that fort's auth service using the stored JWT.
#[tauri::command]
pub async fn get_user(
    state: State<'_, AppState>,
    fort: String,
) -> Result<Option<UserInfo>, String> {
    let tokens = match state.tokens.get(&fort) {
        Some(t) => t,
        None => return Ok(None),
    };

    let target = format!("{}/v1/auth/me", tokens.auth_url);

    let resp = state.client
        .get(&target)
        .header("Authorization", format!("Bearer {}", tokens.jwt))
        .send()
        .await
        .map_err(|e| format!("Get user failed: {e}"))?;

    if resp.status().as_u16() == 401 {
        // Token expired, try refresh
        if crate::proxy::try_refresh(&state, &fort).await {
            // Retry with new JWT
            let new_tokens = state.tokens.get(&fort).unwrap();
            let target = format!("{}/v1/auth/me", new_tokens.auth_url);
            let resp = state.client
                .get(&target)
                .header("Authorization", format!("Bearer {}", new_tokens.jwt))
                .send()
                .await
                .map_err(|e| format!("Get user retry failed: {e}"))?;

            if resp.status().is_success() {
                let user: UserInfo = resp.json().await
                    .map_err(|e| format!("Invalid user response: {e}"))?;
                return Ok(Some(user));
            }
        }
        // Refresh failed or retry failed — user is not authenticated for this fort
        state.tokens.remove(&fort);
        return Ok(None);
    }

    if !resp.status().is_success() {
        return Err(format!("Get user failed: {}", resp.status()));
    }

    let user: UserInfo = resp.json().await
        .map_err(|e| format!("Invalid user response: {e}"))?;
    Ok(Some(user))
}
