package ui

// Repro harness for "age advancement splash freezes, no key dismisses it".
//
// Drives the real App (NewApp) on a tcell SimulationScreen, starts a new game
// through the real splash menu, and advances ages by typing `advance` into the
// real command input — the same path a player uses. After every advance it
// waits for the "age_splash" page, injects a key, and asserts the splash is
// removed and focus returns to the command input. Any stall of the tview event
// loop is detected via a QueueUpdate timeout and reported with a full
// goroutine dump.

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

type reproHarness struct {
	t      *testing.T
	eng    *game.GameEngine
	a      *App
	sim    tcell.SimulationScreen
	runErr chan error
}

func newReproHarness(t *testing.T) *reproHarness {
	t.Helper()
	// Written only once, before the first App exists, and never reset: a
	// stopped App's event loop can still be draining a queued refresh (which
	// reads DevModeActive) after app.Stop() returns.
	if !game.DevModeActive {
		game.DevModeActive = true
	}

	eng := game.NewGameEngine()
	a := NewApp(eng, "dev") // "dev" skips the network update check
	sim := tcell.NewSimulationScreen("UTF-8")
	a.tviewApp.SetScreen(sim) // calls sim.Init()
	sim.SetSize(200, 60)

	h := &reproHarness{t: t, eng: eng, a: a, sim: sim, runErr: make(chan error, 1)}
	go func() { h.runErr <- a.Run() }()
	t.Cleanup(func() {
		eng.Stop()
		stopped := make(chan struct{})
		go func() { a.Stop(); close(stopped) }()
		select {
		case <-stopped:
		case <-time.After(3 * time.Second):
			t.Logf("app.Stop() did not return (event loop wedged)")
		}
	})

	// Real "New Game" path: the splash List has shortcut 'n'.
	h.key(tcell.KeyRune, 'n')
	h.waitFor("dashboard visible with input focus", 5*time.Second, func() bool {
		return h.frontPage() == "dashboard" && h.inputHasFocus()
	})
	// Raise storage once, high enough for every age's resource requirements,
	// then wait for recalculateRates() (runs each tick) to apply it. (/speed is
	// avoided: it writes speedMultiplier unlocked-read by getTickInterval → race.)
	maxReq := 0.0
	for _, age := range config.Ages() {
		for _, v := range age.ResourceReqs {
			maxReq = math.Max(maxReq, v)
		}
	}
	sKey, sVal := bigStorageKey()
	for i := 0; i < int(math.Ceil(maxReq*1.5/sVal))+1; i++ {
		h.dev("/build " + sKey)
	}
	t0 := eng.GetState().Tick
	h.waitFor("storage recalculated", 10*time.Second, func() bool { return eng.GetState().Tick > t0+1 })
	return h
}

// dumpGoroutines returns the stacks of all goroutines.
func dumpGoroutines() string {
	buf := make([]byte, 1<<22)
	n := runtime.Stack(buf, true)
	return string(buf[:n])
}

// onUI runs f on the tview event loop. If the loop does not execute f within
// the timeout the event loop is blocked: dump every goroutine and fail.
func (h *reproHarness) onUI(f func()) {
	h.t.Helper()
	done := make(chan struct{})
	go func() {
		h.a.tviewApp.QueueUpdate(f)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(4 * time.Second):
		h.t.Fatalf("tview event loop BLOCKED: QueueUpdate did not run within 4s\n\n=== GOROUTINES ===\n%s", dumpGoroutines())
	}
}

func (h *reproHarness) key(k tcell.Key, r rune) {
	h.t.Helper()
	done := make(chan struct{})
	go func() {
		h.sim.InjectKey(k, r, tcell.ModNone)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(4 * time.Second):
		h.t.Fatalf("InjectKey blocked (event queue full → event loop wedged)\n\n=== GOROUTINES ===\n%s", dumpGoroutines())
	}
	// Give the poll goroutine a moment to forward the event.
	time.Sleep(15 * time.Millisecond)
}

func (h *reproHarness) typeText(s string) {
	h.t.Helper()
	for _, r := range s {
		h.key(tcell.KeyRune, r)
	}
}

func (h *reproHarness) frontPage() string {
	var name string
	h.onUI(func() { name, _ = h.a.pages.GetFrontPage() })
	return name
}

func (h *reproHarness) hasPage(name string) bool {
	var ok bool
	h.onUI(func() { ok = h.a.pages.HasPage(name) })
	return ok
}

func (h *reproHarness) inputHasFocus() bool {
	var ok bool
	h.onUI(func() { ok = h.a.tviewApp.GetFocus() == h.a.dashboard.inputField })
	return ok
}

func (h *reproHarness) inputText() string {
	var s string
	h.onUI(func() { s = h.a.dashboard.inputField.GetText() })
	return s
}

// describeUI summarises page stack + focus for diagnostics.
func (h *reproHarness) describeUI() string {
	var sb strings.Builder
	h.onUI(func() {
		fmt.Fprintf(&sb, "visible pages (front→back): %v\n", h.a.pages.GetPageNames(true))
		fmt.Fprintf(&sb, "all pages (front→back):     %v\n", h.a.pages.GetPageNames(false))
		f := h.a.tviewApp.GetFocus()
		fmt.Fprintf(&sb, "focused primitive: %T (%p)\n", f, f)
		fmt.Fprintf(&sb, "root(pages).HasFocus(): %v  (false ⇒ tview drops ALL key events)\n", h.a.pages.HasFocus())
		for _, n := range h.a.pages.GetPageNames(false) {
			fmt.Fprintf(&sb, "  page %-12q item.HasFocus()=%v\n", n, h.a.pages.GetPage(n).HasFocus())
		}
		fmt.Fprintf(&sb, "overlayMgr.active=%q\n", h.a.dashboard.overlayMgr.active)
	})
	return sb.String()
}

func (h *reproHarness) waitFor(what string, timeout time.Duration, cond func() bool) {
	h.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	h.t.Fatalf("timed out waiting for %s\n%s", what, h.describeUI())
}

func (h *reproHarness) dev(cmd string) string {
	return game.DevExecCommand(cmd, h.eng)
}

// submit types a command into the real input field and presses Enter. The
// input has autocomplete: the first Enter may only accept a suggestion, so keep
// pressing Enter until the field clears (the DoneFunc clears it).
func (h *reproHarness) submit(cmd string) {
	h.t.Helper()
	h.waitFor("input focus before typing "+cmd, 3*time.Second, h.inputHasFocus)
	h.typeText(cmd)
	h.waitFor("typed text to land", 3*time.Second, func() bool { return h.inputText() == cmd })
	// First Enter either submits (field clears) or accepts the autocomplete
	// suggestion (field becomes "cmd "). Wait for one of those before deciding.
	h.key(tcell.KeyEnter, 0)
	h.waitFor("Enter processed", 3*time.Second, func() bool { t := h.inputText(); return t == "" || t == cmd+" " })
	if h.inputText() == "" {
		return
	}
	h.key(tcell.KeyEnter, 0)
	h.waitFor("command submitted", 3*time.Second, func() bool { return h.inputText() == "" })
}

// bigStorageKey returns the storage building with the largest "all" bonus.
func bigStorageKey() (string, float64) {
	best, bestV := "", 0.0
	for _, b := range config.BuildingByKey() {
		for _, e := range b.Effects {
			if e.Type == "storage" && e.Target == "all" && e.Value > bestV {
				best, bestV = b.Key, e.Value
			}
		}
	}
	return best, bestV
}

// grantNextAge satisfies every requirement of the next age using the dev
// commands (all engine mutation happens under the engine lock). Returns false
// if already at the final age.
func (h *reproHarness) grantNextAge() bool {
	h.t.Helper()
	st := h.eng.GetState()
	if st.NextAge == "" {
		return false
	}
	for res, v := range st.NextAgeResReqs {
		if rs, ok := st.Resources[res]; ok && rs.Storage < v {
			h.t.Fatalf("storage for %s too low: %.0f < %.0f", res, rs.Storage, v)
		}
	}
	for bld, n := range st.NextAgeBldReqs {
		for c := st.Buildings[bld].Count; c < n; c++ {
			h.dev("/build " + bld)
		}
	}
	if st.CurrentAgeWonderKey != "" {
		h.dev("/build " + st.CurrentAgeWonderKey)
	}
	for res := range st.NextAgeResReqs {
		h.dev("/give " + res + " 1e18")
	}
	return true
}

var dismissKeys = []struct {
	k tcell.Key
	r rune
}{
	{tcell.KeyRune, 'x'},
	{tcell.KeyEnter, 0},
	{tcell.KeyEsc, 0},
	{tcell.KeyRune, ' '},
}

// advanceAndCheck performs one real `advance` and checks the splash lifecycle.
func (h *reproHarness) advanceAndCheck(i int, catChoice rune) {
	h.t.Helper()
	h.advanceAndCheckWith(i, catChoice, func() { h.submit("advance") })
}

// advanceAndCheckWith runs doAdvance, waits for the age splash, then plays the
// resulting modal sequence the way a player would until the dashboard is back:
//   - catastrophe modal in front → press catChoice ('d' defer / 'e' endure);
//     it must close within 2s.
//   - age splash in front → press "any key" (rotating x / Enter / Esc / space);
//     it must close within 2s.
//
// Works whether the catastrophe modal is stacked on the splash (old behaviour)
// or shown after it (fixed behaviour). Finally input focus must be restored.
func (h *reproHarness) advanceAndCheckWith(i int, catChoice rune, doAdvance func()) {
	h.t.Helper()
	before := h.eng.GetState()
	doAdvance()
	h.waitFor("age_splash page", 3*time.Second, func() bool { return h.hasPage("age_splash") })
	after := h.eng.GetState()
	label := fmt.Sprintf("#%d %s→%s (epoch %s→%s, pendingCat=%q)", i, before.Age, after.Age, before.EpochKey, after.EpochKey, after.PendingCatastrophe)
	h.t.Logf("%s front=%s", label, h.frontPage())

	for step := 0; step < 6; step++ {
		// Let refresh() (500ms) surface anything pending (e.g. catastrophe modal).
		time.Sleep(550 * time.Millisecond)
		switch h.frontPage() {
		case "catastrophe":
			h.t.Logf("%s catastrophe modal in front; pressing %q\n%s", label, catChoice, h.describeUI())
			h.key(tcell.KeyRune, catChoice)
			deadline := time.Now().Add(2 * time.Second)
			for h.hasPage("catastrophe") {
				if time.Now().After(deadline) {
					h.t.Fatalf("[%s] FROZEN: catastrophe modal ignored %q\n%s", label, catChoice, h.describeUI())
				}
				time.Sleep(50 * time.Millisecond)
			}
			h.t.Logf("%s after closing catastrophe:\n%s", label, h.describeUI())
		case "age_splash":
			dk := dismissKeys[(i+step)%len(dismissKeys)]
			h.key(dk.k, dk.r)
			deadline := time.Now().Add(2 * time.Second)
			for h.hasPage("age_splash") {
				if time.Now().After(deadline) {
					h.t.Fatalf("[%s] FROZEN: splash not dismissed by key %v/%q within 2s\n%s", label, dk.k, dk.r, h.describeUI())
				}
				time.Sleep(50 * time.Millisecond)
			}
		case "dashboard":
			if !h.inputHasFocus() {
				h.t.Fatalf("[%s] back on dashboard but input field lost focus\n%s", label, h.describeUI())
			}
			return
		default:
			h.t.Fatalf("[%s] unexpected front page\n%s", label, h.describeUI())
		}
	}
	h.t.Fatalf("[%s] modal sequence did not settle\n%s", label, h.describeUI())
}

// TestReproAgeSplashAllAges walks every age via the typed `advance` command.
func TestReproAgeSplashAllAges(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI regression test; skipped with -short")
	}
	h := newReproHarness(t)
	for i := 0; ; i++ {
		if !h.grantNextAge() {
			break
		}
		h.advanceAndCheck(i, 'd')
	}
	if st := h.eng.GetState(); st.NextAge != "" {
		t.Fatalf("did not reach final age: %s", st.Age)
	}
}

// TestReproAgeSplashWithCatastrophe forces a pending catastrophe to surface in
// the same refresh() as the age splash (exactly what happens when an epoch
// transition rolls a catastrophe inside advanceAge), for each catastrophe
// choice a player can make.
func TestReproAgeSplashWithCatastrophe(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI regression test; skipped with -short")
	}
	for _, choice := range []rune{'d', 'e'} {
		t.Run(string(choice), func(t *testing.T) {
			h := newReproHarness(t)
			for i := 0; i < 4; i++ {
				if !h.grantNextAge() {
					break
				}
				// Invoke (if still allowed this epoch) and advance atomically
				// w.r.t. refresh() (both on the UI goroutine, like a typed
				// command) so both surface in the same refresh — the same
				// state advanceAge's epoch roll produces.
				h.advanceAndCheckWith(i, choice, func() {
					h.onUI(func() {
						_ = h.eng.InvokeCatastrophe()
						if err := h.eng.AdvanceAge(); err != nil {
							h.t.Errorf("AdvanceAge: %v", err)
						}
					})
				})
			}
		})
	}
}

// screenText renders the simulation screen to plain text (post-draw).
func (h *reproHarness) screenText() string {
	var sb strings.Builder
	h.onUI(func() { h.a.tviewApp.ForceDraw() })
	cells, w, ht := h.sim.GetContents()
	for y := 0; y < ht; y++ {
		line := make([]byte, 0, w)
		for x := 0; x < w; x++ {
			b := cells[y*w+x].Bytes
			if len(b) == 0 {
				b = []byte{' '}
			}
			line = append(line, b...)
		}
		sb.WriteString(strings.TrimRight(string(line), " "))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// TestReproAgeSplashCatastropheAutoDismiss: an advance that also leaves a
// catastrophe pending (what an epoch-transition roll does) and a player who
// just reads the screen for >20s, so the splash's auto-dismiss timer fires.
// Afterwards whatever modal is on screen must still own the keyboard, and the
// player must be able to get back to the dashboard.
func TestReproAgeSplashCatastropheAutoDismiss(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI regression test; skipped with -short")
	}
	h := newReproHarness(t)
	h.grantNextAge()
	h.onUI(func() {
		_ = h.eng.InvokeCatastrophe()
		if err := h.eng.AdvanceAge(); err != nil {
			t.Errorf("AdvanceAge: %v", err)
		}
	})
	h.waitFor("age_splash", 3*time.Second, func() bool { return h.hasPage("age_splash") })
	time.Sleep(700 * time.Millisecond)
	t.Logf("shown:\n%s", h.describeUI())
	time.Sleep(21 * time.Second) // splash auto-dismiss timer fires
	t.Logf("after 20s auto-dismiss:\n%s", h.describeUI())
	if h.hasPage("age_splash") {
		t.Fatalf("splash still present after auto-dismiss\n%s", h.describeUI())
	}
	if !h.hasPage("catastrophe") {
		t.Fatalf("catastrophe modal never shown\n%s", h.describeUI())
	}
	if h.inputHasFocus() {
		t.Logf("SCREEN (modal visible, keyboard on hidden input):\n%s", h.screenText())
	}
	// Player presses 'd' (Defer) on the visible modal.
	h.key(tcell.KeyRune, 'd')
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !h.hasPage("catastrophe") && h.inputHasFocus() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("FROZEN: catastrophe modal left on screen but unreachable by keyboard after splash auto-dismiss\n%s", h.describeUI())
}

var _ = tview.NewBox // keep tview import for diagnostics helpers
