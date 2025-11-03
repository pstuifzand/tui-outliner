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

func TestMoveUpDownSymmetry(t *testing.T) {
	// Test that moving down then up returns to original position
	itemA := model.NewItem("Item A")
	itemB := model.NewItem("Item B")
	itemC := model.NewItem("Item C")
	itemD := model.NewItem("Item D")
	itemE := model.NewItem("Item E")

	// Collapse all items to keep moves at root level only
	itemA.Expanded = false
	itemB.Expanded = false
	itemC.Expanded = false
	itemD.Expanded = false
	itemE.Expanded = false

	items := []*model.Item{itemA, itemB, itemC, itemD, itemE}
	tv := NewTreeView(items)

	// Test 1: Move down then up should return to original position
	tv.SelectItem(1) // Select Item B
	initialParentB := itemB.Parent
	initialIdxB := 1 // B starts at index 1

	// Move down
	if !tv.MoveItemDown() {
		t.Error("MoveItemDown failed")
	}

	// Move up
	if !tv.MoveItemUp() {
		t.Error("MoveItemUp failed")
	}

	// Verify we're back at original position
	finalParentB := itemB.Parent
	finalIdxB := -1
	if finalParentB == nil {
		// Find B in the items array
		for i, item := range tv.items {
			if item.ID == itemB.ID {
				finalIdxB = i
				break
			}
		}
	}

	if finalParentB != initialParentB {
		t.Errorf("Item B parent changed after down-up: initial=%v, final=%v", initialParentB, finalParentB)
	}
	if finalIdxB != initialIdxB {
		t.Errorf("Item B index changed after down-up: initial=%d, final=%d", initialIdxB, finalIdxB)
	}

	// Test 2: Move up then down should also be symmetric
	// But we need to re-find C's position since moves might have changed the order
	var currentCIdx int
	for i, item := range tv.items {
		if item.ID == itemC.ID {
			currentCIdx = i
			break
		}
	}
	tv.SelectItem(currentCIdx) // Select Item C
	initialParentC := itemC.Parent

	// Move up
	if !tv.MoveItemUp() {
		t.Error("MoveItemUp failed")
	}

	// Move down
	if !tv.MoveItemDown() {
		t.Error("MoveItemDown failed")
	}

	// Verify we're back at original position
	finalParentC := itemC.Parent
	finalIdxC := -1
	if finalParentC == nil {
		for i, item := range tv.items {
			if item.ID == itemC.ID {
				finalIdxC = i
				break
			}
		}
	}

	if finalParentC != initialParentC {
		t.Errorf("Item C parent changed after up-down: initial=%v, final=%v", initialParentC, finalParentC)
	}
	if finalIdxC != currentCIdx {
		t.Errorf("Item C index changed after up-down: initial=%d, final=%d", currentCIdx, finalIdxC)
	}
}

func TestMoveDoesNotExpandCollapsedNodes(t *testing.T) {
	// Create tree with collapsed node
	itemA := model.NewItem("Item A")
	itemB := model.NewItem("Item B")
	itemC := model.NewItem("Item C")
	itemD := model.NewItem("Item D")

	itemB.AddChild(itemC)
	itemB.Expanded = false // Collapsed

	items := []*model.Item{itemA, itemB, itemD}
	tv := NewTreeView(items)

	// Store initial state - B should be collapsed
	if itemB.Expanded {
		t.Error("Item B should start as collapsed")
	}

	// Try to move D up (should not go into collapsed B)
	tv.SelectItem(2) // Select D
	initialParentD := itemD.Parent

	success := tv.MoveItemUp()
	if success && itemD.Parent == itemB {
		t.Error("Item D should not have moved into collapsed node B")
	}

	// B should remain collapsed
	if itemB.Expanded {
		t.Error("Collapsed node B was expanded by move operation")
	}

	// D should stay at root level (or at least not in B)
	if itemD.Parent == itemB {
		t.Error("Item D moved into collapsed node, violating requirements")
	}

	// If move succeeded, D should have same parent as before or at root
	if success && itemD.Parent != initialParentD {
		if itemD.Parent != nil {
			t.Error("Item D parent unexpectedly changed")
		}
	}
}

func TestMoveWithCollapsedNodeHavingChildren(t *testing.T) {
	// Test that collapsed nodes with children don't accept new children via move
	itemA := model.NewItem("Item A")
	itemB := model.NewItem("Item B")
	itemC := model.NewItem("Item C - Child of B")
	itemD := model.NewItem("Item D")
	itemE := model.NewItem("Item E")

	itemB.AddChild(itemC)
	itemB.Expanded = false // Collapsed

	items := []*model.Item{itemA, itemB, itemD, itemE}
	tv := NewTreeView(items)

	// Verify B is collapsed
	if itemB.Expanded {
		t.Error("Item B should be collapsed")
	}

	// Try to move E up - it should skip over collapsed B
	tv.SelectItem(3) // Select E
	success := tv.MoveItemUp()

	// E should not be a child of B
	if itemE.Parent == itemB {
		t.Error("Item E should not have moved into collapsed node B (even though B has children)")
	}

	// B should still be collapsed
	if itemB.Expanded {
		t.Error("Move operation should not expand collapsed nodes")
	}

	if success {
		t.Logf("Move succeeded. E is now at parent: %v (B is parent: %v)", itemE.Parent, itemE.Parent == itemB)
	}
}

func TestMoveItemADownAndUp(t *testing.T) {
	// Test moving Item A through all positions down and then back up
	// This mirrors the structure from examples/move_demo.json with expanded nodes:
	// A (root, collapsed)
	// B (root, EXPANDED) -> C (child)
	// D (root, collapsed)
	// E (root, EXPANDED) -> F (child)
	//
	// With B and E expanded, the DFS position sequence is:
	// (root,0), (root,1), (B,0), (B,1), (root,2), (root,3), (E,0), (E,1), (root,4)
	//
	// A starts at root[0] and should move through: root[0] → root[1] → B[0] → B[1] → root[2] → root[3] → E[0] → E[1] → root[4]

	itemA := model.NewItem("Item A")
	itemA.Expanded = false // Keep collapsed to prevent nesting

	itemB := model.NewItem("Item B")
	itemC := model.NewItem("Item C")
	itemB.AddChild(itemC)
	itemB.Expanded = true // Expanded - allows A to move into it

	itemD := model.NewItem("Item D")
	itemD.Expanded = false // Collapsed - skip its positions

	itemE := model.NewItem("Item E")
	itemF := model.NewItem("Item F")
	itemE.AddChild(itemF)
	itemE.Expanded = true // Expanded - allows A to move into it

	items := []*model.Item{itemA, itemB, itemD, itemE}
	tv := NewTreeView(items)

	// Track all positions of Item A as we move it down
	type Position struct {
		ParentText string
		Index      int
	}
	var positionsDown []Position
	var positionsUp []Position

	// Helper to find A's current position (recursive search through all descendants)
	var findAPosition func(*model.Item) Position
	findAPosition = func(parent *model.Item) Position {
		if parent.ID == itemA.ID {
			return Position{parent.Text, -1} // Found as a parent itself - shouldn't happen for A
		}
		for i, child := range parent.Children {
			if child.ID == itemA.ID {
				return Position{parent.Text, i}
			}
			// Recursively search in this child's descendants
			pos := findAPosition(child)
			if pos.Index != -1 || pos.ParentText != "not found" {
				return pos
			}
		}
		return Position{"not found", -1}
	}

	getAPosition := func() Position {
		// Check root level first
		for i, item := range tv.items {
			if item.ID == itemA.ID {
				return Position{"root", i}
			}
		}
		// Recursively check all descendants
		for _, item := range tv.items {
			pos := findAPosition(item)
			if pos.Index != -1 {
				return pos
			}
		}
		return Position{"not found", -1}
	}

	// Record initial position
	initialPos := getAPosition()
	positionsDown = append(positionsDown, initialPos)

	// Move Item A down through all valid positions
	moveCount := 0
	prevPos := initialPos
	for {
		// Find and select A before trying to move
		currentPos := getAPosition()
		if currentPos.Index == -1 {
			t.Fatal("Item A not found in tree")
		}

		tv.SelectItemByID(itemA.ID)
		if !tv.MoveItemDown() {
			break // Can't move down anymore
		}
		moveCount++

		// Get new position
		newPos := getAPosition()
		positionsDown = append(positionsDown, newPos)

		// Check if we're making progress
		if newPos.ParentText == prevPos.ParentText && newPos.Index == prevPos.Index {
			t.Logf("Item A stuck at %s[%d], stopping", newPos.ParentText, newPos.Index)
			break
		}
		prevPos = newPos

		// Safety check to prevent infinite loops
		if moveCount > 20 {
			t.Logf("Positions down: %v", positionsDown)
			t.Fatal("Too many moves down, infinite loop detected")
		}
	}

	// A should successfully move through all positions when using stable position ordering
	// Expected: (root,0), (B,0), (B,1), (root,1), (root,2), (E,0), (E,1), (root,3)
	// = 8 positions total, so 7 moves to get from position 1 to position 8
	if moveCount < 7 {
		t.Errorf("Expected at least 7 moves down (through all expanded nodes), got %d", moveCount)
	}

	// Now move Item A back up to verify symmetry
	for i := 0; i < moveCount; i++ {
		// Find and select A before trying to move
		currentPos := getAPosition()
		if currentPos.Index == -1 {
			t.Errorf("Item A not found at step %d", i)
			break
		}

		tv.SelectItemByID(itemA.ID)
		if !tv.MoveItemUp() {
			t.Logf("MoveUp failed at step %d (A was at %s[%d])", i+1, currentPos.ParentText, currentPos.Index)
			break
		}

		// Get new position after move
		newPos := getAPosition()
		positionsUp = append(positionsUp, newPos)
	}

	// Verify symmetry: for each successful down move, we should have a corresponding up move
	if len(positionsUp) != moveCount {
		t.Errorf("Expected %d up moves (same as down), got %d", moveCount, len(positionsUp))
	}

	// Check that each position when going up matches the reverse of going down
	// positionsDown[0] is initial, positionsDown[1..moveCount] are after each down move
	// positionsUp[0..moveCount-1] are after each up move
	// So positionsUp[i] should equal positionsDown[moveCount - 1 - i]
	for i, posUp := range positionsUp {
		expectedDownIdx := moveCount - 1 - i
		if expectedDownIdx < 0 || expectedDownIdx >= len(positionsDown) {
			t.Errorf("Index out of bounds at step %d (expectedDownIdx=%d)", i, expectedDownIdx)
			continue
		}

		expectedPos := positionsDown[expectedDownIdx]
		if posUp.ParentText != expectedPos.ParentText || posUp.Index != expectedPos.Index {
			t.Errorf("Position mismatch at up[%d]: expected %s[%d] (from down[%d]), got %s[%d]",
				i, expectedPos.ParentText, expectedPos.Index, expectedDownIdx, posUp.ParentText, posUp.Index)
		}
	}

	// Final check: A should be back at original position
	finalPos := getAPosition()
	if finalPos.ParentText != initialPos.ParentText || finalPos.Index != initialPos.Index {
		t.Errorf("Item A not back at original position. Expected %s[%d], got %s[%d]",
			initialPos.ParentText, initialPos.Index, finalPos.ParentText, finalPos.Index)
	}
}

func TestPositionBuildingForCollapsedNodes(t *testing.T) {
	// Test that position building doesn't create positions inside collapsed nodes
	// Also test that expanded nodes with no children behave like collapsed nodes
	itemA := model.NewItem("Item A")
	itemB := model.NewItem("Item B")
	itemC := model.NewItem("Item C")
	itemD := model.NewItem("Item D") // Child of A for second test

	// All default to collapsed now
	items := []*model.Item{itemA, itemB, itemC}
	tv := NewTreeView(items)

	// Get positions for a simple case (all collapsed)
	positions := tv.buildAllPositions()

	t.Logf("Positions with all collapsed: %d positions", len(positions))
	for i, p := range positions {
		parent := "root"
		if p.Parent != nil {
			parent = p.Parent.Text
		}
		t.Logf("[%d] (%s,%d)", i, parent, p.Index)
	}

	// Should only have root-level positions
	// (root,0), (root,1), (root,2), (root,3)
	expectedRootPositions := 4
	actualRootPositions := 0
	for _, p := range positions {
		if p.Parent == nil {
			actualRootPositions++
		}
	}

	if actualRootPositions != expectedRootPositions {
		t.Errorf("Expected %d root positions, got %d", expectedRootPositions, actualRootPositions)
	}

	// No positions should be inside collapsed nodes
	for _, p := range positions {
		if p.Parent != nil && !p.Parent.Expanded {
			t.Errorf("Found position inside collapsed node %s", p.Parent.Text)
		}
	}

	// Now expand A (without children) and verify it still doesn't create child positions
	itemA.Expanded = true
	positions2 := tv.buildAllPositions()

	t.Logf("\nPositions with A expanded (no children): %d positions", len(positions2))
	for i, p := range positions2 {
		parent := "root"
		if p.Parent != nil {
			parent = p.Parent.Text
		}
		t.Logf("[%d] (%s,%d)", i, parent, p.Index)
	}

	// Expanded nodes with no children should NOT create child positions
	foundAChild := false
	for _, p := range positions2 {
		if p.Parent == itemA {
			foundAChild = true
			break
		}
	}

	if foundAChild {
		t.Error("Expected NO child positions for expanded item A with no children")
	}

	// Now add a child to A, expand it, and verify child positions appear
	itemA.AddChild(itemD)
	itemA.Expanded = true
	positions3 := tv.buildAllPositions()

	t.Logf("\nPositions with A expanded (with child): %d positions", len(positions3))
	for i, p := range positions3 {
		parent := "root"
		if p.Parent != nil {
			parent = p.Parent.Text
		}
		t.Logf("[%d] (%s,%d)", i, parent, p.Index)
	}

	// Now we should have positions inside A
	foundAChild = false
	for _, p := range positions3 {
		if p.Parent == itemA {
			foundAChild = true
			break
		}
	}

	if !foundAChild {
		t.Error("Expected to find child positions for expanded item A with children")
	}

	// Collapse A again - positions should be gone
	itemA.Expanded = false
	positions4 := tv.buildAllPositions()

	t.Logf("\nPositions with A re-collapsed: %d positions", len(positions4))
	for _, p := range positions4 {
		if p.Parent == itemA {
			t.Errorf("Found position inside re-collapsed node A")
		}
	}
}
