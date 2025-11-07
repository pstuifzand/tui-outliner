package socket

import (
	"os"
	"testing"
	"time"
)

func TestServerClient(t *testing.T) {
	// Create a server
	pid := os.Getpid()
	server, err := NewServer(pid)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	// Start the server
	server.Start()

	// Wait a bit for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create a client
	client, err := NewClient(server.SocketPath())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Send a message
	msg := Message{
		Command: CommandAddNode,
		Text:    "Test item",
		Target:  "inbox",
	}

	response, err := client.Send(msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got success=false: %s", response.Message)
	}

	// Receive the message from the server
	select {
	case receivedMsg := <-server.Messages():
		if receivedMsg.Command != msg.Command {
			t.Errorf("Expected command=%s, got command=%s", msg.Command, receivedMsg.Command)
		}
		if receivedMsg.Text != msg.Text {
			t.Errorf("Expected text=%s, got text=%s", msg.Text, receivedMsg.Text)
		}
		if receivedMsg.Target != msg.Target {
			t.Errorf("Expected target=%s, got target=%s", msg.Target, receivedMsg.Target)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestFindRunningInstance(t *testing.T) {
	// Create a server
	pid := os.Getpid()
	server, err := NewServer(pid)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	server.Start()

	// Wait a bit for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Find the running instance
	socketPath, foundPid, err := FindRunningInstance()
	if err != nil {
		t.Fatalf("Failed to find running instance: %v", err)
	}

	if socketPath != server.SocketPath() {
		t.Errorf("Expected socketPath=%s, got socketPath=%s", server.SocketPath(), socketPath)
	}

	if foundPid != pid {
		t.Errorf("Expected pid=%d, got pid=%d", pid, foundPid)
	}
}

func TestSendAddNode(t *testing.T) {
	// Create a server
	pid := os.Getpid()
	server, err := NewServer(pid)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	server.Start()

	// Wait a bit for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create a client
	client, err := NewClient(server.SocketPath())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Use the convenience method
	response, err := client.SendAddNode("Test item", "inbox", nil)
	if err != nil {
		t.Fatalf("Failed to send add_node: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got success=false: %s", response.Message)
	}

	// Receive the message
	select {
	case msg := <-server.Messages():
		if msg.Command != CommandAddNode {
			t.Errorf("Expected command=%s, got command=%s", CommandAddNode, msg.Command)
		}
		if msg.Text != "Test item" {
			t.Errorf("Expected text='Test item', got text='%s'", msg.Text)
		}
		if msg.Target != "inbox" {
			t.Errorf("Expected target='inbox', got target='%s'", msg.Target)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestSendAddNodeWithAttributes(t *testing.T) {
	// Create a server
	pid := os.Getpid()
	server, err := NewServer(pid)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	server.Start()

	// Wait a bit for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Create a client
	client, err := NewClient(server.SocketPath())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create attributes
	attributes := map[string]string{
		"type":     "task",
		"priority": "high",
		"status":   "todo",
	}

	// Use the convenience method with attributes
	response, err := client.SendAddNode("Test item with attributes", "inbox", attributes)
	if err != nil {
		t.Fatalf("Failed to send add_node: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got success=false: %s", response.Message)
	}

	// Receive the message
	select {
	case msg := <-server.Messages():
		if msg.Command != CommandAddNode {
			t.Errorf("Expected command=%s, got command=%s", CommandAddNode, msg.Command)
		}
		if msg.Text != "Test item with attributes" {
			t.Errorf("Expected text='Test item with attributes', got text='%s'", msg.Text)
		}
		if msg.Target != "inbox" {
			t.Errorf("Expected target='inbox', got target='%s'", msg.Target)
		}
		if len(msg.Attributes) != 3 {
			t.Errorf("Expected 3 attributes, got %d", len(msg.Attributes))
		}
		if msg.Attributes["type"] != "task" {
			t.Errorf("Expected type='task', got type='%s'", msg.Attributes["type"])
		}
		if msg.Attributes["priority"] != "high" {
			t.Errorf("Expected priority='high', got priority='%s'", msg.Attributes["priority"])
		}
		if msg.Attributes["status"] != "todo" {
			t.Errorf("Expected status='todo', got status='%s'", msg.Attributes["status"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}
