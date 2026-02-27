package chat

import (
	"strings"

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

func (ib *InputBar) SetReadOnly(ro bool) {
	if ro {
		ib.textinput.Placeholder = "read-only channel"
	} else {
		ib.textinput.Placeholder = "type a message..."
	}
}

func (ib *InputBar) UpdateTextInput(msg interface{}) {
	// Type assert to tea.Msg and update
	ib.textinput, _ = ib.textinput.Update(msg)
}

// TryComplete attempts @mention tab-completion against the given usernames.
// Returns true if a completion was applied.
func (ib *InputBar) TryComplete(usernames []string) bool {
	val := ib.textinput.Value()
	pos := ib.textinput.Position()
	if pos == 0 || len(usernames) == 0 {
		return false
	}

	// Scan backwards from cursor to find @ preceded by whitespace or at position 0
	atIdx := -1
	for i := pos - 1; i >= 0; i-- {
		if val[i] == ' ' {
			break // hit whitespace before finding @
		}
		if val[i] == '@' {
			if i == 0 || val[i-1] == ' ' {
				atIdx = i
			}
			break
		}
	}
	if atIdx < 0 {
		return false
	}

	prefix := strings.ToLower(val[atIdx+1 : pos])

	// Find matching usernames
	var matches []string
	for _, u := range usernames {
		if strings.HasPrefix(strings.ToLower(u), prefix) {
			matches = append(matches, u)
		}
	}
	if len(matches) == 0 {
		return false
	}

	var completion string
	if len(matches) == 1 {
		completion = matches[0] + " "
	} else {
		completion = longestCommonPrefix(matches)
		if len(completion) <= len(prefix) {
			return false // can't extend further
		}
	}

	// Splice completion into the value
	newVal := val[:atIdx+1] + completion + val[pos:]
	ib.textinput.SetValue(newVal)
	ib.textinput.SetCursor(atIdx + 1 + len(completion))
	return true
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strings.ToLower(strs[0])
	for _, s := range strs[1:] {
		s = strings.ToLower(s)
		for len(prefix) > 0 && !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

func (ib InputBar) View() string {
	return ib.textinput.View()
}
