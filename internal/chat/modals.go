package chat

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/Work-Fort/WorkFort/pkg/ui"
)

// ModalType identifies which modal is active.
type ModalType int

const (
	ModalNone ModalType = iota
	ModalChannelCreate
	ModalUserInvite
)

// Modal holds the state for overlay modals.
type Modal struct {
	Type      ModalType
	textinput textinput.Model
	public    bool // for channel create
	err       string
}

func NewModal(modalType ModalType) Modal {
	ti := textinput.New()
	ti.TextStyle = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	ti.CharLimit = 64
	ti.Focus()

	switch modalType {
	case ModalChannelCreate:
		ti.Placeholder = "channel-name"
		ti.Prompt = "  # "
	case ModalUserInvite:
		ti.Placeholder = "username"
		ti.Prompt = "  @ "
	}

	return Modal{
		Type:      modalType,
		textinput: ti,
		public:    true,
	}
}

func (m *Modal) TogglePublic() {
	m.public = !m.public
}

func (m *Modal) Value() string {
	return m.textinput.Value()
}

func (m *Modal) IsPublic() bool {
	return m.public
}

func (m *Modal) SetError(err string) {
	m.err = err
}

func (m *Modal) UpdateTextInput(msg interface{}) {
	m.textinput, _ = m.textinput.Update(msg)
}

func (m Modal) View(totalW, totalH int) string {
	modalW := 44

	titleStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.TextDim)

	errStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E05252"))

	var title string
	var help string

	switch m.Type {
	case ModalChannelCreate:
		title = "Create Channel"
		visStr := "public"
		if !m.public {
			visStr = "private"
		}
		help = "Enter: create  Tab: toggle " + visStr + "  Esc: cancel"
	case ModalUserInvite:
		title = "Invite User"
		help = "Enter: invite  Esc: cancel"
	}

	var content string
	content += titleStyle.Render("  "+title) + "\n\n"
	content += m.textinput.View() + "\n"

	if m.Type == ModalChannelCreate {
		visLabel := "public"
		if !m.public {
			visLabel = "private"
		}
		visStyle := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Secondary)
		content += "\n" + visStyle.Render("  Visibility: "+visLabel) + "\n"
	}

	if m.err != "" {
		content += "\n" + errStyle.Render("  "+m.err) + "\n"
	}

	content += "\n" + helpStyle.Render("  "+help)

	box := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ui.CurrentTheme.Primary).
		Padding(1, 2).
		Width(modalW).
		Render(content)

	return lipgloss.Place(totalW, totalH, lipgloss.Center, lipgloss.Center, box)
}
