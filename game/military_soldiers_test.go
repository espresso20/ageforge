package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// militarySoldierCap returns the per-building soldier storage cap (the existing
// {capacity, military} value) for the given military building key, sourced from
// config so the test stays in sync with the building definitions.
func militarySoldierCap(t *testing.T, key string) float64 {
	t.Helper()
	def, ok := config.BuildingByKey()[key]
	if !ok {
		t.Fatalf("building %q not found in config", key)
	}
	for _, eff := range def.Effects {
		if eff.Type == "storage" && eff.Target == "soldiers" {
			return eff.Value
		}
	}
	t.Fatalf("building %q has no soldiers storage effect", key)
	return 0
}

// TestMilitaryBuildings_ProduceAndStoreSoldiers verifies the Stage-1 military
// rework foundation: military buildings produce the `soldiers` resource, the
// soldiers storage cap equals the sum of built military buildings' storage
// effects, and the soldiers amount is clamped at that cap.
func TestMilitaryBuildings_ProduceAndStoreSoldiers(t *testing.T) {
	ge := NewGameEngine()

	const barracks = "barracks"
	def := config.BuildingByKey()[barracks]
	if def.WorkerDomain != "military" {
		t.Fatalf("expected barracks WorkerDomain=military, got %q", def.WorkerDomain)
	}

	// Build 2 barracks, unlock + fully staff them, and unlock the soldiers resource.
	ge.mu.Lock()
	const numBarracks = 2
	ge.Buildings.counts[barracks] = numBarracks
	ge.Buildings.unlocked[barracks] = true
	ge.Resources.UnlockResource("soldiers")
	ge.Resources.UnlockResource("food")
	ge.Resources.AddStorage("food", 1e9)
	ge.Workers.UnlockType("worker")
	// Recruit and assign a full crew (popCap high enough not to gate the test).
	totalSlots := def.WorkerCapacity * numBarracks
	if !ge.Workers.Recruit("worker", totalSlots, 100000) {
		ge.mu.Unlock()
		t.Fatalf("failed to recruit %d workers", totalSlots)
	}
	if !ge.Workers.Assign("worker", barracks, totalSlots) {
		ge.mu.Unlock()
		t.Fatalf("failed to assign %d workers to barracks", totalSlots)
	}
	ge.mu.Unlock()

	expectedCap := militarySoldierCap(t, barracks) * numBarracks

	// --- Soldiers production rate is positive once buildings are worked. ---
	ge.mu.Lock()
	ge.recalculateRates()
	rate := ge.Resources.resources["soldiers"].Rate
	gotCap := ge.Resources.resources["soldiers"].Storage
	ge.mu.Unlock()

	if rate <= 0 {
		t.Errorf("soldiers production rate = %v, want > 0", rate)
	}

	// --- Soldiers storage cap == sum of built military buildings' storage. ---
	if gotCap != expectedCap {
		t.Errorf("soldiers storage cap = %v, want %v (sum of %d barracks)", gotCap, expectedCap, numBarracks)
	}

	// tick advances one engine tick. Food is topped up and morale pinned to 1.0
	// before each tick so this test isolates soldiers production/clamping from the
	// food-starvation and high-military-ratio morale penalties (those are exercised
	// by the morale tests, and an all-military workforce would otherwise drag morale
	// to its floor and throttle production).
	tick := func() {
		ge.mu.Lock()
		ge.Resources.Add("food", 1e6)
		ge.morale = 1.0
		ge.mu.Unlock()
		ge.doTick()
		ge.mu.Lock()
		ge.morale = 1.0
		ge.mu.Unlock()
	}

	// --- Soldiers amount increases after a few ticks. ---
	for i := 0; i < 5; i++ {
		tick()
	}
	afterFew := ge.Resources.Get("soldiers")
	if afterFew <= 0 {
		t.Errorf("soldiers after 5 ticks = %v, want > 0", afterFew)
	}

	// --- Soldiers amount clamps at the storage cap. ---
	// fully-worked 2 barracks produce ~0.8/tick into a cap of 40 → ~50 ticks to
	// fill; 300 ticks guarantees it saturates and holds.
	for i := 0; i < 300; i++ {
		tick()
	}
	atCap := ge.Resources.Get("soldiers")
	if atCap != expectedCap {
		t.Errorf("soldiers after saturation = %v, want clamped at cap %v", atCap, expectedCap)
	}
}
