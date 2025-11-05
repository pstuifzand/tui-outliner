// Package ui contains terminal UI components
package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/pstuifzand/tui-outliner/internal/config"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// ExternalEditorData represents the data structure for TOML frontmatter
type ExternalEditorData struct {
	Tags       []string            `toml:"tags"`
	Attributes map[string]string   `toml:"attributes"`
}

// ValidateAttributesFunc is a callback to validate attributes
// Returns error message if validation fails, empty string if valid
type ValidateAttributesFunc func(attributes map[string]string) string

// EditItemInExternalEditor opens the item in an external editor, updates it with changes
// If validateAttrs is provided, it will be called to validate the edited attributes
func EditItemInExternalEditor(item *model.Item, cfg *config.Config, validateAttrs ValidateAttributesFunc) error {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "tuo-edit-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Serialize item to temp file with TOML frontmatter
	if err := serializeItemToFile(tmpFile, item); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to serialize item: %w", err)
	}
	tmpFile.Close()

	// Get original content to detect changes
	originalContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read temp file: %w", err)
	}

	// Resolve editor command
	editorCmd := resolveEditor(cfg)

	// Launch editor using shell to properly handle commands with arguments like "vim --clean"
	// We use sh -c to allow complex editor commands with flags
	// The terminal is temporarily released for the editor to use
	cmd := exec.Command("sh", "-c", editorCmd+" "+tmpPath)

	// Inherit stdin/stdout/stderr from the current process for proper terminal interaction
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set process group to ensure proper signal handling
	// This allows the editor to function properly in the terminal
	if err := cmd.Run(); err != nil {
		// Editor exited with error, but we should still try to read the file
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("failed to launch editor: %w", err)
		}
	}

	// Read edited content
	editedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Check if file was unchanged or empty (editor was closed without saving)
	if bytes.Equal(originalContent, editedContent) || len(editedContent) == 0 {
		// No changes - keep original
		return nil
	}

	// Parse edited content
	text, data, err := deserializeItemFromFile(editedContent)
	if err != nil {
		// Parse error - keep original
		return fmt.Errorf("failed to parse edited content: %w (keeping original)", err)
	}

	// Validate attributes if validator provided
	if validateAttrs != nil && len(data.Attributes) > 0 {
		if errMsg := validateAttrs(data.Attributes); errMsg != "" {
			return fmt.Errorf("attribute validation failed: %s (keeping original)", errMsg)
		}
	}

	// Update item with parsed data
	item.Text = text
	if len(data.Tags) > 0 {
		item.Metadata.Tags = data.Tags
	}
	if len(data.Attributes) > 0 {
		item.Metadata.Attributes = data.Attributes
	}
	item.Metadata.Modified = time.Now()

	return nil
}

// serializeItemToFile writes the item to a temp file with TOML frontmatter
func serializeItemToFile(file *os.File, item *model.Item) error {
	// Create the frontmatter data
	data := ExternalEditorData{
		Tags:       item.Metadata.Tags,
		Attributes: item.Metadata.Attributes,
	}

	// Encode TOML frontmatter
	var buf bytes.Buffer
	buf.WriteString("+++\n")

	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return err
	}

	buf.WriteString("+++\n")

	// Write frontmatter to file
	if _, err := file.Write(buf.Bytes()); err != nil {
		return err
	}

	// Write item text
	if _, err := file.WriteString(item.Text); err != nil {
		return err
	}

	return file.Sync()
}

// deserializeItemFromFile parses the edited file content and returns text and metadata
func deserializeItemFromFile(content []byte) (string, *ExternalEditorData, error) {
	contentStr := string(content)

	// Find TOML frontmatter delimiters
	startDelim := strings.Index(contentStr, "+++\n")
	if startDelim == -1 {
		// No frontmatter, treat whole content as text
		return contentStr, &ExternalEditorData{
			Tags:       []string{},
			Attributes: make(map[string]string),
		}, nil
	}

	startPos := startDelim + 4 // Skip "+++\n"
	endDelim := strings.Index(contentStr[startPos:], "+++\n")
	if endDelim == -1 {
		// No closing delimiter, treat as text
		return contentStr, &ExternalEditorData{
			Tags:       []string{},
			Attributes: make(map[string]string),
		}, nil
	}

	// Extract TOML and text sections
	tomlSection := contentStr[startPos : startPos+endDelim]
	textSection := contentStr[startPos+endDelim+4:] // Skip closing "+++\n"

	// Parse TOML
	data := &ExternalEditorData{
		Tags:       []string{},
		Attributes: make(map[string]string),
	}
	if err := toml.Unmarshal([]byte(tomlSection), data); err != nil {
		return "", nil, err
	}

	// Trim leading/trailing whitespace from text
	text := strings.TrimSpace(textSection)

	return text, data, nil
}

// resolveEditor determines which editor to use
func resolveEditor(cfg *config.Config) string {
	// Check if editor is configured via :set editor
	if editorVal := cfg.Get("editor"); editorVal != "" {
		return editorVal
	}

	// Check EDITOR environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Default fallback
	return "vi"
}
