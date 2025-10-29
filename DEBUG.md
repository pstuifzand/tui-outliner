# Debugging Keybindings

If keybindings don't seem to work, follow these steps:

## 1. Verify the App Starts

```bash
./tui-outliner examples/sample.json
```

You should see:
- A title at the top showing "Project Planning"
- Several items displayed with tree characters (▼ for expanded, ▶ for collapsed)
- The first item should be highlighted (reverse video)
- Status line at the bottom showing "Ready"

## 2. Test Basic Navigation

Try these keys in order:
1. **`j` or `↓`** - Should move selection down to next item
2. **`k` or `↑`** - Should move selection up to previous item
3. Watch the highlighted line (reverse video) move up and down

## 3. Test Expand/Collapse

1. Navigate to an item with children (has a ▼ symbol)
2. Press **`l` or `→`** - The ▼ should change to ▶ and children should disappear
3. Press **`l` or `→`** again - Should show children again

## 4. Test Editing

1. Select an item with **`j`/`k`**
2. Press **`i`** to enter edit mode
   - The cursor should appear in the item text
   - Status line should update
3. Type some text
4. Press **`Escape`** or **`Enter`** to exit edit mode

## 5. Test Item Creation

1. Press **`o`** - A new item should appear below current item
   - Status should say "Created new item after"
   - The new item should be selected
2. Press **`a`** - A new item should appear as a child of selected item
   - Status should say "Created new child item"

## 6. Check Status Messages

The status line at the bottom should update when you:
- Create items (`o`, `a`)
- Delete items (`d`)
- Indent/outdent (`>`, `<`)
- Edit items (`i`)
- Save items (`Ctrl+S`)

## 7. Terminal Compatibility

Some terminal emulators may have issues with:
- Unicode characters (▼, ▶) - If these don't show, try a different terminal
- Special key codes (Ctrl+I, Ctrl+U) - These may be mapped differently in some terminals
- Color/style support - Make sure your terminal supports ANSI colors

## 8. Check for Errors

Run with error output visible:

```bash
./tui-outliner examples/sample.json 2>&1 | tee debug.log
```

## Common Issues

### Keybindings don't work at all
- Make sure the window has focus
- Try arrow keys instead of hjkl
- Try Ctrl+S to save - this usually works if any keybinding works

### Status messages not updating
- Check that the status line is visible at the bottom
- Status messages fade after 3 seconds

### Display corruption
- Resize the terminal window
- Try a different terminal emulator
- Check that you're using a monospace font

### Items not showing
- Make sure the file exists and is valid JSON
- Try with `examples/sample.json` first

## Development Build

For development with logging, edit `internal/app/app.go` and add debug output to `handleKeypress()`:

```go
func (a *App) handleKeypress(ev *tcell.EventKey) {
	// Add this line for debugging:
	// a.SetStatus(fmt.Sprintf("Key: %v Rune: %q", ev.Key(), ev.Rune()))

	selected := a.tree.GetSelected()
	// ... rest of function
}
```

This will show what key/rune was detected in the status line.
