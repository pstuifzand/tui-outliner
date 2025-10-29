package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/storage"
	"github.com/pstuifzand/tui-outliner/internal/ui"
)

// Mode represents the current editor mode
type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

// App is the main application controller
type App struct {
	screen       *ui.Screen
	outline      *model.Outline
	store        *storage.JSONStore
	tree         *ui.TreeView
	editor       *ui.Editor
	search       *ui.Search
	help         *ui.HelpScreen
	command      *ui.CommandMode
	statusMsg    string
	statusTime   time.Time
	dirty        bool
	autoSaveTime time.Time
	quit         bool
	debugMode    bool
	mode         Mode // Current editor mode (NormalMode or InsertMode)
	clipboard    *model.Item // For cut/paste operations
	keybindings  []KeyBinding // All keybindings
}

// NewApp creates a new App instance
func NewApp(filePath string) (*App, error) {
	screen, err := ui.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	store := storage.NewJSONStore(filePath)
	outline, err := store.Load()
	if err != nil {
		screen.Close()
		return nil, fmt.Errorf("failed to load outline: %w", err)
	}

	// If title is empty, set it based on filename
	if outline.Title == "" {
		outline.Title = "Untitled"
	}

	// If no items exist, create a sample item to start with
	if len(outline.Items) == 0 {
		outline.Items = append(outline.Items, model.NewItem("Welcome to tui-outliner"))
	}

	tree := ui.NewTreeView(outline.Items)
	help := ui.NewHelpScreen()
	command := ui.NewCommandMode()

	app := &App{
		screen:       screen,
		outline:      outline,
		store:        store,
		tree:         tree,
		editor:       nil,
		search:       ui.NewSearch(outline.GetAllItems()),
		help:         help,
		command:      command,
		statusMsg:    "Ready",
		statusTime:   time.Now(),
		dirty:        false,
		autoSaveTime: time.Now(),
		quit:         false,
		mode:         NormalMode,
	}

	// Initialize keybindings
	app.keybindings = app.InitializeKeybindings()

	// Convert keybindings to KeyBindingInfo for help screen
	var helpKeybindings []ui.KeyBindingInfo
	for i := range app.keybindings {
		helpKeybindings = append(helpKeybindings, &app.keybindings[i])
	}
	app.help.SetKeybindings(helpKeybindings)

	return app, nil
}

// Run starts the main event loop
func (a *App) Run() error {
	defer a.Close()

	// Create a channel for events
	eventChan := make(chan tcell.Event)

	// Start event polling goroutine
	go func() {
		for {
			event := a.screen.PollEvent()
			eventChan <- event
			if event == nil {
				break
			}
		}
	}()

	// Create a ticker for rendering and auto-save checks
	ticker := time.NewTicker(50 * time.Millisecond) // ~20 FPS
	defer ticker.Stop()

	for !a.quit {
		select {
		case ev := <-eventChan:
			if ev != nil {
				a.handleRawEvent(ev)
			}
		case <-ticker.C:
			a.render()

			// Auto-save every 5 seconds if dirty
			if a.dirty && time.Since(a.autoSaveTime) > 5*time.Second {
				if err := a.Save(); err != nil {
					a.SetStatus("Failed to save: " + err.Error())
				} else {
					a.SetStatus("Saved")
				}
			}
		}
	}

	return nil
}

// Close closes the application
func (a *App) Close() error {
	if a.screen != nil {
		return a.screen.Close()
	}
	return nil
}

// render renders the current state to the screen
func (a *App) render() {
	a.screen.Clear()

	width := a.screen.GetWidth()
	height := a.screen.GetHeight()

	// Fill background for entire screen
	bgStyle := a.screen.BackgroundStyle()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			a.screen.SetCell(x, y, ' ', bgStyle)
		}
	}

	// Draw header (title)
	headerStyle := a.screen.HeaderStyle()
	header := fmt.Sprintf(" %s ", a.outline.Title)
	a.screen.DrawString(0, 0, header, headerStyle)

	// Draw tree
	treeStartY := 1
	treeEndY := height - 2
	if a.search.IsActive() {
		treeEndY -= 2
	}

	// If search is active, show filtered results
	if a.search.IsActive() {
		results := a.search.GetResults()
		if len(results) > 0 {
			// Create a temporary tree with search results
			tempTree := ui.NewTreeView(results)
			tempTree.Render(a.screen, treeStartY)
		} else {
			a.screen.DrawString(0, treeStartY, "No results", ui.DefaultStyle())
		}
	} else {
		a.tree.Render(a.screen, treeStartY)
	}

	// Render editor inline if active
	if a.editor != nil && a.editor.IsActive() {
		selectedIdx := a.tree.GetSelectedIndex()
		if selectedIdx >= 0 {
			// Calculate the Y position of the selected item on screen
			itemY := treeStartY + selectedIdx
			if itemY < treeEndY {
				// Calculate X position after the tree prefix (indentation + arrow + space)
				selected := a.tree.GetSelected()
				if selected != nil {
					// Get depth from tree view (need to find it)
					depth := a.tree.GetSelectedDepth()
					editorX := depth*2 + 2  // indentation + arrow + space
					maxWidth := width - editorX
					if maxWidth > 0 {
						a.editor.Render(a.screen, editorX, itemY, maxWidth)
					}
				}
			}
		}
	}

	// Draw search bar if active
	if a.search.IsActive() {
		a.search.Render(a.screen, height-3)
	}

	// Draw command line if active
	if a.command.IsActive() {
		a.command.Render(a.screen, height-2)
	}

	// Draw status line with mode indicator
	modeStyle := a.screen.StatusModeStyle()
	messageStyle := a.screen.StatusMessageStyle()
	modifiedStyle := a.screen.StatusModifiedStyle()

	statusLine := ""
	lineX := 0

	// Show mode indicator
	if a.mode == InsertMode {
		statusLine = "-- INSERT --"
	} else {
		statusLine = "-- NORMAL --"
	}
	a.screen.DrawString(lineX, height-1, statusLine, modeStyle)
	lineX += len(statusLine)

	// Append status message if not the default "Ready"
	if a.statusMsg != "Ready" {
		if time.Since(a.statusTime) <= 3*time.Second {
			msg := " " + a.statusMsg
			a.screen.DrawString(lineX, height-1, msg, messageStyle)
			lineX += len(msg)
		}
	}

	// Append modified indicator
	if a.dirty {
		modified := " (modified)"
		a.screen.DrawString(lineX, height-1, modified, modifiedStyle)
		lineX += len(modified)
	}

	// Clear remainder of status line
	for lineX < width {
		a.screen.SetCell(lineX, height-1, ' ', modeStyle)
		lineX++
	}

	// Draw help overlay if visible
	a.help.Render(a.screen)

	a.screen.Show()
}

// handleRawEvent processes raw input events
func (a *App) handleRawEvent(ev tcell.Event) {
	// Handle command mode input
	if a.command.IsActive() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			cmd, done := a.command.HandleKey(keyEv)
			if done {
				a.handleCommand(cmd)
			}
		}
		return
	}

	// Handle search input
	if a.search.IsActive() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			if keyEv.Key() == tcell.KeyEscape {
				a.search.Stop()
			} else {
				a.search.HandleKey(keyEv)
			}
		}
		return
	}

	// Handle editor input
	if a.editor != nil && a.editor.IsActive() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			if !a.editor.HandleKey(keyEv) {
				// Check if Enter or Escape was pressed
				enterPressed := a.editor.WasEnterPressed()
				escapePressed := a.editor.WasEscapePressed()
				editedItem := a.editor.GetItem()

				// Exit edit mode
				a.editor.Stop()
				a.editor = nil
				a.dirty = true
				a.mode = NormalMode
				a.SetStatus("Modified")

				// If Escape was pressed and item is empty, delete it
				if escapePressed && editedItem.Text == "" {
					a.tree.DeleteItem(editedItem)
					a.SetStatus("Deleted empty item")
					a.dirty = true
				} else if enterPressed {
					// If Enter was pressed, create new node below and enter insert mode
					a.tree.AddItemAfter("")
					a.SetStatus("Created new item below")
					a.dirty = true
					// Enter insert mode for the new item
					selected := a.tree.GetSelected()
					if selected != nil {
						a.editor = ui.NewEditor(selected)
						a.editor.Start()
						a.mode = InsertMode
					}
				}
			}
		}
		return
	}

	// Handle help screen
	if a.help.IsVisible() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			if keyEv.Key() == tcell.KeyEscape || keyEv.Rune() == '?' {
				a.help.Toggle()
			}
		}
		return
	}

	// Handle normal input
	if keyEv, ok := ev.(*tcell.EventKey); ok {
		a.handleKeypress(keyEv)
	}
}

// handleKeypress handles a single keypress in normal mode
func (a *App) handleKeypress(ev *tcell.EventKey) {
	// Debug mode: show key information
	if a.debugMode {
		a.SetStatus(fmt.Sprintf("Key: %v | Rune: %q | Modifiers: %v", ev.Key(), ev.Rune(), ev.Modifiers()))
	}

	// Handle special keys first
	switch ev.Key() {
	case tcell.KeyDown:
		a.tree.SelectNext()
		return
	case tcell.KeyUp:
		a.tree.SelectPrev()
		return
	case tcell.KeyLeft:
		a.tree.Collapse()
		return
	case tcell.KeyRight:
		a.tree.Expand()
		return
	case tcell.KeyCtrlI:
		if a.tree.Indent() {
			a.SetStatus("Indented")
			a.dirty = true
		}
		return
	case tcell.KeyCtrlU:
		if a.tree.Outdent() {
			a.SetStatus("Outdented")
			a.dirty = true
		}
		return
	case tcell.KeyCtrlS:
		if err := a.Save(); err != nil {
			a.SetStatus("Failed to save: " + err.Error())
		} else {
			a.SetStatus("Saved")
			a.dirty = false
		}
		return
	case tcell.KeyEscape:
		// Can be used for various purposes (just ignore for now)
		return
	}

	// Handle rune (character) keys using keybinding map
	r := ev.Rune()

	// Check for keybinding (also handle . and , as alternates for > and <)
	kb := a.GetKeybindingByKey(r)
	if kb != nil {
		kb.Handler(a)
		return
	}

	// Handle alternate keybindings for indent/outdent
	switch r {
	case '.':  // . as alternate for indent
		if a.tree.Indent() {
			a.SetStatus("Indented")
			a.dirty = true
		}
	case ',':  // , as alternate for outdent
		if a.tree.Outdent() {
			a.SetStatus("Outdented")
			a.dirty = true
		}
	}
}

// handleCommand processes a command from command mode
func (a *App) handleCommand(cmd string) {
	if cmd == "" {
		return
	}

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "q", "quit":
		if a.dirty {
			a.SetStatus("Unsaved changes! Use :q! to force quit or :w to save")
		} else {
			a.quit = true
		}
	case "q!", "quit!":
		a.quit = true
	case "w", "write":
		if err := a.Save(); err != nil {
			a.SetStatus("Failed to save: " + err.Error())
		} else {
			a.SetStatus("Saved")
		}
	case "wq":
		if err := a.Save(); err != nil {
			a.SetStatus("Failed to save: " + err.Error())
		} else {
			a.quit = true
		}
	case "help":
		a.help.Toggle()
	case "debug":
		a.debugMode = !a.debugMode
		if a.debugMode {
			a.SetStatus("Debug mode ON")
		} else {
			a.SetStatus("Debug mode OFF")
		}
	default:
		a.SetStatus("Unknown command: " + parts[0])
	}
}

// Save saves the outline to disk
func (a *App) Save() error {
	if err := a.store.Save(a.outline); err != nil {
		return err
	}
	a.dirty = false
	a.autoSaveTime = time.Now()
	return nil
}

// SetStatus sets the status message
func (a *App) SetStatus(msg string) {
	a.statusMsg = msg
	a.statusTime = time.Now()
}

// Quit signals the app to quit
func (a *App) Quit() {
	a.quit = true
}

// SetDebugMode enables or disables debug mode
func (a *App) SetDebugMode(debug bool) {
	a.debugMode = debug
}
