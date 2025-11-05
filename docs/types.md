# Template System

The template system in tuo allows you to define reusable item templates with default attributes and child structure. Templates are powerful for creating consistent item types across your outline.

## Overview

Templates enable you to:
- Define standard item types (tasks, projects, bugs, etc.)
- Automatically apply default attributes when creating items of a type
- Include predefined child structures (subtasks, checklists, sections)
- Validate attribute values against type definitions
- Streamline creation of complex, multi-level items

## Components

### 1. Type Definitions

Type definitions specify what attributes are valid and what values they can have. They are stored in the outline as a special `__types__` item.

**Type Specification Format:**

```
enum|value1|value2|value3    → One of the listed values
number|min-max               → Number in range (e.g., 1-5)
date                         → Date in YYYY-MM-DD format
list|itemtype                → List of items
string                       → Any string value
reference|@filter            → Reference with filter expression
```

**Type Definition Commands:**

- `:typedef list` - Show all defined types
- `:typedef add <key> <spec>` - Add a type definition
- `:typedef remove <key>` - Remove a type definition

**Examples:**

```
:typedef add status enum|todo|in-progress|done
:typedef add priority number|1-5
:typedef add deadline date
:typedef remove priority
```


## Type Definition Reference

### Enum Type

Restricts attribute to one of several values:

```
:typedef add status enum|todo|in-progress|done
:typedef add priority enum|low|medium|high|critical
```

**Validation:** Value must exactly match one of the listed values.

### Number Type

Restricts attribute to a numeric range:

```
:typedef add priority number|1-5
:typedef add effort_points number|0-13
```

**Validation:** Value must be an integer within the specified range (inclusive).

### Date Type

Requires attribute to be in YYYY-MM-DD format:

```
:typedef add deadline date
:typedef add due_date date
```

**Validation:** Value must match YYYY-MM-DD format.

### String Type

Allows any string value:

```
:typedef add description string
:typedef add notes string
```

**Validation:** Always valid (no restrictions).

### List Type

Indicates a list of items (advanced):

```
:typedef add components list|string
```

**Validation:** Value cannot be empty.

### Reference Type

Indicates a reference to another item (advanced):

```
:typedef add parent reference|@type=project
```

