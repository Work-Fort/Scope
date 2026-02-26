package config

import (
	"fmt"
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
