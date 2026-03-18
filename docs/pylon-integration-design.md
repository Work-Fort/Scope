# Pylon Integration Design

Scope changes needed to support fetching service listings from a Pylon server
instead of probing individual services directly.

See also: `~/Work/WorkFort/pylon/lead/docs/service-listing-requirements.md`

## Overview

For forts with `local: false` and a `pylon` URL, Scope fetches the service
list from Pylon at that URL rather than probing each service's `/ui/health`.
Pylon is **not a proxy** — it only serves listings. Scope's BFF still proxies
directly to each service's `base_url`.

## Config (no changes needed)

The config already supports this:

```yaml
forts:
  acme:
    local: false
    pylon: "https://pylon.acme.workfort.dev"
```

`FortYaml.pylon` and `Fort.pylon` are parsed and stored but currently
unused. This work activates them.

## Backend changes

### Discovery (`crates/scope-core/src/infra/discovery/mod.rs`)

Currently `probe_all()` iterates `fort.services` and probes each URL at
`/ui/health`. For Pylon forts, discovery should instead fetch the full
service list from the pylon.

When `!fort.local && fort.pylon.is_some()`:
- `GET {pylon}/api/services` → parse the `services` array
- Map each entry directly to `TrackedService` (the response format matches)
- Skip `/ui/health` probing entirely — Pylon already did it

When `fort.local`:
- Existing behavior, no changes

### Polling interval (`crates/scope-server/src/main.rs`)

Currently hard-coded to 10 seconds (line 111). For Pylon forts, poll every
**2 minutes** (120 seconds). Local forts keep the 10-second interval since
they're probing LAN services.

### Shell WebSocket (`crates/scope-server/src/routes/shell_ws.rs`)

Currently pushes `services_changed` only on initial connection (line 24).
When discovery detects a change in the service list (new service added,
service removed, `connected` status changed), it should broadcast an update
so the frontend doesn't have to wait for the next HTTP poll.

This requires discovery to compare the previous and current service lists
after each poll cycle and send a notification on the existing broadcast
channel when they differ.

### Proxy routing (`crates/scope-server/src/routes/proxy.rs`)

No changes needed. The proxy already resolves service URLs from discovery's
`TrackedService.base_url` and forwards directly. This works the same whether
the `base_url` came from a local health probe or from Pylon's listing.

## Frontend changes

### Fort picker (`web/shell/src/components/fort-picker.tsx`)

Currently calls `fetchForts()` and immediately displays the list.

**Change:** Before displaying the fort list, contact each non-local fort's
Pylon server to refresh its service URLs. This ensures the listings are
current when the user is choosing a fort. The fort list should show a
loading state while this check is in progress.

**HTTP warning:** If any service in a fort's listing has a `base_url` using
`http://` (not `https://`), display a warning on that fort's entry in the
selector. The warning should indicate that traffic to some services will
not be encrypted.

### Service polling (`web/shell/src/stores/services.ts`)

Currently polls every 30 seconds (`POLL_INTERVAL = 30_000` at line 8).

**Change:** For Pylon forts, poll every **2 minutes** (120,000ms). The
shell WebSocket already handles real-time pushes for service changes and
notifications, so the HTTP poll is a fallback. Local forts can keep the
current 30-second interval.

### MF remote registration (`web/shell/src/lib/remotes.ts`)

No changes needed. `registerNewRemotes()` already handles new services
appearing at runtime — it filters by `enabled && ui && !already_registered`
and calls `registerRemotes()`. Services discovered via Pylon will flow
through the same path.

## API surface

No new endpoints. The existing contract is sufficient:

| Endpoint | Purpose |
|----------|---------|
| `GET /api/forts` | List forts (already returns pylon field) |
| `GET /forts/{fort}/api/services` | Service list (already returns TrackedService format) |
| `GET /forts/{fort}/api/session` | Auth check (unchanged) |

The only difference is where the service data comes from internally — local
probing vs. Pylon fetch. The API response format is the same either way.

## Files to change

| File | Change |
|------|--------|
| `crates/scope-core/src/infra/discovery/mod.rs` | Add pylon fetch path alongside local probing |
| `crates/scope-server/src/main.rs` | Use 120s poll interval for non-local forts |
| `crates/scope-server/src/routes/shell_ws.rs` | Broadcast service changes on discovery update |
| `web/shell/src/components/fort-picker.tsx` | Pre-check Pylon, HTTP warning display |
| `web/shell/src/stores/services.ts` | 2-minute poll interval for Pylon forts |
