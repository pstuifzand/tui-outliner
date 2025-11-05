# Types, Functions, and Ideas

This document tracks ideas and suggested enhancements for the type system and function capabilities in TUI Outliner.

## Key Strengths Identified

✅ **Well-architected with clear separation of concerns**
- Clean separation between types, templates, and application logic
- Modular design with dedicated packages (`internal/template/`)

✅ **Extensible design for adding new type kinds**
- Easy to add new type kinds through the type specification system
- Plugin-style architecture for validators and selectors

✅ **Comprehensive validation at multiple layers**
- Type definition validation
- Attribute value validation
- Template application validation

✅ **Good error messages and user feedback**
- Clear error messages for validation failures
- Helpful status messages during operations

✅ **Proper testing infrastructure**
- Comprehensive test suites for types and templates
- Good test coverage for edge cases

✅ **Seamless integration with attribute system**
- Types work naturally with existing attribute commands
- Tab-based type-aware selectors integrate smoothly

## Areas for Enhancement

### 1. Type Composition and Inheritance

- **Type inheritance**: Allow types to extend other types
  - Example: `project` type extends `task` type with additional fields
  - Syntax: `typedef add project extends:task ...`

- **Composite types**: Support for complex nested structures
  - Example: `contact` type with `address` sub-type
  - Better modeling of hierarchical data

- **Type aliases**: Create alternative names for existing types
  - Example: `typedef alias priority importance`

### 2. Performance Optimization

- **Caching type registry**: Avoid re-parsing type definitions
  - Cache parsed type specs in memory
  - Invalidate cache only when `__types__` item changes

- **Lazy loading**: Load types on-demand rather than all at once

- **Index type usage**: Track which items use which types for faster lookups

### 3. Type Usage Tracking and Migration Tools

- **Usage statistics**: Show how many items use each type
  - Command: `:typedef stats` - Show usage counts

- **Type migration**: Rename or update types across all items
  - Command: `:typedef rename oldname newname`
  - Command: `:typedef update typename` - Re-apply validation

- **Find items by type**: Quick search for all items of a given type
  - Already supported via search: `@type=task`

- **Orphaned type detection**: Find items with types that no longer exist

### 4. Import/Export Capabilities

- **Export type definitions**: Save types to separate file
  - Command: `:typedef export types.json`
  - Shareable type libraries across outlines

- **Import type definitions**: Load types from file
  - Command: `:typedef import types.json`
  - Merge with existing types or replace

- **Type bundles**: Predefined type sets for common use cases
  - GTD (Getting Things Done) bundle
  - Project management bundle
  - Personal knowledge management bundle

### 5. Advanced Features

- **Constraints and validation rules**: Beyond basic types
  - Cross-field validation: "If status=done, completed_date is required"
  - Custom validation functions

- **Computed attributes**: Derive values from other attributes
  - Example: `effort_remaining = effort_estimate - effort_spent`

- **Type-specific behaviors**: Custom actions for types
  - Example: Setting `@type=task` auto-adds to task list view

- **Relationships between items**: Formalize item references
  - `depends_on` relationship type
  - `blocks` relationship type
  - Bidirectional reference tracking

## Recommended Functions Based on Types

This section proposes comprehensive functions for all attribute types. Each function follows the same clear syntax: `:command @attribute [context] [filter]`

### Function Categories Overview

| Category | Description | Example Functions |
|----------|-------------|-------------------|
| **Aggregation** | Numeric calculations across items | `sum`, `avg`, `count`, `max`, `min` |
| **Date** | Date range queries and calculations | `earliest`, `latest`, `duein`, `overdue`, `age` |
| **Reference** | Relationship validation and analysis | `checkref`, `refcount`, `refby`, `circular` |
| **String** | Text queries and statistics | `contains`, `unique`, `strlen`, `wordcount` |
| **List** | Multi-value operations | `listhas`, `listsize`, `listunion` |
| **Boolean** | True/false counting | `counttrue`, `countfalse`, `percent` |

All functions support multiple execution contexts:
- **Global**: All items in outline (default)
- **Subtree**: Current item and descendants
- **Search**: Filtered items matching search expression
- **Tagged**: Manually marked items

### Aggregation Functions

Functions to compute values across multiple items based on their attributes.

#### Syntax Structure

**All functions follow this clear, parseable syntax:**

```
:command @attribute [context] [filter]
```

**Components:**
1. `command` - The function to execute (`sum`, `avg`, `count`, `max`, `min`)
2. `@attribute` - The attribute to operate on (required, starts with @)
3. `context` - Optional scope: `subtree`, `search`, `tagged` (omit for global)
4. `filter` - Search expression (only when using `search` context)

**Examples:**
```
:sum @priority                          # Global: sum all priority values
:avg @progress subtree                  # Subtree: average progress in current branch
:count @status=done search @type=task   # Search: count done tasks
:sum @effort tagged                     # Tagged: sum effort of marked items
```

#### Available Commands

- `:sum @<attribute>` - Sum numeric attribute
  - Example: `:sum @priority` - Total of all priority values
  - Example: `:sum @effort search @type=task` - Total effort for all tasks

- `:avg @<attribute>` - Average of numeric attribute
  - Example: `:avg @progress` - Average progress percentage
  - Example: `:avg @priority subtree` - Average priority in subtree

- `:count @<attribute>=<value>` - Count items with specific attribute value
  - Example: `:count @status=done` - How many done items
  - Example: `:count @type=bug search @status=open` - Count open bugs

- `:max @<attribute>` - Find maximum value
  - Example: `:max @priority` - Highest priority value
  - Example: `:max @effort search @type=task` - Highest task effort

- `:min @<attribute>` - Find minimum value
  - Example: `:min @progress` - Lowest progress value
  - Example: `:min @priority tagged` - Lowest priority among marked items

#### Context-Aware Functions

Functions should work in multiple contexts:

1. **Global scope**: All items in outline
   - `:sum @priority` - Sum across entire outline
   - `:avg @progress` - Average progress across all items

2. **Subtree scope**: Current item and all descendants
   - `:sum @priority subtree` - Sum for current branch
   - `:avg @effort subtree` - Average effort in subtree

3. **Search-based scope**: Filtered items
   - `:sum @priority search @type=task +@status=done` - Sum priority of done tasks
   - `:avg @progress search @type=bug +@status=open` - Average progress of open bugs

4. **Tagged nodes**: Items marked for batch operations
   - `m` to mark/tag items, then `:sum @priority tagged`
   - `:count @status=done tagged` - Count done items among tagged

#### Examples

```
# Sum all effort estimates for tasks in current project
:sum @effort subtree

# Count completed tasks across entire outline
:count @status=done

# Average priority of open bugs
:avg @priority search @type=bug +@status=open

# Sum priority of marked items
:sum @priority tagged
```

### Date Functions

Functions for working with date-type attributes.

#### Date Range and Comparison

- `:earliest @<date-attr>` - Find item with earliest date
  - Example: `:earliest @due_date` - Find earliest due date
  - Example: `:earliest @due_date search @type=task +@status=todo` - Earliest due among open tasks

- `:latest @<date-attr>` - Find item with latest date
  - Example: `:latest @due_date subtree` - Latest due date in subtree
  - Example: `:latest @created` - Most recently created item

- `:daterange @<date-attr>` - Show date range (earliest to latest)
  - Example: `:daterange @due_date subtree` - Due date range for project
  - Output: "2025-11-05 to 2025-12-15 (41 days)"

- `:duein @<date-attr> <days>` - Find items with date within N days
  - Example: `:duein @due_date 7` - Items due in next 7 days
  - Example: `:duein @due_date 7 search @status=todo` - Open items due soon

- `:overdue @<date-attr>` - Find items with past dates
  - Example: `:overdue @due_date` - All overdue items
  - Example: `:overdue @due_date search @type=task` - Overdue tasks

#### Date Grouping and Statistics

- `:groupby @<date-attr> <period>` - Group items by date period
  - Periods: `day`, `week`, `month`, `year`
  - Example: `:groupby @created week` - Items created per week
  - Output: Shows histogram/counts per period

- `:timeline @<date-attr>` - Show visual timeline of dates
  - Example: `:timeline @due_date subtree` - Project timeline
  - Output: ASCII timeline showing date distribution

- `:datecount @<date-attr>` - Count items per unique date
  - Example: `:datecount @due_date` - How many items per due date
  - Output: List of dates with counts

#### Date Calculations

- `:daysbetween @<date-attr1> @<date-attr2>` - Calculate days between two dates
  - Example: `:daysbetween @start_date @due_date` - Task duration
  - Calculates for each item with both attributes

- `:daysuntil @<date-attr>` - Days until date (from today)
  - Example: `:daysuntil @due_date subtree` - Days until each due date
  - Negative for past dates

- `:age @<date-attr>` - Days since date (from today)
  - Example: `:age @created` - Age of all items
  - Example: `:age @created search @type=bug +@status=open` - Age of open bugs

### Reference Functions

Functions for working with reference-type attributes and item relationships.

#### Reference Validation

- `:checkref @<ref-attr>` - Check for broken references
  - Example: `:checkref @depends_on` - Find items with invalid dependencies
  - Output: List of items with references to non-existent items

- `:resolveref @<ref-attr>` - Show what each reference points to
  - Example: `:resolveref @depends_on subtree` - Show dependency targets
  - Output: Item → Referenced Item mapping

- `:orphanref @<ref-attr>` - Find items not referenced by anything
  - Example: `:orphanref @parent_project` - Items with no parent project
  - Inverse of reference checking

#### Reference Counting and Analysis

- `:refcount @<ref-attr>` - Count how many items reference each target
  - Example: `:refcount @project` - How many items per project
  - Output: Target item with count of referencing items

- `:refby <item-id>` - Find all items that reference this item
  - Example: `:refby 12345` - What references item 12345
  - Search all reference-type attributes

- `:refto` - Show all items current item references
  - Example: `:refto` - Show all outbound references from current item
  - Lists all reference attributes and their targets

- `:circular @<ref-attr>` - Detect circular references
  - Example: `:circular @depends_on` - Find dependency cycles
  - Output: List of items forming circular reference chains

#### Reference Graph Analysis

- `:refchain @<ref-attr>` - Show reference chain from item
  - Example: `:refchain @depends_on` - Full dependency chain
  - Output: Current → Ref1 → Ref2 → ... (until end or cycle)

- `:refgraph @<ref-attr>` - Visualize reference graph
  - Example: `:refgraph @depends_on subtree` - Dependency graph for project
  - Output: ASCII graph showing relationships

- `:refdepth @<ref-attr>` - Calculate depth of reference chain
  - Example: `:refdepth @parent_project` - How deep in project hierarchy
  - Output: Number of levels in reference chain

### String Functions

Functions for working with string-type attributes.

#### String Queries

- `:contains @<str-attr> <text>` - Find items where attribute contains text
  - Example: `:contains @notes "urgent"` - Notes containing "urgent"
  - Case-insensitive substring search

- `:startswith @<str-attr> <text>` - Attribute starts with text
  - Example: `:startswith @title "FIX:"` - Titles starting with FIX:
  - Useful for prefixed/tagged attributes

- `:endswith @<str-attr> <text>` - Attribute ends with text
  - Example: `:endswith @filename ".md"` - Markdown files
  - Pattern matching for suffixes

- `:matches @<str-attr> <regex>` - Regex pattern matching
  - Example: `:matches @email ".*@example\\.com"` - Specific domain emails
  - Full regex support

#### String Statistics

- `:strlen @<str-attr>` - Calculate string length
  - Example: `:strlen @description` - Length of each description
  - Can use with `:max strlen @description` for longest

- `:wordcount @<str-attr>` - Count words in string attribute
  - Example: `:wordcount @notes subtree` - Word count per note
  - Useful for text content attributes

- `:unique @<str-attr>` - List unique values
  - Example: `:unique @status` - All unique status values
  - With counts: `:unique @status --count`

### List Functions

Functions for working with list-type attributes (comma-separated or multi-value).

#### List Queries

- `:listhas @<list-attr> <value>` - Items where list contains value
  - Example: `:listhas @tags "urgent"` - Items tagged with "urgent"
  - Works with comma-separated values

- `:listsize @<list-attr>` - Count elements in list
  - Example: `:listsize @tags` - Number of tags per item
  - Example: `:avg listsize @tags` - Average tags per item

- `:listunion @<list-attr>` - Union of all list values
  - Example: `:listunion @tags subtree` - All unique tags in subtree
  - Output: Complete set of all values across items

#### List Set Operations

- `:listintersect @<list-attr> <values>` - Items with all specified values
  - Example: `:listintersect @tags urgent,important` - Items tagged both
  - AND operation on list membership

- `:listexclude @<list-attr> <values>` - Items without any specified values
  - Example: `:listexclude @tags archived,deleted` - Items not archived/deleted
  - NOT operation on list membership

- `:listcount @<list-attr> <value>` - Count items containing value
  - Example: `:listcount @tags urgent` - How many urgent items
  - Similar to `:count` but for list membership

### Boolean Functions

Functions for working with boolean-type attributes.

- `:counttrue @<bool-attr>` - Count items where attribute is true
  - Example: `:counttrue @completed subtree` - Completed items in subtree
  - Example: `:counttrue @archived` - How many archived items

- `:countfalse @<bool-attr>` - Count items where attribute is false
  - Example: `:countfalse @reviewed` - Unreviewed items
  - Complement of counttrue

- `:percent @<bool-attr>` - Percentage of items where attribute is true
  - Example: `:percent @completed subtree` - Completion percentage
  - Output: "65% (13/20 items)"

### Type-Based Search Enhancements

Already well-supported through the search system! Examples:

- `@type=task +@status=todo` - Find all todo tasks
- `@priority>5 +d:>2` - High priority items deeply nested
- `@type=project +children:0` - Projects with no sub-items
- `@due_date>-7d` - Items due in next 7 days (date comparison)

### Validation Reports

Commands to check and fix type-related issues:

- `:validate` - Check all items against their type definitions
  - Shows items with validation errors
  - Lists attributes that don't match type specs

- `:validate --fix` - Auto-fix validation errors where possible
  - Remove invalid attributes
  - Set default values for required attributes

- `:validate @<attribute>` - Validate specific attribute across all items
  - Example: `:validate @date` - Check all date attributes

- `:validate type:<typename>` - Validate all items of a specific type
  - Example: `:validate type:task` - Check all task items

### Attribute Templates

Quick commands to apply common attribute sets:

- `:template <typename>` - Apply type's default attributes to current item
  - Example: `:template task` - Add status, priority, due_date attributes

- `:template list` - Show available templates

- `:template add <name>` - Create new template from current item's attributes
  - Saves current attributes as a reusable template

## Better Input Fields for Different Types

### Design Philosophy

Type-specific input helpers should be **automatic and always active** when editing attributes. Instead of a generic text field that requires Tab to switch modes, each type gets its own specialized input helper that appears as a **small popup** near the input field (except the calendar widget for dates, which is larger).

### Current State
- Generic text input for all attribute types
- Tab key switches to type-specific selector (enum, number, date)
- Type hints shown below input field

### Proposed Design

Input helpers should:
- **Appear automatically** based on attribute type
- Be **non-intrusive** small popups (except calendar)
- Show **inline suggestions** without blocking the tree view
- Allow **both keyboard and text entry** for flexibility
- Display **usage statistics** to guide selection

### Type-Specific Input Helpers

#### 1. Enum Types - Dropdown Selector (Small Popup)

**Current**: Tab to selector, then arrow navigation

**Proposed**: Automatic dropdown popup below input field
- **Small popup** showing all enum values as a list
- Auto-complete dropdown as you type
- First-letter quick jump (already implemented in selector)
- Show usage count for each value: `todo (15) | in-progress (8) | done (42)`
- Arrow keys to navigate, Enter to select
- Can type value directly or select from list
- Popup dimensions: ~5 lines tall, width matches longest value

#### 2. Number Types - Slider Popup (Small Popup)

**Current**: Slider with up/down navigation

**Proposed**: Automatic slider popup with visual feedback
- **Small popup** below input field showing slider
- Show histogram of currently used values
- Suggest common values: "Most common: 5 (used 12 times)"
- Visual bar showing current value position in range
- Arrow keys ↑/↓ to adjust, Home/End for min/max
- Can type number directly or use slider
- Popup dimensions: ~3-4 lines tall

#### 3. Date Types - Calendar Widget (Large Popup)

**Current**: Date picker with keyboard navigation

**Proposed**: **Use existing calendar widget** (opened with `gc`)
- **Calendar widget** appears as modal overlay (larger than other popups)
- Calendar navigation: h/j/k/l for day/week movement
- J/K for month, H/L for year navigation
- Quick shortcuts: `t` for today
- Can also type date directly in YYYY-MM-DD format
- Calendar shows existing dates with dot indicators
- Relative shortcuts: `+7` for 7 days from now, `-3d` for 3 days ago
- Natural language: "tomorrow", "next week", "end of month"

#### 4. String Types - Suggestions Popup (Small Popup)

**Current**: Free text input

**Proposed**: Automatic suggestions dropdown
- **Small popup** below input showing previously used values
- Auto-complete dropdown as you type
- Show frequency: "urgent (5 times), normal (12 times)"
- Arrow keys to navigate suggestions, Enter to accept
- Can type new value or select from suggestions
- Popup dimensions: ~5-8 lines tall, shows most common values first

#### 5. Reference Types - Node Picker Popup (Medium Popup)

**Current**: Manual text entry

**Proposed**: Interactive node picker with live search
- **Medium popup** showing filtered tree view
- Use the filter from type definition to find available nodes
- Example: Type defined as `reference|@type=project`
  - Show live search results matching `@type=project`
  - Navigate with j/k, select with Enter
  - Show item path/hierarchy for context
- Filter items in real-time as you type
- Visual mini-tree showing available reference targets
- Keyboard navigation (j/k)
- Multi-select for list of references (Space to toggle)
- Popup dimensions: ~10-15 lines tall, shows tree structure

#### 6. List Types - Multi-Value Entry (Inline with Small Popup)

**Current**: Not well supported

**Proposed**: Inline tags with suggestions
- Text input field shows current values as visual tags
- **Small popup** below showing suggestions for next value
- Add multiple values with comma or Enter
- Each value validated against list element type
- Visual tags showing added values (colored boxes)
- Arrow left/right to navigate tags, Backspace to remove
- Tab through values, Delete to remove selected tag
- Popup shows suggestions for additional values

### Implementation Ideas

#### Automatic Helper Detection
- When entering attribute edit mode, detect type from registry
- Automatically show appropriate helper widget
- No Tab key needed - helper is always visible
- Can still type directly if preferred

#### Popup Positioning
- Small popups appear directly below input field
- Calendar widget appears centered as modal overlay
- Popups auto-position to stay within screen bounds
- Semi-transparent background for non-modal popups

#### Visual Design
- Small popups: 3-8 lines tall, minimal border
- Medium popups: 10-15 lines tall for tree views
- Large popups: Calendar widget (full modal with background dim)
- Consistent styling: border, highlight for selection
- Usage statistics in gray/dim color

#### Keyboard Shortcuts
- Arrow keys: Navigate within helper widget
- Enter: Accept selection
- Esc: Close helper, return to text input
- Can type directly without dismissing helper
- Helper updates in real-time based on typed text

#### Value Browser Command
- `:values @<attribute>` - Browse all used values for an attribute
  - Shows frequency histogram
  - Filter and search values
  - Jump to items using each value
  - Useful for exploring attribute usage patterns

## Function Context Execution

All functions should support multiple execution contexts with clear, parseable syntax.

### Syntax Structure

```
:command @attribute [context] [filter]
```

Components:
1. **Command** - Function name (sum, avg, count, etc.)
2. **Attribute** - The attribute to operate on (always starts with @)
3. **Context** - Optional scope keyword (subtree, search, tagged)
4. **Filter** - Optional search expression (only for search context)

### 1. Global Context (Default)
Execute function across entire outline.

```
:sum @priority              # All items
:count @status=done         # Count items with status=done across entire outline
:avg @progress              # Average across all items
:max @priority              # Find highest priority value
```

### 2. Subtree Context
Execute function on current item and all descendants.

```
:sum @priority subtree      # Current branch only
:count @status=done subtree # Count in current subtree
:avg @effort subtree        # Average effort in subtree
:validate subtree           # Validate current branch
```

### 3. Search Context
Execute function on filtered results. Filter comes after the context keyword.

```
:sum @priority search @type=task +@status=done      # Sum priority of done tasks
:count @status=open search @type=bug                # Count open bugs
:avg @progress search @type=bug +@status=open       # Average progress of open bugs
:validate search @type=project                      # Validate all projects
```

### 4. Tagged Nodes Context
Execute function on manually marked items.

```
# Mark items with 'm' keybinding, then:
:sum @priority tagged       # Sum priority of marked items
:count @status=done tagged  # Count done items among marked
:avg @progress tagged       # Average progress of marked items
:validate tagged            # Validate marked items
:template task tagged       # Apply template to all marked items
```

### Implementation Considerations

- **Parse order**: Command → Attribute → Context → Filter
  - Easy to parse: split on spaces, check for @ prefix
  - Context keyword is one of: `subtree`, `search`, `tagged`
  - Everything after `search` keyword is the filter expression

- **Context detection**:
  - If no context keyword: global scope
  - If second token is context keyword: use that context
  - If second token is not context keyword and not @: error

- **Visual feedback**: Show which items are in scope
  - Highlight items being processed
  - Show progress bar for long operations

- **Result display**: Context-aware result formatting
  - "Sum of @priority (15 items in subtree): 127"
  - "Average @progress (8 items matching '@type=bug +@status=open'): 65%"

### Example Workflows

#### Project Progress Tracking
```
# Navigate to project item
j j j

# Sum effort for all items in current project subtree
:sum @effort subtree

# Count completed tasks in current project
:count @status=done subtree

# Calculate average progress in project
:avg @progress subtree

# Sum effort specifically for tasks (using search)
:sum @effort search @type=task +d:>2
```

#### Tag and Analyze
```
# Mark interesting items with 'm'
j m
j j m
k k k m

# Analyze marked items
:sum @priority tagged       # Total priority
:count @status=done tagged  # How many done
:avg @progress tagged       # Average progress
```

#### Validation Sweep
```
# Validate specific item types
:validate search @type=task

# Fix validation errors for tasks
:validate --fix search @type=task

# Validate entire outline
:validate
```

#### Complex Queries
```
# Sum priority of high-priority open bugs
:sum @priority search @type=bug +@priority>5 +@status=open

# Average effort of unstarted tasks in subtree
:avg @effort subtree search @type=task +@status=todo

# Count items with missing required attributes
:count @type=task search -@status
```

## Priority Ranking

### High Priority

**Essential Functions and Input Helpers:**
1. Reference type input with filter-based selection
   - Interactive node picker using search filters
   - Most impactful for usability
2. Better enum/string input with automatic popups and usage statistics
   - Small automatic popups for all types
   - Calendar widget integration for dates
3. Context-aware function execution (subtree, search, tagged)
   - Foundation for all function types
4. Basic aggregation functions (sum, avg, count, max, min)
   - Most commonly needed numeric functions

**Essential Date Functions:**
5. Date range queries (earliest, latest, duein, overdue)
   - Critical for task/project management
6. Date calculations (daysuntil, age, daysbetween)
   - Useful for tracking and planning

### Medium Priority

**Reference and Validation Functions:**
7. Reference validation functions (checkref, resolveref, circular)
   - Maintain data integrity
8. Reference counting and analysis (refcount, refby, refto)
   - Understand item relationships
9. Validation reports and auto-fix
   - Type-system health checks

**String and List Functions:**
10. String query functions (contains, startswith, unique)
    - Common text operations
11. List operations (listhas, listsize, listunion)
    - Working with multi-value attributes
12. Boolean functions (counttrue, percent)
    - Progress tracking and completion metrics

**System Improvements:**
13. Type usage tracking and statistics
14. Performance optimization (caching)
15. Import/export type definitions

### Lower Priority

**Advanced Date Functions:**
16. Date grouping and timeline (groupby, timeline, datecount)
    - Visualization and analysis

**Advanced Reference Functions:**
17. Reference graph analysis (refchain, refgraph, refdepth)
    - Complex dependency visualization

**Advanced String Functions:**
18. Regex matching and word counting
    - Specialized text analysis

**Type System Enhancements:**
19. Type composition and inheritance
20. Computed attributes and constraints
21. Type-specific behaviors
22. Attribute templates

## Function Implementation Order

**Phase 1: Foundation (High Priority 1-6)**
- Input helpers with automatic popups
- Context-aware execution framework
- Basic aggregation and date functions

**Phase 2: Data Integrity (Medium Priority 7-9)**
- Reference validation
- Type validation and auto-fix

**Phase 3: Analysis Tools (Medium Priority 10-15)**
- String/list/boolean functions
- Usage tracking and statistics

**Phase 4: Advanced Features (Lower Priority 16-22)**
- Advanced visualization
- Type system enhancements
- Specialized analysis functions

## Related Documentation

- [Templates Guide](templates.md) - Template system overview
- [Search Syntax](search-syntax.md) - Search filter reference
- [Attribute Value Selection](attribute-value-selection.md) - Type-aware selectors
- [Configuration](configuration.md) - Settings and configuration
