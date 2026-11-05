package mapstyle

import (
	"testing"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// TestWideViewportSpaceNeverDuplicatesEarth verifies that on a terminal wider than
// the world grid (e.g. W = 360, WorldW = 240), out-of-bounds coordinates never mirror
// Earth or duplicate unique celestial bodies. Instead, they seamlessly render procedural space.
func TestWideViewportSpaceNeverDuplicatesEarth(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := busyModel(t, "space_age")
	v := New().(*view)
	f := Frame{Model: m, Anim: 0, Tier: mapmodel.TierUnicode}
	sc := v.scene(f)
	if sc.SpaceCell == nil {
		t.Fatal("space_age scene must have SpaceCell func")
	}

	w, h := 360, 100
	x0 := -(w - mapgen.WorldW) / 2 // -60
	y0 := -(h - mapgen.WorldH) / 2 // -2

	win := windowFilled(sc, x0, y0, w, h)

	// Earth center in world coordinates is at (120, 48).
	// In the window, this corresponds to x = 120 - (-60) = 180, y = 48 - (-2) = 50.
	earthCountInBounds := 0
	earthCountOutOfBounds := 0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			wx, wy := x0+x, y0+y
			hit := win.Hit[y*w+x]
			th := win.Things[hit]

			inWorld := wx >= 0 && wx < sc.Grid.W && wy >= 0 && wy < sc.Grid.H
			if !inWorld {
				// Out of bounds cells must have Hit == 0 and not reference any world Thing
				if hit != 0 {
					t.Errorf("out-of-bounds cell at window (%d,%d) world (%d,%d) has Hit=%d (%+v), want 0", x, y, wx, wy, hit, th)
				}
				cell := win.Cells[y*w+x]
				// Earth rim rune quadrantRunes or home key should never be present out of bounds
				if th.Key == "home" {
					earthCountOutOfBounds++
				}
				// Glyph must not be an Earth surface rune or building
				if cell.R == '█' || cell.R == '▌' || cell.R == '▐' || cell.R == '▖' || cell.R == '▗' || cell.R == '▘' || cell.R == '▝' {
					t.Errorf("out-of-bounds cell at window (%d,%d) has planet rune %c", x, y, cell.R)
				}
			} else {
				if th.Key == "home" {
					earthCountInBounds++
				}
			}
		}
	}

	if earthCountInBounds == 0 {
		t.Error("expected Earth cells inside world bounds, found 0")
	}
	if earthCountOutOfBounds > 0 {
		t.Errorf("Earth was duplicated %d times in out-of-bounds margins!", earthCountOutOfBounds)
	}
}

// TestWideViewportGroundNeverDuplicatesHomeOrBuildings verifies that in ground view,
// unique landmarks ("home", buildings, wonders) are never mirrored into out-of-bounds margins.
func TestWideViewportGroundNeverDuplicatesHomeOrBuildings(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := busyModel(t, "medieval_age")
	v := New().(*view)
	f := Frame{Model: m, Anim: 0, Tier: mapmodel.TierUnicode}
	sc := v.scene(f)
	if sc.SpaceCell != nil {
		t.Fatal("ground scale should have nil SpaceCell")
	}

	w, h := 360, 110
	x0 := -(w - mapgen.WorldW) / 2
	y0 := -(h - mapgen.WorldH) / 2

	win := windowFilled(sc, x0, y0, w, h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			wx, wy := x0+x, y0+y
			inWorld := wx >= 0 && wx < sc.Grid.W && wy >= 0 && wy < sc.Grid.H
			if !inWorld {
				hit := win.Hit[y*w+x]
				th := win.Things[hit]
				if th.Key == "home" {
					t.Errorf("unique 'home' was duplicated at out-of-bounds (%d,%d)", wx, wy)
				}
				if th.Kind == mapgen.KindBuilding || th.Kind == mapgen.KindWonder {
					t.Errorf("building/wonder %+v was duplicated at out-of-bounds (%d,%d)", th, wx, wy)
				}
			}
		}
	}
}

// TestWideViewportSpaceNebulaAndStarsRendered verifies that out-of-bounds cells
// in space contain stars and nebula background variation.
func TestWideViewportSpaceNebulaAndStarsRendered(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := busyModel(t, "space_age")
	v := New().(*view)
	f := Frame{Model: m, Anim: 0, Tier: mapmodel.TierUnicode}
	sc := v.scene(f)

	w, h := 360, 100
	x0 := -(w - mapgen.WorldW) / 2
	y0 := -(h - mapgen.WorldH) / 2

	win := windowFilled(sc, x0, y0, w, h)

	starsCount := 0
	nebulaCount := 0
	outOfBoundsTotal := 0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			wx, wy := x0+x, y0+y
			if wx < 0 || wx >= sc.Grid.W || wy < 0 || wy >= sc.Grid.H {
				outOfBoundsTotal++
				cell := win.Cells[y*w+x]
				if cell.R != ' ' && cell.R != 0 {
					starsCount++
				}
				if cell.Bg.Hue == mapgen.HNebula {
					nebulaCount++
				}
			}
		}
	}

	if starsCount == 0 {
		t.Error("expected procedural stars in out-of-bounds margins, found 0")
	}
	if nebulaCount == 0 {
		t.Error("expected procedural nebula in out-of-bounds margins, found 0")
	}
}
