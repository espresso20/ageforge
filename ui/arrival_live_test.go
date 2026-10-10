package ui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// TestTheGameRunsUnderTheCelebration is the arrival screen on the running
// app, where this screen has frozen the game before (age_splash_repro_test.go).
//
// Two advances land between two refreshes of the dashboard, as they would
// for a player who was not looking. One screen comes up, for the further
// age, with the age passed listed and the town drawn from the age that was
// left. While its celebration plays, with no key pressed, the engine keeps
// ticking, the dashboard under it keeps taking the game's state and the
// screen's own frames keep moving; the celebration then gives way to the
// information by itself, and the game still runs. Two keys later the prompt
// has the keyboard, and no second screen follows.
//
// So that a slow machine has the time to look, the celebration plays at
// half speed here (its clock is the real one, slowed) and its timers are
// held: it ends on its script's last frame, as it does when a draw comes
// before the timer.
func TestTheGameRunsUnderTheCelebration(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI test; skipped with -short")
	}
	h := newReproHarness(t)
	om := h.a.dashboard.overlayMgr
	h.sim.SetSize(120, 40)
	h.onUI(func() {
		om.after = func(time.Duration, func()) func() { return func() {} }
	})

	// Ticks as fast as the game allows from the next one on, so that
	// several fall inside the celebration.
	h.dev("/speed 10")

	h.quietEra()
	if !h.grantNextAge() {
		t.Fatal("no age to advance into")
	}
	h.onUI(func() {
		if err := h.eng.AdvanceAge(); err != nil {
			t.Errorf("the first advance: %v", err)
			return
		}
		h.grantNextAge()
		if err := h.eng.AdvanceAge(); err != nil {
			t.Errorf("the second advance: %v", err)
		}
	})
	if got := h.eng.GetState().Age; got != "bronze_age" {
		t.Fatalf("two advances from the first age ended in %s", got)
	}

	// The first look at the screen, on the event loop: what it says, the
	// game's tick at that moment, and its clock slowed from here on.
	var a *arrival
	var view arrivalView
	var from, into string
	stage, frame0, seen0, tick0 := arrClosed, 0, 0, 0
	h.waitFor("the arrival screen", 5*time.Second, func() bool {
		h.onUI(func() {
			if a = om.arrival; a == nil {
				return
			}
			view, stage = a.view, a.stage
			if a.oldTown != nil && a.newTown != nil {
				from, into = a.oldTown.age, a.newTown.age
			}
			if a.sc != nil {
				frame0 = a.sc.frame
			}
			if om.seen != nil {
				seen0 = om.seen.Tick
			}
			tick0 = h.eng.GetState().Tick
			began, slowed := a.start, time.Now()
			a.now = func() time.Time { return slowed.Add(time.Since(slowed) / 2) }
			a.start = began // the frames played so far stay played
		})
		return a != nil
	})
	if stage != arrCelebrating {
		t.Fatalf("the screen did not come up on its celebration (stage %d)\n%s", stage, h.describeUI())
	}
	if view.age != "Bronze Age" || view.passed != "Ages passed: Stone Age" || view.epoch {
		t.Errorf("the screen is for %q, %q (epoch %v)", view.age, view.passed, view.epoch)
	}
	if from != "Primitive Age" || into != "Bronze Age" {
		t.Errorf("the town turns from %q into %q", from, into)
	}

	// While the celebration plays: each look is taken on the event loop,
	// so what it reads belongs to the stage it saw.
	ticks, frames := 0, 0
	began := time.Now()
	for stage == arrCelebrating {
		if time.Since(began) > 30*time.Second {
			t.Fatalf("the celebration never ended\n%s", h.describeUI())
		}
		time.Sleep(40 * time.Millisecond)
		h.onUI(func() {
			if stage = a.stage; stage != arrCelebrating {
				return
			}
			ticks = h.eng.GetState().Tick - tick0
			if a.sc != nil {
				frames = a.sc.frame - frame0
			}
		})
	}
	if ticks < 1 {
		t.Errorf("the engine did not tick while the celebration played")
	}
	if frames < 4 {
		t.Errorf("the celebration moved %d frames", frames)
	}

	// It gave way to the information by itself, and the game still runs
	// and the dashboard under the screen still takes its state.
	auto := false
	h.onUI(func() { stage, auto = a.stage, a.auto })
	if stage != arrInforming || !auto || !h.hasPage(arrivalPageName) {
		t.Fatalf("the celebration did not give way to the information by itself (stage %d)\n%s", stage, h.describeUI())
	}
	tick1 := h.eng.GetState().Tick
	h.waitFor("a tick under the information", 6*time.Second, func() bool { return h.eng.GetState().Tick > tick1 })
	h.waitFor("the dashboard to take the game's state under the screen", 6*time.Second, func() bool {
		seen := 0
		h.onUI(func() {
			if om.seen != nil {
				seen = om.seen.Tick
			}
		})
		return seen > seen0 && h.hasPage(arrivalPageName)
	})

	h.dismissArrival("the information", tcell.KeyEnter, 0)
	h.waitFor("the prompt to have the keyboard", 3*time.Second, func() bool {
		return h.frontPage() == "dashboard" && h.inputHasFocus()
	})
	// One screen for two advances: the dashboard has refreshed again by now.
	time.Sleep(700 * time.Millisecond)
	if h.hasPage(arrivalPageName) {
		t.Errorf("a second screen followed the first\n%s", h.describeUI())
	}
}
