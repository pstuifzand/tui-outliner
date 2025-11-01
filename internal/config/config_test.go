package config

import (
	"testing"
)

func TestSet(t *testing.T) {
	cfg := &Config{
		sessionSettings: make(map[string]string),
	}

	cfg.Set("visattr", "date")
	if cfg.Get("visattr") != "date" {
		t.Errorf("Expected 'date', got '%s'", cfg.Get("visattr"))
	}
}

func TestGet(t *testing.T) {
	cfg := &Config{
		sessionSettings: make(map[string]string),
	}

	// Test getting a value that doesn't exist
	if cfg.Get("nonexistent") != "" {
		t.Errorf("Expected empty string for nonexistent key, got '%s'", cfg.Get("nonexistent"))
	}

	// Set and then get
	cfg.Set("test", "value")
	if cfg.Get("test") != "value" {
		t.Errorf("Expected 'value', got '%s'", cfg.Get("test"))
	}
}

func TestGetAll(t *testing.T) {
	cfg := &Config{
		Settings:        make(map[string]string),
		sessionSettings: make(map[string]string),
	}

	cfg.Set("key1", "value1")
	cfg.Set("key2", "value2")

	all := cfg.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 settings, got %d", len(all))
	}

	if all["key1"] != "value1" {
		t.Errorf("Expected 'value1', got '%s'", all["key1"])
	}

	if all["key2"] != "value2" {
		t.Errorf("Expected 'value2', got '%s'", all["key2"])
	}
}

func TestGetAllReturnsACopy(t *testing.T) {
	cfg := &Config{
		Settings:        make(map[string]string),
		sessionSettings: make(map[string]string),
	}

	cfg.Set("original", "value")

	// Modify the returned map
	all := cfg.GetAll()
	all["original"] = "modified"

	// Verify the original config was not modified
	if cfg.Get("original") != "value" {
		t.Errorf("GetAll() should return a copy, not a reference")
	}
}

func TestNilSessionSettings(t *testing.T) {
	cfg := &Config{
		Settings: make(map[string]string),
	}
	// sessionSettings is nil

	// Set should initialize it
	cfg.Set("key", "value")
	if cfg.Get("key") != "value" {
		t.Errorf("Set should initialize nil sessionSettings")
	}

	// Get should handle nil gracefully
	cfg2 := &Config{
		Settings: make(map[string]string),
	}
	if cfg2.Get("key") != "" {
		t.Errorf("Get should return empty string for nil sessionSettings")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Theme != "tokyo-night" {
		t.Errorf("Expected default theme 'tokyo-night', got '%s'", cfg.Theme)
	}

	if cfg.sessionSettings == nil {
		t.Errorf("defaultConfig should initialize sessionSettings")
	}

	if cfg.Settings == nil {
		t.Errorf("defaultConfig should initialize Settings")
	}
}

// TestGetPersistedSettings tests retrieving persisted settings from TOML
func TestGetPersistedSettings(t *testing.T) {
	cfg := &Config{
		Settings: map[string]string{
			"visattr": "date,status",
			"theme":   "custom-theme",
		},
		sessionSettings: make(map[string]string),
	}

	if cfg.Get("visattr") != "date,status" {
		t.Errorf("Expected 'date,status', got '%s'", cfg.Get("visattr"))
	}

	if cfg.Get("theme") != "custom-theme" {
		t.Errorf("Expected 'custom-theme', got '%s'", cfg.Get("theme"))
	}
}

// TestSessionSettingsOverridePersisted tests that session settings override persisted settings
func TestSessionSettingsOverridePersisted(t *testing.T) {
	cfg := &Config{
		Settings: map[string]string{
			"visattr": "date",
		},
		sessionSettings: make(map[string]string),
	}

	// Get persisted value
	if cfg.Get("visattr") != "date" {
		t.Errorf("Expected persisted 'date', got '%s'", cfg.Get("visattr"))
	}

	// Override with session setting
	cfg.Set("visattr", "status")
	if cfg.Get("visattr") != "status" {
		t.Errorf("Expected session 'status' to override, got '%s'", cfg.Get("visattr"))
	}
}

// TestGetAllMergesPersistedAndSession tests that GetAll merges both sources
func TestGetAllMergesPersistedAndSession(t *testing.T) {
	cfg := &Config{
		Settings: map[string]string{
			"visattr": "date",
			"other":   "value",
		},
		sessionSettings: make(map[string]string),
	}

	cfg.Set("newsession", "sessionvalue")

	all := cfg.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 settings, got %d", len(all))
	}

	if all["visattr"] != "date" {
		t.Errorf("Expected 'date', got '%s'", all["visattr"])
	}

	if all["newsession"] != "sessionvalue" {
		t.Errorf("Expected 'sessionvalue', got '%s'", all["newsession"])
	}
}

// TestGetAllSessionOverridesInMerge tests that session overrides persisted in GetAll
func TestGetAllSessionOverridesInMerge(t *testing.T) {
	cfg := &Config{
		Settings: map[string]string{
			"key": "persisted",
		},
		sessionSettings: make(map[string]string),
	}

	cfg.Set("key", "session")

	all := cfg.GetAll()
	if all["key"] != "session" {
		t.Errorf("Expected session override 'session', got '%s'", all["key"])
	}
}
