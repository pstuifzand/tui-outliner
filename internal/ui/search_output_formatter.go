package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// OutputFormat specifies how search results should be formatted
type OutputFormat int

const (
	OutputFormatText OutputFormat = iota
	OutputFormatFields
	OutputFormatJSON
	OutputFormatJSONL
)


// SearchOutputFormatter handles formatting search results in different formats
type SearchOutputFormatter struct{}

// NewSearchOutputFormatter creates a new search output formatter
func NewSearchOutputFormatter() *SearchOutputFormatter {
	return &SearchOutputFormatter{}
}

// FormatResults formats search results according to the specified format
func (f *SearchOutputFormatter) FormatResults(
	items []*model.Item,
	format OutputFormat,
	fields []string,
	outline *model.Outline,
) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	switch format {
	case OutputFormatFields:
		return f.formatFields(items, fields, outline), nil
	case OutputFormatJSON:
		return f.formatJSON(items, fields, outline)
	case OutputFormatJSONL:
		return f.formatJSONL(items, fields, outline)
	case OutputFormatText:
		fallthrough
	default:
		return "", fmt.Errorf("text format should not use formatter (display as search node instead)")
	}
}

// formatFields formats results as tab-separated values
func (f *SearchOutputFormatter) formatFields(items []*model.Item, fields []string, outline *model.Outline) string {
	if len(fields) == 0 {
		fields = []string{"id", "text", "attributes"}
	}

	var lines []string
	for _, item := range items {
		line := f.formatItemAsFields(item, fields, outline)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

// formatItemAsFields formats a single item as tab-separated fields
func (f *SearchOutputFormatter) formatItemAsFields(item *model.Item, fields []string, outline *model.Outline) string {
	var values []string
	for _, field := range fields {
		val := f.getFieldValue(item, field, outline)
		// Handle multi-value fields
		switch v := val.(type) {
		case []map[string]interface{}:
			// For path, extract text and join with " > "
			var pathTexts []string
			for _, node := range v {
				if text, ok := node["text"].(string); ok {
					pathTexts = append(pathTexts, text)
				}
			}
			values = append(values, strings.Join(pathTexts, " > "))
		case []string:
			// For tags, join with comma
			values = append(values, strings.Join(v, ","))
		case map[string]string:
			// For attributes, format as @key=value space-separated
			var attrs []string
			for k, v := range v {
				attrs = append(attrs, fmt.Sprintf("@%s=%s", k, v))
			}
			values = append(values, strings.Join(attrs, " "))
		default:
			// For all other types, convert to string
			values = append(values, fmt.Sprintf("%v", v))
		}
	}
	return strings.Join(values, "\t")
}

// formatJSON formats results as a JSON array
func (f *SearchOutputFormatter) formatJSON(items []*model.Item, fields []string, outline *model.Outline) (string, error) {
	if len(fields) == 0 {
		fields = []string{"id", "text", "attributes", "created", "modified", "tags", "depth", "path"}
	}

	var result []interface{}
	for _, item := range items {
		obj := f.getItemAsObject(item, fields, outline)
		result = append(result, obj)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	return string(data), err
}

// formatJSONL formats results as JSON Lines (one JSON object per line)
func (f *SearchOutputFormatter) formatJSONL(items []*model.Item, fields []string, outline *model.Outline) (string, error) {
	if len(fields) == 0 {
		fields = []string{"id", "text", "attributes", "created", "modified", "tags", "depth", "path"}
	}

	var lines []string
	for _, item := range items {
		obj := f.getItemAsObject(item, fields, outline)
		data, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}
		lines = append(lines, string(data))
	}

	return strings.Join(lines, "\n"), nil
}

// getItemAsObject converts an item to a map for JSON output
func (f *SearchOutputFormatter) getItemAsObject(item *model.Item, fields []string, outline *model.Outline) map[string]interface{} {
	obj := make(map[string]interface{})
	for _, field := range fields {
		obj[field] = f.getFieldValue(item, field, outline)
	}
	return obj
}

// getFieldValue extracts a field value from an item
func (f *SearchOutputFormatter) getFieldValue(item *model.Item, field string, outline *model.Outline) interface{} {
	// Handle special attr:name syntax
	if strings.HasPrefix(field, "attr:") {
		attrName := strings.TrimPrefix(field, "attr:")
		if item.Metadata != nil && item.Metadata.Attributes != nil {
			return item.Metadata.Attributes[attrName]
		}
		return ""
	}

	switch field {
	case "id":
		return item.ID
	case "text":
		return item.Text
	case "attributes":
		if item.Metadata == nil || item.Metadata.Attributes == nil {
			return make(map[string]string)
		}
		return item.Metadata.Attributes
	case "created":
		if item.Metadata == nil {
			return ""
		}
		return item.Metadata.Created.Format("2006-01-02T15:04:05Z07:00")
	case "modified":
		if item.Metadata == nil {
			return ""
		}
		return item.Metadata.Modified.Format("2006-01-02T15:04:05Z07:00")
	case "tags":
		if item.Metadata == nil || item.Metadata.Tags == nil {
			return []string{}
		}
		return item.Metadata.Tags
	case "depth":
		return f.getItemDepth(item)
	case "path":
		return f.getItemPath(item, outline)
	case "parent_id":
		if item.Parent == nil {
			return ""
		}
		return item.Parent.ID
	default:
		return ""
	}
}

// getItemDepth calculates the depth of an item in the tree
func (f *SearchOutputFormatter) getItemDepth(item *model.Item) int {
	depth := 0
	current := item.Parent
	for current != nil {
		depth++
		current = current.Parent
	}
	return depth
}

// getItemPath builds the hierarchical path to an item as an array of node objects
func (f *SearchOutputFormatter) getItemPath(item *model.Item, outline *model.Outline) []map[string]interface{} {
	var parts []map[string]interface{}
	current := item
	for current != nil {
		// Build a node object with key fields
		node := map[string]interface{}{
			"id":   current.ID,
			"text": current.Text,
		}
		// Include attributes if present
		if current.Metadata != nil && current.Metadata.Attributes != nil {
			node["attributes"] = current.Metadata.Attributes
		}
		parts = append([]map[string]interface{}{node}, parts...)
		current = current.Parent
	}
	return parts
}

// ParseFormatFlag parses the format flag and returns the corresponding OutputFormat
func ParseFormatFlag(flagValue string) (OutputFormat, error) {
	switch strings.ToLower(flagValue) {
	case "text":
		return OutputFormatText, nil
	case "fields":
		return OutputFormatFields, nil
	case "json":
		return OutputFormatJSON, nil
	case "jsonl":
		return OutputFormatJSONL, nil
	default:
		return OutputFormatText, fmt.Errorf("invalid format: %s (valid options: text, fields, json, jsonl)", flagValue)
	}
}

// ParseFieldsFlag parses the --fields flag into a list of field names
func ParseFieldsFlag(flagValue string) []string {
	if flagValue == "" {
		return nil
	}
	var fields []string
	for _, field := range strings.Split(flagValue, ",") {
		field = strings.TrimSpace(field)
		if field != "" {
			fields = append(fields, field)
		}
	}
	return fields
}
