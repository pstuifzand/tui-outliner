# Typedef Debugging Guide

This document explains how to debug the `:typedef` commands using the built-in logging system.

## Overview

The template system includes comprehensive debug logging that writes to `/tmp/tuo-template-debug.log`. This helps track what happens when you execute typedef commands.

## How Logging Works

When the app starts, it creates a logger that writes all template operations to `/tmp/tuo-template-debug.log`. Every typedef command is logged with detailed information about:

1. Command arguments received
2. Validation checks
3. Type registry operations
4. Outline modifications
5. Error conditions

## Log Levels and Messages

### Main Command Handler (`handleTypedefCommand`)

```
[TEMPLATE] handleTypedefCommand called with parts: [typedef add status enum|todo|done]
[TEMPLATE] Creating new type registry and loading from outline
[TEMPLATE] Loaded 0 types from outline
[TEMPLATE] Processing subcommand: add
[TEMPLATE] Executing add command: key=status, spec=enum|todo|done
```

This shows:
- The exact command and arguments received
- How many types were already in the outline
- Which subcommand is being executed

### Add Type Handler (`handleTypedefAdd`)

```
[TEMPLATE] handleTypedefAdd called: key=status, spec=enum|todo|done
[TEMPLATE] Attempting to add type to registry
[TEMPLATE] Type added to registry successfully
[TEMPLATE] Saving types back to outline
[TEMPLATE] Types saved to outline successfully
[TEMPLATE] Checking outline structure after save
[TEMPLATE] Outline has 1 items
[TEMPLATE]   Item 0: text=__types__
[TEMPLATE]     Attributes: map[status:enum|todo|done]
[TEMPLATE] App marked as dirty, status set
```

This shows:
- The exact key and spec being added
- Success of parsing and addition to registry
- Success of saving to outline
- The outline structure **after** the save
- Which item holds the types (__types__)
- The actual attributes that were saved

### List Types Handler (`handleTypedefList`)

```
[TEMPLATE] handleTypedefList called
[TEMPLATE] Found 2 types
[TEMPLATE]   Type: status (kind: enum)
[TEMPLATE]   Type: priority (kind: number)
[TEMPLATE] Status message: Types: status: enum(todo,done) | priority: number(1-5)
```

This shows:
- How many types were found
- The name and kind of each type
- The formatted message displayed to user

## Testing the typedef Commands

### Step 1: Prepare Test Outline

```bash
cat > /tmp/test_outline.json << 'EOF'
{
  "items": [
    {
      "id": "item_test_1",
      "text": "Test Item",
      "metadata": {
        "created": "2025-11-05T00:00:00Z",
        "modified": "2025-11-05T00:00:00Z",
        "attributes": {}
      }
    }
  ]
}
EOF
```

### Step 2: Clear Old Logs

```bash
rm -f /tmp/tuo-template-debug.log
```

### Step 3: Start the Application

```bash
cd /home/peter/work/tui-outliner
./tuo /tmp/test_outline.json
```

### Step 4: Run Commands

Inside the app, run these commands:

```
:typedef list
:typedef add status enum|todo|in-progress|done
:typedef list
:typedef add priority number|1-5
:typedef list
:typedef remove status
:typedef list
```

### Step 5: Check Logs

```bash
cat /tmp/tuo-template-debug.log
```

Or follow logs in real-time:

```bash
tail -f /tmp/tuo-template-debug.log
```

## Expected Log Output Sequence

For the command `:typedef add status enum|todo|done`:

1. **Command entry:**
   ```
   handleTypedefCommand called with parts: [typedef add status enum|todo|done]
   ```

2. **Registry loading:**
   ```
   Creating new type registry and loading from outline
   Loaded X types from outline
   ```

3. **Subcommand execution:**
   ```
   Processing subcommand: add
   Executing add command: key=status, spec=enum|todo|done
   ```

4. **Type addition:**
   ```
   handleTypedefAdd called: key=status, spec=enum|todo|done
   Attempting to add type to registry
   Type added to registry successfully
   ```

5. **Outline save:**
   ```
   Saving types back to outline
   Types saved to outline successfully
   ```

6. **Verification:**
   ```
   Checking outline structure after save
   Outline has 1 items
     Item 0: text=__types__
       Attributes: map[status:enum|todo|done]
   ```

7. **Completion:**
   ```
   App marked as dirty, status set
   ```

## Common Issues and How to Spot Them in Logs

### Issue: Type not being saved

**Symptom:** Command succeeds but type doesn't appear in `:typedef list`

**What to look for in logs:**
- "Type added to registry successfully" - if present, parsing worked
- "Saving types back to outline" section - look for any errors
- "Checking outline structure" - verify the `__types__` item exists and has attributes

**Example problem log:**
```
[TEMPLATE] handleTypedefAdd called: key=status, spec=enum|todo|done
[TEMPLATE] Attempting to add type to registry
[TEMPLATE] Type added to registry successfully
[TEMPLATE] Saving types back to outline
[TEMPLATE] Failed to save types to outline: __types__ item is nil
```

This shows the type was added to registry but couldn't be saved because __types__ item doesn't exist.

### Issue: Type validation failing

**Symptom:** "Invalid type definition" error message

**What to look for in logs:**
```
[TEMPLATE] Attempting to add type to registry
[TEMPLATE] Failed to add type: enum type requires at least one value
```

This shows the spec parsing failed.

### Issue: Registry not loading existing types

**Symptom:** Command says "No type definitions" even though types were previously added

**What to look for in logs:**
```
[TEMPLATE] Creating new type registry and loading from outline
[TEMPLATE] Loaded 0 types from outline
```

If this shows 0 but you added types previously, check:
1. Are you opening the correct file?
2. Was the file saved after adding types?
3. Check the outline structure:
   ```
   [TEMPLATE] Checking outline structure after save
   [TEMPLATE] Outline has 1 items
   [TEMPLATE]   Item 0: text=__types__
   [TEMPLATE]     Attributes: map[status:enum|todo|done]
   ```

## Reading the Outline Structure Logs

After a typedef add command, you'll see:

```
[TEMPLATE] Checking outline structure after save
[TEMPLATE] Outline has 1 items
[TEMPLATE]   Item 0: text=__types__
[TEMPLATE]     Attributes: map[status:enum|todo|done]
```

This means:
- The outline has 1 item total
- That item has text "__types__" (the special types container)
- Its attributes contain the type definition: `status` → `enum|todo|done`

If attributes show empty:
```
[TEMPLATE] Outline has 2 items
[TEMPLATE]   Item 0: text=__types__
[TEMPLATE]   Item 1: text=Test Item
```

This means types weren't saved - the __types__ item has no attributes.

## Disabling Debug Logs

If you want to disable logs, comment out or remove the init() function in `template_commands.go`:

```go
// var debugLog *log.Logger
//
// func init() {
//     // Logging disabled
// }
```

Or redirect them to `/dev/null`:

```bash
rm /tmp/tuo-template-debug.log
touch /dev/null
ln -s /dev/null /tmp/tuo-template-debug.log
```

## Getting Help with Logs

If typedef commands aren't working:

1. Run the test commands above
2. Share the output of `/tmp/tuo-template-debug.log`
3. Include what you expected vs. what you got
4. Include the current state of your outline file

The logs will show exactly where the problem is occurring.

