package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// BackupManager handles backup creation for outline files
type BackupManager struct {
	backupDir string
}

// NewBackupManager creates a new backup manager
func NewBackupManager() (*BackupManager, error) {
	// Ensure backup directory exists
	backupDir := getBackupDir()
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &BackupManager{
		backupDir: backupDir,
	}, nil
}

// CreateBackup creates a timestamped backup of the outline before saving
// It stores both the outline data and the original filename
func (bm *BackupManager) CreateBackup(outline *model.Outline, originalPath string, sessionID string) error {
	// Generate backup filename
	filename := bm.generateBackupFilename(sessionID)

	// Set the original filename in the outline before saving
	outline.OriginalFilename = originalPath

	// Create backup file path
	backupPath := filepath.Join(bm.backupDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(outline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup JSON: %w", err)
	}

	// Write backup file
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// generateBackupFilename creates a filename in the format: YYYYMMDD_HHMMSS_<sessionID>.tuo
func (bm *BackupManager) generateBackupFilename(sessionID string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.tuo", timestamp, sessionID)
}

// getBackupDir returns the path to the backup directory
func getBackupDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp if home directory cannot be determined
		return filepath.Join("/tmp", ".tui-outliner", "backups")
	}
	return filepath.Join(homeDir, ".local", "share", "tui-outliner", "backups")
}
