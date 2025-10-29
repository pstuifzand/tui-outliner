package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// Editor manages inline text editing of outline items
type Editor struct {
	item      *model.Item
	text      string
	cursorPos int
	active    bool
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

	switch ev.Key() {
	case tcell.KeyEscape:
		return false // Signal to exit edit mode
	case tcell.KeyEnter:
		return false // Signal to exit edit mode
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.cursorPos > 0 {
			e.text = e.text[:e.cursorPos-1] + e.text[e.cursorPos:]
			e.cursorPos--
		}
	case tcell.KeyDelete:
		if e.cursorPos < len(e.text) {
			e.text = e.text[:e.cursorPos] + e.text[e.cursorPos+1:]
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
		ch := ev.Rune()
		if ch > 0 && ch < 127 { // Printable ASCII
			// If this is a new item with placeholder text, clear it on first character
			if e.item.IsNew && e.text == "Type here..." {
				e.text = string(ch)
				e.cursorPos = 1
				e.item.IsNew = false
			} else {
				e.text = e.text[:e.cursorPos] + string(ch) + e.text[e.cursorPos:]
				e.cursorPos++
			}
		}
	}

	return true
}

// Render renders the editor on the screen
func (e *Editor) Render(screen *Screen, x, y int, maxWidth int) {
	textStyle := screen.EditorStyle()
	cursorStyle := screen.EditorCursorStyle()

	// Determine which portion of text to display
	displayText := e.text
	startIdx := 0
	if len(displayText) > maxWidth {
		// Show portion around cursor
		startIdx = e.cursorPos - maxWidth/2
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx+maxWidth > len(displayText) {
			startIdx = len(displayText) - maxWidth
		}
		if startIdx < 0 {
			startIdx = 0
		}
		displayText = displayText[startIdx:]
	}

	// Draw the text
	for i, r := range displayText {
		screen.SetCell(x+i, y, r, textStyle)
	}

	// Clear remainder (except cursor position if it's at the end)
	cursorScreenX := e.cursorPos - startIdx
	for i := len(displayText); i < maxWidth; i++ {
		if x+i < screen.GetWidth() {
			// Show cursor as a block at the end
			if i == cursorScreenX && e.cursorPos == len(e.text) {
				screen.SetCell(x+i, y, ' ', cursorStyle)
			} else {
				screen.SetCell(x+i, y, ' ', textStyle)
			}
		}
	}

	// Draw cursor on character if it's within the displayed text
	if cursorScreenX >= 0 && cursorScreenX < len(displayText) {
		// Cursor is on a character - highlight it in reverse
		r := rune(displayText[cursorScreenX])
		screen.SetCell(x+cursorScreenX, y, r, cursorStyle)
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
