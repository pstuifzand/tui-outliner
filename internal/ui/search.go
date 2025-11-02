package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/history"
	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/search"
)

// Search manages search and filter functionality
type Search struct {
	query           string
	results         []*model.Item
	matchIndices    []int              // Indices of matches in allItems
	currentMatchIdx int                // Index in matchIndices array (which match we're on)
	cursorPos       int
	active          bool
	allItems        []*model.Item
	filterExpr      search.FilterExpr  // Parsed filter expression
	parseError      string             // Error from parsing the query
	history         *History           // Search history manager
}

// NewSearch creates a new Search without history persistence
func NewSearch(items []*model.Item) *Search {
	return &Search{
		query:    "",
		results:  items,
		cursorPos: 0,
		active:   false,
		allItems: items,
		history:  NewHistory(50),
	}
}

// NewSearchWithHistory creates a new Search with history persistence
func NewSearchWithHistory(items []*model.Item, manager *history.Manager) (*Search, error) {
	h, err := NewHistoryWithManager(50, manager, "search.toml")
	if err != nil {
		// If history loading fails, continue with empty history
		h = NewHistory(50)
	}

	return &Search{
		query:    "",
		results:  items,
		cursorPos: 0,
		active:   false,
		allItems: items,
		history:  h,
	}, nil
}

// Start starts search mode
func (s *Search) Start() {
	s.active = true
	s.query = ""
	s.cursorPos = 0
	s.matchIndices = nil
	s.currentMatchIdx = 0
	s.history.Reset()
}

// Stop stops search mode
func (s *Search) Stop() {
	s.active = false
	s.history.Reset()
}

// IsActive returns whether search mode is active
func (s *Search) IsActive() bool {
	return s.active
}

// GetHistory returns a copy of the search history
func (s *Search) GetHistory() []string {
	return s.history.GetAll()
}

// HandleKey handles key presses during search mode
// Returns true if the key was a navigation command that should stay in search
// Returns false otherwise
func (s *Search) HandleKey(ev *tcell.EventKey) bool {
	if !s.active {
		return false
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		s.Stop()
		return false
	case tcell.KeyEnter:
		// Enter key updates results and adds to history, then exits search mode
		s.updateResults()
		s.history.Add(s.query)
		s.Stop()
		// If there are matches, navigate to first one
		if len(s.matchIndices) > 0 {
			return true // Signal to navigate to first match in normal mode
		}
		return false
	case tcell.KeyUp:
		// Store current query before navigating history (on first Up press)
		if !s.history.IsNavigating() {
			s.history.SetTemporary(s.query)
		}
		// Navigate to previous search in history
		if prevQuery, ok := s.history.Previous(); ok {
			s.query = prevQuery
			s.cursorPos = len(s.query)
			s.updateResults()
		}
		return false
	case tcell.KeyDown:
		// Navigate to next search in history
		if nextQuery, ok := s.history.Next(); ok {
			s.query = nextQuery
			s.cursorPos = len(s.query)
			s.updateResults()
		}
		return false
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if s.cursorPos > 0 {
			s.query = s.query[:s.cursorPos-1] + s.query[s.cursorPos:]
			s.cursorPos--
			// Update results immediately as user deletes characters (incremental search)
			s.updateResults()
		}
		return false
	case tcell.KeyDelete:
		if s.cursorPos < len(s.query) {
			s.query = s.query[:s.cursorPos] + s.query[s.cursorPos+1:]
			// Update results immediately as user deletes characters (incremental search)
			s.updateResults()
		}
		return false
	case tcell.KeyLeft:
		if s.cursorPos > 0 {
			s.cursorPos--
		}
		return false
	case tcell.KeyRight:
		if s.cursorPos < len(s.query) {
			s.cursorPos++
		}
		return false
	case tcell.KeyHome:
		s.cursorPos = 0
		return false
	case tcell.KeyEnd:
		s.cursorPos = len(s.query)
		return false
	default:
		ch := ev.Rune()
		// Handle '/' to clear search and start new search
		if ch == '/' {
			s.query = ""
			s.cursorPos = 0
			s.results = s.allItems // Clear results back to all items
			s.matchIndices = nil
			s.currentMatchIdx = 0
			return false
		}
		// Regular character input
		if ch > 0 && ch < 127 {
			s.query = s.query[:s.cursorPos] + string(ch) + s.query[s.cursorPos:]
			s.cursorPos++
			// Update results immediately as user types (incremental search)
			s.updateResults()
		}
		return false
	}
}

// updateResults filters results based on query and tracks match indices
func (s *Search) updateResults() {
	s.matchIndices = nil
	s.currentMatchIdx = 0
	s.parseError = ""
	s.filterExpr = nil

	if s.query == "" {
		s.results = s.allItems
		// Parse empty query to get AlwaysMatch expression
		var err error
		s.filterExpr, err = search.ParseQuery("")
		if err != nil {
			s.parseError = err.Error()
		}
		return
	}

	// Parse the query
	var err error
	s.filterExpr, err = search.ParseQuery(s.query)
	if err != nil {
		s.parseError = err.Error()
		// Return no results on parse error
		s.results = nil
		return
	}

	// Apply the filter to all items
	var filtered []*model.Item
	for idx, item := range s.allItems {
		if s.filterExpr.Matches(item) {
			filtered = append(filtered, item)
			s.matchIndices = append(s.matchIndices, idx)
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

// NextMatch moves to the next search match
func (s *Search) NextMatch() bool {
	if len(s.matchIndices) == 0 {
		return false
	}
	s.currentMatchIdx++
	if s.currentMatchIdx >= len(s.matchIndices) {
		s.currentMatchIdx = 0 // Wrap around to first match
	}
	return true
}

// PrevMatch moves to the previous search match
func (s *Search) PrevMatch() bool {
	if len(s.matchIndices) == 0 {
		return false
	}
	s.currentMatchIdx--
	if s.currentMatchIdx < 0 {
		s.currentMatchIdx = len(s.matchIndices) - 1 // Wrap around to last match
	}
	return true
}

// GetCurrentMatchIndex returns the index in allItems of the current match
func (s *Search) GetCurrentMatchIndex() int {
	if len(s.matchIndices) == 0 || s.currentMatchIdx >= len(s.matchIndices) {
		return -1
	}
	return s.matchIndices[s.currentMatchIdx]
}

// GetCurrentMatch returns the current match item
func (s *Search) GetCurrentMatch() *model.Item {
	idx := s.GetCurrentMatchIndex()
	if idx >= 0 && idx < len(s.allItems) {
		return s.allItems[idx]
	}
	return nil
}

// GetMatchCount returns the number of matches
func (s *Search) GetMatchCount() int {
	return len(s.matchIndices)
}

// GetCurrentMatchNumber returns the current match number (1-based) or 0 if no matches
func (s *Search) GetCurrentMatchNumber() int {
	if len(s.matchIndices) == 0 {
		return 0
	}
	return s.currentMatchIdx + 1
}

// GetParseError returns the last parse error, if any
func (s *Search) GetParseError() string {
	return s.parseError
}

// HasResults returns true if there are active search results
func (s *Search) HasResults() bool {
	return len(s.matchIndices) > 0
}

// Render renders the search bar on the screen
func (s *Search) Render(screen *Screen, y int) {
	labelStyle := screen.SearchLabelStyle()
	textStyle := screen.SearchTextStyle()
	cursorStyle := screen.SearchCursorStyle()
	resultStyle := screen.SearchResultCountStyle()

	// Draw label
	screen.DrawString(0, y, "Search: ", labelStyle)

	// Draw query
	x := 8
	maxWidth := screen.GetWidth() - x
	displayQuery := s.query
	if len(displayQuery) > maxWidth {
		displayQuery = displayQuery[len(displayQuery)-maxWidth:]
	}

	for i, r := range displayQuery {
		charStyle := textStyle
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
			screen.SetCell(x+i, y, ' ', textStyle)
		}
	}

	// Draw result count with current match number or error
	var resultText string
	if s.parseError != "" {
		// Display parse error
		resultText = " (error: " + s.parseError + ")"
	} else if len(s.results) == 0 {
		resultText = " (no matches)"
	} else {
		currentNum := s.GetCurrentMatchNumber()
		totalCount := s.GetMatchCount()
		// Format: (1 of 5 matches)
		resultText = " (" + fmt.Sprintf("%d", currentNum) + " of " + fmt.Sprintf("%d", totalCount) + " matches)"
	}
	// Truncate error message if it's too long
	if len(resultText) > screen.GetWidth()/2 {
		resultText = " (error: syntax)"
	}
	screen.DrawString(screen.GetWidth()-len(resultText), y, resultText, resultStyle)
}
