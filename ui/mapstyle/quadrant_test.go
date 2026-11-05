package mapstyle

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

const quadRunes = "▘▝▖▗▀▄▌▐▚▞▛▜▙▟█"

// TestInnerSystemIsQuadrantBlocks: the inner system at 120 by 40 is drawn in
// quadrant glyphs, with the Sun's warm colour in some cell.
func TestInnerSystemIsQuadrantBlocks(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "interstellar_age", Seed: 7})
	v := New().(*view)
	f := Frame{Model: m}
	v.zoomTo(v.scene(f), zInner)
	scr := drawFrame(v, f, 120, 40, false)
	defer scr.Fini()
	quads := 0
	for _, r := range capture.Text(scr) {
		if strings.ContainsRune(quadRunes, r) {
			quads++
		}
	}
	if quads < 100 {
		t.Errorf("%d quadrant glyphs on the inner system, want many", quads)
	}
	cells, w, h := scr.GetContents()
	warm := false
	for _, c := range cells[:w*h] {
		fg, bg, _ := c.Style.Decompose()
		for _, col := range [2]interface{ RGB() (int32, int32, int32) }{fg, bg} {
			r, g, b := col.RGB()
			if r > 220 && g > 110 && b < 140 && r > b+100 {
				warm = true
			}
		}
	}
	if !warm {
		t.Error("no Sun-coloured cell on the inner system")
	}
}

// TestInspectFindsAPlanetFromAnyOfItsCells: with the cursor over each cell of
// a planet in the overview, the inspect line names a planet.
func TestInspectFindsAPlanetFromAnyOfItsCells(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "interstellar_age", Seed: 7})
	for _, z := range []int{zCosmos, zOverview, zInner} {
		v := New().(*view)
		v.inspect = true
		f := Frame{Model: m}
		v.zoomTo(v.scene(f), z)
		scr := drawFrame(v, f, 120, 40, false)
		scr.Fini()
		g := v.ov.Grid
		seen, want := 0, ""
		for y := 0; y < g.H; y++ {
			for x := 0; x < g.W; x++ {
				if th := g.ThingAt(x, y); th.Kind == mapgen.KindTerrain && th.Key == "home" {
					v.cur = v.ov.Pick(x, y)
					in, ok := v.Inspect(f)
					if !ok || in.Title == "" {
						t.Fatalf("zoom %d: nothing to inspect at cell %d,%d of the home planet", z, x, y)
					}
					if want == "" {
						want = in.Title
					} else if in.Title != want {
						t.Errorf("zoom %d: cell %d,%d says %q, others say %q", z, x, y, in.Title, want)
					}
					seen++
				}
			}
		}
		if seen == 0 && z != zCosmos {
			t.Errorf("zoom %d: the home planet is not drawn", z)
		}
	}
}
