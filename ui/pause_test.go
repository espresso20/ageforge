package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

func pausedDashboard(t *testing.T) (*Dashboard, *game.GameEngine) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	keepTheme(t)
	eng := game.NewGameEngine()
	acct, err := game.CreateAccount("Pause Tester")
	must(t, err)
	eng.SetAccount(acct)
	must(t, eng.StartNewNamedGame("pause-test"))
	d, _ := newTestDashboard(eng)
	return d, eng
}

func isPaused(d *Dashboard) bool {
	front, _ := d.pages.GetFrontPage()
	return front == pausePage
}

func TestEscOpensAndClosesThePauseCard(t *testing.T) {
	d, eng := pausedDashboard(t)
	sendKey(d, kit.KeyEsc)
	if !isPaused(d) || !eng.Held() {
		t.Fatalf("Esc left paused=%v held=%v", isPaused(d), eng.Held())
	}
	rootKey(d.pages, tk(kit.KeyEsc, 0, 0))
	if isPaused(d) || eng.Held() {
		t.Errorf("Esc on the card left paused=%v held=%v", isPaused(d), eng.Held())
	}
}

func TestEscWithAPanelClosesThePanelOnly(t *testing.T) {
	d, eng := pausedDashboard(t)
	d.overlayMgr.Show("help", eng.GetState())
	if sendKey(d, kit.KeyEsc) != nil || d.overlayMgr.HasActive() || isPaused(d) || eng.Held() {
		t.Errorf("Esc over a panel: active=%v paused=%v held=%v", d.overlayMgr.HasActive(), isPaused(d), eng.Held())
	}
}

func TestPauseSaveWritesAndMainMenuLeaves(t *testing.T) {
	d, eng := pausedDashboard(t)
	sendKey(d, kit.KeyEsc)
	d.pages.AddPage("splash", newNativeFlex(), true, false)
	before := eng.SavedAt()
	rootKey(d.pages, tk(kit.KeyRune, 'v', 0))
	if !eng.SavedAt().After(before) {
		t.Error("Save did not write a save")
	}
	if front, _ := d.pages.GetFrontPage(); front != pausePage {
		t.Fatalf("Save left the card: front is %q", front)
	}
	rootKey(d.pages, tk(kit.KeyRune, 'm', 0))
	if front, _ := d.pages.GetFrontPage(); front != "splash" || eng.Held() {
		t.Errorf("Main menu left front %q held=%v, want the splash", front, eng.Held())
	}
}

func TestPauseSettingsTabsEscAndMotion(t *testing.T) {
	d, _ := pausedDashboard(t)
	sendKey(d, kit.KeyEsc)
	press := func(c kit.KeyCode, r rune) { rootKey(d.pages, tk(c, r, 0)) }
	press(kit.KeyRune, 's')
	_, item := d.pages.GetFrontPage()
	p := item.(*pauseLayer)
	press(kit.KeyTab, 0)
	press(kit.KeyEnter, 0)
	if p.view != pauseMotion {
		t.Fatalf("view %d after pressing the Motion tab", p.view)
	}
	was := d.mapSettings().Motion
	press(kit.KeyTab, 0) // past the other tabs to the Motion row's button
	press(kit.KeyTab, 0)
	press(kit.KeyEnter, 0)
	if d.mapSettings().Motion == was {
		t.Error("the Motion button changed nothing")
	}
	press(kit.KeyEsc, 0)
	if p.view != pauseMain || !isPaused(d) {
		t.Errorf("Esc in settings: view %d paused=%v, want the card", p.view, isPaused(d))
	}
}

func TestPauseCardFits(t *testing.T) {
	for _, sz := range [][2]int{{279, 74}, {157, 46}} {
		d, _ := pausedDashboard(t)
		d.openPause()
		_, item := d.pages.GetFrontPage()
		p := item.(*pauseLayer)
		for _, v := range []pauseView{pauseMain, pauseDisplay, pauseMotion, pauseKeysTab} {
			p.go2(v)
			p.Layout(image.Rect(0, 0, sz[0], sz[1]))
			body := image.Rect(0, 0, sz[0], sz[1]-barRows(sz[0], sz[1]))
			if p.box.Empty() || !p.box.In(body) {
				t.Errorf("%dx%d view %d: card %v outside the body %v", sz[0], sz[1], v, p.box, body)
			}
			for _, b := range p.btns {
				if r := b.Bounds(); !r.Empty() && !r.In(p.box) {
					t.Errorf("%dx%d view %d: button %q at %v outside the card %v", sz[0], sz[1], v, b.GetLabel(), r, p.box)
				}
			}
		}
	}
}

// A splash waiting for the next refresh does not open over the card: its close
// would leave the keyboard on the Command box behind the dimmed town.
func TestPauseHoldsBackTheRefresh(t *testing.T) {
	d, _ := pausedDashboard(t)
	sendKey(d, kit.KeyEsc)
	d.pendingAgeSplash = "bronze_age"
	d.refresh()
	if d.pendingAgeSplash == "" || !isPaused(d) {
		t.Errorf("refresh ran while paused: pending %q paused=%v", d.pendingAgeSplash, isPaused(d))
	}
}

// TestMapGlyphsLeavesTheBuildingsBoxAlone: the Map glyphs setting changes
// the map and the mini map only. With the pause card open the Buildings
// box draws the same cells, dimmed the same, in either mode.
func TestMapGlyphsLeavesTheBuildingsBoxAlone(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	keepTheme(t)
	const w, h = 279, 74
	draw := func(mode string) ([]simCell, [4]int) {
		_, _, eng := stagedDashboard(t, "medieval_age")
		acct, err := game.CreateAccount("Glyph Tester " + mode)
		must(t, err)
		eng.SetAccount(acct)
		must(t, acct.SetMapGlyphs(mode))
		d, pages := newTestDashboard(eng)
		drawnDashboard(t, d, pages, w, h) // refresh: a paused dashboard does not
		d.openPause()
		cells := drawnDashboard(t, d, pages, w, h)
		if !strings.Contains(strings.Join(rectText(cells, w, 0, 0, w, h), "\n"), "x10") {
			t.Fatalf("%s: no buy chips on the screen", mode)
		}
		x, y, bw, bh := d.economyTab.buildBox.GetRect()
		return cells, [4]int{x, y, bw, bh}
	}
	uni, r1 := draw("unicode")
	asc, r2 := draw("ascii")
	if r1 != r2 || r1[2] < 20 || r1[3] < 10 {
		t.Fatalf("the Buildings box is %v in unicode and %v in ascii", r1, r2)
	}
	for y := r1[1]; y < r1[1]+r1[3]; y++ {
		for x := r1[0]; x < r1[0]+r1[2]; x++ {
			a, b := uni[y*w+x], asc[y*w+x]
			if string(a.Runes) != string(b.Runes) || a.Style != b.Style {
				t.Fatalf("cell %d,%d of the Buildings box differs: %q %v in unicode, %q %v in ascii", x, y, string(a.Runes), a.Style, string(b.Runes), b.Style)
			}
		}
	}
}
