package mapgen

import (
	"strings"
	"testing"
)

func closeUp(age string, seed int64, anim int) *Scene {
	return (&Generator{}).DrawAt(model(state(age, seed, noon)), Moment{Anim: anim, Visit: -1, Close: true})
}

// The calm close-up is drawn in quadrant blocks, shows Earth's colours and
// holds nothing of the player's.
func TestCalmCloseUpIsEarthAndNothingElse(t *testing.T) {
	for _, age := range []string{"space_age"} { // the Interstellar close-up is the port
		g := closeUp(age, 7, 40).Grid
		quad, hydro := 0, 0
		for _, c := range g.Cells {
			if strings.ContainsRune("▘▝▀▖▌▞▛▗▚▐▜▄▙▟█", c.R) {
				quad++
			}
			if c.Fg.Hue == HHydro || c.Bg.Hue == HHydro {
				hydro++
			}
		}
		if quad < 200 || hydro == 0 {
			t.Errorf("%s: %d quadrant cells, %d ocean cells", age, quad, hydro)
		}
		for i, id := range g.Hit {
			if id == 0 {
				continue
			}
			if k := g.Things[id].Kind; k == KindBuilding || k == KindWonder || k == KindAmbient {
				t.Errorf("%s: cell %d shows a %v at the close-up", age, i, k)
			}
		}
	}
}

func TestCalmCloseUpSameSeedSameFrame(t *testing.T) {
	a, b := closeUp("space_age", 5, 33).Grid, closeUp("space_age", 5, 33).Grid
	for i := range a.Cells {
		if a.Cells[i] != b.Cells[i] || a.Hit[i] != b.Hit[i] {
			t.Fatalf("cell %d differs between two draws of one seed", i)
		}
	}
}

func TestCalmCloseUpVariesBySeed(t *testing.T) {
	seen := map[string]bool{}
	for seed := int64(1); seed <= 6; seed++ {
		var sb strings.Builder
		for _, c := range closeUp("space_age", seed, 0).Grid.Cells {
			sb.WriteRune(c.R)
		}
		seen[sb.String()] = true
	}
	if len(seen) < 5 {
		t.Errorf("only %d different pictures from 6 seeds", len(seen))
	}
}

// The Moon, the Earth and a station can each be found by their cells.
func TestCalmCloseUpNamesItsBodies(t *testing.T) {
	g := closeUp("space_age", 7, 40).Grid
	keys := map[string]bool{}
	for _, id := range g.Hit {
		if id != 0 {
			keys[g.Things[id].Key] = true
		}
	}
	for _, k := range []string{"home", "satellite", "space_station"} {
		if !keys[k] {
			t.Errorf("no %s in the close-up", k)
		}
	}
}
