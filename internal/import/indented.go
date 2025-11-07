package import_parser

import (
	"bufio"
	"strings"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// IndentedTextParser imports plain text files with indentation-based hierarchy
type IndentedTextParser struct{}

func (p *IndentedTextParser) Name() string {
	return "Indented Text"
}

// Parse converts indented text to outline items
func (p *IndentedTextParser) Parse(content string) ([]*model.Item, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var rootItems []*model.Item
	var stack []*model.Item // Stack to track parent at each indentation level
	prevIndent := -1

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Calculate indentation level
		indent := getIndentLevel(line)
		text := strings.TrimSpace(line)

		// Skip empty content after trimming
		if text == "" {
			continue
		}

		// Create new item
		item := model.NewItem(text)

		// Determine parent based on indentation
		if indent == 0 {
			// Root level item
			rootItems = append(rootItems, item)
			stack = []*model.Item{item}
		} else if indent > prevIndent {
			// Indented - child of previous item
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, item)
				item.Parent = parent
				stack = append(stack, item)
			} else {
				// No parent available, treat as root
				rootItems = append(rootItems, item)
				stack = []*model.Item{item}
			}
		} else if indent == prevIndent {
			// Same level - sibling of previous
			if len(stack) > 1 {
				// Remove previous item from stack, add to parent
				stack = stack[:len(stack)-1]
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, item)
				item.Parent = parent
				stack = append(stack, item)
			} else if len(stack) == 1 {
				// Root level sibling
				rootItems = append(rootItems, item)
				stack = []*model.Item{item}
			}
		} else {
			// Outdented - go back to appropriate level
			// Pop stack until we're at the right level
			targetLevel := indent
			if targetLevel > len(stack) {
				targetLevel = len(stack)
			}
			stack = stack[:targetLevel]

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, item)
				item.Parent = parent
			} else {
				rootItems = append(rootItems, item)
			}
			stack = append(stack, item)
		}

		prevIndent = indent
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rootItems, nil
}

// getIndentLevel calculates the indentation level (0-based)
// Counts tabs and spaces (tab = 2 spaces)
func getIndentLevel(line string) int {
	indent := 0
	for i := 0; i < len(line); i++ {
		if line[i] == '\t' {
			indent += 2
		} else if line[i] == ' ' {
			indent++
		} else {
			break
		}
	}
	// Convert to level (2 spaces = 1 level)
	return indent / 2
}
