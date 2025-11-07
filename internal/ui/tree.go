package ui

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/config"
	"github.com/pstuifzand/tui-outliner/internal/links"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// debugLogger is initialized once to log move operations
var debugLogger *log.Logger

func init() {
	debugFile, err := os.OpenFile("/tmp/tuo-move-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fallback to stderr if we can't open the debug file
		debugLogger = log.New(os.Stderr, "[TUO] ", log.LstdFlags)
	} else {
		debugLogger = log.New(debugFile, "[TUO] ", log.LstdFlags)
	}
}

// GetAllItemsRecursive returns all items in a subtree (depth-first)
func GetAllItemsRecursive(item *model.Item) []*model.Item {
	items := []*model.Item{item}
	for _, child := range item.Children {
		items = append(items, GetAllItemsRecursive(child)...)
	}
	return items
}

// CalculateProgressFromChildren calculates completion metrics from children with type=todo
// Returns: total todo children, count with last status (done), count with middle statuses (doing)
func CalculateProgressFromChildren(item *model.Item, todoStatuses []string) (total, done, doing int) {
	if len(todoStatuses) == 0 {
		return 0, 0, 0
	}

	lastStatus := todoStatuses[len(todoStatuses)-1]

	for _, child := range item.Children {
		// Only count children with type=todo
		if child.Metadata == nil || child.Metadata.Attributes == nil {
			continue
		}
		if childType, ok := child.Metadata.Attributes["type"]; !ok || childType != "todo" {
			continue
		}

		total++
		status := child.Metadata.Attributes["status"]

		if status == lastStatus {
			done++
		} else if status != todoStatuses[0] {
			// Status is not first and not last, so it's in progress
			doing++
		}
	}

	return total, done, doing
}

// ProgressBarBlock represents a single block in the progress bar with its status
type ProgressBarBlock struct {
	Status string // Status of the child (todo, doing, done, etc)
}

// RenderProgressBar generates progress bar blocks for an item with todo children
// Returns a slice of ProgressBarBlocks, one per todo child, in order
// Only returns blocks if item has type=todo and has todo children
func RenderProgressBar(item *model.Item, todoStatuses []string) []ProgressBarBlock {
	if item.Metadata == nil || item.Metadata.Attributes == nil {
		return nil
	}

	itemType, hasType := item.Metadata.Attributes["type"]
	if !hasType || itemType != "todo" {
		return nil
	}

	var blocks []ProgressBarBlock
	for _, child := range item.Children {
		// Only include blocks for children with type=todo
		if child.Metadata == nil || child.Metadata.Attributes == nil {
			continue
		}
		childType, ok := child.Metadata.Attributes["type"]
		if !ok || childType != "todo" {
			continue
		}

		status := child.Metadata.Attributes["status"]
		blocks = append(blocks, ProgressBarBlock{Status: status})
	}

	if len(blocks) == 0 {
		return nil
	}
	return blocks
}

// UpdateParentStatusIfTodo updates parent item's status if it has type=todo
// Implements progressive status matching based on children's statuses
// Recursively updates ancestors that also have type=todo
func UpdateParentStatusIfTodo(item *model.Item, todoStatuses []string) {
	if item.Parent == nil || len(todoStatuses) == 0 {
		return
	}

	parent := item.Parent

	// Check if parent has type=todo
	if parent.Metadata == nil || parent.Metadata.Attributes == nil {
		return
	}
	parentType, hasType := parent.Metadata.Attributes["type"]
	if !hasType || parentType != "todo" {
		return
	}

	// Calculate progress from children
	total, done, doing := CalculateProgressFromChildren(parent, todoStatuses)

	if total == 0 {
		// No todo children, don't update
		return
	}

	firstStatus := todoStatuses[0]
	lastStatus := todoStatuses[len(todoStatuses)-1]
	middleStatus := firstStatus
	if len(todoStatuses) > 1 {
		middleStatus = todoStatuses[1]
	}

	var newStatus string

	// Simple status matching logic based on all children
	if done == total {
		// All children are done
		newStatus = lastStatus
	} else if done == 0 && doing == 0 {
		// All children are todo (no progress)
		newStatus = firstStatus
	} else {
		// Mix of statuses
		newStatus = middleStatus
	}

	// Update parent status if changed
	if parent.Metadata.Attributes["status"] != newStatus {
		parent.Metadata.Attributes["status"] = newStatus
		parent.Metadata.Modified = time.Now()
	}

	// Calculate and store progress metrics
	progressPct := 0
	if total > 0 {
		progressPct = (done * 100) / total
	}
	parent.Metadata.Attributes["progress_count"] = fmt.Sprintf("%d/%d", done, total)
	parent.Metadata.Attributes["progress_pct"] = fmt.Sprintf("%d%%", progressPct)

	// Recursively update grandparent if it's also a todo
	UpdateParentStatusIfTodo(parent, todoStatuses)
}

// TreeView manages the display and navigation of the outline tree
type TreeView struct {
	items          []*model.Item
	selectedIdx    int // Index of currently selected item (in terms of items, not display lines)
	filterText     string
	filteredView   []*displayItem
	displayLines   []*DisplayLine // Multi-line aware display for rendering
	viewportOffset int            // Index of first visible display line in the viewport
	maxWidth       int            // Maximum width for text wrapping (0 = no wrapping)

	// Hoisting state
	hoistedItem   *model.Item   // Current hoisted node (nil if not hoisted)
	originalItems []*model.Item // Saved root items before hoisting
}

type displayItem struct {
	Item             *model.Item
	Depth            int
	IsVirtual        bool          // True if this is a virtual child reference
	OriginalItem     *model.Item   // Points to the original if IsVirtual (for virtual references)
	SearchNodeParent *model.Item   // If IsVirtual, points to the search node that owns this virtual reference
	VirtualAncestors []*model.Item // Chain of virtual ancestors for nested virtual items
}

// DisplayLine represents a single visual line in the tree view
// Multiple DisplayLines can belong to the same Item if it has multiple lines of text
type DisplayLine struct {
	Item             *model.Item // The underlying item
	TextLineIndex    int         // Which line within the item's text (0-based, split by \n)
	TextLine         string      // The actual text to display for this line
	ItemStartLine    bool        // True if this is the first line of the item (shows indent/arrow/metadata)
	IsWrapped        bool        // True if this is a wrapped continuation of a long line
	Depth            int
	IsVirtual        bool
	OriginalItem     *model.Item
	SearchNodeParent *model.Item
	VirtualAncestors []*model.Item
	ParentDisplayItem *displayItem // Reference to parent displayItem (for comparing selected items)
}

// NewTreeView creates a new TreeView
func NewTreeView(items []*model.Item) *TreeView {
	tv := &TreeView{
		items:       items,
		selectedIdx: 0,
	}
	tv.RebuildView()
	return tv
}

// SetItems updates the tree view's items and rebuilds the view
func (tv *TreeView) SetItems(items []*model.Item) {
	tv.items = items
	tv.RebuildView()
}

// wrapTextAtWidth wraps a single text line to the specified display width
// Returns a slice of wrapped text portions
// Properly handles multi-byte Unicode characters
func wrapTextAtWidth(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	if StringWidth(text) <= maxWidth {
		return []string{text}
	}

	var result []string
	remaining := text

	for StringWidth(remaining) > maxWidth {
		// Use CalculateBreakPoint to find proper break location
		// This handles word boundaries and multi-byte characters correctly
		byteIdx, _ := CalculateBreakPoint(remaining, maxWidth)

		if byteIdx <= 0 {
			// Edge case: even first character exceeds maxWidth
			// Just take the first rune
			runes := []rune(remaining)
			if len(runes) > 0 {
				result = append(result, string(runes[0]))
				remaining = string(runes[1:])
			} else {
				break
			}
		} else {
			// Slice safely at byte boundary
			line := remaining[:byteIdx]
			result = append(result, strings.TrimRight(line, " "))

			// Skip spaces at the beginning of next line
			remaining = remaining[byteIdx:]
			remaining = strings.TrimLeft(remaining, " ")
		}
	}

	// Add any remaining text
	if len(remaining) > 0 {
		result = append(result, remaining)
	}

	return result
}

// buildDisplayLines converts a list of displayItems into display lines,
// expanding multi-line items into multiple DisplayLine entries with word wrapping
// maxWidth specifies the maximum width for text wrapping (0 = no wrapping)
func (tv *TreeView) buildDisplayLines(displayItems []*displayItem, maxWidth int) []*DisplayLine {
	var lines []*DisplayLine
	for _, dispItem := range displayItems {
		// Split item text by hard newlines first
		textLines := strings.Split(dispItem.Item.Text, "\n")
		for lineIdx, textLine := range textLines {
			// Apply word wrapping if maxWidth is specified
			var wrappedLines []string
			if maxWidth > 0 {
				wrappedLines = wrapTextAtWidth(textLine, maxWidth)
			} else {
				wrappedLines = []string{textLine}
			}

			// Create display lines for each wrapped portion
			for wrapIdx, wrappedText := range wrappedLines {
				isFirstLine := lineIdx == 0 && wrapIdx == 0
				isWrapped := wrapIdx > 0 // True if this is a wrapped continuation

				line := &DisplayLine{
					Item:              dispItem.Item,
					TextLineIndex:     lineIdx,
					TextLine:          wrappedText,
					ItemStartLine:     isFirstLine,
					IsWrapped:         isWrapped,
					Depth:             dispItem.Depth,
					IsVirtual:         dispItem.IsVirtual,
					OriginalItem:      dispItem.OriginalItem,
					SearchNodeParent:  dispItem.SearchNodeParent,
					VirtualAncestors:  dispItem.VirtualAncestors,
					ParentDisplayItem: dispItem,
				}
				lines = append(lines, line)
			}
		}
	}
	return lines
}

// getFirstDisplayLineForItem returns the index of the first display line for a given item
// Returns -1 if item not found
func (tv *TreeView) getFirstDisplayLineForItem(item *model.Item) int {
	if item == nil {
		return -1
	}
	for idx, line := range tv.displayLines {
		if line.Item.ID == item.ID && line.ItemStartLine {
			return idx
		}
	}
	return -1
}

// getLastDisplayLineForItem returns the index of the last display line for a given item
// Returns -1 if item not found
func (tv *TreeView) getLastDisplayLineForItem(item *model.Item) int {
	if item == nil {
		return -1
	}
	lastIdx := -1
	for idx, line := range tv.displayLines {
		if line.Item.ID == item.ID {
			lastIdx = idx
		} else if lastIdx >= 0 {
			// We've moved past this item's lines
			break
		}
	}
	return lastIdx
}

// GetItemFromDisplayLine returns the item index in filteredView for a given display line
// Used for converting display line clicks to item selection
func (tv *TreeView) GetItemFromDisplayLine(displayLineIdx int) int {
	if displayLineIdx < 0 || displayLineIdx >= len(tv.displayLines) {
		return -1
	}
	item := tv.displayLines[displayLineIdx].Item
	for idx, dispItem := range tv.filteredView {
		if dispItem.Item.ID == item.ID {
			return idx
		}
	}
	return -1
}

// RebuildView rebuilds the filtered/display view
func (tv *TreeView) RebuildView() {
	tv.filteredView = tv.buildDisplayItems(tv.items, 0)
	tv.displayLines = tv.buildDisplayLines(tv.filteredView, tv.maxWidth)
	if tv.selectedIdx >= len(tv.filteredView) && len(tv.filteredView) > 0 {
		tv.selectedIdx = len(tv.filteredView) - 1
	}
	tv.viewportOffset = 0 // Reset viewport when rebuilding
}

// SetMaxWidth sets the maximum width for text wrapping and rebuilds the view
// Pass 0 to disable wrapping (use truncation instead)
func (tv *TreeView) SetMaxWidth(width int) {
	if width < 0 {
		width = 0
	}
	if tv.maxWidth != width {
		tv.maxWidth = width
		tv.RebuildView()
	}
}

func (tv *TreeView) buildDisplayItems(items []*model.Item, depth int) []*displayItem {
	return tv.buildDisplayItemsInternal(items, depth, false, nil, nil, nil)
}

func (tv *TreeView) buildDisplayItemsInternal(items []*model.Item, depth int, parentIsVirtual bool, searchNodeParent *model.Item, virtualAncestors []*model.Item, directVirtualParent *model.Item) []*displayItem {
	var result []*displayItem
	for _, item := range items {
		// Check if item should be displayed
		shouldDisplay := false
		if depth == 0 {
			shouldDisplay = true
		} else if parentIsVirtual {
			// Virtual children are always displayed (parent is already expanded)
			shouldDisplay = true
		} else if item.Parent != nil && item.Parent.Expanded {
			// Real children are displayed if parent is expanded
			shouldDisplay = true
		}

		if shouldDisplay {
			// If parent was virtual, this item is also shown as virtual
			ancestors := virtualAncestors
			if parentIsVirtual && directVirtualParent != nil {
				// Build the ancestor chain for this virtual item
				ancestors = append([]*model.Item{}, virtualAncestors...)
				ancestors = append(ancestors, directVirtualParent)
			}

			result = append(result, &displayItem{
				Item:             item,
				Depth:            depth,
				IsVirtual:        parentIsVirtual,
				OriginalItem:     item,
				SearchNodeParent: searchNodeParent,
				VirtualAncestors: ancestors,
			})

			// Determine if this item should show its children
			var shouldShowChildren bool
			if parentIsVirtual && searchNodeParent != nil {
				// For virtual children, ONLY check if it's collapsed in the search node's display
				// Do NOT check the original item's Expanded state
				shouldShowChildren = !searchNodeParent.IsVirtualChildCollapsed(item.ID)
			} else {
				// For real children (not virtual), use the original item's Expanded state
				shouldShowChildren = item.Expanded
			}

			if shouldShowChildren {
				// Add real children
				if len(item.Children) > 0 {
					result = append(result, tv.buildDisplayItemsInternal(item.Children, depth+1, parentIsVirtual, searchNodeParent, ancestors, item)...)
				}
				// Add virtual children (only if this item is not already virtual)
				if !parentIsVirtual {
					virtualChildren := item.GetVirtualChildren()
					if len(virtualChildren) > 0 {
						// Update searchNodeParent if this is a search node
						currentSearchNode := searchNodeParent
						if item.IsSearchNode() {
							currentSearchNode = item
						}

						for _, virtualChild := range virtualChildren {
							result = append(result, tv.buildDisplayItemsInternal([]*model.Item{virtualChild}, depth+1, true, currentSearchNode, nil, virtualChild)...)
						}
					}
				}
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
	// Move selection up by pageSize items
	tv.selectedIdx -= pageSize
	if tv.selectedIdx < 0 {
		tv.selectedIdx = 0
	}
	// Adjust viewport offset to show the selected item's first line at the top
	selectedItem := tv.GetSelected()
	if selectedItem != nil {
		firstLineIdx := tv.getFirstDisplayLineForItem(selectedItem)
		if firstLineIdx >= 0 {
			tv.viewportOffset = firstLineIdx
		}
	}
}

// ScrollPageDown scrolls the viewport down by pageSize items and moves selection
func (tv *TreeView) ScrollPageDown(pageSize int) {
	if pageSize <= 0 {
		pageSize = 1
	}
	// Move selection down by pageSize items
	tv.selectedIdx += pageSize
	maxIdx := len(tv.filteredView) - 1
	if tv.selectedIdx > maxIdx {
		tv.selectedIdx = maxIdx
	}
	// Adjust viewport offset to show the selected item's last line at the bottom of viewport
	selectedItem := tv.GetSelected()
	if selectedItem != nil {
		lastLineIdx := tv.getLastDisplayLineForItem(selectedItem)
		if lastLineIdx >= 0 {
			tv.viewportOffset = lastLineIdx - pageSize + 1
			if tv.viewportOffset < 0 {
				tv.viewportOffset = 0
			}
		}
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
// For virtual children, clears the collapsed flag in the search node without affecting the original item
// Never modifies the actual item's Expanded state for virtual items
func (tv *TreeView) Expand(move bool) {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		dispItem := tv.filteredView[tv.selectedIdx]
		item := dispItem.Item

		// Special handling for virtual children: expand only in the search node's view
		// Do NOT modify the actual item's Expanded state
		if dispItem.IsVirtual && dispItem.SearchNodeParent != nil {
			// Clear the collapsed flag for this virtual item in the search node's display
			dispItem.SearchNodeParent.SetVirtualChildCollapsed(item.ID, false)
			tv.RebuildView()
			// Move to first child if requested
			if move && (len(item.Children) > 0 || len(item.GetVirtualChildren()) > 0) && tv.selectedIdx < len(tv.filteredView)-1 {
				tv.selectedIdx++
			}
			return
		}

		// For non-virtual items, expand normally
		hasChildren := len(item.Children) > 0 || len(item.GetVirtualChildren()) > 0
		if !item.Expanded && hasChildren {
			item.Expanded = true
			tv.RebuildView()
		}
		// Always move to the first child if the item has children
		if move && hasChildren && tv.selectedIdx < len(tv.filteredView)-1 {
			tv.selectedIdx++
		}
	}
}

// Collapse collapses the selected item
// Smart behavior: if item has no children, collapses parent instead and moves selection to parent
// For virtual children (search results), marks them as collapsed in the search node without
// affecting the original item's expand state
func (tv *TreeView) Collapse() {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		dispItem := tv.filteredView[tv.selectedIdx]
		item := dispItem.Item

		// Check if this item has children
		hasChildren := len(item.Children) > 0 || len(item.GetVirtualChildren()) > 0

		// Special handling for virtual children: collapse only in the search node's view
		if dispItem.IsVirtual && dispItem.SearchNodeParent != nil {
			// If virtual item has children, collapse it in the search node
			if hasChildren {
				dispItem.SearchNodeParent.SetVirtualChildCollapsed(item.ID, true)
				tv.RebuildView()
				return
			}
			// If virtual item has no children, try to collapse its virtual parent instead
			// Find the virtual parent by looking at VirtualAncestors
			if len(dispItem.VirtualAncestors) > 0 {
				// There's a virtual parent, collapse it in the search node
				virtualParent := dispItem.VirtualAncestors[len(dispItem.VirtualAncestors)-1]
				dispItem.SearchNodeParent.SetVirtualChildCollapsed(virtualParent.ID, true)
				tv.RebuildView()

				// Move selection to the virtual parent
				virtualParentID := virtualParent.ID
				for idx, dItem := range tv.filteredView {
					if dItem.Item.ID == virtualParentID {
						tv.selectedIdx = idx
						break
					}
				}
				return
			}
			// If no virtual parent, this is a direct virtual child with no children
			// Collapse its parent (the search node's parent in the real tree)
			if item.Parent != nil && item.Parent.Expanded {
				item.Parent.Expanded = false
				tv.RebuildView()
				return
			}
		}

		// If item has children and is expanded, collapse it
		if item.Expanded && hasChildren {
			item.Expanded = false
			tv.RebuildView()
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
			tv.RebuildView()

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
	tv.RebuildView()
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
	tv.RebuildView()
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
		tv.RebuildView()
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
	tv.RebuildView()
}

// findPreviousSiblingForIndent finds an item at the same depth as the given index
// This is the appropriate target for indenting (will become the parent)
func (tv *TreeView) findPreviousSiblingForIndent(currentIdx int) *model.Item {
	if currentIdx < 1 {
		return nil
	}

	currentDepth := tv.filteredView[currentIdx].Depth

	// Search backward for an item at the same depth
	for i := currentIdx - 1; i >= 0; i-- {
		if tv.filteredView[i].Depth == currentDepth {
			return tv.filteredView[i].Item
		}
	}

	return nil
}

// Indent indents the selected item (increases nesting level)
func (tv *TreeView) Indent() bool {
	if tv.selectedIdx < 1 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item

	// Find appropriate parent at the same depth (will become the parent after indenting)
	newParent := tv.findPreviousSiblingForIndent(tv.selectedIdx)
	if newParent == nil {
		return false // Can't indent if no sibling at same depth found
	}

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

	// Add to new parent as child
	newParent.AddChild(current)

	// Expand new parent to show the moved item
	newParent.Expanded = true

	tv.RebuildView()
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

	tv.RebuildView()
	return true
}

// SendItemToNode moves the currently selected item to be a child of the destination item
// Returns true if successful, false if there's no selected item or destination is invalid
// After moving, the selection stays at the same position to facilitate moving multiple items
func (tv *TreeView) SendItemToNode(destination *model.Item) bool {
	if tv.selectedIdx >= len(tv.filteredView) || destination == nil {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item
	if current == nil {
		return false
	}

	// Prevent sending an item to itself
	if current.ID == destination.ID {
		return false
	}

	// Prevent creating circular references (sending parent to its own descendant)
	if isDescendant(current, destination) {
		return false
	}

	// Remember the current selection position
	originalIdx := tv.selectedIdx

	// Remove from current parent or root
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

	// Add as child of destination
	destination.AddChild(current)

	// Expand destination to show the moved item
	destination.Expanded = true

	// Rebuild view
	tv.RebuildView()

	// Keep selection at the same position (or last item if we're now past the end)
	// This allows easy sequential sending with s.s.s.
	if originalIdx < len(tv.filteredView) {
		tv.SelectItem(originalIdx)
	} else if len(tv.filteredView) > 0 {
		tv.SelectItem(len(tv.filteredView) - 1)
	}

	return true
}

// isDescendant checks if potentialDescendant is a descendant of ancestor
func isDescendant(ancestor *model.Item, potentialDescendant *model.Item) bool {
	for _, child := range ancestor.Children {
		if child.ID == potentialDescendant.ID {
			return true
		}
		if isDescendant(child, potentialDescendant) {
			return true
		}
	}
	return false
}

// MoveItemDown moves the selected item down to the next position in linear order
// Positions cycle through all possible (parent, index) pairs in depth-first order
func (tv *TreeView) MoveItemDown() bool {
	if tv.selectedIdx >= len(tv.filteredView) || tv.selectedIdx < 0 {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item

	// Build list of all possible positions in DFS order, excluding invalid positions for this item
	positions := tv.buildAllPositionsForItem(current)

	// Find current position
	currentPos := -1
	currentParentIdx := 0
	if current.Parent != nil {
		for idx, child := range current.Parent.Children {
			if child.ID == current.ID {
				currentParentIdx = idx
				break
			}
		}
	} else {
		for idx, child := range tv.items {
			if child.ID == current.ID {
				currentParentIdx = idx
				break
			}
		}
	}

	for i, pos := range positions {
		if pos.Parent == current.Parent && pos.Index == currentParentIdx {
			currentPos = i
			break
		}
	}

	// Log all positions with current marked
	logPositions(positions, currentPos, current.Text, "MoveDown")

	if currentPos == -1 || currentPos >= len(positions)-1 {
		return false // Can't move further down
	}

	// Find next position
	// All positions in the list are valid (collapsed nodes are already filtered out)
	var nextPos Position
	searchIdx := currentPos + 1
	if searchIdx < len(positions) {
		nextPos = positions[searchIdx]
	}

	// If we couldn't find a valid position, can't move down
	if searchIdx >= len(positions) {
		return false
	}

	// Use the target index directly - positions are built before removal
	nextIdx := nextPos.Index

	// Remember the previous parent for comparison
	prevParent := current.Parent

	// Remove item from current parent
	if current.Parent != nil {
		current.Parent.Children = append(current.Parent.Children[:currentParentIdx], current.Parent.Children[currentParentIdx+1:]...)
	} else {
		tv.items = append(tv.items[:currentParentIdx], tv.items[currentParentIdx+1:]...)
	}

	// Insert at next position
	if nextPos.Parent != nil {
		children := nextPos.Parent.Children
		// Clamp nextIdx to valid range after array has been modified
		insertIdx := nextIdx
		if insertIdx > len(children) {
			insertIdx = len(children)
		}
		newChildren := make([]*model.Item, 0, len(children)+1)
		newChildren = append(newChildren, children[:insertIdx]...)
		newChildren = append(newChildren, current)
		newChildren = append(newChildren, children[insertIdx:]...)
		nextPos.Parent.Children = newChildren
	} else {
		children := tv.items
		// Clamp nextIdx to valid range after array has been modified
		insertIdx := nextIdx
		if insertIdx > len(children) {
			insertIdx = len(children)
		}
		newChildren := make([]*model.Item, 0, len(children)+1)
		newChildren = append(newChildren, children[:insertIdx]...)
		newChildren = append(newChildren, current)
		newChildren = append(newChildren, children[insertIdx:]...)
		tv.items = newChildren
	}

	current.Parent = nextPos.Parent

	// After insertion, find the actual index where current ended up
	var actualIdx int
	if nextPos.Parent != nil {
		for i, child := range nextPos.Parent.Children {
			if child.ID == current.ID {
				actualIdx = i
				break
			}
		}
	} else {
		for i, item := range tv.items {
			if item.ID == current.ID {
				actualIdx = i
				break
			}
		}
	}

	// Verify we actually moved the item. If we ended up in the same parent at the same index, return false
	if nextPos.Parent == prevParent && actualIdx == currentParentIdx {
		return false
	}

	tv.RebuildView()
	// Find the moved item in the new view and keep selection on it
	for idx, displayItem := range tv.filteredView {
		if displayItem.Item.ID == current.ID {
			tv.selectedIdx = idx
			break
		}
	}

	return true
}

// MoveItemUp moves the selected item up to the previous position in linear order
// Positions cycle through all possible (parent, index) pairs in depth-first order
func (tv *TreeView) MoveItemUp() bool {
	if tv.selectedIdx >= len(tv.filteredView) || tv.selectedIdx < 0 {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item

	// Build list of all possible positions in DFS order, excluding invalid positions for this item
	positions := tv.buildAllPositionsForItem(current)

	// Find current position
	currentPos := -1
	currentParentIdx := 0
	if current.Parent != nil {
		for idx, child := range current.Parent.Children {
			if child.ID == current.ID {
				currentParentIdx = idx
				break
			}
		}
	} else {
		for idx, child := range tv.items {
			if child.ID == current.ID {
				currentParentIdx = idx
				break
			}
		}
	}

	for i, pos := range positions {
		if pos.Parent == current.Parent && pos.Index == currentParentIdx {
			currentPos = i
			break
		}
	}

	// Log all positions with current marked
	logPositions(positions, currentPos, current.Text, "MoveUp")

	if currentPos <= 0 {
		return false // Can't move further up
	}

	// Find previous position
	// All positions in the list are valid (collapsed nodes are already filtered out)
	var prevPos Position
	searchIdx := currentPos - 1
	if searchIdx >= 0 {
		prevPos = positions[searchIdx]
	}

	// If we couldn't find a valid position, can't move up
	if searchIdx < 0 {
		return false
	}

	// Use the target index directly - positions are built before removal
	prevIdx := prevPos.Index

	// Remove item from current parent
	if current.Parent != nil {
		current.Parent.Children = append(current.Parent.Children[:currentParentIdx], current.Parent.Children[currentParentIdx+1:]...)
	} else {
		tv.items = append(tv.items[:currentParentIdx], tv.items[currentParentIdx+1:]...)
	}

	// Insert at previous position
	if prevPos.Parent != nil {
		children := prevPos.Parent.Children
		// Clamp prevIdx to valid range after array has been modified
		insertIdx := prevIdx
		if insertIdx > len(children) {
			insertIdx = len(children)
		}
		newChildren := make([]*model.Item, 0, len(children)+1)
		newChildren = append(newChildren, children[:insertIdx]...)
		newChildren = append(newChildren, current)
		newChildren = append(newChildren, children[insertIdx:]...)
		prevPos.Parent.Children = newChildren
	} else {
		children := tv.items
		// Clamp prevIdx to valid range after array has been modified
		insertIdx := prevIdx
		if insertIdx > len(children) {
			insertIdx = len(children)
		}
		newChildren := make([]*model.Item, 0, len(children)+1)
		newChildren = append(newChildren, children[:insertIdx]...)
		newChildren = append(newChildren, current)
		newChildren = append(newChildren, children[insertIdx:]...)
		tv.items = newChildren
	}

	current.Parent = prevPos.Parent

	tv.RebuildView()
	// Find the moved item in the new view and keep selection on it
	for idx, displayItem := range tv.filteredView {
		if displayItem.Item.ID == current.ID {
			tv.selectedIdx = idx
			break
		}
	}
	return true
}

// Position represents a location where an item can be placed: (parent, index)
type Position struct {
	Parent *model.Item // nil for root level
	Index  int         // 0 to len(Parent.Children)
}

// buildAllPositions builds a list of all possible positions in depth-first order
// Each position is defined as (parent, index) where index is the position in parent's children array
func (tv *TreeView) buildAllPositions() []Position {
	var positions []Position
	tv.buildPositionsRecursive(nil, &positions)
	return positions
}

// buildAllPositionsForItem builds a list of all valid positions for moving a specific item
// It excludes positions where the item would become its own parent or ancestor
// It builds positions as if the item was temporarily removed from the tree (to get stable ordering)
func (tv *TreeView) buildAllPositionsForItem(item *model.Item) []Position {
	// Temporarily remove the item from its current location to get stable positions
	// that don't depend on where the item currently is
	var wasParent *model.Item
	var wasIndex int
	var wasRemoved bool

	if item.Parent != nil {
		wasParent = item.Parent
		for i, child := range item.Parent.Children {
			if child.ID == item.ID {
				wasIndex = i
				wasParent.Children = append(wasParent.Children[:i], wasParent.Children[i+1:]...)
				wasRemoved = true
				break
			}
		}
	} else {
		for i, rootItem := range tv.items {
			if rootItem.ID == item.ID {
				wasIndex = i
				tv.items = append(tv.items[:i], tv.items[i+1:]...)
				wasRemoved = true
				break
			}
		}
	}

	// Build positions for the tree WITHOUT the item
	allPositions := tv.buildAllPositions()

	// Restore the item to its original position
	if wasRemoved {
		if wasParent != nil {
			newChildren := make([]*model.Item, 0, len(wasParent.Children)+1)
			newChildren = append(newChildren, wasParent.Children[:wasIndex]...)
			newChildren = append(newChildren, item)
			newChildren = append(newChildren, wasParent.Children[wasIndex:]...)
			wasParent.Children = newChildren
		} else {
			newItems := make([]*model.Item, 0, len(tv.items)+1)
			newItems = append(newItems, tv.items[:wasIndex]...)
			newItems = append(newItems, item)
			newItems = append(newItems, tv.items[wasIndex:]...)
			tv.items = newItems
		}
	}

	// Build set of descendants of the item (items we can't move it to)
	descendants := make(map[*model.Item]bool)
	tv.collectDescendants(item, descendants)

	// Filter out invalid positions
	var validPositions []Position
	for _, pos := range allPositions {
		// Can't move item to be a child of itself or any of its descendants
		if pos.Parent != item && !descendants[pos.Parent] {
			validPositions = append(validPositions, pos)
		}
	}

	return validPositions
}

// collectDescendants recursively collects all descendants of an item
func (tv *TreeView) collectDescendants(item *model.Item, descendants map[*model.Item]bool) {
	for _, child := range item.Children {
		descendants[child] = true
		tv.collectDescendants(child, descendants)
	}
}

// buildPositionsRecursive recursively builds all positions starting from a parent
func (tv *TreeView) buildPositionsRecursive(parent *model.Item, positions *[]Position) {
	var children []*model.Item
	if parent == nil {
		children = tv.items
	} else {
		// Only include positions for expanded nodes with children
		// If a node is Expanded but has no children, treat it as if it's collapsed
		if !parent.Expanded || len(parent.Children) == 0 {
			// Don't create any positions for collapsed nodes or expanded nodes with no children
			return
		}
		children = parent.Children
	}

	// Add positions for all insertion points in this parent's children
	for i := 0; i <= len(children); i++ {
		*positions = append(*positions, Position{Parent: parent, Index: i})
		// Recursively add positions for the child at this index
		if i < len(children) {
			tv.buildPositionsRecursive(children[i], positions)
		}
	}
}

// logPositions logs all positions in a readable format with the current position marked as [CUR]
func logPositions(positions []Position, currentIdx int, itemText string, operation string) {
	var posStrs []string
	for i, pos := range positions {
		var parentName string
		if pos.Parent == nil {
			parentName = "root"
		} else {
			parentName = pos.Parent.Text
		}

		posStr := fmt.Sprintf("(%s,%d)", parentName, pos.Index)

		if i == currentIdx {
			posStr += " [CUR]"
		}

		posStrs = append(posStrs, posStr)
	}

	debugLogger.Printf("[%s] '%s': %s", operation, itemText, strings.Join(posStrs, ", "))
}

// AddItemAfter adds a new item after the selected item
func (tv *TreeView) AddItemAfter(newItem *model.Item) {
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
	tv.RebuildView()
	// Find and select the new item in the filtered view
	for idx, dispItem := range tv.filteredView {
		if dispItem.Item.ID == newItem.ID {
			tv.selectedIdx = idx
			return
		}
	}
}

// AddItemAsChild adds a new item as a child of the selected item
func (tv *TreeView) AddItemAsChild(newItem *model.Item) {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		selected := tv.filteredView[tv.selectedIdx].Item
		selected.AddChild(newItem)
		selected.Expanded = true
	} else {
		tv.items = append(tv.items, newItem)
	}
	tv.RebuildView()
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
	tv.RebuildView()
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

	displayItem := tv.filteredView[tv.selectedIdx]
	if displayItem.IsVirtual {
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

	tv.RebuildView()
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

	tv.RebuildView()
	// Move selection back if needed
	if tv.selectedIdx >= len(tv.filteredView) && len(tv.filteredView) > 0 {
		tv.selectedIdx = len(tv.filteredView) - 1
	}
	return true
}

// PasteAfter pastes an item after the selected item and returns the pasted item (or nil on failure)
func (tv *TreeView) PasteAfter(item *model.Item) *model.Item {
	if item == nil || len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return nil
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
				tv.RebuildView()
				return item
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
				tv.RebuildView()
				return item
			}
		}
	}
	return nil
}

// PasteBefore pastes an item before the selected item and returns the pasted item (or nil on failure)
func (tv *TreeView) PasteBefore(item *model.Item) *model.Item {
	if item == nil || len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return nil
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
				tv.RebuildView()
				return item
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
				tv.RebuildView()
				return item
			}
		}
	}
	return nil
}

// GetSelected returns the currently selected item
func (tv *TreeView) GetSelected() *model.Item {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		return tv.filteredView[tv.selectedIdx].Item
	}
	return nil
}

// GetSelectedDisplayItem returns the selected displayItem (includes virtual/real status and parent info)
func (tv *TreeView) GetSelectedDisplayItem() *displayItem {
	if len(tv.filteredView) > 0 && tv.selectedIdx < len(tv.filteredView) {
		return tv.filteredView[tv.selectedIdx]
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

func (tv *TreeView) SelectItemByID(id string) {
	idx := -1
	for itemIdx, item := range tv.filteredView {
		if item.Item.ID == id {
			idx = itemIdx
			break
		}
	}
	if idx >= 0 {
		tv.SelectItem(idx)
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

// GetDisplayLines returns the current display lines (multi-line aware)
func (tv *TreeView) GetDisplayLines() []*DisplayLine {
	return tv.displayLines
}

// GetViewportOffset returns the index of the first visible display line in the viewport
func (tv *TreeView) GetViewportOffset() int {
	return tv.viewportOffset
}

// Render renders the tree to the screen
func (tv *TreeView) Render(screen *Screen, startY, endY int, visualAnchor int, cfg *config.Config) {
	tv.RenderWithSearchQuery(screen, startY, endY, visualAnchor, "", nil, cfg)
}

// RenderWithSearchQuery renders the tree with optional search query highlighting
func (tv *TreeView) RenderWithSearchQuery(screen *Screen, startY, endY int, visualAnchor int, searchQuery string, currentMatchItem *model.Item, cfg *config.Config) {
	screenWidth := screen.GetWidth()
	screenHeight := screen.GetHeight()

	// Calculate max width for text wrapping
	// Reserve space for indentation (max 6 levels * 3 chars) and arrow/indicator/space (3 chars)
	// This ensures we have at least some reasonable width for text
	maxTextWidth := screenWidth - 21 // 6 levels * 3 + 3 for arrow area
	if maxTextWidth < 20 {
		maxTextWidth = 20 // Minimum wrap width
	}

	// Update max width if it changed
	tv.SetMaxWidth(maxTextWidth)

	defaultStyle := screen.TreeNormalStyle()
	selectedStyle := screen.TreeSelectedStyle()
	visualStyle := screen.TreeVisualSelectionStyle()
	visualCursorStyle := screen.TreeVisualCursorStyle()
	newItemStyle := screen.TreeNewItemStyle()
	highlightStyle := screen.SearchHighlightStyle()

	// Add background to non-selected styles
	bgColor := screen.Theme.Colors.Background
	defaultStyle = defaultStyle.Background(bgColor)
	newItemStyle = newItemStyle.Background(bgColor)

	// Calculate available viewport height
	viewportHeight := endY

	// Get the display line range for the selected item
	selectedItem := tv.GetSelected()
	selectedDisplayItem := tv.GetSelectedDisplayItem()
	var firstLineOfSelected, lastLineOfSelected int
	if selectedItem != nil {
		firstLineOfSelected = tv.getFirstDisplayLineForItem(selectedItem)
		lastLineOfSelected = tv.getLastDisplayLineForItem(selectedItem)
	}

	// Ensure viewport offset keeps selected item visible (show first line)
	if firstLineOfSelected >= 0 {
		if firstLineOfSelected < tv.viewportOffset {
			tv.viewportOffset = firstLineOfSelected
		} else if lastLineOfSelected >= tv.viewportOffset+viewportHeight {
			tv.viewportOffset = lastLineOfSelected - viewportHeight + 1
		}
	}

	// Clamp viewport offset
	maxOffset := max(len(tv.displayLines)-viewportHeight, 0)
	if tv.viewportOffset > maxOffset {
		tv.viewportOffset = maxOffset
	}
	if tv.viewportOffset < 0 {
		tv.viewportOffset = 0
	}

	// For visual selection, convert item indices to display line ranges
	var visualStart, visualEnd int
	var hasVisualSelection bool
	if visualAnchor >= 0 {
		hasVisualSelection = true
		// Get the display line ranges for both anchor and selected items
		anchorItem := tv.filteredView[visualAnchor].Item
		selectedItemObj := tv.filteredView[tv.selectedIdx].Item

		anchorFirst := tv.getFirstDisplayLineForItem(anchorItem)
		selectedFirst := tv.getFirstDisplayLineForItem(selectedItemObj)
		selectedLast := tv.getLastDisplayLineForItem(selectedItemObj)

		if anchorFirst >= 0 && selectedFirst >= 0 {
			visualStart = anchorFirst
			visualEnd = selectedLast
			if visualStart > visualEnd {
				visualStart, visualEnd = visualEnd, visualStart
			}
		} else {
			hasVisualSelection = false
		}
	}

	// Render display lines starting from viewportOffset
	screenY := startY
	for i := tv.viewportOffset; i < len(tv.displayLines) && screenY < screenHeight-1; i++ {
		displayLine := tv.displayLines[i]
		y := screenY

		// Determine if this line's item is selected
		// Highlight only if this display line is from the same displayItem as the selected item
		// This prevents multiple references to the same item from all being highlighted
		isLinePartOfSelected := selectedItem != nil && displayLine.Item.ID == selectedItem.ID && displayLine.ParentDisplayItem == selectedDisplayItem

		// Select style based on selection, visual selection, and new item status
		style := defaultStyle

		// Check if in visual selection range
		inVisualRange := hasVisualSelection && i >= visualStart && i <= visualEnd

		// Determine appropriate background color for all line elements
		lineBackgroundColor := bgColor
		var leafArrowStyle, expandableArrowStyle tcell.Style

		if inVisualRange {
			leafArrowStyle = visualCursorStyle
			expandableArrowStyle = visualCursorStyle
			if isLinePartOfSelected {
				style = visualCursorStyle
			} else {
				style = visualStyle
			}
		} else if isLinePartOfSelected {
			style = selectedStyle
			// Use selected background color for all elements on this line
			lineBackgroundColor = screen.Theme.Colors.TreeSelectedBg
			leafArrowStyle = screen.TreeLeafArrowStyle().Background(lineBackgroundColor)
			expandableArrowStyle = screen.TreeExpandableArrowStyle().Background(lineBackgroundColor)
		} else {
			// Normal unselected line
			leafArrowStyle = screen.TreeLeafArrowStyle().Background(bgColor)
			expandableArrowStyle = screen.TreeExpandableArrowStyle().Background(bgColor)
		}

		// Only render item metadata (indent, arrow, attributes, progress) on the first line
		if displayLine.ItemStartLine {
			// Add indentation for parent levels (3 spaces per level)
			prefix := strings.Repeat("   ", displayLine.Depth)

			// Draw indentation
			if displayLine.Depth > 0 {
				screen.DrawString(0, y, prefix, style)
			}

			// Always draw an arrow
			// Use different colors for leaf vs expandable nodes
			arrowStyle := leafArrowStyle // Default to leaf (dimmer)
			hasChildren := len(displayLine.Item.Children) > 0 || len(displayLine.Item.GetVirtualChildren()) > 0
			if hasChildren {
				// For nodes with children, use brighter expandable arrow style
				arrowStyle = expandableArrowStyle
			}
			if !inVisualRange && isLinePartOfSelected {
				arrowStyle = selectedStyle // Use selected style if item is selected
			}

			// Determine which arrow to show
			arrow := "▶"
			if displayLine.IsVirtual {
				// For virtual children, check if it's collapsed in the search node
				if displayLine.SearchNodeParent != nil && displayLine.SearchNodeParent.IsVirtualChildCollapsed(displayLine.Item.ID) {
					// Collapsed virtual item: show right arrow
					arrow = "→"
				} else if hasChildren {
					// Expanded virtual item: show down arrow
					arrow = "↓"
				} else {
					// Virtual leaf with no children
					arrow = "→"
				}
			} else if hasChildren && displayLine.Item.Expanded {
				// Real items: show down arrow if expanded and has children
				arrow = "▼"
			}

			prefixX := displayLine.Depth * 3
			screen.DrawString(prefixX, y, arrow, arrowStyle)

			// Draw attribute indicator or space to maintain alignment
			indicatorStyle := screen.TreeAttributeIndicatorStyle().Background(lineBackgroundColor)
			if isLinePartOfSelected {
				indicatorStyle = selectedStyle // Use selected style if item is selected
			}

			hasAttributes := displayLine.Item.Metadata != nil && len(displayLine.Item.Metadata.Attributes) > 0
			if hasAttributes {
				screen.SetCell(prefixX+1, y, '●', indicatorStyle) // Filled circle for items with attributes
			} else {
				screen.SetCell(prefixX+1, y, ' ', style) // Space for items without attributes
			}

			// Text starts at fixed position
			textX := prefixX + 3                     // Position after the arrow, indicator, and space
			screen.SetCell(prefixX+2, y, ' ', style) // Space after indicator

			// Calculate max width available for text with truncation
			maxTextWidth := screenWidth - textX
			if maxTextWidth < 0 {
				maxTextWidth = 0
			}

			// Truncate with ellipsis if text exceeds max width
			text := displayLine.TextLine
			if len(text) > maxTextWidth {
				if maxTextWidth > 1 {
					text = text[:maxTextWidth-1] + "…"
				} else {
					text = "…"
				}
			}

			// Draw the text with link and search highlighting
			// Links are always highlighted, search highlighting is applied only to current match
			// Don't highlight virtual references (items shown in search nodes)
			linkStyle := screen.TreeLinkStyle()
			var displayLen int
			if searchQuery != "" && currentMatchItem != nil && displayLine.Item == currentMatchItem && !displayLine.IsVirtual {
				_, displayLen = tv.drawTextWithLinksAndSearch(screen, textX, y, text, style, highlightStyle, linkStyle, searchQuery)
			} else {
				_, displayLen = tv.drawTextWithLinksAndSearch(screen, textX, y, text, style, highlightStyle, linkStyle, "")
			}

			// Draw visible attributes if configured (only on item start line)
			// Use displayLen (actual displayed length after link compression) instead of original text length
			totalLen := textX + displayLen
			if cfg != nil {
				visattrConfig := cfg.Get("visattr")
				if visattrConfig != "" {
					// Parse comma-separated attribute names
					attrNames := strings.Split(visattrConfig, ",")
					var visibleAttrs []string

					if displayLine.Item.Metadata != nil && len(displayLine.Item.Metadata.Attributes) > 0 {
						for _, attrName := range attrNames {
							attrName = strings.TrimSpace(attrName)
							if value, exists := displayLine.Item.Metadata.Attributes[attrName]; exists && value != "" {
								visibleAttrs = append(visibleAttrs, attrName+":"+value)
							}
						}
					}

					// Draw attributes in gray if any are found
					if len(visibleAttrs) > 0 {
						attrStr := "  [" + strings.Join(visibleAttrs, ", ") + "]"
						attrStyle := screen.TreeAttributeStyle().Background(lineBackgroundColor) // Gray/dim style with background color
						if isLinePartOfSelected {
							attrStyle = selectedStyle // Use selected style if item is selected
						}

						// Draw the attribute string if it fits on screen
						// Use StringWidth for proper display width calculation with Unicode characters
						attrX := totalLen
						attrWidth := StringWidth(attrStr)
						if attrX+attrWidth <= screenWidth {
							screen.DrawString(attrX, y, attrStr, attrStyle)
							totalLen = attrX + attrWidth
						}
					}
				}
			}

			// Draw progress bar if configured and if item has todo children (only on item start line)
			if cfg != nil && cfg.Get("showprogress") != "false" {
				statusesStr := cfg.Get("todostatuses")
				if statusesStr == "" {
					statusesStr = "todo,doing,done"
				}
				statuses := strings.Split(statusesStr, ",")

				blocks := RenderProgressBar(displayLine.Item, statuses)
				if len(blocks) > 0 {
					// Add spacing before progress bar
					barStartX := totalLen + 2
					if barStartX < screenWidth {
						screen.SetCell(barStartX-2, y, ' ', style)
						screen.SetCell(barStartX-1, y, ' ', style)

						// Draw each block with appropriate color
						firstStatus := statuses[0]
						lastStatus := statuses[len(statuses)-1]

						for j, block := range blocks {
							blockX := barStartX + j
							if blockX >= screenWidth {
								break
							}

							blockStyle := screen.GrayStyle().Background(lineBackgroundColor) // Default to gray for todo
							switch block.Status {
							case lastStatus:
								blockStyle = screen.GreenStyle().Background(lineBackgroundColor) // Green for done
							case firstStatus:
								blockStyle = screen.GrayStyle().Background(lineBackgroundColor) // Gray for todo
							default:
								blockStyle = screen.OrangeStyle().Background(lineBackgroundColor) // Orange for doing/in-progress
							}

							screen.SetCell(blockX, y, '■', blockStyle)
						}
						totalLen = barStartX + len(blocks)
					}
				}
			}

			// Pad to wrap width with background color on first line only
			// Use the same wrap width that the editor uses for consistent alignment
			wrapEndX := textX + tv.maxWidth
			if wrapEndX > screenWidth {
				wrapEndX = screenWidth
			}
			bgStyle := screen.BackgroundStyle()
			for x := totalLen; x < wrapEndX; x++ {
				padStyle := bgStyle
				if isLinePartOfSelected {
					// Fill entire padding with selected background
					padStyle = selectedStyle
				}
				screen.SetCell(x, y, ' ', padStyle)
			}
		} else {
			// Align with first line's text position
			textX := displayLine.Depth*3 + 3

			// Calculate wrap width for continuation lines
			// Use the same wrap width that the editor uses for consistent alignment
			wrapEndX := textX + tv.maxWidth
			if wrapEndX > screenWidth {
				wrapEndX = screenWidth
			}

			// For continuation lines, fill entire wrap width with background or selection color first
			bgStyle := screen.BackgroundStyle()
			lineStyle := style
			for x := 0; x < wrapEndX; x++ {
				fillStyle := bgStyle
				if isLinePartOfSelected {
					fillStyle = selectedStyle
				}
				screen.SetCell(x, y, ' ', fillStyle)
			}
			if isLinePartOfSelected {
				lineStyle = selectedStyle
			}

			// Render continuation line text with full width available
			maxTextWidth := wrapEndX - textX
			if maxTextWidth < 0 {
				maxTextWidth = 0
			}

			// Truncate with ellipsis if text exceeds max width
			text := displayLine.TextLine
			if len(text) > maxTextWidth {
				if maxTextWidth > 1 {
					text = text[:maxTextWidth-1] + "…"
				} else {
					text = "…"
				}
			}

			// Draw continuation line text
			screen.DrawString(textX, y, text, lineStyle)
		}

		screenY++ // Move to next screen line
	}

	// Clear remaining lines with background color
	bgStyle := screen.BackgroundStyle()
	for y := screenY; y < screen.GetHeight()-1; y++ {
		clearLine := strings.Repeat(" ", screenWidth)
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

// IndentItem indents a specific item (makes it a child of sibling at same depth)
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

	// Find appropriate parent at the same depth (will become the parent after indenting)
	newParent := tv.findPreviousSiblingForIndent(itemIdx)
	if newParent == nil {
		return false // Can't indent if no sibling at same depth found
	}

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

	// Add to new parent as child
	newParent.AddChild(item)

	// Expand new parent to show the moved item
	newParent.Expanded = true

	tv.RebuildView()
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

	tv.RebuildView()
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

// SelectParent moves selection to the parent of the current item.
// Returns: true if parent was found and selected, false otherwise.
// Special case: when hoisted and at root of hoisted view, returns false
// (caller should handle hoisting to parent).
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

// IsAtRootOfHoistedView returns true if we're hoisted and the current selection
// is at the root level of the hoisted view (i.e., a direct child of the hoisted item)
func (tv *TreeView) IsAtRootOfHoistedView() bool {
	if !tv.IsHoisted() {
		return false
	}
	if len(tv.filteredView) == 0 || tv.selectedIdx >= len(tv.filteredView) {
		return false
	}

	current := tv.filteredView[tv.selectedIdx].Item
	// Check if current item's parent is the hoisted item
	return current.Parent != nil && current.Parent.ID == tv.hoistedItem.ID
}

// HoistToParent moves the hoist up to the parent of the currently hoisted node.
// If the hoisted node has no parent, unhoist completely.
// Returns true if successful, false if not hoisted.
func (tv *TreeView) HoistToParent() bool {
	if !tv.IsHoisted() {
		return false
	}

	hoistedItem := tv.GetHoistedItem()
	if hoistedItem == nil {
		return false
	}

	// If hoisted item has a parent, hoist to that parent
	if hoistedItem.Parent != nil {
		// The originalItems should contain the root, which has the parent item
		if tv.originalItems == nil {
			return false
		}

		// Find the parent item in originalItems (recursively if needed)
		var parentItem *model.Item
		for _, item := range tv.originalItems {
			if item.ID == hoistedItem.Parent.ID {
				parentItem = item
				break
			}
			// Also check recursively in case parent is nested deeper
			if found := findItemRecursive(item, hoistedItem.Parent.ID); found != nil {
				parentItem = found
				break
			}
		}

		if parentItem == nil {
			return false
		}

		// Now set the parent as the new hoisted item
		// Keep originalItems pointing to the root for proper unhoist
		tv.hoistedItem = parentItem
		tv.items = parentItem.Children
		tv.selectedIdx = 0
		tv.viewportOffset = 0
		tv.RebuildView()

		return true
	}

	// No parent, so unhoist completely
	return tv.Unhoist()
}

// findItemRecursive is a helper to find an item by ID in a tree
func findItemRecursive(item *model.Item, targetID string) *model.Item {
	if item.ID == targetID {
		return item
	}
	for _, child := range item.Children {
		if found := findItemRecursive(child, targetID); found != nil {
			return found
		}
	}
	return nil
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
	tv.RebuildView()
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
	tv.RebuildView()

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
	tv.RebuildView()

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

// drawTextWithSearchHighlight draws text with highlighted search matches
// It intelligently detects if the search is fuzzy or text-based and highlights accordingly
func (tv *TreeView) drawTextWithSearchHighlight(screen *Screen, x int, y int, text string, defaultStyle tcell.Style, highlightStyle tcell.Style, searchQuery string) {
	if searchQuery == "" {
		screen.DrawString(x, y, text, defaultStyle)
		return
	}

	// Check if this is a fuzzy search (starts with ~)
	if strings.HasPrefix(searchQuery, "~") {
		fuzzyTerm := strings.TrimPrefix(searchQuery, "~")
		tv.drawTextWithFuzzyHighlight(screen, x, y, text, defaultStyle, highlightStyle, fuzzyTerm)
	} else {
		// Regular text search highlighting
		tv.drawTextWithHighlight(screen, x, y, text, defaultStyle, highlightStyle, searchQuery)
	}
}

// drawTextWithFuzzyHighlight draws text with fuzzy match highlighting
// It highlights individual characters that match the fuzzy query in order
func (tv *TreeView) drawTextWithFuzzyHighlight(screen *Screen, x int, y int, text string, defaultStyle tcell.Style, highlightStyle tcell.Style, fuzzyTerm string) {
	if fuzzyTerm == "" {
		screen.DrawString(x, y, text, defaultStyle)
		return
	}

	// Find positions of fuzzy matches
	lowerText := strings.ToLower(text)
	lowerTerm := strings.ToLower(fuzzyTerm)
	var matchPositions []int
	textIdx := 0

	// For each character in the search term, find it in the text
	for _, termChar := range lowerTerm {
		found := false
		for i := textIdx; i < len(lowerText); i++ {
			if rune(lowerText[i]) == termChar {
				matchPositions = append(matchPositions, i)
				textIdx = i + 1
				found = true
				break
			}
		}
		if !found {
			// Shouldn't happen if the match is valid, but fall back to normal text display
			screen.DrawString(x, y, text, defaultStyle)
			return
		}
	}

	// Create a set of matched positions for quick lookup
	matchSet := make(map[int]bool)
	for _, pos := range matchPositions {
		matchSet[pos] = true
	}

	// Draw the text, highlighting matched characters
	currentX := x
	for i, r := range text {
		if matchSet[i] {
			screen.SetCell(currentX, y, r, highlightStyle)
		} else {
			screen.SetCell(currentX, y, r, defaultStyle)
		}
		currentX++
	}
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

// drawTextWithLinksAndSearch draws text with both wiki-style links and search highlighting
// Links are shown in cyan with underline, but displayed text hides the [[id|...]] brackets
// Search matches have background highlight
// Returns: parsed links, display text length (accounting for link compression)
func (tv *TreeView) drawTextWithLinksAndSearch(screen *Screen, x int, y int, text string,
	defaultStyle tcell.Style, highlightStyle tcell.Style, linkStyle tcell.Style, searchQuery string) ([]links.Link, int) {

	// Parse links from the text
	itemLinks := links.ParseLinks(text)

	// If no links, just draw normally (with search highlighting if needed)
	if len(itemLinks) == 0 {
		if searchQuery == "" {
			screen.DrawString(x, y, text, defaultStyle)
			return itemLinks, len(text)
		}
		// Draw with search highlighting only
		tv.drawTextWithHighlight(screen, x, y, text, defaultStyle, highlightStyle, searchQuery)
		return itemLinks, len(text)
	}

	// Build display text and track where links appear in display
	var displayParts []string
	var displayLinkRanges []struct {
		start, end int
	}
	displayPos := 0

	lastEnd := 0
	for _, link := range itemLinks {
		// Add text before this link
		if link.StartPos > lastEnd {
			displayParts = append(displayParts, text[lastEnd:link.StartPos])
			displayPos += link.StartPos - lastEnd
		}

		// Add link display text (with link styling)
		linkDisplay := link.GetDisplayText()
		displayParts = append(displayParts, linkDisplay)
		displayLinkRanges = append(displayLinkRanges, struct {
			start, end int
		}{displayPos, displayPos + len(linkDisplay)})
		displayPos += len(linkDisplay)

		lastEnd = link.EndPos
	}
	// Add remaining text after last link
	if lastEnd < len(text) {
		displayParts = append(displayParts, text[lastEnd:])
	}

	displayText := strings.Join(displayParts, "")

	// Build map of which characters in original text are in search matches
	searchMatches := make([]bool, len(text))
	if searchQuery != "" {
		lowerText := strings.ToLower(text)
		lowerQuery := strings.ToLower(searchQuery)
		idx := 0
		for {
			matchIdx := strings.Index(lowerText[idx:], lowerQuery)
			if matchIdx == -1 {
				break
			}
			matchIdx += idx
			for i := 0; i < len(searchQuery); i++ {
				if matchIdx+i < len(text) {
					searchMatches[matchIdx+i] = true
				}
			}
			idx = matchIdx + len(searchQuery)
		}
	}

	// Draw display text character by character with appropriate styling
	currentX := x
	for i, r := range displayText {
		charStyle := defaultStyle

		// Check if this display position is in a link
		inLink := false
		for _, linkRange := range displayLinkRanges {
			if i >= linkRange.start && i < linkRange.end {
				// Apply link color with underline, but preserve background from defaultStyle
				// Start with defaultStyle and apply link foreground and underline
				fg, _, _ := linkStyle.Decompose()
				charStyle = defaultStyle.Foreground(fg).Underline(true)
				inLink = true
				break
			}
		}

		// Apply search highlighting if not already in a link
		if !inLink && searchQuery != "" {
			// For search matching, we'd need to map back to original positions
			// For now, skip search highlighting in links (or could apply both)
		}

		screen.SetCell(currentX, y, r, charStyle)
		currentX++
	}

	// Return the display text length (shorter than original due to link compression)
	return itemLinks, len(displayText)
}
