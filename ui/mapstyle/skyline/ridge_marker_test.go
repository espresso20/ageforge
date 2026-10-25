package skyline

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// TestRidgeMarkerSitsOnTheRidge: a civ you have met stands on the far ridge
// as a small town with a pennant on a pole and its name just above, at every
// zoom. The name is drawn on the screen (text keeps its size), so it must
// follow the pole's top and not the world's own height: at the close level
// the scene is twice as tall and a name left where the unscaled scene put it
// hangs in empty sky. Under the name there is the pole, under the pole the
// town, under the town the ridge: no sky between.
func TestRidgeMarkerSitsOnTheRidge(t *testing.T) {
	checked := map[zoom]int{}
	for _, age := range []string{"stone_age", "bronze_age", "iron_age", "medieval_age"} {
		m := build(fixture.Options{Age: age, Seed: 3, Routes: 4, Tick: tickAt(0.45)}, 1)
		var names []string
		for _, f := range m.Factions {
			if f.Discovered {
				names = append(names, f.Name)
			}
		}
		if len(names) == 0 {
			continue // no civ met yet
		}
		for _, z := range zoomSteps {
			for _, w := range []int{120, 160} {
				// a camera at every district, and the present's own
				cams := []int{-1}
				for _, d := range m.Skyline.Districts {
					cams = append(cams, d.Centre)
				}
				for _, cam := range cams {
					v := newView()
					v.zoom, v.sw = z, w
					if cam >= 0 {
						v.follow = false
						v.cam = cam - z.cols(w)/2
					}
					scr := draw(v, m, w, 45, 40, mapmodel.TierUnicode, false)
					rows := strings.Split(capture.Text(scr), "\n")
					for _, name := range names {
						for r, line := range rows {
							idx := strings.Index(line, name)
							if idx < 0 || r < 1 || r > 44-3 {
								continue
							}
							c := len([]rune(line[:idx]))
							n := len([]rune(name))
							if !anchoredOnRidge(v, c+n/2, r) {
								t.Errorf("%s %s %dx45 cam %d: %q at row %d column %d has sky between it and the ridge", age, z, w, cam, name, r, c)
							}
							checked[z]++
						}
					}
				}
			}
		}
	}
	for _, z := range zoomSteps {
		if checked[z] == 0 {
			t.Errorf("%s: no civ's name was on screen to check", z)
		}
	}
}

// anchoredOnRidge says whether, in one of the columns at or beside the name's
// middle, everything below the name's row down to the ridge is solid (the
// pole, the town, the ridge): nothing of the sky shows between.
func anchoredOnRidge(v *view, x, r int) bool {
	for _, dx := range []int{0, -1, 1, 2, -2} {
		col, ok := x+dx, false
		for y := r + 1; y < v.fb.h-3; y++ {
			c := v.fb.at(col, y)
			if c == nil || c.d > dRidge {
				break // sky, cloud or aircraft
			}
			if c.d == dRidge || c.d == dHills || c.d <= dTown {
				ok = true // the ridge, the hills or a nearer thing in front of it
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
