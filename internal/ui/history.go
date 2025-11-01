package ui

// History manages input history for search and command inputs
// It allows navigating backward and forward through previous entries
type History struct {
	entries         []string // All stored history entries
	currentIndex    int      // Current position while navigating (-1 = not navigating)
	maxEntries      int      // Maximum number of entries to keep
	temporaryInput  string   // Stores current input before navigating history
}

// NewHistory creates a new History with a maximum number of entries
func NewHistory(maxEntries int) *History {
	return &History{
		entries:      []string{},
		currentIndex: -1,
		maxEntries:   maxEntries,
	}
}

// Add adds an entry to the history
// Avoids empty entries and duplicate consecutive entries
// Automatically removes oldest entries if max size is exceeded
func (h *History) Add(entry string) {
	if entry == "" {
		return
	}

	// Don't add if it's the same as the most recent entry
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == entry {
		return
	}

	h.entries = append(h.entries, entry)

	// Trim history if it exceeds maxEntries
	if len(h.entries) > h.maxEntries {
		h.entries = h.entries[len(h.entries)-h.maxEntries:]
	}

	// Reset navigation index when adding new entry
	h.currentIndex = -1
	h.temporaryInput = ""
}

// Previous returns the previous entry in history
// Call SetTemporary() before calling this if you want to restore current input when navigating forward
func (h *History) Previous() (string, bool) {
	if len(h.entries) == 0 {
		return "", false
	}

	// If we're not in history, start from the end
	if h.currentIndex < 0 {
		h.currentIndex = len(h.entries) - 1
	} else if h.currentIndex > 0 {
		h.currentIndex--
	}

	// Clamp to valid range
	if h.currentIndex >= len(h.entries) {
		h.currentIndex = len(h.entries) - 1
	}
	if h.currentIndex < 0 {
		return "", false
	}

	return h.entries[h.currentIndex], true
}

// Next returns the next entry in history
// Returns the temporarily stored input if we navigate past the end of history
func (h *History) Next() (string, bool) {
	if len(h.entries) == 0 {
		return "", false
	}

	// If we're not in history, return empty
	if h.currentIndex < 0 {
		return "", false
	}

	h.currentIndex++

	// If we go past the end, return temporary input and reset navigation
	if h.currentIndex >= len(h.entries) {
		h.currentIndex = -1
		temp := h.temporaryInput
		h.temporaryInput = ""
		return temp, true
	}

	return h.entries[h.currentIndex], true
}

// Reset resets the navigation state
// Call this when entering input mode or exiting history navigation
func (h *History) Reset() {
	h.currentIndex = -1
	h.temporaryInput = ""
}

// SetTemporary stores the current input before navigating history
// This allows restoring the original input when navigating forward past the end
func (h *History) SetTemporary(input string) {
	h.temporaryInput = input
}

// GetAll returns a copy of all history entries
func (h *History) GetAll() []string {
	entries := make([]string, len(h.entries))
	copy(entries, h.entries)
	return entries
}

// Clear removes all history entries
func (h *History) Clear() {
	h.entries = nil
	h.currentIndex = -1
	h.temporaryInput = ""
}

// Len returns the number of entries in history
func (h *History) Len() int {
	return len(h.entries)
}

// IsNavigating returns true if currently navigating through history
func (h *History) IsNavigating() bool {
	return h.currentIndex >= 0
}
