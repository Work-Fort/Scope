package chat

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/Work-Fort/WorkFort/pkg/ui"
)

const maxInputLines = 8

type InputBar struct {
	textarea textarea.Model
	width    int
}

func NewInputBar() InputBar {
	ta := textarea.New()
	ta.Placeholder = "type a message..."
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	ta.FocusedStyle.Base = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	ta.BlurredStyle.Base = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)
	ta.Prompt = "> "
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	ta.CharLimit = 2000
	ta.ShowLineNumbers = false
	ta.SetHeight(1)
	ta.KeyMap.InsertNewline.SetEnabled(false) // we handle newline manually

	// Add ctrl+arrow word navigation (textarea only has alt+arrow by default)
	ta.KeyMap.WordForward = key.NewBinding(key.WithKeys("alt+right", "ctrl+right", "alt+f"))
	ta.KeyMap.WordBackward = key.NewBinding(key.WithKeys("alt+left", "ctrl+left", "alt+b"))

	return InputBar{
		textarea: ta,
	}
}

func (ib *InputBar) SetWidth(w int) {
	ib.width = w
	ib.textarea.SetWidth(w - 4) // border + padding
}

func (ib *InputBar) Focus() {
	ib.textarea.Focus()
}

func (ib *InputBar) Blur() {
	ib.textarea.Blur()
}

func (ib *InputBar) Focused() bool {
	return ib.textarea.Focused()
}

func (ib *InputBar) Value() string {
	return ib.textarea.Value()
}

func (ib *InputBar) Reset() {
	ib.textarea.Reset()
	ib.textarea.SetHeight(1)
}

func (ib *InputBar) SetReadOnly(ro bool) {
	if ro {
		ib.textarea.Placeholder = "read-only channel"
	} else {
		ib.textarea.Placeholder = "type a message..."
	}
}

func (ib *InputBar) InsertNewline() {
	ib.textarea.SetHeight(maxInputLines)
	ib.textarea.InsertString("\n")
	ib.updateHeight()
}

// Height returns the total height including borders (2 for border).
func (ib *InputBar) Height() int {
	return ib.textarea.Height() + 2
}

func (ib *InputBar) updateHeight() {
	lines := ib.visualLineCount()
	if lines < 1 {
		lines = 1
	}
	if lines > maxInputLines {
		lines = maxInputLines
	}
	ib.textarea.SetHeight(lines)
}

// visualLineCount returns the total number of visual lines accounting for
// word wrap. Each logical line may occupy multiple visual lines when its
// display width exceeds the textarea's internal wrap width.
func (ib *InputBar) visualLineCount() int {
	val := ib.textarea.Value()
	if val == "" {
		return 1
	}

	// The textarea wraps at: SetWidth arg (ib.width - 4) minus prompt width (2).
	wrapW := ib.width - 6
	if wrapW < 1 {
		wrapW = 1
	}

	total := 0
	for _, line := range strings.Split(val, "\n") {
		w := lipgloss.Width(line)
		if w <= wrapW {
			total++
		} else {
			total += (w + wrapW - 1) / wrapW // ceil(w / wrapW)
		}
	}
	return total
}

func (ib *InputBar) UpdateTextInput(msg interface{}) {
	// Temporarily expand to max so the textarea's internal viewport never
	// scrolls during Update (repositionView uses the current height).
	// After Update, shrink back to the actual visual line count.
	ib.textarea.SetHeight(maxInputLines)
	ib.textarea, _ = ib.textarea.Update(msg)
	ib.updateHeight()
}

// TryComplete attempts @mention tab-completion against the given usernames.
// Works on the current line at the cursor position.
func (ib *InputBar) TryComplete(usernames []string) bool {
	val := ib.textarea.Value()
	if val == "" || len(usernames) == 0 {
		return false
	}

	// Get current line and column.
	// LineInfo().ColumnOffset is the visual column within the current
	// wrapped sub-line. When a line soft-wraps, we must add back the
	// width of all preceding visual sub-lines to get the logical offset.
	lines := strings.Split(val, "\n")
	row := ib.textarea.Line()
	li := ib.textarea.LineInfo()
	col := li.ColumnOffset
	if li.RowOffset > 0 && li.Width > 0 {
		col += li.RowOffset * li.Width
	}
	log.Debug("tab_complete", "val", val, "row", row, "lines", len(lines),
		"col", col, "charOffset", li.CharOffset, "colOffset", li.ColumnOffset,
		"rowOffset", li.RowOffset, "width", li.Width)
	if row >= len(lines) {
		log.Debug("tab_complete_bail", "reason", "row>=lines", "row", row, "lines", len(lines))
		return false
	}
	lineText := lines[row]
	if col > len(lineText) {
		col = len(lineText)
	}
	if col == 0 {
		log.Debug("tab_complete_bail", "reason", "col==0")
		return false
	}

	// Scan backwards from cursor to find @ preceded by whitespace or at position 0
	atIdx := -1
	for i := col - 1; i >= 0; i-- {
		if lineText[i] == ' ' {
			break
		}
		if lineText[i] == '@' {
			if i == 0 || lineText[i-1] == ' ' {
				atIdx = i
			}
			break
		}
	}
	if atIdx < 0 {
		log.Debug("tab_complete_bail", "reason", "no_@", "lineText", lineText, "col", col)
		return false
	}

	prefix := strings.ToLower(lineText[atIdx+1 : col])

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
			return false
		}
	}

	// Rebuild the line with the completion
	newLine := lineText[:atIdx+1] + completion + lineText[col:]
	lines[row] = newLine
	newVal := strings.Join(lines, "\n")
	ib.textarea.SetValue(newVal)
	// SetValue resets cursor to end; reposition to after completion
	// Navigate to the correct row then set column
	for i := len(lines) - 1; i > row; i-- {
		ib.textarea.CursorUp()
	}
	ib.textarea.SetCursor(atIdx + 1 + len(completion))
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
	return ib.textarea.View()
}
