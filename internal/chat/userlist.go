package chat

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/Work-Fort/WorkFort/pkg/sharkfin"
	"github.com/Work-Fort/WorkFort/pkg/ui"
)

type UserList struct {
	users        []sharkfin.User
	cursor       int
	scrollOffset int
	width        int
	height       int
	dmChannels   map[string]string // username -> DM channel name
	dmUnread     map[string]int    // DM channel -> unread count
	myUsername   string
}

func NewUserList(username string) UserList {
	return UserList{
		dmChannels: make(map[string]string),
		dmUnread:   make(map[string]int),
		myUsername:  username,
	}
}

func (ul *UserList) SetSize(w, h int) {
	ul.width = w
	ul.height = h
}

func (ul *UserList) MoveUp() {
	if ul.cursor > 0 {
		ul.cursor--
		ul.ensureCursorVisible()
	}
}

func (ul *UserList) MoveDown() {
	if ul.cursor < len(ul.users)-1 {
		ul.cursor++
		ul.ensureCursorVisible()
	}
}

func (ul *UserList) ensureCursorVisible() {
	if ul.height <= 0 {
		return
	}
	if ul.cursor < ul.scrollOffset {
		ul.scrollOffset = ul.cursor
	}
	if ul.cursor >= ul.scrollOffset+ul.height {
		ul.scrollOffset = ul.cursor - ul.height + 1
	}
}

// SelectedUsername returns the username at the cursor.
func (ul *UserList) SelectedUsername() string {
	if len(ul.users) == 0 {
		return ""
	}
	return ul.users[ul.cursor].Username
}

// SelectedDMChannel returns the DM channel for the selected user, or "".
func (ul *UserList) SelectedDMChannel() string {
	return ul.dmChannels[ul.SelectedUsername()]
}

func (ul *UserList) SetUsers(users []sharkfin.User) {
	// Filter out self, sort online first then alphabetical
	var filtered []sharkfin.User
	for _, u := range users {
		if u.Username != ul.myUsername {
			filtered = append(filtered, u)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Online != filtered[j].Online {
			return filtered[i].Online
		}
		return filtered[i].Username < filtered[j].Username
	})
	ul.users = filtered
	if ul.cursor >= len(filtered) {
		ul.cursor = max(0, len(filtered)-1)
	}
}

// UpdatePresence updates a single user's online status and re-sorts.
func (ul *UserList) UpdatePresence(username string, online bool) {
	for i := range ul.users {
		if ul.users[i].Username == username {
			ul.users[i].Online = online
			break
		}
	}
	// Re-sort to maintain online-first order
	selectedUser := ul.SelectedUsername()
	sort.Slice(ul.users, func(i, j int) bool {
		if ul.users[i].Online != ul.users[j].Online {
			return ul.users[i].Online
		}
		return ul.users[i].Username < ul.users[j].Username
	})
	// Restore cursor to same user after sort
	for i, u := range ul.users {
		if u.Username == selectedUser {
			ul.cursor = i
			ul.ensureCursorVisible()
			break
		}
	}
}

// SetDMMapping updates the username -> DM channel mapping from the DM list.
func (ul *UserList) SetDMMapping(dms []sharkfin.DM, myUsername string) {
	ul.dmChannels = make(map[string]string, len(dms))
	for _, dm := range dms {
		log.Debug("dm_mapping_raw", "channel", dm.Channel, "participant", dm.Participant, "participants", dm.Participants, "member", dm.Member)
		// WS dm_list is admin-scoped (returns all DMs). Only map DMs
		// where we are a participant to avoid cross-mapping other users' DMs.
		if len(dm.Participants) > 0 {
			isMember := false
			for _, p := range dm.Participants {
				if p == myUsername {
					isMember = true
					break
				}
			}
			if !isMember {
				log.Debug("dm_mapping_skip", "channel", dm.Channel, "reason", "not_member")
				continue
			}
			for _, p := range dm.Participants {
				if p != myUsername {
					log.Debug("dm_mapping_set", "username", p, "channel", dm.Channel)
					ul.dmChannels[p] = dm.Channel
				}
			}
		} else if dm.Participant != "" && dm.Participant != myUsername {
			// MCP path: Participant is already user-scoped (the other user)
			log.Debug("dm_mapping_set", "username", dm.Participant, "channel", dm.Channel)
			ul.dmChannels[dm.Participant] = dm.Channel
		}
	}
}

func (ul *UserList) SetDMUnread(channel string, count int) {
	if count > 0 {
		ul.dmUnread[channel] = count
	} else {
		delete(ul.dmUnread, channel)
	}
}

func (ul *UserList) ClearDMUnread(channel string) {
	delete(ul.dmUnread, channel)
}

func (ul *UserList) SelectIndex(i int) {
	idx := i + ul.scrollOffset
	if idx >= 0 && idx < len(ul.users) {
		ul.cursor = idx
		ul.ensureCursorVisible()
	}
}

func (ul UserList) View() string {
	if len(ul.users) == 0 {
		return ui.CurrentTheme.TextDimStyle().Render("  No users")
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Text)

	onlineDot := lipgloss.NewStyle().Foreground(lipgloss.Color("#50C878")).Render("● ")
	offlineDot := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Muted).Render("○ ")

	unreadCountStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Accent)

	innerW := ul.width - 4
	if innerW < 1 {
		innerW = 1
	}

	visEnd := ul.scrollOffset + ul.height
	if visEnd > len(ul.users) {
		visEnd = len(ul.users)
	}
	visible := ul.users[ul.scrollOffset:visEnd]

	var lines []string
	for j, user := range visible {
		i := ul.scrollOffset + j

		prefix := offlineDot
		if user.Online {
			prefix = onlineDot
		}

		name := user.Username
		dmChan := ul.dmChannels[name]
		unread := ul.dmUnread[dmChan]

		maxNameW := innerW - 2
		suffix := ""
		if unread > 0 && i != ul.cursor {
			badge := fmt.Sprintf(" (%d)", unread)
			maxNameW -= len(badge)
			suffix = badge
		}
		if len(name) > maxNameW {
			name = name[:maxNameW-1] + "…"
		}

		var line string
		if i == ul.cursor {
			line = prefix + selectedStyle.Render(name)
		} else if suffix != "" {
			line = prefix + normalStyle.Render(name) + " " + unreadCountStyle.Render(suffix)
		} else {
			line = prefix + normalStyle.Render(name)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
