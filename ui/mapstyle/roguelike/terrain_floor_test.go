package roguelike

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// TestTerrainFloorInTheRoguelike holds the roguelike's own palette to the
// map floor: its night dial and the Earth arc's city inks both rebuild the
// terrain colours, on every theme, epoch and age.
func TestTerrainFloorInTheRoguelike(t *testing.T) {
	for th := range fixture.Themes(t) {
		v := &view{}
		for epoch := 0; epoch < 7; epoch++ {
			for age := -1; age <= 25; age++ {
				p := v.palette(epoch, age)
				for _, c := range mapstyle.TerrainClasses {
					if r := theme.ContrastRatio(p.Fg[c], p.Bg); r < mapstyle.TerrainFloor {
						t.Errorf("%s epoch %d age %d: class %d reads %.2f:1, want >= %.1f",
							th.Key, epoch, age, c, r, mapstyle.TerrainFloor)
					}
				}
			}
		}
	}
}
