package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"

	"github.com/Work-Fort/Scope/cmd/chat"
	"github.com/Work-Fort/Scope/pkg/config"
	"github.com/Work-Fort/Scope/pkg/ui"
)

var (
	Version = "dev"

	logLevel    string
	useTUI      bool
	debugLogger *log.Logger
)

var rootCmd = &cobra.Command{
	Use:   "workfort",
	Short: "Scope CLI",
	Long:  "The Scope command-line interface for team collaboration.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitDirs(); err != nil {
			return err
		}

		if err := config.LoadConfig(); err != nil {
			return err
		}

		if viper.GetBool("auto-theme") {
			_ = ui.LoadOmarchyTheme() // silently fall back to default theme
		}

		useTUI = config.GetUseTUI()

		if err := setupLogging(); err != nil {
			return fmt.Errorf("failed to setup logging: %w", err)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func GetDebugLogger() *log.Logger {
	if debugLogger == nil {
		return log.Default()
	}
	return debugLogger
}

func init() {
	config.InitViper()

	rootCmd.AddCommand(chat.NewChatCmd())

	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "debug", "Log level (debug, info, warn, error, disabled)")
	rootCmd.PersistentFlags().BoolVar(&useTUI, "use-tui", true, "Enable TUI mode")

	config.BindFlags(rootCmd.PersistentFlags())
}

func setupLogging() error {
	level := config.GetLogLevel()

	var logLevelParsed log.Level
	switch level {
	case "debug":
		logLevelParsed = log.DebugLevel
	case "info":
		logLevelParsed = log.InfoLevel
	case "warn":
		logLevelParsed = log.WarnLevel
	case "error":
		logLevelParsed = log.ErrorLevel
	case "disabled":
		debugLogger = log.NewWithOptions(io.Discard, log.Options{})
		log.SetDefault(debugLogger)
		return nil
	default:
		logLevelParsed = log.DebugLevel
	}

	logFile := filepath.Join(config.GlobalPaths.DataDir, "debug.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logFile, err)
	}

	debugLogger = log.NewWithOptions(f, log.Options{
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02T15:04:05.000Z07:00",
		Level:           logLevelParsed,
		ReportCaller:    true,
		Formatter:       log.JSONFormatter,
	})

	log.SetDefault(debugLogger)

	return nil
}
