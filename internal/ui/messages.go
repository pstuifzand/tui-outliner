package ui

import (
	"sync"
	"time"
)

// Message represents a status message with timestamp
type Message struct {
	Text      string
	Timestamp time.Time
}

// MessageLogger tracks the last N status messages
type MessageLogger struct {
	messages []*Message
	maxSize  int
	mu       sync.Mutex
}

// NewMessageLogger creates a new message logger with the specified max size
func NewMessageLogger(maxSize int) *MessageLogger {
	return &MessageLogger{
		messages: make([]*Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

// AddMessage adds a new status message to the history
func (ml *MessageLogger) AddMessage(text string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if text == "" {
		return // Don't log empty messages
	}

	// Add new message at the end
	ml.messages = append(ml.messages, &Message{
		Text:      text,
		Timestamp: time.Now(),
	})

	// Keep only the last maxSize messages
	if len(ml.messages) > ml.maxSize {
		ml.messages = ml.messages[len(ml.messages)-ml.maxSize:]
	}
}

// GetMessages returns a copy of all messages in chronological order
func (ml *MessageLogger) GetMessages() []*Message {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	result := make([]*Message, len(ml.messages))
	copy(result, ml.messages)
	return result
}

// GetMessagesReverse returns a copy of all messages in reverse chronological order (newest first)
func (ml *MessageLogger) GetMessagesReverse() []*Message {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	result := make([]*Message, len(ml.messages))
	for i, msg := range ml.messages {
		result[len(ml.messages)-1-i] = msg
	}
	return result
}

// Clear clears all messages
func (ml *MessageLogger) Clear() {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.messages = ml.messages[:0]
}

// Count returns the number of messages in the logger
func (ml *MessageLogger) Count() int {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return len(ml.messages)
}
