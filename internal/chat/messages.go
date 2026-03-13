package chat

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"

	"github.com/Work-Fort/Scope/pkg/sharkfin"
	"github.com/Work-Fort/Scope/pkg/ui"
)

type MessagePane struct {
	channel    string
	messages   map[string][]sharkfin.Message
	viewport   viewport.Model
	width      int
	height     int
	autoScroll bool
	mdRenderer *glamour.TermRenderer
}

func NewMessagePane() MessagePane {
	vp := viewport.New()

	return MessagePane{
		messages:   make(map[string][]sharkfin.Message),
		viewport:   vp,
		autoScroll: true,
	}
}

// Height returns the inner height of the message pane in lines.
func (mp *MessagePane) Height() int {
	return mp.height
}

func (mp *MessagePane) SetSize(w, h int) {
	oldW := mp.width
	oldH := mp.viewport.Height()
	mp.width = w
	mp.height = h
	mp.viewport.SetWidth(w - 2 - 1) // border + scrollbar
	mp.viewport.SetHeight(h)
	// Only recreate markdown renderer when width changes
	if w != oldW {
		contentW := w - 2 - 1
		if contentW < 1 {
			contentW = 1
		}
		style := styles.DarkStyleConfig
		noMargin := uint(0)
		style.Document.Margin = &noMargin
		style.CodeBlock.StyleBlock.Margin = &noMargin
		r, _ := glamour.NewTermRenderer(
			glamour.WithStyles(style),
			glamour.WithWordWrap(contentW-1), // -1 for left padding
		)
		mp.mdRenderer = r
	}
	mp.refreshContent()

	// Anchor scroll: when height changes, keep the viewport anchored appropriately.
	if oldH > 0 && h != oldH {
		if mp.viewport.AtBottom() {
			// Stay pinned to bottom when already there
			mp.viewport.GotoBottom()
		} else {
			// Keep the same bottom line visible
			bottomLine := mp.viewport.YOffset() + oldH
			newOffset := bottomLine - h
			if newOffset < 0 {
				newOffset = 0
			}
			mp.viewport.SetYOffset(newOffset)
		}
	}
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
			mp.viewport.SetYOffset(mp.viewport.YOffset() + addedLines)
		}
	}
}

// LatestID returns the largest message ID for the channel, or 0 if none.
func (mp *MessagePane) LatestID(channel string) int {
	msgs := mp.messages[channel]
	if len(msgs) == 0 {
		return 0
	}
	return msgs[len(msgs)-1].ID
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

	contentW := mp.width - 2 - 1 // border + scrollbar
	if contentW < 1 {
		contentW = 1
	}

	nameStyle := ui.CurrentTheme.MessageNameStyle()
	timeStyle := ui.CurrentTheme.MessageTimeStyle()

	var lines []string
	for _, m := range msgs {
		header := fmt.Sprintf(" %s %s",
			nameStyle.Render(m.From),
			timeStyle.Render("["+ui.FormatShortDateTime(m.SentAt)+"]"),
		)
		body := m.Body
		if mp.mdRenderer != nil {
			if rendered, err := mp.mdRenderer.Render(body); err == nil {
				body = strings.TrimRight(rendered, "\n")
			}
		}
		// Indent each line by 1 space
		indented := make([]string, 0)
		for _, line := range strings.Split(body, "\n") {
			indented = append(indented, " "+line)
		}
		body = strings.Join(indented, "\n")
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
	h := mp.viewport.Height()
	if h < 1 {
		return ""
	}

	trackStyle := ui.CurrentTheme.ScrollTrackStyle()
	thumbStyle := ui.CurrentTheme.ScrollThumbStyle()

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
