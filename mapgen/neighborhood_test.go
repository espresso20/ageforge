package mapgen

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
)

func nearScene(seed int64, cs []Colony) *Scene {
	s := (&Generator{}).DrawAt(model(state("galactic_age", seed, noon)), Moment{Visit: -1})
	s.Colonies = cs
	return s
}

// starCell is where a named star is drawn: the mean of the cells that pick it.
func starCell(o *Overview, key string) (x, y int, ok bool) {
	for i, nr := range NearStars {
		if nr.Key != key {
			continue
		}
		n := 0
		for cy := 0; cy < o.Grid.H; cy++ {
			for cx := 0; cx < o.Grid.W; cx++ {
				if th := o.Grid.ThingAt(cx, cy); th.Key == "neighbor" && th.Ref == i+1 {
					x, y, n = x+cx, y+cy, n+1
				}
			}
		}
		if n > 0 {
			return x / n, y / n, true
		}
	}
	return 0, 0, false
}

// TestNeighborhoodIsToScale: distance from Sol is linear on screen, so Sirius
// lies farther out than Alpha Centauri, and Vega farther still.
func TestNeighborhoodIsToScale(t *testing.T) {
	o := nearScene(7, nil).OverviewNeighborhood(120, 40)
	sx, sy := 60, 20
	dist := func(key string) int {
		x, y, ok := starCell(o, key)
		if !ok {
			t.Fatalf("%s is not drawn", key)
		}
		return sq(x-sx) + 4*sq(y-sy)
	}
	if a, s, v := dist("alpha_centauri"), dist("sirius"), dist("vega"); !(a < s && s < v) {
		t.Errorf("distances on screen: Alpha Centauri %d, Sirius %d, Vega %d", a, s, v)
	}
}

// TestNeighborhoodLightsColonies: a colonized star carries its colony and is
// joined to Sol by a lane; the same seed and frame draw the same grid.
func TestNeighborhoodLightsColonies(t *testing.T) {
	cs := []Colony{{Line: mapmodel.LinFood, Count: 12, Systems: 3, Tier: 2}, {Line: mapmodel.LinHousing, Count: 3, Systems: 1, Tier: 0}}
	for _, c := range cs {
		if _, ok := lineOf[c.Line]; !ok {
			t.Fatalf("no line %q", c.Line)
		}
	}
	o := nearScene(7, cs).OverviewNeighborhood(120, 40)
	lit, lanes := 0, 0
	for _, th := range o.Grid.Things {
		if th.Kind == KindTerrain && th.Key == "neighbor" && th.Lineage != "" {
			lit++
		}
		if th.Kind == KindLane {
			lanes++
		}
	}
	if lit != len(cs) || lanes < len(cs) {
		t.Errorf("%d lit stars and %d lanes for %d colonies", lit, lanes, len(cs))
	}
	again := nearScene(7, cs).OverviewNeighborhood(120, 40)
	for i := range o.Grid.Cells {
		if o.Grid.Cells[i] != again.Grid.Cells[i] {
			t.Fatalf("cell %d differs between two draws of one seed", i)
		}
	}
}
