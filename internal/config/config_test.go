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
	cfg := &Config{}
	// sessionSettings is nil

	// Set should initialize it
	cfg.Set("key", "value")
	if cfg.Get("key") != "value" {
		t.Errorf("Set should initialize nil sessionSettings")
	}

	// Get should handle nil gracefully
	cfg2 := &Config{}
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
}
