package theme

import (
	"github.com/gdamore/tcell/v2"
)

// Colors holds all the color definitions for the theme
type Colors struct {
	// Tree view colors
	TreeNormalText    tcell.Color
	TreeSelectedItem  tcell.Color
	TreeNewItem       tcell.Color
	TreeLeafArrow     tcell.Color
	TreeExpandedArrow tcell.Color
	TreeCollapsedArrow tcell.Color

	// Editor colors
	EditorText   tcell.Color
	EditorCursor tcell.Color

	// Search bar colors
	SearchLabel     tcell.Color
	SearchText      tcell.Color
	SearchCursor    tcell.Color
	SearchResultCount tcell.Color

	// Command line colors
	CommandPrompt   tcell.Color
	CommandText     tcell.Color
	CommandCursor   tcell.Color

	// Help overlay colors
	HelpBackground tcell.Color
	HelpBorder     tcell.Color
	HelpTitle      tcell.Color
	HelpContent    tcell.Color

	// Status line colors
	StatusMode       tcell.Color
	StatusMessage    tcell.Color
	StatusModified   tcell.Color

	// Header colors
	HeaderTitle tcell.Color
}

// Theme represents a complete color theme
type Theme struct {
	Name   string
	Colors Colors
}

// Default returns a default theme using terminal defaults
func Default() *Theme {
	return &Theme{
		Name: "default",
		Colors: Colors{
			// Use tcell default for most elements
			TreeNormalText:     tcell.ColorDefault,
			TreeSelectedItem:   tcell.ColorDefault,
			TreeNewItem:        tcell.ColorDefault,
			TreeLeafArrow:      tcell.ColorDefault,
			TreeExpandedArrow:  tcell.ColorDefault,
			TreeCollapsedArrow: tcell.ColorDefault,
			EditorText:         tcell.ColorDefault,
			EditorCursor:       tcell.ColorDefault,
			SearchLabel:        tcell.ColorDefault,
			SearchText:         tcell.ColorDefault,
			SearchCursor:       tcell.ColorDefault,
			SearchResultCount:  tcell.ColorDefault,
			CommandPrompt:      tcell.ColorDefault,
			CommandText:        tcell.ColorDefault,
			CommandCursor:      tcell.ColorDefault,
			HelpBackground:     tcell.ColorDefault,
			HelpBorder:         tcell.ColorDefault,
			HelpTitle:          tcell.ColorDefault,
			HelpContent:        tcell.ColorDefault,
			StatusMode:         tcell.ColorDefault,
			StatusMessage:      tcell.ColorDefault,
			StatusModified:     tcell.ColorDefault,
			HeaderTitle:        tcell.ColorDefault,
		},
	}
}

// TokyoNight returns the Tokyo Night theme
func TokyoNight() *Theme {
	return &Theme{
		Name: "tokyo-night",
		Colors: Colors{
			// Tokyo Night palette
			// Base colors
			TreeNormalText:     HexToColor("#c0caf5"),      // Light gray-blue
			TreeSelectedItem:   HexToColor("#7aa2f7"),      // Blue
			TreeNewItem:        HexToColor("#565f89"),      // Comment gray
			TreeLeafArrow:      HexToColor("#7dcfff"),      // Cyan
			TreeExpandedArrow:  HexToColor("#7dcfff"),      // Cyan
			TreeCollapsedArrow: HexToColor("#7dcfff"),      // Cyan
			EditorText:         HexToColor("#c0caf5"),      // Light gray-blue
			EditorCursor:       HexToColor("#7aa2f7"),      // Blue
			SearchLabel:        HexToColor("#bb9af7"),      // Magenta
			SearchText:         HexToColor("#c0caf5"),      // Light gray-blue
			SearchCursor:       HexToColor("#7aa2f7"),      // Blue
			SearchResultCount:  HexToColor("#9ece6a"),      // Green
			CommandPrompt:      HexToColor("#bb9af7"),      // Magenta
			CommandText:        HexToColor("#c0caf5"),      // Light gray-blue
			CommandCursor:      HexToColor("#7aa2f7"),      // Blue
			HelpBackground:     HexToColor("#1a1b26"),      // Dark background
			HelpBorder:         HexToColor("#7dcfff"),      // Cyan
			HelpTitle:          HexToColor("#bb9af7"),      // Magenta
			HelpContent:        HexToColor("#c0caf5"),      // Light gray-blue
			StatusMode:         HexToColor("#bb9af7"),      // Magenta
			StatusMessage:      HexToColor("#9ece6a"),      // Green
			StatusModified:     HexToColor("#f7768e"),      // Red
			HeaderTitle:        HexToColor("#bb9af7"),      // Magenta
		},
	}
}
