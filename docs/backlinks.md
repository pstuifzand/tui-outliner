# Backlinks Feature

## Overview

The backlinks feature allows you to find all items that link to a specific item. This is useful for understanding the connections between items in your outline and discovering related content.

## Usage

### Keybinding: `gb`

Press `gb` while an item is selected to show all items that link to the current item.

- **g**: Go to... (prefix key)
- **b**: Show backlinks

### How it works

1. Select an item in your outline
2. Press `gb`
3. The search widget will activate with a `ref:<itemid>` query
4. All items containing links to the selected item will be displayed
5. Navigate through results using `n` (next) and `N` (previous)

### Example

Given the following outline:
```
- Target Item [ID: item_123]
- Item A - Links to [[item_123]]
- Item B - Also links: [[item_123|target]]
- Item C - No link to target
```

If you select "Target Item" and press `gb`, the search will find:
- Item A
- Item B

Item C will not be found because it doesn't contain a link to the target.

## Search Syntax: `ref:<itemid>`

You can also manually use the `ref:` search filter to find backlinks.

### Basic Syntax

```
ref:<item_id>
```

Or use the short alias:

```
r:<item_id>
```

### Examples

**Find all items linking to a specific item:**
```
ref:item_20250110120000_test1
```

**Combine with other filters using OR (`|`):**
```
ref:item_123 | @tag=important
```

This finds items that either:
- Link to item_123, OR
- Have the tag "important"

**Combine with other filters using AND (space or `+`):**
```
ref:item_123 @status=done
```

This finds items that:
- Link to item_123, AND
- Have status=done

**Negate to find items NOT linking to target:**
```
-ref:item_123
```

This finds all items that do NOT link to item_123.

## Link Format

Links use the wiki-style format:

- `[[item_id]]` - Basic link (displays item text)
- `[[item_id|custom text]]` - Link with custom display text

The `ref:` filter matches both formats.

## Related Features

- **Follow Link (`gf`)**: Navigate to the target of a link in the current item
- **Search (`/`)**: General search with advanced filter syntax
- **Go to Referenced (`gr`)**: Navigate to the original item when viewing a virtual child

## Technical Details

The backlinks search:
1. Parses the `ref:<itemid>` query
2. Searches all items for `[[<itemid>` pattern
3. Returns all matching items
4. Automatically navigates to the first result

The search respects:
- Hoisted views (only searches within the hoisted subtree)
- All link formats (basic and custom text)
- Case-sensitive item IDs

## Tips

- Use `n` and `N` to navigate between backlink results
- Press `Escape` to exit search mode
- Combine `ref:` with other filters for powerful queries
- Use the test file `examples/backlinks_test.json` to try out the feature
