package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// panels_full_test.go holds every panel to one rule: it opens over the whole
// screen but the command bar's rows, and Esc closes it back to the dashboard.
// The dashboard is covered by a page of poison cells for these tests, so a
// panel that leaves any of the screen to the dashboard shows poison above the
// bar.

// poisonRune fills the page that stands in for the dashboard.
const poisonRune = 'א'

// fullSizes are the terminal sizes the panels are drawn at.
var fullSizes = [][2]int{{80, 24}, {100, 30}, {160, 45}}

// panelCommands are the commands that open the panels of the Panels list, and
// the two more overlays a command opens (the Buildings list and the old name
// of the Factions panel).
func panelCommands() []string {
	return append(append([]string(nil), sidebarPanels...), "buildings", "diplomacy")
}

// poisonedStage is a dashboard in the Medieval Age with a page of poison over
// it, so a panel's reach shows.
func poisonedStage(t *testing.T) (*Dashboard, *tview.Pages) {
	t.Helper()
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "medieval_age")
	poison := tview.NewBox()
	poison.SetDrawFunc(func(scr tcell.Screen, x, y, w, h int) (int, int, int, int) {
		for yy := y; yy < y+h; yy++ {
			for xx := x; xx < x+w; xx++ {
				scr.SetContent(xx, yy, poisonRune, nil, tcell.StyleDefault)
			}
		}
		return x, y, w, h
	})
	pages.AddPage("poison", poison, true, true)
	return d, pages
}

// openPanel types cmd at the prompt as the player does, at a terminal w by h.
func openPanel(t *testing.T, d *Dashboard, cmd string, w, h int) {
	t.Helper()
	d.overlayMgr.screenW, d.overlayMgr.screenH = w, h
	d.inputField.SetText(cmd)
	d.submitInput()
	if d.overlayMgr.ActiveName() == "" {
		t.Fatalf("%q opened no panel", cmd)
	}
}

// TestEveryPanelOpensFullScreen: each panel of the Panels list, and the
// others a command opens, fills the screen above the command bar at 80x24,
// 100x30 and 160x45 (nothing of the dashboard shows round it, and the bar's
// three rows are left to the dashboard), and Esc closes it back to the
// dashboard with the command bar holding the keyboard.
func TestEveryPanelOpensFullScreen(t *testing.T) {
	d, pages := poisonedStage(t)
	for _, name := range panelCommands() {
		for _, sz := range fullSizes {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", name, w, h)
			openPanel(t, d, name, w, h)
			cells := drawnDashboard(t, d, pages, w, h)
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					c := cells[y*w+x]
					poisoned := len(c.Runes) > 0 && c.Runes[0] == poisonRune
					if y < h-promptRows && poisoned {
						t.Errorf("%s: the dashboard shows at (%d,%d), above the command bar", where, x, y)
						x, y = w, h // one report a panel
					} else if y >= h-promptRows && !poisoned {
						t.Errorf("%s: the panel covers the command bar at (%d,%d)", where, x, y)
						x, y = w, h
					}
				}
			}
			pressKey(d, pages, tcell.KeyEsc, 0)
			if d.overlayMgr.HasActive() {
				t.Errorf("%s: Esc left %q open", where, d.overlayMgr.ActiveName())
			}
			if front, _ := pages.GetFrontPage(); front != "poison" {
				t.Errorf("%s: after Esc the front page is %q, not the dashboard", where, front)
			}
			if !d.inputField.HasFocus() {
				t.Errorf("%s: after Esc the command bar does not hold the keyboard", where)
			}
		}
	}
}

// frameLooks reads a drawn text panel: its top row names it, its sides and
// its corners are the frame, and its foot says how to close it.
func frameLooks(t *testing.T, where string, cells []tcell.SimCell, w, h int, title string, plain bool) {
	t.Helper()
	edge := func(r rune) rune {
		if plain {
			return mapmodel.Fold(r, mapmodel.TierASCII)
		}
		return r
	}
	rows := rectText(cells, w, 0, 0, w, h-promptRows)
	last := len(rows) - 1
	for y, row := range rows {
		cs := []rune(row)
		if len(cs) != w {
			t.Errorf("%s: row %d is %d cells, the frame is %d wide: %q", where, y, len(cs), w, row)
			return
		}
		var l, r rune = '│', '│'
		switch y {
		case 0:
			l, r = '┌', '┐'
		case last:
			l, r = '└', '┘'
		}
		if cs[0] != edge(l) || cs[w-1] != edge(r) {
			t.Errorf("%s: row %d is %q...%q, not a frame edge", where, y, string(cs[0]), string(cs[w-1]))
			return
		}
	}
	if !strings.Contains(rows[0], title) {
		t.Errorf("%s: the frame's top edge does not name the panel %q: %q", where, title, rows[0])
	}
	if foot := rows[last-1]; !strings.Contains(foot, "Esc") {
		t.Errorf("%s: the foot does not say how to close the panel: %q", where, foot)
	}
}

// TestTextPanelsShareOneFrame: each text panel is drawn in the same frame
// at the three sizes, in every glyph tier and in every theme, with the text
// inside it and nothing beyond its edges.
func TestTextPanelsShareOneFrame(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "medieval_age")
	st := eng.GetState()
	tiers := []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd}
	var cur mapmodel.GlyphTier
	// a dashboard whose pages hold only the panel: the frame is judged on its own
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	d.overlayMgr.plain = func() bool { return cur == mapmodel.TierASCII }
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for _, sz := range fullSizes {
			w, h := sz[0], sz[1]
			// every theme at every size is a lot of draws: the other tiers
			// at the middle size are enough
			for _, tier := range tiers {
				if tier != mapmodel.TierUnicode && sz != fullSizes[1] {
					continue
				}
				cur = tier
				for name, e := range d.overlayMgr.entries {
					where := fmt.Sprintf("%s at %dx%d, %s, %v", name, w, h, th.Key, tier)
					d.overlayMgr.screenW, d.overlayMgr.screenH = w, h
					d.overlayMgr.Show(name, st)
					cells := drawnPages(t, pages, w, h)
					frameLooks(t, where, cells, w, h, e.title.String(), tier == mapmodel.TierASCII)
					for _, c := range cells {
						if tier == mapmodel.TierASCII && len(c.Runes) > 0 && c.Runes[0] == poisonRune {
							t.Fatalf("%s: stray poison", where)
						}
					}
					d.overlayMgr.Hide()
				}
			}
		}
	}
}

// drawnPages draws pages at w by h on a simulated screen.
func drawnPages(t *testing.T, pages *tview.Pages, w, h int) []tcell.SimCell {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	screen := theme.WrapScreen(sim)
	for i := 0; i < 2; i++ {
		screen.Clear()
		pages.SetRect(0, 0, w, h)
		pages.Draw(screen)
		screen.Show()
	}
	cells, _, _ := sim.GetContents()
	return append([]tcell.SimCell(nil), cells...)
}

// TestTextPanelsScrollWithTheBarEmpty: a text panel that leaves the keyboard
// with the command bar takes the keys that print nothing while the bar is
// empty, hands every key to the bar when it has text, and the bar gets the
// letters.
func TestTextPanelsScrollWithTheBarEmpty(t *testing.T) {
	d, pages := poisonedStage(t)
	openPanel(t, d, "help", 100, 30)
	if !d.inputField.HasFocus() {
		t.Fatal("the command bar lost the keyboard to the Help panel")
	}
	drawnDashboard(t, d, pages, 100, 30)
	tv := d.overlayMgr.entries["help"].tv
	row := func() int { r, _ := tv.GetScrollOffset(); return r }
	if row() != 0 {
		t.Fatalf("the panel opened scrolled to row %d", row())
	}
	for _, k := range []tcell.Key{tcell.KeyDown, tcell.KeyPgDn} {
		before := row()
		if ev := sendKey(d, k); ev != nil || row() <= before {
			t.Errorf("key %v with the bar empty: taken=%v, row %d -> %d", k, ev == nil, before, row())
		}
	}
	if sendKey(d, tcell.KeyHome) != nil || row() != 0 {
		t.Errorf("Home did not scroll to the top (row %d)", row())
	}
	taken := sendKey(d, tcell.KeyEnd) == nil
	drawnDashboard(t, d, pages, 100, 30) // the text view follows its end when it is drawn
	if !taken || row() == 0 {
		t.Errorf("End did not scroll to the foot (taken %v, row %d)", taken, row())
	}
	// Text in the bar: the keys are the bar's.
	d.inputField.SetText("build")
	before := row()
	if ev := sendKey(d, tcell.KeyPgUp); ev == nil || row() != before {
		t.Errorf("PgUp with text in the bar was taken by the panel (row %d -> %d)", before, row())
	}
	// A letter is the bar's, with the bar empty or not.
	d.inputField.SetText("")
	if ev := d.root.GetInputCapture()(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone)); ev == nil {
		t.Error("a letter was taken by the Help panel")
	}
}

// TestKeyedPanelsKeepTheKeyboard: the Wonders, Harbinger and Plan panels have
// keys that are letters, so they keep the keyboard as they always did, and
// their foot names those keys.
func TestKeyedPanelsKeepTheKeyboard(t *testing.T) {
	d, pages := poisonedStage(t)
	for name, keys := range map[string][]string{
		"wonders":   {"Tab", "Esc"},
		"harbinger": {"A", "B", "I", "Esc"},
		"plan":      {"U D", "X", "C", "Esc"},
	} {
		openPanel(t, d, name, 100, 30)
		if d.inputField.HasFocus() {
			t.Errorf("%s: the command bar has the keyboard, so a letter would print instead of acting", name)
		}
		cells := drawnDashboard(t, d, pages, 100, 30)
		foot := rectText(cells, 100, 0, 30-promptRows-2, 100, 1)[0]
		for _, k := range keys {
			if !strings.Contains(foot, k) {
				t.Errorf("%s: the foot does not name %q: %q", name, k, foot)
			}
		}
		if strings.Contains(foot, "or type") {
			t.Errorf("%s: the foot offers the command bar, which this panel does not give the keyboard: %q", name, foot)
		}
		pressKey(d, pages, tcell.KeyEsc, 0)
		if d.overlayMgr.HasActive() {
			t.Errorf("%s: Esc did not close the panel", name)
		}
	}
}

// TestPanelFooterOffersTheBar: a text panel that leaves the keyboard with the
// command bar says so on its foot, and a game's panels all name Esc.
func TestPanelFooterOffersTheBar(t *testing.T) {
	d, pages := poisonedStage(t)
	for _, name := range []string{"milestones", "army", "trade", "factions", "stats", "logs", "epoch", "history", "buildings", "help"} {
		openPanel(t, d, name, 100, 30)
		cells := drawnDashboard(t, d, pages, 100, 30)
		foot := rectText(cells, 100, 0, 30-promptRows-2, 100, 1)[0]
		for _, want := range []string{"scroll", "Esc", "or type"} {
			if !strings.Contains(foot, want) {
				t.Errorf("%s: the foot does not say %q: %q", name, want, foot)
			}
		}
		pressKey(d, pages, tcell.KeyEsc, 0)
	}
}
