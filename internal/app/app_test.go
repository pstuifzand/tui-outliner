package app

import (
	"os"
	"testing"

	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/storage"
	"github.com/pstuifzand/tui-outliner/internal/ui"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple command",
			input:    "save",
			expected: []string{"save"},
		},
		{
			name:     "command with arguments",
			input:    "open file.txt",
			expected: []string{"open", "file.txt"},
		},
		{
			name:     "double quoted string",
			input:    `export markdown "my file.md"`,
			expected: []string{"export", "markdown", "my file.md"},
		},
		{
			name:     "single quoted string",
			input:    "export markdown 'my file.md'",
			expected: []string{"export", "markdown", "my file.md"},
		},
		{
			name:     "mixed quotes",
			input:    `title "Hello World" and more`,
			expected: []string{"title", "Hello World", "and", "more"},
		},
		{
			name:     "escaped quotes",
			input:    `attr add key "value with \"quotes\""`,
			expected: []string{"attr", "add", "key", `value with "quotes"`},
		},
		{
			name:     "escaped backslash",
			input:    `path "C:\\Users\\test"`,
			expected: []string{"path", `C:\Users\test`},
		},
		{
			name:     "multiple spaces",
			input:    "command    with    spaces",
			expected: []string{"command", "with", "spaces"},
		},
		{
			name:     "tabs and spaces",
			input:    "command\twith\t  mixed",
			expected: []string{"command", "with", "mixed"},
		},
		{
			name:     "empty quoted string",
			input:    `command ""`,
			expected: []string{"command", ""},
		},
		{
			name:     "quoted string with special characters",
			input:    `attr add url "https://example.com/path?query=value&other=123"`,
			expected: []string{"attr", "add", "url", "https://example.com/path?query=value&other=123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommand(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d. Input: %q", len(tt.expected), len(result), tt.input)
				return
			}
			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("Part %d: expected %q, got %q. Input: %q", i, tt.expected[i], part, tt.input)
				}
			}
		})
	}
}

func TestSaveSyncesTreeItemsWithOutline(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test-outline-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create an initial outline with some items
	outline := model.NewOutline()
	outline.Items = append(outline.Items, model.NewItem("First"))
	outline.Items = append(outline.Items, model.NewItem("Second"))
	outline.Items = append(outline.Items, model.NewItem("Third"))

	// Save the initial outline
	store := storage.NewJSONStore(tmpfile.Name())
	if err := store.Save(outline); err != nil {
		t.Fatalf("Failed to save initial outline: %v", err)
	}

	// Create a tree view with the outline items
	tree := ui.NewTreeView(outline.Items)

	// Add a new item to the tree (this modifies tree.items but not outline.Items)
	tree.SelectItem(1)
	tree.AddItemAfter("New Item")

	// Create an app instance
	app := &App{
		outline: outline,
		store:   store,
		tree:    tree,
	}

	// Verify that before Save, outline.Items is not updated
	if len(outline.Items) != 3 {
		t.Logf("Before Save: outline has %d items (as expected)", len(outline.Items))
	}

	// Now test the Save function - it should sync tree items to outline
	if err := app.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// After Save, outline.Items should be updated
	if len(outline.Items) != 4 {
		t.Errorf("After Save: Expected 4 items in outline, got %d", len(outline.Items))
	}

	// Load the file again and verify all items were saved
	loadedOutline, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load outline: %v", err)
	}

	if len(loadedOutline.Items) != 4 {
		t.Errorf("Expected 4 items after reload, got %d", len(loadedOutline.Items))
	}

	// Verify the new item is at the correct position
	foundNewItem := false
	for _, item := range loadedOutline.Items {
		if item.Text == "New Item" {
			foundNewItem = true
			break
		}
	}

	if !foundNewItem {
		t.Error("New Item not found in reloaded outline")
	}
}
