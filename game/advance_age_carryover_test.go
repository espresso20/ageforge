package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestAdvanceAge_CapsHoardedResource verifies that a player who over-accumulates
// a resource the new age uses has it capped to ~CarryoverStarterBuildings of the
// cheapest new-age building's cost — NOT carried over as a flat percentage of the
// hoard. (EPIC: age-pacing economy rebalance, sub-ticket 2.)
func TestAdvanceAge_CapsHoardedResource(t *testing.T) {
	ge := NewGameEngine()

	// HUGE stockpile of a resource stone_age buildings use (wood + stone).
	const hoard = config.Million
	ge.Resources.LoadAmounts(map[string]float64{
		"wood":  hoard,
		"stone": hoard,
	})

	ge.advanceAge("stone_age")

	entry := config.AgeEntryCosts("stone_age")
	for _, res := range []string{"wood", "stone"} {
		cap := CarryoverStarterBuildings * entry[res]
		got := ge.Resources.Get(res)
		if got != cap {
			t.Errorf("%s: hoard should be capped to %.2f (8×%.2f), got %.2f",
				res, cap, entry[res], got)
		}
		// Sanity: the cap must be far below the old flat-10%% carryover.
		if old := hoard * 0.10; got >= old {
			t.Errorf("%s: cap %.2f is not below old 10%% carryover %.2f — model not applied",
				res, got, old)
		}
	}
}

// TestAdvanceAge_PreservesModestStockpile verifies that a player whose stockpile
// is below the starter cap keeps what they had — it is NOT cut to a percentage.
func TestAdvanceAge_PreservesModestStockpile(t *testing.T) {
	ge := NewGameEngine()

	entry := config.AgeEntryCosts("stone_age")
	// Pick a modest amount well under the wood cap (8 × cheapest wood building).
	modest := entry["wood"] * 2 // two starter buildings' worth
	if modest <= 0 {
		t.Fatal("stone_age wood entry cost is zero — test cannot proceed")
	}
	ge.Resources.LoadAmounts(map[string]float64{"wood": modest})

	ge.advanceAge("stone_age")

	got := ge.Resources.Get("wood")
	if got != modest {
		t.Errorf("modest wood stockpile should be preserved: want %.2f, got %.2f", modest, got)
	}
}

// TestAdvanceAge_ResidualForUnusedResource verifies the fallback path: a resource
// that no new-age building uses is reduced to the small residual percentage.
func TestAdvanceAge_ResidualForUnusedResource(t *testing.T) {
	ge := NewGameEngine()

	entry := config.AgeEntryCosts("stone_age")
	// Find a resource that exists on the engine but is NOT in the stone_age entry
	// costs (no stone_age non-wonder building uses it).
	var unused string
	var startAmt float64
	for key := range ge.Resources.resources {
		if key == "faith" {
			continue
		}
		if _, used := entry[key]; !used {
			unused = key
			break
		}
	}
	if unused == "" {
		t.Skip("no unused-by-stone_age resource available to test residual path")
	}

	startAmt = 1000.0
	ge.Resources.LoadAmounts(map[string]float64{unused: startAmt})

	ge.advanceAge("stone_age")

	want := startAmt * CarryoverResidualPct
	got := ge.Resources.Get(unused)
	if got != want {
		t.Errorf("unused resource %q should drop to residual %.2f (%.0f%%), got %.2f",
			unused, want, CarryoverResidualPct*100, got)
	}
}

// TestAdvanceAge_FaithExcluded confirms faith is never reduced on age transition.
func TestAdvanceAge_FaithExcluded(t *testing.T) {
	ge := NewGameEngine()
	const faith = 500.0
	ge.Resources.LoadAmounts(map[string]float64{"faith": faith})

	ge.advanceAge("stone_age")

	if got := ge.Resources.Get("faith"); got != faith {
		t.Errorf("faith should be untouched on age transition: want %.2f, got %.2f", faith, got)
	}
}

// TestAdvanceAge_EndedResourceExpires verifies that when advancing to an age
// where a resource has ended, its remaining bank expires (drops to 0).
func TestAdvanceAge_EndedResourceExpires(t *testing.T) {
	ge := NewGameEngine()
	const startWood = 50000.0
	ge.Resources.LoadAmounts(map[string]float64{"wood": startWood})

	ge.advanceAge("colonial_age")

	if got := ge.Resources.Get("wood"); got != 0 {
		t.Errorf("ended resource wood should expire on advance to colonial_age: want 0, got %.2f", got)
	}
}

// TestAdvanceAge_RetiresEndedResourceBuildings verifies that advancing to an age
// where resources have ended retires the buildings that produce them, sets counts
// and ruins to 0, returns assigned workers to the idle pool, and locks the buildings.
func TestAdvanceAge_RetiresEndedResourceBuildings(t *testing.T) {
	ge := NewGameEngine()

	// Setup wood_camp, sawmill, quarry with counts and stockpiles
	ge.Buildings.counts["wood_camp"] = 16
	ge.Buildings.counts["sawmill"] = 5
	ge.Buildings.counts["quarry"] = 10
	ge.Buildings.UnlockBuilding("wood_camp")
	ge.Buildings.UnlockBuilding("sawmill")
	ge.Buildings.UnlockBuilding("quarry")

	// Recruit workers and assign them
	ge.Workers.Unlock()
	if !ge.Workers.Recruit(100, 200) {
		t.Fatal("failed to recruit workers")
	}

	// wood_camp takes 3 workers per building (capacity 3)
	if !ge.Workers.Assign("wood_camp", 48) {
		t.Fatal("failed to assign workers to wood_camp")
	}
	// quarry takes workers too (capacity 5 per building)
	if !ge.Workers.Assign("quarry", 20) {
		t.Fatal("failed to assign workers to quarry")
	}

	if got := ge.Workers.IdleCount(); got != 32 {
		t.Fatalf("expected 32 idle workers before advance, got %d", got)
	}

	ge.Resources.LoadAmounts(map[string]float64{"wood": 50000, "stone": 30000})
	ge.buildQueue = append(ge.buildQueue, BuildQueueItem{BuildingKey: "wood_camp", TicksLeft: 10, TotalTicks: 10})

	// Advance to space_age where both wood and stone are ended
	ge.advanceAge("space_age")

	// Verify counts are zeroed
	if got := ge.Buildings.GetCount("wood_camp"); got != 0 {
		t.Errorf("wood_camp count after advance to space_age = %d, want 0", got)
	}
	if got := ge.Buildings.GetCount("sawmill"); got != 0 {
		t.Errorf("sawmill count after advance to space_age = %d, want 0", got)
	}
	if got := ge.Buildings.GetCount("quarry"); got != 0 {
		t.Errorf("quarry count after advance to space_age = %d, want 0", got)
	}

	// Verify stockpiles are zeroed
	if got := ge.Resources.Get("wood"); got != 0 {
		t.Errorf("wood stockpile after advance to space_age = %v, want 0", got)
	}
	if got := ge.Resources.Get("stone"); got != 0 {
		t.Errorf("stone stockpile after advance to space_age = %v, want 0", got)
	}

	// Verify assigned workers are 0 and freed back to idle
	if got := ge.Workers.GetAssignedCount("wood_camp"); got != 0 {
		t.Errorf("assigned workers to wood_camp = %d, want 0", got)
	}
	if got := ge.Workers.GetAssignedCount("quarry"); got != 0 {
		t.Errorf("assigned workers to quarry = %d, want 0", got)
	}
	if got := ge.Workers.IdleCount(); got != 100 {
		t.Errorf("idle workers after advance = %d, want 100 (all 68 assigned workers freed)", got)
	}

	// Verify buildings are locked
	if ge.Buildings.IsUnlocked("wood_camp") {
		t.Error("wood_camp should not be unlocked in space_age")
	}
	if ge.Buildings.IsUnlocked("quarry") {
		t.Error("quarry should not be unlocked in space_age")
	}

	// Verify buildQueue item was removed
	for _, item := range ge.buildQueue {
		if item.BuildingKey == "wood_camp" {
			t.Error("wood_camp queue item should have been removed from buildQueue")
		}
	}
}

// TestRetireEndedResources_LoadSaveInSpaceAge verifies that loading an existing
// save in space_age that contains ended-resource buildings immediately retires them
// and frees their assigned workers.
func TestRetireEndedResources_LoadSaveInSpaceAge(t *testing.T) {
	isolateAccountDir(t)
	ge := NewGameEngine()
	ge.age = "space_age"
	ge.currentEpoch = config.EpochForAge("space_age")
	ge.Buildings.counts["wood_camp"] = 16
	ge.Buildings.UnlockBuilding("wood_camp")
	ge.Workers.Unlock()
	ge.Workers.Recruit(100, 200)
	ge.Workers.Assign("wood_camp", 48)
	ge.Resources.LoadAmounts(map[string]float64{"wood": 50000})

	loaded := saveAndLoad(t, ge, "space_ended_res")

	if got := loaded.Buildings.GetCount("wood_camp"); got != 0 {
		t.Errorf("loaded wood_camp count = %d, want 0", got)
	}
	if got := loaded.Workers.GetAssignedCount("wood_camp"); got != 0 {
		t.Errorf("loaded wood_camp assigned workers = %d, want 0", got)
	}
	if got := loaded.Workers.IdleCount(); got != 100 {
		t.Errorf("loaded idle workers = %d, want 100 (48 freed back to idle)", got)
	}
	if got := loaded.Resources.Get("wood"); got != 0 {
		t.Errorf("loaded wood stockpile = %v, want 0", got)
	}
}

// TestRetireEndedResources_DevModeAgeJump verifies that /age space_age retires ended
// resource buildings (freeing workers) and upgrades pre-space buildings like hut to orbital_habitat.
func TestRetireEndedResources_DevModeAgeJump(t *testing.T) {
	ge := NewGameEngine()
	ge.Buildings.counts["wood_camp"] = 16
	ge.Buildings.UnlockBuilding("wood_camp")
	ge.Buildings.counts["hut"] = 5
	ge.Buildings.UnlockBuilding("hut")

	ge.Workers.Unlock()
	ge.Workers.Recruit(100, 200)
	ge.Workers.Assign("wood_camp", 48)

	DevModeActive = true
	defer func() { DevModeActive = false }()

	reply := DevExecCommand("/age space_age", ge)
	if reply != "jumped to space_age" {
		t.Fatalf("DevExecCommand reply = %q, want 'jumped to space_age'", reply)
	}

	// wood_camp retired
	if got := ge.Buildings.GetCount("wood_camp"); got != 0 {
		t.Errorf("wood_camp count after /age space_age = %d, want 0", got)
	}
	if got := ge.Workers.GetAssignedCount("wood_camp"); got != 0 {
		t.Errorf("wood_camp assigned workers = %d, want 0", got)
	}
	if got := ge.Workers.IdleCount(); got != 100 {
		t.Errorf("idle workers after /age space_age = %d, want 100", got)
	}

	// hut upgraded to orbital_habitat
	if got := ge.Buildings.GetCount("hut"); got != 0 {
		t.Errorf("hut count after /age space_age = %d, want 0", got)
	}
	if got := ge.Buildings.GetCount("orbital_habitat"); got != 5 {
		t.Errorf("orbital_habitat count after /age space_age = %d, want 5", got)
	}
}
