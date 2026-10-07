package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestPacingTickMatchesEngine: config derives every payback, build-time and
// research-time cap from its own tick length, so it must be the engine's.
func TestPacingTickMatchesEngine(t *testing.T) {
	if got := BaseTickInterval.Seconds(); got != config.TickSeconds {
		t.Fatalf("config.TickSeconds = %v, engine BaseTickInterval = %vs", config.TickSeconds, got)
	}
}

// TestSoftCapMatchesConfig: the engine applies its production pools through
// the soft cap its ruleset carries, and the core ruleset carries config's:
// the one config.IncomeFactor models the all-production pool with.
func TestSoftCapMatchesConfig(t *testing.T) {
	ge := NewGameEngine()
	soft := ge.rules.SoftCap()
	if soft != config.ProductionSoftCap() {
		t.Fatalf("the engine's soft cap is %+v, config's %+v", soft, config.ProductionSoftCap())
	}
	if soft.Knee != config.ProductionKnee || soft.PastKnee != config.ProductionPastKnee {
		t.Fatalf("config.ProductionSoftCap() = %+v, its constants %v and %v", soft, config.ProductionKnee, config.ProductionPastKnee)
	}
	for _, target := range []string{"production_all", "knowledge_rate", "gold_rate", "food_rate"} {
		if got, want := poolFactor(target, 3.2, soft), 1+soft.Applied(3.2); got != want {
			t.Errorf("poolFactor(%s, +320%%) = %v, want %v", target, got, want)
		}
	}
}

// TestMarketTradesAtParity: two construction resources of the current age
// trade at the age's parity less the fee, including pairs no one listed.
func TestMarketTradesAtParity(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.age = "space_age"
	ge.Buildings.counts["asteroid_market"] = 1
	ge.Resources.UnlockResource("steel")
	ge.Resources.UnlockResource("titanium")
	ge.Resources.resources["steel"].Storage = 1e18
	ge.Resources.resources["titanium"].Storage = 1e18
	ge.Resources.Add("steel", 1e15)
	lv := config.PriceLevels("space_age")
	want := lv["titanium"] / lv["steel"] * (1 - config.ExchangeFee)
	ge.Trade.SetAge(ge.age)
	got, err := ge.Trade.Exchange("steel", "titanium", 1e12, ge.Resources, ge.Buildings, ge.tick)
	if err != nil {
		t.Fatalf("steel -> titanium: %v", err)
	}
	if r := got / 1e12; r < want*0.99 || r > want*1.01 {
		t.Errorf("steel -> titanium paid %.4g per steel, want parity less fee %.4g", r, want)
	}
}
