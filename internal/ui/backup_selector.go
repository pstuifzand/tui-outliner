package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/config"
	"github.com/pstuifzand/tui-outliner/internal/diff"
	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/storage"
)

// ViewMode represents the current view mode in the right panel
type ViewMode int

const (
	ViewModeDiff ViewMode = iota
	ViewModeTree
)

// BackupSelectorWidget allows users to select a backup for comparison with side-by-side diff preview
type BackupSelectorWidget struct {
	visible       bool
	backups       []storage.BackupMetadata
	selectedIndex int
	scrollOffset  int
	maxWidth      int
	maxHeight     int

	currentOutline *model.Outline
	callback       func(backup storage.BackupMetadata)
	onCancel       func()

	// Diff preview state
	diffResults      map[int]*diff.DiffResult // Cached diff results for each backup
	diffLines        []diff.DiffLine
	diffScrollOffset int

	// Range selection state
	visualMode      bool          // Whether in visual/range selection mode
	selectionStart  int           // Start of selection range
	selectedIndices map[int]bool  // Map of selected indices for multi-select

	// Reverse mode state
	reversed bool // Whether backups are displayed in reverse order (newest first)

	// View mode state
	viewMode      ViewMode               // Current view mode (diff or tree)
	treeView      *TreeView              // Tree view for displaying backup outline
	backupOutline *model.Outline         // Currently loaded backup outline
	treeScrollOffset int                  // Scroll position in tree view
	cfg           *config.Config         // Configuration for tree rendering
}

// NewBackupSelectorWidget creates a new backup selector widget
func NewBackupSelectorWidget() *BackupSelectorWidget {
	return &BackupSelectorWidget{
		visible:         false,
		backups:         make([]storage.BackupMetadata, 0),
		selectedIndex:   0,
		scrollOffset:    0,
		diffResults:     make(map[int]*diff.DiffResult),
		diffLines:       make([]diff.DiffLine, 0),
		diffScrollOffset: 0,
		visualMode:      false,
		selectionStart:  0,
		selectedIndices: make(map[int]bool),
		reversed:        true, // Default to newest-first order
		viewMode:        ViewModeDiff,
		treeView:        nil,
		backupOutline:   nil,
		treeScrollOffset: 0,
		cfg:             nil,
	}
}

// SetConfig sets the configuration for tree rendering
func (bs *BackupSelectorWidget) SetConfig(cfg *config.Config) {
	bs.cfg = cfg
}

// Show displays the backup selector with available backups
func (bs *BackupSelectorWidget) Show(backups []storage.BackupMetadata, currentOutline *model.Outline, callback func(storage.BackupMetadata), onCancel func()) {
	if len(backups) == 0 {
		return
	}

	bs.backups = backups
	bs.selectedIndex = 0
	bs.scrollOffset = 0
	bs.currentOutline = currentOutline
	bs.callback = callback
	bs.onCancel = onCancel
	bs.diffScrollOffset = 0
	bs.diffResults = make(map[int]*diff.DiffResult) // Clear cache
	bs.visualMode = false                           // Reset visual mode
	bs.selectedIndices = make(map[int]bool)         // Clear selection
	bs.viewMode = ViewModeDiff                      // Reset to diff view
	bs.treeView = nil                               // Reset tree view
	bs.backupOutline = nil                          // Reset backup outline
	bs.treeScrollOffset = 0                         // Reset tree scroll
	bs.visible = true

	// Preload all backup diffs for summary display
	bs.preloadAllDiffs()

	// Load diff for first backup
	bs.updateDiffPreview()
}

// Hide closes the backup selector
func (bs *BackupSelectorWidget) Hide() {
	bs.visible = false
}

// IsVisible returns whether the widget is currently visible
func (bs *BackupSelectorWidget) IsVisible() bool {
	return bs.visible
}

// SetMaxSize updates the maximum dimensions for rendering
func (bs *BackupSelectorWidget) SetMaxSize(width, height int) {
	bs.maxWidth = width
	bs.maxHeight = height
}

// updateDiffPreview loads and computes diff for currently selected backup
func (bs *BackupSelectorWidget) updateDiffPreview() {
	if bs.selectedIndex < 0 || bs.selectedIndex >= len(bs.backups) {
		bs.diffLines = []diff.DiffLine{}
		bs.backupOutline = nil
		return
	}

	// Convert display index to actual index
	actualIdx := bs.actualIndex(bs.selectedIndex)

	// Check if we already have this diff cached
	if diffResult, ok := bs.diffResults[actualIdx]; ok {
		bs.loadDiffResult(diffResult)
		// Load backup outline for tree view
		bs.loadBackupOutline(actualIdx)
		return
	}

	backup := bs.backups[actualIdx]

	// Check if this is the "(current)" virtual entry - it's always at the end (newest)
	isCurrentEntry := actualIdx == len(bs.backups)-1 && backup.FilePath == "(current)"

	var diffResult *diff.DiffResult
	var err error
	var backupOutlineToUse *model.Outline

	if isCurrentEntry {
		// Current file entry - no changes (it IS the current state)
		diffResult = &diff.DiffResult{
			NewItems:      make(map[string]*diff.ItemData),
			DeletedItems:  make(map[string]*diff.ItemData),
			ModifiedItems: make(map[string]*diff.ItemChange),
		}
		backupOutlineToUse = bs.currentOutline
	} else {
		// Load the backup file
		backupData, err2 := os.ReadFile(backup.FilePath)
		if err2 != nil {
			bs.diffLines = []diff.DiffLine{}
			bs.backupOutline = nil
			return
		}

		var backupOutline model.Outline
		if err2 := json.Unmarshal(backupData, &backupOutline); err2 != nil {
			bs.diffLines = []diff.DiffLine{}
			bs.backupOutline = nil
			return
		}

		backupOutlineToUse = &backupOutline

		// For backups: compare current to backup to show what needs to change to get back to this state
		diffResult, err = diff.ComputeDiff(bs.currentOutline, backupOutlineToUse)
		if err != nil {
			bs.diffLines = []diff.DiffLine{}
			bs.backupOutline = nil
			return
		}
	}

	// Cache the result
	bs.diffResults[actualIdx] = diffResult
	bs.loadDiffResult(diffResult)

	// Store backup outline for tree view
	bs.backupOutline = backupOutlineToUse
}

// loadDiffResult processes a diff result and separates it into content and summary
func (bs *BackupSelectorWidget) loadDiffResult(diffResult *diff.DiffResult) {
	allLines := diff.BuildDiffLines(diffResult, false)
	bs.diffScrollOffset = 0

	// Extract only content lines (not summary)
	bs.diffLines = []diff.DiffLine{}
	for _, line := range allLines {
		if line.Type != diff.DiffTypeSummary {
			bs.diffLines = append(bs.diffLines, line)
		}
	}
}

// loadBackupOutline loads the backup outline when cached
func (bs *BackupSelectorWidget) loadBackupOutline(actualIdx int) {
	backup := bs.backups[actualIdx]

	// Check if this is the "(current)" virtual entry
	isCurrentEntry := actualIdx == len(bs.backups)-1 && backup.FilePath == "(current)"

	if isCurrentEntry {
		bs.backupOutline = bs.currentOutline
		return
	}

	// Load the backup file
	backupData, err := os.ReadFile(backup.FilePath)
	if err != nil {
		bs.backupOutline = nil
		return
	}

	var backupOutline model.Outline
	if err := json.Unmarshal(backupData, &backupOutline); err != nil {
		bs.backupOutline = nil
		return
	}

	bs.backupOutline = &backupOutline
}

// preloadAllDiffs computes and caches diffs for all backups
func (bs *BackupSelectorWidget) preloadAllDiffs() {
	for i := range bs.backups {
		backup := bs.backups[i]

		// Check if this is the "(current)" virtual entry - it's always at the end (newest)
		isCurrentEntry := i == len(bs.backups)-1 && backup.FilePath == "(current)"

		var diffResult *diff.DiffResult
		var err error

		if isCurrentEntry {
			// Current file entry - no changes (it IS the current state)
			diffResult = &diff.DiffResult{
				NewItems:      make(map[string]*diff.ItemData),
				DeletedItems:  make(map[string]*diff.ItemData),
				ModifiedItems: make(map[string]*diff.ItemChange),
			}
		} else {
			// Load the backup file
			backupData, err2 := os.ReadFile(backup.FilePath)
			if err2 != nil {
				continue
			}

			var backupOutline model.Outline
			if err2 := json.Unmarshal(backupData, &backupOutline); err2 != nil {
				continue
			}

			// For backups: compare current to backup to show what needs to change to get back to this state
			diffResult, err = diff.ComputeDiff(bs.currentOutline, &backupOutline)
			if err != nil {
				continue
			}
		}

		// Cache the result
		bs.diffResults[i] = diffResult
	}
}

// HandleKeyEvent processes keyboard input
func (bs *BackupSelectorWidget) HandleKeyEvent(ev *tcell.EventKey) {
	if !bs.visible || len(bs.backups) == 0 {
		return
	}

	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		bs.Hide()
		if bs.onCancel != nil {
			bs.onCancel()
		}
	case tcell.KeyUp, tcell.KeyCtrlK:
		bs.selectPrevious()
	case tcell.KeyDown, tcell.KeyCtrlJ:
		bs.selectNext()
	case tcell.KeyEnter:
		if bs.selectedIndex >= 0 && bs.selectedIndex < len(bs.backups) {
			// Convert display index to actual index
			actualIdx := bs.actualIndex(bs.selectedIndex)
			backup := bs.backups[actualIdx]
			bs.Hide()
			if bs.callback != nil {
				bs.callback(backup)
			}
		}
	case tcell.KeyHome:
		bs.selectedIndex = 0
		bs.scrollOffset = 0
		bs.updateDiffPreview()
	case tcell.KeyEnd:
		bs.selectedIndex = len(bs.backups) - 1
		bs.ensureSelected()
		bs.updateDiffPreview()
	case tcell.KeyPgUp:
		if bs.viewMode == ViewModeDiff {
			bs.diffScrollOffset -= (bs.maxHeight - 6) / 2
			if bs.diffScrollOffset < 0 {
				bs.diffScrollOffset = 0
			}
		} else if bs.viewMode == ViewModeTree {
			bs.treeScrollOffset -= (bs.maxHeight - 6) / 2
			if bs.treeScrollOffset < 0 {
				bs.treeScrollOffset = 0
			}
		}
	case tcell.KeyPgDn:
		if bs.viewMode == ViewModeDiff {
			maxScroll := len(bs.diffLines) - (bs.maxHeight - 6)
			if maxScroll < 0 {
				maxScroll = 0
			}
			bs.diffScrollOffset += (bs.maxHeight - 6) / 2
			if bs.diffScrollOffset > maxScroll {
				bs.diffScrollOffset = maxScroll
			}
		} else if bs.viewMode == ViewModeTree {
			// Tree scroll: will be handled in renderTreePreview
			bs.treeScrollOffset += (bs.maxHeight - 6) / 2
		}
	default:
		// Handle character keys
		if ev.Rune() == 'T' {
			// Toggle between diff and tree view
			if bs.viewMode == ViewModeDiff {
				bs.viewMode = ViewModeTree
				bs.treeScrollOffset = 0
			} else {
				bs.viewMode = ViewModeDiff
				bs.diffScrollOffset = 0
			}
		} else if ev.Rune() == 'R' {
			// Toggle reverse/normal order
			bs.reversed = !bs.reversed
			bs.visualMode = false
			bs.selectedIndices = make(map[int]bool)
			bs.scrollOffset = 0
			bs.selectedIndex = 0
		} else if ev.Rune() == 'V' {
			// Start visual selection mode
			if bs.visualMode {
				// Toggle off visual mode
				bs.visualMode = false
				bs.selectedIndices = make(map[int]bool)
			} else {
				// Toggle on visual mode
				bs.visualMode = true
				bs.selectionStart = bs.selectedIndex
				bs.selectedIndices = make(map[int]bool)
				bs.selectedIndices[bs.selectedIndex] = true
			}
		} else if ev.Rune() == 'j' && bs.visualMode {
			// In visual mode, move down and extend selection
			if bs.selectedIndex < len(bs.backups)-1 {
				bs.selectedIndex++
				bs.ensureSelected()
				bs.updateSelection()
				bs.updateVisualModeDiff()
			}
		} else if ev.Rune() == 'k' && bs.visualMode {
			// In visual mode, move up and extend selection
			if bs.selectedIndex > 0 {
				bs.selectedIndex--
				bs.ensureSelected()
				bs.updateSelection()
				bs.updateVisualModeDiff()
			}
		} else if ev.Key() == tcell.KeyUp && bs.visualMode {
			// In visual mode, arrow up extends selection
			if bs.selectedIndex > 0 {
				bs.selectedIndex--
				bs.ensureSelected()
				bs.updateSelection()
				bs.updateVisualModeDiff()
			}
		} else if ev.Key() == tcell.KeyDown && bs.visualMode {
			// In visual mode, arrow down extends selection
			if bs.selectedIndex < len(bs.backups)-1 {
				bs.selectedIndex++
				bs.ensureSelected()
				bs.updateSelection()
				bs.updateVisualModeDiff()
			}
		} else if ev.Rune() == 'j' {
			bs.selectNext()
		} else if ev.Rune() == 'k' {
			bs.selectPrevious()
		}
	}
}

// selectNext moves selection down
func (bs *BackupSelectorWidget) selectNext() {
	if bs.selectedIndex < len(bs.backups)-1 {
		bs.selectedIndex++
		bs.ensureSelected()
		bs.updateDiffPreview()
	}
}

// selectPrevious moves selection up
func (bs *BackupSelectorWidget) selectPrevious() {
	if bs.selectedIndex > 0 {
		bs.selectedIndex--
		bs.ensureSelected()
		bs.updateDiffPreview()
	}
}

// ensureSelected ensures the selected item is visible in the viewport
func (bs *BackupSelectorWidget) ensureSelected() {
	viewHeight := bs.maxHeight - 6 // Account for header, footer, borders

	if bs.selectedIndex < bs.scrollOffset {
		bs.scrollOffset = bs.selectedIndex
	} else if bs.selectedIndex >= bs.scrollOffset+viewHeight {
		bs.scrollOffset = bs.selectedIndex - viewHeight + 1
	}
}

// updateSelection updates the selected range when in visual mode
func (bs *BackupSelectorWidget) updateSelection() {
	if !bs.visualMode {
		return
	}

	// Clear previous selection
	bs.selectedIndices = make(map[int]bool)

	// Select range from start to current
	start := bs.selectionStart
	end := bs.selectedIndex
	if start > end {
		start, end = end, start
	}

	for i := start; i <= end; i++ {
		bs.selectedIndices[i] = true
	}
}

// updateVisualModeDiff computes diff between first and last selected backup in visual mode
func (bs *BackupSelectorWidget) updateVisualModeDiff() {
	if !bs.visualMode || len(bs.selectedIndices) == 0 {
		return
	}

	// Find first and last selected display indices
	var firstDisplayIdx, lastDisplayIdx int
	for displayIdx := range bs.selectedIndices {
		if firstDisplayIdx == 0 || displayIdx < firstDisplayIdx {
			firstDisplayIdx = displayIdx
		}
		if displayIdx > lastDisplayIdx {
			lastDisplayIdx = displayIdx
		}
	}

	if firstDisplayIdx == lastDisplayIdx {
		// Single backup selected, show its diff
		bs.updateDiffPreview()
		return
	}

	// Convert display indices to actual indices
	firstActualIdx := bs.actualIndex(firstDisplayIdx)
	lastActualIdx := bs.actualIndex(lastDisplayIdx)

	// Load first selected backup
	firstBackup := bs.backups[firstActualIdx]
	firstData, err := os.ReadFile(firstBackup.FilePath)
	if err != nil {
		bs.diffLines = []diff.DiffLine{}
		return
	}

	var firstOutline model.Outline
	if err := json.Unmarshal(firstData, &firstOutline); err != nil {
		bs.diffLines = []diff.DiffLine{}
		return
	}

	// Load last selected backup
	lastBackup := bs.backups[lastActualIdx]
	lastData, err := os.ReadFile(lastBackup.FilePath)
	if err != nil {
		bs.diffLines = []diff.DiffLine{}
		return
	}

	var lastOutline model.Outline
	if err := json.Unmarshal(lastData, &lastOutline); err != nil {
		bs.diffLines = []diff.DiffLine{}
		return
	}

	// Compute diff: compare first selected to last selected
	// This shows what changes are needed to get from the first selected backup to the last selected backup
	var diffResult *diff.DiffResult
	var diffErr error

	// Compare first to last selected backup
	diffResult, diffErr = diff.ComputeDiff(&firstOutline, &lastOutline)

	if diffErr != nil {
		bs.diffLines = []diff.DiffLine{}
		return
	}

	// Load and display the diff
	bs.loadDiffResult(diffResult)
}

// actualIndex converts a display index to the actual backup index based on reverse mode
func (bs *BackupSelectorWidget) actualIndex(displayIdx int) int {
	if bs.reversed {
		return len(bs.backups) - 1 - displayIdx
	}
	return displayIdx
}

// displayIndex converts an actual backup index to display index based on reverse mode
func (bs *BackupSelectorWidget) displayIndex(actualIdx int) int {
	if bs.reversed {
		return len(bs.backups) - 1 - actualIdx
	}
	return actualIdx
}

// Render draws the backup selector and diff preview side-by-side on the screen
func (bs *BackupSelectorWidget) Render(screen *Screen) {
	if !bs.visible || len(bs.backups) == 0 {
		return
	}

	width := screen.GetWidth()
	height := screen.GetHeight()
	bs.SetMaxSize(width, height)

	// Calculate dimensions - full width, split into left and right
	boxHeight := height - 4
	leftPanelWidth := width / 2
	rightPanelWidth := width - leftPanelWidth

	if leftPanelWidth < 20 || rightPanelWidth < 20 || boxHeight < 3 {
		return // Too small to render
	}

	startY := 2

	// Draw left panel (backup selector)
	bs.renderLeftPanel(screen, 1, startY, leftPanelWidth-1, boxHeight)

	// Draw right panel (diff preview)
	bs.renderRightPanel(screen, leftPanelWidth+1, startY, rightPanelWidth-1, boxHeight)
}

// renderLeftPanel draws the backup selector on the left side
func (bs *BackupSelectorWidget) renderLeftPanel(screen *Screen, x, y, width, height int) {
	// Draw border
	drawBox(screen, x, y, width, height, screen.TreeNormalStyle())

	// Draw header
	headerStyle := screen.TreeSelectedStyle()
	headerText := fmt.Sprintf(" Backups (%d) ", len(bs.backups))
	if len(headerText) > width-2 {
		headerText = headerText[:width-4] + " "
	}
	screen.DrawStringLimited(x+1, y, headerText, width-2, headerStyle)

	// Fill remaining header width to prevent show-through
	headerLen := len(headerText)
	for col := headerLen; col < width-2; col++ {
		screen.SetCell(x+1+col, y, ' ', headerStyle)
	}

	// Draw content
	contentStartY := y + 2
	contentHeight := height - 4
	bs.renderBackups(screen, x+1, contentStartY, width-2, contentHeight)

	// Draw footer
	footerStyle := screen.TreeNormalStyle()
	footerText := "j/k/↓/↑: select | Enter: restore | Esc: cancel"
	if len(footerText) > width-2 {
		footerText = footerText[:width-4] + " "
	}
	screen.DrawStringLimited(x+1, y+height-1, footerText, width-2, footerStyle)

	// Fill remaining footer width to prevent show-through
	footerLen := len(footerText)
	for col := footerLen; col < width-2; col++ {
		screen.SetCell(x+1+col, y+height-1, ' ', footerStyle)
	}
}

// renderRightPanel draws the diff preview or tree view on the right side
func (bs *BackupSelectorWidget) renderRightPanel(screen *Screen, x, y, width, height int) {
	// Draw border
	drawBox(screen, x, y, width, height, screen.TreeNormalStyle())

	// Draw header
	headerStyle := screen.TreeSelectedStyle()
	headerText := " To get to "

	// Add mode indicator
	modeStr := "[Diff]"
	if bs.viewMode == ViewModeTree {
		modeStr = "[Tree]"
	}

	if bs.selectedIndex >= 0 && bs.selectedIndex < len(bs.backups) {
		// Convert display index to actual index
		actualIdx := bs.actualIndex(bs.selectedIndex)
		backup := bs.backups[actualIdx]

		// Check if this is the "(current)" virtual entry
		isCurrentEntry := actualIdx == 0 && backup.FilePath == "(current)"
		if isCurrentEntry {
			headerText = fmt.Sprintf(" Current (unsaved) %s ", modeStr)
		} else {
			timeStr := backup.Timestamp.Format("2006-01-02 15:04:05")
			headerText = fmt.Sprintf(" %s: %s %s ", modeStr, timeStr, "")
		}
	}
	if len(headerText) > width-2 {
		headerText = headerText[:width-4] + " "
	}
	screen.DrawStringLimited(x+1, y, headerText, width-2, headerStyle)

	// Fill remaining header width to prevent show-through
	headerLen := len(headerText)
	for col := headerLen; col < width-2; col++ {
		screen.SetCell(x+1+col, y, ' ', headerStyle)
	}

	// Draw content
	contentStartY := y + 2
	contentHeight := height - 4
	if bs.viewMode == ViewModeDiff {
		bs.renderDiffPreview(screen, x+1, contentStartY, width-2, contentHeight)
	} else if bs.viewMode == ViewModeTree {
		bs.renderTreePreview(screen, x+1, contentStartY, width-2, contentHeight)
	}

	// Draw footer
	footerStyle := screen.TreeNormalStyle()
	footerText := "T: toggle | PgUp/PgDn: scroll"
	if len(footerText) > width-2 {
		footerText = footerText[:width-4] + " "
	}
	screen.DrawStringLimited(x+1, y+height-1, footerText, width-2, footerStyle)

	// Fill remaining footer width to prevent show-through
	footerLen := len(footerText)
	for col := footerLen; col < width-2; col++ {
		screen.SetCell(x+1+col, y+height-1, ' ', footerStyle)
	}
}

// renderBackups draws the list of backups
func (bs *BackupSelectorWidget) renderBackups(screen *Screen, x, y, width, height int) {
	displayCount := height
	if displayCount > len(bs.backups) {
		displayCount = len(bs.backups)
	}

	endOffset := bs.scrollOffset + displayCount
	if endOffset > len(bs.backups) {
		endOffset = len(bs.backups)
	}

	for displayIdx := bs.scrollOffset; displayIdx < endOffset; displayIdx++ {
		lineY := y + (displayIdx - bs.scrollOffset)
		if lineY >= y+height {
			break
		}

		isSelected := displayIdx == bs.selectedIndex
		// Convert display index to actual index for rendering
		actualIdx := bs.actualIndex(displayIdx)
		bs.renderBackupLine(screen, x, lineY, width, displayIdx, actualIdx, isSelected)
	}

	// Fill empty space below the last backup with background color
	lastRenderedLine := (endOffset - bs.scrollOffset)
	for i := lastRenderedLine; i < height; i++ {
		lineY := y + i
		// Fill with normal style to cover the outline background
		for col := 0; col < width; col++ {
			screen.SetCell(x+col, lineY, ' ', screen.TreeNormalStyle())
		}
	}

	// Show scrollbar indicator if needed
	if len(bs.backups) > height {
		scrollbarY := y + (bs.scrollOffset * height / len(bs.backups))
		scrollbarStyle := screen.TreeSelectedStyle()
		screen.SetCell(x+width-1, scrollbarY, '█', scrollbarStyle)
	}
}

// renderBackupLine draws a single backup entry with inline summary stats
func (bs *BackupSelectorWidget) renderBackupLine(screen *Screen, x, y, width int, displayIdx int, actualIdx int, isSelected bool) {
	backup := bs.backups[actualIdx]

	// Get diff result for this backup
	diffResult, hasDiff := bs.diffResults[actualIdx]

	// Check if this is the "(current)" virtual entry - it's always at the end (newest)
	isCurrentEntry := actualIdx == len(bs.backups)-1 && backup.FilePath == "(current)"

	var line string
	if isCurrentEntry {
		line = "Current"
	} else {
		// Format main line: "Backup #N: 2025-11-03 15:30:45 (session)"
		backupNum := actualIdx // Don't add 1 since current is at index 0
		timeStr := backup.Timestamp.Format("2006-01-02 15:04:05")
		sessionID := backup.SessionID
		if len(sessionID) > 8 {
			sessionID = sessionID[:8]
		}

		line = fmt.Sprintf("Backup #%d: %s (%s)", backupNum, timeStr, sessionID)
	}

	// Determine base style
	var baseStyle tcell.Style
	isRangeSelected := bs.visualMode && bs.selectedIndices[displayIdx]

	if isSelected {
		baseStyle = screen.TreeSelectedStyle()
	} else if isRangeSelected {
		// Use a slightly different style for range-selected items
		baseStyle = screen.HeaderStyle() // Use header style for range selection
	} else {
		baseStyle = screen.TreeNormalStyle()
	}

	// Add selection indicator
	prefix := "  "
	if isSelected {
		prefix = "> "
	} else if isRangeSelected {
		prefix = "* " // Show asterisk for range-selected items
	}
	content := prefix + line

	// Add summary stats if available
	if hasDiff {
		added := len(diffResult.NewItems)
		deleted := len(diffResult.DeletedItems)
		modified := len(diffResult.ModifiedItems)

		if added > 0 || deleted > 0 || modified > 0 {
			stats := ""
			if added > 0 {
				stats += fmt.Sprintf(" +%d", added)
			}
			if modified > 0 {
				stats += fmt.Sprintf(" ~%d", modified)
			}
			if deleted > 0 {
				stats += fmt.Sprintf(" -%d", deleted)
			}
			content += stats
		}
	}

	// Truncate if necessary
	if len(content) > width {
		content = content[:width-3] + "..."
	}

	// Render main content
	screen.DrawStringLimited(x, y, content, width, baseStyle)

	// Now render colored stats inline if there's space
	if hasDiff && width > len(prefix+line)+10 {
		added := len(diffResult.NewItems)
		deleted := len(diffResult.DeletedItems)
		modified := len(diffResult.ModifiedItems)

		xPos := x + len(prefix+line)

		// Render added in green
		if added > 0 {
			addedStr := fmt.Sprintf(" +%d", added)
			greenFg, _, _ := screen.GreenStyle().Decompose()
			_, bg, _ := baseStyle.Decompose()
			greenStyle := tcell.StyleDefault.Foreground(greenFg).Background(bg)
			for i, ch := range addedStr {
				screen.SetCell(xPos+i, y, ch, greenStyle)
			}
			xPos += len(addedStr)
		}

		// Render modified in yellow
		if modified > 0 {
			modStr := fmt.Sprintf(" ~%d", modified)
			yellowFg, _, _ := screen.YellowStyle().Decompose()
			_, bg, _ := baseStyle.Decompose()
			yellowStyle := tcell.StyleDefault.Foreground(yellowFg).Background(bg)
			for i, ch := range modStr {
				screen.SetCell(xPos+i, y, ch, yellowStyle)
			}
			xPos += len(modStr)
		}

		// Render deleted in red
		if deleted > 0 {
			delStr := fmt.Sprintf(" -%d", deleted)
			redFg, _, _ := screen.RedStyle().Decompose()
			_, bg, _ := baseStyle.Decompose()
			redStyle := tcell.StyleDefault.Foreground(redFg).Background(bg)
			for i, ch := range delStr {
				screen.SetCell(xPos+i, y, ch, redStyle)
			}
			xPos += len(delStr)
		}

		// Fill remaining space with base style
		for xPos < x+width {
			screen.SetCell(xPos, y, ' ', baseStyle)
			xPos++
		}
	} else {
		// No stats or not enough space - fill remaining width with background
		fillStart := x + len(content)
		for fillStart < x+width {
			screen.SetCell(fillStart, y, ' ', baseStyle)
			fillStart++
		}
	}
}

// renderDiffPreview draws the diff lines for the currently selected backup
func (bs *BackupSelectorWidget) renderDiffPreview(screen *Screen, x, y, width, height int) {
	if len(bs.diffLines) == 0 {
		emptyStyle := screen.TreeNormalStyle()
		screen.DrawStringLimited(x, y, "(no changes)", width, emptyStyle)
		// Fill empty space below "(no changes)" message
		for i := 1; i < height; i++ {
			lineY := y + i
			for col := 0; col < width; col++ {
				screen.SetCell(x+col, lineY, ' ', emptyStyle)
			}
		}
		return
	}

	displayCount := height
	if displayCount > len(bs.diffLines) {
		displayCount = len(bs.diffLines)
	}

	endOffset := bs.diffScrollOffset + displayCount
	if endOffset > len(bs.diffLines) {
		endOffset = len(bs.diffLines)
	}

	for i := bs.diffScrollOffset; i < endOffset; i++ {
		lineY := y + (i - bs.diffScrollOffset)
		if lineY >= y+height {
			break
		}

		bs.renderDiffLine(screen, x, lineY, width, bs.diffLines[i])
	}

	// Fill empty space below the last diff line with background color
	lastRenderedLine := (endOffset - bs.diffScrollOffset)
	for i := lastRenderedLine; i < height; i++ {
		lineY := y + i
		// Fill with normal style to cover the outline background
		for col := 0; col < width; col++ {
			screen.SetCell(x+col, lineY, ' ', screen.TreeNormalStyle())
		}
	}

	// Show scrollbar indicator if needed
	if len(bs.diffLines) > height {
		scrollbarY := y + (bs.diffScrollOffset * height / len(bs.diffLines))
		scrollbarStyle := screen.TreeSelectedStyle()
		screen.SetCell(x+width-1, scrollbarY, '█', scrollbarStyle)
	}
}

// renderTreePreview draws the tree view of the currently selected backup
func (bs *BackupSelectorWidget) renderTreePreview(screen *Screen, x, y, width, height int) {
	// If no outline loaded, show empty message
	if bs.backupOutline == nil || len(bs.backupOutline.Items) == 0 {
		emptyStyle := screen.TreeNormalStyle()
		screen.DrawStringLimited(x, y, "(outline is empty)", width, emptyStyle)
		// Fill empty space below "(outline is empty)" message
		for i := 1; i < height; i++ {
			lineY := y + i
			for col := 0; col < width; col++ {
				screen.SetCell(x+col, lineY, ' ', emptyStyle)
			}
		}
		return
	}

	// Create a new tree view with the backup outline items
	bs.treeView = NewTreeView(bs.backupOutline.Items)

	// Set max width for the tree view
	bs.treeView.SetMaxWidth(width)

	// Render the tree (read-only view, no selection)
	bs.treeView.Render(screen, y, y+height, -1, bs.cfg)

	// Get display lines for scrollbar
	displayLines := bs.treeView.GetDisplayLines()
	displayLineCount := len(displayLines)

	// Show scrollbar indicator if needed
	if displayLineCount > height {
		viewportOffset := bs.treeView.GetViewportOffset()
		scrollbarY := y + (viewportOffset * height / displayLineCount)
		if scrollbarY < y {
			scrollbarY = y
		}
		if scrollbarY >= y+height {
			scrollbarY = y + height - 1
		}
		scrollbarStyle := screen.TreeSelectedStyle()
		screen.SetCell(x+width-1, scrollbarY, '█', scrollbarStyle)
	}
}

// renderDiffLine draws a single diff line with appropriate coloring
func (bs *BackupSelectorWidget) renderDiffLine(screen *Screen, x, y, width int, line diff.DiffLine) {
	content := line.Content
	indent := line.Indent

	// Apply indentation
	indentStr := strings.Repeat("  ", indent)
	fullContent := indentStr + content

	// Determine style based on line type
	style := bs.getStyleForLineType(screen, line.Type)

	// Draw the line, truncating if necessary
	displayContent := fullContent
	if len(displayContent) > width {
		displayContent = displayContent[:width-3] + "..."
	}

	screen.DrawStringLimited(x, y, displayContent, width, style)

	// Fill remaining width with background style to prevent show-through
	filledLen := len(displayContent)
	for col := filledLen; col < width; col++ {
		screen.SetCell(x+col, y, ' ', style)
	}
}

// getStyleForLineType returns the appropriate style for a diff line type
// Uses only foreground colors with the widget's background color
func (bs *BackupSelectorWidget) getStyleForLineType(screen *Screen, lineType diff.DiffLineType) tcell.Style {
	baseStyle := screen.TreeNormalStyle()
	var fgColor tcell.Color
	var attrs tcell.AttrMask

	switch lineType {
	case diff.DiffTypeHeader:
		// Header style - use foreground from header style
		fg, _, _ := screen.HeaderStyle().Decompose()
		fgColor = fg
		_, _, attrs = screen.HeaderStyle().Decompose()
	case diff.DiffTypeNewSection, diff.DiffTypeNewItem:
		// Green foreground
		fg, _, _ := screen.GreenStyle().Decompose()
		fgColor = fg
	case diff.DiffTypeDeletedSection, diff.DiffTypeDeletedItem:
		// Red foreground
		fg, _, _ := screen.RedStyle().Decompose()
		fgColor = fg
	case diff.DiffTypeModifiedSection, diff.DiffTypeModifiedItem:
		// Yellow foreground
		fg, _, _ := screen.YellowStyle().Decompose()
		fgColor = fg
	case diff.DiffTypeItemDetail:
		// Use header color for better visibility (same as summary)
		fg, _, _ := screen.HeaderStyle().Decompose()
		fgColor = fg
		_, _, attrs = screen.HeaderStyle().Decompose()
	case diff.DiffTypeSummary:
		// Header style foreground
		fg, _, _ := screen.HeaderStyle().Decompose()
		fgColor = fg
		_, _, attrs = screen.HeaderStyle().Decompose()
	case diff.DiffTypeBlank:
		return baseStyle
	default:
		return baseStyle
	}

	// Get background from base style and apply only foreground color
	_, bg, attrs2 := baseStyle.Decompose()
	if attrs == 0 {
		attrs = attrs2
	}
	style := tcell.StyleDefault.Foreground(fgColor).Background(bg)
	// Apply attributes if present
	if attrs&tcell.AttrBold != 0 {
		style = style.Bold(true)
	}
	if attrs&tcell.AttrUnderline != 0 {
		style = style.Underline(true)
	}
	if attrs&tcell.AttrDim != 0 {
		style = style.Dim(true)
	}
	if attrs&tcell.AttrBlink != 0 {
		style = style.Blink(true)
	}
	if attrs&tcell.AttrReverse != 0 {
		style = style.Reverse(true)
	}
	return style
}
