// Package termfix pre-sets lipgloss background detection to avoid
// a 5-second OSC query timeout during Bubble Tea initialization.
// Import this package (blank import) BEFORE bubbletea in the import list.
package termfix

import "github.com/charmbracelet/lipgloss"

func init() {
	lipgloss.SetHasDarkBackground(true)
}
