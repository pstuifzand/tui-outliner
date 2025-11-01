// Package model contains the model for the outline
package model

import "time"

// Item represents a single node in the outline tree
type Item struct {
	ID       string    `json:"id"`
	Text     string    `json:"text"`
	Children []*Item   `json:"children,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
	Parent   *Item     `json:"-"` // Not persisted
	Expanded bool      `json:"-"` // UI state, not persisted
	IsNew    bool      `json:"-"` // UI state: true for newly created placeholder items
}

// Metadata holds rich information about an item
type Metadata struct {
	Tags       []string          `json:"tags,omitempty"`
	Notes      string            `json:"notes,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Created    time.Time         `json:"created"`
	Modified   time.Time         `json:"modified"`
}

// Outline represents the entire outline document
type Outline struct {
	Items []*Item `json:"items"`
}

// NewItem creates a new outline item with a generated ID
func NewItem(text string) *Item {
	return &Item{
		ID:       generateID(),
		Text:     text,
		Children: make([]*Item, 0),
		Metadata: &Metadata{
			Attributes: make(map[string]string),
			Created:    time.Now(),
			Modified:   time.Now(),
		},
		Expanded: true,
		IsNew:    true,
	}
}

// NewOutline creates a new outline with the given title
func NewOutline() *Outline {
	return &Outline{
		Items: make([]*Item, 0),
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
	for _, item := range o.GetAllItems() {
		if item.ID == id {
			return item
		}
	}
	return nil
}

func generateID() string {
	return "item_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

func randomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range length {
		result[i] = chars[int(time.Now().UnixNano()+int64(i))%len(chars)]
	}
	return string(result)
}
