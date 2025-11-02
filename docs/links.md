# Wiki-Style Internal Links

TUI Outliner supports wiki-style internal links that allow you to create references between items in your outline. Links are stored with unique item IDs for reliable navigation, but displayed as clean, readable text.

## Creating Links

### Inserting a Link While Editing

While editing an item's text, type `[[` to open the link autocomplete widget:

1. Press `i`, `A`, `c`, `o`, or `O` to enter edit mode on an item
2. Position your cursor where you want the link
3. Type `[[` - this triggers the link autocomplete widget
4. A modal window appears showing all items in your outline
5. Type to search for the item you want to link to
6. Use `Ctrl+N` / `Ctrl+P` (or arrow keys) to navigate the results
7. Press `Enter` to insert the link
8. Press `Escape` to cancel without inserting a link

**Example:**
```
Check [[item_20251103100530_abc123|Project Planning]] for details
```

In the tree view, this displays as:
```
Check Project Planning for details
```

The link text appears in cyan with underline to indicate it's a link.

## Updating Links

Links are just text in your outline, so you can edit them like any other text:

1. Enter edit mode on the item containing the link
2. Navigate to the link using arrow keys
3. Edit the text or use `Backspace`/`Delete` to remove characters
4. The link will be maintained as long as you keep the `[[id|text]]` format

**Changing Link Display Text:**
```
Before: [[item_123|old text]]
After:  [[item_123|new text]]
```

**Removing a Link:**
Just delete the `[[...]]` part entirely, leaving the text:
```
Before: Check [[item_123|this item]]
After:  Check this item
```

## Using Links

### Following Links

Press `gf` (go file) to navigate to the item referenced by a link:

1. Select an item that contains a link
2. Press `g` then `f`
3. The outline automatically:
   - Expands parent items to make the target visible
   - Navigates to the linked item
   - Shows a status message indicating success

**Example:**
```
Selected item: "Check Project Planning for details"
Press gf → Navigates to the Project Planning item
Status shows: "Followed link to: Project Planning"
```

If the link target doesn't exist (broken link):
```
Status shows: "Broken link: target item not found (item_123)"
```

### Listing Links in an Item

Use the `:links` command to see all links in the currently selected item:

```
:links
```

The status bar displays:
- **Single link:** Full link details → `"Found 1 link: 1. Project Planning -> item_20251103100530_abc123"`
- **Multiple links (2-3):** All links → `"Found 2 links: 1. Project | 2. Notes"`
- **Many links (4+):** Just count → `"Found 5 links (use gf to follow first one)"`

## File Format

Links are stored in the JSON file as `[[id|text]]` within item text:

```json
{
  "text": "Check [[item_20251103100530_abc123|Project Planning]] for details",
  "children": [...],
  "metadata": {...}
}
```

### Link Format Details

- **ID:** Unique item identifier (format: `item_YYYYMMDDHHMMSS_randomtext`)
- **Display text:** Human-readable link text
- **Syntax:** `[[id|display_text]]` or `[[id]]` (if display text not provided, ID is used)

When you insert a link via autocomplete, it automatically uses the target item's text as the display text.

## Tips and Examples

### Create Reference Networks

Link related items together to build a knowledge network:
```
Daily Note (2025-11-03)
  - Worked on [[item_proj_id|Project X]]
  - Discussed with [[item_person_id|Alice]]
  - See [[item_notes_id|General Notes]] for context
```

### Navigate Hierarchies

Link to parent items or important ancestors:
```
Task Details
  - Part of [[item_epic_id|Epic: Q4 Planning]]
  - Related to [[item_goal_id|Company Goal]]
```

### Create Bidirectional References

Link items in both directions for easy exploration:
```
Item A: "See also [[item_b_id|Item B]]"
Item B: "Related to [[item_a_id|Item A]]"
```

### Update References When Restructuring

If you move or rename items, the links still work (they use IDs, not paths). Just update the display text if needed:

```
Before move: [[item_123|old location/name]]
After move:  [[item_123|new location/name]]
```

## Keyboard Shortcuts Summary

| Action | Shortcut |
|--------|----------|
| Insert link while editing | Type `[[` then search and press `Enter` |
| Navigate search results | `Ctrl+N` / `Ctrl+P` or arrow keys |
| Confirm link selection | `Enter` |
| Cancel link insertion | `Escape` |
| Follow first link | `gf` (in normal mode) |
| List all links in item | `:links` (in command mode) |

## Visual Indicators

- **Link text color:** Cyan
- **Link style:** Underlined
- **Background:** Matches the item's state (normal or selected)

Links are visually distinct in the tree view, making them easy to spot while reading your outline.
