package ui

import "fmt"

// HelpScreen manages the help display
type HelpScreen struct {
	visible bool
}

// NewHelpScreen creates a new HelpScreen
func NewHelpScreen() *HelpScreen {
	return &HelpScreen{
		visible: false,
	}
}

// Toggle toggles the help screen visibility
func (h *HelpScreen) Toggle() {
	h.visible = !h.visible
}

// IsVisible returns whether the help screen is visible
func (h *HelpScreen) IsVisible() bool {
	return h.visible
}

// GetKeybindings returns a list of keybindings
func (h *HelpScreen) GetKeybindings() []string {
	return []string{
		"Navigation:",
		"  j/Down      - Move down",
		"  k/Up        - Move up",
		"  h/Left      - Collapse item",
		"  l/Right     - Expand item",
		"  >/Ctrl+I    - Indent item",
		"  </Ctrl+U    - Outdent item",
		"",
		"Editing:",
		"  i           - Edit selected item",
		"  o           - Insert new item after",
		"  O           - Insert new item before",
		"  a           - Append new child item",
		"  d           - Delete selected item",
		"  /           - Search/filter",
		"  m           - Edit metadata",
		"",
		"Other:",
		"  ?           - Toggle help",
		"  :w          - Save",
		"  :q          - Quit",
		"  Escape      - Close dialogs/exit edit mode",
	}
}

// Render renders the help screen
func (h *HelpScreen) Render(screen *Screen) {
	if !h.visible {
		return
	}

	style := DefaultStyle()
	titleStyle := StyleBold()

	// Draw background (semi-transparent with reverse)
	for y := 0; y < screen.GetHeight(); y++ {
		for x := 0; x < screen.GetWidth(); x++ {
			screen.SetCell(x, y, ' ', StyleDim())
		}
	}

	// Draw help box
	startY := 2
	startX := 5
	width := screen.GetWidth() - 10
	height := screen.GetHeight() - 4

	keybindings := h.GetKeybindings()

	// Draw title
	title := " Keybindings (? to close) "
	screen.DrawString(startX, startY, fmt.Sprintf("┌─%s─┐", repeatString("─", len(title))), style)
	screen.DrawString(startX+2, startY+1, title, titleStyle)
	screen.DrawString(startX, startY+2, "├─" + repeatString("─", width-2) + "┤", style)

	// Draw keybindings
	y := startY + 3
	for i, binding := range keybindings {
		if y >= startY+height-1 {
			break
		}
		if i < len(keybindings) {
			screen.DrawString(startX+2, y, binding, style)
			y++
		}
	}

	// Draw bottom border
	screen.DrawString(startX, y, "└"+repeatString("─", width)+"┘", style)
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
