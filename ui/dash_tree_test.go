package ui

import (
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// stageOf puts the pages on a w by h screen as the window shows them: laid
// out and painted, refreshed, and painted again, since the first paint gives
// every box its size and the refresh after it writes what depends on that.
func stageOf(d *Dashboard, pages *nativePages, w, h int) *stage {
	st := newStage(pages, w, h)
	for i := 0; i < 2; i++ {
		d.refresh()
		st.paint()
	}
	return st
}

// A panel leaves the command bar's rows to the dashboard under it, for the
// mouse as for the eye: with the Advance panel open, a click in the command
// bar still puts the cursor where it lands.
func TestTheCommandBarTakesAClickUnderAPanel(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	pages.SwitchToPage("dashboard")
	d.runForTest("charter")
	d.inputField.SetText("build hut")
	st := stageOf(d, pages, 200, 60)
	x, y, _, _ := d.inputField.GetInnerRect()
	st.click(x+kit.Width(d.inputField.GetLabel())+1, y) // on the "u" of "build"
	keyTo(d.inputField, tk(kit.KeyRune, 'X', 0))
	if got := d.inputField.GetText(); got != "bXuild hut" {
		t.Errorf("after a click on the second letter and an X the bar reads %q", got)
	}
	if d.overlayMgr.ActiveName() != "charter" {
		t.Errorf("the click closed the panel (%q is open)", d.overlayMgr.ActiveName())
	}
}

// A panel open over the dashboard takes the mouse through its page: a click
// on a pillar of the Advance panel, at its place on the screen, picks it.
func TestAPanelOverTheDashboardTakesTheMouse(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	pages.SwitchToPage("dashboard")
	d.runForTest("charter")
	if d.overlayMgr.ActiveName() != "charter" {
		t.Fatalf("the Advance panel did not open (%q is open)", d.overlayMgr.ActiveName())
	}
	st := stageOf(d, pages, 200, 60)
	p := d.advancePanel
	for _, hit := range p.grid.hits {
		if hit.id != hitPillar || hit.n != 1 {
			continue
		}
		at := hit.r.Min.Add(p.Bounds().Min)
		st.mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at.X + 1, Y: at.Y})
		if p.ui.sel != 1 {
			t.Errorf("a click on the second pillar picked %d", p.ui.sel)
		}
		return
	}
	t.Fatal("staging: no second pillar on the screen")
}

// The mini map and the Map panel are part of the page's one tree: they are
// painted with everything else, at the place the layout gave them.
func TestTheMapsArePaintedWithThePage(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	pages.SwitchToPage("dashboard")
	st := stageOf(d, pages, 200, 60)
	if !d.body.miniShown {
		t.Fatal("staging: the mini map is not showing at 200x60")
	}
	x, y, w, h := d.miniMap.GetRect()
	if c := st.buf.At(x, y); c.R != '┌' || w < 3 || h < 3 {
		t.Errorf("the mini map's corner at %d,%d holds %q", x, y, c.R)
	}
	d.runForTest("map")
	st = stageOf(d, pages, 200, 60)
	x, y, w, h = d.mapPanel.GetRect()
	if w < 100 || h < 30 {
		t.Fatalf("the Map panel is %dx%d on a 200x60 screen", w, h)
	}
	blank := 0
	for cx := x; cx < x+w; cx++ {
		if st.buf.At(cx, y+h/2).R == ' ' {
			blank++
		}
	}
	if blank == w {
		t.Error("the middle row of the Map panel is empty: the map was not painted")
	}
}
