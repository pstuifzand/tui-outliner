package app

import (
	"fmt"
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
	Prefix      rune                // The first key (e.g., 'g' or 'z')
	Description string              // Description of what the pending key does
	Sequences   map[rune]KeyBinding // Map of second key to keybinding
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
				wasSearching := app.search.IsActive()
				app.search.Start()
				// When hoisted, search only within the hoisted subtree
				searchItems := app.outline.GetAllItems()
				if app.tree.IsHoisted() {
					hoistedItem := app.tree.GetHoistedItem()
					if hoistedItem != nil {
						// Get all items within the hoisted subtree
						searchItems = ui.GetAllItemsRecursive(hoistedItem)
					}
				}
				app.search.SetAllItems(searchItems)
				// Only auto-navigate to first match if we just started a new search
				// (not if we're clearing and restarting an existing search)
				if !wasSearching && app.search.GetMatchCount() > 0 {
					firstMatch := app.search.GetCurrentMatch()
					if firstMatch != nil {
						// Expand all parent nodes of the first match so it becomes visible
						app.tree.ExpandParents(firstMatch)
						// Find and select first match in the main tree
						items := app.tree.GetDisplayItems()
						for idx, dispItem := range items {
							if dispItem.Item.ID == firstMatch.ID {
								app.tree.SelectItem(idx)
								break
							}
						}
					}
				}
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
			Key:         'n',
			Description: "Next search match",
			Handler: func(app *App) {
				if !app.search.HasResults() {
					app.SetStatus("No active search")
					return
				}
				if !app.search.NextMatch() {
					app.SetStatus("No search results")
					return
				}
				// Navigate to the current match
				currentMatch := app.search.GetCurrentMatch()
				if currentMatch != nil {
					app.tree.ExpandParents(currentMatch)
					items := app.tree.GetDisplayItems()
					for idx, dispItem := range items {
						if dispItem.Item.ID == currentMatch.ID {
							app.tree.SelectItem(idx)
							break
						}
					}
					matchNum := app.search.GetCurrentMatchNumber()
					totalMatches := app.search.GetMatchCount()
					app.SetStatus(fmt.Sprintf("Match %d of %d", matchNum, totalMatches))
				}
			},
		},
		{
			Key:         'N',
			Description: "Previous search match",
			Handler: func(app *App) {
				if !app.search.HasResults() {
					app.SetStatus("No active search")
					return
				}
				if !app.search.PrevMatch() {
					app.SetStatus("No search results")
					return
				}
				// Navigate to the current match
				currentMatch := app.search.GetCurrentMatch()
				if currentMatch != nil {
					app.tree.ExpandParents(currentMatch)
					items := app.tree.GetDisplayItems()
					for idx, dispItem := range items {
						if dispItem.Item.ID == currentMatch.ID {
							app.tree.SelectItem(idx)
							break
						}
					}
					matchNum := app.search.GetCurrentMatchNumber()
					totalMatches := app.search.GetMatchCount()
					app.SetStatus(fmt.Sprintf("Match %d of %d", matchNum, totalMatches))
				}
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
			Key:         '@',
			Description: "Edit attributes",
			Handler: func(app *App) {
				selected := app.tree.GetSelected()
				if selected != nil {
					app.attributeEditor.Show(selected)
					app.SetStatus("Editing attributes (q to quit)")
				} else {
					app.SetStatus("No item selected")
				}
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
		{
			Key:         '-',
			Description: "Select parent (or hoist to parent if at hoisted root)",
			Handler: func(app *App) {
				// First try normal parent selection
				if app.tree.SelectParent() {
					return
				}

				// If that failed and we're at root of hoisted view, hoist to parent
				if app.tree.IsAtRootOfHoistedView() {
					if app.tree.HoistToParent() {
						hoistedItem := app.tree.GetHoistedItem()
						if hoistedItem != nil {
							app.SetStatus(fmt.Sprintf("Hoisted to parent: %s", hoistedItem.Text))
						} else {
							app.SetStatus("Unhoisted")
						}
						return
					}
				}

				// Otherwise report we're at root
				app.SetStatus("Already at root")
			},
		},
	}
}

// InitializePendingKeybindings sets up pending key bindings (keys that wait for a second key)
func (a *App) InitializePendingKeybindings() []PendingKeyBinding {
	return []PendingKeyBinding{
		{
			Prefix:      'g',
			Description: "Go to... or go (URL)... (g + key)",
			Sequences: map[rune]KeyBinding{
				'g': {
					Key:         'g',
					Description: "Go to first node",
					Handler: func(app *App) {
						app.tree.SelectFirst()
					},
				},
				'o': {
					Key:         'o',
					Description: "Open URL from 'url' attribute (xdg-open)",
					Handler: func(app *App) {
						app.handleGoCommand()
					},
				},
			},
		},
		{
			Prefix:      'z',
			Description: "Fold/Hoist... (z + key)",
			Sequences: map[rune]KeyBinding{
				'h': {
					Key:         'h',
					Description: "Hoist (focus on subtree)",
					Handler: func(app *App) {
						if app.tree.Hoist() {
							app.SetStatus("Hoisted - showing only this subtree (zu to unhoist)")
						} else {
							app.SetStatus("Cannot hoist: item has no children")
						}
					},
				},
				'u': {
					Key:         'u',
					Description: "Unhoist (return to full view)",
					Handler: func(app *App) {
						if app.tree.Unhoist() {
							app.SetStatus("Unhoisted - showing full tree")
						} else {
							app.SetStatus("Not currently hoisted")
						}
					},
				},
				'C': {
					Key:         'C',
					Description: "Close all (collapse recursively)",
					Handler: func(app *App) {
						app.tree.CollapseRecursive()
						app.SetStatus("Closed all")
						app.dirty = true
					},
				},
				'O': {
					Key:         'O',
					Description: "Open all (expand recursively)",
					Handler: func(app *App) {
						app.tree.ExpandRecursive()
						app.SetStatus("Opened all")
						app.dirty = true
					},
				},
				'c': {
					Key:         'c',
					Description: "Close all children",
					Handler: func(app *App) {
						app.tree.CollapseAllChildren()
						app.SetStatus("Closed all children")
						app.dirty = true
					},
				},
				's': {
					Key:         's',
					Description: "Close all siblings",
					Handler: func(app *App) {
						app.tree.CollapseSiblings()
						app.SetStatus("Closed all siblings")
						app.dirty = true
					},
				},
			},
		},
		{
			Prefix:      '[',
			Description: "Previous... ([ + key)",
			Sequences: map[rune]KeyBinding{
				'[': {
					Key:         '[',
					Description: "Go to previous sibling",
					Handler: func(app *App) {
						if !app.tree.SelectPrevSibling() {
							app.SetStatus("No previous sibling")
						}
					},
				},
				'd': {
					Key:         'd',
					Description: "Go to previous item with date",
					Handler: func(app *App) {
						if !app.tree.FindPrevDateItem() {
							app.SetStatus("No items with dates found")
						} else {
							app.SetStatus("Found previous item with date")
						}
					},
				},
				'w': {
					Key:         'w',
					Description: "Go to previous item this week",
					Handler: func(app *App) {
						if !app.tree.FindPrevItemWithDateInterval("week") {
							app.SetStatus("No items this week found")
						} else {
							app.SetStatus("Found item this week")
						}
					},
				},
				'm': {
					Key:         'm',
					Description: "Go to previous item this month",
					Handler: func(app *App) {
						if !app.tree.FindPrevItemWithDateInterval("month") {
							app.SetStatus("No items this month found")
						} else {
							app.SetStatus("Found item this month")
						}
					},
				},
				'y': {
					Key:         'y',
					Description: "Go to previous item this year",
					Handler: func(app *App) {
						if !app.tree.FindPrevItemWithDateInterval("year") {
							app.SetStatus("No items this year found")
						} else {
							app.SetStatus("Found item this year")
						}
					},
				},
			},
		},
		{
			Prefix:      ']',
			Description: "Next... (] + key)",
			Sequences: map[rune]KeyBinding{
				']': {
					Key:         ']',
					Description: "Go to next sibling",
					Handler: func(app *App) {
						if !app.tree.SelectNextSibling() {
							app.SetStatus("No next sibling")
						}
					},
				},
				'd': {
					Key:         'd',
					Description: "Go to next item with date",
					Handler: func(app *App) {
						if !app.tree.FindNextDateItem() {
							app.SetStatus("No items with dates found")
						} else {
							app.SetStatus("Found next item with date")
						}
					},
				},
				'w': {
					Key:         'w',
					Description: "Go to next item this week",
					Handler: func(app *App) {
						if !app.tree.FindNextItemWithDateInterval("week") {
							app.SetStatus("No items this week found")
						} else {
							app.SetStatus("Found item this week")
						}
					},
				},
				'm': {
					Key:         'm',
					Description: "Go to next item this month",
					Handler: func(app *App) {
						if !app.tree.FindNextItemWithDateInterval("month") {
							app.SetStatus("No items this month found")
						} else {
							app.SetStatus("Found item this month")
						}
					},
				},
				'y': {
					Key:         'y',
					Description: "Go to next item this year",
					Handler: func(app *App) {
						if !app.tree.FindNextItemWithDateInterval("year") {
							app.SetStatus("No items this year found")
						} else {
							app.SetStatus("Found item this year")
						}
					},
				},
			},
		},
		{
			Prefix:      'a',
			Description: "Attribute... (a + key)",
			Sequences: map[rune]KeyBinding{
				'a': {
					Key:         'a',
					Description: "Add attribute (prompt for key and value)",
					Handler: func(app *App) {
						app.SetStatus("Use :attr add <key> <value> to add attributes")
					},
				},
				'd': {
					Key:         'd',
					Description: "Delete attribute (prompt for key)",
					Handler: func(app *App) {
						app.SetStatus("Use :attr del <key> to delete attributes")
					},
				},
				'c': {
					Key:         'c',
					Description: "Change/edit attribute value (prompt for key)",
					Handler: func(app *App) {
						app.SetStatus("Use :attr add <key> <value> to change attributes")
					},
				},
				'v': {
					Key:         'v',
					Description: "View all attributes for this item",
					Handler: func(app *App) {
						selected := app.tree.GetSelected()
						if selected != nil && selected.Metadata != nil && len(selected.Metadata.Attributes) > 0 {
							var attrs []string
							for k, v := range selected.Metadata.Attributes {
								attrs = append(attrs, k+": "+v)
							}
							app.SetStatus("Attributes: " + attrs[0])
						} else {
							app.SetStatus("No attributes for this item")
						}
					},
				},
			},
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
