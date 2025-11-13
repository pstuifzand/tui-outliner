package template

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ncruces/go-strftime"
)

// DateValue is used to pass dates through the pipe chain
type DateValue struct {
	t time.Time
}

// Now returns the current time in ISO 8601 format
func Now() (string, error) {
	return time.Now().Format(time.RFC3339), nil
}

// DateFormat returns the current date formatted according to strftime format string
// Example: DateFormat("%Y-%m-%d") returns "2024-11-13"
// Common formats:
//   %Y - 4-digit year
//   %m - 2-digit month (01-12)
//   %d - 2-digit day (01-31)
//   %H - hour (00-23)
//   %M - minute (00-59)
//   %S - second (00-59)
//   %A - full weekday name
//   %B - full month name
func DateFormat(format string) (string, error) {
	if format == "" {
		format = "%Y-%m-%d"
	}
	t := time.Now()
	result := strftime.Format(format, t)
	return result, nil
}

// Clipboard reads from the system clipboard using a configurable command
// Default command is "wl-paste" for Wayland. Configure via ClipboardCommand.
// Common values:
//   - "wl-paste" for Wayland
//   - "xclip -selection clipboard -o" for X11
//   - "pbpaste" for macOS
func Clipboard(command string) (string, error) {
	if command == "" {
		command = "wl-paste"
	}

	// If user specified wl-copy, convert to wl-paste for reading
	if command == "wl-copy" {
		command = "wl-paste"
	}

	cmd := exec.Command(command)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("clipboard read error: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Weekday returns the date for the specified weekday in the current calendar week
// The current week is defined as the 7-day period containing today
// Week start day is configurable (default: 1 = Monday)
// dayNum: 0=Sunday, 1=Monday, 2=Tuesday, ..., 6=Saturday
// Returns a DateValue that can be piped to date formatting
// Example: If today is Thursday and weekstart=1 (Monday):
//   {{weekday(1)|date:%Y-%m-%d}} returns Monday of the current week
//   {{weekday(4)|date:%Y-%m-%d}} returns Thursday (today)
func WeekdayWithStart(dayNum int, weekStart int) (*DateValue, error) {
	now := time.Now()
	// Go's time.Weekday: 0=Sunday, 1=Monday, ..., 6=Saturday
	currentWeekday := int(now.Weekday())

	// Ensure weekStart is valid (0-6)
	weekStart = weekStart % 7
	if weekStart < 0 {
		weekStart += 7
	}

	// Calculate days to subtract to get to the start of the current week
	daysToWeekStart := (currentWeekday - weekStart + 7) % 7
	weekStartDate := now.AddDate(0, 0, -daysToWeekStart)

	// Now add the offset for the desired weekday
	daysToTarget := (dayNum - weekStart + 7) % 7
	targetDate := weekStartDate.AddDate(0, 0, daysToTarget)

	return &DateValue{t: targetDate}, nil
}

// Weekday is the default function using configuration
// Signature matches the template syntax: weekday(dayNum)
func Weekday(dayNum int) (*DateValue, error) {
	// Default week start is Monday (1)
	// In the future, this should read from config
	return WeekdayWithStart(dayNum, 1)
}

// FormatDateValue formats a DateValue according to strftime format
// This is used in pipe chains like {{weekday(1)|date:%Y-%m-%d}}
func FormatDateValue(dv *DateValue, format string) (string, error) {
	if dv == nil {
		return "", fmt.Errorf("no date value to format")
	}
	if format == "" {
		format = "%Y-%m-%d"
	}
	result := strftime.Format(format, dv.t)
	return result, nil
}

// parseOptions parses a comma-separated list of options, handling quoted strings
// Example: "option1,option2,\"Option with, comma\""
func parseOptions(args string) []string {
	var options []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(args); i++ {
		char := args[i]

		if char == '"' {
			inQuotes = !inQuotes
		} else if char == ',' && !inQuotes {
			opt := strings.TrimSpace(current.String())
			if opt != "" {
				options = append(options, opt)
			}
			current.Reset()
		} else {
			current.WriteByte(char)
		}
	}

	// Add last option
	opt := strings.TrimSpace(current.String())
	if opt != "" {
		options = append(options, opt)
	}

	return options
}
