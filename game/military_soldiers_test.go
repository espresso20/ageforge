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

// TestSoldiersTrainFromTheFirstMilitaryBuilding: soldiers are locked until
// the Iron Age for a town with no military building, and nothing is held or
// counted before then. The first military building to stand unlocks them,
// whatever the age (config.ResourceDef.BuiltUnlocks): a War Camp trains
// soldiers in the Stone Age, staffed or not, and every post in it produces.
func TestSoldiersTrainFromTheFirstMilitaryBuilding(t *testing.T) {
	const camp = "war_camp"
	def := config.BuildingByKey()[camp]
	run := func(ge *GameEngine, ticks int) {
		for i := 0; i < ticks; i++ {
			ge.mu.Lock()
			ge.Resources.Add("food", 1e6)
			ge.morale = 1.0
			ge.mu.Unlock()
			ge.StepTicks(1)
		}
	}
	stone := func() *GameEngine {
		ge := NewGameEngine()
		ge.mu.Lock()
		ge.applyAgeUnlocks("stone_age")
		ge.age, ge.currentEpoch = "stone_age", ge.rules.EraOf("stone_age")
		ge.Workers.SetAge("stone_age")
		ge.Resources.AddStorage("food", 1e9)
		ge.Workers.UnlockType("worker")
		ge.mu.Unlock()
		return ge
	}

	// No military building: soldiers stay locked, and nothing counts.
	ge := stone()
	run(ge, 40)
	st := ge.GetState()
	if st.Resources["soldiers"].Unlocked || st.Resources["soldiers"].Rate != 0 || st.Military.SoldierCount != 0 {
		t.Errorf("a Stone Age town with no War Camp has soldiers: %+v", st.Resources["soldiers"])
	}
	if got := st.Stats.TotalGathered["soldiers"]; got != 0 || ge.Stats.SoldiersTrained != 0 {
		t.Errorf("the stats count %v soldiers gathered and %v trained in a town with no War Camp", got, ge.Stats.SoldiersTrained)
	}

	// One War Camp, unstaffed: soldiers unlock and train at the unstaffed
	// rate.
	ge.mu.Lock()
	ge.Buildings.counts[camp] = 1
	ge.Buildings.unlocked[camp] = true
	ge.mu.Unlock()
	run(ge, 1)
	st = ge.GetState()
	if !st.Resources["soldiers"].Unlocked {
		t.Fatal("a standing War Camp did not unlock soldiers")
	}
	idle := st.Resources["soldiers"].Rate
	if idle <= 0 || st.Military.SoldierRate != idle {
		t.Errorf("an unstaffed War Camp trains %v soldiers a tick (the Army panel says %v), want more than none", idle, st.Military.SoldierRate)
	}

	// Staffed, every post adds to it: no post in a War Camp produces
	// nothing.
	prev := idle
	for post := 1; post <= def.WorkerCapacity; post++ {
		ge.mu.Lock()
		ok := ge.Workers.Recruit("worker", 1, 100000) && ge.Workers.Assign("worker", camp, 1)
		ge.mu.Unlock()
		if !ok {
			t.Fatalf("could not staff post %d of the War Camp", post)
		}
		run(ge, 1)
		rate := ge.GetState().Resources["soldiers"].Rate
		if rate <= prev {
			t.Errorf("post %d of the War Camp adds nothing: %v soldiers a tick, %v before", post, rate, prev)
		}
		prev = rate
	}
	run(ge, 20)
	st = ge.GetState()
	if st.Military.SoldierCount <= 0 || st.Stats.TotalGathered["soldiers"] <= 0 || ge.Stats.SoldiersTrained <= 0 {
		t.Errorf("a staffed War Camp trained nobody: %d held, %v gathered, %v trained", st.Military.SoldierCount, st.Stats.TotalGathered["soldiers"], ge.Stats.SoldiersTrained)
	}
	if st.Military.Mitigation <= 0 {
		t.Errorf("a Stone Age garrison of %d blunts nothing of a raid (threat %v)", st.Military.SoldierCount, st.Military.Threat)
	}

	// The unlock is the run's: it is saved and comes back.
	isolateAccountDir(t)
	if err := ge.SaveGame("garrison"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("garrison"); err != nil {
		t.Fatal(err)
	}
	if got := loaded.GetState(); !got.Resources["soldiers"].Unlocked || got.Military.SoldierCount != st.Military.SoldierCount {
		t.Errorf("after a load the garrison is %d soldiers (unlocked %v), want %d", got.Military.SoldierCount, got.Resources["soldiers"].Unlocked, st.Military.SoldierCount)
	}
}
