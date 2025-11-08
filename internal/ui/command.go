package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/history"
)

// CommandMode manages command line input (`:command`)
type CommandMode struct {
	active    bool
	input     string
	cursorPos int
	history   *History
}

// NewCommandMode creates a new CommandMode without history persistence
func NewCommandMode() *CommandMode {
	return &CommandMode{
		active:    false,
		input:     "",
		cursorPos: 0,
		history:   NewHistory(50),
	}
}

// NewCommandModeWithHistory creates a new CommandMode with history persistence
func NewCommandModeWithHistory(manager *history.Manager) (*CommandMode, error) {
	h, err := NewHistoryWithManager(50, manager, "command.toml")
	if err != nil {
		// If history loading fails, continue with empty history
		h = NewHistory(50)
	}

	return &CommandMode{
		active:    false,
		input:     "",
		cursorPos: 0,
		history:   h,
	}, nil
}

// Start enters command mode
func (c *CommandMode) Start() {
	c.active = true
	c.input = ""
	c.cursorPos = 0
	c.history.Reset()
}

// Stop exits command mode
func (c *CommandMode) Stop() {
	c.active = false
}

// IsActive returns whether command mode is active
func (c *CommandMode) IsActive() bool {
	return c.active
}

// DeleteWordBackwards deletes the word before the cursor
func (c *CommandMode) DeleteWordBackwards() {
	if c.cursorPos == 0 {
		return
	}

	// Start from cursor position and move backwards
	pos := c.cursorPos - 1

	// Skip any trailing whitespace
	for pos >= 0 && (c.input[pos] == ' ' || c.input[pos] == '\t') {
		pos--
	}

	// Skip the word characters
	for pos >= 0 && c.input[pos] != ' ' && c.input[pos] != '\t' {
		pos--
	}

	// Delete from pos+1 to cursorPos
	deleteStart := pos + 1
	c.input = c.input[:deleteStart] + c.input[c.cursorPos:]
	c.cursorPos = deleteStart
}

// HandleKey processes a key press in command mode
func (c *CommandMode) HandleKey(ev *tcell.EventKey) (command string, done bool) {
	switch ev.Key() {
	case tcell.KeyCtrlW:
		// Check for Ctrl+W - delete word backwards
		c.DeleteWordBackwards()
	case tcell.KeyEscape:
		c.Stop()
		return "", true
	case tcell.KeyEnter:
		cmd := strings.TrimSpace(c.input)
		c.history.Add(cmd)
		c.Stop()
		return cmd, true
	case tcell.KeyUp:
		// Store current input before navigating history (on first Up press)
		if !c.history.IsNavigating() {
			c.history.SetTemporary(c.input)
		}
		// Navigate to previous command in history
		if prevCmd, ok := c.history.Previous(); ok {
			c.input = prevCmd
			c.cursorPos = len(c.input)
		}
	case tcell.KeyDown:
		// Navigate to next command in history
		if nextCmd, ok := c.history.Next(); ok {
			c.input = nextCmd
			c.cursorPos = len(c.input)
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if c.cursorPos > 0 {
			// Delete entire character (which may be multiple bytes for UTF-8)
			before := c.input[:c.cursorPos]
			runes := []rune(before)
			if len(runes) > 0 {
				runes = runes[:len(runes)-1] // Remove last rune/character
				newBefore := string(runes)
				deletedBytes := len(before) - len(newBefore)
				c.input = newBefore + c.input[c.cursorPos:]
				c.cursorPos -= deletedBytes
			}
		} else if c.input == "" {
			// Exit command mode when backspace is pressed on empty command line
			c.Stop()
			return "", true
		}
	case tcell.KeyDelete:
		if c.cursorPos < len(c.input) {
			// Delete entire character (which may be multiple bytes for UTF-8)
			after := c.input[c.cursorPos:]
			runes := []rune(after)
			if len(runes) > 0 {
				// Calculate bytes to delete (the first rune)
				deletedBytes := len(string(runes[:1]))
				c.input = c.input[:c.cursorPos] + c.input[c.cursorPos+deletedBytes:]
			}
		}
	case tcell.KeyLeft:
		if c.cursorPos > 0 {
			c.cursorPos--
		}
	case tcell.KeyRight:
		if c.cursorPos < len(c.input) {
			c.cursorPos++
		}
	case tcell.KeyHome:
		c.cursorPos = 0
	case tcell.KeyEnd:
		c.cursorPos = len(c.input)
	case tcell.KeyCtrlU:
		c.input = c.input[c.cursorPos:]
		c.cursorPos = 0
	case tcell.KeyCtrlK:
		c.input = c.input[:c.cursorPos]
	default:
		ch := ev.Rune()
		if ch > 0 { // Accept all valid Unicode characters
			s := string(ch)
			c.input = c.input[:c.cursorPos] + s + c.input[c.cursorPos:]
			c.cursorPos += len(s) // Increment by byte length, not character count
		}
	}

	return "", false
}

// GetInput returns the current command input
func (c *CommandMode) GetInput() string {
	return strings.TrimSpace(c.input)
}

// Render renders the command line
func (c *CommandMode) Render(screen *Screen, y int) {
	if !c.active {
		return
	}

	promptStyle := screen.CommandPromptStyle()
	textStyle := screen.CommandTextStyle()
	cursorStyle := screen.CommandCursorStyle()
	screenWidth := screen.GetWidth()

	// Draw colon and input
	prefix := ":"
	x := 0
	screen.DrawString(x, y, prefix, promptStyle)
	x += len(prefix)

	// Draw input with cursor
	for i, r := range c.input {
		charStyle := textStyle
		if i == c.cursorPos {
			charStyle = cursorStyle
		}
		if x < screenWidth {
			screen.SetCell(x, y, r, charStyle)
			x++
		}
	}

	// Draw cursor at end if needed
	if c.cursorPos >= len(c.input) && x < screenWidth {
		screen.SetCell(x, y, ' ', cursorStyle)
		x++
	}

	// Clear remainder of line
	for x < screenWidth {
		screen.SetCell(x, y, ' ', textStyle)
		x++
	}
}
