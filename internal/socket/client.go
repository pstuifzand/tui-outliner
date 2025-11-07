package socket

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Client represents a Unix socket client for sending commands
type Client struct {
	socketPath string
}

// FindRunningInstance finds the socket path for a running tuo instance
// Returns the socket path and PID, or an error if not found
func FindRunningInstance() (string, int, error) {
	socketDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "tui-outliner")

	// Look for socket files
	var sockets []string
	err := filepath.WalkDir(socketDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Ignore errors, directory might not exist
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "tuo-") && strings.HasSuffix(d.Name(), ".sock") {
			sockets = append(sockets, path)
		}
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return "", 0, fmt.Errorf("error scanning socket directory: %w", err)
	}

	if len(sockets) == 0 {
		return "", 0, fmt.Errorf("no running tuo instance found")
	}

	// If multiple sockets, use the most recent one
	if len(sockets) > 1 {
		var newestSocket string
		var newestTime time.Time
		for _, sock := range sockets {
			info, err := os.Stat(sock)
			if err != nil {
				continue
			}
			if info.ModTime().After(newestTime) {
				newestTime = info.ModTime()
				newestSocket = sock
			}
		}
		if newestSocket == "" {
			return "", 0, fmt.Errorf("no accessible socket found")
		}
		sockets = []string{newestSocket}
	}

	socketPath := sockets[0]

	// Extract PID from filename
	filename := filepath.Base(socketPath)
	pidStr := strings.TrimPrefix(filename, "tuo-")
	pidStr = strings.TrimSuffix(pidStr, ".sock")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		pid = 0 // Unknown PID
	}

	return socketPath, pid, nil
}

// NewClient creates a new client connected to the specified socket
func NewClient(socketPath string) (*Client, error) {
	// Verify socket exists
	if _, err := os.Stat(socketPath); err != nil {
		return nil, fmt.Errorf("socket not found: %w", err)
	}

	return &Client{
		socketPath: socketPath,
	}, nil
}

// Send sends a message to the server and returns the response
func (c *Client) Send(msg Message) (*Response, error) {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket: %w", err)
	}
	defer conn.Close()

	// Set a timeout for the operation
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Send message
	if err := encoder.Encode(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Receive response
	var response Response
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &response, nil
}

// SendAddNode is a convenience method to send an add_node command
func (c *Client) SendAddNode(text, target string, attributes map[string]string) (*Response, error) {
	if target == "" {
		target = "inbox"
	}

	msg := Message{
		Command:    CommandAddNode,
		Text:       text,
		Target:     target,
		Attributes: attributes,
	}

	return c.Send(msg)
}
