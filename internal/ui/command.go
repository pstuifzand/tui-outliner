package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// CommandMode manages command line input (`:command`)
type CommandMode struct {
	active   bool
	input    string
	cursorPos int
}

// NewCommandMode creates a new CommandMode
func NewCommandMode() *CommandMode {
	return &CommandMode{
		active:    false,
		input:     "",
		cursorPos: 0,
	}
}

// Start enters command mode
func (c *CommandMode) Start() {
	c.active = true
	c.input = ""
	c.cursorPos = 0
}

// Stop exits command mode
func (c *CommandMode) Stop() {
	c.active = false
}

// IsActive returns whether command mode is active
func (c *CommandMode) IsActive() bool {
	return c.active
}

// HandleKey processes a key press in command mode
func (c *CommandMode) HandleKey(ev *tcell.EventKey) (command string, done bool) {
	switch ev.Key() {
	case tcell.KeyEscape:
		c.Stop()
		return "", true
	case tcell.KeyEnter:
		cmd := strings.TrimSpace(c.input)
		c.Stop()
		return cmd, true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if c.cursorPos > 0 {
			c.input = c.input[:c.cursorPos-1] + c.input[c.cursorPos:]
			c.cursorPos--
		}
	case tcell.KeyDelete:
		if c.cursorPos < len(c.input) {
			c.input = c.input[:c.cursorPos] + c.input[c.cursorPos+1:]
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
		if ch > 0 && ch < 127 {
			c.input = c.input[:c.cursorPos] + string(ch) + c.input[c.cursorPos:]
			c.cursorPos++
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
