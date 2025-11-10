package search

import (
	"testing"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

func TestRefExpr(t *testing.T) {
	// Create test items
	targetID := "item_12345"

	tests := []struct {
		name     string
		itemText string
		targetID string
		want     bool
	}{
		{
			name:     "Simple link matches",
			itemText: "This item links to [[item_12345]]",
			targetID: targetID,
			want:     true,
		},
		{
			name:     "Link with custom text matches",
			itemText: "Check this out: [[item_12345|custom text]]",
			targetID: targetID,
			want:     true,
		},
		{
			name:     "Multiple links, one matches",
			itemText: "Links: [[item_99999]] and [[item_12345]]",
			targetID: targetID,
			want:     true,
		},
		{
			name:     "No link to target",
			itemText: "This has no links",
			targetID: targetID,
			want:     false,
		},
		{
			name:     "Link to different item",
			itemText: "Links to [[item_99999]]",
			targetID: targetID,
			want:     false,
		},
		{
			name:     "Partial match should not match",
			itemText: "Contains item_12345 as plain text",
			targetID: targetID,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := model.NewItem(tt.itemText)
			expr := NewRefExpr(tt.targetID)

			got := expr.Matches(item)
			if got != tt.want {
				t.Errorf("RefExpr.Matches() = %v, want %v for text: %q", got, tt.want, tt.itemText)
			}
		})
	}
}

func TestRefExprParsing(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantError bool
	}{
		{
			name:      "Valid ref query",
			query:     "ref:item_12345",
			wantError: false,
		},
		{
			name:      "Ref in combination with other filters",
			query:     "ref:item_12345 | @tag=important",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseQuery(tt.query)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseQuery(%q) error = %v, wantError %v", tt.query, err, tt.wantError)
			}
		})
	}
}
