# Tauri Setup Screen — Design Spec

**Goal:** First-launch setup screen for the Tauri app that lets users connect to their WorkFort server via URL input, QR code scan, or deep link.

**Key Principle:** The setup screen appears when the boot provider's `configure()` returns `needsSetup: true`. Once a server URL is stored, the app re-boots into the normal fort picker flow. Three connection methods, all resolve to the same outcome: a server URL stored in the Rust proxy.

---

## Connection Methods

### 1. URL Input

Text input field + "Connect" button. User types or pastes their server URL (e.g. `https://acme.workfort.dev`). The connect action:

1. Calls `invoke('set_server_url', { url })`
2. Validates the URL by hitting `{url}/api/forts`
3. If valid → store URL, re-boot app
4. If invalid → show error ("Could not connect to server")

### 2. QR Code Scanner

JS-based camera scanner. The QR code encodes a URL in the format `workfort://connect?server=https://acme.workfort.dev`. The scanner:

1. Opens the device camera (or webcam on desktop)
2. Decodes the QR code
3. Extracts the `server` parameter
4. Pre-populates the URL input
5. Auto-connects (same flow as URL input)

Library: Use a lightweight JS QR scanner (e.g. `html5-qrcode` or `qr-scanner`) — no native Tauri plugin needed. Works on desktop and mobile since it uses standard `getUserMedia` API.

WorkFort servers can display their QR code in their admin panel for easy sharing.

### 3. Deep Link

`workfort://connect?server=https://acme.workfort.dev`

When tapped from a browser, email, or message:

1. Android/desktop opens the Tauri app via `tauri-plugin-deep-link`
2. The plugin fires an event with the URL
3. The app extracts the `server` parameter
4. Pre-populates the URL input
5. Auto-connects

Requires `tauri-plugin-deep-link` in `Cargo.toml` and the `workfort` scheme registered in `tauri.conf.json`.

---

## UI Layout

### First Boot (no servers)

```
┌─────────────────────────────────┐
│                                 │
│         WorkFort logo           │
│                                 │
│    Connect to your server       │
│                                 │
│  ┌───────────────────────────┐  │
│  │ https://                  │  │
│  └───────────────────────────┘  │
│                                 │
│       [ Connect ]               │
│                                 │
│  ─────── or ───────             │
│                                 │
│    [ Scan QR Code ]             │
│                                 │
└─────────────────────────────────┘
```

### Has Servers (server list)

```
┌─────────────────────────────────┐
│                                 │
│    Your Servers                 │
│                                 │
│  ┌───────────────────────────┐  │
│  │ ● acme.workfort.dev       │  │
│  ├───────────────────────────┤  │
│  │ ○ dev.workfort.dev        │  │
│  └───────────────────────────┘  │
│                                 │
│     [ Add Server ]              │
│                                 │
└─────────────────────────────────┘
```

Tapping a server selects it and proceeds to the fort picker for that server. "Add Server" opens a `wf-dialog` modal with the same URL input + QR scanner from the first-boot screen.

- **Error state** — red border on input + error message below if connection fails
- **Loading state** — spinner on the Connect button while validating

---

## Component

The setup screen is a SolidJS component at `web/shell/src/components/setup-screen.tsx`. It uses `@workfort/ui` components:

- `wf-input` for the URL field
- `wf-button` for Connect and Scan QR Code
- `wf-list` / `wf-list-item` for recent servers
- `wf-panel` as the container
- `wf-spinner` for loading state
- `wf-banner` for errors

The QR scanner modal uses `wf-dialog` when the camera is active.

---

## Rust Commands

| Command | Params | Returns | Notes |
|---------|--------|---------|-------|
| `set_server_url` | `url: String` | `Result<(), String>` | Sets the active server, already exists |
| `get_server_url` | — | `Result<Option<String>, String>` | Gets the active server, already exists |
| `get_servers` | — | `Result<Vec<ServerEntry>, String>` | New — list all configured servers |
| `add_server` | `url: String` | `Result<(), String>` | New — validates and adds a server |
| `remove_server` | `url: String` | `Result<(), String>` | New — removes a server |

```rust
struct ServerEntry {
    url: String,
    name: Option<String>,  // friendly name, populated after first connect
}
```

Servers are stored in a JSON file at the Tauri app data directory (`app_data_dir/servers.json`). The active server is whichever the user last selected.

---

## Deep Link Registration

`src-tauri/tauri.conf.json`:
```json
{
  "plugins": {
    "deep-link": {
      "desktop": { "schemes": ["workfort"] },
      "mobile": { "schemes": ["workfort"] }
    }
  }
}
```

`src-tauri/Cargo.toml`:
```toml
tauri-plugin-deep-link = "2"
```

The deep link handler in `lib.rs` listens for the URL, extracts the `server` query param, and stores it. If the app is already running and configured, it switches servers.

---

## Flow Integration

The `SetupScreen` component in `app.tsx` (currently a stub) is replaced with this full implementation.

### First boot (no servers)

1. `configure()` → no servers → `{ needsSetup: true }`
2. App renders `SetupScreen` (full screen URL input + QR)
3. User connects via URL/QR/deep link
4. `add_server` validates and stores it, `set_server_url` makes it active
5. Component calls `window.location.reload()` to re-boot
6. `configure()` → server URL found → `{ needsSetup: false }`
7. App renders fort picker → normal flow

### Has servers

1. `configure()` → active server URL found → `{ needsSetup: false }`
2. App renders fort picker for that server
3. User can access server list via settings/menu
4. Server list shows all configured servers with "Add Server" button
5. "Add Server" opens modal with URL input + QR
6. Tapping a different server calls `set_server_url` and re-boots
