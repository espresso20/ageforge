package game

import "sync"

// Event types
const (
	EventBuildingBuilt      = "building_built"
	EventAgeAdvanced        = "age_advanced"
	EventResearchDone       = "research_done"
	EventGameLoaded         = "game_loaded"
	EventMilestoneCompleted = "milestone_completed"
	EventChainCompleted     = "chain_completed"
	EventEpochAdvanced      = "epoch_advanced"
	EventEpochEventFired    = "epoch_event_fired"
	EventHarbingerArrived   = "harbinger_arrived"
	// EventRunEnded carries the record of a run that just ended in a
	// prestige or a fall (run_ending.go), as Payload["ending"], a RunEnding.
	EventRunEnded = "run_ended"
)

// EventData carries data for an event
type EventData struct {
	Type    string
	Payload map[string]interface{}
}

// EventBus provides pub/sub communication between game systems
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]func(EventData)
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(EventData)),
	}
}

// Subscribe registers a handler for an event type
func (eb *EventBus) Subscribe(eventType string, handler func(EventData)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

// Publish sends an event to all subscribers
func (eb *EventBus) Publish(event EventData) {
	eb.mu.RLock()
	handlers := eb.subscribers[event.Type]
	eb.mu.RUnlock()
	for _, h := range handlers {
		h(event)
	}
}
