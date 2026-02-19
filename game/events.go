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
	defs          []config.EventDef
	defMap        map[string]config.EventDef
	lastFired     map[string]int // key -> last tick fired
	active        []ActiveEvent
	nextEventTick int // global cooldown: earliest tick the next event can fire
	goodStreak    int // consecutive good events (reset on bad/mixed)
	badStreak     int // consecutive bad events (reset on good/mixed)
}

const (
	eventMinDelay = 150 // 5 minutes (150 ticks * 2s)
	eventMaxDelay = 600 // 20 minutes (600 ticks * 2s)
)

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	// Schedule first event between 150-600 ticks from start
	firstDelay := eventMinDelay + rand.Intn(eventMaxDelay-eventMinDelay+1)
	return &EventManager{
		defs:          config.RandomEvents(),
		defMap:        config.EventByKey(),
		lastFired:     make(map[string]int),
		nextEventTick: firstDelay,
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

	// Only check for new events after the global cooldown expires
	if tick < em.nextEventTick {
		return
	}

	// Determine sentiment constraints based on streaks
	forceSentiment := em.requiredSentiment()

	// Check for new random events (one per tick max)
	eligible := em.getEligible(tick, currentAge, ageOrder, forceSentiment)
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
			em.lastFired[def.Key] = tick
			triggered = append(triggered, def)

			// Update streak tracking
			em.updateStreaks(def.Sentiment)

			// If duration > 0, add to active
			if def.Duration > 0 {
				em.active = append(em.active, ActiveEvent{
					Key:       def.Key,
					Name:      def.Name,
					TicksLeft: def.Duration,
					Effects:   def.Effects,
				})
			}

			// Schedule next event 5-20 minutes from now
			em.nextEventTick = tick + eventMinDelay + rand.Intn(eventMaxDelay-eventMinDelay+1)
			break
		}
	}

	return
}

// requiredSentiment returns a sentiment filter based on current streaks.
// "" means no constraint, "good" means only good/mixed, "bad" means only bad/mixed.
func (em *EventManager) requiredSentiment() string {
	// Hard rule: never more than 2 bad in a row → force good
	if em.badStreak >= 2 {
		return "good"
	}
	// After 3 good in a row, force bad (with a tiny 3% chance to reset and allow more good)
	if em.goodStreak >= 3 {
		if rand.Intn(100) < 3 {
			em.goodStreak = 0 // lucky reset
			return ""
		}
		return "bad"
	}
	return ""
}

// updateStreaks updates the good/bad consecutive counters after an event fires.
func (em *EventManager) updateStreaks(sentiment string) {
	switch sentiment {
	case "good":
		em.goodStreak++
		em.badStreak = 0
	case "bad":
		em.badStreak++
		em.goodStreak = 0
	default: // "mixed" — resets both
		em.goodStreak = 0
		em.badStreak = 0
	}
}

// getEligible returns events that can trigger right now.
// forceSentiment filters: "good" = only good/mixed, "bad" = only bad/mixed, "" = any.
func (em *EventManager) getEligible(tick int, currentAge string, ageOrder map[string]int, forceSentiment string) []config.EventDef {
	var eligible []config.EventDef
	for _, def := range em.defs {
		// Sentiment filter
		if forceSentiment == "good" && def.Sentiment == "bad" {
			continue
		}
		if forceSentiment == "bad" && def.Sentiment == "good" {
			continue
		}
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
func (em *EventManager) LoadState(lastFired map[string]int, active []ActiveEvent, nextEventTick int, goodStreak int, badStreak int) {
	if lastFired != nil {
		em.lastFired = lastFired
	}
	em.active = active
	if nextEventTick > 0 {
		em.nextEventTick = nextEventTick
	}
	em.goodStreak = goodStreak
	em.badStreak = badStreak
}

// GetNextEventTick returns the next event tick for saving
func (em *EventManager) GetNextEventTick() int {
	return em.nextEventTick
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
