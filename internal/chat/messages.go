package chat

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/Work-Fort/WorkFort/pkg/sharkfin"
	"github.com/Work-Fort/WorkFort/pkg/ui"
)

type MessagePane struct {
	channel    string
	messages   map[string][]sharkfin.Message
	viewport   viewport.Model
	width      int
	height     int
	autoScroll bool
}

func NewMessagePane() MessagePane {
	vp := viewport.New(0, 0)

	return MessagePane{
		messages:   make(map[string][]sharkfin.Message),
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

func (mp *MessagePane) AppendMessage(channel string, msg sharkfin.Message) {
	mp.messages[channel] = append(mp.messages[channel], msg)

	if channel == mp.channel {
		atBottom := mp.viewport.AtBottom()
		mp.refreshContent()
		if atBottom {
			mp.viewport.GotoBottom()
		}
	}
}

// MergeMessages adds messages not already present, maintaining sort order by ID.
func (mp *MessagePane) MergeMessages(channel string, msgs []sharkfin.Message) {
	if len(msgs) == 0 {
		return
	}

	existing := mp.messages[channel]
	seen := make(map[int]bool, len(existing))
	for _, m := range existing {
		seen[m.ID] = true
	}
	for _, m := range msgs {
		if !seen[m.ID] {
			existing = append(existing, m)
		}
	}
	slices.SortFunc(existing, func(a, b sharkfin.Message) int {
		return a.ID - b.ID
	})
	mp.messages[channel] = existing

	if channel == mp.channel {
		atBottom := mp.viewport.AtBottom()
		mp.refreshContent()
		if atBottom {
			mp.viewport.GotoBottom()
		}
	}
}

// PrependHistory adds older messages to the front, preserving scroll position.
func (mp *MessagePane) PrependHistory(channel string, msgs []sharkfin.Message) {
	if len(msgs) == 0 {
		return
	}

	existing := mp.messages[channel]
	all := make([]sharkfin.Message, 0, len(msgs)+len(existing))
	all = append(all, msgs...)
	all = append(all, existing...)
	mp.messages[channel] = all

	if channel == mp.channel {
		prevTotal := mp.viewport.TotalLineCount()
		mp.refreshContent()
		addedLines := mp.viewport.TotalLineCount() - prevTotal
		if addedLines > 0 {
			mp.viewport.SetYOffset(mp.viewport.YOffset + addedLines)
		}
	}
}

// OldestID returns the smallest message ID for the channel, or 0 if none.
func (mp *MessagePane) OldestID(channel string) int {
	msgs := mp.messages[channel]
	if len(msgs) == 0 {
		return 0
	}
	return msgs[0].ID
}

// AtTop returns true if the viewport is scrolled to the top.
func (mp *MessagePane) AtTop() bool {
	return mp.viewport.AtTop()
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
