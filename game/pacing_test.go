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

// TestProductionCapMatchesConfig: config.FlowIncome models the production_all
// bonus with the engine's cap.
func TestProductionCapMatchesConfig(t *testing.T) {
	if productionCap != config.ProductionAllCap {
		t.Fatalf("config.ProductionAllCap = %v, engine productionCap = %v", config.ProductionAllCap, productionCap)
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
