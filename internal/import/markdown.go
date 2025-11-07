package import_parser

import (
	"bufio"
	"strings"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// MarkdownParser imports markdown files
type MarkdownParser struct{}

func (p *MarkdownParser) Name() string {
	return "Markdown"
}

// Parse converts markdown content to outline items
func (p *MarkdownParser) Parse(content string) ([]*model.Item, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var rootItems []*model.Item
	var stack []*model.Item // Stack to track current parent at each level

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check for header
		if strings.HasPrefix(line, "#") {
			level, text := parseHeader(line)
			if level >= 0 {
				item := model.NewItem(text)

				// Add to appropriate parent based on level
				if level == 0 {
					rootItems = append(rootItems, item)
					stack = []*model.Item{item}
				} else if level <= len(stack) {
					// Pop stack to appropriate level
					stack = stack[:level]
					parent := stack[len(stack)-1]
					parent.Children = append(parent.Children, item)
					item.Parent = parent
					stack = append(stack, item)
				} else {
					// Nested deeper than previous - attach to last item
					if len(stack) > 0 {
						parent := stack[len(stack)-1]
						parent.Children = append(parent.Children, item)
						item.Parent = parent
						stack = append(stack, item)
					}
				}
				continue
			}
		}

		// Check for unordered list item
		if listLevel, text := parseListItem(line); listLevel >= 0 {
			item := model.NewItem(text)

			// Determine parent based on indentation
			if listLevel == 0 && len(stack) == 0 {
				// Root level list
				rootItems = append(rootItems, item)
				stack = []*model.Item{item}
			} else if listLevel < len(stack) {
				// Outdented - pop stack
				stack = stack[:listLevel+1]
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, item)
				item.Parent = parent
				stack = append(stack, item)
			} else if listLevel == len(stack)-1 {
				// Same level as previous - add as sibling
				if len(stack) > 1 {
					parent := stack[len(stack)-2]
					parent.Children = append(parent.Children, item)
					item.Parent = parent
					stack[len(stack)-1] = item
				} else {
					rootItems = append(rootItems, item)
					stack = []*model.Item{item}
				}
			} else {
				// Indented - add as child of previous
				if len(stack) > 0 {
					parent := stack[len(stack)-1]
					parent.Children = append(parent.Children, item)
					item.Parent = parent
					stack = append(stack, item)
				} else {
					rootItems = append(rootItems, item)
					stack = []*model.Item{item}
				}
			}
			continue
		}

		// Plain text - add as item at appropriate level
		text := strings.TrimSpace(line)
		if text != "" {
			item := model.NewItem(text)
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, item)
				item.Parent = parent
			} else {
				rootItems = append(rootItems, item)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rootItems, nil
}

// parseHeader extracts level and text from markdown header
func parseHeader(line string) (level int, text string) {
	level = 0
	for i := 0; i < len(line) && line[i] == '#'; i++ {
		level++
	}

	if level == 0 || level > len(line) {
		return -1, ""
	}

	text = strings.TrimSpace(line[level:])
	return level - 1, text // Convert to 0-based level
}

// parseListItem extracts indentation level and text from list item
func parseListItem(line string) (level int, text string) {
	// Count leading spaces/tabs
	indent := 0
	for i := 0; i < len(line); i++ {
		if line[i] == ' ' {
			indent++
		} else if line[i] == '\t' {
			indent += 2 // Treat tab as 2 spaces
		} else {
			break
		}
	}

	trimmed := strings.TrimSpace(line)

	// Check for list markers
	if len(trimmed) > 2 && (trimmed[0] == '-' || trimmed[0] == '*' || trimmed[0] == '+') && trimmed[1] == ' ' {
		text = strings.TrimSpace(trimmed[2:])
		level = indent / 2 // 2 spaces per level
		return level, text
	}

	return -1, ""
}
