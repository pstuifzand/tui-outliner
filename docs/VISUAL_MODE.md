# Visual Mode Guide

Visual mode allows you to select and manipulate multiple items at once, similar to Vim's visual mode.

## Entering Visual Mode

Press `V` in normal mode to enter visual mode. The status line will display `-- VISUAL --`.

## Navigation

Once in visual mode, use the following keys to move the selection cursor:

- `j` or `↓` - Extend selection down (move cursor down)
- `k` or `↑` - Extend selection up (move cursor up)
- `h` or `←` - Collapse item (if hovering over a parent)
- `l` or `→` - Expand item (if hovering over a parent)

The selection spans from the anchor point (where you entered visual mode) to the current cursor position.

## Performing Operations

With items selected, you can perform these operations:

### Delete
- `d` or `x` - Delete all selected items
  - Status: "Deleted N items"
  - Exits visual mode automatically

### Copy (Yank)
- `y` - Copy selected items to clipboard
  - Status: "Yanked N items"
  - Exits visual mode automatically
  - Note: Currently stores the first item in clipboard

### Indent
- `>` - Indent all selected items
  - Makes them children of the previous item
  - Status: "Indented N items"
  - Exits visual mode automatically

### Outdent
- `<` - Outdent all selected items
  - Moves them up one nesting level
  - Status: "Outdented N items"
  - Exits visual mode automatically

## Exiting Visual Mode

- `V` - Exit visual mode (toggle off)
- `Escape` - Exit visual mode without performing any action

After performing an operation (delete, yank, indent, outdent), visual mode automatically exits.

## Visual Feedback

- **Selected items**: Displayed with white text on blue background
- **Cursor position**: Displayed with bold black text on blue background
- The selection always includes complete nodes with all their children

## Example Workflow

1. Press `V` to enter visual mode on the first item you want to select
2. Press `j` three times to extend selection to include 4 items total
3. Press `>` to indent all 4 items at once
4. Visual mode exits automatically, and you're back in normal mode

## Limitations

- Visual mode only works in the main outline view (not in search results)
- Multi-item clipboard is not yet fully implemented (only first item stored)
- Operations don't have undo/redo support yet
