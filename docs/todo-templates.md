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

### 2. Template Structure

A template consists of two levels:

```
Template: Task (@type=template, @applies_to=task, @name=default-task)
  └─ Task (@type=task, @status=todo, @priority=1)
       └─ Subtask 1
       └─ Subtask 2
```

**Template Container** (outer level):
- Has `@type=template` attribute
- Has `@applies_to=<typename>` attribute
- Has `@name=<template-name>` for identification
- Can have `@description` attribute
- Not directly applied to items

**Template Node** (first child):
- The actual structure that gets copied
- Has attributes that become defaults
- Can have children (deep copied when applied)
- Can have nested subtemplates

### 3. Applying Templates

Templates can be applied in two ways:

#### Auto-Apply (When Setting @type)

When you set the `@type` attribute on an item, all templates with matching `@applies_to` are automatically applied:

```
1. Create new item: "My Task"
2. Run: :attr add type task
3. Template automatically applies:
   - Status set to "todo"
   - Priority set to "1"
   - Child "Subtask" item added (if template has children)
```

#### Manual Apply (Explicit Command)

Apply a specific template by name to current item:

```
:apply-template default-task
:apply-template bug-template
```

## Attribute Merging Rules

When a template is applied:

1. **Existing attributes are preserved** - Template only fills in missing attributes
2. **Same-value attributes are skipped** - If target and template have same value, no change
3. **Multiple templates**: Last wins - Later templates override earlier ones
4. **Children are appended** - Template children added below existing children

## Validation

All attributes are validated against type definitions when applying templates:

```
Example: @status must be one of: todo, in-progress, done
If validation fails, the template is not applied and an error is shown
```

Validation happens:
- On manual `:apply-template` command
- On auto-apply when setting `@type` attribute
- Blocks application if any attribute is invalid

## Examples

### Example 1: Simple Task Template

**Step 1: Define types**
```
:typedef add status enum|todo|in-progress|done
:typedef add priority number|1-5
```

**Step 2: Create template**
```
Template: Task (@type=template, @applies_to=task, @name=default-task)
  └─ Task (@type=task, @status=todo, @priority=1)
       └─ Subtask
```

**Step 3: Apply template**
```
# Auto-apply:
:attr add type task

# Or manual:
:apply-template default-task
```

### Example 2: Project Template with Sections

**Template Definition:**
```
Template: Project (@type=template, @applies_to=project, @name=default-project)
  └─ Project (@type=project, @status=todo)
       ├─ Checklist
       │  └─ Task item
       └─ Notes
```

**Result When Applied:**
```
My Project (@type=project, @status=todo)
  ├─ Checklist
  │  └─ Task item
  └─ Notes
```

### Example 3: Bug Report Template

**Template:**
```
Template: Bug Report (@type=template, @applies_to=bug, @name=bug)
  └─ Bug: [Title] (@priority=1, @status=todo)
       ├─ Steps to reproduce
       ├─ Expected behavior
       └─ Actual behavior
```

**Apply with:**
```
:apply-template bug
```

## Template Best Practices

1. **Name templates clearly** - Use `@name` attribute for easy reference
2. **Add descriptions** - Use `@description` for documenting what template does
3. **Define types first** - Set up type definitions before creating templates
4. **Test validation** - Ensure all default attribute values pass validation
5. **Organize templates** - Keep all templates at top-level or in a dedicated section
6. **Use meaningful nesting** - Template children should represent useful structure
7. **Avoid circular refs** - Templates can contain templates, but avoid infinite nesting

## Common Commands

| Command | Purpose |
|---------|---------|
| `:typedef list` | Show all type definitions |
| `:typedef add status enum\|todo\|done` | Define status type |
| `:attr add type task` | Auto-apply task template |
| `:apply-template default-task` | Manually apply template |
| `:typedef remove status` | Remove type definition |

## File Format

Templates are stored as regular items in the JSON outline. Example:

```json
{
  "text": "Template: Task",
  "metadata": {
    "attributes": {
      "type": "template",
      "applies_to": "task",
      "name": "default-task"
    }
  },
  "children": [
    {
      "text": "Task",
      "metadata": {
        "attributes": {
          "type": "task",
          "status": "todo",
          "priority": "1"
        }
      },
      "children": [
        {
          "text": "Subtask"
        }
      ]
    }
  ]
}
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

## Tips & Tricks

### Create Multiple Variants

You can create multiple templates for the same type:

```
Template: Task (Simple) (@applies_to=task, @name=simple-task)
Template: Task (Detailed) (@applies_to=task, @name=detailed-task)
```

Then choose which to apply manually.

### Template Hierarchy

Templates can contain other templates:

```
Template: Sprint (@type=template, @applies_to=sprint)
  └─ Sprint
      └─ Task Template (@type=template, @applies_to=task)
```

Children that are templates won't auto-apply during parent application.

### Validation Error Recovery

If template apply fails due to validation:
1. Check error message
2. Verify type definitions
3. Correct attribute values in template
4. Try again

## See Also

- `examples/template_demo.json` - Complete template example
- `:attr` command - Attribute management
- `:set` command - Configuration

