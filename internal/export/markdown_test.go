package export

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

func TestExportToMarkdown(t *testing.T) {
	// Create a test outline
	outline := &model.Outline{
		Title: "Test Outline",
		Items: []*model.Item{
			{
				ID:   "1",
				Text: "First Item",
				Children: []*model.Item{
					{
						ID:       "1.1",
						Text:     "Nested Item 1",
						Children: []*model.Item{},
					},
					{
						ID:   "1.2",
						Text: "Nested Item 2",
						Children: []*model.Item{
							{
								ID:       "1.2.1",
								Text:     "Deep Item",
								Children: []*model.Item{},
							},
						},
					},
				},
			},
			{
				ID:       "2",
				Text:     "Second Item",
				Children: []*model.Item{},
			},
		},
	}

	// Create a temporary file for output
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_output.md")

	// Export to markdown
	err := ExportToMarkdown(outline, outputFile)
	if err != nil {
		t.Fatalf("ExportToMarkdown failed: %v", err)
	}

	// Read the output file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedContent := `# Test Outline

- First Item
  - Nested Item 1
  - Nested Item 2
    - Deep Item
- Second Item
`

	if string(content) != expectedContent {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedContent, string(content))
	}
}

func TestExportToMarkdownWithEmptyItems(t *testing.T) {
	// Create a test outline with empty items
	outline := &model.Outline{
		Title: "Test Empty Items",
		Items: []*model.Item{
			{
				ID:   "1",
				Text: "Item with content",
				Children: []*model.Item{
					{
						ID:       "1.1",
						Text:     "", // Empty item
						Children: []*model.Item{},
					},
					{
						ID:       "1.2",
						Text:     "Another item",
						Children: []*model.Item{},
					},
				},
			},
		},
	}

	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_empty.md")

	err := ExportToMarkdown(outline, outputFile)
	if err != nil {
		t.Fatalf("ExportToMarkdown failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedContent := `# Test Empty Items

- Item with content
  - Another item
`

	if string(content) != expectedContent {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedContent, string(content))
	}
}
