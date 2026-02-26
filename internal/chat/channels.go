package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Work-Fort/WorkFort/pkg/sharkfin"
	"github.com/Work-Fort/WorkFort/pkg/ui"
)

type ChannelList struct {
	channels     []sharkfin.Channel
	cursor       int
	scrollOffset int
	width        int
	height       int
	unread       map[string]int
}

func NewChannelList() ChannelList {
	return ChannelList{
		unread: make(map[string]int),
	}
}

func (cl *ChannelList) SetSize(w, h int) {
	cl.width = w
	cl.height = h
}

func (cl *ChannelList) MoveUp() {
	if cl.cursor > 0 {
		cl.cursor--
		cl.ensureCursorVisible()
	}
}

func (cl *ChannelList) MoveDown() {
	if cl.cursor < len(cl.channels)-1 {
		cl.cursor++
		cl.ensureCursorVisible()
	}
}

func (cl *ChannelList) ensureCursorVisible() {
	if cl.height <= 0 {
		return
	}
	if cl.cursor < cl.scrollOffset {
		cl.scrollOffset = cl.cursor
	}
	if cl.cursor >= cl.scrollOffset+cl.height {
		cl.scrollOffset = cl.cursor - cl.height + 1
	}
}

func (cl *ChannelList) Selected() string {
	if len(cl.channels) == 0 {
		return ""
	}
	return cl.channels[cl.cursor].Name
}

func (cl *ChannelList) SetChannels(channels []sharkfin.Channel) {
	cl.channels = channels
	if cl.cursor >= len(channels) {
		cl.cursor = max(0, len(channels)-1)
	}
}

func (cl *ChannelList) IncrementUnread(channel string) {
	cl.unread[channel]++
}

func (cl *ChannelList) ClearUnread(channel string) {
	delete(cl.unread, channel)
}

func (cl *ChannelList) SelectIndex(i int) {
	idx := i + cl.scrollOffset
	if idx >= 0 && idx < len(cl.channels) {
		cl.cursor = idx
		cl.ensureCursorVisible()
	}
}

func (cl ChannelList) View() string {
	if len(cl.channels) == 0 {
		return ui.CurrentTheme.TextDimStyle().Render("  No channels")
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Text)

	unreadStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Text).
		Bold(true)

	// Available inner width (pane width minus border chars minus padding)
	innerW := cl.width - 4
	if innerW < 1 {
		innerW = 1
	}

	unreadDot := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Accent).Render("● ")
	countStyle := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Accent)

	// Determine visible window
	visEnd := cl.scrollOffset + cl.height
	if visEnd > len(cl.channels) {
		visEnd = len(cl.channels)
	}
	visible := cl.channels[cl.scrollOffset:visEnd]

	var lines []string
	for j, ch := range visible {
		i := cl.scrollOffset + j
		prefix := "  "
		name := "#" + ch.Name

		count := cl.unread[ch.Name]
		if count > 0 && i != cl.cursor {
			prefix = unreadDot
		}

		// Truncate name before styling to avoid cutting ANSI codes
		// prefix is 2 chars, leave room for suffix
		maxNameW := innerW - 2
		suffix := ""
		if count > 0 && i != cl.cursor {
			badge := fmt.Sprintf(" (%d)", count)
			maxNameW -= len(badge)
			suffix = badge
		}
		if len(name) > maxNameW {
			name = name[:maxNameW-1] + "…"
		}

		var line string
		if i == cl.cursor {
			line = prefix + selectedStyle.Render(name)
		} else if count > 0 {
			line = prefix + unreadStyle.Render(name) + " " + countStyle.Render(suffix)
		} else {
			line = prefix + normalStyle.Render(name)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
