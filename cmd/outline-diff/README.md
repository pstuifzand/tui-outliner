# outline-diff

Compare two TUI Outliner JSON files and show meaningful, per-item changes.

## Purpose

This command intelligently compares backup versions of the same outline document at different times, making it immediately clear what changed on each specific item. Instead of raw line-based diffs, it analyzes changes by item ID and shows:

- New items added (with their metadata)
- Deleted items removed
- Modified items with specific changes:
  - Text edits (old → new)
  - Attribute additions, removals, and changes (old → new values)
  - Tag additions and removals
  - Structure changes (parent moves, position changes)
  - Modification timestamps

## Building

```bash
cd /home/peter/work/tui-outliner
go build -o outline-diff ./cmd/outline-diff
```

## Usage

### Single-File Mode (Backup History)

Show all changes across the backup history of a single file:

```bash
# View changes across all backups
outline-diff my_outline.json

# Show summary of changes only (no item details)
outline-diff -s my_outline.json

# Verbose mode for detailed output
outline-diff -v my_outline.json
```

This automatically finds all backups for the specified file and shows consecutive diffs between them, making it easy to review your editing history.

### Two-File Mode (Direct Comparison)

Compare two specific outline files:

```bash
# Compare any two outline files
outline-diff backup1.json backup2.json

# Summary only
outline-diff -s backup1.json backup2.json

# Verbose output
outline-diff -v backup1.json backup2.json
```

## Output Format

The output is organized by type of change and item ID:

```
=== Outline Diff: backup1.json → backup2.json ===

New Items:

  task-4: Fresh vegetables
    PARENT: task-1 at position 1
    TAGS: produce
    ATTR: type = fresh

Modified Items:

  task-1: Buy groceries and supplies
    TEXT: Buy groceries → Buy groceries and supplies
    TAGS added: urgent
    ATTR added: status = in-progress
    MODIFIED: 2025-11-01T10:00:00 → 2025-11-03T14:30:00

  task-2: Milk, eggs, and butter
    TEXT: Milk and eggs → Milk, eggs, and butter
    ATTR added: aisle = 3
    MODIFIED: 2025-11-01T10:15:00 → 2025-11-03T14:20:00

  task-3: Write report
    ATTR changed: status: todo → done
    MODIFIED: 2025-11-01T11:00:00 → 2025-11-03T15:00:00

Deleted Items:

  task-5: Old item

=== Summary ===
  3 items modified
  1 items added
  1 items deleted
```

## Understanding the Output

Each item is shown with its ID and current text. Changes are indented and clearly labeled:

- **TEXT**: Shows old text → new text
- **TAGS added**: Lists specific tags that were added
- **TAGS removed**: Lists specific tags that were removed
- **ATTR added**: Shows attribute name = value for new attributes
- **ATTR changed**: Shows attribute name: old value → new value
- **ATTR removed**: Shows attribute name (was: old value) for removed attributes
- **MOVED**: Shows parent change when item moved to different parent
- **POSITION**: Shows position change (0-based index within parent)
- **MODIFIED**: Shows timestamp change (YYYY-MM-DDTHH:MM:SS format)

## Use Cases

### Review Editing History of a File

```bash
# See all changes across the entire backup history of a file
outline-diff my_important_outline.json

# This will show a sequence of diffs:
# --- 2025-11-01 10:05:00 (backup 1)
# +++ 2025-11-01 10:15:32 (backup 2)
# ... (changes)
#
# --- 2025-11-01 10:15:32 (backup 2)
# +++ 2025-11-01 10:32:45 (backup 3)
# ... (changes)
```

### Quick Summary of Session Activity

```bash
# See just the counts of modifications without details
outline-diff -s my_outline.json
```

### Compare Two Specific Backups

```bash
# See exact differences between two specific points in time
outline-diff ~/.local/share/tui-outliner/backups/20251103_140405_abc12345.tuo 20251103_143015_abc12345.tuo
```

### Analyze Specific Item Changes

The output clearly shows which attributes changed on which items, making it easy to:
- Verify your edits were saved correctly
- See what metadata was added/removed
- Understand item reorganization (moves and reordering)
- Track when items were last modified
- Review the complete history of how an item evolved

## How It Works

1. Reads both JSON files using streaming JSON decoder
2. Converts each to the diff-optimized format (same as `tuo-to-diffformat`)
3. Parses diff format into structured data by item ID
4. Intelligently analyzes changes:
   - Identifies new items (ID only in file2)
   - Identifies deleted items (ID only in file1)
   - Identifies modified items (ID in both, compares all fields)
5. Groups and displays changes by item with clear labels

The approach of using item IDs as the key ensures:
- Changes are clearly associated with specific items
- Text changes, attribute changes, and structure changes are all shown together
- It's immediately obvious what happened to each item
- No noise from unchanged items (unless in verbose mode)

## Technical Details

- Uses the diff-optimized format as an intermediate for consistent parsing
- Preserves all metadata: text, tags, attributes, structure, timestamps
- Handles multi-line text by showing only the first line (truncated if long)
- Timestamps displayed in YYYY-MM-DDTHH:MM:SS format for readability
- Item IDs are always shown for reference and traceability

## See Also

- `tuo-to-diffformat` - Convert JSON to diff format
- `diffformat-to-json` - Convert diff format back to JSON
- `DIFFFORMAT.md` - Complete format specification
- `internal/storage/backup.go` - Automatic backup system
