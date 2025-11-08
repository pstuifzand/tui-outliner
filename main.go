package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pstuifzand/tui-outliner/internal/app"
	"github.com/pstuifzand/tui-outliner/internal/export"
	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/search"
	"github.com/pstuifzand/tui-outliner/internal/socket"
	"github.com/pstuifzand/tui-outliner/internal/storage"
	"github.com/pstuifzand/tui-outliner/internal/ui"
)

func main() {
	logFile, err := os.Create("tuo.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Check for subcommands
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "add":
			handleAddCommand()
			return
		case "export":
			handleExportCommand()
			return
		case "search":
			handleSearchCommand()
			return
		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	// Parse flags for main app
	debug := flag.Bool("debug", false, "Enable debug mode (shows key events in status)")
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	var filePath string

	if len(args) > 0 {
		filePath = args[0]
	}
	// filePath will be empty if no argument provided, which is allowed

	application, err := app.NewApp(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *debug {
		application.SetDebugMode(true)
	}

	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}

// attrFlags allows multiple --attr flags
type attrFlags []string

func (a *attrFlags) String() string {
	return strings.Join(*a, ", ")
}

func (a *attrFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

// handleAddCommand handles the 'add' subcommand
func handleAddCommand() {
	var attrs attrFlags
	var todoFlag bool
	var runningFlag bool
	var fileFlag string
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addCmd.Var(&attrs, "attr", "Set an attribute (key=value, can be used multiple times)")
	addCmd.Var(&attrs, "a", "Set an attribute (key=value, shorthand)")
	addCmd.BoolVar(&todoFlag, "t", false, "Add as a todo item (sets type=todo)")
	addCmd.BoolVar(&runningFlag, "r", false, "Add to running tuo instance")
	addCmd.StringVar(&fileFlag, "f", "", "Add to file")
	addCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tuo add [options] <text>\n")
		fmt.Fprintf(os.Stderr, "Add a node to the inbox of a running tuo instance or to a file\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -r                      Add to running tuo instance\n")
		fmt.Fprintf(os.Stderr, "  -f file                 Add to file\n")
		fmt.Fprintf(os.Stderr, "  -a, --attr key=value    Set an attribute (can be used multiple times)\n")
		fmt.Fprintf(os.Stderr, "  -t                      Add as todo item (sets type=todo)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  tuo add -r \"Buy milk\"                         # Add to running instance\n")
		fmt.Fprintf(os.Stderr, "  tuo add -r -t \"Call dentist\"                  # Add as todo to running instance\n")
		fmt.Fprintf(os.Stderr, "  tuo add -f notes.json \"Buy milk\"              # Add to file\n")
		fmt.Fprintf(os.Stderr, "  tuo add -f notes.json -t \"Call dentist\"       # Add as todo to file\n")
		fmt.Fprintf(os.Stderr, "  tuo add -r -a priority=high \"Important task\"\n")
	}

	if err := addCmd.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	// Get remaining args as the text to add
	text := strings.Join(addCmd.Args(), " ")
	text = strings.TrimSpace(text)

	if text == "" {
		fmt.Fprintf(os.Stderr, "Error: node text cannot be empty\n\n")
		addCmd.Usage()
		os.Exit(1)
	}

	// Validate that exactly one of -r or -f is specified
	if runningFlag && fileFlag != "" {
		fmt.Fprintf(os.Stderr, "Error: cannot specify both -r and -f\n\n")
		addCmd.Usage()
		os.Exit(1)
	}
	if !runningFlag && fileFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: must specify either -r or -f\n\n")
		addCmd.Usage()
		os.Exit(1)
	}

	// Parse attributes
	attributes := make(map[string]string)
	for _, attr := range attrs {
		parts := strings.SplitN(attr, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Error: invalid attribute format '%s' (expected key=value)\n\n", attr)
			addCmd.Usage()
			os.Exit(1)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			fmt.Fprintf(os.Stderr, "Error: attribute key cannot be empty\n\n")
			addCmd.Usage()
			os.Exit(1)
		}
		attributes[key] = value
	}

	// Handle -t flag for todo items
	if todoFlag {
		attributes["type"] = "todo"
	}

	// Check if we should add to a file or running instance
	if fileFlag != "" {
		// Add to file
		if err := addToFile(fileFlag, text, attributes); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Node added to %s\n", fileFlag)
	} else {
		// Add to running instance
		if err := sendAddNode(text, attributes); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Node added to inbox")
	}
}

// handleExportCommand handles the 'export' subcommand
func handleExportCommand() {
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	fileFlag := exportCmd.String("f", "", "Input outline file to export")
	outputFlag := exportCmd.String("o", "", "Output file (defaults to stdout)")
	exportCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tuo export -f <input.json> [-o output.md]\n")
		fmt.Fprintf(os.Stderr, "Export an outline file to markdown format\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -f file      Input outline file to export\n")
		fmt.Fprintf(os.Stderr, "  -o file      Output file (defaults to stdout)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  tuo export -f notes.json              # Output to stdout\n")
		fmt.Fprintf(os.Stderr, "  tuo export -f notes.json -o notes.md  # Output to file\n")
		fmt.Fprintf(os.Stderr, "  tuo export -f notes.json | less       # Pipe to pager\n")
	}

	if err := exportCmd.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	// Validate that -f is specified
	if *fileFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: -f flag is required\n\n")
		exportCmd.Usage()
		os.Exit(1)
	}

	inputFile := strings.TrimSpace(*fileFlag)
	if inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: input filename cannot be empty\n\n")
		exportCmd.Usage()
		os.Exit(1)
	}

	// Load the outline from the input file
	store := storage.NewJSONStore(inputFile)
	outline, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading outline: %v\n", err)
		os.Exit(1)
	}

	// Determine output destination
	if *outputFlag != "" {
		// Output to file
		outputFile := strings.TrimSpace(*outputFlag)
		if outputFile == "" {
			fmt.Fprintf(os.Stderr, "Error: output filename cannot be empty\n\n")
			exportCmd.Usage()
			os.Exit(1)
		}

		if err := export.ExportToMarkdown(outline, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting to markdown: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Exported %s to %s\n", inputFile, outputFile)
	} else {
		// Output to stdout
		if err := export.ExportToMarkdownWriter(outline, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting to markdown: %v\n", err)
			os.Exit(1)
		}
	}
}

// handleSearchCommand handles the 'search' subcommand
func handleSearchCommand() {
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	runningFlag := searchCmd.Bool("r", false, "Search in running tuo instance")
	fileFlag := searchCmd.String("f", "", "Search in file")
	jsonFlag := searchCmd.Bool("json", false, "Output results as JSON (legacy, use -ff json)")
	ffFlag := searchCmd.String("ff", "", "Output format: text, fields, json, jsonl")
	fieldsFlag := searchCmd.String("fields", "", "Comma-separated fields: id,text,created,etc")
	searchCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tuo search [options] <query>\n")
		fmt.Fprintf(os.Stderr, "Search for nodes matching the query\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -r               Search in running tuo instance\n")
		fmt.Fprintf(os.Stderr, "  -f file          Search in file\n")
		fmt.Fprintf(os.Stderr, "  -ff format       Output format: text, fields, json, jsonl (default: text)\n")
		fmt.Fprintf(os.Stderr, "  --fields list    Comma-separated fields for json/jsonl/fields output\n")
		fmt.Fprintf(os.Stderr, "  -json            Output results as JSON (deprecated, use -ff json)\n\n")
		fmt.Fprintf(os.Stderr, "Output Formats:\n")
		fmt.Fprintf(os.Stderr, "  text   - Human-readable text format (default)\n")
		fmt.Fprintf(os.Stderr, "  fields - Tab-separated values for piping to Unix tools\n")
		fmt.Fprintf(os.Stderr, "  json   - Pretty-printed JSON array\n")
		fmt.Fprintf(os.Stderr, "  jsonl  - JSON Lines format (one object per line)\n\n")
		fmt.Fprintf(os.Stderr, "Available Fields:\n")
		fmt.Fprintf(os.Stderr, "  id, text, attributes, created, modified, tags, depth, path, parent_id\n")
		fmt.Fprintf(os.Stderr, "  Use attr:<name> for specific attributes (e.g., attr:status)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  tuo search -f notes.json \"todo\"\n")
		fmt.Fprintf(os.Stderr, "  tuo search -r \"@type=todo\" -ff json\n")
		fmt.Fprintf(os.Stderr, "  tuo search -f work.json \"@status=done\" -ff fields\n")
		fmt.Fprintf(os.Stderr, "  tuo search -f work.json \"task\" -ff json --fields id,text,created\n")
	}

	if err := searchCmd.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	args := searchCmd.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: query required\n\n")
		searchCmd.Usage()
		os.Exit(1)
	}

	query := args[0]

	// Validate that exactly one of -r or -f is specified
	if *runningFlag && *fileFlag != "" {
		fmt.Fprintf(os.Stderr, "Error: cannot specify both -r and -f\n\n")
		searchCmd.Usage()
		os.Exit(1)
	}
	if !*runningFlag && *fileFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: must specify either -r or -f\n\n")
		searchCmd.Usage()
		os.Exit(1)
	}

	// Determine output format (support both legacy -json and new -ff)
	var outputFormat string
	if *ffFlag != "" {
		outputFormat = *ffFlag
	} else if *jsonFlag {
		outputFormat = "json"
	} else {
		outputFormat = "text"
	}

	if *runningFlag {
		// Search in running instance
		if err := searchRunningInstance(query, outputFormat, *fieldsFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Search in file
		if err := searchFile(query, *fileFlag, outputFormat, *fieldsFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// searchRunningInstance searches in a running tuo instance via socket
func searchRunningInstance(query string, outputFormat string, fieldsStr string) error {
	// Find running instance
	socketPath, pid, err := socket.FindRunningInstance()
	if err != nil {
		return fmt.Errorf("no running tuo instance found: %w", err)
	}

	log.Printf("Found running instance at PID %d: %s", pid, socketPath)

	// Create client
	client, err := socket.NewClient(socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Send search command
	response, err := client.SendSearch(query)
	if err != nil {
		return fmt.Errorf("failed to send search: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("search failed: %s", response.Message)
	}

	// Output results based on format
	// Note: socket response only has basic fields, so we support text and json formats
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response.Results); err != nil {
			return fmt.Errorf("failed to encode results: %w", err)
		}
	case "fields", "jsonl":
		// These formats need full item metadata, only available when searching files
		return fmt.Errorf("format '%s' is only supported when searching files (-f)", outputFormat)
	default:
		// text format (default)
		if len(response.Results) == 0 {
			fmt.Println("No matches found")
		} else {
			fmt.Printf("Found %d match(es):\n\n", len(response.Results))
			for i, result := range response.Results {
				fmt.Printf("%d. %s\n", i+1, result.Text)
				if len(result.Path) > 0 {
					fmt.Printf("   Path: %s\n", strings.Join(result.Path, " > "))
				}
				if len(result.Attrs) > 0 {
					fmt.Printf("   Attributes: ")
					first := true
					for k, v := range result.Attrs {
						if !first {
							fmt.Printf(", ")
						}
						fmt.Printf("%s=%s", k, v)
						first = false
					}
					fmt.Println()
				}
				fmt.Println()
			}
		}
	}

	return nil
}

// searchFile searches in an outline file
func searchFile(query, filePath string, outputFormat string, fieldsStr string) error {
	// Load the outline file
	store := storage.NewJSONStore(filePath)
	outline, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load outline: %w", err)
	}

	// Parse the search query
	filterExpr, err := search.ParseQuery(query)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	// Get matching items
	matches, err := search.GetAlllByQuery(outline, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No matches found")
		return nil
	}

	// Ensure outline has indexed items for proper parent references
	outline.BuildIndex()

	// Determine output format
	switch outputFormat {
	case "fields":
		// Use SearchOutputFormatter for fields format
		formatter := ui.NewSearchOutputFormatter()
		fields := ui.ParseFieldsFlag(fieldsStr)
		output, err := formatter.FormatResults(matches, ui.OutputFormatFields, fields, outline)
		if err != nil {
			return fmt.Errorf("failed to format results: %w", err)
		}
		if output != "" {
			fmt.Println(output)
		}

	case "json":
		// Use SearchOutputFormatter for JSON format
		formatter := ui.NewSearchOutputFormatter()
		fields := ui.ParseFieldsFlag(fieldsStr)
		output, err := formatter.FormatResults(matches, ui.OutputFormatJSON, fields, outline)
		if err != nil {
			return fmt.Errorf("failed to format results: %w", err)
		}
		if output != "" {
			fmt.Println(output)
		}

	case "jsonl":
		// Use SearchOutputFormatter for JSON Lines format
		formatter := ui.NewSearchOutputFormatter()
		fields := ui.ParseFieldsFlag(fieldsStr)
		output, err := formatter.FormatResults(matches, ui.OutputFormatJSONL, fields, outline)
		if err != nil {
			return fmt.Errorf("failed to format results: %w", err)
		}
		if output != "" {
			fmt.Println(output)
		}

	default:
		// text format (default) - use SearchResult objects for backward compatibility
		results := make([]socket.SearchResult, 0, len(matches))
		for _, item := range matches {
			result := socket.SearchResult{
				Text: item.Text,
				Path: buildItemPathForCLI(item),
			}
			if item.Metadata != nil && item.Metadata.Attributes != nil {
				result.Attrs = item.Metadata.Attributes
			}
			results = append(results, result)
		}

		fmt.Printf("Found %d match(es):\n\n", len(results))
		for i, result := range results {
			fmt.Printf("%d. %s\n", i+1, result.Text)
			if len(result.Path) > 0 {
				fmt.Printf("   Path: %s\n", strings.Join(result.Path, " > "))
			}
			if len(result.Attrs) > 0 {
				fmt.Printf("   Attributes: ")
				first := true
				for k, v := range result.Attrs {
					if !first {
						fmt.Printf(", ")
					}
					fmt.Printf("%s=%s", k, v)
					first = false
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}

	// Suppress unused variable warning
	_ = filterExpr

	return nil
}

// buildItemPathForCLI constructs a path array for an item showing its hierarchy
func buildItemPathForCLI(item *model.Item) []string {
	var path []string
	current := item
	for current != nil {
		path = append([]string{current.Text}, path...)
		current = current.Parent
	}
	return path
}

// printUsage prints the main usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "tuo - TUI Outliner\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  tuo [options] [file]                  Start tuo with optional file\n")
	fmt.Fprintf(os.Stderr, "  tuo add -r|-f <file> [options] <text> Add node to running instance or file\n")
	fmt.Fprintf(os.Stderr, "  tuo export -f <file> [-o output]      Export outline to markdown\n")
	fmt.Fprintf(os.Stderr, "  tuo search -r|-f <file> <query>       Search for nodes\n")
	fmt.Fprintf(os.Stderr, "  tuo help                              Show this help message\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  --debug                               Enable debug mode\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  tuo                                   Start with empty outline\n")
	fmt.Fprintf(os.Stderr, "  tuo notes.json                        Open notes.json\n")
	fmt.Fprintf(os.Stderr, "  tuo --debug test.json                 Open test.json in debug mode\n")
	fmt.Fprintf(os.Stderr, "  tuo add -r \"Buy milk\"                 Add item to running instance\n")
	fmt.Fprintf(os.Stderr, "  tuo add -f notes.json \"Buy milk\"      Add item to file\n")
	fmt.Fprintf(os.Stderr, "  tuo export -f notes.json              Export to stdout\n")
	fmt.Fprintf(os.Stderr, "  tuo export -f notes.json -o notes.md  Export to file\n")
	fmt.Fprintf(os.Stderr, "  tuo search -f notes.json \"todo\"       Search for 'todo' in file\n")
	fmt.Fprintf(os.Stderr, "  tuo search -r \"@type=todo\"            Search running instance\n")
}

// addToFile adds a node directly to a file's inbox
func addToFile(filePath, text string, attributes map[string]string) error {
	// Load the outline from file
	store := storage.NewJSONStore(filePath)
	outline, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load outline: %w", err)
	}

	// Ensure items is initialized
	if outline.Items == nil {
		outline.Items = []*model.Item{}
	}

	// Find or create inbox node
	inbox := findInboxInOutline(outline)
	if inbox == nil {
		// Create new inbox at root level
		inbox = model.NewItem("Inbox")
		if inbox.Metadata.Attributes == nil {
			inbox.Metadata.Attributes = make(map[string]string)
		}
		inbox.Metadata.Attributes["type"] = "inbox"
		inbox.Expanded = true
		outline.Items = append(outline.Items, inbox)
	}

	// Create new item
	newItem := model.NewItem(text)
	if len(attributes) > 0 {
		if newItem.Metadata.Attributes == nil {
			newItem.Metadata.Attributes = make(map[string]string)
		}
		for key, value := range attributes {
			newItem.Metadata.Attributes[key] = value
		}
	}

	// Add to inbox
	inbox.AddChild(newItem)

	// Save the file
	if err := store.Save(outline); err != nil {
		return fmt.Errorf("failed to save outline: %w", err)
	}

	return nil
}

// findInboxInOutline searches for a node marked with type=inbox attribute
func findInboxInOutline(outline *model.Outline) *model.Item {
	var search func([]*model.Item) *model.Item
	search = func(items []*model.Item) *model.Item {
		for _, item := range items {
			if item.Metadata != nil && item.Metadata.Attributes != nil {
				if typeVal, ok := item.Metadata.Attributes["type"]; ok && typeVal == "inbox" {
					return item
				}
			}
			if len(item.Children) > 0 {
				if found := search(item.Children); found != nil {
					return found
				}
			}
		}
		return nil
	}
	return search(outline.Items)
}

// sendAddNode sends an add_node command to a running tuo instance
func sendAddNode(text string, attributes map[string]string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("node text cannot be empty")
	}

	// Find running instance
	socketPath, pid, err := socket.FindRunningInstance()
	if err != nil {
		return fmt.Errorf("no running tuo instance found: %w", err)
	}

	log.Printf("Found running instance at PID %d: %s", pid, socketPath)

	// Create client
	client, err := socket.NewClient(socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Send add_node command
	response, err := client.SendAddNode(text, "inbox", attributes)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("server error: %s", response.Message)
	}

	log.Printf("Successfully sent add_node command: %s", text)
	if len(attributes) > 0 {
		log.Printf("Attributes: %v", attributes)
	}
	return nil
}
