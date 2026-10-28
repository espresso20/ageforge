package ui

import (
	"slices"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
)

// TestTheVisitorOnTheMapEarnsItsBadge: with the real map in the Map panel, a
// player who steers the cursor onto the rare visitor with the arrow keys
// earns the secret badge, with the motion setting on or off. No stand-in
// stands in for the map: this is the path the badge is meant to be reached by.
func TestTheVisitorOnTheMapEarnsItsBadge(t *testing.T) {
	const badge = "special.we_are_not_alone"
	for _, motion := range []bool{true, false} {
		d, eng := mapTestDashboard(t, true)
		must(t, eng.StartNewNamedGame("run"))
		must(t, eng.Account().SetMotion(motion))
		p := d.mapPanel
		p.set = d.mapSettings()
		p.update(eng.GetState())
		drawMapPanel(t, d, 160, 48)

		// the world's clock stands at the middle of the first visit
		sg := mapmodel.NextSighting(p.model.Seed, false, 0)
		mid := sg.Start + sg.Frames/2
		pinned := p.start
		p.now = func() time.Time { return pinned.Add(time.Duration(mid) * mapAnimStep) }
		p.update(eng.GetState())
		if slices.Contains(eng.Account().EarnedBadges(), badge) {
			t.Fatal("the badge was earned before the cursor moved")
		}

		// where the visitor stands, from the engine itself
		f := p.frame()
		var gen mapgen.Generator
		g := gen.DrawAt(p.model, mapgen.Moment{Anim: f.Anim, Visit: f.VisitFrame()}).Grid
		var target mapgen.Pt
		for i := range g.Cells {
			if g.Things[g.Hit[i]].Kind == mapgen.KindVisitor {
				target = mapgen.Pt{X: i % g.W, Y: i / g.W}
			}
		}
		if target == (mapgen.Pt{}) {
			t.Fatalf("motion %v: no visitor in the middle of the first visit", motion)
		}

		// the arrows, as a player presses them
		for i := 0; i < 400 && !slices.Contains(eng.Account().EarnedBadges(), badge); i++ {
			at, _ := p.view.Pointer()
			key, along := tcell.KeyRight, target.X-at.X
			switch {
			case target.X < at.X:
				key = tcell.KeyLeft
			case target.X == at.X && target.Y < at.Y:
				key, along = tcell.KeyUp, target.Y-at.Y
			case target.X == at.X:
				key, along = tcell.KeyDown, target.Y-at.Y
			}
			mod := tcell.ModNone
			if along >= 8 || along <= -8 {
				mod = tcell.ModShift
			}
			p.routeKey(tcell.NewEventKey(key, 0, mod), "")
		}
		if !slices.Contains(eng.Account().EarnedBadges(), badge) {
			at, _ := p.view.Pointer()
			t.Fatalf("motion %v: the cursor at %v never earned %s on a visitor at %v", motion, at, badge, target)
		}
	}
}
