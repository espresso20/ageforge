package game

import (
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// oneXInterval is the tick interval at 1x for a tick_speed bonus: what a
// player runs at, whatever speed was set or saved.
func oneXInterval(bonus float64) time.Duration {
	return max(time.Duration(float64(BaseTickInterval)/(1.0+bonus)), MinTickInterval)
}

// TestPrestigeResetsSpeed: completePrestige reset every manager but not the
// speed multiplier, so a 7x run started its next run at 7x. Succumb and Reset
// already went back to 1x.
func TestPrestigeResetsSpeed(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(1)
	ge.speedMultiplier = 7
	ge.age = "modern_age"
	ge.currentEpoch = config.EpochForAge("modern_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
	if ge.Prestige.GetLevel() != 1 {
		t.Fatalf("prestige level %d after DoPrestige, want 1", ge.Prestige.GetLevel())
	}
	if ge.speedMultiplier != 1 {
		t.Errorf("speed after prestige = %vx, want 1x", ge.speedMultiplier)
	}
	if got, want := ge.getTickInterval(), oneXInterval(ge.tickSpeedBonus); got != want {
		t.Errorf("tick interval after prestige = %v, want the 1x %v", got, want)
	}
}
