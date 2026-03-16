# Tauri Mobile App — Design Spec

**Goal:** Run the WorkFort shell as a native mobile app (iOS + Android) via Tauri v2, with the Rust backend acting as the BFF proxy.

**Key Principle:** The Tauri Rust backend IS the BFF. It holds the JWT in memory and proxies all webview requests to the real API with auth headers attached. The webview never touches tokens. The shell web app stays almost unchanged.

> **Note:** The Tauri-specific auth client lives in `@workfort/auth-tauri` (`web/packages/auth/` in this repo). The general `@workfort/auth` package is published from the Passport repo and should not be duplicated here.

---

## Architecture

The Go BFF (`cmd/web/`) converts session cookies to JWTs before proxying requests to backend services. The Tauri Rust backend does exactly the same thing — it just skips the cookie step because it already has the token in memory.

```
┌──────────────────────────────────────────────┐
│  Mobile Device                               │
│                                              │
│  ┌──────────────────────────────────────┐    │
│  │  Tauri Rust Backend (BFF)            │    │
│  │                                      │    │
│  │  ┌────────────┐  ┌───────────────┐   │    │
│  │  │ JWT Store   │  │ HTTP Proxy    │   │    │
│  │  │ Arc<Mutex>  │  │ /api/* → API  │   │    │
│  │  └────────────┘  │ /forts/* → API │   │    │
│  │                   │ + Bearer hdr  │   │    │
│  │  ┌────────────┐  └───────────────┘   │    │
│  │  │ Auth Module │                     │    │
│  │  │ OAuth2/PKCE │                     │    │
│  │  └────────────┘                      │    │
│  └──────────┬───────────────────────────┘    │
│             │ custom protocol / IPC          │
│  ┌──────────▼───────────────────────────┐    │
│  │  Webview (WKWebView / Chromium)      │    │
│  │                                      │    │
│  │  SolidJS Shell (same code)           │    │
│  │  ├── @workfort/ui web components     │    │
│  │  ├── @workfort/auth (getAuthClient)  │    │
│  │  ├── Module Federation remotes       │    │
│  │  └── relative fetch("/api/...")      │    │
│  └──────────────────────────────────────┘    │
└──────────────────────────────────────────────┘
          │
          │ HTTPS (JWT in Authorization header)
          ▼
    ┌─────────────┐
    │ WorkFort API │
    └─────────────┘
```

### Why this works

1. **The shell already uses relative paths.** `web/shell/src/lib/api.ts` calls `/api/forts`, `/forts/:fort/api/services`, etc. Tauri's custom protocol handler intercepts these and proxies them to the real API with auth headers attached.

2. **Module Federation remotes load via relative URLs.** `web/shell/src/lib/remotes.ts` registers remotes at runtime. The proxy handles `remoteEntry.js` fetches the same way — auth headers get attached automatically.

3. **Web components work in any webview.** Lit light DOM components, CSS custom properties, UnoCSS — all standard web platform features supported by WKWebView (iOS 15+) and Android WebView (Chromium).

4. **Auth state stays the same.** `@workfort/auth` exposes `getAuthClient()` with event-based user state. On Tauri, the auth client calls Tauri commands (`get_user`, `login`, `logout`) instead of relying on cookie-based sessions. The interface is identical.

---

## Auth Flow

### Login

```
User taps "Login"
  → Webview calls Tauri command: invoke("login", { email, password })
  → Rust backend POSTs to /api/auth/login (or starts OAuth2/PKCE)
  → Rust backend receives JWT + refresh token
  → Stores JWT in Arc<Mutex<Option<String>>>
  → Stores refresh token in Arc<Mutex<Option<String>>>
  → Returns user info to webview
  → @workfort/auth emits "authenticated" event
```

### Token refresh

```
Proxy gets 401 from API
  → Rust backend uses refresh token to get new JWT
  → Retries the original request with new JWT
  → If refresh fails → clears tokens, emits "unauthenticated" to webview
```

### Logout

```
User taps "Logout"
  → Webview calls Tauri command: invoke("logout")
  → Rust backend clears JWT + refresh token from memory
  → Returns success
  → @workfort/auth emits "unauthenticated" event
```

### Token storage: memory only

The JWT and refresh token live in Rust memory (`Arc<Mutex<Option<String>>>`). Not in the webview, not in localStorage, not in Tauri Stronghold. When the app is killed, tokens are gone and the user logs in again. This is the simplest secure approach — no persistence means no extraction risk.

If persistent sessions become a requirement later, Stronghold is the correct upgrade path (encrypted keychain-backed storage). But start with memory.

---

## Proxy Design

### Custom protocol handler

Tauri v2 supports custom protocol handlers via `tauri::Builder::register_asynchronous_uri_scheme_protocol`. All requests from the webview to the app's origin go through this handler.

```rust
// Pseudocode
tauri::Builder::default()
    .register_asynchronous_uri_scheme_protocol("https", |_ctx, request, responder| {
        let url = request.uri();
        if url.path().starts_with("/api/") || url.path().starts_with("/forts/") {
            // Proxy to real API with JWT
            let mut proxied = reqwest::Request::new(method, api_base.join(url.path()));
            if let Some(jwt) = token_store.lock().unwrap().as_ref() {
                proxied.headers_mut().insert(
                    "Authorization",
                    format!("Bearer {}", jwt).parse().unwrap(),
                );
            }
            // Forward request body, content-type, etc.
            let response = client.execute(proxied).await;
            responder.respond(convert_response(response));
        } else {
            // Serve from embedded frontend assets
            responder.respond(serve_asset(url.path()));
        }
    })
```

### Route matching

| Pattern | Action |
|---------|--------|
| `/api/*` | Proxy to API with JWT |
| `/forts/*/api/*` | Proxy to API with JWT |
| `/forts/*/remoteEntry.js` | Proxy to API with JWT |
| `/*` (other) | Serve from embedded frontend assets |

### Request forwarding

The proxy forwards:
- HTTP method
- Request body
- `Content-Type` header
- Query parameters

The proxy adds:
- `Authorization: Bearer <jwt>`

The proxy strips:
- `Cookie` headers (not needed)
- `Origin` / `Referer` (prevent leaking app protocol)

---

## Shell Changes

### What stays the same

| Area | Why it works |
|------|-------------|
| `app.tsx`, all components | No auth/network code in components |
| `lib/api.ts` relative fetches | Proxy intercepts them transparently |
| `lib/remotes.ts` Module Federation | Remote URLs go through proxy |
| `stores/services.ts` polling | `setInterval` + `fetch` works in webview |
| `stores/theme.ts` | `localStorage` works in Tauri webview |
| `@workfort/ui` web components | Standard web platform, works in any webview |
| `global.css` + design tokens | CSS custom properties, `color-mix()` — all supported |
| UnoCSS utility classes | Build-time generated CSS, no runtime dependency |

### What changes

1. **`@workfort/auth` — Tauri adapter.** Add a Tauri-specific auth client implementation that calls Tauri commands instead of relying on cookie-based sessions. The `getAuthClient()` factory detects the Tauri environment and returns the right implementation.

```typescript
// @workfort/auth — new TauriAuthClient
import { invoke } from "@tauri-apps/api/core";

class TauriAuthClient implements AuthClient {
    async getUser() { return invoke("get_user"); }
    async login(creds) { return invoke("login", creds); }
    async logout() { return invoke("logout"); }
}

export function getAuthClient(): AuthClient {
    if (window.__TAURI_INTERNALS__) return new TauriAuthClient();
    return new WebAuthClient(); // existing cookie-based client
}
```

2. **Vite config — Tauri plugin.** Add `@tauri-apps/vite-plugin` to `web/shell/vite.config.ts` for dev mode (HMR over the Tauri dev server).

3. **`index.html` — conditional Tauri script.** Tauri injects its IPC bridge script automatically; no manual changes needed.

---

## Mobile UX

### Responsive layout

The shell currently uses a fixed 240px sidebar grid (`global.css`). For mobile:

```css
/* Breakpoint: 768px */
@media (max-width: 768px) {
    .shell-grid {
        grid-template-columns: 1fr;
        grid-template-rows: auto 1fr;
    }
    .shell-sidebar {
        position: fixed;
        inset: 0;
        z-index: 100;
        transform: translateX(-100%);
        transition: transform 0.2s ease;
    }
    .shell-sidebar[data-open] {
        transform: translateX(0);
    }
}
```

### iOS safe areas

```css
:root {
    --wf-safe-top: env(safe-area-inset-top);
    --wf-safe-bottom: env(safe-area-inset-bottom);
    --wf-safe-left: env(safe-area-inset-left);
    --wf-safe-right: env(safe-area-inset-right);
}

.shell-header {
    padding-top: var(--wf-safe-top);
}

.shell-content {
    padding-bottom: var(--wf-safe-bottom);
}
```

### Touch targets

All interactive elements must be at least 44x44pt (iOS HIG) / 48x48dp (Material). The existing `@workfort/ui` components mostly meet this. Audit needed for:
- Sidebar navigation items
- Fort picker items
- Service tabs

### Platform configuration

| Item | iOS | Android |
|------|-----|---------|
| Webview | WKWebView (iOS 15+) | Chromium WebView |
| Min OS | iOS 15 | Android 8 (API 26) |
| App icon | 1024x1024 + sizes | Adaptive icon |
| Splash | Storyboard-based | `windowSplashScreenBackground` |
| Deep links | `workfort://` via Associated Domains | `workfort://` via intent filter |
| Status bar | Light/dark based on theme | Edge-to-edge, themed |

---

## Phasing

### Phase A: Scaffolding + Proxy

**Goal:** Shell renders on iOS/Android simulator.

- [ ] Initialize Tauri v2 project in `src-tauri/`
- [ ] Configure `tauri.conf.json` with mobile targets
- [ ] Implement custom protocol handler for asset serving
- [ ] Implement HTTP proxy for `/api/*` and `/forts/*/api/*` routes
- [ ] Add `reqwest` for outbound HTTP
- [ ] Wire Vite build output as embedded frontend assets
- [ ] Verify shell loads and renders (without auth) on simulator

### Phase B: Auth Flow

**Goal:** User can log in and access authenticated endpoints.

- [ ] Implement `Arc<Mutex<Option<String>>>` token store
- [ ] Add Tauri commands: `login`, `logout`, `get_user`
- [ ] Implement token refresh on 401
- [ ] Add `TauriAuthClient` to `@workfort/auth`
- [ ] Wire `getAuthClient()` to detect Tauri environment
- [ ] Test full login → browse services → logout flow

### Phase C: Responsive Layout

**Goal:** Shell is usable on phone-sized screens.

- [ ] Add mobile breakpoints to `global.css`
- [ ] Implement collapsible sidebar with hamburger toggle
- [ ] Add iOS safe area inset CSS variables
- [ ] Audit touch target sizes on `@workfort/ui` components
- [ ] Test fort picker on mobile viewport

### Phase D: Platform Polish

**Goal:** App store ready.

- [ ] App icon generation (iOS + Android)
- [ ] Splash screen configuration
- [ ] Deep linking (`workfort://` URL scheme)
- [ ] Status bar theming (light/dark)
- [ ] Network transition handling (offline indicator)
- [ ] CSP audit for Module Federation script injection

---

## Risks

### 1. Module Federation + Tauri webview CSP

**Risk:** Module Federation injects `<script>` tags at runtime via `document.createElement('script')`. Tauri's default CSP may block dynamically created scripts.

**Mitigation:** Configure `tauri.conf.json` security policy to allow scripts from the proxied origin. Since all remote URLs go through the local proxy, the CSP only needs to allow the app's own origin — no external domains.

```json
{
    "app": {
        "security": {
            "csp": "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'"
        }
    }
}
```

`unsafe-eval` may be needed for Module Federation's dynamic module loading. Test without it first.

### 2. WebKit on iOS

**Risk:** WKWebView uses Safari's engine. CSS features like `color-mix()` (used in `primitives.css`) require Safari 15+.

**Mitigation:** iOS 15 is the minimum target (released 2021, ~98% adoption). All CSS features used in `@workfort/ui` are supported. Lit light DOM rendering has no WebKit-specific issues — no Shadow DOM compatibility concerns.

### 3. Network transitions

**Risk:** Mobile networks switch between WiFi and cellular. Requests in flight may fail silently.

**Mitigation:** The existing 30-second polling in `stores/services.ts` naturally recovers from transient failures. The Rust proxy should implement request timeouts (10s) and surface network errors to the webview for an offline indicator.

### 4. App size

**Risk:** Tauri bundles a Rust binary + the webview assets. Rust binary adds ~5-10 MB.

**Mitigation:** Tauri apps are significantly smaller than Electron. iOS WKWebView and Android WebView are system-provided — not bundled. Expected total app size: 15-25 MB.

### 5. Hot reload during development

**Risk:** Tauri mobile dev requires `tauri dev` which starts a dev server. HMR must work through the Tauri webview.

**Mitigation:** Use `@tauri-apps/vite-plugin` which configures Vite's dev server to work with Tauri's mobile dev flow. The webview connects to the Vite HMR websocket directly during development.
