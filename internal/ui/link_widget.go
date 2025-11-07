package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/search"
)

// LinkAutocompleteWidget provides inline autocomplete for [[links]] while editing
type LinkAutocompleteWidget struct {
	visible     bool
	query       string
	allItems    []*model.Item
	matches     []*model.Item
	selectedIdx int
	cursorPos   int
	maxResults  int
	parseError  string             // Error from parsing advanced search query
	filterExpr  search.FilterExpr  // Parsed filter expression
	onSelect    func(*model.Item)  // Called when link is selected, should insert [[id|text]]
}

func NewLinkAutocompleteWidget() *LinkAutocompleteWidget {
	w := &LinkAutocompleteWidget{
		visible:     false,
		query:       "",
		selectedIdx: 0,
		cursorPos:   0,
		maxResults:  10,
	}
	return w
}

func (w *LinkAutocompleteWidget) SetItems(items []*model.Item) {
	w.allItems = items
	w.updateMatches()
}

func (w *LinkAutocompleteWidget) SetOnSelect(onSelect func(*model.Item)) {
	w.onSelect = onSelect
}

// Show opens the widget with empty query, triggered when user types [[
func (w *LinkAutocompleteWidget) Show() {
	w.visible = true
	w.query = ""
	w.cursorPos = 0
	w.selectedIdx = 0
	w.updateMatches()
}

func (w *LinkAutocompleteWidget) Hide() {
	w.visible = false
}

func (w *LinkAutocompleteWidget) IsVisible() bool {
	return w.visible
}

func (w *LinkAutocompleteWidget) GetParseError() string {
	return w.parseError
}

// updateMatches filters items by advanced search or text search on query
func (w *LinkAutocompleteWidget) updateMatches() {
	w.matches = nil
	w.selectedIdx = 0
	w.parseError = ""
	w.filterExpr = nil

	if w.query == "" {
		// Show all items when query is empty
		w.matches = w.allItems
		if len(w.matches) > w.maxResults {
			w.matches = w.matches[:w.maxResults]
		}
		return
	}

	// Try to parse as advanced search query
	expr, err := search.ParseQuery(w.query)
	if err == nil {
		// Successfully parsed - use the filter expression
		w.filterExpr = expr
		for _, item := range w.allItems {
			if expr.Matches(item) {
				w.matches = append(w.matches, item)
				if len(w.matches) >= w.maxResults {
					break
				}
			}
		}
		return
	}

	// Parse error - fall back to text-only matching
	w.parseError = err.Error()
	queryLower := strings.ToLower(w.query)
	for _, item := range w.allItems {
		if strings.Contains(strings.ToLower(item.Text), queryLower) {
			w.matches = append(w.matches, item)
			if len(w.matches) >= w.maxResults {
				break
			}
		}
	}
}

// UpdateQuery updates the search query (called while user types)
func (w *LinkAutocompleteWidget) UpdateQuery(newQuery string) {
	w.query = newQuery
	w.cursorPos = len(newQuery)
	w.updateMatches()
}

// deleteWordBackward deletes from cursor position back to the start of the previous word
func (w *LinkAutocompleteWidget) deleteWordBackward() {
	if w.cursorPos == 0 {
		return
	}

	// Start from cursor position and move backward
	endPos := w.cursorPos

	// Skip any trailing whitespace
	for w.cursorPos > 0 && (w.query[w.cursorPos-1] == ' ' || w.query[w.cursorPos-1] == '\t') {
		w.cursorPos--
	}

	// Delete until we hit whitespace or start of string
	for w.cursorPos > 0 && w.query[w.cursorPos-1] != ' ' && w.query[w.cursorPos-1] != '\t' {
		w.cursorPos--
	}

	// Delete the word
	w.query = w.query[:w.cursorPos] + w.query[endPos:]
}

func (w *LinkAutocompleteWidget) HandleKeyEvent(ev *tcell.EventKey) bool {
	if !w.visible {
		return false
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		w.Hide()
		return true

	case tcell.KeyEnter:
		// Select the current match and insert link
		if len(w.matches) > 0 && w.selectedIdx < len(w.matches) {
			selected := w.matches[w.selectedIdx]
			w.Hide()
			if w.onSelect != nil {
				w.onSelect(selected)
			}
		}
		return true

	case tcell.KeyCtrlN:
		// Ctrl+N: Move down in results
		if len(w.matches) > 0 {
			w.selectedIdx++
			if w.selectedIdx >= len(w.matches) {
				w.selectedIdx = 0 // Wrap to top
			}
		}
		return true

	case tcell.KeyCtrlP:
		// Ctrl+P: Move up in results
		if len(w.matches) > 0 {
			w.selectedIdx--
			if w.selectedIdx < 0 {
				w.selectedIdx = len(w.matches) - 1 // Wrap to bottom
			}
		}
		return true

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Backspace: Delete character
		if w.cursorPos > 0 {
			w.query = w.query[:w.cursorPos-1] + w.query[w.cursorPos:]
			w.cursorPos--
			w.updateMatches()
		}
		return true

	case tcell.KeyDelete:
		// Delete: Delete character at cursor
		if w.cursorPos < len(w.query) {
			w.query = w.query[:w.cursorPos] + w.query[w.cursorPos+1:]
			w.updateMatches()
		}
		return true

	case tcell.KeyCtrlW:
		// Ctrl+W: Delete word backward
		w.deleteWordBackward()
		w.updateMatches()
		return true

	case tcell.KeyLeft:
		// Move cursor left
		if w.cursorPos > 0 {
			w.cursorPos--
		}
		return true

	case tcell.KeyRight:
		// Move cursor right
		if w.cursorPos < len(w.query) {
			w.cursorPos++
		}
		return true

	case tcell.KeyHome:
		// Move cursor to start
		w.cursorPos = 0
		return true

	case tcell.KeyEnd:
		// Move cursor to end
		w.cursorPos = len(w.query)
		return true

	default:
		// Regular character input
		ch := ev.Rune()
		if ch > 0 && ch != 27 { // Accept all valid Unicode characters (excluding Escape)
			w.query = w.query[:w.cursorPos] + string(ch) + w.query[w.cursorPos:]
			w.cursorPos++
			w.updateMatches()
			return true
		}
	}

	return false
}

func (w *LinkAutocompleteWidget) Render(screen *Screen) {
	if !w.visible {
		return
	}

	width := screen.GetWidth()
	height := screen.GetHeight()

	// Modal box dimensions
	boxWidth := (width * 2) / 3
	if boxWidth > width {
		boxWidth = width - 4
	}
	boxStartX := (width - boxWidth) / 2

	// Box height: 1 header + 1 input + 8 results + 1 footer = 11 lines
	boxHeight := 11
	boxStartY := (height - boxHeight) / 2

	// Styles
	borderStyle := screen.TreeNormalStyle()
	bgStyle := screen.BackgroundStyle()
	selectedStyle := screen.TreeSelectedStyle()

	// Draw the box background
	for y := boxStartY; y < boxStartY+boxHeight && y < height; y++ {
		for x := boxStartX; x < boxStartX+boxWidth && x < width; x++ {
			screen.SetCell(x, y, ' ', bgStyle)
		}
	}

	// Draw border
	// Top and bottom
	for x := boxStartX; x < boxStartX+boxWidth && x < width; x++ {
		if boxStartY >= 0 {
			screen.SetCell(x, boxStartY, '─', borderStyle)
		}
		if boxStartY+boxHeight-1 < height {
			screen.SetCell(x, boxStartY+boxHeight-1, '─', borderStyle)
		}
	}

	// Left and right
	for y := boxStartY; y < boxStartY+boxHeight && y < height; y++ {
		if boxStartX >= 0 {
			screen.SetCell(boxStartX, y, '│', borderStyle)
		}
		if boxStartX+boxWidth-1 < width {
			screen.SetCell(boxStartX+boxWidth-1, y, '│', borderStyle)
		}
	}

	// Corners
	if boxStartY >= 0 && boxStartX >= 0 {
		screen.SetCell(boxStartX, boxStartY, '┌', borderStyle)
	}
	if boxStartY >= 0 && boxStartX+boxWidth-1 < width {
		screen.SetCell(boxStartX+boxWidth-1, boxStartY, '┐', borderStyle)
	}
	if boxStartY+boxHeight-1 < height && boxStartX >= 0 {
		screen.SetCell(boxStartX, boxStartY+boxHeight-1, '└', borderStyle)
	}
	if boxStartY+boxHeight-1 < height && boxStartX+boxWidth-1 < width {
		screen.SetCell(boxStartX+boxWidth-1, boxStartY+boxHeight-1, '┘', borderStyle)
	}

	// Draw title
	titleY := boxStartY + 1
	title := "Insert Link"
	titleX := boxStartX + 2
	screen.DrawStringLimited(titleX, titleY, title, boxWidth-4, borderStyle)

	// Draw search input
	inputY := boxStartY + 2
	inputX := boxStartX + 2
	inputWidth := boxWidth - 4

	screen.DrawString(inputX, inputY, "Find: ", bgStyle)

	// Draw query with cursor
	queryX := inputX + 6
	queryDisplay := w.query
	if len(queryDisplay) > inputWidth-7 {
		// Truncate from the left side for long queries
		queryDisplay = queryDisplay[len(queryDisplay)-(inputWidth-7):]
		cursorPosDisplay := inputWidth - 7 - 1
		screen.DrawString(queryX, inputY, queryDisplay, bgStyle)
		if cursorPosDisplay >= 0 && cursorPosDisplay < len(queryDisplay) {
			cursorStyle := bgStyle.Reverse(true)
			screen.SetCell(queryX+cursorPosDisplay, inputY, rune(queryDisplay[cursorPosDisplay]), cursorStyle)
		}
	} else {
		screen.DrawStringLimited(queryX, inputY, queryDisplay, inputWidth-7, bgStyle)
		// Draw cursor
		if w.cursorPos <= len(w.query) {
			cursorStyle := bgStyle.Reverse(true)
			if w.cursorPos == len(w.query) {
				// Cursor at end - show as empty box
				screen.SetCell(queryX+w.cursorPos, inputY, ' ', cursorStyle)
			} else {
				screen.SetCell(queryX+w.cursorPos, inputY, rune(w.query[w.cursorPos]), cursorStyle)
			}
		}
	}

	// Draw error message if parse error exists
	resultsStartY := boxStartY + 3
	if w.parseError != "" {
		errorMsg := "Parse error: " + w.parseError
		if len(errorMsg) > inputWidth {
			errorMsg = errorMsg[:inputWidth]
		}
		errorStyle := borderStyle.Foreground(tcell.ColorRed)
		screen.DrawStringLimited(inputX, resultsStartY, errorMsg, inputWidth, errorStyle)
		resultsStartY++ // Offset results below the error
	}

	// Draw results
	maxDisplayResults := 7
	resultCount := 0
	for i := 0; i < len(w.matches) && i < maxDisplayResults && resultCount < maxDisplayResults; i++ {
		resultY := resultsStartY + resultCount
		if resultY >= boxStartY+boxHeight-2 {
			break
		}

		item := w.matches[i]
		isSelected := i == w.selectedIdx

		// Format the result line: show first 60 chars of text
		resultText := item.Text
		if len(resultText) > 60 {
			resultText = resultText[:60] + "…"
		}

		var resultLine string
		if isSelected {
			resultLine = " > " + resultText
		} else {
			resultLine = "   " + resultText
		}

		// Truncate if too long
		if len(resultLine) > inputWidth {
			resultLine = resultLine[:inputWidth]
		}

		resultStyle := bgStyle
		if isSelected {
			resultStyle = selectedStyle
		}

		screen.DrawStringLimited(inputX, resultY, resultLine, inputWidth, resultStyle)
		resultCount++
	}

	// Draw footer with match count
	footerY := boxStartY + boxHeight - 2
	matchCount := len(w.matches)
	totalCount := len(w.allItems)
	footer := fmt.Sprintf(" %d of %d items | Enter: link, Esc: cancel", matchCount, totalCount)
	if len(footer) > inputWidth {
		footer = footer[:inputWidth]
	}
	screen.DrawStringLimited(inputX, footerY, footer, inputWidth, borderStyle)
}
