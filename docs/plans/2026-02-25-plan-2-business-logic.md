# Plan 2: WebSocket Client & Business Logic

Wire up the sharkfin WebSocket client, replace all placeholder data with live data, and make the chat fully functional.

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
- `RequestChannels()`, `RequestHistory(channel, before, limit)`, `RequestUsers()` — convenience methods

### 3. internal/chat/msgtypes.go
Update message types to carry real sharkfin data (ChannelListMsg, HistoryMsg, MessageNewMsg, PresenceMsg, DisconnectedMsg, etc.)

### 4. internal/chat/model.go — connect to client
- Accept `*sharkfin.Client` in constructor
- `Init()` returns tea.Cmd that requests channel_list
- Update handlers for all inbound message types
- Channel switch triggers history request
- Send message writes to client outbound
- DisconnectedMsg shows disconnect modal

### 5. internal/chat/channels.go — live data
- Replace hardcoded channels with data from ChannelListMsg
- Unread indicators from MessageNewMsg on non-active channels
- Channel create modal triggers client.Send("channel_create", ...)

### 6. internal/chat/messages.go — live data
- Replace hardcoded messages with HistoryMsg data
- Append from MessageNewMsg in real time
- Auto-scroll behavior (only if at bottom)

### 7. internal/chat/input.go — send messages
- On Enter, send via client, clear input
- MessageSentMsg confirms delivery

### 8. cmd/chat/chat.go — full wiring
- Create client, connect, identify
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
- Channel switching loads history
- New messages appear live
- Presence updates show in user list
- Disconnect modal works on daemon stop
- Clean shutdown on Ctrl+q
