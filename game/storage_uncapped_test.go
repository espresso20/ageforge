package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Stock has no ceiling. Every way stock arrives keeps the whole amount, no
// building is refused for costing more than a town once stored, and nothing
// overflows into a wonder.

// baseStore is the amount of res a store of the base size once held: the
// stock these tests start from and measure "past" against.
func baseStore(res string) float64 { return config.ResourceByKey()[res].BaseStorage }

// shrineEngine is a Primitive Age game whose wood (a base store's worth) is
// less than a Shrine's price.
func shrineEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	if cost, store := ge.Buildings.GetCost("shrine")["wood"], baseStore("wood"); cost <= store {
		t.Fatalf("setup: a shrine (%v wood) is within a base store of wood (%v)", cost, store)
	}
	return ge
}

// A copy is refused for what it costs against what is held, and for nothing
// else: every way of buying says "Cannot afford", none speaks of a store, and
// with the stock to pay it every way goes through.
func TestUncapped_EveryPurchasePathAsksOnlyForTheStock(t *testing.T) {
	paths := []struct {
		name string
		buy  func(ge *GameEngine) (int, error)
	}{
		{"build", func(ge *GameEngine) (int, error) { return 1, ge.BuildBuilding("shrine") }},
		{"build 1", func(ge *GameEngine) (int, error) { return ge.BuildMultiple("shrine", 1) }},
		{"build 5", func(ge *GameEngine) (int, error) { return ge.BuildMultiple("shrine", 5) }},
		{"build max", func(ge *GameEngine) (int, error) { return ge.BuildMultiple("shrine", 10000) }},
	}
	for _, p := range paths {
		t.Run(p.name+" without the stock", func(t *testing.T) {
			ge := shrineEngine(t)
			setAmount(ge, "wood", 40)
			n, err := p.buy(ge)
			if err == nil || !strings.HasPrefix(err.Error(), "Cannot afford") || strings.Contains(err.Error(), "storage") {
				t.Errorf("error = %v, want the plain cannot-afford refusal", err)
			}
			if ge.Buildings.GetCount("shrine")+queued(ge, "shrine") != 0 || n > 0 && p.name != "build" {
				t.Errorf("bought %d: nothing may be bought without the stock", n)
			}
		})
		t.Run(p.name+" with the stock", func(t *testing.T) {
			ge := shrineEngine(t)
			setAmount(ge, "wood", 100*ge.Buildings.GetCost("shrine")["wood"])
			if _, err := p.buy(ge); err != nil {
				t.Fatalf("with plenty of wood: %v", err)
			}
			if ge.Buildings.GetCount("shrine")+queued(ge, "shrine") < 1 {
				t.Error("nothing was bought with the stock to pay for it")
			}
		})
	}
}

// An upgrade priced over a base store goes through when the stock covers it.
func TestUncapped_UpgradeOverTheStoreGoesThrough(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["wood_camp"] = 3
	ge.mu.Lock()
	ge.advanceAge("stone_age")
	ge.mu.Unlock()
	one := ge.Buildings.UpgradeCopyCost("wood_camp", 0, nil)
	over := false
	for res, c := range one {
		if c > baseStore(res) {
			over = true
		}
		setAmount(ge, res, 100*c)
	}
	if !over {
		t.Fatalf("setup: one upgrade costs %v and is within every base store", one)
	}
	if err := ge.UpgradeBuilding("wood_camp", 1, false); err != nil {
		t.Fatalf("upgrade with the stock to pay: %v", err)
	}
	if ge.Buildings.GetCount("woodcutter_camp") != 1 {
		t.Error("the upgrade did not happen")
	}
}

// A wonder is paid through its bank, so `build <wonder> <n>` takes the single
// path and its rules, and a banked wonder takes nothing from stock.
func TestUncapped_BuildMultipleOnAWonderTakesTheBankPath(t *testing.T) {
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
	}
	ge.Buildings.wonderBanks[w] = ge.Buildings.GetCost(w)
	before := ge.Resources.GetAll()
	built, err := ge.BuildMultiple(w, 5)
	if err != nil || built != 1 {
		t.Fatalf("build %s 5 with its bank full = (%d, %v), want (1, nil)", w, built, err)
	}
	for res, v := range ge.Resources.GetAll() {
		if v != before[res] {
			t.Errorf("%s = %v, was %v: a banked wonder takes nothing from stock", res, v, before[res])
		}
	}
}

// `max` buys what the stock pays for, the same copies a run of single builds
// would.
func TestUncapped_MaxBuysWhatTheStockPaysFor(t *testing.T) {
	setup := func() *GameEngine {
		ge := newSeededEngine(1)
		setAmount(ge, "wood", 5000)
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
	if got != one || got < 2 {
		t.Errorf("max bought %d, single builds bought %d: want the same, and a few", got, one)
	}
	if byMax.Resources.Get("wood") != byOne.Resources.Get("wood") {
		t.Errorf("max left %v wood, single builds %v", byMax.Resources.Get("wood"), byOne.Resources.Get("wood"))
	}
	_, errMax := byMax.BuildMultiple("hut", 10000)
	errOne := byMax.BuildBuilding("hut")
	if errMax == nil || errOne == nil || errMax.Error() != errOne.Error() || strings.Contains(errMax.Error(), "storage") {
		t.Errorf("max = %v, build = %v: want one plain refusal", errMax, errOne)
	}
}

// Production, offline catch-up, gathering, loot, a sale's refund and a trade
// route's run all keep the whole amount, past the store.
func TestUncapped_EveryCreditKeepsTheWholeAmount(t *testing.T) {
	ge := newSeededEngine(1)
	store := baseStore("food")
	setAmount(ge, "food", store*10)
	before := ge.Resources.Get("food")

	if got, err := ge.GatherResource("food", 10); err != nil || got <= before {
		t.Errorf("gathering past the store: stock %v (was %v), %v", got, before, err)
	}
	ge.mu.Lock()
	ge.applyExpeditionRewards(map[string]float64{"food": 500})
	ge.mu.Unlock()
	if got := ge.Resources.Get("food"); got < before+500 {
		t.Errorf("loot past the store: food = %v, want at least %v", got, before+500)
	}
	if logHas(ge, "nearly full") || logHas(ge, "storage is") {
		t.Error("a log line speaks of a full store")
	}

	ge.Resources.SetRate("food", 3)
	at := ge.Resources.Get("food")
	ge.Resources.ApplyRates()
	if got := ge.Resources.Get("food"); got != at+3 {
		t.Errorf("a tick past the store: food = %v, want %v", got, at+3)
	}
	at = ge.Resources.Get("food")
	ge.Resources.AddProduced(1000, nil)
	if got := ge.Resources.Get("food"); math.Abs(got-(at+3000)) > 1e-6 {
		t.Errorf("offline credit past the store: food = %v, want %v", got, at+3000)
	}

	ge.mu.Lock()
	ge.advanceAge("stone_age") // buildings are sold from the Stone Age on
	ge.mu.Unlock()
	ge.Buildings.counts["hut"] = 3
	refund, _ := ge.Buildings.SellCost("hut", 3)
	for res := range refund {
		setAmount(ge, res, 10*baseStore(res)+1000)
	}
	held := ge.Resources.GetAll()
	if err := ge.SellBuilding("hut", 3); err != nil {
		t.Fatal(err)
	}
	for res, r := range refund {
		if got := ge.Resources.Get(res); got < held[res]+r-1e-6 {
			t.Errorf("refund of %v %s past the store: %v, was %v", r, res, got, held[res])
		}
	}
	if logHas(ge, "part of the refund was lost") {
		t.Error("a sale says part of its refund was lost")
	}
}

// Losing storage buildings takes nothing from the stock held.
func TestUncapped_LosingStorageKeepsTheStock(t *testing.T) {
	ge := newSeededEngine(1)
	ge.age = "stone_age"
	ge.Buildings.counts["storage_pit"] = 4
	ge.recalculateRates()
	held := config.Million
	setAmount(ge, "wood", held)
	if n := ge.Buildings.RemoveBuilding("storage_pit", 2); n != 2 {
		t.Fatalf("removed %d storage pits, want 2", n)
	}
	ge.recalculateRates()
	if got := ge.Resources.Get("wood"); got != held {
		t.Errorf("after losing storage: wood %v (was %v)", got, held)
	}
}

// A town that out-produces its stores piles stock up past them, and nothing
// is banked into the age's wonder on the side: the bank moves only by hand.
func TestUncapped_NothingOverflowsIntoTheWonder(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["wood_camp"] = 20
	ge.Buildings.counts["gathering_camp"] = 20
	ge.recalculateRates()
	w := ge.progress.WonderForAge(ge.age)
	if w == "" {
		t.Fatalf("setup: the %s has no wonder", ge.age)
	}
	for key, r := range ge.Resources.resources {
		r.Amount = baseStore(key)
	}
	for tick := 0; tick < 300; tick++ {
		ge.doTick()
	}
	if got, store := ge.Resources.Get("wood"), baseStore("wood"); got <= store {
		t.Errorf("wood = %v after 300 ticks of income from a base store of %v, want it past the store", got, store)
	}
	for res, v := range ge.Buildings.wonderBanks[w] {
		if v != 0 {
			t.Errorf("the %s bank holds %v without a deposit", res, v)
		}
	}
	if logHas(ge, "Overflow") {
		t.Error("a line speaks of overflow")
	}
}

// A deal whose goods are more than the store holds is taken, and charged:
// there is no room to run out of.
func TestUncapped_DealIntoAFullStoreIsTaken(t *testing.T) {
	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 30)
	ge.mu.Lock()
	idx := -1
	for i, d := range ge.Diplomacy.factions["merchant_guild"].Deals {
		if d.Get != "" && !d.Taken {
			idx = i
			break
		}
	}
	if idx < 0 {
		ge.mu.Unlock()
		t.Fatal("setup: the guild offers no deal that pays goods")
	}
	d := ge.Diplomacy.factions["merchant_guild"].Deals[idx]
	ge.Resources.resources[d.Give].Amount = d.GiveAmt * 3
	ge.Resources.resources[d.Get].Amount = baseStore(d.Get) // a base store's worth
	give, get := ge.Resources.Get(d.Give), ge.Resources.Get(d.Get)
	ge.mu.Unlock()

	if _, err := ge.AcceptFactionDeal("merchant_guild", idx+1); err != nil {
		t.Fatalf("deal into a full store: %v", err)
	}
	if got := ge.Resources.Get(d.Get); math.Abs(got-(get+d.GetAmt)) > 1e-6 {
		t.Errorf("%s = %v after the deal, want %v: the whole lot", d.Get, got, get+d.GetAmt)
	}
	if got := ge.Resources.Get(d.Give); math.Abs(got-(give-d.GiveAmt)) > 1e-6 {
		t.Errorf("%s = %v after the deal, want %v", d.Give, got, give-d.GiveAmt)
	}
}

// A trade route's run moves its whole import and export however full the
// import's store is.
func TestUncapped_RouteRunIntoAFullStoreMovesTheWholeRun(t *testing.T) {
	ge := newSeededEngine(9)
	for _, a := range []string{"stone_age", "bronze_age"} {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	def := ge.Trade.routeDefs["local_barter"]
	ge.Buildings.counts[def.RequiredBld] = def.MinCount
	ge.recalculateRates()
	var imp, exp string
	for imp = range def.Import {
	}
	for exp = range def.Export {
	}
	ge.Resources.UnlockResource(imp)
	ge.Resources.resources[exp].Amount = config.Million
	ge.Resources.resources[imp].Amount = baseStore(imp) // a base store's worth
	full := ge.Resources.Get(imp)
	ge.Trade.activeRoutes["local_barter"] = &ActiveRoute{Key: "local_barter", TicksLeft: 1}
	msgs := ge.Trade.Tick(ge.Resources, ge.Buildings, nil, 0, nil)
	if want := full + ge.Trade.RoutePay(def.Import[imp]); math.Abs(ge.Resources.Get(imp)-want) > 1e-6 {
		t.Errorf("%s = %v after a run into a full store, want %v", imp, ge.Resources.Get(imp), want)
	}
	if got := config.Million - ge.Resources.Get(exp); math.Abs(got-def.Export[exp]) > 1e-6 {
		t.Errorf("%s taken = %v, want the whole export %v", exp, got, def.Export[exp])
	}
	if len(msgs) != 0 {
		t.Errorf("the run said %q, want silence", msgs)
	}
	if r := ge.Trade.activeRoutes["local_barter"]; r.CyclesDone != 1 || r.Starved {
		t.Errorf("cycles %d, starved %v, want 1 and false", r.CyclesDone, r.Starved)
	}
}
