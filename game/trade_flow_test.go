package game

import (
	"encoding/json"
	"math"
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

// At the gate a route's flow is the old per-tick value: its run spread over
// its ticks. It grows in step with the buildings past the gate.
func TestRouteFlow_AtTheGateAndWithBuildings(t *testing.T) {
	ge, def := flowEngine(t, "local_barter", 1)
	ge.Buildings.counts[def.RequiredBld] = def.MinCount
	per := float64(def.TicksPerRun)
	for res, amount := range def.Import {
		if got, want := importFlow(ge, res), float64(amount/per); math.Abs(got-want) > 1e-9 {
			t.Errorf("flow of %s at the gate = %v, want %v", res, got, want)
		}
	}
	for res, amount := range def.Export {
		if got, want := ge.Resources.resources[res].Breakdown.TradeFlow, -float64(amount/per); math.Abs(got-want) > 1e-9 {
			t.Errorf("flow of %s at the gate = %v, want %v", res, got, want)
		}
	}
	// Each doubling of the line adds one gate's flow: x3 at four gates.
	ge.Buildings.counts[def.RequiredBld] = 4 * def.MinCount
	for res, amount := range def.Import {
		if got, want := importFlow(ge, res), float64(3*amount/per); math.Abs(got-want) > 1e-9 {
			t.Errorf("flow of %s at four gates = %v, want %v", res, got, want)
		}
	}
}

// A route exports only a surplus: with food scarce and a huge line, the
// route's draw never takes the food below what the town itself uses.
func TestRouteFlow_NeverExportsWhatTheTownUses(t *testing.T) {
	ge, def := flowEngine(t, "local_barter", 5000)
	if _, ok := def.Export["food"]; !ok {
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
	ge.Buildings.counts["market"] = 3
	ge.Buildings.counts["exchange"] = 1
	for _, k := range []string{"gold_caravan", "mercantile_convoy"} {
		if err := ge.Trade.StartRoute(k, ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
			t.Fatal(err)
		}
	}
	for key, r := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
		r.Amount = config.Million
	}
	// With a full store the routes take their whole demand D; the town
	// makes gross of its own.
	stone := ge.Resources.resources["stone"]
	ge.recalculateRates()
	demand := -stone.Breakdown.TradeFlow
	gross := stone.Rate + demand
	if demand <= 0 {
		t.Fatal("the routes take no stone")
	}
	// S = stock + gross: set it to half of D.
	stone.Amount = demand/2 - gross
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
	// With nothing in store the routes take only what the town makes.
	stone.Amount = 0
	ge.recalculateRates()
	if got := -stone.Breakdown.TradeFlow; math.Abs(got-gross) > 1e-9 {
		t.Errorf("with no stone in store the routes take %v, want what is made, %v", got, gross)
	}
}

// A paused route, a route short of its building and a route cut by war move
// nothing, and stay listed. A pause is saved; the rest of the state is not.
func TestRouteFlow_PausedAndHeldMoveNothing(t *testing.T) {
	ge, def := flowEngine(t, "local_barter", 2)
	var res string
	for res = range def.Import {
	}
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
		ge, def := flowEngine(t, "local_barter", 1)
		var imp, exp string
		for imp = range def.Import {
		}
		for exp = range def.Export {
		}
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
	var imp string
	for imp = range def.Import {
	}
	want := float64(def.Import[imp]/float64(def.TicksPerRun)) * ticks * OfflineEfficiency
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
