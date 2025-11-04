# Variable References in Query Filters

## Overview

Enable query nodes to reference attributes from parent nodes and template contexts, allowing dynamic filter composition without manual updates. This feature bridges hierarchical structure with query logic, making templates reusable and date/context-aware.

## Problem Statement

Currently, search nodes and filters require hardcoded values. When organizing by month or project, users must manually create separate queries for each time period or context, duplicating logic across the outliner.

**Example of current limitation:**

```
October 2025
├─ Daily Notes
├─ Search: Incomplete Todos
│  @query=@type=day @date>=2025-10-01 @date<=2025-10-31 +child:(@type=todo -@status=done)

November 2025
├─ Daily Notes
├─ Search: Incomplete Todos
│  @query=@type=day @date>=2025-11-01 @date<=2025-11-30 +child:(@type=todo -@status=done)
```

The date ranges must be manually maintained separately for each month.

## Solution

### Syntax Options

Three candidate syntaxes for variable reference, ranked by preference:

#### Option 1: Dollar Prefix (Recommended)

Simple, familiar, minimal syntax overhead.

```
$attribute_name
```

**Examples:**

```
@query=@type=day @date>=$start @date<=$end +child:(@type=todo -@status=done)
@query=child:(@priority>$min_priority)
@query=@tags:$required_tags
```

#### Option 2: Dollar with Braces

Explicit boundaries for complex identifiers, clearer parsing.

```
${attribute_name}
```

**Examples:**

```
@query=@type=day @date>=${start} @date<=${end}
@query=child:(@priority>${min_priority})
```

#### Option 3: Dot Notation

Explicit parent reference, allows nested access (future-proofed).

```
@parent.attribute_name
@parent.parent.attribute_name
```

**Examples:**

```
@query=@type=day @date>=@parent.start @date<=@parent.end
@query=@parent.parent.context_filter
```

**Recommendation:** Use **Option 1** (`$attribute`) as primary syntax. Allow **Option 2** (`${attribute}`) as explicit alternative for complex cases. Defer **Option 3** for future nested hierarchy traversal.

---

## Variable Resolution Scope

### Resolution Order

1. **Local node attributes** (highest priority)
   ```
   Current node @start=2025-01-01 $start → 2025-01-01
   ```

2. **Parent node attributes**
   ```
   Parent @month=October $month → October
   ```

3. **Ancestor chain** (depth-first, closest first)
   ```
   Great-grandparent @project=Alpha $project → Alpha
   ```

4. **Undefined variables** (error state)
   ```
   $undefined → Error: undefined variable
   OR fallback to literal string "$undefined"
   ```

### Example Resolution

```
Project Alpha                          (@project=Alpha)
├─ October 2025                       (@start=2025-10-01, @end=2025-10-31)
│  ├─ Search: Incomplete Todos
│  │  @query=@type=day @date>=$start @date<=$end +child:(@type=todo -@status=done)
│  │  # $start resolves to 2025-10-01 (from parent)
│  │  # $end resolves to 2025-10-31 (from parent)
│  │
│  └─ Priority Filter
│     @query=$project_filter
│     @project_filter=@priority>2
│     # $project_filter resolves to @priority>2 (from self)
```

---

## Use Cases

### 1. Time-Based Templates

**Template structure:**

```
October 2025
  @type=month
  @start=2025-10-01
  @end=2025-10-31
  @search_query=@type=day @date>=$start @date<=$end +child:(@type=todo -@status=done)
  ├─ Daily Notes
  ├─ Search: Incomplete Todos
  │  @type=search @query=$search_query
  └─ Search: Completed Items
     @type=search @query=@type=day @date>=$start @date<=$end +child:@status=done
```

**Benefit:** Duplicate the month node; dates auto-adjust via variables.

### 2. Project Context Filters

**Template structure:**

```
Project: Web Dashboard
  @project=web-dashboard
  @min_priority=2
  @owner=alice
  ├─ Tasks
  ├─ Search: My High Priority Items
  │  @query=@assignee=$owner +child:(@priority>=$min_priority)
  └─ Search: Blockers
     @query=@status=blocked @project=$project
```

**Benefit:** Context filters propagate to all search nodes automatically.

### 3. Hierarchical Metadata Propagation

```
Company
  @type=org
  @timezone=UTC
  ├─ Engineering
    @team=eng
    ├─ Backend Team
      @subteam=backend
      ├─ Sprint Planning
        @query=@team=$team @subteam=$subteam @status=planned
```

### 4. Reusable Named Filters

```
My Filters
  @type=meta
  @urgent_filter=@priority>3 @status!=done
  @this_week=@date>=-7d @date<0d
  @my_items=@assignee=peter
  ├─ Search: Urgent This Week
  │  @query=$my_items $this_week $urgent_filter
  └─ Search: Review Queue
     @query=$my_items @type=review
```

---

## Implementation Details

### Variable Substitution

**Timing:** Variables substitute at filter execution time, not at creation time.

```
# User creates query at 2025-10-15
@query=@type=day @date>=$start
# When executed with parent @start=2025-10-01:
# → @type=day @date>=2025-10-01 (evaluated dynamically)
```

**Benefit:** If parent attributes change, queries auto-update.

### Escaping and Literals

Allow users to reference literal `$` when needed:

```
@query=@price=\$100              # Literal $100
@query=amount=$amount            # Variable reference
```

### Type Coercion

Variables should maintain their types:

```
@count=5                         # Integer
@query=child:($count)            # Interpreted as count

@date=2025-10-01                 # Date/string
@query=@date>=$date              # Interpreted as date comparison
```

### Error Handling

**Option A: Strict** (recommended)

```
Undefined variable $undefined → Parse error, query fails
Message: "Variable $undefined not found in scope"
```

**Option B: Permissive**

```
Undefined variable $undefined → Treated as literal string "$undefined"
Warning: "Variable $undefined not found; treating as literal"
```

**Recommendation:** Strict mode with clear error messages. Avoids silent logic failures.

---

## Query Composition Examples

### Example 1: Monthly Review Dashboard

```
October 2025
  @start=2025-10-01
  @end=2025-10-31
  @type=month
  
  ├─ Search: All Tasks This Month
  │  @query=@type=day @date>=$start @date<=$end +child:@type=todo
  │
  ├─ Search: Incomplete
  │  @query=@type=day @date>=$start @date<=$end +child:(@type=todo -@status=done)
  │
  ├─ Search: High Priority Blockers
  │  @query=@type=day @date>=$start @date<=$end child:(@type=todo @status=blocked @priority>2)
  │
  └─ Search: Completed This Month
     @query=@type=day @date>=$start @date<=$end +child:@status=done
```

**Copy behavior:** Duplicate `October 2025` node, rename to `November 2025`, update `@start` and `@end`. All queries auto-update.

### Example 2: Project Hierarchy with Filters

```
Alpha Project
  @project=alpha
  @status=active
  @lead=alice
  
  ├─ Backend
    @team=backend
    @language=go
    
    ├─ Search: Open Issues
    │  @query=@project=$project @team=$team @status=open
    │
    └─ Search: Code Review Queue
       @query=@project=$project @team=$team @type=review -@status=approved
  
  ├─ Frontend
    @team=frontend
    @language=typescript
    
    └─ Search: Open Issues
       @query=@project=$project @team=$team @status=open
```

**Benefit:** Change `@project=alpha` once; all child searches update.

### Example 3: Conditional Logic with Variables

```
Dashboard
  @current_quarter=Q4
  @min_urgency=2
  @deadline_threshold=-14d
  
  ├─ Search: Urgent & Overdue
  │  @query=@priority>=$min_urgency @deadline<$deadline_threshold
  │
  └─ Search: Quarter-End Items
     @query=@quarter=$current_quarter @status!=done
```

---

## Migration & Backward Compatibility

- **Existing queries** without variables continue to work unchanged
- **New syntax** is opt-in; hardcoded values remain valid
- **Parser** must handle both simultaneously
- **No breaking changes** to current filter syntax

---

## Future Extensions

### 1. Multi-Level Traversal

```
$parent.attribute
$grandparent.attribute
$ancestor(type=project).attribute
```

### 2. Computed Variables

```
@computed_start=$start + 1d       # Date arithmetic
@count=count($child:@type=todo)   # Dynamic counts
```

### 3. Scoped Variables

```
@query=for $item in child:@type=todo | $item.priority>2
```

### 4. Variable Templates

```
@templates.monthly_review=@type=day @date>=$start @date<=$end
@query=$templates.monthly_review
```

---

## Design Rationale

**Why variables at the attribute level?**
- Attributes already exist as the primary metadata layer
- Queries already reference attributes (`@date`, `@type`, etc.)
- Natural extension of existing query syntax
- Minimal cognitive overhead

**Why dynamic substitution (not static)?**
- Supports iterative refinement (change parent, queries auto-update)
- Reduces duplication across hierarchy
- More maintainable for long-term templates

**Why strict error handling?**
- Silent failures in filters are dangerous (queries match nothing unexpectedly)
- Explicit errors guide user to fix variable references
- Better debugging experience

---

## Summary

Variable references enable:

✅ **Template reusability** — Copy structures, variables auto-adapt
✅ **Hierarchical context** — Queries inherit parent metadata  
✅ **Reduced duplication** — Single source of truth for values
✅ **Dynamic updates** — Change parent, all child queries refresh
✅ **Cleaner composition** — Complex filters become readable

**Recommended syntax:** `$attribute_name` with optional `${attribute_name}` for clarity.
