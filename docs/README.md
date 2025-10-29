# Documentation Index

This directory contains comprehensive documentation for the tui-outliner project.

## Quick Start Guides

### [VISUAL_MODE.md](VISUAL_MODE.md)
User-friendly guide to visual mode. Start here if you want to learn how to use visual mode features.

**Topics:**
- How to enter visual mode
- Navigation in visual mode
- Performing bulk operations (delete, copy, indent, outdent)
- Examples and workflow

### [KEYBINDINGS_REFERENCE.md](KEYBINDINGS_REFERENCE.md)
Quick reference table for all keybindings organized by mode.

**Includes:**
- Normal mode keybindings
- Visual mode keybindings
- Insert mode keybindings
- Search and command mode keybindings
- Tips and best practices

## Technical Documentation

### [KEYBINDINGS_ARCHITECTURE.md](KEYBINDINGS_ARCHITECTURE.md)
Architecture and design of the keybindings system.

**Topics:**
- Mode-based keybinding organization
- KeyBinding struct and lookup functions
- Event flow and routing
- How to add new keybindings
- Special key handling

### [VISUAL_MODE_IMPLEMENTATION.md](VISUAL_MODE_IMPLEMENTATION.md)
Deep dive into the visual mode implementation.

**Topics:**
- State management (mode enum, visual anchor)
- Event routing
- Selection calculation and rendering
- Operation implementations (delete, yank, indent, outdent)
- Tree methods for visual operations
- Theme integration
- Detailed flow examples
- Known limitations and future enhancements

## File Structure

```
docs/
├── README.md (this file)
├── VISUAL_MODE.md (user guide)
├── KEYBINDINGS_REFERENCE.md (quick reference)
├── KEYBINDINGS_ARCHITECTURE.md (technical)
└── VISUAL_MODE_IMPLEMENTATION.md (detailed technical)
```

## For Different Audiences

### Users
1. Start with [VISUAL_MODE.md](VISUAL_MODE.md) to learn visual mode
2. Check [KEYBINDINGS_REFERENCE.md](KEYBINDINGS_REFERENCE.md) for quick key lookup

### Developers
1. Read [KEYBINDINGS_ARCHITECTURE.md](KEYBINDINGS_ARCHITECTURE.md) to understand the system
2. Check [VISUAL_MODE_IMPLEMENTATION.md](VISUAL_MODE_IMPLEMENTATION.md) for implementation details
3. Use [KEYBINDINGS_REFERENCE.md](KEYBINDINGS_REFERENCE.md) as a checklist when modifying keybindings

### Contributors
1. Start with [KEYBINDINGS_ARCHITECTURE.md](KEYBINDINGS_ARCHITECTURE.md) to understand design patterns
2. Read [VISUAL_MODE_IMPLEMENTATION.md](VISUAL_MODE_IMPLEMENTATION.md) before making changes
3. Follow the established patterns when adding features

## Key Concepts

### Modes
The application uses vim-style modal editing:
- **Normal Mode**: Navigation and command execution
- **Visual Mode**: Multi-item selection and bulk operations
- **Insert Mode**: Text editing
- **Search Mode**: Finding items
- **Command Mode**: Running commands (save, quit, etc.)

### Keybindings
Each mode has dedicated keybindings defined as structs with:
- Key: The rune to bind
- Description: What it does
- Handler: The function to execute

### Visual Mode Features
- Line-wise selection (select complete items with their children)
- Bulk delete, copy, indent, and outdent operations
- Visual feedback with distinct highlighting
- Automatic mode exit after operations

## Related Documentation

See also:
- `../README.md` - Main project documentation
- `../CLAUDE.md` - Development guide
- `../internal/*/` - Source code comments

## Contributing

When adding new features:
1. Update relevant documentation
2. Add keybindings following established patterns
3. Update [KEYBINDINGS_REFERENCE.md](KEYBINDINGS_REFERENCE.md) with new shortcuts
4. Add detailed docs in this directory if adding new modes or major features
