package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// ActiveEvent represents a currently active timed event
type ActiveEvent struct {
	Key           string
	Name          string
	TicksLeft     int
	Effects       []config.Effect
	WorkersLost   int                // accumulated workers lost during this event
	ResourcesLost map[string]float64 // resource key → total amount lost during this event
}

// EventManager handles random event triggering and management of active
// timed events. Events are drawn from two pools: universal (base) events
// and epoch-exclusive events for the player's current epoch.
//
// Anti-streak system: goodStreak and badStreak track consecutive same-sentiment
// events. After 3 good in a row the next event is forced bad; after 2 bad it
// is forced good. This prevents extended lucky or punishing streaks.
//
// A global cooldown (nextEventTick) ensures at most one event fires per 5-20
// minutes of real time, preventing event spam.
//
// NOTE: InjectEvent bypasses all eligibility checks and fires immediately.
// It is used for milestone chain boosts and epoch event effects; calling it
// inside a Bus handler is safe because the EventManager is not locked separately.
type EventManager struct {
	defs          []config.EventDef
	defMap        map[string]config.EventDef
	lastFired     map[string]int // event key -> last tick fired (for per-event cooldowns)
	active        []ActiveEvent
	nextEventTick int // global cooldown: earliest tick the next random event can fire
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
// Returns list of newly triggered events and list of expired ActiveEvents (with accumulated losses).
func (em *EventManager) Tick(tick int, currentAge string, ageOrder map[string]int, currentEpoch string) (triggered []config.EventDef, expired []ActiveEvent) {
	// Process active events first - decrement durations
	var stillActive []ActiveEvent
	for _, ae := range em.active {
		ae.TicksLeft--
		if ae.TicksLeft <= 0 {
			expired = append(expired, ae)
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
	eligible := em.getEligible(tick, currentAge, ageOrder, forceSentiment, currentEpoch)
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
func (em *EventManager) getEligible(tick int, currentAge string, ageOrder map[string]int, forceSentiment string, currentEpoch string) []config.EventDef {
	// Build candidate pool: universal events + epoch-exclusive events for the current epoch
	pool := make([]config.EventDef, len(em.defs))
	copy(pool, em.defs)
	for _, ev := range config.EpochExclusiveEvents() {
		if ev.EpochKey == currentEpoch {
			pool = append(pool, ev)
		}
	}

	var eligible []config.EventDef
	for _, def := range pool {
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

// InjectEvent adds an event directly to the active list, bypassing all
// eligibility, cooldown, and streak checks. Used for milestone chain speed
// boosts and epoch event side-effects that must fire unconditionally.
// IMPORTANT: Must be called under the engine write lock (same as doTick).
func (em *EventManager) InjectEvent(event ActiveEvent) {
	em.active = append(em.active, event)
}

// GetActiveEffects returns all effects from currently active timed events
func (em *EventManager) GetActiveEffects() []config.Effect {
	var effects []config.Effect
	for _, ae := range em.active {
		effects = append(effects, ae.Effects...)
	}
	return effects
}

// Modifiers emits OpAdd Modifiers for the multiplier-bucket effects carried by
// currently active events, attributed to Source "event:<name>". The engine reads
// these by effect Type (it accumulates eff.Value into the "production_all" and
// "tick_speed" buckets based on eff.Type, not eff.Target), so the Modifier Target
// is the effect Type. Only the bucket types that feed recalculateRates /
// recalculateTickSpeed today are emitted; per-resource "production" effects are
// flat additions handled elsewhere and are not multiplier modifiers.
func (em *EventManager) Modifiers() []Modifier {
	var out []Modifier
	for _, ae := range em.active {
		src := "event:" + ae.Name
		if ae.Name == "" {
			src = "event"
		}
		for _, eff := range ae.Effects {
			switch eff.Type {
			case "production_all", "tick_speed":
				out = append(out, Modifier{Source: src, Target: eff.Type, Op: OpAdd, Value: eff.Value})
			}
		}
	}
	return out
}

// GetActive returns active events for UI display
func (em *EventManager) GetActive() []ActiveEventState {
	var out []ActiveEventState
	for _, ae := range em.active {
		// Surface only ongoing-rate effects. Instant/one-shot types
		// ("instant_resource", "steal_resource", "worker_loss") fired once at
		// trigger and are not ongoing, so they don't belong in the panel.
		var effects []EventEffectInfo
		for _, eff := range ae.Effects {
			switch eff.Type {
			case "production", "production_all", "tick_speed":
				effects = append(effects, EventEffectInfo{
					Type:   eff.Type,
					Target: eff.Target,
					Value:  eff.Value,
				})
			}
		}
		out = append(out, ActiveEventState{
			Name:      ae.Name,
			Key:       ae.Key,
			TicksLeft: ae.TicksLeft,
			Effects:   effects,
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

// RecordWorkerLoss accumulates workers lost for the active event matching key.
func (em *EventManager) RecordWorkerLoss(key string, count int) {
	for i := range em.active {
		if em.active[i].Key == key {
			em.active[i].WorkersLost += count
			return
		}
	}
}

// RecordResourceLoss accumulates resource stolen for the active event matching key.
func (em *EventManager) RecordResourceLoss(key string, resource string, amount float64) {
	for i := range em.active {
		if em.active[i].Key == key {
			if em.active[i].ResourcesLost == nil {
				em.active[i].ResourcesLost = make(map[string]float64)
			}
			em.active[i].ResourcesLost[resource] += amount
			return
		}
	}
}

// buildLossSuffix returns a tview-coloured loss summary string for an expired event,
// or "" if no losses were recorded.
func buildLossSuffix(event ActiveEvent) string {
	var parts []string
	if event.WorkersLost > 0 {
		parts = append(parts, fmt.Sprintf("[yellow]%d workers fled[-]", event.WorkersLost))
	}
	if len(event.ResourcesLost) > 0 {
		keys := make([]string, 0, len(event.ResourcesLost))
		for k := range event.ResourcesLost {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			amt := event.ResourcesLost[k]
			parts = append(parts, fmt.Sprintf("[yellow]%.0f %s stolen[-]", amt, k))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, ", ") + "."
}
