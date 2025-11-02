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
│   │   └── json.go                  # File I/O
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

## Recent Changes (Latest Session)

1. **Optional Filename Argument**: tuo can now be started without a filename (`./tuo`) to work in memory
2. **Buffer Mode**: Start with an empty outline in memory, no backing file required
3. **Save with Filename**: Use `:w <filename>` to save the buffer to disk and set that as the working file
4. **SaveAs Functionality**: After saving with `:w filename`, subsequent `:w` commands save to the same file
5. **Empty Buffer Saves**: Attempting `:w` without a filename on a new buffer shows error message directing to use `:w <filename>`
6. **Markdown Export**: Added `:export markdown <filename>` command to export outlines as markdown with unordered list format
   - Exports hierarchy as nested bullet lists using `-` characters
   - 2 spaces per indentation level for clean formatting
   - Exports text content only (no metadata)
   - Includes outline title as a top-level markdown header
7. **Title Command**: Added `:title <text>` command to set the outline title
   - `:title` with no argument shows the current title
   - `:title My Title` sets the outline title to "My Title"
   - Supports multi-word titles
   - Marks outline as dirty when title is changed
8. **Page Up/Down Scrolling**: Added `Ctrl+U` (page up) and `Ctrl+D` (page down) keybindings for scrolling
   - Selection moves with the viewport (stays visible)
   - Page size adapts to terminal height
   - Ctrl+U previously outdented items; now use `<` or `,` for outdenting instead
   - Smart viewport management keeps selection within visible area
9. **Generic Key-Value Attributes System**: Added flexible attributes support to items
   - Extended Metadata struct with `Attributes map[string]string` field
   - Fully persisted to JSON with `json:"attributes,omitempty"`
   - Attributes are initialized for all new items
10. **Attribute Management Commands**:
   - `:attr add <key> <value>` - Add or update an attribute
   - `:attr del <key>` - Delete an attribute
   - `:attr` or `:attr list` - View all attributes for current item
11. **Attribute Keybindings** (with `a` prefix):
   - `aa` - Add attribute (maps to `:attr add` command instruction)
   - `ad` - Delete attribute (maps to `:attr del` command instruction)
   - `ac` - Change attribute (maps to `:attr add` command instruction)
   - `av` - View all attributes for current item
12. **URL Opening Feature**:
   - `go` keybinding (g then o) opens URL from `url` attribute
   - Uses `xdg-open` to launch URL in default application
   - Checks for `url` attribute, provides error message if not found
13. **Daily Notes Integration with Attributes**:
   - `:dailynote` command now automatically adds `type="day"` and `date="YYYY-MM-DD"` attributes
   - Allows navigation between daily notes based on date
14. **Date Navigation Enhancements**:
   - Date navigation functions (`[d`, `]d`, `[w`, `]w`, etc.) now recognize date attributes
   - Supports items with `date` attribute in YYYY-MM-DD format
   - Maintains backward compatibility with DueDate field
15. **Example Outline**:
   - Created `examples/attributes_demo.json` demonstrating all attribute features
   - Examples include daily notes, URLs, custom task attributes, and navigation
16. **Advanced Search Filter Syntax**: Complete rewrite of search functionality with powerful filter language
   - Created `internal/search/` package with modular design
   - **Parser** (`parser.go`): Tokenizer and recursive descent parser that builds s-expression trees
   - **Filter Expressions** (`expr.go`): Comprehensive FilterExpr interface with 10+ filter types
   - **Filter Types**:
     - Text search (case-insensitive substring)
     - Depth filters: `d:>2`, `d:<=1`, etc.
     - Attribute filters: `a:type=day`, `a:status=done`, `a:url`
     - Date filters: `c:>-7d` (created), `m:<-30d` (modified)
     - Children count: `children:0` (leaf nodes), `children:>5`
     - Parent/Ancestor: `p:d:0`, `ancestor:a:type=project`
   - **Boolean Operators**:
     - Implicit AND: `task project` → `task AND project`
     - Explicit AND: `task +project`
     - OR: `task | project`
     - NOT: `-task`, `-children:0`
     - Grouping: `(task | project) d:>0`
   - **Debug Function** (`debug.go`): Pretty-print expressions and explain match results
   - **Integration** (internal/ui/search.go): Replaced simple substring search with new parser
   - **Testing**: Comprehensive test suite with 100% parser coverage
   - **Documentation**:
     - `docs/search-syntax.md` - Complete syntax reference with examples
     - Updated README.md with search examples and syntax overview
     - Help screen shows common search patterns
   - **Example** (`examples/search_demo.json`): Demonstrates various attributes, depths, and dates

## Implementation Details

### Search Package Architecture (internal/search/)
- **Tokenizer** (`parser.go:Tokenizer`): Converts query string to tokens
  - Handles filters (e.g., `d:>2`, `a:type=day`), operators (`+`, `|`, `-`), and text
  - Recognizes quoted strings and complex filter criteria
- **Parser** (`parser.go:Parser`): Builds s-expression tree from tokens
  - Recursive descent parser with proper operator precedence
  - Precedence: OR < AND < NOT < Atoms
  - Supports parentheses for explicit grouping
- **Filter Expressions** (`expr.go`):
  - `FilterExpr` interface: `Matches(item) bool` and `String()` for debug output
  - Binary operators: `AndExpr`, `OrExpr`, `NotExpr`
  - Filter implementations: `TextExpr`, `DepthFilter`, `AttributeFilter`, `DateFilter`, `ChildrenFilter`, `ParentFilter`, `AncestorFilter`
  - Helper functions for depth calculation, date parsing, and comparisons
- **Debug Module** (`debug.go`):
  - `DebugMatch()` returns detailed match information
  - `evaluateWithReason()` explains why items matched/didn't match
  - `ExpressionString()` pretty-prints s-expressions

### Search Integration (internal/ui/search.go)
- `updateResults()` now parses query and evaluates filter expression
- Error handling: Parse errors shown in search bar (e.g., "error: missing value")
- Fields added:
  - `filterExpr FilterExpr` - Parsed filter expression
  - `parseError string` - Error from parsing query
- `GetParseError()` method for accessing parse errors

### Attribute Date Filtering (internal/search/expr.go)
- New `AttributeDateFilter` type for date comparisons on attributes
- Automatically detects when attribute filter value is a date (YYYY-MM-DD or relative like `-7d`)
- Supports all comparison operators: `>`, `>=`, `<`, `<=`, `=`, `!=`
- Examples:
  - `@deadline>-7d` - Attributes with dates in next 7 days
  - `@date>=2025-11-01` - Attribute dates on or after November 1st
  - `@completed<-30d` - Attribute dates older than 30 days
- Integrated into `parseAttrFilter()` in parser which auto-detects date values

### Search Syntax Updates
- Changed attribute filter prefix from `a:` to `@` (no colon needed)
  - Old: `a:type=day`, `a:url`, `a:date>-7d`
  - New: `@type=day`, `@url`, `@date>-7d`
- Changed ancestor filter prefix from `ancestor:` to `a:` (similar to parent `p:`)
  - Old: `ancestor:a:type=project`
  - New: `a:@type=project`
- Tokenizer updated to recognize `@` as filter start character
  - `readAttrFilter()` method handles `@key` syntax without requiring colon
  - Parser routes `@` prefixed filters through attribute filter logic

### Data Model Changes (internal/model/outline.go)
- Added `Attributes map[string]string` field to Metadata struct
- Initialized attributes map in NewItem() to prevent nil pointer errors
- JSON serialization with `omitempty` tag for clean JSON output

### Command Handling (internal/app/app.go)
- `handleAttrCommand(parts []string)` processes `:attr` commands
- `showAttributes(item *model.Item)` displays attributes in status bar
- `handleGoCommand()` opens URLs with xdg-open
- Modified `:dailynote` to auto-create `type` and `date` attributes

### Keybindings (internal/app/keybindings.go)
- Added `'a'` as pending key prefix for attribute operations
- Added `'o'` to `'g'` prefix for `go` command (URL opening)
- Keybindings provide status messages directing users to command mode

### Navigation Functions (internal/ui/tree.go)
- `FindNextDateItem()` now checks both DueDate and date attribute
- `FindPrevDateItem()` now checks both DueDate and date attribute
- `FindNextItemWithDateInterval()` parses date attributes in YYYY-MM-DD format
- `FindPrevItemWithDateInterval()` parses date attributes in YYYY-MM-DD format

17. **Node Search Widget Enhancement (Ctrl+K)**:
   - Integrated advanced search parser into Node Search Widget
   - Widget now supports all filter expressions: depth, attributes, dates, children, parent, ancestor
   - Features:
     - Real-time search as you type
     - Advanced filter syntax same as main search (`/`)
     - Error handling for invalid queries (falls back to text search)
     - Results limited to 10 matches for performance
     - Display of parse errors in red in the widget
     - Support for hoisting (Alt+Enter) and selection (Enter)
   - Implementation (internal/ui/node_search_widget.go):
     - Added `parseError` and `filterExpr` fields
     - `updateMatches()` now uses `search.ParseQuery()` instead of fuzzy search
     - Error handling with fallback to text-only matching on parse failure
     - Render function updated to display parse errors in red
   - Documentation: `docs/node-search-widget.md` - complete guide with examples

18. **Vim-like Configuration System**:
   - Added `:set` command for configurable settings (session-only by default)
   - Extended `internal/config/config.go` with `Set()`, `Get()`, and `GetAll()` methods
   - Added `cfg` field to `App` struct to store configuration reference
   - Implemented `handleSetCommand()` in `internal/app/app.go` with support for:
     - `:set key value` - Set a configuration value
     - `:set key` - Show value of a specific setting
     - `:set` - Show all currently configured settings
     - Automatic quote removal for quoted values
   - Example usage:
     - `:set visattr date` - Configure visible attributes
     - `:set visattr "my long value"` - Quoted values supported
     - `:set` - List all settings
   - Configuration values stored in memory (session-only)
   - Comprehensive test suite in `internal/config/config_test.go`
   - Documentation: `docs/configuration.md` - complete configuration guide
   - Example: `examples/config_demo.json` - demonstrates configuration features

19. **Visible Attributes Display (visattr)**:
   - Added support for displaying item attributes inline with tree view
   - Use `:set visattr <attribute-names>` to specify which attributes to display
   - Multiple attributes supported: `:set visattr date,status,type`
   - Comma-separated attribute names in the configuration
   - Attributes display format: `Item text  [attr1:value1, attr2:value2]` in gray
   - Gray/dim style for attribute values maintains visual hierarchy
   - Attributes only display if they exist and have non-empty values
   - Selected items show attributes in the selected item style
   - Implementation:
     - Added `TreeAttributeValue` color to theme system
     - Updated `TreeView.Render()` and `RenderWithSearchQuery()` signatures to accept config
     - Attribute rendering logic in `internal/ui/tree.go` (lines 868-902)
     - New `TreeAttributeStyle()` method in `internal/ui/screen.go` for attribute styling
     - Theme color support: `tree_attribute_value` in TOML theme files
   - Example usage:
     - `:set visattr date` - Show date attribute for items that have it
     - `:set visattr date,status,priority` - Show multiple attributes
     - Items without the configured attributes simply don't show them
     - Empty attribute values are not displayed
   - Example file: `examples/visattr_demo.json` demonstrates the feature

20. **Go to Referenced Item Keybinding (gr)**:
   - Added `gr` keybinding to navigate from a virtual reference (item shown in search nodes) to its original item in the tree
   - Implementation:
     - Added `'r'` keybinding to the `'g'` pending keybinding group (gr = go referenced)
     - Implemented `handleGoReferencedCommand()` in `internal/app/app.go`
     - Checks if current item is a virtual reference using `displayItem.IsVirtual` flag
     - Shows error in status if not a reference or if reference can't be resolved
     - Navigates to the original item by:
       1. Getting the original item from `displayItem.OriginalItem`
       2. Expanding parent nodes to make it visible with `tree.ExpandParents()`
       3. Finding and selecting the original item in the display items
       4. Displaying a status message with the referenced item's text
   - Error messages:
     - "No items in tree" - No items exist
     - "No item selected" - Selection out of bounds
     - "Current item is not a reference" - Not a virtual reference
     - "Reference has no original item" - Broken reference
     - "Could not navigate to referenced item" - Item not found after expanding parents
   - Works with search nodes that have virtual children (created with virtual_children field)
   - Example: In `examples/reference_test.json`, search nodes show referenced items; press `gr` to jump to original

## Notes

- The application uses the `tcell` library for terminal UI
- All files use UTF-8 encoding
- JSON is the persistence format for outlines
- The application includes auto-save (5 seconds of inactivity)
- When adding docs, add these to the ./docs/ directory.
- always create test outlines in examples/ directory
