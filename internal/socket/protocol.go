package socket

// Message represents a command sent to the running tuo instance
type Message struct {
	Command    string            `json:"command"`
	Text       string            `json:"text,omitempty"`
	Target     string            `json:"target,omitempty"`     // Default: "inbox"
	Attributes map[string]string `json:"attributes,omitempty"` // Attributes to set on the new item
	ExportPath string            `json:"export_path,omitempty"` // Path for export commands
}

// Response represents the response from the server
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Command types
const (
	CommandAddNode        = "add_node"
	CommandExportMarkdown = "export_markdown"
)
