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
# Global background (optional)
background = "#1a1b26"

# Tree view colors
tree_normal_text = "#c0caf5"
tree_selected_item = "#7aa2f7"
tree_selected_bg = "#283457"         # Background for selected items
tree_new_item = "#565f89"
tree_leaf_arrow = "#565f89"
tree_expandable_arrow = "#7dcfff"
tree_expanded_arrow = "#7dcfff"
tree_collapsed_arrow = "#7dcfff"

# Editor colors
editor_text = "#c0caf5"
editor_cursor = "#1a1b26"
editor_cursor_bg = "#7aa2f7"         # Background for cursor

# Search bar colors
search_label = "#bb9af7"
search_text = "#c0caf5"
search_cursor = "#1a1b26"
search_cursor_bg = "#7aa2f7"         # Background for cursor
search_result_count = "#9ece6a"

# Command line colors
command_prompt = "#bb9af7"
command_text = "#c0caf5"
command_cursor = "#1a1b26"
command_cursor_bg = "#7aa2f7"        # Background for cursor

# Help overlay colors
help_background = "#1a1b26"
help_border = "#7dcfff"
help_title = "#bb9af7"
help_content = "#c0caf5"

# Status line colors
status_mode = "#bb9af7"
status_mode_bg = "#1a1b26"           # Background for mode indicator
status_message = "#9ece6a"
status_modified = "#f7768e"

# Header colors
header_title = "#bb9af7"
header_bg = "#1a1b26"                # Background for header
```

**Note:** All color fields are optional. If a color is not specified in the TOML file, the Tokyo Night theme's default will be used.

## Built-in Themes

### Tokyo Night

The application comes with a built-in Tokyo Night theme that uses the official Tokyo Night color palette:

- **Background**: #1a1b26 (Dark midnight blue)
- **Foreground**: #c0caf5 (Light gray-blue)
- **Selected Items**: Bright cyan text (#7dcfff) on selection blue background (#283457) for high contrast
- **Accents**:
  - Blue (#7aa2f7) for cursor highlights and marks
  - Cyan (#7dcfff) for tree navigation arrows on expandable nodes and selected text
  - Gray (#565f89) for tree navigation arrows on leaf nodes (dimmer)
  - Magenta (#bb9af7) for headers and labels
  - Green (#9ece6a) for status messages
  - Red (#f7768e) for modified indicators

**Tree Navigation Arrows:**
- Leaf nodes (no children): Dim gray (#565f89) - visually indicates they cannot be expanded
- Expandable nodes (with children): Bright cyan (#7dcfff) - visually indicates expandable content

**Selection Highlight:**
- Selected items use bright cyan text (#7dcfff) on selection blue background (#283457)
- This provides excellent contrast for visibility while maintaining the Tokyo Night aesthetic

### Osaka Jade

A warm, earthy theme with jade and gold accents, inspired by traditional Japanese aesthetics:

- **Background**: #111C18 (Dark jade)
- **Foreground**: #C1C497 (Light warm jade)
- **Selected Items**: Dark jade text (#111C18) on light jade background (#C1C497) - inverted for contrast
- **Accents**:
  - Bright cyan (#2DD5B7) for tree navigation arrows on expandable nodes
  - Magenta (#D2689C) for headers and labels
  - Green (#549E6A) for status messages
  - Red (#FF5345) for modified indicators
  - Warm gold (#D7C995) for cursor backgrounds with black text for maximum contrast
  - Dark gray-green (#53685B) for new/placeholder items and leaf arrows

**Color Palette:**
The Osaka Jade theme uses a sophisticated color palette that evokes natural materials:
- Warm jade tones for primary text
- Rich earth tones for accents
- High-contrast cursor backgrounds in warm gold
- Bright cyan highlights for interactive elements

The Osaka Jade theme is available via TOML file configuration: `~/.local/share/tui-outliner/themes/osaka-jade.toml`

## Usage

### Selecting a Theme

You can select which theme to use by editing the configuration file at:

```
~/.config/tui-outliner/config.toml
```

**Example config file:**

```toml
# Theme selection - choose from available themes
theme = "tokyo-night"
```

### Available Options

**Built-in Themes:**
- `tokyo-night` - Cool blue and purple tones (default)
- `osaka-jade` - Warm jade and gold earth tones
- `default` - Use terminal default colors

**Custom Themes:**
To create or use a custom theme, place a TOML file in:
```
~/.local/share/tui-outliner/themes/{themeName}.toml
```

Then reference it in the config file:
```toml
theme = "my-custom-theme"
```

### First Time Setup

If the config file doesn't exist, the application will:
1. Use the Tokyo Night theme by default
2. You can create `~/.config/tui-outliner/config.toml` to customize the theme

### Creating a Custom Theme

See the **Theme Configuration** section above for the complete TOML format. You can use Tokyo Night or Osaka Jade as a template.

## Color Elements

The theme system supports both foreground and background colors for UI elements:

### Global
- Background (used as fallback for various elements)

### Tree View
- Normal text (foreground color)
- Selected items (foreground color + background)
- New/placeholder items (foreground color)
- Leaf node arrows (foreground color, dimmer)
- Expandable node arrows (foreground color, brighter)
- Expanded/collapsed arrows (foreground color)

### Editor
- Text color (foreground)
- Cursor color + background

### Search Bar
- Label (foreground)
- Search text (foreground)
- Cursor (foreground + background)
- Result count (foreground)

### Command Line
- Prompt (foreground)
- Command text (foreground)
- Cursor (foreground + background)

### Help Overlay
- Background (background color)
- Borders (foreground)
- Title (foreground)
- Content text (foreground)

### Status Line
- Mode indicator (foreground + background)
- Status messages (foreground)
- Modified indicator (foreground)

### Header
- Title (foreground + background)

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
