package export

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// ExportToMarkdown exports an outline to a markdown file with full markdown format.
// Headers are exported as markdown headers (# ## ###), regular items as bullets.
func ExportToMarkdown(outline *model.Outline, filePath string) error {
	content := GenerateMarkdownWithHeaders(outline)

	// Write to file
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

// ExportToMarkdownList exports an outline to a markdown file with unordered list format.
// All items (including headers) are represented as bullets with indentation based on depth.
func ExportToMarkdownList(outline *model.Outline, filePath string) error {
	content := GenerateMarkdownList(outline)

	// Write to file
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

// ExportToMarkdownWriter exports an outline to markdown format and writes to the given writer.
// Uses the full markdown format with headers.
func ExportToMarkdownWriter(outline *model.Outline, w io.Writer) error {
	content := GenerateMarkdownWithHeaders(outline)
	_, err := w.Write([]byte(content))
	return err
}

// GenerateMarkdownWithHeaders generates markdown content from an outline with headers as markdown headers.
// Headers are exported as # ## ### etc., regular items as bullets.
func GenerateMarkdownWithHeaders(outline *model.Outline) string {
	var sb strings.Builder

	// TODO: When writing Hugo markdown, write the attributes from the root node
	// as frontmatter yaml

	// Write all items as markdown with header support
	for _, item := range outline.Items {
		writeItemAsMarkdownWithHeaders(&sb, item, 0, 1)
	}

	return sb.String()
}

// GenerateMarkdownList generates markdown content from an outline as an unordered list.
// All items (including headers) are exported as bullets.
func GenerateMarkdownList(outline *model.Outline) string {
	var sb strings.Builder

	// TODO: When writing Hugo markdown, write the attributes from the root node
	// as frontmatter yaml

	// Write all items as markdown bullets
	for _, item := range outline.Items {
		writeItemAsMarkdownList(&sb, item, 0)
	}

	return sb.String()
}

// GenerateMarkdown is deprecated, use GenerateMarkdownWithHeaders or GenerateMarkdownList instead.
// Kept for backwards compatibility.
func GenerateMarkdown(outline *model.Outline) string {
	return GenerateMarkdownList(outline)
}

// writeItemAsMarkdownWithHeaders recursively writes an item and its children as markdown.
// Headers are written as markdown headers (# ## ###), regular items as bullets.
// depth determines the bullet indentation level (2 spaces per level).
// headerLevel determines the header level (1 = #, 2 = ##, etc.).
func writeItemAsMarkdownWithHeaders(sb *strings.Builder, item *model.Item, depth int, headerLevel int) {
	if item == nil {
		return
	}

	// Skip empty items
	if strings.TrimSpace(item.Text) == "" {
		// Still process children even if this item is empty
		for _, child := range item.Children {
			writeItemAsMarkdownWithHeaders(sb, child, depth, headerLevel)
		}
		return
	}

	if item.IsHeader() {
		// Write as markdown header
		sb.WriteString(strings.Repeat("#", headerLevel))
		sb.WriteString(" ")
		sb.WriteString(item.Text)
		sb.WriteString("\n\n")

		// Write children with increased header level
		for _, child := range item.Children {
			writeItemAsMarkdownWithHeaders(sb, child, 0, headerLevel+1)
		}
	} else {
		// Write as bullet with indentation (2 spaces per level)
		indent := strings.Repeat("  ", depth)
		sb.WriteString(indent)
		sb.WriteString("- ")
		sb.WriteString(item.Text)
		sb.WriteString("\n")

		// Write children with increased depth
		for _, child := range item.Children {
			writeItemAsMarkdownWithHeaders(sb, child, depth+1, headerLevel)
		}
	}
}

// writeItemAsMarkdownList recursively writes an item and its children as markdown bullets.
// depth determines the indentation level (2 spaces per level).
func writeItemAsMarkdownList(sb *strings.Builder, item *model.Item, depth int) {
	if item == nil {
		return
	}

	// Skip empty items
	if strings.TrimSpace(item.Text) == "" {
		// Still process children even if this item is empty
		for _, child := range item.Children {
			writeItemAsMarkdownList(sb, child, depth)
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
		writeItemAsMarkdownList(sb, child, depth+1)
	}
}

// writeItemAsMarkdown is deprecated, use writeItemAsMarkdownList or writeItemAsMarkdownWithHeaders instead.
// Kept for backwards compatibility.
func writeItemAsMarkdown(sb *strings.Builder, item *model.Item, depth int) {
	writeItemAsMarkdownList(sb, item, depth)
}
