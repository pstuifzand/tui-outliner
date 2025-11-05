# Parent/Child Filters Demo

This example outline demonstrates the new parent/child filter functionality with quantifiers.

## Structure

The outline contains:
- **Project Alpha** - A completed project with all tasks and subtasks marked as "done"
- **Project Beta** - An active project with mixed status (some done, some in progress, some todo)
- **Archive** - An archived section containing old completed projects

## Example Queries

### Finding Completed Projects

**Query:** `+child:@status=done @type=project`
- **Meaning:** Projects where ALL immediate children have status=done
- **Matches:** Project Alpha (all phases are done)
- **Doesn't match:** Project Beta (has in_progress child)

### Finding Projects With All Work Complete

**Query:** `+child*:@status=done @type=project`
- **Meaning:** Projects where ALL descendants (at any depth) have status=done
- **Matches:** Project Alpha (all tasks and subtasks are done)
- **Doesn't match:** Project Beta (has subtasks that are todo/in_progress)

### Finding Items With No Incomplete Descendants

**Query:** `-child*:@status=todo`
- **Meaning:** Items with NO descendants marked as "todo"
- **Matches:** Project Alpha, Archive section, and all items without todo descendants
- **Doesn't match:** Project Beta (has "Gather feedback" subtask as todo)

### Finding Tasks Under Active Projects

**Query:** `parent:@status=active`
- **Meaning:** Items whose immediate parent has status=active
- **Matches:** "Design phase", "Implementation phase", "Testing phase" (under Project Alpha), "Research", "Prototype" (under Project Beta)
- **Doesn't match:** Items under Archive

### Finding Items With No Archived Ancestors

**Query:** `-parent*:@archived=true`
- **Meaning:** Items where NONE of the ancestors have archived=true
- **Matches:** Project Alpha, Project Beta, and all their descendants
- **Doesn't match:** Archive section and everything under it

### Finding Deep Nodes Under Active Projects

**Query:** `d:>2 +parent*:@type=project -parent*:@archived=true`
- **Meaning:** Items deeper than level 2, where ALL ancestors are projects (not archived)
- **Matches:** Subtasks under Project Alpha and Project Beta that are at depth 3+
- **Doesn't match:** Archived subtasks, shallow items

### Finding Projects With Some Incomplete Work

**Query:** `@type=project child*:@status=todo`
- **Meaning:** Projects that have AT LEAST ONE descendant with status=todo
- **Matches:** Project Beta (has "Gather feedback" as todo)
- **Doesn't match:** Project Alpha (all done), Archive projects

### Finding Phases With All Subtasks Done

**Query:** `+child:@status=done parent:@type=project`
- **Meaning:** Items under a project where ALL immediate children are done
- **Matches:** "Design phase", "Implementation phase", "Testing phase" under Project Alpha
- **Doesn't match:** "Prototype" under Project Beta (has todo child)

### Finding Items With No Children

**Query:** `children:0 -parent*:@archived=true`
- **Meaning:** Leaf nodes (no children) that are not under an archived ancestor
- **Matches:** All subtasks under Project Alpha and Project Beta
- **Doesn't match:** Parent items, archived subtasks

### Finding Non-Archived Active Work

**Query:** `@status=in_progress -parent*:@archived=true`
- **Meaning:** Items marked as in_progress that have no archived ancestors
- **Matches:** "Prototype" task under Project Beta
- **Doesn't match:** Completed or archived items

## Quantifier Comparison

### `child:` vs `+child:` vs `-child:`

| Query | Meaning | Project Alpha | Project Beta |
|-------|---------|---------------|--------------|
| `child:@status=done` | Some children done | ✓ | ✓ |
| `+child:@status=done` | All children done | ✓ | ✗ |
| `-child:@status=done` | No children done | ✗ | ✗ |

### `child*:` vs `+child*:` vs `-child*:`

| Query | Meaning | Project Alpha | Project Beta |
|-------|---------|---------------|--------------|
| `child*:@status=done` | Some descendants done | ✓ | ✓ |
| `+child*:@status=done` | All descendants done | ✓ | ✗ |
| `-child*:@status=done` | No descendants done | ✗ | ✗ |

### `parent*:` vs `+parent*:` vs `-parent*:`

| Query (from subtask level) | Meaning | Under Project Alpha | Under Archive |
|-----------------------------|---------|---------------------|---------------|
| `parent*:@type=project` | Some ancestor is project | ✓ | ✓ |
| `+parent*:@type=project` | All ancestors are projects | ✗ | ✗ |
| `-parent*:@archived=true` | No ancestors archived | ✓ | ✗ |

## Testing the Example

1. Open the example file:
   ```bash
   ./tuo examples/parent_child_filters_demo.json
   ```

2. Press `/` to open search

3. Try the queries listed above

4. Observe which items are highlighted/matched

5. Use `:set visattr status,type` to see attributes inline

## Key Takeaways

1. **Quantifiers control matching:**
   - Default (no prefix): "some" - at least one matches
   - `+` prefix: "all" - all must match
   - `-` prefix: "none" - none must match

2. **`*` suffix means recursive:**
   - `child:` = immediate children only
   - `child*:` = all descendants
   - `parent*:` = all ancestors

3. **Combine with other filters:**
   - Mix with depth filters: `d:>2 +child:@status=done`
   - Mix with attributes: `@type=project +child*:@status=done`
   - Mix with boolean operators: `(child:todo | child:in_progress) @type=project`

4. **Empty set semantics:**
   - Items with no children: `+child:` always false, `-child:` always true
   - Root items: `+parent*:` always true (vacuously), `-parent*:` always true (vacuously)
