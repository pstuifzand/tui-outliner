# Kitty Keyboard Protocol Support

## Overview

The TUI Outliner (`tuo`) supports the Kitty keyboard protocol for enhanced key reporting. This modern terminal protocol provides better keyboard input handling, including:

- **Key disambiguation**: Distinguish between keys that traditionally have the same escape codes (e.g., Tab vs Ctrl-I, Enter vs Ctrl-M)
- **Better modifier support**: More accurate reporting of modifier keys (Ctrl, Shift, Alt, Super)
- **Enhanced key events**: Support for press, repeat, and release events

## Status: Experimental & Disabled by Default

⚠️ **Important**: The Kitty keyboard protocol support is **disabled by default** because tcell v2.9.0 does not yet fully support parsing the enhanced escape sequences. Enabling it may cause issues with keys like Escape and Shift-Tab.

Only enable this feature if:
1. You are using a terminal that supports the Kitty keyboard protocol (Kitty, WezTerm, foot, etc.)
2. You have verified that it works correctly with your setup
3. You are willing to troubleshoot any keyboard input issues

## Enabling the Protocol

To enable the Kitty keyboard protocol, add the following to your config file at `~/.config/tui-outliner/config.toml`:

```toml
[settings]
kitty_keyboard_protocol = "true"
```

## Implementation

The Kitty keyboard protocol is implemented in `internal/ui/screen.go`. When enabled, it activates when the application starts and properly disables when it exits.

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

The Kitty keyboard protocol is **backward compatible** at the terminal level:
- Terminals that support the protocol will enable enhanced key reporting
- Terminals that don't support it will silently ignore the escape sequences

**However**, the application layer (tcell library) may not yet fully support parsing the enhanced sequences, which is why this feature is disabled by default. Future versions of tcell may improve support.

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

## Troubleshooting

If you enable the protocol and experience issues:

1. **Escape key not working**: The enhanced protocol may change how Escape is encoded
2. **Shift-Tab not working**: Enhanced key codes may not be parsed correctly by tcell
3. **Other key combinations broken**: Some modifier combinations may behave unexpectedly

**Solution**: Disable the protocol by removing or setting to "false" in your config:

```toml
[settings]
kitty_keyboard_protocol = "false"
```

## Future Enhancements

Potential future improvements:

1. **Wait for tcell support**: Monitor tcell Issue #671 for native protocol support
2. **Query terminal support**: Send `CSI ? u` to detect protocol support before enabling
3. **Higher enhancement levels**: Enable levels 2+ for additional features once tcell supports them
4. **Enhanced key event handling**: Leverage press/release events for advanced features
5. **Auto-detection**: Automatically enable when both terminal and tcell support is detected
