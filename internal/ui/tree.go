package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// GetAllItemsRecursive returns all items in a subtree (depth-first)
func GetAllItemsRecursive(item *model.Item) []*model.Item {
	items := []*model.Item{item}
	for _, child := range item.Children {
		items = append(items, GetAllItemsRecursive(child)...)
	}
	return items
}

// TreeView manages the display and navigation of the outline tree
type TreeView struct {
	items          []*model.Item
	selectedIdx    int
	filterText     string
	filteredView   []*displayItem
	viewportOffset int // Index of first visible item in the viewport

	// Hoisting state
	hoistedItem   *model.Item   // Current hoisted node (nil if not hoisted)
	originalItems []*model.Item // Saved root items before hoisting
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

// ScrollPageUp scrolls the viewport up by pageSize items and moves selection
func (tv *TreeView) ScrollPageUp(pageSize int) {
	if pageSize <= 0 {
		pageSize = 1
	}
	// Move selection up by pageSize
	tv.selectedIdx -= pageSize
	if tv.selectedIdx < 0 {
		tv.selectedIdx = 0
	}
	// Adjust viewport offset to show the selected item at the top
	tv.viewportOffset = tv.selectedIdx
}

// ScrollPageDown scrolls the viewport down by pageSize items and moves selection
func (tv *TreeView) ScrollPageDown(pageSize int) {
	if pageSize <= 0 {
		pageSize = 1
	}
	// Move selection down by pageSize
	tv.selectedIdx += pageSize
	maxIdx := len(tv.filteredView) - 1
	if tv.selectedIdx > maxIdx {
		tv.selectedIdx = maxIdx
	}
	// Adjust viewport offset to show the selected item at the bottom of viewport
	tv.viewportOffset = tv.selectedIdx - pageSize + 1
	if tv.viewportOffset < 0 {
		tv.viewportOffset = 0
	}
}

// ensureVisible keeps the selected item within the visible viewport
func (tv *TreeView) ensureVisible() {
	// This would need to know the viewport size, which we'll handle in Render
	// For now, just ensure selectedIdx is valid
	if tv.selectedIdx >= len(tv.filteredView) && len(tv.filteredView) > 0 {
		tv.selectedIdx = len(tv.filteredView) - 1
	}
	if tv.selectedIdx < 0 {
		tv.selectedIdx = 0
	}
}

// Expand expands the selected item and moves to the first child
func (tv *TreeView) Expand() {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		item := tv.filteredView[tv.selectedIdx].Item
		if !item.Expanded && len(item.Children) > 0 {
			item.Expanded = true
			tv.rebuildView()
		}
		// Always move to the first child if the item has children
		if len(item.Children) > 0 && tv.selectedIdx < len(tv.filteredView)-1 {
			tv.selectedIdx++
		}
	}
}

// Collapse collapses the selected item
// Smart behavior: if item has no children, collapses parent instead and moves selection to parent
func (tv *TreeView) Collapse() {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		item := tv.filteredView[tv.selectedIdx].Item

		// If item has children and is expanded, collapse it
		if item.Expanded && len(item.Children) > 0 {
			item.Expanded = false
			tv.rebuildView()
			return
		}

		// If item has no children, try to collapse parent instead and move selection to parent
		if item.Parent != nil && item.Parent.Expanded {
			parent := item.Parent
			parentID := parent.ID

			// Find parent's index BEFORE collapsing to know where to go
			parentIdx := -1
			for idx, dispItem := range tv.filteredView {
				if dispItem.Item.ID == parentID {
					parentIdx = idx
					break
				}
			}

			parent.Expanded = false
			tv.rebuildView()

			// If we found the parent, select it
			if parentIdx >= 0 {
				// Parent is likely still visible, but its children are hidden
				// Find it again in case indices shifted
				for idx, dispItem := range tv.filteredView {
					if dispItem.Item.ID == parentID {
						tv.selectedIdx = idx
						break
					}
				}
			}
		}
	}
}

// CollapseRecursive recursively collapses all items in the tree
func (tv *TreeView) CollapseRecursive() {
	// Collapse all root items and their descendants
	for _, item := range tv.items {
		tv.collapseItemRecursive(item)
	}
	tv.rebuildView()
}

// collapseItemRecursive is a helper that recursively collapses an item and all descendants
func (tv *TreeView) collapseItemRecursive(item *model.Item) {
	if item == nil {
		return
	}
	item.Expanded = false
	for _, child := range item.Children {
		tv.collapseItemRecursive(child)
	}
}

// ExpandRecursive recursively expands all items in the tree
func (tv *TreeView) ExpandRecursive() {
	// Expand all root items and their descendants
	for _, item := range tv.items {
		tv.expandItemRecursive(item)
	}
	tv.rebuildView()
}

// expandItemRecursive is a helper that recursively expands an item and all descendants
func (tv *TreeView) expandItemRecursive(item *model.Item) {
	if item == nil {
		return
	}
	if len(item.Children) > 0 {
		item.Expanded = true
		for _, child := range item.Children {
			tv.expandItemRecursive(child)
		}
	}
}

// CollapseAllChildren collapses all direct children of the selected item
func (tv *TreeView) CollapseAllChildren() {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		item := tv.filteredView[tv.selectedIdx].Item
		for _, child := range item.Children {
			child.Expanded = false
		}
		tv.rebuildView()
	}
}

// CollapseSiblings collapses all siblings of the selected item (items at same level with same parent)
func (tv *TreeView) CollapseSiblings() {
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return
	}

	selected := tv.filteredView[tv.selectedIdx].Item
	if selected.Parent == nil {
		// At root level, collapse all root siblings
		for _, item := range tv.items {
			if item.ID != selected.ID {
				item.Expanded = false
			}
		}
	} else {
		// Collapse all siblings (children of same parent) except selected
		for _, sibling := range selected.Parent.Children {
			if sibling.ID != selected.ID {
				sibling.Expanded = false
			}
		}
	}
	tv.rebuildView()
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
					// When hoisted and we modify the hoisted node's children, update tv.items
					if tv.hoistedItem != nil && parent == tv.hoistedItem {
						tv.items = newChildren
					}
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
					// When hoisted and we modify the hoisted node's children, update tv.items
					if tv.hoistedItem != nil && parent == tv.hoistedItem {
						tv.items = newChildren
					}
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
		parent := item.Parent
		item.Parent.RemoveChild(item)
		// When hoisted and we delete from the hoisted node's children, update tv.items
		if tv.hoistedItem != nil && parent == tv.hoistedItem {
			tv.items = parent.Children
		}
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
				// When hoisted and we modify the hoisted node's children, update tv.items
				if tv.hoistedItem != nil && parent == tv.hoistedItem {
					tv.items = newChildren
				}
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
				// When hoisted and we modify the hoisted node's children, update tv.items
				if tv.hoistedItem != nil && parent == tv.hoistedItem {
					tv.items = newChildren
				}
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
	tv.RenderWithSearchQuery(screen, startY, visualAnchor, "", nil)
}

// RenderWithSearchQuery renders the tree with optional search query highlighting
func (tv *TreeView) RenderWithSearchQuery(screen *Screen, startY int, visualAnchor int, searchQuery string, currentMatchItem *model.Item) {
	defaultStyle := screen.TreeNormalStyle()
	selectedStyle := screen.TreeSelectedStyle()
	visualStyle := screen.TreeVisualSelectionStyle()
	visualCursorStyle := screen.TreeVisualCursorStyle()
	newItemStyle := screen.TreeNewItemStyle()
	highlightStyle := screen.SearchHighlightStyle()
	screenWidth := screen.GetWidth()
	screenHeight := screen.GetHeight()

	// Get background color for adding to text styles
	bgColor := screen.Theme.Colors.Background

	// Add background to non-selected styles
	defaultStyle = defaultStyle.Background(bgColor)
	newItemStyle = newItemStyle.Background(bgColor)

	// Calculate available viewport height Reserve 1 line for status bar
	viewportHeight := max(screenHeight-startY-1, 1)

	// Ensure viewport offset keeps selected item visible
	if tv.selectedIdx < tv.viewportOffset {
		tv.viewportOffset = tv.selectedIdx
	} else if tv.selectedIdx >= tv.viewportOffset+viewportHeight {
		tv.viewportOffset = tv.selectedIdx - viewportHeight + 1
	}

	// Clamp viewport offset
	maxOffset := max(len(tv.filteredView)-viewportHeight, 0)
	if tv.viewportOffset > maxOffset {
		tv.viewportOffset = maxOffset
	}
	if tv.viewportOffset < 0 {
		tv.viewportOffset = 0
	}

	// Determine visual selection range (adjust for viewport offset)
	visualStart, visualEnd := -1, -1
	if visualAnchor >= 0 {
		visualStart = visualAnchor
		visualEnd = tv.selectedIdx
		if visualStart > visualEnd {
			visualStart, visualEnd = visualEnd, visualStart
		}
	}

	// Render items starting from viewportOffset
	screenY := startY
	for i := tv.viewportOffset; i < len(tv.filteredView) && screenY < screenHeight-1; i++ {
		dispItem := tv.filteredView[i]
		idx := i // Keep track of actual index in filteredView for selection/visual comparisons
		y := screenY

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

		// Prepare arrow style with background color
		leafArrowStyle := screen.TreeLeafArrowStyle().Background(bgColor)
		expandableArrowStyle := screen.TreeExpandableArrowStyle().Background(bgColor)

		// Build the prefix: 2 spaces per nesting level
		prefix := ""

		// Add indentation for parent levels (2 spaces per level)
		for i := 0; i < dispItem.Depth; i++ {
			prefix += "  " // 2 spaces per nesting level
		}

		// Draw indentation
		if dispItem.Depth > 0 {
			screen.DrawString(0, y, prefix, style)
		}

		// Always draw an arrow
		// Use different colors for leaf vs expandable nodes
		arrowStyle := leafArrowStyle // Default to leaf (dimmer)
		if len(dispItem.Item.Children) > 0 {
			// For nodes with children, use brighter expandable arrow style
			arrowStyle = expandableArrowStyle
		}
		if idx == tv.selectedIdx {
			arrowStyle = selectedStyle // Use selected style if item is selected
		}

		arrow := "▶"
		if len(dispItem.Item.Children) > 0 && dispItem.Item.Expanded {
			arrow = "▼"
		}

		prefixX := dispItem.Depth * 3
		screen.DrawString(prefixX, y, arrow, arrowStyle)

		// Draw attribute indicator or space to maintain alignment
		indicatorStyle := screen.TreeAttributeIndicatorStyle()
		if idx == tv.selectedIdx {
			indicatorStyle = selectedStyle // Use selected style if item is selected
		}

		hasAttributes := dispItem.Item.Metadata != nil && len(dispItem.Item.Metadata.Attributes) > 0
		if hasAttributes {
			screen.SetCell(prefixX+1, y, '●', indicatorStyle) // Filled circle for items with attributes
		} else {
			screen.SetCell(prefixX+1, y, ' ', style) // Space for items without attributes
		}

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
		textX := prefixX + 3                     // Position after the arrow, indicator, and space
		screen.SetCell(prefixX+2, y, ' ', style) // Space after indicator

		// Highlight search matches in the text only if this is the current match
		if searchQuery != "" && currentMatchItem != nil && dispItem.Item == currentMatchItem {
			tv.drawTextWithHighlight(screen, textX, y, text, style, highlightStyle, searchQuery)
		} else {
			screen.DrawString(textX, y, text, style)
		}

		// Pad to screen width with background color
		totalLen := textX + len(text)
		bgStyle := screen.BackgroundStyle()
		for x := totalLen; x < screenWidth; x++ {
			screen.SetCell(x, y, ' ', bgStyle)
		}

		screenY++ // Move to next screen line
	}

	// Clear remaining lines with background color
	bgStyle := screen.BackgroundStyle()
	for y := screenY; y < screen.GetHeight()-1; y++ {
		clearLine := ""
		for i := 0; i < screenWidth; i++ {
			clearLine += " "
		}
		screen.DrawString(0, y, clearLine, bgStyle)
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

// SelectParent moves selection to the parent of the current item
func (tv *TreeView) SelectParent() bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item
	if current.Parent == nil {
		return false // Already at root level
	}

	// Find parent in the filtered view
	for idx, dispItem := range tv.filteredView {
		if dispItem.Item.ID == current.Parent.ID {
			tv.selectedIdx = idx
			return true
		}
	}
	return false
}

// ExpandParents expands all parent nodes of the given item so it becomes visible
func (tv *TreeView) ExpandParents(item *model.Item) {
	if item == nil {
		return
	}
	// Walk up the parent chain and expand each parent
	current := item.Parent
	for current != nil {
		current.Expanded = true
		current = current.Parent
	}
	tv.rebuildView()
}

// SelectNextSibling moves selection to the next sibling
func (tv *TreeView) SelectNextSibling() bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item
	currentDepth := tv.filteredView[tv.selectedIdx].Depth

	// Look for next item at the same depth with the same parent
	for i := tv.selectedIdx + 1; i < len(tv.filteredView); i++ {
		dispItem := tv.filteredView[i]

		// Stop if we go to a shallower depth (we've left this branch)
		if dispItem.Depth < currentDepth {
			return false
		}

		// Found next sibling at same depth
		if dispItem.Depth == currentDepth && dispItem.Item.Parent == current.Parent {
			tv.selectedIdx = i
			return true
		}
	}
	return false
}

// SelectPrevSibling moves selection to the previous sibling
func (tv *TreeView) SelectPrevSibling() bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item
	currentDepth := tv.filteredView[tv.selectedIdx].Depth

	// Look for previous item at the same depth with the same parent
	for i := tv.selectedIdx - 1; i >= 0; i-- {
		dispItem := tv.filteredView[i]

		// Skip items at greater depth (children/descendants)
		if dispItem.Depth > currentDepth {
			continue
		}

		// Stop if we go to a shallower depth (we've left this branch)
		if dispItem.Depth < currentDepth {
			return false
		}

		// Found previous sibling at same depth
		if dispItem.Depth == currentDepth && dispItem.Item.Parent == current.Parent {
			tv.selectedIdx = i
			return true
		}
	}
	return false
}

// FindNextDateItem finds the next item with a date attribute, starting after current selection
func (tv *TreeView) FindNextDateItem() bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	// Search forward from current position
	for i := tv.selectedIdx + 1; i < len(tv.filteredView); i++ {
		item := tv.filteredView[i].Item
		if item.Metadata != nil {
			// Check for date attribute
			if _, hasDate := item.Metadata.Attributes["date"]; hasDate {
				tv.selectedIdx = i
				return true
			}
		}
	}
	return false
}

// FindPrevDateItem finds the previous item with a date attribute, starting before current selection
func (tv *TreeView) FindPrevDateItem() bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	// Search backward from current position
	for i := tv.selectedIdx - 1; i >= 0; i-- {
		item := tv.filteredView[i].Item
		if item.Metadata != nil {
			// Check for date attribute
			if _, hasDate := item.Metadata.Attributes["date"]; hasDate {
				tv.selectedIdx = i
				return true
			}
		}
	}
	return false
}

// FindNextItemWithDateInterval finds next item within specified interval (day, week, month, year)
// Interval: "day", "week", "month", "year"
func (tv *TreeView) FindNextItemWithDateInterval(interval string) bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	now := time.Now()
	var targetStart, targetEnd time.Time

	switch interval {
	case "day":
		targetStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		targetEnd = targetStart.AddDate(0, 0, 1)
	case "week":
		// Start of current week (Monday)
		days := int(now.Weekday()) - 1
		if days < 0 {
			days = 6
		}
		targetStart = now.AddDate(0, 0, -days)
		targetStart = time.Date(targetStart.Year(), targetStart.Month(), targetStart.Day(), 0, 0, 0, 0, targetStart.Location())
		targetEnd = targetStart.AddDate(0, 0, 7)
	case "month":
		targetStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		targetEnd = targetStart.AddDate(0, 1, 0)
	case "year":
		targetStart = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		targetEnd = targetStart.AddDate(1, 0, 0)
	default:
		return false
	}

	// Search forward from current position
	for i := tv.selectedIdx + 1; i < len(tv.filteredView); i++ {
		item := tv.filteredView[i].Item
		if item.Metadata != nil {
			// Check for date attribute (in YYYY-MM-DD format)
			if dateStr, hasDate := item.Metadata.Attributes["date"]; hasDate {
				if dateTime, err := time.Parse("2006-01-02", dateStr); err == nil {
					// Normalize to midnight for comparison
					dateTime = time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, dateTime.Location())
					if dateTime.After(targetStart) && dateTime.Before(targetEnd) {
						tv.selectedIdx = i
						return true
					}
				}
			}
		}
	}
	return false
}

// FindPrevItemWithDateInterval finds previous item within specified interval (day, week, month, year)
func (tv *TreeView) FindPrevItemWithDateInterval(interval string) bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	now := time.Now()
	var targetStart, targetEnd time.Time

	switch interval {
	case "day":
		targetStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		targetEnd = targetStart.AddDate(0, 0, 1)
	case "week":
		// Start of current week (Monday)
		days := int(now.Weekday()) - 1
		if days < 0 {
			days = 6
		}
		targetStart = now.AddDate(0, 0, -days)
		targetStart = time.Date(targetStart.Year(), targetStart.Month(), targetStart.Day(), 0, 0, 0, 0, targetStart.Location())
		targetEnd = targetStart.AddDate(0, 0, 7)
	case "month":
		targetStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		targetEnd = targetStart.AddDate(0, 1, 0)
	case "year":
		targetStart = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		targetEnd = targetStart.AddDate(1, 0, 0)
	default:
		return false
	}

	// Search backward from current position
	for i := tv.selectedIdx - 1; i >= 0; i-- {
		item := tv.filteredView[i].Item
		if item.Metadata != nil {
			// Check for date attribute (in YYYY-MM-DD format)
			if dateStr, hasDate := item.Metadata.Attributes["date"]; hasDate {
				if dateTime, err := time.Parse("2006-01-02", dateStr); err == nil {
					// Normalize to midnight for comparison
					dateTime = time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, dateTime.Location())
					if dateTime.After(targetStart) && dateTime.Before(targetEnd) {
						tv.selectedIdx = i
						return true
					}
				}
			}
		}
	}
	return false
}

// GetItems returns the root-level items (for saving back to outline)
func (tv *TreeView) GetItems() []*model.Item {
	// When hoisted, return original items to ensure full tree is saved
	if tv.hoistedItem != nil {
		return tv.originalItems
	}
	return tv.items
}

// Hoist makes the selected item the temporary root, showing only its children
func (tv *TreeView) Hoist() bool {
	selected := tv.GetSelected()
	if selected == nil || len(selected.Children) == 0 {
		return false
	}

	// Save original root items
	tv.originalItems = tv.items

	// Set hoisted item and replace items with its children
	tv.hoistedItem = selected
	tv.items = selected.Children

	// Rebuild view and reset selection to first child
	tv.selectedIdx = 0
	tv.viewportOffset = 0
	tv.rebuildView()

	return true
}

// Unhoist returns to the full tree view
func (tv *TreeView) Unhoist() bool {
	if tv.hoistedItem == nil {
		return false
	}

	// Restore original items
	tv.items = tv.originalItems
	tv.hoistedItem = nil
	tv.originalItems = nil

	// Rebuild view
	tv.rebuildView()

	return true
}

// IsHoisted returns whether we're currently in hoisted mode
func (tv *TreeView) IsHoisted() bool {
	return tv.hoistedItem != nil
}

// GetHoistedItem returns the current hoist root (nil if not hoisted)
func (tv *TreeView) GetHoistedItem() *model.Item {
	return tv.hoistedItem
}

// GetHoistBreadcrumbs returns the full path to the hoisted item as a breadcrumb string
// e.g., "Project A > Development > Build frontend"
// Returns empty string if not hoisted
func (tv *TreeView) GetHoistBreadcrumbs() string {
	if tv.hoistedItem == nil {
		return ""
	}

	// Build path from root to hoisted item by traversing parents
	var path []*model.Item
	current := tv.hoistedItem
	for current != nil {
		path = append(path, current)
		current = current.Parent
	}

	// Reverse to get root-to-leaf order
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// Build breadcrumb string with separator
	breadcrumbs := make([]string, 0, len(path))
	for _, item := range path {
		breadcrumbs = append(breadcrumbs, item.Text)
	}

	// Join with " > " separator
	result := ""
	for i, crumb := range breadcrumbs {
		if i > 0 {
			result += " > "
		}
		result += crumb
	}

	return result
}

// drawTextWithHighlight draws text with highlighted search matches
// It draws the text while highlighting case-insensitive substring matches of the search query
func (tv *TreeView) drawTextWithHighlight(screen *Screen, x int, y int, text string, defaultStyle tcell.Style, highlightStyle tcell.Style, searchQuery string) {
	if searchQuery == "" {
		screen.DrawString(x, y, text, defaultStyle)
		return
	}

	// Convert to lowercase for case-insensitive comparison
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(searchQuery)

	// Find all matches
	currentX := x
	lastIdx := 0

	for {
		// Find next match
		matchIdx := strings.Index(lowerText[lastIdx:], lowerQuery)
		if matchIdx == -1 {
			// No more matches, draw remaining text
			if lastIdx < len(text) {
				screen.DrawString(currentX, y, text[lastIdx:], defaultStyle)
			}
			break
		}

		// Adjust matchIdx to be relative to the full string
		matchIdx += lastIdx

		// Draw text before match with default style
		if matchIdx > lastIdx {
			beforeText := text[lastIdx:matchIdx]
			screen.DrawString(currentX, y, beforeText, defaultStyle)
			currentX += len(beforeText)
		}

		// Draw matched text with highlight style
		matchedText := text[matchIdx : matchIdx+len(searchQuery)]
		screen.DrawString(currentX, y, matchedText, highlightStyle)
		currentX += len(searchQuery)

		// Move to next position after this match
		lastIdx = matchIdx + len(searchQuery)
	}
}
