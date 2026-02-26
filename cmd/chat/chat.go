package chat

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	chatui "github.com/Work-Fort/WorkFort/internal/chat"
)

var (
	sharkfinHost string
	username     string
)

func NewChatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Open the chat TUI",
		Long:  "Connect to a sharkfin daemon and open an interactive chat interface.",
		RunE:  runChat,
	}

	cmd.Flags().StringVar(&sharkfinHost, "sharkfin-host", "ws://127.0.0.1:16000/ws", "Sharkfin daemon WebSocket URL")
	cmd.Flags().StringVar(&username, "username", "", "Username for chat")
	cmd.MarkFlagRequired("username")

	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	// Plan 2: connect to sharkfin here
	// client := sharkfin.New(sharkfinHost)
	// client.Connect()
	// client.Identify(username)

	model := chatui.NewModel(username)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Plan 2: launch goroutines
	// go client.ReadPump(p)
	// go client.WritePump()
	// defer client.Close()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("chat exited with error: %w", err)
	}

	return nil
}
