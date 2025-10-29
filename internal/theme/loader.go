package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// ThemeConfig represents the raw TOML theme configuration
type ThemeConfig struct {
	Name   string `toml:"name"`
	Colors struct {
		TreeNormalText     string `toml:"tree_normal_text"`
		TreeSelectedItem   string `toml:"tree_selected_item"`
		TreeNewItem        string `toml:"tree_new_item"`
		TreeLeafArrow      string `toml:"tree_leaf_arrow"`
		TreeExpandedArrow  string `toml:"tree_expanded_arrow"`
		TreeCollapsedArrow string `toml:"tree_collapsed_arrow"`
		EditorText         string `toml:"editor_text"`
		EditorCursor       string `toml:"editor_cursor"`
		SearchLabel        string `toml:"search_label"`
		SearchText         string `toml:"search_text"`
		SearchCursor       string `toml:"search_cursor"`
		SearchResultCount  string `toml:"search_result_count"`
		CommandPrompt      string `toml:"command_prompt"`
		CommandText        string `toml:"command_text"`
		CommandCursor      string `toml:"command_cursor"`
		HelpBackground     string `toml:"help_background"`
		HelpBorder         string `toml:"help_border"`
		HelpTitle          string `toml:"help_title"`
		HelpContent        string `toml:"help_content"`
		StatusMode         string `toml:"status_mode"`
		StatusMessage      string `toml:"status_message"`
		StatusModified     string `toml:"status_modified"`
		HeaderTitle        string `toml:"header_title"`
	} `toml:"colors"`
}

// getThemePaths returns the search paths for theme files
func getThemePaths() []string {
	paths := []string{}

	// User config directory
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "tui-outliner", "themes"))
	}

	// User local share directory
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".local", "share", "tui-outliner", "themes"))
	}

	return paths
}

// findThemeFile searches for a theme file in standard locations
func findThemeFile(themeName string) (string, error) {
	filename := themeName + ".toml"

	for _, dir := range getThemePaths() {
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("theme file not found: %s", filename)
}

// LoadThemeFromFile loads a theme from a TOML file
func LoadThemeFromFile(filePath string) (*Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	var config ThemeConfig
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	return configToTheme(config), nil
}

// LoadTheme loads a theme by name, searching standard theme directories
func LoadTheme(themeName string) (*Theme, error) {
	filePath, err := findThemeFile(themeName)
	if err != nil {
		return nil, err
	}

	return LoadThemeFromFile(filePath)
}

// configToTheme converts a ThemeConfig to a Theme, with fallback to Tokyo Night for missing colors
func configToTheme(config ThemeConfig) *Theme {
	// Start with Tokyo Night as base
	tokyoNight := TokyoNight()

	// Override with config values
	if config.Colors.TreeNormalText != "" {
		tokyoNight.Colors.TreeNormalText = ParseColorString(config.Colors.TreeNormalText)
	}
	if config.Colors.TreeSelectedItem != "" {
		tokyoNight.Colors.TreeSelectedItem = ParseColorString(config.Colors.TreeSelectedItem)
	}
	if config.Colors.TreeNewItem != "" {
		tokyoNight.Colors.TreeNewItem = ParseColorString(config.Colors.TreeNewItem)
	}
	if config.Colors.TreeLeafArrow != "" {
		tokyoNight.Colors.TreeLeafArrow = ParseColorString(config.Colors.TreeLeafArrow)
	}
	if config.Colors.TreeExpandedArrow != "" {
		tokyoNight.Colors.TreeExpandedArrow = ParseColorString(config.Colors.TreeExpandedArrow)
	}
	if config.Colors.TreeCollapsedArrow != "" {
		tokyoNight.Colors.TreeCollapsedArrow = ParseColorString(config.Colors.TreeCollapsedArrow)
	}
	if config.Colors.EditorText != "" {
		tokyoNight.Colors.EditorText = ParseColorString(config.Colors.EditorText)
	}
	if config.Colors.EditorCursor != "" {
		tokyoNight.Colors.EditorCursor = ParseColorString(config.Colors.EditorCursor)
	}
	if config.Colors.SearchLabel != "" {
		tokyoNight.Colors.SearchLabel = ParseColorString(config.Colors.SearchLabel)
	}
	if config.Colors.SearchText != "" {
		tokyoNight.Colors.SearchText = ParseColorString(config.Colors.SearchText)
	}
	if config.Colors.SearchCursor != "" {
		tokyoNight.Colors.SearchCursor = ParseColorString(config.Colors.SearchCursor)
	}
	if config.Colors.SearchResultCount != "" {
		tokyoNight.Colors.SearchResultCount = ParseColorString(config.Colors.SearchResultCount)
	}
	if config.Colors.CommandPrompt != "" {
		tokyoNight.Colors.CommandPrompt = ParseColorString(config.Colors.CommandPrompt)
	}
	if config.Colors.CommandText != "" {
		tokyoNight.Colors.CommandText = ParseColorString(config.Colors.CommandText)
	}
	if config.Colors.CommandCursor != "" {
		tokyoNight.Colors.CommandCursor = ParseColorString(config.Colors.CommandCursor)
	}
	if config.Colors.HelpBackground != "" {
		tokyoNight.Colors.HelpBackground = ParseColorString(config.Colors.HelpBackground)
	}
	if config.Colors.HelpBorder != "" {
		tokyoNight.Colors.HelpBorder = ParseColorString(config.Colors.HelpBorder)
	}
	if config.Colors.HelpTitle != "" {
		tokyoNight.Colors.HelpTitle = ParseColorString(config.Colors.HelpTitle)
	}
	if config.Colors.HelpContent != "" {
		tokyoNight.Colors.HelpContent = ParseColorString(config.Colors.HelpContent)
	}
	if config.Colors.StatusMode != "" {
		tokyoNight.Colors.StatusMode = ParseColorString(config.Colors.StatusMode)
	}
	if config.Colors.StatusMessage != "" {
		tokyoNight.Colors.StatusMessage = ParseColorString(config.Colors.StatusMessage)
	}
	if config.Colors.StatusModified != "" {
		tokyoNight.Colors.StatusModified = ParseColorString(config.Colors.StatusModified)
	}
	if config.Colors.HeaderTitle != "" {
		tokyoNight.Colors.HeaderTitle = ParseColorString(config.Colors.HeaderTitle)
	}

	if config.Name != "" {
		tokyoNight.Name = config.Name
	}

	return tokyoNight
}

// LoadThemeOrDefault loads a theme by name, or returns Tokyo Night if not found
func LoadThemeOrDefault(themeName string) *Theme {
	if themeName == "default" {
		return Default()
	}

	theme, err := LoadTheme(themeName)
	if err != nil {
		// Fall back to Tokyo Night
		return TokyoNight()
	}

	return theme
}
