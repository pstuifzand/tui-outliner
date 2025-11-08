package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// Editor manages inline text editing of outline items
type Editor struct {
	item             *model.Item
	text             string
	cursorPos        int
	active           bool
	enterPressed     bool // Track if Enter was pressed to create new node
	escapePressed    bool // Track if Escape was pressed
	backspaceOnEmpty bool // Track if Backspace was pressed on an empty item
	indentPressed    bool // Track if Tab was pressed to indent
	outdentPressed   bool // Track if Shift+Tab was pressed to outdent
}

// NewEditor creates a new Editor
func NewEditor(item *model.Item) *Editor {
	return &Editor{
		item:      item,
		text:      item.Text,
		cursorPos: len(item.Text),
		active:    false,
	}
}

// Start starts editing mode
func (e *Editor) Start() {
	e.active = true
	e.cursorPos = len(e.text)
}

// Stop stops editing mode and returns the final text
func (e *Editor) Stop() string {
	e.active = false
	e.item.Text = e.text
	return e.text
}

// Cancel cancels editing and discards changes
func (e *Editor) Cancel() string {
	e.active = false
	return e.item.Text
}

// IsActive returns whether the editor is active
func (e *Editor) IsActive() bool {
	return e.active
}

// HandleKey handles a key press during editing
func (e *Editor) HandleKey(ev *tcell.EventKey) bool {
	if !e.active {
		return false
	}

	ch := ev.Rune()
	key := ev.Key()

	// Check for Ctrl+; using key code 256 (which is what tcell sends for Ctrl+; in many terminals)
	// The terminal doesn't send the Ctrl modifier with ;, so we check the raw key code instead
	if key == 256 && ch == ';' {
		// Ctrl+; - Insert current time
		e.InsertCurrentTime()
		return true
	}

	switch key {
	case tcell.KeyCtrlW:
		// Check for Ctrl+W - delete word backwards
		e.DeleteWordBackwards()
		return true
	case tcell.KeyEscape:
		e.escapePressed = true
		return false // Signal to exit edit mode
	case tcell.KeyEnter:
		// Check if Shift is held (Shift+Enter = newline, plain Enter = new node)
		if ev.Modifiers()&tcell.ModShift != 0 {
			// Shift+Enter - insert newline for multi-line text
			e.text = e.text[:e.cursorPos] + "\n" + e.text[e.cursorPos:]
			e.cursorPos++
			return true
		}
		// Plain Enter - exit edit mode and create new node
		e.enterPressed = true
		return false // Signal to exit edit mode and create new node
	case tcell.KeyTab:
		// Tab pressed - indent the current item
		e.indentPressed = true
		return false // Signal to exit edit mode and perform indent
	case tcell.KeyBacktab:
		// Shift+Tab pressed (sent as KeyBacktab) - outdent the current item
		e.outdentPressed = true
		return false // Signal to exit edit mode and perform outdent
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.cursorPos > 0 {
			// Delete entire character (which may be multiple bytes for UTF-8)
			before := e.text[:e.cursorPos]
			runes := []rune(before)
			if len(runes) > 0 {
				runes = runes[:len(runes)-1] // Remove last rune/character
				newBefore := string(runes)
				deletedBytes := len(before) - len(newBefore)
				e.text = newBefore + e.text[e.cursorPos:]
				e.cursorPos -= deletedBytes
			}
		} else if e.cursorPos == 0 && e.text == "" {
			// Backspace pressed on empty item - signal to merge with previous item
			e.backspaceOnEmpty = true
			return false // Signal to exit edit mode
		}
	case tcell.KeyDelete:
		if e.cursorPos < len(e.text) {
			// Delete entire character (which may be multiple bytes for UTF-8)
			after := e.text[e.cursorPos:]
			runes := []rune(after)
			if len(runes) > 0 {
				// Calculate bytes to delete (the first rune)
				deletedBytes := len(string(runes[:1]))
				e.text = e.text[:e.cursorPos] + e.text[e.cursorPos+deletedBytes:]
			}
		}
	case tcell.KeyLeft:
		if e.cursorPos > 0 {
			e.cursorPos--
		}
	case tcell.KeyRight:
		if e.cursorPos < len(e.text) {
			e.cursorPos++
		}
	case tcell.KeyHome:
		e.cursorPos = 0
	case tcell.KeyEnd:
		e.cursorPos = len(e.text)
	case tcell.KeyCtrlA:
		e.cursorPos = 0
	case tcell.KeyCtrlE:
		e.cursorPos = len(e.text)
	case tcell.KeyCtrlU:
		// Delete from start to cursor
		e.text = e.text[e.cursorPos:]
		e.cursorPos = 0
	case tcell.KeyCtrlK:
		// Delete from cursor to end
		e.text = e.text[:e.cursorPos]
	default:
		// Regular character input
		if ch > 0 { // Accept all valid Unicode characters
			s := string(ch)
			e.text = e.text[:e.cursorPos] + s + e.text[e.cursorPos:]
			e.cursorPos += len(s) // Increment by byte length, not character count
		}
	}

	return true
}

// Render renders the editor on the screen
func (e *Editor) Render(screen *Screen, x, y int, maxWidth int) {
	textStyle := screen.EditorStyle()
	cursorStyle := screen.EditorCursorStyle()

	// Determine which portion of text to display, accounting for character widths
	displayText := e.text
	startIdx := 0

	// If text is wider than maxWidth, pan the viewport
	textWidth := StringWidth(displayText)
	if textWidth > maxWidth {
		// Calculate cursor position in display columns (up to byte position e.cursorPos)
		// Find display width at cursor position
		cursorDisplayWidth := 0
		for i, r := range displayText {
			if i >= e.cursorPos {
				break
			}
			cursorDisplayWidth += RuneWidth(r)
		}

		// Pan to show cursor at approximately center of viewport
		targetStartWidth := cursorDisplayWidth - maxWidth/2
		if targetStartWidth < 0 {
			targetStartWidth = 0
		}

		// Find byte offset for target display width
		targetStartIdx := 0
		displayWidth := 0
		for i, r := range displayText {
			if displayWidth >= targetStartWidth {
				targetStartIdx = i
				break
			}
			displayWidth += RuneWidth(r)
		}
		startIdx = targetStartIdx

		// Ensure we don't pan past the end
		if startIdx > 0 {
			endDisplayWidth := StringWidth(displayText[startIdx:])
			if endDisplayWidth < maxWidth && startIdx > 0 {
				// We have less than maxWidth of text remaining, pan back
				for startIdx > 0 {
					startIdx--
					endDisplayWidth = StringWidth(displayText[startIdx:])
					if endDisplayWidth <= maxWidth {
						break
					}
				}
			}
		}

		displayText = displayText[startIdx:]
	}

	// Draw the text with proper character width handling
	screenCol := 0
	cursorScreenCol := -1
	for byteIdx, r := range displayText {
		charWidth := RuneWidth(r)
		// Find if cursor is at this position
		if startIdx+byteIdx == e.cursorPos {
			cursorScreenCol = screenCol
		}

		// Draw character(s)
		screen.SetCell(x+screenCol, y, r, textStyle)
		screenCol += charWidth

		// For wide characters that take 2 columns, fill second column
		if charWidth == 2 && screenCol < maxWidth {
			// Wide character fills 2 columns, second is handled by tcell
		}
	}

	// Draw cursor at end of text if needed
	if e.cursorPos == len(e.text) && screenCol <= maxWidth {
		cursorScreenCol = screenCol
	}

	// Clear remainder of line
	for col := screenCol; col < maxWidth; col++ {
		if x+col < screen.GetWidth() {
			if col == cursorScreenCol {
				screen.SetCell(x+col, y, ' ', cursorStyle)
			} else {
				screen.SetCell(x+col, y, ' ', textStyle)
			}
		}
	}

	// Draw cursor on character if it's within the displayed text
	if cursorScreenCol >= 0 && cursorScreenCol < screenCol {
		// Re-draw character at cursor position with cursor style
		col := 0
		for _, r := range displayText {
			if col == cursorScreenCol {
				screen.SetCell(x+col, y, r, cursorStyle)
				break
			}
			col += RuneWidth(r)
		}
	}
}

// GetText returns the current text
func (e *Editor) GetText() string {
	return e.text
}

// SetText sets the text
func (e *Editor) SetText(text string) {
	e.text = text
	if e.cursorPos > len(e.text) {
		e.cursorPos = len(e.text)
	}
}

// GetCursorPos returns the cursor position
func (e *Editor) GetCursorPos() int {
	return e.cursorPos
}

// WasEnterPressed returns whether Enter was pressed and resets the flag
func (e *Editor) WasEnterPressed() bool {
	pressed := e.enterPressed
	e.enterPressed = false
	return pressed
}

// WasEscapePressed returns whether Escape was pressed and resets the flag
func (e *Editor) WasEscapePressed() bool {
	pressed := e.escapePressed
	e.escapePressed = false
	return pressed
}

// WasBackspaceOnEmpty returns whether Backspace was pressed on an empty item and resets the flag
func (e *Editor) WasBackspaceOnEmpty() bool {
	pressed := e.backspaceOnEmpty
	e.backspaceOnEmpty = false
	return pressed
}

// WasIndentPressed returns whether Tab was pressed to indent and resets the flag
func (e *Editor) WasIndentPressed() bool {
	pressed := e.indentPressed
	e.indentPressed = false
	return pressed
}

// WasOutdentPressed returns whether Shift+Tab was pressed to outdent and resets the flag
func (e *Editor) WasOutdentPressed() bool {
	pressed := e.outdentPressed
	e.outdentPressed = false
	return pressed
}

// GetItem returns the item being edited
func (e *Editor) GetItem() *model.Item {
	return e.item
}

// SetCursorFromScreenX sets the cursor position based on a screen X coordinate
// relativeX is the X coordinate relative to the start of the text (after indentation, arrow, etc.)
func (e *Editor) SetCursorFromScreenX(relativeX int) {
	if relativeX < 0 {
		relativeX = 0
	}
	if relativeX > len(e.text) {
		relativeX = len(e.text)
	}
	e.cursorPos = relativeX
}

// SetCursorToStart positions the cursor at the beginning of the text
func (e *Editor) SetCursorToStart() {
	e.cursorPos = 0
}

// InsertCurrentDate inserts the current date at the cursor position (YYYY-MM-DD format)
func (e *Editor) InsertCurrentDate() {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	e.text = e.text[:e.cursorPos] + dateStr + e.text[e.cursorPos:]
	e.cursorPos += len(dateStr)
}

// InsertCurrentTime inserts the current time at the beginning with a space (HH:MM format)
func (e *Editor) InsertCurrentTime() {
	now := time.Now()
	timeStr := now.Format("15:04 ") // Add space after time
	// Always insert at the beginning
	e.text = timeStr + e.text
	e.cursorPos += len(timeStr)
}

// InsertCurrentDateTime inserts the current date and time at the cursor position (YYYY-MM-DD HH:MM:SS format)
func (e *Editor) InsertCurrentDateTime() {
	now := time.Now()
	dateTimeStr := now.Format("2006-01-02 15:04:05")
	e.text = e.text[:e.cursorPos] + dateTimeStr + e.text[e.cursorPos:]
	e.cursorPos += len(dateTimeStr)
}

// DeleteWordBackwards deletes the word before the cursor
func (e *Editor) DeleteWordBackwards() {
	if e.cursorPos == 0 {
		return
	}

	// Start from cursor position and move backwards
	pos := e.cursorPos - 1

	// Skip any trailing whitespace
	for pos >= 0 && (e.text[pos] == ' ' || e.text[pos] == '\t') {
		pos--
	}

	// Skip the word characters
	for pos >= 0 && e.text[pos] != ' ' && e.text[pos] != '\t' {
		pos--
	}

	// Delete from pos+1 to cursorPos
	deleteStart := pos + 1
	e.text = e.text[:deleteStart] + e.text[e.cursorPos:]
	e.cursorPos = deleteStart
}
