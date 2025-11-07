package app

import (
	"log"

	"github.com/pstuifzand/tui-outliner/internal/socket"
)

// handleSocketMessage processes messages received from the Unix socket
func (app *App) handleSocketMessage(msg socket.Message) {
	log.Printf("Received socket message: command=%s, text=%s, target=%s", msg.Command, msg.Text, msg.Target)

	switch msg.Command {
	case socket.CommandAddNode:
		app.handleAddNodeCommand(msg)
	default:
		log.Printf("Unknown socket command: %s", msg.Command)
	}
}

// handleAddNodeCommand processes an add_node command
func (app *App) handleAddNodeCommand(msg socket.Message) {
	// Validate text
	if msg.Text == "" {
		log.Printf("Add node command missing text")
		return
	}

	// Determine target (default to inbox)
	target := msg.Target
	if target == "" {
		target = "inbox"
	}

	// Currently only support inbox target
	if target != "inbox" {
		log.Printf("Unsupported target: %s (only 'inbox' is supported)", target)
		app.SetStatus("Error: Only 'inbox' target is supported")
		return
	}

	log.Printf("Adding item to inbox: '%s'", msg.Text)
	log.Printf("Search active: %v, Hoisted: %v", app.search.IsActive(), app.tree.IsHoisted())
	if len(msg.Attributes) > 0 {
		log.Printf("Attributes: %v", msg.Attributes)
	}

	// Add to inbox
	if err := app.addToInbox(msg.Text, msg.Attributes); err != nil {
		log.Printf("Failed to add item to inbox: %v", err)
		app.SetStatus("Error adding item to inbox")
		return
	}

	log.Printf("Successfully added item to inbox: %s", msg.Text)
	log.Printf("Tree now has %d root items", len(app.outline.Items))
}
