package chat

import "github.com/Work-Fort/WorkFort/pkg/ui"

func ChatKeyBindings() ui.KeyBindingSet {
	return ui.KeyBindingSet{
		Bindings: []ui.KeyBinding{
			{Key: "Ctrl+j/k", Keys: []string{"ctrl+j", "ctrl+k"}, Description: "channels"},
			{Key: "Ctrl+l", Keys: []string{"ctrl+l"}, Description: "input"},
			{Key: "Ctrl+o", Keys: []string{"ctrl+o"}, Description: "channels"},
			{Key: "Ctrl+n", Keys: []string{"ctrl+n"}, Description: "new channel"},
			{Key: "Ctrl+u", Keys: []string{"ctrl+u"}, Description: "invite user"},
			{Key: "Ctrl+q", Keys: []string{"ctrl+q"}, Description: "quit"},
		},
	}
}
