package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/Work-Fort/WorkFort/pkg/ui"
)

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
			{ID: 1, From: "bob", Body: "Hey team, standup in 5", SentAt: "10:32"},
			{ID: 2, From: "alice", Body: "On it", SentAt: "10:33"},
			{ID: 3, From: "charlie", Body: "Finishing up a PR, be there in a sec", SentAt: "10:34"},
			{ID: 4, From: "bob", Body: "No rush", SentAt: "10:34"},
		},
		"engineering": {
			{ID: 5, From: "alice", Body: "Pushed the new auth middleware", SentAt: "09:15"},
			{ID: 6, From: "charlie", Body: "Nice, I'll review after lunch", SentAt: "09:20"},
		},
		"ops": {
			{ID: 7, From: "charlie", Body: "Deploy to staging went clean", SentAt: "08:45"},
		},
		"random": {
			{ID: 8, From: "alice", Body: "Has anyone tried the new coffee machine?", SentAt: "11:00"},
			{ID: 9, From: "bob", Body: "It's life-changing", SentAt: "11:02"},
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
	mp.viewport.Width = w - 4 // border + padding
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

	contentW := mp.width - 4 // border + padding
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
			timeStyle.Render("["+m.SentAt+"]"),
		)
		body := "  " + bodyStyle.Render(m.Body)
		lines = append(lines, header, body, "")
	}

	mp.viewport.SetContent(strings.Join(lines, "\n"))
}

func (mp *MessagePane) Update(msg interface{}) {
	// viewport key handling is delegated from model.go
}

func (mp MessagePane) View() string {
	return mp.viewport.View()
}
