package executor

import (
	"sync"
	"time"
)

// ActionEvent represents a single event in a command execution lifecycle.
type ActionEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`    // action_start, action_output, action_error, action_end
	Action    string    `json:"action"`  // Human-readable label, e.g., "Deploy go-api"
	Command   string    `json:"command"` // The actual command being run
	Output    string    `json:"output,omitempty"`
	Stream    string    `json:"stream,omitempty"` // stdout or stderr
	ExitCode  *int      `json:"exitCode,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Broadcaster fans out ActionEvents to all registered WebSocket listeners.
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[chan ActionEvent]struct{}
}

// NewBroadcaster creates a new Broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan ActionEvent]struct{}),
	}
}

// Subscribe returns a channel that receives all future ActionEvents.
func (b *Broadcaster) Subscribe() chan ActionEvent {
	ch := make(chan ActionEvent, 64)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a listener channel.
func (b *Broadcaster) Unsubscribe(ch chan ActionEvent) {
	b.mu.Lock()
	delete(b.clients, ch)
	close(ch)
	b.mu.Unlock()
}

// Send broadcasts an event to all subscribers.
func (b *Broadcaster) Send(event ActionEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- event:
		default:
			// Drop event if client is slow
		}
	}
}
