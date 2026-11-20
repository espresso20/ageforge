package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// miniTestGame is a game with something to show: a few buildings, and a
// civ met since the save was loaded, so the recap has news.
func miniTestGame(t testing.TB) *game.GameEngine {
	t.Helper()
	eng := game.NewGameEngine()
	for i := 0; i < 12; i++ {
		HandleCommand("gather wood 10", eng)
		HandleCommand("gather food 10", eng)
		HandleCommand("gather stone 10", eng)
	}
	HandleCommand("build hut 2", eng)
	HandleCommand("build stash", eng)
	must(t, eng.SaveGame("minimap"))
	must(t, eng.LoadGame("minimap"))
	must(t, eng.MeetFactionForTest("riverlands_tribes", 30))
	return eng
}

// drawDashboard lays the dashboard out at w x h and draws it (through the
// theme screen wrapper, as the app does), returning the screen text.
func drawDashboard(t testing.TB, d *Dashboard, pages *nativePages, w, h int) string {
	t.Helper()
	sim := newSimScreen(t, w, h)
	scr := sim
	pages.SetRect(0, 0, w, h)
	drawOn(scr, pages)
	scr.Show()
	return simCells(sim).Text()
}

func newMiniDashboard(t testing.TB) (*Dashboard, *nativePages) {
	t.Helper()
	d, pages := newTestDashboard(miniTestGame(t))
	miniOn(t, d)
	return d, pages
}

// miniOn turns the dashboard's mini map on, as "minimap on" does. It is on
// unless the player turned it off, so this pins the setting.
func miniOn(t testing.TB, d *Dashboard) {
	t.Helper()
	d.inputField.SetText("minimap on")
	d.submitInput()
	if !d.mapSettings().Minimap {
		t.Fatal("minimap on did not turn the mini map on")
	}
	d.refresh()
}

// TestMiniMapFitsAndHides: the mini map shows on a roomy terminal, with the
// news line, and hides where even the Panels list's shortest form leaves it under 8
// picture rows (100x30, 80x24).
func TestMiniMapFitsAndHides(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	d, pages := newMiniDashboard(t)
	for _, sz := range []struct {
		w, h int
		show bool
	}{{200, 60, true}, {153, 41, true}, {120, 36, true}, {80, 36, true}, {100, 30, false}, {80, 24, false}} {
		txt := drawDashboard(t, d, pages, sz.w, sz.h)
		d.refresh()
		txt = drawDashboard(t, d, pages, sz.w, sz.h)
		if d.body.miniShown != sz.show {
			t.Errorf("%dx%d: mini map shown = %v, want %v\n%s", sz.w, sz.h, d.body.miniShown, sz.show, txt)
		}
		if strings.Contains(txt, "┌ Map ─") != sz.show {
			t.Errorf("%dx%d: map title on screen = %v, want %v", sz.w, sz.h, !sz.show, sz.show)
		}
		if !strings.Contains(txt, "Buildings") {
			t.Errorf("%dx%d: the Buildings list is gone\n%s", sz.w, sz.h, txt)
		}
		if sz.show && !strings.Contains(txt, "Riverlands") {
			t.Errorf("%dx%d: no since-last-visit news on the mini map\n%s", sz.w, sz.h, txt)
		}
	}
}

// TestMiniMapIsASquareUnderPanels: the mini map is in the right column under
// the Panels list, about square on screen (columns to rows 2.1 to 1), and the
// Buildings list keeps the whole middle column. On a small window the mini
// map is not shown and the Panels list is whole; minimap off hides it.
func TestMiniMapIsASquareUnderPanels(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	d, pages := newMiniDashboard(t)
	draw := func(w, h int) string {
		drawDashboard(t, d, pages, w, h)
		d.refresh()
		return drawDashboard(t, d, pages, w, h)
	}
	for _, sz := range [][2]int{{275, 54}, {200, 60}} {
		draw(sz[0], sz[1])
		if !d.body.miniShown {
			t.Fatalf("%dx%d: no mini map", sz[0], sz[1])
		}
		px, py, pw, ph := d.sidebar.GetRect()
		mx, my, mw, mh := d.miniMap.GetRect()
		bx, by, bw, bh := d.economyTab.buildBox.GetRect()
		if mx != px || mw != pw || my != py+ph {
			t.Errorf("%v: mini map %d,%d %dx%d is not under the Panels list %d,%d %dx%d", sz, mx, my, mw, mh, px, py, pw, ph)
		}
		if inner, want := mh-2, float64(mw)/2.1; float64(inner) < want-1 || float64(inner) > want+1 {
			t.Errorf("%v: mini map picture is %d rows for %d columns, want about %.1f", sz, inner, mw, want)
		}
		if by != py || bx+bw > mx || bh != ph+mh {
			t.Errorf("%v: Buildings at %d,%d %dx%d is not the whole middle column (Panels %d,%d, %d rows with the map)", sz, bx, by, bw, bh, px, py, ph+mh)
		}
	}
	// 153x41 and 120x37, the default windows of a 1080p monitor and a
	// 14-inch laptop: the map keeps its square (the list goes to two
	// columns for it), and every panel is listed.
	for _, sz := range [][2]int{{153, 41}, {120, 37}} {
		txt := draw(sz[0], sz[1])
		_, _, mw, mh := d.miniMap.GetRect()
		if want := float64(mw) / 2.1; !d.body.miniShown || float64(mh-2) < want-1.5 || strings.Contains(txt, "PROGRESS") {
			t.Errorf("%v: shown = %v with %d picture rows for %d columns; want a square map\n%s", sz, d.body.miniShown, mh-2, mw, txt)
		}
		for _, name := range sidebarPanels {
			if !strings.Contains(txt, name) {
				t.Errorf("%v: the Panels list has lost %q", sz, name)
			}
		}
	}
	// 192x52: the list keeps its own form and the box is shorter than
	// square, with at least ten picture rows (case B). (The header is two
	// rows shorter than the top it replaced on a tall window, so the rows
	// the case had at 51 are these at 49.)
	draw(192, 52)
	if _, _, _, mh := d.miniMap.GetRect(); !d.body.miniShown || mh-2 < 10 || mh >= 18 {
		t.Errorf("192x52: mini map shown = %v, %d rows; want a box under square with 10 picture rows or more", d.body.miniShown, mh)
	}
	if txt := draw(192, 52); !strings.Contains(txt, "PROGRESS") {
		t.Errorf("192x52: the Panels list lost its grouped form\n%s", txt)
	}
	// 160x42: the list steps down a form (a name a row, no groups) so the
	// map can show, and every panel is still listed (case D).
	txt := draw(160, 42)
	if !d.body.miniShown || strings.Contains(txt, "PROGRESS") {
		t.Errorf("160x42: shown = %v, or the list kept its grouped form\n%s", d.body.miniShown, txt)
	}
	for _, name := range sidebarPanels {
		if !strings.Contains(txt, name) {
			t.Errorf("160x42: the Panels list has lost %q\n%s", name, txt)
		}
	}
	// With the map off, the list gets its own form back.
	d.runForTest("minimap off")
	if txt := draw(160, 42); d.body.miniShown || !strings.Contains(txt, "PROGRESS") {
		t.Errorf("160x42 with minimap off: shown = %v, or the list is not in its grouped form\n%s", d.body.miniShown, txt)
	}
	d.runForTest("minimap on")
	// Too small: not shown, and the Panels list is whole.
	txt = draw(100, 30)
	if d.body.miniShown || strings.Contains(txt, "┌ Map ─") {
		t.Errorf("100x30: the mini map shows\n%s", txt)
	}
	for _, name := range sidebarPanels {
		if !strings.Contains(txt, name) {
			t.Errorf("100x30: the Panels list has lost %q\n%s", name, txt)
		}
	}
	// Back to a big window: it comes back.
	if draw(200, 60); !d.body.miniShown {
		t.Error("200x60 after 100x30: the mini map did not come back")
	}
	// Off.
	d.runForTest("minimap off")
	if txt := draw(200, 60); d.body.miniShown || strings.Contains(txt, "┌ Map ─") {
		t.Errorf("minimap off: the mini map still shows at 200x60")
	}
}

// TestMinimapSetting: the mini map is on until the player turns it off.
// minimap off gives the Buildings list the whole column back and is saved to
// the account; minimap on brings it back.
func TestMinimapSetting(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	pages := newNativePages()
	pages.AddPage("dashboard", d.Root(), true, true)
	show := func() bool {
		drawDashboard(t, d, pages, 200, 60)
		d.refresh()
		txt := drawDashboard(t, d, pages, 200, 60)
		if strings.Contains(txt, "┌ Map ─") != d.body.miniShown {
			t.Errorf("the map title on screen disagrees with shown=%v", d.body.miniShown)
		}
		return d.body.miniShown
	}
	// On until turned off, with nothing chosen.
	if !show() {
		t.Fatal("the mini map does not show on a new account")
	}
	if !eng.Account().MinimapOn() {
		t.Error("a new account has the mini map off")
	}
	if res := HandleCommand("minimap", eng); !strings.Contains(res.Message, "Mini map: on") {
		t.Errorf("bare minimap: %+v", res)
	}
	// Off, and remembered by the account: the Buildings list has the whole
	// column, and no model is built for a map nobody sees.
	d.runForTest("minimap off")
	if show() {
		t.Error("minimap off: the mini map still shows")
	}
	if eng.Account().MinimapOn() {
		t.Error("minimap off was not saved to the account")
	}
	if d.body.wantsModel() {
		t.Error("the mini map still builds models while off")
	}
	if res := HandleCommand("minimap", eng); !strings.Contains(res.Message, "Mini map: off") {
		t.Errorf("bare minimap: %+v", res)
	}
	d.runForTest("minimap on")
	if !show() || !eng.Account().MinimapOn() {
		t.Error("minimap on did not bring it back")
	}
	if res := HandleCommand("minimap sideways", eng); res.Type != "error" {
		t.Errorf("minimap sideways: %+v", res)
	}
}

// TestMiniMapFollowsSettings: map glyphs switches the mini map too.
func TestMiniMapFollowsSettings(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	d.runForTest("map glyphs ascii")
	d.refresh()
	if d.miniMap.set.Tier != mapmodel.TierASCII {
		t.Errorf("mini map glyphs %s after map glyphs ascii", d.miniMap.set.Tier)
	}
}

// BenchmarkMiniMapDraw is the mini map's own draw (what a keypress redraw
// adds to the dashboard).
func BenchmarkMiniMapDraw(b *testing.B) {
	b.Cleanup(game.SetDataDirForTest(b.TempDir()))
	d, _ := newMiniDashboard(b)
	sim := newSimScreen(b, 160, 48)
	st := d.engine.GetState()
	d.miniMap.update(d.mapSettings(), &st)
	d.miniMap.SetRect(0, 0, 69, 18)
	for i := 0; i < b.N; i++ {
		drawOn(sim, d.miniMap)
	}
}

// TestDumpDashboardWithMiniMap writes text captures of the dashboard with
// the mini map. Opt-in:
//
//	MINIMAP_DUMP=/tmp/dump go test ./ui/ -run TestDumpDashboardWithMiniMap
func TestDumpDashboardWithMiniMap(t *testing.T) {
	dir := os.Getenv("MINIMAP_DUMP")
	if dir == "" {
		t.Skip("set MINIMAP_DUMP=<dir> to dump dashboard captures")
	}
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	d, pages := newMiniDashboard(t)
	for _, sz := range [][2]int{{160, 48}, {120, 40}, {80, 24}} {
		drawDashboard(t, d, pages, sz[0], sz[1])
		d.refresh()
		txt := drawDashboard(t, d, pages, sz[0], sz[1])
		name := filepath.Join(dir, fmt.Sprintf("dashboard_%dx%d.txt", sz[0], sz[1]))
		must(t, os.WriteFile(name, []byte(txt+"\n"), 0o644))
	}
}

// TestMiniMapNotShownBeforeFirstRefresh: a dashboard that has been drawn but
// not yet refreshed (the frame between showing it and the first tick) must
// not show the mini map: it has no map to draw until a refresh builds one.
func TestMiniMapNotShownBeforeFirstRefresh(t *testing.T) {
	d, pages := newTestDashboard(nil)
	txt := drawDashboard(t, d, pages, 180, 56)
	if d.body.miniShown || strings.Contains(txt, "┌ Map ─") {
		t.Fatalf("the mini map shows before the first refresh:\n%s", txt)
	}
}
