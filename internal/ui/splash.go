package ui

import (
	"strings"
)

// SplashScreen displays a welcome screen when no file is provided
type SplashScreen struct {
	visible bool
}

// NewSplashScreen creates a new SplashScreen
func NewSplashScreen() *SplashScreen {
	return &SplashScreen{
		visible: false,
	}
}

// Show makes the splash screen visible
func (s *SplashScreen) Show() {
	s.visible = true
}

// Hide makes the splash screen invisible
func (s *SplashScreen) Hide() {
	s.visible = false
}

// IsVisible returns whether the splash screen is visible
func (s *SplashScreen) IsVisible() bool {
	return s.visible
}

// GetContent returns the lines to display on the splash screen
func (s *SplashScreen) GetContent() []string {
	return []string{
		"",
		"",
		"    ~~ TUI Outliner ~~",
		"",
		"       Version 1.0",
		"",
		"",
		"    A terminal-based outliner",
		"    inspired by Vim",
		"",
		"",
		"    Commands:",
		"    :e <filename>  - Open or create a file",
		"    :help          - Show keybindings",
		"    :q             - Quit (use :q! to force)",
		"    :wq            - Save and quit",
		"",
		"",
		"    Type :e filename to get started",
		"",
	}
}

// Render renders the splash screen
func (s *SplashScreen) Render(screen *Screen) {
	if !s.visible {
		return
	}

	bgStyle := screen.BackgroundStyle()
	textStyle := screen.HeaderStyle()
	dimStyle := screen.StatusMessageStyle()

	width := screen.GetWidth()
	height := screen.GetHeight()

	// Fill background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			screen.SetCell(x, y, ' ', bgStyle)
		}
	}

	// Get content lines
	content := s.GetContent()

	// Find the longest line to determine block width
	maxWidth := 0
	for _, line := range content {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Calculate vertical centering
	totalLines := len(content)
	startY := (height - totalLines) / 2
	if startY < 0 {
		startY = 0
	}

	// Calculate horizontal centering for the entire block
	startX := (width - maxWidth) / 2
	if startX < 0 {
		startX = 0
	}

	// Render content as a centered block
	for i, line := range content {
		y := startY + i
		if y >= height {
			break
		}

		// Determine style based on content
		style := textStyle
		if strings.Contains(line, "Commands:") || strings.Contains(line, "Type :e") {
			style = dimStyle
		}

		screen.DrawString(startX, y, line, style)
	}
}
