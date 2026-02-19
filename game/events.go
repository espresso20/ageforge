package game

import (
	"math/rand"

	"github.com/user/ageforge/config"
)

// ActiveEvent represents a currently active timed event
type ActiveEvent struct {
	Key       string
	Name      string
	TicksLeft int
	Effects   []config.Effect
}

// EventManager handles random event triggering and processing
type EventManager struct {
	defs      []config.EventDef
	defMap    map[string]config.EventDef
	lastFired map[string]int // key -> last tick fired
	active    []ActiveEvent
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		defs:      config.RandomEvents(),
		defMap:    config.EventByKey(),
		lastFired: make(map[string]int),
	}
}

// Tick processes one tick: checks for new events, processes active event durations.
// Returns list of newly triggered events and list of expired events.
func (em *EventManager) Tick(tick int, currentAge string, ageOrder map[string]int) (triggered []config.EventDef, expired []string) {
	// Process active events first - decrement durations
	var stillActive []ActiveEvent
	for _, ae := range em.active {
		ae.TicksLeft--
		if ae.TicksLeft <= 0 {
			expired = append(expired, ae.Key)
		} else {
			stillActive = append(stillActive, ae)
		}
	}
	em.active = stillActive

	// Check for new random events (one per tick max)
	eligible := em.getEligible(tick, currentAge, ageOrder)
	if len(eligible) == 0 {
		return
	}

	// Weighted random selection
	totalWeight := 0
	for _, def := range eligible {
		totalWeight += def.Weight
	}
	if totalWeight == 0 {
		return
	}

	roll := rand.Intn(totalWeight)
	cumulative := 0
	for _, def := range eligible {
		cumulative += def.Weight
		if roll < cumulative {
			// Only trigger with ~8% chance per tick to avoid spam
			if rand.Intn(100) < 8 {
				em.lastFired[def.Key] = tick
				triggered = append(triggered, def)

				// If duration > 0, add to active
				if def.Duration > 0 {
					em.active = append(em.active, ActiveEvent{
						Key:       def.Key,
						Name:      def.Name,
						TicksLeft: def.Duration,
						Effects:   def.Effects,
					})
				}
			}
			break
		}
	}

	return
}

// getEligible returns events that can trigger right now
func (em *EventManager) getEligible(tick int, currentAge string, ageOrder map[string]int) []config.EventDef {
	var eligible []config.EventDef
	for _, def := range em.defs {
		// Check min tick
		if tick < def.MinTick {
			continue
		}
		// Check age requirement
		if ageOrder[def.MinAge] > ageOrder[currentAge] {
			continue
		}
		// Check cooldown
		if lastTick, ok := em.lastFired[def.Key]; ok {
			if tick-lastTick < def.Cooldown {
				continue
			}
		}
		// Check not already active
		alreadyActive := false
		for _, ae := range em.active {
			if ae.Key == def.Key {
				alreadyActive = true
				break
			}
		}
		if alreadyActive {
			continue
		}
		eligible = append(eligible, def)
	}
	return eligible
}

// GetActiveEffects returns all effects from currently active timed events
func (em *EventManager) GetActiveEffects() []config.Effect {
	var effects []config.Effect
	for _, ae := range em.active {
		effects = append(effects, ae.Effects...)
	}
	return effects
}

// GetActive returns active events for UI display
func (em *EventManager) GetActive() []ActiveEventState {
	var out []ActiveEventState
	for _, ae := range em.active {
		out = append(out, ActiveEventState{
			Name:      ae.Name,
			Key:       ae.Key,
			TicksLeft: ae.TicksLeft,
		})
	}
	return out
}

// LoadState restores event manager state from save
func (em *EventManager) LoadState(lastFired map[string]int, active []ActiveEvent) {
	if lastFired != nil {
		em.lastFired = lastFired
	}
	em.active = active
}

// GetLastFired returns the last-fired map for saving
func (em *EventManager) GetLastFired() map[string]int {
	out := make(map[string]int)
	for k, v := range em.lastFired {
		out[k] = v
	}
	return out
}

// GetActiveForSave returns active events for saving
func (em *EventManager) GetActiveForSave() []ActiveEvent {
	out := make([]ActiveEvent, len(em.active))
	copy(out, em.active)
	return out
}
