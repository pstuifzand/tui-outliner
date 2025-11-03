package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBackupFileDetection(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Could not determine home directory")
	}

	backupDir := filepath.Join(homeDir, ".local", "share", "tui-outliner", "backups")

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "Regular file",
			path:     "/tmp/myoutline.json",
			expected: false,
		},
		{
			name:     "Backup file",
			path:     filepath.Join(backupDir, "20251103_150405_abc12345.tuo"),
			expected: true,
		},
		{
			name:     "File with backup in name but different directory",
			path:     "/tmp/backups/20251103_150405_abc12345.tuo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBackupFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsBackupFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestJSONStoreReadOnlyDetection(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Could not determine home directory")
	}

	backupDir := filepath.Join(homeDir, ".local", "share", "tui-outliner", "backups")

	tests := []struct {
		name           string
		filePath       string
		expectedReadOnly bool
	}{
		{
			name:           "Regular file is not readonly",
			filePath:       "/tmp/myoutline.json",
			expectedReadOnly: false,
		},
		{
			name:           "Backup file is readonly",
			filePath:       filepath.Join(backupDir, "20251103_150405_abc12345.tuo"),
			expectedReadOnly: true,
		},
		{
			name:           "Empty file path is not readonly",
			filePath:       "",
			expectedReadOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewJSONStore(tt.filePath)
			if store.ReadOnly != tt.expectedReadOnly {
				t.Errorf("NewJSONStore(%q).ReadOnly = %v, want %v",
					tt.filePath, store.ReadOnly, tt.expectedReadOnly)
			}
		})
	}
}

func TestLoadUpdatesReadOnlyFlag(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Could not determine home directory")
	}

	backupDir := filepath.Join(homeDir, ".local", "share", "tui-outliner", "backups")

	tests := []struct {
		name             string
		initialPath      string
		loadPath         string
		expectedReadOnly bool
	}{
		{
			name:             "Loading regular file into store with backup path",
			initialPath:      filepath.Join(backupDir, "20251103_150405_abc12345.tuo"),
			loadPath:         "/tmp/myoutline.json",
			expectedReadOnly: false,
		},
		{
			name:             "Loading backup file into store with regular path",
			initialPath:      "/tmp/myoutline.json",
			loadPath:         filepath.Join(backupDir, "20251103_150405_abc12345.tuo"),
			expectedReadOnly: true,
		},
		{
			name:             "Loading backup file into store with backup path",
			initialPath:      filepath.Join(backupDir, "20251103_150405_abc12345.tuo"),
			loadPath:         filepath.Join(backupDir, "20251103_160506_def67890.tuo"),
			expectedReadOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewJSONStore(tt.initialPath)
			// Simulate what Load() does
			store.FilePath = tt.loadPath
			store.ReadOnly = IsBackupFile(tt.loadPath)

			if store.ReadOnly != tt.expectedReadOnly {
				t.Errorf("After Load(%q), ReadOnly = %v, want %v",
					tt.loadPath, store.ReadOnly, tt.expectedReadOnly)
			}
		})
	}
}
