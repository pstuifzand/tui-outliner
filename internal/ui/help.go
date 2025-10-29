package ui

import "fmt"

// KeyBindingInfo represents a keybinding for display
type KeyBindingInfo interface {
	GetKey() rune
	GetDescription() string
}

// PendingKeyBindingInfo represents a pending keybinding for display
type PendingKeyBindingInfo interface {
	GetKey() rune
	GetDescription() string
	GetSequences() map[rune]string // Returns map of second key to description
}

// HelpScreen manages the help display
type HelpScreen struct {
	visible      bool
	keybindings  []KeyBindingInfo
}

// NewHelpScreen creates a new HelpScreen
func NewHelpScreen() *HelpScreen {
	return &HelpScreen{
		visible:     false,
		keybindings: []KeyBindingInfo{},
	}
}

// SetKeybindings sets the keybindings to display
func (h *HelpScreen) SetKeybindings(keybindings []KeyBindingInfo) {
	h.keybindings = keybindings
}

// Toggle toggles the help screen visibility
func (h *HelpScreen) Toggle() {
	h.visible = !h.visible
}

// IsVisible returns whether the help screen is visible
func (h *HelpScreen) IsVisible() bool {
	return h.visible
}

// GetKeybindings returns a formatted list of keybindings
func (h *HelpScreen) GetKeybindings() []string {
	var result []string

	result = append(result, "Keybindings:")
	result = append(result, "")

	for _, kb := range h.keybindings {
		// Check if this is a PendingKeyBinding by trying to cast it
		if pkb, ok := kb.(PendingKeyBindingInfo); ok {
			// Display the pending key with its sequences
			line := fmt.Sprintf("  %c  - %s", pkb.GetKey(), pkb.GetDescription())
			result = append(result, line)

			// Show the sub-sequences
			sequences := pkb.GetSequences()
			for seqKey, seqDesc := range sequences {
				line := fmt.Sprintf("    %c%c  - %s", pkb.GetKey(), seqKey, seqDesc)
				result = append(result, line)
			}
		} else {
			line := fmt.Sprintf("  %c  - %s", kb.GetKey(), kb.GetDescription())
			result = append(result, line)
		}
	}

	result = append(result, "")
	result = append(result, "Special Keys:")
	result = append(result, "  Ctrl+S      - Save")
	result = append(result, "  Escape      - Exit edit mode")
	result = append(result, "  Enter       - Confirm/Exit edit mode")
	result = append(result, "  Arrow Keys  - Navigate (alternative to hjkl)")

	return result
}

// Render renders the help screen
func (h *HelpScreen) Render(screen *Screen) {
	if !h.visible {
		return
	}

	contentStyle := screen.HelpStyle()
	borderStyle := screen.HelpBorderStyle()
	titleStyle := screen.HelpTitleStyle()

	// Draw background (semi-transparent with reverse)
	for y := 0; y < screen.GetHeight(); y++ {
		for x := 0; x < screen.GetWidth(); x++ {
			screen.SetCell(x, y, ' ', contentStyle)
		}
	}

	// Draw help box
	startY := 2
	startX := 5
	boxWidth := screen.GetWidth() - 10
	height := screen.GetHeight() - 4

	keybindings := h.GetKeybindings()

	// Draw top border
	screen.SetCell(startX, startY, '┌', borderStyle)
	for i := 1; i < boxWidth-1; i++ {
		screen.SetCell(startX+i, startY, '─', borderStyle)
	}
	screen.SetCell(startX+boxWidth-1, startY, '┐', borderStyle)

	// Draw title with side borders
	title := " Keybindings (? to close) "
	screen.SetCell(startX, startY+1, '│', borderStyle)
	screen.DrawString(startX+2, startY+1, title, titleStyle)
	screen.SetCell(startX+boxWidth-1, startY+1, '│', borderStyle)

	// Draw middle border
	screen.SetCell(startX, startY+2, '├', borderStyle)
	for i := 1; i < boxWidth-1; i++ {
		screen.SetCell(startX+i, startY+2, '─', borderStyle)
	}
	screen.SetCell(startX+boxWidth-1, startY+2, '┤', borderStyle)

	// Draw keybindings with side borders
	y := startY + 3
	for i, binding := range keybindings {
		if y >= startY+height-1 {
			break
		}
		if i < len(keybindings) {
			// Draw left border
			screen.SetCell(startX, y, '│', borderStyle)
			// Draw content
			screen.DrawString(startX+2, y, binding, contentStyle)
			// Draw right border
			screen.SetCell(startX+boxWidth-1, y, '│', borderStyle)
			y++
		}
	}

	// Draw bottom border
	screen.SetCell(startX, y, '└', borderStyle)
	for i := 1; i < boxWidth-1; i++ {
		screen.SetCell(startX+i, y, '─', borderStyle)
	}
	screen.SetCell(startX+boxWidth-1, y, '┘', borderStyle)
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
