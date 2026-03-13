package chat

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/Work-Fort/Scope/pkg/ui"
)

// ModalType identifies which modal is active.
type ModalType int

const (
	ModalNone ModalType = iota
	ModalChannelCreate
	ModalUserInvite
	ModalDMOpen
	ModalShortcuts
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
	hint      string
	boxX      int // screen X of the rendered modal box (set during View)
	boxY      int // screen Y of the rendered modal box
	boxW      int
	boxH      int
}

func NewModal(modalType ModalType) Modal {
	ti := textinput.New()
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	styles.Blurred.Text = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	styles.Cursor.Color = ui.CurrentTheme.Primary
	ti.SetStyles(styles)
	ti.KeyMap.DeleteWordBackward = key.NewBinding(key.WithKeys("alt+backspace", "ctrl+backspace", "ctrl+w"))
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

	m := Modal{
		Type:      modalType,
		textinput: ti,
		public:    true,
	}
	return m
}

func (m *Modal) TogglePublic() {
	m.public = !m.public
}

func (m *Modal) Value() string {
	return m.textinput.Value()
}

func (m *Modal) SetValue(v string) {
	m.textinput.SetValue(v)
	m.textinput.CursorEnd()
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
		return []modalButton{
			{key: "Enter", label: "go", action: ModalActionSubmit},
			{key: "Tab", label: "complete", action: ModalActionToggle},
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
	case ModalShortcuts:
		return []modalButton{
			{key: "Esc", label: "close", action: ModalActionCancel},
		}
	}
	return nil
}

func renderModalButtons(buttons []modalButton) string {
	keyStyle := ui.CurrentTheme.KeyStyle()
	descStyle := ui.CurrentTheme.KeyDescStyle()
	btnStyle := ui.CurrentTheme.ButtonStyle()

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
	keyStyle := ui.CurrentTheme.KeyStyle()
	descStyle := ui.CurrentTheme.KeyDescStyle()
	btnStyle := ui.CurrentTheme.ButtonStyle()

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
	modalW := 56 // v2: Width includes border (54 content + 2 border)
	if m.Type == ModalShortcuts {
		modalW = 62
	}

	titleStyle := ui.CurrentTheme.ModalTitleStyle()
	errStyle := ui.CurrentTheme.ErrorStyle()

	var title string
	switch m.Type {
	case ModalChannelCreate:
		title = "Go to Channel"
	case ModalUserInvite:
		title = "Invite User"
	case ModalDMOpen:
		title = "Direct Message"
	case ModalShortcuts:
		title = "Keyboard Shortcuts"
	}

	var content string
	content += titleStyle.Render("  "+title) + "\n\n"

	if m.Type == ModalShortcuts {
		content += renderShortcuts()
	} else {
		content += m.textinput.View() + "\n"
	}

	if m.Type == ModalChannelCreate && m.hint != "" {
		hintStyle := ui.CurrentTheme.ModalHintStyle()
		content += "\n" + hintStyle.Render("  "+m.hint) + "\n"
	}

	if m.err != "" {
		content += "\n" + errStyle.Render("  "+m.err) + "\n"
	}

	content += "\n" + renderModalButtons(m.buttons())

	box := ui.CurrentTheme.ModalBoxStyle(modalW).Render(content)

	// Store box position for click hit testing
	m.boxW = lipgloss.Width(box)
	m.boxH = lipgloss.Height(box)
	m.boxX = (totalW - m.boxW) / 2
	m.boxY = (totalH - m.boxH) / 2

	return lipgloss.Place(totalW, totalH, lipgloss.Center, lipgloss.Center, box)
}

func renderShortcuts() string {
	groups := AllShortcuts()

	groupStyle := ui.CurrentTheme.ModalLabelStyle()
	keyStyle := ui.CurrentTheme.KeyStyle().Width(24)
	descStyle := ui.CurrentTheme.TextStyle()

	var lines []string
	for i, g := range groups {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "  "+groupStyle.Render(g.Title))
		for _, item := range g.Items {
			lines = append(lines, "  "+keyStyle.Render(item.Key)+descStyle.Render(item.Description))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
