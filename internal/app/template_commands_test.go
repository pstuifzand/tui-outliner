package app

import (
	"strings"
	"testing"

	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/storage"
	"github.com/pstuifzand/tui-outliner/internal/ui"
)

// TestTypedefListEmpty tests :typedef list with no types defined
func TestTypedefListEmpty(t *testing.T) {
	app := createTestApp()

	parts := []string{"typedef", "list"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "No type definitions") {
		t.Errorf("Expected 'No type definitions' message, got: %s", app.statusMsg)
	}
}

// TestTypedefAdd tests adding a type definition
func TestTypedefAddEnum(t *testing.T) {
	app := createTestApp()

	parts := []string{"typedef", "add", "status", "enum|todo|in-progress|done"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "Added type: status") {
		t.Errorf("Expected success message, got: %s", app.statusMsg)
	}

	if !app.dirty {
		t.Errorf("App should be marked dirty after adding type")
	}

	// Verify type was saved
	parts = []string{"typedef", "list"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "status") {
		t.Errorf("Status type should appear in list, got: %s", app.statusMsg)
	}
}

// TestTypedefAddNumber tests adding a number type
func TestTypedefAddNumber(t *testing.T) {
	app := createTestApp()

	parts := []string{"typedef", "add", "priority", "number|1-5"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "Added type: priority") {
		t.Errorf("Expected success message, got: %s", app.statusMsg)
	}

	// List and verify
	parts = []string{"typedef", "list"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "priority") {
		t.Errorf("Priority type should appear in list, got: %s", app.statusMsg)
	}
}

// TestTypedefAddDate tests adding a date type
func TestTypedefAddDate(t *testing.T) {
	app := createTestApp()

	parts := []string{"typedef", "add", "deadline", "date"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "Added type: deadline") {
		t.Errorf("Expected success message, got: %s", app.statusMsg)
	}
}

// TestTypedefAddInvalidSpec tests adding invalid type spec
func TestTypedefAddInvalidSpec(t *testing.T) {
	app := createTestApp()

	// Enum without values should fail
	parts := []string{"typedef", "add", "status", "enum"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "Invalid type definition") {
		t.Errorf("Should reject enum without values, got: %s", app.statusMsg)
	}
}

// TestTypedefAddInvalidNumberRange tests invalid number range
func TestTypedefAddInvalidNumberRange(t *testing.T) {
	app := createTestApp()

	parts := []string{"typedef", "add", "priority", "number|abc-def"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "Invalid type definition") {
		t.Errorf("Should reject invalid number range, got: %s", app.statusMsg)
	}
}

// TestTypedefAddMissingArgs tests missing arguments
func TestTypedefAddMissingArgs(t *testing.T) {
	app := createTestApp()

	parts := []string{"typedef", "add", "status"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, ":typedef add") {
		t.Errorf("Should show usage, got: %s", app.statusMsg)
	}
}

// TestTypedefRemove tests removing a type definition
func TestTypedefRemove(t *testing.T) {
	app := createTestApp()

	// Add a type first
	parts := []string{"typedef", "add", "status", "enum|todo|done"}
	app.handleTypedefCommand(parts)

	// Verify it was added
	parts = []string{"typedef", "list"}
	app.handleTypedefCommand(parts)
	if !strings.Contains(app.statusMsg, "status") {
		t.Fatalf("Type should be added before removing")
	}

	// Remove it
	parts = []string{"typedef", "remove", "status"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "Removed type: status") {
		t.Errorf("Expected remove success message, got: %s", app.statusMsg)
	}

	// Verify it was removed
	parts = []string{"typedef", "list"}
	app.handleTypedefCommand(parts)
	if strings.Contains(app.statusMsg, "status") {
		t.Errorf("Type should be removed from list, got: %s", app.statusMsg)
	}
}

// TestTypedefRemoveNonexistent tests removing non-existent type
func TestTypedefRemoveNonexistent(t *testing.T) {
	app := createTestApp()

	parts := []string{"typedef", "remove", "nonexistent"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "Type not found") {
		t.Errorf("Should report type not found, got: %s", app.statusMsg)
	}
}

// TestTypedefRemoveMissingArgs tests missing arguments for remove
func TestTypedefRemoveMissingArgs(t *testing.T) {
	app := createTestApp()

	parts := []string{"typedef", "remove"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, ":typedef remove") {
		t.Errorf("Should show usage, got: %s", app.statusMsg)
	}
}

// TestTypedefListShowsMultiple tests listing multiple types
func TestTypedefListShowsMultiple(t *testing.T) {
	app := createTestApp()

	// Add multiple types
	app.handleTypedefCommand([]string{"typedef", "add", "status", "enum|todo|done"})
	app.handleTypedefCommand([]string{"typedef", "add", "priority", "number|1-5"})
	app.handleTypedefCommand([]string{"typedef", "add", "deadline", "date"})

	// List them
	app.handleTypedefCommand([]string{"typedef", "list"})

	if !strings.Contains(app.statusMsg, "status") {
		t.Errorf("Should show status type, got: %s", app.statusMsg)
	}
	if !strings.Contains(app.statusMsg, "priority") {
		t.Errorf("Should show priority type, got: %s", app.statusMsg)
	}
	if !strings.Contains(app.statusMsg, "deadline") {
		t.Errorf("Should show deadline type, got: %s", app.statusMsg)
	}
}

// TestReadOnlyBlocked tests that readonly files block typedef changes
func TestReadOnlyBlocked(t *testing.T) {
	app := createTestApp()
	app.readOnly = true

	parts := []string{"typedef", "add", "status", "enum|todo|done"}
	app.handleTypedefCommand(parts)

	if !strings.Contains(app.statusMsg, "readonly") {
		t.Errorf("Should block readonly modifications, got: %s", app.statusMsg)
	}
}

// Helper function to create a test app
func createTestApp() *App {
	// Create a default item for testing
	item := model.NewItem("Test Item")

	app := &App{
		outline:          model.NewOutline(),
		statusMsg:        "",
		dirty:            false,
		readOnly:         false,
		screen:           createMockScreen(),
		tree:             ui.NewTreeView([]*model.Item{}),
		store:            &storage.JSONStore{},
		editor:           ui.NewMultiLineEditor(item),
		nodeSearchWidget: &ui.NodeSearchWidget{},
	}

	// Add the default item to outline
	app.outline.Items = []*model.Item{item}
	app.tree = ui.NewTreeView(app.outline.Items)

	return app
}

// Mock screen for testing
func createMockScreen() *ui.Screen {
	// Create a minimal mock screen
	screen, _ := ui.NewScreen()
	return screen
}
