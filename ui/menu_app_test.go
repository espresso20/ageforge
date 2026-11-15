package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// menu_app_test.go runs the main menu in the real App, on its own event
// loop and a simulated terminal, with real time passing: what the unit
// tests take on trust (tview focusing the menu when a page closes, the
// redraw clock, the strike's short wait) is checked here end to end.

// bootMenuApp starts the App at w by h on an account that has a game.
func bootMenuApp(t *testing.T, w, h int, games ...string) (*reproHarness, *game.Account) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	t.Cleanup(func() { _ = theme.SetActive(theme.DefaultKey) })
	acct, err := game.CreateAccount("Menu App")
	must(t, err)
	for _, name := range games {
		mk := game.NewGameEngine()
		mk.SetAccount(acct)
		must(t, mk.StartNewNamedGame(name))
	}
	eng := game.NewGameEngine()
	eng.SetAccount(acct)
	a := NewApp(eng, "dev") // "dev" skips the network update check
	sim := newSim()
	a.SetScreen(sim)
	sim.SetSize(w, h)
	hr := &reproHarness{t: t, eng: eng, a: a, sim: sim, runErr: make(chan error, 1)}
	go func() { hr.runErr <- a.Run() }()
	t.Cleanup(hr.teardown) // before the data dir is put back
	hr.waitFor("the main menu", 5*time.Second, func() bool { return hr.frontPage() == "splash" })
	return hr, acct
}

// TestMenuInTheRunningApp: the menu moves by itself, every page it opens
// stops it and gives the keyboard back when it closes, Continue opens the
// game and Esc in the game returns to a menu that moves again, a resize is
// taken in stride, and motion off stops the clock for good.
func TestMenuInTheRunningApp(t *testing.T) {
	h, acct := bootMenuApp(t, 120, 40, "ashford")
	var m *mainMenu
	h.onUI(func() {
		_, p := h.a.pages.GetFrontPage()
		m, _ = p.(*mainMenu)
	})
	if m == nil {
		t.Fatal("the page in front is not the main menu")
	}
	frame := func() (n int) {
		h.onUI(func() {
			if m.scene != nil {
				n = m.scene.frame
			}
		})
		return n
	}
	clockRuns := func() (on bool) {
		h.onUI(func() { on = m.stop != nil })
		return on
	}
	screen := func() (s string) {
		h.onUI(func() { s = simCells(h.sim).Text() })
		return s
	}
	moves := func(what string) {
		t.Helper()
		from := frame()
		h.waitFor(what+": the menu moving", 4*time.Second, func() bool { return frame() >= from+3 })
		if !clockRuns() {
			t.Errorf("%s: the menu moved but no clock runs", what)
		}
	}

	moves("at launch")
	if s := screen(); !strings.Contains(s, "▸ Continue") || !strings.Contains(s, "ashford") || !strings.Contains(s, "ages, one terminal") {
		t.Fatalf("the menu with a game is not on the screen:\n%s", s)
	}

	for _, c := range []struct {
		key  rune
		page string
	}{{'l', loadGamePage}, {'a', accountsPage}, {'b', menuBadgesPage}} {
		h.key(kit.KeyRune, c.key)
		h.waitFor(c.page+" in front", 5*time.Second, func() bool { return h.frontPage() == c.page })
		if clockRuns() {
			t.Errorf("the menu's clock still runs under %s", c.page)
		}
		at := frame()
		time.Sleep(3 * mapAnimStep)
		if frame() != at {
			t.Errorf("the menu moved under %s", c.page)
		}
		if strings.Contains(screen(), "ages, one terminal") {
			t.Errorf("the menu shows through %s", c.page)
		}
		h.key(kit.KeyEsc, 0)
		h.waitFor("the menu after "+c.page, 5*time.Second, func() bool { return h.frontPage() == "splash" })
		moves("back from " + c.page)
	}

	// Continue: the letter moved the selection to the entry it opened, so
	// Home puts it back on the first entry and Enter opens the game.
	h.key(kit.KeyHome, 0)
	h.key(kit.KeyEnter, 0)
	h.waitFor("the game", 5*time.Second, func() bool { return h.frontPage() == "dashboard" && h.inputHasFocus() })
	waitRunning(t, h.eng)
	if got := h.eng.ActiveSaveName(); got != "ashford" {
		t.Errorf("Continue opened %q", got)
	}
	if clockRuns() {
		t.Error("the menu's clock runs behind the game")
	}
	// Esc in the game saves, stops it and returns to the menu.
	h.key(kit.KeyEsc, 0)
	h.waitFor("the menu after the game", 5*time.Second, func() bool { return h.frontPage() == "splash" })
	moves("back from the game")
	var sel string
	h.onUI(func() { sel = m.view.items[m.view.sel].id })
	if sel != miContinue {
		t.Errorf("back from the game the selection is on %q", sel)
	}

	// The smallest terminal, live.
	h.sim.SetSize(80, 24)
	h.waitFor("the menu at 80x24", 5*time.Second, func() bool {
		s := screen()
		return strings.Contains(s, "▸ Continue") && strings.Contains(s, "Quit") && strings.Contains(s, "choose")
	})
	moves("at 80x24")

	// Motion off (the setting is the account's): the next time the menu is
	// returned to it holds still, with no clock.
	must(t, acct.SetMotion(false))
	h.key(kit.KeyRune, 's')
	h.waitFor("the Settings page", 5*time.Second, func() bool { return strings.Contains(screen(), "Minimap") })
	h.key(kit.KeyRune, 't')
	h.waitFor("the theme picker", 5*time.Second, func() bool { return h.frontPage() == themePickerPage })
	h.key(kit.KeyEsc, 0)
	h.waitFor("Settings after the picker", 5*time.Second, func() bool { return h.frontPage() == "splash" })
	h.key(kit.KeyEsc, 0)
	h.waitFor("the menu with motion off", 5*time.Second, func() bool { return h.frontPage() == "splash" })
	h.waitFor("the clock stopping", 3*time.Second, func() bool { return !clockRuns() })
	still, at := screen(), frame()
	time.Sleep(4 * mapAnimStep)
	if frame() != at || screen() != still {
		t.Error("with motion off the menu moved")
	}
	// And with motion off a choice opens at once, with no strike.
	h.key(kit.KeyRune, 'l')
	h.waitFor("the Load Game browser", 5*time.Second, func() bool { return h.frontPage() == loadGamePage })
	h.key(kit.KeyEsc, 0)
	h.waitFor("the menu again", 5*time.Second, func() bool { return h.frontPage() == "splash" })
}

// TestMenuFirstRunInTheRunningApp: a launch with an account and no game
// opens on the first page, and New game from it starts a game whose town is
// behind the menu on the way back.
func TestMenuFirstRunInTheRunningApp(t *testing.T) {
	h, _ := bootMenuApp(t, 100, 30)
	var m *mainMenu
	h.onUI(func() {
		_, p := h.a.pages.GetFrontPage()
		m, _ = p.(*mainMenu)
	})
	forge := func() (f bool) {
		h.onUI(func() { f = m.view.forge })
		return f
	}
	if !forge() {
		t.Fatal("a launch with no game is not on the first page")
	}
	h.key(kit.KeyRune, 'n')
	h.waitFor("the name prompt", 5*time.Second, func() bool { return h.frontPage() == newGameNamePage })
	h.key(kit.KeyEnter, 0)
	h.waitFor("the game", 5*time.Second, func() bool { return h.frontPage() == "dashboard" && h.inputHasFocus() })
	waitRunning(t, h.eng)
	h.key(kit.KeyEsc, 0)
	h.waitFor("the menu after the game", 5*time.Second, func() bool { return h.frontPage() == "splash" })
	h.waitFor("the town page", 3*time.Second, func() bool { return !forge() })
	var first, name string
	h.onUI(func() { first, name = m.view.items[m.view.sel].id, m.cur.Save.Name })
	if first != miContinue || name != h.eng.ActiveSaveName() {
		t.Errorf("after a first game the menu selects %q and Continue opens %q (the game was %q)", first, name, h.eng.ActiveSaveName())
	}
}

// TestARightClickGoesBackInTheRunningApp: a right click does what Esc does on
// a page, a Settings page and a panel, and nothing where Esc would leave: the
// main menu's first page and the dashboard with no panel open.
func TestARightClickGoesBackInTheRunningApp(t *testing.T) {
	h, _ := bootMenuApp(t, 120, 40, "ashford")
	var m *mainMenu
	h.onUI(func() {
		_, p := h.a.pages.GetFrontPage()
		m, _ = p.(*mainMenu)
	})
	if m == nil {
		t.Fatal("the page in front is not the main menu")
	}
	rightClick := func() {
		h.a.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonRight, X: 2, Y: 2})
		h.a.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonRight, X: 2, Y: 2})
		h.onUI(func() {}) // the click has run
		h.onUI(func() {}) // and so has the key it pressed
	}
	inSettings := func() (on bool) {
		h.onUI(func() { on = m.settings })
		return on
	}

	rightClick()
	if h.frontPage() != "splash" || inSettings() {
		t.Fatalf("a right click on the main menu's first page did something: front page %q, in Settings %v", h.frontPage(), inSettings())
	}
	h.key(kit.KeyRune, 's')
	h.waitFor("the Settings page", 5*time.Second, inSettings)
	rightClick()
	if inSettings() {
		t.Error("a right click on the Settings page did not go back")
	}
	h.key(kit.KeyRune, 'l')
	h.waitFor("the Load game page in front", 5*time.Second, func() bool { return h.frontPage() == loadGamePage })
	rightClick()
	h.waitFor("the menu after a right click on Load game", 5*time.Second, func() bool { return h.frontPage() == "splash" })

	// Into the game: on the bare dashboard a right click must not leave it.
	h.key(kit.KeyHome, 0)
	h.key(kit.KeyEnter, 0)
	h.waitFor("the dashboard in front", 10*time.Second, func() bool { return h.frontPage() == "dashboard" })
	rightClick() // returns once the click, and any key it pressed, have run
	if h.frontPage() != "dashboard" {
		t.Fatalf("a right click on the dashboard with no panel open left it for %q", h.frontPage())
	}
	// With a panel open it closes the panel and stays in the game.
	h.submit("help")
	open := func() (on bool) {
		h.onUI(func() { on = h.a.dashboard.overlayMgr.HasActive() })
		return on
	}
	h.waitFor("the Help panel open", 5*time.Second, open)
	rightClick()
	h.waitFor("the Help panel closed by a right click", 5*time.Second, func() bool { return !open() })
	if h.frontPage() != "dashboard" {
		t.Errorf("after the right click the page in front is %q, want the dashboard", h.frontPage())
	}
}
