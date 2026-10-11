package ui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// film is the film of a run's ending on the harness's app, nil when none
// is up.
func (h *reproHarness) film() (e *ending) {
	h.onUI(func() { e = h.a.dashboard.overlayMgr.film })
	return e
}

// TestTheNewRunTicksUnderTheFilm is a prestige on the running app, typed at
// the prompt, with the film left to play: no key is pressed. One film comes
// up, on the ending, with the town that was given up and the new run's
// land. While it plays through its four beats the new run's ticks advance
// from zero, the dashboard under it keeps taking the game's state and the
// film's own frames keep moving. It closes by itself, the prompt has the
// keyboard, the new run is in its first age at the next prestige level,
// and nothing else is left on the screen. This is the one film that plays
// moving under the race detector.
func TestTheNewRunTicksUnderTheFilm(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI test; skipped with -short")
	}
	h := newReproHarness(t)
	om := h.a.dashboard.overlayMgr
	h.sim.SetSize(120, 40)
	h.dev("/age medieval_age")
	h.waitFor("the dashboard to see the Medieval Age", 3*time.Second, func() bool {
		seen := ""
		h.onUI(func() {
			if om.seen != nil {
				seen = om.seen.Age
			}
		})
		return seen == "medieval_age"
	})
	level := h.eng.GetState().Prestige.Level
	h.submit("prestige confirm yes")

	var e *ending
	var kind, from, into string
	beat, frame0, tick0 := -1, 0, 0
	h.waitFor("the film", 5*time.Second, func() bool {
		h.onUI(func() {
			if e = om.film; e == nil {
				return
			}
			kind, beat = e.view.end.Kind, e.beat().beat
			if e.oldTown != nil && e.newTown != nil {
				from, into = e.oldTown.age, e.newTown.age
			}
			if e.sc != nil {
				frame0 = e.sc.frame
			}
			tick0 = h.eng.GetState().Tick
		})
		return e != nil
	})
	if kind != game.RunEndPrestige || beat != beatEnding || h.frontPage() != endingPageName {
		t.Fatalf("the film did not come up on a prestige's ending: kind %q, beat %d\n%s", kind, beat, h.describeUI())
	}
	if from != "Medieval Age" || into != "Primitive Age" {
		t.Errorf("the film goes from %q to %q", from, into)
	}

	// While it plays: every look is taken on the event loop.
	seen := map[int]bool{}
	ticks, frames, refreshed, closed := 0, 0, false, false
	began := time.Now()
	for !closed {
		if time.Since(began) > 40*time.Second {
			t.Fatalf("the film never ended\n%s", h.describeUI())
		}
		time.Sleep(60 * time.Millisecond)
		h.onUI(func() {
			if closed = e.closed; closed {
				return
			}
			seen[e.beat().beat] = true
			ticks = h.eng.GetState().Tick - tick0
			if e.sc != nil {
				frames = e.sc.frame - frame0
			}
			if om.seen != nil && om.seen.Age == "primitive_age" && om.seen.Tick > tick0 {
				refreshed = true
			}
		})
	}
	for _, b := range []int{beatEnding, beatReckoning, beatStrike, beatBeginning} {
		if !seen[b] {
			t.Errorf("the film never showed beat %d", b)
		}
	}
	if ticks < 2 {
		t.Errorf("the new run ticked %d times while the film played", ticks)
	}
	if !refreshed {
		t.Error("the dashboard under the film did not take the new run's state")
	}
	if frames < 40 {
		t.Errorf("the film moved %d frames", frames)
	}
	if took := time.Since(began); took < 6*time.Second {
		t.Errorf("the film was over in %v with no key pressed", took)
	}

	// It closed by itself, and the game is there.
	h.waitFor("the prompt to have the keyboard", 3*time.Second, func() bool {
		return h.frontPage() == "dashboard" && h.inputHasFocus()
	})
	if st := h.eng.GetState(); st.Age != "primitive_age" || st.Prestige.Level != level+1 {
		t.Errorf("after the film: %s at prestige level %d", st.Age, st.Prestige.Level)
	}
	time.Sleep(700 * time.Millisecond)
	if h.hasPage(endingPageName) || h.hasPage(arrivalPageName) || h.film() != nil {
		t.Errorf("something followed the film\n%s", h.describeUI())
	}
}

// TestReproFallFilm is the other way a run ends, through the windows that
// lead to it, with the screen held still: the arrival of an age, then the
// catastrophe's window (which waits for the arrival), then Succumb, then
// the film of the fall (which the next run's first window would wait for).
// Each key must reach the film, and the last gives the prompt the keyboard.
func TestReproFallFilm(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI test; skipped with -short")
	}
	h := newReproHarness(t)
	h.still()
	h.onUI(func() {
		if err := h.eng.EnterAgeForTest("iron_age"); err != nil {
			t.Errorf("entering the Iron Age: %v", err)
		}
		if err := h.eng.ForceCatastropheForTest(); err != nil {
			t.Errorf("forcing the catastrophe: %v", err)
		}
	})
	h.waitFor("the arrival screen", 3*time.Second, func() bool { return h.hasPage(arrivalPageName) })
	time.Sleep(550 * time.Millisecond)
	if h.hasPage("catastrophe") {
		t.Fatalf("the catastrophe's window opened over the arrival\n%s", h.describeUI())
	}
	h.dismissArrival("into the Iron Age", tcell.KeyEnter, 0)
	h.waitFor("the catastrophe's window", 3*time.Second, func() bool { return h.frontPage() == "catastrophe" })
	h.key(tcell.KeyRune, 's')
	h.waitFor("the film of the fall", 3*time.Second, func() bool { return h.film() != nil })
	e := h.film()
	var kind string
	var beats int
	look := func() (beat, keys int, closed bool) {
		h.onUI(func() { beat, keys, closed = e.beat().beat, e.keys, e.closed })
		return
	}
	h.onUI(func() { kind, beats = e.view.end.Kind, len(e.spans()) })
	if kind != game.RunEndFallen || beats != 3 || h.frontPage() != endingPageName || h.hasPage("catastrophe") {
		t.Fatalf("the film of a fall: kind %q, %d beats\n%s", kind, beats, h.describeUI())
	}
	if st := h.eng.GetState(); st.Age != "primitive_age" || st.PendingCatastrophe != "" {
		t.Fatalf("after Succumb: %s, pending %q", st.Age, st.PendingCatastrophe)
	}
	// A key a beat, each a press of its own, until the film is closed.
	for _, want := range []int{beatReckoning, beatBeginning, -1} {
		time.Sleep(arrivalKeyGap + 20*time.Millisecond)
		_, before, _ := look()
		h.key(tcell.KeyEnter, 0)
		deadline := time.Now().Add(2 * time.Second)
		for {
			beat, keys, closed := look()
			if keys > before {
				if want >= 0 && (closed || beat != want) || want < 0 && !closed {
					t.Fatalf("a key took the film to beat %d (closed %v), want %d\n%s", beat, closed, want, h.describeUI())
				}
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("FROZEN: a key did not reach the film within 2s\n%s", h.describeUI())
			}
			time.Sleep(15 * time.Millisecond)
		}
	}
	h.waitFor("the prompt to have the keyboard", 3*time.Second, func() bool {
		return h.frontPage() == "dashboard" && h.inputHasFocus()
	})
}

// TestDashboardPlaysTheFilm: the dashboard takes the record of a run's
// ending from the engine and plays its film at the next refresh, from the
// last picture it had of the run that ended. The Last Passage's window
// comes before it, as the choice does; while the film is up nothing else
// is put over it; and one ending is one film.
func TestDashboardPlaysTheFilm(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	eng.SeedRNG(3)
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	om := d.overlayMgr
	om.engine = eng
	om.after = func(time.Duration, func()) func() { return func() {} }

	if err := eng.ForceLastPassageForTest("interstellar_age"); err != nil {
		t.Fatal(err)
	}
	d.refresh()
	if !pages.HasPage(catastrophePage) || om.film != nil {
		t.Fatal("the Last Passage's window did not open, or a film played before the choice")
	}
	full := eng.GetState().LastPassage.PointsNow
	if err := eng.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	d.closeCatastropheModal()
	d.refresh()
	e := om.film
	if e == nil || om.ActiveName() != endingPageName || !om.FullScreenUp() {
		t.Fatalf("no film after the run ended (active %q)", om.ActiveName())
	}
	end := e.view.end
	if end.Kind != game.RunEndEndured || end.Age != "interstellar_age" || end.Full != full || end.Points <= 0 || end.Points >= full {
		t.Errorf("the film is of %s from %s with %d of %d points (the game had %d)", end.Kind, end.Age, end.Points, end.Full, full)
	}
	if e.oldTown == nil || e.oldTown.age != "Interstellar Age" || e.newTown == nil || e.newTown.age != "Primitive Age" {
		t.Error("the film does not go from the town that was to the new land")
	}
	if len(e.view.lines) < 3 || e.view.lines[0].kind != elVerdict || e.view.complete == "" || len(e.view.captions) == 0 {
		t.Errorf("the film says %d lines, the verdict first %v", len(e.view.lines), len(e.view.lines) > 0 && e.view.lines[0].kind == elVerdict)
	}
	// Refreshes while it plays: the same film, and nothing over it.
	for i := 0; i < 3; i++ {
		d.refresh()
	}
	if front, _ := pages.GetFrontPage(); om.film != e || e.closed || front != endingPageName || pages.HasPage(catastrophePage) {
		t.Errorf("the film did not stay in front (front %q)", front)
	}
	e.close()
	d.refresh()
	if om.film != nil || om.ActiveName() != "" || pages.HasPage(endingPageName) {
		t.Error("the film played again, or did not close")
	}
}
