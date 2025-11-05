// Package template provides template system functionality
package template

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

var typeDebugLog *log.Logger

func init() {
	logFile, err := os.OpenFile("/tmp/tuo-template-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		typeDebugLog = log.New(os.Stderr, "[TYPES] ", log.LstdFlags|log.Lshortfile)
	} else {
		typeDebugLog = log.New(logFile, "[TYPES] ", log.LstdFlags|log.Lshortfile)
	}
}

// TypeSpec represents a type definition for an attribute
type TypeSpec struct {
	Name   string   // e.g., "status"
	Kind   string   // enum, number, date, list, string, reference
	Values []string // For enum: values, for number: [min, max], for list: [itemtype]
}

// ParseTypeSpec parses a type specification string
// Format examples:
// - enum|todo|in-progress|done
// - number|1-5
// - date
// - list|string
// - string
// - reference|@type=task
func ParseTypeSpec(key string, spec string) (*TypeSpec, error) {
	if spec == "" {
		return nil, fmt.Errorf("empty type specification")
	}

	parts := strings.Split(spec, "|")
	kind := parts[0]

	ts := &TypeSpec{
		Name:   key,
		Kind:   kind,
		Values: parts[1:],
	}

	// Validate kind
	validKinds := map[string]bool{
		"enum":      true,
		"number":    true,
		"date":      true,
		"list":      true,
		"string":    true,
		"reference": true,
	}

	if !validKinds[kind] {
		return nil, fmt.Errorf("unknown type kind: %s", kind)
	}

	// Validate specific kinds
	if kind == "enum" && len(ts.Values) == 0 {
		return nil, fmt.Errorf("enum type requires at least one value")
	}

	if kind == "number" {
		if len(ts.Values) != 1 {
			return nil, fmt.Errorf("number type requires min-max specification")
		}
		// Validate min-max format
		parts := strings.Split(ts.Values[0], "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("number type must be in format min-max (e.g., 1-5)")
		}
		if _, err := strconv.Atoi(parts[0]); err != nil {
			return nil, fmt.Errorf("invalid min value for number type: %s", parts[0])
		}
		if _, err := strconv.Atoi(parts[1]); err != nil {
			return nil, fmt.Errorf("invalid max value for number type: %s", parts[1])
		}
	}

	if kind == "date" && len(ts.Values) > 0 {
		return nil, fmt.Errorf("date type should not have values")
	}

	if kind == "string" && len(ts.Values) > 0 {
		return nil, fmt.Errorf("string type should not have values")
	}

	if kind == "list" {
		if len(ts.Values) == 0 {
			return nil, fmt.Errorf("list type requires item type specification (e.g., list|string or list|enum|val1|val2)")
		}
		// Validate the item type specification
		itemKind := ts.Values[0]
		validItemKinds := map[string]bool{
			"string": true,
			"number": true,
			"enum":   true,
			"date":   true,
		}
		if !validItemKinds[itemKind] {
			return nil, fmt.Errorf("invalid item type '%s' for list (valid: string, number, enum, date)", itemKind)
		}

		// For enum items in list, validate that at least one value is provided
		if itemKind == "enum" && len(ts.Values) < 2 {
			return nil, fmt.Errorf("enum items in list require at least one value (e.g., list|enum|val1|val2)")
		}

		// For number items in list, validate min-max if provided
		if itemKind == "number" && len(ts.Values) > 1 {
			minMaxSpec := ts.Values[1]
			parts := strings.Split(minMaxSpec, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("number items in list must have min-max specification (e.g., list|number|1-5)")
			}
			if _, err := strconv.Atoi(parts[0]); err != nil {
				return nil, fmt.Errorf("invalid min value for number items in list: %s", parts[0])
			}
			if _, err := strconv.Atoi(parts[1]); err != nil {
				return nil, fmt.Errorf("invalid max value for number items in list: %s", parts[1])
			}
		}
	}

	return ts, nil
}

// Validate checks if a value is valid for this type spec
func (ts *TypeSpec) Validate(value string) error {
	return ts.validateValue(value, ts.Name)
}

// validateValue is the recursive validation function that handles nested types
func (ts *TypeSpec) validateValue(value, fieldName string) error {
	switch ts.Kind {
	case "enum":
		for _, v := range ts.Values {
			if v == value {
				return nil
			}
		}
		return fmt.Errorf("value '%s' is not a valid %s (must be one of: %s)",
			value, fieldName, strings.Join(ts.Values, ", "))

	case "number":
		num, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("value '%s' for %s is not a valid number", value, fieldName)
		}
		parts := strings.Split(ts.Values[0], "-")
		min, _ := strconv.Atoi(parts[0])
		max, _ := strconv.Atoi(parts[1])
		if num < min || num > max {
			return fmt.Errorf("value %d for %s is out of range (%d-%d)",
				num, fieldName, min, max)
		}
		return nil

	case "date":
		// Use time.Parse to validate YYYY-MM-DD format and actual date validity
		// This handles all edge cases like leap years, month boundaries, etc.
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("value '%s' for %s is not a valid date (format: YYYY-MM-DD)", value, fieldName)
		}
		return nil

	case "list":
		// Lists are stored as comma-separated values
		if value == "" {
			return fmt.Errorf("%s list cannot be empty", fieldName)
		}

		// Parse the list item type
		if len(ts.Values) == 0 {
			return fmt.Errorf("list type specification error: no item type")
		}

		itemKind := ts.Values[0]
		items := strings.Split(value, ",")

		// Create an item type spec from the list item kind
		itemSpec := &TypeSpec{
			Name:   fieldName,
			Kind:   itemKind,
			Values: ts.Values[1:], // Pass remaining values (enum values, number range, etc.)
		}

		// Validate each item in the list recursively
		for i, item := range items {
			item = strings.TrimSpace(item)
			if item == "" {
				return fmt.Errorf("%s list item %d cannot be empty", fieldName, i+1)
			}

			// Validate item using the item type spec (recursive call)
			itemFieldName := fmt.Sprintf("%s list item %d", fieldName, i+1)
			if err := itemSpec.validateValue(item, itemFieldName); err != nil {
				return err
			}
		}

		return nil

	case "string":
		// Any string is valid
		return nil

	case "reference":
		// Format: @type=task or other filter expressions
		// Just check that it's not empty and starts with @
		if value == "" {
			return fmt.Errorf("%s reference cannot be empty", fieldName)
		}
		return nil

	default:
		return fmt.Errorf("unknown type kind: %s", ts.Kind)
	}
}

// TypeRegistry manages all type definitions for an outline
type TypeRegistry struct {
	types map[string]*TypeSpec
}

// NewTypeRegistry creates an empty type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]*TypeSpec),
	}
}

// AddType adds or updates a type definition
func (tr *TypeRegistry) AddType(key string, spec string) error {
	ts, err := ParseTypeSpec(key, spec)
	if err != nil {
		return err
	}
	tr.types[key] = ts
	return nil
}

// RemoveType removes a type definition
func (tr *TypeRegistry) RemoveType(key string) {
	delete(tr.types, key)
}

// GetType gets a type definition by key
func (tr *TypeRegistry) GetType(key string) *TypeSpec {
	return tr.types[key]
}

// Validate validates an attribute against registered types
func (tr *TypeRegistry) Validate(key, value string) error {
	ts, exists := tr.types[key]
	if !exists {
		// No type defined for this attribute - that's OK
		return nil
	}
	return ts.Validate(value)
}

// ValidateItem validates all attributes of an item against registered types
func (tr *TypeRegistry) ValidateItem(item *model.Item) error {
	if item.Metadata == nil || item.Metadata.Attributes == nil {
		return nil
	}

	for key, value := range item.Metadata.Attributes {
		if err := tr.Validate(key, value); err != nil {
			return err
		}
	}
	return nil
}

// LoadFromOutline loads type definitions from the outline's TypeDefinitions field
func (tr *TypeRegistry) LoadFromOutline(outline *model.Outline) error {
	if outline == nil {
		typeDebugLog.Printf("LoadFromOutline: outline is nil")
		return nil
	}

	typeDebugLog.Printf("LoadFromOutline called, outline has %d type definitions", len(outline.TypeDefinitions))

	if len(outline.TypeDefinitions) == 0 {
		typeDebugLog.Printf("No type definitions in outline")
		return nil
	}

	for key, spec := range outline.TypeDefinitions {
		typeDebugLog.Printf("  Loading type: %s = %s", key, spec)
		if err := tr.AddType(key, spec); err != nil {
			typeDebugLog.Printf("  Failed to load type %s: %v", key, err)
			return fmt.Errorf("invalid type definition for %s: %w", key, err)
		}
		typeDebugLog.Printf("  Successfully loaded type: %s", key)
	}

	typeDebugLog.Printf("LoadFromOutline complete, registry now has %d types", len(tr.types))
	return nil
}

// SaveToOutline saves type definitions to the outline's TypeDefinitions field
func (tr *TypeRegistry) SaveToOutline(outline *model.Outline) error {
	if outline == nil {
		typeDebugLog.Printf("SaveToOutline: outline is nil")
		return fmt.Errorf("outline cannot be nil")
	}

	typeDebugLog.Printf("SaveToOutline called with %d types", len(tr.types))

	// Initialize TypeDefinitions if nil
	if outline.TypeDefinitions == nil {
		typeDebugLog.Printf("Initializing outline.TypeDefinitions")
		outline.TypeDefinitions = make(map[string]string)
	}

	// Clear existing type definitions
	typeDebugLog.Printf("Clearing existing type definitions")
	outline.TypeDefinitions = make(map[string]string)

	// Add all types
	typeDebugLog.Printf("Adding %d types to outline.TypeDefinitions", len(tr.types))
	for key, ts := range tr.types {
		spec := ts.Kind
		if len(ts.Values) > 0 {
			spec = spec + "|" + strings.Join(ts.Values, "|")
		}
		outline.TypeDefinitions[key] = spec
		typeDebugLog.Printf("  Saved type: %s = %s", key, spec)
	}

	typeDebugLog.Printf("SaveToOutline complete, outline.TypeDefinitions now has %d types", len(outline.TypeDefinitions))
	return nil
}

// GetAll returns all registered type definitions
func (tr *TypeRegistry) GetAll() map[string]*TypeSpec {
	return tr.types
}

