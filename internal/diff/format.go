package diff

import (
	"fmt"
	"sort"
	"strings"
)

// BuildDiffLines converts a DiffResult into formatted display lines
// This is suitable for both CLI and TUI output
func BuildDiffLines(result *DiffResult, verbose bool) []DiffLine {
	var lines []DiffLine

	// New items section
	if len(result.NewItems) > 0 {
		lines = append(lines, DiffLine{Type: DiffTypeNewSection, Content: "New Items:"})
		lines = append(lines, DiffLine{Type: DiffTypeBlank})

		sortedIDs := getSortedIDs(result.NewItems)
		for _, id := range sortedIDs {
			item := result.NewItems[id]
			lines = append(lines, formatNewItem(item)...)
		}
	}

	// Deleted items section
	if len(result.DeletedItems) > 0 {
		lines = append(lines, DiffLine{Type: DiffTypeDeletedSection, Content: "Deleted Items:"})
		lines = append(lines, DiffLine{Type: DiffTypeBlank})

		sortedIDs := getSortedIDs(result.DeletedItems)
		for _, id := range sortedIDs {
			item := result.DeletedItems[id]
			lines = append(lines, formatDeletedItem(item)...)
		}
	}

	// Modified items section
	if len(result.ModifiedItems) > 0 {
		lines = append(lines, DiffLine{Type: DiffTypeModifiedSection, Content: "Modified Items:"})
		lines = append(lines, DiffLine{Type: DiffTypeBlank})

		sortedIDs := getSortedIDs(result.ModifiedItems)
		for _, id := range sortedIDs {
			change := result.ModifiedItems[id]
			lines = append(lines, formatModifiedItem(change, verbose)...)
		}
	}

	// Summary section
	if len(result.NewItems) > 0 || len(result.DeletedItems) > 0 || len(result.ModifiedItems) > 0 {
		lines = append(lines, DiffLine{Type: DiffTypeBlank})
		lines = append(lines, DiffLine{Type: DiffTypeSummary, Content: "=== Summary ==="})
		lines = append(lines, DiffLine{
			Type: DiffTypeSummary,
			Content: fmt.Sprintf("  %d modified, %d added, %d deleted",
				len(result.ModifiedItems), len(result.NewItems), len(result.DeletedItems)),
		})
	}

	return lines
}

// formatNewItem creates display lines for a newly added item
func formatNewItem(item *ItemData) []DiffLine {
	var lines []DiffLine

	// Item header
	lines = append(lines, DiffLine{
		Type:    DiffTypeNewItem,
		Content: fmt.Sprintf("%s: %s", item.ID, truncateText(item.Text, 60)),
		Indent:  1,
	})

	// Parent/position
	if item.ParentID != "" {
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("PARENT: %s at position %d", item.ParentID, item.Position),
			Indent:  2,
		})
	} else {
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("POSITION: root position %d", item.Position),
			Indent:  2,
		})
	}

	// Tags
	if len(item.Tags) > 0 {
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("TAGS: %s", strings.Join(item.Tags, ", ")),
			Indent:  2,
		})
	}

	// Attributes
	for key, val := range item.Attributes {
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("ATTR: %s = %s", key, val),
			Indent:  2,
		})
	}

	lines = append(lines, DiffLine{Type: DiffTypeBlank})
	return lines
}

// formatDeletedItem creates display lines for a deleted item
func formatDeletedItem(item *ItemData) []DiffLine {
	var lines []DiffLine

	lines = append(lines, DiffLine{
		Type:    DiffTypeDeletedItem,
		Content: fmt.Sprintf("%s: %s", item.ID, truncateText(item.Text, 60)),
		Indent:  1,
	})

	lines = append(lines, DiffLine{Type: DiffTypeBlank})
	return lines
}

// formatModifiedItem creates display lines for a modified item
func formatModifiedItem(change *ItemChange, verbose bool) []DiffLine {
	var lines []DiffLine

	// Item header
	lines = append(lines, DiffLine{
		Type:    DiffTypeModifiedItem,
		Content: fmt.Sprintf("%s: %s", change.Item.ID, truncateText(change.Item.Text, 60)),
		Indent:  1,
	})

	// Text change
	if change.TextChanged {
		lines = append(lines, DiffLine{
			Type: DiffTypeItemDetail,
			Content: fmt.Sprintf("TEXT: %s → %s",
				truncateText(change.OldText, 40),
				truncateText(change.Item.Text, 40)),
			Indent: 2,
		})
	}

	// Structure changes
	if change.StructureChanged {
		oldParent := change.OldParentID
		if oldParent == "" {
			oldParent = "root"
		}
		newParent := change.Item.ParentID
		if newParent == "" {
			newParent = "root"
		}

		if oldParent != newParent {
			lines = append(lines, DiffLine{
				Type:    DiffTypeItemDetail,
				Content: fmt.Sprintf("MOVED: from parent %s to parent %s", oldParent, newParent),
				Indent:  2,
			})
		}

		if change.OldPosition != change.Item.Position {
			lines = append(lines, DiffLine{
				Type:    DiffTypeItemDetail,
				Content: fmt.Sprintf("POSITION: %d → %d", change.OldPosition, change.Item.Position),
				Indent:  2,
			})
		}
	}

	// Tag changes
	if len(change.TagsAdded) > 0 {
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("TAGS added: %s", strings.Join(change.TagsAdded, ", ")),
			Indent:  2,
		})
	}
	if len(change.TagsRemoved) > 0 {
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("TAGS removed: %s", strings.Join(change.TagsRemoved, ", ")),
			Indent:  2,
		})
	}

	// Attribute changes
	for key := range change.AttrsAdded {
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("ATTR added: %s = %s", key, change.AttrsAdded[key]),
			Indent:  2,
		})
	}

	for key := range change.AttrsChanged {
		old, new := change.AttrsChanged[key][0], change.AttrsChanged[key][1]
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("ATTR changed: %s: %s → %s", key, old, new),
			Indent:  2,
		})
	}

	for key := range change.AttrsRemoved {
		lines = append(lines, DiffLine{
			Type:    DiffTypeItemDetail,
			Content: fmt.Sprintf("ATTR removed: %s (was: %s)", key, change.AttrsRemoved[key]),
			Indent:  2,
		})
	}

	// Modified timestamp
	if change.ModifiedChanged {
		lines = append(lines, DiffLine{
			Type: DiffTypeItemDetail,
			Content: fmt.Sprintf("MODIFIED: %s → %s",
				formatTime(change.OldModified),
				formatTime(change.Item.Modified)),
			Indent: 2,
		})
	}

	lines = append(lines, DiffLine{Type: DiffTypeBlank})
	return lines
}

// truncateText limits text length for display
func truncateText(text string, maxLen int) string {
	// Handle multi-line text
	lines := strings.Split(text, "\n")
	text = lines[0]
	if len(lines) > 1 {
		text += " ..."
	}

	if len(text) > maxLen {
		return text[:maxLen] + "..."
	}
	return text
}

// formatTime formats an ISO timestamp for display
func formatTime(ts string) string {
	if len(ts) > 19 {
		return ts[:19] // YYYY-MM-DDTHH:MM:SS
	}
	return ts
}

// getSortedIDs returns a sorted slice of keys from a map
func getSortedIDs[T any](items map[string]T) []string {
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
