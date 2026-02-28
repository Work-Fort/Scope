package chat

import (
	"strings"
	"unicode"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"

	"github.com/Work-Fort/WorkFort/pkg/clipboard"
	"github.com/Work-Fort/WorkFort/pkg/ui"
)

const maxInputLines = 8

type InputBar struct {
	textarea  textarea.Model
	width     int
	sttActive bool // true while STT is replacing text
	sttStart  int  // rune position where STT text begins
}

func NewInputBar() InputBar {
	ta := textarea.New()
	ta.Placeholder = "type a message..."
	styles := textarea.DefaultDarkStyles()
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(ui.CurrentTheme.TextDim)
	styles.Focused.Base = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	styles.Blurred.Base = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	styles.Focused.CursorLine = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	styles.Blurred.CursorLine = lipgloss.NewStyle().Foreground(ui.CurrentTheme.Text)
	ta.SetStyles(styles)
	ta.Prompt = " "
	ta.CharLimit = 2000
	ta.ShowLineNumbers = false
	ta.SetHeight(1)
	ta.KeyMap.InsertNewline.SetEnabled(false) // we handle newline manually
	ta.KeyMap.Paste.SetEnabled(false)         // we handle paste with Wayland support

	// Add ctrl+arrow word navigation (textarea only has alt+arrow by default)
	ta.KeyMap.WordForward = key.NewBinding(key.WithKeys("alt+right", "ctrl+right", "alt+f"))
	ta.KeyMap.WordBackward = key.NewBinding(key.WithKeys("alt+left", "ctrl+left", "alt+b"))

	// Add ctrl+backspace/delete word delete (textarea only has alt+backspace/delete by default)
	ta.KeyMap.DeleteWordBackward = key.NewBinding(key.WithKeys("alt+backspace", "ctrl+backspace", "ctrl+w"))
	ta.KeyMap.DeleteWordForward = key.NewBinding(key.WithKeys("alt+delete", "ctrl+delete", "alt+d"))

	return InputBar{
		textarea: ta,
	}
}

func (ib *InputBar) SetWidth(w int) {
	ib.width = w
	ib.textarea.SetWidth(w - 1) // caller subtracted border; -1 for right margin
	ib.updateHeight()           // recalc in case width changed wrapping
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

func (ib *InputBar) Paste() {
	text, err := clipboard.Read()
	if err != nil || text == "" {
		return
	}
	ib.textarea.SetHeight(maxInputLines)
	ib.textarea.InsertString(text)
	ib.updateHeight()
}

func (ib *InputBar) InsertNewline() {
	ib.textarea.SetHeight(maxInputLines)
	ib.textarea.InsertString("\n")
	ib.updateHeight()
}

func (ib *InputBar) InsertString(s string) {
	ib.textarea.SetHeight(maxInputLines)
	ib.textarea.InsertString(s)
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

// visualLineCount returns the total number of visual lines by replicating
// the textarea's word-wrap algorithm. This must match the textarea's internal
// wrap() function exactly to avoid height mismatches.
func (ib *InputBar) visualLineCount() int {
	val := ib.textarea.Value()
	if val == "" {
		return 1
	}

	// The textarea's internal m.width = SetWidth_arg - promptWidth - baseFrameSize.
	// With SetWidth(ib.width - 1), prompt " " (width 1), no base frame:
	// m.width = (ib.width - 1) - 1 - 0 = ib.width - 2
	wrapW := ib.width - 2
	if wrapW < 1 {
		wrapW = 1
	}

	total := 0
	for _, line := range strings.Split(val, "\n") {
		total += wordWrapLineCount([]rune(line), wrapW)
	}
	return total
}

// wordWrapLineCount replicates the textarea's wrap() to count visual lines.
func wordWrapLineCount(runes []rune, width int) int {
	if len(runes) == 0 {
		return 1
	}

	var (
		lines  = 1
		lineW  int
		word   []rune
		spaces int
	)

	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 {
			wordW := uniseg.StringWidth(string(word))
			if lineW+wordW+spaces > width {
				lines++
				lineW = wordW + spaces
			} else {
				lineW += wordW + spaces
			}
			spaces = 0
			word = nil
		} else {
			lastCharLen := rw.RuneWidth(word[len(word)-1])
			wordW := uniseg.StringWidth(string(word))
			if wordW+lastCharLen > width {
				if lineW > 0 {
					lines++
				}
				lineW = wordW
				word = nil
			}
		}
	}

	// Final check uses >= (matching textarea behavior)
	wordW := uniseg.StringWidth(string(word))
	if lineW+wordW+spaces >= width {
		lines++
	}

	return lines
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
	if row >= len(lines) {
		return false
	}
	lineText := lines[row]
	if col > len(lineText) {
		col = len(lineText)
	}
	if col == 0 {
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
	ib.textarea.SetCursorColumn(atIdx + 1 + len(completion))
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

// BeginSTT marks the current cursor position as the start of STT text.
// All text from this position onward will be replaced by transcription updates.
func (ib *InputBar) BeginSTT() {
	ib.sttActive = true
	ib.sttStart = len([]rune(ib.textarea.Value()))
}

// SetSTTText replaces the STT portion of the input (from sttStart onward)
// with the given transcription text, preserving any prefix typed before
// recording started.
func (ib *InputBar) SetSTTText(text string) {
	if !ib.sttActive {
		return
	}
	runes := []rune(ib.textarea.Value())
	if ib.sttStart > len(runes) {
		ib.sttStart = len(runes)
	}
	prefix := string(runes[:ib.sttStart])
	newVal := prefix + text
	ib.textarea.SetValue(newVal) // SetValue moves cursor to end
	ib.updateHeight()
}

// ClearSTTState resets the STT tracking so the input bar accepts normal typing.
func (ib *InputBar) ClearSTTState() {
	ib.sttActive = false
	ib.sttStart = 0
}

// IsSTTActive reports whether the input bar is in STT mode.
func (ib *InputBar) IsSTTActive() bool {
	return ib.sttActive
}

func (ib InputBar) View() string {
	return ib.textarea.View()
}
