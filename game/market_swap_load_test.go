package game

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// A save from before the market lost its swap, loaded by the game after it.
// The fixture (testdata/pre_market_swap_bronze.json) was written and signed
// by the old game (market_swap_fixture_gen_test.go says how): a Bronze Age
// town that had swapped resources at its markets (supply pressure, exchange
// totals, sold and bought in the save), two sell orders in its plan between
// two builds, trade rows in its run's plan log and in the legacy plan
// template, a running trade route, and a civilization's saved deals.
const marketSwapFixture = "testdata/pre_market_swap_bronze.json"

func marketSwapFixtureSave(t *testing.T) GameSave {
	t.Helper()
	data, err := os.ReadFile(marketSwapFixture)
	if err != nil {
		t.Fatal(err)
	}
	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		t.Fatal(err)
	}
	if len(save.Trade.SupplyPressure) == 0 || len(save.Trade.TotalExchanged) == 0 ||
		len(save.Trade.TotalSold) == 0 || len(save.Trade.TotalBought) == 0 {
		t.Fatal("setup: the fixture holds no swap history")
	}
	sellOrders := 0
	for _, it := range save.Plan {
		if it.Kind == "trade" {
			sellOrders++
		}
	}
	if sellOrders != 2 {
		t.Fatalf("setup: the fixture's plan holds %d sell orders, want 2", sellOrders)
	}
	return save
}

// countLines is how many log lines say exactly msg.
func countLines(ge *GameEngine, msg string) int {
	n := 0
	for _, l := range ge.GetLogs() {
		if l.Message == msg {
			n++
		}
	}
	return n
}

// TestMarketSwapSave_LoadsDropsTheSellOrdersAndSaysSoOnce: the old save
// loads with its signature intact, its sell orders gone and its builds kept
// in order, and the plan template's trade rows skipped without a line; the
// notice is on the first load only; saved again, the swap's fields are
// written empty and the save is signed and valid.
func TestMarketSwapSave_LoadsDropsTheSellOrdersAndSaysSoOnce(t *testing.T) {
	isolateAccountDir(t)
	save := marketSwapFixtureSave(t)
	// Loaded as if just saved: the fixture's own old timestamp would play a
	// day of offline catch-up, and with stock uncapped the plan's builds
	// would be bought along the way; this test is about the load.
	ge := loadFixtureAway(t, marketSwapFixture, "swap", time.Second)
	if ge.cheaterBadge {
		t.Fatal("the old game's signature no longer verifies")
	}

	// The plan: the two builds, in order, and no sell order.
	var got []string
	for _, it := range ge.plan {
		if it.Kind == "trade" {
			t.Errorf("a sell order is still in the plan: %+v", it)
		}
		got = append(got, it.Kind+" "+it.Key)
	}
	if want := "build farm,build lumber_mill"; strings.Join(got, ",") != want {
		t.Errorf("the plan after the load is %v, want %s", got, want)
	}

	// The plan template and the run's plan log: no trade row, the builds stay.
	templates := map[string][]PlanTemplateItem{"the run's plan log": ge.planLog, "the legacy plan template": ge.Prestige.legacyPlan}
	for name, rows := range templates {
		builds := 0
		for _, row := range rows {
			if row.Kind == "trade" {
				t.Errorf("%s still holds a trade row: %+v", name, row)
			}
			if row.Kind == PlanBuild {
				builds++
			}
		}
		if builds != 2 {
			t.Errorf("%s holds %d build rows after the load, want 2", name, builds)
		}
	}

	// The line: once, and only the one line about the market.
	if n := countLines(ge, marketSwapNotice); n != 1 {
		t.Fatalf("the notice appears %d times on the first load, want 1", n)
	}
	if marketSwapNotice != "The market no longer swaps resources. The sell orders in your plan were removed." {
		t.Errorf("the notice reads %q", marketSwapNotice)
	}
	// The template rows went without a line of their own.
	for _, l := range ge.GetLogs() {
		if strings.Contains(l.Message, "sell order") && l.Message != marketSwapNotice {
			t.Errorf("an extra line about the sell orders: %q", l.Message)
		}
	}

	// Saved again: the notice does not come back, the swap's fields are
	// written empty, and nothing else of the town is lost.
	again := saveAndLoad(t, ge, "swap_again")
	if n := countLines(again, marketSwapNotice); n != 0 {
		t.Errorf("the notice appears %d times on the second load, want 0", n)
	}
	data, err := os.ReadFile(savePath("swap_again"))
	if err != nil {
		t.Fatal(err)
	}
	var saved GameSave
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Signature == "" || again.cheaterBadge {
		t.Errorf("the re-saved game is not signed and valid (signature %q, cheater %v)", saved.Signature, again.cheaterBadge)
	}
	if len(saved.Trade.SupplyPressure) != 0 || len(saved.Trade.TotalExchanged) != 0 || len(saved.Trade.TotalSold) != 0 || len(saved.Trade.TotalBought) != 0 {
		t.Errorf("the re-saved trade state still holds the swap: %+v", saved.Trade)
	}
	var raw struct {
		Trade map[string]json.RawMessage `json:"trade"`
		Plan  []PlanItem                 `json:"plan"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"supply_pressure", "total_exchanged"} {
		if v, ok := raw.Trade[key]; !ok || string(v) != "{}" {
			t.Errorf("trade.%s is %s in the re-saved game, want it written empty ({})", key, v)
		}
	}
	for _, key := range []string{"total_sold", "total_bought"} {
		if _, ok := raw.Trade[key]; ok {
			t.Errorf("trade.%s is written in the re-saved game, want it left out", key)
		}
	}
	if len(saved.Plan) != 2 {
		t.Errorf("the re-saved plan holds %d items, want the 2 builds", len(saved.Plan))
	}

	// What the swap never touched is as the old save had it: the route keeps
	// running, the civilization's saved offers are the same, and the builds
	// are as many.
	if _, running := again.Trade.activeRoutes["local_barter"]; !running {
		t.Error("the running trade route did not survive")
	}
	for key, want := range save.Diplomacy.Factions {
		if len(want.Deals) != len(saved.Diplomacy.Factions[key].Deals) {
			t.Errorf("%s: %d saved offers before, %d after", key, len(want.Deals), len(saved.Diplomacy.Factions[key].Deals))
		}
	}
	for key, n := range save.Buildings {
		if again.Buildings.GetCount(key) != n {
			t.Errorf("%s: %d copies after the load, the save has %d", key, again.Buildings.GetCount(key), n)
		}
	}
}

// TestMarketSwapSave_NoSellOrdersNoLine: a save whose plan held no sell
// orders gets no line about them, whatever else the swap left in it.
func TestMarketSwapSave_NoSellOrdersNoLine(t *testing.T) {
	isolateAccountDir(t)
	ge := loadFixture(t, marketSwapFixture, "swap")
	if n := countLines(ge, marketSwapNotice); n != 1 {
		t.Fatalf("setup: the notice appears %d times on the first load", n)
	}
	// The re-saved game still has the signed, empty swap fields; a hand-made
	// plan with no sell orders and pressure left in the file says nothing.
	save := ge.buildSaveSnapshot()
	save.Trade.SupplyPressure = map[string]float64{"wood:gold": 0.4}
	save.Trade.TotalExchanged = map[string]float64{"wood": 10}
	save.Signature = signSave(save, saveHMACKey)
	writeRawSave(t, "pressure_only", save)
	again := NewGameEngine()
	if err := again.LoadGame("pressure_only"); err != nil {
		t.Fatal(err)
	}
	if again.cheaterBadge {
		t.Error("a save with leftover pressure failed its signature check")
	}
	if n := countLines(again, marketSwapNotice); n != 0 {
		t.Errorf("the notice appears %d times for a plan with no sell orders", n)
	}
}
