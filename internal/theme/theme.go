package theme

import (
	"github.com/gdamore/tcell/v2"
)

// Colors holds all the color definitions for the theme
type Colors struct {
	// Global background
	Background tcell.Color

	// Tree view colors
	TreeNormalText       tcell.Color
	TreeSelectedItem     tcell.Color
	TreeSelectedBg       tcell.Color  // Background for selected items
	TreeNewItem          tcell.Color
	TreeLeafArrow        tcell.Color  // Dimmer arrow for leaf nodes (no children)
	TreeExpandableArrow  tcell.Color  // Brighter arrow for nodes with children
	TreeExpandedArrow    tcell.Color
	TreeCollapsedArrow   tcell.Color
	TreeVisualSelection  tcell.Color  // Foreground color for items in visual selection
	TreeVisualSelectionBg tcell.Color // Background color for items in visual selection
	TreeVisualCursor     tcell.Color  // Foreground color for visual mode cursor
	TreeVisualCursorBg   tcell.Color  // Background color for visual mode cursor
	TreeAttributeIndicator tcell.Color // Color for attribute indicator symbol
	TreeAttributeValue   tcell.Color  // Color for visible attribute values (gray/dim)

	// Editor colors
	EditorText        tcell.Color
	EditorCursor      tcell.Color
	EditorCursorBg    tcell.Color

	// Search bar colors
	SearchLabel       tcell.Color
	SearchText        tcell.Color
	SearchCursor      tcell.Color
	SearchCursorBg    tcell.Color
	SearchResultCount tcell.Color
	SearchHighlight   tcell.Color   // Highlight color for matching text in items
	SearchHighlightBg tcell.Color   // Background color for matching text in items

	// Command line colors
	CommandPrompt   tcell.Color
	CommandText     tcell.Color
	CommandCursor   tcell.Color
	CommandCursorBg tcell.Color

	// Help overlay colors
	HelpBackground tcell.Color
	HelpBorder     tcell.Color
	HelpTitle      tcell.Color
	HelpContent    tcell.Color

	// Status line colors
	StatusMode       tcell.Color
	StatusModeBg     tcell.Color
	StatusMessage    tcell.Color
	StatusModified   tcell.Color

	// Header colors
	HeaderTitle tcell.Color
	HeaderBg    tcell.Color
}

// Theme represents a complete color theme
type Theme struct {
	Name   string
	Colors Colors
}

// Default returns a default theme with simple black, white, and gray colors
func Default() *Theme {
	return &Theme{
		Name: "default",
		Colors: Colors{
			// Global background - black
			Background: tcell.ColorBlack,
			// Tree view colors
			TreeNormalText:      tcell.ColorWhite,
			TreeSelectedItem:    tcell.ColorBlack,
			TreeSelectedBg:      tcell.ColorWhite,
			TreeNewItem:         tcell.ColorGray,
			TreeLeafArrow:       tcell.ColorGray,
			TreeExpandableArrow: tcell.ColorWhite,
			TreeExpandedArrow:   tcell.ColorWhite,
			TreeCollapsedArrow:  tcell.ColorWhite,
			TreeVisualSelection: tcell.ColorWhite,
			TreeVisualSelectionBg: tcell.ColorBlue,
			TreeVisualCursor:    tcell.ColorBlack,
			TreeVisualCursorBg:  tcell.ColorBlue,
			TreeAttributeIndicator: tcell.NewRGBColor(0, 208, 128), // Bright teal/green (#00D080)
			TreeAttributeValue:  tcell.ColorGray, // Dim gray for attribute values
			// Editor colors
			EditorText:        tcell.ColorWhite,
			EditorCursor:      tcell.ColorBlack,
			EditorCursorBg:    tcell.ColorWhite,
			// Search bar colors
			SearchLabel:       tcell.ColorWhite,
			SearchText:        tcell.ColorWhite,
			SearchCursor:      tcell.ColorBlack,
			SearchCursorBg:    tcell.ColorWhite,
			SearchResultCount: tcell.ColorWhite,
			SearchHighlight:   tcell.ColorBlack, // Black text on yellow background for search matches
			SearchHighlightBg: tcell.ColorYellow,
			// Command line colors
			CommandPrompt:   tcell.ColorWhite,
			CommandText:     tcell.ColorWhite,
			CommandCursor:   tcell.ColorBlack,
			CommandCursorBg: tcell.ColorWhite,
			// Help overlay colors
			HelpBackground: tcell.ColorBlack,
			HelpBorder:     tcell.ColorWhite,
			HelpTitle:      tcell.ColorWhite,
			HelpContent:    tcell.ColorWhite,
			// Status line colors
			StatusMode:   tcell.ColorWhite,
			StatusModeBg: tcell.ColorBlack,
			StatusMessage: tcell.ColorWhite,
			StatusModified: tcell.ColorWhite,
			// Header colors
			HeaderTitle: tcell.ColorWhite,
			HeaderBg:    tcell.ColorBlack,
		},
	}
}
