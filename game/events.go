package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
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
// minutes of real time in the Primitive and Stone Ages, preventing event spam.
// From the Bronze Age on the ages run config.PacingStretch times longer, and
// so do the delay, each event's duration and its cooldown (the ruleset's
// StretchTicks on the current age), so an age holds as many events as before.
//
// NOTE: InjectEvent bypasses all eligibility checks and fires immediately.
// It is used for milestone chain boosts and epoch event effects; calling it
// inside a Bus handler is safe because the EventManager is not locked separately.
type EventManager struct {
	rules         *rules.Set
	defs          []config.EventDef
	defMap        map[string]config.EventDef
	lastFired     map[string]int // event key -> last tick fired (for per-event cooldowns)
	active        []ActiveEvent
	nextEventTick int // global cooldown: earliest tick the next random event can fire
	goodStreak    int // consecutive good events (reset on bad/mixed)
	badStreak     int // consecutive bad events (reset on good/mixed)
}

// The delay between random events before the age's stretch
// (eventDelay applies it).
const (
	eventMinDelay = 150 // 5 minutes (150 ticks * 2s)
	eventMaxDelay = 600 // 20 minutes (600 ticks * 2s)
)

// eventDelayIn draws the wait until the next random event in age on set's
// clocks: 150-600 ticks, times the age's stretch. One draw from rng
// whatever the age, so the stream keeps its shape.
func eventDelayIn(set *rules.Set, rng *rand.Rand, age string) int {
	return set.StretchTicks(age, eventMinDelay+rng.Intn(eventMaxDelay-eventMinDelay+1))
}

// eventUnscheduled marks an EventManager whose first random event has not been
// scheduled yet. The first delay is drawn on the first Tick, off the rng that
// Tick is handed, rather than at construction: the engine builds its managers
// before it knows the run's seed, so a draw here would come from the wrong
// stream.
const eventUnscheduled = -1

// NewEventManager creates a new event manager on the core ruleset.
func NewEventManager() *EventManager { return NewEventManagerWith(rules.Core()) }

// NewEventManagerWith creates a new event manager with set's events. It makes
// no random draws; the first event is scheduled on the first Tick.
func NewEventManagerWith(set *rules.Set) *EventManager {
	em := &EventManager{
		lastFired:     make(map[string]int),
		nextEventTick: eventUnscheduled,
	}
	em.Rebind(set)
	return em
}

// Rebind moves the manager onto set: it takes set's events. Active events,
// cooldowns and the next event's tick stay.
func (em *EventManager) Rebind(set *rules.Set) {
	em.rules = set
	em.defs = set.Events()
	em.defMap = set.EventMap()
}

// Tick processes one tick: checks for new events, processes active event durations.
// Returns list of newly triggered events and list of expired ActiveEvents (with accumulated losses).
// Every random draw comes from rng, in a fixed order, so a seeded rng gives a
// reproducible event stream.
func (em *EventManager) Tick(rng *rand.Rand, tick int, currentAge string, ageOrder map[string]int, currentEpoch string) (triggered []config.EventDef, expired []ActiveEvent) {
	if em.nextEventTick == eventUnscheduled {
		// First event 150-600 ticks from the start of the run.
		em.nextEventTick = eventDelayIn(em.rules, rng, currentAge)
	}

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
	forceSentiment := em.requiredSentiment(rng)

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

	roll := rng.Intn(totalWeight)
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
					TicksLeft: em.rules.StretchTicks(currentAge, def.Duration),
					Effects:   def.Effects,
				})
			}

			// Schedule the next event (5-20 minutes from now, times the
			// age's stretch)
			em.nextEventTick = tick + eventDelayIn(em.rules, rng, currentAge)
			break
		}
	}

	return
}

// requiredSentiment returns a sentiment filter based on current streaks.
// "" means no constraint, "good" means only good/mixed, "bad" means only bad/mixed.
func (em *EventManager) requiredSentiment(rng *rand.Rand) string {
	// Hard rule: never more than 2 bad in a row → force good
	if em.badStreak >= 2 {
		return "good"
	}
	// After 3 good in a row, force bad (with a tiny 3% chance to reset and allow more good)
	if em.goodStreak >= 3 {
		if rng.Intn(100) < 3 {
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
	for _, ev := range em.rules.EraEvents() {
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
		// Check cooldown (stretched like the age)
		if lastTick, ok := em.lastFired[def.Key]; ok {
			if tick-lastTick < em.rules.StretchTicks(currentAge, def.Cooldown) {
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
// these by effect Type — it accumulates eff.Value into the "production_all",
// "tick_speed", and per-resource "<res>_rate" additive pools keyed by eff.Type
// (not eff.Target) — so the Modifier Target is the effect Type. The "<res>_rate"
// case is what carries a faction-encounter specialty boon (Type "iron_rate", ...)
// into the resolver's per-resource multiplier pool in recalculateRates. Per-resource
// "production" effects are flat additions handled elsewhere and are not multiplier
// modifiers.
func (em *EventManager) Modifiers() []Modifier {
	var out []Modifier
	for _, ae := range em.active {
		src := "event:" + ae.Name
		if ae.Name == "" {
			src = "event"
		}
		for _, eff := range ae.Effects {
			switch {
			case eff.Type == "production_all", eff.Type == "tick_speed",
				strings.HasSuffix(eff.Type, "_rate"):
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
		//
		// The admitted set MUST track Modifiers() above — anything the engine
		// keeps applying every tick is something the panel has to be able to
		// show. The "<res>_rate" suffix case is load-bearing: every faction
		// specialty boon and most setbacks land as a RateBuff, which
		// boon/apply.go maps to Type "<res>_rate". Matching those types
		// exactly (as this switch once did) silently dropped every one of
		// them, so the panel rendered a named event with no magnitude.
		var effects []EventEffectInfo
		for _, eff := range ae.Effects {
			switch {
			case eff.Type == "production", eff.Type == "production_all",
				eff.Type == "tick_speed", strings.HasSuffix(eff.Type, "_rate"):
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

// buildLossSuffix returns a loss summary for an expired event, or "" if none
// was recorded. Losses are logged when an event strikes now (see
// applyEventEffects), so only timed events loaded from older saves still
// carry recorded losses to report here.
func buildLossSuffix(event ActiveEvent) string {
	var parts []string
	if event.WorkersLost > 0 {
		parts = append(parts, fmt.Sprintf("[yellow]%d workers lost[-]", event.WorkersLost))
	}
	if len(event.ResourcesLost) > 0 {
		keys := make([]string, 0, len(event.ResourcesLost))
		for k := range event.ResourcesLost {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			amt := event.ResourcesLost[k]
			parts = append(parts, fmt.Sprintf("[yellow]%s %s lost[-]", amountText(amt), resourceLabel(k)))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, ", ") + "."
}

// applyEventEffects applies a triggered random event's instant and on-trigger
// effects, then logs what the player actually lost (amounts and workers, at
// once, not when a timed event ends). A Raid event meets the garrison first,
// and a second line says what it kept. Under the write lock.
func (ge *GameEngine) applyEventEffects(def config.EventDef) {
	// A raid meets the garrison: it blunts a share of what the raiders
	// take (config/defense.go). 0 with no soldiers, and every loss below
	// is then exactly what it always was.
	guard := 0.0
	if def.Raid {
		guard = ge.raidMitigation()
	}
	var lostRes, keptRes map[string]float64
	var granted, fit map[string]float64 // what the event's text promised, and what storage took
	lostWorkers, keptWorkers := 0, 0
	for _, eff := range def.Effects {
		switch eff.Type {
		case "instant_resource":
			if granted == nil {
				granted, fit = make(map[string]float64), make(map[string]float64)
			}
			granted[eff.Target] += eff.Value
			fit[eff.Target] += ge.grantLocked(eff.Target, eff.Value)
			ge.addLog("debug", fmt.Sprintf("Event effect: %s %s %+.1f", eff.Type, eff.Target, eff.Value))
		case "steal_resource":
			current := ge.Resources.Get(eff.Target)
			loss := eff.Value
			if loss > current {
				loss = current
			}
			if guard > 0 && loss > 0 {
				kept := float64(loss * guard)
				loss -= kept
				if keptRes == nil {
					keptRes = make(map[string]float64)
				}
				keptRes[eff.Target] += kept
			}
			if loss > 0 && ge.Resources.Remove(eff.Target, loss) {
				if lostRes == nil {
					lostRes = make(map[string]float64)
				}
				lostRes[eff.Target] += loss
			}
			ge.addLog("debug", fmt.Sprintf("Event effect: %s %s -%.1f", eff.Type, eff.Target, loss))
		case "worker_loss":
			// Value is a share (0.0-1.0) of the worker pool to remove.
			pct := eff.Value
			if guard > 0 {
				pct = float64(eff.Value * (1 - guard))
				full := int(float64(ge.Workers.TotalPop()) * eff.Value)
				if blunted := int(float64(ge.Workers.TotalPop()) * pct); blunted < full {
					keptWorkers += full - blunted
				}
			}
			before := ge.Workers.TotalPop()
			ge.Workers.RemovePct(pct)
			lostWorkers += before - ge.Workers.TotalPop()
			ge.addLog("debug", fmt.Sprintf("Event effect: worker_loss %.0f%%", pct*100))
		}
	}
	// The event's text states the full grant: say so when a full store
	// took less.
	if line := clippedLine(granted, fit); line != "" {
		ge.addLog("info", line)
	}
	if parts := lossParts(lostRes, lostWorkers); len(parts) > 0 {
		ge.addLog("warning", "  You lost "+joinAnd(parts)+".")
	}
	if line := garrisonSavedLine(guard, keptRes, keptWorkers); line != "" {
		ge.addLog("success", line)
		t := ge.defenseTally()
		t.Raids++
		ge.note(config.BadgeEvRaidBlunted, "")
		t.Workers += keptWorkers
		for _, res := range sortedKeys(keptRes) {
			ge.recordSavedResource(res, keptRes[res])
		}
	}
}
