package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// JSONStore handles JSON file persistence
type JSONStore struct {
	FilePath string
}

// NewJSONStore creates a new JSON store for the given file path
func NewJSONStore(filePath string) *JSONStore {
	return &JSONStore{
		FilePath: filePath,
	}
}

// Load loads an outline from a JSON file
func (s *JSONStore) Load() (*model.Outline, error) {
	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty outline if file doesn't exist
			return model.NewOutline("Untitled"), nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var outline model.Outline
	if err := json.Unmarshal(data, &outline); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Restore parent pointers after deserialization
	restoreParentPointers(outline.Items)

	return &outline, nil
}

// Save saves an outline to a JSON file
func (s *JSONStore) Save(outline *model.Outline) error {
	// Ensure directory exists
	dir := filepath.Dir(s.FilePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	data, err := json.MarshalIndent(outline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(s.FilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// restoreParentPointers reconstructs parent pointers after JSON deserialization
func restoreParentPointers(items []*model.Item) {
	for _, item := range items {
		if len(item.Children) > 0 {
			for _, child := range item.Children {
				child.Parent = item
			}
			restoreParentPointers(item.Children)
		}
	}
}

// FileExists checks if the outline file exists
func (s *JSONStore) FileExists() bool {
	_, err := os.Stat(s.FilePath)
	return err == nil
}
