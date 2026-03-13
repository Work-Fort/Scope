package chat

import (
	"fmt"
	"strings"

	"github.com/Work-Fort/Scope/pkg/sharkfin"
	"github.com/Work-Fort/Scope/pkg/ui"
)

type DMList struct {
	dms          []sharkfin.DM
	cursor       int
	scrollOffset int
	width        int
	height       int
	unread       map[string]int
	mentions     map[string]int
	username     string // current user, used to derive "other" participant
}

func NewDMList(username string) DMList {
	return DMList{
		unread:   make(map[string]int),
		mentions: make(map[string]int),
		username: username,
	}
}

func (dl *DMList) SetSize(w, h int) {
	dl.width = w
	dl.height = h
}

func (dl *DMList) MoveUp() {
	if dl.cursor > 0 {
		dl.cursor--
		dl.ensureCursorVisible()
	}
}

func (dl *DMList) MoveDown() {
	if dl.cursor < len(dl.dms)-1 {
		dl.cursor++
		dl.ensureCursorVisible()
	}
}

func (dl *DMList) ensureCursorVisible() {
	if dl.height <= 0 {
		return
	}
	if dl.cursor < dl.scrollOffset {
		dl.scrollOffset = dl.cursor
	}
	if dl.cursor >= dl.scrollOffset+dl.height {
		dl.scrollOffset = dl.cursor - dl.height + 1
	}
}

func (dl *DMList) Selected() string {
	if len(dl.dms) == 0 {
		return ""
	}
	return dl.dms[dl.cursor].Channel
}

func (dl *DMList) Participant() string {
	if dl.cursor < 0 || dl.cursor >= len(dl.dms) {
		return ""
	}
	return dl.displayName(dl.dms[dl.cursor])
}

// displayName returns both participants joined with " / ".
func (dl *DMList) displayName(dm sharkfin.DM) string {
	if len(dm.Participants) > 0 {
		return strings.Join(dm.Participants, " / ")
	}
	return dm.Participant
}

func (dl *DMList) SetDMs(dms []sharkfin.DM) {
	dl.dms = dms
	if dl.cursor >= len(dms) {
		dl.cursor = max(0, len(dms)-1)
	}
}

func (dl *DMList) IncrementUnread(channel string) {
	dl.unread[channel]++
}

func (dl *DMList) IncrementMention(channel string) {
	dl.unread[channel]++
	dl.mentions[channel]++
}

func (dl *DMList) SetCounts(channel string, unread, mentions int) {
	if unread > 0 {
		dl.unread[channel] = unread
	} else {
		delete(dl.unread, channel)
	}
	if mentions > 0 {
		dl.mentions[channel] = mentions
	} else {
		delete(dl.mentions, channel)
	}
}

func (dl *DMList) ClearUnread(channel string) {
	delete(dl.unread, channel)
	delete(dl.mentions, channel)
}

func (dl *DMList) HasUnreads() bool {
	return len(dl.unread) > 0
}

func (dl *DMList) SelectByChannel(channel string) {
	for i, dm := range dl.dms {
		if dm.Channel == channel {
			dl.cursor = i
			dl.ensureCursorVisible()
			return
		}
	}
}

func (dl *DMList) SelectIndex(i int) {
	idx := i + dl.scrollOffset
	if idx >= 0 && idx < len(dl.dms) {
		dl.cursor = idx
		dl.ensureCursorVisible()
	}
}

func (dl DMList) View() string {
	if len(dl.dms) == 0 {
		return ui.CurrentTheme.TextDimStyle().Render("  No conversations")
	}

	selectedStyle := ui.CurrentTheme.ListSelectedStyle()
	normalStyle := ui.CurrentTheme.ListNormalStyle()
	unreadStyle := ui.CurrentTheme.ListUnreadStyle()
	mentionStyle := ui.CurrentTheme.ListMentionStyle()

	innerW := dl.width - 4
	if innerW < 1 {
		innerW = 1
	}

	unreadDot := ui.CurrentTheme.UnreadDot()
	unreadCountStyle := ui.CurrentTheme.UnreadCountStyle()
	mentionDot := ui.CurrentTheme.MentionDot()
	mentionCountStyle := ui.CurrentTheme.MentionCountStyle()

	visEnd := dl.scrollOffset + dl.height
	if visEnd > len(dl.dms) {
		visEnd = len(dl.dms)
	}
	visible := dl.dms[dl.scrollOffset:visEnd]

	var lines []string
	for j, dm := range visible {
		i := dl.scrollOffset + j
		prefix := "  "
		name := dl.displayName(dm)

		unreadCount := dl.unread[dm.Channel]
		mentionCount := dl.mentions[dm.Channel]
		hasMention := mentionCount > 0 && i != dl.cursor
		hasUnread := unreadCount > 0 && i != dl.cursor

		if hasMention {
			prefix = mentionDot
		} else if hasUnread {
			prefix = unreadDot
		}

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
		if i == dl.cursor {
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
