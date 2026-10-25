package game

import (
	"maps"
	"math/rand"
	"testing"
)

// destroyTown is a town of seven destroyable building types, a storage
// building (spared) and a count with no building behind it.
func destroyTown() *BuildingManager {
	bm := NewBuildingManager()
	for key, n := range map[string]int{
		"advanced_alloy_plant": 2, "aerospace_foundry": 3, "agri_complex": 4,
		"agricultural_complex": 5, "agricultural_works": 6, "altar": 7, "amphitheater": 8,
		"atomic_vault": 3, "no_such_building": 5,
	} {
		bm.counts[key] = n
	}
	return bm
}

// The pool a catastrophe draws from is every destroyable copy, in sorted key
// order, so a seed picks the same buildings every time. The picks below are
// what the seeds gave when the pool was built by hand; they must not move.
func TestDestroyRandomPicksAreFixedBySeed(t *testing.T) {
	bm := destroyTown()
	if got, want := bm.DestroyableCount(), 35; got != want {
		t.Fatalf("DestroyableCount = %d, want %d", got, want)
	}
	if got := len(bm.destroyablePool()); got != 35 {
		t.Fatalf("pool holds %d, want 35", got)
	}
	destroyed, names := bm.DestroyRandom(rand.New(rand.NewSource(99)), 6)
	wantDestroyed := map[string]int{"agri_complex": 1, "agricultural_works": 1, "altar": 1, "amphitheater": 3}
	if !maps.Equal(destroyed, wantDestroyed) {
		t.Errorf("destroyed %v, want %v", destroyed, wantDestroyed)
	}
	wantNames := []string{"1 Agritech Campus", "1 Agricultural Works", "1 Altar", "3 Amphitheaters"}
	if len(names) != len(wantNames) {
		t.Fatalf("names %v, want %v", names, wantNames)
	}
	for i := range names {
		if names[i] != wantNames[i] {
			t.Errorf("names %v, want %v", names, wantNames)
			break
		}
	}
	wantCounts := map[string]int{
		"advanced_alloy_plant": 2, "aerospace_foundry": 3, "agri_complex": 3,
		"agricultural_complex": 5, "agricultural_works": 5, "altar": 6, "amphitheater": 5,
		"atomic_vault": 3, "no_such_building": 5,
	}
	if got := bm.GetAll(); !maps.Equal(got, wantCounts) {
		t.Errorf("counts after %v, want %v", got, wantCounts)
	}

	ruins := bm.GenerateRuins(rand.New(rand.NewSource(5)), 4)
	wantRuins := map[string]int{"aerospace_foundry": 1, "agricultural_complex": 1, "altar": 1, "amphitheater": 1}
	if !maps.Equal(ruins, wantRuins) {
		t.Errorf("ruins %v, want %v", ruins, wantRuins)
	}
	if bm.counts["altar"] != 5 || bm.counts["atomic_vault"] != 3 || bm.counts["no_such_building"] != 5 {
		t.Errorf("counts after ruins %v", bm.GetAll())
	}
}

// Asking for more than the town holds takes every destroyable copy and
// leaves no count below zero.
func TestDestroyRandomTakesNoMoreThanStands(t *testing.T) {
	bm := destroyTown()
	destroyed, names := bm.DestroyRandom(rand.New(rand.NewSource(1)), 1000)
	total := 0
	for key, n := range destroyed {
		total += n
		if bm.counts[key] != 0 {
			t.Errorf("%s left at %d", key, bm.counts[key])
		}
	}
	if total != 35 || len(names) != 7 {
		t.Errorf("destroyed %d copies in %d lines, want 35 in 7", total, len(names))
	}
	if bm.counts["atomic_vault"] != 3 || bm.counts["no_such_building"] != 5 {
		t.Errorf("a spared count moved: %v", bm.GetAll())
	}
	if again, lines := bm.DestroyRandom(rand.New(rand.NewSource(1)), 5); again != nil || lines != nil {
		t.Errorf("an empty town lost %v (%v)", again, lines)
	}
}

// The copies the managers hand out are never nil, even when nothing is held.
func TestBuildingCopiesAreNeverNil(t *testing.T) {
	bm := NewBuildingManager()
	if bm.GetAll() == nil {
		t.Error("GetAll is nil")
	}
	if bank := bm.GetWonderBank("no_such_wonder"); bank == nil || len(bank) != 0 {
		t.Errorf("bank of an unknown wonder = %#v, want an empty map", bank)
	}
	bm.wonderBanks["colosseum"] = nil
	bm.wonderBanks["pyramids"] = map[string]float64{"stone": 4}
	banks := bm.GetWonderBanks()
	if banks == nil || banks["colosseum"] == nil || len(banks["colosseum"]) != 0 {
		t.Errorf("banks = %#v, want an empty map for the empty bank", banks)
	}
	banks["pyramids"]["stone"] = 99
	if bm.wonderBanks["pyramids"]["stone"] != 4 {
		t.Error("a copy of the banks shares storage with the manager")
	}
}

// UpgradeRoom is the room under MaxCount, queued copies counted, never
// negative, and the whole request for a building with no cap.
func TestUpgradeRoom(t *testing.T) {
	bm := NewBuildingManager()
	capped, free := "", ""
	for _, key := range sortedKeys(bm.defs) {
		switch def := bm.defs[key]; {
		case def.MaxCount > 0 && capped == "":
			capped = key
		case def.MaxCount <= 0 && free == "":
			free = key
		}
	}
	if capped == "" || free == "" {
		t.Fatal("no capped or uncapped building to test with")
	}
	limit := bm.defs[capped].MaxCount
	for _, have := range []int{0, limit - 1, limit, limit + 4} {
		for queued := 0; queued <= 2; queued++ {
			for _, n := range []int{0, 1, limit, limit + 3} {
				queue := make([]BuildQueueItem, queued)
				for i := range queue {
					queue[i].BuildingKey = capped
				}
				queue = append(queue, BuildQueueItem{BuildingKey: free})
				bm.counts[capped] = have
				want := limit - have - queued
				want = max(want, 0)
				if n < want {
					want = n
				}
				if got := bm.UpgradeRoom(capped, n, queue); got != want {
					t.Errorf("%d held, %d queued, %d asked (cap %d): room %d, want %d", have, queued, n, limit, got, want)
				}
			}
		}
	}
	if got := bm.UpgradeRoom(free, 7, nil); got != 7 {
		t.Errorf("an uncapped building gave room %d, want 7", got)
	}
	if got := bm.UpgradeRoom("no_such_building", 7, nil); got != 7 {
		t.Errorf("an unknown building gave room %d, want 7", got)
	}
}
