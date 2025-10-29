# TUI Outliner

A powerful, keyboard-driven outliner application for the terminal, built in Go. Organize your thoughts, projects, and tasks in a hierarchical tree structure with rich metadata support.

## Features

- **Hierarchical Tree Structure**: Organize items in a nested, expandable/collapsible tree
- **Rich Metadata**: Add tags, priorities, due dates, and notes to each item
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
go build -o tui-outliner
```

### Run

```bash
./tui-outliner <outline_file.json>
```

If the file doesn't exist, it will be created when you save.

## Quick Start

1. Create or open an outline file:
```bash
./tui-outliner my_outline.json
```

2. Start editing with these basic commands:
- `j/k` or `↓/↑` - Navigate items
- `i` - Edit selected item text
- `o` - Insert new item after selected
- `a` - Insert new child item
- `d` - Delete selected item
- `l/h` or `→/←` - Expand/collapse items
- `>/<` or `Ctrl+I/Ctrl+U` - Indent/outdent items
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
| `i` | Edit selected item |
| `o` | Insert new item after |
| `a` | Insert new child item |
| `d` | Delete selected item |

### Tree Manipulation

| Key | Action |
|-----|--------|
| `>` / `.` / `Ctrl+I` | Indent item (increase nesting) |
| `<` / `,` / `Ctrl+U` | Outdent item (decrease nesting) |

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
| `:w` | `:write` | Save the outline |
| `:q` | `:quit` | Quit (warns if unsaved) |
| `:q!` | `:quit!` | Force quit without saving |
| `:wq` | | Save and quit |
| `:help` | | Show help screen |
| `:debug` | | Toggle debug mode |

Examples:
```
:w           # Save
:q           # Quit (if saved)
:wq          # Save and quit
:q!          # Force quit
:help        # Show keybindings
```

## File Format

Outlines are stored as JSON files with the following structure:

```json
{
  "title": "My Outline",
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
            "priority": "high",
            "due_date": "2025-12-31T00:00:00Z",
            "notes": "Additional notes",
            "created": "2025-10-29T00:00:00Z",
            "modified": "2025-10-29T00:00:00Z"
          }
        }
      ],
      "metadata": {
        "tags": ["tag"],
        "priority": "high",
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
./tui-outliner examples/sample.json
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

When you press `i` to edit an item's text:

| Key | Action |
|-----|--------|
| `Enter` | Save changes and exit edit mode |
| `Escape` | Cancel and exit edit mode |
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
- No metadata editing UI yet (metadata can be added via JSON editing)
- Search is case-insensitive and matches full text

## Troubleshooting

### Keybindings Not Working

If keybindings don't respond:

1. **Run with debug mode** to see what keys are being detected:
   ```bash
   ./tui-outliner -debug examples/sample.json
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

- [ ] Metadata editing UI for tags, priorities, due dates
- [ ] Multiple documents with tabs
- [ ] Export to Markdown/OPML formats
- [ ] Undo/redo functionality
- [ ] Custom keybinding configuration
- [ ] Themes and color customization
- [ ] Quick filters (by priority, tags, due date)
- [ ] Vi command line mode (:w, :q, :wq)

## Development

### Building

```bash
go build -o tui-outliner
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
