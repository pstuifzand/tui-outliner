package storage

import (
	"bytes"
	"testing"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

func TestEncodeDiffFormatBasic(t *testing.T) {
	// Create a simple outline
	outline := &model.Outline{
		Items: []*model.Item{
			{
				ID:   "item1",
				Text: "First item",
				Metadata: &model.Metadata{
					Tags:       []string{"important"},
					Attributes: map[string]string{"status": "done"},
					Created:    time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
					Modified:   time.Date(2025, 11, 3, 15, 0, 0, 0, time.UTC),
				},
				Children: []*model.Item{
					{
						ID:   "item2",
						Text: "Nested item",
						Metadata: &model.Metadata{
							Tags:       []string{},
							Attributes: map[string]string{},
							Created:    time.Date(2025, 11, 2, 10, 0, 0, 0, time.UTC),
							Modified:   time.Date(2025, 11, 2, 10, 0, 0, 0, time.UTC),
						},
						Children: []*model.Item{},
					},
				},
			},
		},
	}

	// Set parent pointers
	outline.Items[0].Children[0].Parent = outline.Items[0]

	// Encode
	var buf bytes.Buffer
	err := EncodeDiffFormat(outline, &buf)
	if err != nil {
		t.Fatalf("EncodeDiffFormat failed: %v", err)
	}

	output := buf.String()

	// Verify sections exist
	if !bytes.Contains(buf.Bytes(), []byte("[TEXT SECTION]")) {
		t.Error("Missing [TEXT SECTION]")
	}
	if !bytes.Contains(buf.Bytes(), []byte("[STRUCTURE SECTION]")) {
		t.Error("Missing [STRUCTURE SECTION]")
	}
	if !bytes.Contains(buf.Bytes(), []byte("[TAGS SECTION]")) {
		t.Error("Missing [TAGS SECTION]")
	}
	if !bytes.Contains(buf.Bytes(), []byte("[ATTRIBUTES SECTION]")) {
		t.Error("Missing [ATTRIBUTES SECTION]")
	}
	if !bytes.Contains(buf.Bytes(), []byte("[TIMESTAMPS SECTION]")) {
		t.Error("Missing [TIMESTAMPS SECTION]")
	}

	t.Logf("Encoded output:\n%s", output)
}

func TestTextEscaping(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		encoded string
	}{
		{
			name:    "Simple text",
			input:   "Hello world",
			encoded: "Hello world",
		},
		{
			name:    "Text with newline",
			input:   "Hello\nworld",
			encoded: "Hello\\nworld",
		},
		{
			name:    "Text with backslash",
			input:   "C:\\path\\to\\file",
			encoded: "C:\\\\path\\\\to\\\\file",
		},
		{
			name:    "Text with newline and backslash",
			input:   "Line1\nPath: C:\\test",
			encoded: "Line1\\nPath: C:\\\\test",
		},
		{
			name:    "Multiple newlines",
			input:   "Line1\nLine2\nLine3",
			encoded: "Line1\\nLine2\\nLine3",
		},
		{
			name:    "Empty string",
			input:   "",
			encoded: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encodeTextValue(tt.input)
			if encoded != tt.encoded {
				t.Errorf("encodeTextValue(%q) = %q, want %q", tt.input, encoded, tt.encoded)
			}

			decoded := decodeTextValue(encoded)
			if decoded != tt.input {
				t.Errorf("decodeTextValue(%q) = %q, want %q", encoded, decoded, tt.input)
			}
		})
	}
}

func TestRoundTripEncoding(t *testing.T) {
	// Create a more complex outline
	item1 := &model.Item{
		ID:   "id-001",
		Text: "First task",
		Metadata: &model.Metadata{
			Tags:       []string{"work", "urgent"},
			Attributes: map[string]string{"priority": "high", "status": "in-progress"},
			Created:    time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC),
			Modified:   time.Date(2025, 11, 3, 15, 30, 0, 0, time.UTC),
		},
		Children: []*model.Item{},
	}

	item2 := &model.Item{
		ID:   "id-002",
		Text: "Subtask with\nnewlines",
		Metadata: &model.Metadata{
			Tags:       []string{},
			Attributes: map[string]string{},
			Created:    time.Date(2025, 11, 2, 9, 0, 0, 0, time.UTC),
			Modified:   time.Date(2025, 11, 3, 16, 0, 0, 0, time.UTC),
		},
		Children: []*model.Item{},
	}
	item2.Parent = item1

	item3 := &model.Item{
		ID:   "id-003",
		Text: "Path: C:\\Users\\Documents",
		Metadata: &model.Metadata{
			Tags:       []string{"reference"},
			Attributes: map[string]string{"type": "path"},
			Created:    time.Date(2025, 11, 3, 8, 0, 0, 0, time.UTC),
			Modified:   time.Date(2025, 11, 3, 8, 0, 0, 0, time.UTC),
		},
		Children: []*model.Item{},
	}

	item1.Children = []*model.Item{item2}

	outline := &model.Outline{
		Items: []*model.Item{item1, item3},
	}

	// Encode
	var buf bytes.Buffer
	err := EncodeDiffFormat(outline, &buf)
	if err != nil {
		t.Fatalf("EncodeDiffFormat failed: %v", err)
	}

	t.Logf("Encoded output:\n%s", buf.String())

	// Decode
	decoded, err := DecodeDiffFormat(&buf)
	if err != nil {
		t.Fatalf("DecodeDiffFormat failed: %v", err)
	}

	// Verify structure
	if len(decoded.Items) != 2 {
		t.Errorf("Expected 2 root items, got %d", len(decoded.Items))
		for i, item := range decoded.Items {
			t.Logf("Root item %d: ID=%s, Text=%s", i, item.ID, item.Text)
		}
	}

	// Find items by ID for comparison (order might differ)
	itemMap := make(map[string]*model.Item)
	var collectItems func([]*model.Item)
	collectItems = func(items []*model.Item) {
		for _, item := range items {
			itemMap[item.ID] = item
			if len(item.Children) > 0 {
				collectItems(item.Children)
			}
		}
	}
	collectItems(decoded.Items)

	// Verify item1
	if item1Data, ok := itemMap["id-001"]; ok {
		if item1Data.Text != "First task" {
			t.Errorf("item1 text mismatch: got %q, want %q", item1Data.Text, "First task")
		}
		if len(item1Data.Metadata.Tags) != 2 {
			t.Errorf("item1 tags count: got %d, want 2", len(item1Data.Metadata.Tags))
		}
		if len(item1Data.Children) != 1 {
			t.Errorf("item1 children count: got %d, want 1", len(item1Data.Children))
		}
	} else {
		t.Error("item1 (id-001) not found in decoded outline")
	}

	// Verify item2
	if item2Data, ok := itemMap["id-002"]; ok {
		if item2Data.Text != "Subtask with\nnewlines" {
			t.Errorf("item2 text mismatch: got %q, want %q", item2Data.Text, "Subtask with\nnewlines")
		}
		if item2Data.Parent == nil {
			t.Error("item2 parent is nil")
		} else if item2Data.Parent.ID != "id-001" {
			t.Errorf("item2 parent ID: got %q, want %q", item2Data.Parent.ID, "id-001")
		}
	} else {
		t.Error("item2 (id-002) not found in decoded outline")
	}

	// Verify item3
	if item3Data, ok := itemMap["id-003"]; ok {
		if item3Data.Text != "Path: C:\\Users\\Documents" {
			t.Errorf("item3 text mismatch: got %q, want %q", item3Data.Text, "Path: C:\\Users\\Documents")
		}
		if item3Data.Metadata.Attributes["type"] != "path" {
			t.Errorf("item3 attribute 'type': got %q, want %q", item3Data.Metadata.Attributes["type"], "path")
		}
	} else {
		t.Error("item3 (id-003) not found in decoded outline")
	}
}

func TestParsingEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		encoded  string
		expected string
	}{
		{
			name:     "Only backslash",
			encoded:  "\\\\",
			expected: "\\",
		},
		{
			name:     "Only newline",
			encoded:  "\\n",
			expected: "\n",
		},
		{
			name:     "Backslash followed by literal n",
			encoded:  "\\\\n",
			expected: "\\n",
		},
		{
			name:     "Escaped backslash at end",
			encoded:  "test\\\\",
			expected: "test\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeTextValue(tt.encoded)
			if result != tt.expected {
				t.Errorf("decodeTextValue(%q) = %q, want %q", tt.encoded, result, tt.expected)
			}
		})
	}
}

func TestEmptyOutline(t *testing.T) {
	outline := &model.Outline{
		Items: []*model.Item{},
	}

	var buf bytes.Buffer
	err := EncodeDiffFormat(outline, &buf)
	if err != nil {
		t.Fatalf("EncodeDiffFormat failed: %v", err)
	}

	decoded, err := DecodeDiffFormat(&buf)
	if err != nil {
		t.Fatalf("DecodeDiffFormat failed: %v", err)
	}

	if len(decoded.Items) != 0 {
		t.Errorf("Expected empty outline, got %d items", len(decoded.Items))
	}
}

func TestComplexNesting(t *testing.T) {
	// Create a deeply nested structure
	root := &model.Item{
		ID:   "root",
		Text: "Root",
		Metadata: &model.Metadata{
			Tags:       []string{},
			Attributes: map[string]string{},
			Created:    time.Now(),
			Modified:   time.Now(),
		},
		Children: []*model.Item{},
	}

	level1 := &model.Item{
		ID:   "level1",
		Text: "Level 1",
		Metadata: &model.Metadata{
			Tags:       []string{},
			Attributes: map[string]string{},
			Created:    time.Now(),
			Modified:   time.Now(),
		},
		Children: []*model.Item{},
		Parent:   root,
	}

	level2 := &model.Item{
		ID:   "level2",
		Text: "Level 2",
		Metadata: &model.Metadata{
			Tags:       []string{},
			Attributes: map[string]string{},
			Created:    time.Now(),
			Modified:   time.Now(),
		},
		Children: []*model.Item{},
		Parent:   level1,
	}

	level1.Children = []*model.Item{level2}
	root.Children = []*model.Item{level1}

	outline := &model.Outline{
		Items: []*model.Item{root},
	}

	// Encode and decode
	var buf bytes.Buffer
	if err := EncodeDiffFormat(outline, &buf); err != nil {
		t.Fatalf("EncodeDiffFormat failed: %v", err)
	}

	decoded, err := DecodeDiffFormat(&buf)
	if err != nil {
		t.Fatalf("DecodeDiffFormat failed: %v", err)
	}

	// Verify nesting
	if len(decoded.Items) != 1 || decoded.Items[0].ID != "root" {
		t.Error("Root item not preserved")
	}

	root2 := decoded.Items[0]
	if len(root2.Children) != 1 || root2.Children[0].ID != "level1" {
		t.Error("Level 1 nesting not preserved")
	}

	level1_2 := root2.Children[0]
	if len(level1_2.Children) != 1 || level1_2.Children[0].ID != "level2" {
		t.Error("Level 2 nesting not preserved")
	}
}
