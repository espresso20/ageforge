package game

import (
	"math"
	"testing"
)

// Nanobot System (Trello 0kLti5GR) — engine-level checks that the new content
// actually plugs into the live economy: the producer makes nanobots, and the
// Nanofabrication tech's build-cost reduction flows through recalculateRates.

// TestNanobots_FoundryProducesThroughEngine verifies the nano_foundry building
// is known to the engine's BuildingManager and carries a positive nanobot
// production effect (the engine sources building production from these defs).
func TestNanobots_FoundryProducesThroughEngine(t *testing.T) {
	ge := NewGameEngine()
	def, ok := ge.Buildings.defs["nano_foundry"]
	if !ok {
		t.Fatal("engine BuildingManager has no nano_foundry def")
	}
	found := false
	for _, eff := range def.Effects {
		if eff.Type == "production" && eff.Target == "nanobots" && eff.Value > 0 {
			found = true
		}
	}
	if !found {
		t.Errorf("nano_foundry def carries no positive nanobots production effect: %+v", def.Effects)
	}
}

// TestNanobots_NanofabricationReducesBuildCost researches Nanofabrication and
// asserts the charged build cost drops by ~8% — proving the tech's build_cost
// effect flows through the resolver into the BuildingManager cost multiplier.
func TestNanobots_NanofabricationReducesBuildCost(t *testing.T) {
	ge := NewGameEngine()
	base := ge.Buildings.defs["hut"].BaseCost["wood"]

	// Baseline: no tech → full base cost.
	ge.recalculateRates()
	if got, _ := ge.Buildings.BuildBatchCost("hut", 1, nil); got["wood"] != base {
		t.Fatalf("pre-research: charged wood = %v, want base %v", got["wood"], base)
	}

	// Research Nanofabrication (LoadState marks it researched AND applies its
	// effects into the research bonus pool that buildResolver reads).
	ge.Research.LoadState([]string{"nanofabrication"}, "", 0, 0)
	ge.recalculateRates()

	want := math.Floor(base * 0.92) // -8% build_cost
	charged, _ := ge.Buildings.BuildBatchCost("hut", 1, nil)
	if charged["wood"] != want {
		t.Errorf("post-Nanofabrication: charged wood = %v, want %v (-8%% of %v)", charged["wood"], want, base)
	}
	if charged["wood"] >= base {
		t.Errorf("Nanofabrication did not reduce build cost: %v >= base %v", charged["wood"], base)
	}
}
