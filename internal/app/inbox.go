package app

import (
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// findInboxNode searches for a node marked with @type=inbox attribute
// It searches recursively through all items in the outline
func (app *App) findInboxNode() *model.Item {
	return app.findInboxNodeRecursive(app.outline.Items)
}

// findInboxNodeRecursive recursively searches for inbox node in item list
func (app *App) findInboxNodeRecursive(items []*model.Item) *model.Item {
	for _, item := range items {
		// Check if this item has type=inbox attribute
		if item.Metadata.Attributes != nil {
			if typeVal, ok := item.Metadata.Attributes["type"]; ok && typeVal == "inbox" {
				return item
			}
		}
		// Recursively search children
		if len(item.Children) > 0 {
			if found := app.findInboxNodeRecursive(item.Children); found != nil {
				return found
			}
		}
	}
	return nil
}

// getOrCreateInboxNode finds an existing inbox node or creates a new one
// Returns the inbox node and a boolean indicating if it was created
func (app *App) getOrCreateInboxNode() (*model.Item, bool) {
	// Try to find existing inbox
	if inbox := app.findInboxNode(); inbox != nil {
		// Ensure inbox is expanded so new items are visible
		inbox.Expanded = true
		return inbox, false
	}

	// Create new inbox at root level
	inbox := model.NewItem("Inbox")
	inbox.Metadata.Attributes = make(map[string]string)
	inbox.Metadata.Attributes["type"] = "inbox"
	inbox.Expanded = true

	// Add to root level
	app.outline.Items = append(app.outline.Items, inbox)
	app.dirty = true

	return inbox, true
}

// addToInbox adds a new item to the inbox node
// If no inbox exists, one will be created
func (app *App) addToInbox(text string) error {
	// Clear search mode if active to ensure inbox is visible
	if app.search.IsActive() {
		app.search.Stop()
	}

	// Clear hoisting to ensure inbox is visible
	if app.tree.IsHoisted() {
		app.tree.Unhoist()
	}

	// Ensure Items is initialized (not nil)
	if app.outline.Items == nil {
		app.outline.Items = []*model.Item{}
	}

	inbox, created := app.getOrCreateInboxNode()

	// Create new item
	newItem := model.NewItem(text)
	inbox.AddChild(newItem)

	// Mark as dirty to trigger save
	app.dirty = true

	// Reset auto-save timer to save soon
	app.autoSaveTime = time.Now()

	// Rebuild the tree view to reflect changes
	app.tree.RebuildView()

	// Try to navigate to the inbox to make it visible
	items := app.tree.GetDisplayItems()
	for idx, item := range items {
		if item.Item.ID == inbox.ID {
			app.tree.SelectItem(idx)
			break
		}
	}

	// Set status message
	if created {
		app.SetStatus("Added to new inbox node")
	} else {
		app.SetStatus("Added to inbox")
	}

	// Force immediate screen update
	app.render()

	return nil
}
