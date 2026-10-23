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

// TestProductionPoolsApplyInFullLikeConfig: the engine applies its production
// pools in full, 1 + the pool, the way config.IncomeFactor models the
// all-production pool.
func TestProductionPoolsApplyInFullLikeConfig(t *testing.T) {
	for _, target := range []string{"production_all", "knowledge_rate", "gold_rate", "food_rate"} {
		if got, want := poolFactor(target, 3.2), 4.2; got != want {
			t.Errorf("poolFactor(%s, +320%%) = %v, want %v", target, got, want)
		}
	}
}
