package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
)

func TestMenuShowsSettingsNotThemes(t *testing.T) {
	r := newMenuRig(t)
	got := ids(r.m.view.items)
	if !strings.Contains(got, miSettings) || strings.Contains(got, "themes") {
		t.Fatalf("main menu entries are %q", got)
	}
	scr := r.draw(120, 40)
	if !scr.has("Settings") || scr.has("Themes") {
		t.Errorf("the page should show Settings and not Themes")
	}
}

func TestMenuSettingsRowsSave(t *testing.T) {
	r := newMenuRig(t)
	t.Cleanup(func() { vibe.Set(false) })
	r.press('s')
	if !r.m.settings || ids(r.m.view.items) != "set_theme set_vibe set_glyphs set_motion set_minimap set_back" {
		t.Fatalf("Settings page not open: %v %q", r.m.settings, ids(r.m.view.items))
	}
	// Motion: Enter flips it and the account holds it.
	if !r.acct.MotionOn() {
		t.Fatal("motion should start on")
	}
	r.press('m')
	if r.acct.MotionOn() {
		t.Error("motion row did not turn motion off")
	}
	r.press('m')
	if !r.acct.MotionOn() {
		t.Error("motion row did not turn motion back on")
	}
	// Minimap.
	r.press('i')
	if r.acct.MinimapOn() {
		t.Error("minimap row did not turn it off")
	}
	// Glyphs: right steps forward from unicode, left steps back.
	r.press('g')
	if g, _ := r.acct.MapPrefs(); g != "nerd" {
		t.Errorf("glyphs after one step: %q", g)
	}
	r.key(tcell.KeyLeft, 0)
	if g, _ := r.acct.MapPrefs(); g != "unicode" {
		t.Errorf("glyphs after left: %q", g)
	}
	r.key(tcell.KeyRight, 0)
	r.key(tcell.KeyRight, 0)
	if g, _ := r.acct.MapPrefs(); g != "ascii" || r.m.set.Tier != mapmodel.TierASCII {
		t.Errorf("glyphs after wrapping: %q", g)
	}
	// VibeMode.
	r.press('v')
	if !r.acct.VibeOn() || !vibe.On() {
		t.Error("vibe row did not turn VibeMode on")
	}
	r.press('v')
	if r.acct.VibeOn() {
		t.Error("vibe row did not turn VibeMode off")
	}
}

func TestMenuSettingsEscAndThemeReturn(t *testing.T) {
	r := newMenuRig(t)
	r.press('s')
	r.key(tcell.KeyDown, 0)
	r.key(tcell.KeyUp, 0)
	r.key(tcell.KeyEsc, 0)
	if r.m.settings || !strings.Contains(ids(r.m.view.items), miSettings) {
		t.Fatalf("Esc did not return to the main menu: %q", ids(r.m.view.items))
	}
	if r.m.view.items[r.m.view.sel].id != miSettings {
		t.Error("the main menu should land on Settings")
	}
	// The theme row opens the picker, and back returns to Settings.
	r.press('s')
	r.press('t')
	if got := r.front(); got != themePickerPage {
		t.Fatalf("theme row opened %q", got)
	}
	r.key(tcell.KeyEsc, 0)
	if r.front() != "splash" || !r.m.settings {
		t.Errorf("back from the picker: front %q, settings %v", r.front(), r.m.settings)
	}
}
