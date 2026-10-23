package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The Primitive Age gate, played on the real engine through the same commands
// a player types: a new game, stores topped up to their caps each step (the
// store is the limit, never raised by the test), every building and the
// wonder bought exactly as the gate asks, in an order that buys storage first.
// No purchase may be refused as over a store, and the advance must go through.
func TestStoneAgeIsReachableBuyingExactlyTheGate(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	ge := NewGameEngine()
	ge.SeedRNG(1)
	age := config.Ages()[1]
	wonder := ge.progress.WonderForAge(ge.age)

	fill := func() {
		ge.mu.Lock()
		for res := range ge.Resources.GetAll() {
			ge.Resources.Add(res, ge.Resources.GetStorage(res))
		}
		ge.mu.Unlock()
	}
	owned := func(k string) int { return ge.Buildings.GetCount(k) + queued(ge, k) }
	// Storage first, then the rest, cheapest keys last.
	order := []string{"stash", "hut", "story_circle", "gathering_camp", "wood_camp", "shrine"}
	for step := 0; step < 4000; step++ {
		fill()
		for _, k := range order {
			for owned(k) < age.BuildingReqs[k] {
				err := ge.BuildBuilding(k)
				if err != nil && strings.Contains(err.Error(), "more than your") && queued(ge, "stash") == 0 {
					// Nothing is on its way to raise the store: a dead end.
					t.Fatalf("step %d: %s #%d refused with no storage building coming: %v", step, k, owned(k)+1, err)
				}
				if err != nil {
					break // waiting for the stash in the queue, or for stock
				}
			}
		}
		if wonder != "" && ge.Buildings.GetCount(wonder)+queued(ge, wonder) == 0 {
			for _, res := range []string{"wood", "food"} {
				_, _ = ge.BankWonderMax(wonder, res)
			}
			if err := ge.BuildBuilding(wonder); err != nil && strings.Contains(err.Error(), "more than your") {
				t.Fatalf("step %d: wonder refused: %v", step, err)
			}
		}
		ge.StepTicks(5)
		if err := ge.AdvanceAge(); err == nil {
			if got := ge.GetState().Age; got != age.Key {
				t.Fatalf("advanced to %s, want %s", got, age.Key)
			}
			return
		}
	}
	st := ge.GetState()
	for k, n := range age.BuildingReqs {
		t.Logf("%s: have %d of %d", k, st.Buildings[k].Count, n)
	}
	t.Fatalf("never reached the Stone Age; wood store %v, advance: %v", st.Resources["wood"].Storage, ge.AdvanceAge())
}

// woodBase is the wood store a new game holds; stashSize is what one Stash adds.
func woodBase() float64 { return config.ResourceByKey()["wood"].BaseStorage }

func stashSize() float64 {
	for _, e := range config.BuildingByKey()["stash"].Effects {
		if e.Type == "storage" && e.Target == "all" {
			return e.Value
		}
	}
	return 0
}

// stashTown is a Primitive Age game holding n Stashes and a full wood store.
func stashTown(t *testing.T, n int) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	ge.SetTownForTest(map[string]int{"stash": n}, 0)
	flat := woodBase() + stashSize()*float64(n)
	next, _ := ge.Buildings.BuildBatchCost("stash", 1, ge.buildQueue)
	want := flat
	if n > 0 && next["wood"] > want {
		want = next["wood"]
	}
	if got := ge.Resources.GetStorage("wood"); got != want {
		t.Fatalf("setup: %d Stashes hold %v wood, want %v", n, got, want)
	}
	setAmount(ge, "wood", ge.Resources.GetStorage("wood"))
	return ge
}

func TestStorageNextCopyAlwaysFits(t *testing.T) {
	ge := stashTown(t, 5)
	for copy := 6; copy <= 7; copy++ {
		cost, _ := ge.Buildings.BuildBatchCost("stash", 1, ge.buildQueue)
		if store := ge.Resources.GetStorage("wood"); store < cost["wood"] {
			t.Fatalf("Stash #%d costs %v wood, store holds %v", copy, cost["wood"], store)
		}
		setAmount(ge, "wood", ge.Resources.GetStorage("wood"))
		if err := ge.BuildBuilding("stash"); err != nil {
			t.Fatalf("build Stash #%d: %v", copy, err)
		}
		queuedNext, _ := ge.Buildings.BuildBatchCost("stash", 1, ge.buildQueue)
		if store := ge.Resources.GetStorage("wood"); store < queuedNext["wood"] {
			t.Fatalf("with Stash #%d queued, next costs %v wood but store holds %v", copy, queuedNext["wood"], store)
		}
		for queued(ge, "stash") > 0 {
			ge.StepTicks(1)
		}
		next, _ := ge.Buildings.BuildBatchCost("stash", 1, ge.buildQueue)
		if store := ge.Resources.GetStorage("wood"); store < next["wood"] {
			t.Fatalf("after Stash #%d, next costs %v wood but store holds %v", copy, next["wood"], store)
		}
	}
}

func TestStoragePriceFloorOnlyLiftsWhenNeeded(t *testing.T) {
	ge := stashTown(t, 1)
	if got, want := ge.Resources.GetStorage("wood"), woodBase()+stashSize(); got != want {
		t.Errorf("one Stash store = %v, want flat bonus %v", got, want)
	}

	ge = newSeededEngine(1)
	if got, want := ge.Resources.GetStorage("wood"), woodBase(); got != want {
		t.Errorf("no Stash store = %v, want base store %v", got, want)
	}
}

// The last storage tier has no successor in the tables. Its own next copy now
// fits, so the storage-limit refusal itself is unreachable.
func TestLastStorageTierHasNoSuccessor(t *testing.T) {
	ages, defs := config.Ages(), config.BaseBuildings()
	for i := range ages {
		if i+1 >= len(ages) {
			break
		}
		has := config.AgeHasStorageBuilding(defs, ages[i+1].Key)
		if want := ages[i+1].Key != "transcendent_age"; has != want {
			t.Errorf("AgeHasStorageBuilding(%s) = %v, want %v", ages[i+1].Key, has, want)
		}
	}
	var storage string
	for _, d := range defs {
		if d.Category == "storage" && d.RequiredAge == "quantum_age" {
			storage = d.Key
		}
	}
	ge := newSeededEngine(1)
	if err := ge.EnterAgeForTest("quantum_age"); err != nil {
		t.Fatal(err)
	}
	ge.SetTownForTest(map[string]int{storage: 60}, 0)
	cost, _ := ge.Buildings.BuildBatchCost(storage, 1, ge.buildQueue)
	for res, amount := range cost {
		if store := ge.Resources.GetStorage(res); store < amount {
			t.Errorf("next %s costs %v %s but store holds %v", storage, amount, res, store)
		}
	}
}
