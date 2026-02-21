package game

import "sync"

// Event types
const (
	EventBuildingBuilt  = "building_built"
	EventVillagerAdded  = "villager_added"
	EventAgeAdvanced    = "age_advanced"
	EventResourceDepleted = "resource_depleted"
	EventResearchDone   = "research_done"
	EventGameSaved           = "game_saved"
	EventGameLoaded          = "game_loaded"
	EventMilestoneCompleted  = "milestone_completed"
	EventChainCompleted      = "chain_completed"
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
