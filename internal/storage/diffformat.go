package storage

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

// DiffFormat provides encoding/decoding of outlines in a diff-optimized format
// Format contains multiple sections, all sorted by ID for deterministic output:
//
//   [TEXT SECTION]
//   id: escaped text content
//   id: escaped text content
//
//   [STRUCTURE SECTION]
//   id: parent_id:position
//   id: parent_id:position
//
//   [TAGS SECTION]
//   id: tag1,tag2,tag3
//   id: tag1
//
//   [ATTRIBUTES SECTION]
//   id: key1=value1,key2=value2
//   id: key1=value1
//
//   [TIMESTAMPS SECTION]
//   id: created_timestamp modified_timestamp
//   id: created_timestamp modified_timestamp
//
// Text escaping:
//   - \ (backslash) is encoded as \\
//   - newline is encoded as \n
//   - Must be decoded in reverse order to handle escapes correctly

// EncodeDiffFormat encodes an outline to the diff-optimized format
func EncodeDiffFormat(outline *model.Outline, w io.Writer) error {
	writer := bufio.NewWriter(w)

	// Build lookup maps for efficient access
	itemsByID := make(map[string]*model.Item)
	var allItems []*model.Item
	collectAllItems(outline.Items, itemsByID, &allItems)

	// Sort items by ID for consistent output
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].ID < allItems[j].ID
	})

	// Build structure data: map of id -> (parentID, position)
	structureData := make(map[string][2]string) // id -> [parentID, position]

	// Walk entire tree to build structure
	var walkStructure func([]*model.Item, string)
	walkStructure = func(items []*model.Item, parentID string) {
		for i, item := range items {
			position := fmt.Sprintf("%d", i)
			structureData[item.ID] = [2]string{parentID, position}
			if len(item.Children) > 0 {
				walkStructure(item.Children, item.ID)
			}
		}
	}
	walkStructure(outline.Items, "")

	// Write TEXT SECTION
	if _, err := writer.WriteString("[TEXT SECTION]\n"); err != nil {
		return err
	}

	for _, item := range allItems {
		encodedText := encodeTextValue(item.Text)
		line := fmt.Sprintf("%s: %s\n", item.ID, encodedText)
		if _, err := writer.WriteString(line); err != nil {
			return err
		}
	}

	// Write STRUCTURE SECTION
	if _, err := writer.WriteString("\n[STRUCTURE SECTION]\n"); err != nil {
		return err
	}

	for _, item := range allItems {
		structure := structureData[item.ID]
		parentID := structure[0]
		position := structure[1]
		line := fmt.Sprintf("%s: %s:%s\n", item.ID, parentID, position)
		if _, err := writer.WriteString(line); err != nil {
			return err
		}
	}

	// Write TAGS SECTION
	if _, err := writer.WriteString("\n[TAGS SECTION]\n"); err != nil {
		return err
	}

	for _, item := range allItems {
		if item.Metadata != nil && len(item.Metadata.Tags) > 0 {
			tagStr := strings.Join(item.Metadata.Tags, ",")
			line := fmt.Sprintf("%s: %s\n", item.ID, tagStr)
			if _, err := writer.WriteString(line); err != nil {
				return err
			}
		}
	}

	// Write ATTRIBUTES SECTION
	if _, err := writer.WriteString("\n[ATTRIBUTES SECTION]\n"); err != nil {
		return err
	}

	for _, item := range allItems {
		if item.Metadata != nil && len(item.Metadata.Attributes) > 0 {
			// Sort attributes for consistent output
			var keys []string
			for k := range item.Metadata.Attributes {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var pairs []string
			for _, k := range keys {
				pairs = append(pairs, fmt.Sprintf("%s=%s", k, item.Metadata.Attributes[k]))
			}
			attrStr := strings.Join(pairs, ",")
			line := fmt.Sprintf("%s: %s\n", item.ID, attrStr)
			if _, err := writer.WriteString(line); err != nil {
				return err
			}
		}
	}

	// Write TIMESTAMPS SECTION
	if _, err := writer.WriteString("\n[TIMESTAMPS SECTION]\n"); err != nil {
		return err
	}

	for _, item := range allItems {
		if item.Metadata != nil {
			created := item.Metadata.Created.Format(time.RFC3339Nano)
			modified := item.Metadata.Modified.Format(time.RFC3339Nano)
			line := fmt.Sprintf("%s: %s %s\n", item.ID, created, modified)
			if _, err := writer.WriteString(line); err != nil {
				return err
			}
		}
	}

	return writer.Flush()
}

// DecodeDiffFormat decodes an outline from the diff-optimized format
func DecodeDiffFormat(r io.Reader) (*model.Outline, error) {
	scanner := bufio.NewScanner(r)
	outline := &model.Outline{
		Items: make([]*model.Item, 0),
	}

	textData := make(map[string]string)
	structureData := make(map[string][2]string) // id -> [parentID, position]
	tagsData := make(map[string][]string)
	attributesData := make(map[string]map[string]string)
	timestampsData := make(map[string][2]time.Time)

	var bufferedLine *string // Store a line for the next iteration

	// Helper to read a section, storing the next section header in bufferedLine
	readSection := func(sectionName string) error {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "[") {
				// Start of next section - buffer it
				bufferedLine = &line
				return nil
			}
			if line == "" {
				continue
			}

			switch sectionName {
			case "TEXT":
				id, text, err := parseTextLine(line)
				if err != nil {
					return err
				}
				textData[id] = text

			case "STRUCTURE":
				id, parentID, position, err := parseStructureLine(line)
				if err != nil {
					return err
				}
				structureData[id] = [2]string{parentID, position}

			case "TAGS":
				id, tags, err := parseTagsLine(line)
				if err != nil {
					return err
				}
				tagsData[id] = tags

			case "ATTRIBUTES":
				id, attrs, err := parseAttributesLine(line)
				if err != nil {
					return err
				}
				attributesData[id] = attrs

			case "TIMESTAMPS":
				id, created, modified, err := parseTimestampsLine(line)
				if err != nil {
					return err
				}
				timestampsData[id] = [2]time.Time{created, modified}
			}
		}
		return scanner.Err()
	}

	// Read all sections
	for {
		var line string
		if bufferedLine != nil {
			line = *bufferedLine
			bufferedLine = nil
		} else {
			if !scanner.Scan() {
				break
			}
			line = strings.TrimSpace(scanner.Text())
		}

		if line == "[TEXT SECTION]" {
			if err := readSection("TEXT"); err != nil {
				return nil, fmt.Errorf("error reading TEXT SECTION: %w", err)
			}
		} else if line == "[STRUCTURE SECTION]" {
			if err := readSection("STRUCTURE"); err != nil {
				return nil, fmt.Errorf("error reading STRUCTURE SECTION: %w", err)
			}
		} else if line == "[TAGS SECTION]" {
			if err := readSection("TAGS"); err != nil {
				return nil, fmt.Errorf("error reading TAGS SECTION: %w", err)
			}
		} else if line == "[ATTRIBUTES SECTION]" {
			if err := readSection("ATTRIBUTES"); err != nil {
				return nil, fmt.Errorf("error reading ATTRIBUTES SECTION: %w", err)
			}
		} else if line == "[TIMESTAMPS SECTION]" {
			if err := readSection("TIMESTAMPS"); err != nil {
				return nil, fmt.Errorf("error reading TIMESTAMPS SECTION: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// Build items from parsed data
	itemsByID := make(map[string]*model.Item)
	for id, text := range textData {
		var metadata *model.Metadata
		if ts, ok := timestampsData[id]; ok {
			metadata = &model.Metadata{
				Tags:       tagsData[id],
				Attributes: attributesData[id],
				Created:    ts[0],
				Modified:   ts[1],
			}
		} else {
			metadata = &model.Metadata{
				Tags:       tagsData[id],
				Attributes: attributesData[id],
				Created:    time.Now(),
				Modified:   time.Now(),
			}
		}

		if metadata.Tags == nil {
			metadata.Tags = make([]string, 0)
		}
		if metadata.Attributes == nil {
			metadata.Attributes = make(map[string]string)
		}

		item := &model.Item{
			ID:       id,
			Text:     text,
			Children: make([]*model.Item, 0),
			Metadata: metadata,
		}
		itemsByID[id] = item
	}

	// Build parent-child relationships
	rootItems := make([]*model.Item, 0)
	childrenByParent := make(map[string][]*model.Item)

	// Sort items by parent and position for proper ordering
	type itemWithPosition struct {
		item     *model.Item
		position int
	}

	for id, item := range itemsByID {
		structure := structureData[id]
		parentID := structure[0]

		if parentID == "" {
			// Root item
			rootItems = append(rootItems, item)
		} else {
			parent := itemsByID[parentID]
			if parent != nil {
				item.Parent = parent
				childrenByParent[parentID] = append(childrenByParent[parentID], item)
			}
		}
	}

	// Parse and sort root items by position
	rootWithPos := make([]struct {
		item     *model.Item
		position int
	}, len(rootItems))
	for i, item := range rootItems {
		pos := parsePosition(structureData[item.ID][1])
		rootWithPos[i] = struct {
			item     *model.Item
			position int
		}{item, pos}
	}
	sort.SliceStable(rootWithPos, func(i, j int) bool {
		return rootWithPos[i].position < rootWithPos[j].position
	})
	rootItems = make([]*model.Item, len(rootWithPos))
	for i, rp := range rootWithPos {
		rootItems[i] = rp.item
	}

	// Parse and sort children by position
	for parentID, children := range childrenByParent {
		parent := itemsByID[parentID]
		if parent != nil {
			childWithPos := make([]struct {
				item     *model.Item
				position int
			}, len(children))
			for i, item := range children {
				pos := parsePosition(structureData[item.ID][1])
				childWithPos[i] = struct {
					item     *model.Item
					position int
				}{item, pos}
			}
			sort.SliceStable(childWithPos, func(i, j int) bool {
				return childWithPos[i].position < childWithPos[j].position
			})
			sortedChildren := make([]*model.Item, len(childWithPos))
			for i, cp := range childWithPos {
				sortedChildren[i] = cp.item
			}
			parent.Children = sortedChildren
		}
	}

	outline.Items = rootItems
	return outline, nil
}

// collectAllItems recursively collects all items from the outline tree
func collectAllItems(items []*model.Item, byID map[string]*model.Item, all *[]*model.Item) {
	for _, item := range items {
		byID[item.ID] = item
		*all = append(*all, item)
		if len(item.Children) > 0 {
			collectAllItems(item.Children, byID, all)
		}
	}
}

// encodeTextValue encodes a text value with proper escape sequence handling
// Backslashes are escaped first, then newlines
func encodeTextValue(text string) string {
	var result strings.Builder
	for _, ch := range text {
		switch ch {
		case '\\':
			result.WriteString("\\\\")
		case '\n':
			result.WriteString("\\n")
		default:
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// decodeTextValue decodes a text value with proper escape sequence parsing
// Reads character by character to handle \n and \\ correctly
func decodeTextValue(text string) string {
	var result strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] == '\\' && i+1 < len(text) {
			next := text[i+1]
			if next == 'n' {
				result.WriteRune('\n')
				i++ // Skip the 'n'
			} else if next == '\\' {
				result.WriteRune('\\')
				i++ // Skip the second backslash
			} else {
				// Unrecognized escape sequence, treat as literal
				result.WriteByte('\\')
			}
		} else {
			result.WriteByte(text[i])
		}
	}
	return result.String()
}

// parseTextLine parses a line from the TEXT SECTION
// Format: id: text
func parseTextLine(line string) (id string, text string, err error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid text line format: %s", line)
	}

	id = strings.TrimSpace(parts[0])
	rawText := strings.TrimSpace(parts[1])
	text = decodeTextValue(rawText)

	return id, text, nil
}

// parseStructureLine parses a line from the STRUCTURE SECTION
// Format: id: parent_id:position
func parseStructureLine(line string) (id string, parentID string, position string, err error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid structure line format: %s", line)
	}

	id = strings.TrimSpace(parts[0])
	rest := strings.SplitN(parts[1], ":", 2)
	if len(rest) != 2 {
		return "", "", "", fmt.Errorf("invalid structure format (missing position): %s", line)
	}

	parentID = strings.TrimSpace(rest[0])
	position = strings.TrimSpace(rest[1])

	return id, parentID, position, nil
}

// parseTagsLine parses a line from the TAGS SECTION
// Format: id: tag1,tag2,tag3
func parseTagsLine(line string) (id string, tags []string, err error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid tags line format: %s", line)
	}

	id = strings.TrimSpace(parts[0])
	tagsStr := strings.TrimSpace(parts[1])

	if tagsStr == "" {
		tags = make([]string, 0)
	} else {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	return id, tags, nil
}

// parseAttributesLine parses a line from the ATTRIBUTES SECTION
// Format: id: key1=value1,key2=value2
func parseAttributesLine(line string) (id string, attrs map[string]string, err error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid attributes line format: %s", line)
	}

	id = strings.TrimSpace(parts[0])
	attrStr := strings.TrimSpace(parts[1])
	attrs = make(map[string]string)

	if attrStr != "" {
		pairs := strings.Split(attrStr, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				attrs[key] = value
			}
		}
	}

	return id, attrs, nil
}

// parsePosition parses a position string to an integer
func parsePosition(posStr string) int {
	var pos int
	fmt.Sscanf(posStr, "%d", &pos)
	return pos
}

// parseTimestampsLine parses a line from the TIMESTAMPS SECTION
// Format: id: created modified
func parseTimestampsLine(line string) (id string, created, modified time.Time, err error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", time.Time{}, time.Time{}, fmt.Errorf("invalid timestamps line format: %s", line)
	}

	id = strings.TrimSpace(parts[0])
	timesStr := strings.TrimSpace(parts[1])
	times := strings.SplitN(timesStr, " ", 2)

	if len(times) != 2 {
		return "", time.Time{}, time.Time{}, fmt.Errorf("invalid timestamps format: %s", line)
	}

	created, err = time.Parse(time.RFC3339Nano, strings.TrimSpace(times[0]))
	if err != nil {
		return "", time.Time{}, time.Time{}, fmt.Errorf("invalid created timestamp: %w", err)
	}

	modified, err = time.Parse(time.RFC3339Nano, strings.TrimSpace(times[1]))
	if err != nil {
		return "", time.Time{}, time.Time{}, fmt.Errorf("invalid modified timestamp: %w", err)
	}

	return id, created, modified, nil
}
