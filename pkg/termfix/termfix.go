// Package termfix pre-configures lipgloss v1 (used by glamour) so it does not
// attempt to query the terminal when bubbletea v2 already owns stdin/stdout.
// Import this package (blank import) BEFORE bubbletea in the import list.
package termfix

import (
	lipglossv1 "github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	lipglossv1.SetColorProfile(termenv.TrueColor)
	lipglossv1.SetHasDarkBackground(true)
}
