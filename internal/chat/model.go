package chat

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/Work-Fort/WorkFort/pkg/sharkfin"
	"github.com/Work-Fort/WorkFort/pkg/ui"
)

// Pane identifies which panel has focus.
type Pane int

const (
	PaneChannels Pane = iota
	PaneInput
)

// SidebarTab identifies which sidebar list is visible.
type SidebarTab int

const (
	TabChannels SidebarTab = iota
	TabDMs
)

// ChatModel is the root Bubble Tea model for the chat UI.
type ChatModel struct {
	width    int
	height   int
	tooSmall bool

	client   *sharkfin.Client
	channels ChannelList
	dmList   DMList
	messages MessagePane
	input    InputBar
	modal    *Modal

	activePane      Pane
	sidebarTab      SidebarTab
	selectedChannel string
	username        string
	users           []sharkfin.User

	lastChannelScroll time.Time
	loadingHistory    bool
	historyExhausted  map[string]bool // channels that have no more history to load

	customSidebarW int  // user-dragged sidebar width (0 = auto)
	dragging       bool // true while dragging the divider
}

// NewModel creates a ChatModel wired to a sharkfin client.
func NewModel(client *sharkfin.Client, username string) ChatModel {
	cl := NewChannelList()
	dl := NewDMList(username)
	mp := NewMessagePane()
	ib := NewInputBar()

	return ChatModel{
		client:           client,
		channels:         cl,
		dmList:           dl,
		messages:         mp,
		input:            ib,
		activePane:       PaneChannels,
		username:         username,
		historyExhausted: make(map[string]bool),
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

	case sharkfin.ChannelListMsg:
		log.Debug("channel_list", "count", len(msg.Channels))
		m.channels.SetChannels(msg.Channels)
		m.client.RequestUnreadCounts()
		if len(msg.Channels) > 0 && m.selectedChannel == "" {
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
			isMention := containsUser(msg.Mentions, m.username)
			if msg.ChannelType == "dm" {
				if isMention {
					m.dmList.IncrementMention(msg.Channel)
				} else {
					m.dmList.IncrementUnread(msg.Channel)
				}
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
		return m, nil

	case sharkfin.UserListMsg:
		m.users = msg.Users
		return m, nil

	case sharkfin.MessageSentMsg:
		return m, nil

	case sharkfin.DMListMsg:
		log.Debug("dm_list", "count", len(msg.DMs))
		m.dmList.SetDMs(msg.DMs)
		return m, nil

	case sharkfin.DMOpenMsg:
		log.Debug("dm_open", "channel", msg.Channel, "participant", msg.Participant, "created", msg.Created)
		if msg.Created {
			m.client.RequestDMList()
		}
		m.sidebarTab = TabDMs
		m.selectedChannel = msg.Channel
		m.dmList.ClearUnread(msg.Channel)
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

	case "ctrl+a":
		if m.sidebarTab == TabChannels {
			m.sidebarTab = TabDMs
		} else {
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
		modal := NewModal(ModalShortcuts)
		m.modal = &modal
		m.input.Blur()
		return m, nil

	case "ctrl+j":
		if m.sidebarTab == TabChannels {
			m.channels.MoveDown()
		} else {
			m.dmList.MoveDown()
		}
		m.switchChannel()
		return m, nil

	case "ctrl+k":
		if m.sidebarTab == TabChannels {
			m.channels.MoveUp()
		} else {
			m.dmList.MoveUp()
		}
		m.switchChannel()
		return m, nil

	case "ctrl+l":
		canWrite := m.sidebarTab == TabDMs || m.channels.IsMember()
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
			return m, nil
		}

	case "enter":
		canWrite := m.sidebarTab == TabDMs || m.channels.IsMember()
		if m.activePane == PaneInput && m.input.Value() != "" && canWrite {
			body := m.input.Value()
			log.Debug("send_message", "channel", m.selectedChannel, "body", body)
			m.client.SendMessage(m.selectedChannel, body)
			// Append locally — server doesn't echo message.new to sender
			m.messages.AppendMessage(m.selectedChannel, sharkfin.Message{
				From:   m.username,
				Body:   body,
				SentAt: time.Now().UTC(),
			})
			m.input.Reset()
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
	// Bubble Tea's parser can split SGR mouse events across reads, delivering fragments
	// like "[<64;87;52M" as KeyRunes. Drop these before they reach the text input.
	if m.activePane == PaneInput && !isLeakedMouseSeq(key) {
		m.input.UpdateTextInput(msg)
	}

	return m, nil
}

func (m ChatModel) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m.modal = nil
		return m, nil

	case "enter":
		if m.modal.Value() != "" {
			m.submitModal()
			m.modal = nil
		}
		return m, nil

	case "tab":
		if m.modal.Type == ModalChannelCreate {
			m.modal.TogglePublic()
		}
		return m, nil
	}

	if m.modal.Type != ModalShortcuts {
		m.modal.UpdateTextInput(msg)
	}
	return m, nil
}

func (m *ChatModel) submitModal() {
	switch m.modal.Type {
	case ModalChannelCreate:
		m.client.CreateChannel(m.modal.Value(), m.modal.IsPublic(), nil)
	case ModalUserInvite:
		m.client.InviteUser(m.selectedChannel, m.modal.Value())
	case ModalDMOpen:
		m.client.DMOpen(m.modal.Value())
	}
}

func (m ChatModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Modal click handling
	if m.modal != nil {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			action := m.modal.HitTest(msg.X, msg.Y)
			switch action {
			case ModalActionSubmit:
				if m.modal.Value() != "" {
					m.submitModal()
					m.modal = nil
				}
			case ModalActionToggle:
				if m.modal.Type == ModalChannelCreate {
					m.modal.TogglePublic()
				}
			case ModalActionCancel:
				m.modal = nil
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
			if m.sidebarTab == TabChannels {
				m.channels.MoveUp()
			} else {
				m.dmList.MoveUp()
			}
			m.switchChannel()
		} else {
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
			if m.sidebarTab == TabChannels {
				m.channels.MoveDown()
			} else {
				m.dmList.MoveDown()
			}
			m.switchChannel()
		} else {
			m.messages.ScrollDown(3)
		}
		return m, nil

	case tea.MouseButtonLeft:
		switch msg.Action {
		case tea.MouseActionPress:
			// Check if pressing on the divider gap (±1 char hit zone)
			gap := layout.SidebarW
			if msg.X >= gap-1 && msg.X <= gap+1 {
				m.dragging = true
				return m, nil
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
				// Tab bar is the first row inside the sidebar border
				tabBarY := contentTop + 1 // border top
				if msg.Y == tabBarY {
					// Click on tab bar — determine which tab
					// Tab bar: " Channels │ DMs"
					// "Channels" starts at X=2 (border+space), ~10 chars
					// Separator at ~11, "DMs" starts at ~14
					relX := msg.X - 1 // subtract border
					if relX < 11 {
						m.sidebarTab = TabChannels
					} else {
						m.sidebarTab = TabDMs
					}
					m.switchChannel()
					return m, nil
				}

				// Click in sidebar list: select item from active tab
				// Account for border (1) + tab bar (1)
				row := msg.Y - contentTop - 1 - 1
				if row >= 0 {
					if m.sidebarTab == TabChannels {
						m.channels.SelectIndex(row)
					} else {
						m.dmList.SelectIndex(row)
					}
					m.switchChannel()
				}
				return m, nil
			}

			// Click in right pane — focus input if writable
			canWrite := m.sidebarTab == TabDMs || m.channels.IsMember()
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
		if m.sidebarTab == TabChannels {
			m.sidebarTab = TabDMs
		} else {
			m.sidebarTab = TabChannels
		}
		m.switchChannel()
	case "ctrl+d":
		modal := NewModal(ModalDMOpen)
		m.modal = &modal
		m.input.Blur()
	case "ctrl+s":
		modal := NewModal(ModalShortcuts)
		m.modal = &modal
		m.input.Blur()
	case "ctrl+j":
		if m.sidebarTab == TabChannels {
			m.channels.MoveDown()
		} else {
			m.dmList.MoveDown()
		}
		m.switchChannel()
	case "ctrl+k":
		if m.sidebarTab == TabChannels {
			m.channels.MoveUp()
		} else {
			m.dmList.MoveUp()
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
	case "ctrl+up":
		m.messages.ScrollUp(3)
		m.maybeLoadHistory()
	case "ctrl+down":
		m.messages.ScrollDown(3)
	}
	return m, nil
}

func (m *ChatModel) switchChannel() {
	var selected string
	if m.sidebarTab == TabChannels {
		selected = m.channels.Selected()
	} else {
		selected = m.dmList.Selected()
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
	if m.sidebarTab == TabChannels {
		m.channels.ClearUnread(selected)
	} else {
		m.dmList.ClearUnread(selected)
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
	return ui.CalculateChatLayoutWithSidebar(m.width, m.height, m.customSidebarW)
}

func (m *ChatModel) updateLayout() {
	layout := m.layout()

	// Inner heights subtract 2 for border, PaneTitleH for tab bar+gap
	sidebarInnerH := layout.ContentH - 2 - ui.PaneTitleH
	inputH := m.input.Height()
	msgPaneOuterH := layout.ContentH - ui.ChannelHeaderH - inputH
	msgInnerH := msgPaneOuterH - 2

	m.channels.SetSize(layout.SidebarW, sidebarInnerH)
	m.dmList.SetSize(layout.SidebarW, sidebarInnerH)
	m.messages.SetSize(layout.MessageW, msgInnerH)
	m.input.SetWidth(layout.MessageW - 2) // minus border
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
	var listView string
	if m.sidebarTab == TabChannels {
		listView = m.channels.View()
	} else {
		listView = m.dmList.View()
	}
	tabBar := m.renderSidebarTabs(layout.SidebarW - 4) // minus border+padding
	sidebarStyle := ui.CreatePaneStyle(m.activePane == PaneChannels, layout.SidebarW, layout.ContentH)
	sidebar := sidebarStyle.Render(tabBar + "\n" + listView)

	// Channel header bar (mirrors input bar)
	var chanLabel string
	if m.selectedChannel == "" {
		chanLabel = "No channel"
	} else if m.sidebarTab == TabDMs {
		chanLabel = m.dmList.Participant()
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
		Width(layout.MessageW - 2).
		Height(m.input.Height() - 2)
	inputPane := inputStyle.Render(m.input.View())

	// Compose right side
	rightPane := lipgloss.JoinVertical(lipgloss.Left, chanHeader, msgPane, inputPane)

	// Compose main layout
	gap := strings.Repeat(" ", ui.PaneGap)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, gap, rightPane)

	// Header bar
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.CurrentTheme.Muted).
		Width(m.width - 2).
		Align(lipgloss.Center)
	header := headerStyle.Render(
		lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary).Bold(true).Render("WorkFort"),
	)

	// Help bar
	helpBar := ChatKeyBindings().Render()

	fullUI := lipgloss.JoinVertical(lipgloss.Left, header, mainContent, helpBar)

	// Modal overlay
	if m.modal != nil {
		return m.modal.View(m.width, m.height)
	}

	// Place constrains output to exact terminal dimensions
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, fullUI)
}

// renderSidebarTabs renders tab labels at the bottom of the sidebar.
func (m ChatModel) renderSidebarTabs(width int) string {
	activeStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)
	inactiveStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.TextDim)
	dotStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Accent)

	chLabel := "Channels"
	dmLabel := "DMs"

	// Add unread dot to inactive tab if it has unreads
	if m.sidebarTab == TabChannels {
		chLabel = activeStyle.Render(chLabel)
		if m.dmList.HasUnreads() {
			dmLabel = inactiveStyle.Render(dmLabel) + dotStyle.Render("●")
		} else {
			dmLabel = inactiveStyle.Render(dmLabel)
		}
	} else {
		dmLabel = activeStyle.Render(dmLabel)
		if m.channels.HasUnreads() {
			chLabel = inactiveStyle.Render(chLabel) + dotStyle.Render("●")
		} else {
			chLabel = inactiveStyle.Render(chLabel)
		}
	}

	sep := inactiveStyle.Render(" │ ")
	return " " + chLabel + sep + dmLabel
}

// isLeakedMouseSeq detects SGR mouse escape sequence fragments that Bubble Tea's
// parser failed to consume. These look like "[<64;87;52M" — digits and semicolons
// bracketed by "[<" and "M"/"m".
func isLeakedMouseSeq(s string) bool {
	if !strings.HasPrefix(s, "[<") {
		return false
	}
	for _, r := range s[2:] {
		switch {
		case r >= '0' && r <= '9', r == ';':
			continue
		case r == 'M' || r == 'm':
			return true
		default:
			return false
		}
	}
	// Partial sequence (no trailing M/m yet) — still drop it
	return len(s) > 2
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
