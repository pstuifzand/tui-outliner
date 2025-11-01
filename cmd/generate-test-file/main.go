package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pstuifzand/tui-outliner/internal/model"
)

func main() {
	numNodes := flag.Int("nodes", 1000, "Number of nodes to generate")
	output := flag.String("output", "large_test.json", "Output file path")
	depth := flag.Int("depth", 3, "Maximum nesting depth")
	flag.Parse()

	if *numNodes < 1 {
		fmt.Fprintf(os.Stderr, "nodes must be at least 1\n")
		os.Exit(1)
	}

	outline := generateOutline(*numNodes, *depth)

	// Marshal to JSON with nice formatting
	data, err := json.MarshalIndent(outline, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	// Ensure directory exists
	dir := filepath.Dir(*output)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Write to file
	err = os.WriteFile(*output, data, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write file: %v\n", err)
		os.Exit(1)
	}

	countNodes := countAllNodes(&outline)
	fmt.Printf("Generated outline with %d nodes\n", countNodes)
	fmt.Printf("Saved to: %s\n", *output)
	fmt.Printf("File size: %.2f MB\n", float64(len(data))/(1024*1024))
}

func generateOutline(totalNodes int, maxDepth int) model.Outline {
	outline := model.Outline{
		Items: []*model.Item{},
	}

	remaining := totalNodes
	depth := 0

	// Create a balanced tree structure
	for remaining > 0 {
		item := generateItemRecursive(&remaining, depth, maxDepth)
		if item.Text != "" {
			outline.Items = append(outline.Items, item)
		}
	}

	return outline
}

func generateItemRecursive(remaining *int, currentDepth int, maxDepth int) *model.Item {
	if *remaining <= 0 {
		return nil
	}

	item := model.NewItem(generateUniqueText(*remaining))
	*remaining--

	// Add children if we haven't reached max depth and still have nodes left
	if currentDepth < maxDepth && *remaining > 0 {
		numChildren := getChildCount(*remaining, maxDepth-currentDepth)
		item.Children = make([]*model.Item, 0, numChildren)

		for i := 0; i < numChildren && *remaining > 0; i++ {
			child := generateItemRecursive(remaining, currentDepth+1, maxDepth)
			if child != nil {
				item.Children = append(item.Children, child)
			}
		}
	}

	return item
}

func getChildCount(remaining int, depthLeft int) int {
	// Distribute nodes across children based on remaining nodes
	if depthLeft == 1 {
		// Leaf level: create fewer children
		if remaining > 10 {
			return 5
		}
		return remaining / 2
	}
	// Internal levels: create 2-3 children
	if remaining > 50 {
		return 3
	}
	return 2
}

func generateUniqueText(index int) string {
	// Generate unique, descriptive text for each node
	categories := []string{
		"Task", "Note", "Idea", "Bug", "Feature", "Enhancement",
		"Documentation", "Refactor", "Test", "Optimization",
		"Research", "Design", "Implementation", "Review",
	}

	category := categories[index%len(categories)]
	return fmt.Sprintf("%s #%d - %s", category, index,
		generateDescription(index))
}

func generateDescription(index int) string {
	descriptions := []string{
		"Core functionality",
		"User interface",
		"Performance improvement",
		"Bug fix",
		"New capability",
		"API integration",
		"Data validation",
		"Error handling",
		"Caching layer",
		"Database schema",
		"Authentication",
		"Configuration",
		"Logging system",
		"Monitoring",
		"Security audit",
	}

	return descriptions[index%len(descriptions)]
}

func countAllNodes(outline *model.Outline) int {
	count := 0
	for i := range outline.Items {
		count += countItemNodes(outline.Items[i])
	}
	return count
}

func countItemNodes(item *model.Item) int {
	count := 1 // Count this item
	for i := range item.Children {
		count += countItemNodes(item.Children[i])
	}
	return count
}
