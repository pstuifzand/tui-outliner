package search

import (
	"fmt"
	"strings"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// DebugInfo contains detailed information about why an item matched or didn't match
type DebugInfo struct {
	Item    *model.Item
	Expr    FilterExpr
	Matched bool
	Reason  string
	Details map[string]string // Additional details about the evaluation
}

// ExpressionString returns a pretty-printed representation of the filter expression
func ExpressionString(expr FilterExpr) string {
	return prettyPrintExpr(expr, 0)
}

func prettyPrintExpr(expr FilterExpr, indent int) string {
	indentStr := strings.Repeat("  ", indent)

	switch e := expr.(type) {
	case *AndExpr:
		left := prettyPrintExpr(e.left, indent+1)
		right := prettyPrintExpr(e.right, indent+1)
		return fmt.Sprintf("%s(and\n%s\n%s\n%s)", indentStr, left, right, indentStr)

	case *OrExpr:
		left := prettyPrintExpr(e.left, indent+1)
		right := prettyPrintExpr(e.right, indent+1)
		return fmt.Sprintf("%s(or\n%s\n%s\n%s)", indentStr, left, right, indentStr)

	case *NotExpr:
		inner := prettyPrintExpr(e.expr, indent+1)
		return fmt.Sprintf("%s(not\n%s\n%s)", indentStr, inner, indentStr)

	default:
		return indentStr + expr.String()
	}
}

// DebugMatch returns detailed information about why an item matched or didn't match
func DebugMatch(item *model.Item, expr FilterExpr) *DebugInfo {
	matched := expr.Matches(item)
	reason := evaluateWithReason(item, expr)

	debug := &DebugInfo{
		Item:    item,
		Expr:    expr,
		Matched: matched,
		Reason:  reason,
		Details: make(map[string]string),
	}

	// Add detailed information for common filter types
	addItemDetails(item, debug)

	return debug
}

func evaluateWithReason(item *model.Item, expr FilterExpr) string {
	switch e := expr.(type) {
	case *TextExpr:
		if e.Matches(item) {
			return fmt.Sprintf("Text contains %q", e.term)
		}
		return fmt.Sprintf("Text does not contain %q", e.term)

	case *DepthFilter:
		depth := calculateDepth(item)
		if e.Matches(item) {
			return fmt.Sprintf("Depth %d matches condition %s%d", depth, e.op, e.value)
		}
		return fmt.Sprintf("Depth %d does not match condition %s%d", depth, e.op, e.value)

	case *AttributeFilter:
		if item.Metadata == nil || item.Metadata.Attributes == nil {
			return "No attributes defined"
		}
		if val, exists := item.Metadata.Attributes[e.key]; exists {
			if e.op == "" {
				return fmt.Sprintf("Has attribute %q = %q", e.key, val)
			}
			if e.Matches(item) {
				return fmt.Sprintf("Attribute %q %s %q (value: %q)", e.key, e.op, e.value, val)
			}
			return fmt.Sprintf("Attribute %q = %q does not match %s %q", e.key, val, e.op, e.value)
		}
		return fmt.Sprintf("No attribute %q", e.key)

	case *DateFilter:
		typeStr := "created"
		var compareTime string
		if e.filterType == FilterTypeModified {
			typeStr = "modified"
			if item.Metadata != nil {
				compareTime = item.Metadata.Modified.Format("2006-01-02")
			}
		} else {
			if item.Metadata != nil {
				compareTime = item.Metadata.Created.Format("2006-01-02")
			}
		}
		if e.Matches(item) {
			return fmt.Sprintf("%s date %s matches %s%s", typeStr, compareTime, e.op, e.value)
		}
		return fmt.Sprintf("%s date %s does not match %s%s", typeStr, compareTime, e.op, e.value)

	case *ChildrenFilter:
		count := len(item.Children)
		if e.Matches(item) {
			return fmt.Sprintf("Has %d children, matches %s%d", count, e.op, e.value)
		}
		return fmt.Sprintf("Has %d children, does not match %s%d", count, e.op, e.value)

	case *ParentFilter:
		if item.Parent == nil {
			return "No parent"
		}
		return fmt.Sprintf("Parent matches: %s", evaluateWithReason(item.Parent, e.inner))

	case *AncestorFilter:
		current := item.Parent
		depth := 0
		for current != nil {
			depth++
			if e.inner.Matches(current) {
				return fmt.Sprintf("Ancestor at level %d matches: %s", depth, evaluateWithReason(current, e.inner))
			}
			current = current.Parent
		}
		return "No ancestor matches"

	case *AndExpr:
		leftMatch := e.left.Matches(item)
		rightMatch := e.right.Matches(item)
		if leftMatch && rightMatch {
			return fmt.Sprintf("Both conditions match: %s AND %s", evaluateWithReason(item, e.left), evaluateWithReason(item, e.right))
		}
		if !leftMatch {
			return fmt.Sprintf("Left condition fails: %s", evaluateWithReason(item, e.left))
		}
		return fmt.Sprintf("Right condition fails: %s", evaluateWithReason(item, e.right))

	case *OrExpr:
		leftMatch := e.left.Matches(item)
		if leftMatch {
			return fmt.Sprintf("Left condition matches: %s", evaluateWithReason(item, e.left))
		}
		return fmt.Sprintf("Right condition matches: %s", evaluateWithReason(item, e.right))

	case *NotExpr:
		innerMatch := e.expr.Matches(item)
		if !innerMatch {
			return fmt.Sprintf("Condition is false (inverted): %s", evaluateWithReason(item, e.expr))
		}
		return fmt.Sprintf("Condition is true (inverted): %s", evaluateWithReason(item, e.expr))

	default:
		return expr.String()
	}
}

func addItemDetails(item *model.Item, debug *DebugInfo) {
	debug.Details["text"] = item.Text
	debug.Details["depth"] = fmt.Sprintf("%d", calculateDepth(item))
	debug.Details["children_count"] = fmt.Sprintf("%d", len(item.Children))

	if item.Parent != nil {
		debug.Details["parent"] = item.Parent.Text
	} else {
		debug.Details["parent"] = "(root)"
	}

	if item.Metadata != nil {
		debug.Details["created"] = item.Metadata.Created.Format("2006-01-02")
		debug.Details["modified"] = item.Metadata.Modified.Format("2006-01-02")

		if len(item.Metadata.Tags) > 0 {
			debug.Details["tags"] = strings.Join(item.Metadata.Tags, ", ")
		}

		if len(item.Metadata.Attributes) > 0 {
			attrs := make([]string, 0, len(item.Metadata.Attributes))
			for k, v := range item.Metadata.Attributes {
				attrs = append(attrs, fmt.Sprintf("%s=%s", k, v))
			}
			debug.Details["attributes"] = strings.Join(attrs, ", ")
		}
	}
}

// FormatDebugInfo returns a formatted string representation of debug information
func FormatDebugInfo(debug *DebugInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Item: %s\n", debug.Item.Text))
	sb.WriteString(fmt.Sprintf("Matched: %v\n", debug.Matched))
	sb.WriteString(fmt.Sprintf("Reason: %s\n", debug.Reason))

	if len(debug.Details) > 0 {
		sb.WriteString("\nDetails:\n")
		for key, value := range debug.Details {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	return sb.String()
}
