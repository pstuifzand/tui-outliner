# Configuration System Implementation Summary

## Overview

A Vim-like configuration system has been implemented for tui-outliner, allowing users to configure the application with commands like `:set visattr "date"`.

## Files Modified

### 1. **internal/config/config.go**
- Added `sessionSettings map[string]string` field to `Config` struct
- Implemented three new methods:
  - `Set(key, value string)` - Sets a configuration value
  - `Get(key string) string` - Retrieves a configuration value
  - `GetAll() map[string]string` - Returns a copy of all settings
- Updated `Load()`, `LoadFromFile()`, and `defaultConfig()` to initialize `sessionSettings`

### 2. **internal/app/app.go**
- Added import: `"github.com/pstuifzand/tui-outliner/internal/config"`
- Added `cfg *config.Config` field to `App` struct
- Updated `NewApp()` to load configuration via `config.Load()`
- Added case for "set" in `handleCommand()` switch statement
- Implemented `handleSetCommand(parts []string)` function with:
  - Show all settings: `:set`
  - Show specific setting: `:set key`
  - Set a value: `:set key value`
  - Support for quoted values (quotes are automatically removed)

## Files Created

### 1. **internal/config/config_test.go**
Comprehensive test suite covering:
- `TestSet()` - Tests setting and retrieving values
- `TestGet()` - Tests getting values including non-existent keys
- `TestGetAll()` - Tests retrieving all settings
- `TestGetAllReturnsACopy()` - Ensures `GetAll()` returns a copy, not a reference
- `TestNilSessionSettings()` - Tests handling of uninitialized sessionSettings
- `TestDefaultConfig()` - Tests default configuration initialization

### 2. **docs/configuration.md**
Complete user documentation covering:
- Usage examples for all `:set` command variants
- Available settings (with `visattr` as primary example)
- Session vs. persistent configuration explanation
- Technical implementation details
- Integration examples for developers

### 3. **examples/config_demo.json**
Example outline file demonstrating:
- Items with various attributes (date, priority, status)
- Nested items
- Proper JSON structure for configuration examples

## Key Features

✅ **Session Configuration** - Settings are stored in memory for the current session
✅ **Vim-like Syntax** - Familiar `:set` command interface
✅ **Flexible Values** - Supports any key-value pair
✅ **Quote Support** - Automatically handles quoted values
✅ **Introspection** - Can view all or specific settings
✅ **No Persistence** - Settings are lost when quitting (can be extended later)

## Usage Examples

```bash
# Set a configuration value
:set visattr date

# With quotes (quotes are removed)
:set visattr "my long value"

# View a specific setting
:set visattr
# Output: visattr=date

# View all settings
:set
# Output: Settings: visattr=date, other_key=value

# Show non-existent setting
:set nonexistent
# Output: Setting 'nonexistent' is not set
```

## How to Use in Code

To access configuration values in other parts of the application:

```go
// Get a configuration value
visattrValue := a.cfg.Get("visattr")
if visattrValue != "" {
    // Use the configuration value
    fmt.Println("Visible attribute:", visattrValue)
}

// Set a value programmatically
a.cfg.Set("my_setting", "my_value")

// Get all settings
allSettings := a.cfg.GetAll()
for key, value := range allSettings {
    fmt.Printf("%s = %s\n", key, value)
}
```

## Testing

All tests pass successfully:

```bash
$ go test ./internal/config -v
=== RUN   TestSet
--- PASS: TestSet (0.00s)
=== RUN   TestGet
--- PASS: TestGet (0.00s)
=== RUN   TestGetAll
--- PASS: TestGetAll (0.00s)
=== RUN   TestGetAllReturnsACopy
--- PASS: TestGetAllReturnsACopy (0.00s)
=== RUN   TestNilSessionSettings
--- PASS: TestNilSessionSettings (0.00s)
=== RUN   TestDefaultConfig
--- PASS: TestDefaultConfig (0.00s)
PASS
```

## Future Enhancements

Potential improvements:
- Persist configuration to `~/.config/tui-outliner/session.toml`
- Add configuration validation (e.g., enum values for certain settings)
- Support for `:unset` command to remove settings
- Configuration profiles for different use cases
- Integration with tree view rendering to use `visattr` configuration
