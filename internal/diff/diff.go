package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/storage"
)

// ComputeDiff compares two outlines and returns a DiffResult
// This is the main entry point for computing differences
func ComputeDiff(outline1, outline2 *model.Outline) (*DiffResult, error) {
	// Convert both outlines to diff format for consistent parsing
	buf1, err := serializeOutlineToDiffFormat(outline1)
	if err != nil {
		return nil, fmt.Errorf("failed to encode first outline: %w", err)
	}

	buf2, err := serializeOutlineToDiffFormat(outline2)
	if err != nil {
		return nil, fmt.Errorf("failed to encode second outline: %w", err)
	}

	// Parse both into structured data
	data1 := parseDiffFormat(buf1)
	data2 := parseDiffFormat(buf2)

	// Analyze differences
	return analyzeChanges(data1, data2), nil
}

// serializeOutlineToDiffFormat converts an outline to diff format string
func serializeOutlineToDiffFormat(outline *model.Outline) (string, error) {
	var buf strings.Builder
	if err := storage.EncodeDiffFormat(outline, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// parseDiffFormat parses the diff format into structured data by ID
func parseDiffFormat(content string) map[string]*ItemData {
	items := make(map[string]*ItemData)
	lines := strings.Split(content, "\n")

	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line
			continue
		}

		// Parse line: id: value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		id := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if _, exists := items[id]; !exists {
			items[id] = &ItemData{
				ID:         id,
				Attributes: make(map[string]string),
			}
		}

		switch currentSection {
		case "[TEXT SECTION]":
			items[id].Text = decodeTextValue(value)

		case "[STRUCTURE SECTION]":
			// Format: parent_id:position
			structParts := strings.SplitN(value, ":", 2)
			if len(structParts) == 2 {
				items[id].ParentID = structParts[0]
				fmt.Sscanf(structParts[1], "%d", &items[id].Position)
			}

		case "[TAGS SECTION]":
			if value != "" {
				items[id].Tags = strings.Split(value, ",")
				for i := range items[id].Tags {
					items[id].Tags[i] = strings.TrimSpace(items[id].Tags[i])
				}
			}

		case "[ATTRIBUTES SECTION]":
			parseAttributes(value, items[id].Attributes)

		case "[TIMESTAMPS SECTION]":
			// Format: created modified
			parts := strings.SplitN(value, " ", 2)
			if len(parts) == 2 {
				items[id].Created = strings.TrimSpace(parts[0])
				items[id].Modified = strings.TrimSpace(parts[1])
			}
		}
	}

	return items
}

// decodeTextValue decodes escaped text values
func decodeTextValue(text string) string {
	var result strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] == '\\' && i+1 < len(text) {
			next := text[i+1]
			if next == 'n' {
				result.WriteRune('\n')
				i++
			} else if next == '\\' {
				result.WriteRune('\\')
				i++
			} else {
				result.WriteByte('\\')
			}
		} else {
			result.WriteByte(text[i])
		}
	}
	return result.String()
}

// parseAttributes parses "key1=value1,key2=value2" format
func parseAttributes(value string, attrs map[string]string) {
	if value == "" {
		return
	}
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			attrs[key] = val
		}
	}
}

// analyzeChanges compares two sets of item data
func analyzeChanges(data1, data2 map[string]*ItemData) *DiffResult {
	result := &DiffResult{
		NewItems:      make(map[string]*ItemData),
		DeletedItems:  make(map[string]*ItemData),
		ModifiedItems: make(map[string]*ItemChange),
	}

	// Find new and modified items
	for id, item2 := range data2 {
		if item1, exists := data1[id]; !exists {
			result.NewItems[id] = item2
		} else {
			change := compareItems(item1, item2)
			if change != nil {
				result.ModifiedItems[id] = change
			}
		}
	}

	// Find deleted items
	for id, item1 := range data1 {
		if _, exists := data2[id]; !exists {
			result.DeletedItems[id] = item1
		}
	}

	return result
}

// compareItems checks if an item changed and returns the changes
func compareItems(old, new *ItemData) *ItemChange {
	change := &ItemChange{
		Item:           new,
		OldItem:        old,
		AttrsAdded:     make(map[string]string),
		AttrsRemoved:   make(map[string]string),
		AttrsChanged:   make(map[string][2]string),
	}

	hasChange := false

	// Check text
	if old.Text != new.Text {
		change.TextChanged = true
		change.OldText = old.Text
		hasChange = true
	}

	// Check structure
	if old.ParentID != new.ParentID || old.Position != new.Position {
		change.StructureChanged = true
		change.OldParentID = old.ParentID
		change.OldPosition = old.Position
		hasChange = true
	}

	// Check tags
	oldTagSet := make(map[string]bool)
	for _, tag := range old.Tags {
		oldTagSet[tag] = true
	}
	newTagSet := make(map[string]bool)
	for _, tag := range new.Tags {
		newTagSet[tag] = true
	}

	for tag := range newTagSet {
		if !oldTagSet[tag] {
			change.TagsAdded = append(change.TagsAdded, tag)
			hasChange = true
		}
	}
	sort.Strings(change.TagsAdded)

	for tag := range oldTagSet {
		if !newTagSet[tag] {
			change.TagsRemoved = append(change.TagsRemoved, tag)
			hasChange = true
		}
	}
	sort.Strings(change.TagsRemoved)

	// Check attributes
	for key, newVal := range new.Attributes {
		if oldVal, exists := old.Attributes[key]; !exists {
			change.AttrsAdded[key] = newVal
			hasChange = true
		} else if oldVal != newVal {
			change.AttrsChanged[key] = [2]string{oldVal, newVal}
			hasChange = true
		}
	}

	for key := range old.Attributes {
		if _, exists := new.Attributes[key]; !exists {
			change.AttrsRemoved[key] = old.Attributes[key]
			hasChange = true
		}
	}

	// Check modified timestamp
	if old.Modified != new.Modified {
		change.ModifiedChanged = true
		change.OldModified = old.Modified
		hasChange = true
	}

	if !hasChange {
		return nil
	}

	return change
}
