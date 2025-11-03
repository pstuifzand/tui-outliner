# Diff-Optimized Format for TUI Outliner

This document describes the diff-optimized format for TUI Outliner outlines, designed for easy version control, diffing, and change tracking.

## Overview

The diff-optimized format is a line-based text format that represents an outline in a way that's perfect for:

- **Git integration**: Each change appears clearly in `git diff`
- **Patch files**: Standard `patch` and `diff3` tools work natively
- **Change tracking**: Text changes, structural changes, and metadata changes are all visible
- **Merging**: Three-way merges with `diff3` are straightforward
- **History analysis**: Easy to parse and analyze with standard Unix tools

## Format Structure

The format consists of 5 sections, each sorted alphabetically by item ID:

### 1. TEXT SECTION

```
[TEXT SECTION]
task-1: Buy groceries
task-2: Milk and eggs
task-3: Fresh vegetables\nCarrots\nBroccoli
task-4: Write report (C:\\path\\to\\file.txt)
```

**Purpose**: Stores the text content of each item

**Format**: `id: escaped_text`

**Escaping Rules**:
- Newline characters (`\n`) are represented as literal `\n`
- Backslash characters (`\`) are represented as `\\`
- All other characters are literal

**Sorting**: Alphabetically by item ID

### 2. STRUCTURE SECTION

```
[STRUCTURE SECTION]
task-1: :0
task-2: task-1:0
task-3: task-1:1
task-4: :1
```

**Purpose**: Defines the tree hierarchy and ordering

**Format**: `id: parent_id:position`

**Details**:
- `parent_id` is empty for root items
- `position` is the 0-based index within the parent's children array
- Root items are ordered at the top level
- Changes to position indicate reordering
- Changes to parent indicate moving items

**Sorting**: Alphabetically by item ID

### 3. TAGS SECTION

```
[TAGS SECTION]
task-1: shopping,urgent
task-3: produce
task-4: work
```

**Purpose**: Stores tags associated with items

**Format**: `id: tag1,tag2,tag3`

**Details**:
- Only items with tags appear in this section
- Tags are comma-separated
- Sorted alphabetically within each item (by tag name)
- Sorted alphabetically by item ID

### 4. ATTRIBUTES SECTION

```
[ATTRIBUTES SECTION]
task-1: priority=high,status=todo
task-2: status=done
task-3: status=todo
task-4: assigned_to=John\Doe,status=in-progress
```

**Purpose**: Stores custom attributes (key-value pairs)

**Format**: `id: key1=value1,key2=value2`

**Details**:
- Only items with attributes appear
- Attribute pairs are comma-separated
- Keys are sorted alphabetically within each item
- Sorted alphabetically by item ID
- Values preserve backslashes and other characters

### 5. TIMESTAMPS SECTION

```
[TIMESTAMPS SECTION]
task-1: 2025-11-01T10:00:00Z 2025-11-03T15:30:00Z
task-2: 2025-11-01T10:15:00Z 2025-11-02T09:00:00Z
task-3: 2025-11-01T10:20:00Z 2025-11-01T10:20:00Z
task-4: 2025-11-02T08:00:00Z 2025-11-03T14:00:00Z
```

**Purpose**: Tracks when items were created and last modified

**Format**: `id: created_timestamp modified_timestamp`

**Details**:
- Timestamps are in ISO8601 format (RFC3339Nano)
- Separated by a single space
- Both creation and modification times are preserved
- Sorted alphabetically by item ID

## Usage

### Converting from JSON to Diff Format

Use the `tuo-to-diffformat` tool:

```bash
# Convert to stdout
tuo-to-diffformat outline.json

# Convert to file
tuo-to-diffformat outline.json outline.txt
```

### Converting Back to JSON

Use the `diffformat-to-json` tool:

```bash
# Convert to stdout
diffformat-to-json outline.txt

# Convert to file
diffformat-to-json outline.txt outline.json
```

### Round-Trip Verification

All data is preserved in a round-trip conversion:

```bash
# Create a test outline
echo '{"items":[{"id":"1","text":"test"}]}' > test.json

# Convert to diff format
tuo-to-diffformat test.json test.txt

# Convert back to JSON
diffformat-to-json test.txt restored.json

# Verify the data is identical
diff <(jq -S . test.json) <(jq -S . restored.json)
```

## Why This Format is Perfect for Diffing

### 1. Line-Based Representation

Each piece of information is on a single line, making it compatible with all standard diff tools.

### 2. Alphabetical Sorting

All items are sorted by ID within each section. This means:
- New items show as consecutive new lines (not scattered throughout)
- Deleted items show as consecutive deleted lines
- Same items appear on the same line numbers in different versions
- Order changes within a parent are immediately visible

### 3. Structured Sections

Different types of changes appear in different sections:
- **TEXT changes**: Visible in TEXT SECTION (content changes)
- **STRUCTURE changes**: Visible in STRUCTURE SECTION (reordering, moving)
- **METADATA changes**: Visible in TAGS, ATTRIBUTES, TIMESTAMPS sections

Example:

```diff
--- version1.txt
+++ version2.txt
 [TEXT SECTION]
 task-1: Buy groceries
-task-2: Milk and eggs
+task-2: Milk, eggs, and butter
 task-3: Vegetables

 [STRUCTURE SECTION]
 task-1: :0
-task-2: task-1:0
-task-3: task-1:1
+task-2: task-1:1
+task-3: task-1:0

 [ATTRIBUTES SECTION]
 task-2: status=todo
+task-2: quantity=1 carton,status=todo
```

This diff clearly shows:
- task-2 text was updated
- task-2 and task-3 were reordered (positions swapped)
- task-2 got a new attribute

### 4. Semantic Clarity

Each type of change is immediately meaningful:

```
- task-1: :0              # Was at position 0 in root
+ task-1: parent-a:2      # Now at position 2 under parent-a
                          # Item was moved!

- task-2: status=todo     # Had status=todo
+ task-2: status=done     # Now has status=done
                          # Status was updated!

- task-3: Incomplete task  # Old text
+ task-3: Completed task   # New text
                          # Description was updated!
```

## Use Cases

### Version Control with Git

```bash
# Initialize tracking
git init
tuo-to-diffformat outline.json outline.txt
git add outline.txt
git commit -m "Initial outline"

# Later, after editing in TUI Outliner...
tuo-to-diffformat outline.json outline.txt
git diff outline.txt  # See exactly what changed
git add outline.txt
git commit -m "Updated project timeline"
```

### Merging Changes

```bash
# Your version
tuo-to-diffformat yours.json yours.txt

# Their version
tuo-to-diffformat theirs.json theirs.txt

# Common ancestor
tuo-to-diffformat original.json original.txt

# Three-way merge
diff3 yours.txt original.txt theirs.txt

# Or with patch
diff original.txt theirs.txt > their_changes.patch
patch yours.txt < their_changes.patch
```

### Analyzing Changes

```bash
# Find all items modified today
grep "2025-11-03T" outline.txt

# Find all items with status=done
grep "status=done" outline.txt

# Find items in a specific project
grep "project-a:" outline.txt

# Count items by tag
grep "^[^:]*: " outline.txt | grep ":" | cut -d: -f2 | tr ',' '\n' | sort | uniq -c
```

### Backup and Recovery

```bash
# Regular backups
for date in $(seq 1 30); do
  tuo-to-diffformat outline.json backups/outline_day${date}.txt
done

# Easily see progression
diff backups/outline_day1.txt backups/outline_day30.txt

# Restore from specific date
diffformat-to-json backups/outline_day15.txt outline_restored.json
```

## Implementation Details

### Text Escaping

The format uses a proper escape parser that handles:

```
Original Text           Encoded           Decoded Back
Hello                   Hello             Hello
Hello\nWorld            Hello\nWorld      Hello
                                          World
C:\path\file            C:\\path\\file    C:\path\file
Line1\nC:\test          Line1\nC:\\test   Line1
                                          C:\test
```

The decoder is character-by-character aware:
- When it sees `\`, it checks the next character
- `\n` → newline character
- `\\` → single backslash
- `\x` (other) → error or literal handling

### ID Sorting

Items are sorted using standard alphabetical ordering:

```
item-1
item-2
item-10
item-20
task-1
task-2
zzz-item
```

This ensures consistent output and efficient diffing.

### Parent-Child Relationships

The format preserves both:
- **Parent identity**: Which item is the parent
- **Child ordering**: Position within parent's children array

Examples:

```
Root level items:       item-1: :0
                       item-2: :1

Children of item-1:     item-3: item-1:0
                       item-4: item-1:1

Grandchildren:         item-5: item-3:0
                       item-6: item-3:1
```

Reordering changes the position:
```
- item-3: item-1:0     # Was first child
- item-4: item-1:1     # Was second child
+ item-3: item-1:1     # Now second child
+ item-4: item-1:0     # Now first child
```

Moving changes the parent:
```
- item-3: item-1:0     # Was under item-1
+ item-3: item-2:0     # Now under item-2
```

## File Extension

The recommended file extension is `.txt`, though any extension can be used:
- `.txt` - General text outline format
- `.tuo.txt` - TUI Outliner diff format
- `.diff` - For version control history
- `.patch` - For diff/patch files

## Performance Characteristics

- **Encoding**: O(n) where n is number of items
- **Decoding**: O(n) with one pass through the file
- **Diff**: O(n) where n is total lines in both files
- **File size**: Typically 30-50% of JSON size (less whitespace, no structure overhead)

## Compatibility

The format is compatible with:
- ✓ Standard `diff` and `patch` utilities
- ✓ Git's built-in diff and merge tools
- ✓ `diff3` for three-way merges
- ✓ All standard text editors
- ✓ grep, awk, sed, and other Unix tools
- ✓ Python, Go, and other scripting languages (simple parsing)

## Limitations

- **Circular references**: Not supported (assumes valid tree structure)
- **Virtual children**: Not preserved in this format (structure is literal tree)
- **Metadata beyond ID/text**: Only tags, attributes, and timestamps preserved

## Future Extensions

Possible extensions to the format:

```
[VIRTUAL CHILDREN SECTION]
search-node-1: item-1,item-5,item-9

[CUSTOM FIELDS SECTION]
item-1: field1=value1

[BINARY ATTACHMENTS SECTION]
item-1: image.png base64...
```

These can be added without breaking the core format.
