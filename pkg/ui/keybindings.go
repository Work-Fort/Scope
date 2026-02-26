package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

// Render formats the key binding set as a help bar.
func (kbs KeyBindingSet) Render() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(CurrentTheme.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(CurrentTheme.TextDim)

	sepStyle := lipgloss.NewStyle().
		Foreground(CurrentTheme.Muted)

	var parts []string
	for _, kb := range kbs.Bindings {
		part := keyStyle.Render(kb.Key) + " " + descStyle.Render(kb.Description)
		parts = append(parts, part)
	}

	return strings.Join(parts, sepStyle.Render("  "))
}
