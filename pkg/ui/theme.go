package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme holds the application color palette and provides all style builders.
type Theme struct {
	Primary   color.Color // Active elements, selection
	Secondary color.Color // Borders, inactive text
	Muted     color.Color // Dimmed elements, help text
	Accent    color.Color // Unread indicators, alerts
	Text      color.Color // Primary text
	TextDim   color.Color // Secondary text
	BgDark    color.Color // Dark background areas
	Error     color.Color // Error states, recording, disconnected
	Success   color.Color // Success states, online indicators
}

// CurrentTheme is the active application theme.
var CurrentTheme = Theme{
	Primary:   lipgloss.Color("#D4A04A"), // Muted amber/gold
	Secondary: lipgloss.Color("#5A5E6B"), // Cool gray
	Muted:     lipgloss.Color("#3E4250"), // Dim gray
	Accent:    lipgloss.Color("#E8B84B"), // Bright amber
	Text:      lipgloss.Color("#C8CCD4"), // Light gray text
	TextDim:   lipgloss.Color("#6B7080"), // Dimmed text
	BgDark:    lipgloss.Color("#1A1C24"), // Dark panel bg
	Error:     lipgloss.Color("#E05252"), // Red
	Success:   lipgloss.Color("#50C878"), // Green
}

// --- Color style getters ---

func (t Theme) PrimaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary)
}

func (t Theme) SecondaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Secondary)
}

func (t Theme) MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Muted)
}

func (t Theme) AccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Accent)
}

func (t Theme) TextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Text)
}

func (t Theme) TextDimStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.TextDim)
}

func (t Theme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Error)
}

func (t Theme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Success)
}

// --- List item styles ---

func (t Theme) ListSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
}

func (t Theme) ListNormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Text)
}

func (t Theme) ListUnreadStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Text).Bold(true)
}

func (t Theme) ListMentionStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Secondary).Bold(true)
}

// --- Indicator dots (rendered strings) ---

func (t Theme) UnreadDot() string {
	return lipgloss.NewStyle().Foreground(t.Accent).Render("● ")
}

func (t Theme) MentionDot() string {
	return lipgloss.NewStyle().Foreground(t.Secondary).Render("● ")
}

func (t Theme) OnlineDot() string {
	return lipgloss.NewStyle().Foreground(t.Success).Render("● ")
}

func (t Theme) OfflineDot() string {
	return lipgloss.NewStyle().Foreground(t.Muted).Render("○ ")
}

// --- Badge count styles ---

func (t Theme) UnreadCountStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Accent)
}

func (t Theme) MentionCountStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Secondary)
}

// --- Button styles (help bar + modals) ---

func (t Theme) KeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
}

func (t Theme) KeyDescStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.TextDim)
}

func (t Theme) ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Muted).
		PaddingLeft(1).
		PaddingRight(1)
}

// --- Message pane styles ---

func (t Theme) MessageNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
}

func (t Theme) MessageTimeStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.TextDim)
}

func (t Theme) ScrollTrackStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Muted)
}

func (t Theme) ScrollThumbStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary)
}

// --- Modal styles ---

func (t Theme) ModalTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
}

func (t Theme) ModalBoxStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2).
		Width(width)
}

func (t Theme) ModalLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Secondary).Bold(true)
}

func (t Theme) ModalHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.TextDim)
}

// --- Layout styles ---

func (t Theme) HeaderStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Muted).
		Width(width).
		Align(lipgloss.Center)
}

func (t Theme) HeaderTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
}

func (t Theme) ChannelHeaderStyle(width, height int, borderColor color.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height)
}

func (t Theme) TabActiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
}

func (t Theme) TabInactiveStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.TextDim)
}

func (t Theme) TabBaseStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Muted).
		Align(lipgloss.Center)
}

func (t Theme) ActionButtonStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Muted).
		Width(width).
		Align(lipgloss.Center)
}
