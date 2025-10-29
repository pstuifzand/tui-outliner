package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config holds application configuration
type Config struct {
	Theme string `toml:"theme"`
}

// Load loads the config file from the standard location
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return defaultConfig(), nil // Return default if can't find config path
	}

	return LoadFromFile(configPath)
}

// LoadFromFile loads config from a specific file
func LoadFromFile(filePath string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return defaultConfig(), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults if not specified
	if config.Theme == "" {
		config.Theme = "tokyo-night"
	}

	return &config, nil
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "tui-outliner", "config.toml"), nil
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		Theme: "tokyo-night",
	}
}

// GetConfigDir returns the config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "tui-outliner")
	return configDir, nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(configDir, 0755)
}
