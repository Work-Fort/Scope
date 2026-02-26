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

// ChatModel is the root Bubble Tea model for the chat UI.
type ChatModel struct {
	width    int
	height   int
	tooSmall bool

	client   *sharkfin.Client
	channels ChannelList
	messages MessagePane
	input    InputBar
	modal    *Modal

	activePane      Pane
	selectedChannel string
	username        string
	users           []sharkfin.User

	lastChannelScroll time.Time
	loadingHistory    bool
}

// NewModel creates a ChatModel wired to a sharkfin client.
func NewModel(client *sharkfin.Client, username string) ChatModel {
	cl := NewChannelList()
	mp := NewMessagePane()
	ib := NewInputBar()

	return ChatModel{
		client:     client,
		channels:   cl,
		messages:   mp,
		input:      ib,
		activePane: PaneChannels,
		username:   username,
	}
}

func (m ChatModel) Init() tea.Cmd {
	m.client.RequestChannels()
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
		if len(msg.Channels) > 0 && m.selectedChannel == "" {
			m.selectedChannel = msg.Channels[0].Name
			m.messages.SetChannel(m.selectedChannel)
			m.client.RequestHistory(m.selectedChannel, 0, 50)
			m.client.RequestUnread(m.selectedChannel)
		}
		return m, nil

	case sharkfin.HistoryMsg:
		log.Debug("history", "channel", msg.Channel, "count", len(msg.Messages), "scrollback", m.loadingHistory)
		if m.loadingHistory {
			// Scrollback: prepend older messages, preserve scroll position
			m.messages.PrependHistory(msg.Channel, msg.Messages)
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
			m.channels.IncrementUnread(msg.Channel)
		}
		return m, nil

	case sharkfin.PresenceMsg:
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

	case "ctrl+j":
		m.channels.MoveDown()
		m.switchChannel()
		return m, nil

	case "ctrl+k":
		m.channels.MoveUp()
		m.switchChannel()
		return m, nil

	case "ctrl+l":
		m.activePane = PaneInput
		m.input.Focus()
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

	case "enter":
		if m.activePane == PaneInput && m.input.Value() != "" {
			body := m.input.Value()
			log.Debug("send_message", "channel", m.selectedChannel, "body", body)
			m.client.SendMessage(m.selectedChannel, body)
			// Append locally — server doesn't echo message.new to sender
			m.messages.AppendMessage(m.selectedChannel, sharkfin.Message{
				From:   m.username,
				Body:   body,
				SentAt: time.Now(),
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

	// Pass remaining keys to focused component
	if m.activePane == PaneInput {
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

	m.modal.UpdateTextInput(msg)
	return m, nil
}

func (m *ChatModel) submitModal() {
	switch m.modal.Type {
	case ModalChannelCreate:
		m.client.CreateChannel(m.modal.Value(), m.modal.IsPublic(), nil)
	case ModalUserInvite:
		m.client.InviteUser(m.selectedChannel, m.modal.Value())
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

	layout := ui.CalculateChatLayout(m.width, m.height)

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if msg.X < layout.SidebarW {
			if time.Since(m.lastChannelScroll) < 80*time.Millisecond {
				return m, nil
			}
			m.lastChannelScroll = time.Now()
			m.channels.MoveUp()
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
			m.channels.MoveDown()
			m.switchChannel()
		} else {
			m.messages.ScrollDown(3)
		}
		return m, nil

	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
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
			// Click in sidebar: select channel
			// Account for header + border (1) + title line (1)
			row := msg.Y - contentTop - 1 - 1
			if row >= 0 {
				m.channels.SelectIndex(row)
				m.switchChannel()
			}
			return m, nil
		}

		// Click in right pane — check if input area
		inputTop := contentBottom - ui.InputHeight
		if msg.Y >= inputTop && msg.Y < contentBottom {
			m.activePane = PaneInput
			m.input.Focus()
			return m, nil
		}

		// Click in messages area — focus input (natural chat behavior)
		if msg.Y < inputTop {
			m.activePane = PaneInput
			m.input.Focus()
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
	case "ctrl+j":
		m.channels.MoveDown()
		m.switchChannel()
	case "ctrl+k":
		m.channels.MoveUp()
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
	selected := m.channels.Selected()
	if selected != m.selectedChannel {
		log.Debug("switch_channel", "from", m.selectedChannel, "to", selected)
		m.selectedChannel = selected
		m.channels.ClearUnread(selected)
		m.messages.SetChannel(selected)
		m.loadingHistory = false
		// Load recent history if no messages cached for this channel
		if m.messages.OldestID(selected) == 0 {
			m.client.RequestHistory(selected, 0, 50)
		}
		m.client.RequestUnread(selected)
	}
}

func (m *ChatModel) maybeLoadHistory() {
	if m.loadingHistory {
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

func (m *ChatModel) updateLayout() {
	layout := ui.CalculateChatLayout(m.width, m.height)

	// Inner heights subtract 2 for border, then PaneTitleH for the title+gap rendered inside
	sidebarInnerH := layout.ContentH - 2 - ui.PaneTitleH
	msgPaneOuterH := layout.ContentH - ui.InputHeight
	msgInnerH := msgPaneOuterH - 2 - ui.PaneTitleH

	m.channels.SetSize(layout.SidebarW, sidebarInnerH)
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

	layout := ui.CalculateChatLayout(m.width, m.height)

	// Sidebar
	sidebarTitle := ui.RenderPaneTitle(" Channels ", m.activePane == PaneChannels)
	sidebarStyle := ui.CreatePaneStyle(m.activePane == PaneChannels, layout.SidebarW, layout.ContentH)
	sidebar := sidebarStyle.Render(sidebarTitle + "\n" + m.channels.View())

	// Message pane
	chanLabel := "#" + m.selectedChannel
	if m.selectedChannel == "" {
		chanLabel = "No channel"
	}
	channelTitle := ui.RenderPaneTitle(" "+chanLabel+" ", m.activePane == PaneInput)
	msgStyle := ui.CreatePaneStyle(m.activePane == PaneInput, layout.MessageW, layout.ContentH-ui.InputHeight)
	msgPane := msgStyle.Render(channelTitle + "\n" + m.messages.View())

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
		Height(ui.InputHeight - 2)
	inputPane := inputStyle.Render(m.input.View())

	// Compose right side
	rightPane := lipgloss.JoinVertical(lipgloss.Left, msgPane, inputPane)

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
