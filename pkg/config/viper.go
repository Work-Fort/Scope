package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// InitViper sets up Viper with defaults and environment variable support.
// Called once during application init(), before directories are created.
func InitViper() {
	viper.SetConfigType(ConfigType)

	viper.SetDefault("use-tui", true)
	viper.SetDefault("log-level", "debug")
	viper.SetDefault("sharkfin-host", "ws://127.0.0.1:16000/ws")
	viper.SetDefault("username", "")
	viper.SetDefault("auto-theme", true)
	viper.SetDefault("time-display.use-24h", false)
	viper.SetDefault("time-display.show-seconds", false)

	// Fort defaults
	viper.SetDefault("active-fort", "local")
	viper.SetDefault("forts.local.local", true)
	viper.SetDefault("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.SetDefault("forts.local.services.auth.enabled", true)
	viper.SetDefault("forts.local.services.sharkfin.url", "http://127.0.0.1:16000")
	viper.SetDefault("forts.local.services.sharkfin.enabled", true)
	viper.SetDefault("forts.local.services.sharkfin.ws-paths", []string{"/ws", "/presence"})
	viper.SetDefault("forts.local.services.nexus.url", "http://127.0.0.1:9600")
	viper.SetDefault("forts.local.services.nexus.enabled", true)
	viper.SetDefault("forts.local.services.hive.url", "http://127.0.0.1:17000")
	viper.SetDefault("forts.local.services.hive.enabled", false)

	viper.SetEnvPrefix(EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

// LoadConfig reads config files in precedence order.
// Precedence: ENV > ./workfort.yaml > ~/.config/workfort/config.yaml > defaults
func LoadConfig() error {
	viper.SetConfigName(UserConfigFileName)
	viper.AddConfigPath(GlobalPaths.ConfigDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read user config: %w", err)
		}
	}

	viper.SetConfigName(LocalConfigFileName)
	viper.AddConfigPath(".")

	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to merge local config: %w", err)
		}
	}

	return nil
}

// BindFlags binds application flags to Viper.
func BindFlags(flags *pflag.FlagSet) error {
	flagsToBind := []string{
		"use-tui",
		"log-level",
		"sharkfin-host",
		"username",
	}

	for _, flagName := range flagsToBind {
		if err := viper.BindPFlag(flagName, flags.Lookup(flagName)); err != nil {
			return fmt.Errorf("failed to bind flag %s: %w", flagName, err)
		}
	}

	return nil
}

func GetUseTUI() bool {
	return viper.GetBool("use-tui")
}

func GetLogLevel() string {
	return viper.GetString("log-level")
}

// SaveSetting sets a key in viper and persists it to the user config file.
func SaveSetting(key, value string) error {
	viper.Set(key, value)
	configPath := filepath.Join(GlobalPaths.ConfigDir, UserConfigFileName+"."+ConfigType)
	return viper.WriteConfigAs(configPath)
}
