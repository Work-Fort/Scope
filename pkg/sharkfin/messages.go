package sharkfin

import (
	"encoding/json"
	"time"
)

// Envelope is the top-level JSON structure for all sharkfin WS messages.
type Envelope struct {
	Type string          `json:"type"`
	D    json.RawMessage `json:"d,omitempty"`
	Ref  string          `json:"ref,omitempty"`
	OK   *bool           `json:"ok,omitempty"`
}

// Hello is sent by the server on connection.
type Hello struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// IdentifyRequest is sent to authenticate as an existing user.
type IdentifyRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest is sent to create a new user.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Channel represents a chat channel from the server.
type Channel struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
	Member bool   `json:"member"`
}

// ChannelListResponse is the payload of a channel_list response.
type ChannelListResponse struct {
	Channels []Channel `json:"channels"`
}

// ChannelCreateRequest creates a new channel.
type ChannelCreateRequest struct {
	Name    string   `json:"name"`
	Public  bool     `json:"public"`
	Members []string `json:"members,omitempty"`
}

// ChannelInviteRequest invites a user to a channel.
type ChannelInviteRequest struct {
	Channel  string `json:"channel"`
	Username string `json:"username"`
}

// Message represents a chat message from the server.
type Message struct {
	ID       int       `json:"id"`
	Channel  string    `json:"channel"`
	From     string    `json:"from"`
	Body     string    `json:"body"`
	SentAt   time.Time `json:"sent_at"`
	ThreadID *int      `json:"thread_id,omitempty"`
}

// HistoryRequest requests message history for a channel.
type HistoryRequest struct {
	Channel string `json:"channel"`
	Before  int    `json:"before,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// HistoryResponse is the payload of a history response.
type HistoryResponse struct {
	Channel  string    `json:"channel"`
	Messages []Message `json:"messages"`
}

// UnreadRequest requests unread messages.
type UnreadRequest struct {
	Channel      string `json:"channel,omitempty"`
	MentionsOnly bool   `json:"mentions_only,omitempty"`
	ThreadID     *int   `json:"thread_id,omitempty"`
}

// UnreadResponse is the payload of an unread_messages response.
type UnreadResponse struct {
	Channel  string    `json:"channel,omitempty"`
	Messages []Message `json:"messages"`
}

// SendMessageRequest sends a message to a channel.
type SendMessageRequest struct {
	Channel  string `json:"channel"`
	Body     string `json:"body"`
	ThreadID *int   `json:"thread_id,omitempty"`
}

// MessageNewEvent is a server push for new messages.
type MessageNewEvent struct {
	ID       int       `json:"id"`
	Channel  string    `json:"channel"`
	From     string    `json:"from"`
	Body     string    `json:"body"`
	SentAt   time.Time `json:"sent_at"`
	ThreadID *int      `json:"thread_id,omitempty"`
}

// User represents a user from the server.
type User struct {
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

// UserListResponse is the payload of a user_list response.
type UserListResponse struct {
	Users []User `json:"users"`
}

// PresenceEvent is a server push for user presence changes.
type PresenceEvent struct {
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

// MarshalEnvelope creates an Envelope with the given type and data payload.
func MarshalEnvelope(msgType string, data any, ref string) ([]byte, error) {
	var d json.RawMessage
	if data != nil {
		var err error
		d, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}
	env := Envelope{
		Type: msgType,
		D:    d,
		Ref:  ref,
	}
	return json.Marshal(env)
}

// ParseEnvelope parses a raw JSON message into an Envelope.
func ParseEnvelope(raw []byte) (Envelope, error) {
	var env Envelope
	err := json.Unmarshal(raw, &env)
	return env, err
}

// ParseData unmarshals the D field of an envelope into the given target.
func ParseData[T any](env Envelope) (T, error) {
	var t T
	err := json.Unmarshal(env.D, &t)
	return t, err
}
