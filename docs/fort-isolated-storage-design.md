# Fort-Isolated Storage — Design Spec

**Goal:** Prevent service remotes in one fort from reading another fort's localStorage data by encrypting storage entries with per-fort keys held in the BFF/Tauri proxy.

**Key Principle:** The encryption key never enters the browser's JavaScript context. The BFF (Go) or Tauri (Rust) proxy owns the key and performs encrypt/decrypt operations. The webview only ever sees ciphertext in localStorage.

**Status:** Design backlog. Build when forts start hosting untrusted third-party service remotes.

---

## Threat Model

All forts are served from the same origin (same BFF/Tauri proxy). Browser storage APIs (`localStorage`, `sessionStorage`, `IndexedDB`) are origin-scoped — any JavaScript running on the page can access all stored data.

Service remotes are loaded via Module Federation and execute in the same JavaScript context as the shell. A hostile remote could:
- Read all localStorage keys
- Extract user preferences, cached data, or tokens from other forts
- Write malicious data to keys used by other services

---

## Solution

### Per-Fort Encryption Keys

Each fort has a symmetric encryption key (AES-256-GCM). The key is:
- Generated or provided by the fort's auth service during login
- Stored in the BFF/Tauri proxy's memory alongside the fort's JWT
- Never sent to the webview
- Lost when the app is killed (keys are session-scoped)

### Storage Abstraction

A `FortStorage` API replaces direct `localStorage` access:

```typescript
// @workfort/ui or shell utility
class FortStorage {
  private fort: string;

  constructor(fort: string) {
    this.fort = fort;
  }

  async setItem(key: string, value: string): Promise<void> {
    const encrypted = await invoke('storage_encrypt', {
      fort: this.fort,
      data: value,
    });
    localStorage.setItem(`${this.fort}:${key}`, encrypted);
  }

  async getItem(key: string): Promise<string | null> {
    const encrypted = localStorage.getItem(`${this.fort}:${key}`);
    if (!encrypted) return null;
    return invoke('storage_decrypt', {
      fort: this.fort,
      data: encrypted,
    });
  }

  removeItem(key: string): void {
    localStorage.removeItem(`${this.fort}:${key}`);
  }
}
```

### Proxy Commands

**Tauri (Rust):**

```rust
#[tauri::command]
async fn storage_encrypt(
    fort: String,
    data: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    let keys = state.storage_keys.lock().unwrap();
    let key = keys.get(&fort).ok_or("No storage key for fort")?;
    // AES-256-GCM encrypt, return base64
    encrypt_aes_gcm(key, data.as_bytes())
}

#[tauri::command]
async fn storage_decrypt(
    fort: String,
    data: String,
    state: State<'_, AppState>,
) -> Result<String, String> {
    let keys = state.storage_keys.lock().unwrap();
    let key = keys.get(&fort).ok_or("No storage key for fort")?;
    // AES-256-GCM decrypt from base64
    decrypt_aes_gcm(key, &data)
}
```

**Go BFF:** Same pattern — HTTP endpoints or a middleware that intercepts storage-related API calls.

### Key Lifecycle

```
User logs into fort "acme"
  → Auth service returns JWT + storage_key
  → Proxy stores storage_key in memory: Map<fort_name, AES key>
  → Proxy stores JWT (existing behavior)

User accesses storage
  → FortStorage.setItem() → invoke('storage_encrypt') → proxy encrypts → ciphertext to localStorage
  → FortStorage.getItem() → read ciphertext → invoke('storage_decrypt') → proxy decrypts → plaintext

User switches fort
  → Previous fort's key remains in memory (can still decrypt if user switches back)
  → New fort's key loaded on login

App killed
  → All keys lost from memory
  → localStorage still has ciphertext (encrypted at rest)
  → Next launch: user logs in again → keys restored → can decrypt again
```

---

## What This Protects Against

| Attack | Protected? | How |
|--------|-----------|-----|
| Hostile remote reads another fort's localStorage | Yes | Data is encrypted, key not in JS context |
| Hostile remote writes to another fort's keys | Partial | Can write ciphertext but can't produce valid encrypted data |
| XSS within the shell | No | If the shell itself is compromised, the attacker can call invoke() |
| Proxy compromise | No | Keys are in proxy memory — if proxy is compromised, game over |

---

## What This Does NOT Replace

- **Auth tokens must stay in the proxy** — never in localStorage, encrypted or not
- **Session management stays server-side** — this is only for UI preferences, cached data
- **CSP and Module Federation sandboxing** are still needed for defense in depth

---

## Migration Path

1. **Now:** Services use `localStorage` directly with fort-prefixed keys (convention-based)
2. **When needed:** Introduce `FortStorage` API, services migrate to use it
3. **Enforcement:** Shell can intercept direct `localStorage` calls via a Proxy wrapper and log/warn

No breaking changes — `FortStorage` is opt-in. Services that don't use sensitive per-fort data can continue using plain `localStorage`.

---

## Dependencies

- Tauri: `aes-gcm` crate (or `ring`)
- Go BFF: `crypto/aes` + `crypto/cipher` (stdlib)
- Auth service: new endpoint or field in login response to provide the storage key
