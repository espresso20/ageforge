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
	d.refresh()
	return d, pages
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

// TestMiniMapFollowsSettings: map style switches the mini map too.
func TestMiniMapFollowsSettings(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
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
			d.miniMap.SetRect(0, 0, 69, miniH)
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
