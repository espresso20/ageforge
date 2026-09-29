package roguelike

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// TestWallRingSteady: as a town grows the wall ring keeps its centre and
// its radius only ever steps outward along the ladder, never shrinking or
// creeping a tile at a time.
func TestWallRingSteady(t *testing.T) {
	for _, age := range []string{"bronze_age", "medieval_age", "renaissance_age", "industrial_age"} {
		for _, seed := range []int64{1, 4, 9} {
			st := fixture.State(fixture.Options{Age: age, Seed: seed})
			var prevR float64
			var prevC mapmodel.Pt
			steps := 0
			for g := 0; g < 12; g++ {
				s := newScene(testBuilder.Build(&st, nil))
				onLadder := false
				for _, x := range wallRadii {
					onLadder = onLadder || x == s.wallR
				}
				if !onLadder {
					t.Errorf("%s/%d grow %d: radius %v is off the ladder", age, seed, g, s.wallR)
				}
				if g > 0 {
					if s.wallC != prevC {
						t.Errorf("%s/%d grow %d: the ring moved from %v to %v", age, seed, g, prevC, s.wallC)
					}
					if s.wallR < prevR {
						t.Errorf("%s/%d grow %d: the ring shrank from %v to %v", age, seed, g, prevR, s.wallR)
					}
					if s.wallR != prevR {
						steps++
					}
				}
				prevR, prevC = s.wallR, s.wallC
				st = fixture.Grow(st, 1)
			}
			if steps > 6 {
				t.Errorf("%s/%d: the ring changed size %d times in 12 growth steps", age, seed, steps)
			}
		}
	}
}

// TestDistrictLabelBudget: the district zoom names at most its budget of
// building types, nearest the cursor first, and still names some.
func TestDistrictLabelBudget(t *testing.T) {
	for _, age := range []string{"medieval_age", "industrial_age", "galactic_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 7})
		for _, sz := range [][2]int{{80, 24}, {160, 48}, {220, 60}} {
			v := newView()
			v.SetOption(mapstyle.OptInspect, true)
			v.zoom = zDistrict
			draw(v, m, sz[0], sz[1], 0, mapmodel.TierUnicode, false)
			budget := districtLabelBudget(v.g.w, v.g.h)
			if v.names > budget {
				t.Errorf("%s %dx%d: %d names, budget %d", age, sz[0], sz[1], v.names, budget)
			}
			if v.names == 0 {
				t.Errorf("%s %dx%d: no building names at district zoom", age, sz[0], sz[1])
			}
		}
	}
}
