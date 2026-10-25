package game

import (
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// Game speed is fixed at 1x so the calendar paces the game. The player speed
// setting is retired (wonders used to raise its cap by 0.5x each, 7x by the
// Atomic Age), and only the dev console's /speed goes past 1x. These tests
// hold the pieces that keep a player at 1x.

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

// TestDevSpeedNeedsDevMode: the dev console's /speed still sets the speed
// multiplier, and does nothing at all without dev mode.
func TestDevSpeedNeedsDevMode(t *testing.T) {
	prev := DevModeActive
	t.Cleanup(func() { DevModeActive = prev })
	ge := newSeededEngine(1)
	base := ge.getTickInterval()

	DevModeActive = false
	if got := DevConsoleCommand("/speed 4", ge); got != "" {
		t.Errorf("/speed without dev mode replied %q, want nothing", got)
	}
	if ge.speedMultiplier != 1 || ge.getTickInterval() != base {
		t.Errorf("/speed without dev mode changed the speed to %vx (interval %v)", ge.speedMultiplier, ge.getTickInterval())
	}

	DevModeActive = true
	if got, want := DevConsoleCommand("/speed 4", ge), "tick speed set to 4x"; got != want {
		t.Errorf("/speed 4 with dev mode replied %q, want %q", got, want)
	}
	if ge.speedMultiplier != 4 {
		t.Errorf("speed after /speed 4 = %vx, want 4x", ge.speedMultiplier)
	}
	if got, want := ge.getTickInterval(), base/4; got != want {
		t.Errorf("tick interval at 4x = %v, want %v", got, want)
	}
}
