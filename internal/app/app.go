package app

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gdamore/tcell/v2"
	"github.com/pstuifzand/tui-outliner/internal/config"
	"github.com/pstuifzand/tui-outliner/internal/export"
	"github.com/pstuifzand/tui-outliner/internal/history"
	import_parser "github.com/pstuifzand/tui-outliner/internal/import"
	"github.com/pstuifzand/tui-outliner/internal/links"
	"github.com/pstuifzand/tui-outliner/internal/model"
	search "github.com/pstuifzand/tui-outliner/internal/search"
	"github.com/pstuifzand/tui-outliner/internal/socket"
	"github.com/pstuifzand/tui-outliner/internal/storage"
	tmpl "github.com/pstuifzand/tui-outliner/internal/template"
	"github.com/pstuifzand/tui-outliner/internal/ui"
)

// Mode represents the current editor mode
type Mode int

const (
	NormalMode Mode = iota
	InsertMode
	VisualMode
)

// App is the main application controller
type App struct {
	screen                 *ui.Screen
	outline                *model.Outline
	store                  *storage.JSONStore
	tree                   *ui.TreeView
	editor                 *ui.MultiLineEditor
	search                 *ui.Search
	help                   *ui.HelpScreen
	splash                 *ui.SplashScreen
	command                *ui.CommandMode
	attributeEditor        *ui.AttributeEditor // Attribute editing modal
	nodeSearchWidget       *ui.NodeSearchWidget
	linkAutocompleteWidget *ui.LinkAutocompleteWidget // Wiki-style link autocomplete
	calendarWidget         *ui.CalendarWidget         // Calendar date picker widget
	backupSelectorWidget   *ui.BackupSelectorWidget   // Backup selector with side-by-side diff preview
	messageLogger          *ui.MessageLogger          // Message history for :messages command
	historyManager         *history.Manager           // Manager for persisting command and search history
	socketServer           *socket.Server             // Unix socket server for external commands
	cfg                    *config.Config             // Application configuration
	sessionID              string                     // 8-character session ID for backups
	readOnly               bool                       // Whether the file is readonly (e.g., backup file)
	originalFilePath       string                     // Original file path for backup filtering
	currentBackupPath      string                     // Current backup file path if viewing a backup
	statusMsg              string
	statusTime             time.Time
	dirty                  bool
	autoSaveTime           time.Time
	quit                   bool
	debugMode              bool
	messagesViewActive     bool                // Whether messages view is currently displayed
	messagesViewMessages   []*ui.Message       // Messages to display
	messagesViewScroll     int                 // Scroll position for messages view
	mode                   Mode                // Current editor mode (NormalMode, InsertMode, or VisualMode)
	clipboard              *model.Item         // For cut/paste operations
	visualAnchor           int                 // For visual mode selection (index in filteredView, -1 when not in visual mode)
	lastSendDestination    *model.Item         // Last destination node used with 'ss' for repeat with 's.'
	keybindings            []KeyBinding        // All keybindings
	pendingKeybindings     []PendingKeyBinding // Pending key definitions (g, z, etc)
	pendingKeySeq          rune                // Current pending key waiting for second character
	hasFile                bool                // Whether a file was provided in arguments
}

// NewApp creates a new App instance
func NewApp(filePath string) (*App, error) {
	screen, err := ui.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	// Enable mouse support if available
	screen.EnableMouse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v\n", err)
		cfg = &config.Config{}
		cfg.Set("", "") // Initialize sessionSettings
	}

	// Set default todostatuses if not already configured
	if cfg.Get("todostatuses") == "" {
		cfg.Set("todostatuses", "todo,doing,done")
	}

	// Set default showprogress if not already configured
	if cfg.Get("showprogress") == "" {
		cfg.Set("showprogress", "true")
	}

	store := storage.NewJSONStore(filePath)
	sessionID := generateSessionID()
	store.SetSessionID(sessionID)

	outline, err := store.Load()
	if err != nil {
		screen.Close()
		return nil, fmt.Errorf("failed to load outline: %w", err)
	}

	tree := ui.NewTreeView(outline.Items)
	help := ui.NewHelpScreen()
	splash := ui.NewSplashScreen()
	attributeEditor := ui.NewAttributeEditor()
	nodeSearchWidget := ui.NewNodeSearchWidget()
	linkAutocompleteWidget := ui.NewLinkAutocompleteWidget()
	calendarWidget := ui.NewCalendarWidget()
	backupSelectorWidget := ui.NewBackupSelectorWidget()
	messageLogger := ui.NewMessageLogger(10) // Track last 10 messages

	// Connect calendar widget to attribute editor
	attributeEditor.SetCalendarWidget(calendarWidget)

	// Apply weekstart config if set (0=Sunday, 1=Monday, etc.)
	if weekstartStr := cfg.Get("weekstart"); weekstartStr != "" {
		if weekstart, err := strconv.Atoi(weekstartStr); err == nil && weekstart >= 0 && weekstart <= 6 {
			calendarWidget.SetWeekStart(weekstart)
		}
	}

	// Initialize history manager
	historyManager, err := history.NewManager()
	if err != nil {
		log.Printf("Warning: Failed to initialize history manager: %v\n", err)
		historyManager = nil
	}

	// Initialize command mode with history persistence
	var command *ui.CommandMode
	if historyManager != nil {
		command, err = ui.NewCommandModeWithHistory(historyManager)
		if err != nil {
			log.Printf("Warning: Failed to load command history: %v\n", err)
			command = ui.NewCommandMode()
		}
	} else {
		command = ui.NewCommandMode()
	}

	// Initialize search with history persistence
	var searchMode *ui.Search
	if historyManager != nil {
		searchMode, err = ui.NewSearchWithHistory(outline.GetAllItems(), historyManager)
		if err != nil {
			log.Printf("Warning: Failed to load search history: %v\n", err)
			searchMode = ui.NewSearch(outline.GetAllItems())
		}
	} else {
		searchMode = ui.NewSearch(outline.GetAllItems())
	}

	// Show splash screen if no file was provided
	hasFile := filePath != ""
	if !hasFile {
		splash.Show()
	}

	// Determine if loading a backup file and extract original filename
	isBackup := store.ReadOnly
	var originalFilePath string
	var currentBackupPath string
	if isBackup {
		currentBackupPath = filePath
		originalFilePath = outline.OriginalFilename
	} else if filePath != "" {
		// Convert to absolute path for consistent backup searching
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			absPath = filePath
		}
		originalFilePath = absPath
	}

	app := &App{
		screen:                 screen,
		outline:                outline,
		store:                  store,
		tree:                   tree,
		editor:                 nil,
		search:                 searchMode,
		help:                   help,
		splash:                 splash,
		command:                command,
		attributeEditor:        attributeEditor,
		nodeSearchWidget:       nodeSearchWidget,
		linkAutocompleteWidget: linkAutocompleteWidget,
		calendarWidget:         calendarWidget,
		backupSelectorWidget:   backupSelectorWidget,
		messageLogger:          messageLogger,
		historyManager:         historyManager,
		cfg:                    cfg,
		sessionID:              sessionID,
		readOnly:               store.ReadOnly,
		originalFilePath:       originalFilePath,
		currentBackupPath:      currentBackupPath,
		statusMsg:              "Ready",
		statusTime:             time.Now(),
		dirty:                  false,
		autoSaveTime:           time.Now(),
		quit:                   false,
		debugMode:              false,
		messagesViewActive:     false,
		messagesViewMessages:   []*ui.Message{},
		messagesViewScroll:     0,
		mode:                   NormalMode,
		visualAnchor:           -1,
		pendingKeySeq:          0,
		hasFile:                hasFile,
	}

	// Set callback for attribute editor modifications
	attributeEditor.SetOnModified(func() {
		app.dirty = true
	})

	// Set validation callback for attribute editor
	attributeEditor.SetValidateAttribute(func(key, value string) string {
		// Load type registry from outline
		registry := tmpl.NewTypeRegistry()
		if err := registry.LoadFromOutline(app.outline); err != nil {
			// If we can't load types, allow the value (type system is optional)
			return ""
		}

		// Get the type definition for this key
		typeSpec := registry.GetType(key)
		if typeSpec == nil {
			// No type definition for this key - that's OK
			return ""
		}

		// Validate the value against the type
		if err := typeSpec.Validate(value); err != nil {
			return err.Error()
		}

		return ""
	})

	// Set up type registry for type-aware value selection
	typeRegistry := tmpl.NewTypeRegistry()
	if err := typeRegistry.LoadFromOutline(app.outline); err == nil {
		attributeEditor.SetTypeRegistry(typeRegistry)
	}

	// Set callbacks for node search widget
	nodeSearchWidget.SetOnSelect(func(item *model.Item) {
		// Expand parents and navigate to item in current tree
		app.tree.ExpandParents(item)
		items := app.tree.GetDisplayItems()
		for idx, dispItem := range items {
			if dispItem.Item.ID == item.ID {
				app.tree.SelectItem(idx)
				app.SetStatus(fmt.Sprintf("Selected: %s", item.Text))
				break
			}
		}
	})

	nodeSearchWidget.SetOnHoist(func(item *model.Item) {
		// Navigate to item and hoist it
		app.tree.ExpandParents(item)

		// Find the item in the display and select it
		items := app.tree.GetDisplayItems()
		found := false
		for idx, dispItem := range items {
			if dispItem.Item.ID == item.ID {
				app.tree.SelectItem(idx)
				found = true
				break
			}
		}

		// Now hoist the selected item
		if found {
			if app.tree.Hoist() {
				app.SetStatus(fmt.Sprintf("Hoisted: %s", item.Text))
			} else {
				app.SetStatus("Cannot hoist (no children)")
			}
		} else {
			app.SetStatus("Item not found in tree")
		}
	})

	// Set callback for link autocomplete widget
	linkAutocompleteWidget.SetOnSelect(func(item *model.Item) {
		// Insert the link into the editor
		if app.editor != nil && app.editor.IsActive() {
			app.editor.InsertLink(item.ID, item.Text)
			app.dirty = true
		}
	})

	// Set up calendar widget with all items for date detection
	calendarWidget.SetItems(outline.GetAllItems())

	// Calendar date selection handler (context-dependent)
	calendarWidget.SetOnDateSelected(func(selectedDate time.Time) {
		dateStr := selectedDate.Format("2006-01-02")
		formattedDate := selectedDate.Format("Mon, Jan 2, 2006") // Short day name format for display

		// Check context mode
		if app.calendarWidget.GetContextMode() == ui.CalendarAttributeMode {
			// Set the date attribute on current item
			selected := app.tree.GetSelected()
			if selected != nil {
				selected.Metadata.Attributes[app.calendarWidget.GetAttributeName()] = dateStr
				app.dirty = true
				app.SetStatus(fmt.Sprintf("Set %s to %s", app.calendarWidget.GetAttributeName(), dateStr))
			} else {
				app.SetStatus("No item selected")
			}
		} else {
			// Search mode: look for daily note (type=day) with matching date
			query := fmt.Sprintf("@type=day @date=%s", dateStr)
			matchingItem, err := search.GetFirstByQuery(app.outline, query)
			if err != nil {
				app.SetStatus(fmt.Sprintf("Error while searching: %v", err))
				log.Println(err)
				return
			}

			if matchingItem != nil {
				// Navigate to existing item
				app.tree.ExpandParents(matchingItem)
				items := app.tree.GetDisplayItems()
				for idx, dispItem := range items {
					if dispItem.Item.ID == matchingItem.ID {
						app.tree.SelectItem(idx)
						app.SetStatus(fmt.Sprintf("Selected: %s", matchingItem.Text))
						return
					}
				}
			} else {
				matchingItem, err := search.GetFirstByQuery(app.outline, "@type=dailynotes")
				if err != nil {
					app.SetStatus(fmt.Sprintf("Error while searching: %v", err))
					log.Println(err)
					return
				}
				if matchingItem == nil {
					// Create new daily notes container if it doesn't exist
					newItem := model.NewItem("Daily Notes")
					newItem.Metadata.Attributes["type"] = "dailynotes"
					app.tree.AddItemAfter(newItem)
					matchingItem = newItem
				} else {
					spew.Fdump(log.Default().Writer(), matchingItem)
					app.tree.SelectItemByID(matchingItem.ID)
				}

				// Create new daily note with formatted date
				newItem := model.NewItem(formattedDate)
				newItem.Metadata.Attributes["type"] = "day"
				newItem.Metadata.Attributes["date"] = dateStr

				app.tree.AddItemAsChild(newItem)

				app.dirty = true
				app.SetStatus(fmt.Sprintf("Created: %s", formattedDate))
			}
		}
	})

	// Initialize keybindings
	app.keybindings = app.InitializeKeybindings()
	app.pendingKeybindings = app.InitializePendingKeybindings()

	// Convert keybindings to KeyBindingInfo for help screen
	var helpKeybindings []ui.KeyBindingInfo
	for i := range app.keybindings {
		helpKeybindings = append(helpKeybindings, &app.keybindings[i])
	}
	// Add pending keybindings to help
	for i := range app.pendingKeybindings {
		helpKeybindings = append(helpKeybindings, &app.pendingKeybindings[i])
	}
	app.help.SetKeybindings(helpKeybindings)

	// Initialize socket server for external commands
	socketServer, err := socket.NewServer(os.Getpid())
	if err != nil {
		log.Printf("Warning: Failed to create socket server: %v", err)
		// Don't fail app creation if socket server fails
		app.socketServer = nil
	} else {
		app.socketServer = socketServer
		log.Printf("Socket server initialized: %s", socketServer.SocketPath())
	}

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

	// Start socket server if available
	if a.socketServer != nil {
		a.socketServer.Start()
		log.Printf("Socket server started")
	}

	// Create a ticker for rendering and auto-save checks
	ticker := time.NewTicker(50 * time.Millisecond) // ~20 FPS
	defer ticker.Stop()

	// Get socket message channel (nil if no server)
	var socketChan <-chan socket.Message
	if a.socketServer != nil {
		socketChan = a.socketServer.Messages()
	}

	for !a.quit {
		select {
		case ev := <-eventChan:
			if ev != nil {
				a.handleRawEvent(ev)
			}
		case msg := <-socketChan:
			a.handleSocketMessage(msg)
		case <-ticker.C:
			a.render()

			// Auto-save every 5 seconds if dirty (skip for readonly files)
			if a.dirty && !a.readOnly && time.Since(a.autoSaveTime) > 5*time.Second {
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
	// Stop socket server if running
	if a.socketServer != nil {
		a.socketServer.Stop()
		log.Printf("Socket server stopped")
	}

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
	for y := range height {
		for x := range width {
			a.screen.SetCell(x, y, ' ', bgStyle)
		}
	}

	// Draw splash screen if visible
	if a.splash.IsVisible() {
		a.splash.Render(a.screen)
		// Still draw command line if active
		if a.command.IsActive() {
			a.command.Render(a.screen, height-1)
		}
		a.screen.Show()
		return
	}

	// Draw messages view if active
	if a.messagesViewActive {
		a.renderMessagesView()
		if a.command.IsActive() {
			a.command.Render(a.screen, height-1)
		}
		a.screen.Show()
		return
	}

	// Draw header (title)
	headerStyle := a.screen.HeaderStyle()
	var header string

	// Add hoisting indicator with breadcrumbs if hoisted
	if a.tree.IsHoisted() {
		breadcrumbs := a.tree.GetHoistBreadcrumbs()
		if breadcrumbs != "" {
			header = fmt.Sprintf("[%s] ", breadcrumbs)
		}
	}

	a.screen.DrawString(0, 0, header, headerStyle)

	// Draw tree
	treeStartY := 1
	if header == "" {
		treeStartY = 0
	}
	treeEndY := height - 2
	if a.search.IsActive() {
		treeEndY -= 3
	}

	// Render the main tree (search is active but doesn't filter items)
	searchQuery := ""
	currentMatchItem := (*model.Item)(nil)
	if a.search.IsActive() {
		searchQuery = a.search.GetQuery()
		currentMatchItem = a.search.GetCurrentMatch()
	}
	a.tree.RenderWithSearchQuery(a.screen, treeStartY, treeEndY, a.visualAnchor, searchQuery, currentMatchItem, a.cfg)

	// Render editor inline if active
	if a.editor != nil && a.editor.IsActive() {
		selected := a.tree.GetSelected()
		if selected != nil {
			// Get the first display line for the selected item
			displayLines := a.tree.GetDisplayLines()
			displayLineIdx := -1
			for idx, line := range displayLines {
				if line.Item.ID == selected.ID && line.ItemStartLine {
					displayLineIdx = idx
					break
				}
			}

			if displayLineIdx >= 0 {
				// Calculate the Y position based on display line (accounting for viewport offset)
				viewportOffset := a.tree.GetViewportOffset()
				screenIdx := displayLineIdx - viewportOffset
				if screenIdx >= 0 && screenIdx < treeEndY-treeStartY {
					itemY := treeStartY + screenIdx
					// Calculate X position after the tree prefix (indentation + arrow + space)
					depth := a.tree.GetSelectedDepth()
					editorX := depth*3 + 3 // indentation + arrow + attribute indicator + space
					// Use same max width calculation as tree view for consistent wrapping
					maxWidth := width - 21 // 6 levels * 3 + 3 for arrow area
					if maxWidth < 20 {
						maxWidth = 20 // Minimum wrap width
					}
					// Render editor (may span multiple lines)
					// Render call will call SetMaxWidth internally
					a.editor.Render(a.screen, editorX, itemY, maxWidth)

					// Clear remaining lines if editor is shorter than original item
					editorLineCount := a.editor.GetWrappedLineCount()
					editorEndY := itemY + editorLineCount

					// Find how many display lines the original item occupied
					itemLineCount := 0
					for idx := displayLineIdx; idx < len(displayLines); idx++ {
						if idx > displayLineIdx && displayLines[idx].ItemStartLine {
							// Next item found
							break
						}
						itemLineCount++
					}

					// Clear lines below editor if item was taller
					itemEndY := itemY + itemLineCount
					if editorEndY < itemEndY {
						editingStyle := a.screen.BackgroundStyle()
						for y := editorEndY; y < itemEndY && y < treeEndY; y++ {
							clearLine := strings.Repeat(" ", width)
							a.screen.DrawString(0, y, clearLine, editingStyle)
						}
					}
				}
			}
		}
	}

	// Draw search bar if active
	if a.search.IsActive() {
		a.search.Render(a.screen, height-2)
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
	switch a.mode {
	case InsertMode:
		statusLine = "-- INSERT --"
	case VisualMode:
		statusLine = "-- VISUAL --"
	default:
		statusLine = "NORMAL"
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

	// Append readonly indicator
	if a.readOnly {
		readonly := " [READONLY]"
		a.screen.DrawString(lineX, height-1, readonly, modifiedStyle)
		lineX += len(readonly)
	}

	// Clear remainder of status line
	for lineX < width {
		a.screen.SetCell(lineX, height-1, ' ', modeStyle)
		lineX++
	}

	// Draw help overlay if visible
	a.help.Render(a.screen)

	// Draw attribute editor if visible
	a.attributeEditor.Render(a.screen)

	// Draw node search widget if visible
	a.nodeSearchWidget.Render(a.screen)

	// Draw link autocomplete widget if visible
	a.linkAutocompleteWidget.Render(a.screen)

	// Draw calendar widget if visible
	a.calendarWidget.Render(a.screen)

	// Draw backup selector widget if visible (includes side-by-side diff preview)
	a.backupSelectorWidget.Render(a.screen)

	a.screen.Show()
}

// renderMessagesView renders the message history view on the screen
func (a *App) renderMessagesView() {
	width := a.screen.GetWidth()
	height := a.screen.GetHeight()

	// Draw title area
	titleStyle := a.screen.HelpTitleStyle()
	title := " Message History "
	a.screen.DrawString(0, 0, title, titleStyle)

	// Clear rest of title line
	for x := len(title); x < width; x++ {
		a.screen.SetCell(x, 0, ' ', titleStyle)
	}

	// Draw messages starting from line 1
	contentStyle := a.screen.HelpStyle()
	startY := 1
	maxY := height - 2 // Leave room for status bar

	if len(a.messagesViewMessages) == 0 {
		a.screen.DrawString(1, startY, "No messages", contentStyle)
		return
	}

	// Calculate visible range based on scroll
	visibleLines := maxY - startY
	if visibleLines <= 0 {
		return
	}

	// Display messages from scroll position
	currentY := startY
	for i := a.messagesViewScroll; i < len(a.messagesViewMessages) && currentY < maxY; i++ {
		msg := a.messagesViewMessages[i]

		// Format message with timestamp
		timestamp := msg.Timestamp.Format("15:04:05")
		msgText := fmt.Sprintf("[%s] %s", timestamp, msg.Text)

		// Truncate to screen width if needed
		if len(msgText) > width-2 {
			msgText = msgText[:width-2]
		}

		a.screen.DrawString(1, currentY, msgText, contentStyle)
		currentY++
	}

	// Fill rest of view with blank lines
	for currentY < maxY {
		a.screen.DrawString(1, currentY, "", contentStyle)
		currentY++
	}

	// Draw status bar showing help and scroll position
	statusStyle := a.screen.StatusMessageStyle()
	statusMsg := fmt.Sprintf("Messages: %d | [q]Close  [j/k]Scroll", len(a.messagesViewMessages))
	a.screen.DrawString(0, height-1, statusMsg, statusStyle)

	// Fill rest of status line
	for x := len(statusMsg); x < width; x++ {
		a.screen.SetCell(x, height-1, ' ', statusStyle)
	}
}

// handleRawEvent processes raw input events
func (a *App) handleRawEvent(ev tcell.Event) {
	// Handle splash screen
	if a.splash.IsVisible() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			// Allow colon to enter command mode and hide splash screen
			if keyEv.Rune() == ':' {
				a.splash.Hide()
				a.command.Start()
				return
			}
			// Allow ESC to dismiss splash screen
			if keyEv.Key() == tcell.KeyEscape {
				a.splash.Hide()
				return
			}
		}
		return
	}

	// Handle messages view input
	if a.messagesViewActive {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			ch := keyEv.Rune()
			switch ch {
			case 'q':
				// Close messages view
				a.messagesViewActive = false
				a.messagesViewScroll = 0
				return
			case 'j':
				// Scroll down
				if a.messagesViewScroll < len(a.messagesViewMessages)-1 {
					a.messagesViewScroll++
				}
				return
			case 'k':
				// Scroll up
				if a.messagesViewScroll > 0 {
					a.messagesViewScroll--
				}
				return
			case ':':
				// Allow command mode even in messages view
				a.messagesViewActive = false
				a.command.Start()
				return
			}
		}
		return
	}

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

	// Handle attribute editor input
	if a.attributeEditor.IsVisible() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			a.attributeEditor.HandleKeyEvent(keyEv)
		}
		if mouseEv, ok := ev.(*tcell.EventMouse); ok {
			x, y := mouseEv.Position()
			// Only handle left mouse button clicks
			if mouseEv.Buttons()&tcell.Button1 != 0 {
				a.attributeEditor.HandleMouseEvent(x, y)
			}
		}
		return
	}

	// Handle link autocomplete widget input
	if a.linkAutocompleteWidget.IsVisible() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			a.linkAutocompleteWidget.HandleKeyEvent(keyEv)
		}
		return
	}

	// Handle node search widget input
	if a.nodeSearchWidget.IsVisible() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			a.nodeSearchWidget.HandleKeyEvent(keyEv)
		}
		return
	}

	// Handle calendar widget input
	if a.calendarWidget.IsVisible() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			a.calendarWidget.HandleKeyEvent(keyEv)
		}
		if mouseEv, ok := ev.(*tcell.EventMouse); ok {
			x, y := mouseEv.Position()
			// Only handle left mouse button clicks
			if mouseEv.Buttons()&tcell.Button1 != 0 {
				a.calendarWidget.HandleMouseEvent(x, y)
			}
		}
		return
	}

	// Handle backup selector widget input
	if a.backupSelectorWidget.IsVisible() {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			a.backupSelectorWidget.HandleKeyEvent(keyEv)
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
				// After handling any search key, navigate to first match if there are results
				if a.search.HasResults() {
					currentMatch := a.search.GetCurrentMatch()
					if currentMatch != nil {
						// Expand all parent nodes of the match so it becomes visible
						a.tree.ExpandParents(currentMatch)
						// Find and select this item in the main tree
						items := a.tree.GetDisplayItems()
						for idx, dispItem := range items {
							if dispItem.Item.ID == currentMatch.ID {
								a.tree.SelectItem(idx)
								break
							}
						}
						matchNum := a.search.GetCurrentMatchNumber()
						totalMatches := a.search.GetMatchCount()
						a.SetStatus(fmt.Sprintf("Match %d of %d", matchNum, totalMatches))
					}
				} else if a.search.GetQuery() != "" {
					// Query is not empty but no matches
					a.SetStatus("No matches")
				}
			}
		}
		return
	}

	// Handle editor input (keyboard and mouse)
	if a.editor != nil && a.editor.IsActive() {
		// Handle mouse clicks to position cursor
		if mouseEv, ok := ev.(*tcell.EventMouse); ok {
			a.handleEditorMouseClick(mouseEv)
			return
		}

		if keyEv, ok := ev.(*tcell.EventKey); ok {
			if !a.editor.HandleKey(keyEv) {
				// Check if link autocomplete was triggered ([[ was typed)
				if a.editor.WasLinkAutocompleteTriggered() {
					// Show link autocomplete widget
					a.linkAutocompleteWidget.SetItems(a.outline.GetAllItems())
					a.linkAutocompleteWidget.Show()
					return
				}

				// Check if Enter, Escape, Backspace on empty, or indent/outdent was pressed
				enterPressed := a.editor.WasEnterPressed()
				escapePressed := a.editor.WasEscapePressed()
				backspaceOnEmpty := a.editor.WasBackspaceOnEmpty()
				indentPressed := a.editor.WasIndentPressed()
				outdentPressed := a.editor.WasOutdentPressed()
				editedItem := a.editor.GetItem()

				// Exit edit mode (except for indent/outdent which continue in insert mode)
				a.editor.Stop()
				a.editor = nil
				a.dirty = true
				a.mode = NormalMode
				a.SetStatus("Modified")
				// Refresh any expanded search nodes with new/updated items
				a.refreshSearchNodes()

				// Auto-detect todo checkbox pattern "[] "
				if strings.HasPrefix(editedItem.Text, "[] ") {
					// Initialize metadata if needed
					if editedItem.Metadata == nil {
						editedItem.Metadata = &model.Metadata{
							Attributes: make(map[string]string),
							Created:    time.Now(),
							Modified:   time.Now(),
						}
					}
					// Set type and status attributes, strip prefix from text
					editedItem.Metadata.Attributes["type"] = "todo"
					editedItem.Metadata.Attributes["status"] = "todo"
					editedItem.Text = strings.TrimPrefix(editedItem.Text, "[] ")
					editedItem.Metadata.Modified = time.Now()
				}

				// If Escape was pressed and item is empty (and has no children), delete it
				if escapePressed && editedItem.Text == "" && len(editedItem.Children) == 0 {
					// Move to previous item before deleting
					currentIdx := a.tree.GetSelectedIndex()
					a.tree.DeleteItem(editedItem)
					// Select the previous item if it exists
					if currentIdx > 0 {
						a.tree.SelectItem(currentIdx - 1)
					}
					a.SetStatus("Deleted empty item")
					a.dirty = true
				} else if backspaceOnEmpty {
					// Backspace pressed on empty item - merge with previous item
					prevIdx := a.tree.GetSelectedIndex() - 1
					if prevIdx >= 0 {
						a.tree.DeleteItem(editedItem)
						a.tree.SelectItem(prevIdx)
						a.SetStatus("Merged with previous item")
						a.dirty = true

						// Enter insert mode on previous item with cursor at end
						prevItem := a.tree.GetSelected()
						if prevItem != nil {
							a.editor = ui.NewMultiLineEditor(prevItem)
							a.editor.Start()
							a.mode = InsertMode
						}
					}
				} else if indentPressed {
					// Tab pressed - indent the current item
					if a.tree.Indent() {
						a.SetStatus("Indented")
						a.dirty = true
					} else {
						a.SetStatus("Cannot indent (no previous item)")
					}
					// Re-enter insert mode on the same item
					a.editor = ui.NewMultiLineEditor(editedItem)
					a.editor.Start()
					a.mode = InsertMode
				} else if outdentPressed {
					// Shift+Tab pressed - outdent the current item
					if a.tree.Outdent() {
						a.SetStatus("Outdented")
						a.dirty = true
					} else {
						a.SetStatus("Cannot outdent (already at root level)")
					}
					// Re-enter insert mode on the same item
					a.editor = ui.NewMultiLineEditor(editedItem)
					a.editor.Start()
					a.mode = InsertMode
				} else if enterPressed {
					// If Enter was pressed, create new node below and enter insert mode
					item := model.NewItem("")
					a.tree.AddItemAfter(item)
					a.SetStatus("Created new item below")
					a.dirty = true
					// Enter insert mode for the new item
					selected := a.tree.GetSelected()
					if selected != nil {
						a.editor = ui.NewMultiLineEditor(selected)
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

	// Handle visual mode
	if a.mode == VisualMode {
		if keyEv, ok := ev.(*tcell.EventKey); ok {
			a.handleVisualMode(keyEv)
		}
		return
	}

	// Handle normal mode (keyboard and mouse)
	if keyEv, ok := ev.(*tcell.EventKey); ok {
		a.handleKeypress(keyEv)
		return
	}

	// Handle mouse clicks in normal mode
	if mouseEv, ok := ev.(*tcell.EventMouse); ok {
		a.handleTreeMouseClick(mouseEv)
	}
}

// handleKeypress handles a single keypress in normal mode
func (a *App) handleKeypress(ev *tcell.EventKey) {
	// Debug mode: show key information
	if a.debugMode {
		log.Printf("Key: %v | Rune: %q | Modifiers: %v", ev.Key(), ev.Rune(), ev.Modifiers())
	}

	// Handle special keys first
	switch ev.Key() {
	case tcell.KeyDown:
		a.tree.SelectNext()
		a.pendingKeySeq = 0 // Clear pending sequence on other keys
		return
	case tcell.KeyUp:
		a.tree.SelectPrev()
		a.pendingKeySeq = 0
		return
	case tcell.KeyLeft:
		a.tree.Collapse()
		a.pendingKeySeq = 0
		return
	case tcell.KeyRight:
		a.tree.Expand(true)
		a.pendingKeySeq = 0
		return
	case tcell.KeyCtrlI:
		if a.tree.Indent() {
			a.SetStatus("Indented")
			a.dirty = true
		}
		a.pendingKeySeq = 0
		return
	case tcell.KeyCtrlU:
		// Page up - scroll viewport
		height := a.screen.GetHeight()
		treeStartY := 0
		treeEndY := height - 2
		if a.search.IsActive() {
			treeEndY -= 2
		}
		pageSize := treeEndY - treeStartY
		if pageSize < 1 {
			pageSize = 1
		}
		a.tree.ScrollPageUp(pageSize)
		a.pendingKeySeq = 0
		return
	case tcell.KeyCtrlD:
		// Page down - scroll viewport
		height := a.screen.GetHeight()
		treeStartY := 0
		treeEndY := height - 2
		if a.search.IsActive() {
			treeEndY -= 2
		}
		pageSize := max(treeEndY-treeStartY, 1)
		a.tree.ScrollPageDown(pageSize)
		a.pendingKeySeq = 0
		return
	case tcell.KeyCtrlS:
		if err := a.Save(); err != nil {
			a.SetStatus("Failed to save: " + err.Error())
		} else {
			a.SetStatus("Saved")
			a.dirty = false
		}
		a.pendingKeySeq = 0
		return
	case tcell.KeyCtrlK:
		a.pendingKeySeq = 0
		// Collect all searchable items
		var allItems []*model.Item
		for _, item := range a.tree.GetItems() {
			allItems = append(allItems, ui.GetAllItemsRecursive(item)...)
		}
		a.nodeSearchWidget.SetItems(allItems)
		a.nodeSearchWidget.Show()
		return
	case tcell.KeyEscape:
		// Can be used for various purposes (just ignore for now)
		a.pendingKeySeq = 0
		return
	}

	// Handle rune (character) keys using keybinding map
	r := ev.Rune()

	// Check if we're waiting for a second key of a pending key sequence
	if a.pendingKeySeq != 0 {
		pendingKey := a.GetPendingKeyBindingByPrefix(a.pendingKeySeq)
		if pendingKey != nil {
			if seqBinding, ok := pendingKey.Sequences[r]; ok {
				// Execute the pending key sequence
				seqBinding.Handler(a)
				a.pendingKeySeq = 0
				return
			}
		}
		// Clear pending sequence if second key didn't match
		a.pendingKeySeq = 0
	}

	// Check if this is a pending key prefix
	if a.IsPendingKeyPrefix(r) {
		a.pendingKeySeq = r
		return
	}

	// Check for regular keybinding (also handle . and , as alternates for > and <)
	kb := a.GetKeybindingByKey(r)
	if kb != nil {
		kb.Handler(a)
		return
	}

	// Handle alternate keybindings for indent/outdent
	switch r {
	case '.': // . as alternate for indent
		if a.tree.Indent() {
			a.SetStatus("Indented")
			a.dirty = true
		}
	case ',': // , as alternate for outdent
		if a.tree.Outdent() {
			a.SetStatus("Outdented")
			a.dirty = true
		}
	}
}

// parseCommand parses a command string into parts, respecting quoted strings
// Handles both single and double quotes, and allows escaping quotes with backslash
func parseCommand(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)
	wasQuoted := false
	i := 0

	for i < len(cmd) {
		r := rune(cmd[i])

		// Handle escape sequences
		if r == '\\' && i+1 < len(cmd) {
			nextR := rune(cmd[i+1])
			// Escape quote or backslash
			if nextR == '"' || nextR == '\'' || nextR == '\\' {
				current.WriteRune(nextR)
				i += 2
				continue
			}
		}

		// Handle quotes
		if (r == '"' || r == '\'') && (quoteChar == 0 || quoteChar == r) {
			if inQuote {
				inQuote = false
				quoteChar = 0
			} else {
				inQuote = true
				quoteChar = r
				wasQuoted = true
			}
			i++
			continue
		}

		// Handle whitespace (outside quotes)
		if !inQuote && (r == ' ' || r == '\t') {
			if current.Len() > 0 || wasQuoted {
				parts = append(parts, current.String())
				current.Reset()
				wasQuoted = false
			}
			i++
			continue
		}

		// Regular character
		current.WriteRune(r)
		i++
	}

	// Add final part
	if current.Len() > 0 || wasQuoted {
		parts = append(parts, current.String())
	}

	return parts
}

// handleCommand processes a command from command mode
func (a *App) handleCommand(cmd string) {
	if cmd == "" {
		return
	}

	parts := parseCommand(cmd)
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
	case "e", "edit":
		if len(parts) != 2 {
			a.SetStatus(":edit <filename>")
		} else {
			filename := parts[1]
			if err := a.Load(parts[1]); err != nil {
				a.SetStatus(fmt.Sprintf("Failed to edit %s: %s", filename, err.Error()))
			} else {
				a.SetStatus(fmt.Sprintf("Opened %s", filename))
				a.splash.Hide()
				a.hasFile = true
			}
		}
	case "w", "write":
		var filename string
		if len(parts) > 1 {
			filename = parts[1]
		}
		if err := a.SaveAs(filename); err != nil {
			a.SetStatus("Failed to save: " + err.Error())
		} else {
			if filename != "" {
				a.SetStatus("Saved to " + filename)
			} else {
				a.SetStatus("Saved")
			}
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
	case "messages":
		a.handleMessagesCommand()
	case "export":
		if len(parts) < 3 {
			a.SetStatus("Usage: :export <format> <filename>")
			return
		}
		format := parts[1]
		filename := parts[2]
		// Sync tree items back to outline before exporting
		a.outline.Items = a.tree.GetItems()

		switch format {
		case "markdown":
			if err := export.ExportToMarkdown(a.outline, filename); err != nil {
				a.SetStatus("Failed to export: " + err.Error())
			} else {
				a.SetStatus("Exported to " + filename)
			}
		default:
			a.SetStatus("Unknown export format: " + format + " (use 'markdown' or 'text')")
		}
	case "import":
		if a.readOnly {
			a.SetStatus("Cannot modify readonly file")
			return
		}
		if len(parts) < 2 {
			a.SetStatus("Usage: :import <filename> [format]")
			return
		}
		filename := parts[1]
		format := import_parser.FormatAuto
		if len(parts) >= 3 {
			switch parts[2] {
			case "markdown", "md":
				format = import_parser.FormatMarkdown
			case "indented", "text", "txt":
				format = import_parser.FormatIndentedText
			default:
				a.SetStatus("Unknown import format: " + parts[2] + " (use 'markdown' or 'indented')")
				return
			}
		} else {
			// Auto-detect format from extension
			format = import_parser.DetectFormat(filename)
		}

		// Read file content
		content, err := os.ReadFile(filename)
		if err != nil {
			a.SetStatus("Failed to read file: " + err.Error())
			return
		}

		// Parse content
		items, err := import_parser.ImportFile(string(content), format)
		if err != nil {
			a.SetStatus("Failed to import: " + err.Error())
			return
		}

		if len(items) == 0 {
			a.SetStatus("No items found in file")
			return
		}

		// Add items to tree at current position
		currentIdx := a.tree.GetSelectedIndex()
		if currentIdx < 0 {
			// No selection, add to root
			for _, item := range items {
				a.tree.AddItemAfter(item)
			}
		} else {
			// Add as children of current item
			selectedItem := a.tree.GetSelected()
			if selectedItem != nil {
				for _, item := range items {
					selectedItem.Children = append(selectedItem.Children, item)
					item.Parent = selectedItem
				}
				selectedItem.Expanded = true // Auto-expand to show imported items
			}
		}

		// Sync outline and update tree
		a.outline.Items = a.tree.GetItems()
		a.tree.SetItems(a.outline.Items) // Update tree's items and rebuild view
		a.dirty = true

		// Force a complete redraw
		a.screen.Clear()

		a.SetStatus(fmt.Sprintf("Imported %d items from %s", len(items), filename))
	case "dailynote":
		if a.readOnly {
			a.SetStatus("Cannot modify readonly file")
			return
		}
		// Create or navigate to today's daily note
		now := time.Now()
		today := now.Format("2006-01-02")               // ISO format for attribute storage
		formattedDate := now.Format("Mon, Jan 2, 2006") // Short day name format for display

		// Look for existing daily note with today's date (search by date attribute)
		var foundItem *model.Item
		for _, item := range a.tree.GetItems() {
			if item.Metadata != nil && item.Metadata.Attributes != nil {
				if item.Metadata.Attributes["date"] == today {
					foundItem = item
					break
				}
			}
		}

		// If not found, create new item with formatted date
		if foundItem == nil {
			item := model.NewItem(formattedDate)
			a.tree.AddItemAfter(item)
			// Find the newly created item
			for _, dispItem := range a.tree.GetDisplayItems() {
				if dispItem.Item.Text == formattedDate {
					foundItem = dispItem.Item
					break
				}
			}
			if foundItem != nil {
				// Add type and date attributes to the daily note
				if foundItem.Metadata == nil {
					foundItem.Metadata = &model.Metadata{
						Attributes: make(map[string]string),
						Created:    time.Now(),
						Modified:   time.Now(),
					}
				}
				if foundItem.Metadata.Attributes == nil {
					foundItem.Metadata.Attributes = make(map[string]string)
				}
				foundItem.Metadata.Attributes["type"] = "day"
				foundItem.Metadata.Attributes["date"] = today
			}
			a.dirty = true
			a.SetStatus("Created daily note for " + formattedDate)
		} else {
			// Navigate to existing daily note
			for idx, dispItem := range a.tree.GetDisplayItems() {
				if dispItem.Item.ID == foundItem.ID {
					a.tree.SelectItem(idx)
					break
				}
			}
			a.SetStatus("Navigated to daily note for " + formattedDate)
		}
	case "attr":
		a.handleAttrCommand(parts)
	case "set":
		a.handleSetCommand(parts)
	case "search":
		a.handleSearchCommand(parts)
	case "calendar":
		a.handleCalendarCommand(parts)
	case "links":
		a.handleLinksCommand(parts)
	case "diff":
		a.handleDiffCommand(parts)
	case "typedef":
		a.handleTypedefCommand(parts)
	default:
		a.SetStatus("Unknown command: " + parts[0])
	}
}

// Save saves the outline to disk
func (a *App) Save() error {
	// Sync tree items back to outline before saving
	a.outline.Items = a.tree.GetItems()

	if err := a.store.Save(a.outline); err != nil {
		return err
	}
	a.dirty = false
	a.autoSaveTime = time.Now()
	return nil
}

func (a *App) Load(filename string) error {
	a.store.FilePath = filename
	// Re-check if the file is readonly (e.g., a backup file)
	a.store.ReadOnly = storage.IsBackupFile(filename)
	outline, err := a.store.Load()
	if err != nil {
		return err
	}

	a.outline = outline
	a.tree = ui.NewTreeView(outline.Items)
	a.dirty = false
	a.autoSaveTime = time.Now()
	// Update the app's readonly flag based on the store's status
	a.readOnly = a.store.ReadOnly
	return nil
}

// SaveAs saves the outline to a specified file path
// If filename is empty, uses the current store's file path
// If successful, updates the store's FilePath to the new location
func (a *App) SaveAs(filename string) error {
	if filename == "" {
		return a.Save()
	}

	// Sync tree items back to outline before saving
	a.outline.Items = a.tree.GetItems()

	// Save to the specified filename
	if err := a.store.SaveToFile(a.outline, filename); err != nil {
		return err
	}

	// Update the store's file path for future saves
	a.store.FilePath = filename

	// Hide splash screen when saving to a file
	if !a.hasFile {
		a.splash.Hide()
		a.hasFile = true
	}

	a.dirty = false
	a.autoSaveTime = time.Now()
	return nil
}

// handleVisualMode handles input while in visual mode
func (a *App) handleVisualMode(ev *tcell.EventKey) {
	// Handle special keys for visual mode
	switch ev.Key() {
	case tcell.KeyDown:
		a.tree.SelectNext()
		a.pendingKeySeq = 0
		return
	case tcell.KeyUp:
		a.tree.SelectPrev()
		a.pendingKeySeq = 0
		return
	case tcell.KeyLeft:
		a.tree.Collapse()
		a.pendingKeySeq = 0
		return
	case tcell.KeyRight:
		a.tree.Expand(true)
		a.pendingKeySeq = 0
		return
	case tcell.KeyEscape:
		// Exit visual mode
		a.mode = NormalMode
		a.visualAnchor = -1
		a.pendingKeySeq = 0
		a.SetStatus("Exited visual mode")
		return
	}

	// Handle character keys
	key := ev.Rune()

	// Check if we're waiting for a second key of a pending key sequence
	if a.pendingKeySeq != 0 {
		pendingKey := a.GetPendingKeyBindingByPrefix(a.pendingKeySeq)
		if pendingKey != nil {
			if seqBinding, ok := pendingKey.Sequences[key]; ok {
				// Execute the pending key sequence
				seqBinding.Handler(a)
				a.pendingKeySeq = 0
				return
			}
		}
		// Clear pending sequence if second key didn't match
		a.pendingKeySeq = 0
	}

	// Check if this is a pending key prefix
	if a.IsPendingKeyPrefix(key) {
		a.pendingKeySeq = key
		return
	}

	// Handle visual keybindings
	kb := a.GetVisualKeybindingByKey(key)
	if kb != nil {
		kb.Handler(a)
	}
}

// deleteVisualSelection deletes all items in the visual selection range
func (a *App) deleteVisualSelection() {
	start, end := a.getVisualSelectionRange()
	if start < 0 || end < 0 {
		a.SetStatus("No selection")
		return
	}

	// Get all items in the selection range
	items := a.tree.GetItemsInRange(start, end)
	if len(items) == 0 {
		a.SetStatus("Nothing to delete")
		return
	}

	// Delete each item
	for _, item := range items {
		a.tree.DeleteItem(item)
	}

	a.mode = NormalMode
	a.visualAnchor = -1
	a.SetStatus(fmt.Sprintf("Deleted %d items", len(items)))
	a.dirty = true
}

// yankVisualSelection yanks (copies) all items in the visual selection range
func (a *App) yankVisualSelection() {
	start, end := a.getVisualSelectionRange()
	if start < 0 || end < 0 {
		a.SetStatus("No selection")
		return
	}

	// Get all items in the selection range
	items := a.tree.GetItemsInRange(start, end)
	if len(items) == 0 {
		a.SetStatus("Nothing to yank")
		return
	}

	// For now, store just the first item in clipboard
	// TODO: Extend clipboard to support multiple items
	if len(items) > 0 {
		a.clipboard = items[0]
	}

	a.mode = NormalMode
	a.visualAnchor = -1
	a.SetStatus(fmt.Sprintf("Yanked %d items", len(items)))
}

// indentVisualSelection indents all items in the visual selection range
func (a *App) indentVisualSelection() {
	start, end := a.getVisualSelectionRange()
	if start < 0 || end < 0 {
		a.SetStatus("No selection")
		return
	}

	// Get all items in the selection range
	items := a.tree.GetItemsInRange(start, end)
	if len(items) == 0 {
		a.SetStatus("Nothing to indent")
		return
	}

	// Indent each item
	count := 0
	for _, item := range items {
		if a.tree.IndentItem(item) {
			count++
		}
	}

	a.mode = NormalMode
	a.visualAnchor = -1
	a.SetStatus(fmt.Sprintf("Indented %d items", count))
	a.dirty = true
}

// outdentVisualSelection outdents all items in the visual selection range
func (a *App) outdentVisualSelection() {
	start, end := a.getVisualSelectionRange()
	if start < 0 || end < 0 {
		a.SetStatus("No selection")
		return
	}

	// Get all items in the selection range
	items := a.tree.GetItemsInRange(start, end)
	if len(items) == 0 {
		a.SetStatus("Nothing to outdent")
		return
	}

	// Outdent each item
	count := 0
	for _, item := range items {
		if a.tree.OutdentItem(item) {
			count++
		}
	}

	a.mode = NormalMode
	a.visualAnchor = -1
	a.SetStatus(fmt.Sprintf("Outdented %d items", count))
	a.dirty = true
}

// getVisualSelectionRange returns the start and end indices of the visual selection
// Returns -1, -1 if not in visual selection
func (a *App) getVisualSelectionRange() (int, int) {
	if a.visualAnchor < 0 {
		return -1, -1
	}

	current := a.tree.GetSelectedIndex()
	start := a.visualAnchor
	end := current

	if start > end {
		start, end = end, start
	}

	return start, end
}

// handleTreeMouseClick handles mouse clicks in the tree view
func (a *App) handleTreeMouseClick(mouseEv *tcell.EventMouse) {
	// Only handle left mouse button
	if mouseEv.Buttons()&tcell.Button1 == 0 {
		return
	}

	x, y := mouseEv.Position()

	// FIXME: Tree starts at Y = 0
	treeStartY := 0

	// Check if click is within tree area
	if y < treeStartY {
		return
	}

	// Calculate which display line was clicked
	// Need to add viewport offset to account for scrolling
	viewportOffset := a.tree.GetViewportOffset()
	displayLineIdx := y - treeStartY + viewportOffset

	// Check if we're in search mode
	if a.search.IsActive() {
		// For search results, just select the item
		results := a.search.GetResults()
		if displayLineIdx >= 0 && displayLineIdx < len(results) {
			a.tree = ui.NewTreeView(results)
			if displayLineIdx < len(a.tree.GetDisplayItems()) {
				for i := 0; i < displayLineIdx; i++ {
					a.tree.SelectNext()
				}
			}
		}
		return
	}

	// Convert display line index to item index in filtered view
	itemIdx := a.tree.GetItemFromDisplayLine(displayLineIdx)
	if itemIdx < 0 {
		return
	}

	displayItems := a.tree.GetDisplayItems()
	if itemIdx >= len(displayItems) {
		return
	}

	// Check if click was on the arrow (expand/collapse)
	// Only process arrow clicks on the first line of the item
	displayLine := a.tree.GetDisplayLines()[displayLineIdx]
	if displayLine.ItemStartLine {
		dispItem := displayItems[itemIdx]
		arrowX := dispItem.Depth * 3

		// Arrow is at position arrowX, click is on it if within those bounds
		if x >= arrowX && x < arrowX+1 && len(dispItem.Item.Children) > 0 {
			// Click was on the arrow
			a.tree.SelectItem(itemIdx)
			if dispItem.Item.Expanded {
				a.tree.Collapse()
			} else {
				a.tree.Expand(false)
			}
			return
		}
	}

	// Otherwise, just select the item
	a.tree.SelectItem(itemIdx)
}

// handleEditorMouseClick handles mouse clicks while editing
func (a *App) handleEditorMouseClick(mouseEv *tcell.EventMouse) {
	// Only handle left mouse button
	if mouseEv.Buttons()&tcell.Button1 == 0 {
		return
	}

	x, _ := mouseEv.Position()

	// Get editor position info from the last render
	selectedIdx := a.tree.GetSelectedIndex()
	if selectedIdx < 0 {
		return
	}

	// Get editor details from the app's last render state
	// We need to calculate the editor's screen position
	selected := a.tree.GetSelected()
	if selected == nil {
		return
	}

	depth := a.tree.GetSelectedDepth()
	editorX := depth*3 + 3 // indentation + arrow + space

	// Calculate cursor position from click
	if x >= editorX {
		relativeX := x - editorX
		a.editor.SetCursorFromScreenX(relativeX)
	}
}

// SetStatus sets the status message
func (a *App) SetStatus(msg string) {
	a.statusMsg = msg
	a.statusTime = time.Now()
	// Add message to history
	if a.messageLogger != nil {
		a.messageLogger.AddMessage(msg)
	}
}

// Quit signals the app to quit
func (a *App) Quit() {
	a.quit = true
}

// SetDebugMode enables or disables debug mode
func (a *App) SetDebugMode(debug bool) {
	a.debugMode = debug
}

// handleMessagesCommand displays the message history
func (a *App) handleMessagesCommand() {
	if a.messageLogger == nil {
		a.SetStatus("Message history not available")
		return
	}

	messages := a.messageLogger.GetMessagesReverse() // Newest first
	if len(messages) == 0 {
		a.SetStatus("No messages in history")
		return
	}

	// Store messages in a temporary state for rendering
	a.messagesViewActive = true
	a.messagesViewMessages = messages
	a.messagesViewScroll = 0
}

// handleAttrCommand processes attribute-related commands
func (a *App) handleAttrCommand(parts []string) {
	// Check if trying to modify attributes on a readonly file
	if len(parts) > 1 && parts[1] != "list" && parts[1] != "" {
		if a.readOnly {
			a.SetStatus("Cannot modify readonly file")
			return
		}
	}

	selected := a.tree.GetSelected()
	if selected == nil {
		a.SetStatus("No item selected")
		return
	}

	// Ensure metadata exists
	if selected.Metadata == nil {
		selected.Metadata = &model.Metadata{
			Attributes: make(map[string]string),
			Created:    time.Now(),
			Modified:   time.Now(),
		}
	}

	// Ensure attributes map exists
	if selected.Metadata.Attributes == nil {
		selected.Metadata.Attributes = make(map[string]string)
	}

	if len(parts) < 2 {
		// Show all attributes
		a.showAttributes(selected)
		return
	}

	switch parts[1] {
	case "add", "set":
		if len(parts) < 4 {
			a.SetStatus("Usage: :attr add <key> <value>")
			return
		}
		key := parts[2]
		value := strings.Join(parts[3:], " ")

		// Validate attribute value against type definitions if they exist
		if a.validateAttributeValue(key, value) == false {
			// validateAttributeValue sets the status message on error
			return
		}

		selected.Metadata.Attributes[key] = value
		selected.Metadata.Modified = time.Now()
		a.dirty = true
		a.SetStatus(fmt.Sprintf("Attribute '%s' set to '%s'", key, value))

	case "del", "delete", "remove":
		if len(parts) < 3 {
			a.SetStatus("Usage: :attr del <key>")
			return
		}
		key := parts[2]
		if _, exists := selected.Metadata.Attributes[key]; !exists {
			a.SetStatus(fmt.Sprintf("Attribute '%s' not found", key))
			return
		}
		delete(selected.Metadata.Attributes, key)
		selected.Metadata.Modified = time.Now()
		a.dirty = true
		a.SetStatus(fmt.Sprintf("Attribute '%s' deleted", key))

	case "list", "show", "view":
		a.showAttributes(selected)

	default:
		a.SetStatus("Unknown attr command: " + parts[1])
	}
}

// showAttributes displays all attributes for an item
func (a *App) showAttributes(item *model.Item) {
	if item.Metadata == nil || len(item.Metadata.Attributes) == 0 {
		a.SetStatus("No attributes for this item")
		return
	}

	// Build a formatted string of all attributes
	var lines []string
	for key, value := range item.Metadata.Attributes {
		lines = append(lines, fmt.Sprintf("%s: %s", key, value))
	}

	// Show all attributes in status bar (limit to first line for now)
	if len(lines) > 0 {
		a.SetStatus("Attributes: " + lines[0])
	}
}

// handleGoCommand opens a URL from the 'url' attribute using xdg-open
func (a *App) handleGoCommand() {
	selected := a.tree.GetSelected()
	if selected == nil {
		a.SetStatus("No item selected")
		return
	}

	// Check if item has attributes
	if selected.Metadata == nil || selected.Metadata.Attributes == nil {
		a.SetStatus("Item has no attributes")
		return
	}

	// Look for 'url' attribute
	url, exists := selected.Metadata.Attributes["url"]
	if !exists || url == "" {
		a.SetStatus("No 'url' attribute found for this item")
		return
	}

	// Try to open the URL with xdg-open
	cmd := exec.Command("xdg-open", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		a.SetStatus(fmt.Sprintf("Failed to open URL: %v", err))
	} else {
		a.SetStatus(fmt.Sprintf("Opening URL: %s", url))
	}
}

// handleGoReferencedCommand navigates to the referenced (original) item if the current item is a virtual reference
func (a *App) handleGoReferencedCommand() {
	// Get the currently selected display item to check if it's a virtual reference
	displayItems := a.tree.GetDisplayItems()
	if len(displayItems) == 0 {
		a.SetStatus("No items in tree")
		return
	}

	selectedIdx := a.tree.GetSelectedIndex()
	if selectedIdx < 0 || selectedIdx >= len(displayItems) {
		a.SetStatus("No item selected")
		return
	}

	displayItem := displayItems[selectedIdx]

	// Check if this is a virtual reference
	if !displayItem.IsVirtual {
		a.SetStatus("Current item is not a reference")
		return
	}

	// Get the original item
	originalItem := displayItem.OriginalItem
	if originalItem == nil {
		a.SetStatus("Reference has no original item")
		return
	}

	// Expand parents to make the original item visible
	a.tree.ExpandParents(originalItem)

	// Find and select the original item in the display items
	updatedDisplayItems := a.tree.GetDisplayItems()
	for idx, dispItem := range updatedDisplayItems {
		if dispItem.Item.ID == originalItem.ID {
			a.tree.SelectItem(idx)
			a.SetStatus(fmt.Sprintf("Navigated to referenced item: %s", originalItem.Text))
			return
		}
	}

	// If we couldn't find it after expanding parents, something went wrong
	a.SetStatus("Could not navigate to referenced item")
}

// handleFollowLinkCommand follows a wiki-style link [[id]] in the current item
func (a *App) handleFollowLinkCommand() {
	selected := a.tree.GetSelected()
	if selected == nil {
		a.SetStatus("No item selected")
		return
	}

	// Parse links from the selected item's text
	itemLinks := links.ParseLinks(selected.Text)
	if len(itemLinks) == 0 {
		a.SetStatus("No links found in current item")
		return
	}

	// For now, navigate to the first link
	// TODO: If multiple links, could show a selection menu
	link := itemLinks[0]

	// Find the target item by ID
	targetItem := a.outline.FindItemByID(link.ID)
	if targetItem == nil {
		a.SetStatus(fmt.Sprintf("Broken link: target item not found (%s)", link.ID))
		return
	}

	// Expand parents to make the target item visible
	a.tree.ExpandParents(targetItem)

	// Find and select the target item in the display items
	displayItems := a.tree.GetDisplayItems()
	for idx, dispItem := range displayItems {
		if dispItem.Item.ID == targetItem.ID {
			a.tree.SelectItem(idx)
			displayText := link.GetDisplayText()
			a.SetStatus(fmt.Sprintf("Followed link to: %s", displayText))
			return
		}
	}

	// If we couldn't find it after expanding parents, something went wrong
	a.SetStatus("Could not navigate to linked item")
}

// handleLinksCommand displays all links in the current item
func (a *App) handleLinksCommand(parts []string) {
	selected := a.tree.GetSelected()
	if selected == nil {
		a.SetStatus("No item selected")
		return
	}

	// Parse links from the selected item's text
	itemLinks := links.ParseLinks(selected.Text)
	if len(itemLinks) == 0 {
		a.SetStatus("No links found in current item")
		return
	}

	// Build a list of link descriptions
	var linkDescs []string
	for i, link := range itemLinks {
		displayText := link.GetDisplayText()
		linkDescs = append(linkDescs, fmt.Sprintf("%d. %s -> %s", i+1, displayText, link.ID))
	}

	// Show the first few links in status (for longer lists, show just count)
	if len(linkDescs) == 1 {
		a.SetStatus("Found 1 link: " + linkDescs[0])
	} else if len(linkDescs) <= 3 {
		a.SetStatus(fmt.Sprintf("Found %d links: %s", len(linkDescs), strings.Join(linkDescs, " | ")))
	} else {
		a.SetStatus(fmt.Sprintf("Found %d links (use gf to follow first one)", len(linkDescs)))
	}
}

// handlePasteAsChildCommand pastes the clipboard item as a child of the selected item
func (a *App) handlePasteAsChildCommand() {
	if a.clipboard == nil {
		a.SetStatus("Nothing to paste (clipboard is empty)")
		return
	}

	newItem := model.NewItemFrom(a.clipboard)

	a.tree.AddItemAsChild(newItem)
	a.dirty = true
	a.SetStatus(fmt.Sprintf("Pasted as child: %s", newItem.Text))
}

// handleSendToNode opens the node search widget to select a destination and sends the current item there
func (a *App) handleSendToNode() {
	if a.readOnly {
		a.SetStatus("File is readonly")
		return
	}

	selected := a.tree.GetSelected()
	if selected == nil {
		a.SetStatus("No item selected")
		return
	}

	// Set up the node search widget
	a.nodeSearchWidget.SetItems(a.outline.GetAllItems())
	a.nodeSearchWidget.SetOnSelect(func(destination *model.Item) {
		if destination == nil {
			a.SetStatus("No destination selected")
			return
		}

		// Attempt to send the item
		if a.tree.SendItemToNode(destination) {
			a.lastSendDestination = destination
			a.dirty = true
			// Truncate destination text if it's too long for status display
			destText := destination.Text
			if len(destText) > 40 {
				destText = destText[:37] + "..."
			}
			a.SetStatus(fmt.Sprintf("Sent to: %s", destText))
		} else {
			a.SetStatus("Cannot send item (circular reference or invalid destination)")
		}
	})
	a.nodeSearchWidget.Show()
	a.SetStatus("Select destination node (Enter to select, Escape to cancel)")
}

// handleSendToLastNode sends the current item to the last destination used with 'ss'
func (a *App) handleSendToLastNode() {
	if a.readOnly {
		a.SetStatus("File is readonly")
		return
	}

	if a.lastSendDestination == nil {
		a.SetStatus("No previous send destination (use 'ss' first)")
		return
	}

	selected := a.tree.GetSelected()
	if selected == nil {
		a.SetStatus("No item selected")
		return
	}

	// Attempt to send the item
	if a.tree.SendItemToNode(a.lastSendDestination) {
		a.dirty = true
		// Truncate destination text if it's too long for status display
		destText := a.lastSendDestination.Text
		if len(destText) > 40 {
			destText = destText[:37] + "..."
		}
		a.SetStatus(fmt.Sprintf("Sent to: %s", destText))
	} else {
		a.SetStatus("Cannot send item (circular reference or destination no longer valid)")
	}
}

// handleExternalEdit opens the current item in an external editor
func (a *App) handleExternalEdit() {
	selected := a.tree.GetSelected()
	if selected == nil {
		a.SetStatus("No item selected")
		return
	}

	// Suspend tcell to release terminal control for the editor
	a.screen.Suspend()

	// Create validation callback for external editor
	validateAttrs := func(attributes map[string]string) string {
		// Load type registry from outline
		registry := tmpl.NewTypeRegistry()
		if err := registry.LoadFromOutline(a.outline); err != nil {
			// If we can't load types, allow the attributes (type system is optional)
			return ""
		}

		// Validate each attribute
		for key, value := range attributes {
			if err := registry.Validate(key, value); err != nil {
				return err.Error()
			}
		}

		return ""
	}

	// Edit the item in external editor (editor now has full terminal control)
	err := ui.EditItemInExternalEditor(selected, a.cfg, validateAttrs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Editor error: %v\n", err)
	}

	// Resume tcell and restore terminal control
	if err := a.screen.Resume(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resume terminal: %v\n", err)
		os.Exit(1)
	}

	// Rebuild the tree view to show updated item text
	a.tree.RebuildView()

	// Force a complete redraw
	a.screen.Clear()
	a.screen.Show()

	// Mark as modified and update status
	a.dirty = true
	a.SetStatus("Item updated from external editor")
}

// handleSetCommand processes :set configuration commands
// Examples:
//
//	:set visattr "date"
//	:set visattr date
//	:set (shows all settings)
//	:set visattr (shows visattr value)
func (a *App) handleSetCommand(parts []string) {
	if len(parts) == 1 {
		// Show all settings
		allSettings := a.cfg.GetAll()
		if len(allSettings) == 0 {
			a.SetStatus("No configuration settings set")
			return
		}
		var settingsList []string
		for key, value := range allSettings {
			settingsList = append(settingsList, fmt.Sprintf("%s=%s", key, value))
		}
		a.SetStatus("Settings: " + strings.Join(settingsList, ", "))
		return
	}

	key := parts[1]

	if len(parts) == 2 {
		// Show specific setting
		value := a.cfg.Get(key)
		if value == "" {
			a.SetStatus(fmt.Sprintf("Setting '%s' is not set", key))
		} else {
			a.SetStatus(fmt.Sprintf("%s=%s", key, value))
		}
		return
	}

	// Set the configuration value
	// Combine all parts after the key to support values with spaces
	value := strings.Join(parts[2:], " ")

	// Remove surrounding quotes if present
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'')) {
		value = value[1 : len(value)-1]
	}

	a.cfg.Set(key, value)

	// Handle special settings that need to be applied immediately
	if key == "weekstart" {
		// Validate and apply weekstart to calendar widget
		if weekstart, err := strconv.Atoi(value); err == nil && weekstart >= 0 && weekstart <= 6 {
			a.calendarWidget.SetWeekStart(weekstart)
			a.SetStatus(fmt.Sprintf("Set %s = %s (0=Sun, 1=Mon, 2=Tue, 3=Wed, 4=Thu, 5=Fri, 6=Sat)", key, value))
		} else {
			a.SetStatus(fmt.Sprintf("Invalid weekstart value '%s'. Use 0-6 (0=Sunday, 1=Monday, ...)", value))
		}
	} else {
		a.SetStatus(fmt.Sprintf("Set %s = %s", key, value))
	}
}

// handleSearchCommand creates a new search node with the given query
// Usage: :search <query>
func (a *App) handleSearchCommand(parts []string) {
	if len(parts) < 2 {
		a.SetStatus("Usage: :search <query>")
		return
	}

	// Combine all parts after the command to form the query
	query := strings.Join(parts[1:], " ")

	// Create a new search node
	searchNode := model.NewItem("[Search] " + query)
	searchNode.Metadata.Attributes["type"] = "search"
	searchNode.Metadata.Attributes["query"] = query

	a.tree.AddItemAfter(searchNode)
	a.refreshSearchNodes()
	a.SetStatus(fmt.Sprintf("Created search node for: %s (use l to expand)", query))
	a.dirty = true
}

// populateSearchNode updates a single search node with current matching results
// Returns the count of matches found, or 0 if query is empty/invalid
func (a *App) populateSearchNode(item *model.Item) int {
	queryStr := item.GetSearchQuery()
	if queryStr == "" {
		return 0
	}

	// Parse and execute the query
	filterExpr, err := search.ParseQuery(queryStr)
	if err != nil {
		return 0
	}

	// Find matching items
	var matchingIDs []string
	for _, candidate := range a.outline.GetAllItems() {
		// Don't include the search node itself
		if candidate.ID == item.ID {
			continue
		}
		if filterExpr.Matches(candidate) {
			matchingIDs = append(matchingIDs, candidate.ID)
		}
	}

	return a.outline.PopulateSearchNode(item, matchingIDs)
}

// refreshSearchNodes refreshes all expanded search nodes with current results
func (a *App) refreshSearchNodes() {
	// Sync outline with tree before searching
	a.outline.Items = a.tree.GetItems()
	a.outline.BuildIndex()

	// Find all search nodes and refresh them
	for _, item := range a.outline.GetAllItems() {
		if item.IsSearchNode() && item.Expanded {
			a.populateSearchNode(item)
		}
	}

	// Rebuild the tree view to show updated results
	a.tree.RebuildView()
}

// handleCalendarCommand handles the :calendar command
// Usage:
//
//	:calendar - opens calendar in search mode
//	:calendar attr <name> - opens calendar in attribute mode for setting an attribute
func (a *App) handleCalendarCommand(parts []string) {
	if len(parts) == 1 {
		// Open calendar in search mode
		a.calendarWidget.Show()
		a.SetStatus("Calendar opened (press Esc to close)")
		return
	}

	if len(parts) >= 2 && parts[1] == "attr" {
		if len(parts) < 3 {
			a.SetStatus("Usage: :calendar attr <attribute-name>")
			return
		}
		attrName := parts[2]

		// Get current item and its attribute value if it exists
		currentItem := a.tree.GetSelected()
		dateValue := ""
		if currentItem != nil && currentItem.Metadata != nil && currentItem.Metadata.Attributes != nil {
			if val, ok := currentItem.Metadata.Attributes[attrName]; ok {
				dateValue = val
			}
		}

		a.calendarWidget.ShowForAttributeWithValue(attrName, dateValue)
		a.SetStatus(fmt.Sprintf("Calendar opened to set '%s' attribute", attrName))
		return
	}

	a.SetStatus("Usage: :calendar or :calendar attr <name>")
}

// handleDiffCommand shows diff view with selected backup
func (a *App) handleDiffCommand(parts []string) {
	if a.originalFilePath == "" {
		a.SetStatus("No file to compare backups for")
		return
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found for this file")
		return
	}

	if len(backups) < 1 {
		a.SetStatus("No backups found for this file")
		return
	}

	// Append the current file as a virtual "backup" entry at the end
	// Since backups are stored oldest-first, appending at the end makes it the newest
	// In reversed mode (newest-first display), it will appear at the top
	currentEntry := storage.BackupMetadata{
		FilePath:     "(current)",
		Timestamp:    time.Now(),
		SessionID:    "current",
		OriginalFile: a.originalFilePath,
	}
	backupsWithCurrent := make([]storage.BackupMetadata, 0, len(backups)+1)
	backupsWithCurrent = append(backupsWithCurrent, backups...)
	backupsWithCurrent = append(backupsWithCurrent, currentEntry)

	// Show backup selector with side-by-side diff preview
	a.backupSelectorWidget.Show(backupsWithCurrent, a.outline,
		func(backup storage.BackupMetadata) {
			// Don't restore if it's the "(current)" virtual entry
			if backup.FilePath == "(current)" {
				a.SetStatus("Already at current state")
				return
			}

			// Load the selected backup
			backupData, err := os.ReadFile(backup.FilePath)
			if err != nil {
				a.SetStatus(fmt.Sprintf("Failed to read backup: %v", err))
				return
			}

			var restoredOutline model.Outline
			if err := json.Unmarshal(backupData, &restoredOutline); err != nil {
				a.SetStatus(fmt.Sprintf("Failed to parse backup: %v", err))
				return
			}

			// Close any active editor
			if a.editor != nil {
				a.editor = nil
			}

			// Replace current outline with restored backup
			a.outline = &restoredOutline
			a.dirty = true

			// Recreate the tree view with restored items
			a.tree = ui.NewTreeView(a.outline.Items)

			// Close search widget if open
			if a.search != nil && a.search.IsActive() {
				a.search.Stop()
			}

			// Close node search widget if open
			if a.nodeSearchWidget != nil {
				a.nodeSearchWidget.Hide()
			}

			a.SetStatus(fmt.Sprintf("Restored backup from %s", backup.Timestamp.Format("2006-01-02 15:04:05")))
		},
		func() {
			a.SetStatus("Diff cancelled")
		})
}

// handlePreviousBackupSameSession loads the previous backup with the same session ID
func (a *App) handlePreviousBackupSameSession() bool {
	if a.originalFilePath == "" {
		a.SetStatus("No file to find backups for")
		return false
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return false
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found")
		return false
	}

	// Find current backup position in the list
	currentIdx := -1
	for i, b := range backups {
		if b.FilePath == a.currentBackupPath {
			currentIdx = i
			break
		}
	}

	// If not viewing a backup, go to the most recent one
	if currentIdx == -1 {
		// Find backups with same session ID
		for i := len(backups) - 1; i >= 0; i-- {
			if backups[i].SessionID == a.sessionID {
				if a.loadBackupFile(backups[i]) {
					a.SetStatus(fmt.Sprintf("Backup: %s (%s)", backups[i].Timestamp.Format("2006-01-02 15:04:05"), backups[i].SessionID))
					return true
				}
			}
		}
		a.SetStatus("No backups with current session ID")
		return false
	}

	// Find previous backup with same session ID
	for i := currentIdx - 1; i >= 0; i-- {
		if backups[i].SessionID == a.sessionID {
			if a.loadBackupFile(backups[i]) {
				a.SetStatus(fmt.Sprintf("Backup: %s (%s)", backups[i].Timestamp.Format("2006-01-02 15:04:05"), backups[i].SessionID))
				return true
			}
		}
	}

	a.SetStatus("No older backups with same session ID")
	return false
}

// handleNextBackupSameSession loads the next backup with the same session ID
func (a *App) handleNextBackupSameSession() bool {
	if a.originalFilePath == "" {
		a.SetStatus("No file to find backups for")
		return false
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return false
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found")
		return false
	}

	// Find current backup position
	currentIdx := -1
	for i, b := range backups {
		if b.FilePath == a.currentBackupPath {
			currentIdx = i
			break
		}
	}

	// Must be viewing a backup to go to next one
	if currentIdx == -1 {
		a.SetStatus("Not viewing a backup")
		return false
	}

	// Find next backup with same session ID
	for i := currentIdx + 1; i < len(backups); i++ {
		if backups[i].SessionID == a.sessionID {
			if a.loadBackupFile(backups[i]) {
				a.SetStatus(fmt.Sprintf("Backup: %s (%s)", backups[i].Timestamp.Format("2006-01-02 15:04:05"), backups[i].SessionID))
				return true
			}
		}
	}

	a.SetStatus("No newer backups with same session ID")
	return false
}

// handlePreviousBackupAnySession loads the previous backup regardless of session ID
func (a *App) handlePreviousBackupAnySession() bool {
	if a.originalFilePath == "" {
		a.SetStatus("No file to find backups for")
		return false
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return false
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found")
		return false
	}

	// Find current backup position
	currentIdx := -1
	for i, b := range backups {
		if b.FilePath == a.currentBackupPath {
			currentIdx = i
			break
		}
	}

	// If not viewing a backup, go to most recent
	if currentIdx == -1 {
		if a.loadBackupFile(backups[len(backups)-1]) {
			b := backups[len(backups)-1]
			a.SetStatus(fmt.Sprintf("Backup: %s (%s)", b.Timestamp.Format("2006-01-02 15:04:05"), b.SessionID))
			return true
		}
		return false
	}

	// Find previous backup
	if currentIdx > 0 {
		if a.loadBackupFile(backups[currentIdx-1]) {
			b := backups[currentIdx-1]
			a.SetStatus(fmt.Sprintf("Backup: %s (%s)", b.Timestamp.Format("2006-01-02 15:04:05"), b.SessionID))
			return true
		}
	}

	a.SetStatus("No older backups")
	return false
}

// handleNextBackupAnySession loads the next backup regardless of session ID
func (a *App) handleNextBackupAnySession() bool {
	if a.originalFilePath == "" {
		a.SetStatus("No file to find backups for")
		return false
	}

	backupMgr, err := storage.NewBackupManager()
	if err != nil {
		a.SetStatus("Failed to access backups")
		return false
	}

	backups, err := backupMgr.FindBackupsForFile(a.originalFilePath)
	if err != nil || len(backups) == 0 {
		a.SetStatus("No backups found")
		return false
	}

	// Find current backup position
	currentIdx := -1
	for i, b := range backups {
		if b.FilePath == a.currentBackupPath {
			currentIdx = i
			break
		}
	}

	// Must be viewing a backup to go to next one
	if currentIdx == -1 {
		a.SetStatus("Not viewing a backup")
		return false
	}

	// Find next backup
	if currentIdx < len(backups)-1 {
		if a.loadBackupFile(backups[currentIdx+1]) {
			b := backups[currentIdx+1]
			a.SetStatus(fmt.Sprintf("Backup: %s (%s)", b.Timestamp.Format("2006-01-02 15:04:05"), b.SessionID))
			return true
		}
	}

	a.SetStatus("No newer backups")
	return false
}

// loadBackupFile loads a backup file and updates the app state
func (a *App) loadBackupFile(backup storage.BackupMetadata) bool {
	if err := a.Load(backup.FilePath); err != nil {
		a.SetStatus(fmt.Sprintf("Failed to load backup: %s", err.Error()))
		return false
	}

	a.currentBackupPath = backup.FilePath
	a.originalFilePath = backup.OriginalFile
	a.sessionID = backup.SessionID

	return true
}

// generateSessionID creates a random 8-character session ID for backup naming
func generateSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
