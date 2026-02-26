package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Work-Fort/WorkFort/pkg/ui"
)

type ChannelList struct {
	channels []ChannelInfo
	cursor   int
	width    int
	height   int
	unread   map[string]bool
}

func NewChannelList() ChannelList {
	return ChannelList{
		channels: []ChannelInfo{
			{Name: "general", Public: true},
			{Name: "engineering", Public: true},
			{Name: "ops", Public: true},
			{Name: "random", Public: true},
		},
		unread: make(map[string]bool),
	}
}

func (cl *ChannelList) SetSize(w, h int) {
	cl.width = w
	cl.height = h
}

func (cl *ChannelList) MoveUp() {
	if cl.cursor > 0 {
		cl.cursor--
	}
}

func (cl *ChannelList) MoveDown() {
	if cl.cursor < len(cl.channels)-1 {
		cl.cursor++
	}
}

func (cl *ChannelList) Selected() string {
	if len(cl.channels) == 0 {
		return ""
	}
	return cl.channels[cl.cursor].Name
}

func (cl *ChannelList) SetChannels(channels []ChannelInfo) {
	cl.channels = channels
	if cl.cursor >= len(channels) {
		cl.cursor = max(0, len(channels)-1)
	}
}

func (cl *ChannelList) MarkUnread(channel string) {
	cl.unread[channel] = true
}

func (cl *ChannelList) ClearUnread(channel string) {
	delete(cl.unread, channel)
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
		Foreground(ui.CurrentTheme.Accent).
		Bold(true)

	// Available inner width (pane width minus border chars minus padding)
	innerW := cl.width - 4
	if innerW < 1 {
		innerW = 1
	}

	var lines []string
	for i, ch := range cl.channels {
		prefix := "  "
		name := fmt.Sprintf("# %s", ch.Name)

		if cl.unread[ch.Name] {
			prefix = ui.CurrentTheme.AccentStyle().Render("* ")
		}

		var line string
		if i == cl.cursor {
			line = prefix + selectedStyle.Render(name)
		} else if cl.unread[ch.Name] {
			line = prefix + unreadStyle.Render(name)
		} else {
			line = prefix + normalStyle.Render(name)
		}

		// Truncate if needed
		if lipgloss.Width(line) > innerW {
			line = line[:innerW]
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
