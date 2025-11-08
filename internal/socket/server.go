package socket

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Server represents a Unix socket server for accepting external commands
type Server struct {
	socketPath string
	listener   net.Listener
	msgChan    chan Message
	stopChan   chan struct{}
}

// NewServer creates a new Unix socket server
func NewServer(pid int) (*Server, error) {
	// Use XDG_RUNTIME_DIR if available, otherwise fall back to ~/.local/share
	var socketDir string
	if xdgRuntime := os.Getenv("XDG_RUNTIME_DIR"); xdgRuntime != "" {
		// Use XDG_RUNTIME_DIR/tui-outliner for sockets
		socketDir = filepath.Join(xdgRuntime, "tui-outliner")
	} else {
		// Fallback to ~/.local/share/tui-outliner
		socketDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "tui-outliner")
	}

	// Create socket directory if it doesn't exist
	if err := os.MkdirAll(socketDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	socketPath := filepath.Join(socketDir, fmt.Sprintf("tuo-%d.sock", pid))

	// Remove existing socket if it exists
	if err := os.RemoveAll(socketPath); err != nil {
		return nil, fmt.Errorf("failed to remove existing socket: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on socket: %w", err)
	}

	log.Printf("Socket server listening on: %s", socketPath)

	return &Server{
		socketPath: socketPath,
		listener:   listener,
		msgChan:    make(chan Message, 10), // Buffer up to 10 messages
		stopChan:   make(chan struct{}),
	}, nil
}

// Start begins accepting connections on the socket
func (s *Server) Start() {
	go s.acceptLoop()
}

// acceptLoop continuously accepts new connections
func (s *Server) acceptLoop() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				// Check if we're shutting down
				select {
				case <-s.stopChan:
					return
				default:
					log.Printf("Error accepting connection: %v", err)
					continue
				}
			}
			go s.handleConnection(conn)
		}
	}
}

// handleConnection processes a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		if err != io.EOF {
			log.Printf("Error decoding message: %v", err)
		}
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Invalid message format: %v", err),
		}
		encoder.Encode(response)
		return
	}

	// Validate command
	if msg.Command == "" {
		response := Response{
			Success: false,
			Message: "Missing command field",
		}
		encoder.Encode(response)
		return
	}

	// For synchronous commands (like search), create a response channel
	if msg.Command == CommandSearch {
		msg.ResponseChan = make(chan *Response, 1)
	}

	// Send message to channel for processing
	select {
	case s.msgChan <- msg:
		// For synchronous commands, wait for response
		if msg.ResponseChan != nil {
			select {
			case response := <-msg.ResponseChan:
				encoder.Encode(response)
			case <-time.After(10 * time.Second):
				response := Response{
					Success: false,
					Message: "Command timed out",
				}
				encoder.Encode(response)
			}
		} else {
			// For async commands, acknowledge immediately
			response := Response{
				Success: true,
				Message: "Command queued",
			}
			encoder.Encode(response)
		}
	case <-s.stopChan:
		response := Response{
			Success: false,
			Message: "Server is shutting down",
		}
		encoder.Encode(response)
	}
}

// Messages returns the channel for receiving messages
func (s *Server) Messages() <-chan Message {
	return s.msgChan
}

// SocketPath returns the path to the Unix socket
func (s *Server) SocketPath() string {
	return s.socketPath
}

// Stop stops the server and cleans up resources
func (s *Server) Stop() {
	close(s.stopChan)
	if s.listener != nil {
		s.listener.Close()
	}
	// Clean up socket file
	if s.socketPath != "" {
		os.Remove(s.socketPath)
	}
	log.Printf("Socket server stopped")
}
