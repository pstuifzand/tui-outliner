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

	// Session settings (not persisted)
	sessionSettings map[string]string
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

	// Initialize session settings
	config.sessionSettings = make(map[string]string)

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
		Theme:            "tokyo-night",
		sessionSettings:  make(map[string]string),
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

// Set sets a session configuration value
func (c *Config) Set(key, value string) {
	if c.sessionSettings == nil {
		c.sessionSettings = make(map[string]string)
	}
	c.sessionSettings[key] = value
}

// Get retrieves a session configuration value, returns empty string if not found
func (c *Config) Get(key string) string {
	if c.sessionSettings == nil {
		return ""
	}
	return c.sessionSettings[key]
}

// GetAll returns all session configuration values
func (c *Config) GetAll() map[string]string {
	if c.sessionSettings == nil {
		return make(map[string]string)
	}
	// Return a copy to prevent external modifications
	result := make(map[string]string)
	for k, v := range c.sessionSettings {
		result[k] = v
	}
	return result
}
