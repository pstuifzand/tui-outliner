# Multi-line Text in Tree Nodes

The TUI Outliner now supports multi-line text in tree nodes. This allows you to create items with multiple lines of content, displayed naturally in the tree view.

## Features

### Multi-line Display Rendering
- Items with newline characters (`\n`) are displayed across multiple visual lines
- Each text line maintains the item's structure and styling
- Only the first line shows the item's metadata (indent, arrow, attributes)
- Subsequent lines are properly indented to match their parent's nesting level
- All lines of a multi-line item are highlighted together when selected

### Width Handling
- Long lines that exceed terminal width are truncated with an ellipsis (`…`)
- Maximum width is calculated as: `terminal_width - indentation - prefix`
- Truncation is applied intelligently on both first and continuation lines
- Maintains visual alignment and hierarchy

### Editing Multi-line Text
- Press **Shift+Enter** to insert a newline while editing
- Plain **Enter** still creates a new item (preserves existing behavior)
- Newlines are preserved in the item's text field as `\n` characters

### Mouse Support
- Clicking any line of a multi-line item selects that item
- Expand/collapse arrows work only on the first line
- Multi-line items respond to expand/collapse operations as a single unit

### Keyboard Navigation
- **Up/Down arrows**: Move between items (not individual lines)
- **Ctrl+U/Ctrl+D**: Page up/down operations work with multi-line aware viewport
- All navigation commands treat multi-line items as single entities
- Selection always applies to the entire item, highlighting all its lines

## JSON Format

Multi-line text is stored as a single string with newline characters:

```json
{
  "text": "First line\nSecond line\nThird line"
}
```

Example with nested items:

```json
{
  "id": "1",
  "text": "Project title",
  "expanded": true,
  "children": [
    {
      "id": "2",
      "text": "Task with details\nLine 2 of the task\nLine 3 of the task",
      "children": []
    }
  ]
}
```

## Usage Example

### Creating Multi-line Items

1. Start editing an item (press `i`, `a`, `c`, or `o`)
2. Type the first line of text
3. Press **Shift+Enter** to add a newline
4. Type the next line
5. Repeat as needed
6. Press **Escape** to finish editing (discards empty items)
7. Or press **Enter** to save and create a new item

### Viewing Multi-line Items

In the tree view:
```
▼ Project title
  ▶ Task with details
    Second line of the task
    Third line of the task
  ▶ Another task
    With two lines
```

The first line shows all metadata (arrow, indicator, attributes, progress bar).
Continuation lines are aligned with the text start position.

## Display Behavior

### Multi-line Item Structure
- **First Line (ItemStartLine=true)**:
  - Shows indentation and arrow for expand/collapse
  - Shows attribute indicator (●) if item has attributes
  - Displays visible attributes (if configured with `:set visattr`)
  - Shows progress bar (if item has todo children)
  - Text is truncated if it exceeds screen width

- **Continuation Lines (ItemStartLine=false)**:
  - Only show the text content
  - Indented to match the first line's text start position
  - No metadata, arrows, or indicators
  - Selected items highlight all continuation lines
  - Text is truncated independently per line

### Visual Selection
- When an item is selected, ALL lines of that item are highlighted
- Visual range selection (V mode) highlights complete items
- Status bar and search highlighting work across all lines

## Examples

See `examples/multiline_demo.json` for comprehensive examples of multi-line text usage.

## Implementation Details

### Data Structures
- **DisplayLine struct**: Represents a single visual line in the tree view
  - `Item`: Reference to the underlying Item
  - `TextLine`: The actual text to display for this line
  - `TextLineIndex`: Which line within the item's text (0 = first line)
  - `ItemStartLine`: True if this is the first line of the item
  - `Depth`: Nesting level

### Key Functions
- `buildDisplayLines()`: Converts display items into multi-line aware display lines
- `getFirstDisplayLineForItem()`: Gets the first display line of an item
- `getLastDisplayLineForItem()`: Gets the last display line of an item
- `GetItemFromDisplayLine()`: Maps display line to item index for mouse clicks

### Rendering
- Display lines are calculated fresh on each `RebuildView()`
- Each item's text is split by `\n` to create display lines
- Rendering handles both first lines (with metadata) and continuation lines
- Width truncation with ellipsis is applied to both types of lines
- Viewport management uses display line indices while selection uses item indices

## Text Wrapping

### Overview
Long lines that exceed the terminal width are automatically wrapped at word boundaries. This preserves the natural flow of text while keeping items readable.

### Wrapping Behavior
- **Automatic**: Text wrapping is enabled by default based on terminal width
- **Word-boundary aware**: Attempts to wrap at spaces and word boundaries
- **Character-boundary fallback**: If no suitable word boundary exists, wraps at character boundary
- **Dynamic**: Recalculates on terminal resize (when RebuildView is called)

### Wrapped vs Hard-break Continuations
- **Hard-break continuation** (after explicit `\n`):
  - Indented same as first line text position
  - Represents a new logical section/paragraph
  - Example: item with `Line 1\nLine 2`

- **Wrapped continuation** (from word wrapping):
  - Indented 2 extra spaces beyond hard-break indentation
  - Visual distinction shows it's part of the same logical line
  - Example: a very long single-line item split across multiple visual lines

### Width Calculation
- Maximum wrap width = `terminal_width - 21` characters
  - 18 characters reserved: 6 nesting levels × 3 characters each
  - 3 characters for arrow/indicator area
- Minimum wrap width: 20 characters (for very narrow terminals)

### Visual Example
```
▼ Project with long description line that continues and wraps to the next
line here with no indentation
  ▶ Item with explicit line breaks
    First paragraph
    Second paragraph (hard break after \n)
  ▶ Another item
```

Notice:
- First `▼` shows arrow, indent, and start of text
- Continuation lines ("line here...") have no indentation, extending full width
- "Second paragraph" is indented (hard break after explicit `\n`)
- Multi-line selection highlights all lines of the item

## Limitations and Future Enhancements

### Current Limitations
- Editor only shows the first line during editing (full text is preserved)
- Wrapping width is fixed based on screen dimensions (no user configuration option yet)
- Continuation lines don't show in search highlights (only text match)

### Possible Enhancements
1. **Multi-line editor modal**: For editing long multi-line content
2. **Configurable wrapping**: `:set wrapping on/off` or `:set wrapwidth <N>`
3. **Alternative display modes**: Option to truncate instead of wrap
4. **Line-specific operations**: Delete/edit individual lines within an item
5. **Export formats**: Better preservation of line structure in markdown/text export
6. **Search result display**: Show context around matches in multi-line items
