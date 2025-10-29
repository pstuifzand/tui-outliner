package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Screen manages the tcell screen and rendering
type Screen struct {
	tcellScreen tcell.Screen
	width       int
	height      int
}

// NewScreen creates a new Screen instance
func NewScreen() (*Screen, error) {
	tcellScreen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	if err := tcellScreen.Init(); err != nil {
		return nil, fmt.Errorf("failed to init screen: %w", err)
	}

	width, height := tcellScreen.Size()
	return &Screen{
		tcellScreen: tcellScreen,
		width:       width,
		height:      height,
	}, nil
}

// Close closes the screen
func (s *Screen) Close() error {
	s.tcellScreen.Fini()
	return nil
}

// Clear clears the entire screen
func (s *Screen) Clear() {
	s.tcellScreen.Clear()
}

// SetCell sets a cell at the given position
func (s *Screen) SetCell(x, y int, r rune, style tcell.Style) {
	if x >= 0 && x < s.width && y >= 0 && y < s.height {
		s.tcellScreen.SetContent(x, y, r, nil, style)
	}
}

// DrawString draws a string at the given position with the given style
func (s *Screen) DrawString(x, y int, text string, style tcell.Style) {
	for i, r := range text {
		s.SetCell(x+i, y, r, style)
	}
}

// DrawStringLimited draws a string, truncating it if it exceeds maxWidth
func (s *Screen) DrawStringLimited(x, y int, text string, maxWidth int, style tcell.Style) {
	if maxWidth <= 0 {
		return
	}
	if len(text) > maxWidth {
		text = text[:maxWidth]
	}
	s.DrawString(x, y, text, style)
}

// PollEvent polls for the next event (key press, mouse, etc.)
func (s *Screen) PollEvent() tcell.Event {
	return s.tcellScreen.PollEvent()
}

// PollEventWithTimeout polls for an event with a timeout
// Note: tcell's PollEvent is already blocking, so we just call it
// The timeout is handled at the application level by checking elapsed time
func (s *Screen) PollEventWithTimeout(timeout time.Duration) tcell.Event {
	// tcell.PollEvent() blocks until an event is available
	// We can't interrupt it from here, so we just call it directly
	// The application loop should handle timing out the check
	return s.tcellScreen.PollEvent()
}

// Show shows the screen
func (s *Screen) Show() {
	s.tcellScreen.Show()
}

// Size returns the width and height of the screen
func (s *Screen) Size() (int, int) {
	w, h := s.tcellScreen.Size()
	s.width = w
	s.height = h
	return w, h
}

// GetWidth returns the width of the screen
func (s *Screen) GetWidth() int {
	s.width, _ = s.tcellScreen.Size()
	return s.width
}

// GetHeight returns the height of the screen
func (s *Screen) GetHeight() int {
	_, s.height = s.tcellScreen.Size()
	return s.height
}

// HasMouse returns true if mouse is supported
func (s *Screen) HasMouse() bool {
	return s.tcellScreen.HasMouse()
}

// DefaultStyle returns the default terminal style
func DefaultStyle() tcell.Style {
	return tcell.StyleDefault
}

// StyleBold returns a bold style
func StyleBold() tcell.Style {
	return tcell.StyleDefault.Bold(true)
}

// StyleReverse returns a reverse video style
func StyleReverse() tcell.Style {
	return tcell.StyleDefault.Reverse(true)
}

// StyleDim returns a dim style
func StyleDim() tcell.Style {
	return tcell.StyleDefault.Dim(true)
}
