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

	debug := flag.Bool("debug", false, "Enable debug mode (shows key events in status)")
	addNode := flag.String("add", "", "Add a node to the inbox of a running tuo instance")
	flag.Parse()

	// Handle add node command
	if *addNode != "" {
		if err := sendAddNode(*addNode); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Node added to inbox")
		return
	}

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

// sendAddNode sends an add_node command to a running tuo instance
func sendAddNode(text string) error {
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
	response, err := client.SendAddNode(text, "inbox")
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("server error: %s", response.Message)
	}

	log.Printf("Successfully sent add_node command: %s", text)
	return nil
}
