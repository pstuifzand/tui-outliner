package app

import (
	"fmt"
	"log"
	"os"
	"strings"

	tmpl "github.com/pstuifzand/tui-outliner/internal/template"
)

var debugLog *log.Logger

func init() {
	// Initialize debug logger
	logFile, err := os.OpenFile("/tmp/tuo-template-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		// If we can't open log file, use stderr
		debugLog = log.New(os.Stderr, "[TEMPLATE] ", log.LstdFlags|log.Lshortfile)
	} else {
		debugLog = log.New(logFile, "[TEMPLATE] ", log.LstdFlags|log.Lshortfile)
	}
}

// handleTypedefCommand handles :typedef command
// Subcommands:
//
//	:typedef list           - Show all type definitions
//	:typedef add <key> <spec> - Add type definition
//	:typedef remove <key>   - Remove type definition
func (a *App) handleTypedefCommand(parts []string) {
	debugLog.Printf("handleTypedefCommand called with parts: %v", parts)

	if a.readOnly {
		debugLog.Printf("File is readonly, blocking modification")
		a.SetStatus("Cannot modify readonly file")
		return
	}

	if len(parts) < 2 {
		debugLog.Printf("Not enough arguments, expected at least 2, got %d", len(parts))
		a.SetStatus(":typedef list|add|remove ...")
		return
	}

	debugLog.Printf("Creating new type registry and loading from outline")
	registry := tmpl.NewTypeRegistry()
	if err := registry.LoadFromOutline(a.outline); err != nil {
		debugLog.Printf("Error loading types: %v", err)
		a.SetStatus(fmt.Sprintf("Error loading types: %s", err.Error()))
		return
	}
	debugLog.Printf("Loaded %d types from outline", len(registry.GetAll()))

	subcommand := parts[1]
	debugLog.Printf("Processing subcommand: %s", subcommand)

	switch subcommand {
	case "list":
		debugLog.Printf("Executing list command")
		a.handleTypedefList(registry)

	case "add":
		if len(parts) < 4 {
			debugLog.Printf("Add command missing arguments: got %d parts", len(parts))
			a.SetStatus(":typedef add <key> <spec> (e.g., :typedef add status enum|todo|done)")
			return
		}
		key := parts[2]
		spec := strings.Join(parts[3:], "|")
		debugLog.Printf("Executing add command: key=%s, spec=%s", key, spec)
		a.handleTypedefAdd(registry, key, spec)

	case "remove":
		if len(parts) < 3 {
			debugLog.Printf("Remove command missing arguments: got %d parts", len(parts))
			a.SetStatus(":typedef remove <key>")
			return
		}
		key := parts[2]
		debugLog.Printf("Executing remove command: key=%s", key)
		a.handleTypedefRemove(registry, key)

	default:
		debugLog.Printf("Unknown subcommand: %s", subcommand)
		a.SetStatus(fmt.Sprintf("Unknown typedef subcommand: %s", subcommand))
	}
}

// handleTypedefList shows all type definitions
func (a *App) handleTypedefList(registry *tmpl.TypeRegistry) {
	debugLog.Printf("handleTypedefList called")
	types := registry.GetAll()
	debugLog.Printf("Found %d types", len(types))

	if len(types) == 0 {
		debugLog.Printf("No types defined, displaying empty message")
		a.SetStatus("No type definitions defined")
		return
	}

	var msg strings.Builder
	msg.WriteString("Types: ")

	typeList := make([]string, 0, len(types))
	for key := range types {
		typeList = append(typeList, key)
		debugLog.Printf("  Type: %s (kind: %s)", key, types[key].Kind)
	}

	for i, key := range typeList {
		if i > 0 {
			msg.WriteString(" | ")
		}
		ts := types[key]
		spec := ts.Kind
		if len(ts.Values) > 0 {
			spec = spec + "(" + strings.Join(ts.Values, ",") + ")"
		}
		msg.WriteString(fmt.Sprintf("%s: %s", key, spec))
	}

	debugLog.Printf("Status message: %s", msg.String())
	a.SetStatus(msg.String())
}

// handleTypedefAdd adds a new type definition
func (a *App) handleTypedefAdd(registry *tmpl.TypeRegistry, key string, spec string) {
	debugLog.Printf("handleTypedefAdd called: key=%s, spec=%s", key, spec)

	if key == "" {
		debugLog.Printf("Type key is empty, rejecting")
		a.SetStatus("Type key cannot be empty")
		return
	}

	if spec == "" {
		debugLog.Printf("Type spec is empty, rejecting")
		a.SetStatus("Type spec cannot be empty")
		return
	}

	debugLog.Printf("Attempting to add type to registry")
	if err := registry.AddType(key, spec); err != nil {
		debugLog.Printf("Failed to add type: %v", err)
		a.SetStatus(fmt.Sprintf("Invalid type definition: %s", err.Error()))
		return
	}
	debugLog.Printf("Type added to registry successfully")

	// Save back to outline
	debugLog.Printf("Saving types back to outline")
	if err := registry.SaveToOutline(a.outline); err != nil {
		debugLog.Printf("Failed to save types to outline: %v", err)
		a.SetStatus(fmt.Sprintf("Failed to save type definitions: %s", err.Error()))
		return
	}
	debugLog.Printf("Types saved to outline successfully")

	debugLog.Printf("Checking outline structure after save")
	debugLog.Printf("Outline has %d items", len(a.outline.Items))
	for i, item := range a.outline.Items {
		debugLog.Printf("  Item %d: text=%s", i, item.Text)
		if item.Metadata != nil && item.Metadata.Attributes != nil {
			debugLog.Printf("    Attributes: %v", item.Metadata.Attributes)
		}
	}

	a.dirty = true
	debugLog.Printf("App marked as dirty, status set")
	a.SetStatus(fmt.Sprintf("Added type: %s", key))
}

// handleTypedefRemove removes a type definition
func (a *App) handleTypedefRemove(registry *tmpl.TypeRegistry, key string) {
	if key == "" {
		a.SetStatus("Type key cannot be empty")
		return
	}

	if registry.GetType(key) == nil {
		a.SetStatus(fmt.Sprintf("Type not found: %s", key))
		return
	}

	registry.RemoveType(key)

	// Save back to outline
	if err := registry.SaveToOutline(a.outline); err != nil {
		a.SetStatus(fmt.Sprintf("Failed to save type definitions: %s", err.Error()))
		return
	}

	a.dirty = true
	a.SetStatus(fmt.Sprintf("Removed type: %s", key))
}

// validateAttributeValue validates an attribute value against type definitions
// Returns true if valid (or no type definition exists), false if invalid
// Sets status message on error
func (a *App) validateAttributeValue(key string, value string) bool {
	// Load type registry from outline
	registry := tmpl.NewTypeRegistry()
	if err := registry.LoadFromOutline(a.outline); err != nil {
		// If we can't load types, allow the value (type system is optional)
		return true
	}

	// Get the type definition for this key
	typeSpec := registry.GetType(key)
	if typeSpec == nil {
		// No type definition for this key - that's OK
		return true
	}

	// Validate the value against the type
	if err := typeSpec.Validate(value); err != nil {
		debugLog.Printf("Validation failed for attribute %s=%s: %v", key, value, err)
		a.SetStatus(fmt.Sprintf("Invalid value for attribute '%s': %s", key, err.Error()))
		return false
	}

	debugLog.Printf("Validation passed for attribute %s=%s", key, value)
	return true
}
