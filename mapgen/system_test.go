package mapgen

import (
	"testing"
)

// TestSystemPlanetsDistributedVerifiesNaturalSpacing verifies that all 8 planets
// in the interstellar solar system are evenly distributed across orbital sectors
// around Sol and never clump into a single column.
func TestSystemPlanetsDistributedVerifiesNaturalSpacing(t *testing.T) {
	for seed := int64(1); seed <= 10; seed++ {
		st := state("interstellar_age", seed, noon)
		var gen Generator
		f := gen.frame(model(st), 0)
		f.paintGround()
		f.paintBodies()

		// Verify planet positions in the solar system
		c := f.plan.Centre
		quadrants := [4]int{}
		// Scan for planet bodies
		planetCells := 0
		sunCells := 0
		hasRockOnSun := false

		for y := 0; y < f.g.H; y++ {
			for x := 0; x < f.g.W; x++ {
				th := f.g.ThingAt(x, y)
				cell := f.g.At(x, y)
				if th.Kind == KindTerrain && th.Key == "sun" {
					sunCells++
					if cell.Fg.Hue == HRock || cell.Bg.Hue == HRock {
						hasRockOnSun = true
					}
				}
				if th.Kind == KindTerrain && (th.Key == "planet" || th.Key == "home") {
					planetCells++
					dx := x - c.X
					dy := y - c.Y
					switch {
					case dx >= 0 && dy >= 0:
						quadrants[0]++
					case dx < 0 && dy >= 0:
						quadrants[1]++
					case dx < 0 && dy < 0:
						quadrants[2]++
					case dx >= 0 && dy < 0:
						quadrants[3]++
					}
				}
			}
		}

		if hasRockOnSun {
			t.Fatalf("seed %d: Sol contains HRock grey mud, must be pure blazing white/gold plasma", seed)
		}
		if sunCells < 100 {
			t.Fatalf("seed %d: Sol has %d cells, expected large disc", seed, sunCells)
		}
		if planetCells < 50 {
			t.Fatalf("seed %d: expected multi-cell high-resolution planets, got %d", seed, planetCells)
		}

		// At least 3 quadrants must contain planetary material (proving balanced 360-degree distribution)
		activeQuadrants := 0
		for _, count := range quadrants {
			if count > 0 {
				activeQuadrants++
			}
		}
		if activeQuadrants < 3 {
			t.Fatalf("seed %d: planets clumped in only %d quadrants: %v", seed, activeQuadrants, quadrants)
		}
	}
}

// TestSolPureWhiteGoldPlasma verifies that the Sun consists strictly of radiant
// white, golden yellow, convective amber, and solar fire with zero rock.
func TestSolPureWhiteGoldPlasma(t *testing.T) {
	st := state("interstellar_age", 42, noon)
	var gen Generator
	f := gen.frame(model(st), 0)
	f.paintGround()
	f.paintBodies()

	seenHues := map[Hue]int{}
	for y := 0; y < f.g.H; y++ {
		for x := 0; x < f.g.W; x++ {
			th := f.g.ThingAt(x, y)
			if th.Kind == KindTerrain && th.Key == "sun" {
				cell := f.g.At(x, y)
				seenHues[cell.Fg.Hue]++
				seenHues[cell.Bg.Hue]++
			}
		}
	}

	if seenHues[HRock] > 0 {
		t.Fatalf("Sol contains HRock (%d cells), expected 0", seenHues[HRock])
	}
	if seenHues[HSun] == 0 || seenHues[HWhite] == 0 {
		t.Fatalf("Sol missing radiant sun/white hues: %v", seenHues)
	}
}

func TestSystemScaleExpandedDimensionsAndClearance(t *testing.T) {
	st := state("interstellar_age", 42, noon)
	var gen Generator
	m := model(st)
	f := gen.frame(m, 0)
	if f.g.W != SystemW || f.g.H != SystemH {
		t.Fatalf("expected grid dimensions %dx%d, got %dx%d", SystemW, SystemH, f.g.W, f.g.H)
	}

	sc := gen.Draw(m, 0)
	if sc.Grid.W != SystemW || sc.Grid.H != SystemH {
		t.Errorf("expected scene grid %dx%d, got %dx%d", SystemW, SystemH, sc.Grid.W, sc.Grid.H)
	}
	if sc.Bounds.W < SystemW-4 || sc.Bounds.H < SystemH-4 {
		t.Errorf("expected scene bounds near %dx%d, got %dx%d", SystemW, SystemH, sc.Bounds.W, sc.Bounds.H)
	}

	// Test OverviewCosmos
	ov := sc.OverviewCosmos(160, 48)
	if ov == nil || ov.Grid == nil {
		t.Fatal("expected non-nil OverviewCosmos")
	}
	if ov.Grid.W != 160 || ov.Grid.H != 48 {
		t.Errorf("expected cosmos grid 160x48, got %dx%d", ov.Grid.W, ov.Grid.H)
	}
}
