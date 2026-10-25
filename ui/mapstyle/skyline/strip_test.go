package skyline

import (
	"sort"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// stripSizes are the inner sizes of the dashboard's skyline strip (the box
// is half the Log's height at a 100x30 terminal or larger) and the sidebar's.
var stripSizes = [][2]int{{40, 8}, {50, 10}, {60, 12}, {80, 14}, {40, 15}}

// builtWonder is a model of the age whose wonder is already built: a landmark
// far taller than the houses, which used to set the strip's scale.
func builtWonder(age, key string) *mapmodel.Model {
	st := fixture.State(fixture.Options{Age: age, Seed: 3, Routes: 4, Tick: tickAt(0.45)})
	bs := st.Buildings[key]
	bs.Count, bs.WonderBank = 1, nil
	st.Buildings[key] = bs
	st.CurrentAgeWonderKey, st.CurrentAgeWonderName = "", ""
	return mapmodel.NewBuilder(catalog()).Build(&st, nil)
}

// TestStripReadsAtItsHeight: the dashboard's skyline strip (the compact
// view) shows a town whose buildings fill most of the strip's height, with
// the tallest landmark near its top, however tall that landmark is next to
// the houses; people are specks, not white faces; and nothing hangs above the
// skyline but sky (the marks for new buildings, the wonder and fire sit in
// the building's top cell).
func TestStripReadsAtItsHeight(t *testing.T) {
	type tc struct {
		name string
		m    *mapmodel.Model
	}
	var cases []tc
	for _, age := range []string{"primitive_age", "stone_age", "bronze_age", "iron_age", "medieval_age", "industrial_age", "modern_age"} {
		cases = append(cases, tc{age, build(fixture.Options{Age: age, Seed: 3, Routes: 4, Tick: tickAt(0.45)}, 1)})
		cases = append(cases, tc{age + " at night", build(fixture.Options{Age: age, Seed: 5, Routes: 2, Tick: tickAt(0.95), Harbinger: true}, 1)})
	}
	cases = append(cases, tc{"iron, wonder built", builtWonder("iron_age", "colosseum")}, tc{"classical, wonder built", builtWonder("classical_age", "parthenon")})
	for _, c := range cases {
		for _, sz := range stripSizes {
			w, h := sz[0], sz[1]
			v := newView()
			draw(v, c.m, w, h, 40, mapmodel.TierUnicode, true)
			S := h - 2 // sky and city rows; the road is the row under them
			road := h - 1
			tops := make([]int, w) // the first row of a building in each column (road: none)
			var heights []int
			for x := 0; x < w; x++ {
				tops[x] = road
				for y := 1; y < road; y++ {
					if cl := v.fb.at(x, y); cl.d <= dRow0 {
						tops[x] = y
						break
					}
				}
				if tops[x] < road {
					heights = append(heights, road-tops[x])
				}
			}
			if len(heights) < w/2 {
				t.Errorf("%s at %dx%d: buildings stand in %d of %d columns", c.name, w, h, len(heights), w)
				continue
			}
			sort.Ints(heights)
			tall, typical := heights[len(heights)-1], heights[len(heights)*3/4]
			if float64(tall) < 0.7*float64(S) {
				t.Errorf("%s at %dx%d: the tallest building is %d rows of the strip's %d", c.name, w, h, tall, S)
			}
			if float64(typical) < 0.45*float64(S) {
				t.Errorf("%s at %dx%d: the top quarter of the buildings is under %d rows of the strip's %d", c.name, w, h, typical, S)
			}
			for x := 0; x < w; x++ {
				for y := 1; y < tops[x]; y++ {
					if cl := v.fb.at(x, y); cl.d < dAirLow {
						t.Errorf("%s at %dx%d: %q hangs over the skyline at column %d row %d (depth %d)", c.name, w, h, cl.ch, x, y, cl.d)
					}
				}
				for y := 1; y < h; y++ {
					if ch := v.fb.at(x, y).ch; ch == '☻' || ch == '☺' {
						t.Errorf("%s at %dx%d: a person is a face %q at column %d row %d", c.name, w, h, ch, x, y)
					}
				}
			}
		}
	}
}
