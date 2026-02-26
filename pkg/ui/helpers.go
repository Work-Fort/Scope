package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const (
	MinWidth  = 80
	MinHeight = 24

	SidebarRatio = 0.25
	MinSidebarW  = 20
	MaxSidebarW  = 40

	InputHeight = 3
	HelpHeight  = 3
	PaneTitleH  = 2 // title line + 1 blank line
)

// ChatLayout holds calculated dimensions for the chat UI.
type ChatLayout struct {
	SidebarW int
	MessageW int
	ContentH int // Height available for sidebar and message pane
	InputH   int
	HelpH    int
	TotalW   int
	TotalH   int
}

// CalculateChatLayout computes panel dimensions from terminal size.
func CalculateChatLayout(termW, termH int) ChatLayout {
	sidebarW := int(float64(termW) * SidebarRatio)
	if sidebarW < MinSidebarW {
		sidebarW = MinSidebarW
	}
	if sidebarW > MaxSidebarW {
		sidebarW = MaxSidebarW
	}

	// Account for borders: 2 chars per pane (left+right)
	messageW := termW - sidebarW

	// Vertical: content + help bar
	contentH := termH - HelpHeight

	return ChatLayout{
		SidebarW: sidebarW,
		MessageW: messageW,
		ContentH: contentH,
		InputH:   InputHeight,
		HelpH:    HelpHeight,
		TotalW:   termW,
		TotalH:   termH,
	}
}

// CreatePaneStyle returns a bordered style for active or inactive panes.
func CreatePaneStyle(isActive bool, width, height int) lipgloss.Style {
	var borderColor lipgloss.Color
	var border lipgloss.Border

	if isActive {
		borderColor = CurrentTheme.Primary
		border = lipgloss.ThickBorder()
	} else {
		borderColor = CurrentTheme.Muted
		border = lipgloss.NormalBorder()
	}

	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(borderColor).
		Width(width - 2).  // subtract border chars
		Height(height - 2) // subtract border chars
}

// RenderSizeError renders a centered error box when the terminal is too small.
func RenderSizeError(minW, minH, curW, curH int) string {
	msg := fmt.Sprintf(
		"  Window too small\n  Minimum: %dx%d\n  Current: %dx%d",
		minW, minH, curW, curH,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(CurrentTheme.Secondary).
		Padding(1, 2).
		Foreground(CurrentTheme.Text).
		Render(msg)

	return lipgloss.Place(curW, curH, lipgloss.Center, lipgloss.Center, box)
}

// FillTerminal places content centered in the terminal dimensions.
func FillTerminal(content string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, content)
}

// RenderPaneTitle renders a pane title with theme styling.
func RenderPaneTitle(title string, isActive bool) string {
	if isActive {
		return CurrentTheme.PrimaryStyle().Bold(true).Render(title)
	}
	return CurrentTheme.SecondaryStyle().Render(title)
}
