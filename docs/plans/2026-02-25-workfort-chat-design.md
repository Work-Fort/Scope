# workfort chat — Design Document

## Overview

`workfort chat` is a Bubble Tea TUI that connects to a sharkfin daemon via WebSocket, providing a Slack/Discord-style chat interface with a channel listing sidebar and a main message area.

## Architecture

### Package Layout

```
cmd/chat/chat.go           — Command wiring: flags, client, tea.Program
pkg/sharkfin/client.go     — WebSocket client: ReadPump, WritePump, protocol
pkg/sharkfin/messages.go   — Protocol types: Envelope, requests, responses
pkg/ui/helpers.go          — Layout calculations, pane styles, size error
pkg/ui/keybindings.go      — KeyBinding type, rendering
pkg/ui/theme.go            — Color palette, lipgloss styles
internal/chat/model.go     — Root ChatModel: Init/Update/View
internal/chat/channels.go  — Channel list panel (left sidebar)
internal/chat/messages.go  — Message display panel (viewport)
internal/chat/input.go     — Text input bar
internal/chat/keys.go      — Chat-specific keybinding sets
internal/chat/msgtypes.go  — Bubble Tea message types for WS events
```

### Data Flow

```
WebSocket ──ReadPump──→ p.Send(msg) ──→ model.Update() ──→ View()
                                              │
                                              └──→ outbound channel ──→ WritePump ──→ WebSocket
```

- `pkg/sharkfin/` has no Bubble Tea dependency
- Bridge: `p.Send()` inbound, Go channel outbound
- `internal/chat/` contains all TUI-specific code
- `pkg/ui/` contains reusable rendering primitives

## Configuration

| Setting | Flag | Env | Config key | Default |
|---------|------|-----|------------|---------|
| Username | `--username` | `WORKFORT_USERNAME` | `username` | (required) |
| Daemon | `--sharkfin-host` | `WORKFORT_SHARKFIN_HOST` | `sharkfin-host` | `ws://127.0.0.1:16000/ws` |

Registration: auto-register (try identify first, fall back to register).

## UI Layout

```
┌─ workfort chat ──────────────────────────────────────────────┐
│  ┌─ Channels ──────┐  ┌─ #general ───────────────────────┐  │
│  │  # general       │  │  bob [10:32]                     │  │
│  │  # engineering   │  │  Hey team, standup in 5          │  │
│  │                  │  │                                  │  │
│  │  ── Direct ──    │  │  alice [10:33]                    │  │
│  │  alice           │  │  On it                           │  │
│  │                  │  │                                  │  │
│  │                  │  ├──────────────────────────────────┤  │
│  │                  │  │ > type a message...              │  │
│  └──────────────────┘  └──────────────────────────────────┘  │
│  [Ctrl+j/k] channels  [Ctrl+l] input  [Ctrl+q] quit         │
└──────────────────────────────────────────────────────────────┘
```

- Channel sidebar: ~25% width. Message area: ~75%.
- Input bar: fixed 3 lines at bottom of right pane.
- Minimum terminal size: 80x24. Below that, render centered size error.
- Active pane: thick border in primary color. Inactive: normal border, muted.

### Modals

- Channel create/edit modal (centered overlay)
- User invite modal (centered overlay)

## Keybindings

| Key | Action |
|-----|--------|
| `Ctrl+j` | Channel list: move down |
| `Ctrl+k` | Channel list: move up |
| `Ctrl+l` | Focus input |
| `Ctrl+o` | Focus channel list |
| `Enter` | Send message (input focused) / Select channel (channels focused) |
| `Ctrl+n` | Open new channel modal |
| `Ctrl+q` | Quit |

## Sharkfin Protocol

Endpoint: `ws://host/ws`

Envelope: `{"type": "...", "d": {...}, "ref": "...", "ok": true/false}`

Lifecycle: connect → hello → identify/register → requests

Request types: user_list, channel_list, channel_create, channel_invite, send_message, history, ping

Server push: message.new, presence

## Error Handling

- Pre-connection failure: print error, exit before TUI starts
- Mid-session disconnect: DisconnectedMsg renders modal, press q to quit
- No auto-reconnect in v1

## Theme

Industrial/utilitarian aesthetic. Dark background, alt-screen.

- Primary: muted amber/gold (active elements, selected channel)
- Secondary: cool gray (borders, inactive text)
- Accent: bright highlight (unread indicators)
