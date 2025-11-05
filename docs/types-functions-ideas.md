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

### Aggregation Functions

Functions to compute values across multiple items based on their attributes.

#### Global Commands
- `:sum @<attribute>` - Sum numeric attribute across all items
  - Example: `:sum @priority` - Total of all priority values

- `:avg @<attribute>` - Average of numeric attribute
  - Example: `:avg @progress` - Average progress percentage

- `:count @<attribute>=<value>` - Count items with specific attribute value
  - Example: `:count @status=done` - How many done items

- `:max @<attribute>` - Find maximum value
  - Example: `:max @priority` - Highest priority value

- `:min @<attribute>` - Find minimum value
  - Example: `:min @progress` - Lowest progress value

#### Context-Aware Functions

Functions should work in multiple contexts:

1. **Global scope**: All items in outline
   - `:sum @priority` - Sum across entire outline

2. **Subtree scope**: Current item and all descendants
   - `:sum subtree @priority` - Sum for current branch

3. **Search-based scope**: Filtered items
   - `:sum search @type=task +@status=done @priority` - Sum priority of done tasks

4. **Tagged nodes**: Items marked for batch operations
   - `m` to mark/tag items, then `:sum tagged @priority`

#### Examples

```
# Sum all effort estimates for tasks in current project
:sum subtree @effort

# Count completed tasks across entire outline
:count @type=task +@status=done

# Average priority of open bugs
:sum search @type=bug +@status=open @priority
```

### Type-Based Search Enhancements

Already well-supported through the search system! Examples:

- `@type=task +@status=todo` - Find all todo tasks
- `@priority>5 +d:>2` - High priority items deeply nested
- `@type=project +children:0` - Projects with no sub-items

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

### Current State
- Generic text input for all attribute types
- Tab key switches to type-specific selector (enum, number, date)
- Type hints shown below input field

### Enhancements Needed

#### 1. Enum Types - Prespecified Values

**Current**: Tab to selector, then arrow navigation
**Enhancement**:
- Auto-complete dropdown showing all values as you type
- First-letter quick jump (already implemented in selector)
- Show usage count for each value: `todo (15) | in-progress (8) | done (42)`

#### 2. Number Types - Context-Aware Ranges

**Current**: Slider with up/down navigation
**Enhancement**:
- Show histogram of currently used values
- Suggest common values: "Most common: 5 (used 12 times)"
- Visual bar showing current value position in range

#### 3. Date Types - Relative Shortcuts

**Current**: Date picker with keyboard navigation
**Enhancement**:
- Quick shortcuts: `+7` for 7 days from now, `-3d` for 3 days ago
- Natural language: "tomorrow", "next week", "end of month"
- Show existing dates: "3 items use 2025-11-10"

#### 4. String Types - Used Values

**Current**: Free text input
**Enhancement**:
- Suggest previously used values
- Auto-complete from existing values
- Show frequency: "urgent (5 times), normal (12 times)"

#### 5. Reference Types - Filter-Based Selection

**Current**: Manual text entry
**Enhancement**:
- Use the filter from type definition to find available nodes
- Example: Type defined as `reference|@type=project`
  - Show live search results matching `@type=project`
  - Navigate with j/k, select with Enter
  - Show item path/hierarchy for context

- Interactive node picker:
  - Filter items in real-time as you type
  - Visual tree showing available reference targets
  - Multi-select for list of references

#### 6. List Types - Multi-Value Entry

**Current**: Not well supported
**Enhancement**:
- Add multiple values with comma separation
- Each value validated against list element type
- Visual tags showing added values
- Easy removal: Tab through values, Delete to remove

### Implementation Ideas

#### Smart Input Widget
- Unified input widget that adapts based on type
- Shows relevant suggestions inline
- Keyboard shortcuts consistent across types
- Visual indicators for type (icon or color)

#### Value Browser
- Command: `:values @<attribute>` - Browse all used values
  - Shows frequency histogram
  - Filter and search values
  - Jump to items using each value

#### Reference Picker
- Interactive widget for `reference` type attributes
- Uses search filter from type definition
- Real-time filtering as you type
- Shows item context (parent path)
- Keyboard navigation (j/k)
- Multi-select support for list of references

## Function Context Execution

All functions should support multiple execution contexts:

### 1. Global Context (Default)
Execute function across entire outline.

```
:sum @priority              # All items
:count @status=done         # Count across entire outline
:avg @progress              # Average across all items
```

### 2. Subtree Context
Execute function on current item and all descendants.

```
:sum subtree @priority      # Current branch only
:count subtree @status=done # Count in current subtree
:validate subtree           # Validate current branch
```

### 3. Search Context
Execute function on filtered results.

```
:sum search @type=task @priority          # Sum priority of all tasks
:count search @status=done @type=bug      # Count done bugs
:validate search @type=project            # Validate all projects
```

### 4. Tagged Nodes Context
Execute function on manually marked items.

```
# Mark items with 'm' keybinding, then:
:sum tagged @priority       # Sum priority of marked items
:count tagged               # Count marked items
:validate tagged            # Validate marked items
:template tagged task       # Apply template to all marked items
```

### Implementation Considerations

- **Context keyword**: First argument specifies context
  - `subtree`, `search`, `tagged`, or omit for global

- **Context chaining**: Combine contexts
  - `:sum subtree search @type=task @priority` - Tasks in subtree

- **Visual feedback**: Show which items are in scope
  - Highlight items being processed
  - Show progress bar for long operations

- **Result display**: Context-aware result formatting
  - "Sum of @priority (15 items in subtree): 127"
  - "Average @progress (8 items matching search): 65%"

### Example Workflows

#### Project Progress Tracking
```
# Navigate to project item
j j j

# Sum effort for all tasks in project
:sum subtree @type=task @effort

# Count completed tasks
:count subtree @status=done

# Calculate completion percentage
:avg subtree @type=task @progress
```

#### Tag and Analyze
```
# Mark interesting items with 'm'
j m
j j m
k k k m

# Analyze marked items
:sum tagged @priority       # Total priority
:count tagged               # How many marked
:values tagged @status      # Status distribution
```

#### Validation Sweep
```
# Validate specific item types
:validate search @type=task

# Fix validation errors
:validate --fix search @type=task

# Check entire outline
:validate
```

## Priority Ranking

### High Priority
1. Reference type input with filter-based selection
2. Better enum/string input with usage statistics
3. Context-aware function execution (subtree, search, tagged)
4. Aggregation functions (sum, avg, count)

### Medium Priority
5. Validation reports and auto-fix
6. Type usage tracking and statistics
7. Performance optimization (caching)
8. Import/export type definitions

### Lower Priority
9. Type composition and inheritance
10. Computed attributes and constraints
11. Type-specific behaviors
12. Attribute templates

## Related Documentation

- [Templates Guide](templates.md) - Template system overview
- [Search Syntax](search-syntax.md) - Search filter reference
- [Attribute Value Selection](attribute-value-selection.md) - Type-aware selectors
- [Configuration](configuration.md) - Settings and configuration
