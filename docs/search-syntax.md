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
@date>-7d       # Attribute 'date' is more recent than 7 days ago
@date<-30d      # Attribute 'date' is older than 30 days ago
@date>=-1d      # Attribute 'date' is within the last day
@date=2025-11-01 # Attribute 'date' is exactly this date
@date>=2025-10-01 # Attribute 'date' is on or after this date
@date<=2025-11-30 # Attribute 'date' is on or before this date
```

**Examples with date attributes:**

```
@deadline>-7d          # Items with deadlines in the next 7 days
@completed<-7d         # Items completed more than a week ago
@start>=-1d m:>-3d     # Recently started items, modified in last 3 days
```

### Creation Date Filter: `c:`

Match nodes created within a time window.

**Syntax:** `c:DATE` | `c:>=DATE` | `c:>DATE` | `c:<=DATE` | `c:<DATE`

Date formats:
- Relative: `-1d` (1 day ago), `-7d` (7 days ago), `-30d` (30 days ago), `-1w`, `-4w`, `-1m`, `-6m`, `-1y`
- Absolute: `2025-11-01` (YYYY-MM-DD format)

```
c:>-7d           # Created in the last 7 days
c:<-30d          # Created more than 30 days ago
c:>=2025-11-01   # Created on or after this date
```

### Modified Date Filter: `m:`

Match nodes modified within a time window. Same syntax as creation date filter.

```
m:>-7d           # Modified in the last 7 days
m:<-1d           # Modified more than 1 day ago
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

Match nodes whose parent matches criteria. Allows filtering by parent's properties.

**Syntax:** `p:FILTER`

```
p:d:0            # Nodes whose parent is at root level
p:@type=project  # Nodes whose parent has 'type=project' attribute
p:d:>=2          # Nodes whose parent is at depth 2 or deeper
```

### Ancestor Filter: `a:`

Match nodes that have an ancestor matching criteria. Use `a:` followed by another filter.

**Syntax:** `a:FILTER`

```
a:@type=project    # Nodes anywhere under a node with 'type=project' attribute
a:d:0              # Any node with root as ancestor (all non-root)
a:@status=active   # Nodes under a node with 'status=active' attribute
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
```

## Notes

- Text searches are **case-insensitive** substring matches
- Dates use relative format (relative to current time) or absolute YYYY-MM-DD format
- Comparisons: `>`, `>=`, `<`, `<=`, `=`, `!=`
- Invalid syntax will display an error message; the search will not execute
- Empty search matches all nodes
