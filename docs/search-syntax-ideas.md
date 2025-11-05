# Search Syntax Enhancement Ideas

This document lists potential missing features and enhancements to the TUI Outliner search syntax.

## High-Priority Missing Features

### 1. ✅ Descendant/Subtree Filter (IMPLEMENTED)

**Status:** Implemented as `child:` and `child*:` filters

Find nodes that have children or descendants matching criteria. This is the logical opposite of ancestor filtering.

**Syntax:**
- `child:FILTER` - Direct children only (one level down)
- `child*:FILTER` - All descendants recursively (entire subtree)

**Quantifiers:**
```
-child:FILTER                    # No children match FILTER
child:FILTER                     # Some children match FILTER
+child:FILTER                    # All children match FILTER

-child*:FILTER                   # No descendants match FILTER (at any depth)
child*:FILTER                    # Some descendant matches FILTER (at any depth)
+child*:FILTER                   # All descendants match FILTER (entire subtree)
```

**Real examples:**
```
child:@status=done               # Items with at least one done child
+child:@status=done              # Items where all children are done
-child:@archived                 # Items with no archived children
child*:@urgent                   # Items with urgent descendants anywhere in subtree
+child*:@status=done             # Items where ALL descendants are done
-child*:@archived                # Items with no archived descendants at any level
@type=project child*:@deadline>-7d  # Projects with upcoming deadlines in subtree
```

**Complete Hierarchy Navigation System:**

**Parent (singular, one level up):**
```
-parent:FILTER                   # Parent doesn't match FILTER
parent:FILTER                    # Parent matches FILTER
+parent:FILTER                   # Redundant (parent is always singular)
```

**Parent* (recursive, all ancestors to root):**
```
-parent*:FILTER                  # No ancestors match FILTER
parent*:FILTER                   # At least one ancestor matches FILTER
+parent*:FILTER                  # All ancestors (entire chain to root) match FILTER
```

**Child (single level, multiple children):**
```
-child:FILTER                    # No children match FILTER
child:FILTER                     # Some children match FILTER
+child:FILTER                    # All children match FILTER
```

**Child* (recursive, all descendants):**
```
-child*:FILTER                   # No descendants match FILTER (at any depth)
child*:FILTER                    # Some descendant matches FILTER (at any depth)
+child*:FILTER                   # All descendants match FILTER (entire subtree)
```

**Multi-level examples:**
```
+parent*:@type=project           # Every ancestor up to root is a project
parent*:@status=active           # At least one ancestor is active
+child*:@status=done             # All descendants recursively are done
child*:@urgent                   # Has some urgent item somewhere in subtree
-child*:@archived                # Nothing in subtree is archived
```

**Use cases:**
- Find parent nodes whose children match criteria
- Find projects containing specific items in subtree
- Validate completion (all descendants done)
- Find containers with specific content patterns

**Documentation:** See docs/search-syntax.md for full details

---

### 2. ✅ Sibling Filter (IMPLEMENTED)

**Status:** Implemented in commit e61683c

Find nodes sharing the same parent as nodes matching criteria.

**Syntax:** `sibling:FILTER` or `s:FILTER`

**Quantifiers:**
```
-sibling:FILTER                  # No siblings match FILTER
sibling:FILTER                   # Some siblings match FILTER
+sibling:FILTER                  # All siblings match FILTER
```

**Real examples:**
```
sibling:@type=project            # Find items with at least one project sibling
+sibling:@status=done            # Find nodes where all siblings are done
-sibling:@archived               # Find nodes with no archived siblings
sibling:@urgent                  # Find nodes that have urgent siblings
+sibling:@priority=high          # Find nodes where all siblings are high priority
@status=todo +sibling:@status=done # Todo items where all siblings are done
```

**Implementation details:**
- Both `sibling:` and `s:` shorthand syntax supported
- Quantifiers work with all filter types (attributes, depth, text, etc.)
- Empty set semantics: items with no siblings return false for Some/All, true for None
- Root items have no siblings (no parent)
- Comprehensive test coverage in internal/search/parser_test.go

**Documentation:** See docs/search-syntax.md for full details

---

### 3. Regex Text Search

Enable pattern matching for text without needing to list all variations.

**Syntax:** `/pattern/` or `regex:pattern`

**Examples:**

```
/^TODO/                      # Items starting with TODO
/\d{4}-\d{2}-\d{2}/          # Items containing dates in text
/@[a-z]+/                    # Items mentioning usernames
/^(?!SKIP)/                  # Items NOT starting with SKIP
```

**Use cases:**
- Match complex text patterns
- Extract structured data from text
- Negative lookahead patterns

---

### 4. Text Search Qualifiers

Enhance text matching with additional options beyond substring matching.

**Syntax:** Text filters with modifier flags

**Examples:**
```
task@word                    # Whole-word match: "task" but not "multitask"
task@case                    # Case-sensitive match: "Task" != "task"
todo^                        # Prefix match: "todo" but not "mytodo"
^done                        # Suffix match: "done" but not "ongoing"
word@accent-insensitive      # Ignore accents: "café" matches "cafe"
```

**Use cases:**
- Precision matching when substring matching is too broad
- Case-sensitive search for structured data
- Word boundary matching for variable names

---

### 5. Text Length Filter

Find nodes based on their text content length.

**Syntax:** `len:N` | `len:>N` | `len:<N` | `len:>=N` | `len:<=N`

**Examples:**
```
len:0                        # Empty items (no text)
len:<50                      # Short items (quick notes)
len:>200                     # Long items (detailed descriptions)
len:>=100 len:<=500          # Medium-length items
children:0 len:0             # Empty leaf nodes (placeholders)
```

**Use cases:**
- Find overly verbose items
- Find empty placeholders to fill
- Manage content organization by size
- Find incomplete entries (empty)

---

## Medium-Priority Missing Features

### 6. Recursive Depth Limits on Ancestor/Descendant

Currently `a:` searches unlimited depth up the tree. Allow limiting recursion depth.

**Syntax:** `a1:FILTER`, `a2:FILTER`, `a<3:FILTER`

**Examples:**
```
a1:@type=project             # Only direct parent is project type
a2:@status=active            # Parent or grandparent is active
a<3:d:0                       # Ancestor is root (direct or one level deep)
desc1:@done                   # Only direct children are done
```

**Use cases:**
- "Find items whose immediate parent matches criteria"
- "Limit ancestor search to specific depth"
- "Find items at specific hierarchical distance"

---

### 7. Sort/Order Results

Currently results appear in tree traversal order only. Allow custom sorting.

**Syntax:** `+sort:FIELD`, `+order:FIELD`, `+reverse`

**Examples:**
```
task +sort:depth             # Results ordered by depth (shallowest first)
task +sort:modified          # Results ordered by modification date
task +sort:created -reverse   # Results reverse ordered by creation
d:>0 +sort:children:desc     # Sorted by number of children descending
```

**Use cases:**
- Find most recently modified matching items first
- Order results by relevance/depth
- Reverse chronological order for recent changes

---

### 8. Virtual/Real Item Filter

Filter for virtual reference items (created by search nodes) versus real items.

**Syntax:** `virtual:true`, `virtual:false`, `@virtual`

**Examples:**
```
-virtual:true                # Only real items (exclude references)
virtual:true                 # Only virtual items (search node results)
@type=day virtual:false      # Real daily notes (not search references)
```

**Use cases:**
- Hide search node references from results
- Find only original items (not duplicates in search nodes)
- Filter search-generated virtual items

---

### 9. Empty Item Detection

Find nodes with no text or only whitespace content.

**Syntax:** `empty:true`, `empty:false`, `text:empty`

**Examples:**
```
empty:true                   # Placeholder nodes with no content
children:>0 empty:true       # Parent nodes with no text (just containers)
-empty:true                  # Non-empty items only
```

**Use cases:**
- Find incomplete/placeholder entries
- Clean up outline by finding empty nodes
- Find container nodes vs content nodes

---

### 10. Time-of-Day Filtering

Currently only date-based filtering (`c:`, `m:`). Enable time-of-day filtering for timestamps.

**Syntax:** `c:time>HH:MM`, `m:time<18:00`

**Examples:**
```
c:time>09:00                 # Created after 9 AM
m:time<18:00                 # Modified before 6 PM
c:2025-11-03 time:09:30      # Created on specific date at specific time
```

**Use cases:**
- Find entries created/modified during work hours
- Capture time-sensitive entries
- Track when changes were made (more precise)

---

### 11. Descendant Text Search

Find nodes whose children or descendants contain specific text (different from ancestor search).

**Syntax:** `child-text:keyword`, `desc-text:keyword`

**Examples:**
```
child-text:urgent            # Nodes whose children mention "urgent"
desc-text:TODO               # Nodes whose descendants contain "TODO"
@type=project desc-text:DONE # Projects with "DONE" items somewhere below
```

**Use cases:**
- Find parent containers by their content
- Search parent nodes by what children contain
- Navigate to parent through child content

---

### 12. Compound Attribute Matching

For attributes storing comma-separated lists or tag collections.

**Syntax:** `@attr+=value`, `@attr-=value`, `@attr~=pattern`

**Examples:**
```
@tags+=urgent                # Tags attribute contains "urgent"
@tags+=bug,feature           # Tags contain either "bug" or "feature"
@categories-=archived        # Categories does NOT include "archived"
@flags~=^REQUIRED            # Flags contains value starting with REQUIRED
```

**Use cases:**
- Match within multi-value attributes
- Check membership in comma-separated lists
- Tag-based filtering with flexible matching

---

## Lower-Priority/Niche Features

### 13. Search Result Position Indicator (UI)

Show "X of Y results" during search navigation.

**Format:** Display count indicator in search widget

**Use cases:**
- Navigate through large result sets
- Understand search result scope
- Know when you're at last result

---

## Prioritization Summary

**Recommend implementing in this order:**

1. **Regex Text Search** - Unlocks advanced pattern matching, high user demand
2. ~~**Descendant Filter**~~ - ✅ **IMPLEMENTED** (`child:` and `child*:` filters)
3. **Text Qualifiers** - Improves precision with minimal parser changes
4. **Depth Limits on Ancestor** - More control over existing ancestor functionality
5. **Text Length Filter** - Practical for content management and cleanup
6. **Empty Item Detection** - Quick win for finding placeholders
7. ~~**Sibling Filter**~~ - ✅ **IMPLEMENTED** (commit e61683c)
8. **Sort Results** - Quality of life improvement
9. **Virtual Item Filter** - Niche but useful for search node users

---

## Implementation Notes

- Most features require **parser enhancements** in `internal/search/parser.go`
- New filter types need **FilterExpr implementations** in `internal/search/expr.go`
- Regex support requires adding regex compilation and matching logic
- Sort/order features may benefit from **storing results in a list** rather than streaming
- Depth-limited ancestor should reuse existing depth calculation with limit parameter
- Text qualifiers could be flags appended to text filters (e.g., `text@word`)
