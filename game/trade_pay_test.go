package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestRouteView_SharesMultiplyOutToTickPay: the flow and shares on the state
// view equal what the rates carry, with a harbor and an ally in play, and the
// building shares keep their order.
func TestRouteView_SharesMultiplyOutToTickPay(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "industrial_age"
	ge.Buildings.counts["market"] = 2
	ge.Buildings.counts["harbor"] = 1
	setRes(ge, "food", config.Million)
	if err := ge.Trade.StartRoute("local_barter", ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
		t.Fatal(err)
	}
	var first []PayShare
	for i := 0; i < 5; i++ {
		terms := ge.routeTerms()
		got := ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil, terms).ActiveRoutes[0]
		list := got.Shares["wood"]
		if i == 0 {
			first = list
		}
		if len(list) != len(first) {
			t.Fatalf("share list changed length: %d vs %d", len(list), len(first))
		}
		for j := range list {
			if list[j] != first[j] {
				t.Fatalf("share %d changed between calls: %v vs %v", j, list[j], first[j])
			}
		}
	}
	def := config.TradeRouteByKey()["local_barter"]
	got := ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil, ge.routeTerms()).ActiveRoutes[0]
	sum := 1.0
	for _, s := range got.Shares["wood"] {
		sum += s.Percent / 100
	}
	if want := float64(def.Import["wood"]*sum) / float64(def.TicksPerRun); math.Abs(got.Import["wood"]-want) > 1e-9 {
		t.Fatalf("view flow %.6f, shares give %.6f", got.Import["wood"], want)
	}
	// The rates carry the same flow the view lists.
	flow := importFlow(ge, "wood")
	got = ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil, ge.routeTerms()).ActiveRoutes[0]
	if math.Abs(flow-got.Import["wood"]*got.Scale) > 1e-9 {
		t.Fatalf("the rates carry %.6f, the view says %.6f", flow, got.Import["wood"]*got.Scale)
	}
}

// TestRouteView_ShortShowsAndClears: a route short of its export runs slower
// and the view says by how much; a funded one runs in full.
func TestRouteView_ShortShowsAndClears(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "industrial_age"
	ge.Buildings.counts["market"] = 2
	setRes(ge, "food", 0)
	if err := ge.Trade.StartRoute("local_barter", ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
		t.Fatal(err)
	}
	view := func() ActiveRouteInfo {
		ge.recalculateRates()
		return ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil, RouteTerms{}).ActiveRoutes[0]
	}
	if got := view().Run; got >= 1 {
		t.Fatalf("a route with no stock runs at %v, want under 1", got)
	}
	setRes(ge, "food", config.Million)
	if got := view().Run; got != 1 {
		t.Fatalf("a funded route runs at %v, want 1", got)
	}
}

// TestRouteView_AllyShareInPay: an ally whose specialty a route imports adds
// its share, and the view's flow matches the sum of the terms.
func TestRouteView_AllyShareInPay(t *testing.T) {
	ge := NewGameEngine()
	for _, f := range ge.Diplomacy.factionList {
		for _, def := range ge.Trade.routeList {
			amount, ok := def.Import[f.Specialty]
			if !ok {
				continue
			}
			allyFaction(ge, f.Key)
			ge.Trade.activeRoutes[def.Key] = &ActiveRoute{Key: def.Key}
			view := ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil, ge.routeTerms()).ActiveRoutes[0]
			shares := view.Shares
			var ally float64
			for _, s := range shares[f.Specialty] {
				if s.Kind == ShareAlly {
					ally += s.Percent / 100
				}
			}
			if ally <= 0 {
				t.Fatalf("no ally share for %s on %s", f.Key, def.Key)
			}
			want := float64(ge.Trade.RunPay(amount, ge.harborRouteBonus()+ge.Diplomacy.GetTradeBonus(f.Specialty)) / float64(def.TicksPerRun))
			if math.Abs(view.Import[f.Specialty]-want) > 1e-9 {
				t.Fatalf("flow %.6f, the rates would carry %.6f", view.Import[f.Specialty], want)
			}
			return
		}
	}
	t.Skip("no route imports a faction specialty")
}
