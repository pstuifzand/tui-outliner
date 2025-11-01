# Node Search Widget Guide

The Node Search Widget (activated with `Ctrl+K`) provides a quick way to navigate your outline using the same advanced search syntax as the main search mode (`/`).

## Overview

The Node Search Widget is a modal popup that lets you search for and navigate to nodes in your outline. It uses the same powerful filter expressions as the main search functionality.

**Key Features:**
- Advanced filter syntax (depth, attributes, dates, etc.)
- Real-time search as you type
- Quick navigation with keyboard shortcuts
- Optional node hoisting (Alt+Enter)
- Error messages for invalid queries
- Limits results to 10 matches for performance

## Activating the Widget

Press `Ctrl+K` to open the Node Search Widget. The widget appears as a centered modal box with:
- A search query input field
- A list of matching nodes (up to 10 results)
- Navigation and action keys

## Search Syntax

The Node Search Widget supports the complete advanced search syntax:

### Text Search

Plain text searches match node text content (case-insensitive substring):

```
task                # Find nodes containing "task"
important           # Find nodes containing "important"
```

### Filter Types

#### Depth Filter: `d:`
Match nodes at specific depth levels:
```
d:0                 # Root level only
d:>1                # Deeper than level 1
d:<=2               # Level 2 or shallower
```

#### Attribute Filter: `@`
Match nodes with attributes:
```
@status=done        # Nodes with status=done
@type=project       # Nodes with type=project
@url                # Nodes with any url attribute
@deadline>-7d       # Deadline in next 7 days
```

#### Date Filters
- `c:DATE` - Created date
- `m:DATE` - Modified date
- `@KEY>DATE` - Attribute date comparison

#### Children Filter: `children:`
```
children:0          # Leaf nodes (no children)
children:>0         # Parent nodes (has children)
children:>=3        # Nodes with 3+ children
```

#### Parent/Ancestor Filters
- `p:FILTER` - Match nodes whose parent matches
- `a:FILTER` - Match nodes under ancestor matching filter

### Boolean Operators

#### AND (implicit or `+`)
Filters separated by spaces are AND-ed:
```
task d:>0           # task AND depth > 0
```

#### OR: `|`
Match nodes satisfying ANY criteria:
```
project | urgent    # project OR urgent
```

#### NOT: `-`
Exclude matches:
```
-done               # Not done
-children:0         # Has children
```

## Keyboard Controls

### Navigation
| Key | Action |
|-----|--------|
| `Ctrl+N` | Move down in results (next match) |
| `Ctrl+P` | Move up in results (previous match) |
| `Home` | Move cursor to start of query |
| `End` | Move cursor to end of query |
| `Left Arrow` | Move cursor left |
| `Right Arrow` | Move cursor right |

### Editing
| Key | Action |
|-----|--------|
| `Backspace` | Delete character before cursor |
| `Delete` | Delete character at cursor |
| `Ctrl+W` | Delete word before cursor |

### Selection & Actions
| Key | Action |
|-----|--------|
| `Enter` | Select highlighted node and jump to it |
| `Alt+Enter` | Hoist the selected node (if it has children) |
| `Escape` | Close the widget without selecting |

## Examples

### Find by task status
```
@status=done        # Completed tasks
-@status=done       # Incomplete tasks
@priority=high      # High priority items
```

### Find by structure
```
children:0          # All leaf nodes
children:>0         # All parent nodes
d:2                 # Nodes at depth level 2
```

### Find by dates
```
c:>-7d              # Created in last 7 days
m:>-1d              # Modified today
@deadline>-3d       # Deadline in next 3 days
```

### Complex queries
```
task d:>0           # Tasks that aren't root level
project children:>0 # Projects with children
(@type=day | @type=note) m:>-7d # Recent daily notes or notes
```

## Display

The widget shows:
- **Query Line**: Current search query with cursor position
- **Results**: Up to 9 matching nodes (currently selected highlighted with `>`)
- **Footer**: Match count and available actions
- **Error Messages**: Parse errors displayed in red (e.g., invalid syntax)

### Error Handling

If the search query is invalid:
- A parse error message appears
- The widget falls back to simple text search
- Valid parts of the query are ignored

Example invalid query feedback:
```
Parse error: missing value
```

## Performance Considerations

- The widget limits results to 10 matches for performance
- Search applies filters instantly as you type
- Large outlines may have slight latency during complex searches
- Use specific filters (e.g., `d:>2` + `@type=task`) for faster results

## Tips

1. **Quick navigation**: Use `Ctrl+K` + type partial text + Enter to jump nodes
2. **Hoisting**: Alt+Enter to focus on a subtree
3. **Complex searches**: Combine multiple filters for precise results
4. **Attribute searches**: Use `@KEY=VALUE` to filter by custom attributes
5. **Date filtering**: Use relative dates like `-7d` for recent items

## Comparison with Main Search (`/`)

| Feature | Node Widget (Ctrl+K) | Main Search (/) |
|---------|---------------------|-----------------|
| Syntax | Advanced filters | Advanced filters |
| Modal/Overlay | Modal popup | Inline bar |
| Max results | 10 | All matches |
| Navigation | Ctrl+N/P | n/N keys |
| Hoisting | Alt+Enter | Not available |
| Real-time search | Yes | Yes |
| Error display | Red message | Inline in bar |

## Troubleshooting

### No results appear
- Check query syntax (use `/` for detailed help)
- Try simpler text search
- Verify items exist with that text/attributes

### Error "missing value"
- Filter requires a value, e.g. `d:` needs a number
- Use `@KEY` without value to check existence
- Check syntax documentation

### Slow search
- Simplify complex queries
- Use depth filters to narrow scope
- Avoid wildcard-like text searches on large outlines

For more detailed information on the search syntax, see `docs/search-syntax.md`.
