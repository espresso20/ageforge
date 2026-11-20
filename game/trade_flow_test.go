package game

import (
	"encoding/json"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// flowEngine is a town with the buildings for a route (count of them) and
// plenty of everything it exports.
func flowEngine(t *testing.T, key string, count int) (*GameEngine, config.TradeRouteDef) {
	t.Helper()
	ge := NewGameEngine()
	ge.age = "industrial_age"
	townMakes(ge)
	def := ge.Trade.routeDefs[key]
	ge.Buildings.counts[def.RequiredBld] = count
	for _, r := range ge.Resources.resources {
		r.Amount = config.Million
	}
	if err := ge.Trade.StartRoute(key, ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
		t.Fatal(err)
	}
	return ge, def
}

// At the gate a route sends RouteShare of the town's gross production of its
// export and brings RouteMargin times its value in the import. It grows by one gate's
// flow for each doubling of the line.
func TestRouteFlow_AtTheGateAndWithBuildings(t *testing.T) {
	ge, def := flowEngine(t, "local_barter", 1)
	ge.Buildings.counts[def.RequiredBld] = def.MinCount
	ge.recalculateRates()
	check := func(scale float64, when string) {
		gout, gin := gate(ge, def)
		out := float64(gout["food"] * scale)
		if got := -ge.Resources.resources["food"].Breakdown.TradeFlow; math.Abs(got-out) > 1e-9 {
			t.Errorf("food sent %s = %v, want %v", when, got, out)
		}
		if got, want := importFlow(ge, "wood"), float64(gin["wood"]*scale); math.Abs(got-want) > 1e-9 {
			t.Errorf("wood brought %s = %v, want %v", when, got, want)
		}
	}
	check(1, "at the gate")
	ge.Buildings.counts[def.RequiredBld] = 4 * def.MinCount
	ge.recalculateRates()
	check(3, "at four gates")
}

// A route exports only a surplus: with food scarce and a huge line, the
// route's draw never takes the food below what the town itself uses.
func TestRouteFlow_NeverExportsWhatTheTownUses(t *testing.T) {
	ge, def := flowEngine(t, "local_barter", 5000)
	if !slices.Contains(def.Export, "food") {
		t.Fatal("the route no longer exports food")
	}
	food := ge.Resources.resources["food"]
	food.Amount = 3
	for key := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
	}
	ge.recalculateRates()
	pre := food.Rate - food.Breakdown.TradeFlow
	if pre+food.Amount >= 0 {
		if left := food.Amount + food.Rate; left < -1e-9 {
			t.Errorf("trade left %v food after a tick: it took what the town uses (net %v, store %v)", left, pre, food.Amount)
		}
	} else if food.Breakdown.TradeFlow != 0 {
		t.Errorf("the town already runs short of food, yet the route moved %v", food.Breakdown.TradeFlow)
	}
}

// Routes that export the same resource share its stock: each runs at
// min(1, S/D), both legs, and together they spend what there is.
func TestRouteFlow_ThrottleSharesTheStock(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "renaissance_age"
	townMakes(ge)
	// A line this long sends more than the town makes, so the stock throttles.
	ge.Buildings.counts["market"] = 3 << 12
	ge.Buildings.counts["exchange"] = 1 << 12
	for _, k := range []string{"gold_caravan", "mercantile_convoy"} {
		if err := ge.Trade.StartRoute(k, ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
			t.Fatal(err)
		}
	}
	for key, r := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
		r.Amount = 1e18 // so only the stone throttles
	}
	// With a full store the routes take their whole demand D; the town
	// makes gross of its own.
	stone := ge.Resources.resources["stone"]
	ge.recalculateRates()
	demand := -stone.Breakdown.TradeFlow
	gross := ge.Trade.gross["stone"]
	if demand <= 0 {
		t.Fatal("the routes take no stone")
	}
	// S = stock + gross, less the share the town keeps of each (twice):
	// set it to half of D.
	keep := recruitFoodMarginShare * gross
	stone.Amount = demand/2 - gross + 2*keep
	if stone.Amount < 0 {
		t.Skipf("the town makes %v stone a tick, more than half the demand %v", gross, demand)
	}
	ge.recalculateRates()
	for _, k := range []string{"gold_caravan", "mercantile_convoy"} {
		if got := ge.Trade.activeRoutes[k].Run; math.Abs(got-0.5) > 1e-9 {
			t.Errorf("%s runs at %v, want 0.5", k, got)
		}
	}
	if got := stone.Breakdown.TradeFlow; math.Abs(got+demand/2) > 1e-9 {
		t.Errorf("the routes take %v stone a tick, want half the demand %v", -got, demand/2)
	}
	// With nothing in store the routes take what the town makes, less its share.
	stone.Amount = 0
	ge.recalculateRates()
	if got := -stone.Breakdown.TradeFlow; math.Abs(got-(gross-keep)) > 1e-9 {
		t.Errorf("with no stone in store the routes take %v, want what is made less the share kept, %v", got, gross-keep)
	}
}

// A paused route, a route short of its building and a route cut by war move
// nothing, and stay listed. A pause is saved; the rest of the state is not.
func TestRouteFlow_PausedAndHeldMoveNothing(t *testing.T) {
	ge, def := flowEngine(t, "local_barter", 2)
	var res string
	res = def.Import[0]
	if importFlow(ge, res) <= 0 {
		t.Fatal("a running route moves nothing")
	}
	if err := ge.Trade.StopRoute(def.Key); err != nil {
		t.Fatal(err)
	}
	if got := importFlow(ge, res); got != 0 {
		t.Errorf("a paused route moves %v", got)
	}
	if err := ge.Trade.StopRoute(def.Key); err == nil {
		t.Error("a paused route stopped twice")
	}
	raw, _ := json.Marshal(ge.Trade.activeRoutes[def.Key])
	if string(raw) != `{"Key":"local_barter","Paused":true}` {
		t.Errorf("a route is saved as %s", raw)
	}
	if err := ge.Trade.StartRoute(def.Key, ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
		t.Fatal(err)
	}
	if importFlow(ge, res) <= 0 {
		t.Error("a resumed route moves nothing")
	}
	ge.Buildings.counts[def.RequiredBld] = 0
	if got := importFlow(ge, res); got != 0 {
		t.Errorf("a route with no building moves %v", got)
	}
	if n := len(ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil, RouteTerms{}).ActiveRoutes); n != 1 {
		t.Errorf("the held route is listed %d times, want 1", n)
	}
}

// Time away runs the flows too, at the offline efficiency, and spends the
// stock once: a store that holds a few ticks' worth of the export ends at 0
// and the import comes in with it.
func TestRouteFlow_OfflineAppliesAndThrottles(t *testing.T) {
	away := func(food float64) (gained, left float64, ge *GameEngine) {
		ge, def := flowEngine(t, "local_barter", 1<<14)
		ge.Buildings.counts["test_wood"] = 0 // only the route brings wood
		var imp, exp string
		imp, exp = def.Import[0], def.Export[0]
		ge.Resources.resources[exp].Amount = food
		before := ge.Resources.Get(imp)
		ge.mu.Lock()
		ge.applyOfflineProgress(time.Hour)
		ge.mu.Unlock()
		return ge.Resources.Get(imp) - before, ge.Resources.Get(exp), ge
	}
	full, _, ge := away(config.Million)
	def := ge.Trade.routeDefs["local_barter"]
	ticks := float64(int(time.Hour / ge.tickIntervalLocked()))
	imp := def.Import[0]
	_, gin := gate(ge, def)
	want := float64(gin[imp]*ge.Trade.activeRoutes[def.Key].Scale) * ticks * OfflineEfficiency
	if math.Abs(full-want) > want*0.01 {
		t.Errorf("an hour away brought %v %s, want about %v", full, imp, want)
	}
	short, left, _ := away(50)
	if short <= 0 || short >= full {
		t.Errorf("a short store brought %v %s, want between 0 and %v", short, imp, full)
	}
	if left < 0 {
		t.Errorf("the store went to %v", left)
	}
}

// gate is what def moves a tick at its gate, before the techs, the harbor,
// the allies and the speed cut: the first import the town makes comes in at
// RouteShare of its make and the export is worth that less RouteMargin; with
// none of the imports made, RouteShare of each export's make is sent and
// RouteMargin times its value comes in, by price level, over the imports.
func gate(ge *GameEngine, def config.TradeRouteDef) (out, in map[string]float64) {
	lv := func(res string) float64 { return ge.Trade.rules.RoutePriceLevel(res, ge.age) }
	out, in = map[string]float64{}, map[string]float64{}
	worth := 0.0
	for _, res := range def.Import {
		if g := ge.Trade.gross[res]; g > 0 {
			worth = config.RouteShare * g / lv(res)
			break
		}
	}
	if worth > 0 {
		for _, res := range def.Export {
			out[res] = worth * float64(len(def.Import)) / config.RouteMargin / float64(len(def.Export)) * lv(res)
		}
	} else {
		for _, res := range def.Export {
			out[res] = config.RouteShare * ge.Trade.gross[res]
			worth += out[res] / lv(res)
		}
		worth *= config.RouteMargin / float64(len(def.Import))
	}
	for _, res := range def.Import {
		in[res] = worth * lv(res)
	}
	return out, in
}

// townMakes gives the town some of everything routes export, so a route has
// income to take its share of.
func townMakes(ge *GameEngine) {
	for _, res := range []string{"food", "wood", "stone", "gold", "culture", "iron", "coal"} {
		addProductionBuilding(ge, "test_"+res, res, 10, 1)
	}
}

// A town that makes none of the export sends nothing, and the route says so.
func TestRouteFlow_NoIncomeHoldsTheRoute(t *testing.T) {
	ge, def := flowEngine(t, "local_barter", 2)
	ge.Buildings.counts["test_food"] = 0
	ge.recalculateRates()
	if importFlow(ge, "wood") != 0 || ge.Resources.resources["food"].Breakdown.TradeFlow != 0 {
		t.Error("a route with no income to take a share of still moved goods")
	}
	if got := ge.Trade.activeRoutes[def.Key].NoIncome; got != "food" {
		t.Errorf("NoIncome = %q, want food", got)
	}
}

// A food route on a long line spends the stock down but never runs it dry:
// the town keeps a share of what it makes, so it never starves under a route.
func TestRouteFlow_FoodStockNeverRunsDry(t *testing.T) {
	ge, _ := flowEngine(t, "local_barter", 5000)
	for key := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
	}
	food := ge.Resources.resources["food"]
	food.Amount = 2000
	for i := 0; i < 2000; i++ {
		ge.recalculateRates()
		ge.Resources.ApplyRates()
	}
	if food.Amount <= 0 || food.Rate < 0 {
		t.Errorf("food %v at a rate of %v after the routes spent the stock, want a store above 0 and a rate that is not negative", food.Amount, food.Rate)
	}
}

// The board reads a route's share of income off the state, against the same
// gross the engine sized the route on, so it can exceed 100%: a route that
// brings 6 times the wood the town cuts is 600% of its wood income.
func TestRouteFlow_ShareIsAgainstGrossIncome(t *testing.T) {
	ge, _ := flowEngine(t, "local_barter", 2)
	for key := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
	}
	ge.recalculateRates()
	wood := ge.Resources.resources["wood"]
	made := wood.Rate - wood.Breakdown.TradeFlow // what the town makes, trade left out
	st := ge.GetState().Trade
	if got, want := st.Share("wood", wood.Breakdown.TradeFlow), wood.Breakdown.TradeFlow/made; math.Abs(got-want) > 1e-9 || got <= 0 {
		t.Errorf("share of wood income = %v, want %v", got, want)
	}
	if st.Share("nothing", 1) != 0 {
		t.Error("a good the town makes none of has a share")
	}
}

// Research has a price level (its scholars' buildings, like any flow
// resource), so a route that imports it brings some.
func TestRouteFlow_ResearchImportPays(t *testing.T) {
	ge, def := flowEngine(t, "triangular_trade", 1)
	ge.Buildings.counts[def.RequiredBld] = def.MinCount
	ge.recalculateRates()
	if got := ge.Trade.activeRoutes[def.Key].Research; got <= 0 {
		t.Errorf("research brought = %v, want more than 0", got)
	}
}

// A route is sized on its import: at the gate it brings RouteShare of what the
// town makes of it, whatever the ratio of the two goods, and more of the
// building multiplies it. The pay terms raise the import and not the export.
func TestRouteFlow_ShareOfIncomeAtTheGate(t *testing.T) {
	ge, def := flowEngine(t, "local_barter", 1)
	ge.recalculateRates()
	gross := ge.Trade.gross["wood"]
	if got := importFlow(ge, "wood") / gross; math.Abs(got-config.RouteShare) > 1e-9 {
		t.Errorf("at the gate the route brings %v of the wood made, want %v", got, config.RouteShare)
	}
	ge.Buildings.counts[def.RequiredBld] = 4 * def.MinCount
	ge.recalculateRates()
	if got := importFlow(ge, "wood") / gross; math.Abs(got-3*config.RouteShare) > 1e-9 {
		t.Errorf("at four gates the route brings %v of the wood made, want %v", got, 3*config.RouteShare)
	}
	// A bonus raises what comes in and not what goes out.
	out0, in0, _ := ge.Trade.flows(def, RouteTerms{}, ge.age)
	out1, in1, _ := ge.Trade.flows(def, RouteTerms{Harbor: 1}, ge.age)
	if out1["food"] != out0["food"] || math.Abs(in1["wood"]-2*in0["wood"]) > 1e-9 {
		t.Errorf("a bonus of 100%% took food %v to %v and wood %v to %v", out0["food"], out1["food"], in0["wood"], in1["wood"])
	}
}

// Every route can run: in a town of its own age that makes its export and has
// its building, it is not held for want of income or a price, and it moves
// goods.
func TestRouteFlow_EveryRouteCanRun(t *testing.T) {
	for _, def := range NewGameEngine().Rules().TradeRoutes() {
		ge := NewGameEngine()
		ge.age = def.MinAge
		for _, res := range def.Export {
			addProductionBuilding(ge, "test_"+res, res, 10, 1)
		}
		ge.Buildings.counts[def.RequiredBld] = def.MinCount
		for key, r := range ge.Resources.resources {
			ge.Resources.UnlockResource(key)
			r.Amount = 1e18
		}
		if err := ge.Trade.StartRoute(def.Key, ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
			t.Errorf("%s: %v", def.Key, err)
			continue
		}
		ge.recalculateRates()
		route := ge.Trade.activeRoutes[def.Key]
		if route.NoIncome != "" || route.Run <= 0 {
			t.Errorf("%s: held for %q, runs at %v", def.Key, route.NoIncome, route.Run)
			continue
		}
		moved := route.Research
		for _, res := range def.Import {
			if r := ge.Resources.resources[res]; r != nil {
				moved += r.Breakdown.TradeFlow
			}
		}
		if moved <= 0 {
			t.Errorf("%s runs but brings nothing in", def.Key)
		}
	}
}
