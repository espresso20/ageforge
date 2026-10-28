package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
)

// TestMapZoomKeys: + - = _ are the map's while the prompt is empty and the
// prompt's otherwise. + and = zoom in and - and _ zoom out (a keyboard with
// no Page keys zooms with them), between the region and the settlement. [ and
// ] are not the map's.
func TestMapZoomKeys(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	p := d.mapPanel
	d.runForTest("map")
	if !strings.Contains(drawMapPanel(t, d, 160, 48), "SETTLEMENT") {
		t.Fatal("the map does not open at the settlement zoom")
	}
	for _, step := range []struct {
		r    rune
		want string
	}{{'-', "REGION"}, {'-', "REGION"}, {'=', "SETTLEMENT"}, {'+', "SETTLEMENT"}, {'_', "REGION"}} {
		if !p.routeKey(key(step.r), "") {
			t.Fatalf("%q did not go to the map", step.r)
		}
		if got := drawMapPanel(t, d, 160, 48); !strings.Contains(got, step.want) {
			t.Errorf("after %q the map does not show %s:\n%s", step.r, step.want, got)
		}
	}
	for _, r := range "[]" {
		if p.routeKey(key(r), "") {
			t.Errorf("%q went to the map; it has no district keys", r)
		}
	}
	for _, r := range "+-=_" {
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
	if !p.routeKey(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), "") || !strings.Contains(drawMapPanel(t, d, 160, 48), "SETTLEMENT") {
		t.Error("PgDn did not zoom in")
	}
}

// TestHelpListsMapZoomKeys: the Help panel's Map section lists the cursor
// and zoom keys at any width, and names no second map.
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
		for _, want := range []string{"Arrows - Move the cursor; Shift moves further", "+/- - Zoom in and out with no Page keys (= works as +)",
			"PgUp/PgDn - Zoom out and in (Option+Up and Option+Down do the same)", "Home/End - The town square", "and + - = while the prompt is empty",
			"Settings are commands: map glyphs, map flows, minimap."} {
			if !strings.Contains(flat, want) {
				t.Errorf("%d columns: the Map section does not say %q:\n%s", screenW, want, section)
			}
		}
		for _, bad := range []string{"—", "!", "skyline", "Skyline", "style"} {
			if strings.Contains(section, bad) {
				t.Errorf("%d columns: the Map section has %q", screenW, bad)
			}
		}
	}
}
