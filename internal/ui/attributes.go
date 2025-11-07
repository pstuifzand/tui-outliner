package ui

import (
	"fmt"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
	tmpl "github.com/pstuifzand/tui-outliner/internal/template"
)

// ValidateAttributeFunc is a function that validates an attribute value
// Returns error message if invalid, empty string if valid
type ValidateAttributeFunc func(key, value string) string

// AttributeEditor manages the attribute editing modal
type AttributeEditor struct {
	visible         bool
	item            *model.Item
	attributes      []string // Sorted list of attribute keys
	selectedIdx     int
	editingIdx      int    // -1 if not editing, >= 0 if editing key, -2 if editing value
	editingKey      string // Current key being edited
	editingValue    string // Current value being edited
	mode            string // "view", "edit", "add_key", "add_value"
	statusMessage   string
	keyEditor       *Editor               // Editor for key input
	valueEditor     *Editor               // Editor for value input
	onModified      func()                // Callback when attributes are modified
	validateAttr    ValidateAttributeFunc // Callback to validate attributes
	calendarWidget  *CalendarWidget       // Calendar for date selection
	calendarVisible bool                  // Whether calendar is open
	typeRegistry    *tmpl.TypeRegistry    // Type definitions for attribute validation
}

// NewAttributeEditor creates a new AttributeEditor
func NewAttributeEditor() *AttributeEditor {
	// Create a temporary item for the editors (they will be replaced when needed)
	tempItem := model.NewItem("")
	return &AttributeEditor{
		visible:      false,
		item:         nil,
		attributes:   []string{},
		selectedIdx:  0,
		editingIdx:   -1,
		mode:         "view",
		keyEditor:    NewEditor(tempItem),
		valueEditor:  NewEditor(tempItem),
		typeRegistry: tmpl.NewTypeRegistry(),
	}
}

// SetOnModified sets a callback to be called when attributes are modified
func (ae *AttributeEditor) SetOnModified(callback func()) {
	ae.onModified = callback
}

// SetValidateAttribute sets the validation callback for attribute values
func (ae *AttributeEditor) SetValidateAttribute(validate ValidateAttributeFunc) {
	ae.validateAttr = validate
}

// ValidateAttribute validates an attribute using the validation callback
// Returns error message if invalid, empty string if valid
func (ae *AttributeEditor) ValidateAttribute(key, value string) string {
	if ae.validateAttr == nil {
		return ""
	}
	return ae.validateAttr(key, value)
}

// SetCalendarWidget sets the calendar widget for date selection
func (ae *AttributeEditor) SetCalendarWidget(calendar *CalendarWidget) {
	ae.calendarWidget = calendar
	if ae.calendarWidget != nil {
		// Set up callback for when a date is selected from the calendar
		ae.calendarWidget.SetOnDateSelected(func(selectedDate time.Time) {
			// Format the date as YYYY-MM-DD
			dateStr := selectedDate.Format("2006-01-02")
			// Update the value editor with the selected date
			ae.valueEditor.SetText(dateStr)
			// Reset the cursor position to the end of the new text
			ae.valueEditor.Stop()
			ae.valueEditor.SetText(dateStr)
			ae.valueEditor.Start()
			ae.editingValue = dateStr
			ae.calendarVisible = false
		})
	}
}

// SetTypeRegistry sets the type registry for type-aware value selection
func (ae *AttributeEditor) SetTypeRegistry(registry *tmpl.TypeRegistry) {
	if registry != nil {
		ae.typeRegistry = registry
	}
}

// Show shows the attribute editor for an item
func (ae *AttributeEditor) Show(item *model.Item) {
	ae.item = item
	ae.visible = true
	ae.selectedIdx = 0
	ae.editingIdx = -1
	ae.mode = "view"
	ae.statusMessage = ""
	ae.refreshAttributes()
}

// ShowInAddMode shows the attribute editor for an item and immediately starts adding a new attribute
func (ae *AttributeEditor) ShowInAddMode(item *model.Item) {
	ae.item = item
	ae.visible = true
	ae.selectedIdx = 0
	ae.editingIdx = -1
	ae.mode = "add_key"
	ae.statusMessage = "Enter attribute key"
	ae.refreshAttributes()

	// Start the key editor
	ae.keyEditor.SetText("")
	ae.keyEditor.Start()
}

// Hide hides the attribute editor
func (ae *AttributeEditor) Hide() {
	ae.visible = false
	ae.item = nil
	ae.statusMessage = ""
}

// IsVisible returns whether the editor is visible
func (ae *AttributeEditor) IsVisible() bool {
	return ae.visible
}

// refreshAttributes refreshes the sorted list of attributes
func (ae *AttributeEditor) refreshAttributes() {
	ae.attributes = []string{}
	if ae.item == nil || ae.item.Metadata == nil || ae.item.Metadata.Attributes == nil {
		return
	}

	for key := range ae.item.Metadata.Attributes {
		ae.attributes = append(ae.attributes, key)
	}
	sort.Strings(ae.attributes)

	if ae.selectedIdx >= len(ae.attributes) {
		ae.selectedIdx = len(ae.attributes) - 1
	}
	if ae.selectedIdx < 0 {
		ae.selectedIdx = 0
	}
}

// HandleKeyEvent processes keyboard input in the modal
func (ae *AttributeEditor) HandleKeyEvent(ev *tcell.EventKey) bool {
	if !ae.visible {
		return false
	}

	// Handle calendar input if calendar is visible
	if ae.calendarVisible && ae.calendarWidget != nil {
		if ae.calendarWidget.HandleKeyEvent(ev) {
			// Calendar closed or date selected
			if !ae.calendarWidget.IsVisible() {
				ae.calendarVisible = false
			}
			return true
		}
	}

	key := ev.Key()

	// Escape to close in view mode
	if key == tcell.KeyEscape && ae.mode == "view" {
		ae.Hide()
		return true
	}

	switch ae.mode {
	case "view":
		return ae.handleViewMode(ev)
	case "edit":
		return ae.handleEditMode(ev)
	case "add_key":
		return ae.handleAddKeyMode(ev)
	case "add_value":
		return ae.handleAddValueMode(ev)
	}

	return false
}

// HandleMouseEvent processes mouse input in the attribute editor
func (ae *AttributeEditor) HandleMouseEvent(x, y int) bool {
	if !ae.visible || !ae.calendarVisible || ae.calendarWidget == nil {
		return false
	}

	// Route mouse events to the calendar widget
	return ae.calendarWidget.HandleMouseEvent(x, y)
}

// handleViewMode handles keys in view mode
func (ae *AttributeEditor) handleViewMode(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyCtrlD && ae.calendarWidget != nil {
		ae.openCalendarForEdit()
		return true
	}

	ch := ev.Rune()
	switch ch {
	case 'j':
		// Move down
		if ae.selectedIdx < len(ae.attributes)-1 {
			ae.selectedIdx++
		}
		return true
	case 'k':
		// Move up
		if ae.selectedIdx > 0 {
			ae.selectedIdx--
		}
		return true
	case 'e':
		// Edit selected attribute
		if ae.selectedIdx < len(ae.attributes) {
			ae.editingIdx = ae.selectedIdx
			ae.editingKey = ae.attributes[ae.selectedIdx]
			ae.editingValue = ae.item.Metadata.Attributes[ae.editingKey]
			ae.mode = "edit"
			ae.valueEditor.SetText(ae.editingValue)
			ae.valueEditor.Start()
			ae.statusMessage = "Editing value"
		}
		return true
	case 'd':
		// Delete selected attribute
		if ae.selectedIdx < len(ae.attributes) {
			key := ae.attributes[ae.selectedIdx]
			delete(ae.item.Metadata.Attributes, key)
			ae.statusMessage = fmt.Sprintf("Deleted attribute '%s'", key)
			ae.refreshAttributes()
			if ae.onModified != nil {
				ae.onModified()
			}
		}
		return true
	case 'a':
		// Add new attribute
		ae.mode = "add_key"
		ae.keyEditor.SetText("")
		ae.keyEditor.Start()
		ae.statusMessage = "Enter attribute key"
		return true
	case 'q':
		// Quit
		ae.Hide()
		return true
	}

	return false
}

// handleEditMode handles keys while editing an attribute value
func (ae *AttributeEditor) handleEditMode(ev *tcell.EventKey) bool {
	ae.valueEditor.HandleKey(ev)

	if ae.valueEditor.WasEnterPressed() {
		// Save the edited value
		if ae.item != nil && ae.item.Metadata != nil && ae.item.Metadata.Attributes != nil {
			newValue := ae.valueEditor.GetText()

			// Validate the new value if validation is set
			if ae.validateAttr != nil {
				if errMsg := ae.validateAttr(ae.editingKey, newValue); errMsg != "" {
					ae.statusMessage = fmt.Sprintf("Invalid: %s", errMsg)
					ae.valueEditor.Stop()
					ae.valueEditor.Start()
					return true
				}
			}

			ae.item.Metadata.Attributes[ae.editingKey] = newValue
			ae.statusMessage = fmt.Sprintf("Updated '%s'", ae.editingKey)
			ae.mode = "view"
			ae.editingIdx = -1
			ae.refreshAttributes()
			if ae.onModified != nil {
				ae.onModified()
			}
		} else {
			ae.statusMessage = "Error: invalid state"
		}
		return true
	}

	if ae.valueEditor.WasEscapePressed() {
		// Cancel editing
		ae.mode = "view"
		ae.editingIdx = -1
		ae.statusMessage = "Cancelled"
		return true
	}

	return true
}

// handleAddKeyMode handles keys while entering the key for a new attribute
func (ae *AttributeEditor) handleAddKeyMode(ev *tcell.EventKey) bool {
	ae.keyEditor.HandleKey(ev)

	if ae.keyEditor.WasEnterPressed() {
		keyText := ae.keyEditor.GetText()
		if keyText == "" {
			ae.statusMessage = "Key cannot be empty"
			ae.keyEditor.Start() // Re-activate for more input
			return true
		}
		ae.editingKey = keyText
		ae.mode = "add_value"
		ae.valueEditor.SetText("")
		ae.valueEditor.Start()
		ae.statusMessage = "Enter attribute value"
		return true
	}

	if ae.keyEditor.WasEscapePressed() {
		ae.mode = "view"
		ae.statusMessage = "Cancelled"
		return true
	}

	return true
}

// handleAddValueMode handles keys while entering the value for a new attribute
func (ae *AttributeEditor) handleAddValueMode(ev *tcell.EventKey) bool {
	ae.valueEditor.HandleKey(ev)

	if ae.valueEditor.WasEnterPressed() {
		// Add the new attribute
		// Ensure metadata and attributes map are initialized
		if ae.item == nil {
			ae.statusMessage = "Error: no item selected"
			return true
		}
		if ae.item.Metadata == nil {
			ae.item.Metadata = &model.Metadata{}
		}
		if ae.item.Metadata.Attributes == nil {
			ae.item.Metadata.Attributes = make(map[string]string)
		}

		newValue := ae.valueEditor.GetText()

		// Validate the new value if validation is set
		if ae.validateAttr != nil {
			if errMsg := ae.validateAttr(ae.editingKey, newValue); errMsg != "" {
				ae.statusMessage = fmt.Sprintf("Invalid: %s", errMsg)
				ae.valueEditor.Stop()
				ae.valueEditor.Start()
				return true
			}
		}

		ae.item.Metadata.Attributes[ae.editingKey] = newValue
		ae.statusMessage = fmt.Sprintf("Added attribute '%s'", ae.editingKey)
		ae.mode = "view"
		ae.refreshAttributes()
		if ae.onModified != nil {
			ae.onModified()
		}
		return true
	}

	if ae.valueEditor.WasEscapePressed() {
		ae.mode = "view"
		ae.statusMessage = "Cancelled"
		return true
	}

	return true
}

// openCalendarForEdit opens the calendar widget for editing the current attribute value
func (ae *AttributeEditor) openCalendarForEdit() {
	if ae.calendarWidget == nil {
		ae.statusMessage = "Calendar not available"
		return
	}
	key := ae.attributes[ae.selectedIdx]
	value := ae.item.Metadata.Attributes[key]

	ae.calendarWidget.ShowForAttributeWithValue(key, value)
	ae.calendarVisible = true
	ae.statusMessage = "Opening calendar..."
}

// Render renders the attribute editor modal
func (ae *AttributeEditor) Render(screen *Screen) {
	if !ae.visible {
		return
	}

	contentStyle := screen.HelpStyle()
	borderStyle := screen.HelpBorderStyle()
	titleStyle := screen.HelpTitleStyle()

	// Draw background
	for y := 0; y < screen.GetHeight(); y++ {
		for x := 0; x < screen.GetWidth(); x++ {
			screen.SetCell(x, y, ' ', contentStyle)
		}
	}

	// Draw modal box
	startY := 1
	startX := 3
	boxWidth := screen.GetWidth() - 6
	height := screen.GetHeight() - 3

	// Draw top border
	screen.SetCell(startX, startY, '┌', borderStyle)
	for i := 1; i < boxWidth-1; i++ {
		screen.SetCell(startX+i, startY, '─', borderStyle)
	}
	screen.SetCell(startX+boxWidth-1, startY, '┐', borderStyle)

	// Draw title
	title := " Attributes "
	screen.DrawString(startX+2, startY+1, title, titleStyle)
	screen.SetCell(startX, startY+1, '│', borderStyle)
	screen.SetCell(startX+boxWidth-1, startY+1, '│', borderStyle)

	// Draw middle border
	screen.SetCell(startX, startY+2, '├', borderStyle)
	for i := 1; i < boxWidth-1; i++ {
		screen.SetCell(startX+i, startY+2, '─', borderStyle)
	}
	screen.SetCell(startX+boxWidth-1, startY+2, '┤', borderStyle)

	// Draw content based on mode
	contentHeight := height - 4

	switch ae.mode {
	case "view":
		ae.renderViewMode(screen, startX, startY, boxWidth, height, contentHeight, contentStyle, borderStyle, screen.TreeSelectedStyle())
	case "edit":
		ae.renderEditMode(screen, startX, startY, boxWidth, height, contentHeight, contentStyle, borderStyle, titleStyle)
	case "add_key", "add_value":
		ae.renderAddMode(screen, startX, startY, boxWidth, height, contentHeight, contentStyle, borderStyle, titleStyle)
	}

	// Draw status message at bottom
	screen.SetCell(startX, startY+height-2, '├', borderStyle)
	for i := 1; i < boxWidth-1; i++ {
		screen.SetCell(startX+i, startY+height-2, '─', borderStyle)
	}
	screen.SetCell(startX+boxWidth-1, startY+height-2, '┤', borderStyle)

	msgLine := " " + ae.statusMessage
	if len(msgLine) > boxWidth-3 {
		msgLine = msgLine[:boxWidth-3]
	}
	screen.DrawString(startX+1, startY+height-1, msgLine, contentStyle)
	screen.SetCell(startX, startY+height-1, '│', borderStyle)
	screen.SetCell(startX+boxWidth-1, startY+height-1, '│', borderStyle)

	// Draw bottom border
	screen.SetCell(startX, startY+height, '└', borderStyle)
	for i := 1; i < boxWidth-1; i++ {
		screen.SetCell(startX+i, startY+height, '─', borderStyle)
	}
	screen.SetCell(startX+boxWidth-1, startY+height, '┘', borderStyle)

	// Render calendar if visible
	if ae.calendarVisible && ae.calendarWidget != nil {
		ae.calendarWidget.Render(screen)
	}
}

// renderViewMode renders the view mode content
func (ae *AttributeEditor) renderViewMode(screen *Screen, startX, startY, boxWidth, height, contentHeight int, contentStyle, borderStyle, selectionStyle tcell.Style) {
	y := startY + 3

	// If no attributes
	if len(ae.attributes) == 0 {
		msg := " No attributes "
		// Clear the line first, then draw borders
		screen.SetCell(startX, y, '│', borderStyle)
		screen.DrawString(startX+1, y, msg, contentStyle)
		screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
		y++

		// Fill empty lines with borders
		for y < startY+height-2 {
			screen.SetCell(startX, y, '│', borderStyle)
			screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
			y++
		}
		// Update status message instead of drawing help here
		ae.statusMessage = "[a] Add  [q] Close"
		return
	}

	// Draw attributes
	for i, key := range ae.attributes {
		if y >= startY+height-2 {
			break
		}

		value := ae.item.Metadata.Attributes[key]
		// Truncate long values
		maxLen := boxWidth - 10
		if len(value) > maxLen {
			value = value[:maxLen] + "..."
		}

		line := fmt.Sprintf(" %s: %s", key, value)
		if len(line) > boxWidth-2 {
			line = line[:boxWidth-2]
		}

		screen.SetCell(startX, y, '│', borderStyle)
		if i == ae.selectedIdx {
			screen.DrawString(startX+1, y, line, selectionStyle)
		} else {
			screen.DrawString(startX+1, y, line, contentStyle)
		}
		screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
		y++
	}

	// Fill empty lines with borders
	for y < startY+height-2 {
		screen.SetCell(startX, y, '│', borderStyle)
		screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
		y++
	}

	// Update status message with help text
	ae.statusMessage = "[j/k]Select [e]Edit [d]Delete [a]Add [q]Quit"
}

// renderEditMode renders the edit mode content
func (ae *AttributeEditor) renderEditMode(screen *Screen, startX, startY, boxWidth, height, contentHeight int, contentStyle, borderStyle, titleStyle tcell.Style) {
	y := startY + 3

	// Draw key line (read-only)
	keyLine := fmt.Sprintf(" Key: %s", ae.editingKey)
	if len(keyLine) > boxWidth-2 {
		keyLine = keyLine[:boxWidth-2]
	}
	screen.SetCell(startX, y, '│', borderStyle)
	screen.DrawString(startX+1, y, keyLine, contentStyle)
	screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
	y++

	// Draw value line using Editor
	screen.SetCell(startX, y, '│', borderStyle)
	valuePrefix := " Value: "
	screen.DrawString(startX+1, y, valuePrefix, contentStyle)
	// Render the editor for the remaining width - use StringWidth for proper Unicode handling
	valuePrefixWidth := StringWidth(valuePrefix)
	maxValueWidth := boxWidth - 2 - valuePrefixWidth
	ae.valueEditor.Render(screen, startX+1+valuePrefixWidth, y, maxValueWidth)
	screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
	y++

	// Fill empty lines with borders
	for y < startY+height-2 {
		screen.SetCell(startX, y, '│', borderStyle)
		screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
		y++
	}

	// Update status message with help text
	if ae.calendarWidget != nil {
		ae.statusMessage = "[Enter]Save [Escape]Cancel [Ctrl+D]Calendar"
	} else {
		ae.statusMessage = "[Enter]Save [Escape]Cancel"
	}
}

// renderAddMode renders the add mode content
func (ae *AttributeEditor) renderAddMode(screen *Screen, startX, startY, boxWidth, height, contentHeight int, contentStyle, borderStyle, titleStyle tcell.Style) {
	y := startY + 3

	if ae.mode == "add_key" {
		// Draw key input using Editor
		screen.SetCell(startX, y, '│', borderStyle)
		keyPrefix := " Key: "
		screen.DrawString(startX+1, y, keyPrefix, contentStyle)
		// Use StringWidth for proper Unicode handling
		keyPrefixWidth := StringWidth(keyPrefix)
		maxKeyWidth := boxWidth - 2 - keyPrefixWidth
		ae.keyEditor.Render(screen, startX+1+keyPrefixWidth, y, maxKeyWidth)
		screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
		y++
	} else {
		// Add key (display only)
		keyLine := fmt.Sprintf(" Key: %s", ae.editingKey)
		// Use StringWidth and TruncateToWidth for proper Unicode handling
		keyLine = TruncateToWidth(keyLine, boxWidth-2)
		screen.SetCell(startX, y, '│', borderStyle)
		screen.DrawString(startX+1, y, keyLine, contentStyle)
		screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
		y++

		// Draw value input using Editor
		screen.SetCell(startX, y, '│', borderStyle)
		valuePrefix := " Value: "
		screen.DrawString(startX+1, y, valuePrefix, contentStyle)
		// Use StringWidth for proper Unicode handling
		valuePrefixWidth := StringWidth(valuePrefix)
		maxValueWidth := boxWidth - 2 - valuePrefixWidth
		ae.valueEditor.Render(screen, startX+1+valuePrefixWidth, y, maxValueWidth)
		screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
		y++
	}

	// Fill empty lines with borders
	for y < startY+height-2 {
		screen.SetCell(startX, y, '│', borderStyle)
		screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
		y++
	}

	// Update status message with help text
	if ae.mode == "add_key" {
		ae.statusMessage = "[Enter]Next [Escape]Cancel"
	} else {
		ae.statusMessage = "[Enter]Save [Escape]Cancel"
	}
}
