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
