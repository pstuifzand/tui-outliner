package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

	// Convert original path to absolute path before storing
	absPath, err := filepath.Abs(originalPath)
	if err != nil {
		// If we can't get absolute path, use the original
		absPath = originalPath
	}

	// Set the original filename in the outline before saving
	outline.OriginalFilename = absPath

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

// GetBackupDir is a public function to get the backup directory
func GetBackupDir() string {
	return getBackupDir()
}

// BackupMetadata holds parsed information about a backup file
type BackupMetadata struct {
	FilePath     string    // Full path to backup file
	Timestamp    time.Time // Parsed timestamp from filename
	SessionID    string    // 8-character session ID
	OriginalFile string    // Original filename stored in backup
}

// FindBackupsForFile returns all backup files for a given original filename, sorted chronologically
func (bm *BackupManager) FindBackupsForFile(originalFilePath string) ([]BackupMetadata, error) {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupMetadata

	// Normalize the search path to absolute for consistent comparison
	var searchPath string
	if originalFilePath != "" {
		absPath, err := filepath.Abs(originalFilePath)
		if err != nil {
			searchPath = originalFilePath
		} else {
			// Use Clean to normalize the path (remove . and .. and duplicate slashes)
			searchPath = filepath.Clean(absPath)
		}
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tuo") {
			continue
		}

		// Parse metadata from file
		metadata, err := parseBackupFilename(entry.Name(), filepath.Join(bm.backupDir, entry.Name()))
		if err != nil {
			continue // Skip files that can't be parsed
		}

		// If originalFilePath specified, filter by it
		if searchPath != "" {
			// Normalize the backup's original file path too
			backupPath := filepath.Clean(metadata.OriginalFile)
			if backupPath != searchPath {
				continue
			}
		}

		backups = append(backups, metadata)
	}

	// Sort chronologically by timestamp
	sortBackupsByTimestamp(backups)
	return backups, nil
}

// parseBackupFilename extracts metadata from a backup filename
// Expected format: YYYYMMDD_HHMMSS_<sessionID>.tuo
func parseBackupFilename(filename string, fullPath string) (BackupMetadata, error) {
	// Parse filename format: YYYYMMDD_HHMMSS_XXXXXXXX.tuo
	if len(filename) < 22 { // Min length for valid format
		return BackupMetadata{}, fmt.Errorf("filename too short")
	}

	// Extract timestamp: YYYYMMDD_HHMMSS (15 characters)
	timestampStr := filename[:15]

	// Extract session ID: 8 characters after the second underscore
	sessionID := filename[16 : 16+8]

	// Parse timestamp
	timestamp, err := time.Parse("20060102_150405", timestampStr)
	if err != nil {
		return BackupMetadata{}, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Read the backup file to get original filename
	var originalFile string
	data, err := os.ReadFile(fullPath)
	if err == nil {
		var outline model.Outline
		if err := json.Unmarshal(data, &outline); err == nil {
			originalFile = outline.OriginalFilename
		}
	}

	return BackupMetadata{
		FilePath:     fullPath,
		Timestamp:    timestamp,
		SessionID:    sessionID,
		OriginalFile: originalFile,
	}, nil
}

// sortBackupsByTimestamp sorts backups chronologically (oldest first)
func sortBackupsByTimestamp(backups []BackupMetadata) {
	slices.SortFunc(backups, func(a, b BackupMetadata) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
}
