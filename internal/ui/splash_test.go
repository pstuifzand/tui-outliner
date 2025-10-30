package ui

import (
	"testing"
)

func TestSplashScreenVisibility(t *testing.T) {
	splash := NewSplashScreen()

	// Initially should be hidden
	if splash.IsVisible() {
		t.Error("Splash screen should be hidden initially")
	}

	// Show it
	splash.Show()
	if !splash.IsVisible() {
		t.Error("Splash screen should be visible after Show()")
	}

	// Hide it
	splash.Hide()
	if splash.IsVisible() {
		t.Error("Splash screen should be hidden after Hide()")
	}
}

func TestSplashScreenContent(t *testing.T) {
	splash := NewSplashScreen()
	content := splash.GetContent()

	// Verify content is not empty
	if len(content) == 0 {
		t.Error("Splash screen content should not be empty")
	}

	// Verify key information is present
	contentStr := ""
	for _, line := range content {
		contentStr += line + "\n"
	}

	requiredStrings := []string{"TUI Outliner", "Version", ":e", ":q", ":wq"}
	for _, required := range requiredStrings {
		if len(contentStr) == 0 || (contentStr != "" && !containsString(contentStr, required)) {
			// If content is not empty and doesn't contain the required string, that's an error
			if len(contentStr) > 0 {
				t.Errorf("Splash screen content should contain '%s'", required)
			}
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
