package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// zoomLevel is the open Map panel's style zoom ("" when it has none).
func zoomLevel(d *Dashboard) string {
	if z, ok := d.mapPanel.current().(mapstyle.Zoomable); ok {
		return z.ZoomLevel()
	}
	return ""
}

// TestMapZoomKeys: + - = [ ] are the map's while the prompt is empty and the
// prompt's otherwise. + and = zoom in and - and _ zoom out in both styles,
// the roguelike too (a keyboard with no Page keys zooms with them); [ and ]
// are the skyline's district keys, so they stay the prompt's in the
// roguelike.
func TestMapZoomKeys(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	p := d.mapPanel
	d.runForTest("map")
	if !strings.Contains(drawMapPanel(t, d, 160, 48), "SETTLEMENT") {
		t.Fatal("the roguelike does not open at the settlement zoom")
	}
	for _, step := range []struct {
		r    rune
		want string
	}{{'-', "REGION"}, {'-', "REGION"}, {'=', "SETTLEMENT"}, {'+', "DISTRICT"}, {'_', "SETTLEMENT"}} {
		if !p.routeKey(key(step.r), "") {
			t.Fatalf("%q did not go to the roguelike map", step.r)
		}
		if got := drawMapPanel(t, d, 160, 48); !strings.Contains(got, step.want) {
			t.Errorf("after %q the roguelike does not show %s:\n%s", step.r, step.want, got)
		}
	}
	for _, r := range "[]" {
		if p.routeKey(key(r), "") {
			t.Errorf("%q went to the roguelike map; it has no district keys", r)
		}
	}
	for _, r := range "+-=_" {
		if p.routeKey(key(r), "bui") {
			t.Errorf("%q went to the roguelike map with text in the prompt", r)
		}
	}
	d.overlayMgr.Hide()
	d.runForTest("map style skyline")
	d.runForTest("map")
	drawMapPanel(t, d, 160, 48)
	if zoomLevel(d) != "mid" {
		t.Fatalf("the skyline opens at %q", zoomLevel(d))
	}
	for _, step := range []struct {
		r    rune
		want string
	}{{'+', "close"}, {'-', "mid"}, {'-', "wide"}, {'-', "wide"}, {'=', "mid"}} { // (this town is too small for a far level of its own: wide shows all of it)
		if !p.routeKey(key(step.r), "") {
			t.Fatalf("%q did not go to the skyline", step.r)
		}
		if zoomLevel(d) != step.want {
			t.Errorf("after %q the zoom is %q, want %q", step.r, zoomLevel(d), step.want)
		}
		drawMapPanel(t, d, 160, 48)
	}
	for _, r := range "+-=_[]" {
		if p.routeKey(key(r), "bui") {
			t.Errorf("%q went to the map with text in the prompt", r)
		}
	}
	for _, r := range "sgzxhjklfnc?" { // letters are still the prompt's
		if p.routeKey(key(r), "") {
			t.Errorf("%q went to the map", r)
		}
	}
	if p.routeKey(tcell.NewEventKey(tcell.KeyRune, '+', tcell.ModCtrl), "") || p.routeKey(tcell.NewEventKey(tcell.KeyRune, '[', tcell.ModAlt), "") {
		t.Error("a Ctrl or Alt chord went to the map")
	}
	before := zoomLevel(d)
	if !p.routeKey(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), "") || zoomLevel(d) == before {
		t.Error("PgDn did not zoom in")
	}
}

// TestMapZoomPersists: the zoom level is a map setting: it is saved to the
// account when a key changes it and the next open (and the next session) has
// it, and an account switch swaps it.
func TestMapZoomPersists(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	first := eng.Account()
	d.runForTest("map style skyline")
	d.runForTest("map")
	drawMapPanel(t, d, 160, 48)
	if first.MapZoom() != "" {
		t.Fatalf("a new account has zoom %q", first.MapZoom())
	}
	d.mapPanel.routeKey(key('+'), "")
	if first.MapZoom() != "close" || d.mapSettings().Zoom != "close" {
		t.Fatalf("after +: account %q, settings %q", first.MapZoom(), d.mapSettings().Zoom)
	}
	d.overlayMgr.Hide()
	// a fresh panel (a new session) opens at the saved level
	d.mapPanel.styles = styleSet{reg: d.mapViews.reg}
	d.runForTest("map")
	drawMapPanel(t, d, 160, 48)
	if zoomLevel(d) != "close" {
		t.Errorf("the next open is at %q, want close", zoomLevel(d))
	}
	d.mapPanel.routeKey(key('-'), "")
	d.mapPanel.routeKey(key('-'), "")
	if first.MapZoom() != "wide" {
		t.Errorf("the account has %q, want wide", first.MapZoom())
	}
	d.overlayMgr.Hide()

	second, err := game.CreateAccount("Second")
	if err != nil {
		t.Fatal(err)
	}
	eng.SetAccount(second)
	d.mapPanel.styles = styleSet{reg: d.mapViews.reg}
	d.runForTest("map style skyline")
	d.runForTest("map")
	drawMapPanel(t, d, 160, 48)
	if zoomLevel(d) != "mid" || d.mapSettings().Zoom != "" {
		t.Errorf("a new account opens at %q (setting %q), want mid", zoomLevel(d), d.mapSettings().Zoom)
	}
	d.overlayMgr.Hide()
	back, err := game.SwitchAccount(first.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	eng.SetAccount(back)
	d.mapPanel.styles = styleSet{reg: d.mapViews.reg}
	d.runForTest("map")
	drawMapPanel(t, d, 160, 48)
	if zoomLevel(d) != "wide" {
		t.Errorf("switching back (read from disk): %q, want wide", zoomLevel(d))
	}
}

// TestMapZoomWithoutAccount: with no account the level is kept for the
// session.
func TestMapZoomWithoutAccount(t *testing.T) {
	d, _ := mapTestDashboard(t, false)
	d.runForTest("map style skyline")
	d.runForTest("map")
	drawMapPanel(t, d, 160, 48)
	d.mapPanel.routeKey(key('+'), "")
	if d.mapSettings().Zoom != "close" {
		t.Errorf("session zoom %q, want close", d.mapSettings().Zoom)
	}
	d.overlayMgr.Hide()
	d.runForTest("map")
	drawMapPanel(t, d, 160, 48)
	if zoomLevel(d) != "close" {
		t.Errorf("the next open is at %q", zoomLevel(d))
	}
}

// TestMapZoomFooter: the open skyline's footer says the zoom and its keys, at
// every level.
func TestMapZoomFooter(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	d.runForTest("map style skyline")
	d.runForTest("map")
	drawMapPanel(t, d, 160, 48)
	for _, step := range []struct {
		want string
		next rune // the key that goes on to the next level
	}{{"mid", '+'}, {"close", '-'}, {"mid", '-'}, {"wide", 0}} {
		txt := drawMapPanel(t, d, 160, 48)
		if !strings.Contains(txt, "zoom: "+step.want) || !strings.Contains(txt, "[ ] district") {
			t.Errorf("the footer does not say zoom: %s and the district keys:\n%s", step.want, txt)
		}
		if step.next != 0 {
			d.mapPanel.routeKey(key(step.next), "")
		}
	}
}

// TestHelpListsMapZoomKeys: the Help panel's Map section lists the zoom and
// district keys at any width.
func TestHelpListsMapZoomKeys(t *testing.T) {
	for _, screenW := range []int{0, 80, 120, 200} {
		help := visible(safeTags(helpProvider(game.GameState{}, screenW)))
		at := strings.Index(help, "The Map panel")
		if at < 0 {
			t.Fatal("the Help panel has no Map section")
		}
		section := help[at:]
		if end := strings.Index(section, "═══ The Research panel"); end > 0 {
			section = section[:end]
		}
		flat := strings.Join(strings.Fields(section), " ")
		for _, want := range []string{"+/- - Zoom in and out with no Page keys, in either style", "far shows the whole town", "close every building twice the size", "[ ] - Skyline: jump to the previous and next district",
			"PgUp/PgDn - Zoom out and in (Option+Up and Option+Down do the same)", "Shift moves further, half a screen in the skyline", "and + - = [ ] while the prompt is empty"} {
			if !strings.Contains(flat, want) {
				t.Errorf("%d columns: the Map section does not say %q:\n%s", screenW, want, section)
			}
		}
		for _, bad := range []string{"—", "!"} {
			if strings.Contains(section, bad) {
				t.Errorf("%d columns: the Map section has %q", screenW, bad)
			}
		}
	}
}
