package ui

import (
	"testing"
)

func TestWrapTextWithLinks_PreservesLinks(t *testing.T) {
	tests := []struct {
		name     string
		rawText  string
		maxWidth int
	}{
		{
			name:     "Link at end of line that would be split",
			rawText:  "This is a long line with [[item_123456789012345|a link]] here",
			maxWidth: 30,
		},
		{
			name:     "Link in middle that would be split",
			rawText:  "Text before [[item_999999999999999|link text]] and after",
			maxWidth: 25,
		},
		{
			name:     "Multiple links",
			rawText:  "First [[item_111111111111111|link]] and second [[item_222222222222222|link]]",
			maxWidth: 30,
		},
		{
			name:     "Very long link that exceeds maxWidth",
			rawText:  "Before [[item_123456789012345678901234567890|very long display text]] after",
			maxWidth: 20,
		},
		{
			name:     "Link with short display text but long ID",
			rawText:  "Check [[item_20240115123456789012345|here]] for details",
			maxWidth: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to display text and extract link ranges
			displayText, linkRanges := convertLinksToDisplayText(tt.rawText)

			// Wrap with link awareness
			wrappedLines := wrapTextWithLinks(displayText, linkRanges, tt.maxWidth)

			t.Logf("Raw text: %q (maxWidth=%d)", tt.rawText, tt.maxWidth)
			t.Logf("Display text: %q", displayText)
			t.Logf("Link ranges: %+v", linkRanges)

			// Verify each wrapped line
			for i, wrapped := range wrappedLines {
				t.Logf("  Line %d (width=%d): %q", i+1, StringWidth(wrapped.Text), wrapped.Text)
				t.Logf("    Link ranges: %+v", wrapped.LinkRanges)

				// Verify link ranges are valid
				for _, lr := range wrapped.LinkRanges {
					if lr.Start < 0 || lr.End > len([]rune(wrapped.Text)) {
						t.Errorf("Invalid link range: Start=%d, End=%d, text length=%d",
							lr.Start, lr.End, len([]rune(wrapped.Text)))
					}
					if lr.Start >= lr.End {
						t.Errorf("Invalid link range: Start=%d >= End=%d", lr.Start, lr.End)
					}
				}
			}

			// Verify links are not split - each link should appear complete in one line
			totalLinkRanges := 0
			for _, wrapped := range wrappedLines {
				totalLinkRanges += len(wrapped.LinkRanges)
			}
			if totalLinkRanges != len(linkRanges) {
				t.Errorf("Expected %d total link ranges across all lines, got %d",
					len(linkRanges), totalLinkRanges)
			}
		})
	}
}

func TestConvertLinksToDisplayText(t *testing.T) {
	tests := []struct {
		name        string
		rawText     string
		wantDisplay string
		wantLinks   int
	}{
		{
			name:        "Link with custom display text",
			rawText:     "Check [[item_123|here]] now",
			wantDisplay: "Check here now",
			wantLinks:   1,
		},
		{
			name:        "Link without display text (shows ID)",
			rawText:     "See [[item_456]] please",
			wantDisplay: "See item_456 please",
			wantLinks:   1,
		},
		{
			name:        "No links",
			rawText:     "Plain text",
			wantDisplay: "Plain text",
			wantLinks:   0,
		},
		{
			name:        "Multiple links",
			rawText:     "[[item_1|First]] and [[item_2|second]]",
			wantDisplay: "First and second",
			wantLinks:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			displayText, linkRanges := convertLinksToDisplayText(tt.rawText)

			if displayText != tt.wantDisplay {
				t.Errorf("Display text = %q, want %q", displayText, tt.wantDisplay)
			}

			if len(linkRanges) != tt.wantLinks {
				t.Errorf("Got %d links, want %d", len(linkRanges), tt.wantLinks)
			}

			t.Logf("Raw: %q → Display: %q, Links: %+v", tt.rawText, displayText, linkRanges)
		})
	}
}
