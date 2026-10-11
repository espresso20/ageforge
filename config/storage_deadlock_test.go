package config

import "testing"

// TestStash_FirstCopyAffordableWithinBaseWoodCap guards against the primitive-age
// deadlock where the first stash costs more wood than the player can store.
// Stash is the only building that raises the wood storage cap, so if its first
// copy costs more than the base cap, it can never be built — softlocking the run.
// (The cost-curve rebalance once pushed it to 59 vs a 50 cap; this test exists so
// any future re-tune of normalizeCostCurves or the base cap re-surfaces it loudly.)
func TestStash_FirstCopyAffordableWithinBaseWoodCap(t *testing.T) {
	stash, ok := BuildingByKey()["stash"]
	if !ok {
		t.Fatal("stash building not found")
	}

	var woodCap float64
	found := false
	for _, r := range BaseResources() {
		if r.Key == "wood" {
			woodCap = r.BaseStorage
			found = true
			break
		}
	}
	if !found {
		t.Fatal("wood resource not found")
	}

	// First copy (count 0) costs exactly the (normalized) base.
	firstCopy := stash.BaseCost["wood"]

	if firstCopy > woodCap {
		t.Fatalf("DEADLOCK: stash first copy costs %.0f wood but the base wood cap is %.0f — "+
			"the only building that raises the cap is unbuildable", firstCopy, woodCap)
	}
	// Demand a real margin: a player must be able to gather up to the cost before
	// hitting the cap, so it shouldn't sit right at the ceiling.
	if firstCopy > woodCap*0.9 {
		t.Errorf("stash first copy (%.0f wood) is too close to the wood cap (%.0f) — "+
			"leave margin so it's reachable", firstCopy, woodCap)
	}
}
