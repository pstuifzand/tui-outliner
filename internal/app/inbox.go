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
// The item is added quietly without disrupting the user's current view
func (app *App) addToInbox(text string, attributes map[string]string) error {
	// Ensure Items is initialized (not nil)
	if app.outline.Items == nil {
		app.outline.Items = []*model.Item{}
	}

	inbox, created := app.getOrCreateInboxNode()

	// Create new item
	newItem := model.NewItem(text)

	// Set attributes if provided
	if len(attributes) > 0 {
		if newItem.Metadata.Attributes == nil {
			newItem.Metadata.Attributes = make(map[string]string)
		}
		for key, value := range attributes {
			newItem.Metadata.Attributes[key] = value
		}
	}

	inbox.AddChild(newItem)

	// Mark as dirty to trigger save
	app.dirty = true

	// Reset auto-save timer to save soon
	app.autoSaveTime = time.Now()

	// Update tree view with current outline items (in case slice was reallocated)
	app.tree.SetItems(app.outline.Items)

	// Set status message
	if created {
		app.SetStatus("Added to new inbox node")
	} else {
		app.SetStatus("Added to inbox")
	}

	// Update screen to show new item if inbox is currently visible
	app.render()

	return nil
}
