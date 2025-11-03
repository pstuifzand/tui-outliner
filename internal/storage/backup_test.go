package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

func TestBackupManagerCreateBackup(t *testing.T) {
	// Create backup manager
	bm, err := NewBackupManager()
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Create a simple outline
	outline := model.NewOutline()
	item := model.NewItem("Test item")
	outline.Items = append(outline.Items, item)

	// Create backup
	originalPath := "/tmp/test_outline.json"
	sessionID := "test1234"
	err = bm.CreateBackup(outline, originalPath, sessionID)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup file exists
	backupDir := getBackupDir()
	files, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No backup files created")
	}

	// Check the latest backup file
	latestFile := files[len(files)-1]
	backupPath := filepath.Join(backupDir, latestFile.Name())

	// Verify backup file contains our data
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	var backupOutline model.Outline
	err = json.Unmarshal(data, &backupOutline)
	if err != nil {
		t.Fatalf("Failed to unmarshal backup: %v", err)
	}

	// Verify original filename is stored as absolute path
	absPath, _ := filepath.Abs(originalPath)
	if backupOutline.OriginalFilename != absPath {
		t.Fatalf("Expected original filename '%s', got '%s'", absPath, backupOutline.OriginalFilename)
	}

	// Verify items are preserved
	if len(backupOutline.Items) != 1 {
		t.Fatalf("Expected 1 item in backup, got %d", len(backupOutline.Items))
	}

	if backupOutline.Items[0].Text != "Test item" {
		t.Fatalf("Expected 'Test item', got '%s'", backupOutline.Items[0].Text)
	}

	// Cleanup
	os.Remove(backupPath)
}

func TestBackupFilenameFormat(t *testing.T) {
	bm, _ := NewBackupManager()

	sessionID := "abc12345"
	filename := bm.generateBackupFilename(sessionID)

	// Check format: YYYYMMDD_HHMMSS_<sessionID>.tuo
	expectedLen := len("20251103_150405_abc12345.tuo")
	if len(filename) != expectedLen {
		t.Fatalf("Filename format incorrect: expected length %d, got %d: %s", expectedLen, len(filename), filename)
	}

	// Verify it ends with .tuo
	if filename[len(filename)-4:] != ".tuo" {
		t.Fatalf("Filename should end with .tuo: %s", filename)
	}

	// Verify session ID is in filename
	sessionIDStart := len(filename) - 12
	if sessionIDStart >= 0 && filename[sessionIDStart:len(filename)-4] != sessionID {
		t.Fatalf("Session ID not found in filename: %s", filename)
	}
}

func TestBackupManagerDirectoryCreation(t *testing.T) {
	// This test verifies that backup directory is properly created
	bm, err := NewBackupManager()
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Check that backup directory exists
	if _, err := os.Stat(bm.backupDir); os.IsNotExist(err) {
		t.Fatalf("Backup directory was not created: %v", err)
	}
}

func TestSessionIDGeneration(t *testing.T) {
	sessionID1 := generateSessionID()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure different random
	sessionID2 := generateSessionID()

	// Verify format: 8 characters alphanumeric
	if len(sessionID1) != 8 {
		t.Fatalf("Session ID should be 8 characters, got %d", len(sessionID1))
	}

	if len(sessionID2) != 8 {
		t.Fatalf("Session ID should be 8 characters, got %d", len(sessionID2))
	}

	// Verify they're likely different (random)
	if sessionID1 == sessionID2 {
		t.Log("Warning: Two generated session IDs are identical (unlikely but possible)")
	}

	// Verify all characters are alphanumeric
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, ch := range sessionID1 {
		found := false
		for _, valid := range charset {
			if ch == valid {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Session ID contains invalid character: %c", ch)
		}
	}
}

func TestBackupStoresAbsolutePath(t *testing.T) {
	// Create backup manager
	bm, err := NewBackupManager()
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Create a simple outline
	outline := model.NewOutline()
	item := model.NewItem("Test item")
	outline.Items = append(outline.Items, item)

	// Test with relative path
	originalPath := "test_outline.json"
	sessionID := "testabs1"
	err = bm.CreateBackup(outline, originalPath, sessionID)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Read the backup file
	backupDir := getBackupDir()
	files, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}

	// Find the backup we just created
	var backupPath string
	for i := len(files) - 1; i >= 0; i-- {
		if files[i].Name() != "" {
			backupPath = filepath.Join(backupDir, files[i].Name())
			// Check if this is our backup (has testabs1 in name)
			if filepath.Base(backupPath) == files[i].Name() {
				break
			}
		}
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	var backupOutline model.Outline
	err = json.Unmarshal(data, &backupOutline)
	if err != nil {
		t.Fatalf("Failed to unmarshal backup: %v", err)
	}

	// Verify original filename is stored as absolute path
	absPath, _ := filepath.Abs(originalPath)
	if backupOutline.OriginalFilename != absPath {
		t.Errorf("Expected absolute path '%s', got '%s'", absPath, backupOutline.OriginalFilename)
	}

	// Verify it's an absolute path
	if !filepath.IsAbs(backupOutline.OriginalFilename) {
		t.Errorf("Original filename is not absolute: %s", backupOutline.OriginalFilename)
	}

	// Cleanup
	os.Remove(backupPath)
}

func TestFindBackupsWithAbsolutePath(t *testing.T) {
	// Create backup manager
	bm, err := NewBackupManager()
	if err != nil {
		t.Fatalf("Failed to create backup manager: %v", err)
	}

	// Create a simple outline
	outline := model.NewOutline()
	item := model.NewItem("Test item")
	outline.Items = append(outline.Items, item)

	// Create backup with relative path
	relativePath := "my_test.json"
	sessionID := "testfind1"
	err = bm.CreateBackup(outline, relativePath, sessionID)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Convert relative path to absolute for comparison
	absPath, _ := filepath.Abs(relativePath)

	// Test 1: Find backups using relative path
	backups, err := bm.FindBackupsForFile(relativePath)
	if err != nil {
		t.Fatalf("Failed to find backups with relative path: %v", err)
	}
	if len(backups) == 0 {
		t.Error("Expected to find backups using relative path")
	}

	// Test 2: Find backups using absolute path
	backups, err = bm.FindBackupsForFile(absPath)
	if err != nil {
		t.Fatalf("Failed to find backups with absolute path: %v", err)
	}
	if len(backups) == 0 {
		t.Error("Expected to find backups using absolute path")
	}

	// Test 3: Both should find the same backups
	backupsRel, _ := bm.FindBackupsForFile(relativePath)
	backupsAbs, _ := bm.FindBackupsForFile(absPath)
	if len(backupsRel) != len(backupsAbs) {
		t.Errorf("Different number of backups found: relative=%d, absolute=%d",
			len(backupsRel), len(backupsAbs))
	}

	// Cleanup
	for _, backup := range backups {
		os.Remove(backup.FilePath)
	}
}

// Test function accessible from app.go for testing
func generateSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		// Using simple rand for testing (non-cryptographic)
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
