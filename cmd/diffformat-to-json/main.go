package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pstuifzand/tui-outliner/internal/storage"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: diffformat-to-json [options] <input.txt> [output.json]

Converts a diff-formatted outline file back to JSON.

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Arguments:
  input.txt    Path to the diff-formatted file to convert
  output.json  Path to write the JSON output (optional)
               If not provided, writes to stdout

Examples:
  # Convert and print to stdout
  diffformat-to-json my_outline.txt

  # Convert and save to file
  diffformat-to-json my_outline.txt my_outline.json

  # Round-trip test: JSON → diff → JSON
  tuo-to-diffformat original.json temp.txt
  diffformat-to-json temp.txt restored.json
  diff <(jq -S . original.json) <(jq -S . restored.json)
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

	// Read and parse the diff format file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	outline, err := storage.DecodeDiffFormat(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing diff format: %v\n", err)
		os.Exit(1)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(outline, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to JSON: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if outputPath == "" {
		fmt.Println(string(data))
	} else {
		// Create output directory if needed
		outDir := filepath.Dir(outputPath)
		if outDir != "." && outDir != "" {
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
				os.Exit(1)
			}
		}

		if err := os.WriteFile(outputPath, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Successfully converted: %s → %s\n", inputPath, outputPath)
	}
}
