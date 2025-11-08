package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

func TestParseFormatFlag(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  OutputFormat
		shouldErr bool
	}{
		{"text format", "text", OutputFormatText, false},
		{"fields format", "fields", OutputFormatFields, false},
		{"json format", "json", OutputFormatJSON, false},
		{"jsonl format", "jsonl", OutputFormatJSONL, false},
		{"invalid format", "xml", OutputFormatText, true},
		{"case insensitive", "TEXT", OutputFormatText, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFormatFlag(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestParseFieldsFlag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single field", "id", []string{"id"}},
		{"multiple fields", "id,text,created", []string{"id", "text", "created"}},
		{"fields with spaces", "id, text, created", []string{"id", "text", "created"}},
		{"empty string", "", []string(nil)},
		{"single field with spaces", "  id  ", []string{"id"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFieldsFlag(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d", len(tt.expected), len(result))
				return
			}
			for i, field := range result {
				if field != tt.expected[i] {
					t.Errorf("field %d: expected %q, got %q", i, tt.expected[i], field)
				}
			}
		})
	}
}

func TestFormatFields(t *testing.T) {
	// Create test items
	now := time.Now()
	item1 := &model.Item{
		ID:   "item_1",
		Text: "First task",
		Metadata: &model.Metadata{
			Attributes: map[string]string{
				"type":     "task",
				"status":   "done",
				"priority": "high",
			},
			Tags:     []string{"urgent", "work"},
			Created:  now,
			Modified: now.Add(1 * time.Hour),
		},
	}

	item2 := &model.Item{
		ID:   "item_2",
		Text: "Second item",
		Metadata: &model.Metadata{
			Attributes: map[string]string{"type": "note"},
			Created:    now,
			Modified:   now,
		},
	}

	item2.Parent = item1 // Set parent for depth testing

	items := []*model.Item{item1, item2}
	outline := &model.Outline{}

	formatter := NewSearchOutputFormatter()

	tests := []struct {
		name          string
		fields        []string
		shouldContain []string
	}{
		{
			name:          "default fields",
			fields:        []string{"id", "text", "attributes"},
			shouldContain: []string{"item_1", "First task", "@type=task"},
		},
		{
			name:          "custom fields",
			fields:        []string{"id", "text"},
			shouldContain: []string{"item_1\tFirst task", "item_2\tSecond item"},
		},
		{
			name:          "tags field",
			fields:        []string{"id", "tags"},
			shouldContain: []string{"urgent,work"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatFields(items, tt.fields, outline)
			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("expected output to contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	now := time.Now()
	item := &model.Item{
		ID:   "item_1",
		Text: "Test task",
		Metadata: &model.Metadata{
			Attributes: map[string]string{"type": "task", "status": "done"},
			Tags:       []string{"test"},
			Created:    now,
			Modified:   now,
		},
	}

	items := []*model.Item{item}
	outline := &model.Outline{}

	formatter := NewSearchOutputFormatter()
	result, err := formatter.formatJSON(items, []string{"id", "text", "attributes"}, outline)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that it's valid JSON and contains expected fields
	if !strings.Contains(result, "item_1") {
		t.Errorf("expected JSON to contain item ID")
	}
	if !strings.Contains(result, "Test task") {
		t.Errorf("expected JSON to contain item text")
	}
	if !strings.Contains(result, "type") {
		t.Errorf("expected JSON to contain attributes")
	}

	// Check it starts with [
	if !strings.HasPrefix(strings.TrimSpace(result), "[") {
		t.Errorf("expected JSON array format")
	}
}

func TestFormatJSONL(t *testing.T) {
	now := time.Now()
	item1 := &model.Item{
		ID:   "item_1",
		Text: "First",
		Metadata: &model.Metadata{
			Attributes: map[string]string{"type": "task"},
			Created:    now,
			Modified:   now,
		},
	}

	item2 := &model.Item{
		ID:   "item_2",
		Text: "Second",
		Metadata: &model.Metadata{
			Attributes: map[string]string{"type": "note"},
			Created:    now,
			Modified:   now,
		},
	}

	items := []*model.Item{item1, item2}
	outline := &model.Outline{}

	formatter := NewSearchOutputFormatter()
	result, err := formatter.formatJSONL(items, []string{"id", "text"}, outline)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	// Check each line is valid JSON
	for i, line := range lines {
		if !strings.Contains(line, "item_") {
			t.Errorf("line %d doesn't contain item ID", i)
		}
	}
}

func TestGetFieldValue(t *testing.T) {
	now := time.Now()
	item := &model.Item{
		ID:   "test_id",
		Text: "Test text",
		Metadata: &model.Metadata{
			Attributes: map[string]string{
				"status": "active",
				"type":   "task",
			},
			Tags:     []string{"tag1", "tag2"},
			Created:  now,
			Modified: now.Add(30 * time.Minute),
		},
	}

	parent := &model.Item{ID: "parent_id", Text: "Parent"}
	item.Parent = parent

	formatter := NewSearchOutputFormatter()
	outline := &model.Outline{}

	tests := []struct {
		field    string
		expected interface{}
	}{
		{"id", "test_id"},
		{"text", "Test text"},
		{"parent_id", "parent_id"},
		{"attr:status", "active"},
		{"attr:type", "task"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := formatter.getFieldValue(item, tt.field, outline)
			if result != tt.expected {
				t.Errorf("field %q: expected %v, got %v", tt.field, tt.expected, result)
			}
		})
	}
}

func TestGetItemDepth(t *testing.T) {
	root := &model.Item{ID: "root"}
	level1 := &model.Item{ID: "level1", Parent: root}
	level2 := &model.Item{ID: "level2", Parent: level1}
	level3 := &model.Item{ID: "level3", Parent: level2}

	formatter := NewSearchOutputFormatter()

	tests := []struct {
		item     *model.Item
		expected int
	}{
		{root, 0},
		{level1, 1},
		{level2, 2},
		{level3, 3},
	}

	for _, tt := range tests {
		result := formatter.getItemDepth(tt.item)
		if result != tt.expected {
			t.Errorf("item %q: expected depth %d, got %d", tt.item.ID, tt.expected, result)
		}
	}
}

func TestGetItemPath(t *testing.T) {
	root := &model.Item{ID: "root", Text: "Root"}
	level1 := &model.Item{ID: "level1", Text: "Level1", Parent: root}
	level2 := &model.Item{ID: "level2", Text: "Level2", Parent: level1}

	formatter := NewSearchOutputFormatter()
	outline := &model.Outline{}

	tests := []struct {
		item     *model.Item
		expected []string
	}{
		{root, []string{"Root"}},
		{level1, []string{"Root", "Level1"}},
		{level2, []string{"Root", "Level1", "Level2"}},
	}

	for _, tt := range tests {
		result := formatter.getItemPath(tt.item, outline)
		if len(result) != len(tt.expected) {
			t.Errorf("item %q: expected %d path elements, got %d", tt.item.ID, len(tt.expected), len(result))
			continue
		}
		for i, elem := range result {
			if elem != tt.expected[i] {
				t.Errorf("item %q: expected path element %d to be %q, got %q", tt.item.ID, i, tt.expected[i], elem)
			}
		}
	}
}
