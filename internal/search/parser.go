package search

import (
	"fmt"
	"strings"
)

// TokenType represents the type of a token in the search query
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenText
	TokenFilter
	TokenAnd     // + (explicit)
	TokenOr      // |
	TokenNot     // -
	TokenLParen  // (
	TokenRParen  // )
)

// Token represents a single token in the search query
type Token struct {
	Type  TokenType
	Value string
}

// FilterType represents the type of filter
type FilterType string

const (
	FilterTypeText     FilterType = "text"
	FilterTypeDepth    FilterType = "d"
	FilterTypeAttr     FilterType = "a"
	FilterTypeCreated  FilterType = "c"
	FilterTypeModified FilterType = "m"
	FilterTypeChildren FilterType = "children"
	FilterTypeParent   FilterType = "p"
	FilterTypeAncestor FilterType = "ancestor"
)

// ComparisonOp represents comparison operators
type ComparisonOp string

const (
	OpEqual        ComparisonOp = "="
	OpNotEqual     ComparisonOp = "!="
	OpGreater      ComparisonOp = ">"
	OpGreaterEqual ComparisonOp = ">="
	OpLess         ComparisonOp = "<"
	OpLessEqual    ComparisonOp = "<="
)

// Tokenizer converts a search query string into tokens
type Tokenizer struct {
	input string
	pos   int
}

// NewTokenizer creates a new tokenizer for the given input
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{input: input, pos: 0}
}

// NextToken returns the next token in the input
func (t *Tokenizer) NextToken() Token {
	t.skipWhitespace()

	if t.pos >= len(t.input) {
		return Token{Type: TokenEOF}
	}

	ch := t.input[t.pos]

	switch ch {
	case '(':
		t.pos++
		return Token{Type: TokenLParen, Value: "("}
	case ')':
		t.pos++
		return Token{Type: TokenRParen, Value: ")"}
	case '|':
		t.pos++
		return Token{Type: TokenOr, Value: "|"}
	case '+':
		t.pos++
		return Token{Type: TokenAnd, Value: "+"}
	case '-':
		// Check if it's a NOT operator or part of a number/filter
		if t.pos+1 < len(t.input) {
			next := t.input[t.pos+1]
			// If next char is digit, letter, or ':' it could be a filter
			// Otherwise it's a NOT operator
			if isFilterStart(next) {
				return t.readFilter()
			}
		}
		t.pos++
		return Token{Type: TokenNot, Value: "-"}
	case '"':
		return t.readQuotedText()
	case '@':
		return t.readAttrFilter()
	default:
		if isFilterStart(ch) {
			return t.readFilter()
		}
		return t.readText()
	}
}

// AllTokens returns all tokens in the input
func (t *Tokenizer) AllTokens() []Token {
	var tokens []Token
	for {
		tok := t.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

func (t *Tokenizer) skipWhitespace() {
	for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\t' || t.input[t.pos] == '\n') {
		t.pos++
	}
}

func (t *Tokenizer) readQuotedText() Token {
	t.pos++ // Skip opening quote
	start := t.pos
	for t.pos < len(t.input) && t.input[t.pos] != '"' {
		t.pos++
	}
	value := t.input[start:t.pos]
	if t.pos < len(t.input) {
		t.pos++ // Skip closing quote
	}
	return Token{Type: TokenText, Value: value}
}

func (t *Tokenizer) readFilter() Token {
	// Check if this is a NOT filter
	isNot := false
	start := t.pos
	if t.input[t.pos] == '-' {
		isNot = true
		t.pos++
	}

	// Read the filter identifier (could be word like 'ancestor' or 'children')
	identStart := t.pos
	for t.pos < len(t.input) && (isAlphaNumeric(t.input[t.pos]) || t.input[t.pos] == '_') {
		t.pos++
	}
	ident := t.input[identStart:t.pos]

	// Check if there's a colon (filter with criteria)
	if t.pos < len(t.input) && t.input[t.pos] == ':' {
		t.pos++ // Skip colon
		// Read the criteria part
		criteria := t.readFilterCriteria()
		value := ident + ":" + criteria
		if isNot {
			return Token{Type: TokenFilter, Value: "-" + value}
		}
		return Token{Type: TokenFilter, Value: value}
	}

	// Just a text token that looks like a filter identifier
	value := t.input[start:t.pos]
	if isNot && ident != "" {
		// This was "-" followed by text, so it's NOT operator
		t.pos = start + 1 // Back up to after the -
		return Token{Type: TokenNot, Value: "-"}
	}
	return Token{Type: TokenText, Value: value}
}

func (t *Tokenizer) readFilterCriteria() string {
	start := t.pos
	// Read comparison operator
	if t.pos < len(t.input) && (t.input[t.pos] == '>' || t.input[t.pos] == '<' || t.input[t.pos] == '!' || t.input[t.pos] == '=') {
		t.pos++
		if t.pos < len(t.input) && t.input[t.pos] == '=' {
			t.pos++
		}
	}

	// Read value (letters, digits, dots, dashes, underscores, etc.)
	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if ch == ' ' || ch == '\t' || ch == '|' || ch == '+' || ch == ')' {
			break
		}
		t.pos++
	}

	return t.input[start:t.pos]
}

func (t *Tokenizer) readText() Token {
	start := t.pos
	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if ch == ' ' || ch == '\t' || ch == '|' || ch == '+' || ch == ')' || ch == '(' {
			break
		}
		t.pos++
	}
	return Token{Type: TokenText, Value: t.input[start:t.pos]}
}

func (t *Tokenizer) readAttrFilter() Token {
	t.pos++ // Skip @

	// Read attribute key
	keyStart := t.pos
	for t.pos < len(t.input) && (isAlphaNumeric(t.input[t.pos]) || t.input[t.pos] == '_') {
		t.pos++
	}
	key := t.input[keyStart:t.pos]

	// Check if there's a comparison operator (no colon required for attributes)
	if t.pos < len(t.input) && (t.input[t.pos] == '>' || t.input[t.pos] == '<' || t.input[t.pos] == '!' || t.input[t.pos] == '=') {
		criteria := t.readFilterCriteria()
		value := "@" + key + criteria
		return Token{Type: TokenFilter, Value: value}
	}

	// Just @key with no criteria
	value := "@" + key
	return Token{Type: TokenFilter, Value: value}
}

func isFilterStart(ch byte) bool {
	return isAlpha(ch)
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || (ch >= '0' && ch <= '9') || ch == '_'
}

// Parser converts tokens into a FilterExpr tree
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser creates a new parser for the given tokens
func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// ParseQuery parses a complete search query and returns the root expression
func ParseQuery(query string) (FilterExpr, error) {
	tokenizer := NewTokenizer(query)
	tokens := tokenizer.AllTokens()

	if len(tokens) == 1 && tokens[0].Type == TokenEOF {
		// Empty query
		return NewAlwaysMatchExpr(), nil
	}

	parser := NewParser(tokens)
	expr, err := parser.parseOr()
	if err != nil {
		return nil, err
	}

	if parser.currentToken().Type != TokenEOF {
		return nil, fmt.Errorf("unexpected token: %s", parser.currentToken().Value)
	}

	return expr, nil
}

func (p *Parser) currentToken() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) peek() Token {
	if p.pos+1 >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos+1]
}

// Operator precedence: OR < AND < NOT < Atoms
// We parse from lowest to highest precedence (OR first, then AND, then NOT, then atoms)

func (p *Parser) parseOr() (FilterExpr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.currentToken().Type == TokenOr {
		p.advance() // consume |
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = NewOrExpr(left, right)
	}

	return left, nil
}

func (p *Parser) parseAnd() (FilterExpr, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.currentToken().Type == TokenAnd || (p.currentToken().Type != TokenEOF && p.currentToken().Type != TokenRParen && p.currentToken().Type != TokenOr) {
		if p.currentToken().Type == TokenAnd {
			p.advance() // consume +
		}
		// Implicit AND: just continue parsing if we see another filter/text
		if p.currentToken().Type == TokenEOF || p.currentToken().Type == TokenRParen || p.currentToken().Type == TokenOr {
			break
		}
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = NewAndExpr(left, right)
	}

	return left, nil
}

func (p *Parser) parseNot() (FilterExpr, error) {
	if p.currentToken().Type == TokenNot {
		p.advance() // consume -
		expr, err := p.parseNot() // Allow chaining of NOTs
		if err != nil {
			return nil, err
		}
		return NewNotExpr(expr), nil
	}

	return p.parseAtom()
}

func (p *Parser) parseAtom() (FilterExpr, error) {
	switch p.currentToken().Type {
	case TokenLParen:
		p.advance() // consume (
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.currentToken().Type != TokenRParen {
			return nil, fmt.Errorf("expected ')', got %s", p.currentToken().Value)
		}
		p.advance() // consume )
		return expr, nil

	case TokenText:
		value := p.currentToken().Value
		p.advance()
		return NewTextExpr(value), nil

	case TokenFilter:
		value := p.currentToken().Value
		p.advance()
		return parseFilterValue(value)

	case TokenEOF:
		return nil, fmt.Errorf("unexpected end of input")

	default:
		return nil, fmt.Errorf("unexpected token: %s", p.currentToken().Value)
	}
}

// parseFilterValue converts a filter token value into the appropriate FilterExpr
func parseFilterValue(value string) (FilterExpr, error) {
	// Check for NOT modifier
	isNot := false
	if strings.HasPrefix(value, "-") {
		isNot = true
		value = value[1:]
	}

	var expr FilterExpr
	var err error

	// Check for @ prefix (attribute filter - no colon separator)
	if strings.HasPrefix(value, "@") {
		expr, err = parseAttrFilter(value[1:]) // Strip @ and parse the criteria
		if err != nil {
			return nil, err
		}
		if isNot {
			expr = NewNotExpr(expr)
		}
		return expr, nil
	}

	// Parse the filter type and criteria (with colon separator)
	parts := strings.SplitN(value, ":", 2)
	if len(parts) < 2 {
		// Just a text token that got classified as filter
		expr = NewTextExpr(value)
		if isNot {
			expr = NewNotExpr(expr)
		}
		return expr, nil
	}

	filterType := parts[0]
	criteria := parts[1]

	switch filterType {
	case string(FilterTypeDepth):
		expr, err = parseDepthFilter(criteria)
	case string(FilterTypeCreated):
		expr, err = parseDateFilter(FilterTypeCreated, criteria)
	case string(FilterTypeModified):
		expr, err = parseDateFilter(FilterTypeModified, criteria)
	case string(FilterTypeChildren):
		expr, err = parseChildrenFilter(criteria)
	case string(FilterTypeParent):
		expr, err = parseParentFilter(criteria)
	case "a": // Changed from "ancestor" to "a"
		expr, err = parseAncestorFilter(criteria)
	default:
		// Unknown filter type, treat as text
		expr = NewTextExpr(value)
	}

	if err != nil {
		return nil, err
	}

	if isNot {
		expr = NewNotExpr(expr)
	}

	return expr, nil
}

func parseDepthFilter(criteria string) (FilterExpr, error) {
	op, val, err := parseComparison(criteria)
	if err != nil {
		return nil, err
	}
	return NewDepthFilter(op, val)
}

func parseAttrFilter(criteria string) (FilterExpr, error) {
	// Check for comparison operators (from longest to shortest to avoid partial matches)
	ops := []string{"!=", ">=", "<=", ">", "<", "="}
	var key, op, value string

	for _, o := range ops {
		if idx := strings.Index(criteria, o); idx != -1 {
			key = criteria[:idx]
			op = o
			value = criteria[idx+len(o):]
			break
		}
	}

	// If no operator found, just check for existence
	if op == "" {
		return NewAttrFilter(criteria, "", ""), nil
	}

	// Check if value is a date - if so, use AttributeDateFilter
	if isValidDateValue(value) {
		compOp := ComparisonOp(op)
		return NewAttrDateFilter(key, compOp, value)
	}

	// Otherwise use regular string comparison
	return NewAttrFilter(key, op, value), nil
}

func parseDateFilter(filterType FilterType, criteria string) (FilterExpr, error) {
	op, val, err := parseComparison(criteria)
	if err != nil {
		return nil, err
	}
	return NewDateFilter(filterType, op, val)
}

func parseChildrenFilter(criteria string) (FilterExpr, error) {
	op, val, err := parseComparison(criteria)
	if err != nil {
		return nil, err
	}
	return NewChildrenFilter(op, val)
}

func parseParentFilter(criteria string) (FilterExpr, error) {
	// Parent filter contains another filter expression
	// Recursively parse it
	innerExpr, err := parseFilterValue(criteria)
	if err != nil {
		return nil, err
	}
	return NewParentFilter(innerExpr), nil
}

func parseAncestorFilter(criteria string) (FilterExpr, error) {
	// Ancestor filter contains another filter expression
	// Recursively parse it
	innerExpr, err := parseFilterValue(criteria)
	if err != nil {
		return nil, err
	}
	return NewAncestorFilter(innerExpr), nil
}

// parseComparison extracts the comparison operator and value from criteria
// Examples: "5" -> ("=", "5"), ">2" -> (">", "2"), ">=2025-11-01" -> (">=", "2025-11-01")
func parseComparison(criteria string) (ComparisonOp, string, error) {
	if criteria == "" {
		return "", "", fmt.Errorf("empty criteria")
	}

	// Check for comparison operators
	ops := []ComparisonOp{OpGreaterEqual, OpLessEqual, OpNotEqual, OpGreater, OpLess, OpEqual}
	for _, op := range ops {
		if strings.HasPrefix(criteria, string(op)) {
			val := criteria[len(op):]
			if val == "" {
				return "", "", fmt.Errorf("missing value after operator %s", op)
			}
			return op, val, nil
		}
	}

	// Default to equality
	return OpEqual, criteria, nil
}
