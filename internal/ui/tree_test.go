package ui

import (
	"fmt"
	"testing"

	"github.com/pstuifzand/tui-outliner/internal/config"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

func TestAddItemAfter(t *testing.T) {
	// Create test items
	item1 := model.NewItem("First")
	item2 := model.NewItem("Second")
	item3 := model.NewItem("Third")
	item4 := model.NewItem("Fourth")
	item5 := model.NewItem("Fifth")
	item6 := model.NewItem("Sixth")

	items := []*model.Item{item1, item2, item3, item4, item5, item6}

	// Create tree view
	tv := NewTreeView(items)

	// Select Fifth Item (index 4)
	tv.SelectItem(4)

	// Add item after Fifth
	item := model.NewItem("After fifth")
	tv.AddItemAfter(item)

	// Check the state
	t.Logf("Total items: %d", len(tv.filteredView))
	for i, item := range tv.filteredView {
		t.Logf("[%d] %s (ID: %s)", i, item.Item.Text, item.Item.ID)
	}

	// Verify that Sixth Item is still there
	found := false
	for _, item := range tv.filteredView {
		if item.Item.Text == "Sixth" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Sixth Item is missing after AddItemAfter!")
	}

	// Verify the new item is after Fifth
	if len(tv.filteredView) < 7 {
		t.Errorf("Expected at least 7 items, got %d", len(tv.filteredView))
	}

	// Check order
	if tv.filteredView[4].Item.Text != "Fifth" {
		t.Errorf("Expected Fifth at index 4, got %s", tv.filteredView[4].Item.Text)
	}
	if tv.filteredView[5].Item.Text != "After fifth" {
		t.Errorf("Expected After fifth at index 5, got %s", tv.filteredView[5].Item.Text)
	}
	if tv.filteredView[6].Item.Text != "Sixth" {
		t.Errorf("Expected Sixth at index 6, got %s", tv.filteredView[6].Item.Text)
	}
}

func TestAddItemAfterSimulatingRealScenario(t *testing.T) {
	// Create test items matching the scenario
	item1 := model.NewItem("First Item - Press gg to return here")
	item1.ID = "item_nav_1"

	item2 := model.NewItem("Second Item")
	item2.ID = "item_20251029233403_lrjgf663"

	item3 := model.NewItem("Third Item")
	item3.ID = "item_nav_3"

	item4 := model.NewItem("Fourth Item with Children")
	item4.ID = "item_nav_4"

	item5 := model.NewItem("Fifth Item")
	item5.ID = "item_nav_5"

	item6 := model.NewItem("Sixth Item")
	item6.ID = "item_20251029233541_yxnj6d60"

	items := []*model.Item{item1, item2, item3, item4, item5, item6}

	// Create tree view
	tv := NewTreeView(items)

	t.Logf("Before AddItemAfter:")
	for i, item := range tv.filteredView {
		t.Logf("[%d] %s", i, item.Item.Text)
	}

	// Select Fifth Item (should be at index 4)
	tv.SelectItem(4)
	selectedBefore := tv.GetSelected()
	t.Logf("Selected before: %s (idx: %d)", selectedBefore.Text, tv.GetSelectedIndex())

	// Simulate what happens in the 'o' keybinding
	item := model.NewItem("")
	tv.AddItemAfter(item) // Create new empty item

	t.Logf("After AddItemAfter:")
	for i, item := range tv.filteredView {
		t.Logf("[%d] %s", i, item.Item.Text)
	}

	selectedAfter := tv.GetSelected()
	t.Logf("Selected after: %s (idx: %d)", selectedAfter.Text, tv.GetSelectedIndex())

	// Check that the new item is selected
	if selectedAfter.Text != "" {
		t.Errorf("Expected empty new item to be selected, got '%s'", selectedAfter.Text)
	}

	// Check that Sixth Item is still in the list
	found := false
	foundIdx := -1
	for i, item := range tv.filteredView {
		if item.Item.Text == "Sixth Item" {
			found = true
			foundIdx = i
			break
		}
	}

	if !found {
		t.Error("Sixth Item is missing!")
		t.Logf("Final items:")
		for i, item := range tv.filteredView {
			t.Logf("[%d] %s", i, item.Item.Text)
		}
	} else {
		t.Logf("Sixth Item found at index %d", foundIdx)
	}
}

func TestAddItemAfterWithLargeCapacitySlice(t *testing.T) {
	// This test reproduces the slice capacity issue that was causing items to be lost
	// Create items with extra capacity in the underlying array to trigger the bug with the old code
	items := make([]*model.Item, 0, 10) // Pre-allocate capacity for 10 items

	for i := 0; i < 6; i++ {
		items = append(items, model.NewItem(fmt.Sprintf("Item %d", i)))
	}

	tv := NewTreeView(items)

	// Select the Fifth Item (index 4)
	tv.SelectItem(4)

	// Add a new item after it
	item := model.NewItem("New Item")
	tv.AddItemAfter(item)

	// Verify all items are present
	if len(tv.filteredView) != 7 {
		t.Errorf("Expected 7 items, got %d", len(tv.filteredView))
	}

	// Verify the order
	expectedOrder := []string{"Item 0", "Item 1", "Item 2", "Item 3", "Item 4", "New Item", "Item 5"}
	for i, expected := range expectedOrder {
		if i >= len(tv.filteredView) {
			t.Errorf("Missing item at index %d", i)
			break
		}
		if tv.filteredView[i].Item.Text != expected {
			t.Errorf("At index %d: expected '%s', got '%s'", i, expected, tv.filteredView[i].Item.Text)
		}
	}
}

func TestVisattrConfiguration(t *testing.T) {
	// Create test items with various attributes
	item1 := model.NewItem("Daily Standup")
	item1.Metadata.Attributes["type"] = "meeting"
	item1.Metadata.Attributes["date"] = "2025-11-01"

	item2 := model.NewItem("Project Task")
	item2.Metadata.Attributes["status"] = "in-progress"
	item2.Metadata.Attributes["priority"] = "high"

	item3 := model.NewItem("Item without attributes")
	// item3 has no attributes

	items := []*model.Item{item1, item2, item3}
	_ = NewTreeView(items) // Create tree view (tv not needed for this test)

	// Create config and set visattr
	cfg := &config.Config{}
	cfg.Set("visattr", "type,date")

	// Test that visattr config is correctly retrieved
	visattrConfig := cfg.Get("visattr")
	if visattrConfig != "type,date" {
		t.Errorf("Expected 'type,date', got '%s'", visattrConfig)
	}

	// Verify that items with attributes configured in visattr have them
	if value, exists := item1.Metadata.Attributes["type"]; !exists || value != "meeting" {
		t.Errorf("item1 should have type='meeting'")
	}
	if value, exists := item1.Metadata.Attributes["date"]; !exists || value != "2025-11-01" {
		t.Errorf("item1 should have date='2025-11-01'")
	}

	// Verify item2 has status attribute but not in visattr
	if value, exists := item2.Metadata.Attributes["status"]; !exists || value != "in-progress" {
		t.Errorf("item2 should have status='in-progress'")
	}

	// Verify item3 has no attributes
	if len(item3.Metadata.Attributes) > 0 {
		t.Errorf("item3 should have no attributes")
	}

	// Test with different visattr configuration
	cfg.Set("visattr", "status")
	visattrConfig = cfg.Get("visattr")
	if visattrConfig != "status" {
		t.Errorf("Expected 'status', got '%s'", visattrConfig)
	}

	// Test empty visattr
	cfg.Set("visattr", "")
	visattrConfig = cfg.Get("visattr")
	if visattrConfig != "" {
		t.Errorf("Expected empty string, got '%s'", visattrConfig)
	}
}
