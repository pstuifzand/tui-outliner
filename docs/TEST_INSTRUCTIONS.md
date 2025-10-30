# Testing Instructions

## Quick Test

```bash
./tui-outliner examples/sample.json
```

Expected behavior:
1. Window shows "Project Planning" as title
2. Multiple items visible with tree expand/collapse symbols (▼ / ▶)
3. One item is highlighted (reverse video/inverted colors)
4. Status line at bottom says "Ready"

## Test Keybindings

With the window open, try these in order:

### 1. Navigation Test
Press `j` or down arrow multiple times
- Selection should move down through items
- Watch the reverse video highlight move

Press `k` or up arrow multiple times
- Selection should move back up

### 2. Expand/Collapse Test
Find an item with ▼ (expanded, has children)
Press `h` or left arrow
- The ▼ should change to ▶
- Children should disappear from view

Press `l` or right arrow
- Should show children again (▼ returns)

### 3. Debug Mode Test
```bash
./tui-outliner -debug examples/sample.json
```

Press any key - status line should show:
```
Key: <key-code> | Rune: '<char>' | Modifiers: <mods>
```

This confirms the app is receiving your keypresses.

### 4. Edit Mode Test
Press `i` to edit current item
- Cursor should appear in the item text
- You should be able to type
- Press Enter or Escape to exit

### 5. Create Item Test
Press `o` to create item after
- Status line should say "Created new item after"
- A new item should appear below current
- It should be selected

Press `a` to create child
- Status line should say "Created new child item"
- A new item appears indented below current

### 6. Indent/Outdent Test
Create two items at the same level with `o`
Select the second item
Press `>` or `.` to indent
- The item should move to become a child of the previous item
- Status should say "Indented"
- The parent should show ▼ (expanded) with the child visible

Press `<` or `,` to outdent
- The item should move back to root level
- Status should say "Outdented"
- It should no longer be indented

## If Nothing Works

1. Check that you have a 24-line terminal (height must be at least 3)
2. Try using arrow keys instead of hjkl
3. Run with debug mode: `./tui-outliner -debug examples/sample.json`
4. Try a different terminal emulator
5. Check for errors: `./tui-outliner examples/sample.json 2>&1`

## Information to Report

If keybindings still don't work, please provide:

1. Output from debug mode:
   ```bash
   ./tui-outliner -debug examples/sample.json 2>&1 | head -20
   ```

2. Terminal name:
   ```bash
   echo $TERM
   ```

3. Terminal emulator:
   ```
   (uname -s, name of your terminal app)
   ```

4. Terminal size:
   ```bash
   stty size
   ```

5. What you see when you run the app (screenshot or description)

6. Any error messages that appear
