package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

type NodeSearchWidget struct {
	visible     bool
	query       string
	allItems    []*model.Item
	matches     []*model.Item
	selectedIdx int
	cursorPos   int
	maxResults  int
	onSelect    func(*model.Item)
	onHoist     func(*model.Item)
}

func NewNodeSearchWidget() *NodeSearchWidget {
	w := &NodeSearchWidget{
		visible:     false,
		query:       "",
		selectedIdx: 0,
		cursorPos:   0,
		maxResults:  10,
	}
	return w
}

func (w *NodeSearchWidget) SetItems(items []*model.Item) {
	w.allItems = items
	w.updateMatches()
}

func (w *NodeSearchWidget) SetOnSelect(onSelect func(*model.Item)) {
	w.onSelect = onSelect
}

func (w *NodeSearchWidget) SetOnHoist(onHoist func(*model.Item)) {
	w.onHoist = onHoist
}

func (w *NodeSearchWidget) Show() {
	w.visible = true
	w.query = ""
	w.cursorPos = 0
	w.selectedIdx = 0
	w.updateMatches()
}

func (w *NodeSearchWidget) Hide() {
	w.visible = false
}

func (w *NodeSearchWidget) IsVisible() bool {
	return w.visible
}

// updateMatches performs fuzzy search on all items and updates matches
func (w *NodeSearchWidget) updateMatches() {
	w.matches = nil
	w.selectedIdx = 0

	if w.query == "" {
		return
	}

	// Collect all item texts and map them to items
	type itemWithText struct {
		text string
		item *model.Item
	}

	var candidates []itemWithText
	for _, item := range w.allItems {
		candidates = append(candidates, itemWithText{
			text: item.Text,
			item: item,
		})
	}

	// Use fuzzy ranking to get sorted matches (case-insensitive)
	query := w.query
	var matchedItems []itemWithText

	// Use RankFindFold for case-insensitive scoring and sorting
	rankCandidates := make([]string, len(candidates))
	for i, c := range candidates {
		rankCandidates[i] = c.text
	}

	ranks := fuzzy.RankFindFold(query, rankCandidates)

	// Build matches list from ranked results
	for _, rank := range ranks {
		for _, candidate := range candidates {
			if candidate.text == rank.Target {
				matchedItems = append(matchedItems, candidate)
				break
			}
		}
	}

	// Limit to maxResults
	if len(matchedItems) > w.maxResults {
		matchedItems = matchedItems[:w.maxResults]
	}

	// Extract just the items
	for _, mi := range matchedItems {
		w.matches = append(w.matches, mi.item)
	}
}

// DeleteWordBackwards deletes the word before the cursor
func (w *NodeSearchWidget) DeleteWordBackwards() {
	if w.cursorPos == 0 {
		return
	}

	// Start from cursor position and move backwards
	pos := w.cursorPos - 1

	// Skip any trailing whitespace
	for pos >= 0 && (w.query[pos] == ' ' || w.query[pos] == '\t') {
		pos--
	}

	// Skip the word characters
	for pos >= 0 && w.query[pos] != ' ' && w.query[pos] != '\t' {
		pos--
	}

	// Delete from pos+1 to cursorPos
	deleteStart := pos + 1
	w.query = w.query[:deleteStart] + w.query[w.cursorPos:]
	w.cursorPos = deleteStart
}

func (w *NodeSearchWidget) HandleKeyEvent(ev *tcell.EventKey) bool {
	if !w.visible {
		return false
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		w.Hide()
		return true

	case tcell.KeyEnter:
		if ev.Modifiers()&tcell.ModAlt != 0 {
			// Alt+Enter: Hoist the current match
			if len(w.matches) > 0 && w.selectedIdx < len(w.matches) {
				selected := w.matches[w.selectedIdx]
				w.Hide()
				if w.onHoist != nil {
					w.onHoist(selected)
				}
			}
		} else {
			// Enter: Select the current match
			if len(w.matches) > 0 && w.selectedIdx < len(w.matches) {
				selected := w.matches[w.selectedIdx]
				w.Hide()
				if w.onSelect != nil {
					w.onSelect(selected)
				}
			}
		}
		return true

	case tcell.KeyCtrlW:
		w.DeleteWordBackwards()
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
		if ch > 0 && ch < 127 && ch != 27 { // Printable ASCII (excluding Escape)
			w.query = w.query[:w.cursorPos] + string(ch) + w.query[w.cursorPos:]
			w.cursorPos++
			w.updateMatches()
			return true
		}
	}

	return false
}

func (w *NodeSearchWidget) Render(screen *Screen) {
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

	// Box height: 1 header + 1 input + 10 results + 1 footer = 13 lines
	boxHeight := 13
	boxStartY := (height - boxHeight) / 2

	// Styles - use theme-aware styles matching the tree view
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
	title := "Search Nodes"
	titleX := boxStartX + 2
	screen.DrawStringLimited(titleX, titleY, title, boxWidth-4, borderStyle)

	// Draw search input
	inputY := boxStartY + 2
	inputX := boxStartX + 2
	inputWidth := boxWidth - 4

	screen.DrawString(inputX, inputY, "Query: ", bgStyle)

	// Draw query with cursor
	queryX := inputX + 7
	queryDisplay := w.query
	if len(queryDisplay) > inputWidth-8 {
		// Truncate from the left side for long queries
		queryDisplay = queryDisplay[len(queryDisplay)-(inputWidth-8):]
		cursorPosDisplay := inputWidth - 8 - 1
		screen.DrawString(queryX, inputY, queryDisplay, bgStyle)
		if cursorPosDisplay >= 0 && cursorPosDisplay < len(queryDisplay) {
			cursorStyle := bgStyle.Reverse(true)
			screen.SetCell(queryX+cursorPosDisplay, inputY, rune(queryDisplay[cursorPosDisplay]), cursorStyle)
		}
	} else {
		screen.DrawStringLimited(queryX, inputY, queryDisplay, inputWidth-8, bgStyle)
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

	// Draw results
	resultsY := boxStartY + 3
	maxDisplayResults := 9
	canHoist := true
	for i := 0; i < len(w.matches) && i < maxDisplayResults; i++ {
		resultY := resultsY + i
		if resultY >= boxStartY+boxHeight-2 {
			break
		}

		item := w.matches[i]
		isSelected := i == w.selectedIdx

		if isSelected && len(item.Children) == 0 {
			canHoist = false
		}

		// Format the result line
		var resultLine string
		if isSelected {
			resultLine = " > " + item.Text
		} else {
			resultLine = "   " + item.Text
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
	}

	// Draw footer with match count
	footerY := boxStartY + boxHeight - 2
	matchCount := len(w.matches)
	totalCount := len(w.allItems)
	hoistAction := ""
	if canHoist {
		hoistAction = ", Alt+Enter: hoist"
	}
	footer := fmt.Sprintf(" %d of %d matches | Enter: select%s, Esc: close", matchCount, totalCount, hoistAction)
	if len(footer) > inputWidth {
		footer = footer[:inputWidth]
	}
	screen.DrawStringLimited(inputX, footerY, footer, inputWidth, borderStyle)
}
