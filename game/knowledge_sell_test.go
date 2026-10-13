package game

import (
	"strings"
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
			if res != "knowledge" {
				ge.Resources.AddStorage(res, 1e12)
			}
			ge.Resources.Add(res, 1e12)
		}
		store := ge.Resources.GetStorage("knowledge")
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
		if got := ge.Resources.GetStorage("knowledge"); got != store {
			t.Errorf("knowledge store %v -> %v after the refused upgrade", store, got)
		}
	}
}

// TestKnowledgeStore_NeverFallsAcrossAdvances walks the real engine through
// four advances. In each age the town builds that age's knowledge building
// and some ordinary ones, then tries every player action that could shrink
// the store (upgrade and sell, on everything it owns), then advances. No
// knowledge building changes tier or count, none is offered an upgrade, the
// ordinary buildings still are (so the transform ran), and the store never
// goes down.
func TestKnowledgeStore_NeverFallsAcrossAdvances(t *testing.T) {
	ages := []string{"medieval_age", "renaissance_age", "colonial_age", "industrial_age", "victorian_age"}
	ge := catEngine(t, ages[0], 1)
	var know []config.BuildingDef
	for _, d := range ge.rules.Buildings() {
		if d.KnowledgeStore > 0 {
			know = append(know, d)
		}
	}
	store := func() float64 {
		ge.recalculateRates()
		return ge.Resources.GetStorage("knowledge")
	}
	last := store()
	offered := 0
	for i, age := range ages {
		for _, d := range know {
			if d.RequiredAge == age {
				ge.Buildings.counts[d.Key] = 4
			}
		}
		// Ordinary lineage buildings of the age whose next tier the coming
		// age has: what the advance offers an upgrade for.
		if i+1 < len(ages) {
			for _, d := range ge.rules.Buildings() {
				if d.RequiredAge != age || d.LineageKey == "" || d.LineageKey == "wonder" || keepsItsTier(d) {
					continue
				}
				if _, ok := ge.rules.NextTier(d.LineageKey, d.LineageTier, ages[i+1]); ok {
					ge.Buildings.counts[d.Key] = 2
				}
			}
		}
		for res := range ge.Resources.GetAllStorage() {
			if res != "knowledge" {
				ge.Resources.AddStorage(res, 1e15)
			}
			ge.Resources.Add(res, 1e15)
		}
		if got := store(); got < last {
			t.Fatalf("%s: the store fell from %v to %v on building", age, last, got)
		}
		last = store()
		counts := map[string]int{}
		for _, d := range know {
			counts[d.Key] = ge.Buildings.GetCount(d.Key)
		}
		// Every player action that could shrink it.
		for _, key := range sortedKeys(ge.Buildings.counts) {
			if ge.Buildings.counts[key] == 0 {
				continue
			}
			def := config.BuildingByKey()[key]
			err := ge.UpgradeBuilding(key, 0, true)
			if def.KnowledgeStore > 0 && (err == nil || !strings.Contains(err.Error(), "has no upgrade this age")) {
				t.Errorf("%s: upgrade %s: %v", age, key, err)
			}
			if def.KnowledgeStore > 0 {
				if err := ge.SellBuilding(key, 1); err == nil {
					t.Errorf("%s: selling %s succeeded", age, key)
				}
			}
		}
		for _, d := range know {
			if got := ge.Buildings.GetCount(d.Key); got != counts[d.Key] {
				t.Errorf("%s: %s %d -> %d by a player action", age, d.Key, counts[d.Key], got)
			}
		}
		if got := store(); got < last {
			t.Fatalf("%s: the store fell from %v to %v by a player action", age, last, got)
		}
		if i+1 == len(ages) {
			break
		}
		ge.advanceAge(ages[i+1])
		ge.age = ages[i+1]
		for _, d := range know {
			if got := ge.Buildings.GetCount(d.Key); got != counts[d.Key] {
				t.Errorf("%s -> %s: %s %d -> %d by the advance", age, ages[i+1], d.Key, counts[d.Key], got)
			}
			if to, ok := ge.Buildings.GetPendingUpgrade(d.Key); ok {
				t.Errorf("%s -> %s: %s is offered an upgrade to %s", age, ages[i+1], d.Key, to)
			}
		}
		for _, k := range sortedKeys(ge.Buildings.counts) {
			if _, ok := ge.Buildings.GetPendingUpgrade(k); ok {
				offered++
			}
		}
		if got := store(); got < last {
			t.Fatalf("%s -> %s: the store fell from %v to %v on the advance", age, ages[i+1], last, got)
		}
	}
	if offered == 0 {
		t.Error("no ordinary building was offered an upgrade on any advance: the walk did not run the transform")
	}
}
