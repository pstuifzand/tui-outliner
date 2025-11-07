# Socket Commands - External Command Interface

The tuo application includes a Unix socket-based command interface that allows you to send commands to a running instance from the command line or other applications. This is useful for quickly capturing notes, integrating with other tools, or scripting workflows.

## Overview

When tuo starts, it automatically creates a Unix socket in `~/.local/share/tui-outliner/tuo-<PID>.sock`. Other processes can connect to this socket and send commands.

## Quick Start

### 1. Start tuo with a file

```bash
./tuo myfile.json
```

The application will create a socket at `~/.local/share/tui-outliner/tuo-<PID>.sock` where `<PID>` is the process ID.

### 2. Send commands from another terminal

```bash
# Add a node to the inbox
./tuo add "Buy milk"

# Add another node
./tuo add "Call dentist"

# Multi-word text is supported
./tuo add "Remember to check email"
```

## Features

### Inbox Node

The socket command system uses an **inbox** node as the default target for new items. The inbox node:

- Is automatically created if it doesn't exist
- Is marked with the attribute `@type=inbox`
- Is automatically expanded so new items are visible
- Can be manually created by adding an item with `@type=inbox` attribute

If you create an inbox node manually before sending commands, new items will be added to that node. Otherwise, a new inbox node will be created at the root level automatically.

### Command: Add Node

Add a new item to the inbox of a running tuo instance:

```bash
./tuo add "Text for the new item"
```

**Behavior:**
- Finds the item marked with `@type=inbox`
- If no inbox exists, creates one at the root level
- Adds the new item as a child of the inbox
- Marks the outline as dirty (triggers auto-save after 5 seconds)
- Sets a status message indicating the item was added

**Example workflow:**

```bash
# Start tuo
./tuo notes.json

# From another terminal, add items
./tuo add "Quick note 1"
./tuo add "Quick note 2"
./tuo add "Quick note 3"

# All three items will appear in the inbox node
```

## Socket Protocol

The socket uses a JSON-based protocol for extensibility.

### Message Format

**Request:**
```json
{
  "command": "add_node",
  "text": "Item text here",
  "target": "inbox"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Command queued"
}
```

### Available Commands

#### `add_node`

Adds a new item to the specified target.

**Fields:**
- `command`: Must be `"add_node"`
- `text`: The text content for the new item (required)
- `target`: The target location (currently only `"inbox"` is supported)

## Integration Examples

### Shell Script

```bash
#!/bin/bash
# capture-note.sh - Quick note capture script

if [ -z "$1" ]; then
    echo "Usage: $0 <note text>"
    exit 1
fi

./tuo add "$1"
```

Usage:
```bash
./capture-note.sh "Meeting with Bob at 2pm"
```

### Alfred Workflow / LaunchBar Action

You can create quick capture workflows in launcher apps:

```bash
# In Alfred, create a shell script action
/path/to/tuo add "$1"
```

### Keyboard Shortcut

Using tools like `xbindkeys` or system keyboard shortcuts, you can trigger a script that prompts for input and sends it to tuo:

```bash
#!/bin/bash
# quick-capture.sh
NOTE=$(zenity --entry --title="Quick Capture" --text="Enter note:")
if [ -n "$NOTE" ]; then
    /path/to/tuo add "$NOTE"
fi
```

## Implementation Details

### Socket Location

- Path: `~/.local/share/tui-outliner/tuo-<PID>.sock`
- Created automatically when tuo starts
- Automatically cleaned up when tuo exits
- Multiple instances can run simultaneously (each has its own socket)

### Finding Running Instances

The `--add` command automatically finds the most recently started instance by:
1. Scanning `~/.local/share/tui-outliner/` for socket files
2. Checking modification times
3. Connecting to the most recent socket

### Error Handling

- If no running instance is found, an error message is displayed
- If the connection fails, an error message is displayed
- If the command fails on the server side, the response indicates failure

### Logging

Socket operations are logged to `tuo.log` in the working directory, including:
- Socket creation and cleanup
- Incoming connections
- Received messages
- Command processing results

## Troubleshooting

### "No running tuo instance found"

This error occurs when:
- tuo is not running
- The socket file was manually deleted
- The socket directory doesn't exist

**Solution:** Start tuo first, then run the `--add` command.

### Socket file not cleaned up

If tuo crashes without cleaning up the socket file, you may see stale socket files.

**Solution:** Manually delete stale sockets:
```bash
rm ~/.local/share/tui-outliner/tuo-*.sock
```

### Permission errors

If the socket directory or files have incorrect permissions:

**Solution:** Ensure proper permissions:
```bash
chmod 755 ~/.local/share/tui-outliner/
chmod 644 ~/.local/share/tui-outliner/*.sock
```

## Future Enhancements

Potential future features:
- Support for other target nodes (not just inbox)
- Commands to query outline state
- Commands to modify existing items
- Support for setting attributes on new items
- Batch operations

## See Also

- [README.md](../README.md) - General tuo documentation
- [CLAUDE.md](../CLAUDE.md) - Developer documentation
- [examples/socket_demo.json](../examples/socket_demo.json) - Example outline
