package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
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
	acct, err := game.CreateNamedAccount("Menu App")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range games {
		mk := game.NewGameEngine()
		mk.SetAccount(acct)
		if err := mk.StartNewNamedGame(name); err != nil {
			t.Fatal(err)
		}
	}
	eng := game.NewGameEngine()
	eng.SetAccount(acct)
	a := NewApp(eng, "dev") // "dev" skips the network update check
	sim := tcell.NewSimulationScreen("UTF-8")
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
	screen := func() string {
		var sb strings.Builder
		h.onUI(func() {
			cells, w, hh := h.sim.GetContents()
			for y := 0; y < hh; y++ {
				for x := 0; x < w; x++ {
					if r := cells[y*w+x].Runes; len(r) > 0 {
						sb.WriteRune(r[0])
					} else {
						sb.WriteByte(' ')
					}
				}
				sb.WriteByte('\n')
			}
		})
		return sb.String()
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
	}{{'t', themePickerPage}, {'l', loadGamePage}, {'a', accountsPage}, {'b', menuBadgesPage}} {
		h.key(tcell.KeyRune, c.key)
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
		h.key(tcell.KeyEsc, 0)
		h.waitFor("the menu after "+c.page, 5*time.Second, func() bool { return h.frontPage() == "splash" })
		moves("back from " + c.page)
	}

	// Continue: the letter moved the selection to the entry it opened, so
	// Home puts it back on the first entry and Enter opens the game.
	h.key(tcell.KeyHome, 0)
	h.key(tcell.KeyEnter, 0)
	h.waitFor("the game", 5*time.Second, func() bool { return h.frontPage() == "dashboard" && h.inputHasFocus() })
	waitRunning(t, h.eng)
	if got := h.eng.ActiveSaveName(); got != "ashford" {
		t.Errorf("Continue opened %q", got)
	}
	if clockRuns() {
		t.Error("the menu's clock runs behind the game")
	}
	// Esc in the game saves, stops it and returns to the menu.
	h.key(tcell.KeyEsc, 0)
	h.waitFor("the menu after the game", 5*time.Second, func() bool { return h.frontPage() == "splash" })
	moves("back from the game")
	var sel string
	h.onUI(func() { sel = m.view.items[m.view.sel].id })
	if sel != miContinue {
		t.Errorf("back from the game the selection is on %q", sel)
	}

	// The smallest terminal, live.
	h.sim.SetSize(80, 24)
	if err := h.sim.PostEvent(tcell.NewEventResize(80, 24)); err != nil {
		t.Fatal(err)
	}
	h.waitFor("the menu at 80x24", 5*time.Second, func() bool {
		s := screen()
		return strings.Contains(s, "▸ Continue") && strings.Contains(s, "Quit") && strings.Contains(s, "choose")
	})
	moves("at 80x24")

	// Motion off (the setting is the account's): the next time the menu is
	// returned to it holds still, with no clock.
	if err := acct.SetMotion(false); err != nil {
		t.Fatal(err)
	}
	h.key(tcell.KeyRune, 't')
	h.waitFor("the theme picker", 5*time.Second, func() bool { return h.frontPage() == themePickerPage })
	h.key(tcell.KeyEsc, 0)
	h.waitFor("the menu with motion off", 5*time.Second, func() bool { return h.frontPage() == "splash" })
	h.waitFor("the clock stopping", 3*time.Second, func() bool { return !clockRuns() })
	still, at := screen(), frame()
	time.Sleep(4 * mapAnimStep)
	if frame() != at || screen() != still {
		t.Error("with motion off the menu moved")
	}
	// And with motion off a choice opens at once, with no strike.
	h.key(tcell.KeyRune, 'l')
	h.waitFor("the Load Game browser", 5*time.Second, func() bool { return h.frontPage() == loadGamePage })
	h.key(tcell.KeyEsc, 0)
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
	h.key(tcell.KeyRune, 'n')
	h.waitFor("the name prompt", 5*time.Second, func() bool { return h.frontPage() == newGameNamePage })
	h.key(tcell.KeyEnter, 0)
	h.waitFor("the game", 5*time.Second, func() bool { return h.frontPage() == "dashboard" && h.inputHasFocus() })
	waitRunning(t, h.eng)
	h.key(tcell.KeyEsc, 0)
	h.waitFor("the menu after the game", 5*time.Second, func() bool { return h.frontPage() == "splash" })
	h.waitFor("the town page", 3*time.Second, func() bool { return !forge() })
	var first, name string
	h.onUI(func() { first, name = m.view.items[m.view.sel].id, m.cur.Save.Name })
	if first != miContinue || name != h.eng.ActiveSaveName() {
		t.Errorf("after a first game the menu selects %q and Continue opens %q (the game was %q)", first, name, h.eng.ActiveSaveName())
	}
}
