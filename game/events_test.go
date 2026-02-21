package game

import (
	"testing"

	"github.com/user/ageforge/config"
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
	em.Tick(1, "primitive_age", ageOrder)
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
	_, expired := em.Tick(2, "primitive_age", ageOrder)
	foundExpired := false
	for _, key := range expired {
		if key == "short_boost" {
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
