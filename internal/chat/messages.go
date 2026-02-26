package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/Work-Fort/WorkFort/pkg/ui"
)

// t is a shorthand for building placeholder timestamps.
func t(hour, min int) time.Time {
	return time.Date(2026, 2, 25, hour, min, 0, 0, time.Local)
}

type MessagePane struct {
	channel    string
	messages   map[string][]MessageInfo
	viewport   viewport.Model
	width      int
	height     int
	autoScroll bool
}

func NewMessagePane() MessagePane {
	vp := viewport.New(0, 0)

	// Seed placeholder messages
	msgs := map[string][]MessageInfo{
		"general": {
			{ID: 1, From: "bob", Body: "Hey team, standup in 5", SentAt: t(10, 32)},
			{ID: 2, From: "alice", Body: "On it", SentAt: t(10, 33)},
			{ID: 3, From: "charlie", Body: "Finishing up a PR, be there in a sec", SentAt: t(10, 34)},
			{ID: 4, From: "bob", Body: "No rush", SentAt: t(10, 34)},
		},
		"engineering": {
			{ID: 5, From: "alice", Body: "Pushed the new auth middleware", SentAt: t(9, 15)},
			{ID: 6, From: "charlie", Body: "Nice, I'll review after lunch", SentAt: t(9, 20)},
		},
		"ops": {
			{ID: 7, From: "charlie", Body: "Deploy to staging went clean", SentAt: t(8, 45)},
		},
		"random": {
			{ID: 8, From: "alice", Body: "Has anyone tried the new coffee machine?", SentAt: t(11, 0)},
			{ID: 9, From: "bob", Body: "It's life-changing", SentAt: t(11, 2)},
			{ID: 10, From: "charlie", Body: "The oat milk option is surprisingly good", SentAt: t(11, 3)},
			{ID: 11, From: "alice", Body: "Right? I've had three cups already", SentAt: t(11, 5)},
			{ID: 12, From: "bob", Body: "That explains the commit frequency today", SentAt: t(11, 6)},
			{ID: 13, From: "charlie", Body: "Lol. Anyone up for lunch at the thai place?", SentAt: t(11, 30)},
			{ID: 14, From: "alice", Body: "Yes! Pad see ew for me", SentAt: t(11, 31)},
			{ID: 15, From: "bob", Body: "I'm in. Green curry", SentAt: t(11, 32)},
			{ID: 16, From: "charlie", Body: "Cool, booking for 12:15", SentAt: t(11, 33)},
			{ID: 17, From: "alice", Body: "Has anyone seen the new conference room names?", SentAt: t(13, 0)},
			{ID: 18, From: "bob", Body: "Yeah, they're all named after Go stdlib packages", SentAt: t(13, 1)},
			{ID: 19, From: "charlie", Body: "I had a meeting in 'net/http' this morning", SentAt: t(13, 2)},
			{ID: 20, From: "alice", Body: "I'm in 'encoding/json' right now", SentAt: t(13, 3)},
			{ID: 21, From: "bob", Body: "Please tell me there's a 'crypto/rand' room", SentAt: t(13, 4)},
			{ID: 22, From: "charlie", Body: "There is. It's the one with no windows", SentAt: t(13, 5)},
			{ID: 23, From: "alice", Body: "That's actually perfect", SentAt: t(13, 6)},
			{ID: 24, From: "bob", Body: "The 'os/exec' room has a great view though", SentAt: t(13, 7)},
			{ID: 25, From: "charlie", Body: "Anyone know if we're doing the team offsite in March?", SentAt: t(14, 0)},
			{ID: 26, From: "alice", Body: "I heard it's happening but location TBD", SentAt: t(14, 2)},
			{ID: 27, From: "bob", Body: "Last year's was great, hope we do something similar", SentAt: t(14, 3)},
			{ID: 28, From: "charlie", Body: "Agreed. The hiking day was a highlight", SentAt: t(14, 5)},
			{ID: 29, From: "alice", Body: "Alright back to work, this PR isn't going to review itself", SentAt: t(14, 10)},
			{ID: 30, From: "bob", Body: "Speak for yourself, I'm training an AI to do mine", SentAt: t(14, 11)},
			{ID: 31, From: "charlie", Body: "Famous last words", SentAt: t(14, 12)},
		},
	}

	return MessagePane{
		channel:    "general",
		messages:   msgs,
		viewport:   vp,
		autoScroll: true,
	}
}

func (mp *MessagePane) SetSize(w, h int) {
	mp.width = w
	mp.height = h
	mp.viewport.Width = w - 4 - 1 // border + padding + scrollbar
	mp.viewport.Height = h
	mp.refreshContent()
}

func (mp *MessagePane) SetChannel(name string) {
	mp.channel = name
	mp.autoScroll = true
	mp.refreshContent()
	mp.viewport.GotoBottom()
}

func (mp *MessagePane) AppendMessage(channel string, msg MessageInfo) {
	mp.messages[channel] = append(mp.messages[channel], msg)

	if channel == mp.channel {
		atBottom := mp.viewport.AtBottom()
		mp.refreshContent()
		if atBottom {
			mp.viewport.GotoBottom()
		}
	}
}

// Channel returns the currently displayed channel name.
func (mp *MessagePane) Channel() string {
	return mp.channel
}

func (mp *MessagePane) refreshContent() {
	msgs := mp.messages[mp.channel]
	if len(msgs) == 0 {
		mp.viewport.SetContent(
			ui.CurrentTheme.TextDimStyle().Render("  No messages yet"),
		)
		return
	}

	contentW := mp.width - 4 - 1 // border + padding + scrollbar
	if contentW < 1 {
		contentW = 1
	}

	nameStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Primary).
		Bold(true)

	timeStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.TextDim)

	bodyStyle := lipgloss.NewStyle().
		Foreground(ui.CurrentTheme.Text).
		Width(contentW)

	var lines []string
	for _, m := range msgs {
		header := fmt.Sprintf("  %s %s",
			nameStyle.Render(m.From),
			timeStyle.Render("["+ui.FormatShortDateTime(m.SentAt)+"]"),
		)
		body := "  " + bodyStyle.Render(m.Body)
		lines = append(lines, header, body, "")
	}

	mp.viewport.SetContent(strings.Join(lines, "\n"))
}

func (mp *MessagePane) ScrollUp(n int) {
	mp.viewport.ScrollUp(n)
}

func (mp *MessagePane) ScrollDown(n int) {
	mp.viewport.ScrollDown(n)
}

func (mp MessagePane) View() string {
	content := mp.viewport.View()
	scrollbar := mp.renderScrollbar()
	return lipgloss.JoinHorizontal(lipgloss.Top, content, scrollbar)
}

func (mp MessagePane) renderScrollbar() string {
	h := mp.viewport.Height
	if h < 1 {
		return ""
	}

	trackStyle := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Muted)
	thumbStyle := lipgloss.NewStyle().Foreground(ui.CurrentTheme.Primary)

	totalLines := mp.viewport.TotalLineCount()
	visibleLines := mp.viewport.VisibleLineCount()

	// No scrollbar needed if all content fits
	if totalLines <= visibleLines {
		return strings.Repeat(trackStyle.Render("│")+"\n", h-1) + trackStyle.Render("│")
	}

	// Thumb size: proportional to visible/total ratio, minimum 1
	thumbSize := int(float64(h) * float64(visibleLines) / float64(totalLines))
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > h {
		thumbSize = h
	}

	// Thumb position from scroll percent
	pct := mp.viewport.ScrollPercent()
	maxOffset := h - thumbSize
	thumbStart := int(pct * float64(maxOffset))
	if thumbStart+thumbSize > h {
		thumbStart = h - thumbSize
	}

	var lines []string
	for i := 0; i < h; i++ {
		if i >= thumbStart && i < thumbStart+thumbSize {
			lines = append(lines, thumbStyle.Render("┃"))
		} else {
			lines = append(lines, trackStyle.Render("│"))
		}
	}
	return strings.Join(lines, "\n")
}
