package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
)

// CalendarContextMode defines how the calendar should behave when a date is selected
type CalendarContextMode int

const (
	CalendarSearchMode CalendarContextMode = iota
	CalendarAttributeMode
)

// CalendarWidget provides an interactive calendar for date selection
type CalendarWidget struct {
	visible          bool
	currentMonth     time.Time
	selectedDate     time.Time
	items            []*model.Item
	contextMode      CalendarContextMode
	attributeName    string
	onDateSelected   func(time.Time)
	onCreateDateNode func(time.Time)
	weekStart        int // 0=Sunday, 1=Monday, etc.

	// Cached bounds for mouse handling
	boxStartX, boxStartY int
	boxWidth, boxHeight  int
}

// NewCalendarWidget creates a new calendar widget
func NewCalendarWidget() *CalendarWidget {
	return &CalendarWidget{
		visible:      false,
		currentMonth: time.Now(),
		selectedDate: time.Now(),
		contextMode:  CalendarSearchMode,
		weekStart:    0, // Default to Sunday
	}
}

// Show displays the calendar
func (w *CalendarWidget) Show() {
	w.visible = true
	w.currentMonth = time.Now()
	w.selectedDate = time.Now()
	w.contextMode = CalendarSearchMode
}

// ShowForAttribute displays the calendar in attribute mode
func (w *CalendarWidget) ShowForAttribute(attrName string) {
	w.visible = true
	startDate := time.Now()
	w.contextMode = CalendarAttributeMode
	w.attributeName = attrName
	w.currentMonth = startDate
	w.selectedDate = startDate
}

// ShowForAttributeWithValue displays the calendar in attribute mode with a specific starting date
func (w *CalendarWidget) ShowForAttributeWithValue(attrName string, dateStr string) {
	w.visible = true
	startDate := time.Now()

	// Try to parse the date string if provided
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			startDate = parsed
		}
	}

	w.contextMode = CalendarAttributeMode
	w.attributeName = attrName
	w.currentMonth = startDate
	w.selectedDate = startDate
}

// Hide hides the calendar
func (w *CalendarWidget) Hide() {
	w.visible = false
}

// IsVisible returns whether the calendar is currently shown
func (w *CalendarWidget) IsVisible() bool {
	return w.visible
}

// SetItems sets the items to check for dates
func (w *CalendarWidget) SetItems(items []*model.Item) {
	w.items = items
}

// SetOnDateSelected sets the callback for when a date is selected with Enter
func (w *CalendarWidget) SetOnDateSelected(fn func(time.Time)) {
	w.onDateSelected = fn
}

// SetOnCreateDateNode sets the callback for creating a new item with the selected date
func (w *CalendarWidget) SetOnCreateDateNode(fn func(time.Time)) {
	w.onCreateDateNode = fn
}

// GetContextMode returns the current context mode
func (w *CalendarWidget) GetContextMode() CalendarContextMode {
	return w.contextMode
}

// GetAttributeName returns the attribute name for attribute mode
func (w *CalendarWidget) GetAttributeName() string {
	return w.attributeName
}

// SetWeekStart sets the day the week starts on (0=Sunday, 1=Monday, ..., 6=Saturday)
func (w *CalendarWidget) SetWeekStart(day int) {
	if day < 0 || day > 6 {
		return // Invalid day, ignore
	}
	w.weekStart = day
}

// GetWeekStart returns the day the week starts on
func (w *CalendarWidget) GetWeekStart() int {
	return w.weekStart
}

// handleMonthNavigation updates currentMonth if selectedDate has moved outside it
func (w *CalendarWidget) handleMonthNavigation() {
	// Get first and last day of current month
	firstDay := time.Date(w.currentMonth.Year(), w.currentMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)

	// If selected date is before the month, go to previous months
	if w.selectedDate.Before(firstDay) {
		w.currentMonth = w.selectedDate
		return
	}

	// If selected date is after the month, go to next months
	if w.selectedDate.After(time.Date(w.currentMonth.Year(), w.currentMonth.Month(), lastDay.Day(), 23, 59, 59, 0, time.Local)) {
		w.currentMonth = w.selectedDate
		return
	}
}

// HandleKeyEvent processes keyboard input
func (w *CalendarWidget) HandleKeyEvent(ev *tcell.EventKey) bool {
	if !w.visible {
		return false
	}

	// Check for special keys first
	switch ev.Key() {
	case tcell.KeyEscape:
		w.Hide()
		return true
	case tcell.KeyEnter:
		if w.onDateSelected != nil {
			w.onDateSelected(w.selectedDate)
		}
		w.Hide()
		return true
	case tcell.KeyLeft:
		w.selectedDate = w.selectedDate.AddDate(0, 0, -1) // Previous day
		w.handleMonthNavigation()
		return true
	case tcell.KeyRight:
		w.selectedDate = w.selectedDate.AddDate(0, 0, 1) // Next day
		w.handleMonthNavigation()
		return true
	case tcell.KeyUp:
		w.selectedDate = w.selectedDate.AddDate(0, 0, -7) // Previous week
		w.handleMonthNavigation()
		return true
	case tcell.KeyDown:
		w.selectedDate = w.selectedDate.AddDate(0, 0, 7) // Next week
		w.handleMonthNavigation()
		return true
	}

	// Handle rune keys
	switch ev.Rune() {
	case 'h':
		w.selectedDate = w.selectedDate.AddDate(0, 0, -1) // Day backward
		w.handleMonthNavigation()
		return true
	case 'l':
		w.selectedDate = w.selectedDate.AddDate(0, 0, 1) // Day forward
		w.handleMonthNavigation()
		return true
	case 'j':
		w.selectedDate = w.selectedDate.AddDate(0, 0, 7) // Week forward
		w.handleMonthNavigation()
		return true
	case 'k':
		w.selectedDate = w.selectedDate.AddDate(0, 0, -7) // Week backward
		w.handleMonthNavigation()
		return true
	case 'H': // Shift+h - Previous year
		w.currentMonth = w.currentMonth.AddDate(-1, 0, 0)
		return true
	case 'L': // Shift+l - Next year
		w.currentMonth = w.currentMonth.AddDate(1, 0, 0)
		return true
	case 'J': // Shift+j - Next month
		w.currentMonth = w.currentMonth.AddDate(0, 1, 0)
		return true
	case 'K': // Shift+k - Previous month
		w.currentMonth = w.currentMonth.AddDate(0, -1, 0)
		return true
	case 't':
		w.selectedDate = time.Now()
		w.currentMonth = time.Now()
		return true
	}

	return false
}

// HandleMouseEvent processes mouse clicks
func (w *CalendarWidget) HandleMouseEvent(x, y int) bool {
	if !w.visible || !w.IsPointInside(x, y) {
		return false
	}

	// Check if click is on navigation arrows
	// Previous year button: << at (boxStartX+2, boxStartY+1)
	if y == w.boxStartY+1 {
		if x >= w.boxStartX+2 && x < w.boxStartX+5 {
			w.currentMonth = w.currentMonth.AddDate(-1, 0, 0)
			return true
		}
		// Previous month button: <
		if x >= w.boxStartX+6 && x < w.boxStartX+8 {
			w.currentMonth = w.currentMonth.AddDate(0, -1, 0)
			return true
		}
		// Next month button: >
		if x >= w.boxWidth-8+w.boxStartX && x < w.boxWidth-6+w.boxStartX {
			w.currentMonth = w.currentMonth.AddDate(0, 1, 0)
			return true
		}
		// Next year button: >>
		if x >= w.boxWidth-5+w.boxStartX && x < w.boxWidth-2+w.boxStartX {
			w.currentMonth = w.currentMonth.AddDate(1, 0, 0)
			return true
		}
	}

	// Calculate date at position
	date := w.GetDateAtPosition(x, y)
	if date != nil {
		w.selectedDate = *date
		return true
	}

	return false
}

// IsPointInside checks if a point is within the calendar bounds
func (w *CalendarWidget) IsPointInside(x, y int) bool {
	return x >= w.boxStartX && x < w.boxStartX+w.boxWidth &&
		y >= w.boxStartY && y < w.boxStartY+w.boxHeight
}

// GetDateAtPosition returns the date at the given screen coordinates, or nil
func (w *CalendarWidget) GetDateAtPosition(x, y int) *time.Time {
	// Calendar grid starts at boxStartY+5, with 2 pixels per row, 8 pixels per column
	startX := w.boxStartX + 4
	startY := w.boxStartY + 5

	// Check if within grid bounds
	if x < startX || x >= startX+56 || y < startY || y >= startY+12 {
		return nil
	}

	// Calculate column and row
	col := (x - startX) / 8
	row := (y - startY) / 2

	if col < 0 || col > 6 || row < 0 || row > 5 {
		return nil
	}

	// Get first day of month
	firstDay := time.Date(w.currentMonth.Year(), w.currentMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Calculate starting column, accounting for weekStart setting
	// This must match the calculation in drawCalendarGrid
	dayOfWeek := int(firstDay.Weekday())
	startCol := (dayOfWeek - w.weekStart + 7) % 7

	// Calculate day number from grid position
	dayNum := row*7 + col - startCol + 1

	if dayNum < 1 || dayNum > lastDay.Day() {
		return nil
	}

	date := time.Date(w.currentMonth.Year(), w.currentMonth.Month(), dayNum, 0, 0, 0, 0, time.Local)
	return &date
}

// Render draws the calendar widget to the screen
func (w *CalendarWidget) Render(screen *Screen) {
	if !w.visible {
		return
	}

	width := screen.GetWidth()
	height := screen.GetHeight()

	// Calendar dimensions
	w.boxWidth = 60
	w.boxHeight = 20
	w.boxStartX = (width - w.boxWidth) / 2
	w.boxStartY = (height - w.boxHeight) / 2

	// Get styles from theme
	borderStyle := screen.TreeNormalStyle()
	bgStyle := screen.BackgroundStyle()
	selectedStyle := screen.TreeSelectedStyle()
	todayStyle := screen.GreenStyle()
	dayStyle := screen.CalendarDayStyle()
	inactiveDayStyle := screen.CalendarInactiveDayStyle()
	indicatorStyle := screen.CalendarDayIndicatorStyle()

	// Bounds checking
	maxX := width - 1
	maxY := height - 1

	// Draw background
	for y := w.boxStartY; y < w.boxStartY+w.boxHeight && y <= maxY; y++ {
		for x := w.boxStartX; x < w.boxStartX+w.boxWidth && x <= maxX; x++ {
			screen.SetCell(x, y, ' ', bgStyle)
		}
	}

	// Draw borders
	w.drawBorders(screen, borderStyle, maxX, maxY)

	// Draw navigation and title
	w.drawTitle(screen, borderStyle, maxX, maxY)

	// Draw weekday headers
	w.drawWeekdayHeaders(screen, borderStyle, maxX, maxY)

	// Draw calendar grid
	w.drawCalendarGrid(screen, borderStyle, selectedStyle, todayStyle, dayStyle, inactiveDayStyle, indicatorStyle, maxX, maxY)

	// Draw footer with keybindings
	w.drawFooter(screen, borderStyle, maxX, maxY)
}

func (w *CalendarWidget) drawBorders(screen *Screen, style tcell.Style, maxX, maxY int) {
	// Top and bottom
	for x := w.boxStartX; x < w.boxStartX+w.boxWidth && x <= maxX; x++ {
		screen.SetCell(x, w.boxStartY, '─', style)
		if w.boxStartY+w.boxHeight-1 <= maxY {
			screen.SetCell(x, w.boxStartY+w.boxHeight-1, '─', style)
		}
	}

	// Left and right
	for y := w.boxStartY; y < w.boxStartY+w.boxHeight && y <= maxY; y++ {
		screen.SetCell(w.boxStartX, y, '│', style)
		if w.boxStartX+w.boxWidth-1 <= maxX {
			screen.SetCell(w.boxStartX+w.boxWidth-1, y, '│', style)
		}
	}

	// Corners
	screen.SetCell(w.boxStartX, w.boxStartY, '┌', style)
	if w.boxStartX+w.boxWidth-1 <= maxX {
		screen.SetCell(w.boxStartX+w.boxWidth-1, w.boxStartY, '┐', style)
	}
	if w.boxStartY+w.boxHeight-1 <= maxY {
		screen.SetCell(w.boxStartX, w.boxStartY+w.boxHeight-1, '└', style)
	}
	if w.boxStartX+w.boxWidth-1 <= maxX && w.boxStartY+w.boxHeight-1 <= maxY {
		screen.SetCell(w.boxStartX+w.boxWidth-1, w.boxStartY+w.boxHeight-1, '┘', style)
	}
}

func (w *CalendarWidget) drawTitle(screen *Screen, style tcell.Style, maxX, maxY int) {
	y := w.boxStartY + 1
	if y > maxY {
		return
	}

	// Draw navigation arrows and title
	// Format: << < Month YYYY > >>
	navStart := w.boxStartX + 2

	// Previous year
	if navStart < maxX {
		screen.SetCell(navStart, y, '<', style)
	}
	if navStart+1 <= maxX {
		screen.SetCell(navStart+1, y, '<', style)
	}

	// Previous month
	if navStart+3 <= maxX {
		screen.SetCell(navStart+3, y, '<', style)
	}

	// Title
	title := w.currentMonth.Format("January 2006")
	titleX := w.boxStartX + (w.boxWidth-len(title))/2
	screen.DrawString(titleX, y, title, style.Bold(true))

	// Next month
	nextMonthX := w.boxStartX + w.boxWidth - 6
	if nextMonthX <= maxX {
		screen.SetCell(nextMonthX, y, '>', style)
	}

	// Next year
	if nextMonthX+1 <= maxX {
		screen.SetCell(nextMonthX+1, y, '>', style)
	}
	if nextMonthX+2 <= maxX {
		screen.SetCell(nextMonthX+2, y, '>', style)
	}
	if nextMonthX+3 <= maxX {
		screen.SetCell(nextMonthX+3, y, '>', style)
	}
}

func (w *CalendarWidget) drawWeekdayHeaders(screen *Screen, style tcell.Style, maxX, maxY int) {
	allWeekdays := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	startX := w.boxStartX + 4
	y := w.boxStartY + 3

	if y > maxY {
		return
	}

	// Rotate weekdays based on weekStart
	weekdays := make([]string, 7)
	for i := 0; i < 7; i++ {
		weekdays[i] = allWeekdays[(i+w.weekStart)%7]
	}

	for i, day := range weekdays {
		x := startX + i*8
		if x+len(day) <= maxX {
			screen.DrawString(x, y, day, style.Bold(true))
		}
	}
}

func (w *CalendarWidget) drawCalendarGrid(screen *Screen, borderStyle, selectedStyle, todayStyle, dayStyle, inactiveDayStyle, indicatorStyle tcell.Style, maxX, maxY int) {
	// Get first day of month
	firstDay := time.Date(w.currentMonth.Year(), w.currentMonth.Month(), 1, 0, 0, 0, 0, time.Local)

	// Get last day of month
	lastDay := firstDay.AddDate(0, 1, -1)

	// Start position
	startX := w.boxStartX + 4
	startY := w.boxStartY + 5

	// Calculate starting column based on weekStart
	// firstDay.Weekday() returns 0=Sunday, 1=Monday, etc.
	// We adjust it so that the column depends on weekStart
	dayOfWeek := int(firstDay.Weekday())
	startCol := (dayOfWeek - w.weekStart + 7) % 7

	today := time.Now()

	// First, fill all 42 cells (6 weeks × 7 days) with appropriate style
	for cellRow := 0; cellRow < 6; cellRow++ {
		for cellCol := 0; cellCol < 7; cellCol++ {
			x := startX + cellCol*8
			y := startY + cellRow*2

			if y > maxY || x+8 > maxX {
				continue
			}

			// Calculate day number for this cell
			dayNum := cellRow*7 + cellCol - startCol + 1

			// Determine if this is an active day (in current month)
			isActiveDay := dayNum >= 1 && dayNum <= lastDay.Day()
			cellStyle := dayStyle
			if !isActiveDay {
				cellStyle = inactiveDayStyle
			}

			// Draw 6 spaces as background for each cell (leaving 2 for separator)
			for dx := 0; dx < 6; dx++ {
				if x+dx <= maxX {
					screen.SetCell(x+dx, y, ' ', cellStyle)
				}
			}
		}
	}

	// Second pass: highlight selected dates with selected style background
	for day := 1; day <= lastDay.Day(); day++ {
		currentDate := time.Date(w.currentMonth.Year(), w.currentMonth.Month(), day, 0, 0, 0, 0, time.Local)

		col := (startCol + day - 1) % 7
		row := (startCol + day - 1) / 7

		x := startX + col*8
		y := startY + row*2

		if y > maxY {
			break
		}

		// Check if this date is selected
		if w.selectedDate.Year() == currentDate.Year() &&
			w.selectedDate.Month() == currentDate.Month() &&
			w.selectedDate.Day() == currentDate.Day() {
			// Fill all 6 spaces with selected style
			for dx := 0; dx < 6; dx++ {
				if x+dx <= maxX {
					screen.SetCell(x+dx, y, ' ', selectedStyle)
				}
			}
		}
	}

	// Third pass: draw the day numbers
	for day := 1; day <= lastDay.Day(); day++ {
		currentDate := time.Date(w.currentMonth.Year(), w.currentMonth.Month(), day, 0, 0, 0, 0, time.Local)

		col := (startCol + day - 1) % 7
		row := (startCol + day - 1) / 7

		x := startX + col*8
		y := startY + row*2

		if y > maxY {
			break
		}

		// Determine style for text
		style := dayStyle

		// Check if this date is selected
		if w.selectedDate.Year() == currentDate.Year() &&
			w.selectedDate.Month() == currentDate.Month() &&
			w.selectedDate.Day() == currentDate.Day() {
			style = selectedStyle
		} else if today.Year() == currentDate.Year() &&
			today.Month() == currentDate.Month() &&
			today.Day() == currentDate.Day() {
			style = todayStyle.Bold(true)
		}

		// Draw day number (right-padded to 2 chars, leaving room for dot)
		dayStr := fmt.Sprintf("%3d", day)
		if x+len(dayStr) <= maxX {
			screen.DrawString(x, y, dayStr, style)
		}

		// Check if there are items on this date and draw filled circle (●)
		if w.hasItemsOnDate(currentDate) {
			dotX := x + 4
			if dotX <= maxX {
				// Use indicator style with day background, or selected style if selected
				dotStyle := indicatorStyle
				if w.selectedDate.Year() == currentDate.Year() &&
					w.selectedDate.Month() == currentDate.Month() &&
					w.selectedDate.Day() == currentDate.Day() {
					dotStyle = selectedStyle
				}
				screen.SetCell(dotX, y, '●', dotStyle)
			}
		}
	}
}

func (w *CalendarWidget) drawFooter(screen *Screen, style tcell.Style, maxX, maxY int) {
	footerY := w.boxStartY + w.boxHeight - 2
	if footerY > maxY {
		return
	}

	footer := "h/l:day j/k:week J/K:month H/L:year | Enter:select/create"
	footerX := w.boxStartX + 2

	if footerX+len(footer) <= maxX {
		screen.DrawString(footerX, footerY, footer, style.Dim(true))
	}
}

// hasItemsOnDate checks if any items exist on the given date
func (w *CalendarWidget) hasItemsOnDate(date time.Time) bool {
	dateStr := date.Format("2006-01-02")

	for _, item := range w.items {
		if item == nil || item.Metadata == nil {
			continue
		}

		// Check date attribute
		if item.Metadata.Attributes != nil {
			if itemDate, ok := item.Metadata.Attributes["date"]; ok {
				if itemDate == dateStr {
					return true
				}
			}
		}

		// Check if item text is a date
		if item.Text == dateStr {
			return true
		}
	}

	return false
}
