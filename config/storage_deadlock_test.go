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

// TestStash_MaxCapClearsStoneAgeFoodGate guards the second half of the softlock:
// even if the first stash is buildable, the FULL stack of stashes must raise the
// food cap above the stone-age food requirement, or the player can never store
// enough food to advance. Stash is the only primitive-age building that lifts the
// food cap. The old +300/copy effect topped out at 50 + 50*300 = 15,050, below the
// normalized 16,000 gate — a silent softlock. See yQw8uK8S.
func TestStash_MaxCapClearsStoneAgeFoodGate(t *testing.T) {
	stash, ok := BuildingByKey()["stash"]
	if !ok {
		t.Fatal("stash building not found")
	}

	// Per-copy storage lift (the "all"-targeted storage effect).
	var perCopy float64
	foundEffect := false
	for _, e := range stash.Effects {
		if e.Type == "storage" && (e.Target == "all" || e.Target == "food") {
			perCopy = e.Value
			foundEffect = true
			break
		}
	}
	if !foundEffect {
		t.Fatal("stash has no all/food storage effect")
	}

	// Base food storage cap before any stashes.
	var foodCap float64
	foundFood := false
	for _, r := range BaseResources() {
		if r.Key == "food" {
			foodCap = r.BaseStorage
			foundFood = true
			break
		}
	}
	if !foundFood {
		t.Fatal("food resource not found")
	}

	// Normalized stone-age food requirement (raw 8000 * 2.0 band = 16,000).
	stone, ok := AgeByKey()["stone_age"]
	if !ok {
		t.Fatal("stone_age not found")
	}
	foodReq := stone.ResourceReqs["food"]
	if foodReq <= 0 {
		t.Fatal("stone_age has no food requirement")
	}

	maxCap := foodCap + float64(stash.MaxCount)*perCopy
	if maxCap < foodReq {
		t.Fatalf("SOFTLOCK: max reachable food cap is %.0f (base %.0f + %d stashes * %.0f/copy) "+
			"but the stone-age food gate is %.0f — the player can never store enough to advance",
			maxCap, foodCap, stash.MaxCount, perCopy, foodReq)
	}
}
