package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// editorState represents a single undo/redo state
type editorState struct {
	text      string
	cursorPos int
}

// MultiLineEditor manages multi-line text editing with word wrapping
type MultiLineEditor struct {
	item                      *model.Item
	text                      string
	cursorPos                 int // Absolute position in text
	active                    bool
	enterPressed              bool // Plain Enter - create new node
	escapePressed             bool
	backspaceOnEmpty          bool
	indentPressed             bool
	outdentPressed            bool
	maxWidth                  int // Maximum width for text wrapping
	wrappedLines              []string
	lineStartOffsets          []int // Starting text offset for each wrapped line
	undoStack                 []editorState
	redoStack                 []editorState
	maxUndoLevels             int // Maximum undo history levels
	linkAutocompleteTriggered bool // [[ was typed - app should show link widget
	linkAutocompleteStartPos  int  // Position where [[ started
}

// NewMultiLineEditor creates a new MultiLineEditor
func NewMultiLineEditor(item *model.Item) *MultiLineEditor {
	mle := &MultiLineEditor{
		item:          item,
		text:          item.Text,
		cursorPos:     len(item.Text),
		active:        false,
		maxWidth:      80, // Default width
		maxUndoLevels: 50, // Default undo history levels
	}
	mle.calculateWrappedLines()
	return mle
}

// SetMaxWidth sets the maximum width for wrapping and recalculates wrapped lines
func (mle *MultiLineEditor) SetMaxWidth(width int) {
	if width < 0 {
		width = 0
	}
	if mle.maxWidth != width {
		mle.maxWidth = width
		mle.calculateWrappedLines()
	}
}

// calculateWrappedLines splits text into wrapped lines and tracks their offsets
func (mle *MultiLineEditor) calculateWrappedLines() {
	if mle.maxWidth <= 0 {
		// No wrapping, treat text as single line (with hard breaks)
		mle.wrappedLines = strings.Split(mle.text, "\n")
		mle.lineStartOffsets = make([]int, len(mle.wrappedLines))
		offset := 0
		for i := range mle.wrappedLines {
			mle.lineStartOffsets[i] = offset
			offset += len(mle.wrappedLines[i]) + 1 // +1 for \n
		}
		return
	}

	// Split by hard newlines first
	hardLines := strings.Split(mle.text, "\n")
	mle.wrappedLines = []string{}
	mle.lineStartOffsets = []int{}

	offset := 0
	for _, hardLine := range hardLines {
		// Apply word wrapping to each hard line
		wrappedParts := wrapTextAtWidth(hardLine, mle.maxWidth)

		// Track position within the hard line to account for skipped spaces
		searchPos := 0
		for _, part := range wrappedParts {
			// Find where this part is in the hard line starting from searchPos
			partIdx := strings.Index(hardLine[searchPos:], part)
			if partIdx >= 0 {
				searchPos += partIdx
			}

			// The offset for this line is offset + searchPos (position in hard line)
			mle.lineStartOffsets = append(mle.lineStartOffsets, offset+searchPos)
			mle.wrappedLines = append(mle.wrappedLines, part)

			// Move searchPos past this part
			searchPos += len(part)

			// Skip spaces for the next part
			for searchPos < len(hardLine) && hardLine[searchPos] == ' ' {
				searchPos++
			}
		}
		offset += len(hardLine) + 1 // Account for the hard line length + the \n
	}
}

// getCursorVisualPosition returns the visual (row, col) position of the cursor
// Returns (-1, -1) if cursor is out of bounds
func (mle *MultiLineEditor) getCursorVisualPosition() (row int, col int) {
	if mle.cursorPos < 0 || mle.cursorPos > len(mle.text) {
		return -1, -1
	}

	// If no wrapping, easier calculation
	if mle.maxWidth <= 0 {
		lines := strings.Split(mle.text, "\n")
		offset := 0
		for i, line := range lines {
			lineEnd := offset + len(line)
			if mle.cursorPos <= lineEnd {
				return i, mle.cursorPos - offset
			}
			offset = lineEnd + 1 // +1 for \n
		}
		return len(lines) - 1, len(lines[len(lines)-1])
	}

	// With wrapping: find which wrapped line contains cursor
	for lineIdx, startOffset := range mle.lineStartOffsets {
		lineEnd := startOffset + len(mle.wrappedLines[lineIdx])
		if lineIdx < len(mle.lineStartOffsets)-1 {
			nextLineStart := mle.lineStartOffsets[lineIdx+1]
			// Check if cursor is in this line or in the gap (newline)
			if mle.cursorPos >= startOffset && mle.cursorPos <= lineEnd {
				return lineIdx, mle.cursorPos - startOffset
			}
			// Check if cursor is on the newline after this line
			if mle.cursorPos > lineEnd && mle.cursorPos < nextLineStart {
				return lineIdx, len(mle.wrappedLines[lineIdx])
			}
		} else {
			// Last line
			if mle.cursorPos >= startOffset {
				return lineIdx, mle.cursorPos - startOffset
			}
		}
	}

	// Cursor at end
	if len(mle.wrappedLines) > 0 {
		lastIdx := len(mle.wrappedLines) - 1
		return lastIdx, len(mle.wrappedLines[lastIdx])
	}
	return 0, 0
}

// getCursorTextOffset converts visual position to text offset
// If invalid position, returns -1
func (mle *MultiLineEditor) getCursorTextOffset(row int, col int) int {
	if row < 0 || row >= len(mle.wrappedLines) {
		return -1
	}
	if col < 0 || col > len(mle.wrappedLines[row]) {
		return -1
	}

	return mle.lineStartOffsets[row] + col
}

// Start starts editing mode
func (mle *MultiLineEditor) Start() {
	mle.active = true
	mle.cursorPos = len(mle.text)
	mle.calculateWrappedLines()
}

// Stop stops editing mode and returns the final text
func (mle *MultiLineEditor) Stop() string {
	mle.active = false
	mle.item.Text = mle.text
	return mle.text
}

// Cancel cancels editing and discards changes
func (mle *MultiLineEditor) Cancel() string {
	mle.active = false
	return mle.item.Text
}

// IsActive returns whether the editor is active
func (mle *MultiLineEditor) IsActive() bool {
	return mle.active
}

// HandleKey handles a key press during editing
func (mle *MultiLineEditor) HandleKey(ev *tcell.EventKey) bool {
	if !mle.active {
		return false
	}

	ch := ev.Rune()
	key := ev.Key()

	// Check for Ctrl+; using key code 256
	if key == 256 && ch == ';' {
		// Ctrl+; - Insert current time
		mle.InsertCurrentTime()
		return true
	}

	switch key {
	case tcell.KeyCtrlZ:
		// Undo
		mle.undo()
		return true
	case tcell.KeyCtrlY:
		// Redo
		mle.redo()
		return true
	case tcell.KeyCtrlW:
		// Delete word backwards
		mle.saveUndoState()
		mle.deleteWordBackwards()
		mle.calculateWrappedLines()
		return true
	case tcell.KeyEscape:
		mle.escapePressed = true
		return false
	case tcell.KeyEnter:
		// Check if Shift is held (Shift+Enter = newline, plain Enter = new node)
		if ev.Modifiers()&tcell.ModShift != 0 {
			// Shift+Enter - insert newline for multi-line text
			mle.saveUndoState()
			mle.text = mle.text[:mle.cursorPos] + "\n" + mle.text[mle.cursorPos:]
			mle.cursorPos++
			mle.calculateWrappedLines()
			return true
		}
		// Plain Enter - exit edit mode and create new node
		mle.enterPressed = true
		return false
	case tcell.KeyTab:
		mle.indentPressed = true
		return false
	case tcell.KeyBacktab:
		mle.outdentPressed = true
		return false
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if mle.cursorPos > 0 {
			mle.saveUndoState()
			mle.text = mle.text[:mle.cursorPos-1] + mle.text[mle.cursorPos:]
			mle.cursorPos--
			mle.calculateWrappedLines()
		} else if mle.cursorPos == 0 && mle.text == "" {
			mle.backspaceOnEmpty = true
			return false
		}
		return true
	case tcell.KeyDelete:
		// Check for Ctrl modifier (delete word forward)
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			mle.saveUndoState()
			mle.deleteWordForward()
		} else {
			if mle.cursorPos < len(mle.text) {
				mle.saveUndoState()
				mle.text = mle.text[:mle.cursorPos] + mle.text[mle.cursorPos+1:]
				mle.calculateWrappedLines()
			}
		}
		return true
	case tcell.KeyLeft:
		// Check for Ctrl modifier (word jump)
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			mle.jumpWordBackward()
		} else {
			if mle.cursorPos > 0 {
				mle.cursorPos--
			}
		}
		return true
	case tcell.KeyRight:
		// Check for Ctrl modifier (word jump)
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			mle.jumpWordForward()
		} else {
			if mle.cursorPos < len(mle.text) {
				mle.cursorPos++
			}
		}
		return true
	case tcell.KeyUp:
		// Move to previous wrapped line, same column if possible
		row, col := mle.getCursorVisualPosition()
		if row > 0 {
			newOffset := mle.getCursorTextOffset(row-1, col)
			if newOffset < 0 {
				// Column out of bounds on previous line, go to end of line
				newOffset = mle.getCursorTextOffset(row-1, len(mle.wrappedLines[row-1]))
			}
			mle.cursorPos = newOffset
		}
		return true
	case tcell.KeyDown:
		// Move to next wrapped line, same column if possible
		row, col := mle.getCursorVisualPosition()
		if row < len(mle.wrappedLines)-1 {
			newOffset := mle.getCursorTextOffset(row+1, col)
			if newOffset < 0 {
				// Column out of bounds on next line, go to end of line
				newOffset = mle.getCursorTextOffset(row+1, len(mle.wrappedLines[row+1]))
			}
			mle.cursorPos = newOffset
		}
		return true
	case tcell.KeyHome:
		// Go to start of current wrapped line
		row, _ := mle.getCursorVisualPosition()
		if row >= 0 {
			mle.cursorPos = mle.lineStartOffsets[row]
		}
		return true
	case tcell.KeyEnd:
		// Go to end of current wrapped line
		row, _ := mle.getCursorVisualPosition()
		if row >= 0 {
			mle.cursorPos = mle.lineStartOffsets[row] + len(mle.wrappedLines[row])
		}
		return true
	case tcell.KeyCtrlA:
		mle.cursorPos = 0
		return true
	case tcell.KeyCtrlE:
		mle.cursorPos = len(mle.text)
		return true
	case tcell.KeyCtrlU:
		// Delete from start to cursor
		mle.saveUndoState()
		mle.text = mle.text[mle.cursorPos:]
		mle.cursorPos = 0
		mle.calculateWrappedLines()
		return true
	case tcell.KeyCtrlK:
		// Delete from cursor to end
		mle.saveUndoState()
		mle.text = mle.text[:mle.cursorPos]
		mle.calculateWrappedLines()
		return true
	default:
		// Regular character input
		if ch > 0 { // Accept all valid Unicode characters
			mle.saveUndoState()
			s := string(ch)
			mle.text = mle.text[:mle.cursorPos] + s + mle.text[mle.cursorPos:]
			mle.cursorPos += len(s) // Increment by byte length, not character count

			// Check for [[ trigger for link autocomplete
			if ch == '[' && mle.cursorPos >= 2 && mle.text[mle.cursorPos-2] == '[' {
				mle.linkAutocompleteTriggered = true
				mle.linkAutocompleteStartPos = mle.cursorPos - 2 // Position of first [
				mle.calculateWrappedLines()
				return false // Return false to signal app to open link widget immediately
			}

			mle.calculateWrappedLines()
		}
		return true
	}
}

// Render renders the editor with multi-line support
func (mle *MultiLineEditor) Render(screen *Screen, x, y int, maxWidth int) {
	mle.SetMaxWidth(maxWidth)

	textStyle := screen.EditorStyle()
	cursorStyle := screen.EditorCursorStyle()

	screenWidth := screen.GetWidth()
	cursorRow, cursorCol := mle.getCursorVisualPosition()

	// Render each wrapped line
	for lineIdx, line := range mle.wrappedLines {
		screenY := y + lineIdx
		if screenY >= screen.GetHeight() {
			break
		}

		// Display the line with proper character width handling
		screenCol := 0
		for byteIdx, r := range line {
			if x+screenCol < screenWidth {
				charWidth := RuneWidth(r)
				charStyle := textStyle
				// Check if cursor is at this byte position
				if lineIdx == cursorRow && byteIdx == cursorCol {
					charStyle = cursorStyle
				}
				screen.SetCell(x+screenCol, screenY, r, charStyle)
				screenCol += charWidth
			}
		}

		// Calculate display width of line
		lineDisplayWidth := StringWidth(line)

		// Show cursor at end if cursor is at end of this line
		cursorAtEnd := lineIdx == cursorRow && cursorCol == len(line)
		if cursorAtEnd {
			if x+lineDisplayWidth < screenWidth {
				screen.SetCell(x+lineDisplayWidth, screenY, ' ', cursorStyle)
			}
		}

		// Clear remainder of line (skip cursor position if it's at the end)
		clearStart := lineDisplayWidth
		if cursorAtEnd {
			clearStart = lineDisplayWidth + 1
		}
		for i := clearStart; i < maxWidth; i++ {
			if x+i < screenWidth {
				screen.SetCell(x+i, screenY, ' ', textStyle)
			}
		}
	}
}

// GetItem returns the item being edited
func (mle *MultiLineEditor) GetItem() *model.Item {
	return mle.item
}

// GetText returns the current text
func (mle *MultiLineEditor) GetText() string {
	return mle.text
}

// SetText sets the text and recalculates wrapped lines
func (mle *MultiLineEditor) SetText(text string) {
	mle.text = text
	if mle.cursorPos > len(mle.text) {
		mle.cursorPos = len(mle.text)
	}
	mle.calculateWrappedLines()
}

// GetCursorPos returns the cursor position (text offset)
func (mle *MultiLineEditor) GetCursorPos() int {
	return mle.cursorPos
}

// GetWrappedLineCount returns the number of visual lines
func (mle *MultiLineEditor) GetWrappedLineCount() int {
	return len(mle.wrappedLines)
}

// GetCursorVisualRow returns the visual row of the cursor
func (mle *MultiLineEditor) GetCursorVisualRow() int {
	row, _ := mle.getCursorVisualPosition()
	return row
}

// SetCursorToStart positions the cursor at the beginning of the text
func (mle *MultiLineEditor) SetCursorToStart() {
	mle.cursorPos = 0
}

// SetCursorToEnd positions the cursor at the end of the text
func (mle *MultiLineEditor) SetCursorToEnd() {
	mle.cursorPos = len(mle.text)
}

// SetCursorFromScreenX sets the cursor position based on a screen X coordinate
// For multi-line, this is a simple version that places cursor at X position on current line
func (mle *MultiLineEditor) SetCursorFromScreenX(relativeX int) {
	if relativeX < 0 {
		relativeX = 0
	}
	if relativeX > len(mle.text) {
		relativeX = len(mle.text)
	}
	mle.cursorPos = relativeX
}

// WasEnterPressed returns whether Enter was pressed and resets the flag
func (mle *MultiLineEditor) WasEnterPressed() bool {
	pressed := mle.enterPressed
	mle.enterPressed = false
	return pressed
}

// WasEscapePressed returns whether Escape was pressed and resets the flag
func (mle *MultiLineEditor) WasEscapePressed() bool {
	pressed := mle.escapePressed
	mle.escapePressed = false
	return pressed
}

// WasBackspaceOnEmpty returns whether Backspace was pressed on an empty item and resets the flag
func (mle *MultiLineEditor) WasBackspaceOnEmpty() bool {
	pressed := mle.backspaceOnEmpty
	mle.backspaceOnEmpty = false
	return pressed
}

// WasIndentPressed returns whether Tab was pressed and resets the flag
func (mle *MultiLineEditor) WasIndentPressed() bool {
	pressed := mle.indentPressed
	mle.indentPressed = false
	return pressed
}

// WasOutdentPressed returns whether Shift+Tab was pressed and resets the flag
func (mle *MultiLineEditor) WasOutdentPressed() bool {
	pressed := mle.outdentPressed
	mle.outdentPressed = false
	return pressed
}

// WasLinkAutocompleteTriggered returns whether [[ was typed and resets the flag
func (mle *MultiLineEditor) WasLinkAutocompleteTriggered() bool {
	triggered := mle.linkAutocompleteTriggered
	mle.linkAutocompleteTriggered = false
	return triggered
}

// GetLinkAutocompleteStartPos returns the position where [[ was typed
func (mle *MultiLineEditor) GetLinkAutocompleteStartPos() int {
	return mle.linkAutocompleteStartPos
}

// InsertLink inserts a link in the format [[id|text]] at the autocomplete start position
// and removes the initial [[ characters
func (mle *MultiLineEditor) InsertLink(itemID string, itemText string) {
	if mle.linkAutocompleteStartPos < 0 || mle.linkAutocompleteStartPos > len(mle.text) {
		return
	}

	// Replace [[ with the full link
	linkStr := "[[" + itemID + "|" + itemText + "]]"
	mle.text = mle.text[:mle.linkAutocompleteStartPos] + linkStr + mle.text[mle.cursorPos:]
	mle.cursorPos = mle.linkAutocompleteStartPos + len(linkStr)
	mle.linkAutocompleteTriggered = false
	mle.calculateWrappedLines()
}

// CancelLinkAutocomplete resets the link autocomplete state (e.g., when widget is closed)
func (mle *MultiLineEditor) CancelLinkAutocomplete() {
	mle.linkAutocompleteTriggered = false
	mle.linkAutocompleteStartPos = -1
	// Leave the [[ in the text - user can continue editing or delete it manually
}

// deleteWordBackwards deletes the word before the cursor
func (mle *MultiLineEditor) deleteWordBackwards() {
	if mle.cursorPos == 0 {
		return
	}

	// Move back over any spaces
	pos := mle.cursorPos - 1
	for pos > 0 && mle.text[pos] == ' ' {
		pos--
	}

	// Move back over word characters
	for pos > 0 && mle.text[pos] != ' ' && mle.text[pos] != '\n' {
		pos--
	}

	// If we stopped on a space/newline, move forward one
	if pos > 0 && (mle.text[pos] == ' ' || mle.text[pos] == '\n') {
		pos++
	}

	// Delete from pos to cursor
	mle.text = mle.text[:pos] + mle.text[mle.cursorPos:]
	mle.cursorPos = pos
}

// jumpWordBackward moves cursor to the start of the previous word
func (mle *MultiLineEditor) jumpWordBackward() {
	if mle.cursorPos == 0 {
		return
	}

	pos := mle.cursorPos - 1

	// Skip back over spaces
	for pos > 0 && mle.text[pos] == ' ' {
		pos--
	}

	// Move back to start of word (stop at space or newline)
	for pos > 0 && mle.text[pos] != ' ' && mle.text[pos] != '\n' {
		pos--
	}

	// If we stopped on a space/newline, move forward one
	if pos > 0 && (mle.text[pos] == ' ' || mle.text[pos] == '\n') {
		pos++
	}

	mle.cursorPos = pos
}

// jumpWordForward moves cursor to the start of the next word
func (mle *MultiLineEditor) jumpWordForward() {
	if mle.cursorPos >= len(mle.text) {
		return
	}

	pos := mle.cursorPos

	// Skip current word characters
	for pos < len(mle.text) && mle.text[pos] != ' ' && mle.text[pos] != '\n' {
		pos++
	}

	// Skip spaces
	for pos < len(mle.text) && mle.text[pos] == ' ' {
		pos++
	}

	mle.cursorPos = pos
}

// deleteWordForward deletes the next word after the cursor
func (mle *MultiLineEditor) deleteWordForward() {
	if mle.cursorPos >= len(mle.text) {
		return
	}

	startPos := mle.cursorPos
	pos := mle.cursorPos

	// Skip current word characters
	for pos < len(mle.text) && mle.text[pos] != ' ' && mle.text[pos] != '\n' {
		pos++
	}

	// Skip spaces
	for pos < len(mle.text) && mle.text[pos] == ' ' {
		pos++
	}

	// Delete from startPos to pos
	mle.text = mle.text[:startPos] + mle.text[pos:]
	mle.calculateWrappedLines()
}

// saveUndoState saves current state to undo stack
func (mle *MultiLineEditor) saveUndoState() {
	state := editorState{
		text:      mle.text,
		cursorPos: mle.cursorPos,
	}
	mle.undoStack = append(mle.undoStack, state)

	// Limit undo stack size
	if len(mle.undoStack) > mle.maxUndoLevels {
		mle.undoStack = mle.undoStack[1:]
	}

	// Clear redo stack when a new edit is made
	mle.redoStack = []editorState{}
}

// undo reverts to the previous state
func (mle *MultiLineEditor) undo() {
	if len(mle.undoStack) == 0 {
		return
	}

	// Save current state to redo stack
	state := editorState{
		text:      mle.text,
		cursorPos: mle.cursorPos,
	}
	mle.redoStack = append(mle.redoStack, state)

	// Pop from undo stack
	lastIdx := len(mle.undoStack) - 1
	previousState := mle.undoStack[lastIdx]
	mle.undoStack = mle.undoStack[:lastIdx]

	// Restore state
	mle.text = previousState.text
	mle.cursorPos = previousState.cursorPos
	mle.calculateWrappedLines()
}

// redo restores the next state
func (mle *MultiLineEditor) redo() {
	if len(mle.redoStack) == 0 {
		return
	}

	// Save current state to undo stack
	state := editorState{
		text:      mle.text,
		cursorPos: mle.cursorPos,
	}
	mle.undoStack = append(mle.undoStack, state)

	// Pop from redo stack
	lastIdx := len(mle.redoStack) - 1
	nextState := mle.redoStack[lastIdx]
	mle.redoStack = mle.redoStack[:lastIdx]

	// Restore state
	mle.text = nextState.text
	mle.cursorPos = nextState.cursorPos
	mle.calculateWrappedLines()
}

// InsertCurrentDate inserts the current date at the cursor position (YYYY-MM-DD format)
func (mle *MultiLineEditor) InsertCurrentDate() {
	mle.saveUndoState()
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	mle.text = mle.text[:mle.cursorPos] + dateStr + mle.text[mle.cursorPos:]
	mle.cursorPos += len(dateStr)
	mle.calculateWrappedLines()
}

// InsertCurrentTime inserts the current time at the beginning with a space (HH:MM format)
func (mle *MultiLineEditor) InsertCurrentTime() {
	mle.saveUndoState()
	now := time.Now()
	timeStr := now.Format("15:04 ") // Add space after time
	// Always insert at the beginning
	mle.text = timeStr + mle.text
	mle.cursorPos += len(timeStr)
	mle.calculateWrappedLines()
}

// InsertCurrentDateTime inserts the current date and time at the cursor position (YYYY-MM-DD HH:MM:SS format)
func (mle *MultiLineEditor) InsertCurrentDateTime() {
	mle.saveUndoState()
	now := time.Now()
	dateTimeStr := now.Format("2006-01-02 15:04:05")
	mle.text = mle.text[:mle.cursorPos] + dateTimeStr + mle.text[mle.cursorPos:]
	mle.cursorPos += len(dateTimeStr)
	mle.calculateWrappedLines()
}
