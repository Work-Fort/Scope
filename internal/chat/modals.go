package chat

import (
	"strings"

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
	ModalDMOpen
)

// ModalAction identifies a clickable modal button action.
type ModalAction int

const (
	ModalActionNone ModalAction = iota
	ModalActionSubmit
	ModalActionToggle
	ModalActionCancel
)

// modalButton defines a button in the modal help bar.
type modalButton struct {
	key    string
	label  string
	action ModalAction
}

// Modal holds the state for overlay modals.
type Modal struct {
	Type      ModalType
	textinput textinput.Model
	public    bool // for channel create
	err       string
	boxX      int // screen X of the rendered modal box (set during View)
	boxY      int // screen Y of the rendered modal box
	boxW      int
	boxH      int
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
	case ModalDMOpen:
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

func (m Modal) buttons() []modalButton {
	switch m.Type {
	case ModalChannelCreate:
		visLabel := "public "
		if !m.public {
			visLabel = "private"
		}
		return []modalButton{
			{key: "Enter", label: "create", action: ModalActionSubmit},
			{key: "Tab", label: visLabel, action: ModalActionToggle},
			{key: "Esc", label: "cancel", action: ModalActionCancel},
		}
	case ModalUserInvite:
		return []modalButton{
			{key: "Enter", label: "invite", action: ModalActionSubmit},
			{key: "Esc", label: "cancel", action: ModalActionCancel},
		}
	case ModalDMOpen:
		return []modalButton{
			{key: "Enter", label: "open", action: ModalActionSubmit},
			{key: "Esc", label: "cancel", action: ModalActionCancel},
		}
	}
	return nil
}

func renderModalButtons(buttons []modalButton) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.TextDim)

	btnStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.CurrentTheme.Muted).
		PaddingLeft(1).
		PaddingRight(1)

	var rendered []string
	for _, b := range buttons {
		label := keyStyle.Render(b.key) + " " + descStyle.Render(b.label)
		rendered = append(rendered, btnStyle.Render(label))
	}
	spacer := strings.Repeat(" ", ui.PaneGap)
	var parts []string
	for i, r := range rendered {
		parts = append(parts, r)
		if i < len(rendered)-1 {
			parts = append(parts, spacer)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// HitTest checks if screen coordinates (x, y) land on a modal button.
func (m *Modal) HitTest(screenX, screenY int) ModalAction {
	// Modal-relative coordinates, accounting for border (1) + padding (2)
	relX := screenX - m.boxX - 3
	// Button row is at the bottom of the box content
	// boxY + boxH - 1 = bottom border, -3 = button row (border top, content, border bottom)
	btnRowTop := m.boxY + m.boxH - 4
	if screenY < btnRowTop || screenY > btnRowTop+2 {
		return ModalActionNone
	}

	buttons := m.buttons()
	keyStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.TextDim)

	btnStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ui.CurrentTheme.Muted).
		PaddingLeft(1).
		PaddingRight(1)

	x := 0
	for _, b := range buttons {
		label := keyStyle.Render(b.key) + " " + descStyle.Render(b.label)
		w := lipgloss.Width(btnStyle.Render(label))
		if relX >= x && relX < x+w {
			return b.action
		}
		x += w + ui.PaneGap
	}
	return ModalActionNone
}

func (m *Modal) View(totalW, totalH int) string {
	modalW := 54

	titleStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)

	errStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E05252"))

	var title string
	switch m.Type {
	case ModalChannelCreate:
		title = "Create Channel"
	case ModalUserInvite:
		title = "Invite User"
	case ModalDMOpen:
		title = "Direct Message"
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

	content += "\n" + renderModalButtons(m.buttons())

	box := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ui.CurrentTheme.Primary).
		Padding(1, 2).
		Width(modalW).
		Render(content)

	// Store box position for click hit testing
	m.boxW = lipgloss.Width(box)
	m.boxH = lipgloss.Height(box)
	m.boxX = (totalW - m.boxW) / 2
	m.boxY = (totalH - m.boxH) / 2

	return lipgloss.Place(totalW, totalH, lipgloss.Center, lipgloss.Center, box)
}
