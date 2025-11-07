package ui

import (
	"github.com/mattn/go-runewidth"
)

// TextWidth provides Unicode-aware text width calculations for proper handling
// of wide characters (emoji, CJK, combining marks, etc.)
// All functions work with display width (screen columns) not byte length

// RuneWidth returns the display width of a single rune
// - ASCII and most Unicode: 1 column
// - Wide characters (emoji, CJK): 2 columns
// - Combining marks, zero-width spaces: 0 columns
// - Control characters: 0 columns
func RuneWidth(r rune) int {
	w := runewidth.RuneWidth(r)
	if w < 0 {
		// Negative width means control/combining character, treat as 0
		return 0
	}
	return w
}

// StringWidth returns the display width of a string
// Properly handles multi-byte characters and combining marks
func StringWidth(s string) int {
	return runewidth.StringWidth(s)
}

// StringWidthUpTo returns the display width of a string, stopping at maxWidth
// Returns (width, runes_used) so caller knows how many runes to slice
func StringWidthUpTo(s string, maxWidth int) (width int, runesUsed int) {
	if maxWidth <= 0 {
		return 0, 0
	}

	width = 0
	for i, r := range s {
		rw := RuneWidth(r)
		if width+rw > maxWidth {
			return width, i
		}
		width += rw
	}
	return width, len([]rune(s))
}

// TruncateToWidth safely truncates a string to fit within maxWidth columns
// Properly handles multi-byte characters without splitting them
func TruncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(s)
	width := 0

	for i, r := range runes {
		rw := RuneWidth(r)
		if width+rw > maxWidth {
			return string(runes[:i])
		}
		width += rw
	}

	return s
}

// TruncateToWidthWithEllipsis truncates a string with "..." if it exceeds maxWidth
// Reserves space for ellipsis
func TruncateToWidthWithEllipsis(s string, maxWidth int) string {
	if maxWidth <= 3 {
		return TruncateToWidth(s, maxWidth)
	}

	if StringWidth(s) <= maxWidth {
		return s
	}

	// Reserve 3 columns for "..."
	truncated := TruncateToWidth(s, maxWidth-3)
	return truncated + "..."
}

// PadStringToWidth pads a string to a specific display width with spaces
// If string is already wider, returns unchanged
func PadStringToWidth(s string, width int) string {
	current := StringWidth(s)
	if current >= width {
		return s
	}
	padding := width - current
	for i := 0; i < padding; i++ {
		s += " "
	}
	return s
}

// FindRuneIndexAtWidth finds the byte index that corresponds to a specific display width
// Useful for mapping screen column positions to byte indices
// Returns the byte index of the rune that starts at or after the given column
func FindRuneIndexAtWidth(s string, targetWidth int) int {
	if targetWidth <= 0 {
		return 0
	}

	width := 0
	for i, r := range s {
		if width >= targetWidth {
			return i
		}
		width += RuneWidth(r)
	}

	// String ends before reaching targetWidth
	return len(s)
}

// CalculateBreakPoint finds where to break text for wrapping at maxWidth
// Returns (byteIndex, width) where to break and the actual width used
// Prefers breaking at word boundaries (spaces), falls back to character boundary
func CalculateBreakPoint(s string, maxWidth int) (byteIndex int, actualWidth int) {
	if maxWidth <= 0 {
		return 0, 0
	}

	runes := []rune(s)
	width := 0
	lastSpaceIdx := -1
	lastSpaceWidth := 0

	for i, r := range runes {
		rw := RuneWidth(r)

		// Check if adding this rune would exceed maxWidth
		if width+rw > maxWidth {
			// If we found a space, break there (word boundary)
			if lastSpaceIdx >= 0 {
				// Return index after the space
				return len(string(runes[:lastSpaceIdx+1])), lastSpaceWidth + RuneWidth(runes[lastSpaceIdx])
			}
			// No space found, break at current position (character boundary)
			return len(string(runes[:i])), width
		}

		width += rw

		// Mark space position for word-break preference
		if r == ' ' || r == '\t' || r == '\n' {
			lastSpaceIdx = i
			lastSpaceWidth = width - rw // Width before the space
		}
	}

	// Entire string fits
	return len(s), width
}

// WordBoundaryIndex returns the index of the next or previous word boundary
// For next: returns index of first non-space after current position
// For prev: returns index of first space before current position, or start of word
func WordBoundaryIndex(s string, pos int, next bool) int {
	runes := []rune(s)
	if len(runes) == 0 {
		return 0
	}

	// Clamp position to valid range
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}

	if next {
		// Find next word: skip current word, then skip spaces
		inWord := pos < len(runes) && runes[pos] != ' '
		for i := pos; i < len(runes); i++ {
			isSpace := runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\n'
			if inWord && isSpace {
				// Found space after word
				i++ // Skip the space
				for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\n') {
					i++
				}
				return i
			}
			if !inWord && !isSpace {
				inWord = true
			}
		}
		return len(runes)
	} else {
		// Find previous word: go back to space, then to start of word
		if pos > 0 {
			pos--
		}
		for i := pos; i >= 0; i-- {
			isSpace := runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\n'
			if isSpace {
				return i + 1
			}
		}
		return 0
	}
}
