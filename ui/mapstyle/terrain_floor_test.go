package mapstyle

import (
	"testing"

	"github.com/espresso20/ageforge/theme"
)

// TestTerrainFloorOnEveryThemeAndEpoch holds every terrain and detail class
// to the map floor (TerrainFloor) against the map background, on
// all 16 themes and all seven epochs. Light themes darken the ground and
// dark themes lighten it until it reads.
func TestTerrainFloorOnEveryThemeAndEpoch(t *testing.T) {
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for epoch := 0; epoch < 7; epoch++ {
			p := NewPalette(epoch)
			for _, c := range TerrainClasses {
				if r := theme.ContrastRatio(p.Fg[c], p.Bg); r < TerrainFloor {
					t.Errorf("%s epoch %d: class %d reads %.2f:1 on the map, want >= %.1f",
						th.Key, epoch, c, r, TerrainFloor)
				}
			}
		}
	}
}
