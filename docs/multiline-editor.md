# MultiLineEditor

The `MultiLineEditor` is a specialized text editor component that provides multi-line editing with automatic word wrapping. It has **replaced the single-line `Editor` throughout the application** for editing tree nodes.

Unlike the original single-line `Editor`, it's designed to handle long text with proper wrapping, cursor navigation across wrapped lines, and visual line breaks. Now when you edit any item in the tree, you get full multi-line support automatically.

## Integration with Tree Nodes

The MultiLineEditor is now **automatically used for all tree node editing**. When you press any edit key (`i`, `a`, `c`, `o`, `O`), the MultiLineEditor is created for the selected item.

### How It Works
1. **Edit Key Pressed**: User presses `i` (insert), `a` (append), `c` (change), `o` (new after), or `O` (new before)
2. **MultiLineEditor Created**: `ui.NewMultiLineEditor(item)` is instantiated
3. **Editor Initialized**: Width is calculated based on terminal width and item indentation
4. **Text Displayed**: Text renders with word wrapping across multiple visual lines
5. **Editing**: User types, uses Shift+Enter for newlines, Up/Down for wrapped lines
6. **Save/Cancel**: Press Enter to save or Escape to cancel

### Display During Editing
- Editor text wraps to fit the available width
- Indentation and tree structure are preserved
- Wrapped lines are properly aligned
- Multi-line items display across multiple screen rows
- Selection and visual styling work across all lines

## Overview

### Purpose
- Edit multi-line item text with word wrapping (now integrated with tree)
- Navigate across wrapped lines naturally
- Support all standard editing operations
- Seamless experience for long item descriptions

### Key Features
- **Automatic Word Wrapping**: Text wraps at word boundaries
- **Multi-line Rendering**: Displays multiple visual lines for single items
- **Smart Cursor Navigation**: Up/Down move between wrapped lines, Home/End work per-line
- **Preserved Newlines**: Explicit newlines (`\n`) are preserved from text
- **Cursor Position Mapping**: Converts between text offsets and visual (row, col) positions

## Architecture

### Text Representation
Text is stored as a single string with embedded newlines (`\n`). This matches the tree item's text representation.

### Wrapped Lines
- `wrappedLines []string`: Cached split text into wrapped portions
- `lineStartOffsets []int`: Starting text offset for each wrapped line
- Recalculated whenever text changes
- Based on `maxWidth` parameter

### Cursor Position
- `cursorPos`: Absolute position in the text string (0-based)
- Visual position calculated on-demand via `getCursorVisualPosition()`
- Can be converted back via `getCursorTextOffset(row, col)`

## API Reference

### Initialization
```go
editor := NewMultiLineEditor(item)
editor.SetMaxWidth(80)  // Optional: set wrapping width
editor.Start()          // Begin editing
```

### Text Management
```go
editor.SetText(text)     // Set text and recalculate wrapping
text := editor.GetText() // Get current text
```

### Cursor Control
```go
editor.GetCursorPos()          // Get text offset of cursor
editor.GetCursorVisualRow()    // Get visual row of cursor
editor.GetWrappedLineCount()   // Get number of visual lines
```

### Key Handling
```go
handled := editor.HandleKey(event)  // Process keyboard event
```

### State Management
```go
editor.Start()                      // Begin editing
editor.Stop()                       // Finish editing (saves text)
editor.Cancel()                     // Abandon edits
```

### Flags
```go
editor.WasEnterPressed()       // Plain Enter (create new item)
editor.WasEscapePressed()      // Escape (cancel)
editor.WasBackspaceOnEmpty()   // Backspace on empty item
editor.WasIndentPressed()      // Tab (indent)
editor.WasOutdentPressed()     // Shift+Tab (outdent)
```

## Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| Left/Right | Move cursor within current line |
| Ctrl+Left/Right | Jump to previous/next word |
| Up/Down | Move to previous/next wrapped line (same column) |
| Home | Go to start of current wrapped line |
| End | Go to end of current wrapped line |
| Ctrl+A | Go to start of all text |
| Ctrl+E | Go to end of all text |

### Editing
| Key | Action |
|-----|--------|
| Any Character | Insert character at cursor |
| Shift+Enter | Insert newline (multi-line text) |
| Enter | Finish editing, create new item |
| Backspace | Delete character before cursor |
| Delete | Delete character at cursor |
| Ctrl+Delete | Delete word forward |
| Ctrl+W | Delete word backwards |
| Ctrl+U | Delete from start to cursor |
| Ctrl+K | Delete from cursor to end |

### History
| Key | Action |
|-----|--------|
| Ctrl+Z | Undo last edit |
| Ctrl+Y | Redo last undo |

### Exit
| Key | Action |
|-----|--------|
| Escape | Cancel editing, discard changes |
| Enter | Save and create new item |

## Undo/Redo

The editor supports up to 50 levels of undo history:

- **Ctrl+Z** - Undo the last edit operation
- **Ctrl+Y** - Redo the last undone edit

### Undo Behavior

- Each user action (character insertion, deletion, word operations) creates an undo checkpoint
- Undo stack is limited to 50 entries to prevent excessive memory usage
- When you perform a new edit after undoing, the redo history is cleared
- Undo history is cleared when you exit the editor (either by saving or canceling)

### Word Navigation and Deletion

The editor supports efficient word-based navigation and editing:

- **Ctrl+Left** - Jump to the start of the previous word
- **Ctrl+Right** - Jump to the start of the next word
- **Ctrl+Delete** - Delete the next word
- **Ctrl+W** - Delete the previous word (existing)

Words are defined as sequences of non-space, non-newline characters. Navigation skips over spaces intelligently.

## Integration in Tree Editing

When you edit a tree item in the application:

```
Normal view:
▼ Project with long description text that gets wrapped
Task details continue on next line with no indentation
  ▶ Subtask with details

Editing mode (press 'i' on first item):
▼ Project with long description text that gets
wrapped at word boundaries during editing
  ▶ Subtask with details
```

Note: Continuation lines (from word wrapping) have no indentation and extend full width. This is different from items at the next nesting level, which are indented by 3 spaces.

## Usage Example (Advanced)

For advanced use cases where you create a MultiLineEditor directly:

```go
// Create editor for item
item := model.NewItem("Initial text")
editor := NewMultiLineEditor(item)

// Set maximum wrap width (calculate from screen)
screenWidth := 100
maxTextWidth := screenWidth - 21  // Account for indentation
editor.SetMaxWidth(maxTextWidth)

// Start editing
editor.Start()

// Handle key events
for event := range eventChan {
    if ev, ok := event.(*tcell.EventKey); ok {
        if !editor.HandleKey(ev) {
            // Key wasn't handled by editor (e.g., Escape, Enter)
            if editor.WasEscapePressed() {
                // Discard changes
                text := editor.Cancel()
            } else if editor.WasEnterPressed() {
                // Save and create new
                text := editor.Stop()
                // Create new item...
            }
            break
        }
    }
}

// Render editor
editor.Render(screen, xPos, yPos, maxWidth)
```

## Implementation Details

### Word Wrapping Algorithm
Uses the same `wrapTextAtWidth()` function as the TreeView:
1. Attempts to wrap at word boundaries (spaces)
2. Falls back to character-boundary wrapping for unbreakable words
3. Preserves explicit newlines as hard breaks

### Cursor Position Mapping
**Text Offset → Visual Position:**
- Iterate through `wrappedLines` starting offsets
- Find which line contains the cursor position
- Calculate column within that line

**Visual Position → Text Offset:**
- Validate row and column against wrapped lines
- Look up starting offset of the line
- Add column to get text offset

### Wrapped Line Calculation
When text changes:
1. Split by hard newlines (`\n`)
2. For each hard line, apply word wrapping
3. Track starting offset of each wrapped portion
4. Store in `wrappedLines` and `lineStartOffsets`

### Multi-line Rendering
```
Line 0: ▼ First line of item with metadata
Line 1:   Wrapped continuation at same indent
Line 2:   Another wrapped continuation
Line 3:   Second paragraph after \n
```

## Differences from Editor

| Feature | Editor | MultiLineEditor |
|---------|--------|-----------------|
| Max Lines | 1 (single-line) | Multiple (wrapped) |
| Text Height | 1 screen line | Multiple screen lines |
| Wrapping | No (scrolls off-screen) | Yes (word-wrap) |
| Up/Down Keys | N/A | Navigate wrapped lines |
| Home/End | Not needed | Work per wrapped-line |
| Newlines | Not supported | Shift+Enter support |
| Rendering | Simple horizontal | Multi-line vertical |

## Performance Considerations

- **Wrapped lines cached**: Recalculated only on text changes, not on every render
- **Cursor position**: Calculated on-demand (O(n) where n = number of wrapped lines)
- **Memory**: Stores original text + array of wrapped portions (minimal overhead)
- **Suitable for**: Items with up to several thousand characters

## Limitations

- Cursor shown as visual block only (no blinking animation)
- No support for multi-column wide characters yet
- Rendering limited to contiguous screen area
- No undo/redo functionality (delegated to caller)

## Future Enhancements

1. **Cursor animation**: Blinking cursor indicator
2. **Syntax highlighting**: Color different parts of text
3. **Search/highlight**: Find and highlight text within editor
4. **Undo/redo**: History of edits
5. **Rich text**: Support for formatting markers
6. **Selection**: Highlight text ranges
7. **Copy/paste**: Clipboard integration
