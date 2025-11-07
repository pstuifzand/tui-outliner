package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pstuifzand/tui-outliner/internal/app"
	"github.com/pstuifzand/tui-outliner/internal/socket"
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
		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	// Parse flags for main app
	debug := flag.Bool("debug", false, "Enable debug mode (shows key events in status)")
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
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addCmd.Var(&attrs, "attr", "Set an attribute (key=value, can be used multiple times)")
	addCmd.Var(&attrs, "a", "Set an attribute (key=value, shorthand)")
	addCmd.BoolVar(&todoFlag, "t", false, "Add as a todo item (sets type=todo)")
	addCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tuo add [options] <text>\n")
		fmt.Fprintf(os.Stderr, "Add a node to the inbox of a running tuo instance\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -a, --attr key=value    Set an attribute (can be used multiple times)\n")
		fmt.Fprintf(os.Stderr, "  -t                      Add as todo item (sets type=todo)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  tuo add \"Buy milk\"\n")
		fmt.Fprintf(os.Stderr, "  tuo add -t \"Call dentist\"                     # Add as todo\n")
		fmt.Fprintf(os.Stderr, "  tuo add -t -a status=done \"Completed task\"    # Add as done todo\n")
		fmt.Fprintf(os.Stderr, "  tuo add -a priority=high \"Important task\"\n")
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

	if err := sendAddNode(text, attributes); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Node added to inbox")
}

// printUsage prints the main usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "tuo - TUI Outliner\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  tuo [options] [file]           Start tuo with optional file\n")
	fmt.Fprintf(os.Stderr, "  tuo add <text>                 Add node to running instance\n")
	fmt.Fprintf(os.Stderr, "  tuo help                       Show this help message\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  --debug                        Enable debug mode\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  tuo                            Start with empty outline\n")
	fmt.Fprintf(os.Stderr, "  tuo notes.json                 Open notes.json\n")
	fmt.Fprintf(os.Stderr, "  tuo --debug test.json          Open test.json in debug mode\n")
	fmt.Fprintf(os.Stderr, "  tuo add \"Buy milk\"             Add item to running instance\n")
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
