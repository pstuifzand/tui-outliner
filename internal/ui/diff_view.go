package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/diff"
)

// DiffViewWidget displays a formatted diff between two outlines
type DiffViewWidget struct {
	visible      bool
	diffResult   *diff.DiffResult
	lines        []diff.DiffLine
	scrollOffset int
	maxWidth     int
	maxHeight    int

	file1Path string
	file1Name string
	file2Path string
	file2Name string
}

// NewDiffViewWidget creates a new diff view widget
func NewDiffViewWidget() *DiffViewWidget {
	return &DiffViewWidget{
		visible: false,
		lines:   make([]diff.DiffLine, 0),
	}
}

// Show displays the diff view with the given result
func (dv *DiffViewWidget) Show(result *diff.DiffResult, file1, file2 string) {
	dv.diffResult = result
	dv.file1Path = file1
	dv.file2Path = file2
	dv.file1Name = extractFileName(file1)
	dv.file2Name = extractFileName(file2)
	dv.scrollOffset = 0

	// Build display lines
	dv.lines = diff.BuildDiffLines(result, false)

	dv.visible = true
}

// Hide closes the diff view
func (dv *DiffViewWidget) Hide() {
	dv.visible = false
}

// IsVisible returns whether the widget is currently visible
func (dv *DiffViewWidget) IsVisible() bool {
	return dv.visible
}

// SetMaxSize updates the maximum dimensions for rendering
func (dv *DiffViewWidget) SetMaxSize(width, height int) {
	dv.maxWidth = width
	dv.maxHeight = height
}

// HandleKeyEvent processes keyboard input
func (dv *DiffViewWidget) HandleKeyEvent(ev *tcell.EventKey) {
	if !dv.visible {
		return
	}

	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		dv.Hide()
	case tcell.KeyUp, tcell.KeyCtrlK:
		dv.scroll(-1)
	case tcell.KeyDown, tcell.KeyCtrlJ:
		dv.scroll(1)
	case tcell.KeyPgUp, tcell.KeyCtrlU:
		dv.scroll(-dv.maxHeight / 2)
	case tcell.KeyPgDn, tcell.KeyCtrlD:
		dv.scroll(dv.maxHeight / 2)
	case tcell.KeyHome:
		dv.scrollOffset = 0
	case tcell.KeyEnd:
		maxScroll := len(dv.lines) - dv.maxHeight + 4 // Account for header/footer
		if maxScroll < 0 {
			maxScroll = 0
		}
		dv.scrollOffset = maxScroll
	default:
		// Handle 'q' to close
		if ev.Rune() == 'q' {
			dv.Hide()
		}
	}
}

// scroll moves the view up or down
func (dv *DiffViewWidget) scroll(lines int) {
	newOffset := dv.scrollOffset + lines
	maxScroll := len(dv.lines) - dv.maxHeight + 4

	if maxScroll < 0 {
		maxScroll = 0
	}

	if newOffset < 0 {
		newOffset = 0
	} else if newOffset > maxScroll {
		newOffset = maxScroll
	}

	dv.scrollOffset = newOffset
}

// Render draws the diff view on the screen
func (dv *DiffViewWidget) Render(screen *Screen) {
	if !dv.visible {
		return
	}

	width := screen.GetWidth()
	height := screen.GetHeight()
	dv.SetMaxSize(width, height)

	// Calculate dimensions
	boxWidth := width - 4
	boxHeight := height - 4
	startX := 2
	startY := 2

	if boxWidth < 20 || boxHeight < 5 {
		return // Too small to render
	}

	// Draw border
	drawBox(screen, startX, startY, boxWidth, boxHeight, screen.TreeNormalStyle())

	// Draw header
	headerStyle := screen.TreeSelectedStyle()
	headerText := fmt.Sprintf(" Diff: %s → %s ", dv.file1Name, dv.file2Name)
	if len(headerText) > boxWidth-2 {
		headerText = headerText[:boxWidth-4] + " "
	}
	screen.DrawStringLimited(startX+1, startY, headerText, boxWidth-2, headerStyle)

	// Draw content
	contentStartY := startY + 2
	contentHeight := boxHeight - 4

	dv.renderContent(screen, startX+1, contentStartY, boxWidth-2, contentHeight)

	// Draw footer
	footerStyle := screen.TreeNormalStyle()
	footerText := "j/k/↓/↑: scroll | Ctrl+U/D: page | q/Esc: close"
	if len(footerText) > boxWidth-2 {
		footerText = footerText[:boxWidth-4] + " "
	}
	screen.DrawStringLimited(startX+1, startY+boxHeight-1, footerText, boxWidth-2, footerStyle)
}

// renderContent draws the diff lines in the content area
func (dv *DiffViewWidget) renderContent(screen *Screen, x, y, width, height int) {
	displayLines := height
	if displayLines > len(dv.lines) {
		displayLines = len(dv.lines)
	}

	endOffset := dv.scrollOffset + displayLines
	if endOffset > len(dv.lines) {
		endOffset = len(dv.lines)
	}

	for i := dv.scrollOffset; i < endOffset; i++ {
		lineY := y + (i - dv.scrollOffset)
		if lineY >= y+height {
			break
		}

		dv.renderLine(screen, x, lineY, width, dv.lines[i])
	}

	// Show scrollbar indicator if needed
	if len(dv.lines) > height {
		scrollbarY := y + (dv.scrollOffset * height / len(dv.lines))
		scrollbarStyle := screen.TreeSelectedStyle()
		screen.SetCell(x+width-1, scrollbarY, '█', scrollbarStyle)
	}
}

// renderLine draws a single diff line
func (dv *DiffViewWidget) renderLine(screen *Screen, x, y, width int, line diff.DiffLine) {
	content := line.Content
	indent := line.Indent

	// Apply indentation
	indentStr := strings.Repeat("  ", indent)
	fullContent := indentStr + content

	// Determine style based on line type
	style := dv.getStyleForLineType(screen, line.Type)

	// Draw the line, truncating if necessary
	displayContent := fullContent
	if len(displayContent) > width {
		displayContent = displayContent[:width-3] + "..."
	}

	screen.DrawStringLimited(x, y, displayContent, width, style)
}

// getStyleForLineType returns the appropriate style for a diff line type
func (dv *DiffViewWidget) getStyleForLineType(screen *Screen, lineType diff.DiffLineType) tcell.Style {
	switch lineType {
	case diff.DiffTypeHeader:
		return screen.HeaderStyle()
	case diff.DiffTypeNewSection:
		return screen.GreenStyle()
	case diff.DiffTypeDeletedSection:
		return screen.RedStyle()
	case diff.DiffTypeModifiedSection:
		return screen.YellowStyle()
	case diff.DiffTypeNewItem:
		return screen.GreenStyle()
	case diff.DiffTypeDeletedItem:
		return screen.RedStyle()
	case diff.DiffTypeModifiedItem:
		return screen.YellowStyle()
	case diff.DiffTypeItemDetail:
		return screen.GrayStyle()
	case diff.DiffTypeSummary:
		return screen.HeaderStyle()
	case diff.DiffTypeBlank:
		return screen.TreeNormalStyle()
	default:
		return screen.TreeNormalStyle()
	}
}

// extractFileName extracts the filename from a path
func extractFileName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// drawBox draws a simple box border
func drawBox(screen *Screen, x, y, width, height int, style tcell.Style) {
	// Top border
	screen.SetCell(x, y, '┌', style)
	for i := 1; i < width-1; i++ {
		screen.SetCell(x+i, y, '─', style)
	}
	screen.SetCell(x+width-1, y, '┐', style)

	// Bottom border
	screen.SetCell(x, y+height-1, '└', style)
	for i := 1; i < width-1; i++ {
		screen.SetCell(x+i, y+height-1, '─', style)
	}
	screen.SetCell(x+width-1, y+height-1, '┘', style)

	// Side borders
	for i := 1; i < height-1; i++ {
		screen.SetCell(x, y+i, '│', style)
		screen.SetCell(x+width-1, y+i, '│', style)
	}
}
