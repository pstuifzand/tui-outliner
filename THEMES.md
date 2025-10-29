# Theme System

The TUI Outliner now supports a flexible theming system using TOML configuration files. All colors in the application can be customized through themes.

## Theme Architecture

### Theme Package (`internal/theme/`)

The theme system consists of several components:

- **theme.go**: Core theme structures and built-in theme definitions
  - `Theme` - represents a complete theme with a name and color configuration
  - `Colors` - contains all color definitions for UI elements
  - `Default()` - returns the default terminal theme
  - `TokyoNight()` - returns the built-in Tokyo Night theme

- **colors.go**: Color parsing and style utilities
  - `HexToColor(string)` - converts hex color codes (#RRGGBB or #RGB) to tcell.Color
  - `RGBToColor(r, g, b int)` - converts RGB values to tcell.Color
  - `ParseColorString(string)` - handles multiple color formats
  - `ColorToStyle(tcell.Color)` - creates a style with foreground color
  - `ColorPairToStyle(fg, bg tcell.Color)` - creates style with foreground and background

- **loader.go**: Theme loading from TOML files
  - `LoadTheme(name string)` - loads a theme by name from standard locations
  - `LoadThemeFromFile(path string)` - loads a theme from a specific file
  - `LoadThemeOrDefault(name string)` - loads a theme or falls back to Tokyo Night

### Theme File Locations

Themes are searched in the following locations (in order):

1. `~/.config/tui-outliner/themes/`
2. `~/.local/share/tui-outliner/themes/`

Theme files should be named `{themeName}.toml` (e.g., `tokyo-night.toml`)

## Theme Configuration

### TOML Format

Themes are defined in TOML files with the following structure:

```toml
name = "theme-name"

[colors]
# Tree view colors
tree_normal_text = "#c0caf5"
tree_selected_item = "#7aa2f7"
tree_new_item = "#565f89"
tree_leaf_arrow = "#7dcfff"
tree_expanded_arrow = "#7dcfff"
tree_collapsed_arrow = "#7dcfff"

# Editor colors
editor_text = "#c0caf5"
editor_cursor = "#7aa2f7"

# Search bar colors
search_label = "#bb9af7"
search_text = "#c0caf5"
search_cursor = "#7aa2f7"
search_result_count = "#9ece6a"

# Command line colors
command_prompt = "#bb9af7"
command_text = "#c0caf5"
command_cursor = "#7aa2f7"

# Help overlay colors
help_background = "#1a1b26"
help_border = "#7dcfff"
help_title = "#bb9af7"
help_content = "#c0caf5"

# Status line colors
status_mode = "#bb9af7"
status_message = "#9ece6a"
status_modified = "#f7768e"

# Header colors
header_title = "#bb9af7"
```

## Built-in Themes

### Tokyo Night

The application comes with a built-in Tokyo Night theme that uses the official Tokyo Night color palette:

- **Background**: #1a1b26 (Dark midnight blue)
- **Foreground**: #c0caf5 (Light gray-blue)
- **Accents**:
  - Blue (#7aa2f7) for selections and highlights
  - Cyan (#7dcfff) for tree navigation arrows
  - Magenta (#bb9af7) for headers and labels
  - Green (#9ece6a) for status messages
  - Red (#f7768e) for modified indicators

The Tokyo Night theme is automatically applied when the application starts.

## Usage

### Current Implementation

The theme system is currently hardcoded to use the Tokyo Night theme. To use a different theme:

1. Create a new TOML file in `~/.local/share/tui-outliner/themes/{themeName}.toml`
2. Modify the `NewScreen()` function in `internal/ui/screen.go` to load the desired theme

Future versions may include:
- Config file support for theme selection
- Command-line flag for theme selection
- Theme selection in the application UI

## Color Elements

The theme system supports colors for all UI elements:

### Tree View
- Normal text
- Selected items
- New/placeholder items
- Leaf node arrows
- Expanded/collapsed arrows

### Editor
- Text color
- Cursor color

### Search Bar
- Label
- Search text
- Cursor
- Result count

### Command Line
- Prompt
- Command text
- Cursor

### Help Overlay
- Background
- Borders
- Title
- Content text

### Status Line
- Mode indicator
- Status messages
- Modified indicator

### Header
- Title

## File Locations

After implementation, the following files were created/modified:

### New Files
- `internal/theme/colors.go` - Color parsing utilities
- `internal/theme/theme.go` - Theme definitions
- `internal/theme/loader.go` - Theme loading functionality
- `~/.local/share/tui-outliner/themes/tokyo-night.toml` - Tokyo Night theme

### Modified Files
- `go.mod` - Added github.com/pelletier/go-toml/v2 dependency
- `internal/ui/screen.go` - Added theme-aware style methods
- `internal/ui/tree.go` - Updated to use themed colors
- `internal/ui/editor.go` - Updated to use themed colors
- `internal/ui/search.go` - Updated to use themed colors
- `internal/ui/command.go` - Updated to use themed colors
- `internal/ui/help.go` - Updated to use themed colors
- `internal/app/app.go` - Updated header and status rendering to use themes
