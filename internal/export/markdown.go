package export

import (
	"fmt"
	"os"
	"strings"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// ExportToMarkdown exports an outline to a markdown file with unordered list format.
// Items are represented as bullets with indentation based on depth.
func ExportToMarkdown(outline *model.Outline, filePath string) error {
	var sb strings.Builder

	// TODO: When writing Hugo markdown, write the attributes from the root node
	// as frontmatter yaml

	// Write all items as markdown bullets
	for _, item := range outline.Items {
		writeItemAsMarkdown(&sb, item, 0)
	}

	// Write to file
	content := sb.String()
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

// writeItemAsMarkdown recursively writes an item and its children as markdown bullets.
// depth determines the indentation level (2 spaces per level).
func writeItemAsMarkdown(sb *strings.Builder, item *model.Item, depth int) {
	if item == nil {
		return
	}

	// Skip empty items
	if strings.TrimSpace(item.Text) == "" {
		// Still process children even if this item is empty
		for _, child := range item.Children {
			writeItemAsMarkdown(sb, child, depth)
		}
		return
	}

	// Write indentation (2 spaces per level)
	indent := strings.Repeat("  ", depth)
	sb.WriteString(indent)
	sb.WriteString("- ")
	sb.WriteString(item.Text)
	sb.WriteString("\n")

	// Write children with increased depth
	for _, child := range item.Children {
		writeItemAsMarkdown(sb, child, depth+1)
	}
}
