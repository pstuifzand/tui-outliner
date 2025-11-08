# Kitty Keyboard Protocol Support

## Overview

The TUI Outliner (`tuo`) now supports the Kitty keyboard protocol for enhanced key reporting. This modern terminal protocol provides better keyboard input handling, including:

- **Key disambiguation**: Distinguish between keys that traditionally have the same escape codes (e.g., Tab vs Ctrl-I, Enter vs Ctrl-M)
- **Better modifier support**: More accurate reporting of modifier keys (Ctrl, Shift, Alt, Super)
- **Enhanced key events**: Support for press, repeat, and release events

## Implementation

The Kitty keyboard protocol is implemented in `internal/ui/screen.go` and is automatically enabled when the application starts and disabled when it exits.

### Enabling the Protocol

When the screen is initialized (`NewScreenWithTheme`), the following CSI escape sequence is sent:

```
CSI > 1 u
```

This sequence:
- Pushes the current keyboard flags onto a stack
- Enables level 1 of the protocol (basic key disambiguation)

### Disabling the Protocol

When the screen is closed, the following CSI escape sequence is sent:

```
CSI < u
```

This pops the keyboard flags from the stack, restoring the terminal's previous state.

### Backward Compatibility

The Kitty keyboard protocol is **backward compatible**:
- Terminals that support the protocol will enable enhanced key reporting
- Terminals that don't support it will silently ignore the escape sequences
- The application works normally in both cases

## Technical Details

### Escape Sequences

The implementation uses the following CSI (Control Sequence Introducer) sequences:

| Sequence | Description |
|----------|-------------|
| `\x1b[>1u` | Push current flags and enable level 1 |
| `\x1b[<u` | Pop keyboard flags from stack |

### Progressive Enhancement Levels

The Kitty keyboard protocol supports multiple enhancement levels. Currently, `tuo` uses **level 1**, which provides:
- Disambiguation of keys with overlapping escape codes
- Basic enhanced modifier key reporting

Future enhancements could enable higher levels for additional features.

### Code Location

The relevant code is in `internal/ui/screen.go`:

- `enableKittyKeyboardProtocol()` - Enables the protocol
- `disableKittyKeyboardProtocol()` - Disables the protocol
- `writeEscapeSequence()` - Writes raw escape sequences to the terminal

## Benefits for Users

With the Kitty keyboard protocol enabled, users can:

1. **More reliable key bindings**: Reduced conflicts between key combinations
2. **Better modifier key support**: More accurate detection of Ctrl, Shift, Alt combinations
3. **Terminal compatibility**: Works across modern terminals (Kitty, WezTerm, foot, iTerm2, and more)

## Supported Terminals

The Kitty keyboard protocol is supported by:

- [Kitty](https://sw.kovidgoyal.net/kitty/)
- [WezTerm](https://wezfurlong.org/wezterm/)
- [foot](https://codeberg.org/dnkl/foot)
- [iTerm2](https://iterm2.com/) (with CSI u support)
- Other modern terminals implementing the protocol

Unsupported terminals will continue to work normally with standard keyboard input.

## References

- [Kitty Keyboard Protocol Specification](https://sw.kovidgoyal.net/kitty/keyboard-protocol/)
- [Terminal Protocol Extensions](https://sw.kovidgoyal.net/kitty/protocol-extensions/)
- [CSI u - iTerm2 Documentation](https://iterm2.com/documentation-csiu.html)

## Future Enhancements

Potential future improvements:

1. **Query terminal support**: Send `CSI ? u` to detect protocol support before enabling
2. **Higher enhancement levels**: Enable levels 2+ for additional features
3. **Configuration option**: Allow users to disable the protocol if needed
4. **Enhanced key event handling**: Leverage press/release events for advanced features
