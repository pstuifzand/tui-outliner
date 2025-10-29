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
	style := DefaultStyle()
	cursorStyle := StyleReverse()

	// Draw the text
	displayText := e.text
	if len(displayText) > maxWidth {
		// Show portion around cursor
		start := e.cursorPos - maxWidth/2
		if start < 0 {
			start = 0
		}
		if start+maxWidth > len(displayText) {
			start = len(displayText) - maxWidth
		}
		if start < 0 {
			start = 0
		}
		displayText = displayText[start:]
	}

	for i, r := range displayText {
		charStyle := style
		if i == e.cursorPos {
			charStyle = cursorStyle
		}
		screen.SetCell(x+i, y, r, charStyle)
	}

	// Draw cursor at end if needed
	if e.cursorPos >= len(displayText) && e.cursorPos < maxWidth+x {
		screen.SetCell(x+e.cursorPos, y, ' ', cursorStyle)
	}

	// Clear remainder
	for i := len(displayText); i < maxWidth; i++ {
		if x+i < screen.GetWidth() {
			screen.SetCell(x+i, y, ' ', style)
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
