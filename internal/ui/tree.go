package ui

import (
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// TreeView manages the display and navigation of the outline tree
type TreeView struct {
	items        []*model.Item
	selectedIdx  int
	filterText   string
	filteredView []*displayItem
}

type displayItem struct {
	Item  *model.Item
	Depth int
}

// NewTreeView creates a new TreeView
func NewTreeView(items []*model.Item) *TreeView {
	tv := &TreeView{
		items:       items,
		selectedIdx: 0,
	}
	tv.rebuildView()
	return tv
}

// rebuildView rebuilds the filtered/display view
func (tv *TreeView) rebuildView() {
	tv.filteredView = tv.buildDisplayItems(tv.items, 0)
	if tv.selectedIdx >= len(tv.filteredView) && len(tv.filteredView) > 0 {
		tv.selectedIdx = len(tv.filteredView) - 1
	}
}

func (tv *TreeView) buildDisplayItems(items []*model.Item, depth int) []*displayItem {
	var result []*displayItem
	for _, item := range items {
		// Only show if expanded or at top level
		if depth == 0 || (item.Parent != nil && item.Parent.Expanded) {
			result = append(result, &displayItem{Item: item, Depth: depth})
			if item.Expanded && len(item.Children) > 0 {
				result = append(result, tv.buildDisplayItems(item.Children, depth+1)...)
			}
		}
	}
	return result
}

// SelectNext moves selection down
func (tv *TreeView) SelectNext() {
	if tv.selectedIdx < len(tv.filteredView)-1 {
		tv.selectedIdx++
	}
}

// SelectPrev moves selection up
func (tv *TreeView) SelectPrev() {
	if tv.selectedIdx > 0 {
		tv.selectedIdx--
	}
}

// Expand expands the selected item
func (tv *TreeView) Expand() {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		item := tv.filteredView[tv.selectedIdx].Item
		if !item.Expanded && len(item.Children) > 0 {
			item.Expanded = true
			tv.rebuildView()
		}
	}
}

// Collapse collapses the selected item
func (tv *TreeView) Collapse() {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		item := tv.filteredView[tv.selectedIdx].Item
		if item.Expanded {
			item.Expanded = false
			tv.rebuildView()
		}
	}
}

// Indent indents the selected item (increases nesting level)
func (tv *TreeView) Indent() bool {
	if tv.selectedIdx < 1 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item

	// Get previous item - it becomes the parent
	prevItem := tv.filteredView[tv.selectedIdx-1].Item

	// Remove from current parent
	if current.Parent != nil {
		current.Parent.RemoveChild(current)
	} else {
		// Remove from root
		for idx, item := range tv.items {
			if item.ID == current.ID {
				tv.items = append(tv.items[:idx], tv.items[idx+1:]...)
				break
			}
		}
	}

	// Add to previous item as child
	prevItem.AddChild(current)

	// Expand previous item to show the moved item
	prevItem.Expanded = true

	tv.rebuildView()
	return true
}

// Outdent outdents the selected item (decreases nesting level)
func (tv *TreeView) Outdent() bool {
	if tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item
	if current.Parent == nil {
		return false // Already at top level
	}

	parentParent := current.Parent.Parent
	currentParent := current.Parent

	// Remove from current parent
	currentParent.RemoveChild(current)

	// Add to grandparent or root
	if parentParent != nil {
		parentParent.AddChild(current)
	} else {
		// Add to root level
		current.Parent = nil
		tv.items = append(tv.items, current)
	}

	tv.rebuildView()
	return true
}

// MoveItemDown moves the selected item down in linear order (swaps with next sibling)
func (tv *TreeView) MoveItemDown() bool {
	if tv.selectedIdx >= len(tv.filteredView)-1 || tv.selectedIdx < 0 {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item
	next := tv.filteredView[tv.selectedIdx+1].Item

	// Only swap if they have the same parent and depth
	if tv.filteredView[tv.selectedIdx].Depth != tv.filteredView[tv.selectedIdx+1].Depth {
		return false
	}

	// Get the parent (could be nil for root items)
	parent := current.Parent

	// Find indices in the parent's children array
	var currentIdx, nextIdx int
	var children []*model.Item

	if parent != nil {
		children = parent.Children
	} else {
		children = tv.items
	}

	for idx, child := range children {
		if child.ID == current.ID {
			currentIdx = idx
		}
		if child.ID == next.ID {
			nextIdx = idx
		}
	}

	// Swap only if they are adjacent
	if nextIdx != currentIdx+1 {
		return false
	}

	// Swap items in the slice
	children[currentIdx], children[nextIdx] = children[nextIdx], children[currentIdx]

	tv.rebuildView()
	tv.selectedIdx++ // Move selection to follow the item
	return true
}

// MoveItemUp moves the selected item up in linear order (swaps with previous sibling)
func (tv *TreeView) MoveItemUp() bool {
	if tv.selectedIdx <= 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item
	prev := tv.filteredView[tv.selectedIdx-1].Item

	// Only swap if they have the same parent and depth
	if tv.filteredView[tv.selectedIdx].Depth != tv.filteredView[tv.selectedIdx-1].Depth {
		return false
	}

	// Get the parent (could be nil for root items)
	parent := current.Parent

	// Find indices in the parent's children array
	var currentIdx, prevIdx int
	var children []*model.Item

	if parent != nil {
		children = parent.Children
	} else {
		children = tv.items
	}

	for idx, child := range children {
		if child.ID == current.ID {
			currentIdx = idx
		}
		if child.ID == prev.ID {
			prevIdx = idx
		}
	}

	// Swap only if they are adjacent
	if prevIdx != currentIdx-1 {
		return false
	}

	// Swap items in the slice
	children[currentIdx], children[prevIdx] = children[prevIdx], children[currentIdx]

	tv.rebuildView()
	tv.selectedIdx-- // Move selection to follow the item
	return true
}

// AddItemAfter adds a new item after the selected item
func (tv *TreeView) AddItemAfter(text string) {
	newItem := model.NewItem(text)
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		tv.items = append(tv.items, newItem)
	} else {
		selected := tv.filteredView[tv.selectedIdx].Item
		parent := selected.Parent
		if parent != nil {
			// Find position of selected item in parent's children
			for idx, child := range parent.Children {
				if child.ID == selected.ID {
					// Insert after this position using safe concatenation
					newChildren := make([]*model.Item, 0, len(parent.Children)+1)
					newChildren = append(newChildren, parent.Children[:idx+1]...)
					newItem.Parent = parent
					newChildren = append(newChildren, newItem)
					newChildren = append(newChildren, parent.Children[idx+1:]...)
					parent.Children = newChildren
					break
				}
			}
		} else {
			// Insert at root level using safe concatenation
			for idx, item := range tv.items {
				if item.ID == selected.ID {
					newItems := make([]*model.Item, 0, len(tv.items)+1)
					newItems = append(newItems, tv.items[:idx+1]...)
					newItems = append(newItems, newItem)
					newItems = append(newItems, tv.items[idx+1:]...)
					tv.items = newItems
					break
				}
			}
		}
	}
	tv.rebuildView()
	// Find and select the new item in the filtered view
	for idx, dispItem := range tv.filteredView {
		if dispItem.Item.ID == newItem.ID {
			tv.selectedIdx = idx
			return
		}
	}
}

// AddItemAsChild adds a new item as a child of the selected item
func (tv *TreeView) AddItemAsChild(text string) {
	newItem := model.NewItem(text)
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		selected := tv.filteredView[tv.selectedIdx].Item
		selected.AddChild(newItem)
		selected.Expanded = true
	} else {
		tv.items = append(tv.items, newItem)
	}
	tv.rebuildView()
}

// AddItemBefore adds a new item before the selected item
func (tv *TreeView) AddItemBefore(text string) {
	newItem := model.NewItem(text)
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		tv.items = append(tv.items, newItem)
	} else {
		selected := tv.filteredView[tv.selectedIdx].Item
		parent := selected.Parent
		if parent != nil {
			// Find position of selected item in parent's children
			for idx, child := range parent.Children {
				if child.ID == selected.ID {
					// Insert before this position using safe concatenation
					newChildren := make([]*model.Item, 0, len(parent.Children)+1)
					newChildren = append(newChildren, parent.Children[:idx]...)
					newItem.Parent = parent
					newChildren = append(newChildren, newItem)
					newChildren = append(newChildren, parent.Children[idx:]...)
					parent.Children = newChildren
					break
				}
			}
		} else {
			// Insert at root level using safe concatenation
			for idx, item := range tv.items {
				if item.ID == selected.ID {
					newItems := make([]*model.Item, 0, len(tv.items)+1)
					newItems = append(newItems, tv.items[:idx]...)
					newItems = append(newItems, newItem)
					newItems = append(newItems, tv.items[idx:]...)
					tv.items = newItems
					break
				}
			}
		}
	}
	tv.rebuildView()
	// Find and select the new item in the filtered view
	for idx, dispItem := range tv.filteredView {
		if dispItem.Item.ID == newItem.ID {
			tv.selectedIdx = idx
			return
		}
	}
}

// DeleteSelected removes the selected item
func (tv *TreeView) DeleteSelected() bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	item := tv.filteredView[tv.selectedIdx].Item
	if item.Parent != nil {
		item.Parent.RemoveChild(item)
	} else {
		// Remove from root
		for idx, rootItem := range tv.items {
			if rootItem.ID == item.ID {
				tv.items = append(tv.items[:idx], tv.items[idx+1:]...)
				break
			}
		}
	}

	tv.rebuildView()
	return true
}

// DeleteItem removes a specific item by reference
func (tv *TreeView) DeleteItem(item *model.Item) bool {
	if item == nil {
		return false
	}

	if item.Parent != nil {
		item.Parent.RemoveChild(item)
	} else {
		// Remove from root
		for idx, rootItem := range tv.items {
			if rootItem.ID == item.ID {
				tv.items = append(tv.items[:idx], tv.items[idx+1:]...)
				break
			}
		}
	}

	tv.rebuildView()
	// Move selection back if needed
	if tv.selectedIdx >= len(tv.filteredView) && len(tv.filteredView) > 0 {
		tv.selectedIdx = len(tv.filteredView) - 1
	}
	return true
}

// PasteAfter pastes an item after the selected item
func (tv *TreeView) PasteAfter(item *model.Item) bool {
	if item == nil || len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	selected := tv.filteredView[tv.selectedIdx].Item
	parent := selected.Parent

	if parent != nil {
		// Find position of selected item in parent's children
		for idx, child := range parent.Children {
			if child.ID == selected.ID {
				// Insert after this position
				newChildren := make([]*model.Item, 0, len(parent.Children)+1)
				newChildren = append(newChildren, parent.Children[:idx+1]...)
				item.Parent = parent
				newChildren = append(newChildren, item)
				newChildren = append(newChildren, parent.Children[idx+1:]...)
				parent.Children = newChildren
				tv.rebuildView()
				return true
			}
		}
	} else {
		// Selected item is at root level
		for idx, rootItem := range tv.items {
			if rootItem.ID == selected.ID {
				newItems := make([]*model.Item, 0, len(tv.items)+1)
				newItems = append(newItems, tv.items[:idx+1]...)
				item.Parent = nil
				newItems = append(newItems, item)
				newItems = append(newItems, tv.items[idx+1:]...)
				tv.items = newItems
				tv.rebuildView()
				return true
			}
		}
	}
	return false
}

// PasteBefore pastes an item before the selected item
func (tv *TreeView) PasteBefore(item *model.Item) bool {
	if item == nil || len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	selected := tv.filteredView[tv.selectedIdx].Item
	parent := selected.Parent

	if parent != nil {
		// Find position of selected item in parent's children
		for idx, child := range parent.Children {
			if child.ID == selected.ID {
				// Insert before this position
				newChildren := make([]*model.Item, 0, len(parent.Children)+1)
				newChildren = append(newChildren, parent.Children[:idx]...)
				item.Parent = parent
				newChildren = append(newChildren, item)
				newChildren = append(newChildren, parent.Children[idx:]...)
				parent.Children = newChildren
				tv.rebuildView()
				return true
			}
		}
	} else {
		// Selected item is at root level
		for idx, rootItem := range tv.items {
			if rootItem.ID == selected.ID {
				newItems := make([]*model.Item, 0, len(tv.items)+1)
				newItems = append(newItems, tv.items[:idx]...)
				item.Parent = nil
				newItems = append(newItems, item)
				newItems = append(newItems, tv.items[idx:]...)
				tv.items = newItems
				tv.rebuildView()
				return true
			}
		}
	}
	return false
}

// GetSelected returns the currently selected item
func (tv *TreeView) GetSelected() *model.Item {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		return tv.filteredView[tv.selectedIdx].Item
	}
	return nil
}

// GetSelectedIndex returns the currently selected index
func (tv *TreeView) GetSelectedIndex() int {
	return tv.selectedIdx
}

// SelectItem selects an item by index
func (tv *TreeView) SelectItem(idx int) {
	if idx >= 0 && idx < len(tv.filteredView) {
		tv.selectedIdx = idx
	}
}

// GetSelectedDepth returns the depth (nesting level) of the currently selected item
func (tv *TreeView) GetSelectedDepth() int {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		return tv.filteredView[tv.selectedIdx].Depth
	}
	return 0
}

// GetDisplayItems returns the current display items
func (tv *TreeView) GetDisplayItems() []*displayItem {
	return tv.filteredView
}

// Render renders the tree to the screen
func (tv *TreeView) Render(screen *Screen, startY int, visualAnchor int) {
	defaultStyle := screen.TreeNormalStyle()
	selectedStyle := screen.TreeSelectedStyle()
	visualStyle := screen.TreeVisualSelectionStyle()
	visualCursorStyle := screen.TreeVisualCursorStyle()
	newItemStyle := screen.TreeNewItemStyle()
	screenWidth := screen.GetWidth()

	// Determine visual selection range
	visualStart, visualEnd := -1, -1
	if visualAnchor >= 0 {
		visualStart = visualAnchor
		visualEnd = tv.selectedIdx
		if visualStart > visualEnd {
			visualStart, visualEnd = visualEnd, visualStart
		}
	}

	for idx, dispItem := range tv.filteredView {
		y := startY + idx
		if y >= screen.GetHeight() {
			break
		}

		// Select style based on selection, visual selection, and new item status
		style := defaultStyle
		if dispItem.Item.IsNew && idx != tv.selectedIdx && (visualStart < 0 || idx < visualStart || idx > visualEnd) {
			// Use new item style for new items (dim) when not selected and not in visual range
			style = newItemStyle
		}

		// Check if in visual selection range
		inVisualRange := visualStart >= 0 && idx >= visualStart && idx <= visualEnd

		if inVisualRange {
			if idx == tv.selectedIdx {
				style = visualCursorStyle
			} else {
				style = visualStyle
			}
		} else if idx == tv.selectedIdx {
			style = selectedStyle
		}

		// Build the prefix: 2 spaces per nesting level
		prefix := ""

		// Add indentation for parent levels (2 spaces per level)
		for i := 0; i < dispItem.Depth; i++ {
			prefix += "  "  // 2 spaces per nesting level
		}

		// Draw indentation
		if dispItem.Depth > 0 {
			screen.DrawString(0, y, prefix, style)
		}

		// Always draw an arrow
		// Use different colors for leaf vs expandable nodes
		arrowStyle := screen.TreeLeafArrowStyle()  // Default to leaf (dimmer)
		if len(dispItem.Item.Children) > 0 {
			// For nodes with children, use brighter expandable arrow style
			arrowStyle = screen.TreeExpandableArrowStyle()
		}
		if idx == tv.selectedIdx {
			arrowStyle = selectedStyle  // Use selected style if item is selected
		}

		arrow := "▶"
		if len(dispItem.Item.Children) > 0 && dispItem.Item.Expanded {
			arrow = "▼"
		}

		prefixX := dispItem.Depth * 2
		screen.DrawString(prefixX, y, arrow, arrowStyle)

		// Build the full line
		arrowAndIndent := prefix + arrow
		maxWidth := screenWidth - len(arrowAndIndent)
		if maxWidth < 0 {
			maxWidth = 0
		}

		text := dispItem.Item.Text
		if len(text) > maxWidth {
			text = text[:maxWidth]
		}

		// Draw the text
		textX := prefixX + 2  // Position after the arrow and space
		screen.SetCell(prefixX+1, y, ' ', style)  // Space after arrow
		screen.DrawString(textX, y, text, style)

		// Pad to screen width
		totalLen := textX + len(text)
		for x := totalLen; x < screenWidth; x++ {
			screen.SetCell(x, y, ' ', style)
		}
	}

	// Clear remaining lines
	for y := startY + len(tv.filteredView); y < screen.GetHeight()-1; y++ {
		clearLine := ""
		for i := 0; i < screenWidth; i++ {
			clearLine += " "
		}
		screen.DrawString(0, y, clearLine, defaultStyle)
	}
}

// GetItemCount returns the number of displayed items
func (tv *TreeView) GetItemCount() int {
	return len(tv.filteredView)
}

// GetItemsInRange returns all items in the range from start to end index (inclusive)
func (tv *TreeView) GetItemsInRange(start, end int) []*model.Item {
	if start < 0 || end < 0 || start >= len(tv.filteredView) || end >= len(tv.filteredView) {
		return nil
	}

	if start > end {
		start, end = end, start
	}

	items := make([]*model.Item, 0)
	for i := start; i <= end; i++ {
		if i < len(tv.filteredView) {
			items = append(items, tv.filteredView[i].Item)
		}
	}

	return items
}

// IndentItem indents a specific item (makes it a child of previous item)
func (tv *TreeView) IndentItem(item *model.Item) bool {
	if item == nil {
		return false
	}

	// Find the index of this item in filteredView
	itemIdx := -1
	for idx, dispItem := range tv.filteredView {
		if dispItem.Item.ID == item.ID {
			itemIdx = idx
			break
		}
	}

	if itemIdx < 1 {
		return false // Must have a previous item to indent into
	}

	// Get previous item - it becomes the parent
	prevItem := tv.filteredView[itemIdx-1].Item

	// Remove from current parent
	if item.Parent != nil {
		item.Parent.RemoveChild(item)
	} else {
		// Remove from root
		for idx, rootItem := range tv.items {
			if rootItem.ID == item.ID {
				tv.items = append(tv.items[:idx], tv.items[idx+1:]...)
				break
			}
		}
	}

	// Add to previous item as child
	prevItem.AddChild(item)

	// Expand previous item to show the moved item
	prevItem.Expanded = true

	tv.rebuildView()
	return true
}

// OutdentItem outdents a specific item (decreases nesting level)
func (tv *TreeView) OutdentItem(item *model.Item) bool {
	if item == nil {
		return false
	}

	if item.Parent == nil {
		return false // Already at top level
	}

	parentParent := item.Parent.Parent
	currentParent := item.Parent

	// Remove from current parent
	currentParent.RemoveChild(item)

	// Add to parent's parent or to root
	if parentParent != nil {
		parentParent.AddChild(item)
	} else {
		tv.items = append(tv.items, item)
	}

	tv.rebuildView()
	return true
}

// SelectFirst moves selection to the first item
func (tv *TreeView) SelectFirst() {
	tv.selectedIdx = 0
}

// SelectLast moves selection to the last item
func (tv *TreeView) SelectLast() {
	if len(tv.filteredView) > 0 {
		tv.selectedIdx = len(tv.filteredView) - 1
	}
}

// GetItems returns the root-level items (for saving back to outline)
func (tv *TreeView) GetItems() []*model.Item {
	return tv.items
}
