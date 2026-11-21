package game

import (
	"math"
	"testing"
)

func upgradeTown(t *testing.T) (*GameEngine, string) {
	t.Helper()
	ge, def := flowEngine(t, "local_barter", 2)
	ge.recalculateRates()
	return ge, def.Import[0]
}

func TestRouteUpgrade_CostIsTenMinutesOfTheBaseFlowAndGrows(t *testing.T) {
	ge, good := upgradeTown(t)
	def := ge.Trade.routeDefs["local_barter"]
	route := ge.Trade.activeRoutes["local_barter"]
	_, gin := gate(ge, def)
	want := float64(gin[good]*route.Scale) * float64(minutesToTicks(10))
	_, in, _ := ge.Trade.flows(def, ge.routeTerms(), ge.age)
	g, c := ge.Trade.upgradeCost(route, def, in)
	if g != good || math.Abs(c-want) > want*1e-9 {
		t.Fatalf("cost at level 0 = %v %s, want %v %s", c, g, want, good)
	}
	route.Level = 1
	if _, c1 := ge.Trade.upgradeCost(route, def, in); math.Abs(c1-want*1.15) > want*1e-9 {
		t.Errorf("cost at level 1 = %v, want %v", c1, want*1.15)
	}
}

func TestRouteUpgrade_MultiplierMilestones(t *testing.T) {
	for level, want := range map[int]float64{0: 1, 1: 1.25, 9: 3.25, 10: 7, 25: 2 * 2 * 7.25, 50: 2 * 2 * 2 * 13.5} {
		if got := routeMult(level); math.Abs(got-want) > 1e-12 {
			t.Errorf("routeMult(%d) = %v, want %v", level, got, want)
		}
	}
}

func TestRouteUpgrade_BuyingSpendsAndSurvivesASave(t *testing.T) {
	ge, good := upgradeTown(t)
	def := ge.Trade.routeDefs["local_barter"]
	before, flow0 := ge.Resources.Get(good), importFlow(ge, good)
	_, in, _ := ge.Trade.flows(def, ge.routeTerms(), ge.age)
	_, cost := ge.Trade.upgradeCost(ge.Trade.activeRoutes["local_barter"], def, in)
	if err := ge.UpgradeTradeRoute("local_barter"); err != nil {
		t.Fatal(err)
	}
	if got := ge.Resources.Get(good); math.Abs(before-cost-got) > 1e-6 {
		t.Errorf("%s went %v to %v, want %v spent", good, before, got, cost)
	}
	if lv := ge.Trade.activeRoutes["local_barter"].Level; lv != 1 {
		t.Fatalf("level = %d", lv)
	}
	if got := importFlow(ge, good); math.Abs(got-1.25*flow0) > 1e-9*flow0 {
		t.Errorf("import flow = %v, want %v", got, 1.25*flow0)
	}
	if err := ge.SaveGame("route_upgrade"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("route_upgrade"); err != nil {
		t.Fatal(err)
	}
	if lv := ge2.Trade.activeRoutes["local_barter"].Level; lv != 1 {
		t.Errorf("level after a load = %d", lv)
	}
}

func TestRouteUpgrade_Refusals(t *testing.T) {
	ge, good := upgradeTown(t)
	if err := ge.UpgradeTradeRoute("no_such_route"); err == nil {
		t.Error("an unknown route was upgraded")
	}
	ge.Resources.resources[good].Amount = 0
	if err := ge.UpgradeTradeRoute("local_barter"); err == nil || ge.Trade.activeRoutes["local_barter"].Level != 0 {
		t.Error("an upgrade was bought without the good")
	}
	ge.Resources.resources[good].Amount = 1e12
	if err := ge.StopTradeRoute("local_barter"); err != nil {
		t.Fatal(err)
	}
	if err := ge.UpgradeTradeRoute("local_barter"); err == nil || ge.Trade.activeRoutes["local_barter"].Level != 0 {
		t.Error("a paused route was upgraded")
	}
}

func TestRouteUpgrade_ExportStaysWithinTheSurplus(t *testing.T) {
	ge, _ := flowEngine(t, "local_barter", 5000)
	ge.Trade.activeRoutes["local_barter"].Level = 60
	food := ge.Resources.resources["food"]
	food.Amount = 3
	for key := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
	}
	ge.recalculateRates()
	pre := food.Rate - food.Breakdown.TradeFlow
	if pre+food.Amount >= 0 {
		if left := food.Amount + food.Rate; left < -1e-9 {
			t.Errorf("a high level took the food below what the town uses: %v left", left)
		}
	} else if food.Breakdown.TradeFlow != 0 {
		t.Errorf("the town runs short of food, yet the route moved %v", food.Breakdown.TradeFlow)
	}
}
