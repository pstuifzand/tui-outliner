package app

import (
	"log"
	"strings"

	"github.com/pstuifzand/tui-outliner/internal/export"
	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/search"
	"github.com/pstuifzand/tui-outliner/internal/socket"
)

// handleSocketMessage processes messages received from the Unix socket
func (app *App) handleSocketMessage(msg socket.Message) {
	log.Printf("Received socket message: command=%s, text=%s, target=%s", msg.Command, msg.Text, msg.Target)

	switch msg.Command {
	case socket.CommandAddNode:
		app.handleAddNodeCommand(msg)
	case socket.CommandExportMarkdown:
		app.handleExportMarkdownCommand(msg)
	case socket.CommandSearch:
		app.handleSocketSearchCommand(msg)
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

// handleExportMarkdownCommand processes an export_markdown command
func (app *App) handleExportMarkdownCommand(msg socket.Message) {
	// Validate export path
	if msg.ExportPath == "" {
		log.Printf("Export command missing export path")
		app.SetStatus("Error: Export path required")
		return
	}

	log.Printf("Exporting to markdown: '%s'", msg.ExportPath)

	// Sync tree items back to outline before exporting
	app.outline.Items = app.tree.GetItems()

	// Export to markdown
	if err := export.ExportToMarkdown(app.outline, msg.ExportPath); err != nil {
		log.Printf("Failed to export: %v", err)
		app.SetStatus("Error exporting to markdown: " + err.Error())
		return
	}

	log.Printf("Successfully exported to: %s", msg.ExportPath)
	app.SetStatus("Exported to " + msg.ExportPath)
}

// handleSocketSearchCommand processes a search command from socket
func (app *App) handleSocketSearchCommand(msg socket.Message) {
	// Validate query
	if msg.Query == "" {
		log.Printf("Search command missing query")
		if msg.ResponseChan != nil {
			msg.ResponseChan <- &socket.Response{
				Success: false,
				Message: "Query required",
			}
		}
		return
	}

	log.Printf("Searching with query: '%s', fields: %v, format: %s", msg.Query, msg.Fields, msg.Format)

	// Parse the search query
	filterExpr, err := search.ParseQuery(msg.Query)
	if err != nil {
		log.Printf("Failed to parse search query: %v", err)
		if msg.ResponseChan != nil {
			msg.ResponseChan <- &socket.Response{
				Success: false,
				Message: "Parse error: " + err.Error(),
			}
		}
		return
	}

	// Get matching items
	matches := search.GetMatchingItems(app.outline, filterExpr)
	log.Printf("Found %d matches", len(matches))

	// Determine fields to include (defaults based on backward compatibility)
	fields := msg.Fields
	if len(fields) == 0 {
		// Default to legacy fields for backward compatibility
		fields = []string{"text", "path", "attributes"}
	}

	// For markdown/list format, always include children
	includeChildren := msg.Format == "markdown" || msg.Format == "list"

	// Build results with requested fields
	results := make([]socket.SearchResult, 0, len(matches))
	for _, item := range matches {
		result := buildSearchResult(item, fields)

		// Add children for export formats
		if includeChildren {
			result["children"] = buildChildrenArray(item)
		}

		results = append(results, result)
	}

	// Send results through response channel
	if msg.ResponseChan != nil {
		msg.ResponseChan <- &socket.Response{
			Success: true,
			Message: "Search completed",
			Results: results,
		}
	}

	log.Printf("Search completed with %d results", len(results))
}

// buildChildrenArray recursively builds an array of children for an item
func buildChildrenArray(item *model.Item) []interface{} {
	if len(item.Children) == 0 {
		return nil
	}

	children := make([]interface{}, 0, len(item.Children))
	for _, child := range item.Children {
		childData := map[string]interface{}{
			"id":   child.ID,
			"text": child.Text,
		}

		// Add attributes if present
		if child.Metadata != nil && child.Metadata.Attributes != nil && len(child.Metadata.Attributes) > 0 {
			childData["attributes"] = child.Metadata.Attributes
		}

		// Recursively add children
		if grandchildren := buildChildrenArray(child); grandchildren != nil {
			childData["children"] = grandchildren
		}

		children = append(children, childData)
	}

	return children
}

// buildItemPath constructs a path array for an item showing its hierarchy with full node objects
func buildItemPath(item *model.Item) []interface{} {
	var path []interface{}
	current := item
	for current != nil {
		node := map[string]interface{}{
			"id":   current.ID,
			"text": current.Text,
		}
		if current.Metadata != nil && current.Metadata.Attributes != nil {
			node["attributes"] = current.Metadata.Attributes
		}
		path = append([]interface{}{node}, path...)
		current = current.Parent
	}
	return path
}

// buildSearchResult builds a search result with the requested fields
func buildSearchResult(item *model.Item, fields []string) socket.SearchResult {
	result := make(socket.SearchResult)

	for _, field := range fields {
		// Handle special attr:name syntax
		if strings.HasPrefix(field, "attr:") {
			attrName := strings.TrimPrefix(field, "attr:")
			if item.Metadata != nil && item.Metadata.Attributes != nil {
				result[field] = item.Metadata.Attributes[attrName]
			} else {
				result[field] = ""
			}
			continue
		}

		// Handle standard fields
		switch field {
		case "id":
			result["id"] = item.ID
		case "text":
			result["text"] = item.Text
		case "attributes":
			if item.Metadata == nil || item.Metadata.Attributes == nil {
				result["attributes"] = make(map[string]string)
			} else {
				result["attributes"] = item.Metadata.Attributes
			}
		case "created":
			if item.Metadata == nil {
				result["created"] = ""
			} else {
				result["created"] = item.Metadata.Created.Format("2006-01-02T15:04:05Z07:00")
			}
		case "modified":
			if item.Metadata == nil {
				result["modified"] = ""
			} else {
				result["modified"] = item.Metadata.Modified.Format("2006-01-02T15:04:05Z07:00")
			}
		case "tags":
			if item.Metadata == nil || item.Metadata.Tags == nil {
				result["tags"] = []string{}
			} else {
				result["tags"] = item.Metadata.Tags
			}
		case "depth":
			result["depth"] = getItemDepth(item)
		case "path":
			result["path"] = buildItemPath(item)
		case "parent_id":
			if item.Parent == nil {
				result["parent_id"] = ""
			} else {
				result["parent_id"] = item.Parent.ID
			}
		}
	}

	return result
}

// getItemDepth calculates the depth of an item in the tree
func getItemDepth(item *model.Item) int {
	depth := 0
	current := item.Parent
	for current != nil {
		depth++
		current = current.Parent
	}
	return depth
}
