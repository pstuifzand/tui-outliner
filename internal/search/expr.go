package search

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// FilterExpr represents a filter expression that can match items
type FilterExpr interface {
	Matches(item *model.Item) bool
	String() string // For debug output
}

// Quantifier represents how many items must match a filter (for multi-item filters like children, ancestors)
type Quantifier int

const (
	QuantifierSome Quantifier = iota // At least one must match (default)
	QuantifierAll                    // All must match
	QuantifierNone                   // None must match
)

func (q Quantifier) String() string {
	switch q {
	case QuantifierSome:
		return "some"
	case QuantifierAll:
		return "all"
	case QuantifierNone:
		return "none"
	default:
		return "unknown"
	}
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

// FuzzyExpr matches items whose text fuzzy-matches the search term (case-insensitive)
type FuzzyExpr struct {
	term string
}

func NewFuzzyExpr(term string) *FuzzyExpr {
	return &FuzzyExpr{term: strings.ToLower(term)}
}

func (e *FuzzyExpr) Matches(item *model.Item) bool {
	return fuzzy.MatchFold(e.term, strings.ToLower(item.Text))
}

func (e *FuzzyExpr) String() string {
	return fmt.Sprintf("fuzzy(%q)", e.term)
}

// RegexExpr matches items whose text matches a regular expression pattern
type RegexExpr struct {
	pattern string
	re      *regexp.Regexp
}

func NewRegexExpr(pattern string) (*RegexExpr, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}
	return &RegexExpr{pattern: pattern, re: re}, nil
}

func (e *RegexExpr) Matches(item *model.Item) bool {
	return e.re.MatchString(item.Text)
}

func (e *RegexExpr) String() string {
	return fmt.Sprintf("regex(/%s/)", e.pattern)
}

// GetMatchPositions returns the positions of characters that match the fuzzy query
// Returns a list of indices in the text that correspond to matched characters
func (e *FuzzyExpr) GetMatchPositions(text string) []int {
	if e.term == "" {
		return nil
	}

	lowerText := strings.ToLower(text)
	var positions []int
	textIdx := 0

	// For each character in the search term, find it in the text starting from textIdx
	for _, termChar := range e.term {
		found := false
		for i := textIdx; i < len(lowerText); i++ {
			if rune(lowerText[i]) == termChar {
				positions = append(positions, i)
				textIdx = i + 1
				found = true
				break
			}
		}
		if !found {
			// Character not found - this shouldn't happen if Matches() returned true
			break
		}
	}

	return positions
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

	switch e.filterType {
	case FilterTypeCreated:
		targetTime = item.Metadata.Created
	case FilterTypeModified:
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

// AncestorFilter matches items based on their ancestors (parent* in search syntax)
type AncestorFilter struct {
	inner      FilterExpr
	quantifier Quantifier
}

func NewAncestorFilter(inner FilterExpr) *AncestorFilter {
	return &AncestorFilter{inner: inner, quantifier: QuantifierSome}
}

func NewAncestorFilterWithQuantifier(inner FilterExpr, quantifier Quantifier) *AncestorFilter {
	return &AncestorFilter{inner: inner, quantifier: quantifier}
}

func (e *AncestorFilter) Matches(item *model.Item) bool {
	var ancestors []*model.Item
	current := item.Parent
	for current != nil {
		ancestors = append(ancestors, current)
		current = current.Parent
	}

	switch e.quantifier {
	case QuantifierSome:
		// At least one ancestor must match
		for _, ancestor := range ancestors {
			if e.inner.Matches(ancestor) {
				return true
			}
		}
		return false
	case QuantifierAll:
		// All ancestors must match (vacuously true if no ancestors)
		if len(ancestors) == 0 {
			return true
		}
		for _, ancestor := range ancestors {
			if !e.inner.Matches(ancestor) {
				return false
			}
		}
		return true
	case QuantifierNone:
		// No ancestors must match
		for _, ancestor := range ancestors {
			if e.inner.Matches(ancestor) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (e *AncestorFilter) String() string {
	if e.quantifier == QuantifierSome {
		return fmt.Sprintf("ancestor(%s)", e.inner.String())
	}
	return fmt.Sprintf("ancestor(%s,%s)", e.quantifier.String(), e.inner.String())
}

// ChildFilter matches items based on their immediate children (child in search syntax)
type ChildFilter struct {
	inner      FilterExpr
	quantifier Quantifier
}

func NewChildFilter(inner FilterExpr, quantifier Quantifier) *ChildFilter {
	return &ChildFilter{inner: inner, quantifier: quantifier}
}

func (e *ChildFilter) Matches(item *model.Item) bool {
	children := item.Children

	switch e.quantifier {
	case QuantifierSome:
		// At least one child must match
		for _, child := range children {
			if e.inner.Matches(child) {
				return true
			}
		}
		return false
	case QuantifierAll:
		// All children must match (false if no children)
		if len(children) == 0 {
			return false
		}
		for _, child := range children {
			if !e.inner.Matches(child) {
				return false
			}
		}
		return true
	case QuantifierNone:
		// No children must match
		for _, child := range children {
			if e.inner.Matches(child) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (e *ChildFilter) String() string {
	return fmt.Sprintf("child(%s,%s)", e.quantifier.String(), e.inner.String())
}

// DescendantFilter matches items based on all their descendants (child* in search syntax)
type DescendantFilter struct {
	inner      FilterExpr
	quantifier Quantifier
}

func NewDescendantFilter(inner FilterExpr, quantifier Quantifier) *DescendantFilter {
	return &DescendantFilter{inner: inner, quantifier: quantifier}
}

func (e *DescendantFilter) Matches(item *model.Item) bool {
	var descendants []*model.Item
	e.collectDescendants(item, &descendants)

	switch e.quantifier {
	case QuantifierSome:
		// At least one descendant must match
		for _, descendant := range descendants {
			if e.inner.Matches(descendant) {
				return true
			}
		}
		return false
	case QuantifierAll:
		// All descendants must match (false if no descendants)
		if len(descendants) == 0 {
			return false
		}
		for _, descendant := range descendants {
			if !e.inner.Matches(descendant) {
				return false
			}
		}
		return true
	case QuantifierNone:
		// No descendants must match
		for _, descendant := range descendants {
			if e.inner.Matches(descendant) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (e *DescendantFilter) collectDescendants(item *model.Item, descendants *[]*model.Item) {
	for _, child := range item.Children {
		*descendants = append(*descendants, child)
		e.collectDescendants(child, descendants)
	}
}

func (e *DescendantFilter) String() string {
	return fmt.Sprintf("descendant(%s,%s)", e.quantifier.String(), e.inner.String())
}

// SiblingFilter matches items based on their siblings (items with same parent)
type SiblingFilter struct {
	inner      FilterExpr
	quantifier Quantifier
}

func NewSiblingFilter(inner FilterExpr, quantifier Quantifier) *SiblingFilter {
	return &SiblingFilter{inner: inner, quantifier: quantifier}
}

func (e *SiblingFilter) Matches(item *model.Item) bool {
	// Root items have no siblings
	if item.Parent == nil {
		switch e.quantifier {
		case QuantifierSome:
			return false // No siblings to match
		case QuantifierAll:
			return false // No siblings, can't match all
		case QuantifierNone:
			return true // No siblings means none match
		default:
			return false
		}
	}

	// Get all siblings (excluding the item itself)
	var siblings []*model.Item
	for _, sibling := range item.Parent.Children {
		if sibling != item {
			siblings = append(siblings, sibling)
		}
	}

	switch e.quantifier {
	case QuantifierSome:
		// At least one sibling must match
		for _, sibling := range siblings {
			if e.inner.Matches(sibling) {
				return true
			}
		}
		return false
	case QuantifierAll:
		// All siblings must match (false if no siblings)
		if len(siblings) == 0 {
			return false
		}
		for _, sibling := range siblings {
			if !e.inner.Matches(sibling) {
				return false
			}
		}
		return true
	case QuantifierNone:
		// No siblings must match
		for _, sibling := range siblings {
			if e.inner.Matches(sibling) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (e *SiblingFilter) String() string {
	return fmt.Sprintf("sibling(%s,%s)", e.quantifier.String(), e.inner.String())
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
	// Empty string is not a valid date
	if len(value) == 0 {
		return false
	}

	// Check if ends with a time unit suffix
	suffix := value[len(value)-1:]
	isTimeUnit := suffix == "h" || suffix == "d" || suffix == "w" || suffix == "m" || suffix == "y"

	// Relative dates with explicit sign: -1h, -7d, +30d, etc.
	if strings.HasPrefix(value, "-") || strings.HasPrefix(value, "+") {
		// Need at least 2 characters: "-" + suffix
		if len(value) < 2 {
			return false
		}
		return isTimeUnit
	}

	// Relative dates without prefix (shortcut for "ago"): 1h, 7d, 30d, etc.
	// These are interpreted as "N time units ago" (in the past)
	if isTimeUnit && len(value) >= 2 {
		// Check if the part before suffix is a valid number
		var amount int
		_, err := fmt.Sscanf(value[:len(value)-1], "%d", &amount)
		return err == nil
	}

	// Absolute dates: YYYY-MM-DD
	return len(value) == 10 && value[4] == '-' && value[7] == '-'
}

// parseDate parses a date value (relative or absolute) into a time.Time
func parseDate(value string) time.Time {
	now := time.Now()

	// Try to parse as relative date with explicit sign (format: -/+Nh, -/+Nd, -/+Nw, -/+Nm, -/+Ny)
	if len(value) > 2 && (strings.HasPrefix(value, "-") || strings.HasPrefix(value, "+")) {
		sgn := 1
		if strings.HasPrefix(value, "-") {
			sgn = -1
		}

		var amount int
		var suffix string
		_, err := fmt.Sscanf(value[1:], "%d%s", &amount, &suffix)
		if err == nil && len(suffix) == 1 {
			switch suffix {
			case "h":
				return now.Add(time.Duration(sgn*amount) * time.Hour)
			case "d":
				return now.AddDate(0, 0, sgn*amount)
			case "w":
				return now.AddDate(0, 0, sgn*amount*7)
			case "m":
				return now.AddDate(0, sgn*amount, 0)
			case "y":
				return now.AddDate(sgn*amount, 0, 0)
			}
		}
	}

	// Try to parse as relative date without prefix (format: Nh, Nd, Nw, Nm, Ny)
	// These are interpreted as "N time units ago" (in the past)
	if len(value) >= 2 {
		suffix := value[len(value)-1:]
		if suffix == "h" || suffix == "d" || suffix == "w" || suffix == "m" || suffix == "y" {
			var amount int
			_, err := fmt.Sscanf(value[:len(value)-1], "%d", &amount)
			if err == nil {
				switch suffix {
				case "h":
					return now.Add(time.Duration(-amount) * time.Hour)
				case "d":
					return now.AddDate(0, 0, -amount)
				case "w":
					return now.AddDate(0, 0, -amount*7)
				case "m":
					return now.AddDate(0, -amount, 0)
				case "y":
					return now.AddDate(-amount, 0, 0)
				}
			}
		}
	}

	// Try to parse as absolute date (format: YYYY-MM-DD)
	t, err := time.Parse("2006-01-02", value)
	if err == nil {
		return t
	}

	return time.Time{}
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
