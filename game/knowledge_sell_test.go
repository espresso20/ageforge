package game

import (
	"sort"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A knowledge building is the knowledge store (see isDestroyable). Selling one
// shrinks the store for good: an older age's copy can never be rebuilt, and a
// tech priced over the smaller store could then never be started.

const knowledgeSellRefusal = "Knowledge buildings cannot be sold: once its age has passed, it can never be rebuilt."

// TestSellBuilding_RefusesKnowledgeBuildings: a knowledge building is refused,
// one copy or all of them, both this age's and an older age's, and nothing
// moves: not the copies, not the knowledge store.
func TestSellBuilding_RefusesKnowledgeBuildings(t *testing.T) {
	cases := []struct {
		key, age string
	}{
		{"elders_hall", "stone_age"},  // this age's knowledge building
		{"story_circle", "stone_age"}, // an older age's knowledge building
	}
	for _, c := range cases {
		if d := config.BuildingByKey()[c.key]; d.KnowledgeStore <= 0 {
			t.Fatalf("setup: %s has no knowledge store figure", c.key)
		}
		ge := catEngine(t, c.age, 1)
		ge.Buildings.counts[c.key] = 3
		ge.recalculateRates()
		store := ge.Resources.GetStorage("knowledge")
		for _, n := range []int{1, 3} {
			err := ge.SellBuilding(c.key, n)
			if err == nil {
				t.Errorf("selling %d %s succeeded; knowledge buildings must not be sellable", n, c.key)
				continue
			}
			if err.Error() != knowledgeSellRefusal {
				t.Errorf("selling %s: refusal = %q, want %q", c.key, err.Error(), knowledgeSellRefusal)
			}
			if got := ge.Buildings.GetCount(c.key); got != 3 {
				t.Errorf("%s count = %d after a refused sale of %d, want 3", c.key, got, n)
			}
			ge.recalculateRates()
			if got := ge.Resources.GetStorage("knowledge"); got != store {
				t.Errorf("knowledge store %v -> %v after a refused sale of %s", store, got, c.key)
			}
		}
	}
}

// TestSellBuilding_StillSellsProducers: an ordinary producer still sells, and
// a knowledge building beside it does not change that.
func TestSellBuilding_StillSellsProducers(t *testing.T) {
	ge := catEngine(t, "stone_age", 1)
	food := pickWorkerBuildings(t, "food")[0]
	ge.Buildings.counts[food] = 2
	ge.Buildings.counts["elders_hall"] = 2
	if err := ge.SellBuilding(food, 1); err != nil {
		t.Fatalf("selling %s: %v", food, err)
	}
	if got := ge.Buildings.GetCount(food); got != 1 {
		t.Errorf("%s count = %d after selling 1, want 1", food, got)
	}
}

// TestSellBuilding_StorageKeepsItsRefusal: storage is still refused, with the
// text it had before knowledge buildings joined it.
func TestSellBuilding_StorageKeepsItsRefusal(t *testing.T) {
	ge := catEngine(t, "stone_age", 1)
	ge.Buildings.counts["storage_pit"] = 2
	err := ge.SellBuilding("storage_pit", 1)
	const want = "Storage cannot be sold: once its age has passed, it can never be rebuilt."
	if err == nil || err.Error() != want {
		t.Errorf("selling storage: err = %v, want %q", err, want)
	}
	if got := ge.Buildings.GetCount("storage_pit"); got != 2 {
		t.Errorf("storage_pit count = %d after a refused sale, want 2", got)
	}
}

// TestUpgradeKnowledgeBuilding_RecordsStoreChange records what an upgrade does
// to the knowledge store. It asserts nothing about direction: it logs two real
// upgrades (one up, one down) and, from the tables, every step down a
// knowledge lineage, so the result shows in the test log.
func TestUpgradeKnowledgeBuilding_RecordsStoreChange(t *testing.T) {
	cases := []struct {
		age, from, to string
		count         int
	}{
		{"stone_age", "story_circle", "elders_hall", 4},
		{"colonial_age", "natural_philosophy_hall", "research_institute", 4},
	}
	for _, c := range cases {
		ge := catEngine(t, c.age, 1)
		ge.Buildings.counts[c.from] = c.count
		ge.Buildings.SetPendingUpgrade(c.from, c.to)
		ge.recalculateRates()
		for res := range ge.Resources.GetAllStorage() {
			if res != "knowledge" {
				ge.Resources.AddStorage(res, 1e12)
			}
			ge.Resources.Add(res, 1e12)
		}
		before := ge.Resources.GetStorage("knowledge")
		if err := ge.UpgradeBuilding(c.from, 0, true); err != nil {
			t.Fatalf("UpgradeBuilding %s: %v", c.from, err)
		}
		ge.recalculateRates()
		after := ge.Resources.GetStorage("knowledge")
		t.Logf("upgrade %d %s -> %s (%s): knowledge store %.6g -> %.6g", c.count, c.from, c.to, c.age, before, after)
	}

	byLineage := map[string][]config.BuildingDef{}
	for _, d := range config.BuildingByKey() {
		if d.LineageKey == "knowledge" && d.Category != "wonder" {
			byLineage[d.LineageKey] = append(byLineage[d.LineageKey], d)
		}
	}
	for _, defs := range byLineage {
		sort.Slice(defs, func(i, j int) bool { return defs[i].LineageTier < defs[j].LineageTier })
		for i := 1; i < len(defs); i++ {
			if defs[i].KnowledgeStore < defs[i-1].KnowledgeStore {
				t.Logf("table: %s (tier %d, %.6g each) -> %s (tier %d, %.6g each) lowers the store per copy",
					defs[i-1].Key, defs[i-1].LineageTier, defs[i-1].KnowledgeStore,
					defs[i].Key, defs[i].LineageTier, defs[i].KnowledgeStore)
			}
		}
	}
}
