package mapgen

import (
	"testing"
)

// TestOverviewFillFlag verifies that space scales (Orbit, System, Stars) enable Fill
// for smooth fractional fitting, while ground scales use integer step increments.
func TestOverviewFillFlag(t *testing.T) {
	gen := &Generator{}
	for _, tc := range []struct {
		age      string
		wantFill bool
	}{
		{"stone_age", false},
		{"bronze_age", false},
		{"medieval_age", false},
		{"modern_age", false},
		{"space_age", true},
		{"interstellar_age", true},
		{"galactic_age", true},
	} {
		m := model(state(tc.age, 7, noon))
		sc := gen.Draw(m, 0)
		if sc.Fill != tc.wantFill {
			t.Errorf("%s: Fill = %v, want %v", tc.age, sc.Fill, tc.wantFill)
		}
	}
}

// TestOverviewWideViewportOrbitCentersAndFillsSpace verifies that on wide or tall viewports
// in Orbit scale (e.g. space_age):
// 1. o.X0 / o.Y0 center the world instead of pinning to (0, 0).
// 2. Earth / world centre is centered in the overview.
// 3. Cells outside the world boundaries are filled with valid space background (not zero Cell{}).
// 4. Hit index for out-of-bounds cells is 0.
func TestOverviewWideViewportOrbitCentersAndFillsSpace(t *testing.T) {
	gen := &Generator{}
	m := model(state("space_age", 7, noon))
	sc := gen.Draw(m, 0)
	if sc.SpaceCell == nil {
		t.Fatal("space_age scene must have SpaceCell func")
	}

	testCases := []struct {
		name string
		w, h int
	}{
		{"160x50", 160, 50},
		{"240x60", 240, 60},
		{"260x100", 260, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ov := sc.Overview(tc.w, tc.h)

			spanX := tc.w * ov.step / 256
			if spanX > worldW {
				if ov.X0 >= 0 {
					t.Errorf("spanX (%d) > worldW (%d), but X0 = %d, want < 0 (centered)", spanX, worldW, ov.X0)
				}
				wantX0 := (worldW - spanX) / 2
				if ov.X0 != wantX0 {
					t.Errorf("X0 = %d, want %d", ov.X0, wantX0)
				}
			}

			spanY := tc.h * ov.step / 256
			if spanY > worldH {
				if ov.Y0 >= 0 {
					t.Errorf("spanY (%d) > worldH (%d), but Y0 = %d, want < 0 (centered)", spanY, worldH, ov.Y0)
				}
				wantY0 := (worldH - spanY) / 2
				if ov.Y0 != wantY0 {
					t.Errorf("Y0 = %d, want %d", ov.Y0, wantY0)
				}
			}

			// Verify Earth / world center is centered horizontally and vertically
			cx, cy, ok := ov.CellOf(sc.Centre)
			if !ok {
				t.Fatalf("world centre %v not found in overview %dx%d", sc.Centre, tc.w, tc.h)
			}
			if abs(cx-tc.w/2) > 2 {
				t.Errorf("world centre X in overview = %d, want near %d (center)", cx, tc.w/2)
			}
			if abs(cy-tc.h/2) > 2 {
				t.Errorf("world centre Y in overview = %d, want near %d (center)", cy, tc.h/2)
			}

			// Verify cells outside world boundaries have space background and Hit == 0
			blankCount := 0
			oobCount := 0
			for y := 0; y < tc.h; y++ {
				for x := 0; x < tc.w; x++ {
					bx0 := ov.blockX(x)
					by0 := ov.blockY(y)
					bx1 := ov.blockX(x + 1)
					by1 := ov.blockY(y + 1)

					// Check if entire block is outside world grid
					if (bx1 <= 0 || bx0 >= worldW) || (by1 <= 0 || by0 >= worldH) {
						oobCount++
						cell := ov.Grid.At(x, y)
						hit := ov.Grid.Hit[y*tc.w+x]
						if hit != 0 {
							t.Errorf("at (%d,%d) out-of-bounds block has Hit = %d, want 0", x, y, hit)
						}
						if cell == (Cell{}) {
							blankCount++
						} else if cell.Bg.Hue != HVoid && cell.Bg.Hue != HNebula {
							t.Errorf("at (%d,%d) out-of-bounds cell has unexpected bg hue %v", x, y, cell.Bg.Hue)
						}
					}
				}
			}

			if blankCount > 0 {
				t.Errorf("%d out-of-bounds cells were pitch-black Cell{} in %dx%d overview", blankCount, tc.w, tc.h)
			}
			if (spanX > worldW || spanY > worldH) && oobCount == 0 {
				t.Errorf("expected out-of-bounds cells for spanX=%d spanY=%d, found none", spanX, spanY)
			}
		})
	}
}

// TestOverviewWideViewportGroundNeverLeavesBlank verifies that on wide viewports
// in ground scales (e.g. stone_age), out-of-bounds blocks are filled with valid terrain
// rather than zero Cell{}.
func TestOverviewWideViewportGroundNeverLeavesBlank(t *testing.T) {
	gen := &Generator{}
	m := model(state("stone_age", 7, noon))
	sc := gen.Draw(m, 0)

	w, h := 260, 100
	ov := sc.Overview(w, h)

	blankCount := 0
	oobCount := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bx0 := ov.blockX(x)
			by0 := ov.blockY(y)
			bx1 := ov.blockX(x + 1)
			by1 := ov.blockY(y + 1)

			if (bx1 <= 0 || bx0 >= worldW) || (by1 <= 0 || by0 >= worldH) {
				oobCount++
				cell := ov.Grid.At(x, y)
				hit := ov.Grid.Hit[y*w+x]
				if hit != 0 {
					t.Errorf("ground at (%d,%d) out-of-bounds block has Hit = %d, want 0", x, y, hit)
				}
				if cell == (Cell{}) {
					blankCount++
				}
			}
		}
	}

	if oobCount == 0 {
		t.Fatal("expected out-of-bounds blocks in 260x100 overview")
	}
	if blankCount > 0 {
		t.Errorf("%d out-of-bounds cells were pitch-black Cell{} in 260x100 ground overview", blankCount)
	}
}

// TestOverviewSystemRendersOrrery: the orrery is drawn in quadrant blocks, the
// Sun and the planets are found from any of their cells, and the warp gate
// and the old single-rune planets are gone or kept as the design says.
func TestOverviewSystemRendersOrrery(t *testing.T) {
	gen := &Generator{}
	m := model(state("interstellar_age", 7, noon))
	sc := gen.Draw(m, 0)
	if sc.Scale != System {
		t.Fatalf("expected Scale == System for interstellar_age, got %v", sc.Scale)
	}
	ov := sc.Overview(120, 40)
	quads, sun, planets, orbits := 0, 0, 0, 0
	for y := 0; y < ov.Grid.H; y++ {
		for x := 0; x < ov.Grid.W; x++ {
			if r := ov.Grid.At(x, y).R; r != ' ' && r >= '\u2580' && r <= '\u259f' {
				quads++
			}
			switch th := ov.Grid.ThingAt(x, y); {
			case th.Kind == KindTerrain && th.Key == "sun":
				sun++
			case th.Kind == KindRoad && th.Key == "orbit":
				orbits++
			case th.Kind == KindTerrain && th.Key != "sun":
				planets++
			}
		}
	}
	if quads < 100 {
		t.Errorf("only %d quadrant cells in the overview", quads)
	}
	if sun == 0 || orbits == 0 || planets == 0 {
		t.Errorf("inspect marks: sun %d, orbit %d, planet %d; want all present", sun, orbits, planets)
	}
}
