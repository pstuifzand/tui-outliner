# Todo Items

TUI Outliner includes a flexible todo item system for tracking task status. The system is built on top of the attributes system and integrates seamlessly with the existing outline structure.

## Quick Start

### Creating Todo Items

Type `[] ` at the beginning of an item's text to automatically create a todo:

1. Press `o` to create a new item
2. Type `[] My task name`
3. Press `Escape` to finish editing

The item will be created with:
- Text: `My task name` (the `[] ` prefix is removed)
- Attribute `type=todo`
- Attribute `status=todo` (the first status in your todostatuses config)

### Rotating Status

Press `x` on any item to rotate through todo statuses:

```
todo → doing → done → todo → ...
```

The status cycles through the values in the `todostatuses` configuration (default: `todo,doing,done`).

**Behavior:**
- If an item has no status, pressing `x` sets `type=todo` and initializes status to the first value
- If an item already has a status, `x` advances to the next status in the sequence
- Pressing `x` on the last status wraps around to the first status

### Automatic Parent Status Updates

When you rotate a child task's status with `x`, the parent's status automatically updates based on the status of its children (bottom-up workflow):

**Status Propagation Rules:**
- **Parent must have `type=todo`** to receive automatic updates
- If **all children are todo** → parent is todo
- If **all children are done** → parent is done
- If **mix of statuses** → parent is doing
- **Progress attributes** are automatically stored:
  - `progress_count`: Shows "3/5" (completed/total todo children)
  - `progress_pct`: Shows "60%" (percentage of children done)

**Example Workflow:**
1. Create a parent task with `[] Project` → status=todo
2. Create children: `[] Task 1`, `[] Task 2`, `[] Task 3` → all status=todo
3. Parent "Project" stays todo (all children are todo)
4. Press `x` on Task 1 → status changes to "doing"
5. Parent "Project" automatically changes to "doing" (mix of statuses)
6. Press `x` twice on Task 1 → status changes to "done"
7. Press `x` on Task 2 → status changes to "doing"
8. Parent "Project" stays "doing" (still a mix)
9. Press `x` twice on Tasks 2 and 3 → both become "done"
10. Parent "Project" automatically changes to "done" (all children done)

**Cascading Updates:**
- Updates propagate up through ancestors that also have `type=todo`
- Stops at ancestors without `type=todo` (respects type boundaries)
- Non-todo parent items are never automatically updated

**View Progress:**

Use `:set visattr progress_count,progress_pct` to display progress inline:
```
[] Project                 [progress_count:3/5, progress_pct:60%]
  [] Task 1              [status:done]
  [] Task 2              [status:doing]
  [] Task 3              [status:todo]
```

Or use the visual progress bar (enabled by default):
```
[] Project  ■■■■■  (One ■ block per child, colored by status)
  [] Task 1 ■ (Green = done)
  [] Task 2 ■ (Orange = doing)
  [] Task 3 ■ (Gray = todo)
```

### Progress Bar Visualization

Parent items with todo children automatically display a colored progress bar:
- **■ Gray block**: Child is todo
- **■ Orange block**: Child is doing/in-progress
- **■ Green block**: Child is done
- One block per todo child (dynamic length)
- Inline display after item text

**Configuration:**
```
:set showprogress true   # Enable progress bar (default)
:set showprogress false  # Disable progress bar
```

**Example:**
```
[] Project A  ■■■■■■■■■■  (All children todo - all gray)
[] Project B  ■■■■■■■     (Mix of statuses - gray, orange, green)
[] Project C  ■■■■■       (All children done - all green)
```

## Configuration

### todostatuses

Configure the allowed todo statuses with the `:set` command:

```
:set todostatuses "todo,doing,done"
```

The default value is `todo,doing,done`. Use comma-separated values without spaces between statuses. You can customize this to any values you prefer:

```
:set todostatuses "backlog,in-progress,review,done"
:set todostatuses "not-started,started,completed"
```

## Display Options

### Show Status Inline

Use the `:set visattr status` command to display the status attribute inline with item text:

```
:set visattr status
```

This will show items like:
```
[] My first task                 [status:todo]
[] My second task                [status:doing]
[] My completed task             [status:done]
```

You can also show multiple attributes:

```
:set visattr status,type,priority
```

### View Item Attributes

Press `av` (attribute view) to see all attributes of the current item, including status.

## Advanced Usage

### Searching for Todos

Use the search syntax to find todos by status:

```
/type=todo
/@status=doing
/@status!=done
```

Or search for todos with specific properties:

```
/type=todo d:0        # Find leaf node todos
/type=todo @status=done d:>0  # Find completed todos with children
```

### Setting Status Manually

You can set the status attribute directly using `:attr` commands:

```
:attr add status done
:attr add status doing
```

Or use the attribute editor by pressing `ac` (change attribute).

## Example

See `examples/todo_demo.json` for a complete example demonstrating:
- Todo items with different statuses
- How to create todos with the `[] ` prefix
- Different status values
- How to navigate and rotate statuses

Run:
```bash
./tuo examples/todo_demo.json
```

Then:
1. Navigate with `j`/`k`
2. Press `x` on items to rotate their status
3. Use `:set visattr status` to see statuses
4. Use `/` to search with `@status=doing`
