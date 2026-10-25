package game

import (
	"sort"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Storage is permanent, like wonders: no catastrophe takes it and it can't be
// sold. The age lock never lets an older age's storage be rebuilt, and an
// age's first storage copy can cost more than the storage left after a loss,
// so a lost copy could leave a cap that never rises again. The nightly smoke
// found it: seed 1's second run entered the Victorian Age with 478M of
// storage, a Nuclear Exchange took both Industrial Depots (and 11 older
// storage copies), and the 130M left could never hold a Victorian Vault
// (about 210M steel) or anything the age asked for.

// storageKeys returns every storage building key, sorted.
func storageKeys() []string {
	var out []string
	for k, d := range config.BuildingByKey() {
		if d.Category == "storage" {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func storageCounts(ge *GameEngine) map[string]int {
	out := make(map[string]int)
	for _, k := range storageKeys() {
		out[k] = ge.Buildings.GetCount(k)
	}
	return out
}

func TestDestroyable_SparesWondersStorageAndKnowledgeBuildings(t *testing.T) {
	spared := 0
	byKey := config.BuildingByKey()
	for _, k := range sortedKeys(byKey) {
		d := byKey[k]
		want := d.Category != "wonder" && d.Category != "storage" && !config.IsKnowledgeBuilding(d)
		if got := isDestroyable(d); got != want {
			t.Errorf("isDestroyable(%s, %s) = %v, want %v", k, d.Category, got, want)
		}
		if config.IsKnowledgeBuilding(d) {
			spared++
		}
	}
	if spared == 0 {
		t.Fatal("no knowledge building found: the test proves nothing")
	}
}

// The knowledge buildings come with the scholars: a catastrophe leaves every
// copy standing, so the scholars stay.
func TestEndure_NeverDestroysKnowledgeBuildings(t *testing.T) {
	ge := catEngine(t, "iron_age", 1)
	know := "monastery_library"
	food := pickWorkerBuildings(t, "food")[0]
	ge.Buildings.counts[know] = 10
	ge.Buildings.counts[food] = 1
	ge.recalculateRates()
	scholars := ge.scholarTotal()
	if got := ge.Buildings.DestroyableCount(); got != 1 {
		t.Fatalf("DestroyableCount = %d, want 1 (knowledge buildings don't count)", got)
	}
	ge.pendingCatastrophe = "iron_era"
	if err := ge.Endure(); err != nil {
		t.Fatalf("Endure: %v", err)
	}
	ge.recalculateRates()
	if got := ge.Buildings.GetCount(know); got != 10 {
		t.Errorf("Endure took knowledge buildings: %s 10 -> %d", know, got)
	}
	if got := ge.scholarTotal(); got != scholars {
		t.Errorf("Endure moved the scholars: %v -> %v", scholars, got)
	}
}

// With almost nothing but storage standing, the old pool could only destroy
// storage; now the one other building is all a catastrophe can take.
func TestEndure_NeverDestroysStorage(t *testing.T) {
	ge := catEngine(t, "iron_age", 1)
	for _, k := range []string{"stash", "storage_pit", "warehouse", "granary"} {
		ge.Buildings.counts[k] = 10
	}
	food := pickWorkerBuildings(t, "food")[0]
	ge.Buildings.counts[food] = 1
	before := storageCounts(ge)
	if got := ge.Buildings.DestroyableCount(); got != 1 {
		t.Fatalf("DestroyableCount = %d, want 1 (storage doesn't count)", got)
	}
	ge.pendingCatastrophe = "iron_era"
	if err := ge.Endure(); err != nil {
		t.Fatalf("Endure: %v", err)
	}
	for k, n := range before {
		if got := ge.Buildings.GetCount(k); got != n {
			t.Errorf("Endure took storage: %s %d -> %d", k, n, got)
		}
	}
	if got := ge.Buildings.GetCount(food); got != 0 {
		t.Errorf("%s count = %d, want 0 (the only destroyable building)", food, got)
	}
}

func TestGreatFire_SparesStorage(t *testing.T) {
	ge := catEngine(t, "medieval_age", 1)
	for _, k := range []string{"stash", "storage_pit", "warehouse", "granary", "classical_vault", "keep"} {
		ge.Buildings.counts[k] = 5
	}
	food := pickWorkerBuildings(t, "food")[0]
	ge.Buildings.counts[food] = 3
	before := storageCounts(ge)
	ge.applyChallengingEpochEvent(config.EpochEventByKey()["the_great_fire"], ge.currentEpoch)
	for k, n := range before {
		if got := ge.Buildings.GetCount(k); got != n {
			t.Errorf("the Great Fire burned storage: %s %d -> %d", k, n, got)
		}
	}
	if got := ge.Buildings.GetCount(food); got != 0 {
		t.Errorf("%s count = %d, want 0 (the fire takes up to 8, and only those 3 can burn)", food, got)
	}
}

func TestSuccumbRuins_NeverStorage(t *testing.T) {
	ge := catEngine(t, "iron_age", 1)
	for _, k := range []string{"stash", "storage_pit", "warehouse", "granary"} {
		ge.Buildings.counts[k] = 10
	}
	food := pickWorkerBuildings(t, "food")[0]
	ge.Buildings.counts[food] = 2
	ruins := ge.Buildings.GenerateRuins(ge.gameRNG(), SuccumbRuinCount)
	byKey := config.BuildingByKey()
	for k, n := range ruins {
		if byKey[k].Category == "storage" {
			t.Errorf("Succumb ruined %d %s: storage ruins produce nothing and storage is permanent", n, k)
		}
	}
	if ruins[food] != 2 {
		t.Errorf("ruins = %v, want the 2 %s (the only destroyable buildings)", ruins, food)
	}
}

func TestSellBuilding_RefusesStorage(t *testing.T) {
	ge := catEngine(t, "stone_age", 1)
	ge.Buildings.counts["stash"] = 10      // an older age's storage
	ge.Buildings.counts["storage_pit"] = 4 // this age's
	food := pickWorkerBuildings(t, "food")[0]
	ge.Buildings.counts[food] = 2
	for _, k := range []string{"stash", "storage_pit"} {
		before := ge.Buildings.GetCount(k)
		if err := ge.SellBuilding(k, 1); err == nil {
			t.Errorf("selling %s succeeded; storage must not be sellable", k)
		}
		if got := ge.Buildings.GetCount(k); got != before {
			t.Errorf("%s count %d -> %d after a refused sale", k, before, got)
		}
	}
	if err := ge.SellBuilding(food, 1); err != nil {
		t.Errorf("selling %s (not storage): %v", food, err)
	}
}
