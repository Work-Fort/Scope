package chat

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/Work-Fort/WorkFort/pkg/audio"
	"github.com/Work-Fort/WorkFort/pkg/ui"
)

// ModalType identifies which modal is active.
type ModalType int

const (
	ModalNone ModalType = iota
	ModalChannelCreate
	ModalUserInvite
	ModalDMOpen
	ModalShortcuts
	ModalSettings
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
	// Settings state
	soundCursor int           // index into audio.AllSounds()
	soundChoice audio.Sound   // currently selected sound
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

	m := Modal{
		Type:      modalType,
		textinput: ti,
		public:    true,
	}
	return m
}

// NewSettingsModal creates a settings modal with the given current sound.
func NewSettingsModal(currentSound audio.Sound) Modal {
	m := NewModal(ModalSettings)
	m.soundChoice = currentSound
	sounds := audio.AllSounds()
	for i, s := range sounds {
		if s == currentSound {
			m.soundCursor = i
			break
		}
	}
	return m
}

// SoundCursorUp moves the sound selector up.
func (m *Modal) SoundCursorUp() {
	if m.soundCursor > 0 {
		m.soundCursor--
	}
}

// SoundCursorDown moves the sound selector down.
func (m *Modal) SoundCursorDown() {
	sounds := audio.AllSounds()
	if m.soundCursor < len(sounds)-1 {
		m.soundCursor++
	}
}

// SelectedSound returns the sound at the current cursor.
func (m *Modal) SelectedSound() audio.Sound {
	sounds := audio.AllSounds()
	if m.soundCursor < len(sounds) {
		return sounds[m.soundCursor]
	}
	return audio.SoundTone
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
	case ModalShortcuts:
		return []modalButton{
			{key: "Esc", label: "close", action: ModalActionCancel},
		}
	case ModalSettings:
		return []modalButton{
			{key: "Enter", label: "save", action: ModalActionSubmit},
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

// SoundHitTest checks if screen coordinates land on a sound selector item.
// Returns the index (0-based) or -1 if no hit.
func (m *Modal) SoundHitTest(screenX, screenY int) int {
	if m.Type != ModalSettings {
		return -1
	}
	// Sound items start after: border(1) + padding(1) + title(1) + blank(1) + label(1) + blank(1) = 6 rows
	soundStart := m.boxY + 6
	sounds := audio.AllSounds()
	idx := screenY - soundStart
	if idx >= 0 && idx < len(sounds) && screenX >= m.boxX && screenX < m.boxX+m.boxW {
		return idx
	}
	return -1
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
	if m.Type == ModalShortcuts {
		modalW = 60
	}

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
	case ModalShortcuts:
		title = "Keyboard Shortcuts"
	case ModalSettings:
		title = "Settings"
	}

	var content string
	content += titleStyle.Render("  "+title) + "\n\n"

	if m.Type == ModalShortcuts {
		content += renderShortcuts()
	} else if m.Type == ModalSettings {
		content += m.renderSoundSelector()
	} else {
		content += m.textinput.View() + "\n"
	}

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

func (m *Modal) renderSoundSelector() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Secondary).
		Bold(true)
	selectedStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Text)

	var lines []string
	lines = append(lines, "  "+labelStyle.Render("Notification Sound"))
	lines = append(lines, "")

	sounds := audio.AllSounds()
	for i, s := range sounds {
		cursor := "  "
		style := normalStyle
		if i == m.soundCursor {
			cursor = "  "
			style = selectedStyle
		}
		label := s.Label()
		if i == m.soundCursor {
			label = "> " + label
		} else {
			label = "  " + label
		}
		lines = append(lines, cursor+style.Render(label))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n") + "\n"
}

func renderShortcuts() string {
	groups := AllShortcuts()

	groupStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Secondary).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true).
		Width(24)

	descStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Text)

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
