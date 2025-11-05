# Template System

The template system in tuo allows you to define reusable item templates with default attributes and child structure. Templates are powerful for creating consistent item types across your outline.

## Overview

Templates enable you to:
- Define standard item types (tasks, projects, bugs, etc.)
- Validate attribute values against type definitions

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

