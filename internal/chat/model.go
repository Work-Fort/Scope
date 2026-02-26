package chat

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

	channels ChannelList
	messages MessagePane
	input    InputBar
	modal    *Modal

	activePane      Pane
	selectedChannel string
	username        string

	lastChannelScroll time.Time
}

// NewModel creates a ChatModel with placeholder data.
func NewModel(username string) ChatModel {
	cl := NewChannelList()
	mp := NewMessagePane()
	ib := NewInputBar()

	// Mark some channels as having unreads for demo
	for i := 0; i < 5; i++ {
		cl.IncrementUnread("engineering")
	}

	return ChatModel{
		channels:        cl,
		messages:        mp,
		input:           ib,
		activePane:      PaneChannels,
		selectedChannel: cl.Selected(),
		username:        username,
	}
}

func (m ChatModel) Init() tea.Cmd {
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
		return m, nil

	case "ctrl+down":
		m.messages.ScrollDown(1)
		return m, nil

	case "enter":
		if m.activePane == PaneInput && m.input.Value() != "" {
			// In Plan 2, this sends via WebSocket
			m.messages.AppendMessage(m.selectedChannel, MessageInfo{
				From:   m.username,
				Body:   m.input.Value(),
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
			// In Plan 2, this sends via WebSocket
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

func (m ChatModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Modal click handling
	if m.modal != nil {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			action := m.modal.HitTest(msg.X, msg.Y)
			switch action {
			case ModalActionSubmit:
				if m.modal.Value() != "" {
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

		// Help bar spans full width — check Y first
		if msg.Y >= layout.ContentH {
			return m.handleHelpBarClick(msg.X)
		}

		if msg.X < layout.SidebarW {
			// Click in sidebar: select channel
			// Account for border (1) + title line (1)
			row := msg.Y - 1 - 1
			if row >= 0 {
				m.channels.SelectIndex(row)
				m.switchChannel()
			}
			return m, nil
		}

		// Click in right pane — check if input area
		inputTop := layout.ContentH - ui.InputHeight
		if msg.Y >= inputTop && msg.Y < layout.ContentH {
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
	case "ctrl+down":
		m.messages.ScrollDown(3)
	}
	return m, nil
}

func (m *ChatModel) switchChannel() {
	selected := m.channels.Selected()
	if selected != m.selectedChannel {
		m.selectedChannel = selected
		m.channels.ClearUnread(selected)
		m.messages.SetChannel(selected)
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
	channelTitle := ui.RenderPaneTitle(" #"+m.selectedChannel+" ", m.activePane == PaneInput)
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
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, rightPane)

	// Help bar
	helpBar := ChatKeyBindings().Render()

	fullUI := lipgloss.JoinVertical(lipgloss.Left, mainContent, helpBar)

	// Modal overlay
	if m.modal != nil {
		return m.modal.View(m.width, m.height)
	}

	// Place constrains output to exact terminal dimensions
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, fullUI)
}
