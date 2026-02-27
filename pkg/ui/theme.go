package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme holds the application color palette.
type Theme struct {
	Primary   color.Color // Active elements, selection
	Secondary color.Color // Borders, inactive text
	Muted     color.Color // Dimmed elements, help text
	Accent    color.Color // Unread indicators, alerts
	Text      color.Color // Primary text
	TextDim   color.Color // Secondary text
	BgDark    color.Color // Dark background areas
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
}

// Style helpers

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
