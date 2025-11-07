package import_parser

import (
	"fmt"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// ImportFormat represents different file formats that can be imported
type ImportFormat string

const (
	FormatMarkdown     ImportFormat = "markdown"
	FormatIndentedText ImportFormat = "indented"
	FormatAuto         ImportFormat = "auto" // Auto-detect from extension
)

// ImportOptions configures how files are imported
type ImportOptions struct {
	Format     ImportFormat
	InsertMode string // "append", "prepend", "replace"
}

// Parser interface for different import formats
type Parser interface {
	Parse(content string) ([]*model.Item, error)
	Name() string
}

// ImportFile imports a file and returns the root items
func ImportFile(content string, format ImportFormat) ([]*model.Item, error) {
	var parser Parser

	switch format {
	case FormatMarkdown:
		parser = &MarkdownParser{}
	case FormatIndentedText:
		parser = &IndentedTextParser{}
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}

	items, err := parser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("parse error (%s): %w", parser.Name(), err)
	}

	return items, nil
}

// DetectFormat attempts to detect the file format from extension
func DetectFormat(filename string) ImportFormat {
	// Simple extension-based detection
	if len(filename) > 3 {
		ext := filename[len(filename)-3:]
		if ext == ".md" {
			return FormatMarkdown
		}
		if ext == "txt" {
			return FormatIndentedText
		}
	}

	// Default to indented text
	return FormatIndentedText
}
