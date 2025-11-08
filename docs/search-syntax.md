# TUI Outliner Search Syntax

This document describes the search filter syntax for finding nodes in your outline.

## Overview

The search syntax provides a powerful way to filter nodes using multiple criteria. Filters are combined using operators to create complex search expressions.

**Basic syntax:**
```
filter1 filter2 -filter3 | filter4
```

This means: `(filter1 AND filter2 AND NOT filter3) OR filter4`

## Filter Types

### Text Filter (Implicit)

Plain text searches the node's text content (case-insensitive substring match).

```
task                 # Find nodes containing "task"
"multi word"         # Quoted text for exact phrase search
TODO                 # Case-insensitive
```

### Regex Filter: `/pattern/`

Match nodes using regular expression patterns. Use Go's regex syntax (RE2).

**Syntax:** `/pattern/`

```
/^TODO/              # Items starting with TODO
/\d{4}-\d{2}-\d{2}/  # Items containing dates in YYYY-MM-DD format
/@[a-z]+/            # Items mentioning usernames (@username)
/(?i)bug/            # Case-insensitive match for "bug"
/bug|issue|problem/  # Match any of these words
```

**Important Notes:**
- Uses Go's RE2 regex engine (not all Perl features supported)
- No support for negative lookahead `(?!)` or lookbehind `(?<=)`
- Backslashes must be escaped: `\/` for literal slash
- Supports common features: anchors (`^`, `$`), character classes (`\d`, `\w`, `[a-z]`), quantifiers (`*`, `+`, `?`, `{n,m}`), groups, alternation

**Use cases:**
- Complex text patterns beyond simple substring matching
- Extract structured data from text (dates, emails, etc.)
- Pattern-based filtering with anchors and boundaries

**Examples:**

```
/^TODO/ d:0          # Root-level items starting with TODO
-/^SKIP/             # Items NOT starting with SKIP
/bug|issue/ @status=open  # Open bugs or issues
```

### Depth Filter: `d:`

Match nodes at specific depth levels.

**Syntax:** `d:N` | `d:>=N` | `d:>N` | `d:<=N` | `d:<N`

```
d:0              # Root level nodes only
d:1              # First level (children of root)
d:>2             # Nodes deeper than level 2
d:>=2            # Level 2 and deeper
d:<3             # Shallower than level 3
```

### Attribute Filter: `@`

Match nodes with specific attributes using the `@` prefix.

**Syntax:** `@KEY` | `@KEY=VALUE` | `@KEY!=VALUE` | `@KEY>VALUE` (for dates)

```
@url            # Has 'url' attribute (any value)
@type=day       # Has 'type' attribute with value 'day'
@type!=day      # Has 'type' attribute but NOT 'day'
@status=done    # Exact match on attribute value
```

**Date-based Attribute Filtering:**

If an attribute contains a date value (YYYY-MM-DD format), you can filter using date comparisons:

```
@date>7d        # Attribute 'date' is more recent than 7 days ago (shortcut)
@date>-7d       # Attribute 'date' is more recent than 7 days ago (explicit)
@date<30d       # Attribute 'date' is older than 30 days ago
@date>=1d       # Attribute 'date' is within the last day
@date=2025-11-01 # Attribute 'date' is exactly this date
@date>=2025-10-01 # Attribute 'date' is on or after this date
@date<=2025-11-30 # Attribute 'date' is on or before this date
```

**Examples with date attributes:**

```
@deadline>7d           # Items with deadlines in the next 7 days
@completed<7d          # Items completed more than a week ago
@start>=1d m:>3d       # Recently started items, modified in last 3 days
```

### Creation Date Filter: `c:`

Match nodes created within a time window.

**Syntax:** `c:DATE` | `c:>=DATE` | `c:>DATE` | `c:<=DATE` | `c:<DATE`

Date formats:
- Relative (shortcut): `1h`, `2h`, `1d`, `7d`, `30d`, `1w`, `4w`, `1m`, `6m`, `1y` (means "N units ago")
- Relative (explicit): `-1h`, `-7d`, `-30d`, `-1w`, `-4w`, `-1m`, `-6m`, `-1y` (same as shortcut)
- Absolute: `2025-11-01` (YYYY-MM-DD format)

Time units:
- `h` - hours (e.g., `1h` = 1 hour ago, `24h` = 24 hours ago)
- `d` - days (e.g., `7d` = 7 days ago)
- `w` - weeks (e.g., `1w` = 1 week ago)
- `m` - months (e.g., `1m` = 1 month ago)
- `y` - years (e.g., `1y` = 1 year ago)

```
c:>1h            # Created in the last hour
c:>24h           # Created in the last 24 hours
c:>7d            # Created in the last 7 days (shortcut syntax)
c:>-7d           # Created in the last 7 days (explicit syntax, same as above)
c:<30d           # Created more than 30 days ago
c:>=2025-11-01   # Created on or after this date
```

### Modified Date Filter: `m:`

Match nodes modified within a time window. Same syntax as creation date filter.

```
m:>1h            # Modified in the last hour
m:>2h            # Modified in the last 2 hours
m:>7d            # Modified in the last 7 days
m:<1d            # Modified more than 1 day ago
```

### Children Count Filter: `children:`

Match nodes based on the number of children they have.

**Syntax:** `children:N` | `children:>=N` | `children:>N` | `children:<=N` | `children:<N`

```
children:0       # Leaf nodes (no children)
children:>0      # Nodes with at least one child
children:5       # Nodes with exactly 5 children
children:>=3     # Nodes with 3 or more children
children:<10     # Nodes with fewer than 10 children
```

### Parent Filter: `p:`

Match nodes whose parent matches criteria. Allows filtering by parent's properties. Parent filters support all filter types including regex.

**Syntax:** `p:FILTER`

```
p:d:0            # Nodes whose parent is at root level
p:@type=project  # Nodes whose parent has 'type=project' attribute
p:d:>=2          # Nodes whose parent is at depth 2 or deeper
p:/^TODO/        # Nodes whose parent text starts with TODO (regex)
p:/\d{4}/        # Nodes whose parent contains a 4-digit year
```

### Ancestor Filter: `a:` or `parent*:`

Match nodes that have an ancestor matching criteria. Ancestor filters support all filter types including regex.

**Syntax:** `a:FILTER` or `parent*:FILTER`

```
a:@type=project    # Nodes anywhere under a node with 'type=project' attribute
a:d:0              # Any node with root as ancestor (all non-root)
a:@status=active   # Nodes under a node with 'status=active' attribute
a:/^Project/       # Nodes with an ancestor starting with "Project" (regex)
parent*:project    # Same as a:project (alternate syntax)
```

### Sibling Filter: `sibling:` or `s:`

Match nodes based on their siblings (items sharing the same parent). Siblings are items at the same hierarchical level with the same parent node.

**Syntax:** `sibling:FILTER` | `+sibling:FILTER` | `-sibling:FILTER` | `s:FILTER`

#### Quantifier Prefixes

- **No prefix (default)**: At least one sibling matches (some)
- **`+` prefix**: All siblings must match (all)
- **`-` prefix**: No siblings must match (none)

**Examples:**
```
sibling:@status=done        # At least one sibling has status=done
+sibling:@status=done       # All siblings have status=done
-sibling:@status=done       # No siblings have status=done
s:@priority=high            # Shorthand: at least one sibling has high priority
sibling:task                # At least one sibling contains "task" in text
+sibling:d:2                # All siblings are at depth 2
```

**Special Cases:**
- Root items (nodes with no parent) have no siblings
- Only children (nodes with no other siblings) behave like having empty sibling set
- For items with no siblings:
  - `sibling:FILTER` returns `false` (no siblings to match)
  - `+sibling:FILTER` returns `false` (can't be "all" if none exist)
  - `-sibling:FILTER` returns `true` (vacuously true, no siblings to violate condition)

### Child and Descendant Filters

Match nodes based on their children or descendants. These filters support **quantifiers** to specify how many children/descendants must match.

#### Quantifier Prefixes

- **No prefix (default)**: At least one matches (some)
- **`+` prefix**: All must match (all)
- **`-` prefix**: None must match (none)

#### Child Filter: `child:`

Match nodes based on their immediate children (one level down).

**Syntax:** `child:FILTER` | `+child:FILTER` | `-child:FILTER`

```
child:task              # At least one child contains "task"
+child:@status=done     # All children have status=done
-child:@urgent          # No children have urgent attribute
child:d:>2              # At least one child is at depth > 2
```

#### Descendant Filter: `child*:`

Match nodes based on all their descendants (recursive, all levels down).

**Syntax:** `child*:FILTER` | `+child*:FILTER` | `-child*:FILTER`

```
child*:task             # At least one descendant contains "task"
+child*:@status=done    # All descendants have status=done
-child*:@urgent         # No descendants have urgent attribute
child*:d:>3             # At least one descendant is at depth > 3
```

#### Parent* Filter (Ancestor with Quantifiers): `parent*:`

Match nodes based on all their ancestors (recursive, all levels up).

**Syntax:** `parent*:FILTER` | `+parent*:FILTER` | `-parent*:FILTER`

```
parent*:project         # At least one ancestor contains "project" (same as a:project)
+parent*:@type=section  # All ancestors have type=section
-parent*:@archived      # No ancestors are archived
parent*:d:0             # At least one ancestor is at root (all non-root nodes)
```

#### Empty Set Semantics

When there are no children/siblings/ancestors, quantifiers behave as follows:

- **All (`+`)**:
  - Children/Descendants/Siblings: `false` (can't be "all" if none exist)
  - Ancestors: `true` (vacuously true for root nodes)
- **Some (default)**: Always `false` (none exist to match)
- **None (`-`)**: Always `true` (vacuously true, none exist to match)

**Examples:**
```
+child:task             # Root nodes with no children: false
-child:task             # Root nodes with no children: true
+sibling:@done          # Only children with no siblings: false
-sibling:@done          # Only children with no siblings: true
+parent*:project        # Root nodes with no ancestors: true
-parent*:archived       # Root nodes with no ancestors: true
```

## Operators

### AND Operator (implicit or `+`)

Filters are AND-ed by default when separated by spaces.

```
task d:>2              # Equivalent to: task AND d:>2
+d:>2 children:>0      # Explicit: d:>2 AND children:>0
```

### OR Operator: `|`

Match nodes that satisfy ANY of the criteria.

```
d:0 | d:1             # Root level OR first level nodes
@urgent | @important  # Nodes with 'urgent' OR 'important' attribute
```

### NOT Operator: `-`

Exclude nodes matching the filter.

```
-children:0            # Non-leaf nodes (nodes with children)
task -@done            # "task" but NOT marked as done
```

## Operator Precedence

From highest to lowest:

1. **Filter atoms** (text, `d:`, `a:`, etc.)
2. **NOT** (`-`)
3. **AND** (space or `+`)
4. **OR** (`|`)

**Examples:**

```
a b | c d        # (a AND b) OR (c AND d)
-a b | c         # (NOT a AND b) OR c
a | -b c         # a OR (NOT b AND c)
```

Use parentheses for explicit grouping (if supported by parser):

```
(a | b) +c       # (a OR b) AND c
```

## Common Patterns

### Find nodes by type
```
@type=day               # All daily note items
@type=project           # All project nodes
@type=day d:<2          # Daily notes that are top-level or level 1
```

### Find actionable items
```
-@status=done +children:>0     # Incomplete projects (have children)
-@done task                    # Incomplete tasks
@urgent -@done                 # Urgent incomplete items
```

### Find recently modified
```
m:>-7d -@archived       # Modified this week, not archived
c:>-1d                  # Created today
m:<-30d c:>-90d         # Created in last 90 days but not touched in 30 days
```

### Find structural patterns
```
children:0              # All leaf nodes
children:>0             # All parent nodes
d:2 children:>5         # Nodes at level 2 with many children
```

### Find by hierarchy
```
p:@status=active        # Nodes whose parent is active
a:@type=work            # Nodes under a node with 'type=work' attribute
a:d:0                   # All nodes except root (have a root ancestor)
```

### Find by children/descendants
```
+child:@status=done           # Projects where all immediate children are done
-child:@urgent                # Projects with no urgent children
child*:@bug                   # Any node with a bug descendant
+child*:@status=done          # Projects where all descendants are done
-child*:@status=todo          # Projects with no incomplete descendants
```

### Find by ancestors (advanced)
```
+parent*:@type=project        # Nodes where all ancestors are projects
-parent*:@archived            # Nodes with no archived ancestors
parent*:@type=milestone       # Nodes under at least one milestone
```

### Find by siblings
```
sibling:@status=done          # Items with at least one completed sibling
+sibling:@status=done         # Items where all siblings are completed
-sibling:@status=todo         # Items with no incomplete siblings
@status=todo +sibling:@status=done  # Todo items where all siblings are done
sibling:@priority=high        # Items with high-priority siblings
-sibling:@archived            # Items with no archived siblings
```

### Find by attribute dates
```
@deadline>-7d           # Items with upcoming deadlines (next 7 days)
@deadline<-1d           # Items with overdue deadlines
@date>=-1d              # Daily notes from today
@date>-30d -@archived   # Items from last month, not archived
@start>=2025-11-01      # Items started on or after this date
@review<=2025-11-30     # Items with reviews due by end of month
```

## Examples

```
# Simple text search
urgent

# Combined with depth
TODO d:>1

# Attribute-based search
@type=day @date>-7d

# Complex boolean logic
(d:0 | d:1) +children:>0 -@archived

# Project nodes with unfinished children
@type=project -@status=done +children:>0

# Daily notes from last week
@type=day @date>-7d

# Recently modified tasks
task m:>-3d

# Nodes with specific parent
p:d:0 -children:0        # Top-level children that themselves have children

# Deadline filtering
@deadline>-1d            # Items with overdue deadlines
@deadline>-7d            # Items with deadlines in next week

# Items by start and due dates
@start>=-7d @due<-7d     # Recently started, due in next week

# Attribute date range
@date>=2025-10-01 @date<=2025-10-31  # Items from October

# Projects with all children done
+child:@status=done @type=project

# Items with no incomplete descendants
-child*:@status=todo

# Deep nodes under active projects
d:>3 +parent*:@type=project -parent*:@archived
```

## Output Formats

By default, search results are displayed as an interactive search node that can be expanded. For scripting and integration with other tools, you can specify alternative output formats using the `-ff` flag.

### Command Syntax

```
:search <query> [-ff format] [--fields field1,field2,...]
```

### Formats

#### Text Format (default)
```
:search task
```
Creates an interactive search node with expandable results (current behavior).

#### Fields Format (tab-separated)
```
:search @status=done -ff fields
:search @type=task -ff fields --fields id,text,created
```
Outputs results as tab-separated values, one result per line. Default fields are: `id`, `text`, `attributes`.

**Example output:**
```
id_abc	Buy groceries	@type=task @status=done @priority=high
id_def	Review proposal	@type=task @status=inprogress
```

This format is useful for piping to `grep`, `awk`, `sed`, or other Unix tools:
```
:search @status=done -ff fields | grep "priority=high"
:search @type=task -ff fields --fields text | sort
```

#### JSON Format
```
:search @type=project -ff json
:search task -ff json --fields id,text,attributes,created
```
Outputs results as a pretty-printed JSON array.

**Example output:**
```json
[
  {
    "id": "id_abc",
    "text": "Buy groceries",
    "attributes": {
      "type": "task",
      "status": "done",
      "priority": "high"
    },
    "created": "2024-11-08T14:30:00Z"
  }
]
```

#### JSONL Format (JSON Lines)
```
:search @type=day -ff jsonl
:search -@archived -ff jsonl --fields id,text,depth
```
Outputs results as JSON Lines format (one JSON object per line), useful for streaming large result sets.

**Example output:**
```
{"id":"id_abc","text":"Buy groceries","attributes":{"type":"task","status":"done"}}
{"id":"id_def","text":"Review proposal","attributes":{"type":"task","status":"inprogress"}}
```

### Available Fields

| Field | Description | Example |
|-------|-------------|---------|
| `id` | Item unique identifier | `item_abc123` |
| `text` | Item text content | `Buy groceries` |
| `attributes` | All attributes as object or @key=value | `{"type":"task"}` or `@type=task` |
| `attr:<name>` | Specific attribute value | `attr:status` → `done` |
| `created` | Creation timestamp (ISO 8601) | `2024-11-08T14:30:00Z` |
| `modified` | Modification timestamp (ISO 8601) | `2024-11-08T15:45:30Z` |
| `tags` | Tags list | `["urgent","work"]` or `urgent,work` |
| `depth` | Nesting level (0 = root) | `2` |
| `path` | Hierarchical path | `Projects > Work > Important` |
| `parent_id` | ID of parent item | `parent123` |

### Default Fields

- **Text format**: N/A (uses interactive display)
- **Fields format**: `id`, `text`, `attributes`
- **JSON format**: `id`, `text`, `attributes`, `created`, `modified`, `tags`, `depth`, `path`
- **JSONL format**: `id`, `text`, `attributes`, `created`, `modified`, `tags`, `depth`, `path`

### Clipboard Integration

When using non-text formats (`fields`, `json`, `jsonl`), results are automatically copied to the system clipboard for easy pasting into other applications or terminal commands.

### Examples

**Find all completed tasks and copy results for processing:**
```
:search @status=done -ff fields
```

**Export recent changes as JSON:**
```
:search m:>-7d -ff json --fields id,text,modified
```

**Get task IDs for scripting:**
```
:search @type=task -ff fields --fields id
```

**Stream large result sets:**
```
:search @type=day -ff jsonl
```

**Combine with Unix tools:**
```
# Find all high-priority items
:search @priority=high -ff fields | grep "@status=done"

# Count items by type
:search -ff fields --fields attr:type | sort | uniq -c

# Extract creation dates
:search c:>-30d -ff json --fields id,text,created
```

## Notes

- Text searches are **case-insensitive** substring matches
- Dates use relative format (relative to current time) or absolute YYYY-MM-DD format
- Comparisons: `>`, `>=`, `<`, `<=`, `=`, `!=`
- Invalid syntax will display an error message; the search will not execute
- Empty search matches all nodes
- When using output formats other than text, results are copied to clipboard if a clipboard tool is available (xclip, xsel, or pbcopy)
