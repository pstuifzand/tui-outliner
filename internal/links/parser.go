package links

import (
	"regexp"
	"strings"
)

// Link represents an internal link reference in item text
type Link struct {
	ID          string // Item ID being referenced
	DisplayText string // Custom text to display (if provided)
	StartPos    int    // Start position in original text (inclusive)
	EndPos      int    // End position in original text (exclusive)
}

var linkPattern = regexp.MustCompile(`\[\[([^\]\|]+)(?:\|([^\]]+))?\]\]`)

// ParseLinks extracts all wiki-style links from text
// Supports two formats:
//   - [[item_id]] - displays as resolved item text
//   - [[item_id|custom text]] - displays as custom text
func ParseLinks(text string) []Link {
	matches := linkPattern.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}

	links := make([]Link, 0, len(matches))
	for _, match := range matches {
		// match[0:2] = full match indices
		// match[2:4] = group 1 (item ID)
		// match[4:6] = group 2 (display text, may be -1 if not present)

		startPos := match[0]
		endPos := match[1]
		idStart := match[2]
		idEnd := match[3]

		id := text[idStart:idEnd]
		id = strings.TrimSpace(id)

		displayText := ""
		if match[4] != -1 && match[5] != -1 {
			displayText = text[match[4]:match[5]]
			displayText = strings.TrimSpace(displayText)
		}

		links = append(links, Link{
			ID:          id,
			DisplayText: displayText,
			StartPos:    startPos,
			EndPos:      endPos,
		})
	}

	return links
}

// GetDisplayText returns the text that should be shown for this link
// If DisplayText is set, returns that; otherwise returns the ID
func (l *Link) GetDisplayText() string {
	if l.DisplayText != "" {
		return l.DisplayText
	}
	return l.ID
}

// ContainsPosition checks if a text position falls within this link
func (l *Link) ContainsPosition(pos int) bool {
	return pos >= l.StartPos && pos < l.EndPos
}

// IsValidID checks if an ID matches the expected format
func IsValidID(id string) bool {
	// IDs should start with "item_" followed by timestamp and random text
	return strings.HasPrefix(id, "item_") && len(id) > 30
}
