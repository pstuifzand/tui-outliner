package theme

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// HexToColor converts a hex color string (#RRGGBB or #RGB) to tcell.Color
func HexToColor(hexColor string) tcell.Color {
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Handle short form (#RGB)
	if len(hexColor) == 3 {
		hexColor = string(hexColor[0]) + string(hexColor[0]) +
			string(hexColor[1]) + string(hexColor[1]) +
			string(hexColor[2]) + string(hexColor[2])
	}

	// Parse hex to RGB
	if len(hexColor) != 6 {
		return tcell.ColorDefault
	}

	// Use go-colorful for parsing
	c, err := colorful.Hex("#" + hexColor)
	if err != nil {
		return tcell.ColorDefault
	}

	// Convert to RGB values (0-255)
	r, g, b := c.RGB255()

	// Convert RGB to tcell color
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

// RGBToColor converts RGB values to tcell.Color
func RGBToColor(r, g, b int) tcell.Color {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return tcell.ColorDefault
	}
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

// ParseColorString handles multiple color formats: #RRGGBB, #RGB, or rgb(r,g,b)
func ParseColorString(colorStr string) tcell.Color {
	colorStr = strings.TrimSpace(colorStr)

	// Handle hex colors
	if strings.HasPrefix(colorStr, "#") {
		return HexToColor(colorStr)
	}

	// Handle rgb(r,g,b) format
	if strings.HasPrefix(colorStr, "rgb(") && strings.HasSuffix(colorStr, ")") {
		innerStr := strings.TrimPrefix(colorStr, "rgb(")
		innerStr = strings.TrimSuffix(innerStr, ")")
		parts := strings.Split(innerStr, ",")
		if len(parts) != 3 {
			return tcell.ColorDefault
		}

		r, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		g, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		b, err3 := strconv.Atoi(strings.TrimSpace(parts[2]))

		if err1 == nil && err2 == nil && err3 == nil {
			return RGBToColor(r, g, b)
		}
	}

	return tcell.ColorDefault
}

// ColorToStyle creates a style with a specific foreground color
func ColorToStyle(fgColor tcell.Color) tcell.Style {
	return tcell.StyleDefault.Foreground(fgColor)
}

// ColorPairToStyle creates a style with specific foreground and background colors
func ColorPairToStyle(fgColor, bgColor tcell.Color) tcell.Style {
	return tcell.StyleDefault.Foreground(fgColor).Background(bgColor)
}
