package game

import (
	"encoding/json"
	"math"
	"os"
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

// TestLoadClampsSavedSpeed: a save written at 7x (the cap a player had by
// the Atomic Age) loads at 1x, and its offline catch-up counts the time away
// at 1x too. The save field stays; only the loaded value is clamped.
func TestLoadClampsSavedSpeed(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(1)
	ge.mu.Lock()
	ge.speedMultiplier = 7
	save := ge.buildSaveSnapshot()
	ge.mu.Unlock()
	if save.SpeedMultiplier != 7 {
		t.Fatalf("the save carries %vx, want the 7x it was written at", save.SpeedMultiplier)
	}
	// Written an hour ago and signed like SaveGame signs, so the load runs
	// its offline catch-up and the signature check passes.
	save.Timestamp = time.Now().Add(-time.Hour)
	save.Signature = signSave(save, saveHMACKey)
	data, err := json.Marshal(save)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath("seven_x"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	b := NewGameEngine()
	if err := b.LoadGame("seven_x"); err != nil {
		t.Fatal(err)
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.speedMultiplier != 1 {
		t.Errorf("loaded speed = %vx, want 1x", b.speedMultiplier)
	}
	if b.cheaterBadge {
		t.Error("the signed save was flagged as tampered")
	}
	if got, want := b.tickIntervalLocked(), oneXInterval(b.tickSpeedBonus); got != want {
		t.Errorf("tick interval after load = %v, want the 1x %v", got, want)
	}
	// An hour at 1x is 1,800 two-second ticks; at 7x it would be 12,600.
	// The slack covers a slow machine between writing and loading the save.
	if got := b.tick - save.Tick; got < 1800 || got > 1805 {
		t.Errorf("an hour away caught up %d ticks, want 1800 (the 1x count)", got)
	}
}

// TestClampPlayerSpeed: whatever a save carries, a player loads at 1x.
func TestClampPlayerSpeed(t *testing.T) {
	for _, saved := range []float64{0, -3, 0.5, 1, 1.5, 7, 12, 1e9, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if got := clampPlayerSpeed(saved); got != 1 {
			t.Errorf("clampPlayerSpeed(%v) = %v, want 1", saved, got)
		}
	}
}

// TestWondersDoNotRaiseSpeed: each wonder used to raise the speed cap by
// 0.5x. Now every wonder built leaves the game at 1x: the tick interval does
// not move, and a save carrying 1.5x (what one wonder allowed) loads at 1x.
func TestWondersDoNotRaiseSpeed(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(1)
	ge.mu.Lock()
	before := ge.tickIntervalLocked()
	wonders := 0
	for key, def := range ge.Buildings.defs {
		if def.Category == "wonder" {
			ge.Buildings.counts[key] = 1
			wonders++
		}
	}
	ge.recalculateRates()
	ge.recalculateTickSpeed()
	after := ge.tickIntervalLocked()
	ge.speedMultiplier = 1.5
	ge.mu.Unlock()
	if wonders == 0 {
		t.Fatal("no wonders in the building defs")
	}
	if after != before {
		t.Errorf("%d wonders built moved the tick interval from %v to %v", wonders, before, after)
	}

	if err := ge.SaveGame("wonders"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("wonders"); err != nil {
		t.Fatal(err)
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	built := 0
	for key, def := range b.Buildings.defs {
		if def.Category == "wonder" && b.Buildings.counts[key] > 0 {
			built++
		}
	}
	if built != wonders {
		t.Fatalf("%d wonders after the load, want %d", built, wonders)
	}
	if b.speedMultiplier != 1 {
		t.Errorf("with every wonder built, a 1.5x save loads at %vx, want 1x", b.speedMultiplier)
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
