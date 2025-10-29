# Quick Keybindings Reference

## Normal Mode (Default)

### Navigation
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Collapse item |
| `l` / `→` | Expand item |
| `gg` | Go to first node |
| `G` | Go to last node |
| `z...` | Reserved for future fold/zoom commands |

### Item Operations
| Key | Action |
|-----|--------|
| `i` | Edit item |
| `c` | Change (clear and edit) |
| `A` | Append (edit at end) |
| `o` | Insert new item after |
| `O` | Insert new item before |

### Node Manipulation
| Key | Action |
|-----|--------|
| `J` | Move node down |
| `K` | Move node up |
| `>` | Indent item |
| `<` | Outdent item |

### Clipboard Operations
| Key | Action |
|-----|--------|
| `y` | Yank (copy) item |
| `d` | Delete item (to clipboard) |
| `p` | Paste below |
| `P` | Paste above |

### Other
| Key | Action |
|-----|--------|
| `V` | Enter visual mode |
| `/` | Search |
| `?` | Toggle help |
| `:` | Command mode |

---

## Visual Mode (Multi-item Selection)

Enter with `V` in normal mode.

### Navigation / Selection
| Key | Action |
|-----|--------|
| `j` / `↓` | Extend selection down |
| `k` / `↑` | Extend selection up |
| `h` / `←` | Collapse item |
| `l` / `→` | Expand item |
| `gg` | Extend selection to first node |
| `G` | Extend selection to last node |
| `z...` | Reserved for future fold/zoom commands |

### Operations
| Key | Action |
|-----|--------|
| `d` / `x` | Delete selected items |
| `y` | Yank (copy) selected items |
| `>` | Indent selected items |
| `<` | Outdent selected items |

### Exit
| Key | Action |
|-----|--------|
| `V` | Exit visual mode |
| `Escape` | Exit visual mode (cancel) |

---

## Insert Mode (Text Editing)

Activated by pressing `i`, `c`, `A`, `o`, or `O` in normal mode.

| Key | Action |
|-----|--------|
| `Enter` | Save and create new item below |
| `Escape` | Save and exit to normal mode |
| Standard keys | Edit text (backspace, delete, etc.) |

---

## Search Mode

Activated by pressing `/` in normal mode.

| Key | Action |
|-----|--------|
| `Escape` | Exit search |
| Standard keys | Type search query |

---

## Command Mode

Activated by pressing `:` in normal mode.

| Key | Action |
|-----|--------|
| `Enter` | Execute command |
| `Escape` | Cancel |
| Standard keys | Type command |

### Available Commands
- `:w [filename]` - Save (optionally to new file)
- `:q` - Quit

---

## Tips

1. **Modal Editing**: tui-outliner uses Vim-style modes. Always return to normal mode when you're done editing.
2. **Visual Mode Power**: Use visual mode to perform bulk operations on multiple items at once.
3. **Help Screen**: Press `?` to toggle the help screen showing all available keybindings.
4. **Navigation**: Arrow keys work as alternatives to hjkl in all modes.
