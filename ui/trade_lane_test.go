package ui

import (
	"image"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

func TestLaneSpeedClamp(t *testing.T) {
	if s := laneSpeed(0); s != 2 {
		t.Fatalf("idle lane speed %v, want 2", s)
	}
	if s := laneSpeed(1e12); s != 14 {
		t.Fatalf("top lane speed %v, want 14", s)
	}
	if laneSpeed(10) <= laneSpeed(1) {
		t.Fatal("faster flow must move faster")
	}
}

// The lanes ride the panel's clock, which stops while the game is held.
func TestLaneClockStopsWhileHeld(t *testing.T) {
	keepTheme(t)
	p := newDeckPanel(func(game.GameState) deckPage { return deckPage{} })
	now, held := time.Unix(1000, 0), false
	p.now, p.held = func() time.Time { return now }, func() bool { return held }
	p.open(game.GameState{})
	step := func() float64 {
		now = now.Add(time.Second)
		p.Layout(image.Rect(0, 0, 80, 20))
		return p.clock
	}
	step()
	run := step()
	held = true
	if got := step(); got != run {
		t.Errorf("the clock moved from %v to %v while held", run, got)
	}
	held = false
	step()
	if got := step(); got <= run {
		t.Errorf("the clock stayed at %v after the hold lifted", got)
	}
}
