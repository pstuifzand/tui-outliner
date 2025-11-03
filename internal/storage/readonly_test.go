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
			result := isBackupFile(tt.path)
			if result != tt.expected {
				t.Errorf("isBackupFile(%q) = %v, want %v", tt.path, result, tt.expected)
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
