# tuo - Build Instructions for Claude

This document provides quick reference for building and working with **tuo** (TUI Outliner).

## Application Name

- **Full Name**: TUI Outliner
- **Command Name**: `tuo`
- **Repository**: tui-outliner

## Prerequisites

- Go 1.16 or higher
- Standard Unix build tools (make, gcc, etc.) - optional

## Building

### Basic Build

Build the application from the project root:

```bash
cd /home/peter/work/tui-outliner
go build -o tuo
```

This creates a `tuo` executable in the current directory.

### Build and Install System-wide

To install `tuo` so it's available from anywhere:

```bash
go build -o tuo
sudo mv tuo /usr/local/bin/
```

Then you can run it from anywhere:

```bash
tuo my_outline.json
```

## Running

### Development (from project directory)

```bash
./tuo [outline_file.json]
```

Examples:
```bash
./tuo                           # Start with empty outline in memory
./tuo examples/sample.json      # Open specific file
./tuo my_notes.json             # Open or create file
```

When no file is specified, tuo starts with an empty outline. Use `:w <filename>` to save it to disk.

### After System Installation

```bash
tuo [outline_file.json]
```

### With Debug Mode

To debug key mappings and terminal input:

```bash
./tuo -debug                     # Debug mode with empty outline
./tuo -debug examples/sample.json    # Debug mode with specific file
```

## Project Structure

```
tui-outliner/
├── main.go                          # Entry point
├── go.mod / go.sum                  # Go module files
├── internal/
│   ├── app/
│   │   ├── app.go                   # Main app controller
│   │   ├── keybindings.go           # All keybindings and mode enum
│   │   └── ...
│   ├── model/
│   │   └── outline.go               # Data structures
│   ├── ui/
│   │   ├── editor.go                # Insert mode editor
│   │   ├── tree.go                  # Tree navigation and operations
│   │   ├── screen.go                # Terminal rendering
│   │   ├── search.go                # Search functionality
│   │   └── help.go                  # Help screen
│   ├── storage/
│   │   ├── json.go                  # File I/O
│   │   └── backup.go                # Automatic backup creation
│   ├── export/
│   │   └── markdown.go              # Markdown export functionality
│   ├── theme/
│   │   └── ...
│   └── config/
│       └── ...
├── README.md                        # User documentation
├── CLAUDE.md                        # This file
└── examples/
    └── sample.json                  # Example outline file
```

## Key Implementation Details

### Modes

The application uses an enum-based mode system defined in `internal/app/app.go`:

```go
type Mode int

const (
    NormalMode Mode = iota
    InsertMode
)
```

Replace all string-based mode checks (e.g., `if mode == "INSERT"`) with enum values.

### Editor Behavior

The editor (in `internal/ui/editor.go`) supports:

- **Insert Mode Operations**:
  - `i` - Edit from start
  - `A` - Append (edit from end)
  - `c` - Change (clear and replace)
  - `o` - Insert new item after
  - `O` - Insert new item before

- **Enter Key**: Creates new item below, stays in insert mode
- **Escape Key**: Exits insert mode; deletes empty items, preserves non-empty

### Tree Operations

Key methods in `internal/ui/tree.go`:

- `AddItemAfter(text)` - Insert after current
- `AddItemBefore(text)` - Insert before current (added recently)
- `DeleteItem(item)` - Delete specific item by reference
- `SelectNext()`, `SelectPrev()` - Navigation

### Export Functionality

Export functions are in `internal/export/markdown.go`:

- `ExportToMarkdown(outline *model.Outline, filePath string)` - Exports outline to markdown format with unordered list structure
  - Uses `-` for bullets with proper indentation (2 spaces per nesting level)
  - Exports only text content (no metadata)
  - Skips empty items while preserving structure
  - Includes outline title as a top-level header if present

Commands in `app.go`:
- `:export markdown <filename>` - Export current outline to markdown file

### Import Functionality

Import functions are in `internal/import/`:

- `parser.go` - Common interface and format detection
  - `ImportFile(content, format)` - Main entry point for importing
  - `DetectFormat(filename)` - Auto-detect format from file extension
- `markdown.go` - Markdown import parser
  - Parses markdown headers (# ## ###) as hierarchy levels
  - Parses unordered lists (- * +) with indentation
  - Supports plain text lines
- `indented.go` - Indented text import parser
  - Uses indentation (tabs or spaces) to define hierarchy
  - 2 spaces or 1 tab = 1 indentation level
  - Simple format for plain text outlines

Commands in `app.go`:
- `:import <filename> [format]` - Import items from file
  - Auto-detects format from extension (.md, .txt)
  - Optional format: `markdown`, `md`, `indented`, `text`, `txt`
  - Items imported as children of currently selected item
  - If no item selected, items added to root level
  - Automatically expands parent to show imported items

Supported formats:
- **Markdown** (.md) - Headers and lists with hierarchy
- **Indented Text** (.txt) - Plain text with indentation

Example files:
- `examples/recipe_import.txt` - Indented text format
- `examples/recipe_import.md` - Markdown format

## Common Development Tasks

### Add a New Keybinding

Edit `internal/app/keybindings.go` in the `InitializeKeybindings()` function:

```go
{
    Key:         'X',
    Description: "Description of action",
    Handler: func(app *App) {
        // Implementation
        app.SetStatus("Status message")
        app.dirty = true
    },
},
```

### Change Editor Behavior

Edit `internal/ui/editor.go`:
- `HandleKey()` - Process keypresses
- `WasEnterPressed()`, `WasEscapePressed()` - Check special key states
- `Stop()` - Save changes

### Modify Tree Display Logic

Edit `internal/ui/tree.go`:
- `buildDisplayItems()` - Control what appears on screen
- Filtering and expansion logic

## Testing

Run all tests:

```bash
go test ./...
```

Run tests for specific package:

```bash
go test ./internal/app
go test ./internal/ui
```

## Documentation Updates

After making code changes, update relevant sections in:

- `README.md` - User-facing documentation
- `CLAUDE.md` - This developer guide
- Inline code comments for complex logic

## Build Troubleshooting

### "go: command not found"

Install Go from https://golang.org/dl/ or your package manager:

```bash
# Ubuntu/Debian
sudo apt-get install golang-go

# macOS
brew install go

# Arch
sudo pacman -S go
```

### Module not found errors

Ensure you're in the project root and dependencies are present:

```bash
go mod download
go mod tidy
```

### Build fails with permission errors

Ensure you have write permissions in the directory:

```bash
ls -la /home/peter/work/tui-outliner/
```

## Notes

- The application uses the `tcell` library for terminal UI
- All files use UTF-8 encoding
- JSON is the persistence format for outlines
- The application includes auto-save (5 seconds of inactivity)
- When adding docs, add these to the ./docs/ directory.
- always create test outlines in examples/ directory
- **Kitty Keyboard Protocol**: Optional support for the Kitty keyboard protocol (disabled by default, see `docs/KITTY_KEYBOARD_PROTOCOL.md`)

### MoveUp and MoveDown

Valid positions are:

  1. Root level (parent == nil)
  2. Expanded nodes (parent.Expanded == true)
  3. Sibling positions (parent == current.Parent) - Stay as siblings regardless

