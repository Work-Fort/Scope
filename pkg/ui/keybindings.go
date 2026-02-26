package ui

import "github.com/charmbracelet/lipgloss"

// KeyBinding represents a single key binding with its display info.
type KeyBinding struct {
	Key         string   // Display name (e.g., "Ctrl+j")
	Keys        []string // Actual key strings to match (e.g., "ctrl+j")
	Description string
}

// KeyBindingSet is an ordered collection of key bindings.
type KeyBindingSet struct {
	Bindings []KeyBinding
}

// Contains checks if a keypress string matches any binding in the set.
func (kbs KeyBindingSet) Contains(key string) *KeyBinding {
	for i := range kbs.Bindings {
		for _, k := range kbs.Bindings[i].Keys {
			if k == key {
				return &kbs.Bindings[i]
			}
		}
	}
	return nil
}

// Render formats the key binding set as bordered buttons.
func (kbs KeyBindingSet) Render() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(CurrentTheme.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(CurrentTheme.TextDim)

	btnStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(CurrentTheme.Muted).
		PaddingLeft(1).
		PaddingRight(1)

	var buttons []string
	for _, kb := range kbs.Bindings {
		label := keyStyle.Render(kb.Key) + " " + descStyle.Render(kb.Description)
		buttons = append(buttons, btnStyle.Render(label))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, buttons...)
}

// ButtonRegions returns the X start/end positions for each button in the rendered help bar.
// Used for mouse click hit testing.
func (kbs KeyBindingSet) ButtonRegions() []ButtonRegion {
	keyStyle := lipgloss.NewStyle().
		Foreground(CurrentTheme.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(CurrentTheme.TextDim)

	btnStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(CurrentTheme.Muted).
		PaddingLeft(1).
		PaddingRight(1)

	var regions []ButtonRegion
	x := 0
	for i, kb := range kbs.Bindings {
		label := keyStyle.Render(kb.Key) + " " + descStyle.Render(kb.Description)
		rendered := btnStyle.Render(label)
		w := lipgloss.Width(rendered)
		regions = append(regions, ButtonRegion{
			Index: i,
			XMin:  x,
			XMax:  x + w,
		})
		x += w
	}
	return regions
}

// ButtonRegion describes the horizontal hit area of a rendered button.
type ButtonRegion struct {
	Index int
	XMin  int
	XMax  int
}
