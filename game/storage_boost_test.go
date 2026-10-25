package game

import (
	"math"
	"slices"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A storage building holds nothing (stock has no cap); each standing copy
// adds config.StorageBoostPercent percent to all production, through the
// all-production pool every wonder and milestone bonus joins.

// boostOf is the factor the all-production pool gives the town's buildings'
// output: building rate plus bonus rate over building rate, read off the first
// resource that gets nothing but that pool.
func boostOf(t *testing.T, ge *GameEngine) float64 {
	t.Helper()
	f, ok := truthAllFactor(ge, readTruth(t, ge))
	if !ok {
		t.Fatal("the lab town makes nothing")
	}
	return f
}

// TestEachStorageCopyAddsOnePercentToProduction: n copies are worth n percent,
// copies of older ages' storage buildings count with the newer ones, and the
// storage buildings are the pool's only source here.
func TestEachStorageCopyAddsOnePercentToProduction(t *testing.T) {
	ge := newTruthEngine("bronze_age", truthClean)
	if got := boostOf(t, ge); math.Abs(got-1) > 1e-12 {
		t.Fatalf("a town with no storage building runs at x%v, want x1", got)
	}
	steps := []struct {
		counts map[string]int
		want   float64
	}{
		{map[string]int{"stash": 1}, 1.01},
		{map[string]int{"stash": 4}, 1.04},
		// Older copies keep counting once their age is past (a Stash cannot
		// be built in the Bronze Age): 4 + 3 + 2.
		{map[string]int{"stash": 4, "storage_pit": 3, "warehouse": 2}, 1.09},
	}
	for _, s := range steps {
		for key := range ge.Buildings.defs {
			if ge.Buildings.defs[key].Category == "storage" {
				ge.Buildings.counts[key] = s.counts[key]
			}
		}
		ge.recalculateRates()
		if got := boostOf(t, ge); math.Abs(got-s.want) > 1e-12 {
			t.Errorf("storage copies %v: production runs at x%v, want x%v (%v%% a copy)", s.counts, got, s.want, config.StorageBoostPercent)
		}
	}
}

// TestStorageBoostShowsInTheBonusBreakdown: the boost is one source,
// "storage", of the production_all pool, and nothing else comes of the
// storage buildings: no store grows by them.
func TestStorageBoostShowsInTheBonusBreakdown(t *testing.T) {
	ge := newTruthEngine("bronze_age", truthClean)
	ge.Buildings.counts["stash"] = 5
	ge.Buildings.counts["warehouse"] = 2
	ge.recalculateRates()
	found := 0.0
	for _, m := range ge.buildResolver().Breakdown("production_all") {
		if m.Source == storageBoostSource {
			if m.Op != OpAdd {
				t.Errorf("the storage source is %v, want an additive one", m.Op)
			}
			found += m.Value
		}
	}
	if math.Abs(found-0.07) > 1e-12 {
		t.Errorf("the storage source gives %v to the all-production pool, want 0.07 (7 copies)", found)
	}
	if !slices.ContainsFunc(ge.GetState().Modifiers, func(m Modifier) bool {
		return m.Source == storageBoostSource && m.Target == "production_all"
	}) {
		t.Error("the state's modifier list, which the stats panel reads, does not carry the storage source")
	}
	if ge.buildResolver().AddTotal("production_all") != found {
		t.Errorf("the pool holds %v, want only the storage source's %v", ge.buildResolver().AddTotal("production_all"), found)
	}
}

// TestNoStorageBuildingHoldsAnything: no storage building has a storage
// effect, and a town of them only boosts production.
func TestNoStorageBuildingHoldsAnything(t *testing.T) {
	n := 0
	for _, def := range config.BaseBuildings() {
		if def.Category != "storage" {
			continue
		}
		n++
		for _, e := range def.Effects {
			if e.Type == "storage" {
				t.Errorf("%s has a storage effect", def.Key)
			}
		}
	}
	if n == 0 {
		t.Fatal("no storage buildings")
	}
}
