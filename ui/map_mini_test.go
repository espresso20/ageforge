package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
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
	if err := eng.SaveGame("minimap"); err != nil {
		t.Fatal(err)
	}
	if err := eng.LoadGame("minimap"); err != nil {
		t.Fatal(err)
	}
	if err := eng.MeetFactionForTest("riverlands_tribes", 30); err != nil {
		t.Fatal(err)
	}
	return eng
}

// drawDashboard lays the dashboard out at w x h and draws it (through the
// theme screen wrapper, as the app does), returning the screen text.
func drawDashboard(t testing.TB, d *Dashboard, pages *tview.Pages, w, h int) string {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	scr := theme.WrapScreen(sim)
	pages.SetRect(0, 0, w, h)
	pages.Draw(scr)
	scr.Show()
	return simText(sim)
}

func newMiniDashboard(t testing.TB) (*Dashboard, *tview.Pages) {
	t.Helper()
	eng := miniTestGame(t)
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	miniOn(t, d)
	return d, pages
}

// miniOn turns the dashboard's mini map on, as "minimap on" does. It is
// off until the player asks for it.
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
// news line, and hides at 80x24 and 100x30, where the Buildings list gets
// the whole column back.
func TestMiniMapFitsAndHides(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	d, pages := newMiniDashboard(t)
	for _, sz := range []struct {
		w, h int
		show bool
	}{{160, 48, true}, {120, 40, true}, {100, 30, false}, {80, 24, false}} {
		txt := drawDashboard(t, d, pages, sz.w, sz.h)
		d.refresh()
		txt = drawDashboard(t, d, pages, sz.w, sz.h)
		if d.mapDock.shown != sz.show {
			t.Errorf("%dx%d: mini map shown = %v, want %v\n%s", sz.w, sz.h, d.mapDock.shown, sz.show, txt)
		}
		if strings.Contains(txt, " Map · ") != sz.show {
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

// TestMiniMapSize: the mini map is short enough to leave the Buildings list
// most of the column: at most 9 rows inside its border and about a quarter
// of the column, hidden where it would leave the list under dockMinBody
// rows. Adam found the old 17-row mini map crowding Buildings off a 160x48
// screen.
func TestMiniMapSize(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	d, pages := newMiniDashboard(t)
	for _, sz := range [][2]int{{160, 48}, {120, 40}, {200, 60}} {
		drawDashboard(t, d, pages, sz[0], sz[1])
		d.refresh()
		drawDashboard(t, d, pages, sz[0], sz[1])
		_, _, _, dh := d.mapDock.GetRect()
		_, _, _, mh := d.miniMap.GetRect()
		if !d.mapDock.shown {
			t.Fatalf("%dx%d: no mini map", sz[0], sz[1])
		}
		if mh > miniMaxInnerH+2 || (mh-2)*4 > dh-2 || dh-mh < dockMinBody {
			t.Errorf("%dx%d: mini map %d rows of a %d-row column", sz[0], sz[1], mh, dh)
		}
	}
	for _, c := range []struct{ w, h, want int }{{69, 39, 11}, {49, 32, 9}, {49, 26, 8}, {49, 25, 0}, {41, 39, 0}} {
		if got := miniHeight(c.w, c.h); got != c.want {
			t.Errorf("miniHeight(%d, %d) = %d, want %d", c.w, c.h, got, c.want)
		}
	}
}

// TestMinimapSetting: the mini map is off until the player turns it on.
// minimap on shows it above the Buildings list and is saved to the account;
// minimap off gives the list the whole column back.
func TestMinimapSetting(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	pages := tview.NewPages()
	pages.AddPage("dashboard", d.Root(), true, true)
	show := func() bool {
		drawDashboard(t, d, pages, 160, 48)
		d.refresh()
		txt := drawDashboard(t, d, pages, 160, 48)
		if strings.Contains(txt, " Map · ") != d.mapDock.shown {
			t.Errorf("the map title on screen disagrees with shown=%v", d.mapDock.shown)
		}
		return d.mapDock.shown
	}
	// Off until asked for: the Buildings list has the whole column, and no
	// model is built for a map nobody sees.
	if show() {
		t.Fatal("the mini map shows on a dashboard nobody turned it on for")
	}
	if eng.Account().MinimapOn() || d.mapDock.wantsModel() {
		t.Error("a new account has the mini map on, or builds models for it")
	}
	if res := HandleCommand("minimap", eng); !strings.Contains(res.Message, "Mini map: off") {
		t.Errorf("bare minimap: %+v", res)
	}
	// On, and remembered by the account.
	d.runForTest("minimap on")
	if !show() || !eng.Account().MinimapOn() {
		t.Fatal("minimap on did not show the mini map, or was not saved to the account")
	}
	if res := HandleCommand("minimap", eng); !strings.Contains(res.Message, "Mini map: on") {
		t.Errorf("bare minimap: %+v", res)
	}
	d.runForTest("minimap off")
	if show() {
		t.Error("minimap off: the mini map still shows")
	}
	if eng.Account().MinimapOn() {
		t.Error("minimap off was not saved to the account")
	}
	if d.mapDock.wantsModel() {
		t.Error("the mini map still builds models while off")
	}
	d.runForTest("minimap on")
	if !show() || !eng.Account().MinimapOn() {
		t.Error("minimap on did not bring it back")
	}
	if res := HandleCommand("minimap sideways", eng); res.Type != "error" {
		t.Errorf("minimap sideways: %+v", res)
	}
}

// TestMiniMapFollowsSettings: map style switches the mini map too.
func TestMiniMapFollowsSettings(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	miniOn(t, d)
	d.runForTest("map style skyline")
	d.refresh()
	if d.miniMap.set.Style != "skyline" {
		t.Errorf("mini map style %s after map style skyline", d.miniMap.set.Style)
	}
}

// TestMiniMapFrameTime: a dashboard frame with the mini map stays cheap.
// Budget: 25 ms a frame on average (a CI runner under -race is slower
// than a laptop; the typical cost is a few ms, see BenchmarkDashboardFrame).
func TestMiniMapFrameTime(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	d, pages := newMiniDashboard(t)
	drawDashboard(t, d, pages, 160, 48)
	const n = 20
	start := time.Now()
	for i := 0; i < n; i++ {
		d.refresh()
		drawDashboard(t, d, pages, 160, 48)
	}
	per := time.Since(start) / n
	t.Logf("dashboard refresh+frame at 160x48 with the mini map: %v (mini map draw %v)", per, d.miniMap.lastDraw)
	if per > 250*time.Millisecond {
		t.Errorf("a dashboard frame takes %v", per)
	}
}

// BenchmarkMiniMapDraw is the mini map's own draw (what a keypress redraw
// adds to the dashboard), per style.
func BenchmarkMiniMapDraw(b *testing.B) {
	b.Cleanup(game.SetDataDirForTest(b.TempDir()))
	d, _ := newMiniDashboard(b)
	sim := tcell.NewSimulationScreen("UTF-8")
	_ = sim.Init()
	defer sim.Fini()
	sim.SetSize(160, 48)
	for _, style := range d.mapViews.reg.Names() {
		b.Run(style, func(b *testing.B) {
			s := d.mapSettings()
			s.Style = style
			st := d.engine.GetState()
			d.miniMap.update(s, &st)
			d.miniMap.SetRect(0, 0, 69, miniMaxInnerH+2)
			for i := 0; i < b.N; i++ {
				d.miniMap.Draw(sim)
			}
		})
	}
}

// TestDumpDashboardWithMiniMap writes text captures of the dashboard with
// the mini map, per style. Opt-in:
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
	for _, style := range d.mapViews.reg.Names() {
		d.mapLocal = nil
		s := d.mapSettings()
		s.Style = style
		d.mapLocal = &s
		for _, sz := range [][2]int{{160, 48}, {120, 40}, {80, 24}} {
			drawDashboard(t, d, pages, sz[0], sz[1])
			d.refresh()
			txt := drawDashboard(t, d, pages, sz[0], sz[1])
			name := filepath.Join(dir, fmt.Sprintf("dashboard_%s_%dx%d.txt", style, sz[0], sz[1]))
			if err := os.WriteFile(name, []byte(txt), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}
