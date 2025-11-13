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
		Background            string `toml:"background"`
		TreeNormalText        string `toml:"tree_normal_text"`
		TreeSelectedItem      string `toml:"tree_selected_item"`
		TreeSelectedBg        string `toml:"tree_selected_bg"`
		TreeNewItem           string `toml:"tree_new_item"`
		TreeLeafArrow         string `toml:"tree_leaf_arrow"`
		TreeExpandableArrow   string `toml:"tree_expandable_arrow"`
		TreeExpandedArrow     string `toml:"tree_expanded_arrow"`
		TreeCollapsedArrow    string `toml:"tree_collapsed_arrow"`
		TreeVisualSelection   string `toml:"tree_visual_selection"`
		TreeVisualSelectionBg string `toml:"tree_visual_selection_bg"`
		TreeVisualCursor      string `toml:"tree_visual_cursor"`
		TreeVisualCursorBg    string `toml:"tree_visual_cursor_bg"`
		TreeAttributeIndicator string `toml:"tree_attribute_indicator"`
		TreeAttributeValue    string `toml:"tree_attribute_value"`
		TreeTagValue          string `toml:"tree_tag_value"`
		EditorText        string `toml:"editor_text"`
		EditorCursor      string `toml:"editor_cursor"`
		EditorCursorBg    string `toml:"editor_cursor_bg"`
		SearchLabel       string `toml:"search_label"`
		SearchText        string `toml:"search_text"`
		SearchCursor      string `toml:"search_cursor"`
		SearchCursorBg    string `toml:"search_cursor_bg"`
		SearchResultCount string `toml:"search_result_count"`
		CommandPrompt     string `toml:"command_prompt"`
		CommandText       string `toml:"command_text"`
		CommandCursor     string `toml:"command_cursor"`
		CommandCursorBg   string `toml:"command_cursor_bg"`
		HelpBackground    string `toml:"help_background"`
		HelpBorder        string `toml:"help_border"`
		HelpTitle         string `toml:"help_title"`
		HelpContent       string `toml:"help_content"`
		StatusMode        string `toml:"status_mode"`
		StatusModeBg      string `toml:"status_mode_bg"`
		StatusMessage     string `toml:"status_message"`
		StatusModified    string `toml:"status_modified"`
		HeaderTitle           string `toml:"header_title"`
		HeaderBg              string `toml:"header_bg"`
		CalendarDayText       string `toml:"calendar_day_text"`
		CalendarDayBg         string `toml:"calendar_day_bg"`
		CalendarInactiveDayText string `toml:"calendar_inactive_day_text"`
		CalendarInactiveDayBg   string `toml:"calendar_inactive_day_bg"`
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

// configToTheme converts a ThemeConfig to a Theme, with fallback to Default for missing colors
func configToTheme(config ThemeConfig) *Theme {
	// Start with Default as base
	baseTheme := Default()

	// Override with config values
	if config.Colors.Background != "" {
		baseTheme.Colors.Background = ParseColorString(config.Colors.Background)
	}
	if config.Colors.TreeNormalText != "" {
		baseTheme.Colors.TreeNormalText = ParseColorString(config.Colors.TreeNormalText)
	}
	if config.Colors.TreeSelectedItem != "" {
		baseTheme.Colors.TreeSelectedItem = ParseColorString(config.Colors.TreeSelectedItem)
	}
	if config.Colors.TreeSelectedBg != "" {
		baseTheme.Colors.TreeSelectedBg = ParseColorString(config.Colors.TreeSelectedBg)
	}
	if config.Colors.TreeNewItem != "" {
		baseTheme.Colors.TreeNewItem = ParseColorString(config.Colors.TreeNewItem)
	}
	if config.Colors.TreeLeafArrow != "" {
		baseTheme.Colors.TreeLeafArrow = ParseColorString(config.Colors.TreeLeafArrow)
	}
	if config.Colors.TreeExpandableArrow != "" {
		baseTheme.Colors.TreeExpandableArrow = ParseColorString(config.Colors.TreeExpandableArrow)
	}
	if config.Colors.TreeExpandedArrow != "" {
		baseTheme.Colors.TreeExpandedArrow = ParseColorString(config.Colors.TreeExpandedArrow)
	}
	if config.Colors.TreeCollapsedArrow != "" {
		baseTheme.Colors.TreeCollapsedArrow = ParseColorString(config.Colors.TreeCollapsedArrow)
	}
	if config.Colors.TreeVisualSelection != "" {
		baseTheme.Colors.TreeVisualSelection = ParseColorString(config.Colors.TreeVisualSelection)
	}
	if config.Colors.TreeVisualSelectionBg != "" {
		baseTheme.Colors.TreeVisualSelectionBg = ParseColorString(config.Colors.TreeVisualSelectionBg)
	}
	if config.Colors.TreeVisualCursor != "" {
		baseTheme.Colors.TreeVisualCursor = ParseColorString(config.Colors.TreeVisualCursor)
	}
	if config.Colors.TreeVisualCursorBg != "" {
		baseTheme.Colors.TreeVisualCursorBg = ParseColorString(config.Colors.TreeVisualCursorBg)
	}
	if config.Colors.TreeAttributeIndicator != "" {
		baseTheme.Colors.TreeAttributeIndicator = ParseColorString(config.Colors.TreeAttributeIndicator)
	}
	if config.Colors.TreeAttributeValue != "" {
		baseTheme.Colors.TreeAttributeValue = ParseColorString(config.Colors.TreeAttributeValue)
	}
	if config.Colors.TreeTagValue != "" {
		baseTheme.Colors.TreeTagValue = ParseColorString(config.Colors.TreeTagValue)
	}
	if config.Colors.EditorText != "" {
		baseTheme.Colors.EditorText = ParseColorString(config.Colors.EditorText)
	}
	if config.Colors.EditorCursor != "" {
		baseTheme.Colors.EditorCursor = ParseColorString(config.Colors.EditorCursor)
	}
	if config.Colors.EditorCursorBg != "" {
		baseTheme.Colors.EditorCursorBg = ParseColorString(config.Colors.EditorCursorBg)
	}
	if config.Colors.SearchLabel != "" {
		baseTheme.Colors.SearchLabel = ParseColorString(config.Colors.SearchLabel)
	}
	if config.Colors.SearchText != "" {
		baseTheme.Colors.SearchText = ParseColorString(config.Colors.SearchText)
	}
	if config.Colors.SearchCursor != "" {
		baseTheme.Colors.SearchCursor = ParseColorString(config.Colors.SearchCursor)
	}
	if config.Colors.SearchCursorBg != "" {
		baseTheme.Colors.SearchCursorBg = ParseColorString(config.Colors.SearchCursorBg)
	}
	if config.Colors.SearchResultCount != "" {
		baseTheme.Colors.SearchResultCount = ParseColorString(config.Colors.SearchResultCount)
	}
	if config.Colors.CommandPrompt != "" {
		baseTheme.Colors.CommandPrompt = ParseColorString(config.Colors.CommandPrompt)
	}
	if config.Colors.CommandText != "" {
		baseTheme.Colors.CommandText = ParseColorString(config.Colors.CommandText)
	}
	if config.Colors.CommandCursor != "" {
		baseTheme.Colors.CommandCursor = ParseColorString(config.Colors.CommandCursor)
	}
	if config.Colors.CommandCursorBg != "" {
		baseTheme.Colors.CommandCursorBg = ParseColorString(config.Colors.CommandCursorBg)
	}
	if config.Colors.HelpBackground != "" {
		baseTheme.Colors.HelpBackground = ParseColorString(config.Colors.HelpBackground)
	}
	if config.Colors.HelpBorder != "" {
		baseTheme.Colors.HelpBorder = ParseColorString(config.Colors.HelpBorder)
	}
	if config.Colors.HelpTitle != "" {
		baseTheme.Colors.HelpTitle = ParseColorString(config.Colors.HelpTitle)
	}
	if config.Colors.HelpContent != "" {
		baseTheme.Colors.HelpContent = ParseColorString(config.Colors.HelpContent)
	}
	if config.Colors.StatusMode != "" {
		baseTheme.Colors.StatusMode = ParseColorString(config.Colors.StatusMode)
	}
	if config.Colors.StatusModeBg != "" {
		baseTheme.Colors.StatusModeBg = ParseColorString(config.Colors.StatusModeBg)
	}
	if config.Colors.StatusMessage != "" {
		baseTheme.Colors.StatusMessage = ParseColorString(config.Colors.StatusMessage)
	}
	if config.Colors.StatusModified != "" {
		baseTheme.Colors.StatusModified = ParseColorString(config.Colors.StatusModified)
	}
	if config.Colors.HeaderTitle != "" {
		baseTheme.Colors.HeaderTitle = ParseColorString(config.Colors.HeaderTitle)
	}
	if config.Colors.HeaderBg != "" {
		baseTheme.Colors.HeaderBg = ParseColorString(config.Colors.HeaderBg)
	}
	if config.Colors.CalendarDayText != "" {
		baseTheme.Colors.CalendarDayText = ParseColorString(config.Colors.CalendarDayText)
	}
	if config.Colors.CalendarDayBg != "" {
		baseTheme.Colors.CalendarDayBg = ParseColorString(config.Colors.CalendarDayBg)
	}
	if config.Colors.CalendarInactiveDayText != "" {
		baseTheme.Colors.CalendarInactiveDayText = ParseColorString(config.Colors.CalendarInactiveDayText)
	}
	if config.Colors.CalendarInactiveDayBg != "" {
		baseTheme.Colors.CalendarInactiveDayBg = ParseColorString(config.Colors.CalendarInactiveDayBg)
	}

	if config.Name != "" {
		baseTheme.Name = config.Name
	}

	return baseTheme
}

// LoadThemeOrDefault loads a theme by name, or returns Default if not found
func LoadThemeOrDefault(themeName string) *Theme {
	if themeName == "default" {
		return Default()
	}

	theme, err := LoadTheme(themeName)
	if err != nil {
		// Fall back to Default
		return Default()
	}

	return theme
}
