# TUI-Outliner UTF-8 and Unicode Character Handling Analysis

## Executive Summary

The tui-outliner codebase has **significant and pervasive UTF-8 handling issues** despite having `go-runewidth` as a dependency (which is **never actually used in the code**). The application uses string byte-length operations (`len(text)`) throughout the codebase for width calculations, which fails for multi-byte characters like emoji, CJK (Chinese/Japanese/Korean), and other Unicode characters. This causes text to misalign, misrender, and cause incorrect cursor positioning, especially with wide characters (2+ columns) and combining characters.

---

## Current UTF-8 Support Status

### Positive Aspects:
1. **tcell v2 (indirect) and go-runewidth v0.0.16** are included as dependencies
2. **tcell itself properly handles UTF-8** when using `SetContent()` with rune values
3. **JSON storage preserves UTF-8 correctly** - text is stored as UTF-8 strings without corruption
4. **Basic ASCII editing works** - all keybindings and navigation work for ASCII text

### Critical Issues:
1. **No use of `go-runewidth` library** - dependency exists but never imported/called
2. **Byte-length-based width calculations throughout** - `len(string)` returns byte count, not character count
3. **String byte-indexing for mouse and cursor positioning** - assumes 1 byte = 1 column
4. **Text wrapping uses byte lengths** - wraps at wrong positions for multi-byte text
5. **Attribute rendering uses byte counts** - overflow calculations fail
6. **Editor cursor positioning is byte-based** - clicks misaligned with wide characters
7. **Multi-line editor assumes single-byte characters** - wrapped line offset calculations wrong
8. **Text slicing without rune awareness** - can split multi-byte sequences or combining characters

---

## Detailed Problem Areas

### 1. **String Rendering (Screen.DrawString)** - CRITICAL
**File**: `internal/ui/screen.go` lines 88-92

```go
func (s *Screen) DrawString(x, y int, text string, style tcell.Style) {
	for i, r := range text {
		s.SetCell(x+i, y, r, style)
	}
}
```

**Problem**: 
- Uses `for i, r := range text` which iterates over **Unicode code points**, not bytes
- However, position calculation `x+i` treats index as column offset
- For multi-byte characters (e.g., emoji "😀" = 4 bytes, 2 display columns), this places characters at wrong column
- A single emoji at position 0 would render at columns 0 and 1, but next char placed at column 1 (overlap)

**Impact**: High - All text rendering is affected
**Examples**:
- Emoji: "😀test" renders incorrectly (emoji takes 2 columns but code positions next char at column 1)
- CJK: "你好test" (Chinese characters take 2 columns each)
- Combining: "é" (e + acute combining) wrong column count

**Recommendation**: Use `runewidth.StringWidth(text)` and loop through runes with width tracking

---

### 2. **Text Wrapping at Byte Boundaries** - CRITICAL
**File**: `internal/ui/tree.go` lines 233-273 (`wrapTextAtWidth` function)

```go
func wrapTextAtWidth(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}
	
	for len(remaining) > maxWidth {
		lastSpace := -1
		for i := 0; i < maxWidth && i < len(remaining); i++ {
			if remaining[i] == ' ' {
				lastSpace = i
			}
		}
		
		if lastSpace > 0 && lastSpace < maxWidth {
			result = append(result, remaining[:lastSpace])
			remaining = strings.TrimPrefix(remaining[lastSpace:], " ")
		} else {
			result = append(result, remaining[:maxWidth])  // BUG: Byte-based slicing!
			remaining = remaining[maxWidth:]
		}
	}
}
```

**Problems**:
- `len(text) <= maxWidth` - compares byte length to column width (incompatible units)
- `for i := 0; i < maxWidth` - iterates bytes, not display columns
- `remaining[:maxWidth]` - slices at byte position, not column position
  - "你好world" (6 bytes for "你好", 5 for "world") truncated at byte 5 splits the second Chinese character in half
- `remaining[maxWidth:]` - can split UTF-8 sequences or combining characters

**Impact**: Very High - Causes visual corruption and potential crashes with invalid UTF-8
**Examples**:
- Text "😀😀😀😀😀" with maxWidth=10: Each emoji (4 bytes) wraps at wrong position
- Text "你好你好你好" with maxWidth=10: Gets split in middle of characters
- Mixed text "Hello你好World" - wrapping calculation completely wrong

**Recommendation**: Rewrite using rune iteration with `runewidth` for display width tracking

---

### 3. **Editor Cursor Positioning (Byte-Based)** - CRITICAL
**File**: `internal/ui/editor.go` lines 149-195, 257-265

```go
func (e *Editor) Render(screen *Screen, x, y int, maxWidth int) {
	displayText := e.text
	startIdx := 0
	if len(displayText) > maxWidth {
		startIdx = e.cursorPos - maxWidth/2
		// ... bounds checking ...
		displayText = displayText[startIdx:]
	}
	
	for i, r := range displayText {
		screen.SetCell(x+i, y, r, textStyle)
	}
}

func (e *Editor) SetCursorFromScreenX(relativeX int) {
	if relativeX > len(e.text) {
		relativeX = len(e.text)
	}
	e.cursorPos = relativeX  // BUG: Direct byte position!
}
```

**Problems**:
- Cursor position stored as **byte offset**, not column position
- Mouse click handler in `app.go` line 1843: `a.editor.SetCursorFromScreenX(relativeX)` passes screen column as byte offset
- Click on emoji places cursor at wrong byte position
- Text display position calculation `x+i` assumes 1 byte = 1 column

**Impact**: Very High - Mouse clicking and cursor position broken for wide characters
**Examples**:
- Click after "😀" appears to click at byte 2, but character occupies columns 0-1
- In "Hello你World", clicking between 你 and W places cursor wrong position
- Backspace/Delete operate at byte level, not character level (works, but weird with combining chars)

---

### 4. **MultiLineEditor Cursor and Wrapping** - CRITICAL
**File**: `internal/ui/multiline_editor.go` lines 64-110, 114-160

```go
func (mle *MultiLineEditor) calculateWrappedLines() {
	hardLines := strings.Split(mle.text, "\n")
	wrappedLines := []string{}
	lineStartOffsets := []int{}
	
	offset := 0
	for _, hardLine := range hardLines {
		wrappedParts := wrapTextAtWidth(hardLine, mle.maxWidth)
		// ...
		for _, part := range wrappedParts {
			lineStartOffsets = append(lineStartOffsets, offset+searchPos)
			wrappedLines = append(wrappedLines, part)
			searchPos += len(part)  // BUG: Byte length!
		}
		offset += len(hardLine) + 1  // BUG: Byte length!
	}
}

func (mle *MultiLineEditor) getCursorVisualPosition() (row int, col int) {
	for lineIdx, startOffset := range mle.lineStartOffsets {
		lineEnd := startOffset + len(mle.wrappedLines[lineIdx])  // BUG: Byte length!
		if mle.cursorPos >= startOffset && mle.cursorPos <= lineEnd {
			return lineIdx, mle.cursorPos - startOffset
		}
	}
}
```

**Problems**:
- All offset tracking uses **byte positions**, not character positions
- `searchPos += len(part)` - adds byte length
- `offset += len(hardLine) + 1` - byte-based tracking
- `lineEnd = startOffset + len(mle.wrappedLines[lineIdx])` - byte length comparison
- `mle.cursorPos - startOffset` - assumes 1 byte = 1 column
- Multi-byte characters break cursor positioning across wrapped lines

**Impact**: Very High - Multi-line editing completely broken with wide characters
**Examples**:
- Text "你好你好你好你好你好" wrapped at column 10: Line offsets calculated wrong
- Cursor navigation (Up/Down) between lines misaligned
- Rendered wrapped lines show correct visual positions but cursor/selection invisible

---

### 5. **Text Truncation Without Rune Awareness** - HIGH
**File**: `internal/ui/tree.go` lines 1779-1781, 1931-1933

```go
if len(text) > maxTextWidth {
	text = text[:maxTextWidth]
}
screen.DrawString(textX, y, text, lineStyle)
```

**Problems**:
- `text[:maxTextWidth]` slices at byte maxTextWidth
- Can split multi-byte characters in middle
- Creates invalid UTF-8 strings (tcell may handle gracefully but unpredictable)
- Ellipsis "..." not added, text just truncated

**Impact**: Medium - Text corruption and visual misalignment
**Examples**:
- "这是一个非常长的文本" truncated at byte 10: "这是一个非常" gets split
- No visual indication text was truncated

---

### 6. **Attribute String Length Calculation** - HIGH
**File**: `internal/ui/tree.go` lines 1828-1832

```go
attrStr := "  [" + strings.Join(visibleAttrs, ", ") + "]"
attrStyle := screen.TreeAttributeStyle().Background(bgColor)
attrX := totalLen
if attrX+len(attrStr) <= screenWidth {
	screen.DrawString(attrX, y, attrStr, attrStyle)
	totalLen = attrX + len(attrStr)  // BUG: Byte length!
}
```

**Problems**:
- `len(attrStr)` is byte length
- Comparison `attrX+len(attrStr) <= screenWidth` wrong unit mismatch
- `totalLen = attrX + len(attrStr)` - updates position using byte length
- Attribute names containing non-ASCII characters (e.g., "日期:2025-01-01") break layout
- Overflow calculations completely wrong

**Impact**: Medium - Attribute display corrupted with non-ASCII attribute names
**Examples**:
- Attribute key "日期" (4 bytes, 2 display columns) causes wrong layout
- Progress bar position calculation wrong when attributes present

---

### 7. **DrawStringLimited Truncation** - HIGH
**File**: `internal/ui/screen.go` lines 94-103

```go
func (s *Screen) DrawStringLimited(x, y int, text string, maxWidth int, style tcell.Style) {
	if len(text) > maxWidth {
		text = text[:maxWidth]
	}
	s.DrawString(x, y, text, style)
}
```

**Problems**:
- Compares byte length `len(text)` to column width `maxWidth`
- Slices at byte boundary `text[:maxWidth]`
- Can create invalid UTF-8

**Impact**: Medium - Any use of this function breaks with wide characters
**Examples**:
- Drawing a 4-byte emoji with maxWidth=1: tries `text[:1]` creating invalid UTF-8
- CJK text truncated at odd byte positions

---

### 8. **Mouse Click Position Calculation** - HIGH
**File**: `internal/app/app.go` lines 1816-1844

```go
func (a *App) handleEditorMouseClick(mouseEv *tcell.EventMouse) {
	x, _ := mouseEv.Position()
	
	depth := a.tree.GetSelectedDepth()
	editorX := depth*3 + 3
	
	if x >= editorX {
		relativeX := x - editorX  // Screen column position
		a.editor.SetCursorFromScreenX(relativeX)  // BUG: Passed as byte offset!
	}
}
```

**Problems**:
- `relativeX` is screen column offset (0, 1, 2, ...)
- Passed to `SetCursorFromScreenX()` which treats it as byte offset
- Click on emoji at screen position 1 sets cursor to byte 1 (inside 4-byte sequence)
- No rune width consideration

**Impact**: High - Mouse clicking in editor broken with wide characters
**Examples**:
- Text "😀test" - click at display column 2 (after emoji) goes to byte 2 (middle of emoji)
- Text "Hello你" - click at display column 6 (after 你) goes to byte 8 (middle of character)

---

### 9. **Text Search and Highlighting** - MEDIUM
**File**: `internal/ui/tree.go` lines 2678-2703

```go
currentX := x
for i, r := range displayText {
	// ... determine if in link, search match ...
	screen.SetCell(currentX, y, r, charStyle)
	currentX++  // BUG: Assumes 1 rune = 1 column!
}
```

**Problems**:
- `currentX++` increments by 1 for every rune
- Wide characters take 2+ columns but only increment by 1
- Search highlighting position calculation wrong
- Link range calculations use byte positions in displayText but rune iteration breaks alignment

**Impact**: Medium - Search results misaligned and incorrectly highlighted
**Examples**:
- Search highlighting wrong column for text after wide characters
- Link detection and highlighting breaks with emoji in text

---

### 10. **Editor String Slicing Operations** - MEDIUM
**File**: `internal/ui/editor.go` lines 86-88, 103-104, 112-113, 132-136

```go
// Line 86-88: Shift+Enter insert newline
e.text = e.text[:e.cursorPos] + "\n" + e.text[e.cursorPos:]
e.cursorPos++

// Line 103-104: Backspace
e.text = e.text[:e.cursorPos-1] + e.text[e.cursorPos:]
e.cursorPos--

// Line 112-113: Delete
e.text = e.text[:e.cursorPos] + e.text[e.cursorPos+1:]

// Line 132-136: Ctrl+U (delete to start)
e.text = e.text[e.cursorPos:]
e.cursorPos = 0
```

**Problems**:
- All assume `cursorPos` is accurate byte position
- With wide characters, `cursorPos` is already wrong (see #3)
- Slicing `[:e.cursorPos-1]` can split multi-byte character
- Deleting single rune requires code point boundary detection
- Combining characters: deleting base removes accent, leaving dangling combining mark

**Impact**: Medium-High - Text editing breaks with non-ASCII
**Examples**:
- In text "hello😀test", backspace at wrong position splits emoji
- Deleting in "naïve" (i + combining diaeresis) separates combining mark

---

### 11. **MultiLineEditor String Operations** - MEDIUM
**File**: `internal/ui/multiline_editor.go` lines 238-272, 550-575, etc.

```go
// Line 239-240: Shift+Enter
mle.text = mle.text[:mle.cursorPos] + "\n" + mle.text[mle.cursorPos:]
mle.cursorPos++

// Line 256: Backspace
mle.text = mle.text[:mle.cursorPos-1] + mle.text[mle.cursorPos:]

// Line 558-575: Delete word backwards
for pos >= 0 && mle.text[pos] == ' ' {
	pos--
}
```

**Problems**:
- All string slicing assumes correct `cursorPos` (which is already wrong)
- Word boundary detection using `mle.text[pos]` checks bytes, not runes
- Multi-byte characters corrupt word deletion
- Newline handling in multi-byte text wrong

**Impact**: Medium - Multi-line editing broken with non-ASCII
**Examples**:
- "你好 你好" - delete word backwards deletes wrong characters
- Mixed text "hello你好 test" - line splitting and word operations incorrect

---

### 12. **Column Position vs Byte Position Mismatch** - SYSTEMATIC
**Throughout codebase**:

Pattern: `x + i` where `i` is rune index or byte index
- `screen.go` line 90: `s.SetCell(x+i, y, r, style)`
- `editor.go` line 173: `screen.SetCell(x+i, y, r, textStyle)`
- `tree.go` line 2701: `screen.SetCell(currentX, y, r, charStyle)` with `currentX++`
- Many other places

**Problem**: Sets cell at column `x+i` for each rune, but wide runes need 2+ columns
**Impact**: All text rendering misaligned with wide characters

---

## Width Calculation Issues Summary

| Operation | Current Approach | Issue | Affected Characters |
|-----------|-----------------|-------|-------------------|
| Text wrapping | `len(text)` byte count | Breaks at byte boundaries | CJK, emoji, Latin extended |
| Text truncation | `text[:width]` byte slicing | Cuts mid-character | All non-ASCII |
| Display positioning | `x+i` with rune iteration | Wide chars get 1 column | CJK (2 cols), emoji (2 cols), combining |
| Cursor positioning | Byte offset in text | Wrong column for wide chars | All multi-byte |
| Mouse clicks | Screen column as byte offset | Click misaligned | All multi-byte |
| Attribute layout | `len(attrStr)` bytes | Overflow calculations wrong | Non-ASCII attribute names |
| Screen width checking | `x + len(text)` | Off-by-N columns | Wide characters |
| Word boundaries | Byte iteration checking for space | Splits multi-byte words | CJK text |

---

## Specific Examples of Breakage

### Example 1: Emoji Rendering
**Input**: Item text "😀 Hello"
**Expected Display**: "[😀 Hello" (single display column for emoji at column 0, space at 1, "Hello" at 2-6)
**Actual Display**: Emoji at column 0, space at 1, "Hello" at 2-6 (happens to work because emoji followed by ASCII)

**Problem Area**: If emoji at middle of line, column calculations are off
- Text "Hello 😀 World" at display position ~18
- Rendering code does `x+i` where `i` is rune index (0, 1, 2...)
- Emoji (1 rune) placed at 1 column position, should be 2 for proper spacing

### Example 2: Chinese Text Wrapping
**Input**: "这是一个很长的中文文本需要换行" (all Chinese, should wrap at visual column 20)
**Expected**: Split into 2 lines, each with ~10 characters displayed at ~20 columns
**Actual**: `wrapTextAtWidth(text, 20)` calculates `len(text)` = 36 bytes (12 characters × 3 bytes each)
- `for i := 0; i < 20` - only checks first 20 bytes (about 6 characters)
- Wrapping position calculation completely wrong
- Text gets truncated at byte 20 = 6.67 characters, splitting last character

### Example 3: Cursor Positioning with Mixed Text
**Input**: "Hello你" (5 ASCII bytes + 3 bytes for 你 = 8 bytes total, 7 display columns)
**Action**: Click at display column 6 (should be at/after 你)
**Expected**: Cursor at byte position 8 (after 你)
**Actual**: Cursor at byte position 6 (middle of 你)
**Result**: Backspace deletes partial character, leaving orphaned UTF-8 bytes

### Example 4: Attribute Display Overflow
**Input**: Item with attribute `日期:2025-01-01` (8 bytes for "日期:", 10 for date = 18 bytes, ~12 display columns)
**Layout**: Text=30 chars, attributes start at column 32
**Expected**: `[日期:2025-01-01]` displayed starting at column 32
**Actual**: Calculation: `attrX + len(attrStr) = 32 + 18 = 50` (byte length check!)
- Thinks it has 50 - 32 = 18 columns available
- Actually only has ~12 columns before edge
- Attribute text overflows or gets cut off

### Example 5: Text Wrapping with Emoji
**Input**: "Hello 😀 World" with maxWidth=10
**Expected**: 
  - Line 1: "Hello 😀 " (visual width ~9)
  - Line 2: "World"
**Actual**: `wrapTextAtWidth` with maxWidth=10:
  - Text = 18 bytes (13 characters: "Hello " = 6, "😀" = 4 bytes, " World" = 6)
  - Byte iteration to position 10 finds space at byte 6
  - Wraps at byte 6: Line 1 = "Hello " (6 bytes)
  - Remaining = "😀 World" (12 bytes)
  - Byte 10 is middle of emoji!
  - Tries to slice `remaining[:10]` = "😀 Wor" = "😀 Wor" = 10 bytes
  - Invalid UTF-8 or missing last rune of "World"

---

## Dependencies Analysis

### What's Available:
1. **github.com/gdamore/tcell/v2 v2.9.0** - Has proper rune handling via `SetContent()`
2. **github.com/mattn/go-runewidth v0.0.16** - NOT IMPORTED or USED anywhere
   - Would solve all width calculation issues
   - `runewidth.StringWidth()` - returns display width of string
   - `runewidth.Wcswidth()` - width of rune slice

### What's Missing:
- **No import of runewidth package in any file**
- No usage of `runewidth.StringWidth()` for width calculations
- No usage of rune iteration with width tracking

---

## Recommendations for Fixes

### Priority 1: Critical Rendering Issues

1. **Create utility functions** in new file `internal/ui/textwidth.go`:
   ```go
   import "github.com/mattn/go-runewidth"
   
   // DisplayWidth returns visual column width of string
   func DisplayWidth(s string) int
   
   // TruncateAtWidth truncates string to fit within width
   func TruncateAtWidth(s string, width int) string
   
   // WrapAtWidth properly wraps text at display width boundaries
   func WrapAtWidth(text string, maxWidth int) []string
   
   // StringToRuneIndex converts byte offset to rune index
   func StringToRuneIndex(s string, bytePos int) int
   
   // RuneIndexToString converts rune index to byte offset
   func RuneIndexToString(s string, runeIndex int) int
   ```

2. **Fix Screen.DrawString**:
   - Use rune iteration with width tracking
   - Place each rune at correct column position
   - Handle wide characters (2 columns) correctly

3. **Fix wrapTextAtWidth**:
   - Use rune iteration with width accumulation
   - Track display width, not byte length
   - Wrap at word boundaries using width, not bytes

4. **Fix text truncation**:
   - Use width-aware truncation
   - Add ellipsis "..." when truncating
   - Never split multi-byte sequences

### Priority 2: Cursor and Input Issues

5. **Fix Editor.SetCursorFromScreenX**:
   - Convert screen column to rune index
   - Track visual position with width calculations
   - Store both byte offset and rune index in cursor

6. **Fix MultiLineEditor cursor tracking**:
   - Use rune-based line offsets, not byte offsets
   - Convert between visual row/col and text positions properly
   - Account for wide characters in column calculations

7. **Fix mouse click handling**:
   - Convert screen column click to rune position
   - Consider wide character widths
   - Use width-aware cursor positioning

### Priority 3: Layout and Display

8. **Fix attribute rendering**:
   - Use width-aware calculations for `attrX` position
   - Check display width, not byte length, for overflow
   - Handle non-ASCII attribute names

9. **Fix search highlighting and link detection**:
   - Track positions with width awareness
   - Map byte positions to screen columns properly
   - Use rune iteration with width

10. **Fix word boundary detection**:
    - Operate on runes, not bytes
    - Use Unicode character classification for boundaries
    - Handle combining characters properly

### Priority 4: Data Integrity

11. **Fix string slicing operations**:
    - Always slice at rune boundaries
    - Never split combining characters
    - Use helper function for safe character deletion

12. **Add input validation**:
    - Validate cursor positions are at rune boundaries
    - Prevent creating invalid UTF-8
    - Safe character-level operations

---

## Test Cases Needed

To validate fixes, add tests with:
1. **CJK characters**: "你好世界", "こんにちは", "안녕하세요"
2. **Emoji**: "😀😁😂", "👨‍👩‍👧‍👦" (with zero-width joiners)
3. **Combining characters**: "é" (e + acute), "Å" (A + ring above)
4. **RTL text**: "مرحبا" (Arabic), "שלום" (Hebrew) - if supported
5. **Mixed text**: "Hello你好😀test"
6. **Emoji with skin tone modifiers**: "👋🏻👋🏼👋🏽👋🏾👋🏿"
7. **Edge cases**:
   - Empty string
   - Single emoji
   - Text exactly at wrap width boundary
   - Very long CJK strings
   - Maxwidth = 1 with wide characters

---

## Files Requiring Changes

### Core Issues (Edit Required):
- `internal/ui/screen.go` - DrawString, DrawStringLimited
- `internal/ui/tree.go` - wrapTextAtWidth, text rendering, attribute display, search highlighting
- `internal/ui/editor.go` - Cursor positioning, text operations, rendering
- `internal/ui/multiline_editor.go` - Line offset calculations, cursor positioning, wrapping
- `internal/app/app.go` - Mouse click handling for editor

### New Files (Create):
- `internal/ui/textwidth.go` - Width utility functions
- `internal/ui/textwidth_test.go` - Tests for width functions

### Review (Check for similar issues):
- `internal/ui/attributes.go` - Text rendering in attribute editor
- `internal/ui/calendar.go` - Date formatting and display
- `internal/ui/command.go` - Command input and rendering
- `internal/ui/search.go` - Search query and highlighting
- All other UI files for text rendering patterns

---

## Impact Assessment

### Current State:
- **ASCII text**: Works perfectly (100% of use cases)
- **Single emoji**: Mostly works but positioning inconsistent
- **CJK text**: Broken - wrapping, cursor, rendering all incorrect
- **Mixed text**: Variable breakage depending on character positions
- **Attribute display**: Broken with non-ASCII names
- **Mouse interaction**: Broken with wide characters

### After Fixes:
- **All text types**: Should work correctly
- **Performance**: Minimal impact (width calculation is O(n) like rendering anyway)
- **Compatibility**: No breaking changes to API

---

## Conclusion

The tui-outliner codebase has comprehensive UTF-8 support for **storage and basic display**, but critical flaws in **width calculation and cursor positioning**. The `go-runewidth` dependency exists but is unused - incorporating its functions would solve most issues. The fixes are substantial but straightforward - primarily replacing byte-length checks with display-width calculations throughout the text handling code.

All issues are solvable through systematic refactoring to use proper Unicode-aware width calculations instead of byte-length assumptions.
