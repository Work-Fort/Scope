# DM Refactor Design

## Problem

DMs are modeled as regular channels, polluting the channel list and creating confusion. Channel names like `tpm-workfort-cli-team-lead` are ad-hoc, and the public/private flags are inconsistent for DMs.

## Approach

Add a `type` column (`channel` | `dm`) to the server's channels table. The existing message, history, unread, and cursor infrastructure stays unchanged — it operates on channel names regardless of type. New server endpoints handle DM-specific concerns.

## Server Changes (sharkfin-team-lead)

**Migration**: Add `type TEXT NOT NULL DEFAULT 'channel'` column. Migrate existing DMs (private channels with exactly 2 members) to `type='dm'`.

**Modified endpoints**:
- `channel_list` — filters to `type='channel'` only
- `unread_counts` — adds `type` field to each entry
- `message.new` broadcast — adds `channel_type` field

**New endpoints** (WS + MCP):
- `dm_list` — returns `[{channel, participant, member}]` where `participant` is the other user's username
- `dm_open {username}` — find-or-create DM, returns `{channel, participant, created}`

**Unchanged**: `send_message`, `history`, `unread_messages`, `mark_read` — all work on channel names, agnostic of type.

## TUI Changes (workfort-cli-team-lead)

### Sidebar Tabs

The sidebar gets a tab bar at the bottom (near the help bar) with two tabs: **Channels** and **DMs**.

- **Ctrl+A** toggles between tabs
- Each tab has its own scrollable list and cursor
- Unread/mention badges route to the correct tab based on `channel_type`
- The inactive tab shows a badge count if it has unreads

### DM List Rendering

DMs display the participant's username (e.g. `tpm`) instead of the channel name. No `#` prefix — use a different visual indicator (e.g. the username directly, or a `@` prefix).

### New DM Modal

**Ctrl+D** opens a modal (same pattern as Ctrl+N for new channel) where you type a username. Calls `dm_open` to find-or-create, then switches to the conversation.

### Data Flow

1. On startup: `channel_list` (channels only) + `dm_list` + `unread_counts`
2. Tab toggle: switch visible list, preserve each tab's cursor position
3. Channel/DM selection: same as today — `mark_read`, load history if not cached
4. `message.new` with `channel_type`: route unread badge to correct tab
5. New DM: `dm_open` → add to DM list → switch to it

## New Types

```go
// DM represents a direct message conversation.
type DM struct {
    Channel     string `json:"channel"`
    Participant string `json:"participant"`
    Member      bool   `json:"member"`
}

type DMListResponse struct {
    DMs []DM `json:"dms"`
}

type DMOpenRequest struct {
    Username string `json:"username"`
}

type DMOpenResponse struct {
    Channel     string `json:"channel"`
    Participant string `json:"participant"`
    Created     bool   `json:"created"`
}
```

## Keybinds

| Key | Action |
|-----|--------|
| Ctrl+A | Toggle sidebar tab (Channels / DMs) |
| Ctrl+D | Open new DM modal |
| Ctrl+J/K | Navigate within active tab |

## Implementation Order

1. Server: schema migration + `dm_list` + `dm_open` + modified `channel_list` + enriched `unread_counts` + `message.new` with `channel_type`
2. TUI: new sharkfin types (`DM`, `DMListResponse`, `DMOpenRequest/Response`, `DMListMsg`, `DMOpenMsg`)
3. TUI: client methods (`RequestDMList`, `DMOpen`) + dispatch cases
4. TUI: sidebar tab component (tab bar rendering, tab state, Ctrl+A toggle)
5. TUI: DM list component (like ChannelList but renders participant names)
6. TUI: model wiring (Init requests dm_list, tab switching, message.new routing, Ctrl+D modal)
7. TUI: help bar update with new keybinds
