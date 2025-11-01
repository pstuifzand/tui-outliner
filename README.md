# tuo - TUI Outliner

A powerful, keyboard-driven outliner application for the terminal, built in Go. Organize your thoughts, projects, and tasks in a hierarchical tree structure with rich metadata support.

**tuo** is the command-line application name for the TUI Outliner.

## Features

- **Hierarchical Tree Structure**: Organize items in a nested, expandable/collapsible tree
- **Flexible Attributes**: Add custom key-value attributes to items for rich metadata (tags, URLs, statuses, etc.)
- **File Persistence**: Save and load outlines from JSON files with auto-save support
- **Search & Filter**: Quickly find items by text search
- **Keyboard-Driven**: Vim-style keybindings for efficient navigation and editing
- **TUI Interface**: Full-featured terminal UI built with tcell

## Installation

### Prerequisites

- Go 1.16 or higher

### Build from Source

```bash
cd tui-outliner
go build -o tuo
```

### Run

```bash
./tuo [outline_file.json]
```

- If a filename is provided, it loads that file (or creates it on save)
- If no filename is provided, tuo starts with an empty outline in memory
- Use `:w <filename>` to save the outline to a file

## Quick Start

1. Start tuo with a file or start with an empty outline:
```bash
./tuo my_outline.json    # Open specific file
./tuo                     # Start with empty outline in memory
```

2. Start editing with these basic commands:
- `j/k` or `↓/↑` - Navigate items
- `i` - Edit selected item text
- `o` - Insert new item after selected
- `A` - Append text (edit at end of current item)
- `d` - Delete selected item
- `l/h` or `→/←` - Expand/collapse items
- `>/<` or `Ctrl+I` - Indent/outdent items (also `<` / `,`)
- `Ctrl+U` / `Ctrl+D` - Page up/down (scroll viewport)
- `/` - Search/filter items
- `Ctrl+S` - Save
- `?` - Show help
- `:q` - Quit

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `Down` | Move down |
| `k` / `Up` | Move up |
| `h` / `Left` | Collapse item |
| `l` / `Right` | Expand item |

### Editing

| Key | Action |
|-----|--------|
| `i` | Edit selected item (cursor at start) |
| `A` | Append text (edit at end of item) |
| `c` | Change (replace all item text) |
| `o` | Insert new item after |
| `O` | Insert new item before |
| `d` | Delete selected item |

### Tree Manipulation

| Key | Action |
|-----|--------|
| `>` / `.` / `Ctrl+I` | Indent item (increase nesting) |
| `<` / `,` | Outdent item (decrease nesting) |

### Scrolling

| Key | Action |
|-----|--------|
| `Ctrl+U` | Page up (scroll viewport) |
| `Ctrl+D` | Page down (scroll viewport) |

### Other

| Key | Action |
|-----|--------|
| `/` | Search/filter items |
| `?` | Toggle help screen |
| `:` | Enter command mode |
| `Ctrl+S` | Save outline |
| `Escape` | Close dialogs/cancel edit |

## Command Mode

Press `:` to enter command mode, then type a command and press Enter:

| Command | Alias | Action |
|---------|-------|--------|
| `:w` | `:write` | Save the outline to current file |
| `:w <file>` | `:write <file>` | Save the outline to a specific file |
| `:export markdown <file>` | | Export outline as markdown (unordered list format) |
| `:q` | `:quit` | Quit (warns if unsaved) |
| `:q!` | `:quit!` | Force quit without saving |
| `:wq` | | Save and quit |
| `:help` | | Show help screen |
| `:debug` | | Toggle debug mode |
| `:attr add <key> <value>` | | Add or update an attribute on selected item |
| `:attr del <key>` | | Delete an attribute from selected item |
| `:attr list` (or `:attr`) | | Show all attributes for selected item |

Examples:
```
:w                    # Save to current file
:w backup.json        # Save to a new file
:export markdown notes.md  # Export as markdown
:attr add type task   # Add a 'type' attribute with value 'task'
:attr add url https://example.com  # Add a URL attribute
:attr del type        # Remove the 'type' attribute
:attr                 # Show all attributes for current item
:q                    # Quit (if saved)
:wq                   # Save and quit
:q!                   # Force quit
:help                 # Show keybindings
```

## Attributes

Items can have custom key-value attributes for rich metadata. Attributes are useful for:
- Adding custom properties to items (e.g., status, assignee, effort)
- Storing URLs and other links
- Tagging items with domain-specific metadata
- Creating daily notes with date information

### Managing Attributes

**Via Commands:**
- `:attr add <key> <value>` - Add or update an attribute
- `:attr del <key>` - Delete an attribute
- `:attr` or `:attr list` - View all attributes for the current item

**Via Keybindings:**
- `av` - View all attributes (shortcut for `:attr`)
- `go` - Open URL from `url` attribute with `xdg-open` (g then o)

### Examples

```json
{
  "text": "Important Link",
  "metadata": {
    "attributes": {
      "url": "https://example.com",
      "type": "reference"
    }
  }
}
```

```json
{
  "text": "2025-10-30",
  "metadata": {
    "attributes": {
      "type": "day",
      "date": "2025-10-30"
    }
  }
}
```

```json
{
  "text": "Project Task",
  "metadata": {
    "attributes": {
      "status": "in-progress",
      "priority": "high",
      "assignee": "Alice",
      "effort": "3-days"
    }
  }
}
```

### Special Attributes

- **date**: Items with a date attribute (in YYYY-MM-DD format) can be navigated with date-based commands ([d, ]d, etc.)
- **type**: Custom item type indicators (e.g., "day" for daily notes)
- **url**: URLs that can be opened with the `go` command (uses xdg-open)

## File Format

Outlines are stored as JSON files with the following structure:

```json
{
  "items": [
    {
      "id": "unique_id",
      "text": "Item text",
      "children": [
        {
          "id": "child_id",
          "text": "Child item",
          "metadata": {
            "tags": ["tag1", "tag2"],
            "notes": "Additional notes",
            "attributes": {
              "url": "https://example.com",
              "type": "reference",
              "priority": "high"
            },
            "created": "2025-10-29T00:00:00Z",
            "modified": "2025-10-29T00:00:00Z"
          }
        }
      ],
      "metadata": {
        "tags": ["tag"],
        "created": "2025-10-29T00:00:00Z",
        "modified": "2025-10-29T00:00:00Z"
      }
    }
  ]
}
```

## Examples

Check the `examples/` directory for sample outline files:

```bash
./tuo examples/sample.json
```

## Project Structure

```
tui-outliner/
├── main.go                 # Entry point
├── internal/
│   ├── model/
│   │   └── outline.go      # Data structures
│   ├── storage/
│   │   └── json.go         # JSON persistence
│   ├── ui/
│   │   ├── screen.go       # Terminal screen management
│   │   ├── tree.go         # Tree view and navigation
│   │   ├── editor.go       # Text editing
│   │   ├── search.go       # Search and filter
│   │   └── help.go         # Help screen
│   └── app/
│       └── app.go          # Application controller
└── examples/
    └── sample.json         # Example outline
```

## Keyboard Shortcuts in Detail

### Edit Mode

When you press `i`, `A`, `c`, `o`, or `O` to edit an item's text:

| Key | Action |
|-----|--------|
| `Enter` | Save changes and create new item below |
| `Escape` | Cancel edit (deletes empty items, preserves non-empty) |
| `Ctrl+A` | Move to beginning of line |
| `Ctrl+E` | Move to end of line |
| `Ctrl+U` | Delete from start to cursor |
| `Ctrl+K` | Delete from cursor to end |
| `Backspace` | Delete character before cursor |
| `Delete` | Delete character at cursor |

### Search Mode

When you press `/` to search:

| Key | Action |
|-----|--------|
| `Escape` | Exit search mode |
| `Ctrl+A` / `Home` | Go to start of search query |
| `Ctrl+E` / `End` | Go to end of search query |

## Tips

1. **Auto-save**: The outline is automatically saved after every 5 seconds of inactivity
2. **Persistent expansion state**: Item expansion/collapse state is preserved in memory during the session (but not saved to file)
3. **Search highlights**: When searching, only matching items are shown
4. **Hierarchical operations**: When you indent/outdent items, their entire subtree moves with them

## Limitations

- Currently supports single outline files (no multi-document tabs yet)
- Search is case-insensitive and matches full text
- Attribute display in tree view not yet implemented (view via status bar with `:attr`)

## Troubleshooting

### Keybindings Not Working

If keybindings don't respond:

1. **Run with debug mode** to see what keys are being detected:
   ```bash
   ./tuo -debug examples/sample.json
   ```
   Every keypress will show in the status line with its key code and rune.

2. **Test basic navigation first:**
   - Arrow keys usually work universally
   - If arrow keys work but hjkl don't, your terminal may have a key mapping issue

3. **Check your terminal:**
   - Try a different terminal emulator (xterm, urxvt, GNOME Terminal, etc.)
   - Ensure it supports ANSI color codes
   - Make sure you're using a monospace font

4. **Check for terminal conflicts:**
   - Some terminals reserve certain key combinations
   - Try different key combinations from the help screen (`?`)

See `DEBUG.md` for more detailed troubleshooting steps.

## Future Enhancements

- [ ] Attribute display in tree view (inline rendering)
- [ ] Modal UI for attribute editing (instead of command mode)
- [ ] Multiple documents with tabs
- [ ] Export to Markdown/OPML formats
- [ ] Undo/redo functionality
- [ ] Custom keybinding configuration
- [ ] Themes and color customization
- [ ] Quick filters by attributes
- [ ] Vi command line mode (:w, :q, :wq)

## Development

### Building

```bash
go build -o tuo
```

To install the binary to your system:

```bash
go build -o tuo && mv tuo /usr/local/bin/
```

Then you can run `tuo` from anywhere:

```bash
tuo my_outline.json
```

### Running Tests

```bash
go test ./...
```

### Dependencies

- `github.com/gdamore/tcell/v2` - Terminal UI library

## License

MIT License - feel free to use, modify, and distribute

## Contributing

Contributions are welcome! Feel free to:
- Report issues
- Submit pull requests
- Suggest features
