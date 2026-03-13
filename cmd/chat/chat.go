package chat

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	chatui "github.com/Work-Fort/Scope/internal/chat"
	"github.com/Work-Fort/Scope/pkg/sharkfin"
)

var (
	sharkfinHost string
	username     string
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Open the chat TUI",
		Long:  "Connect to a sharkfin daemon and open an interactive chat interface.",
		RunE:  runChat,
	}

	cmd.Flags().StringVar(&sharkfinHost, "sharkfin-host", "", "Sharkfin daemon WebSocket URL")
	cmd.Flags().StringVar(&username, "username", "", "Username for chat")

	return cmd
}

func runChat(cmd *cobra.Command, args []string) error {
	host := sharkfinHost
	if host == "" {
		host = viper.GetString("sharkfin-host")
	}

	user := username
	if user == "" {
		user = viper.GetString("username")
	}
	if user == "" {
		return fmt.Errorf("username is required (--username flag or WORKFORT_USERNAME env)")
	}

	// Connect to sharkfin
	client := sharkfin.New(host)
	_, err := client.Connect()
	if err != nil {
		return fmt.Errorf("connect to sharkfin: %w", err)
	}
	defer client.Close()

	// Identify (auto-registers on first use)
	if err := client.Identify(user); err != nil {
		return fmt.Errorf("identify: %w", err)
	}

	model := chatui.NewModel(client, user)
	p := tea.NewProgram(model)

	// Start pumps before Run() so Init()'s requests can be sent immediately
	go client.WritePump()
	go client.ReadPump(p)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("chat exited with error: %w", err)
	}

	return nil
}
