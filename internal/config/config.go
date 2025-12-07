package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// ConfigDirName is the directory name under ~/.config
	ConfigDirName = "brewsync"
	// ConfigFileName is the config file name without extension
	ConfigFileName = "config"
	// ConfigFileType is the config file extension
	ConfigFileType = "yaml"
)

var (
	// cfg holds the loaded configuration
	cfg *Config
	// configPath holds the path to the config file
	configPath string
)

// configDir returns the config directory path
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", ConfigDirName), nil
}

// ConfigDir returns the config directory path (exported version)
func ConfigDir() string {
	dir, err := configDir()
	if err != nil {
		// Fallback to a default (should rarely happen)
		return filepath.Join(os.Getenv("HOME"), ".config", "brewsync")
	}
	return dir
}

// ConfigPath returns the full path to the config file
func ConfigPath() (string, error) {
	if configPath != "" {
		return configPath, nil
	}
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName+"."+ConfigFileType), nil
}

// SetConfigPath overrides the default config path
func SetConfigPath(path string) {
	configPath = path
}

// Init initializes viper with defaults and loads config if it exists
func Init() error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	// Set defaults
	setDefaults()

	// Configure viper
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileType)
	viper.AddConfigPath(dir)

	// Allow override via custom path
	if configPath != "" {
		viper.SetConfigFile(configPath)
	}

	// Environment variable support
	viper.SetEnvPrefix("BREWSYNC")
	viper.AutomaticEnv()

	// Bind MACHINE env var to current_machine
	if err := viper.BindEnv("current_machine", "MACHINE"); err != nil {
		return fmt.Errorf("failed to bind MACHINE env var: %w", err)
	}

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Also check for standard file not found error (occurs with SetConfigFile)
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to read config: %w", err)
			}
		}
		// Config file not found is OK - we'll use defaults
	}

	return nil
}

// Load unmarshals the config into the Config struct
func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	if err := Init(); err != nil {
		return nil, err
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Detect current machine if set to "auto"
	if cfg.CurrentMachine == "auto" || cfg.CurrentMachine == "" {
		detected, err := DetectMachine(cfg.Machines)
		if err == nil {
			cfg.CurrentMachine = detected
		}
	}

	// Load ignore file
	ignoreFile, err := LoadIgnoreFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load ignore file: %w", err)
	}
	cfg.ignoreFile = ignoreFile

	return cfg, nil
}

// Get returns the loaded config, loading it if necessary
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}
	return Load()
}

// Exists checks if the config file exists
func Exists() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// EnsureDir creates the config directory if it doesn't exist
func EnsureDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// ProfilesDir returns the path to the profiles directory
func ProfilesDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles"), nil
}

// HistoryPath returns the path to the history log file
func HistoryPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.log"), nil
}
