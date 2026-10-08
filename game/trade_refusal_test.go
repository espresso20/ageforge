package game

import "testing"

// TestExchangeRefusalSaysHowFarShort: amounts print to three figures, so a
// trade refused for a fraction used to read "to give 247 (you have 247)".
// When the two numbers print the same the refusal says how far short
// instead; otherwise it keeps its words.
func TestExchangeRefusalSaysHowFarShort(t *testing.T) {
	ge := newSeededEngine(3)
	for _, a := range []string{"stone_age", "bronze_age"} {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	ge.mu.Lock()
	ge.Buildings.counts["market"] = 1
	ge.Resources.resources["wood"].Storage = 1e6
	ge.mu.Unlock()
	setResource(ge, "wood", 246.9)
	if _, err := ge.ExchangeResources("wood", "gold", 247.3); err == nil || err.Error() != "Not enough wood to give 247: you are 0.4 short." {
		t.Errorf("selling 247.3 wood while holding 246.9: %v", err)
	}
	if _, err := ge.ExchangeResources("wood", "gold", 400); err == nil || err.Error() != "Not enough wood to give 400 (you have 247)." {
		t.Errorf("selling 400 wood while holding 246.9: %v", err)
	}
	if _, err := ge.ExchangeResources("wood", "gold", 246.9); err != nil {
		t.Errorf("selling exactly what is held: %v", err)
	}
}

// TestMarketClosesGoldForWhatTheAgeBuildsWith: from the Digital Age on no
// building costs gold, and the Information Age's hubs still make it by the
// hundreds of millions. The listed gold pairs no longer turn it into what
// the age builds with (data, crypto); it still buys what no age prices, and
// data still buys gold. Nanobots are a build material: the market trades
// none.
func TestMarketClosesGoldForWhatTheAgeBuildsWith(t *testing.T) {
	ge := catEngine(t, "digital_age", 3)
	for _, a := range ageKeys() {
		ge.applyAgeUnlocks(a)
		if a == "digital_age" {
			break
		}
	}
	ge.mu.Lock()
	ge.Buildings.counts["market"] = 1
	ge.mu.Unlock()
	setStock(ge, map[string][2]float64{"gold": {1e12, 1e15}, "data": {1e9, 1e15}, "knowledge": {0, 1e15}, "nanobots": {1e6, 1e15}, "electricity": {1e12, 1e15}})

	if got, err := ge.ExchangeResources("gold", "data", 1e9); err == nil || got != 0 || err.Error() != "The market does not trade gold for data in this age. Type trade list to see the rates." {
		t.Errorf("gold for data in the Digital Age: got %v, err %v; want it refused", got, err)
	}
	if got, err := ge.ExchangeResources("nanobots", "electricity", 1000); err == nil || got != 0 {
		t.Errorf("nanobots for electricity: got %v, err %v; want it refused", got, err)
	}
	if got, err := ge.ExchangeResources("gold", "knowledge", 1000); err != nil || got < 2500 {
		t.Errorf("gold for knowledge: got %v, err %v; want the typed rate", got, err)
	}
	if got, err := ge.ExchangeResources("electricity", "data", 1e9); err != nil || got <= 0 {
		t.Errorf("electricity for data, both built with: got %v, err %v; want parity", got, err)
	}
	st := ge.GetState()
	for key := range st.Trade.ExchangeRates {
		x := st.Trade.ExchangeRates[key]
		if x.From == "gold" && x.To == "data" || x.From == "nanobots" || x.To == "nanobots" {
			t.Errorf("the trade list offers %s for %s in the Digital Age", x.From, x.To)
		}
	}
}
