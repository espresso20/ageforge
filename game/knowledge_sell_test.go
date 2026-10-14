package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A knowledge building comes with its scholars (see isDestroyable). Selling one
// would lose them for good: an older age's copy can never be rebuilt.

const knowledgeSellRefusal = "Knowledge buildings cannot be sold: once its age has passed, it can never be rebuilt."

// TestSellBuilding_RefusesKnowledgeBuildings: a knowledge building is refused,
// one copy or all of them, both this age's and an older age's, and nothing
// moves: not the copies, not the scholars.
func TestSellBuilding_RefusesKnowledgeBuildings(t *testing.T) {
	cases := []struct {
		key, age string
	}{
		{"elders_hall", "stone_age"},  // this age's knowledge building
		{"story_circle", "stone_age"}, // an older age's knowledge building
	}
	for _, c := range cases {
		if d := config.BuildingByKey()[c.key]; d.Scholars <= 0 {
			t.Fatalf("setup: %s has no scholars", c.key)
		}
		ge := catEngine(t, c.age, 1)
		ge.Buildings.counts[c.key] = 3
		ge.recalculateRates()
		scholars := ge.scholarTotal()
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
			if got := ge.scholarTotal(); got != scholars {
				t.Errorf("scholars %v -> %v after a refused sale of %s", scholars, got, c.key)
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

// Knowledge buildings keep their tier, as storage buildings do: the upgrade
// command refuses them with the sentence any building without an upgrade gets,
// and neither it nor an age advance changes their copies or the store.
//
// Before this, an advance offered every lineage building its next tier (but
// storage) and `upgrade` took the offer; the per-copy figures do not rise by
// tier, so upgrading 4 Natural Philosophy Halls to Research Institutes took
// the knowledge store from 3.36e7 to 3.64e6.
func TestUpgradeKnowledgeBuilding_IsRefused(t *testing.T) {
	cases := []struct{ age, from, to string }{
		{"stone_age", "story_circle", "elders_hall"},
		{"colonial_age", "natural_philosophy_hall", "research_institute"},
	}
	for _, c := range cases {
		ge := catEngine(t, c.age, 1)
		ge.Buildings.counts[c.from] = 4
		// Even an offer an older game made stands for nothing.
		ge.Buildings.pendingUpgrades[c.from] = c.to
		ge.recalculateRates()
		for res := range ge.Resources.GetAllStorage() {
			ge.Resources.AddStorage(res, 1e12)
			ge.Resources.Add(res, 1e12)
		}
		scholars := ge.scholarTotal()
		err := ge.UpgradeBuilding(c.from, 0, true)
		def := config.BuildingByKey()[c.from]
		want := def.Name + " has no upgrade this age. Type 'upgrade' to list the ones that do."
		if err == nil || err.Error() != want {
			t.Errorf("upgrade %s: err = %v, want %q", c.from, err, want)
		}
		ge.recalculateRates()
		if got := ge.Buildings.GetCount(c.from); got != 4 {
			t.Errorf("%s: %d copies after the refused upgrade, want 4", c.from, got)
		}
		if got := ge.Buildings.GetCount(c.to); got != 0 {
			t.Errorf("%s: %d copies after the refused upgrade, want 0", c.to, got)
		}
		if got := ge.scholarTotal(); got != scholars {
			t.Errorf("scholars %v -> %v after the refused upgrade", scholars, got)
		}
	}
}
