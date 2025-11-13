package template

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TemplateContext holds data needed for template evaluation
type TemplateContext struct {
	ClipboardCommand string // e.g., "wl-copy"
	AppContext       interface{} // *app.App for interactive operations
}

// InteractionCallback is called when a template expression requires user interaction
// The app should handle the interaction and call the response callback with the result
type InteractionCallback func(interaction *Interaction)

// Interaction represents an async operation that needs user interaction
type Interaction struct {
	Type     string             // "prompt", "select", "search"
	Question string             // for prompt
	Options  []string           // for select
	Query    string             // for search
	OnResult func(result string) // Callback with result
}

// ProcessTemplate evaluates template expressions in text
// Non-interactive expressions (now, date, clipboard) are processed immediately
// Interactive expressions (prompt, select, search) are skipped and returned in pendingInteractions
// Expressions are in the format {{function:arg1,arg2}} or {{function}}
func ProcessTemplate(text string, ctx TemplateContext) (string, []*Interaction, error) {
	result := text
	var pendingInteractions []*Interaction

	// Match expressions like {{...}}
	expr := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	matches := expr.FindAllStringSubmatch(result, -1)
	for _, match := range matches {
		fullExpr := match[0]      // {{...}}
		innerExpr := match[1]     // ...

		value, interaction, err := evaluateExpression(innerExpr, ctx)
		if err != nil {
			return "", nil, err
		}

		if interaction != nil {
			// Store for later processing
			pendingInteractions = append(pendingInteractions, interaction)
		} else {
			// Replace immediately - value is already a string from evaluateExpression
			result = strings.ReplaceAll(result, fullExpr, value)
		}
	}

	return result, pendingInteractions, nil
}

// evaluateExpression evaluates a single expression with support for piping
// Returns: (value, interaction or nil, error)
// If interaction is not nil, the caller should handle user interaction
// Example: weekday(1)|date:%V %Y (%A, %B %d)
func evaluateExpression(expr string, ctx TemplateContext) (string, *Interaction, error) {
	// Split by pipe for chained operations
	parts := strings.Split(expr, "|")

	// Process first function
	rawValue, interaction, err := processFunction(parts[0], ctx)
	if err != nil {
		return "", nil, err
	}

	// If we have no pipes, convert raw value to string
	if len(parts) == 1 {
		return convertToString(rawValue), interaction, nil
	}

	// Apply any pipes - pass previous result as context
	value := rawValue
	for i := 1; i < len(parts); i++ {
		pipeExpr := strings.TrimSpace(parts[i])

		// Handle piping - the value from previous step becomes context for next
		pipeResult, err := applyPipe(pipeExpr, value, ctx)
		if err != nil {
			return "", nil, err
		}
		value = pipeResult
	}

	return convertToString(value), interaction, nil
}

// convertToString converts a value (string or DateValue) to string
func convertToString(val interface{}) string {
	if val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	if dv, ok := val.(*DateValue); ok {
		// Default formatting if DateValue wasn't piped
		formatted, _ := FormatDateValue(dv, "%Y-%m-%d")
		return formatted
	}
	return ""
}

// applyPipe applies a pipe operation to a value from the previous step
// value can be a string or a DateValue reference
func applyPipe(pipeExpr string, prevValue interface{}, ctx TemplateContext) (string, error) {
	// Parse the pipe expression: "function:args" or "function"
	parts := strings.SplitN(pipeExpr, ":", 2)
	function := strings.TrimSpace(parts[0])
	var args string
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	// Handle date formatting on a DateValue
	if function == "date" {
		// Check if prevValue is a DateValue (from weekday)
		if dv, ok := prevValue.(*DateValue); ok {
			result, err := FormatDateValue(dv, args)
			if err != nil {
				return "", err
			}
			return result, nil
		}
	}

	// If we reach here, we don't know how to pipe this value
	return "", fmt.Errorf("unknown pipe function: %s", function)
}

// processFunction handles individual function calls
// Format: "function" or "function:arg1,arg2,arg3" or "function(arg)"
// Returns: (value which can be string or DateValue, interaction or nil, error)
func processFunction(funcExpr string, ctx TemplateContext) (interface{}, *Interaction, error) {
	// Handle function calls with parentheses: function(arg) or function(arg1,arg2)
	funcExpr = strings.TrimSpace(funcExpr)
	var function string
	var args string

	if strings.Contains(funcExpr, "(") && strings.Contains(funcExpr, ")") {
		// Parse function(args) format
		closeIdx := strings.LastIndex(funcExpr, ")")
		if closeIdx > 0 {
			function = strings.TrimSpace(funcExpr[:strings.Index(funcExpr, "(")])
			args = strings.TrimSpace(funcExpr[strings.Index(funcExpr, "(")+1 : closeIdx])
		} else {
			// Malformed, treat as simple function
			function = funcExpr
		}
	} else {
		// Parse function:args format
		parts := strings.SplitN(funcExpr, ":", 2)
		function = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			args = strings.TrimSpace(parts[1])
		}
	}

	switch function {
	case "now":
		val, err := Now()
		return val, nil, err
	case "date":
		if args == "" {
			args = "%Y-%m-%d"
		}
		val, err := DateFormat(args)
		return val, nil, err
	case "clipboard":
		val, err := Clipboard(ctx.ClipboardCommand)
		return val, nil, err
	case "weekday":
		// Parse the numeric argument
		dayNum := 0
		if args != "" {
			if n, err := strconv.Atoi(args); err == nil {
				dayNum = n
			}
		}
		val, err := Weekday(dayNum)
		return val, nil, err
	case "select":
		interaction := &Interaction{
			Type:    "select",
			Options: parseOptions(args),
		}
		return "", interaction, nil
	case "search":
		interaction := &Interaction{
			Type:  "search",
			Query: args,
		}
		return "", interaction, nil
	case "prompt":
		interaction := &Interaction{
			Type:     "prompt",
			Question: unquoteString(args),
		}
		return "", interaction, nil
	default:
		return "", nil, nil // Unknown function, return empty
	}
}

// unquoteString removes surrounding quotes from a string
func unquoteString(s string) string {
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimSuffix(s, "'")
	return s
}
