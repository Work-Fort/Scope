# Plan 2: WebSocket Client & Business Logic

Wire up the sharkfin WebSocket client, replace all placeholder data with live data, and make the chat fully functional.

## Sharkfin Protocol Reference

- Endpoint: `ws://<host>/ws`
- JSON envelope: `{"type":"...","d":{...},"ref":"...","ok":true/false}`
- Flow: connect → hello (server sends heartbeat interval) → identify/register → requests
- Request types: `user_list`, `channel_list`, `channel_create`, `channel_invite`, `send_message`, `history`, `unread_messages`
- Server push: `message.new`, `presence`
- Unread tracking: server stores `(channel, user, last_read_message_id)` cursor. Calling `unread_messages` (no filters) returns messages with `id > cursor` AND advances cursor. Filtered calls (`mentions_only`, `thread_id`) do NOT advance cursor.

## Steps

### 1. pkg/sharkfin/messages.go
Implement all protocol types: Envelope, request/response structs, marshal/unmarshal helpers.

### 2. pkg/sharkfin/client.go
- `New(host string) *Client`
- `Connect() error` — dial WebSocket, read hello, store heartbeat interval
- `Identify(username string) error` — send identify, fall back to register
- `ReadPump(p *tea.Program)` — read loop, parse envelope, dispatch typed tea.Msg via p.Send()
- `WritePump()` — write loop from outbound channel, single writer
- `Send(msgType string, data any)` — marshal envelope, push to outbound
- `Close()` — signal done, close conn
- Convenience methods:
  - `RequestChannels()`
  - `RequestHistory(channel, before, limit)` — paginated history for scrollback
  - `RequestUnread(channel)` — fetch unread messages AND advance read cursor
  - `RequestUsers()`
  - `SendMessage(channel, body)`
  - `CreateChannel(name, public, members)`
  - `InviteUser(channel, username)`

### 3. internal/chat/msgtypes.go
Update message types to carry real sharkfin data (ChannelListMsg, HistoryMsg, UnreadMsg, MessageNewMsg, PresenceMsg, DisconnectedMsg, etc.)

### 4. internal/chat/model.go — connect to client
- Accept `*sharkfin.Client` in constructor
- `Init()` returns tea.Cmd that requests channel_list
- Update handlers for all inbound message types
- Channel switch calls `RequestUnread(channel)` to fetch new messages and advance server read cursor
- Scrollback triggers `RequestHistory(channel, before, limit)` for older messages
- Send message writes to client via `SendMessage()`
- Channel create modal calls `CreateChannel()`
- Invite modal calls `InviteUser()`
- DisconnectedMsg shows disconnect modal

### 5. internal/chat/channels.go — live data
- Replace hardcoded channels with data from ChannelListMsg
- Unread count tracking:
  - Increment on `message.new` for non-active channels (local counter)
  - Clear local counter on channel switch
  - On channel switch, `RequestUnread()` advances server cursor
- Channel create modal triggers `CreateChannel()`

### 6. internal/chat/messages.go — live data
- Replace hardcoded messages with data from UnreadMsg/HistoryMsg
- Append from MessageNewMsg in real time
- Auto-scroll behavior (only if at bottom)
- Prepend older messages from HistoryMsg on scrollback

### 7. internal/chat/input.go — send messages
- On Enter, send via client `SendMessage()`, clear input
- MessageSentMsg confirms delivery

### 8. cmd/chat/chat.go — full wiring
- Create client, connect, identify (auto-register on first use)
- Pass client to model
- Launch ReadPump/WritePump goroutines
- defer client.Close()
- Pre-connection errors: print and exit

### 9. pkg/config/viper.go — new config keys
- Add defaults for sharkfin-host and username
- Add to BindFlags list

### 10. Verify
- mise run build + lint
- Launch against running sharkfin daemon
- Send/receive messages in real time
- Channel switching fetches unreads and advances read cursor
- Scrollback loads history
- New messages appear live with unread counts on inactive channels
- Presence updates show in user list
- Channel create and invite modals work end-to-end
- Disconnect modal works on daemon stop
- Clean shutdown on Ctrl+q
