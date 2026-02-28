package chat

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"

	"github.com/spf13/viper"

	"github.com/Work-Fort/WorkFort/pkg/audio"
	"github.com/Work-Fort/WorkFort/pkg/config"
	"github.com/Work-Fort/WorkFort/pkg/sharkfin"
	"github.com/Work-Fort/WorkFort/pkg/stt"
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

// RecordingState tracks the current state of voice recording.
type RecordingState int

const (
	RecordingIdle         RecordingState = iota
	RecordingDownloading                        // model downloading
	RecordingActive                             // mic is on, capturing audio
	RecordingTranscribing                       // final transcription in progress
)

// STT tea messages — lowercase to keep them package-private.
type (
	modelReadyMsg         struct{ path string }
	modelDownloadErrorMsg struct{ err error }
	streamSegmentMsg      struct {
		text       string
		final      bool
		generation int
	}
	streamTickMsg      struct{ generation int }
	transcribeErrorMsg struct {
		err        error
		generation int
	}
	downloadTickMsg struct{}
)

// dlSpinnerFrames are quarter-fill circles that cycle clockwise.
var dlSpinnerFrames = []string{"◐", "◓", "◑", "◒"}

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

	// STT (speech-to-text) state
	recording      RecordingState
	recorder       *stt.Recorder
	transcriber    *stt.Transcriber // lazy-loaded on first mic press
	sttGeneration  int              // monotonic counter to discard stale results
	sttInFlight    bool             // true while a streaming transcription cmd is running
	recordingStart time.Time        // when recording started (for 30s timeout)
	downloadFrame  int              // spinner frame index during model download

	disconnected     bool // true when WS connection is lost
	reconnectAttempt int  // current reconnect attempt number
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

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case tea.PasteMsg:
		if m.activePane == PaneInput && m.modal == nil && m.recording == RecordingIdle {
			prevH := m.input.Height()
			m.input.InsertString(msg.Content)
			if m.input.Height() != prevH {
				m.updateLayout()
			}
		}
		return m, nil

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
		if !m.disconnected {
			m.disconnected = true
			m.reconnectAttempt = 0
			return m, m.client.Reconnect()
		}
		return m, nil

	case sharkfin.ReconnectingMsg:
		m.reconnectAttempt = msg.Attempt
		return m, nil

	case sharkfin.ConnectedMsg:
		m.disconnected = false
		m.reconnectAttempt = 0
		// Re-fetch all state
		m.client.RequestChannels()
		m.client.RequestDMList()
		m.client.RequestUsers()
		if m.selectedChannel != "" {
			m.client.RequestHistory(m.selectedChannel, 0, 50)
			m.client.RequestUnread(m.selectedChannel)
		}
		return m, nil

	// --- STT messages ---

	case modelReadyMsg:
		log.Debug("stt_model_ready", "path", msg.path)
		lang := viper.GetString("stt-language")
		threads := uint(viper.GetInt("stt-threads"))
		t, err := stt.NewTranscriber(msg.path, lang, threads)
		if err != nil {
			log.Error("stt_transcriber_init", "err", err)
			m.recording = RecordingIdle
			return m, nil
		}
		m.transcriber = t
		// Auto-start recording now that model is loaded
		return m.startRecording()

	case modelDownloadErrorMsg:
		log.Error("stt_model_download", "err", msg.err)
		m.recording = RecordingIdle
		return m, nil

	case downloadTickMsg:
		if m.recording == RecordingDownloading || m.recording == RecordingTranscribing {
			m.downloadFrame++
			return m, downloadTickCmd()
		}
		return m, nil

	case streamTickMsg:
		if msg.generation != m.sttGeneration || m.recording != RecordingActive {
			return m, nil
		}
		// Check 30s timeout
		if time.Since(m.recordingStart) >= 30*time.Second {
			return m.stopRecording()
		}
		if m.sttInFlight {
			// Previous transcription still running — schedule next tick
			return m, sttTickCmd(m.sttGeneration)
		}
		m.sttInFlight = true
		samples := m.recorder.Snapshot()
		gen := m.sttGeneration
		return m, tea.Batch(
			sttTranscribeCmd(m.transcriber, samples, false, gen),
			sttTickCmd(gen),
		)

	case streamSegmentMsg:
		if msg.generation != m.sttGeneration {
			return m, nil // stale result
		}
		if msg.text != "" {
			prevH := m.input.Height()
			m.input.SetSTTText(msg.text)
			if m.input.Height() != prevH {
				m.updateLayout()
			}
		}
		if msg.final {
			m.input.ClearSTTState()
			m.recording = RecordingIdle
			m.sttInFlight = false
			return m, nil
		}
		m.sttInFlight = false
		return m, nil

	case transcribeErrorMsg:
		if msg.generation != m.sttGeneration {
			return m, nil
		}
		log.Error("stt_transcribe", "err", msg.err)
		m.input.ClearSTTState()
		m.recording = RecordingIdle
		m.sttInFlight = false
		return m, nil
	}

	return m, nil
}

func (m ChatModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Modal takes priority if open
	if m.modal != nil {
		return m.handleModalKey(msg)
	}

	switch key {
	case "ctrl+q":
		return m, tea.Quit

	case "ctrl+m":
		return m.toggleMic()

	case "ctrl+,", "ctrl+.":
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
		if m.activePane == PaneInput && m.recording == RecordingIdle {
			m.input.InsertNewline()
			m.updateLayout()
			return m, nil
		}

	case "ctrl+v":
		if m.activePane == PaneInput && m.recording == RecordingIdle {
			m.input.Paste()
			m.updateLayout()
			return m, nil
		}

	case "enter":
		if m.activePane == PaneInput && m.recording == RecordingIdle && m.trySendMessage() {
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
	if m.activePane == PaneInput && !leaked && m.recording == RecordingIdle {
		prevH := m.input.Height()
		m.input.UpdateTextInput(msg)
		if m.input.Height() != prevH {
			m.updateLayout()
		}
	}

	return m, nil
}

func (m ChatModel) handleModalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m.modal = nil
		m.input.Focus()
		return m, nil

	case "ctrl+s", "ctrl+d", "ctrl+n", "ctrl+u", "ctrl+,", "ctrl+.":
		toggleType := map[string]ModalType{
			"ctrl+s": ModalShortcuts,
			"ctrl+d": ModalDMOpen,
			"ctrl+n": ModalChannelCreate,
			"ctrl+u": ModalUserInvite,
			"ctrl+,": ModalSettings,
			"ctrl+.": ModalSettings,
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
	mouse := msg.Mouse()
	m.lastMouseY = mouse.Y
	m.lastMouseTime = time.Now()

	// Modal click handling
	if m.modal != nil {
		if _, ok := msg.(tea.MouseClickMsg); ok && mouse.Button == tea.MouseLeft {
			// Sound selector click
			if idx := m.modal.SoundHitTest(mouse.X, mouse.Y); idx >= 0 {
				m.modal.soundCursor = idx
				audio.Play(m.modal.SelectedSound())
				return m, nil
			}
			action := m.modal.HitTest(mouse.X, mouse.Y)
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

	switch msg.(type) {
	case tea.MouseWheelMsg:
		if mouse.Button == tea.MouseWheelUp {
			if mouse.X < layout.SidebarW {
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
				if m.messages.Height() <= 24 {
					if time.Since(m.lastMsgScroll) < 80*time.Millisecond {
						return m, nil
					}
					m.lastMsgScroll = time.Now()
				}
				m.messages.ScrollUp(3)
				m.maybeLoadHistory()
			}
		} else if mouse.Button == tea.MouseWheelDown {
			if mouse.X < layout.SidebarW {
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
				if m.messages.Height() <= 24 {
					if time.Since(m.lastMsgScroll) < 80*time.Millisecond {
						return m, nil
					}
					m.lastMsgScroll = time.Now()
				}
				m.messages.ScrollDown(3)
			}
		}
		return m, nil

	case tea.MouseClickMsg:
		if mouse.Button != tea.MouseLeft {
			return m, nil
		}

		// Check if pressing on the divider gap (±1 char hit zone)
		if !layout.Skinny && !m.sidebarHidden {
			gap := layout.SidebarW
			if mouse.X >= gap-1 && mouse.X <= gap+1 {
				m.dragging = true
				return m, nil
			}
		}

		contentTop := ui.HeaderHeight
		contentBottom := contentTop + layout.ContentH

		// Help bar spans full width — check Y first
		if mouse.Y >= contentBottom {
			return m.handleHelpBarClick(mouse.X)
		}

		// Click in header — ignore
		if mouse.Y < contentTop {
			return m, nil
		}

		if mouse.X < layout.SidebarW {
			// Tab bar buttons span 3 rows inside the sidebar border
			tabBarTop := contentTop + 1 // after sidebar border
			tabBarBottom := tabBarTop + 2
			if mouse.Y >= tabBarTop && mouse.Y <= tabBarBottom {
				// Click on tab bar — split into thirds
				third := layout.SidebarW / 3
				if mouse.X < third {
					m.sidebarTab = TabChannels
				} else if mouse.X < third*2 {
					m.sidebarTab = TabDMs
				} else {
					m.sidebarTab = TabUsers
				}
				m.switchChannel()
				return m, nil
			}

			// Click in sidebar list: select item from active tab
			// Account for border (1) + tab bar (3)
			row := mouse.Y - contentTop - 1 - 3
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
		if mouse.X >= btnAreaX && mouse.Y >= btnTop && mouse.Y < inputBottom {
			// Each bordered button is 7 wide (border+Width(5)+border)
			// 1 space gap before buttons, so send starts at offset 1+7=8
			if mouse.X >= btnAreaX+8 {
				if m.recording == RecordingIdle {
					m.trySendMessage()
				}
			} else {
				return m.toggleMic()
			}
			return m, nil
		}

		// Click in right pane — focus input if writable
		canWrite := m.sidebarTab == TabDMs || m.sidebarTab == TabUsers || m.channels.IsMember()
		if canWrite {
			m.activePane = PaneInput
			m.input.Focus()
		}
		return m, nil

	case tea.MouseMotionMsg:
		if m.dragging {
			m.customSidebarW = mouse.X
			m.updateLayout()
			return m, nil
		}

	case tea.MouseReleaseMsg:
		if m.dragging {
			m.dragging = false
			return m, nil
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
	case "ctrl+m":
		return m.toggleMic()
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
	case "ctrl+,", "ctrl+.":
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
	if m.disconnected {
		return false
	}
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

func (m ChatModel) View() tea.View {
	view := func(s string) tea.View {
		v := tea.NewView(s)
		v.AltScreen = true
		v.MouseMode = tea.MouseModeCellMotion
		return v
	}

	if m.tooSmall {
		return view(ui.RenderSizeError(ui.MinWidth, ui.MinHeight, m.width, m.height))
	}

	if m.width == 0 || m.height == 0 {
		return view("")
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
	chanHeaderBorderColor := ui.CurrentTheme.Muted
	var chanExtra string
	if m.disconnected {
		chanHeaderBorderColor = ui.CurrentTheme.Error
		status := "Reconnecting"
		if m.reconnectAttempt > 1 {
			status = fmt.Sprintf("Reconnecting (attempt %d)", m.reconnectAttempt)
		}
		chanExtra = ui.CurrentTheme.ErrorStyle().Render("  " + status)
	}
	chanHeaderStyle := ui.CurrentTheme.ChannelHeaderStyle(layout.MessageW, ui.ChannelHeaderH, chanHeaderBorderColor)
	chanHeader := chanHeaderStyle.Render(
		ui.CurrentTheme.HeaderTitleStyle().Render(" "+chanLabel) + chanExtra,
	)

	// Message pane
	msgPaneH := layout.ContentH - ui.ChannelHeaderH - m.input.Height()
	msgStyle := ui.CreatePaneStyle(m.activePane == PaneInput, layout.MessageW, msgPaneH)
	msgView := m.messages.View()
	msgPane := msgStyle.Render(msgView)

	// Input bar — red border means input is locked (recording or transcribing)
	inputBorder := lipgloss.NormalBorder()
	inputBorderColor := ui.CurrentTheme.Muted
	if m.activePane == PaneInput {
		inputBorder = lipgloss.ThickBorder()
		inputBorderColor = ui.CurrentTheme.Primary
	}
	if m.recording != RecordingIdle {
		inputBorderColor = ui.CurrentTheme.Error
	}
	inputStyle := lipgloss.NewStyle().
		Border(inputBorder).
		BorderForeground(inputBorderColor).
		Width(layout.MessageW - inputBtnW).
		Height(m.input.Height())
	inputBox := inputStyle.Render(m.input.View())

	// Action buttons (mic + send) beside input, bordered like help bar
	canWrite := m.sidebarTab == TabDMs || m.sidebarTab == TabUsers || m.channels.IsMember()
	hasText := m.input.Value() != ""
	sendColor := ui.CurrentTheme.Muted
	if canWrite && hasText {
		sendColor = ui.CurrentTheme.Primary
	}
	btnStyle := ui.CurrentTheme.ActionButtonStyle(7)
	// Mic button: orange (ready), green spinner (downloading), red (recording), grey (transcribing)
	micIcon := "󰍬"
	micColor := ui.CurrentTheme.Primary
	switch m.recording {
	case RecordingDownloading:
		micIcon = dlSpinnerFrames[m.downloadFrame%len(dlSpinnerFrames)]
		micColor = ui.CurrentTheme.Success
	case RecordingActive:
		micIcon = "●"
		micColor = ui.CurrentTheme.Error
	case RecordingTranscribing:
		micIcon = dlSpinnerFrames[m.downloadFrame%len(dlSpinnerFrames)]
		micColor = ui.CurrentTheme.Error
	}
	micBorderColor := micColor
	if m.recording == RecordingTranscribing {
		micBorderColor = ui.CurrentTheme.TextDim // grey border, red spinner
	}
	micBtn := btnStyle.Foreground(micColor).BorderForeground(micBorderColor).Render(micIcon)
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
		header := ui.CurrentTheme.HeaderStyle(m.width).Render(
			ui.CurrentTheme.HeaderTitleStyle().Render("WorkFort"),
		)
		helpBar := ChatKeyBindings().Render()
		fullUI = lipgloss.JoinVertical(lipgloss.Left, header, mainContent, helpBar)
	}

	// Modal overlay
	if m.modal != nil {
		return view(m.modal.View(m.width, m.height))
	}

	// Place constrains output to exact terminal dimensions
	return view(lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, fullUI))
}

// renderSidebarTabs renders tab buttons matching the help bar button style.
func (m ChatModel) renderSidebarTabs(innerW int) string {
	activeStyle := ui.CurrentTheme.TabActiveStyle()
	inactiveStyle := ui.CurrentTheme.TabInactiveStyle()
	dotStyle := ui.CurrentTheme.AccentStyle()

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
	btnContentW := available / 3 // v2: Width includes border
	remainder := available % 3
	if btnContentW < 6 {
		btnContentW = 6
	}
	baseBtnStyle := ui.CurrentTheme.TabBaseStyle()

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
func isLeakedMouseSeq(msg tea.KeyPressMsg) bool {
	// Only printable text keys can be leaked mouse fragments
	text := msg.Key().Text
	if text == "" {
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
	// Multi-char starting with "[" (real bracket typing is single char)
	if s[0] == '[' && len(text) > 1 {
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

// toggleMic handles Ctrl+M or mic button click.
func (m ChatModel) toggleMic() (tea.Model, tea.Cmd) {
	switch m.recording {
	case RecordingIdle:
		if m.transcriber == nil {
			modelName := viper.GetString("stt-model")
			if stt.ModelCached(modelName) {
				// Model on disk — load inline, no download spinner
				log.Debug("stt_model_loading")
				return m, sttDownloadModelCmd()
			}
			// Need to download model — show green spinner
			m.recording = RecordingDownloading
			m.downloadFrame = 0
			log.Debug("stt_download_start")
			return m, tea.Batch(sttDownloadModelCmd(), downloadTickCmd())
		}
		return m.startRecording()
	case RecordingActive:
		return m.stopRecording()
	case RecordingDownloading, RecordingTranscribing:
		// Busy — ignore
	}
	return m, nil
}

// startRecording begins audio capture and starts the streaming tick loop.
func (m ChatModel) startRecording() (tea.Model, tea.Cmd) {
	if m.recorder == nil {
		rec, err := stt.NewRecorder()
		if err != nil {
			log.Error("stt_recorder_init", "err", err)
			m.recording = RecordingIdle
			return m, nil
		}
		m.recorder = rec
	}
	if err := m.recorder.Start(); err != nil {
		log.Error("stt_recorder_start", "err", err)
		m.recording = RecordingIdle
		return m, nil
	}
	m.recording = RecordingActive
	m.recordingStart = time.Now()
	m.sttGeneration++
	m.sttInFlight = false
	m.input.BeginSTT()
	log.Debug("stt_recording_start", "generation", m.sttGeneration)
	return m, sttTickCmd(m.sttGeneration)
}

// stopRecording stops audio capture and runs a final transcription pass.
func (m ChatModel) stopRecording() (tea.Model, tea.Cmd) {
	samples := m.recorder.Stop()
	m.recording = RecordingTranscribing
	m.downloadFrame = 0
	m.sttInFlight = true
	gen := m.sttGeneration
	log.Debug("stt_recording_stop", "samples", len(samples), "generation", gen)
	return m, tea.Batch(sttTranscribeCmd(m.transcriber, samples, true, gen), downloadTickCmd())
}

// sttDownloadModelCmd returns a Cmd that downloads the whisper model.
func sttDownloadModelCmd() tea.Cmd {
	return func() tea.Msg {
		modelName := viper.GetString("stt-model")
		path, err := stt.EnsureModel(modelName, nil)
		if err != nil {
			return modelDownloadErrorMsg{err: err}
		}
		return modelReadyMsg{path: path}
	}
}

// downloadTickCmd returns a Cmd that fires a downloadTickMsg every 80ms for spinner animation.
func downloadTickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(_ time.Time) tea.Msg {
		return downloadTickMsg{}
	})
}

// sttTickCmd returns a Cmd that fires a streamTickMsg after 2.5 seconds.
func sttTickCmd(generation int) tea.Cmd {
	return tea.Tick(2500*time.Millisecond, func(_ time.Time) tea.Msg {
		return streamTickMsg{generation: generation}
	})
}

// sttTranscribeCmd returns a Cmd that transcribes audio samples.
func sttTranscribeCmd(t *stt.Transcriber, samples []float32, final bool, generation int) tea.Cmd {
	return func() tea.Msg {
		text, err := t.Transcribe(samples)
		if err != nil {
			return transcribeErrorMsg{err: err, generation: generation}
		}
		return streamSegmentMsg{text: text, final: final, generation: generation}
	}
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
