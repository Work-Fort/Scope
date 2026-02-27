package chat

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// textareaWrap is copied directly from bubbles/textarea/textarea.go wrap()
// to compare against our wordWrapLineCount.
func textareaWrap(runes []rune, width int) [][]rune {
	var (
		lines  = [][]rune{{}}
		word   = []rune{}
		row    int
		spaces int
	)

	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 {
			if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces > width {
				row++
				lines = append(lines, []rune{})
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			} else {
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatSpaces(spaces)...)
				spaces = 0
				word = nil
			}
		} else {
			lastCharLen := rw.RuneWidth(word[len(word)-1])
			wordW := uniseg.StringWidth(string(word))
			if wordW+lastCharLen > width {
				if len(lines[row]) > 0 {
					row++
					lines = append(lines, []rune{})
				}
				lines[row] = append(lines[row], word...)
				word = nil
			}
		}
	}

	if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces >= width {
		lines = append(lines, []rune{})
		lines[row+1] = append(lines[row+1], word...)
		spaces++
		lines[row+1] = append(lines[row+1], repeatSpaces(spaces)...)
	} else {
		lines[row] = append(lines[row], word...)
		spaces++
		lines[row] = append(lines[row], repeatSpaces(spaces)...)
	}

	return lines
}

func repeatSpaces(n int) []rune {
	return []rune(strings.Repeat(string(' '), n))
}

// TestWordWrapLineCountMatchesTextarea types text character by character
// and checks that wordWrapLineCount matches the textarea's wrap() line count.
func TestWordWrapLineCountMatchesTextarea(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
	}{
		{"short text", "hello world", 61},
		{"near wrap", "the quick brown fox jumps over the lazy dog and then some more", 61},
		{"exact wrap", strings.Repeat("a", 61), 61},
		{"over wrap", strings.Repeat("a", 62), 61},
		{"words near boundary", "hello world this is a test of the wrapping algorithm at the boundary point here", 61},
		{"long word", "superlongwordthatdoesnotcontainanyspaces and then more text", 20},
		{"multiple spaces", "hello   world   test", 61},
		{"real width 61", "Build and test your builds before before teling me to test.", 61},
		{"progressive typing", "this is a normal sentence that should wrap around the line boundary correctly", 61},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.text)

			// Test character by character (simulating typing)
			for i := 1; i <= len(runes); i++ {
				partial := runes[:i]
				partialStr := string(partial)

				// Get textarea's wrap result
				wrapResult := textareaWrap(partial, tt.width)
				expectedLines := len(wrapResult)

				// Get our count
				ourCount := wordWrapLineCount(partial, tt.width)

				if ourCount != expectedLines {
					t.Errorf("mismatch at char %d (text=%q): textarea wrap=%d, our count=%d",
						i, partialStr, expectedLines, ourCount)
				}
			}
		})
	}
}

// TestWordWrapLineCountEdgeCases tests specific edge cases.
func TestWordWrapLineCountEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected int // expected from textarea's wrap
	}{
		{"empty", "", 61, 1},
		{"single char", "a", 61, 1},
		{"exactly width", strings.Repeat("a", 61), 61, 2}, // >= check at end wraps
		{"width minus 1", strings.Repeat("a", 60), 61, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.text)

			wrapResult := textareaWrap(runes, tt.width)
			textareaLines := len(wrapResult)

			if textareaLines != tt.expected {
				t.Errorf("textarea wrap(%q, %d) = %d lines, expected %d",
					tt.text, tt.width, textareaLines, tt.expected)
			}

			ourCount := wordWrapLineCount(runes, tt.width)
			if ourCount != textareaLines {
				t.Errorf("wordWrapLineCount(%q, %d) = %d, textarea wrap = %d",
					tt.text, tt.width, ourCount, textareaLines)
			}
		})
	}
}

// TestInputBarHeightDuringTyping creates a real InputBar and simulates
// typing character by character, checking that the height stays correct.
func TestInputBarHeightDuringTyping(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		setWidth int // the w passed to SetWidth
	}{
		{
			"normal sentence width 63",
			"the quick brown fox jumps over the lazy dog and then continues with more text here to test wrapping",
			63,
		},
		{
			"tpm's real message",
			"Build and test your builds before before teling me to test.",
			63,
		},
		{
			"words at boundary",
			"hello world this is a test of the wrapping algorithm at the exact boundary",
			63,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := NewInputBar()
			ib.SetWidth(tt.setWidth)
			ib.Focus()

			wrapW := tt.setWidth - 2 // match visualLineCount calculation

			for i, r := range tt.text {
				// Simulate typing via UpdateTextInput
				msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
				ib.UpdateTextInput(msg)

				// Check: the textarea's Value should have i+1 characters
				val := ib.Value()
				if len(val) != i+1 {
					t.Fatalf("char %d: expected value length %d, got %d (value=%q)", i, i+1, len(val), val)
				}

				// Compute expected height from textarea's wrap
				expectedTotal := 0
				for _, line := range strings.Split(val, "\n") {
					wrapResult := textareaWrap([]rune(line), wrapW)
					expectedTotal += len(wrapResult)
				}
				if expectedTotal < 1 {
					expectedTotal = 1
				}
				if expectedTotal > maxInputLines {
					expectedTotal = maxInputLines
				}

				actualHeight := ib.textarea.Height()

				if actualHeight != expectedTotal {
					t.Errorf("char %d (%c): height=%d, expected=%d (value=%q, wrapW=%d)",
						i, r, actualHeight, expectedTotal, val, wrapW)
				}
			}
		})
	}
}

// TestInputBarWrappingWidth verifies that the wrapW used in visualLineCount
// matches the textarea's internal m.width.
func TestInputBarWrappingWidth(t *testing.T) {
	ib := NewInputBar()
	ib.SetWidth(63)
	ib.Focus()

	// Type enough text to trigger wrapping
	text := "the quick brown fox jumps over the lazy dog and then more text"
	for _, r := range text {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		ib.UpdateTextInput(msg)
	}

	// The textarea's internal width should be SetWidth_arg - promptWidth
	// = 62 - 1 = 61 (prompt is " " with width 1)
	// Our wrapW should also be 61
	wrapW := ib.width - 2 // 63 - 2 = 61

	// Verify by checking that the textarea's LineInfo matches expectations
	li := ib.textarea.LineInfo()
	t.Logf("LineInfo: Width=%d, ColumnOffset=%d, RowOffset=%d, CharOffset=%d",
		li.Width, li.ColumnOffset, li.RowOffset, li.CharOffset)
	t.Logf("ib.width=%d, wrapW=%d, textarea.Height()=%d", ib.width, wrapW, ib.textarea.Height())
	t.Logf("Value length=%d, value=%q", len(ib.Value()), ib.Value())

	// The Width field in LineInfo is the rune count of the current wrapped line,
	// NOT the wrapping width. But for a line that wraps, the max Width should be
	// close to the wrapping width.
	visualLines := ib.visualLineCount()
	textareaLines := 0
	for _, line := range strings.Split(ib.Value(), "\n") {
		textareaLines += len(textareaWrap([]rune(line), wrapW))
	}
	if visualLines != textareaLines {
		t.Errorf("visualLineCount=%d, textareaWrap count=%d", visualLines, textareaLines)
	}

	fmt.Printf("  Height=%d, visualLines=%d, textareaWrap=%d\n", ib.textarea.Height(), visualLines, textareaLines)
}
