package ui

import (
	"testing"
)

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected int
	}{
		// ASCII
		{"ASCII letter", 'A', 1},
		{"ASCII space", ' ', 1},
		{"ASCII digit", '5', 1},

		// Wide characters
		{"Emoji", '😀', 2},
		{"Chinese character", '中', 2},
		{"Japanese hiragana", 'あ', 2},
		{"Korean hangul", '한', 2},

		// Combining marks
		{"Combining acute", '\u0301', 0},
		{"Zero width joiner", '\u200d', 0},

		// Control characters
		{"Tab", '\t', 0},
		{"Newline", '\n', 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RuneWidth(tt.r)
			if got != tt.expected {
				t.Errorf("RuneWidth(%q) = %d, want %d", tt.r, got, tt.expected)
			}
		})
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// ASCII
		{"ASCII only", "Hello", 5},
		{"ASCII with spaces", "Hello World", 11},

		// Mixed ASCII and emoji
		{"Emoji with text", "😀 Hello", 8}, // 2 + 1 + 5
		{"Multiple emoji", "😀😀", 4},

		// CJK characters
		{"Chinese", "中国", 4},
		{"Japanese", "こんにちは", 10},
		{"Mixed CJK and ASCII", "Hello中国", 9}, // 5 + 4

		// Edge cases
		{"Empty string", "", 0},
		{"Single ASCII", "a", 1},
		{"Single emoji", "😀", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringWidth(tt.input)
			if got != tt.expected {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxWidth  int
		expected  string
		expWidth  int
	}{
		// ASCII
		{"ASCII fits", "Hello", 10, "Hello", 5},
		{"ASCII truncated", "Hello", 3, "Hel", 3},
		{"ASCII exact", "Hello", 5, "Hello", 5},

		// Emoji - ensure we don't split them
		{"Emoji fits", "😀Hi", 10, "😀Hi", 4},
		{"Emoji truncated before", "😀Hello", 2, "😀", 2},
		{"Emoji truncated after", "Hi😀", 3, "Hi", 2},
		{"Multiple emoji truncated", "😀😀😀", 5, "😀😀", 4},

		// CJK
		{"Chinese fits", "中国", 10, "中国", 4},
		{"Chinese truncated", "中国", 2, "中", 2},

		// Mixed
		{"Mixed fits", "Hello中国", 20, "Hello中国", 9},
		{"Mixed truncated at ASCII", "Hello中国", 4, "Hell", 4},
		{"Mixed truncated before CJK", "Hello中国", 5, "Hello", 5},

		// Edge cases
		{"Empty string", "", 5, "", 0},
		{"MaxWidth 0", "Hello", 0, "", 0},
		{"MaxWidth negative", "Hello", -1, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateToWidth(tt.input, tt.maxWidth)
			if got != tt.expected {
				t.Errorf("TruncateToWidth(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.expected)
			}

			// Also verify width
			width := StringWidth(got)
			if width != tt.expWidth {
				t.Errorf("Result width %d, want %d", width, tt.expWidth)
			}
		})
	}
}

func TestTruncateToWidthWithEllipsis(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxWidth  int
		expected  string
		checkEllipsis bool
	}{
		// String that fits
		{"Fits no ellipsis", "Hello", 10, "Hello", false},

		// String that needs truncation
		{"Long ASCII", "HelloWorld", 5, "", true},
		{"Long with emoji", "😀HelloWorld", 7, "", true},

		// Edge cases
		{"MaxWidth 5", "HelloWorld", 5, "..", true},
		{"MaxWidth 2", "HelloWorld", 2, "", false},
		{"Empty string", "", 5, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateToWidthWithEllipsis(tt.input, tt.maxWidth)

			if tt.checkEllipsis {
				// Just check that it ends with ellipsis
				if len(got) < 3 || got[len(got)-3:] != "..." {
					t.Errorf("Result should end with '...': %q", got)
				}
			}

			width := StringWidth(got)
			if width > tt.maxWidth {
				t.Errorf("Result width %d exceeds maxWidth %d", width, tt.maxWidth)
			}
		})
	}
}

func TestPadStringToWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		// ASCII
		{"ASCII shorter", "Hi", 5, "Hi   "},
		{"ASCII exact", "Hello", 5, "Hello"},
		{"ASCII longer", "Hello", 3, "Hello"},

		// Emoji (width 2)
		{"Emoji shorter", "😀", 5, "😀   "},
		{"Emoji exact", "😀Hi", 4, "😀Hi"},

		// CJK
		{"CJK shorter", "中", 5, "中   "},

		// Edge cases
		{"Empty string", "", 5, "     "},
		{"Width 0", "Hi", 0, "Hi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PadStringToWidth(tt.input, tt.width)
			if got != tt.expected {
				t.Errorf("PadStringToWidth(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expected)
			}

			width := StringWidth(got)
			if width < tt.width {
				t.Errorf("Result width %d less than requested %d", width, tt.width)
			}
		})
	}
}

func TestFindRuneIndexAtWidth(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		targetWidth int
		expected    int
	}{
		// ASCII
		{"ASCII start", "Hello", 0, 0},
		{"ASCII middle", "Hello", 2, 2},
		{"ASCII end", "Hello", 5, 5},
		{"ASCII beyond", "Hello", 10, 5},

		// Emoji (2 columns each, 4 bytes each)
		{"Emoji start", "😀😀😀", 0, 0},
		{"Emoji after first", "😀😀😀", 2, 4},
		{"Emoji after second", "😀😀😀", 4, 8},
		{"Emoji beyond", "😀😀😀", 6, 12},

		// Mixed (H=1byte+1width, 😀=4bytes+2width)
		{"Mixed ASCII start", "H😀lo", 0, 0},
		{"Mixed ASCII then emoji", "H😀lo", 1, 1},
		{"Mixed after emoji", "H😀lo", 3, 5},

		// Edge cases
		{"Empty string", "", 5, 0},
		{"Width 0", "Hello", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindRuneIndexAtWidth(tt.input, tt.targetWidth)
			if got != tt.expected {
				t.Errorf("FindRuneIndexAtWidth(%q, %d) = %d, want %d", tt.input, tt.targetWidth, got, tt.expected)
			}
		})
	}
}

func TestCalculateBreakPoint(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxWidth  int
		expIdx    int
		expWidth  int
	}{
		// ASCII - word boundaries preferred
		{"ASCII no space", "HelloWorld", 5, 5, 5},
		{"ASCII with space", "Hello World", 6, 6, 6}, // "Hello "
		{"ASCII space boundary", "Hello World", 5, 5, 5},
		{"ASCII multiple words", "Hello Beautiful World", 8, 6, 6}, // "Hello "

		// Emoji
		{"Single emoji", "😀Hello", 3, 5, 3},
		{"Emoji and words", "Hi 😀 World", 4, 3, 3}, // "Hi " (before emoji)

		// Edge cases
		{"Fits entirely", "Hello", 10, 5, 5},
		{"Empty string", "", 5, 0, 0},
		{"MaxWidth 0", "Hello", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIdx, gotWidth := CalculateBreakPoint(tt.input, tt.maxWidth)
			if gotIdx != tt.expIdx || gotWidth != tt.expWidth {
				t.Errorf("CalculateBreakPoint(%q, %d) = (%d, %d), want (%d, %d)",
					tt.input, tt.maxWidth, gotIdx, gotWidth, tt.expIdx, tt.expWidth)
			}
		})
	}
}

func TestWordBoundaryIndex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		pos      int
		next     bool
		expected int
	}{
		// Forward (next=true)
		{"ASCII next from start", "Hello World", 0, true, 6},
		{"ASCII next from word", "Hello World", 2, true, 6},
		{"ASCII next from space", "Hello World", 5, true, 11},
		{"ASCII next at end", "Hello World", 11, true, 11},

		// Backward (next=false)
		{"ASCII prev from space", "Hello World", 6, false, 6},
		{"ASCII prev from word", "Hello World", 8, false, 6},
		{"ASCII prev at start", "Hello World", 0, false, 0},

		// Edge cases
		{"Empty string next", "", 0, true, 0},
		{"Empty string prev", "", 0, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordBoundaryIndex(tt.input, tt.pos, tt.next)
			if got != tt.expected {
				t.Errorf("WordBoundaryIndex(%q, %d, %v) = %d, want %d",
					tt.input, tt.pos, tt.next, got, tt.expected)
			}
		})
	}
}

func TestStringWidthUpTo(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxWidth  int
		expWidth  int
		expRunes  int
	}{
		// ASCII
		{"ASCII fits", "Hello", 10, 5, 5},
		{"ASCII partial", "Hello", 3, 3, 3},
		{"ASCII exact", "Hello", 5, 5, 5},

		// Emoji
		{"Emoji fits", "😀Hi", 10, 4, 3},
		{"Emoji partial", "😀Hello", 3, 3, 5},

		// Mixed
		{"Mixed partial", "Hi😀World", 5, 5, 7},

		// Edge cases
		{"Empty string", "", 5, 0, 0},
		{"MaxWidth 0", "Hello", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidth, gotRunes := StringWidthUpTo(tt.input, tt.maxWidth)
			if gotWidth != tt.expWidth || gotRunes != tt.expRunes {
				t.Errorf("StringWidthUpTo(%q, %d) = (%d, %d), want (%d, %d)",
					tt.input, tt.maxWidth, gotWidth, gotRunes, tt.expWidth, tt.expRunes)
			}
		})
	}
}
