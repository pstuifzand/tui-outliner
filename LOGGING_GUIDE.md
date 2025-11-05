# Template System Logging Guide

## Overview

Comprehensive debug logging has been added to the template system to help diagnose issues with `:typedef` commands. All logs are written to `/tmp/tuo-template-debug.log`.

Type definitions are stored as a hidden field `type_definitions` in the outline JSON file, not as visible nodes in the outline.

## Logging Location

```
/tmp/tuo-template-debug.log
```

View logs in real-time:
```bash
tail -f /tmp/tuo-template-debug.log
```

View all logs:
```bash
cat /tmp/tuo-template-debug.log
```

## What Gets Logged

### Command Entry Point

Every template command logs:
- The exact command and all arguments received
- Whether file is readonly
- Number of types loaded from outline

**Example:**
```
[TEMPLATE] handleTypedefCommand called with parts: [typedef add status enum|todo|done]
[TEMPLATE] Creating new type registry and loading from outline
[TEMPLATE] Loaded 0 types from outline
[TEMPLATE] Processing subcommand: add
[TEMPLATE] Executing add command: key=status, spec=enum|todo|done
```

### Type Addition

When adding a type, logs show:
- The key and spec being added
- Success or failure of parsing
- Success or failure of registry addition
- The outline structure after save
- The exact attributes stored

**Example:**
```
[TEMPLATE] handleTypedefAdd called: key=status, spec=enum|todo|done
[TEMPLATE] Attempting to add type to registry
[TEMPLATE] Type added to registry successfully
[TEMPLATE] Saving types back to outline
[TYPES] SaveToOutline called with 1 types
[TYPES]   Saved type: status = enum|todo|done
[TYPES] SaveToOutline complete, outline.TypeDefinitions now has 1 types
[TEMPLATE] Types saved to outline successfully
[TEMPLATE] App marked as dirty, status set
```

### Type Listing

When listing types, logs show:
- Number of types found
- Name and kind of each type
- The formatted status message

**Example:**
```
[TEMPLATE] handleTypedefList called
[TEMPLATE] Found 2 types
[TEMPLATE]   Type: status (kind: enum)
[TEMPLATE]   Type: priority (kind: number)
[TEMPLATE] Status message: Types: status: enum(todo,done) | priority: number(1-5)
```

### Type Removal

When removing a type, logs show:
- The key being removed
- Success of removal
- Outline structure after save

## Testing Workflow

### 1. Clear Old Logs

```bash
rm -f /tmp/tuo-template-debug.log
```

### 2. Start the App

```bash
cd /home/peter/work/tui-outliner
./tuo /tmp/test_outline.json
```

### 3. Execute Test Commands

Inside the app, run:
```
:typedef list
:typedef add status enum|todo|in-progress|done
:typedef list
:typedef add priority number|1-5
:typedef list
```

### 4. Monitor Logs

In another terminal:
```bash
tail -f /tmp/tuo-template-debug.log
```

### 5. Exit App

Press `q` and examine the logs.

## Understanding Log Messages

### Success Flow for `:typedef add status enum|todo|done`

1. **Command received:**
   ```
   [TEMPLATE] handleTypedefCommand called with parts: [typedef add status enum|todo|done]
   ```

2. **Registry created and loaded:**
   ```
   [TEMPLATE] Creating new type registry and loading from outline
   [TEMPLATE] Loaded 0 types from outline
   ```

3. **Subcommand identified:**
   ```
   [TEMPLATE] Processing subcommand: add
   [TEMPLATE] Executing add command: key=status, spec=enum|todo|done
   ```

4. **Type added to registry:**
   ```
   [TEMPLATE] handleTypedefAdd called: key=status, spec=enum|todo|done
   [TEMPLATE] Attempting to add type to registry
   [TEMPLATE] Type added to registry successfully
   ```

5. **Saved to outline:**
   ```
   [TEMPLATE] Saving types back to outline
   [TEMPLATE] Types saved to outline successfully
   ```

6. **Verified:**
   ```
   [TEMPLATE] Checking outline structure after save
   [TEMPLATE] Outline has 1 items
   [TEMPLATE]   Item 0: text=__types__
   [TEMPLATE]     Attributes: map[status:enum|todo|done]
   ```

7. **Complete:**
   ```
   [TEMPLATE] App marked as dirty, status set
   ```

### Error Flow: Invalid Type Spec

If you run `:typedef add status enum` (missing values):

```
[TEMPLATE] handleTypedefCommand called with parts: [typedef add status enum]
[TEMPLATE] Creating new type registry and loading from outline
[TEMPLATE] Loaded 0 types from outline
[TEMPLATE] Processing subcommand: add
[TEMPLATE] Executing add command: key=status, spec=enum
[TEMPLATE] handleTypedefAdd called: key=status, spec=enum
[TEMPLATE] Attempting to add type to registry
[TEMPLATE] Failed to add type: enum type requires at least one value
```

The "Failed to add type" message shows the validation error.

### Error Flow: Save Failure

If the outline can't be saved:

```
[TEMPLATE] handleTypedefAdd called: key=status, spec=enum|todo|done
[TEMPLATE] Attempting to add type to registry
[TEMPLATE] Type added to registry successfully
[TEMPLATE] Saving types back to outline
[TEMPLATE] Failed to save types to outline: write permission denied
```

The "Failed to save types" message shows the I/O error.

## Troubleshooting Guide

### Problem: Type Added But Not Showing Up

**Check logs for:**
```
[TEMPLATE] Types saved to outline successfully
[TEMPLATE] Outline has 1 items
[TEMPLATE]   Item 0: text=__types__
[TEMPLATE]     Attributes: map[status:enum|todo|done]
```

If you see this, the type WAS saved. Check if:
1. You're opening the same file
2. You saved the file (`:w`)
3. You opened it again (`:e filename`)

### Problem: `:typedef list` Shows Empty

**Check logs for:**
```
[TEMPLATE] handleTypedefList called
[TEMPLATE] Found 0 types
```

This means the types weren't found. Check:
1. Did you actually add types? (look for "Type added to registry successfully")
2. Are they in the __types__ item? (check the outline structure logs)
3. Was the file saved after adding?

### Problem: Type Validation Error

**Check logs for:**
```
[TEMPLATE] Failed to add type: [error message]
```

The error message will tell you what's wrong with your type spec. Examples:
- `enum type requires at least one value` - Use: `enum|val1|val2`
- `invalid min value for number type` - Use: `number|1-5` (both must be numbers)
- `unknown type kind` - Check the kind name is valid

## Log Format

Each log line follows this format:
```
[TEMPLATE] <function_name> message
```

Where:
- `[TEMPLATE]` - Log prefix identifying template system logs
- `<function_name>` - Which function produced the log
- `message` - The actual log message

## Tips for Debugging

1. **Search for errors:**
   ```bash
   grep -i "error\|failed" /tmp/tuo-template-debug.log
   ```

2. **Follow a single command:**
   ```bash
   grep "status" /tmp/tuo-template-debug.log
   ```

3. **See outline state:**
   ```bash
   grep "Outline has\|Item.*:\|Attributes:" /tmp/tuo-template-debug.log
   ```

4. **Track type counts:**
   ```bash
   grep "Found.*types\|Loaded.*types" /tmp/tuo-template-debug.log
   ```

## Disabling Logging

If you want to disable logging, modify `template_commands.go`:

```go
var debugLog *log.Logger

func init() {
    // Disabled - comment out to prevent logging
    // debugLog = log.New(...)
    debugLog = log.New(io.Discard, "", 0)  // Send to /dev/null
}
```

Or simply delete the log file and create a symlink to /dev/null:

```bash
rm /tmp/tuo-template-debug.log
ln -s /dev/null /tmp/tuo-template-debug.log
```

## Getting Help

When reporting typedef issues, include:

1. The commands you ran
2. The output of `/tmp/tuo-template-debug.log`
3. Your current outline file (with :w)
4. What you expected vs. what you got

The logs will show exactly where the problem is.
