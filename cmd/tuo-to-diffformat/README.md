# tuo-to-diffformat

Converts TUI Outliner JSON files to the diff-optimized format for easy diffing and version comparison.

## Overview

The diff-optimized format is designed to work seamlessly with standard diff tools like `diff`, `diff3`, and `patch`. Each section is sorted alphabetically by ID, making structural changes and content changes immediately visible.

## Installation

```bash
go build -o tuo-to-diffformat ./cmd/tuo-to-diffformat
sudo mv tuo-to-diffformat /usr/local/bin/
```

## Usage

### Basic Usage

```bash
# Convert and print to stdout
tuo-to-diffformat my_outline.json

# Convert and save to file
tuo-to-diffformat my_outline.json my_outline.txt
```

### Examples

#### Convert a single file
```bash
tuo-to-diffformat ~/documents/projects.json ~/documents/projects.txt
```

#### Convert all JSON files in a directory
```bash
for f in *.json; do
  tuo-to-diffformat "$f" "${f%.json}.txt"
done
```

#### Create a git history of changes
```bash
# Initial commit
tuo-to-diffformat outline.json outline.txt
git add outline.txt
git commit -m "Initial outline"

# Later, after edits...
tuo-to-diffformat outline.json outline.txt
git diff outline.txt  # See exactly what changed
git add outline.txt
git commit -m "Updated outline"
```

## Format Explanation

### Text Section
```
[TEXT SECTION]
task-1: Buy groceries
task-2: Milk and eggs
task-3: Fresh vegetables\nCarrots\nBroccoli
```

- Each line: `id: escaped_text`
- Newlines in text are escaped as `\n`
- Backslashes are escaped as `\\`
- Sorted alphabetically by ID

### Structure Section
```
[STRUCTURE SECTION]
task-1: :0
task-2: task-1:0
task-3: task-1:1
task-4: :1
```

- Format: `id: parent_id:position`
- Root items have empty parent: `task-1: :0`
- Position is 0-based within the parent's children
- Easily spot when items are reordered (position changes) or moved (parent changes)

### Tags Section
```
[TAGS SECTION]
task-1: shopping,urgent
task-3: produce
```

- Format: `id: tag1,tag2,tag3`
- Only items with tags appear
- Comma-separated tags sorted by ID

### Attributes Section
```
[ATTRIBUTES SECTION]
task-1: priority=high,status=todo
task-3: status=todo
```

- Format: `id: key1=value1,key2=value2`
- Only items with attributes appear
- Key-value pairs sorted by key within each item
- Great for tracking status changes

### Timestamps Section
```
[TIMESTAMPS SECTION]
task-1: 2025-11-01T10:00:00Z 2025-11-03T15:30:00Z
task-2: 2025-11-01T10:15:00Z 2025-11-02T09:00:00Z
```

- Format: `id: created_timestamp modified_timestamp`
- ISO8601 format for easy parsing and diffing
- Shows exactly when items were created and last modified

## Why Use This Format?

### For Version Control

```bash
# See exactly what changed between versions
diff version1.txt version2.txt

# Apply changes with patch
patch < changes.patch

# Merge multiple changes
diff3 mine.txt original.txt theirs.txt
```

### For Analysis

```bash
# Find all items with a specific tag
grep ": .*urgent" outline.txt

# Find all items modified after a specific date
awk -F'[ T]' '$2 > "2025-11-03"' outline.txt

# Count items
grep "^[^:]*: [^:]*$" outline.txt | wc -l
```

### For Diffs

The line-based format with alphabetical sorting means:

1. **Item creation shows as new lines** in the TEXT, STRUCTURE, TAGS, ATTRIBUTES, and TIMESTAMPS sections
2. **Item deletion shows as removed lines** across all sections
3. **Text changes show on the same line** - easy to spot
4. **Status/attribute changes show on the same line** - easy to track
5. **Reordering shows as position changes** - `task-1: parent:0` → `task-1: parent:2`
6. **Moving shows as parent changes** - `task-1: :0` → `task-1: other-parent:0`

## Reverse Conversion

To convert back to JSON, use `diffformat-to-json`:

```bash
diffformat-to-json my_outline.txt my_outline.json
```

This is useful for:
- Restoring from a diff-formatted backup
- Converting between formats
- Verifying round-trip integrity

## Round-Trip Guarantees

The format is fully reversible:

```bash
# Create a test
tuo-to-diffformat original.json temp.txt
diffformat-to-json temp.txt restored.json

# Verify the data is identical
diff <(jq -S . original.json) <(jq -S . restored.json)
```

All data is preserved:
- ✓ Item text (including newlines and backslashes)
- ✓ Tree structure (parent-child relationships and ordering)
- ✓ Tags
- ✓ Attributes (including custom key-value pairs)
- ✓ Timestamps (creation and modification)
