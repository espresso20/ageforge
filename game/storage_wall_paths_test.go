package game

import (
	"strings"
	"testing"
)

// The storage wall's one refusal: every way of buying a building or an
// upgrade asks overStore first and answers with storeTooSmallText, so a
// price larger than a store reads the same whichever command was typed.

// wallShrineEngine is a Primitive Age game whose wood store (50) is full and
// smaller than a Shrine's price (60).
func wallShrineEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	if cost, store := ge.Buildings.GetCost("shrine")["wood"], ge.Resources.GetStorage("wood"); cost <= store {
		t.Fatalf("setup: a shrine (%v wood) fits the wood store (%v)", cost, store)
	}
	setAmount(ge, "wood", ge.Resources.GetStorage("wood"))
	return ge
}

// wallUpgradeEngine is a Stone Age game holding three Wood Camps, whose
// upgrade to Woodcutter Camps costs more wood than the (full) store holds.
func wallUpgradeEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	ge.Buildings.counts["wood_camp"] = 3
	ge.mu.Lock()
	ge.advanceAge("stone_age")
	ge.mu.Unlock()
	if to, ok := ge.Buildings.GetPendingUpgrade("wood_camp"); !ok || to != "woodcutter_camp" {
		t.Fatalf("setup: wood_camp upgrades to %q (%v)", to, ok)
	}
	one, _ := ge.Buildings.UpgradeCost("wood_camp", "woodcutter_camp", 1)
	if res := ge.overStore(one); res == "" {
		t.Fatalf("setup: one upgrade costs %v and fits every store", one)
	}
	for res := range one {
		setAmount(ge, res, ge.Resources.GetStorage(res))
	}
	return ge
}

func TestWall_EveryPurchasePathGivesTheSameRefusal(t *testing.T) {
	ge := wallShrineEngine(t)
	want := ge.BuildBuilding("shrine")
	if want == nil || !strings.Contains(want.Error(), "more than your wood storage holds") {
		t.Fatalf("build shrine = %v, want the store-too-small refusal", want)
	}
	buyPaths := []struct {
		name string
		buy  func(ge *GameEngine) (int, error)
	}{
		{"build", func(ge *GameEngine) (int, error) { return 0, ge.BuildBuilding("shrine") }},
		{"build 1", func(ge *GameEngine) (int, error) { return ge.BuildMultiple("shrine", 1) }},
		{"build 5", func(ge *GameEngine) (int, error) { return ge.BuildMultiple("shrine", 5) }},
		{"build max", func(ge *GameEngine) (int, error) { return ge.BuildMultiple("shrine", 10000) }},
	}
	for _, p := range buyPaths {
		t.Run(p.name, func(t *testing.T) {
			ge := wallShrineEngine(t)
			n, err := p.buy(ge)
			if err == nil || err.Error() != want.Error() {
				t.Errorf("error = %v, want %q", err, want)
			}
			if n != 0 || ge.Buildings.GetCount("shrine") != 0 || queued(ge, "shrine") != 0 {
				t.Errorf("bought %d, have %d, queued %d: nothing may be bought", n, ge.Buildings.GetCount("shrine"), queued(ge, "shrine"))
			}
			if got := ge.Resources.Get("wood"); got != ge.Resources.GetStorage("wood") {
				t.Errorf("wood = %v after the refusal, want the store untouched (%v)", got, ge.Resources.GetStorage("wood"))
			}
		})
	}

	t.Run("plan", func(t *testing.T) {
		ge := wallShrineEngine(t)
		if _, err := ge.PlanAddBuild("shrine", 1); err != nil {
			t.Fatal(err)
		}
		ge.runPlanTick()
		if ge.Buildings.GetCount("shrine")+queued(ge, "shrine") != 0 {
			t.Error("the plan bought a shrine priced over the store")
		}
		v := ge.planViews()
		if len(v) != 1 || v[0].Status != PlanStatusBlocked || v[0].Note != "needs more wood storage" {
			t.Errorf("plan view = %+v, want blocked: needs more wood storage", v)
		}
	})

	for _, p := range []struct {
		name string
		all  bool
		n    int
	}{{"upgrade 1", false, 1}, {"upgrade all", true, 0}} {
		t.Run(p.name, func(t *testing.T) {
			ge := wallUpgradeEngine(t)
			err := ge.UpgradeBuilding("wood_camp", p.n, p.all)
			if err == nil || !strings.Contains(err.Error(), "more than your") || !strings.HasPrefix(err.Error(), "Upgrading ") {
				t.Fatalf("upgrade = %v, want the store-too-small refusal", err)
			}
			if ge.Buildings.GetCount("wood_camp") != 3 || ge.Buildings.GetCount("woodcutter_camp") != 0 {
				t.Error("an upgrade the store cannot hold went through")
			}
		})
	}
}

// A wonder is the one price allowed over a store: it is paid through its
// bank, so `build <wonder> <n>` takes the single path and its rules, not the
// store check and not a second payment from stock.
func TestWall_BuildMultipleOnAWonderTakesTheBankPath(t *testing.T) {
	ge := newSeededEngine(1)
	w := ge.progress.WonderForAge(ge.age)
	if w == "" {
		t.Skip("setup: the age has no wonder")
	}
	want := ge.BuildBuilding(w)
	for _, n := range []int{1, 3, 10000} {
		built, err := ge.BuildMultiple(w, n)
		if built != 0 || err == nil || err.Error() != want.Error() {
			t.Errorf("build %s %d = (%d, %v), want (0, %v)", w, n, built, err, want)
		}
		if err != nil && strings.Contains(err.Error(), "more than your") {
			t.Errorf("build %s %d gave the store refusal for a wonder: %v", w, n, err)
		}
	}
	// With the bank full it builds one wonder and takes nothing from stock.
	ge.Buildings.wonderBanks[w] = ge.Buildings.GetCost(w)
	before := ge.Resources.GetAll()
	built, err := ge.BuildMultiple(w, 5)
	if err != nil || built != 1 {
		t.Fatalf("build %s 5 with its bank full = (%d, %v), want (1, nil)", w, built, err)
	}
	if got := ge.Buildings.GetCount(w) + queued(ge, w); got != 1 {
		t.Errorf("%d wonders built or queued, want 1", got)
	}
	for res, v := range ge.Resources.GetAll() {
		if v != before[res] {
			t.Errorf("%s = %v, was %v: a banked wonder takes nothing from stock", res, v, before[res])
		}
	}
}

// `max` buys what fits under the stock and the store, and the same copies a
// run of single builds would: it neither stops early nor goes over.
func TestWall_MaxBuysWhatFitsAndNoMore(t *testing.T) {
	setup := func() *GameEngine {
		ge := newSeededEngine(1)
		// Park the hut count one copy short of the store: the next copy is
		// the last one that fits, the one after it is over.
		n := 0
		for ; n < 400; n++ {
			if ge.Buildings.NextCost("hut", "wood", nil) > ge.Resources.GetStorage("wood") {
				break
			}
			ge.Buildings.counts["hut"]++
		}
		if n == 400 {
			t.Fatal("setup: a hut never outgrew the wood store")
		}
		ge.Buildings.counts["hut"]--
		setAmount(ge, "wood", ge.Resources.GetStorage("wood"))
		return ge
	}
	byMax, byOne := setup(), setup()
	got, err := byMax.BuildMultiple("hut", 10000)
	if err != nil {
		t.Fatalf("max: %v", err)
	}
	one := 0
	for byOne.BuildBuilding("hut") == nil {
		one++
	}
	if got != one || got < 1 {
		t.Errorf("max bought %d, single builds bought %d: want the same, and at least one", got, one)
	}
	if queued(byMax, "hut") != queued(byOne, "hut") || byMax.Resources.Get("wood") != byOne.Resources.Get("wood") {
		t.Errorf("max left %d queued and %v wood, single builds %d and %v",
			queued(byMax, "hut"), byMax.Resources.Get("wood"), queued(byOne, "hut"), byOne.Resources.Get("wood"))
	}
	// And now the next copy is over the store: max says so, as a build does.
	_, errMax := byMax.BuildMultiple("hut", 10000)
	errOne := byMax.BuildBuilding("hut")
	if errMax == nil || errOne == nil || errMax.Error() != errOne.Error() {
		t.Errorf("max = %v, build = %v: want one refusal", errMax, errOne)
	}
}
