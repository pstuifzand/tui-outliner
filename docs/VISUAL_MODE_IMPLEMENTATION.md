# Visual Mode Implementation Details

## Overview

Visual mode is a Vim-inspired feature that allows selecting and performing operations on multiple items at once. This document describes the technical implementation.

## Architecture

### State Management

Visual mode state is tracked in two places:

1. **Mode Enum** (`internal/app/app.go`):
   ```go
   type Mode int
   const (
       NormalMode Mode = iota
       InsertMode
       VisualMode
   )
   ```

2. **Visual Anchor** (`internal/app/app.go`):
   ```go
   type App struct {
       mode         Mode
       visualAnchor int  // Index in filteredView where selection started (-1 = not active)
   }
   ```

### Event Routing

Visual mode input is routed through `handleRawEvent()`:

```
User Input
    ↓
handleRawEvent()
    ↓
    ├─ Command mode active? → handleCommand()
    ├─ Search mode active? → search handler
    ├─ Editor active? → editor.HandleKey()
    ├─ Visual mode? → handleVisualMode()
    └─ Normal mode → handleKeypress()
```

### Keybindings

Visual mode has dedicated keybindings defined in `InitializeVisualKeybindings()`:

```go
type KeyBinding struct {
    Key         rune
    Description string
    Handler     func(*App)
}
```

Lookup is performed via:
```go
GetVisualKeybindingByKey(key rune) *KeyBinding
```

## Selection Calculation

### Selection Range

The visual selection spans from `visualAnchor` to the current cursor position (`tree.GetSelectedIndex()`):

```go
getVisualSelectionRange() (start, end int)
// Returns indices in filteredView
// Guarantees start <= end
```

### Rendering

The `Render()` method in `tree.go` receives the visual anchor:

```go
func (tv *TreeView) Render(screen *Screen, startY int, visualAnchor int)
```

It calculates the range and applies styles:
- **In range, at cursor**: `visualCursorStyle` (bold, blue background)
- **In range, not cursor**: `visualStyle` (white text, blue background)
- **Not in range, at cursor**: `selectedStyle` (normal selected style)
- **Not in range**: `defaultStyle` (normal style)

## Operations

### DeleteVisualSelection

```go
func (a *App) deleteVisualSelection()
    1. Get selection range
    2. Get all items in range via tree.GetItemsInRange()
    3. Delete each item via tree.DeleteItem()
    4. Exit visual mode (set visualAnchor = -1)
    5. Set dirty flag
```

**Status**: "Deleted N items"

### YankVisualSelection

```go
func (a *App) yankVisualSelection()
    1. Get selection range
    2. Get all items in range
    3. Copy to clipboard (currently first item only)
    4. Exit visual mode
```

**Status**: "Yanked N items"

**Note**: Full multi-item clipboard support is a TODO

### IndentVisualSelection

```go
func (a *App) indentVisualSelection()
    1. Get selection range
    2. For each item in range:
       a. Call tree.IndentItem(item)
       b. Count successful operations
    3. Exit visual mode
    4. Set dirty flag
```

**Status**: "Indented N items"

### OutdentVisualSelection

```go
func (a *App) outdentVisualSelection()
    1. Get selection range
    2. For each item in range:
       a. Call tree.OutdentItem(item)
       b. Count successful operations
    3. Exit visual mode
    4. Set dirty flag
```

**Status**: "Outdented N items"

## Tree Methods

New methods added to support visual operations:

### GetItemsInRange

```go
func (tv *TreeView) GetItemsInRange(start, end int) []*model.Item
```

Returns all items in the visible range (indices in filteredView). This respects the flattened tree structure.

### IndentItem

```go
func (tv *TreeView) IndentItem(item *model.Item) bool
```

Makes a specific item a child of the previous item. Mirrors the existing `Indent()` method but works on arbitrary items.

### OutdentItem

```go
func (tv *TreeView) OutdentItem(item *model.Item) bool
```

Moves a specific item up one nesting level. Mirrors the existing `Outdent()` method.

## Theme Integration

Visual mode colors are fully integrated into the theme system:

### Colors Struct

```go
type Colors struct {
    TreeVisualSelection  tcell.Color
    TreeVisualSelectionBg tcell.Color
    TreeVisualCursor     tcell.Color
    TreeVisualCursorBg   tcell.Color
}
```

### Default Theme

```go
TreeVisualSelection:   tcell.ColorWhite
TreeVisualSelectionBg: tcell.ColorBlue
TreeVisualCursor:      tcell.ColorBlack
TreeVisualCursorBg:    tcell.ColorBlue
```

### TOML Configuration

Visual colors can be customized in theme TOML files:

```toml
[colors]
tree_visual_selection = "#ffffff"
tree_visual_selection_bg = "#0000ff"
tree_visual_cursor = "#000000"
tree_visual_cursor_bg = "#0000ff"
```

## Status Line Display

The status line updates based on current mode:

```go
if a.mode == InsertMode {
    statusLine = "-- INSERT --"
} else if a.mode == VisualMode {
    statusLine = "-- VISUAL --"
} else {
    statusLine = "-- NORMAL --"
}
```

## Initialization

Visual anchor is initialized to -1 in `NewApp()`:

```go
app := &App{
    // ... other fields ...
    visualAnchor: -1,
}
```

This ensures visual selection is disabled on startup.

## Flow Example: Delete Selection

```
User presses 'V'
    ↓
InitializeKeybindings() adds V keybinding
    ↓
Handler sets mode = VisualMode, visualAnchor = selectedIdx
    ↓
Status line now shows "-- VISUAL --"
    ↓
User presses 'j' 3 times
    ↓
handleVisualMode() calls GetVisualKeybindingByKey('j')
    ↓
Handler calls tree.SelectNext() (repeats 3 times)
    ↓
Render() shows selection from visualAnchor to currentIdx
    ↓
User presses 'd'
    ↓
handleVisualMode() calls deleteVisualSelection()
    ↓
deleteVisualSelection():
    - getVisualSelectionRange() → (anchor, cursor)
    - tree.GetItemsInRange(anchor, cursor)
    - For each item, tree.DeleteItem()
    - Set mode = NormalMode, visualAnchor = -1
    ↓
Next render shows normal mode again
```

## Known Limitations

1. **Clipboard**: Only stores first selected item (full array needed)
2. **Search integration**: Visual mode doesn't work in search results view
3. **Undo/Redo**: No undo support for visual operations
4. **Boundary validation**: Simple range-based selection (could be enhanced)
5. **Collapsed children**: Operations affect hidden children (by design)

## Future Enhancements

- [ ] Multi-item clipboard storage
- [ ] Undo/redo support for visual operations
- [ ] Visual mode in search results
- [ ] Character-wise vs line-wise visual modes
- [ ] Visual block selection (rectangular)
- [ ] Macro recording with visual operations
