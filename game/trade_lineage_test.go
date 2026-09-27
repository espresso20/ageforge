package game

import "testing"

// Upgrading your markets into trading posts (the log offers it on reaching the
// Iron Age) used to shut the exchange, which only looked for "market" or
// "port". Any building of the trade lineage now opens it.
func TestExchange_AnyTradeBuildingOpensMarket(t *testing.T) {
	ge := NewGameEngine()
	ge.Resources.UnlockResource("iron")
	ge.Resources.UnlockResource("gold")
	ge.Resources.AddStorage("iron", 1e6)
	ge.Resources.AddStorage("gold", 1e6)
	ge.Resources.Add("iron", 1000)

	if _, err := ge.Trade.Exchange("iron", "gold", 100, ge.Resources, ge.Buildings, 0); err == nil {
		t.Fatal("exchange worked with no trade building at all")
	}
	ge.Buildings.counts["trading_post"] = 1 // upgraded market; no "market" left
	got, err := ge.Trade.Exchange("iron", "gold", 100, ge.Resources, ge.Buildings, 0)
	if err != nil {
		t.Fatalf("exchange with a trading post: %v", err)
	}
	if got <= 0 {
		t.Errorf("exchange returned %v gold", got)
	}
}
