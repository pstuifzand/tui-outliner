package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/storage"
)

// ItemData represents parsed diff format data for a single item
type ItemData struct {
	ID         string
	Text       string
	ParentID   string
	Position   int
	Tags       []string
	Attributes map[string]string
	Created    string
	Modified   string
}

// DiffResult contains the analysis of changes
type DiffResult struct {
	NewItems      map[string]*ItemData
	DeletedItems  map[string]*ItemData
	ModifiedItems map[string]*ItemChange
}

// ItemChange describes what changed for an item
type ItemChange struct {
	Item           *ItemData
	OldItem        *ItemData
	TextChanged    bool
	OldText        string
	StructureChanged bool
	OldParentID    string
	OldPosition    int
	TagsAdded      []string
	TagsRemoved    []string
	AttrsAdded     map[string]string
	AttrsRemoved   map[string]string
	AttrsChanged   map[string][2]string // attrName -> [oldValue, newValue]
	ModifiedChanged bool
	OldModified    string
}

func main() {
	verbose := flag.Bool("v", false, "Verbose output (show full details)")
	summary := flag.Bool("s", false, "Summary only (no item-level details)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: outline-diff [options] <file.json>
       outline-diff [options] <file1.json> <file2.json>

Compares TUI Outliner JSON files and shows meaningful changes by item.

Single-file mode: Shows changes across all backups of that file
Two-file mode: Shows changes between two specific files

Options:
  -v   Verbose output (show full details)
  -s   Summary only (counts without item-level details)

Examples:
  # Compare a file against its backup history
  outline-diff my_outline.json

  # Show changes between two specific files
  outline-diff backup1.json backup2.json

  # Just show summary of changes
  outline-diff -s backup1.json backup2.json

Output shows:
  - New items added to the outline
  - Deleted items removed from the outline
  - Modified items with specific changes (text, attributes, tags, structure)
  - Changes clearly labeled with item IDs for reference
`)
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	if len(args) == 1 {
		// Single-file mode: show history across backups
		handleSingleFileMode(args[0], *verbose, *summary)
	} else {
		// Two-file mode: compare two specific files
		handleTwoFileMode(args[0], args[1], *verbose, *summary)
	}
}

// handleTwoFileMode compares two specific files
func handleTwoFileMode(file1Path, file2Path string, verbose, summary bool) {
	// Open and parse first file
	file1, err := os.Open(file1Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening first file: %v\n", err)
		os.Exit(1)
	}
	defer file1.Close()

	var outline1 model.Outline
	if err := json.NewDecoder(file1).Decode(&outline1); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing first JSON: %v\n", err)
		os.Exit(1)
	}

	// Open and parse second file
	file2, err := os.Open(file2Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening second file: %v\n", err)
		os.Exit(1)
	}
	defer file2.Close()

	var outline2 model.Outline
	if err := json.NewDecoder(file2).Decode(&outline2); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing second JSON: %v\n", err)
		os.Exit(1)
	}

	// Convert to diff format for parsing
	var buf1, buf2 bytes.Buffer
	if err := storage.EncodeDiffFormat(&outline1, &buf1); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding first outline: %v\n", err)
		os.Exit(1)
	}

	if err := storage.EncodeDiffFormat(&outline2, &buf2); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding second outline: %v\n", err)
		os.Exit(1)
	}

	// Parse diff format into structured data
	data1 := parseDiffFormat(buf1.String())
	data2 := parseDiffFormat(buf2.String())

	// Analyze changes
	result := analyzeChanges(data1, data2)

	// Output header
	fmt.Printf("=== Outline Diff: %s → %s ===\n\n", file1Path, file2Path)

	// Output changes
	if !summary {
		if len(result.NewItems) > 0 {
			printNewItems(result.NewItems, verbose)
		}

		if len(result.DeletedItems) > 0 {
			printDeletedItems(result.DeletedItems, verbose)
		}

		if len(result.ModifiedItems) > 0 {
			printModifiedItems(result.ModifiedItems, verbose)
		}

		if len(result.NewItems) == 0 && len(result.DeletedItems) == 0 && len(result.ModifiedItems) == 0 {
			fmt.Println("No changes detected")
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Printf("  %d items modified\n", len(result.ModifiedItems))
	fmt.Printf("  %d items added\n", len(result.NewItems))
	fmt.Printf("  %d items deleted\n", len(result.DeletedItems))
}

// handleSingleFileMode finds backups for a file and shows the diff history
func handleSingleFileMode(filePath string, verbose, summaryOnly bool) {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	// Create backup manager and find backups
	bm, err := storage.NewBackupManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing backup manager: %v\n", err)
		os.Exit(1)
	}

	backups, err := bm.FindBackupsForFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching for backups: %v\n", err)
		os.Exit(1)
	}

	if len(backups) == 0 {
		fmt.Fprintf(os.Stderr, "No backups found for %s\n", absPath)
		fmt.Fprintf(os.Stderr, "Note: Backups are stored in ~/.local/share/tui-outliner/backups/\n")
		fmt.Fprintf(os.Stderr, "Make sure the file has been edited and auto-saved to create backups.\n")
		os.Exit(1)
	}

	if len(backups) < 2 {
		fmt.Fprintf(os.Stderr, "Only found %d backup, need at least 2 to compare\n", len(backups))
		os.Exit(1)
	}

	fmt.Printf("=== Backup History for: %s ===\n", filePath)
	fmt.Printf("Found %d backups\n\n", len(backups))

	// Compare each consecutive pair
	for i := 0; i < len(backups)-1; i++ {
		backup1 := backups[i]
		backup2 := backups[i+1]

		// Load outlines from backups
		var outline1, outline2 model.Outline

		data1, err := os.ReadFile(backup1.FilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading backup 1: %v\n", err)
			continue
		}
		if err := json.Unmarshal(data1, &outline1); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing backup 1: %v\n", err)
			continue
		}

		data2, err := os.ReadFile(backup2.FilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading backup 2: %v\n", err)
			continue
		}
		if err := json.Unmarshal(data2, &outline2); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing backup 2: %v\n", err)
			continue
		}

		// Convert to diff format
		var buf1, buf2 bytes.Buffer
		if err := storage.EncodeDiffFormat(&outline1, &buf1); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding backup 1: %v\n", err)
			continue
		}
		if err := storage.EncodeDiffFormat(&outline2, &buf2); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding backup 2: %v\n", err)
			continue
		}

		// Parse and analyze
		data1Map := parseDiffFormat(buf1.String())
		data2Map := parseDiffFormat(buf2.String())
		result := analyzeChanges(data1Map, data2Map)

		// Print diff header
		fmt.Printf("--- %s (backup %d)\n", formatBackupTime(backup1.Timestamp), i+1)
		fmt.Printf("+++ %s (backup %d)\n", formatBackupTime(backup2.Timestamp), i+2)
		fmt.Println()

		// Output changes
		if !summaryOnly {
			if len(result.NewItems) > 0 {
				printNewItems(result.NewItems, verbose)
			}

			if len(result.DeletedItems) > 0 {
				printDeletedItems(result.DeletedItems, verbose)
			}

			if len(result.ModifiedItems) > 0 {
				printModifiedItems(result.ModifiedItems, verbose)
			}

			if len(result.NewItems) == 0 && len(result.DeletedItems) == 0 && len(result.ModifiedItems) == 0 {
				fmt.Println("No changes between these backups")
			}
		}

		// Summary
		fmt.Println()
		fmt.Printf("  %d modified, %d added, %d deleted\n",
			len(result.ModifiedItems), len(result.NewItems), len(result.DeletedItems))
		fmt.Println()
	}
}

// formatBackupTime formats a backup timestamp for display
func formatBackupTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
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

// printNewItems prints newly added items
func printNewItems(items map[string]*ItemData, verbose bool) {
	if len(items) == 0 {
		return
	}

	fmt.Println("New Items:")
	fmt.Println()

	// Sort by ID for consistent output
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		item := items[id]
		fmt.Printf("  %s: %s\n", id, truncateText(item.Text, 60))

		if item.ParentID != "" {
			fmt.Printf("    PARENT: %s at position %d\n", item.ParentID, item.Position)
		} else {
			fmt.Printf("    POSITION: root position %d\n", item.Position)
		}

		if len(item.Tags) > 0 {
			fmt.Printf("    TAGS: %s\n", strings.Join(item.Tags, ", "))
		}

		if len(item.Attributes) > 0 {
			for key, val := range item.Attributes {
				fmt.Printf("    ATTR: %s = %s\n", key, val)
			}
		}
		fmt.Println()
	}
}

// printDeletedItems prints deleted items
func printDeletedItems(items map[string]*ItemData, verbose bool) {
	if len(items) == 0 {
		return
	}

	fmt.Println("Deleted Items:")
	fmt.Println()

	// Sort by ID for consistent output
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		item := items[id]
		fmt.Printf("  %s: %s\n", id, truncateText(item.Text, 60))
		fmt.Println()
	}
}

// printModifiedItems prints items with changes
func printModifiedItems(items map[string]*ItemChange, verbose bool) {
	if len(items) == 0 {
		return
	}

	fmt.Println("Modified Items:")
	fmt.Println()

	// Sort by ID for consistent output
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		change := items[id]
		fmt.Printf("  %s: %s\n", id, truncateText(change.Item.Text, 60))

		if change.TextChanged {
			fmt.Printf("    TEXT: %s → %s\n",
				truncateText(change.OldText, 40),
				truncateText(change.Item.Text, 40))
		}

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
				fmt.Printf("    MOVED: from parent %s to parent %s\n", oldParent, newParent)
			}
			if change.OldPosition != change.Item.Position {
				fmt.Printf("    POSITION: %d → %d\n", change.OldPosition, change.Item.Position)
			}
		}

		if len(change.TagsAdded) > 0 {
			fmt.Printf("    TAGS added: %s\n", strings.Join(change.TagsAdded, ", "))
		}
		if len(change.TagsRemoved) > 0 {
			fmt.Printf("    TAGS removed: %s\n", strings.Join(change.TagsRemoved, ", "))
		}

		if len(change.AttrsAdded) > 0 {
			for key := range change.AttrsAdded {
				fmt.Printf("    ATTR added: %s = %s\n", key, change.AttrsAdded[key])
			}
		}

		if len(change.AttrsChanged) > 0 {
			for key := range change.AttrsChanged {
				old, new := change.AttrsChanged[key][0], change.AttrsChanged[key][1]
				fmt.Printf("    ATTR changed: %s: %s → %s\n", key, old, new)
			}
		}

		if len(change.AttrsRemoved) > 0 {
			for key := range change.AttrsRemoved {
				fmt.Printf("    ATTR removed: %s (was: %s)\n", key, change.AttrsRemoved[key])
			}
		}

		if change.ModifiedChanged {
			fmt.Printf("    MODIFIED: %s → %s\n", formatTime(change.OldModified), formatTime(change.Item.Modified))
		}

		fmt.Println()
	}
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
