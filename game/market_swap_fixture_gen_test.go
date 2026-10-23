//go:build marketswapfixture

package game

import (
	"os"
	"testing"
)

// The save the market swap's removal has to load: a Bronze Age town that
// swapped resources at its markets (supply pressure and exchange, sold and
// bought totals in the save), with sell orders in its plan, trade rows in its
// run's plan log and in the legacy plan template, a running trade route and a
// civilization's saved deals.
//
// This file builds only with -tags marketswapfixture, because it calls what
// the removal deleted (ExchangeResources, PlanAddTrade). Run it on a checkout
// of the game from before the removal, so the bytes are the old game's,
// written and signed by it:
//
//	UPDATE_FIXTURES=1 go test -tags marketswapfixture ./game -run TestWriteMarketSwapFixture
//
// then commit testdata/pre_market_swap_bronze.json.
func TestWriteMarketSwapFixture(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the pre market swap save")
	}
	isolateAccountDir(t)
	ge := newSeededEngine(81)
	for _, a := range []string{"stone_age", "bronze_age"} {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	learn(ge, "envoys", "the_wheel")
	ge.mu.Lock()
	ge.Buildings.counts["market"] = 2
	ge.Buildings.counts["hut"] = 6
	ge.mu.Unlock()
	setResource(ge, "wood", 60000)
	setResource(ge, "stone", 20000)
	setResource(ge, "gold", 3000)
	for _, s := range []struct {
		give, get string
		amount    float64
	}{{"wood", "gold", 1200}, {"wood", "gold", 900}, {"wood", "stone", 800}, {"stone", "wood", 300}} {
		if _, err := ge.ExchangeResources(s.give, s.get, s.amount); err != nil {
			t.Fatal(err)
		}
	}
	// A route running, so the save holds one.
	def := ge.Trade.routeDefs["local_barter"]
	ge.mu.Lock()
	ge.Buildings.counts[def.RequiredBld] = def.MinCount
	ge.mu.Unlock()
	if err := ge.StartTradeRoute("local_barter"); err != nil {
		t.Fatal(err)
	}
	// Two sell orders (one with an amount, one open-ended) between two builds.
	if _, err := ge.PlanAddBuild("farm", 2); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddTrade("wood", "stone", 1500); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddTrade("stone", "gold", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("lumber_mill", 1); err != nil {
		t.Fatal(err)
	}
	// Offers a civilization had rolled under the old rates.
	meet(ge, "riverlands_tribes", 0)
	// The legacy plan template carries trade rows too.
	ge.mu.Lock()
	ge.Prestige.legacyPlan = []PlanTemplateItem{
		{Age: "stone_age", Kind: PlanBuild, Key: "stone_pit", Count: 3},
		{Age: "bronze_age", Kind: PlanTrade, Key: "wood", To: "gold", Amount: 400},
		{Age: "bronze_age", Kind: PlanBuild, Key: "farm", Count: 2},
	}
	ge.mu.Unlock()
	if err := ge.SaveGame("pre_market_swap_bronze"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(savePath("pre_market_swap_bronze"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("testdata/pre_market_swap_bronze.json", data, 0o644); err != nil {
		t.Fatal(err)
	}
}
