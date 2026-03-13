package chat

import (
	"fmt"
	"strings"

	"github.com/Work-Fort/Scope/pkg/sharkfin"
	"github.com/Work-Fort/Scope/pkg/ui"
)

type ChannelList struct {
	channels     []sharkfin.Channel
	cursor       int
	scrollOffset int
	width        int
	height       int
	unread       map[string]int
	mentions     map[string]int
}

func NewChannelList() ChannelList {
	return ChannelList{
		unread:   make(map[string]int),
		mentions: make(map[string]int),
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

func (cl *ChannelList) IsMember() bool {
	if cl.cursor < 0 || cl.cursor >= len(cl.channels) {
		return false
	}
	return cl.channels[cl.cursor].Member
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

func (cl *ChannelList) IncrementMention(channel string) {
	cl.unread[channel]++
	cl.mentions[channel]++
}

func (cl *ChannelList) SetCounts(channel string, unread, mentions int) {
	if unread > 0 {
		cl.unread[channel] = unread
	} else {
		delete(cl.unread, channel)
	}
	if mentions > 0 {
		cl.mentions[channel] = mentions
	} else {
		delete(cl.mentions, channel)
	}
}

func (cl *ChannelList) ClearUnread(channel string) {
	delete(cl.unread, channel)
	delete(cl.mentions, channel)
}

func (cl *ChannelList) HasUnreads() bool {
	return len(cl.unread) > 0
}

// Names returns all channel names.
func (cl *ChannelList) Names() []string {
	names := make([]string, len(cl.channels))
	for i, ch := range cl.channels {
		names[i] = ch.Name
	}
	return names
}

// FindByName returns the channel and true if found, or zero value and false.
func (cl *ChannelList) FindByName(name string) (sharkfin.Channel, bool) {
	for _, ch := range cl.channels {
		if ch.Name == name {
			return ch, true
		}
	}
	return sharkfin.Channel{}, false
}

// SelectByName selects a channel by name. Returns true if found.
func (cl *ChannelList) SelectByName(name string) bool {
	for i, ch := range cl.channels {
		if ch.Name == name {
			cl.cursor = i
			cl.ensureCursorVisible()
			return true
		}
	}
	return false
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

	selectedStyle := ui.CurrentTheme.ListSelectedStyle()
	normalStyle := ui.CurrentTheme.ListNormalStyle()
	unreadStyle := ui.CurrentTheme.ListUnreadStyle()
	mentionStyle := ui.CurrentTheme.ListMentionStyle()

	// Available inner width (pane width minus border chars minus padding)
	innerW := cl.width - 4
	if innerW < 1 {
		innerW = 1
	}

	unreadDot := ui.CurrentTheme.UnreadDot()
	unreadCountStyle := ui.CurrentTheme.UnreadCountStyle()
	mentionDot := ui.CurrentTheme.MentionDot()
	mentionCountStyle := ui.CurrentTheme.MentionCountStyle()

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

		unreadCount := cl.unread[ch.Name]
		mentionCount := cl.mentions[ch.Name]
		hasMention := mentionCount > 0 && i != cl.cursor
		hasUnread := unreadCount > 0 && i != cl.cursor

		if hasMention {
			prefix = mentionDot
		} else if hasUnread {
			prefix = unreadDot
		}

		// Truncate name before styling to avoid cutting ANSI codes
		// prefix is 2 chars, leave room for suffix
		maxNameW := innerW - 2
		suffix := ""
		if hasMention {
			badge := fmt.Sprintf(" @%d", mentionCount)
			maxNameW -= len(badge)
			suffix = badge
		} else if hasUnread {
			badge := fmt.Sprintf(" (%d)", unreadCount)
			maxNameW -= len(badge)
			suffix = badge
		}
		if len(name) > maxNameW {
			name = name[:maxNameW-1] + "…"
		}

		var line string
		if i == cl.cursor {
			line = prefix + selectedStyle.Render(name)
		} else if hasMention {
			line = prefix + mentionStyle.Render(name) + " " + mentionCountStyle.Render(suffix)
		} else if hasUnread {
			line = prefix + unreadStyle.Render(name) + " " + unreadCountStyle.Render(suffix)
		} else {
			line = prefix + normalStyle.Render(name)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
