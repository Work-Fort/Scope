package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	AppName             = "workfort"
	EnvPrefix           = "WORKFORT"
	UserConfigFileName  = "config"
	LocalConfigFileName = "workfort"
	ConfigType          = "yaml"
)

// Paths holds XDG-compliant directory paths.
type Paths struct {
	ConfigDir string // XDG_CONFIG_HOME/workfort
	DataDir   string // XDG_DATA_HOME/workfort
	CacheDir  string // XDG_CACHE_HOME/workfort
}

var GlobalPaths *Paths

func init() {
	GlobalPaths = GetPaths()
}

func GetPaths() *Paths {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}

	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, _ := os.UserHomeDir()
		dataHome = filepath.Join(home, ".local", "share")
	}

	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		home, _ := os.UserHomeDir()
		cacheHome = filepath.Join(home, ".cache")
	}

	return &Paths{
		ConfigDir: filepath.Join(configHome, AppName),
		DataDir:   filepath.Join(dataHome, AppName),
		CacheDir:  filepath.Join(cacheHome, AppName),
	}
}

func InitDirs() error {
	dirs := []string{
		GlobalPaths.ConfigDir,
		GlobalPaths.DataDir,
		GlobalPaths.CacheDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
