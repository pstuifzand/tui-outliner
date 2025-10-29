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
	item  *model.Item
	depth int
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
			result = append(result, &displayItem{item: item, depth: depth})
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
		item := tv.filteredView[tv.selectedIdx].item
		if !item.Expanded && len(item.Children) > 0 {
			item.Expanded = true
			tv.rebuildView()
		}
	}
}

// Collapse collapses the selected item
func (tv *TreeView) Collapse() {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		item := tv.filteredView[tv.selectedIdx].item
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

	current := tv.filteredView[tv.selectedIdx].item

	// Get previous item - it becomes the parent
	prevItem := tv.filteredView[tv.selectedIdx-1].item

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

	current := tv.filteredView[tv.selectedIdx].item
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

// AddItemAfter adds a new item after the selected item
func (tv *TreeView) AddItemAfter(text string) {
	newItem := model.NewItem(text)
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		tv.items = append(tv.items, newItem)
	} else {
		selected := tv.filteredView[tv.selectedIdx].item
		parent := selected.Parent
		if parent != nil {
			// Find position of selected item in parent's children
			for idx, child := range parent.Children {
				if child.ID == selected.ID {
					// Insert after this position
					parent.Children = append(parent.Children[:idx+1], append([]*model.Item{newItem}, parent.Children[idx+1:]...)...)
					newItem.Parent = parent
					break
				}
			}
		} else {
			// Insert at root level
			for idx, item := range tv.items {
				if item.ID == selected.ID {
					tv.items = append(tv.items[:idx+1], append([]*model.Item{newItem}, tv.items[idx+1:]...)...)
					break
				}
			}
		}
	}
	tv.rebuildView()
	tv.SelectNext()
}

// AddItemAsChild adds a new item as a child of the selected item
func (tv *TreeView) AddItemAsChild(text string) {
	newItem := model.NewItem(text)
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		selected := tv.filteredView[tv.selectedIdx].item
		selected.AddChild(newItem)
		selected.Expanded = true
	} else {
		tv.items = append(tv.items, newItem)
	}
	tv.rebuildView()
}

// DeleteSelected removes the selected item
func (tv *TreeView) DeleteSelected() bool {
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	item := tv.filteredView[tv.selectedIdx].item
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

// GetSelected returns the currently selected item
func (tv *TreeView) GetSelected() *model.Item {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		return tv.filteredView[tv.selectedIdx].item
	}
	return nil
}

// GetSelectedIndex returns the currently selected index
func (tv *TreeView) GetSelectedIndex() int {
	return tv.selectedIdx
}

// GetDisplayItems returns the current display items
func (tv *TreeView) GetDisplayItems() []*displayItem {
	return tv.filteredView
}

// Render renders the tree to the screen
func (tv *TreeView) Render(screen *Screen, startY int) {
	defaultStyle := DefaultStyle()
	selectedStyle := StyleReverse()
	screenWidth := screen.GetWidth()

	for idx, dispItem := range tv.filteredView {
		y := startY + idx
		if y >= screen.GetHeight() {
			break
		}

		// Select style based on selection and new item status
		style := defaultStyle
		if dispItem.item.IsNew && idx != tv.selectedIdx {
			// Use dim style for new items (light gray) when not selected
			style = StyleDim()
		}
		if idx == tv.selectedIdx {
			style = selectedStyle
		}

		// Build the prefix: 2 spaces per nesting level
		prefix := ""

		// Add indentation for parent levels (2 spaces per level)
		for i := 0; i < dispItem.depth; i++ {
			prefix += "  "  // 2 spaces per nesting level
		}

		// Add tree character for current item
		if len(dispItem.item.Children) > 0 {
			if dispItem.item.Expanded {
				prefix += "▼"  // Arrow only, no space after
			} else {
				prefix += "▶"  // Arrow only, no space after
			}
		} else {
			prefix += " "  // 1 space for leaf nodes (to align with parent's arrow)
		}

		// Build the full line
		maxWidth := screenWidth - len(prefix)
		if maxWidth < 0 {
			maxWidth = 0
		}

		text := dispItem.item.Text
		if len(text) > maxWidth {
			text = text[:maxWidth]
		}

		fullLine := prefix + text

		// Pad to screen width
		for len(fullLine) < screenWidth {
			fullLine += " "
		}

		// Draw the full line with consistent style
		screen.DrawString(0, y, fullLine, style)
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
