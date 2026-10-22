package smoke

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// config.RefHousing (which the military buildings' soldiers are sized from)
// is the housing the reference town of smoke's players has at the end of each
// age: the same sum as the room refTables.state works out.
func TestRefHousingIsTheReferenceTownsHousing(t *testing.T) {
	tb := newRefTables()
	housing := config.RefHousing(config.BaseBuildings(), config.Ages())
	for i, a := range tb.ages {
		if want := tb.state(tb.refTown(i), i, 1).room; math.Abs(housing[i]-want) > 1e-9*math.Max(1, want) {
			t.Errorf("%s: config says %v homes, the reference town has %v", a.Key, housing[i], want)
		}
	}
}
