# Configuration Quick Reference

## The `:set` Command

### Syntax
```
:set [key] [value]
```

### Examples

| Command | Effect |
|---------|--------|
| `:set` | Show all configured settings |
| `:set visattr` | Show current value of `visattr` setting |
| `:set visattr date` | Set `visattr` to `date` |
| `:set visattr "my value"` | Set `visattr` to `my value` (quotes removed) |
| `:set debug_mode on` | Set `debug_mode` to `on` |
| `:set theme dark` | Set `theme` to `dark` |

### Common Settings

#### `visattr` - Visible Attributes
Controls which item attributes are displayed in the tree view.

```
:set visattr date          # Show date attribute
:set visattr priority      # Show priority attribute
:set visattr status        # Show status attribute
```

#### Custom Settings
You can create any custom settings for your use case:

```
:set my_custom_key my_custom_value
:set flag_name enabled
```

## Accessing Settings in Code

```go
// In internal/app/app.go or related modules
visattr := a.cfg.Get("visattr")
if visattr != "" {
    // Use the setting
}

// Set a value programmatically
a.cfg.Set("key", "value")

// Get all settings
allSettings := a.cfg.GetAll()
```

## Notes

- Settings are stored in **memory only** (session-only)
- Settings are lost when you quit the application
- No configuration file is modified by `:set` commands
- Quoted values are supported and quotes are automatically removed
- Setting keys and values can contain spaces (when quoted)
