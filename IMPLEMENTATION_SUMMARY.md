# Node Search Widget - Advanced Search Integration

## Summary

Successfully integrated the advanced search parser into the Node Search Widget (Ctrl+K), replacing the previous fuzzy-search-only implementation with full support for advanced filter expressions.

## Changes Made

### 1. Node Search Widget Enhancement (`internal/ui/node_search_widget.go`)

**New Fields:**
- `parseError string` - Stores parse errors from query parsing
- `filterExpr search.FilterExpr` - Stores parsed filter expression

**Modified Functions:**
- `updateMatches()` - Replaced fuzzy search with advanced filter expression parsing
  - Now uses `search.ParseQuery()` to parse advanced syntax
  - Falls back to text-only matching on parse errors
  - Maintains 10-result limit for performance

- `Render()` - Enhanced rendering
  - Added error message display in red
  - Shows parse errors inline in the widget

**Search Features:**
- Text search: `task`, `project`
- Depth filters: `d:>1`, `d:<=2`
- Attribute filters: `@status=done`, `@type=project`, `@deadline>-7d`
- Date filters: `c:>-7d`, `m:<-30d`
- Children filters: `children:0`, `children:>0`
- Parent/Ancestor filters: `p:d:0`, `a:@type=project`
- Boolean operators: Implicit AND, explicit `+`, OR `|`, NOT `-`
- Grouping: `(filter1 | filter2) filter3`

### 2. Documentation

**New Files:**
- `docs/node-search-widget.md` - Comprehensive guide with examples

**Updated Files:**
- `README.md`
  - Added Node Search Widget to features section
  - Added `Ctrl+K` to quick start keybindings
  - Added Node Search Widget to Other keybindings table
  - Added dedicated "Node Search Widget" section with features, keyboard controls, and examples
  - Reference to detailed documentation

- `CLAUDE.md`
  - Documented Node Search Widget enhancement as item #17
  - Listed all features and implementation details
  - Noted documentation location

### 3. Testing & Verification

**Tests Verified:**
- All existing search tests pass (100+ tests)
- UI tests pass
- Build succeeds without warnings or errors

**Example Test Queries:**
- `design` - Text search
- `d:>1` - Depth filter
- `@status=done` - Attribute filter
- `d:>1 @type=task` - Combined filters
- `task | project` - OR operator
- `-project` - NOT operator
- `children:0` - Leaf nodes
- `(task | subtask) @status=done` - Complex grouping

## Features

### Supported Filter Expressions

| Type | Syntax | Examples |
|------|--------|----------|
| Text | Plain text | `task`, `important` |
| Depth | `d:COMP` | `d:>1`, `d:<=2`, `d:0` |
| Attribute | `@KEY[=VALUE]` | `@status=done`, `@url`, `@deadline>-7d` |
| Date (Created) | `c:COMP DATE` | `c:>-7d`, `c:<-30d` |
| Date (Modified) | `m:COMP DATE` | `m:>-1d`, `m:2025-11-01` |
| Children | `children:COMP N` | `children:0`, `children:>5` |
| Parent | `p:FILTER` | `p:d:0`, `p:@status=active` |
| Ancestor | `a:FILTER` | `a:@type=project`, `a:d:0` |

### Operators

| Operator | Syntax | Example |
|----------|--------|---------|
| AND | Space or `+` | `task d:>0` or `task + d:>0` |
| OR | `\|` | `task \| project` |
| NOT | `-` | `-project`, `task -done` |
| Grouping | `()` | `(task \| project) d:>0` |

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl+K` | Open Node Search Widget |
| `Ctrl+N` / `Ctrl+P` | Next/Previous result |
| `Enter` | Select and jump to node |
| `Alt+Enter` | Hoist selected node |
| `Escape` | Close widget |
| `Ctrl+W` | Delete word before cursor |

## Implementation Details

### Parser Integration
The widget uses the existing `search.ParseQuery()` function from `internal/search/parser.go` which:
- Tokenizes the query string
- Parses using recursive descent parser
- Builds s-expression tree with proper operator precedence
- Returns `FilterExpr` interface for matching

### Fallback Behavior
If parsing fails:
- Error message displayed in red in widget
- Falls back to simple text search
- User can still navigate results or close the widget

### Performance Optimization
- Results limited to 10 matches (configurable via `maxResults`)
- Stops adding matches once limit reached
- Suitable for large outlines

## Files Changed

1. **Core Implementation:**
   - `internal/ui/node_search_widget.go` - Updated search logic and rendering

2. **Documentation:**
   - `docs/node-search-widget.md` - New comprehensive guide
   - `README.md` - Updated with Node Search Widget info
   - `CLAUDE.md` - Added implementation notes

3. **No changes to:**
   - Search parser (`internal/search/`)
   - Main search UI (`internal/ui/search.go`)
   - Application core (`internal/app/`)

## Testing

All tests pass:
```bash
$ go test ./...
```

**Test Coverage:**
- Search parser tests: 100+ test cases
- UI tests: Existing tests still pass
- Integration: Widget works with main application

## Backward Compatibility

✓ No breaking changes
✓ Existing search functionality (`/`) unchanged
✓ All previous keybindings work
✓ All existing tests pass

## Future Enhancements

Potential improvements for future versions:
1. Increase result limit dynamically based on terminal height
2. Add live highlight of matching terms in results
3. Support regex patterns for text search
4. Add search history to Node Search Widget
5. Remember last search query between sessions

## Usage Examples

### Basic Usage
```
Ctrl+K              # Open Node Search Widget
task                # Find items with "task"
@status=done        # Completed items
d:>1                # Items deeper than level 1
```

### Find tasks by status
```
@status=done        # Completed items
@status=in-progress # Currently working on
-@status=done       # Incomplete items
```

### Find by hierarchy
```
children:0          # Leaf nodes only
children:>0         # Parent nodes only
d:2                 # Items at depth level 2
p:d:0               # Direct children of root
```

### Find by dates
```
c:>-7d              # Created this week
m:>-1d              # Modified today
@deadline>-3d       # Deadline within 3 days
```

### Complex queries
```
task @status=done d:>1           # Done tasks not at root
(project | task) children:>0     # Parent projects or tasks
@type=day @date>-7d              # Daily notes from past week
```

## Conclusion

The advanced search parser has been successfully integrated into the Node Search Widget, providing users with powerful filtering capabilities directly accessible via `Ctrl+K`. The implementation maintains backward compatibility, passes all tests, and includes comprehensive documentation.
