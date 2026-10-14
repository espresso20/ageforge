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

// stashTown is a Primitive Age game holding n Stashes and a full wood store.
func stashTown(t *testing.T, n int) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	ge.SetTownForTest(map[string]int{"stash": n}, 0)
	if got, want := ge.Resources.GetStorage("wood"), 50+100*float64(n); got != want {
		t.Fatalf("setup: %d Stashes hold %v wood, want %v", n, got, want)
	}
	setAmount(ge, "wood", ge.Resources.GetStorage("wood"))
	return ge
}

// The report (card Pi3AczMR): five Stashes hold 550 wood, the sixth costs 574,
// and the refusal sent the player to "Build more storage first: Stash", the
// building it was refusing. The sixth Stash really cannot be bought, and the
// tables promise nothing past the fifth (the gate asks five, and each of
// those fits), so the fault was the words and the list: a storage building
// that has run out of copies must say so and name nothing to build.
func TestSixthStashIsAtItsLimitAndSendsNobodyToTheStash(t *testing.T) {
	ge := stashTown(t, 5)
	err := ge.BuildBuilding("stash")
	if err == nil || !strings.Contains(err.Error(), "Stash costs 574 wood, more than your wood storage holds (550).") {
		t.Fatalf("build stash = %v, want the 574 against 550 refusal", err)
	}
	if strings.Contains(err.Error(), "Build more storage first") {
		t.Errorf("the refusal sends the player back to the Stash: %q", err)
	}
	if !strings.Contains(err.Error(), "at its limit") || !strings.Contains(err.Error(), "The next age has a bigger storage building.") {
		t.Errorf("the refusal does not say the storage is at its limit and where bigger storage comes from: %q", err)
	}
	if !ge.GetState().Buildings["stash"].StoreLimit {
		t.Error("the Stash state does not carry the store limit")
	}
	if n, err2 := ge.BuildMultiple("stash", 3); n != 0 || err2 == nil || err2.Error() != err.Error() {
		t.Errorf("build stash 3 = (%d, %v), want the same refusal", n, err2)
	}
	// The gate asked five Stashes and the town has them: it can advance.
	// (Its other gate buildings are not staged here; the point is that the
	// refusal above is not what stands between a town and the Stone Age,
	// which TestStoneAgeIsReachableBuyingExactlyTheGate plays out.)

	// A plan item for it says why it waits.
	if _, err := ge.PlanAddBuild("stash", 1); err != nil {
		t.Fatal(err)
	}
	if v := ge.planViews(); len(v) != 1 || v[0].Status != PlanStatusBlocked || v[0].Note != "at its storage limit" {
		t.Errorf("plan view = %+v, want blocked: at its storage limit", v)
	}
}

// Every copy up to the fifth is bought, and the fifth is never told it is at
// a limit: the limit shows only once the sixth is the next copy, counting the
// copies still going up.
func TestStashLimitShowsOnlyAfterTheFifthCopy(t *testing.T) {
	for n := 0; n <= 5; n++ {
		ge := stashTown(t, n)
		limit := ge.GetState().Buildings["stash"].StoreLimit
		if want := n == 5; limit != want {
			t.Errorf("%d Stashes built: StoreLimit = %v, want %v", n, limit, want)
		}
		if n < 5 {
			if err := ge.BuildBuilding("stash"); err != nil {
				t.Errorf("copy %d of the Stash refused: %v", n+1, err)
			}
		}
	}
	// Four built and the fifth going up: the sixth is already out of reach.
	ge := stashTown(t, 4)
	if err := ge.BuildBuilding("stash"); err != nil {
		t.Fatal(err)
	}
	err := ge.BuildBuilding("stash")
	if err == nil || !strings.Contains(err.Error(), "at its limit") {
		t.Errorf("sixth Stash with the fifth in the queue = %v, want the limit", err)
	}
	// One built, the second still waiting on the store the first will raise:
	// wait, it is not a limit.
	ge = newSeededEngine(1)
	setAmount(ge, "wood", 50)
	if err := ge.BuildBuilding("stash"); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "wood", 50)
	err = ge.BuildBuilding("stash")
	if err == nil || strings.Contains(err.Error(), "at its limit") || strings.Contains(err.Error(), "Build more storage first") ||
		!strings.Contains(err.Error(), "already under construction will raise it") {
		t.Errorf("second Stash with the first in the queue = %v, want a wait for the first", err)
	}
}

// A building other than storage priced over a store that can still grow is
// told to build storage; once the storage building has no copy left it is told
// the storage is at its limit instead.
func TestOtherBuildingsAreSentToStorageOnlyWhileStorageCanGrow(t *testing.T) {
	for _, tc := range []struct {
		stashes   int
		wantBuild bool
	}{{2, true}, {5, false}} {
		ge := stashTown(t, tc.stashes)
		ge.Buildings.counts["shrine"] = 25 // priced over any store a Primitive town can build
		cost := ge.Buildings.GetCost("shrine")["wood"]
		if store := ge.Resources.GetStorage("wood"); cost <= store {
			t.Skipf("setup: shrine #26 (%v) fits the store (%v)", cost, store)
		}
		err := ge.BuildBuilding("shrine")
		if err == nil || !strings.Contains(err.Error(), "more than your") {
			t.Fatalf("%d Stashes: build shrine = %v", tc.stashes, err)
		}
		if got := strings.Contains(err.Error(), "Build more storage first: Stash."); got != tc.wantBuild {
			t.Errorf("%d Stashes: refusal %q names the Stash = %v, want %v", tc.stashes, err, got, tc.wantBuild)
		}
	}
}
