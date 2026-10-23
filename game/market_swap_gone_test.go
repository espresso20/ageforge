package game

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestATownCanNoLongerSwapResourcesButRoutesDealsAndTheBlackMarketStillPay:
// the market does not turn one resource into another by any engine call,
// plan item or saved order, and the three things that stood on their own
// still pay: a trade route's run, a faction deal, a black-market win.
func TestATownCanNoLongerSwapResourcesButRoutesDealsAndTheBlackMarketStillPay(t *testing.T) {
	// No engine call: nothing on the engine or the trade manager swaps, or
	// plans a swap.
	for _, typ := range []reflect.Type{reflect.TypeOf(&GameEngine{}), reflect.TypeOf(&TradeManager{})} {
		for i := 0; i < typ.NumMethod(); i++ {
			name := typ.Method(i).Name
			if strings.Contains(name, "Exchange") || strings.Contains(name, "PlanAddTrade") || name == "RateIn" || name == "Pressure" || name == "DecayPressure" {
				t.Errorf("%s still has %s", typ, name)
			}
		}
	}

	// No plan item and no saved order: a town with the old save's sell
	// orders and pressure plays its ticks without one swap, and its route
	// runs on.
	isolateAccountDir(t)
	ge := loadFixture(t, marketSwapFixture, "gone")
	ge.StepTicks(300)
	for _, l := range ge.GetLogs() {
		if strings.HasPrefix(l.Message, "Traded ") || strings.Contains(l.Message, "plan traded") {
			t.Fatalf("a swap happened: %q", l.Message)
		}
	}
	for _, it := range ge.GetState().Plan {
		if it.Kind == "trade" {
			t.Errorf("a sell order is back in the plan: %+v", it)
		}
	}
	if len(ge.GetState().Trade.ActiveRoutes) == 0 {
		t.Error("the route stopped running")
	}

	// Routes still pay: a run takes its export and brings its import.
	re := newSeededEngine(9)
	learn(re, "the_wheel")
	def := re.Trade.routeDefs["local_barter"]
	re.mu.Lock()
	re.Buildings.counts[def.RequiredBld] = def.MinCount
	for _, r := range re.Resources.resources {
		r.Amount = config.Million
	}
	re.Resources.UnlockResource("gold")
	re.Trade.activeRoutes["local_barter"] = &ActiveRoute{Key: "local_barter", TicksLeft: 1}
	var exp, imp string
	for exp = range def.Export {
	}
	for imp = range def.Import {
	}
	exp0, imp0 := re.Resources.Get(exp), re.Resources.Get(imp)
	re.Trade.Tick(re.Resources, re.Buildings, nil, 0, nil)
	paid := re.Resources.Get(imp) - imp0
	took := exp0 - re.Resources.Get(exp)
	re.mu.Unlock()
	if math.Abs(paid-re.Trade.RoutePay(def.Import[imp])) > 1e-9 || math.Abs(took-def.Export[exp]) > 1e-9 {
		t.Errorf("a route run took %v %s and paid %v %s, want %v and %v", took, exp, paid, imp, def.Export[exp], re.Trade.RoutePay(def.Import[imp]))
	}

	// Deals still pay: the offer's goods for its price, once.
	de := dealEngine(t, "colonial_age")
	meet(de, "merchant_guild", 0)
	de.mu.RLock()
	d := de.Diplomacy.factions["merchant_guild"].Deals[0]
	give0, get0 := de.Resources.Get(d.Give), de.Resources.Get(d.Get)
	de.mu.RUnlock()
	if _, err := de.AcceptFactionDeal("merchant_guild", 1); err != nil {
		t.Fatal(err)
	}
	if give1, get1 := de.Resources.Get(d.Give), de.Resources.Get(d.Get); give1 != give0-d.GiveAmt || get1 != get0+d.GetAmt {
		t.Errorf("a deal moved %s %v→%v and %s %v→%v, want -%v and +%v", d.Give, give0, give1, d.Get, get0, get1, d.GiveAmt, d.GetAmt)
	}

	// The black market still pays: a win on iron is the stake's gold value
	// over iron's rate.
	be := bmEngine(50000)
	be.blackMarketRand = blackMarketSeedWin()
	be.Resources.resources["iron"].Amount = 0
	stake := be.BlackMarketStatus().Cost
	won, gain, err := be.DoBlackMarket("iron")
	if err != nil || !won {
		t.Fatalf("a black-market run on iron: won %v, err %v", won, err)
	}
	if want := stake * blackMarketWinMult / 2; math.Abs(gain-want) > 1e-6 || math.Abs(be.Resources.Get("iron")-want) > 1e-6 {
		t.Errorf("a win on iron paid %v (iron %v), want %v", gain, be.Resources.Get("iron"), want)
	}
}

// TestBlackMarketKeepsItsNineRates: the smugglers deal in nine resources at
// their own gold rates (a payout is the stake's gold value over the rate),
// pay gold directly, and refuse everything else, as when they read the
// market's table.
func TestBlackMarketKeepsItsNineRates(t *testing.T) {
	ge := newSeededEngine(4)
	rates := map[string]float64{
		"iron": 2, "coal": 0.08, "steel": 5, "oil": 3, "electricity": 0.5,
		"data": 5, "crypto": 20, "dark_matter": 50, "quantum_flux": 100,
	}
	const stake = 10000.0
	for _, r := range ge.rules.Resources() {
		want := 0.0
		if rate, ok := rates[r.Key]; ok {
			want = stake * blackMarketWinMult / rate
		}
		if r.Key == "gold" {
			want = stake * blackMarketWinMult
		}
		if got := ge.blackMarketReward(r.Key, stake); math.Abs(got-want) > 1e-9*math.Max(want, 1) {
			t.Errorf("a win on %s pays %v, want %v", r.Key, got, want)
		}
	}
	// A resource the smugglers do not deal in is refused with the old words,
	// and nothing is spent.
	be := bmEngine(50000)
	be.blackMarketRand = blackMarketSeedWin()
	if _, _, err := be.DoBlackMarket("food"); err == nil || err.Error() != "Smugglers do not deal in food. Pick another resource." {
		t.Errorf("a run on food: %v", err)
	}
	if be.Resources.Get("culture") != 50000 {
		t.Errorf("culture is %v after a refused run, want it untouched", be.Resources.Get("culture"))
	}
}
