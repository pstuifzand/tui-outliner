package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// Search manages search and filter functionality
type Search struct {
	query      string
	results    []*model.Item
	cursorPos  int
	active     bool
	allItems   []*model.Item
}

// NewSearch creates a new Search
func NewSearch(items []*model.Item) *Search {
	return &Search{
		query:     "",
		results:   items,
		cursorPos: 0,
		active:    false,
		allItems:  items,
	}
}

// Start starts search mode
func (s *Search) Start() {
	s.active = true
	s.query = ""
	s.cursorPos = 0
}

// Stop stops search mode
func (s *Search) Stop() {
	s.active = false
}

// IsActive returns whether search mode is active
func (s *Search) IsActive() bool {
	return s.active
}

// HandleKey handles key presses during search
func (s *Search) HandleKey(ev *tcell.EventKey) {
	if !s.active {
		return
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		s.Stop()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if s.cursorPos > 0 {
			s.query = s.query[:s.cursorPos-1] + s.query[s.cursorPos:]
			s.cursorPos--
			s.updateResults()
		}
	case tcell.KeyDelete:
		if s.cursorPos < len(s.query) {
			s.query = s.query[:s.cursorPos] + s.query[s.cursorPos+1:]
			s.updateResults()
		}
	case tcell.KeyLeft:
		if s.cursorPos > 0 {
			s.cursorPos--
		}
	case tcell.KeyRight:
		if s.cursorPos < len(s.query) {
			s.cursorPos++
		}
	case tcell.KeyHome:
		s.cursorPos = 0
	case tcell.KeyEnd:
		s.cursorPos = len(s.query)
	default:
		ch := ev.Rune()
		if ch > 0 && ch < 127 {
			s.query = s.query[:s.cursorPos] + string(ch) + s.query[s.cursorPos:]
			s.cursorPos++
			s.updateResults()
		}
	}
}

// updateResults filters results based on query
func (s *Search) updateResults() {
	if s.query == "" {
		s.results = s.allItems
		return
	}

	query := strings.ToLower(s.query)
	var filtered []*model.Item
	for _, item := range s.allItems {
		if strings.Contains(strings.ToLower(item.Text), query) {
			filtered = append(filtered, item)
		}
	}
	s.results = filtered
}

// GetResults returns the current search results
func (s *Search) GetResults() []*model.Item {
	return s.results
}

// GetQuery returns the current search query
func (s *Search) GetQuery() string {
	return s.query
}

// SetAllItems sets the items to search in
func (s *Search) SetAllItems(items []*model.Item) {
	s.allItems = items
	s.updateResults()
}

// Render renders the search bar on the screen
func (s *Search) Render(screen *Screen, y int) {
	style := DefaultStyle()
	cursorStyle := StyleReverse()

	// Draw label
	screen.DrawString(0, y, "Search: ", style)

	// Draw query
	x := 8
	maxWidth := screen.GetWidth() - x
	displayQuery := s.query
	if len(displayQuery) > maxWidth {
		displayQuery = displayQuery[len(displayQuery)-maxWidth:]
	}

	for i, r := range displayQuery {
		charStyle := style
		if i == s.cursorPos {
			charStyle = cursorStyle
		}
		screen.SetCell(x+i, y, r, charStyle)
	}

	// Draw cursor
	if s.cursorPos >= len(displayQuery) {
		screen.SetCell(x+len(displayQuery), y, ' ', cursorStyle)
	}

	// Clear remainder
	for i := len(displayQuery); i < maxWidth; i++ {
		if x+i < screen.GetWidth() {
			screen.SetCell(x+i, y, ' ', style)
		}
	}

	// Draw result count
	resultText := " (" + string(rune(len(s.results))) + " matches)"
	if len(s.results) > 9 {
		resultText = " (" + string(rune(len(s.results))) + " matches)"
	}
	screen.DrawString(screen.GetWidth()-len(resultText), y, resultText, style)
}
