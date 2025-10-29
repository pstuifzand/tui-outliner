package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pstuifzand/tui-outliner/internal/app"
)

func main() {
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
