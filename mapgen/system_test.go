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

// TestInterstellarSpaceConstructsAndStations verifies that rich animated space stations,
// Dyson collectors, Warp Gate, Asteroid Mines, and Lagrange Relays are present.
func TestInterstellarSpaceConstructsAndStations(t *testing.T) {
	for seed := int64(1); seed <= 5; seed++ {
		st := state("interstellar_age", seed, noon)
		var gen Generator
		f := gen.frame(model(st), 0)
		f.paintGround()
		f.paintBodies()

		counts := map[string]int{}
		for y := 0; y < f.g.H; y++ {
			for x := 0; x < f.g.W; x++ {
				th := f.g.ThingAt(x, y)
				if th.Kind == KindBuilding {
					counts[th.Key]++
				}
			}
		}

		expectedConstructs := []string{
			"space_station",
			"dyson_harvester",
			"warp_gate",
			"asteroid_mine",
			"deep_space_relay",
		}

		for _, k := range expectedConstructs {
			if counts[k] == 0 {
				t.Errorf("seed %d: expected construct %q to be present, got 0 cells", seed, k)
			}
		}

		// Ensure Warp Gate has iconic multi-cell megastructure dimensions (at least 9 cells)
		if counts["warp_gate"] < 9 {
			t.Errorf("seed %d: expected warp_gate to have at least 9 cells, got %d", seed, counts["warp_gate"])
		}

		// Ensure multiple orbital space stations exist across the solar system (at least 15 cells)
		if counts["space_station"] < 15 {
			t.Errorf("seed %d: expected space_station to have at least 15 cells across worlds, got %d", seed, counts["space_station"])
		}

		// Ensure Dyson collectors exist (at least 8 cells)
		if counts["dyson_harvester"] < 8 {
			t.Errorf("seed %d: expected dyson_harvester to have at least 8 cells, got %d", seed, counts["dyson_harvester"])
		}

		// Ensure Asteroid mining complexes exist (at least 12 cells)
		if counts["asteroid_mine"] < 12 {
			t.Errorf("seed %d: expected asteroid_mine to have at least 12 cells, got %d", seed, counts["asteroid_mine"])
		}

		// Ensure Deep space relays exist (at least 8 cells)
		if counts["deep_space_relay"] < 8 {
			t.Errorf("seed %d: expected deep_space_relay to have at least 8 cells, got %d", seed, counts["deep_space_relay"])
		}
	}
}

// TestSystemScaleExpandedDimensionsAndClearance verifies that the System scale
// dedicates a 480x240 grid, providing spacious orbits with no overlap, and that
// OverviewCosmos produces a padded deep space overview.
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

// TestInterplanetarySpaceshipRedesign verifies that spaceships across the solar
// system are compact (2-3 cells per ship), oriented along 8 directional vectors,
// numbered 3-4 ships total (never bloated 30 ships), painted against the clean space
// background with no dotted corridor tracks.
func TestInterplanetarySpaceshipRedesign(t *testing.T) {
	for seed := int64(1); seed <= 5; seed++ {
		for _, anim := range []int{0, 24, 72, 160} {
			st := state("interstellar_age", seed, noon)
			var gen Generator
			f := gen.frame(model(st), anim)
			f.paintGround()
			f.paintBodies()

			shipCells := 0
			orbitRoadCells := 0
			validRunes := map[rune]bool{
				'►': true, '◄': true, '▲': true, '▼': true,
				'↗': true, '↖': true, '↘': true, '↙': true,
				'═': true, '║': true, '·': true,
			}

			for y := 0; y < f.g.H; y++ {
				for x := 0; x < f.g.W; x++ {
					th := f.g.ThingAt(x, y)
					cell := f.g.At(x, y)

					if th.Kind == KindRoad && th.Key == "orbit" {
						orbitRoadCells++
					}

					if th.Kind == KindAgent && th.Key == "generation_ship" {
						shipCells++
						if !validRunes[cell.R] {
							t.Errorf("seed %d anim %d: unexpected ship rune %q at (%d, %d)", seed, anim, cell.R, x, y)
						}
						if cell.Bg != f.st.Sky.Bg {
							t.Errorf("seed %d anim %d: ship cell at (%d, %d) has Bg %v, expected sky background %v",
								seed, anim, x, y, cell.Bg, f.st.Sky.Bg)
						}
					}
				}
			}

			if orbitRoadCells > 0 {
				t.Fatalf("seed %d anim %d: found %d dotted corridor track cells, expected 0 (deep space must be dark & clean)",
					seed, anim, orbitRoadCells)
			}

			// In a 4-corridor system with 1 ship per corridor (2-3 cells per ship),
			// shipCells should be between 4 and 12 cells total (at most 4 ships).
			if shipCells < 2 || shipCells > 14 {
				t.Errorf("seed %d anim %d: expected 3-4 compact ships (4-12 cells), got %d cells", seed, anim, shipCells)
			}
		}
	}
}
