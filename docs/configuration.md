# Configuration System

The tui-outliner supports a Vim-like configuration system using the `:set` command. This allows you to customize the application behavior during a session.

## Usage

### Setting a configuration value

```
:set key value
```

Example:
```
:set visattr date
:set visattr "my long value"
```

Quotes are optional and automatically removed if present.

### Viewing a specific setting

```
:set key
```

Example:
```
:set visattr
```

Shows the current value of `visattr`, or indicates it's not set.

### Viewing all settings

```
:set
```

Shows all currently configured settings in the format `key=value`.

## Available Settings

The following settings are commonly used:

### `visattr` - Visible Attributes

Controls which item attributes should be displayed when rendering items in the tree view.

**Example:**
```
:set visattr date
```

This tells the application to visually display the `date` attribute for items.

**Common values:**
- `date` - Display date attributes
- `priority` - Display priority attributes
- `status` - Display status attributes
- Or any custom attribute name you've defined

### Custom Settings

You can create and use any custom settings that your application needs. The configuration system is generic and supports any key-value pair.

**Example:**
```
:set my_custom_setting "my value"
:set theme dark
:set debug on
```

## Session vs Persistent Configuration

- **Session settings** (`:set` command) - Stored in memory for the current session only. They are lost when you quit the application.
- **Persistent settings** - Can be configured in the TOML configuration file at `~/.config/tui-outliner/config.toml` (currently supports `theme` setting).

## Technical Details

The configuration system is built into the `App` struct and backed by the `Config` type in `internal/config/config.go`.

- Settings are stored in the application's memory
- The `:set` command handler is in `internal/app/app.go:handleSetCommand()`
- Configuration can be accessed programmatically via:
  - `app.cfg.Set(key, value)` - Set a configuration value
  - `app.cfg.Get(key)` - Get a configuration value
  - `app.cfg.GetAll()` - Get all configuration values

## Examples

**Set visible attribute to date:**
```
:set visattr date
```

**Set multiple configuration values:**
```
:set visattr date
:set debug_mode on
:set theme nord
```

**Check what's currently configured:**
```
:set
```

**Check a specific setting:**
```
:set visattr
```

## Integration with Features

To use configuration values in your features, access them via:

```go
visattrValue := a.cfg.Get("visattr")
if visattrValue != "" {
    // Use the configuration value
}
```

This allows features like the tree view renderer to check if certain attributes should be displayed based on the `visattr` configuration.
