package sharkfin

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
)

// Client manages a WebSocket connection to a sharkfin daemon.
type Client struct {
	host        string
	conn        *websocket.Conn
	outbound    chan []byte
	done        chan struct{}
	closed      atomic.Bool
	refSeq      atomic.Int64
	mu          sync.Mutex
	pendingRefs sync.Map // ref string → request type string
}

// New creates a new sharkfin client for the given WebSocket URL.
func New(host string) *Client {
	return &Client{
		host:     host,
		outbound: make(chan []byte, 64),
		done:     make(chan struct{}),
	}
}

// Connect dials the sharkfin WebSocket and reads the hello message.
func (c *Client) Connect() (*Hello, error) {
	conn, _, err := websocket.DefaultDialer.Dial(c.host, nil)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", c.host, err)
	}
	c.conn = conn

	// Read hello
	_, raw, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("read hello: %w", err)
	}

	env, err := ParseEnvelope(raw)
	if err != nil || env.Type != "hello" {
		conn.Close()
		return nil, fmt.Errorf("expected hello, got: %s", string(raw))
	}

	hello, err := ParseData[Hello](env)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("parse hello: %w", err)
	}

	log.Debug("connected to sharkfin", "host", c.host, "heartbeat_interval", hello.HeartbeatInterval)
	return &hello, nil
}

// Identify authenticates as an existing user, falling back to register.
// Writes directly to the connection (WritePump is not yet running during handshake).
func (c *Client) Identify(username string) error {
	ref := c.nextRef()
	data, _ := MarshalEnvelope("identify", IdentifyRequest{
		Username: username,
		Password: "",
	}, ref)
	if err := c.writeDirect(data); err != nil {
		return fmt.Errorf("write identify: %w", err)
	}

	env, err := c.readResponse()
	if err != nil {
		return err
	}

	if env.OK != nil && *env.OK {
		log.Debug("identified", "username", username)
		return nil
	}

	// Fall back to register
	ref = c.nextRef()
	data, _ = MarshalEnvelope("register", RegisterRequest{
		Username: username,
		Password: "",
	}, ref)
	if err := c.writeDirect(data); err != nil {
		return fmt.Errorf("write register: %w", err)
	}

	env, err = c.readResponse()
	if err != nil {
		return err
	}

	if env.OK != nil && *env.OK {
		log.Debug("registered", "username", username)
		return nil
	}

	return fmt.Errorf("identify/register failed for %s", username)
}

// ReadPump reads messages from the WebSocket and dispatches them as tea.Msg via p.Send().
func (c *Client) ReadPump(p *tea.Program) {
	defer func() {
		c.Close()
		p.Send(DisconnectedMsg{})
	}()

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if !c.closed.Load() {
				log.Error("read error", "err", err)
				p.Send(DisconnectedMsg{Err: err})
			}
			return
		}

		env, err := ParseEnvelope(raw)
		if err != nil {
			log.Error("parse envelope", "err", err, "raw", string(raw))
			continue
		}

		log.Debug("ws_recv", "type", env.Type, "ref", env.Ref)
		msg := c.dispatchMessage(env)
		if msg != nil {
			p.Send(msg)
		}
	}
}

// WritePump writes outbound messages to the WebSocket. Single writer goroutine.
func (c *Client) WritePump() {
	for {
		select {
		case data := <-c.outbound:
			c.mu.Lock()
			err := c.conn.WriteMessage(websocket.TextMessage, data)
			c.mu.Unlock()
			if err != nil {
				if !c.closed.Load() {
					log.Error("write error", "err", err)
				}
				return
			}
		case <-c.done:
			return
		}
	}
}

// Close shuts down the client.
func (c *Client) Close() {
	if c.closed.CompareAndSwap(false, true) {
		close(c.done)
		if c.conn != nil {
			c.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			c.conn.Close()
		}
	}
}

// RequestChannels requests the channel list.
func (c *Client) RequestChannels() {
	c.send("channel_list", nil)
}

// RequestHistory requests paginated message history.
func (c *Client) RequestHistory(channel string, before, limit int) {
	c.send("history", HistoryRequest{
		Channel: channel,
		Before:  before,
		Limit:   limit,
	})
}

// RequestUnread requests unread messages for a channel (advances read cursor).
func (c *Client) RequestUnread(channel string) {
	c.send("unread_messages", UnreadRequest{
		Channel: channel,
	})
}

// RequestUsers requests the user list.
func (c *Client) RequestUsers() {
	c.send("user_list", nil)
}

// SendMessage sends a chat message to a channel.
func (c *Client) SendMessage(channel, body string) {
	c.send("send_message", SendMessageRequest{
		Channel: channel,
		Body:    body,
	})
}

// CreateChannel creates a new channel.
func (c *Client) CreateChannel(name string, public bool, members []string) {
	c.send("channel_create", ChannelCreateRequest{
		Name:    name,
		Public:  public,
		Members: members,
	})
}

// InviteUser invites a user to a channel.
func (c *Client) InviteUser(channel, username string) {
	c.send("channel_invite", ChannelInviteRequest{
		Channel:  channel,
		Username: username,
	})
}

// RequestUnreadCounts requests per-channel unread and mention counts.
func (c *Client) RequestUnreadCounts() {
	c.send("unread_counts", nil)
}

// RequestDMList requests the DM conversation list.
func (c *Client) RequestDMList() {
	c.send("dm_list", nil)
}

// DMOpen opens or creates a DM with the given user.
func (c *Client) DMOpen(username string) {
	c.send("dm_open", DMOpenRequest{
		Username: username,
	})
}

// MarkRead advances the read cursor for a channel.
func (c *Client) MarkRead(channel string, messageID *int) {
	c.send("mark_read", MarkReadRequest{
		Channel:   channel,
		MessageID: messageID,
	})
}

func (c *Client) send(msgType string, data any) {
	ref := c.nextRef()
	c.pendingRefs.Store(ref, msgType)
	raw, err := MarshalEnvelope(msgType, data, ref)
	if err != nil {
		log.Error("marshal", "type", msgType, "err", err)
		c.pendingRefs.Delete(ref)
		return
	}
	c.sendRaw(raw)
}

// writeDirect writes to the connection without going through the outbound channel.
// Used during the handshake phase before WritePump is running.
func (c *Client) writeDirect(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) sendRaw(data []byte) {
	select {
	case c.outbound <- data:
	case <-c.done:
	}
}

func (c *Client) nextRef() string {
	return fmt.Sprintf("ref_%d", c.refSeq.Add(1))
}

// readResponse reads a single response from the WebSocket (used during handshake).
func (c *Client) readResponse() (Envelope, error) {
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer c.conn.SetReadDeadline(time.Time{})

	_, raw, err := c.conn.ReadMessage()
	if err != nil {
		return Envelope{}, fmt.Errorf("read response: %w", err)
	}
	return ParseEnvelope(raw)
}

// dispatchMessage converts an inbound envelope to a typed tea.Msg.
func (c *Client) dispatchMessage(env Envelope) tea.Msg {
	switch env.Type {
	case "reply":
		return c.dispatchReply(env)

	case "message.new":
		msg, err := ParseData[MessageNewEvent](env)
		if err != nil {
			log.Error("parse message.new", "err", err)
			return nil
		}
		return MessageNewMsg(msg)

	case "presence":
		evt, err := ParseData[PresenceEvent](env)
		if err != nil {
			log.Error("parse presence", "err", err)
			return nil
		}
		return PresenceMsg(evt)

	case "heartbeat":
		c.send("heartbeat", nil)
		return nil

	default:
		log.Debug("unhandled message type", "type", env.Type, "raw", string(env.D))
		return nil
	}
}

// dispatchReply routes a "reply" envelope to the correct handler based on the original request type.
func (c *Client) dispatchReply(env Envelope) tea.Msg {
	val, ok := c.pendingRefs.LoadAndDelete(env.Ref)
	if !ok {
		log.Debug("reply with unknown ref", "ref", env.Ref)
		return nil
	}
	reqType := val.(string)

	// Check for error replies (ok: false) on data-bearing requests
	if env.OK != nil && !*env.OK {
		log.Debug("reply error", "reqType", reqType, "ref", env.Ref, "data", string(env.D))
		return nil
	}

	switch reqType {
	case "channel_list":
		resp, err := ParseData[ChannelListResponse](env)
		if err != nil {
			log.Error("parse channel_list reply", "err", err)
			return nil
		}
		return ChannelListMsg{Channels: resp.Channels}

	case "history":
		resp, err := ParseData[HistoryResponse](env)
		if err != nil {
			log.Error("parse history reply", "err", err)
			return nil
		}
		return HistoryMsg{Channel: resp.Channel, Messages: resp.Messages}

	case "unread_messages":
		resp, err := ParseData[UnreadResponse](env)
		if err != nil {
			log.Error("parse unread_messages reply", "err", err)
			return nil
		}
		ch := resp.Channel
		if ch == "" && len(resp.Messages) > 0 {
			ch = resp.Messages[0].Channel
		}
		return UnreadMsg{Channel: ch, Messages: resp.Messages}

	case "user_list":
		resp, err := ParseData[UserListResponse](env)
		if err != nil {
			log.Error("parse user_list reply", "err", err)
			return nil
		}
		return UserListMsg{Users: resp.Users}

	case "send_message":
		return MessageSentMsg{}

	case "channel_create":
		if env.OK != nil && *env.OK {
			c.RequestChannels()
		}
		return nil

	case "channel_invite":
		return nil

	case "unread_counts":
		resp, err := ParseData[UnreadCountsResponse](env)
		if err != nil {
			log.Error("parse unread_counts reply", "err", err)
			return nil
		}
		return UnreadCountsMsg{Counts: resp.Counts}

	case "mark_read":
		return nil

	case "dm_list":
		resp, err := ParseData[DMListResponse](env)
		if err != nil {
			log.Error("parse dm_list reply", "err", err)
			return nil
		}
		return DMListMsg{DMs: resp.DMs}

	case "dm_open":
		resp, err := ParseData[DMOpenResponse](env)
		if err != nil {
			log.Error("parse dm_open reply", "err", err)
			return nil
		}
		return DMOpenMsg{Channel: resp.Channel, Participant: resp.Participant, Created: resp.Created}

	default:
		log.Debug("unhandled reply type", "reqType", reqType, "ref", env.Ref)
		return nil
	}
}

// Tea messages dispatched by the client.

type ConnectedMsg struct{}

type DisconnectedMsg struct {
	Err error
}

type ChannelListMsg struct {
	Channels []Channel
}

type HistoryMsg struct {
	Channel  string
	Messages []Message
}

type UnreadMsg struct {
	Channel  string
	Messages []Message
}

type MessageNewMsg = MessageNewEvent

type PresenceMsg = PresenceEvent

type UserListMsg struct {
	Users []User
}

type MessageSentMsg struct{}

type UnreadCountsMsg struct {
	Counts []ChannelUnreadCount
}

type DMListMsg struct {
	DMs []DM
}

type DMOpenMsg struct {
	Channel     string
	Participant string
	Created     bool
}
