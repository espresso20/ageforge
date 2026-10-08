package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// menu_behaviour_test.go drives the real main menu: a named account and
// its saves in a temp data root (never the player's), the page's own keys,
// and a clock the test holds.

// menuRig is a main menu on a stage.
type menuRig struct {
	t     *testing.T
	eng   *game.GameEngine
	acct  *game.Account
	app   *tview.Application
	pages *tview.Pages
	m     *mainMenu
	// held is the choices waiting out their strike, clock the menu's time,
	// checks how often the update check was opened.
	held   []func()
	clock  time.Time
	checks int
}

// newMenuRig makes an account, a game for each name (the last one played
// last) and the menu.
func newMenuRig(t *testing.T, games ...string) *menuRig {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	acct, err := game.CreateNamedAccount("Menu Player")
	if err != nil {
		t.Fatal(err)
	}
	r := &menuRig{t: t, acct: acct, clock: time.Unix(1_000_000, 0)}
	for _, name := range games {
		r.newGame(name)
	}
	r.eng = game.NewGameEngine()
	r.eng.SetAccount(acct)
	r.app, r.pages = tview.NewApplication(), tview.NewPages()
	r.m = CreateSplashPage(r.app, r.pages, r.eng, "dev").(*mainMenu)
	r.m.later = func(fn func()) { r.held = append(r.held, fn) }
	r.m.checkUpdates = func() { r.checks++ }
	r.m.now = func() time.Time { return r.clock }
	r.m.start = r.clock
	r.pages.AddPage("splash", r.m, true, true)
	r.pages.AddPage("dashboard", tview.NewBox(), true, false)
	r.app.SetRoot(r.pages, true)
	t.Cleanup(func() {
		r.m.leave()
		r.eng.Stop()
	})
	return r
}

// newGame starts and saves a game on the account, from an engine of its own.
func (r *menuRig) newGame(name string) {
	r.t.Helper()
	mk := game.NewGameEngine()
	mk.SetAccount(r.acct)
	if err := mk.StartNewNamedGame(name); err != nil {
		r.t.Fatal(err)
	}
}

// front is the name of the page in front.
func (r *menuRig) front() string {
	name, _ := r.pages.GetFrontPage()
	return name
}

// key sends a key to whatever has the keyboard.
func (r *menuRig) key(k tcell.Key, ch rune) {
	r.t.Helper()
	// From the root down, as the app sends it, so a page's own key capture
	// has its say.
	r.pages.InputHandler()(tcell.NewEventKey(k, ch, tcell.ModNone), func(p tview.Primitive) { r.app.SetFocus(p) })
}

// release lets the choices that were waiting out their strike act.
func (r *menuRig) release() {
	held := r.held
	r.held = nil
	for _, fn := range held {
		fn()
	}
}

// press presses a letter on the menu and lets the choice act.
func (r *menuRig) press(ch rune) {
	r.t.Helper()
	r.key(tcell.KeyRune, ch)
	r.release()
}

// screen is a drawn menu: its cells.
type menuScreen struct {
	w, h  int
	rows  []string
	cells []string // glyph and style, to tell two frames apart
}

func (s menuScreen) has(text string) bool {
	for _, row := range s.rows {
		if strings.Contains(row, text) {
			return true
		}
	}
	return false
}

// draw draws the menu at w by h, as the app would.
func (r *menuRig) draw(w, h int) menuScreen {
	r.t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		r.t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	r.m.SetRect(0, 0, w, h)
	r.m.Draw(theme.WrapScreen(sim))
	out := menuScreen{w: w, h: h}
	for y := 0; y < h; y++ {
		var sb strings.Builder
		for x := 0; x < w; x++ {
			ch, _, st, _ := sim.GetContent(x, y)
			sb.WriteRune(ch)
			fg, bg, attr := st.Decompose()
			out.cells = append(out.cells, fmt.Sprintf("%c %06x %06x %d", ch, fg.Hex(), bg.Hex(), attr&tcell.AttrBold))
		}
		out.rows = append(out.rows, sb.String())
	}
	return out
}

func ids(items []menuItem) string {
	var out []string
	for _, it := range items {
		out = append(out, it.id)
	}
	return strings.Join(out, " ")
}

// TestMenuFirstVisitThenAGame: an account with no game sees the forge page,
// New game first and selected, no Continue. With a game it sees its town,
// Continue first and selected, with the town, the age and the people under
// it and the save's name beside it.
func TestMenuFirstVisitThenAGame(t *testing.T) {
	r := newMenuRig(t)
	m := r.m
	if !m.view.forge || m.hasCur {
		t.Fatalf("an account with no game: forge page %v, a current game %v", m.view.forge, m.hasCur)
	}
	if got := ids(m.view.items); got != "new load badges themes accounts updates quit" {
		t.Errorf("the first visit's entries are %q", got)
	}
	if m.view.sel != 0 {
		t.Errorf("the first visit selects entry %d, want New game", m.view.sel)
	}
	scr := r.draw(120, 40)
	for _, want := range []string{"CONTENTS", "▸ New game", "Plate I. The forge", "0 of ", "dev"} {
		if !scr.has(want) {
			t.Errorf("the first visit's page has no %q", want)
		}
	}
	if scr.has("Continue") {
		t.Error("the first visit's page offers Continue")
	}

	// A game is started and saved (as play would); the menu is returned to.
	r.newGame("ashford")
	m.Focus(func(tview.Primitive) {})
	if m.view.forge || !m.hasCur || m.town == nil {
		t.Fatalf("an account with a game: forge page %v, a current game %v, a town %v", m.view.forge, m.hasCur, m.town != nil)
	}
	if got := ids(m.view.items); got != "continue new load badges themes accounts updates quit" {
		t.Errorf("the entries with a game are %q", got)
	}
	if m.view.sel != 0 || m.view.items[0].id != miContinue {
		t.Errorf("with a game the menu selects entry %d (%s), want Continue", m.view.sel, m.view.items[m.view.sel].id)
	}
	for _, size := range [][2]int{{120, 40}, {80, 24}} {
		scr = r.draw(size[0], size[1])
		for _, want := range []string{"▸ Continue", "ashford", "Primitive Age", "0 people", "New game", "ages, one terminal", m.town.name + " · Primitive Age"} {
			if !scr.has(want) {
				t.Errorf("at %dx%d the page with a game has no %q", size[0], size[1], want)
			}
		}
		if scr.has("CONTENTS") {
			t.Errorf("at %dx%d the page with a game is the first-visit page", size[0], size[1])
		}
	}
	if m.town.name == "" {
		t.Error("the town has no name")
	}

	// A selection made by hand stays put when the menu reads again.
	r.key(tcell.KeyDown, 0)
	r.key(tcell.KeyDown, 0)
	m.Focus(func(tview.Primitive) {})
	if got := m.view.items[m.view.sel].id; got != miLoad {
		t.Errorf("after two downs and a return the selection is on %q, want load", got)
	}
	r.key(tcell.KeyUp, 0)
	r.key(tcell.KeyUp, 0)
	r.key(tcell.KeyUp, 0)
	if got := m.view.items[m.view.sel].id; got != miQuit {
		t.Errorf("up from the first entry lands on %q, want quit", got)
	}
	r.key(tcell.KeyHome, 0)
	if m.view.sel != 0 {
		t.Errorf("Home selects entry %d", m.view.sel)
	}
}

// TestMenuKeysOpenEveryEntry: every entry opens by its letter, the pages
// come back with Esc, and the letter that used to delete every save does
// nothing on the front page.
func TestMenuKeysOpenEveryEntry(t *testing.T) {
	r := newMenuRig(t, "ashford")
	m := r.m
	saves := func() int {
		list, _ := game.ListSaveDetails()
		return len(list)
	}

	for _, c := range []struct {
		key  rune
		page string
	}{{'l', loadGamePage}, {'t', themePickerPage}, {'a', accountsPage}, {'b', menuBadgesPage}, {'L', loadGamePage}, {'B', menuBadgesPage}} {
		r.press(c.key)
		if got := r.front(); got != c.page {
			t.Fatalf("%q opens %q, want %q", c.key, got, c.page)
		}
		if m.active || m.stop != nil {
			t.Errorf("%q: the menu is still running under the page it opened", c.key)
		}
		r.key(tcell.KeyEsc, 0)
		if got := r.front(); got != "splash" {
			t.Fatalf("Esc on %q leaves %q in front", c.page, got)
		}
		if !m.active {
			t.Errorf("back from %q the menu is not in use", c.page)
		}
	}

	// Check for updates stays on the front page: a dialog over the menu.
	r.press('u')
	if r.checks != 1 || r.front() != "splash" || !m.active {
		t.Errorf("u: the update check opened %d times, front %q, menu in use %v", r.checks, r.front(), m.active)
	}

	// x deleted every save from the old menu. It is not on this page.
	r.press('x')
	if r.front() != "splash" || saves() != 1 || len(r.held) != 0 {
		t.Errorf("x did something on the front page: front %q, %d saves", r.front(), saves())
	}

	// New game asks for a name first; Esc leaves the menu as it was.
	r.press('n')
	if r.front() != newGameNamePage {
		t.Fatalf("n opens %q, want the name prompt", r.front())
	}
	if !m.active {
		t.Error("the name prompt is a dialog over the menu, which should keep running")
	}
	r.key(tcell.KeyEsc, 0)
	if r.front() != "splash" || saves() != 1 {
		t.Errorf("Esc on the name prompt: front %q, %d saves", r.front(), saves())
	}

	// Continue opens the current game directly.
	r.press('c')
	if r.front() != "dashboard" || r.eng.ActiveSaveName() != "ashford" {
		t.Fatalf("c: front %q, the game in play %q", r.front(), r.eng.ActiveSaveName())
	}
	waitRunning(t, r.eng)
	if m.active {
		t.Error("the menu is still in use behind the game")
	}
	// Back at the menu (Esc in the game saves, stops and returns).
	r.eng.Stop()
	r.pages.SwitchToPage("splash")
	if !m.active || m.view.items[m.view.sel].id != miContinue {
		t.Errorf("back from the game: menu in use %v, selection %q", m.active, m.view.items[m.view.sel].id)
	}
	// Enter on the selection is Continue too.
	r.key(tcell.KeyEnter, 0)
	r.release()
	if r.front() != "dashboard" {
		t.Fatalf("Enter on Continue: front %q", r.front())
	}
	waitRunning(t, r.eng)
	r.eng.Stop()
	r.pages.SwitchToPage("splash")

	r.press('q')
	if m.active {
		t.Error("q did not leave the menu")
	}
}

// TestMenuStrikeAndMotion: choosing an entry strikes the title (a flare and
// a burst of sparks) and the entry opens a moment later; with the motion
// setting off nothing flares, nothing is thrown and the entry opens at once.
func TestMenuStrikeAndMotion(t *testing.T) {
	r := newMenuRig(t, "ashford")
	m := r.m
	r.draw(120, 40)
	sparks := len(m.scene.sparks.a)

	r.key(tcell.KeyRune, 'l')
	if m.scene.strike != 1 || len(m.scene.sparks.a) <= sparks+40 {
		t.Errorf("a choice with motion on: strike %.2f, %d sparks (from %d)", m.scene.strike, len(m.scene.sparks.a), sparks)
	}
	if !m.pending || len(r.held) != 1 || r.front() != "splash" {
		t.Fatalf("the choice should wait out its strike: pending %v, %d waiting, front %q", m.pending, len(r.held), r.front())
	}
	// A key while it waits is not a second choice.
	r.key(tcell.KeyRune, 't')
	r.key(tcell.KeyDown, 0)
	if len(r.held) != 1 || m.view.items[m.view.sel].id != miLoad {
		t.Errorf("a key during the strike was taken: %d waiting, selection %q", len(r.held), m.view.items[m.view.sel].id)
	}
	// The flare is on the page: the wordmark's hottest cell is white heat.
	hot := rampAt(&m.pal.heat, 1)
	flared := false
	g, _ := m.frameGrid(120, 40)
	for _, c := range g.c {
		flared = flared || c.r == '█' && c.fg == hot
	}
	if !flared {
		t.Error("the strike does not show on the wordmark")
	}
	r.release()
	if r.front() != loadGamePage {
		t.Fatalf("after the strike the entry did not open: front %q", r.front())
	}
	r.key(tcell.KeyEsc, 0)
	if menuStrikeHold >= mapAnimStep {
		t.Errorf("a choice waits %v, a whole animation frame (%v) or more", menuStrikeHold, mapAnimStep)
	}

	// Motion off: the setting is the account's.
	if err := r.acct.SetMotion(false); err != nil {
		t.Fatal(err)
	}
	m.Focus(func(tview.Primitive) {})
	r.draw(120, 40)
	before := *m.scene
	nSparks := len(m.scene.sparks.a)
	r.key(tcell.KeyRune, 'l')
	if r.front() != loadGamePage || len(r.held) != 0 {
		t.Fatalf("with motion off the entry should open at once: front %q, %d waiting", r.front(), len(r.held))
	}
	if m.scene.strike != before.strike || len(m.scene.sparks.a) != nSparks || m.scene.frame != before.frame {
		t.Errorf("with motion off a choice moved the scene: strike %.2f, sparks %d to %d, frame %d to %d",
			m.scene.strike, nSparks, len(m.scene.sparks.a), before.frame, m.scene.frame)
	}
}

// TestMenuClockRunsOnlyWhileInUse: the page moves at the map's rate while
// the menu is the page in use, stops when it is left and goes on from where
// it stopped when it is returned to. With the motion setting off there is
// no clock at all and every draw is the same still frame.
func TestMenuClockRunsOnlyWhileInUse(t *testing.T) {
	r := newMenuRig(t, "ashford")
	m := r.m
	a := r.draw(120, 40)
	if m.stop == nil {
		t.Fatal("the menu is in use with motion on and no clock runs")
	}
	if got := r.draw(120, 40); !reflect.DeepEqual(a.cells, got.cells) {
		n := 0
		for i := range a.cells {
			if a.cells[i] != got.cells[i] && n < 12 {
				t.Logf("cell (%d,%d): %q then %q", i%120, i/120, a.cells[i], got.cells[i])
				n++
			}
		}
		t.Error("two draws at the same moment differ")
	}
	r.clock = r.clock.Add(3 * mapAnimStep)
	b := r.draw(120, 40)
	if m.scene.frame != 3 {
		t.Errorf("three frames of time moved the scene %d frames", m.scene.frame)
	}
	if reflect.DeepEqual(a.cells, b.cells) {
		t.Error("the page did not move in three frames")
	}
	// A long stall is skipped, not replayed.
	r.clock = r.clock.Add(1000 * mapAnimStep)
	r.draw(120, 40)
	if m.scene.frame != 1003 {
		t.Errorf("after a stall the scene is at frame %d, want 1003", m.scene.frame)
	}

	// Left: no clock, and the menu draws nothing under the page over it.
	m.leave()
	if m.stop != nil {
		t.Error("the clock still runs after the menu was left")
	}
	r.clock = r.clock.Add(40 * mapAnimStep)
	blank := r.draw(120, 40)
	if m.scene.frame != 1003 {
		t.Errorf("the scene moved while the menu was left: frame %d", m.scene.frame)
	}
	for _, row := range blank.rows {
		if strings.TrimSpace(row) != "" {
			t.Fatalf("the menu drew while it was left: %q", row)
		}
	}
	// Returned to: it goes on from where it stopped.
	m.Focus(func(tview.Primitive) {})
	r.draw(120, 40)
	if m.scene.frame != 1003 || m.stop == nil {
		t.Errorf("on return the scene is at frame %d (want 1003) and the clock runs: %v", m.scene.frame, m.stop != nil)
	}
	r.clock = r.clock.Add(2 * mapAnimStep)
	r.draw(120, 40)
	if m.scene.frame != 1005 {
		t.Errorf("two frames after the return the scene is at frame %d, want 1005", m.scene.frame)
	}

	// Motion off: one still frame, whatever the time, and no clock.
	if err := r.acct.SetMotion(false); err != nil {
		t.Fatal(err)
	}
	m.Focus(func(tview.Primitive) {})
	m.scene = nil // a menu opened with motion off starts from the rest frame
	still := r.draw(120, 40)
	if m.stop != nil {
		t.Error("with motion off a clock runs")
	}
	for _, d := range []time.Duration{time.Second, 7 * time.Second, time.Hour} {
		r.clock = r.clock.Add(d)
		if got := r.draw(120, 40); !reflect.DeepEqual(still.cells, got.cells) {
			t.Fatalf("with motion off the page changed after %v", d)
		}
	}
	if m.scene.frame != 0 {
		t.Errorf("with motion off the scene moved to frame %d", m.scene.frame)
	}
	// The still frame is not an empty one: the sparks are in the air.
	if len(m.scene.sparks.a) == 0 {
		t.Error("the still frame has no sparks")
	}
	// And it is the same frame every time the menu is made.
	other := newMenuScene(m.scene.L)
	if !reflect.DeepEqual(other.sparks.a, m.scene.sparks.a) || other.t != m.scene.t {
		t.Error("two scenes at rest differ: the still frame is not always the same")
	}
}

// dataTree reads every file under root: its bytes and its modification time.
func dataTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() {
			out[rel+"/"] = "dir"
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out[rel] = hex.EncodeToString(sum[:]) + " " + fi.ModTime().Format(time.RFC3339Nano)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestMenuLooksAndWritesNothing: the menu is made, drawn, moved, struck,
// walked up and down, left for the badge case and returned to, and read
// again from scratch. Not one file changes (not a byte, not a time), and
// the engine in play is as it was: no game loaded into it, no tick, no
// save name taken, nothing earned.
func TestMenuLooksAndWritesNothing(t *testing.T) {
	root := t.TempDir()
	t.Cleanup(game.SetDataDirForTest(root))
	acct, err := game.CreateNamedAccount("Careful Player")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ashford", "brackwater"} {
		mk := game.NewGameEngine()
		mk.SetAccount(acct)
		if err := mk.StartNewNamedGame(name); err != nil {
			t.Fatal(err)
		}
		mk.StepTicks(40)
		if err := mk.SaveGame(name); err != nil {
			t.Fatal(err)
		}
	}
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	// Make the saves old, so that a load would have hours of time away to
	// apply. (Only the files' times: the menu must not care either way.)
	old := time.Now().Add(-6 * time.Hour)
	saves := filepath.Join(game.DataDir(), "saves")
	for _, name := range []string{"ashford", "brackwater"} {
		if err := os.Chtimes(filepath.Join(saves, name+".json"), old, old); err != nil {
			t.Fatal(err)
		}
	}

	eng := game.NewGameEngine()
	eng.SetAccount(acct)
	stBefore := eng.GetState()
	_, sumBefore := eng.Badges()
	before := dataTree(t, root)

	app, pages := tview.NewApplication(), tview.NewPages()
	m := CreateSplashPage(app, pages, eng, "dev").(*mainMenu)
	clock := time.Unix(2_000_000, 0)
	m.now, m.start = func() time.Time { return clock }, clock
	var held []func()
	m.later = func(fn func()) { held = append(held, fn) }
	m.checkUpdates = func() {}
	pages.AddPage("splash", m, true, true)
	app.SetRoot(pages, true)
	t.Cleanup(m.leave)

	if !m.hasCur || m.view.forge || m.cur.Save.Name != "brackwater" {
		t.Fatalf("the menu should show the town of brackwater: current %q, forge page %v", m.cur.Save.Name, m.view.forge)
	}
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	scr := theme.WrapScreen(sim)
	key := func(k tcell.Key, ch rune) {
		pages.InputHandler()(tcell.NewEventKey(k, ch, tcell.ModNone), func(p tview.Primitive) { app.SetFocus(p) })
	}
	for _, size := range [][2]int{{120, 40}, {80, 24}, {144, 46}} {
		sim.SetSize(size[0], size[1])
		m.SetRect(0, 0, size[0], size[1])
		for i := 0; i < 12; i++ {
			clock = clock.Add(mapAnimStep)
			m.Draw(scr)
		}
	}
	key(tcell.KeyDown, 0)
	key(tcell.KeyUp, 0)
	key(tcell.KeyRune, 'b') // the badge case, and back
	for _, fn := range held {
		fn()
	}
	if name, _ := pages.GetFrontPage(); name != menuBadgesPage {
		t.Fatalf("b opened %q", name)
	}
	_, front := pages.GetFrontPage()
	front.SetRect(0, 0, 120, 40)
	front.Draw(scr)
	key(tcell.KeyEsc, 0)
	key(tcell.KeyRune, 'u')
	m.refresh(true)
	m.Draw(scr)
	m.leave()

	if after := dataTree(t, root); !reflect.DeepEqual(before, after) {
		for k, v := range after {
			if before[k] != v {
				t.Errorf("the menu changed %s", k)
			}
		}
		for k := range before {
			if _, ok := after[k]; !ok {
				t.Errorf("the menu removed %s", k)
			}
		}
	}
	stAfter := eng.GetState()
	if stAfter.Tick != stBefore.Tick || stAfter.Age != stBefore.Age || eng.Running() || eng.ActiveSaveName() != game.AutosaveName ||
		!reflect.DeepEqual(stAfter.Resources, stBefore.Resources) || len(stAfter.Buildings) != len(stBefore.Buildings) {
		t.Errorf("the engine in play changed: tick %d to %d, age %q, running %v, save %q",
			stBefore.Tick, stAfter.Tick, stAfter.Age, eng.Running(), eng.ActiveSaveName())
	}
	if _, sumAfter := eng.Badges(); !reflect.DeepEqual(sumBefore, sumAfter) {
		t.Errorf("looking at the menu earned something: %+v to %+v", sumBefore, sumAfter)
	}
}

// panicStyle is a map style that fails when it draws.
type panicStyle struct{ mapstyle.Style }

func (panicStyle) Draw(tcell.Screen, mapstyle.Rect, mapstyle.Frame) {
	panic("the map cannot draw this")
}

// noErrorOnPage fails if a drawn page says anything went wrong.
func noErrorOnPage(t *testing.T, where string, scr menuScreen) {
	t.Helper()
	for _, row := range scr.rows {
		low := strings.ToLower(row)
		for _, word := range []string{"error", "could not", "cannot", "failed", "damaged", "corrupt", "panic"} {
			if strings.Contains(low, word) {
				t.Errorf("%s: the page says %q: %q", where, word, strings.TrimSpace(row))
			}
		}
	}
}

// TestMenuFallsBackWithoutAnError: a save that cannot be read, a map that
// cannot be drawn and a save that vanishes all leave the player on the
// first page with nothing said about it on screen.
func TestMenuFallsBackWithoutAnError(t *testing.T) {
	// The account's only save is damaged: no Continue, the first page.
	r := newMenuRig(t)
	if err := os.MkdirAll(filepath.Join(game.DataDir(), "saves"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(game.DataDir(), "saves", "torn.json"), []byte("{not a save"), 0644); err != nil {
		t.Fatal(err)
	}
	r.m.Focus(func(tview.Primitive) {})
	if r.m.hasCur || !r.m.view.forge || r.m.view.items[0].id != miNew {
		t.Errorf("a damaged save alone: a current game %v, forge page %v, first entry %q", r.m.hasCur, r.m.view.forge, r.m.view.items[0].id)
	}
	noErrorOnPage(t, "a damaged save", r.draw(120, 40))

	// A damaged save beside a good one: the good one is the current game.
	r.newGame("ashford")
	r.m.Focus(func(tview.Primitive) {})
	if !r.m.hasCur || r.m.cur.Save.Name != "ashford" || r.m.view.forge {
		t.Errorf("a damaged save beside a good one: current %q, forge page %v", r.m.cur.Save.Name, r.m.view.forge)
	}

	// The map fails while it draws: the first page takes over in that very
	// frame, with Continue still offered (the save itself is fine).
	r.m.town.style = panicStyle{r.m.town.style}
	scr := r.draw(120, 40)
	if !r.m.town.failed {
		t.Fatal("a map that panics was not marked as failed")
	}
	noErrorOnPage(t, "a map that cannot be drawn", scr)
	if !r.m.view.forge || r.m.view.items[0].id != miContinue {
		t.Errorf("after the map failed: forge page %v, first entry %q", r.m.view.forge, r.m.view.items[0].id)
	}
	if !scr.has("CONTENTS") || !scr.has("▸ Continue") {
		t.Error("after the map failed the page is not the first page with Continue on it")
	}
	// And it stays the first page, without trying the map again.
	scr = r.draw(80, 24)
	noErrorOnPage(t, "the first page after the map failed", scr)
	if !scr.has("CONTENTS") {
		t.Error("the page went back to a map that cannot be drawn")
	}

	// The save vanishes after the menu read it: Continue says so in a
	// dialog (the one thing that is said), and the menu reads again.
	if err := game.DeleteSave("ashford"); err != nil {
		t.Fatal(err)
	}
	r.press('c')
	if r.front() != updateModalPage {
		t.Fatalf("Continue on a save that is gone: front %q, want a dialog", r.front())
	}
	if r.eng.Running() || r.m.hasCur {
		t.Errorf("after a failed Continue: engine running %v, a current game %v", r.eng.Running(), r.m.hasCur)
	}
}

// TestMenuBadgesOpenWithoutAGame: the badge case opens from the menu with
// no game loaded, draws the account's badges, and Esc closes an open badge
// first, then the case.
func TestMenuBadgesOpenWithoutAGame(t *testing.T) {
	r := newMenuRig(t)
	r.press('b')
	if r.front() != menuBadgesPage {
		t.Fatalf("b opens %q", r.front())
	}
	_, front := r.pages.GetFrontPage()
	b := front.(*menuBadges)
	if !b.account || len(b.views) == 0 {
		t.Fatalf("the case has no badges to show: account %v, %d badges", b.account, len(b.views))
	}
	if r.eng.Running() || r.eng.GetState().Tick != 0 {
		t.Error("opening the case started a game")
	}
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(100, 30)
	b.SetRect(0, 0, 100, 30)
	b.Draw(theme.WrapScreen(sim))
	var text strings.Builder
	for y := 0; y < 30; y++ {
		for x := 0; x < 100; x++ {
			ch, _, _, _ := sim.GetContent(x, y)
			text.WriteRune(ch)
		}
		text.WriteByte('\n')
	}
	if !strings.Contains(text.String(), "earned") {
		t.Errorf("the case does not show its count:\n%s", text.String())
	}
	// The case's own keys work, and Esc closes the open badge first.
	r.key(tcell.KeyRight, 0)
	r.key(tcell.KeyEnter, 0)
	if !b.view.card {
		t.Fatal("Enter did not open the selected badge")
	}
	r.key(tcell.KeyEsc, 0)
	if b.view.card || r.front() != menuBadgesPage {
		t.Fatalf("the first Esc should close the badge and leave the case open: card %v, front %q", b.view.card, r.front())
	}
	r.key(tcell.KeyEsc, 0)
	if r.front() != "splash" || !r.m.active {
		t.Errorf("the second Esc: front %q, menu in use %v", r.front(), r.m.active)
	}
}

// TestMenuTownFollowsTheSettings: the town is drawn in the account's map
// style and glyph set, and the plain glyph set puts no block on the page.
func TestMenuTownFollowsTheSettings(t *testing.T) {
	r := newMenuRig(t, "ashford")
	m := r.m
	if got := m.town.style.Name(); got != "roguelike" {
		t.Errorf("by default the town is drawn by %q", got)
	}
	if err := r.acct.SetMapStyle("skyline"); err != nil {
		t.Fatal(err)
	}
	if err := r.acct.SetMapGlyphs("ascii"); err != nil {
		t.Fatal(err)
	}
	m.Focus(func(tview.Primitive) {})
	if got := m.town.style.Name(); got != "skyline" {
		t.Errorf("after map style skyline the town is drawn by %q", got)
	}
	if m.set.Tier != mapmodel.TierASCII || !m.view.plain {
		t.Errorf("after map glyphs ascii: tier %v, plain %v", m.set.Tier, m.view.plain)
	}
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		scr := r.draw(size[0], size[1])
		for y, row := range scr.rows {
			for x, ch := range []rune(row) {
				if ch > 0x7e {
					t.Fatalf("at %dx%d with plain glyphs cell (%d,%d) holds %q", size[0], size[1], x, y, ch)
				}
			}
		}
		if !scr.has("> Continue") || !scr.has("#") {
			t.Errorf("at %dx%d the plain page is missing its menu or its wordmark", size[0], size[1])
		}
	}
}

// TestMenuCaptionAndTheLinesThatComeAndGo: the caption names the town and
// its age, and the prestige level once the game has one; a newer release
// is announced beside the version; and the forge master's line shows only
// for an account that holds a save with the proof.
func TestMenuCaptionAndTheLinesThatComeAndGo(t *testing.T) {
	r := newMenuRig(t, "ashford")
	m := r.m
	want := m.town.name + " · Primitive Age"
	if len(m.view.captions) == 0 || m.view.captions[0] != want {
		t.Errorf("the caption is %q, want %q first", m.view.captions, want)
	}
	if scr := r.draw(120, 40); scr.has("Prestige") || scr.has("update available") || scr.has(m.eliteMsg[1]) {
		t.Error("a first game shows a prestige level, an update or the forge master's line")
	}

	// A game with prestige behind it.
	st := game.GameState{Age: "bronze_age", AgeName: "Bronze Age"}
	st.Prestige.Level = 3
	caps := menuCaptions(m.town, &st)
	if caps[0] != m.town.name+" · Bronze Age · ★ Prestige level 3" || caps[len(caps)-1] != "Bronze Age" {
		t.Errorf("with prestige the captions are %q", caps)
	}
	m.view.captions = caps
	if scr := r.draw(120, 40); !scr.has("★ Prestige level 3") {
		t.Error("the prestige level is not on the page at 120x40")
	}
	if scr := r.draw(80, 24); !scr.has(m.town.name + " · Bronze Age") {
		t.Error("at 80x24 the caption lost the town and its age")
	}

	// A newer release: beside the version, with its key.
	m.update = true
	m.buildView()
	for _, size := range [][2]int{{120, 40}, {80, 24}} {
		if scr := r.draw(size[0], size[1]); !scr.has("dev · update available (u)") {
			t.Errorf("at %dx%d the update notice is not beside the version", size[0], size[1])
		}
	}
	if got := menuEdition("v4.2.1"); got != "fourth" {
		t.Errorf("v4.2.1 is the %q edition", got)
	}
	if got := menuEdition("dev"); got != "" {
		t.Errorf("a dev build is the %q edition", got)
	}
	if numberWords(len(config.Ages())) != "twenty-two" {
		t.Logf("the game has %d ages; the title's line says %q", len(config.Ages()), numberWords(len(config.Ages())))
	}
	if !r.draw(120, 40).has(numberWords(len(config.Ages())) + " ages, one terminal") {
		t.Error("the line under the wordmark does not count the game's ages")
	}
}
