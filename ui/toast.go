package ui

import (
	"fmt"
	"sync"
	"time"
)

// Toast represents a single toast notification
type Toast struct {
	Message  string
	Color    string
	Duration time.Duration
	Expiry   time.Time
}

// ToastManager manages toast notifications with a queue
type ToastManager struct {
	mu      sync.Mutex
	current *Toast
	queue   []Toast
}

// NewToastManager creates a new toast manager
func NewToastManager() *ToastManager {
	return &ToastManager{}
}

// Show queues a toast notification (thread-safe)
func (tm *ToastManager) Show(message, color string, duration time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	toast := Toast{
		Message:  message,
		Color:    color,
		Duration: duration,
		Expiry:   time.Now().Add(duration),
	}
	if tm.current == nil || time.Now().After(tm.current.Expiry) {
		tm.current = &toast
	} else {
		tm.queue = append(tm.queue, toast)
	}
}

// GetCurrent returns the current toast text or empty string if none active
func (tm *ToastManager) GetCurrent() string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()

	// Check if current is expired, promote from queue
	for tm.current != nil && now.After(tm.current.Expiry) {
		if len(tm.queue) > 0 {
			next := tm.queue[0]
			next.Expiry = now.Add(next.Duration)
			tm.current = &next
			tm.queue = tm.queue[1:]
		} else {
			tm.current = nil
		}
	}

	if tm.current == nil {
		return ""
	}
	return fmt.Sprintf("[%s]%s[-]", tm.current.Color, tm.current.Message)
}
