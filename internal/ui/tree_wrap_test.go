package ui

import (
	"strings"
	"testing"
)

func TestWrapTextAtWidth_PreservesLinkBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		wantErr  bool
	}{
		{
			name:     "Link at end of line that would be split",
			text:     "This is a long line with [[item_123456789012345|a link]] here",
			maxWidth: 30,
			wantErr:  false,
		},
		{
			name:     "Link in middle that would be split",
			text:     "Text before [[item_999999999999999|link text]] and after",
			maxWidth: 25,
			wantErr:  false,
		},
		{
			name:     "Multiple links",
			text:     "First [[item_111111111111111|link]] and second [[item_222222222222222|link]]",
			maxWidth: 30,
			wantErr:  false,
		},
		{
			name:     "Very long link that exceeds maxWidth",
			text:     "Before [[item_123456789012345678901234567890|very long display text]] after",
			maxWidth: 20,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapTextAtWidth(tt.text, tt.maxWidth)

			// Verify links are not split
			for _, line := range result {
				// Count open and close brackets
				openCount := strings.Count(line, "[[")
				closeCount := strings.Count(line, "]]")

				// Each line should have equal opens and closes (or none)
				// This ensures links are not split
				if openCount != closeCount {
					t.Errorf("wrapTextAtWidth() line %q has mismatched brackets: [[=%d, ]]=%d",
						line, openCount, closeCount)
				}

				// If there's a [[, ensure it's followed by ]] in the same line
				if strings.Contains(line, "[[") && !strings.Contains(line, "]]") {
					t.Errorf("wrapTextAtWidth() line %q has incomplete link markup", line)
				}
			}

			// Print results for manual inspection
			t.Logf("Input: %q (maxWidth=%d)", tt.text, tt.maxWidth)
			for i, line := range result {
				t.Logf("  Line %d (width=%d): %q", i+1, StringWidth(line), line)
			}
		})
	}
}

func TestWrapTextAtWidth_EmptyAndShortText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     []string
	}{
		{
			name:     "Empty text",
			text:     "",
			maxWidth: 10,
			want:     []string{""},
		},
		{
			name:     "Text shorter than maxWidth",
			text:     "Short [[item_123|link]]",
			maxWidth: 50,
			want:     []string{"Short [[item_123|link]]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapTextAtWidth(tt.text, tt.maxWidth)
			if len(got) != len(tt.want) {
				t.Errorf("wrapTextAtWidth() returned %d lines, want %d", len(got), len(tt.want))
			}
		})
	}
}
