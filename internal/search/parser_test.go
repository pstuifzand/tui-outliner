package search

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
		{
			input:  "~task",
			tokens: []TokenType{TokenFilter, TokenEOF},
		},
		{
			input:  "~task ~project",
			tokens: []TokenType{TokenFilter, TokenFilter, TokenEOF},
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
		{
			query:    "~task",
			exprType: "*search.FuzzyExpr",
		},
		{
			query:    "~task ~project",
			exprType: "*search.AndExpr",
		},
		{
			query:    "-~task",
			exprType: "*search.NotExpr",
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
		query      string
		childCount int
		matches    bool
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

func TestAttributeDateExprParser(t *testing.T) {
	tests := []struct {
		query    string
		key      string
		value    string
		op       ComparisonOp
		shouldErr bool
	}{
		{
			query:    "@date>=2025-10-10",
			key:      "date",
			value:    "2025-10-10",
			op:       OpGreaterEqual,
		},
		{
			query:    "@date>2025-10-10",
			key:      "date",
			value:    "2025-10-10",
			op:       OpGreater,
		},
		{
			query:    "@date<2025-10-10",
			key:      "date",
			value:    "2025-10-10",
			op:       OpLess,
		},
		{
			query:    "@date<=2025-10-10",
			key:      "date",
			value:    "2025-10-10",
			op:       OpLessEqual,
		},
		{
			query:    "@date=2025-10-10",
			key:      "date",
			value:    "2025-10-10",
			op:       OpEqual,
		},
		{
			query:    "@date!=2025-10-10",
			key:      "date",
			value:    "2025-10-10",
			op:       OpNotEqual,
		},
		{
			query:    "@deadline>-7d",
			key:      "deadline",
			value:    "-7d",
			op:       OpGreater,
		},
		{
			query:    "@duedate<=-30d",
			key:      "duedate",
			value:    "-30d",
			op:       OpLessEqual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			assert.NoErrorf(t, err, "%q should parse without error", tt.query)
			assert.Equal(t, tt.key, expr.(*AttributeDateFilter).key, "key should match")
			assert.Equal(t, tt.value, expr.(*AttributeDateFilter).value, "value should match")
			assert.Equal(t, tt.op, expr.(*AttributeDateFilter).op, "operator should match")
		})
	}
}

func TestAttributeDateComparisonOpsAllOperators(t *testing.T) {
	// Test that all date comparison operators parse and can be used
	// Since parseDate has timing issues with relative dates, this test focuses on
	// verifying that each operator is correctly parsed and applied to filter expressions

	tests := []struct {
		name             string
		query            string
		expectsOp        ComparisonOp
		expectsAttribute string
	}{
		// All six comparison operators
		{
			name:             "OpGreater",
			query:            "@date>-7d",
			expectsOp:        OpGreater,
			expectsAttribute: "date",
		},
		{
			name:             "OpGreaterEqual",
			query:            "@date>=-7d",
			expectsOp:        OpGreaterEqual,
			expectsAttribute: "date",
		},
		{
			name:             "OpLess",
			query:            "@date<-7d",
			expectsOp:        OpLess,
			expectsAttribute: "date",
		},
		{
			name:             "OpLessEqual",
			query:            "@date<=-7d",
			expectsOp:        OpLessEqual,
			expectsAttribute: "date",
		},
		{
			name:             "OpEqual",
			query:            "@date=-7d",
			expectsOp:        OpEqual,
			expectsAttribute: "date",
		},
		{
			name:             "OpNotEqual",
			query:            "@date!=-7d",
			expectsOp:        OpNotEqual,
			expectsAttribute: "date",
		},
		// Different attribute keys
		{
			name:             "OpGreater with deadline",
			query:            "@deadline>-30d",
			expectsOp:        OpGreater,
			expectsAttribute: "deadline",
		},
		{
			name:             "OpLess with duedate",
			query:            "@duedate<-14d",
			expectsOp:        OpLess,
			expectsAttribute: "duedate",
		},
		{
			name:             "OpEqual with expiry",
			query:            "@expiry=-0d",
			expectsOp:        OpEqual,
			expectsAttribute: "expiry",
		},
		{
			name:             "OpNotEqual with modified",
			query:            "@modified!=-1w",
			expectsOp:        OpNotEqual,
			expectsAttribute: "modified",
		},
		// Multiple different operators
		{
			name:             "OpGreater with month range",
			query:            "@created>-1m",
			expectsOp:        OpGreater,
			expectsAttribute: "created",
		},
		{
			name:             "OpLessEqual with year range",
			query:            "@archived<=-1y",
			expectsOp:        OpLessEqual,
			expectsAttribute: "archived",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			// Verify it's an AttributeDateFilter
			dateFilter, ok := expr.(*AttributeDateFilter)
			if !ok {
				t.Fatalf("expected *AttributeDateFilter, got %T", expr)
			}

			// Verify the operator is correct
			if dateFilter.op != tt.expectsOp {
				t.Errorf("expected operator %q, got %q", tt.expectsOp, dateFilter.op)
			}

			// Verify the attribute key is correct
			if dateFilter.key != tt.expectsAttribute {
				t.Errorf("expected attribute key %q, got %q", tt.expectsAttribute, dateFilter.key)
			}
		})
	}
}

func TestAttributeDatePositiveFilter(t *testing.T) {
	// Test absolute date filters (YYYY-MM-DD format)
	// The "Positive" name refers to testing the filter with dates that should match
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	weekAgo := now.AddDate(0, 0, -7)

	tests := []struct {
		query   string
		dateVal string
		matches bool
	}{
		// Absolute date comparisons
		{
			query:   fmt.Sprintf("@date<=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			matches: true,
		},
		{
			query:   fmt.Sprintf("@date<=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", weekAgo.Year(), weekAgo.Month(), weekAgo.Day()),
			matches: true,
		},
		{
			query:   fmt.Sprintf("@date>%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", yesterday.Year(), yesterday.Month(), yesterday.Day()),
			matches: false,
		},
		{
			query:   fmt.Sprintf("@date>%d-%02d-%02d", weekAgo.Year(), weekAgo.Month(), weekAgo.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			matches: true,
		},
		// Equality with absolute dates
		{
			query:   fmt.Sprintf("@date=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			matches: true,
		},
		{
			query:   fmt.Sprintf("@date=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", yesterday.Year(), yesterday.Month(), yesterday.Day()),
			matches: false,
		},
		// Negative equality
		{
			query:   fmt.Sprintf("@date!=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", yesterday.Year(), yesterday.Month(), yesterday.Day()),
			matches: true,
		},
		{
			query:   fmt.Sprintf("@date!=%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
			dateVal: fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day()),
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

func TestAttributeDateNegativeFilter(t *testing.T) {
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

func TestAttributeDateFutureRelativeDates(t *testing.T) {
	// Test all date comparison operators with future relative date queries
	// Item dates are always stored as absolute dates (YYYY-MM-DD)
	// Queries use relative dates (+Nd format) to compare against items
	now := time.Now()
	future7d := now.AddDate(0, 0, 7)
	future14d := now.AddDate(0, 0, 14)
	future30d := now.AddDate(0, 0, 30)
	past7d := now.AddDate(0, 0, -7)
	past30d := now.AddDate(0, 0, -30)
	future1y := now.AddDate(1, 0, 0)

	tests := []struct {
		name     string
		query    string
		itemDate time.Time // stored as YYYY-MM-DD
		matches  bool
	}{
		// OpGreater: date > reference (item date is MORE in future)
		{
			name:     "OpGreater: 30d future > 7d future",
			query:    "@date>+7d",
			itemDate: future30d,
			matches:  true,
		},
		{
			name:     "OpGreater: 7d future > 7d future (same boundary)",
			query:    "@date>+7d",
			itemDate: future7d,
			matches:  false,
		},
		{
			name:     "OpGreater: 1d future > 7d future (earlier)",
			query:    "@date>+7d",
			itemDate: now.AddDate(0, 0, 1),
			matches:  false,
		},

		// OpGreaterEqual: date >= reference
		{
			name:     "OpGreaterEqual: 14d future >= 7d past",
			query:    "@date>=-7d",
			itemDate: future14d,
			matches:  true,
		},
		{
			name:     "OpGreaterEqual: today >= 7d past",
			query:    "@date>=-7d",
			itemDate: now,
			matches:  true,
		},
		{
			name:     "OpGreaterEqual: 1d future >= 7d past (more recent)",
			query:    "@date>=-7d",
			itemDate: now.AddDate(0, 0, 1),
			matches:  true,
		},

		// OpLess: date < reference (item date is LESS in future, or in past)
		{
			name:     "OpLess: 1d future < 7d future",
			query:    "@date<+7d",
			itemDate: now.AddDate(0, 0, 1),
			matches:  true,
		},
		{
			name:     "OpLess: 7d past < 7d future",
			query:    "@date<+7d",
			itemDate: past7d,
			matches:  true,
		},
		{
			name:     "OpLess: 14d future < 7d future (later)",
			query:    "@date<+7d",
			itemDate: future14d,
			matches:  false,
		},

		// OpLessEqual: date <= reference
		{
			name:     "OpLessEqual: 1d future <= 7d future",
			query:    "@date<=+7d",
			itemDate: now.AddDate(0, 0, 1),
			matches:  true,
		},
		{
			name:     "OpLessEqual: 7d future <= 7d future (equal boundary)",
			query:    "@date<=+7d",
			itemDate: future7d,
			matches:  true,
		},
		{
			name:     "OpLessEqual: 14d future <= 7d future",
			query:    "@date<=+7d",
			itemDate: future14d,
			matches:  false,
		},

		// OpEqual: date equals reference (same day)
		{
			name:     "OpEqual: 7d future = 7d future",
			query:    "@date=+7d",
			itemDate: future7d,
			matches:  true,
		},
		{
			name:     "OpEqual: 1d future = 7d future",
			query:    "@date=+7d",
			itemDate: now.AddDate(0, 0, 1),
			matches:  false,
		},
		{
			name:     "OpEqual: 7d past = 7d future",
			query:    "@date=+7d",
			itemDate: past7d,
			matches:  false,
		},

		// OpNotEqual: date does not equal reference
		{
			name:     "OpNotEqual: 1d future != 7d future",
			query:    "@date!=+7d",
			itemDate: now.AddDate(0, 0, 1),
			matches:  true,
		},
		{
			name:     "OpNotEqual: 7d future != 7d future",
			query:    "@date!=+7d",
			itemDate: future7d,
			matches:  false,
		},
		{
			name:     "OpNotEqual: 7d past != 7d future",
			query:    "@date!=+7d",
			itemDate: past7d,
			matches:  true,
		},

		// Mixed past and future comparisons
		{
			name:     "OpGreater: future date > past reference",
			query:    "@date>-7d",
			itemDate: future7d,
			matches:  true,
		},
		{
			name:     "OpLess: past date < future reference",
			query:    "@date<+7d",
			itemDate: past7d,
			matches:  true,
		},
		{
			name:     "OpGreaterEqual: today >= 7d past",
			query:    "@date>=-7d",
			itemDate: now,
			matches:  true,
		},
		{
			name:     "OpLessEqual: 7d past <= 7d future",
			query:    "@date<=+7d",
			itemDate: past7d,
			matches:  true,
		},

		// Different attribute keys
		{
			name:     "OpGreater with deadline: future deadline",
			query:    "@deadline>+7d",
			itemDate: future14d,
			matches:  true,
		},
		{
			name:     "OpLess with duedate: past due date",
			query:    "@duedate<-7d",
			itemDate: past30d,
			matches:  true,
		},
		{
			name:     "OpEqual with expiry: exact expiry date",
			query:    "@expiry=+30d",
			itemDate: future30d,
			matches:  true,
		},

		// Week, month, and year ranges
		{
			name:     "OpGreater: 14d future > 1 week past",
			query:    "@date>-1w",
			itemDate: future14d,
			matches:  true,
		},
		{
			name:     "OpLess: 15d past < 30d past (approximately 1 month)",
			query:    "@date<-30d",
			itemDate: now.AddDate(0, 0, -15),
			matches:  false,
		},
		{
			name:     "OpLess: 60d past < 1 month past",
			query:    "@date<-1m",
			itemDate: now.AddDate(0, 0, -60),
			matches:  true,
		},
		{
			name:     "OpGreaterEqual: 1 year future >= 1 year past",
			query:    "@date>=-1y",
			itemDate: future1y,
			matches:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			// Extract attribute key from query
			var attrKey string
			if strings.Contains(tt.query, "@date") {
				attrKey = "date"
			} else if strings.Contains(tt.query, "@deadline") {
				attrKey = "deadline"
			} else if strings.Contains(tt.query, "@duedate") {
				attrKey = "duedate"
			} else if strings.Contains(tt.query, "@expiry") {
				attrKey = "expiry"
			}

			// Convert time.Time to YYYY-MM-DD format for storage
			dateStr := tt.itemDate.Format("2006-01-02")

			item := &model.Item{
				ID:   "test-item",
				Text: "test",
				Metadata: &model.Metadata{
					Created:  time.Now(),
					Modified: time.Now(),
					Attributes: map[string]string{
						attrKey: dateStr,
					},
				},
			}

			matches := expr.Matches(item)
			if matches != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

func TestBooleanOperators(t *testing.T) {
	tests := []struct {
		query   string
		text    string
		depth   int
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
			ID:     fmt.Sprintf("child-%d", i),
			Text:   fmt.Sprintf("child-%d", i),
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

func TestFuzzyFilter(t *testing.T) {
	tests := []struct {
		query   string
		text    string
		matches bool
	}{
		// Exact matches
		{"~task", "task", true},
		{"~task", "Task", true}, // Case-insensitive
		// Fuzzy matches (letters in order, not necessarily consecutive)
		{"~tsk", "task", true},
		{"~tst", "test", true},
		{"~abc", "a b c", true},
		{"~hlo", "hello", true},
		// Non-matches
		{"~task", "project", false},
		{"~xyz", "abc", false},
		{"~aaa", "ab", false},
		// Partial fuzzy matching
		{"~write", "rewrite", true},
		{"~doc", "documentation", true},
		// Case insensitive fuzzy
		{"~TsK", "task", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_matches_%s", tt.query, tt.text), func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := createModelItemWithText(tt.text, 0)
			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("query %s with text %q: expected %v, got %v", tt.query, tt.text, tt.matches, matches)
			}
		})
	}
}

func TestFuzzyHighlightPositions(t *testing.T) {
	tests := []struct {
		text     string
		query    string
		expected []int
	}{
		{
			text:     "task",
			query:    "task",
			expected: []int{0, 1, 2, 3},
		},
		{
			text:     "task",
			query:    "tsk",
			expected: []int{0, 2, 3},
		},
		{
			text:     "documentation",
			query:    "doc",
			expected: []int{0, 1, 2},
		},
		{
			text:     "documentation",
			query:    "dcm",
			expected: []int{0, 2, 4}, // d, c, m (first occurrence of each)
		},
		{
			text:     "hello world",
			query:    "hw",
			expected: []int{0, 6},
		},
		{
			text:     "programming",
			query:    "png",
			expected: []int{0, 9, 10}, // p, n, g (first occurrence of each)
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.text, tt.query), func(t *testing.T) {
			expr := NewFuzzyExpr(tt.query)
			positions := expr.GetMatchPositions(tt.text)

			if len(positions) != len(tt.expected) {
				t.Errorf("expected %d positions, got %d", len(tt.expected), len(positions))
			}

			for i, pos := range positions {
				if i < len(tt.expected) && pos != tt.expected[i] {
					t.Errorf("position %d: expected %d, got %d", i, tt.expected[i], pos)
				}
			}
		})
	}
}

func typeOf(v interface{}) string {
	return string([]rune(string([]rune(fmt.Sprintf("%T", v)))))
}

// Tests for parent/child filters with quantifiers

func TestChildFilterQuantifiers(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		childTexts []string
		matches    bool
	}{
		// Some quantifier (default)
		{
			name:       "child:task with one match",
			query:      "child:task",
			childTexts: []string{"task", "other"},
			matches:    true,
		},
		{
			name:       "child:task with no matches",
			query:      "child:task",
			childTexts: []string{"other", "more"},
			matches:    false,
		},
		{
			name:       "child:task with no children",
			query:      "child:task",
			childTexts: []string{},
			matches:    false,
		},
		// All quantifier
		{
			name:       "+child:task all match",
			query:      "+child:task",
			childTexts: []string{"task", "task"},
			matches:    true,
		},
		{
			name:       "+child:task some don't match",
			query:      "+child:task",
			childTexts: []string{"task", "other"},
			matches:    false,
		},
		{
			name:       "+child:task no children",
			query:      "+child:task",
			childTexts: []string{},
			matches:    false,
		},
		// None quantifier
		{
			name:       "-child:task none match",
			query:      "-child:task",
			childTexts: []string{"other", "more"},
			matches:    true,
		},
		{
			name:       "-child:task one matches",
			query:      "-child:task",
			childTexts: []string{"task", "other"},
			matches:    false,
		},
		{
			name:       "-child:task no children",
			query:      "-child:task",
			childTexts: []string{},
			matches:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := createItemWithChildren(tt.childTexts)
			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

func TestDescendantFilterQuantifiers(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		buildTree    func() *model.Item
		matches      bool
		description  string
	}{
		// Some quantifier (default)
		{
			name:  "child*:task some match",
			query: "child*:task",
			buildTree: func() *model.Item {
				// root -> child1("task") -> grandchild1("other")
				// root -> child2("other")
				root := &model.Item{ID: "root", Text: "root", Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				child1 := &model.Item{ID: "child1", Text: "task", Parent: root, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				grandchild1 := &model.Item{ID: "grandchild1", Text: "other", Parent: child1, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				child1.Children = []*model.Item{grandchild1}
				child2 := &model.Item{ID: "child2", Text: "other", Parent: root, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				root.Children = []*model.Item{child1, child2}
				return root
			},
			matches:     true,
			description: "Some descendants match",
		},
		{
			name:  "child*:task none match",
			query: "child*:task",
			buildTree: func() *model.Item {
				root := &model.Item{ID: "root", Text: "root", Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				child1 := &model.Item{ID: "child1", Text: "other", Parent: root, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				root.Children = []*model.Item{child1}
				return root
			},
			matches:     false,
			description: "No descendants match",
		},
		// All quantifier
		{
			name:  "+child*:task all match",
			query: "+child*:task",
			buildTree: func() *model.Item {
				root := &model.Item{ID: "root", Text: "root", Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				child1 := &model.Item{ID: "child1", Text: "task", Parent: root, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				grandchild1 := &model.Item{ID: "grandchild1", Text: "task", Parent: child1, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				child1.Children = []*model.Item{grandchild1}
				root.Children = []*model.Item{child1}
				return root
			},
			matches:     true,
			description: "All descendants match",
		},
		{
			name:  "+child*:task some don't match",
			query: "+child*:task",
			buildTree: func() *model.Item {
				root := &model.Item{ID: "root", Text: "root", Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				child1 := &model.Item{ID: "child1", Text: "task", Parent: root, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				child2 := &model.Item{ID: "child2", Text: "other", Parent: root, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				root.Children = []*model.Item{child1, child2}
				return root
			},
			matches:     false,
			description: "Not all descendants match",
		},
		// None quantifier
		{
			name:  "-child*:task none match",
			query: "-child*:task",
			buildTree: func() *model.Item {
				root := &model.Item{ID: "root", Text: "root", Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				child1 := &model.Item{ID: "child1", Text: "other", Parent: root, Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()}}
				root.Children = []*model.Item{child1}
				return root
			},
			matches:     true,
			description: "No descendants match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := tt.buildTree()
			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.matches, matches)
			}
		})
	}
}

func TestAncestorFilterQuantifiers(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		ancestorTexts []string  // From immediate parent to root
		matches     bool
	}{
		// Some quantifier (default) - equivalent to current ancestor behavior
		{
			name:          "parent*:project some match",
			query:         "parent*:project",
			ancestorTexts: []string{"project", "other"},
			matches:       true,
		},
		{
			name:          "parent*:project no match",
			query:         "parent*:project",
			ancestorTexts: []string{"other", "more"},
			matches:       false,
		},
		{
			name:          "parent*:project no ancestors (root)",
			query:         "parent*:project",
			ancestorTexts: []string{},
			matches:       false,
		},
		// All quantifier
		{
			name:          "+parent*:project all match",
			query:         "+parent*:project",
			ancestorTexts: []string{"project", "project"},
			matches:       true,
		},
		{
			name:          "+parent*:project some don't match",
			query:         "+parent*:project",
			ancestorTexts: []string{"project", "other"},
			matches:       false,
		},
		{
			name:          "+parent*:project no ancestors (root)",
			query:         "+parent*:project",
			ancestorTexts: []string{},
			matches:       true,  // Vacuously true
		},
		// None quantifier
		{
			name:          "-parent*:project none match",
			query:         "-parent*:project",
			ancestorTexts: []string{"other", "more"},
			matches:       true,
		},
		{
			name:          "-parent*:project one matches",
			query:         "-parent*:project",
			ancestorTexts: []string{"project", "other"},
			matches:       false,
		},
		{
			name:          "-parent*:project no ancestors (root)",
			query:         "-parent*:project",
			ancestorTexts: []string{},
			matches:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := createItemWithAncestors(tt.ancestorTexts)
			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

func TestParentFilterWithNegation(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		parentText string
		hasParent  bool
		matches    bool
	}{
		{
			name:       "parent:project matches",
			query:      "parent:project",
			parentText: "project",
			hasParent:  true,
			matches:    true,
		},
		{
			name:       "parent:project doesn't match",
			query:      "parent:project",
			parentText: "other",
			hasParent:  true,
			matches:    false,
		},
		{
			name:       "-parent:project matches (parent is other)",
			query:      "-parent:project",
			parentText: "other",
			hasParent:  true,
			matches:    true,
		},
		{
			name:       "-parent:project doesn't match (parent is project)",
			query:      "-parent:project",
			parentText: "project",
			hasParent:  true,
			matches:    false,
		},
		{
			name:       "parent:project no parent",
			query:      "parent:project",
			hasParent:  false,
			matches:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			var item *model.Item
			if tt.hasParent {
				item = createItemWithAncestors([]string{tt.parentText})
			} else {
				item = &model.Item{
					ID:       "root",
					Text:     "root",
					Metadata: &model.Metadata{Created: time.Now(), Modified: time.Now()},
				}
			}

			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

func TestComplexParentChildQueries(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		buildTree func() *model.Item
		matches   bool
	}{
		{
			name:  "item with all children done",
			query: "+child:@status=done",
			buildTree: func() *model.Item {
				root := &model.Item{
					ID:   "root",
					Text: "root",
					Metadata: &model.Metadata{
						Created:  time.Now(),
						Modified: time.Now(),
					},
				}
				child1 := &model.Item{
					ID:   "child1",
					Text: "task1",
					Parent: root,
					Metadata: &model.Metadata{
						Created:    time.Now(),
						Modified:   time.Now(),
						Attributes: map[string]string{"status": "done"},
					},
				}
				child2 := &model.Item{
					ID:   "child2",
					Text: "task2",
					Parent: root,
					Metadata: &model.Metadata{
						Created:    time.Now(),
						Modified:   time.Now(),
						Attributes: map[string]string{"status": "done"},
					},
				}
				root.Children = []*model.Item{child1, child2}
				return root
			},
			matches: true,
		},
		{
			name:  "item with some children not done",
			query: "+child:@status=done",
			buildTree: func() *model.Item {
				root := &model.Item{
					ID:   "root",
					Text: "root",
					Metadata: &model.Metadata{
						Created:  time.Now(),
						Modified: time.Now(),
					},
				}
				child1 := &model.Item{
					ID:   "child1",
					Text: "task1",
					Parent: root,
					Metadata: &model.Metadata{
						Created:    time.Now(),
						Modified:   time.Now(),
						Attributes: map[string]string{"status": "done"},
					},
				}
				child2 := &model.Item{
					ID:   "child2",
					Text: "task2",
					Parent: root,
					Metadata: &model.Metadata{
						Created:    time.Now(),
						Modified:   time.Now(),
						Attributes: map[string]string{"status": "todo"},
					},
				}
				root.Children = []*model.Item{child1, child2}
				return root
			},
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			item := tt.buildTree()
			matches := expr.Matches(item)

			if matches != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, matches)
			}
		})
	}
}

// Helper functions

func createItemWithChildren(childTexts []string) *model.Item {
	item := &model.Item{
		ID:       "parent",
		Text:     "parent",
		Children: make([]*model.Item, len(childTexts)),
		Metadata: &model.Metadata{
			Created:  time.Now(),
			Modified: time.Now(),
		},
	}

	for i, text := range childTexts {
		child := &model.Item{
			ID:     fmt.Sprintf("child-%d", i),
			Text:   text,
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

func createItemWithAncestors(ancestorTexts []string) *model.Item {
	item := &model.Item{
		ID:       "item",
		Text:     "item",
		Children: make([]*model.Item, 0),
		Metadata: &model.Metadata{
			Created:  time.Now(),
			Modified: time.Now(),
		},
	}

	current := item
	for i, text := range ancestorTexts {
		parent := &model.Item{
			ID:       fmt.Sprintf("ancestor-%d", i),
			Text:     text,
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
