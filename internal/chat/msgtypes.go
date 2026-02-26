package chat

// ChannelInfo represents a chat channel.
type ChannelInfo struct {
	Name   string
	Public bool
}

// MessageInfo represents a single chat message.
type MessageInfo struct {
	ID     int
	From   string
	Body   string
	SentAt string
}

// Inbound messages (from sharkfin in Plan 2, mocked for now)

type ConnectedMsg struct{}

type DisconnectedMsg struct {
	Err error
}

type ChannelListMsg struct {
	Channels []ChannelInfo
}

type HistoryMsg struct {
	Channel  string
	Messages []MessageInfo
}

type MessageNewMsg struct {
	ID      int
	Channel string
	From    string
	Body    string
	SentAt  string
}

type PresenceMsg struct {
	Username string
	Online   bool
}

// Internal UI messages

type ChannelSelectedMsg struct {
	Name string
}

type MessageSentMsg struct{}

type ModalCloseMsg struct{}
