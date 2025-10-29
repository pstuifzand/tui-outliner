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
			Key:         'i',
			Description: "Edit item",
			Handler: func(app *App) {
				selected := app.tree.GetSelected()
				if selected != nil {
					app.editor = ui.NewEditor(selected)
					app.editor.Start()
					app.mode = "INSERT"
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
					app.mode = "INSERT"
				}
			},
		},
		{
			Key:         'o',
			Description: "Insert new item after",
			Handler: func(app *App) {
				app.tree.AddItemAfter("Type here...")
				app.SetStatus("Created new item after")
				app.dirty = true
			},
		},
		{
			Key:         'a',
			Description: "Insert new item as child",
			Handler: func(app *App) {
				app.tree.AddItemAsChild("Type here...")
				app.SetStatus("Created new child item")
				app.dirty = true
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
