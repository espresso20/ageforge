package game

import (
	"testing"
)

// autoTown is a town with the route's age reached and the routes open, but
// nothing started.
func autoTown(t *testing.T, key string, count int) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	ge.age = "industrial_age"
	townMakes(ge)
	ge.grantedFeatures = map[string]bool{}
	for _, f := range routeFeatures(key) {
		ge.grantedFeatures[f] = true
	}
	ge.Buildings.counts[ge.Trade.routeDefs[key].RequiredBld] = count
	return ge
}

func running(ge *GameEngine, key string) bool {
	r, ok := ge.Trade.activeRoutes[key]
	return ok && !r.Paused
}

func TestAutoStart_ARouteStartsWhenItBecomesReady(t *testing.T) {
	ge := autoTown(t, "local_barter", 0)
	ge.startReadyRoutes()
	if running(ge, "local_barter") {
		t.Fatal("a route started without its building")
	}
	ge.Buildings.counts[ge.Trade.routeDefs["local_barter"].RequiredBld] = ge.Trade.routeDefs["local_barter"].MinCount
	ge.mu.Lock()
	ge.processTrade()
	ge.mu.Unlock()
	if !running(ge, "local_barter") {
		t.Fatal("a ready route did not start")
	}
}

func TestAutoStart_AStoppedRouteStaysStoppedAcrossASaveUntilStarted(t *testing.T) {
	ge := autoTown(t, "local_barter", 2)
	ge.startReadyRoutes()
	if err := ge.StopTradeRoute("local_barter"); err != nil {
		t.Fatal(err)
	}
	ge.startReadyRoutes()
	if running(ge, "local_barter") {
		t.Fatal("an auto-start took up a route the player stopped")
	}
	if err := ge.SaveGame("autostart_stop"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("autostart_stop"); err != nil {
		t.Fatal(err)
	}
	ge2.startReadyRoutes()
	if r, ok := ge2.Trade.activeRoutes["local_barter"]; !ok || !r.Paused {
		t.Fatal("the stop was not saved")
	}
	ge2.grantedFeatures = ge.grantedFeatures
	if err := ge2.StartTradeRoute("local_barter"); err != nil {
		t.Fatal(err)
	}
	if !running(ge2, "local_barter") {
		t.Fatal("start did not clear the stop")
	}
}

func TestAutoStart_ALostBuildingHoldsTheRouteAndItResumesOnItsOwn(t *testing.T) {
	ge := autoTown(t, "local_barter", 2)
	def := ge.Trade.routeDefs["local_barter"]
	ge.startReadyRoutes()
	ge.Buildings.counts[def.RequiredBld] = 0
	ge.recalculateRates()
	ge.startReadyRoutes()
	if r := ge.Trade.activeRoutes["local_barter"]; r.Paused || !r.NoBuilding {
		t.Fatalf("a route with no building: paused %v, held %v", r.Paused, r.NoBuilding)
	}
	ge.Buildings.counts[def.RequiredBld] = def.MinCount
	ge.recalculateRates()
	if r := ge.Trade.activeRoutes["local_barter"]; r.NoBuilding || r.Run <= 0 {
		t.Error("a route did not resume with its building back")
	}
}
