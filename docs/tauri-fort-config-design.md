# Tauri Fort Configuration — Design Spec

**Goal:** Let Tauri app users create and configure forts by manually entering service URLs. This replicates the Go BFF's config file approach through a UI. The design supports a future transition to gateway-based service discovery.

**Key Principle:** The BFF is the source of truth for how forts work. The Tauri app replicates the same `Fort` model (name, local flag, gateway, services list) but stores it in a JSON file instead of a viper config. The fort config UI is the Tauri equivalent of editing the BFF's config file.

---

## Fort Model (matches Go domain)

```rust
struct FortConfig {
    name: String,
    local: bool,              // true = direct service URLs, false = gateway
    gateway: Option<String>,  // single origin URL when local=false
    services: Vec<ServiceConfig>,
}

struct ServiceConfig {
    url: String,              // e.g. "http://192.168.1.50:16000"
    // name/label are discovered at runtime by probing the service
}
```

This mirrors `domain.Fort` in the Go codebase exactly.

---

## Two Modes

### Manual Mode (now)

`local: true` — the user enters individual service URLs. The Tauri proxy connects directly to each service, just like the Go BFF does for local forts.

```
Fort "home-lab"
  local: true
  services:
    - http://192.168.1.50:16000   (Sharkfin)
    - http://192.168.1.50:9600    (Nexus)
```

The Rust proxy probes each service URL (like the Go `ServiceTracker` does) to discover the service name, check connectivity, and find the auth service.

### Gateway Mode (future)

`local: false` — the user enters one gateway URL. The gateway handles routing to all services behind it.

```
Fort "acme-corp"
  local: false
  gateway: https://acme.workfort.dev
```

The Rust proxy sends all requests through the gateway. Service discovery happens via the gateway's `/api/services` endpoint.

---

## UI Flow

### First Boot

1. Boot provider returns `needsSetup: true` (no forts configured)
2. App shows the fort setup screen
3. User creates their first fort

### Fort Setup Screen

```
┌─────────────────────────────────┐
│                                 │
│         WorkFort logo           │
│                                 │
│      Create a Fort              │
│                                 │
│  Fort name                      │
│  ┌───────────────────────────┐  │
│  │ home-lab                  │  │
│  └───────────────────────────┘  │
│                                 │
│  Connection type                │
│  ○ Direct (enter service URLs)  │
│  ○ Gateway (single URL)         │
│                                 │
│  ─── Services ───               │
│                                 │
│  ┌───────────────────────────┐  │
│  │ http://192.168.1.50:16000 │ ×│
│  ├───────────────────────────┤  │
│  │ http://192.168.1.50:9600  │ ×│
│  └───────────────────────────┘  │
│                                 │
│  ┌───────────────────────────┐  │
│  │ Add service URL...        │  │
│  └───────────────────────────┘  │
│                                 │
│       [ Create Fort ]           │
│                                 │
│  ─────── or ───────             │
│                                 │
│    [ Scan QR Code ]             │
│                                 │
└─────────────────────────────────┘
```

When "Gateway" is selected, the services list disappears and a single gateway URL input appears instead.

### Fort List (has forts)

Same as the setup screen design — list of configured forts, "Add Fort" button opens the creation dialog. Tapping a fort selects it and enters the fort picker / service view.

---

## QR Code & Deep Link

The QR/deep link format encodes the full fort config:

```
workfort://fort?name=home-lab&local=true&services=http://192.168.1.50:16000,http://192.168.1.50:9600
```

Or for gateway mode:

```
workfort://fort?name=acme-corp&gateway=https://acme.workfort.dev
```

This replaces the simpler `workfort://connect?server=URL` from the previous design. The URL pre-populates the fort creation form so the user just reviews and taps "Create Fort."

---

## Rust Commands

| Command | Params | Returns |
|---------|--------|---------|
| `get_forts` | — | `Vec<FortConfig>` |
| `add_fort` | `FortConfig` | `Result<()>` |
| `remove_fort` | `name: String` | `Result<()>` |
| `set_active_fort` | `name: String` | `Result<()>` |
| `get_active_fort` | — | `Option<String>` |

Replaces the `get_servers`/`add_server`/`remove_server` commands from the previous design. The data model is richer — forts have names, connection modes, and service lists.

---

## Storage

Forts are stored in `app_data_dir/forts.json`:

```json
{
  "active": "home-lab",
  "forts": [
    {
      "name": "home-lab",
      "local": true,
      "services": [
        { "url": "http://192.168.1.50:16000" },
        { "url": "http://192.168.1.50:9600" }
      ]
    }
  ]
}
```

---

## Proxy Behavior

When a fort is active, the Rust proxy uses its config to route requests:

**Local fort:** Each service URL is tracked individually. The proxy matches the request path to the service (same as Go's `ServiceTracker`):
- Probes each service URL on fort activation
- Discovers service names and capabilities
- Routes `/forts/{fort}/api/{service}/*` to the matching service URL
- Finds the auth service for per-fort JWT management

**Gateway fort (future):** All requests go through the single gateway URL:
- Routes `/forts/{fort}/api/*` to `{gateway}/api/*`
- Gateway handles internal service routing

---

## Transition Path

The UI supports both modes from day one. The "Gateway" radio option works as soon as a gateway is available — no UI changes needed. The transition is:

1. **Now:** Users create local forts with manual service URLs
2. **When gateways ship:** Users create gateway forts with a single URL
3. **Migration:** Users can edit existing local forts to switch to gateway mode, or create new gateway forts alongside local ones

No breaking changes, no migration scripts. Both modes coexist.

---

## Boot Provider Integration

The Tauri boot provider's `configure()` changes:

```typescript
async configure() {
  const activeFort = await invoke<string | null>('get_active_fort');
  if (!activeFort) {
    return { apiBase: '', environment: 'tauri', needsSetup: true };
  }
  return { apiBase: '', environment: 'tauri', activeFort };
}
```

The `apiBase` stays empty because the Rust proxy handles all routing internally based on the active fort's config. The shell doesn't need to know service URLs — it just makes relative requests and the proxy figures out where to send them.

---

## Supersedes

This design supersedes `docs/setup-screen-design.md`. The "server URL" concept is replaced by the richer "fort configuration" model that matches the Go BFF's architecture.
