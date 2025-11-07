package socket

// Message represents a command sent to the running tuo instance
type Message struct {
	Command string `json:"command"`
	Text    string `json:"text,omitempty"`
	Target  string `json:"target,omitempty"` // Default: "inbox"
}

// Response represents the response from the server
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Command types
const (
	CommandAddNode = "add_node"
)
