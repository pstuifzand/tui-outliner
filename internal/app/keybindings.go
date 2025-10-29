package app

import (
	"github.com/pstuifzand/tui-outliner/internal/ui"
)

// KeyBinding represents a key binding with its description and handler
type KeyBinding struct {
	Key         rune
	Description string
	Handler     func(*App)
}

// GetKey returns the key of this keybinding
func (kb *KeyBinding) GetKey() rune {
	return kb.Key
}

// GetDescription returns the description of this keybinding
func (kb *KeyBinding) GetDescription() string {
	return kb.Description
}

// PendingKeyBinding represents a pending key (like 'g' or 'z') that waits for a second key
type PendingKeyBinding struct {
	Prefix      rune                          // The first key (e.g., 'g' or 'z')
	Description string                        // Description of what the pending key does
	Sequences   map[rune]KeyBinding           // Map of second key to keybinding
}

// GetKey returns the prefix key
func (pkb *PendingKeyBinding) GetKey() rune {
	return pkb.Prefix
}

// GetDescription returns the description
func (pkb *PendingKeyBinding) GetDescription() string {
	return pkb.Description
}

// GetSequences returns a map of second key to description for display in help
func (pkb *PendingKeyBinding) GetSequences() map[rune]string {
	result := make(map[rune]string)
	for key, binding := range pkb.Sequences {
		result[key] = binding.Description
	}
	return result
}

// InitializeKeybindings sets up all the key bindings
func (a *App) InitializeKeybindings() []KeyBinding {
	return []KeyBinding{
		{
			Key:         'j',
			Description: "Move down",
			Handler: func(app *App) {
				app.tree.SelectNext()
			},
		},
		{
			Key:         'k',
			Description: "Move up",
			Handler: func(app *App) {
				app.tree.SelectPrev()
			},
		},
		{
			Key:         'h',
			Description: "Collapse item",
			Handler: func(app *App) {
				app.tree.Collapse()
			},
		},
		{
			Key:         'l',
			Description: "Expand item",
			Handler: func(app *App) {
				app.tree.Expand()
			},
		},
		{
			Key:         'J',
			Description: "Move node down",
			Handler: func(app *App) {
				if app.tree.MoveItemDown() {
					app.SetStatus("Moved item down")
					app.dirty = true
				}
			},
		},
		{
			Key:         'K',
			Description: "Move node up",
			Handler: func(app *App) {
				if app.tree.MoveItemUp() {
					app.SetStatus("Moved item up")
					app.dirty = true
				}
			},
		},
		{
			Key:         'i',
			Description: "Edit item (cursor at start)",
			Handler: func(app *App) {
				selected := app.tree.GetSelected()
				if selected != nil {
					app.editor = ui.NewEditor(selected)
					app.editor.Start()
					app.editor.SetCursorToStart()
					app.mode = InsertMode
				}
			},
		},
		{
			Key:         'c',
			Description: "Change (replace) item text",
			Handler: func(app *App) {
				selected := app.tree.GetSelected()
				if selected != nil {
					app.editor = ui.NewEditor(selected)
					app.editor.SetText("")
					app.editor.Start()
					app.mode = InsertMode
				}
			},
		},
		{
			Key:         'A',
			Description: "Append (edit at end of text)",
			Handler: func(app *App) {
				selected := app.tree.GetSelected()
				if selected != nil {
					app.editor = ui.NewEditor(selected)
					app.editor.Start()
					app.mode = InsertMode
				}
			},
		},
		{
			Key:         'O',
			Description: "Insert new item before",
			Handler: func(app *App) {
				app.tree.AddItemBefore("")
				app.SetStatus("Created new item before")
				app.dirty = true
				// Enter insert mode for the new item
				selected := app.tree.GetSelected()
				if selected != nil {
					app.editor = ui.NewEditor(selected)
					app.editor.Start()
					app.mode = InsertMode
				}
			},
		},
		{
			Key:         'o',
			Description: "Insert new item after",
			Handler: func(app *App) {
				app.tree.AddItemAfter("")
				app.SetStatus("Created new item after")
				app.dirty = true
				// Enter insert mode for the new item
				selected := app.tree.GetSelected()
				if selected != nil {
					app.editor = ui.NewEditor(selected)
					app.editor.Start()
					app.mode = InsertMode
				}
			},
		},
		{
			Key:         'd',
			Description: "Delete item",
			Handler: func(app *App) {
				selected := app.tree.GetSelected()
				app.clipboard = selected
				if app.tree.DeleteSelected() {
					app.SetStatus("Deleted item")
					app.dirty = true
				}
			},
		},
		{
			Key:         'y',
			Description: "Yank (copy) item",
			Handler: func(app *App) {
				selected := app.tree.GetSelected()
				if selected != nil {
					app.clipboard = selected
					app.SetStatus("Yanked item")
				}
			},
		},
		{
			Key:         'p',
			Description: "Paste item below",
			Handler: func(app *App) {
				if app.clipboard != nil {
					if app.tree.PasteAfter(app.clipboard) {
						app.SetStatus("Pasted item")
						app.dirty = true
						app.clipboard = nil
					}
				}
			},
		},
		{
			Key:         'P',
			Description: "Paste item above",
			Handler: func(app *App) {
				if app.clipboard != nil {
					if app.tree.PasteBefore(app.clipboard) {
						app.SetStatus("Pasted item")
						app.dirty = true
						app.clipboard = nil
					}
				}
			},
		},
		{
			Key:         '>',
			Description: "Indent item",
			Handler: func(app *App) {
				if app.tree.Indent() {
					app.SetStatus("Indented")
					app.dirty = true
				}
			},
		},
		{
			Key:         '<',
			Description: "Outdent item",
			Handler: func(app *App) {
				if app.tree.Outdent() {
					app.SetStatus("Outdented")
					app.dirty = true
				}
			},
		},
		{
			Key:         '/',
			Description: "Search",
			Handler: func(app *App) {
				app.search.Start()
				app.search.SetAllItems(app.outline.GetAllItems())
			},
		},
		{
			Key:         '?',
			Description: "Toggle help",
			Handler: func(app *App) {
				app.help.Toggle()
			},
		},
		{
			Key:         ':',
			Description: "Command mode",
			Handler: func(app *App) {
				app.command.Start()
			},
		},
		{
			Key:         'V',
			Description: "Visual mode (line-wise selection)",
			Handler: func(app *App) {
				if app.mode == NormalMode {
					app.mode = VisualMode
					app.visualAnchor = app.tree.GetSelectedIndex()
					app.SetStatus("Visual mode")
				}
			},
		},
		{
			Key:         'G',
			Description: "Go to last node",
			Handler: func(app *App) {
				app.tree.SelectLast()
			},
		},
	}
}

// InitializePendingKeybindings sets up pending key bindings (keys that wait for a second key)
func (a *App) InitializePendingKeybindings() []PendingKeyBinding {
	return []PendingKeyBinding{
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
			},
		},
		{
			Prefix:      'z',
			Description: "Fold... (z + key)",
			Sequences:   map[rune]KeyBinding{}, // Placeholder for future fold operations
		},
	}
}

// InitializeVisualKeybindings sets up all the key bindings for visual mode
func (a *App) InitializeVisualKeybindings() []KeyBinding {
	return []KeyBinding{
		{
			Key:         'j',
			Description: "Extend selection down",
			Handler: func(app *App) {
				app.tree.SelectNext()
			},
		},
		{
			Key:         'k',
			Description: "Extend selection up",
			Handler: func(app *App) {
				app.tree.SelectPrev()
			},
		},
		{
			Key:         'h',
			Description: "Collapse item",
			Handler: func(app *App) {
				app.tree.Collapse()
			},
		},
		{
			Key:         'l',
			Description: "Expand item",
			Handler: func(app *App) {
				app.tree.Expand()
			},
		},
		{
			Key:         'V',
			Description: "Exit visual mode",
			Handler: func(app *App) {
				app.mode = NormalMode
				app.visualAnchor = -1
				app.SetStatus("Exited visual mode")
			},
		},
		{
			Key:         'd',
			Description: "Delete selected items",
			Handler: func(app *App) {
				app.deleteVisualSelection()
			},
		},
		{
			Key:         'x',
			Description: "Delete selected items",
			Handler: func(app *App) {
				app.deleteVisualSelection()
			},
		},
		{
			Key:         'y',
			Description: "Yank (copy) selected items",
			Handler: func(app *App) {
				app.yankVisualSelection()
			},
		},
		{
			Key:         '>',
			Description: "Indent selected items",
			Handler: func(app *App) {
				app.indentVisualSelection()
			},
		},
		{
			Key:         '<',
			Description: "Outdent selected items",
			Handler: func(app *App) {
				app.outdentVisualSelection()
			},
		},
		{
			Key:         'G',
			Description: "Extend selection to last node",
			Handler: func(app *App) {
				app.tree.SelectLast()
			},
		},
	}
}

// GetKeybindingByKey returns a keybinding for a given key
func (a *App) GetKeybindingByKey(key rune) *KeyBinding {
	for _, kb := range a.keybindings {
		if kb.Key == key {
			return &kb
		}
	}
	return nil
}

// GetVisualKeybindingByKey returns a visual mode keybinding for a given key
func (a *App) GetVisualKeybindingByKey(key rune) *KeyBinding {
	visualKeybindings := a.InitializeVisualKeybindings()
	for _, kb := range visualKeybindings {
		if kb.Key == key {
			return &kb
		}
	}
	return nil
}

// GetPendingKeyBindingByPrefix returns a pending keybinding for a prefix key
func (a *App) GetPendingKeyBindingByPrefix(prefix rune) *PendingKeyBinding {
	for i := range a.pendingKeybindings {
		if a.pendingKeybindings[i].Prefix == prefix {
			return &a.pendingKeybindings[i]
		}
	}
	return nil
}

// IsPendingKeyPrefix checks if a key is a pending key prefix
func (a *App) IsPendingKeyPrefix(key rune) bool {
	return a.GetPendingKeyBindingByPrefix(key) != nil
}
