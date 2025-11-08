package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/storage"
	"github.com/pstuifzand/tui-outliner/internal/ui"
)

// handleDiffCommand shows diff view with selected backup
func (a *App) handleDiffCommand(parts []string) {
	if a.originalFilePath == "" {
		a.SetStatus("No file to compare backups for")
		return
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found for this file")
		return
	}

	if len(backups) < 1 {
		a.SetStatus("No backups found for this file")
		return
	}

	// Append the current file as a virtual "backup" entry at the end
	// Since backups are stored oldest-first, appending at the end makes it the newest
	// In reversed mode (newest-first display), it will appear at the top
	currentEntry := storage.BackupMetadata{
		FilePath:     "(current)",
		Timestamp:    time.Now(),
		SessionID:    "current",
		OriginalFile: a.originalFilePath,
	}
	backupsWithCurrent := make([]storage.BackupMetadata, 0, len(backups)+1)
	backupsWithCurrent = append(backupsWithCurrent, backups...)
	backupsWithCurrent = append(backupsWithCurrent, currentEntry)

	// Show backup selector with side-by-side diff preview
	a.backupSelectorWidget.Show(backupsWithCurrent, a.outline,
		func(backup storage.BackupMetadata) {
			// Don't restore if it's the "(current)" virtual entry
			if backup.FilePath == "(current)" {
				a.SetStatus("Already at current state")
				return
			}

			// Load the selected backup
			backupData, err := os.ReadFile(backup.FilePath)
			if err != nil {
				a.SetStatus(fmt.Sprintf("Failed to read backup: %v", err))
				return
			}

			var restoredOutline model.Outline
			if err := json.Unmarshal(backupData, &restoredOutline); err != nil {
				a.SetStatus(fmt.Sprintf("Failed to parse backup: %v", err))
				return
			}

			// Close any active editor
			if a.editor != nil {
				a.editor = nil
			}

			// Replace current outline with restored backup
			a.outline = &restoredOutline
			a.dirty = true

			// Recreate the tree view with restored items
			a.tree = ui.NewTreeView(a.outline.Items)

			// Close search widget if open
			if a.search != nil && a.search.IsActive() {
				a.search.Stop()
			}

			// Close node search widget if open
			if a.nodeSearchWidget != nil {
				a.nodeSearchWidget.Hide()
			}

			a.SetStatus(fmt.Sprintf("Restored backup from %s", backup.Timestamp.Format("2006-01-02 15:04:05")))
		},
		func() {
			a.SetStatus("Diff cancelled")
		})
}

// handlePreviousBackupSameSession loads the previous backup with the same session ID
func (a *App) handlePreviousBackupSameSession() bool {
	if a.originalFilePath == "" {
		a.SetStatus("No file to find backups for")
		return false
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return false
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found")
		return false
	}

	// Find current backup position in the list
	currentIdx := -1
	for i, b := range backups {
		if b.FilePath == a.currentBackupPath {
			currentIdx = i
			break
		}
	}

	// If not viewing a backup, go to the most recent one
	if currentIdx == -1 {
		// Find backups with same session ID
		for i := len(backups) - 1; i >= 0; i-- {
			if backups[i].SessionID == a.sessionID {
				if a.loadBackupFile(backups[i]) {
					a.SetStatus(fmt.Sprintf("Backup: %s (%s)", backups[i].Timestamp.Format("2006-01-02 15:04:05"), backups[i].SessionID))
					return true
				}
			}
		}
		a.SetStatus("No backups with current session ID")
		return false
	}

	// Find previous backup with same session ID
	for i := currentIdx - 1; i >= 0; i-- {
		if backups[i].SessionID == a.sessionID {
			if a.loadBackupFile(backups[i]) {
				a.SetStatus(fmt.Sprintf("Backup: %s (%s)", backups[i].Timestamp.Format("2006-01-02 15:04:05"), backups[i].SessionID))
				return true
			}
		}
	}

	a.SetStatus("No older backups with same session ID")
	return false
}

// handleNextBackupSameSession loads the next backup with the same session ID
func (a *App) handleNextBackupSameSession() bool {
	if a.originalFilePath == "" {
		a.SetStatus("No file to find backups for")
		return false
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return false
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found")
		return false
	}

	// Find current backup position
	currentIdx := -1
	for i, b := range backups {
		if b.FilePath == a.currentBackupPath {
			currentIdx = i
			break
		}
	}

	// Must be viewing a backup to go to next one
	if currentIdx == -1 {
		a.SetStatus("Not viewing a backup")
		return false
	}

	// Find next backup with same session ID
	for i := currentIdx + 1; i < len(backups); i++ {
		if backups[i].SessionID == a.sessionID {
			if a.loadBackupFile(backups[i]) {
				a.SetStatus(fmt.Sprintf("Backup: %s (%s)", backups[i].Timestamp.Format("2006-01-02 15:04:05"), backups[i].SessionID))
				return true
			}
		}
	}

	a.SetStatus("No newer backups with same session ID")
	return false
}

// handlePreviousBackupAnySession loads the previous backup regardless of session ID
func (a *App) handlePreviousBackupAnySession() bool {
	if a.originalFilePath == "" {
		a.SetStatus("No file to find backups for")
		return false
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return false
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found")
		return false
	}

	// Find current backup position
	currentIdx := -1
	for i, b := range backups {
		if b.FilePath == a.currentBackupPath {
			currentIdx = i
			break
		}
	}

	// If not viewing a backup, go to most recent
	if currentIdx == -1 {
		if a.loadBackupFile(backups[len(backups)-1]) {
			b := backups[len(backups)-1]
			a.SetStatus(fmt.Sprintf("Backup: %s (%s)", b.Timestamp.Format("2006-01-02 15:04:05"), b.SessionID))
			return true
		}
		return false
	}

	// Find previous backup
	if currentIdx > 0 {
		if a.loadBackupFile(backups[currentIdx-1]) {
			b := backups[currentIdx-1]
			a.SetStatus(fmt.Sprintf("Backup: %s (%s)", b.Timestamp.Format("2006-01-02 15:04:05"), b.SessionID))
			return true
		}
	}

	a.SetStatus("No older backups")
	return false
}

// handleNextBackupAnySession loads the next backup regardless of session ID
func (a *App) handleNextBackupAnySession() bool {
	if a.originalFilePath == "" {
		a.SetStatus("No file to find backups for")
		return false
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return false
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found")
		return false
	}

	// Find current backup position
	currentIdx := -1
	for i, b := range backups {
		if b.FilePath == a.currentBackupPath {
			currentIdx = i
			break
		}
	}

	// Must be viewing a backup to go to next one
	if currentIdx == -1 {
		a.SetStatus("Not viewing a backup")
		return false
	}

	// Find next backup
	if currentIdx < len(backups)-1 {
		if a.loadBackupFile(backups[currentIdx+1]) {
			b := backups[currentIdx+1]
			a.SetStatus(fmt.Sprintf("Backup: %s (%s)", b.Timestamp.Format("2006-01-02 15:04:05"), b.SessionID))
			return true
		}
	}

	a.SetStatus("No newer backups")
	return false
}

// loadBackupFile loads a backup file and updates the app state
func (a *App) loadBackupFile(backup storage.BackupMetadata) bool {
	if err := a.Load(backup.FilePath); err != nil {
		a.SetStatus(fmt.Sprintf("Failed to load backup: %s", err.Error()))
		return false
	}

	a.currentBackupPath = backup.FilePath
	a.originalFilePath = backup.OriginalFile
	a.sessionID = backup.SessionID

	return true
}

// generateSessionID creates a random 8-character session ID for backup naming
func generateSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
