# Calendar Widget

The calendar widget is an interactive date picker that helps you select dates for tasks and notes. It can be used in two modes:

1. **Search Mode** - Search for and navigate to items with a specific date
2. **Attribute Mode** - Set a date attribute on the currently selected item

## Opening the Calendar

### Via Keybinding
Press `gc` (g + c) to open the calendar in search mode.

### Via Command
- `:calendar` - Open calendar in search mode
- `:calendar attr <attribute-name>` - Open calendar in attribute mode to set a specific date attribute

Example:
```
:calendar attr deadline
```

## Navigation

Use these keybindings to navigate the calendar:

### Day Navigation
- `h` / `l` - Previous / next day
- Arrow Left / Right - Previous / next day

### Week Navigation
- `j` / `k` - Next / previous week
- Arrow Down / Up - Next / previous week

### Month Navigation
- `J` / `K` - Previous / next month (Shift+j / Shift+k)

### Year Navigation
- `H` / `L` - Previous / next year (Shift+h / Shift+l)

### Quick Navigation
- `t` - Jump to today

## Configuration

### Week Start Day
You can configure which day the week starts on using the `:set weekstart` command.

- `:set weekstart 0` - Start on Sunday (default)
- `:set weekstart 1` - Start on Monday
- `:set weekstart 2` - Start on Tuesday
- `:set weekstart 3` - Start on Wednesday
- `:set weekstart 4` - Start on Thursday
- `:set weekstart 5` - Start on Friday
- `:set weekstart 6` - Start on Saturday

This setting affects both the weekday headers and how dates are positioned in the calendar grid.

Example:
```
:set weekstart 1
```

This is useful if your region traditionally uses Monday as the first day of the week.

## Selection

- **Enter** - Smart date selection
  - In search mode:
    - If an item with this date exists: Navigate to it
    - If no item exists: Create a new item with this date
  - In attribute mode: Sets the attribute on the current item to this date

- **Escape** - Close calendar without selecting

## Mouse Support

You can also use your mouse to interact with the calendar:

- **Click on a date** - Select that date
- **Click navigation arrows**:
  - `<<` - Previous year
  - `<` - Previous month
  - `>` - Next month
  - `>>` - Next year

## Visual Indicators

The calendar displays dates with visual indicators:

- **Currently Selected Date** - Highlighted with the tree selection color
- **Today** - Shown in green and bold
- **Dates with Items** - Displayed with a dot (•) to indicate there are existing items with that date
  - Items are detected by:
    - Matching the `date` attribute (YYYY-MM-DD format)
    - Item text that is a date (YYYY-MM-DD format)

## Examples

### Example 1: Set a deadline for today's selected item
1. Select an item in the tree
2. Press `:calendar attr deadline` (or `:calendar attr deadline`)
3. Navigate to the desired date
4. Press Enter to set the deadline

### Example 2: Find or create items from a specific date
1. Press `gc` to open the calendar
2. Navigate to the desired date
3. Press Enter:
   - If an item with that date exists, you'll navigate to it
   - If not, a new dated item will be created automatically

### Example 3: Set multiple date attributes
You can set different date attributes on an item:
```
:calendar attr date
```
Then later:
```
:calendar attr deadline
```
Then later:
```
:calendar attr review-date
```

## Integration with Search

After using the calendar in search mode, you can use the regular search commands to navigate between matching items:
- `n` - Next match
- `N` - Previous match

This makes it easy to find all items from a specific date.

## Date Formats

Dates are always stored and displayed in ISO 8601 format: **YYYY-MM-DD**

Examples:
- 2025-11-02 (November 2, 2025)
- 2025-12-25 (December 25, 2025)
- 2026-01-01 (January 1, 2026)

## Tips

- Items with dates that fall outside the currently displayed month/year are still searchable via the calendar
- The calendar persists its selection as you navigate, making it easy to compare dates
- If you accidentally close the calendar, just reopen it - your selection will be preserved
- Combine the calendar with the attribute system for powerful date-based organization
