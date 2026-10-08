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
