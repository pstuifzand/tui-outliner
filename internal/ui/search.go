package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
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
	queryLocked     bool               // Whether query is locked after pressing Enter
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
	s.matchIndices = nil
	s.currentMatchIdx = 0
	s.queryLocked = false
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
// Returns true if the key was handled and should navigate to a match
// Returns false if the key was normal search input
func (s *Search) HandleKey(ev *tcell.EventKey) bool {
	if !s.active {
		return false
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		s.Stop()
		return false
	case tcell.KeyEnter:
		// Enter key updates results and starts navigation to matches (stays in search mode for n/N)
		s.updateResults()
		s.queryLocked = true // Lock the query after pressing Enter
		if len(s.matchIndices) > 0 {
			return true // Signal to navigate to first match
		}
		return false
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Only allow backspace if query is not locked
		if !s.queryLocked && s.cursorPos > 0 {
			s.query = s.query[:s.cursorPos-1] + s.query[s.cursorPos:]
			s.cursorPos--
		}
		return false
	case tcell.KeyDelete:
		// Only allow delete if query is not locked
		if !s.queryLocked && s.cursorPos < len(s.query) {
			s.query = s.query[:s.cursorPos] + s.query[s.cursorPos+1:]
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
			s.queryLocked = false // Unlock query for new search
			return false // Not a navigation command
		}
		// Handle 'n' and 'N' for navigation only if we have matches and user is done entering query
		// (i.e., only treat as navigation if there are matches, not while typing)
		if (ch == 'n' || ch == 'N') && len(s.matchIndices) > 0 {
			if ch == 'n' {
				s.NextMatch()
			} else {
				s.PrevMatch()
			}
			return true // Indicate this was a navigation command
		}
		// Regular character input - only allow if query is not locked
		if !s.queryLocked && ch > 0 && ch < 127 {
			s.query = s.query[:s.cursorPos] + string(ch) + s.query[s.cursorPos:]
			s.cursorPos++
		}
		return false
	}
}

// updateResults filters results based on query and tracks match indices
func (s *Search) updateResults() {
	s.matchIndices = nil
	s.currentMatchIdx = 0

	if s.query == "" {
		s.results = s.allItems
		return
	}

	query := strings.ToLower(s.query)
	var filtered []*model.Item
	for idx, item := range s.allItems {
		if strings.Contains(strings.ToLower(item.Text), query) {
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

	// Draw result count with current match number
	var resultText string
	if len(s.results) == 0 {
		resultText = " (no matches)"
	} else {
		currentNum := s.GetCurrentMatchNumber()
		totalCount := s.GetMatchCount()
		// Format: (1 of 5 matches)
		resultText = " (" + fmt.Sprintf("%d", currentNum) + " of " + fmt.Sprintf("%d", totalCount) + " matches)"
	}
	screen.DrawString(screen.GetWidth()-len(resultText), y, resultText, resultStyle)
}
