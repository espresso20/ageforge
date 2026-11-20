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

	// Routes still pay: a running route takes its export and brings its
	// import as a flow in the rates.
	re := newSeededEngine(9)
	learn(re, "the_wheel")
	def := re.Trade.routeDefs["local_barter"]
	re.mu.Lock()
	re.Buildings.counts[def.RequiredBld] = def.MinCount
	for _, r := range re.Resources.resources {
		r.Amount = config.Million
	}
	re.Resources.UnlockResource("gold")
	re.Trade.activeRoutes["local_barter"] = &ActiveRoute{Key: "local_barter"}
	var exp, imp string
	for exp = range def.Export {
	}
	for imp = range def.Import {
	}
	re.recalculateRates()
	paid := re.Resources.resources[imp].Breakdown.TradeFlow
	took := -re.Resources.resources[exp].Breakdown.TradeFlow
	re.mu.Unlock()
	per := runLength(re, def)
	if math.Abs(paid*per-re.Trade.RoutePay(def.Import[imp])) > 1e-9 || math.Abs(took*per-def.Export[exp]) > 1e-9 {
		t.Errorf("a route took %v %s and paid %v %s a run, want %v and %v", took*per, exp, paid*per, imp, def.Export[exp], re.Trade.RoutePay(def.Import[imp]))
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

	// The black market still pays: a win on iron is the stake's minutes of
	// culture income, times the win multiplier, in minutes of iron income.
	be := bmEngine(50000)
	be.blackMarketRand = blackMarketSeedWin()
	be.Resources.resources["iron"].Amount = 0
	stake := be.BlackMarketStatus().Cost
	won, gain, err := be.DoBlackMarket("iron")
	if err != nil || !won {
		t.Fatalf("a black-market run on iron: won %v, err %v", won, err)
	}
	if want := bmFloorPayout(be, "iron", stake); math.Abs(gain-want) > 1e-9*want || math.Abs(be.Resources.Get("iron")-want) > 1e-9*want {
		t.Errorf("a win on iron paid %v (iron %v), want %v", gain, be.Resources.Get("iron"), want)
	}
}

// TestBlackMarketPaysTheSameWorthInEveryResource: the smugglers deal in gold
// and nine resources, and a win pays the same worth in each, as minutes of
// the town's income of it: winMult times the stake's minutes of culture
// income. They refuse everything else, and a resource the age makes none of,
// as when they read the market's table.
func TestBlackMarketPaysTheSameWorthInEveryResource(t *testing.T) {
	deals := map[string]bool{
		"gold": true, "iron": true, "coal": true, "steel": true, "oil": true, "electricity": true,
		"data": true, "crypto": true, "dark_matter": true, "quantum_flux": true,
	}
	const stake = 10000.0
	ge := newSeededEngine(4)
	ge.age = "quantum_age"
	typC := ge.rules.TypicalIncome("culture", ge.age)
	paid := 0
	for _, r := range ge.rules.Resources() {
		typR := ge.rules.TypicalIncome(r.Key, ge.age)
		got := ge.blackMarketReward(r.Key, stake)
		if !deals[r.Key] || typR <= 0 {
			if got != 0 {
				t.Errorf("a win on %s pays %v, want nothing (not dealt in, or none made)", r.Key, got)
			}
			continue
		}
		paid++
		// This town makes none of either: both incomes stand at a quarter of
		// the typical, so the minutes of the resource's typical income a win
		// pays are the stake's minutes of the culture's, times the multiplier.
		minutes := got / (config.TicksPerMinute * typR)
		want := blackMarketWinMult * stake / (config.TicksPerMinute * typC)
		if math.Abs(minutes-want) > 1e-9*want {
			t.Errorf("a win on %s pays %v minutes of its typical income, want %v", r.Key, minutes, want)
		}
	}
	if paid != len(deals) {
		t.Errorf("%d resources pay in the Quantum Age, want %d", paid, len(deals))
	}
	// A town making more of the resource is paid more of it, held to four
	// times the typical income; one making less is paid as at a quarter.
	typR := ge.rules.TypicalIncome("coal", ge.age)
	coal := ge.Resources.resources["coal"]
	for _, c := range []struct{ rate, held float64 }{{10 * typR, 4 * typR}, {2 * typR, 2 * typR}, {0.01 * typR, 0.25 * typR}} {
		coal.Rate = c.rate
		want := blackMarketWinMult * stake * c.held / (config.EventIncomeFloor * typC)
		if got := ge.blackMarketReward("coal", stake); math.Abs(got-want) > 1e-9*want {
			t.Errorf("a town making %v coal is paid %v, want %v", c.rate, got, want)
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
