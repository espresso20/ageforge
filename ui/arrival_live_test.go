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
func TestTheGameRunsUnderTheCelebration(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI test; skipped with -short")
	}
	h := newReproHarness(t)
	om := h.a.dashboard.overlayMgr

	// Ticks every half second, so that several fall inside the celebration.
	h.dev("/speed 4")
	t0 := h.eng.GetState().Tick
	h.waitFor("a tick at the new speed", 6*time.Second, func() bool { return h.eng.GetState().Tick > t0+1 })

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
	h.waitFor("the arrival screen", 3*time.Second, func() bool { return h.hasPage(arrivalPageName) })

	var a *arrival
	var view arrivalView
	var from, into string
	frame0, seen0 := 0, 0
	h.onUI(func() {
		if a = om.arrival; a == nil {
			return
		}
		view = a.view
		if a.oldTown != nil && a.newTown != nil {
			from, into = a.oldTown.age, a.newTown.age
		}
		if a.sc != nil {
			frame0 = a.sc.frame
		}
		if om.seen != nil {
			seen0 = om.seen.Tick
		}
	})
	if a == nil || h.arrivalStage() != arrCelebrating {
		t.Fatalf("the screen did not come up on its celebration\n%s", h.describeUI())
	}
	if view.age != "Bronze Age" || view.passed != "Ages passed: Stone Age" || view.epoch {
		t.Errorf("the screen is for %q, %q (epoch %v)", view.age, view.passed, view.epoch)
	}
	if from != "Primitive Age" || into != "Bronze Age" {
		t.Errorf("the town turns from %q into %q", from, into)
	}

	// While the celebration plays.
	tick0 := h.eng.GetState().Tick
	ticks, refreshed, frames := 0, 0, 0
	began := time.Now()
	for h.arrivalStage() == arrCelebrating {
		if time.Since(began) > 15*time.Second {
			t.Fatalf("the celebration never ended\n%s", h.describeUI())
		}
		ticks = h.eng.GetState().Tick - tick0
		h.onUI(func() {
			if a.sc != nil {
				frames = a.sc.frame - frame0
			}
			if om.seen != nil {
				refreshed = om.seen.Tick - seen0
			}
		})
		time.Sleep(40 * time.Millisecond)
	}
	if ticks < 2 {
		t.Errorf("the engine ticked %d times while the celebration played", ticks)
	}
	if refreshed < 1 {
		t.Errorf("the dashboard under the celebration did not take the game's state")
	}
	if frames < 8 {
		t.Errorf("the celebration moved %d frames", frames)
	}

	// It gave way to the information by itself, and the game still runs.
	if h.arrivalStage() != arrInforming || !h.hasPage(arrivalPageName) {
		t.Fatalf("the celebration did not give way to the information\n%s", h.describeUI())
	}
	tick1 := h.eng.GetState().Tick
	h.waitFor("a tick under the information", 4*time.Second, func() bool { return h.eng.GetState().Tick > tick1 })

	h.dismissArrival("the information", tcell.KeyEnter, 0)
	h.waitFor("the prompt to have the keyboard", 2*time.Second, func() bool {
		return h.frontPage() == "dashboard" && h.inputHasFocus()
	})
	// One screen for two advances: the dashboard has refreshed again by now.
	time.Sleep(700 * time.Millisecond)
	if h.hasPage(arrivalPageName) {
		t.Errorf("a second screen followed the first\n%s", h.describeUI())
	}
}
