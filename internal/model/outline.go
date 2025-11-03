// Package model contains the model for the outline
package model

import (
	"crypto/rand"
	"maps"
	"slices"
	"time"
)

// Item represents a single node in the outline tree
type Item struct {
	ID               string    `json:"id"`
	Text             string    `json:"text"`
	Children         []*Item   `json:"children,omitempty"`
	VirtualChildRefs []string  `json:"virtual_children,omitempty"` // IDs of items to show as children (not duplicated)
	Metadata         *Metadata `json:"metadata,omitempty"`
	Parent           *Item     `json:"-"` // Not persisted
	Expanded         bool      `json:"-"` // UI state, not persisted
	virtualChildren  []*Item   `json:"-"` // Resolved virtual child pointers (runtime only)
	// CollapsedVirtualChildren tracks which virtual children are collapsed (for display-only)
	// Maps virtual child item ID -> true if collapsed. Only used for search nodes to avoid
	// collapsing the original items. This is session-only state.
	CollapsedVirtualChildren map[string]bool `json:"-"`
}

// Metadata holds rich information about an item
type Metadata struct {
	Tags       []string          `json:"tags,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Created    time.Time         `json:"created"`
	Modified   time.Time         `json:"modified"`
}

// Outline represents the entire outline document
type Outline struct {
	Items                 []*Item          `json:"items"`
	OriginalFilename      string           `json:"original_filename,omitempty"`
	itemIndex             map[string]*Item `json:"-"` // Fast O(1) ID lookup cache
}

// NewItem creates a new outline item with a generated ID
func NewItem(text string) *Item {
	return &Item{
		ID:   generateID(),
		Text: text,
		Metadata: &Metadata{
			Attributes: make(map[string]string),
			Created:    time.Now(),
			Modified:   time.Now(),
		},
		Expanded:                 false,
		CollapsedVirtualChildren: make(map[string]bool),
	}
}

func NewItemFrom(item *Item) *Item {
	yanked := NewItem(item.Text)
	yanked.Metadata.Attributes = maps.Clone(yanked.Metadata.Attributes)
	for _, c := range item.Children {
		yc := NewItemFrom(c)
		yanked.Children = append(yanked.Children, yc)
		yc.Parent = yanked
	}
	return yanked
}

// NewOutline creates a new outline with the given title
func NewOutline() *Outline {
	return &Outline{
		Items:     nil,
		itemIndex: make(map[string]*Item),
	}
}

// AddChild adds a child item to this item
func (i *Item) AddChild(child *Item) {
	child.Parent = i
	i.Children = append(i.Children, child)
}

// RemoveChild removes a child item from this item
func (i *Item) RemoveChild(child *Item) {
	for idx, c := range i.Children {
		if c.ID == child.ID {
			i.Children = append(i.Children[:idx], i.Children[idx+1:]...)
			child.Parent = nil
			// If no children remain, set Expanded to false
			if len(i.Children) == 0 {
				i.Expanded = false
			}
			break
		}
	}
}

// GetAllItems returns all items in the outline (depth-first)
func (o *Outline) GetAllItems() []*Item {
	var items []*Item
	for _, item := range o.Items {
		items = append(items, getAllItemsRecursive(item)...)
	}
	return items
}

func getAllItemsRecursive(item *Item) []*Item {
	items := []*Item{item}
	for _, child := range item.Children {
		items = append(items, getAllItemsRecursive(child)...)
	}
	return items
}

// FindItemByID finds an item by its ID in the outline
func (o *Outline) FindItemByID(id string) *Item {
	// Use index if available (O(1))
	if o.itemIndex != nil {
		if item, ok := o.itemIndex[id]; ok {
			return item
		}
		return nil
	}
	// Fallback to linear search (O(n))
	for _, item := range o.GetAllItems() {
		if item.ID == id {
			o.itemIndex[id] = item
			return item
		}
	}
	return nil
}

// BuildIndex creates a fast lookup cache for item IDs
func (o *Outline) BuildIndex() {
	o.itemIndex = make(map[string]*Item)
	for _, item := range o.GetAllItems() {
		o.itemIndex[item.ID] = item
	}
}

// ResolveVirtualChildren resolves all virtual child references to actual item pointers
// Detects and prevents circular references
func (o *Outline) ResolveVirtualChildren() {
	visited := make(map[string]bool)
	for _, item := range o.Items {
		o.resolveVirtualChildrenRecursive(item, visited)
	}
}

func (o *Outline) resolveVirtualChildrenRecursive(item *Item, visitedPath map[string]bool) {
	if visitedPath[item.ID] {
		// Circular reference detected, skip
		item.virtualChildren = nil
		return
	}

	// Mark as visited in this path
	visitedPath[item.ID] = true
	defer delete(visitedPath, item.ID)

	// Clear and rebuild virtual children
	item.virtualChildren = make([]*Item, 0, len(item.VirtualChildRefs))

	for _, refID := range item.VirtualChildRefs {
		if refItem := o.FindItemByID(refID); refItem != nil {
			item.virtualChildren = append(item.virtualChildren, refItem)
		}
	}

	// Recursively resolve for all children
	for _, child := range item.Children {
		o.resolveVirtualChildrenRecursive(child, visitedPath)
	}
}

// GetVirtualChildren returns the resolved virtual child references
func (i *Item) GetVirtualChildren() []*Item {
	if i.virtualChildren == nil {
		return make([]*Item, 0)
	}
	return i.virtualChildren
}

// IsVirtualChildCollapsed checks if a virtual child is collapsed in the search node's display
func (i *Item) IsVirtualChildCollapsed(virtualChildID string) bool {
	if i.CollapsedVirtualChildren == nil {
		return false
	}
	return i.CollapsedVirtualChildren[virtualChildID]
}

// SetVirtualChildCollapsed marks a virtual child as collapsed in the search node's display
func (i *Item) SetVirtualChildCollapsed(virtualChildID string, collapsed bool) {
	if i.CollapsedVirtualChildren == nil {
		i.CollapsedVirtualChildren = make(map[string]bool)
	}
	if collapsed {
		i.CollapsedVirtualChildren[virtualChildID] = true
	} else {
		delete(i.CollapsedVirtualChildren, virtualChildID)
	}
}

// AddVirtualChild adds a virtual child reference by ID
func (i *Item) AddVirtualChild(itemID string) {
	if slices.Contains(i.VirtualChildRefs, itemID) {
		return // Already exists
	}
	i.VirtualChildRefs = append(i.VirtualChildRefs, itemID)
}

// RemoveVirtualChild removes a virtual child reference by ID
func (i *Item) RemoveVirtualChild(itemID string) {
	for idx, ref := range i.VirtualChildRefs {
		if ref == itemID {
			i.VirtualChildRefs = append(i.VirtualChildRefs[:idx], i.VirtualChildRefs[idx+1:]...)
			break
		}
	}
}

// ClearVirtualChildren clears all virtual child references
func (i *Item) ClearVirtualChildren() {
	i.VirtualChildRefs = nil
	i.virtualChildren = nil
}

// IsSearchNode returns true if this item is a search node (has type="search" attribute)
func (i *Item) IsSearchNode() bool {
	if i.Metadata == nil || i.Metadata.Attributes == nil {
		return false
	}
	return i.Metadata.Attributes["type"] == "search"
}

// GetSearchQuery returns the search query from the @query attribute
func (i *Item) GetSearchQuery() string {
	if i.Metadata == nil || i.Metadata.Attributes == nil {
		return ""
	}
	return i.Metadata.Attributes["query"]
}

// PopulateSearchNode sets the virtual children for a search node from a list of matching item IDs
// Returns the number of results set
func (o *Outline) PopulateSearchNode(item *Item, matchingIDs []string) int {
	if !item.IsSearchNode() {
		return 0
	}

	// Clear and set virtual children
	item.VirtualChildRefs = matchingIDs

	// Resolve the virtual child pointers
	visitedPath := make(map[string]bool)
	o.resolveVirtualChildrenRecursive(item, visitedPath)

	return len(matchingIDs)
}

func generateID() string {
	return "item_" + time.Now().Format("20060102150405") + "_" + rand.Text()
}
