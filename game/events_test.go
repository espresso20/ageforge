package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

func TestEventManager_InjectEvent(t *testing.T) {
	em := NewEventManager()

	em.InjectEvent(ActiveEvent{
		Key:       "test_boost",
		Name:      "Test Boost",
		TicksLeft: 10,
		Effects: []config.Effect{
			{Type: "tick_speed", Target: "tick_speed", Value: 2.0},
		},
	})

	effects := em.GetActiveEffects()
	found := false
	for _, e := range effects {
		if e.Type == "tick_speed" && e.Value == 2.0 {
			found = true
		}
	}
	if !found {
		t.Error("injected event effect not found in GetActiveEffects")
	}

	active := em.GetActive()
	if len(active) < 1 {
		t.Fatal("expected at least 1 active event")
	}
	foundActive := false
	for _, a := range active {
		if a.Key == "test_boost" && a.TicksLeft == 10 {
			foundActive = true
		}
	}
	if !foundActive {
		t.Error("injected event not found in GetActive")
	}
}

// TestGetActive_SurfacesOngoingEffects proves the view struct now carries each
// active event's ongoing-rate effects (production / production_all / tick_speed)
// and drops instant/one-shot types that already fired at trigger.
func TestGetActive_SurfacesOngoingEffects(t *testing.T) {
	em := NewEventManager()
	em.InjectEvent(ActiveEvent{
		Key:       "famine",
		Name:      "Famine",
		TicksLeft: 8,
		Effects: []config.Effect{
			{Type: "production", Target: "food", Value: -3.0},
			{Type: "instant_resource", Target: "wood", Value: 100.0}, // one-shot, must be dropped
		},
	})

	active := em.GetActive()
	if len(active) != 1 {
		t.Fatalf("expected 1 active event, got %d", len(active))
	}
	got := active[0]

	if len(got.Effects) != 1 {
		t.Fatalf("expected 1 ongoing effect (instant dropped), got %d: %+v", len(got.Effects), got.Effects)
	}
	eff := got.Effects[0]
	if eff.Type != "production" || eff.Target != "food" || eff.Value != -3.0 {
		t.Errorf("ongoing effect = %+v, want {production food -3.0}", eff)
	}
	for _, e := range got.Effects {
		if e.Type == "instant_resource" {
			t.Errorf("instant_resource effect must NOT be surfaced in the view: %+v", e)
		}
	}
}

func TestEventManager_InjectedEventExpires(t *testing.T) {
	em := NewEventManager()
	em.InjectEvent(ActiveEvent{
		Key:       "short_boost",
		Name:      "Short Boost",
		TicksLeft: 2,
		Effects: []config.Effect{
			{Type: "tick_speed", Target: "tick_speed", Value: 1.0},
		},
	})

	ageOrder := map[string]int{"primitive_age": 0}

	// Tick 1: still active
	em.Tick(1, "primitive_age", ageOrder, "stone_era")
	active := em.GetActive()
	found := false
	for _, a := range active {
		if a.Key == "short_boost" {
			found = true
		}
	}
	if !found {
		t.Error("event should still be active after 1 tick")
	}

	// Tick 2: should expire
	_, expired := em.Tick(2, "primitive_age", ageOrder, "stone_era")
	foundExpired := false
	for _, ae := range expired {
		if ae.Key == "short_boost" {
			foundExpired = true
		}
	}
	if !foundExpired {
		t.Error("event should have expired after 2 ticks")
	}
}

func TestEventManager_SaveLoadRoundTrip(t *testing.T) {
	em := NewEventManager()
	em.InjectEvent(ActiveEvent{
		Key:       "save_test",
		Name:      "Save Test",
		TicksLeft: 50,
		Effects:   []config.Effect{{Type: "production", Target: "food", Value: 1.0}},
	})

	// Save
	lastFired := em.GetLastFired()
	activeForSave := em.GetActiveForSave()
	nextTick := em.GetNextEventTick()

	// Load into fresh
	em2 := NewEventManager()
	em2.LoadState(lastFired, activeForSave, nextTick, 0, 0)

	active := em2.GetActive()
	found := false
	for _, a := range active {
		if a.Key == "save_test" {
			found = true
		}
	}
	if !found {
		t.Error("loaded event manager should have save_test active")
	}
}

// TestGetActive_SurfacesRateEffects pins the fix for a silent data loss in the
// UI-facing view: GetActive matched effect types by EXACT name
// ("production" / "production_all" / "tick_speed"), but the boon engine emits
// faction specialty buffs and setbacks as "<res>_rate" (boon/apply.go maps a
// RateBuff on food to Type "food_rate"). Those matched no case, so
// ActiveEventState.Effects came back EMPTY for the most common boon shape in the
// game and the panel could only render a name with no magnitude.
//
// The admitted set must track Modifiers(): if the engine keeps applying an
// effect every tick, the player has to be able to see how big it is.
func TestGetActive_SurfacesRateEffects(t *testing.T) {
	em := NewEventManager()
	em.InjectEvent(ActiveEvent{
		Key:       factionBuffKey("dawnfolk"),
		Name:      "Dawnfolk Favour",
		TicksLeft: 120,
		Effects: []config.Effect{
			{Type: "food_rate", Target: "food", Value: 0.13},
			{Type: "instant_resource", Target: "gold", Value: 500}, // one-shot, must stay dropped
		},
	})

	active := em.GetActive()
	if len(active) != 1 {
		t.Fatalf("got %d active events, want 1", len(active))
	}
	effs := active[0].Effects
	if len(effs) != 1 {
		t.Fatalf("got %d surfaced effects %+v, want exactly the food_rate one", len(effs), effs)
	}
	if effs[0].Type != "food_rate" || effs[0].Target != "food" || effs[0].Value != 0.13 {
		t.Errorf("surfaced effect = %+v, want {food_rate food 0.13}", effs[0])
	}
}

// TestGetActive_RateSetbackKeepsSign proves a NEGATIVE rate effect (a faction
// setback) survives the filter with its sign intact — the panel colours off the
// sign, so dropping or flattening it would render a penalty as a gift.
func TestGetActive_RateSetbackKeepsSign(t *testing.T) {
	em := NewEventManager()
	em.InjectEvent(ActiveEvent{
		Key:       factionMalusKey("dawnfolk"),
		Name:      "Dawnfolk Reprisal",
		TicksLeft: 60,
		Effects: []config.Effect{
			{Type: "iron_rate", Target: "iron", Value: -0.20},
		},
	})

	active := em.GetActive()
	if len(active) != 1 || len(active[0].Effects) != 1 {
		t.Fatalf("setback did not surface: %+v", active)
	}
	if got := active[0].Effects[0].Value; got != -0.20 {
		t.Errorf("Value = %v, want -0.20", got)
	}
}

// TestGetActive_AdmitsWhatModifiersApplies is the guard that keeps the two
// filters from drifting apart again: every multiplier-bucket effect the engine
// keeps applying must be renderable.
func TestGetActive_AdmitsWhatModifiersApplies(t *testing.T) {
	em := NewEventManager()
	em.InjectEvent(ActiveEvent{
		Key:       "mixed",
		Name:      "Mixed Bag",
		TicksLeft: 30,
		Effects: []config.Effect{
			{Type: "production_all", Value: 0.10},
			{Type: "tick_speed", Value: 0.05},
			{Type: "gold_rate", Target: "gold", Value: 0.25},
			{Type: "worker_loss", Value: 3}, // one-shot: applied, never ongoing
		},
	})

	surfaced := map[string]bool{}
	for _, eff := range em.GetActive()[0].Effects {
		surfaced[eff.Type] = true
	}
	for _, m := range em.Modifiers() {
		if !surfaced[m.Target] {
			t.Errorf("Modifiers() applies %q every tick but GetActive() does not surface it", m.Target)
		}
	}
	if surfaced["worker_loss"] {
		t.Error("one-shot worker_loss leaked into the ongoing-effects view")
	}
}
