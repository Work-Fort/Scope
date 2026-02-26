package chat

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/Work-Fort/WorkFort/pkg/ui"
)

type InputBar struct {
	textinput textinput.Model
	width     int
}

func NewInputBar() InputBar {
	ti := textinput.New()
	ti.Placeholder = "type a message..."
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	ti.TextStyle = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)
	ti.Prompt = "  > "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)
	ti.CharLimit = 1000

	return InputBar{
		textinput: ti,
	}
}

func (ib *InputBar) SetWidth(w int) {
	ib.width = w
	ib.textinput.Width = w - 8 // border + prompt + padding
}

func (ib *InputBar) Focus() {
	ib.textinput.Focus()
}

func (ib *InputBar) Blur() {
	ib.textinput.Blur()
}

func (ib *InputBar) Focused() bool {
	return ib.textinput.Focused()
}

func (ib *InputBar) Value() string {
	return ib.textinput.Value()
}

func (ib *InputBar) Reset() {
	ib.textinput.Reset()
}

func (ib *InputBar) UpdateTextInput(msg interface{}) {
	// Type assert to tea.Msg and update
	ib.textinput, _ = ib.textinput.Update(msg)
}

func (ib InputBar) View() string {
	return ib.textinput.View()
}
