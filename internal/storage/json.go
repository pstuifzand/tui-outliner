package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// JSONStore handles JSON file persistence
type JSONStore struct {
	FilePath       string
	backupManager  *BackupManager
	sessionID      string
}

// NewJSONStore creates a new JSON store for the given file path
func NewJSONStore(filePath string) *JSONStore {
	backupManager, err := NewBackupManager()
	if err != nil {
		log.Printf("Warning: Failed to initialize backup manager: %v\n", err)
	}

	return &JSONStore{
		FilePath:      filePath,
		backupManager: backupManager,
		sessionID:     "",
	}
}

// SetSessionID sets the session ID for backup naming
func (s *JSONStore) SetSessionID(sessionID string) {
	s.sessionID = sessionID
}

// Load loads an outline from a JSON file
func (s *JSONStore) Load() (*model.Outline, error) {
	// If no file path specified, return a new empty outline
	if s.FilePath == "" {
		return model.NewOutline(), nil
	}

	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty outline if file doesn't exist
			return model.NewOutline(), nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var outline model.Outline
	if err := json.Unmarshal(data, &outline); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Restore parent pointers after deserialization
	restoreParentPointers(outline.Items)

	// Build ID index for fast lookups
	outline.BuildIndex()

	// Resolve virtual child references
	outline.ResolveVirtualChildren()

	return &outline, nil
}

// Save saves an outline to a JSON file
func (s *JSONStore) Save(outline *model.Outline) error {
	if s.FilePath == "" {
		return fmt.Errorf("no file path specified. Use :w <filename> to save")
	}
	return s.SaveToFile(outline, s.FilePath)
}

// SaveToFile saves an outline to a specified file path
func (s *JSONStore) SaveToFile(outline *model.Outline, filePath string) error {
	// Create backup before saving (fail gracefully if backup fails)
	if s.backupManager != nil && s.sessionID != "" {
		originalPath := filePath
		// For buffer mode, use a placeholder name
		if originalPath == "" {
			originalPath = "unsaved_buffer"
		}

		if err := s.backupManager.CreateBackup(outline, originalPath, s.sessionID); err != nil {
			log.Printf("Warning: Failed to create backup: %v\n", err)
			// Don't return error - backup failure shouldn't prevent saving
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	data, err := json.MarshalIndent(outline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// restoreParentPointers reconstructs parent pointers after JSON deserialization
func restoreParentPointers(items []*model.Item) {
	for _, item := range items {
		if item.ID == "" {
			log.Println("Item has no parent pointers")
		}
		for _, child := range item.Children {
			child.Parent = item
		}
		restoreParentPointers(item.Children)
	}
}

// FileExists checks if the outline file exists
func (s *JSONStore) FileExists() bool {
	_, err := os.Stat(s.FilePath)
	return err == nil
}
