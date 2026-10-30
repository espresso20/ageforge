package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// TestMenuTownFillsAnyScreen: the town behind the menu reaches the right
// edge of the land at any terminal size, wide ones included.
func TestMenuTownFillsAnyScreen(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {200, 60}, {420, 110}} {
		v, town := stagedMenuView("primitive", false, 0)
		g, sc := menuPage(v, town, sz[0], sz[1], 40)
		L := sc.L
		if L.forge || L.compact {
			continue
		}
		for x := 0; x < L.landW; x++ {
			if x >= L.px && x < L.px+L.pw {
				continue // under the menu box
			}
			live := false
			for y := L.top; y < L.h && !live; y++ {
				c := g.c[y*g.w+x]
				live = (c.r != ' ' && c.r != 0) || c.hasBg
			}
			if !live {
				t.Fatalf("%dx%d: column %d of the land (width %d) is blank", sz[0], sz[1], x, L.landW)
			}
		}
	}
}

// TestDashboardMapFillsWideScreen: the map panel of the game paints its
// whole picture area on a wide terminal too, not just the world's width.
func TestDashboardMapFillsWideScreen(t *testing.T) {
	_, town := stagedMenuView("primitive", false, 0)
	for _, sz := range [][2]int{{80, 24}, {420, 110}} {
		w, h := sz[0], sz[1]
		sim := tcell.NewSimulationScreen("UTF-8")
		if err := sim.Init(); err != nil {
			t.Fatal(err)
		}
		sim.SetSize(w, h)
		st := mapstyle.New()
		st.Draw(theme.WrapScreen(sim), mapstyle.Rect{W: w, H: h}, mapstyle.Frame{Model: town.model, Tier: mapmodel.TierUnicode})
		picW := w - 40 // the legend sits at the right
		for x := 0; x < picW; x++ {
			live := false
			for y := 1; y < h-5 && !live; y++ {
				r, _, style, _ := sim.GetContent(x, y)
				_, bg, _ := style.Decompose()
				live = (r != ' ' && r != 0) || bg != tcell.ColorDefault
			}
			if !live {
				t.Fatalf("%dx%d: map column %d is blank", w, h, x)
			}
		}
		sim.Fini()
	}
}
