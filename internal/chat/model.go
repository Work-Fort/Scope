package chat

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/spf13/viper"

	"github.com/Work-Fort/WorkFort/pkg/audio"
	"github.com/Work-Fort/WorkFort/pkg/config"
	"github.com/Work-Fort/WorkFort/pkg/sharkfin"
	"github.com/Work-Fort/WorkFort/pkg/ui"
)

// Pane identifies which panel has focus.
type Pane int

const (
	PaneChannels Pane = iota
	PaneInput
)

const inputBtnW = 15 // width reserved for two bordered action buttons beside input

// SidebarTab identifies which sidebar list is visible.
type SidebarTab int

const (
	TabChannels SidebarTab = iota
	TabDMs
	TabUsers
)

// ChatModel is the root Bubble Tea model for the chat UI.
type ChatModel struct {
	width    int
	height   int
	tooSmall bool

	client   *sharkfin.Client
	channels ChannelList
	dmList   DMList
	userList UserList
	messages MessagePane
	input    InputBar
	modal    *Modal

	activePane      Pane
	sidebarTab      SidebarTab
	selectedChannel string
	username        string
	users           []sharkfin.User

	lastChannelScroll time.Time
	lastMsgScroll     time.Time
	loadingHistory    bool
	historyExhausted  map[string]bool // channels that have no more history to load

	customSidebarW int  // user-dragged sidebar width (0 = auto)
	sidebarHidden  bool // true when sidebar is toggled off (Ctrl+R)
	dragging       bool // true while dragging the divider
	lastMouseY    int       // last known mouse Y position from tea.MouseMsg
	lastMouseTime time.Time // when last mouse event was received

	notifSound     audio.Sound // current notification sound
	pendingJoinCh  string      // channel name to select after join completes
}

// NewModel creates a ChatModel wired to a sharkfin client.
func NewModel(client *sharkfin.Client, username string) ChatModel {
	cl := NewChannelList()
	dl := NewDMList(username)
	ul := NewUserList(username)
	mp := NewMessagePane()
	ib := NewInputBar()

	return ChatModel{
		client:           client,
		channels:         cl,
		dmList:           dl,
		userList:         ul,
		messages:         mp,
		input:            ib,
		activePane:       PaneChannels,
		username:         username,
		historyExhausted: make(map[string]bool),
		notifSound:       audio.Sound(viper.GetString("notification-sound")),
	}
}

func (m ChatModel) Init() tea.Cmd {
	m.client.RequestChannels()
	m.client.RequestDMList()
	m.client.RequestUsers()
	return nil
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tooSmall = msg.Width < ui.MinWidth || msg.Height < ui.MinHeight
		if !m.tooSmall {
			m.updateLayout()
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	// --- Sharkfin inbound messages ---

	case sharkfin.ChannelJoinMsg:
		// Join succeeded — channel list refresh is already requested by the client.
		// pendingJoinCh will be selected when the refreshed list arrives.
		return m, nil

	case sharkfin.ChannelListMsg:
		log.Debug("channel_list", "count", len(msg.Channels))
		m.channels.SetChannels(msg.Channels)
		m.client.RequestUnreadCounts()
		if m.pendingJoinCh != "" {
			if m.channels.SelectByName(m.pendingJoinCh) {
				m.switchChannel()
				m.activePane = PaneInput
				m.input.Focus()
			}
			m.pendingJoinCh = ""
		} else if len(msg.Channels) > 0 && m.selectedChannel == "" {
			m.selectedChannel = msg.Channels[0].Name
			m.messages.SetChannel(m.selectedChannel)
			m.client.RequestHistory(m.selectedChannel, 0, 50)
			m.client.RequestUnread(m.selectedChannel)
		}
		return m, nil

	case sharkfin.UnreadCountsMsg:
		log.Debug("unread_counts", "count", len(msg.Counts))
		for _, c := range msg.Counts {
			if c.Channel == m.selectedChannel {
				continue
			}
			if c.ChannelType == "dm" {
				m.dmList.SetCounts(c.Channel, c.UnreadCount, c.MentionCount)
				m.userList.SetDMUnread(c.Channel, c.UnreadCount)
			} else {
				m.channels.SetCounts(c.Channel, c.UnreadCount, c.MentionCount)
			}
		}
		return m, nil

	case sharkfin.HistoryMsg:
		log.Debug("history", "channel", msg.Channel, "count", len(msg.Messages), "scrollback", m.loadingHistory)
		if m.loadingHistory {
			if len(msg.Messages) == 0 {
				// No older messages — channel history is fully loaded
				m.historyExhausted[msg.Channel] = true
			} else {
				m.messages.PrependHistory(msg.Channel, msg.Messages)
			}
		} else {
			// Initial load: merge and auto-scroll to bottom
			m.messages.MergeMessages(msg.Channel, msg.Messages)
		}
		m.loadingHistory = false
		return m, nil

	case sharkfin.UnreadMsg:
		log.Debug("unread", "channel", msg.Channel, "count", len(msg.Messages))
		if len(msg.Messages) > 0 {
			m.messages.MergeMessages(msg.Channel, msg.Messages)
		}
		return m, nil

	case sharkfin.MessageNewMsg:
		log.Debug("message.new", "channel", msg.Channel, "from", msg.From, "body", msg.Body, "active", m.selectedChannel)
		// Skip messages from ourselves — we already appended locally on send
		if msg.From == m.username {
			return m, nil
		}
		newMsg := sharkfin.Message{
			ID:       msg.ID,
			Channel:  msg.Channel,
			From:     msg.From,
			Body:     msg.Body,
			SentAt:   msg.SentAt,
			ThreadID: msg.ThreadID,
		}
		m.messages.AppendMessage(msg.Channel, newMsg)
		if msg.Channel != m.selectedChannel {
			audio.Play(m.notifSound)
			isMention := containsUser(msg.Mentions, m.username)
			if msg.ChannelType == "dm" {
				if isMention {
					m.dmList.IncrementMention(msg.Channel)
				} else {
					m.dmList.IncrementUnread(msg.Channel)
				}
				m.userList.SetDMUnread(msg.Channel, m.dmList.unread[msg.Channel])
			} else {
				if isMention {
					m.channels.IncrementMention(msg.Channel)
				} else {
					m.channels.IncrementUnread(msg.Channel)
				}
			}
		}
		return m, nil

	case sharkfin.PresenceMsg:
		log.Debug("presence", "user", msg.Username, "online", msg.Online)
		for i := range m.users {
			if m.users[i].Username == msg.Username {
				m.users[i].Online = msg.Online
				break
			}
		}
		m.userList.UpdatePresence(msg.Username, msg.Online)
		return m, nil

	case sharkfin.UserListMsg:
		m.users = msg.Users
		m.userList.SetUsers(msg.Users)
		return m, nil

	case sharkfin.MessageSentMsg:
		return m, nil

	case sharkfin.DMListMsg:
		log.Debug("dm_list", "count", len(msg.DMs))
		m.dmList.SetDMs(msg.DMs)
		m.userList.SetDMMapping(msg.DMs, m.username)
		if m.sidebarTab == TabDMs && m.selectedChannel != "" {
			m.dmList.SelectByChannel(m.selectedChannel)
		}
		return m, nil

	case sharkfin.DMOpenMsg:
		log.Debug("dm_open", "channel", msg.Channel, "participant", msg.Participant, "created", msg.Created)
		if msg.Created {
			m.client.RequestDMList()
		}
		m.selectedChannel = msg.Channel
		// Sync both DM-aware lists without switching tabs
		m.dmList.SelectByChannel(msg.Channel)
		m.dmList.ClearUnread(msg.Channel)
		m.userList.ClearDMUnread(msg.Channel)
		m.messages.SetChannel(msg.Channel)
		m.client.RequestHistory(msg.Channel, 0, 50)
		m.input.SetReadOnly(false)
		m.activePane = PaneInput
		m.input.Focus()
		return m, nil

	case sharkfin.DisconnectedMsg:
		// TODO: show disconnect overlay
		return m, nil
	}

	return m, nil
}

func (m ChatModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Modal takes priority if open
	if m.modal != nil {
		return m.handleModalKey(msg)
	}

	switch key {
	case "ctrl+q":
		return m, tea.Quit

	case "ctrl+,":
		modal := NewSettingsModal(m.notifSound)
		m.modal = &modal
		m.input.Blur()
		return m, nil

	case "ctrl+r":
		m.sidebarHidden = !m.sidebarHidden
		m.updateLayout()
		return m, nil

	case "ctrl+a":
		switch m.sidebarTab {
		case TabChannels:
			m.sidebarTab = TabDMs
		case TabDMs:
			m.sidebarTab = TabUsers
		default:
			m.sidebarTab = TabChannels
		}
		m.switchChannel()
		return m, nil

	case "ctrl+d":
		modal := NewModal(ModalDMOpen)
		m.modal = &modal
		m.input.Blur()
		return m, nil

	case "ctrl+s":
		if m.modal != nil && m.modal.Type == ModalShortcuts {
			m.modal = nil
			m.input.Focus()
		} else {
			modal := NewModal(ModalShortcuts)
			m.modal = &modal
			m.input.Blur()
		}
		return m, nil

	case "ctrl+j":
		switch m.sidebarTab {
		case TabChannels:
			m.channels.MoveDown()
		case TabDMs:
			m.dmList.MoveDown()
		case TabUsers:
			m.userList.MoveDown()
		}
		m.switchChannel()
		return m, nil

	case "ctrl+k":
		switch m.sidebarTab {
		case TabChannels:
			m.channels.MoveUp()
		case TabDMs:
			m.dmList.MoveUp()
		case TabUsers:
			m.userList.MoveUp()
		}
		m.switchChannel()
		return m, nil

	case "ctrl+l":
		canWrite := m.sidebarTab == TabDMs || m.sidebarTab == TabUsers || m.channels.IsMember()
		if canWrite {
			m.activePane = PaneInput
			m.input.Focus()
		}
		return m, nil

	case "ctrl+o":
		m.activePane = PaneChannels
		m.input.Blur()
		return m, nil

	case "ctrl+n":
		modal := NewModal(ModalChannelCreate)
		m.modal = &modal
		m.input.Blur()
		return m, nil

	case "ctrl+u":
		modal := NewModal(ModalUserInvite)
		m.modal = &modal
		m.input.Blur()
		return m, nil

	case "ctrl+up":
		m.messages.ScrollUp(1)
		m.maybeLoadHistory()
		return m, nil

	case "ctrl+down":
		m.messages.ScrollDown(1)
		return m, nil

	case "tab":
		if m.activePane == PaneInput {
			m.input.TryComplete(m.usernameList())
			return m, nil
		}

	case "alt+enter":
		if m.activePane == PaneInput {
			m.input.InsertNewline()
			m.updateLayout()
			return m, nil
		}

	case "ctrl+v":
		if m.activePane == PaneInput {
			m.input.Paste()
			m.updateLayout()
			return m, nil
		}

	case "enter":
		if m.activePane == PaneInput && m.trySendMessage() {
			return m, nil
		}
		if m.activePane == PaneChannels {
			m.switchChannel()
			m.activePane = PaneInput
			m.input.Focus()
			return m, nil
		}
	}

	// Pass remaining keys to focused component, filtering leaked SGR mouse sequences.
	// Bubble Tea's parser splits SGR mouse events across reads: "[" arrives as one
	// KeyMsg, then "<64;87;52M" as another. We detect these fragments and also
	// suppress lone "[" characters when the mouse is not over the input bar.
	leaked := isLeakedMouseSeq(msg)
	if !leaked && key == "[" {
		// "[" is ambiguous — it's both a valid character and an SGR sequence opener.
		// Suppress it when a mouse event was received recently, since rapid
		// scrolling splits SGR sequences and the "[" is almost certainly a fragment.
		if time.Since(m.lastMouseTime) < 100*time.Millisecond {
			leaked = true
		}
	}
	if m.activePane == PaneInput && !leaked {
		prevH := m.input.Height()
		m.input.UpdateTextInput(msg)
		if m.input.Height() != prevH {
			m.updateLayout()
		}
	}

	return m, nil
}

func (m ChatModel) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m.modal = nil
		m.input.Focus()
		return m, nil

	case "ctrl+s", "ctrl+d", "ctrl+n", "ctrl+u":
		toggleType := map[string]ModalType{
			"ctrl+s": ModalShortcuts,
			"ctrl+d": ModalDMOpen,
			"ctrl+n": ModalChannelCreate,
			"ctrl+u": ModalUserInvite,
		}[key]
		if m.modal.Type == toggleType {
			m.modal = nil
			m.input.Focus()
			return m, nil
		}

	case "enter":
		if m.modal.Type == ModalSettings {
			m.submitModal()
			m.modal = nil
			m.input.Focus()
			return m, nil
		}
		if m.modal.Value() != "" {
			m.submitModal()
			m.modal = nil
		}
		return m, nil

	case "up", "k":
		if m.modal.Type == ModalSettings {
			m.modal.SoundCursorUp()
			audio.Play(m.modal.SelectedSound())
			return m, nil
		}

	case "down", "j":
		if m.modal.Type == ModalSettings {
			m.modal.SoundCursorDown()
			audio.Play(m.modal.SelectedSound())
			return m, nil
		}

	case "tab":
		if m.modal.Type == ModalChannelCreate {
			m.tryChannelModalComplete()
		} else if m.modal.Type == ModalDMOpen || m.modal.Type == ModalUserInvite {
			m.tryModalComplete()
		}
		return m, nil
	}

	if m.modal.Type != ModalShortcuts && m.modal.Type != ModalSettings {
		m.modal.UpdateTextInput(msg)
		// Clear hint when typing in channel modal
		if m.modal.Type == ModalChannelCreate {
			m.modal.hint = ""
		}
	}
	return m, nil
}

func (m *ChatModel) submitModal() {
	switch m.modal.Type {
	case ModalChannelCreate:
		name := m.modal.Value()
		if ch, found := m.channels.FindByName(name); found {
			if ch.Member {
				// Already a member — switch to it
				m.channels.SelectByName(name)
				m.switchChannel()
				m.activePane = PaneInput
				m.input.Focus()
			} else {
				// Not a member — join it
				m.pendingJoinCh = name
				m.client.JoinChannel(name)
			}
		} else {
			// Channel doesn't exist — create it (public by default)
			m.client.CreateChannel(name, true, nil)
		}
	case ModalUserInvite:
		m.client.InviteUser(m.selectedChannel, m.modal.Value())
	case ModalDMOpen:
		m.client.DMOpen(m.modal.Value())
	case ModalSettings:
		m.notifSound = m.modal.SelectedSound()
		if err := config.SaveSetting("notification-sound", string(m.notifSound)); err != nil {
			log.Error("failed to save notification sound setting", "err", err)
		}
	}
}

func (m ChatModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.lastMouseY = msg.Y
	m.lastMouseTime = time.Now()

	// Modal click handling
	if m.modal != nil {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			// Sound selector click
			if idx := m.modal.SoundHitTest(msg.X, msg.Y); idx >= 0 {
				m.modal.soundCursor = idx
				audio.Play(m.modal.SelectedSound())
				return m, nil
			}
			action := m.modal.HitTest(msg.X, msg.Y)
			switch action {
			case ModalActionSubmit:
				if m.modal.Type == ModalSettings || m.modal.Value() != "" {
					m.submitModal()
					m.modal = nil
					m.input.Focus()
				}
			case ModalActionToggle:
				if m.modal.Type == ModalChannelCreate {
					m.modal.TogglePublic()
				}
			case ModalActionCancel:
				m.modal = nil
				m.input.Focus()
			}
		}
		return m, nil
	}

	layout := m.layout()

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if msg.X < layout.SidebarW {
			if time.Since(m.lastChannelScroll) < 80*time.Millisecond {
				return m, nil
			}
			m.lastChannelScroll = time.Now()
			switch m.sidebarTab {
			case TabChannels:
				m.channels.MoveUp()
			case TabDMs:
				m.dmList.MoveUp()
			case TabUsers:
				m.userList.MoveUp()
			}
			m.switchChannel()
		} else {
			if layout.Skinny {
				if time.Since(m.lastMsgScroll) < 80*time.Millisecond {
					return m, nil
				}
				m.lastMsgScroll = time.Now()
			}
			m.messages.ScrollUp(3)
			m.maybeLoadHistory()
		}
		return m, nil

	case tea.MouseButtonWheelDown:
		if msg.X < layout.SidebarW {
			if time.Since(m.lastChannelScroll) < 80*time.Millisecond {
				return m, nil
			}
			m.lastChannelScroll = time.Now()
			switch m.sidebarTab {
			case TabChannels:
				m.channels.MoveDown()
			case TabDMs:
				m.dmList.MoveDown()
			case TabUsers:
				m.userList.MoveDown()
			}
			m.switchChannel()
		} else {
			if layout.Skinny {
				if time.Since(m.lastMsgScroll) < 80*time.Millisecond {
					return m, nil
				}
				m.lastMsgScroll = time.Now()
			}
			m.messages.ScrollDown(3)
		}
		return m, nil

	case tea.MouseButtonLeft:
		switch msg.Action {
		case tea.MouseActionPress:
			// Check if pressing on the divider gap (±1 char hit zone)
			if !layout.Skinny && !m.sidebarHidden {
				gap := layout.SidebarW
				if msg.X >= gap-1 && msg.X <= gap+1 {
					m.dragging = true
					return m, nil
				}
			}

			contentTop := ui.HeaderHeight
			contentBottom := contentTop + layout.ContentH

			// Help bar spans full width — check Y first
			if msg.Y >= contentBottom {
				return m.handleHelpBarClick(msg.X)
			}

			// Click in header — ignore
			if msg.Y < contentTop {
				return m, nil
			}

			if msg.X < layout.SidebarW {
				// Tab bar buttons span 3 rows inside the sidebar border
				tabBarTop := contentTop + 1 // after sidebar border
				tabBarBottom := tabBarTop + 2
				if msg.Y >= tabBarTop && msg.Y <= tabBarBottom {
					// Click on tab bar — split into thirds
					third := layout.SidebarW / 3
					if msg.X < third {
						m.sidebarTab = TabChannels
					} else if msg.X < third*2 {
						m.sidebarTab = TabDMs
					} else {
						m.sidebarTab = TabUsers
					}
					m.switchChannel()
					return m, nil
				}

				// Click in sidebar list: select item from active tab
				// Account for border (1) + tab bar (3)
				row := msg.Y - contentTop - 1 - 3
				if row >= 0 {
					switch m.sidebarTab {
					case TabChannels:
						m.channels.SelectIndex(row)
					case TabDMs:
						m.dmList.SelectIndex(row)
					case TabUsers:
						m.userList.SelectIndex(row)
					}
					m.switchChannel()
				}
				return m, nil
			}

			// Check for action button clicks (bordered mic/send beside input)
			rightPaneX := layout.SidebarW + ui.PaneGap
			if layout.Skinny || m.sidebarHidden {
				rightPaneX = 0
			}
			btnAreaX := rightPaneX + layout.MessageW - inputBtnW
			inputBottom := contentTop + layout.ContentH
			// Buttons are 3 rows tall (border+content+border), bottom-aligned
			btnTop := inputBottom - 3
			if msg.X >= btnAreaX && msg.Y >= btnTop && msg.Y < inputBottom {
				// Each bordered button is 7 wide (border+Width(5)+border)
				// 1 space gap before buttons, so send starts at offset 1+7=8
				if msg.X >= btnAreaX+8 {
					m.trySendMessage()
				}
				// Mic button (left) — unwired
				return m, nil
			}

			// Click in right pane — focus input if writable
			canWrite := m.sidebarTab == TabDMs || m.sidebarTab == TabUsers || m.channels.IsMember()
			if canWrite {
				m.activePane = PaneInput
				m.input.Focus()
			}
			return m, nil

		case tea.MouseActionMotion:
			if m.dragging {
				m.customSidebarW = msg.X
				m.updateLayout()
				return m, nil
			}

		case tea.MouseActionRelease:
			if m.dragging {
				m.dragging = false
				return m, nil
			}
		}
	}

	return m, nil
}

func (m ChatModel) handleHelpBarClick(x int) (tea.Model, tea.Cmd) {
	bindings := ChatKeyBindings()
	regions := bindings.ButtonRegions()

	for _, r := range regions {
		if x >= r.XMin && x < r.XMax {
			kb := bindings.Bindings[r.Index]
			if len(kb.Keys) > 0 {
				return m.dispatchAction(kb.Keys[0])
			}
		}
	}
	return m, nil
}

func (m ChatModel) dispatchAction(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+q":
		return m, tea.Quit
	case "ctrl+a":
		switch m.sidebarTab {
		case TabChannels:
			m.sidebarTab = TabDMs
		case TabDMs:
			m.sidebarTab = TabUsers
		default:
			m.sidebarTab = TabChannels
		}
		m.switchChannel()
	case "ctrl+r":
		m.sidebarHidden = !m.sidebarHidden
		m.updateLayout()
	case "ctrl+d":
		modal := NewModal(ModalDMOpen)
		m.modal = &modal
		m.input.Blur()
	case "ctrl+s":
		if m.modal != nil && m.modal.Type == ModalShortcuts {
			m.modal = nil
			m.input.Focus()
		} else {
			modal := NewModal(ModalShortcuts)
			m.modal = &modal
			m.input.Blur()
		}
	case "ctrl+j":
		switch m.sidebarTab {
		case TabChannels:
			m.channels.MoveDown()
		case TabDMs:
			m.dmList.MoveDown()
		case TabUsers:
			m.userList.MoveDown()
		}
		m.switchChannel()
	case "ctrl+k":
		switch m.sidebarTab {
		case TabChannels:
			m.channels.MoveUp()
		case TabDMs:
			m.dmList.MoveUp()
		case TabUsers:
			m.userList.MoveUp()
		}
		m.switchChannel()
	case "ctrl+l":
		m.activePane = PaneInput
		m.input.Focus()
	case "ctrl+o":
		m.activePane = PaneChannels
		m.input.Blur()
	case "ctrl+n":
		modal := NewModal(ModalChannelCreate)
		m.modal = &modal
		m.input.Blur()
	case "ctrl+u":
		modal := NewModal(ModalUserInvite)
		m.modal = &modal
		m.input.Blur()
	case "ctrl+,":
		modal := NewSettingsModal(m.notifSound)
		m.modal = &modal
		m.input.Blur()
	case "ctrl+up":
		m.messages.ScrollUp(3)
		m.maybeLoadHistory()
	case "ctrl+down":
		m.messages.ScrollDown(3)
	}
	return m, nil
}

// trySendMessage sends the current input if the channel is writable and has text.
func (m *ChatModel) trySendMessage() bool {
	canWrite := m.sidebarTab == TabDMs || m.sidebarTab == TabUsers || m.channels.IsMember()
	if !canWrite || m.input.Value() == "" {
		return false
	}
	body := m.input.Value()
	log.Debug("send_message", "channel", m.selectedChannel, "body", body)
	m.client.SendMessage(m.selectedChannel, body)
	m.messages.AppendMessage(m.selectedChannel, sharkfin.Message{
		From:   m.username,
		Body:   body,
		SentAt: time.Now().UTC(),
	})
	m.input.Reset()
	m.updateLayout()
	return true
}

func (m *ChatModel) switchChannel() {
	var selected string
	switch m.sidebarTab {
	case TabChannels:
		selected = m.channels.Selected()
	case TabDMs:
		selected = m.dmList.Selected()
	case TabUsers:
		selected = m.userList.SelectedDMChannel()
		if selected == "" {
			// No DM exists yet with this user — open one
			if user := m.userList.SelectedUsername(); user != "" {
				m.client.DMOpen(user)
			}
			return
		}
	}
	if selected == "" || selected == m.selectedChannel {
		return
	}

	log.Debug("switch_channel", "from", m.selectedChannel, "to", selected, "tab", m.sidebarTab)

	// Advance server-side read cursor
	if latestID := m.messages.LatestID(selected); latestID > 0 {
		m.client.MarkRead(selected, &latestID)
	} else {
		m.client.MarkRead(selected, nil)
	}

	m.selectedChannel = selected
	switch m.sidebarTab {
	case TabChannels:
		m.channels.ClearUnread(selected)
	case TabDMs:
		m.dmList.ClearUnread(selected)
	case TabUsers:
		m.dmList.ClearUnread(selected)
		m.userList.ClearDMUnread(selected)
	}
	m.messages.SetChannel(selected)
	m.loadingHistory = false
	// Load recent history if no messages cached for this channel
	if m.messages.OldestID(selected) == 0 {
		m.client.RequestHistory(selected, 0, 50)
	}
	m.client.RequestUnread(selected)
	// DMs are always member; channels check membership
	if m.sidebarTab == TabChannels {
		readOnly := !m.channels.IsMember()
		m.input.SetReadOnly(readOnly)
		if readOnly {
			m.input.Reset()
			m.input.Blur()
			if m.activePane == PaneInput {
				m.activePane = PaneChannels
			}
		}
	} else {
		m.input.SetReadOnly(false)
	}
}

func (m *ChatModel) maybeLoadHistory() {
	if m.loadingHistory {
		return
	}
	if m.historyExhausted[m.selectedChannel] {
		return
	}
	if !m.messages.AtTop() {
		return
	}
	oldest := m.messages.OldestID(m.selectedChannel)
	if oldest > 0 {
		m.client.RequestHistory(m.selectedChannel, oldest, 50)
		m.loadingHistory = true
	}
}

func (m *ChatModel) layout() ui.ChatLayout {
	l := ui.CalculateChatLayoutWithSidebar(m.width, m.height, m.customSidebarW)
	if !l.Skinny && m.sidebarHidden {
		l.SidebarW = 0
		l.MessageW = m.width
	}
	return l
}

func (m *ChatModel) updateLayout() {
	layout := m.layout()

	if layout.SidebarW > 0 {
		// Inner heights subtract 2 for sidebar border, 3 for tab bar buttons (bordered)
		sidebarInnerH := layout.ContentH - 2 - 3
		m.channels.SetSize(layout.SidebarW, sidebarInnerH)
		m.dmList.SetSize(layout.SidebarW, sidebarInnerH)
		m.userList.SetSize(layout.SidebarW, sidebarInnerH)
	}

	inputH := m.input.Height()
	msgPaneOuterH := layout.ContentH - ui.ChannelHeaderH - inputH
	msgInnerH := msgPaneOuterH - 2

	m.messages.SetSize(layout.MessageW, msgInnerH)
	m.input.SetWidth(layout.MessageW - 2 - inputBtnW) // minus border minus action buttons
}

func (m ChatModel) View() string {
	if m.tooSmall {
		return ui.RenderSizeError(ui.MinWidth, ui.MinHeight, m.width, m.height)
	}

	if m.width == 0 || m.height == 0 {
		return ""
	}

	layout := m.layout()

	// Sidebar — tab bar as title, content based on active tab
	var sidebar string
	if !layout.Skinny && !m.sidebarHidden {
		var listView string
		switch m.sidebarTab {
		case TabChannels:
			listView = m.channels.View()
		case TabDMs:
			listView = m.dmList.View()
		case TabUsers:
			listView = m.userList.View()
		}
		tabBar := m.renderSidebarTabs(layout.SidebarW - 2) // minus sidebar border
		sidebarStyle := ui.CreatePaneStyle(m.activePane == PaneChannels, layout.SidebarW, layout.ContentH)
		sidebar = sidebarStyle.Render(tabBar + "\n" + listView)
	}

	// Channel header bar (mirrors input bar)
	var chanLabel string
	if m.selectedChannel == "" {
		chanLabel = "No channel"
	} else if m.sidebarTab == TabDMs {
		chanLabel = m.dmList.Participant()
	} else if m.sidebarTab == TabUsers {
		chanLabel = m.userList.SelectedUsername()
	} else {
		chanLabel = "#" + m.selectedChannel
	}
	chanHeaderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.CurrentTheme.Muted).
		Width(layout.MessageW - 2).
		Height(ui.ChannelHeaderH - 2)
	chanHeader := chanHeaderStyle.Render(
		lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary).Bold(true).Render(" " + chanLabel),
	)

	// Message pane
	msgStyle := ui.CreatePaneStyle(m.activePane == PaneInput, layout.MessageW, layout.ContentH-ui.ChannelHeaderH-m.input.Height())
	msgPane := msgStyle.Render(m.messages.View())

	// Input bar
	inputBorder := lipgloss.NormalBorder()
	inputBorderColor := ui.CurrentTheme.Muted
	if m.activePane == PaneInput {
		inputBorder = lipgloss.ThickBorder()
		inputBorderColor = ui.CurrentTheme.Primary
	}
	inputStyle := lipgloss.NewStyle().
		Border(inputBorder).
		BorderForeground(inputBorderColor).
		Width(layout.MessageW - 2 - inputBtnW).
		Height(m.input.Height() - 2)
	inputBox := inputStyle.Render(m.input.View())

	// Action buttons (mic + send) beside input, bordered like help bar
	canWrite := m.sidebarTab == TabDMs || m.sidebarTab == TabUsers || m.channels.IsMember()
	hasText := m.input.Value() != ""
	sendColor := ui.CurrentTheme.Muted
	if canWrite && hasText {
		sendColor = ui.CurrentTheme.Primary
	}
	btnStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.CurrentTheme.Muted).
		Width(5).
		Align(lipgloss.Center)
	micBtn := btnStyle.Foreground(ui.CurrentTheme.TextDim).Render("󰍬")
	sendBtn := btnStyle.BorderForeground(sendColor).Foreground(sendColor).Render("󰒊")
	btnGroup := lipgloss.JoinHorizontal(lipgloss.Top, micBtn, sendBtn)
	inputPane := lipgloss.JoinHorizontal(lipgloss.Bottom, inputBox, " ", btnGroup)

	// Compose right side
	rightPane := lipgloss.JoinVertical(lipgloss.Left, chanHeader, msgPane, inputPane)

	// Compose main layout
	var mainContent string
	if layout.Skinny || m.sidebarHidden {
		mainContent = rightPane
	} else {
		gap := strings.Repeat(" ", ui.PaneGap)
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, gap, rightPane)
	}

	// Compose full UI — skinny mode hides header and help bar
	var fullUI string
	if layout.Skinny {
		fullUI = mainContent
	} else {
		headerStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ui.CurrentTheme.Muted).
			Width(m.width - 2).
			Align(lipgloss.Center)
		header := headerStyle.Render(
			lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary).Bold(true).Render("WorkFort"),
		)
		helpBar := ChatKeyBindings().Render()
		fullUI = lipgloss.JoinVertical(lipgloss.Left, header, mainContent, helpBar)
	}

	// Modal overlay
	if m.modal != nil {
		return m.modal.View(m.width, m.height)
	}

	// Place constrains output to exact terminal dimensions
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, fullUI)
}

// renderSidebarTabs renders tab buttons matching the help bar button style.
func (m ChatModel) renderSidebarTabs(innerW int) string {
	activeStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)
	inactiveStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.TextDim)
	dotStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Accent)

	type tabDef struct {
		label     string
		tab       SidebarTab
		hasUnread bool
	}
	tabs := []tabDef{
		{"Channels", TabChannels, m.channels.HasUnreads()},
		{"DMs", TabDMs, m.dmList.HasUnreads()},
		{"Users", TabUsers, false},
	}

	// Three buttons filling available width (minus 1-col margins on each side)
	available := innerW - 2 // 1 left margin + 1 right margin
	btnContentW := (available - 6) / 3 // 6 = 2 border chars per button x 3
	remainder := (available - 6) % 3
	if btnContentW < 4 {
		btnContentW = 4
	}
	baseBtnStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.CurrentTheme.Muted).
		Align(lipgloss.Center)

	var rendered []string
	for i, t := range tabs {
		w := btnContentW
		if i == len(tabs)-1 {
			w += remainder // last button absorbs integer division slack
		}
		var text string
		if m.sidebarTab == t.tab {
			text = activeStyle.Render(t.label)
		} else if t.hasUnread {
			text = inactiveStyle.Render(t.label) + dotStyle.Render(" ●")
		} else {
			text = inactiveStyle.Render(t.label)
		}
		rendered = append(rendered, baseBtnStyle.Width(w).Render(text))
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	return lipgloss.NewStyle().MarginLeft(1).MarginRight(1).Render(bar)
}

// isLeakedMouseSeq detects SGR mouse escape sequence fragments that Bubble Tea's
// parser failed to consume. These look like "[<64;87;52M" — digits and semicolons
// bracketed by "[<" and "M"/"m". Also catches multi-rune fragments where the
// sequence is partially split.
func isLeakedMouseSeq(msg tea.KeyMsg) bool {
	// Only KeyRunes can be leaked mouse fragments
	if msg.Type != tea.KeyRunes {
		return false
	}
	// Bracketed paste wraps content in "[...]" — never a mouse sequence
	if tea.Key(msg).Paste {
		return false
	}
	s := msg.String()
	if len(s) < 2 {
		return false
	}
	// Full or partial SGR: "[<..."
	if strings.HasPrefix(s, "[<") {
		return true
	}
	// Multi-char starting with "[" (real bracket typing is single rune)
	if s[0] == '[' && len(msg.Runes) > 1 {
		return true
	}
	// Fragment starting with "<" containing semicolons
	if s[0] == '<' && strings.ContainsAny(s, ";") {
		return true
	}
	// Digits+semicolons ending in M/m
	if strings.ContainsAny(s, ";") {
		last := s[len(s)-1]
		if last == 'M' || last == 'm' {
			return true
		}
	}
	return false
}

func (m *ChatModel) tryModalComplete() {
	prefix := strings.ToLower(m.modal.Value())
	if prefix == "" {
		return
	}
	var matches []string
	for _, u := range m.users {
		if strings.HasPrefix(strings.ToLower(u.Username), prefix) {
			matches = append(matches, u.Username)
		}
	}
	if len(matches) == 0 {
		return
	}
	if len(matches) == 1 {
		m.modal.SetValue(matches[0])
	} else {
		lcp := longestCommonPrefix(matches)
		if len(lcp) > len(prefix) {
			m.modal.SetValue(lcp)
		}
	}
}

func (m *ChatModel) tryChannelModalComplete() {
	prefix := strings.ToLower(m.modal.Value())
	if prefix == "" {
		return
	}
	names := m.channels.Names()
	var matches []string
	for _, n := range names {
		if strings.HasPrefix(strings.ToLower(n), prefix) {
			matches = append(matches, n)
		}
	}
	if len(matches) == 0 {
		m.modal.hint = "New channel — Enter to create"
		return
	}
	if len(matches) == 1 {
		m.modal.SetValue(matches[0])
		if ch, found := m.channels.FindByName(matches[0]); found && ch.Member {
			m.modal.hint = "Switch to #" + matches[0]
		} else {
			m.modal.hint = "Join #" + matches[0]
		}
	} else {
		lcp := longestCommonPrefix(matches)
		if len(lcp) > len(prefix) {
			m.modal.SetValue(lcp)
		}
		m.modal.hint = fmt.Sprintf("%d channels match", len(matches))
	}
}

func (m *ChatModel) usernameList() []string {
	names := make([]string, len(m.users))
	for i, u := range m.users {
		names[i] = u.Username
	}
	return names
}

func containsUser(mentions []string, username string) bool {
	for _, m := range mentions {
		if strings.EqualFold(m, username) {
			return true
		}
	}
	return false
}
