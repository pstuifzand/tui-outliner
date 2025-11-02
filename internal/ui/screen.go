package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/config"
	"github.com/pstuifzand/tui-outliner/internal/theme"
)

// Screen manages the tcell screen and rendering
type Screen struct {
	tcellScreen tcell.Screen
	width       int
	height      int
	Theme       *theme.Theme
}

// NewScreen creates a new Screen instance with the configured theme
func NewScreen() (*Screen, error) {
	// Load config to get the theme name
	cfg, err := config.Load()
	if err != nil {
		// If config fails to load, use Default as fallback
		return NewScreenWithTheme(theme.Default())
	}

	// Load the theme based on config
	// Try to load from TOML files first, fall back to built-in Default
	t := theme.LoadThemeOrDefault(cfg.Theme)
	return NewScreenWithTheme(t)
}

// NewScreenWithTheme creates a new Screen instance with a specific theme
func NewScreenWithTheme(t *theme.Theme) (*Screen, error) {
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
		Theme:       t,
	}, nil
}

// Close closes the screen
func (s *Screen) Close() error {
	s.tcellScreen.Fini()
	return nil
}

// Suspend releases terminal control temporarily
func (s *Screen) Suspend() error {
	// Use tcell's built-in Suspend which properly handles the terminal state
	// without breaking the event polling loop
	return s.tcellScreen.Suspend()
}

// Resume restores terminal control after suspension
func (s *Screen) Resume() error {
	// Use tcell's built-in Resume which restores proper operation
	// The event polling loop remains intact
	return s.tcellScreen.Resume()
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

// EnableMouse enables mouse support on the screen
func (s *Screen) EnableMouse() {
	s.tcellScreen.EnableMouse()
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

// Theme-aware style methods

// TreeNormalStyle returns the style for normal tree items
func (s *Screen) TreeNormalStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.TreeNormalText, s.Theme.Colors.Background)
}

// TreeSelectedStyle returns the style for selected tree items
func (s *Screen) TreeSelectedStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.TreeSelectedItem, s.Theme.Colors.TreeSelectedBg).Bold(true)
}

// TreeNewItemStyle returns the style for new/placeholder tree items
func (s *Screen) TreeNewItemStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.TreeNewItem).Dim(true)
}

// TreeLeafArrowStyle returns the style for leaf node arrows (dimmer)
func (s *Screen) TreeLeafArrowStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.TreeLeafArrow)
}

// TreeExpandableArrowStyle returns the style for expandable node arrows (brighter)
func (s *Screen) TreeExpandableArrowStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.TreeExpandableArrow)
}

// TreeVisualSelectionStyle returns the style for items in visual selection
func (s *Screen) TreeVisualSelectionStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.TreeVisualSelection, s.Theme.Colors.TreeVisualSelectionBg)
}

// TreeVisualCursorStyle returns the style for the cursor position in visual selection
func (s *Screen) TreeVisualCursorStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.TreeVisualCursor, s.Theme.Colors.TreeVisualCursorBg).Bold(true)
}

// TreeAttributeIndicatorStyle returns the style for attribute indicator symbol
func (s *Screen) TreeAttributeIndicatorStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.TreeAttributeIndicator)
}

// TreeAttributeStyle returns the style for visible attribute values (gray/dim)
func (s *Screen) TreeAttributeStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.TreeAttributeValue)
}

// EditorStyle returns the style for editor text
func (s *Screen) EditorStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.EditorText)
}

// EditorCursorStyle returns the style for editor cursor
func (s *Screen) EditorCursorStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.EditorCursor, s.Theme.Colors.EditorCursorBg)
}

// SearchLabelStyle returns the style for search label
func (s *Screen) SearchLabelStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.SearchLabel)
}

// SearchTextStyle returns the style for search text
func (s *Screen) SearchTextStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.SearchText)
}

// SearchCursorStyle returns the style for search cursor
func (s *Screen) SearchCursorStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.SearchCursor, s.Theme.Colors.SearchCursorBg)
}

// SearchResultCountStyle returns the style for search result count
func (s *Screen) SearchResultCountStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.SearchResultCount)
}

// SearchHighlightStyle returns the style for highlighted search matches in items
func (s *Screen) SearchHighlightStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.SearchHighlight, s.Theme.Colors.SearchHighlightBg)
}

// CommandPromptStyle returns the style for command prompt
func (s *Screen) CommandPromptStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.CommandPrompt)
}

// CommandTextStyle returns the style for command text
func (s *Screen) CommandTextStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.CommandText)
}

// CommandCursorStyle returns the style for command cursor
func (s *Screen) CommandCursorStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.CommandCursor, s.Theme.Colors.CommandCursorBg)
}

// HelpStyle returns the style for help background
func (s *Screen) HelpStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.HelpContent, s.Theme.Colors.HelpBackground)
}

// HelpBorderStyle returns the style for help borders
func (s *Screen) HelpBorderStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.HelpBorder, s.Theme.Colors.HelpBackground)
}

// HelpTitleStyle returns the style for help title
func (s *Screen) HelpTitleStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.HelpTitle, s.Theme.Colors.HelpBackground).Bold(true)
}

// StatusModeStyle returns the style for mode indicator
func (s *Screen) StatusModeStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.StatusMode, s.Theme.Colors.StatusModeBg).Bold(true)
}

// StatusMessageStyle returns the style for status messages
func (s *Screen) StatusMessageStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.StatusMessage)
}

// StatusModifiedStyle returns the style for modified indicator
func (s *Screen) StatusModifiedStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.StatusModified)
}

// HeaderStyle returns the style for header title
func (s *Screen) HeaderStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.HeaderTitle, s.Theme.Colors.HeaderBg).Bold(true)
}

// Standard color styles
func (s *Screen) YellowStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.ColorYellow)
}

func (s *Screen) OrangeStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.ColorOrange)
}

func (s *Screen) RedStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.ColorRed)
}

func (s *Screen) GreenStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.ColorGreen)
}

func (s *Screen) BlueStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.ColorBlue)
}

func (s *Screen) PurpleStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.ColorPurple)
}

func (s *Screen) GrayStyle() tcell.Style {
	return theme.ColorToStyle(s.Theme.Colors.ColorGray)
}

// BackgroundStyle returns the default background style for the application
func (s *Screen) BackgroundStyle() tcell.Style {
	return tcell.StyleDefault.Background(s.Theme.Colors.Background)
}

// CalendarDayStyle returns the style for calendar day cells
func (s *Screen) CalendarDayStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.CalendarDayText, s.Theme.Colors.CalendarDayBg)
}

// CalendarInactiveDayStyle returns the style for inactive calendar days (prev/next month)
func (s *Screen) CalendarInactiveDayStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.CalendarInactiveDayText, s.Theme.Colors.CalendarInactiveDayBg)
}

// CalendarDayIndicatorStyle returns the style for indicator dots with indicator foreground and day background
func (s *Screen) CalendarDayIndicatorStyle() tcell.Style {
	return theme.ColorPairToStyle(s.Theme.Colors.TreeAttributeIndicator, s.Theme.Colors.CalendarDayBg)
}
