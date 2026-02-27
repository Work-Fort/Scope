package chat

import "github.com/Work-Fort/WorkFort/pkg/ui"

// ChatKeyBindings returns the bindings shown in the bottom help bar.
func ChatKeyBindings() ui.KeyBindingSet {
	return ui.KeyBindingSet{
		Bindings: []ui.KeyBinding{
			{Key: "Ctrl+n", Keys: []string{"ctrl+n"}, Description: "new channel"},
			{Key: "Ctrl+d", Keys: []string{"ctrl+d"}, Description: "new dm"},
			{Key: "Ctrl+u", Keys: []string{"ctrl+u"}, Description: "invite user"},
			{Key: "Ctrl+s", Keys: []string{"ctrl+s"}, Description: "shortcuts"},
			{Key: "Ctrl+,", Keys: []string{"ctrl+,"}, Description: "settings"},
			{Key: "Ctrl+q", Keys: []string{"ctrl+q"}, Description: "quit"},
		},
	}
}

// AllShortcuts returns the full list of shortcuts for the shortcuts modal.
func AllShortcuts() []ShortcutGroup {
	return []ShortcutGroup{
		{
			Title: "Navigation",
			Items: []ShortcutItem{
				{Key: "Ctrl+j / Ctrl+k", Description: "next / prev channel"},
				{Key: "Ctrl+a", Description: "switch channels / DMs tab"},
				{Key: "Ctrl+l", Description: "focus input"},
				{Key: "Ctrl+r", Description: "toggle sidebar"},
			{Key: "Ctrl+o", Description: "focus channel list"},
				{Key: "Ctrl+Up / Ctrl+Down", Description: "scroll messages"},
			},
		},
		{
			Title: "Actions",
			Items: []ShortcutItem{
				{Key: "Enter", Description: "send message"},
				{Key: "Alt+Enter", Description: "new line"},
				{Key: "Tab", Description: "@mention completion"},
				{Key: "Ctrl+n", Description: "go to channel"},
				{Key: "Ctrl+d", Description: "open DM"},
				{Key: "Ctrl+u", Description: "invite user"},
				{Key: "Ctrl+,", Description: "settings"},
			},
		},
		{
			Title: "Editing",
			Items: []ShortcutItem{
				{Key: "Ctrl+V", Description: "paste from clipboard"},
				{Key: "Ctrl+Left / Ctrl+Right", Description: "word skip"},
				{Key: "Alt+Backspace", Description: "delete word"},
			},
		},
	}
}

// ShortcutGroup is a named group of shortcuts.
type ShortcutGroup struct {
	Title string
	Items []ShortcutItem
}

// ShortcutItem is a single shortcut entry.
type ShortcutItem struct {
	Key         string
	Description string
}
