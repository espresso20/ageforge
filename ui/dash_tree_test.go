package ui

import (
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
	"github.com/rivo/tview"
)

// drawOnWindow draws the pages on a screen that knows which kit host drew
// each cell, as the window's does.
func drawOnWindow(t *testing.T, d *Dashboard, pages *nativePages, w, h int) *nativeScreen {
	t.Helper()
	scr := newNativeScreen(theme.WrapScreen(newSimScreen(t, w, h)))
	for i := 0; i < 2; i++ {
		d.refresh()
		scr.Clear()
		pages.SetRect(0, 0, w, h)
		pages.Draw(scr)
		scr.Show()
	}
	return scr
}

// The whole screen is one kit tree: the dashboard's boxes, the mini map
// among them, are all painted by the one host, the page stack's.
func TestTheDashboardIsOneKitTree(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	pages.SwitchToPage("dashboard")
	scr := drawOnWindow(t, d, pages, 150, 48)
	boxes := map[string]interface{ GetRect() (int, int, int, int) }{
		"the status ribbon": d.statusTV, "the Ledger": d.ledgerTV, "the command bar": d.inputField,
		"the log": d.logTV, "the Panels list": d.sidebar,
		"the Resources box": d.economyTab.resourceTV, "the Buildings list": d.economyTab.buildingTV,
		"Under construction": d.economyTab.constructionTV,
	}
	for name, box := range boxes {
		x, y, w, h := box.GetRect()
		if w <= 0 || h <= 0 {
			t.Errorf("%s has no room on a 150x48 screen", name)
			continue
		}
		if got := scr.hostAt(x+w/2, y+h/2); got != pages.nativeHost {
			t.Errorf("%s is not painted by the page stack's host (%p, want %p)", name, got, pages.nativeHost)
		}
	}
	if !d.mapDock.shown {
		t.Fatal("staging: the mini map is not showing at 150x48")
	}
	x, y, w, h := d.miniMap.GetRect()
	if scr.hostAt(x+w/2, y+h/2) != pages.nativeHost {
		t.Error("the mini map is not painted by the page stack's host")
	}
}

// The Map panel, open over the dashboard, is painted by its page's host: the
// map's cells reach the screen through the kit.
func TestTheMapPanelIsPaintedByTheKit(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	pages.SwitchToPage("dashboard")
	d.runForTest("map")
	if d.overlayMgr.ActiveName() != "map" {
		t.Fatalf("the Map panel did not open (%q is open)", d.overlayMgr.ActiveName())
	}
	scr := drawOnWindow(t, d, pages, 150, 48)
	x, y, w, h := d.mapPanel.GetRect()
	if w < 100 || h < 30 {
		t.Fatalf("the Map panel is %dx%d on a 150x48 screen", w, h)
	}
	if host := scr.hostAt(x+w/2, y+h/2); host != pages.nativeHost {
		t.Errorf("the middle of the map belongs to %p, not the page stack's host", host)
	}
	if scr.hostAt(x+1, y+h-1) != pages.nativeHost {
		t.Error("the map's key bar is not painted with the map")
	}
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
	scr := drawOnWindow(t, d, pages, 150, 48)
	x, y, _, _ := d.inputField.GetInnerRect()
	at := x + kit.Width(d.inputField.GetLabel()) + 1 // on the "u" of "build"
	scr.mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at, Y: y})
	scr.mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: at, Y: y})
	d.inputField.InputHandler()(tev(kit.KeyRune, 'X', 0), func(Primitive) {})
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
	scr := drawOnWindow(t, d, pages, 150, 48)
	p := d.advancePanel
	px, py, _, _ := p.GetRect()
	for _, hit := range p.grid.hits {
		if hit.id != hitPillar || hit.n != 1 {
			continue
		}
		at := hit.r.Min.Add(p.Bounds().Min)
		// The panel's Bounds are in the page's cells; the page is at the
		// screen's corner, so they are the screen's too.
		if bx, by := p.Bounds().Min.X, p.Bounds().Min.Y; bx != px || by != py {
			t.Fatalf("the panel is at %d,%d in its page but %d,%d on the screen", bx, by, px, py)
		}
		scr.mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at.X + 1, Y: at.Y})
		if p.ui.sel != 1 {
			t.Errorf("a click on the second pillar picked %d", p.ui.sel)
		}
		return
	}
	t.Fatal("staging: no second pillar on the screen")
}
