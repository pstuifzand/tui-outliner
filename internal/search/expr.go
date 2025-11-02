package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// FilterExpr represents a filter expression that can match items
type FilterExpr interface {
	Matches(item *model.Item) bool
	String() string // For debug output
}

// TextExpr matches items whose text contains the search term (case-insensitive)
type TextExpr struct {
	term string
}

func NewTextExpr(term string) *TextExpr {
	return &TextExpr{term: strings.ToLower(term)}
}

func (e *TextExpr) Matches(item *model.Item) bool {
	return strings.Contains(strings.ToLower(item.Text), e.term)
}

func (e *TextExpr) String() string {
	return fmt.Sprintf("text(%q)", e.term)
}

// AlwaysMatchExpr matches all items (for empty queries)
type AlwaysMatchExpr struct{}

func NewAlwaysMatchExpr() *AlwaysMatchExpr {
	return &AlwaysMatchExpr{}
}

func (e *AlwaysMatchExpr) Matches(item *model.Item) bool {
	return true
}

func (e *AlwaysMatchExpr) String() string {
	return "always-match"
}

// AndExpr matches if both left and right match
type AndExpr struct {
	left  FilterExpr
	right FilterExpr
}

func NewAndExpr(left, right FilterExpr) *AndExpr {
	return &AndExpr{left: left, right: right}
}

func (e *AndExpr) Matches(item *model.Item) bool {
	return e.left.Matches(item) && e.right.Matches(item)
}

func (e *AndExpr) String() string {
	return fmt.Sprintf("(and %s %s)", e.left.String(), e.right.String())
}

// OrExpr matches if either left or right matches
type OrExpr struct {
	left  FilterExpr
	right FilterExpr
}

func NewOrExpr(left, right FilterExpr) *OrExpr {
	return &OrExpr{left: left, right: right}
}

func (e *OrExpr) Matches(item *model.Item) bool {
	return e.left.Matches(item) || e.right.Matches(item)
}

func (e *OrExpr) String() string {
	return fmt.Sprintf("(or %s %s)", e.left.String(), e.right.String())
}

// NotExpr matches if the wrapped expression does not match
type NotExpr struct {
	expr FilterExpr
}

func NewNotExpr(expr FilterExpr) *NotExpr {
	return &NotExpr{expr: expr}
}

func (e *NotExpr) Matches(item *model.Item) bool {
	return !e.expr.Matches(item)
}

func (e *NotExpr) String() string {
	return fmt.Sprintf("(not %s)", e.expr.String())
}

// DepthFilter matches items at specific depth levels
type DepthFilter struct {
	op    ComparisonOp
	value int
}

func NewDepthFilter(op ComparisonOp, value string) (*DepthFilter, error) {
	var depth int
	_, err := fmt.Sscanf(value, "%d", &depth)
	if err != nil {
		return nil, fmt.Errorf("invalid depth value: %s", value)
	}
	return &DepthFilter{op: op, value: depth}, nil
}

func (e *DepthFilter) Matches(item *model.Item) bool {
	itemDepth := calculateDepth(item)
	return compare(itemDepth, e.op, e.value)
}

func (e *DepthFilter) String() string {
	return fmt.Sprintf("depth(%s%d)", e.op, e.value)
}

// AttributeFilter matches items with specific attributes
type AttributeFilter struct {
	key   string
	op    string
	value string
}

func NewAttrFilter(key, op, value string) *AttributeFilter {
	return &AttributeFilter{key: key, op: op, value: value}
}

func (e *AttributeFilter) Matches(item *model.Item) bool {
	if item.Metadata == nil || item.Metadata.Attributes == nil {
		return false
	}

	attrVal, exists := item.Metadata.Attributes[e.key]

	if e.op == "" {
		// Just checking for existence
		return exists
	}

	if !exists {
		return e.op == "!="
	}

	switch e.op {
	case "=":
		return attrVal == e.value
	case "!=":
		return attrVal != e.value
	default:
		return false
	}
}

func (e *AttributeFilter) String() string {
	if e.op == "" {
		return fmt.Sprintf("attr(%s)", e.key)
	}
	return fmt.Sprintf("attr(%s%s%s)", e.key, e.op, e.value)
}

// AttributeDateFilter matches items where an attribute contains a date value that matches a date comparison
type AttributeDateFilter struct {
	key   string
	op    ComparisonOp
	value string
}

func NewAttrDateFilter(key string, op ComparisonOp, value string) (*AttributeDateFilter, error) {
	// Validate that the date value is valid
	if !isValidDateValue(value) {
		return nil, fmt.Errorf("invalid date value for attribute filter: %s", value)
	}
	return &AttributeDateFilter{key: key, op: op, value: value}, nil
}

func (e *AttributeDateFilter) Matches(item *model.Item) bool {
	if item.Metadata == nil || item.Metadata.Attributes == nil {
		return false
	}

	attrVal, exists := item.Metadata.Attributes[e.key]
	if !exists {
		return false
	}

	// Parse the attribute value as a date
	attrDate := parseDate(attrVal)
	if attrDate.IsZero() {
		// Attribute value is not a valid date
		return false
	}

	// Parse the comparison date
	compareDate := parseDate(e.value)
	if compareDate.IsZero() {
		return false
	}

	// Perform the comparison
	switch e.op {
	case OpGreater:
		return attrDate.After(compareDate)
	case OpGreaterEqual:
		return attrDate.After(compareDate) || attrDate.Equal(compareDate)
	case OpLess:
		return attrDate.Before(compareDate)
	case OpLessEqual:
		return attrDate.Before(compareDate) || attrDate.Equal(compareDate)
	case OpEqual:
		return attrDate.Format("2006-01-02") == compareDate.Format("2006-01-02")
	case OpNotEqual:
		return attrDate.Format("2006-01-02") != compareDate.Format("2006-01-02")
	default:
		return false
	}
}

func (e *AttributeDateFilter) String() string {
	return fmt.Sprintf("attr(%s%s%s)", e.key, e.op, e.value)
}

// DateFilter matches items based on creation or modification date
type DateFilter struct {
	filterType FilterType
	op         ComparisonOp
	value      string
}

func NewDateFilter(filterType FilterType, op ComparisonOp, value string) (*DateFilter, error) {
	// Validate the date value format
	if !isValidDateValue(value) {
		return nil, fmt.Errorf("invalid date value: %s", value)
	}
	return &DateFilter{filterType: filterType, op: op, value: value}, nil
}

func (e *DateFilter) Matches(item *model.Item) bool {
	var targetTime time.Time
	if item.Metadata == nil {
		return false
	}

	if e.filterType == FilterTypeCreated {
		targetTime = item.Metadata.Created
	} else if e.filterType == FilterTypeModified {
		targetTime = item.Metadata.Modified
	}

	if targetTime.IsZero() {
		return false
	}

	compareTime := parseDate(e.value)
	if compareTime.IsZero() {
		return false
	}

	switch e.op {
	case OpGreater:
		return targetTime.After(compareTime)
	case OpGreaterEqual:
		return targetTime.After(compareTime) || targetTime.Equal(compareTime)
	case OpLess:
		return targetTime.Before(compareTime)
	case OpLessEqual:
		return targetTime.Before(compareTime) || targetTime.Equal(compareTime)
	case OpEqual:
		return targetTime.Format("2006-01-02") == compareTime.Format("2006-01-02")
	case OpNotEqual:
		return targetTime.Format("2006-01-02") != compareTime.Format("2006-01-02")
	default:
		return false
	}
}

func (e *DateFilter) String() string {
	typeStr := "created"
	if e.filterType == FilterTypeModified {
		typeStr = "modified"
	}
	return fmt.Sprintf("%s(%s%s)", typeStr, e.op, e.value)
}

// ChildrenFilter matches items based on the number of children
type ChildrenFilter struct {
	op    ComparisonOp
	value int
}

func NewChildrenFilter(op ComparisonOp, value string) (*ChildrenFilter, error) {
	var count int
	_, err := fmt.Sscanf(value, "%d", &count)
	if err != nil {
		return nil, fmt.Errorf("invalid children count: %s", value)
	}
	return &ChildrenFilter{op: op, value: count}, nil
}

func (e *ChildrenFilter) Matches(item *model.Item) bool {
	childCount := len(item.Children)
	return compare(childCount, e.op, e.value)
}

func (e *ChildrenFilter) String() string {
	return fmt.Sprintf("children(%s%d)", e.op, e.value)
}

// ParentFilter matches items whose parent matches the inner filter
type ParentFilter struct {
	inner FilterExpr
}

func NewParentFilter(inner FilterExpr) *ParentFilter {
	return &ParentFilter{inner: inner}
}

func (e *ParentFilter) Matches(item *model.Item) bool {
	if item.Parent == nil {
		return false
	}
	return e.inner.Matches(item.Parent)
}

func (e *ParentFilter) String() string {
	return fmt.Sprintf("parent(%s)", e.inner.String())
}

// AncestorFilter matches items that have an ancestor matching the inner filter
type AncestorFilter struct {
	inner FilterExpr
}

func NewAncestorFilter(inner FilterExpr) *AncestorFilter {
	return &AncestorFilter{inner: inner}
}

func (e *AncestorFilter) Matches(item *model.Item) bool {
	current := item.Parent
	for current != nil {
		if e.inner.Matches(current) {
			return true
		}
		current = current.Parent
	}
	return false
}

func (e *AncestorFilter) String() string {
	return fmt.Sprintf("ancestor(%s)", e.inner.String())
}

// Helper functions

// calculateDepth returns the depth of an item in the tree (root = 0)
func calculateDepth(item *model.Item) int {
	depth := 0
	current := item
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	return depth
}

// compare performs a comparison between two integers based on the operator
func compare(a int, op ComparisonOp, b int) bool {
	switch op {
	case OpGreater:
		return a > b
	case OpGreaterEqual:
		return a >= b
	case OpLess:
		return a < b
	case OpLessEqual:
		return a <= b
	case OpEqual:
		return a == b
	case OpNotEqual:
		return a != b
	default:
		return false
	}
}

// isValidDateValue checks if a date value is in a valid format
func isValidDateValue(value string) bool {
	// Relative dates: -1d, -7d, -30d, -1w, -4w, -1m, -6m, -1y
	if strings.HasPrefix(value, "-") {
		suffix := value[len(value)-1:]
		return suffix == "d" || suffix == "w" || suffix == "m" || suffix == "y"
	}
	// Absolute dates: YYYY-MM-DD
	return len(value) == 10 && value[4] == '-' && value[7] == '-'
}

// parseDate parses a date value (relative or absolute) into a time.Time
func parseDate(value string) time.Time {
	now := time.Now()

	// Relative dates
	if strings.HasPrefix(value, "-") {
		var amount int
		var suffix string
		_, err := fmt.Sscanf(value, "-%d%s", &amount, &suffix)
		if err != nil || len(suffix) != 1 {
			return time.Time{}
		}

		switch suffix {
		case "d":
			return now.AddDate(0, 0, -amount)
		case "w":
			return now.AddDate(0, 0, -amount*7)
		case "m":
			return now.AddDate(0, -amount, 0)
		case "y":
			return now.AddDate(-amount, 0, 0)
		default:
			return time.Time{}
		}
	}

	// Absolute dates: YYYY-MM-DD
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}
	}
	return t
}

// GetMatchingItems returns all items that match the given filter expression
func GetMatchingItems(outline *model.Outline, filterExpr FilterExpr) []*model.Item {
	var matches []*model.Item
	for _, item := range outline.GetAllItems() {
		if filterExpr.Matches(item) {
			matches = append(matches, item)
		}
	}
	return matches
}

// GetFirstMatchingItem returns the first item that matches the given filter expression, or nil if none match
func GetFirstMatchingItem(outline *model.Outline, filterExpr FilterExpr) *model.Item {
	for _, item := range outline.GetAllItems() {
		if filterExpr.Matches(item) {
			return item
		}
	}
	return nil
}

func GetFirstByQuery(outline *model.Outline, query string) (*model.Item, error) {
	filterExpr, err := ParseQuery(query)
	if err != nil {
		return nil, err
	}
	return GetFirstMatchingItem(outline, filterExpr), nil
}

func GetAlllByQuery(outline *model.Outline, query string) ([]*model.Item, error) {
	filterExpr, err := ParseQuery(query)
	if err != nil {
		return nil, err
	}
	return GetMatchingItems(outline, filterExpr), nil
}
