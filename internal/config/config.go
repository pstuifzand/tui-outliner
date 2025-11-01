package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config holds application configuration
type Config struct {
	Theme    string            `toml:"theme"`
	Settings map[string]string `toml:"settings"`

	// Session settings (not persisted to TOML, overrides persisted settings)
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

	// Initialize persisted settings if not present
	if config.Settings == nil {
		config.Settings = make(map[string]string)
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
		Settings:         make(map[string]string),
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

// Get retrieves a configuration value, checking session settings first (which override persisted settings)
// Returns empty string if not found in either source
func (c *Config) Get(key string) string {
	// Check session settings first (they override persisted settings)
	if c.sessionSettings != nil {
		if val, ok := c.sessionSettings[key]; ok {
			return val
		}
	}

	// Fall back to persisted settings
	if c.Settings != nil {
		if val, ok := c.Settings[key]; ok {
			return val
		}
	}

	return ""
}

// GetAll returns all configuration values (both persisted and session)
// Session settings override persisted settings with the same key
func (c *Config) GetAll() map[string]string {
	result := make(map[string]string)

	// First, add all persisted settings
	if c.Settings != nil {
		for k, v := range c.Settings {
			result[k] = v
		}
	}

	// Then override with session settings (they take precedence)
	if c.sessionSettings != nil {
		for k, v := range c.sessionSettings {
			result[k] = v
		}
	}

	return result
}

// Save persists the configuration to the TOML file
// Note: This only persists the Settings map, not session settings
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure the config directory exists
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshall the config to TOML
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
