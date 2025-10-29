# Keybindings Architecture

## Overview

The keybindings system in tui-outliner is organized by mode, with each mode having its own keybindings map. This allows different key behaviors depending on the current editor mode.

## Mode-Based Keybindings

### Normal Mode Keybindings
Location: `internal/app/keybindings.go` - `InitializeKeybindings()`

These are the default keybindings active when not in Insert or Visual mode.

**Key bindings include:**
- Navigation: `j`, `k`, `h`, `l` (and arrow keys), `G` (go to last node)
- Pending keys: `g...` (go to commands), `z...` (fold/zoom commands - reserved)
  - `gg` (go to first node)
- Node operations: `J`, `K` (move up/down), `i`, `c`, `A` (edit), `o`, `O` (create)
- Deletion: `d` (delete current)
- Paste: `p`, `P` (paste below/above)
- Indentation: `>`, `<` (indent/outdent)
- Search: `/` (search), `?` (help)
- Command: `:` (command mode)
- Visual: `V` (enter visual mode)

### Visual Mode Keybindings
Location: `internal/app/keybindings.go` - `InitializeVisualKeybindings()`

These keybindings are active when in visual mode (multi-item selection).

**Key bindings include:**
- Navigation: `j`, `k`, `h`, `l` (extend selection / expand collapse), `G` (extend to last node)
- Pending keys: `g...` (go to commands), `z...` (fold/zoom commands - reserved)
  - `gg` (extend selection to first node)
- Operations: `d`, `x` (delete), `y` (yank), `>`, `<` (indent/outdent)
- Exit: `V` (exit visual mode)

### Insert Mode (Editor)
Not defined in keybindings.go - handled by the `Editor` component in `internal/ui/editor.go`

Supports standard text editing operations.

## KeyBinding Structure

```go
type KeyBinding struct {
    Key         rune                  // The key to bind (e.g., 'j', 'k')
    Description string               // Human-readable description
    Handler     func(*App)           // Function to execute
}
```

### Example

```go
{
    Key:         'j',
    Description: "Move down",
    Handler: func(app *App) {
        app.tree.SelectNext()
    },
},
```

## Lookup Functions

### Normal Mode
```go
GetKeybindingByKey(key rune) *KeyBinding
```
Returns the keybinding for a given key in normal mode.

### Visual Mode
```go
GetVisualKeybindingByKey(key rune) *KeyBinding
```
Returns the keybinding for a given key in visual mode.

## Event Flow

1. **Event Reception**: `handleRawEvent()` receives raw terminal input
2. **Mode Routing**: Routes to appropriate handler based on current mode:
   - Command mode → `handleCommand()`
   - Search mode → search handler
   - Editor/Insert mode → `editor.HandleKey()`
   - Visual mode → `handleVisualMode()`
   - Normal mode → `handleKeypress()`
3. **Keybinding Lookup**: Character keys are looked up in the mode-specific keybindings
4. **Handler Execution**: The matching handler function is called with the App instance

## Adding New Keybindings

### To Normal Mode

1. Edit `internal/app/keybindings.go`
2. Add a new KeyBinding struct to the `InitializeKeybindings()` function
3. Implement the handler function
4. Rebuild the application

Example:
```go
{
    Key:         'Z',
    Description: "My new action",
    Handler: func(app *App) {
        // Your implementation here
        app.SetStatus("Action executed")
    },
},
```

### To Visual Mode

1. Edit `internal/app/keybindings.go`
2. Add a new KeyBinding struct to the `InitializeVisualKeybindings()` function
3. Implement the handler function
4. Rebuild the application

## Pending Keys

The keybindings system supports "pending keys" - keys that wait for a second character to complete the command (like Vim's `g`, `z`, etc).

### How It Works

Pending keys are defined in `InitializePendingKeybindings()` as a `PendingKeyBinding` structure:

```go
type PendingKeyBinding struct {
    Prefix      rune                // The first key (e.g., 'g' or 'z')
    Description string              // Description for help screen
    Sequences   map[rune]KeyBinding // Map of second key to keybinding
}
```

### Event Flow

1. When a pending key prefix is pressed (e.g., 'g'), it's stored in `pendingKeySeq`
2. On the next keypress, the system checks if it matches a registered sequence
3. If it matches, the handler executes and `pendingKeySeq` is cleared
4. If no match, `pendingKeySeq` is cleared and the key is processed normally

### Adding New Pending Key Sequences

To add a new `g_` or `z_` command, edit `InitializePendingKeybindings()`:

```go
{
    Prefix:      'g',
    Description: "Go to... (g + key)",
    Sequences: map[rune]KeyBinding{
        'g': {
            Key:         'g',
            Description: "Go to first node",
            Handler: func(app *App) {
                app.tree.SelectFirst()
            },
        },
        // Add more sequences here (ga, gb, etc.)
    },
},
```

### Currently Implemented Sequences

- `gg` - Go to first node (Normal mode and Visual mode)
- `z...` - Reserved prefix for future fold/zoom operations

## Special Key Handling

Some keys don't have rune representations and are handled separately via `tcell.EventKey.Key()`:

- Arrow keys: `KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`
- `Escape`: `KeyEscape`
- `Ctrl` combinations: `KeyCtrlA`, `KeyCtrlU`, etc.

These are typically handled with switch statements before keybinding lookup:

```go
switch ev.Key() {
case tcell.KeyDown:
    // Handle arrow down
case tcell.KeyEscape:
    // Handle escape
}
```

## Keybinding Conflicts

Currently, the system allows the same key to be bound in multiple modes, but they are separate:
- The same key can have different meanings in normal vs visual mode
- This is by design and enables vim-like modal behavior

## Future Enhancements

Potential improvements to the keybindings system:
- Dynamic keybinding remapping
- Keybinding profiles
- Macro recording
- Keybinding help display integrated with help screen
- Customizable keybindings from config file
