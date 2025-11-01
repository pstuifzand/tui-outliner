package search

import (
	"fmt"
	"testing"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

func TestTokenizer(t *testing.T) {
	tests := []struct {
		input  string
		tokens []TokenType
	}{
		{
			input:  "task",
			tokens: []TokenType{TokenText, TokenEOF},
		},
		{
			input:  "task project",
			tokens: []TokenType{TokenText, TokenText, TokenEOF},
		},
		{
			input:  "task | project",
			tokens: []TokenType{TokenText, TokenOr, TokenText, TokenEOF},
		},
		{
			input:  "task +project",
			tokens: []TokenType{TokenText, TokenAnd, TokenText, TokenEOF},
		},
		{
			input:  "-task",
			tokens: []TokenType{TokenNot, TokenText, TokenEOF},
		},
		{
			input:  "d:>2",
			tokens: []TokenType{TokenFilter, TokenEOF},
		},
		{
			input:  "d:>=2 @type=day",
			tokens: []TokenType{TokenFilter, TokenFilter, TokenEOF},
		},
		{
			input:  "(task | project)",
			tokens: []TokenType{TokenLParen, TokenText, TokenOr, TokenText, TokenRParen, TokenEOF},
		},
		{
			input:  `"multi word"`,
			tokens: []TokenType{TokenText, TokenEOF},
		},
		{
			input:  "@url",
			tokens: []TokenType{TokenFilter, TokenEOF},
		},
		{
			input:  "@date>-7d",
			tokens: []TokenType{TokenFilter, TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			tokens := tokenizer.AllTokens()

			if len(tokens) != len(tt.tokens) {
				t.Fatalf("expected %d tokens, got %d", len(tt.tokens), len(tokens))
			}

			for i, expectedType := range tt.tokens {
				if tokens[i].Type != expectedType {
					t.Errorf("token %d: expected %d, got %d", i, expectedType, tokens[i].Type)
				}
			}
		})
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		query       string
		shouldError bool
		exprType    string
	}{
		{
			query:    "task",
			exprType: "*search.TextExpr",
		},
		{
			query:    "task project",
			exprType: "*search.AndExpr",
		},
		{
			query:    "task | project",
			exprType: "*search.OrExpr",
		},
		{
			query:    "-task",
			exprType: "*search.NotExpr",
		},
		{
			query:    "d:>2",
			exprType: "*search.DepthFilter",
		},
		{
			query:    "@type=day",
			exprType: "*search.AttributeFilter",
		},
		{
			query:    "children:>0",
			exprType: "*search.ChildrenFilter",
		},
		{
			query:    "task d:>2",
			exprType: "*search.AndExpr",
		},
		{
			query:    "(task | project) d:>2",
			exprType: "*search.AndExpr",
		},
		{
			query:    "",
			exprType: "*search.AlwaysMatchExpr",
		},
		{
			query:       "(task",
			shouldError: true,
		},
		{
			query:       "d:>",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)

			if tt.shouldError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err != nil {
				return
			}

			actualType := string([]rune(typeOf(expr))[:20])
			expectedType := string([]rune(tt.exprType)[:20])
			if actualType != expectedType {
				t.Errorf("expected type %s, got %s", tt.exprType, typeOf(expr))
			}
		})
	}
}

func TestDepthFilter(t *testing.T) {
	tests := []struct {
		query   string
		depth   int
		matches bool
	}{
		{"d:0", 0, true},
		{"d:0", 1, false},
		{"d:>0", 0, false},
		{"d:>0", 1, true},
		{"d:>=1", 1, true},
		{"d:<2", 1, true},
		{"d:<2", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := createModelItemAtDepth(tt.depth)
			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("query %s with depth %d: expected %v, got %v", tt.query, tt.depth, tt.matches, matches)
			}
		})
	}
}

func TestChildrenFilter(t *testing.T) {
	tests := []struct {
		query        string
		childCount   int
		matches      bool
	}{
		{"children:0", 0, true},
		{"children:0", 1, false},
		{"children:>0", 0, false},
		{"children:>0", 1, true},
		{"children:5", 5, true},
		{"children:5", 4, false},
		{"children:>=3", 3, true},
		{"children:>=3", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := createModelItemWithChildren(tt.childCount)
			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("query %s with children %d: expected %v, got %v", tt.query, tt.childCount, tt.matches, matches)
			}
		})
	}
}

func TestAttributeDateFilter(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	tests := []struct {
		query   string
		dateVal string
		matches bool
	}{
		// Recent dates
		{
			query:   fmt.Sprintf("@date>=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			matches: true,
		},
		// Older dates
		{
			query:   fmt.Sprintf("@date>=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", weekAgo.Year(), weekAgo.Month(), weekAgo.Day()),
			matches: false,
		},
		// Date range with relative dates
		{
			query:   "@date>-7d",
			dateVal: yesterday.Format("2006-01-02"),
			matches: true,
		},
		{
			query:   "@date<-7d",
			dateVal: monthAgo.Format("2006-01-02"),
			matches: true,
		},
		{
			query:   "@date<-7d",
			dateVal: yesterday.Format("2006-01-02"),
			matches: false,
		},
		// Equality
		{
			query:   fmt.Sprintf("@date=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			matches: true,
		},
		{
			query:   fmt.Sprintf("@date=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: yesterday.Format("2006-01-02"),
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := &model.Item{
				ID:   "test-item",
				Text: "test",
				Metadata: &model.Metadata{
					Created:  time.Now(),
					Modified: time.Now(),
					Attributes: map[string]string{
						"date": tt.dateVal,
					},
				},
			}

			matches := expr.Matches(item)
			if matches != tt.matches {
				t.Errorf("query %s with date %s: expected %v, got %v", tt.query, tt.dateVal, tt.matches, matches)
			}
		})
	}
}

func TestBooleanOperators(t *testing.T) {
	tests := []struct {
		query string
		text  string
		depth int
		matches bool
	}{
		{"task d:>0", "task", 1, true},
		{"task d:>0", "task", 0, false},
		{"task | project", "task", 0, true},
		{"task | project", "project", 0, true},
		{"task | project", "other", 0, false},
		{"-task", "task", 0, false},
		{"-task", "other", 0, true},
		{"(task | project) d:>0", "task", 1, true},
		{"(task | project) d:>0", "other", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := createModelItemWithText(tt.text, tt.depth)
			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("query %s with text %q depth %d: expected %v, got %v", tt.query, tt.text, tt.depth, tt.matches, matches)
			}
		})
	}
}

// Helper functions for tests

func createModelItemAtDepth(depth int) *model.Item {
	item := &model.Item{
		ID:       "test-item",
		Text:     "test",
		Children: make([]*model.Item, 0),
		Metadata: &model.Metadata{
			Created:  time.Now(),
			Modified: time.Now(),
		},
	}

	current := item
	for i := 0; i < depth; i++ {
		parent := &model.Item{
			ID:       fmt.Sprintf("parent-%d", i),
			Text:     fmt.Sprintf("parent-%d", i),
			Children: []*model.Item{current},
			Metadata: &model.Metadata{
				Created:  time.Now(),
				Modified: time.Now(),
			},
		}
		current.Parent = parent
		current = parent
	}

	return item
}

func createModelItemWithChildren(count int) *model.Item {
	item := &model.Item{
		ID:       "parent-item",
		Text:     "parent",
		Children: make([]*model.Item, count),
		Metadata: &model.Metadata{
			Created:  time.Now(),
			Modified: time.Now(),
		},
	}

	for i := 0; i < count; i++ {
		child := &model.Item{
			ID:   fmt.Sprintf("child-%d", i),
			Text: fmt.Sprintf("child-%d", i),
			Parent: item,
			Metadata: &model.Metadata{
				Created:  time.Now(),
				Modified: time.Now(),
			},
		}
		item.Children[i] = child
	}

	return item
}

func createModelItemWithText(text string, depth int) *model.Item {
	item := &model.Item{
		ID:       "test-item",
		Text:     text,
		Children: make([]*model.Item, 0),
		Metadata: &model.Metadata{
			Created:  time.Now(),
			Modified: time.Now(),
		},
	}

	current := item
	for i := 0; i < depth; i++ {
		parent := &model.Item{
			ID:       fmt.Sprintf("parent-%d", i),
			Text:     fmt.Sprintf("parent-%d", i),
			Children: []*model.Item{current},
			Metadata: &model.Metadata{
				Created:  time.Now(),
				Modified: time.Now(),
			},
		}
		current.Parent = parent
		current = parent
	}

	return item
}

func typeOf(v interface{}) string {
	return string([]rune(string([]rune(fmt.Sprintf("%T", v)))))
}
