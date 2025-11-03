package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pstuifzand/tui-outliner/internal/model"
	"github.com/pstuifzand/tui-outliner/internal/storage"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: tuo-to-diffformat [options] <input.json> [output.txt]

Converts a TUI Outliner JSON file to the diff-optimized format.

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Arguments:
  input.json   Path to the outline JSON file to convert
  output.txt   Path to write the diff-formatted output (optional)
               If not provided, writes to stdout

Examples:
  # Convert and print to stdout
  tuo-to-diffformat my_outline.json

  # Convert and save to file
  tuo-to-diffformat my_outline.json my_outline.txt

  # Convert all JSON files in a directory
  for f in *.json; do tuo-to-diffformat "$f" "${f%.json}.txt"; done
`)
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := args[0]
	outputPath := ""
	if len(args) > 1 {
		outputPath = args[1]
	}

	// Read the input JSON file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Parse the JSON into an Outline
	var outline model.Outline
	if err := json.Unmarshal(data, &outline); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Open output file
	var outFile *os.File
	if outputPath == "" {
		outFile = os.Stdout
	} else {
		// Create output directory if needed
		outDir := filepath.Dir(outputPath)
		if outDir != "." && outDir != "" {
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
				os.Exit(1)
			}
		}

		file, err := os.Create(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		outFile = file
	}

	// Convert to diff format
	if err := storage.EncodeDiffFormat(&outline, outFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding to diff format: %v\n", err)
		os.Exit(1)
	}

	if outputPath != "" {
		fmt.Fprintf(os.Stderr, "Successfully converted: %s → %s\n", inputPath, outputPath)
	}
}
