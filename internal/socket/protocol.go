package socket

// Message represents a command sent to the running tuo instance
type Message struct {
	Command    string            `json:"command"`
	Text       string            `json:"text,omitempty"`
	Target     string            `json:"target,omitempty"`     // Default: "inbox"
	Attributes map[string]string `json:"attributes,omitempty"` // Attributes to set on the new item
	ExportPath string            `json:"export_path,omitempty"` // Path for export commands
	Query      string            `json:"query,omitempty"`      // Search query
	Fields     []string          `json:"fields,omitempty"`     // Fields to include in search results
	Format     string            `json:"format,omitempty"`     // Output format for search results

	// Internal field for synchronous responses (not sent over the wire)
	ResponseChan chan *Response `json:"-"`
}

// SearchResult represents a single search result item with flexible fields
// The actual fields included depend on the Fields parameter in the search request
type SearchResult map[string]interface{}

// Response represents the response from the server
type Response struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Results []SearchResult `json:"results,omitempty"` // For search commands
}

// Command types
const (
	CommandAddNode        = "add_node"
	CommandExportMarkdown = "export_markdown"
	CommandSearch         = "search"
)
