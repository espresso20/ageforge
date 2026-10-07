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
	"sync/atomic"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

type reproHarness struct {
	t      *testing.T
	eng    *game.GameEngine
	a      *App
	sim    tcell.SimulationScreen
	runErr chan error
	// built is the high-water building count per key. An endured catastrophe
	// destroys a random 20% of buildings (Buildings.DestroyRandom; seeded, but
	// each new game rolls a fresh seed), and when the dice took a storage building the next age's
	// requirement no longer fit under the cap: the storage check below failed
	// roughly one run in four under -race. grantNextAge rebuilds back to this
	// mark, so the test exercises the modal lifecycle it is about rather than
	// the catastrophe's dice.
	built map[string]int
	// catsSeen counts catastrophe modals the sequence helper resolved.
	catsSeen int
	// loopExited is set by teardown once app.Run has returned: from then on
	// nothing of this App reads game.DevModeActive, and it can be put back.
	loopExited *atomic.Bool
}

func newReproHarness(t *testing.T) *reproHarness {
	t.Helper()
	// Dev mode is on for the harness and put back afterwards, so it does not
	// leak into the tests that run after this one (it used to stay on for
	// the rest of the package). The write back must wait for the App's event
	// loop: a stopped App can still be draining a queued refresh, which
	// reads DevModeActive, after app.Stop() returns. This cleanup is
	// registered first, so it runs last, after teardown, and it only writes
	// once teardown has seen app.Run return.
	prevDev := game.DevModeActive
	loopExited := &atomic.Bool{}
	if !prevDev {
		game.DevModeActive = true
		t.Cleanup(func() {
			if loopExited.Load() {
				game.DevModeActive = prevDev
			}
		})
	}

	// Isolate the data root (account + saves) so the harness never touches a
	// real ./data tree, and wire an established account up front so NewApp
	// skips the first-run account-name prompt.
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	acct, err := game.CreateNamedAccount("Repro Harness")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}

	eng := game.NewGameEngine()
	eng.SetAccount(acct)
	a := NewApp(eng, "dev") // "dev" skips the network update check
	sim := tcell.NewSimulationScreen("UTF-8")
	a.SetScreen(sim) // theme-wrapped; calls sim.Init()
	sim.SetSize(200, 60)

	h := &reproHarness{t: t, eng: eng, a: a, sim: sim, runErr: make(chan error, 1), loopExited: loopExited}
	go func() { h.runErr <- a.Run() }()
	// Registered after SetDataDirForTest, so (cleanups run last in, first
	// out) it runs before the data dir is restored.
	t.Cleanup(h.teardown)

	// Real "New Game" path: the splash List has shortcut 'n', which opens the
	// civilization-name prompt pre-filled with a generated name; Enter accepts.
	h.waitFor("splash menu in front", 5*time.Second, func() bool { return h.frontPage() == "splash" })
	h.key(tcell.KeyRune, 'n')
	h.waitFor("new-game name prompt", 5*time.Second, func() bool { return h.frontPage() == newGameNamePage })
	h.key(tcell.KeyEnter, 0)
	h.waitFor("dashboard visible with input focus", 5*time.Second, func() bool {
		return h.frontPage() == "dashboard" && h.inputHasFocus()
	})
	// Raise storage once, high enough for every age's resource requirements,
	// then wait for recalculateRates() (runs each tick) to apply it.
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

// teardown stops the engine and the App and waits for both to finish, so
// nothing they started can touch package state after the test's earlier-
// registered cleanups run. CI once caught SetDataDirForTest's restore racing
// a goroutine that was still resolving a path through the old data dir.
func (h *reproHarness) teardown() {
	h.eng.Stop() // returns once the tick loop (autosave, account flush) has exited
	stopped := make(chan struct{})
	go func() { h.a.Stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		h.t.Errorf("app.Stop() did not return (event loop wedged)")
		return
	}
	// app.Stop returns before the event loop does, and a refresh still in the
	// queue calls GetState, which stats the save file under the data dir. Only
	// Run returning means the loop is done.
	select {
	case <-h.runErr:
		h.loopExited.Store(true)
	case <-time.After(5 * time.Second):
		h.t.Errorf("app.Run() did not return after Stop; its event loop could still touch the data dir")
	}
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

// submit types a command into the real input field and presses Enter once.
// Enter runs the typed text even while the autocomplete dropdown is open, and
// the submit clears the field.
func (h *reproHarness) submit(cmd string) {
	h.t.Helper()
	h.waitFor("input focus before typing "+cmd, 3*time.Second, h.inputHasFocus)
	h.typeText(cmd)
	h.waitFor("typed text to land", 3*time.Second, func() bool { return h.inputText() == cmd })
	h.key(tcell.KeyEnter, 0)
	h.waitFor("command submitted on the first Enter", 3*time.Second, func() bool { return h.inputText() == "" })
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
	if h.built == nil {
		h.built = map[string]int{}
	}
	rebuilt := false
	for key, b := range st.Buildings {
		for c := b.Count; c < h.built[key]; c++ {
			h.dev("/build " + key)
			rebuilt = true
		}
	}
	if rebuilt {
		// Storage caps are recomputed on the next tick, not on /build.
		h.waitFor("storage to recover after rebuilding", 10*time.Second, func() bool {
			s := h.eng.GetState()
			for res, v := range s.NextAgeResReqs {
				if rs, ok := s.Resources[res]; ok && rs.Storage < v {
					return false
				}
			}
			return true
		})
		st = h.eng.GetState()
	}
	defer h.recordBuilt()
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

// recordBuilt raises the building high-water mark to the current counts.
func (h *reproHarness) recordBuilt() {
	for key, b := range h.eng.GetState().Buildings {
		if b.Count > h.built[key] {
			h.built[key] = b.Count
		}
	}
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
	h.quietEra()
	h.advanceAndCheckWith(i, catChoice, func() { h.submit("advance") })
}

// quietEra makes the current era's hidden fate quiet (no doom, no false
// prophet), so a fated doom cannot hold an advance up at random: the walk
// tests the splash sequence, not the fate (game/fate_test.go does).
func (h *reproHarness) quietEra() {
	h.t.Helper()
	_ = h.eng.ForceQuietFateForTest(h.eng.GetState().EpochKey) // none in the final epoch
}

// advanceAndCheckWith runs doAdvance, waits for the age splash, then plays the
// resulting sequence the way a player would until the dashboard is back:
//
//  1. The splash says "Press any key to continue", so the first key pressed
//     (rotating x / Enter / Esc / space) must dismiss it within 2s, whatever
//     else the same advance left pending. A catastrophe modal stacked on top
//     of the splash swallows that key: the reported freeze.
//  2. If a catastrophe modal appears afterwards, press catChoice ('e' endure;
//     there is no defer any more, and a pending catastrophe would block the
//     next advance); it must close within 2s.
//  3. Back on the dashboard, the command input must own focus.
func (h *reproHarness) advanceAndCheckWith(i int, catChoice rune, doAdvance func()) {
	h.t.Helper()
	before := h.eng.GetState()
	doAdvance()
	h.waitFor("age_splash page", 3*time.Second, func() bool { return h.hasPage("age_splash") })
	after := h.eng.GetState()
	label := fmt.Sprintf("#%d %s→%s (epoch %s→%s, pendingCat=%q)", i, before.Age, after.Age, before.EpochKey, after.EpochKey, after.PendingCatastrophe)

	// Let refresh() (500ms) surface anything else pending (e.g. catastrophe modal).
	time.Sleep(550 * time.Millisecond)
	h.t.Logf("%s front=%s", label, h.frontPage())

	dk := dismissKeys[i%len(dismissKeys)]
	h.key(dk.k, dk.r)
	deadline := time.Now().Add(2 * time.Second)
	for h.hasPage("age_splash") {
		if time.Now().After(deadline) {
			h.t.Fatalf("[%s] FROZEN: splash not dismissed by key %v/%q within 2s\n%s", label, dk.k, dk.r, h.describeUI())
		}
		time.Sleep(50 * time.Millisecond)
	}

	for step := 0; step < 4; step++ {
		// Let refresh() surface a catastrophe held back while the splash was up.
		time.Sleep(550 * time.Millisecond)
		switch h.frontPage() {
		case "catastrophe":
			h.catsSeen++
			h.t.Logf("%s catastrophe modal in front; pressing %q", label, catChoice)
			h.key(tcell.KeyRune, catChoice)
			deadline := time.Now().Add(2 * time.Second)
			for h.hasPage("catastrophe") {
				if time.Now().After(deadline) {
					h.t.Fatalf("[%s] FROZEN: catastrophe modal ignored %q\n%s", label, catChoice, h.describeUI())
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
		// 'e' would Endure a catastrophe; the quiet walk brings none.
		h.advanceAndCheck(i, 'e')
	}
	st := h.eng.GetState()
	if st.NextAge != "" {
		t.Fatalf("did not reach final age: %s", st.Age)
	}
	// The walk kept every era quiet (quietEra), so the only harbinger is the
	// Cosmic Era's, which warns of the Last Passage (prestige): it arrived on
	// entering the era and handed off at each later age, all without taking
	// the front page or focus from the splash sequence checked above (it is
	// non-blocking), and is still running at the final age, voiced by the
	// fourth figure.
	epochs := config.Epochs()
	cosmic := epochs[len(epochs)-1]
	if h := st.Harbinger; h == nil || !h.LastPassage || h.Age != cosmic.Ages[len(cosmic.Ages)-1] ||
		len(h.Earlier) != len(cosmic.Ages)-1 {
		t.Errorf("final epoch harbinger = %+v, want the Last Passage thread in its last age", st.Harbinger)
	}
	if n := len(st.HarbingerHistory); n != 0 {
		t.Errorf("a quiet walk resolved %d harbinger threads: %+v", n, st.HarbingerHistory)
	}
}

// TestReproAgeSplashWithCatastrophe forces a pending catastrophe to surface in
// the same refresh() as the age splash (a doom that struck just as the player
// advanced, or the dev console's /catastrophe) and resolves it with Endure. Catastrophes only exist from the Iron Era on, and at most once per
// epoch per run, so the walk goes far enough to cross two epoch boundaries.
func TestReproAgeSplashWithCatastrophe(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI regression test; skipped with -short")
	}
	h := newReproHarness(t)
	for i := 0; i < 6; i++ { // primitive → … → renaissance (Iron and Steel eras)
		if !h.grantNextAge() {
			break
		}
		// Advance, then force the new epoch's catastrophe with the dev console
		// (refused before the Iron Era, and while the transition's own roll is
		// pending), atomically w.r.t. refresh() (both on the UI goroutine, like
		// a typed command) so the splash and the modal surface in the same
		// refresh.
		h.quietEra()
		h.advanceAndCheckWith(i, 'e', func() {
			h.onUI(func() {
				before := h.eng.GetState().EpochKey
				if err := h.eng.AdvanceAge(); err != nil {
					h.t.Errorf("AdvanceAge: %v", err)
				}
				if h.eng.GetState().EpochKey != before {
					game.DevConsoleCommand("/catastrophe", h.eng)
				}
			})
		})
	}
	if h.catsSeen < 2 {
		t.Fatalf("expected a catastrophe modal in the Iron and Steel eras, saw %d", h.catsSeen)
	}
	if st := h.eng.GetState(); st.PendingCatastrophe != "" {
		t.Fatalf("catastrophe still pending after Endure: %q", st.PendingCatastrophe)
	}
}

// reachBronze advances until the next advance crosses into the Iron Era,
// where catastrophes become possible.
func (h *reproHarness) reachBronze() {
	h.t.Helper()
	for i := 0; h.eng.GetState().Age != "bronze_age"; i++ {
		if !h.grantNextAge() || i > 5 {
			h.t.Fatalf("could not reach the Bronze Age (at %s)", h.eng.GetState().Age)
		}
		h.advanceAndCheck(i, 'e')
	}
}

// advanceIntoIronWithCatastrophe crosses into the Iron Era and forces its
// catastrophe (dev console) in the same UI update.
func (h *reproHarness) advanceIntoIronWithCatastrophe() {
	h.t.Helper()
	h.grantNextAge()
	h.quietEra()
	h.onUI(func() {
		if err := h.eng.AdvanceAge(); err != nil {
			h.t.Errorf("AdvanceAge: %v", err)
		}
		game.DevConsoleCommand("/catastrophe", h.eng) // refused only if the transition already rolled one
	})
	if h.eng.GetState().PendingCatastrophe == "" {
		h.t.Fatal("no catastrophe pending after entering the Iron Era")
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
// player must be able to get back to the dashboard (Esc) and then return to
// the choice with the `catastrophe` command.
func TestReproAgeSplashCatastropheAutoDismiss(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI regression test; skipped with -short")
	}
	h := newReproHarness(t)
	h.reachBronze()
	h.advanceIntoIronWithCatastrophe()
	h.waitFor("age_splash", 3*time.Second, func() bool { return h.hasPage("age_splash") })
	time.Sleep(700 * time.Millisecond)
	t.Logf("shown:\n%s", h.describeUI())
	// No key pressed: wait out the splash's 20s auto-dismiss timer, then for
	// the catastrophe modal (shown now, or held back until the splash left).
	h.waitFor("splash auto-dismiss", 25*time.Second, func() bool { return !h.hasPage("age_splash") })
	h.waitFor("catastrophe modal", 3*time.Second, func() bool { return h.hasPage("catastrophe") })
	time.Sleep(100 * time.Millisecond) // let any focus change queued with it land
	t.Logf("after 20s auto-dismiss:\n%s", h.describeUI())
	if h.inputHasFocus() {
		t.Logf("SCREEN (modal visible, keyboard on hidden input):\n%s", h.screenText())
	}
	// Player presses Esc on the visible modal: it must close and hand the
	// keyboard back, leaving the catastrophe pending.
	h.key(tcell.KeyEsc, 0)
	deadline := time.Now().Add(2 * time.Second)
	for !(!h.hasPage("catastrophe") && h.inputHasFocus()) {
		if time.Now().After(deadline) {
			t.Fatalf("FROZEN: catastrophe modal left on screen but unreachable by keyboard after splash auto-dismiss\n%s", h.describeUI())
		}
		time.Sleep(50 * time.Millisecond)
	}
	if h.eng.GetState().PendingCatastrophe == "" {
		t.Fatal("Esc must not resolve the catastrophe")
	}
	h.submit("catastrophe")
	h.waitFor("catastrophe modal reopened", 3*time.Second, func() bool { return h.frontPage() == "catastrophe" })
	h.key(tcell.KeyRune, 'e')
	h.waitFor("modal closed after Endure", 3*time.Second, func() bool { return !h.hasPage("catastrophe") && h.inputHasFocus() })
	if st := h.eng.GetState(); st.PendingCatastrophe != "" {
		t.Fatalf("still pending after Endure: %q", st.PendingCatastrophe)
	}
}

// TestReproCatastropheEscBadgeBlockReopen: Esc closes the modal without
// deciding; the status bar shows the pending badge; `advance` is refused while
// pending; the bare `catastrophe` command reopens the modal; refresh() does
// not re-pop it on its own in between; Endure clears everything.
func TestReproCatastropheEscBadgeBlockReopen(t *testing.T) {
	if testing.Short() {
		t.Skip("slow end-to-end UI regression test; skipped with -short")
	}
	h := newReproHarness(t)
	h.reachBronze()
	h.advanceIntoIronWithCatastrophe()
	h.waitFor("age_splash", 3*time.Second, func() bool { return h.hasPage("age_splash") })
	time.Sleep(550 * time.Millisecond)
	h.key(tcell.KeyEnter, 0)
	h.waitFor("splash dismissed", 2*time.Second, func() bool { return !h.hasPage("age_splash") })
	h.waitFor("catastrophe modal", 3*time.Second, func() bool { return h.frontPage() == "catastrophe" })

	h.key(tcell.KeyEsc, 0)
	h.waitFor("Esc closes the modal", 2*time.Second, func() bool { return !h.hasPage("catastrophe") && h.inputHasFocus() })
	// Two refresh cycles: the modal must stay closed until asked for.
	time.Sleep(1100 * time.Millisecond)
	if h.hasPage("catastrophe") {
		t.Fatalf("refresh() re-popped the modal after Esc\n%s", h.describeUI())
	}
	if scr := h.screenText(); !strings.Contains(scr, "Catastrophe pending") {
		t.Errorf("pending badge missing from the status bar:\n%s", scr)
	}

	age := h.eng.GetState().Age
	grantOK := h.grantNextAge()
	h.submit("advance")
	time.Sleep(200 * time.Millisecond)
	if st := h.eng.GetState(); st.Age != age {
		t.Fatalf("advance went through while a catastrophe was pending: %s → %s (requirements granted: %v)", age, st.Age, grantOK)
	}

	h.submit("catastrophe")
	h.waitFor("catastrophe reopens the modal", 3*time.Second, func() bool { return h.frontPage() == "catastrophe" })
	h.key(tcell.KeyRune, 'e')
	h.waitFor("Endure closes the modal", 3*time.Second, func() bool { return !h.hasPage("catastrophe") && h.inputHasFocus() })
	if st := h.eng.GetState(); st.PendingCatastrophe != "" {
		t.Fatalf("still pending after Endure: %q", st.PendingCatastrophe)
	}
	time.Sleep(600 * time.Millisecond) // one refresh so the status bar updates
	if scr := h.screenText(); strings.Contains(scr, "Catastrophe pending") {
		t.Errorf("pending badge still shown after Endure:\n%s", scr)
	}
}

// TestReproHarnessPutsDevModeBack: the harness turns dev mode on and used to
// leave it on for every test that ran after it in the package. It is put
// back once the harness's App has stopped.
func TestReproHarnessPutsDevModeBack(t *testing.T) {
	prev := game.DevModeActive
	game.DevModeActive = false
	t.Cleanup(func() { game.DevModeActive = prev })

	t.Run("harness", func(t *testing.T) {
		newReproHarness(t)
		if !game.DevModeActive {
			t.Error("the harness runs with dev mode on")
		}
	})
	if game.DevModeActive {
		t.Error("dev mode is still on after the harness's test ended")
	}
}
