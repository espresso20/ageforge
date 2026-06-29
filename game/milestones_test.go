package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// fullAgeOrder returns a complete age order map for tests
func fullAgeOrder() map[string]int {
	pm := NewProgressManager()
	return pm.GetAgeOrder()
}

func TestMilestoneManager_CheckFirstShelter(t *testing.T) {
	mm := NewMilestoneManager()
	rm := NewResourceManager()
	bm := NewBuildingManager()
	bm.UnlockBuilding("hut")

	ageOrder := fullAgeOrder()

	// No hut — should not complete
	completed := mm.CheckMilestones(1, "primitive_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	if len(completed) != 0 {
		t.Errorf("expected 0 completions with no hut, got %d", len(completed))
	}

	// Build a hut
	bm.counts["hut"] = 1
	completed = mm.CheckMilestones(2, "primitive_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)

	found := false
	for _, ms := range completed {
		if ms.Key == "first_shelter" {
			found = true
		}
	}
	if !found {
		t.Error("first_shelter should complete when hut count >= 1")
	}

	// Should not trigger again
	completed = mm.CheckMilestones(3, "primitive_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	for _, ms := range completed {
		if ms.Key == "first_shelter" {
			t.Error("first_shelter should not trigger twice")
		}
	}
}

func TestMilestoneManager_PopulationMilestone(t *testing.T) {
	mm := NewMilestoneManager()
	rm := NewResourceManager()
	bm := NewBuildingManager()
	ageOrder := fullAgeOrder()

	// small_village requires pop 5,000 (hardened threshold)
	completed := mm.CheckMilestones(1, "primitive_age", ageOrder, rm, bm, 4999, 0, 0, nil, 0, 0, 0)
	for _, ms := range completed {
		if ms.Key == "small_village" {
			t.Error("small_village should not trigger at pop 4999")
		}
	}

	completed = mm.CheckMilestones(2, "primitive_age", ageOrder, rm, bm, 5000, 0, 0, nil, 0, 0, 0)
	found := false
	for _, ms := range completed {
		if ms.Key == "small_village" {
			found = true
		}
	}
	if !found {
		t.Error("small_village should trigger at pop 5000")
	}
}

func TestMilestoneManager_AgeMilestone(t *testing.T) {
	mm := NewMilestoneManager()
	rm := NewResourceManager()
	bm := NewBuildingManager()
	ageOrder := fullAgeOrder()

	// bronze_pioneer requires bronze_age
	completed := mm.CheckMilestones(1, "stone_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	for _, ms := range completed {
		if ms.Key == "bronze_pioneer" {
			t.Error("bronze_pioneer should not trigger in stone_age")
		}
	}

	completed = mm.CheckMilestones(2, "bronze_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	found := false
	for _, ms := range completed {
		if ms.Key == "bronze_pioneer" {
			found = true
		}
	}
	if !found {
		t.Error("bronze_pioneer should trigger in bronze_age")
	}
}

func TestMilestoneManager_ChainCompletion(t *testing.T) {
	mm := NewMilestoneManager()

	// Military chain now requires: first_soldiers, war_machine, iron_legion,
	// fortress_state, military_superpower
	mm.completed["first_soldiers"] = true
	mm.completed["war_machine"] = true
	mm.completed["iron_legion"] = true
	mm.completed["fortress_state"] = true
	mm.completed["military_superpower"] = true

	chains := mm.CheckChains()
	found := false
	for _, c := range chains {
		if c.Key == "military_chain" {
			found = true
			if c.Title != "The Conquerors" {
				t.Errorf("military chain title = %v, want The Conquerors", c.Title)
			}
		}
	}
	if !found {
		t.Error("military_chain should complete when all military milestones are done")
	}

	// Should not trigger again
	chains = mm.CheckChains()
	for _, c := range chains {
		if c.Key == "military_chain" {
			t.Error("military_chain should not trigger twice")
		}
	}
}

func TestMilestoneManager_TitleRecalculation(t *testing.T) {
	mm := NewMilestoneManager()

	// No milestones — no title
	mm.recalculateTitle()
	if mm.currentTitle != "" {
		t.Errorf("title with 0 milestones = %v, want empty", mm.currentTitle)
	}

	// 5 milestones = "Aspiring" (MilestoneTitles() requires MinMilestones: 5)
	mm.completed["first_shelter"] = true
	mm.completed["small_village"] = true
	mm.completed["knowledge_seeker"] = true
	mm.completed["first_research"] = true
	mm.completed["stone_mason"] = true
	mm.recalculateTitle()
	if mm.currentTitle != "Aspiring" {
		t.Errorf("title with 5 milestones = %v, want Aspiring", mm.currentTitle)
	}

	// Complete a chain — chain title overrides
	mm.completed["war_machine"] = true
	mm.chainsCompleted["military_chain"] = true
	mm.recalculateTitle()
	if mm.currentTitle != "The Conquerors" {
		t.Errorf("title with chain = %v, want The Conquerors", mm.currentTitle)
	}
}

func TestMilestoneManager_Snapshot(t *testing.T) {
	mm := NewMilestoneManager()
	mm.completed["first_shelter"] = true

	params := MilestoneSnapshotParams{
		Age:      "primitive_age",
		AgeOrder: fullAgeOrder(),
	}
	snap := mm.Snapshot(params)

	if snap.CompletedCount != 1 {
		t.Errorf("snapshot completed = %v, want 1", snap.CompletedCount)
	}
	if !snap.Milestones["first_shelter"].Completed {
		t.Error("first_shelter should be completed in snapshot")
	}
	if snap.Milestones["first_shelter"].RewardText == "" {
		t.Error("completed milestone should have reward text")
	}
}

func TestMilestoneManager_HiddenVisibility(t *testing.T) {
	mm := NewMilestoneManager()

	params := MilestoneSnapshotParams{
		Age:      "primitive_age",
		AgeOrder: fullAgeOrder(),
	}
	snap := mm.Snapshot(params)

	// metropolis is hidden, should not be visible at primitive age with no progress
	if snap.Milestones["metropolis"].Visible {
		t.Error("metropolis should be hidden at primitive_age with no progress")
	}

	// first_shelter is NOT hidden, should be visible
	if !snap.Milestones["first_shelter"].Visible {
		t.Error("first_shelter should be visible (not hidden)")
	}
}

func TestMilestoneManager_SaveLoadRoundTrip(t *testing.T) {
	mm := NewMilestoneManager()
	mm.completed["first_shelter"] = true
	mm.completed["war_machine"] = true
	mm.chainsCompleted["military_chain"] = true
	mm.currentTitle = "The Conquerors"

	// Save
	completed := mm.GetCompleted()
	chains := mm.GetChainsCompleted()
	title := mm.GetCurrentTitle()

	// Load into fresh
	mm2 := NewMilestoneManager()
	mm2.LoadState(completed, chains, title)

	if !mm2.IsCompleted("first_shelter") {
		t.Error("loaded manager should have first_shelter completed")
	}
	if !mm2.IsCompleted("war_machine") {
		t.Error("loaded manager should have war_machine completed")
	}
	if mm2.currentTitle != "The Conquerors" {
		t.Errorf("loaded title = %v, want The Conquerors", mm2.currentTitle)
	}
}

// TestMilestoneChain_CompletionBoosts guards the chain rebalance: every chain's
// completion speed-boost must match the normalized values. Military used to be
// 18x weaker than Settlement; this pins all six so a future edit can't silently
// regress the balance.
func TestMilestoneChain_CompletionBoosts(t *testing.T) {
	want := map[string]struct {
		value    float64
		duration int
	}{
		"settlement_chain":   {3.0, 180},
		"scholar_chain":      {3.0, 180},
		"builder_chain":      {2.5, 150},
		"military_chain":     {2.5, 150},
		"trade_chain":        {2.5, 150},
		"ancient_ages_chain": {2.5, 150},
	}

	chains := config.MilestoneChainByKey()
	if len(chains) != len(want) {
		t.Fatalf("chain count = %d, want %d", len(chains), len(want))
	}
	for key, exp := range want {
		c, ok := chains[key]
		if !ok {
			t.Errorf("missing chain %q", key)
			continue
		}
		if c.BoostValue != exp.value {
			t.Errorf("%s BoostValue = %v, want %v", key, c.BoostValue, exp.value)
		}
		if c.BoostDuration != exp.duration {
			t.Errorf("%s BoostDuration = %v, want %v", key, c.BoostDuration, exp.duration)
		}
	}
}

// TestTradeChain_Structure verifies the Trade chain was expanded from 3 to 6
// milestones in the correct order (the Trade Expansion added maritime_empire as
// the capstone), and that the new milestones exist, are categorized as trade,
// and map back to trade_chain.
func TestTradeChain_Structure(t *testing.T) {
	chain := config.MilestoneChainByKey()["trade_chain"]
	wantKeys := []string{
		"first_market", "merchant_guild", "caravan_network",
		"merchant_princes", "trade_empire", "maritime_empire",
	}
	if len(chain.MilestoneKeys) != len(wantKeys) {
		t.Fatalf("trade_chain has %d milestones, want %d", len(chain.MilestoneKeys), len(wantKeys))
	}
	for i, k := range wantKeys {
		if chain.MilestoneKeys[i] != k {
			t.Errorf("trade_chain key[%d] = %q, want %q", i, chain.MilestoneKeys[i], k)
		}
	}

	byKey := config.MilestoneByKey()
	for _, k := range []string{"caravan_network", "merchant_princes", "maritime_empire"} {
		def, ok := byKey[k]
		if !ok {
			t.Errorf("new trade milestone %q not found in Milestones()", k)
			continue
		}
		if def.Category != "trade" {
			t.Errorf("%s category = %q, want trade", k, def.Category)
		}
	}
}

// TestTradeChain_NewMilestonesComplete drives the two new Trade milestones
// through the real engine check path: they must stay incomplete below threshold
// and complete once buildings + age conditions are met.
func TestTradeChain_NewMilestonesComplete(t *testing.T) {
	mm := NewMilestoneManager()
	rm := NewResourceManager()
	bm := NewBuildingManager()
	ageOrder := fullAgeOrder()

	// caravan_network: classical_age + 5 trading_post.
	// Below threshold (4 posts, classical) → no completion.
	bm.counts["trading_post"] = 4
	completed := mm.CheckMilestones(1, "classical_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	for _, ms := range completed {
		if ms.Key == "caravan_network" {
			t.Error("caravan_network should not complete with only 4 trading posts")
		}
	}

	// 5 posts but still iron_age → age gate blocks it.
	bm.counts["trading_post"] = 5
	completed = mm.CheckMilestones(2, "iron_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	for _, ms := range completed {
		if ms.Key == "caravan_network" {
			t.Error("caravan_network should not complete before classical_age")
		}
	}

	// 5 posts in classical_age → completes.
	completed = mm.CheckMilestones(3, "classical_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	if !containsKey(completed, "caravan_network") {
		t.Error("caravan_network should complete with 5 trading posts in classical_age")
	}

	// merchant_princes: medieval_age + 12 trading_post + 4 merchant_quarter.
	bm.counts["trading_post"] = 12
	bm.counts["merchant_quarter"] = 3 // one short
	completed = mm.CheckMilestones(4, "medieval_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	for _, ms := range completed {
		if ms.Key == "merchant_princes" {
			t.Error("merchant_princes should not complete with only 3 merchant quarters")
		}
	}

	bm.counts["merchant_quarter"] = 4
	completed = mm.CheckMilestones(5, "medieval_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0, 0)
	if !containsKey(completed, "merchant_princes") {
		t.Error("merchant_princes should complete with 12 trading posts + 4 merchant quarters in medieval_age")
	}
}

// TestMilestoneChain_InjectsBoostEvent reproduces the engine's chain-completion
// path (CheckChains -> InjectEvent) and asserts the injected tick_speed event
// carries the chain's normalized BoostValue and BoostDuration. Previously the
// boost injection had no test coverage.
func TestMilestoneChain_InjectsBoostEvent(t *testing.T) {
	mm := NewMilestoneManager()
	em := NewEventManager()

	// Complete the military chain (now 2.5 / 150).
	for _, k := range []string{
		"first_soldiers", "war_machine", "iron_legion",
		"fortress_state", "military_superpower",
	} {
		mm.completed[k] = true
	}

	newChains := mm.CheckChains()
	if len(newChains) != 1 || newChains[0].Key != "military_chain" {
		t.Fatalf("expected military_chain to complete, got %v", newChains)
	}

	// Mirror engine.go: inject the speed boost for each newly completed chain.
	for _, chain := range newChains {
		em.InjectEvent(ActiveEvent{
			Key:       chain.Key + "_boost",
			Name:      chain.Name + " Speed Boost",
			TicksLeft: chain.BoostDuration,
			Effects: []config.Effect{
				{Type: "tick_speed", Target: "tick_speed", Value: chain.BoostValue},
			},
		})
	}

	// The injected event must expose a tick_speed effect of +2.5.
	var found bool
	for _, eff := range em.GetActiveEffects() {
		if eff.Type == "tick_speed" {
			found = true
			if eff.Value != 2.5 {
				t.Errorf("military chain boost value = %v, want 2.5", eff.Value)
			}
		}
	}
	if !found {
		t.Error("expected a tick_speed effect from the injected chain boost")
	}
}

// containsKey reports whether a completed-milestone slice includes the given key.
func containsKey(defs []config.MilestoneDef, key string) bool {
	for _, d := range defs {
		if d.Key == key {
			return true
		}
	}
	return false
}
