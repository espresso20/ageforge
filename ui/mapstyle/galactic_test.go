package mapstyle

import (
	"testing"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// galacticView draws the Galactic map once, at the zoom it opens on.
func galacticView(t *testing.T) (*view, Frame) {
	t.Helper()
	fixture.WithTheme(t, "forge")
	f := Frame{Model: modelFor(t, fixture.Options{Age: "galactic_age", Seed: 7})}
	v := New().(*view)
	drawFrame(v, f, 120, 40, false).Fini()
	return v, f
}

// inspectThing puts the cursor on the first cell of the picture that shows a
// thing of the kind and key, and says what the map says of it.
func inspectThing(t *testing.T, v *view, f Frame, kind mapgen.Kind, key string) Inspection {
	t.Helper()
	g := v.ov.Grid
	for i, h := range g.Hit {
		if th := g.Things[h]; th.Kind == kind && th.Key == key {
			v.cur = v.ov.Pick(i%g.W, i/g.W)
			in, _ := v.Inspect(f)
			return in
		}
	}
	t.Fatalf("no %s in the picture", key)
	return Inspection{}
}

// TestGalacticOpensOnTheNeighborhood: the map opens at the inner zoom, a star
// inspects as its name and its distance from Sol, Sol is the only star to
// dive into, and the close-up of Sol names Sol and the Earth.
func TestGalacticOpensOnTheNeighborhood(t *testing.T) {
	v, f := galacticView(t)
	if z := v.shownZoom(); z != zInner {
		t.Fatalf("opens at zoom %d, want %d", z, zInner)
	}
	g := v.ov.Grid
	for i, h := range g.Hit {
		if th := g.Things[h]; th.Key == "neighbor" && th.Ref == 1 {
			v.cur = v.ov.Pick(i%g.W, i/g.W)
			in, _ := v.Inspect(f)
			if in.Title != "Alpha Centauri" || len(in.Lines) == 0 || in.Lines[0] != "4.4 light-years from Sol" {
				t.Errorf("Alpha Centauri inspects as %+v", in)
			}
			v.zoomTo(v.scene(f), zCloseUp)
			if v.shownZoom() != zInner {
				t.Errorf("zoom in on a star went to zoom %d", v.shownZoom())
			}
			break
		}
	}
	if in := inspectThing(t, v, f, mapgen.KindTerrain, "sun"); in.Title != "Sol" {
		t.Errorf("the middle of the map inspects as %q, want Sol", in.Title)
	}
	v.zoomTo(v.scene(f), zCloseUp)
	if v.shownZoom() != zCloseUp {
		t.Fatalf("zoom in on Sol stays at zoom %d", v.shownZoom())
	}
	drawFrame(v, f, 120, 40, false).Fini()
	if in := inspectThing(t, v, f, mapgen.KindTerrain, "home"); in.Title != "Earth" {
		t.Errorf("the home world inspects as %q, want Earth", in.Title)
	}
	if in := inspectThing(t, v, f, mapgen.KindTerrain, "sun"); in.Title != "Sol" {
		t.Errorf("the Sun inspects as %q, want Sol", in.Title)
	}
}
