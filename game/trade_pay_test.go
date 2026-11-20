package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestRouteView_SharesMultiplyOutToTickPay: the pay and shares on the state
// view equal what Tick pays, with a harbor and an ally in play, and the
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
	if want := def.Import["wood"] * sum; math.Abs(got.Pay["wood"]-want) > 1e-9 {
		t.Fatalf("view pay %.6f, shares give %.6f", got.Pay["wood"], want)
	}
	before := ge.Resources.Get("wood")
	for i := 0; i < def.TicksPerRun; i++ {
		ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, ge.harborRouteBonus(), nil)
	}
	if paid := ge.Resources.Get("wood") - before; math.Abs(paid-got.Pay["wood"]) > 1e-9 {
		t.Fatalf("Tick paid %.6f, view says %.6f", paid, got.Pay["wood"])
	}
}

// TestRouteView_StarvedShowsAndClears: a route short of its export is flagged
// on the view; a funded one is not.
func TestRouteView_StarvedShowsAndClears(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "industrial_age"
	ge.Buildings.counts["market"] = 2
	setRes(ge, "food", 0)
	if err := ge.Trade.StartRoute("local_barter", ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
		t.Fatal(err)
	}
	def := config.TradeRouteByKey()["local_barter"]
	for i := 0; i < def.TicksPerRun; i++ {
		ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, 0, nil)
	}
	view := func() ActiveRouteInfo {
		return ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil, RouteTerms{}).ActiveRoutes[0]
	}
	if !view().Starved {
		t.Fatal("route short of its export is not flagged")
	}
	setRes(ge, "food", config.Million)
	for i := 0; i < def.TicksPerRun; i++ {
		ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, 0, nil)
	}
	if view().Starved {
		t.Fatal("a funded run left the route flagged")
	}
}

// TestRouteView_AllyShareInPay: an ally whose specialty a route imports adds
// its share, and the view's pay matches the sum of Tick's terms.
func TestRouteView_AllyShareInPay(t *testing.T) {
	ge := NewGameEngine()
	for _, f := range ge.Diplomacy.factionList {
		for _, def := range ge.Trade.routeList {
			amount, ok := def.Import[f.Specialty]
			if !ok {
				continue
			}
			allyFaction(ge, f.Key)
			pay, shares := ge.Trade.payFor(def, ge.routeTerms())
			var ally float64
			for _, s := range shares[f.Specialty] {
				if s.Kind == ShareAlly {
					ally += s.Percent / 100
				}
			}
			if ally <= 0 {
				t.Fatalf("no ally share for %s on %s", f.Key, def.Key)
			}
			want := ge.Trade.RunPay(amount, ge.harborRouteBonus()+ge.Diplomacy.GetTradeBonus(f.Specialty))
			if math.Abs(pay[f.Specialty]-want) > 1e-9 {
				t.Fatalf("pay %.6f, Tick would pay %.6f", pay[f.Specialty], want)
			}
			return
		}
	}
	t.Skip("no route imports a faction specialty")
}
