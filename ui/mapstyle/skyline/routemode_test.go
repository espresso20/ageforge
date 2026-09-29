package skyline

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// TestRouteVehiclesFollowMode: a sea route sails (a ship on a bay when the
// town has one), an air route flies once the age has aircraft, and a land
// route stays on the ground.
func TestRouteVehiclesFollowMode(t *testing.T) {
	base := build(fixture.Options{Age: "galactic_age", Seed: 4, Routes: 1}, 0)
	for _, c := range []struct {
		mode mapmodel.RouteMode
		want func(v vehicle) bool
		what string
	}{
		{mapmodel.ModeLand, func(v vehicle) bool { return v.t == &tHoverTram }, "a hover tram"},
		{mapmodel.ModeAir, func(v vehicle) bool { return v.t == &tWarpShip }, "a warp ship"},
	} {
		m := *base
		m.Routes = []mapmodel.Route{{Key: "r", Mode: c.mode}}
		m.Activity.Routes = 1
		found := false
		for _, v := range trafficFor(&m, 3, 0, 160, 39) {
			if v.kind == vkRoute {
				found = found || c.want(v)
				if !c.want(v) {
					t.Errorf("%s route drew %p, want %s", c.mode, v.t, c.what)
				}
			}
		}
		if !found {
			t.Errorf("%s route drew no route vehicle", c.mode)
		}
	}

	// A sea route with a bay in the town is a ship on it.
	for _, age := range []string{"colonial_age", "industrial_age", "modern_age"} {
		for seed := int64(1); seed < 40; seed++ {
			m := build(fixture.Options{Age: age, Seed: seed, Routes: 1}, 0)
			var bay *mapmodel.District
			for i := range m.Skyline.Districts {
				if m.Skyline.Districts[i].Bay {
					bay = &m.Skyline.Districts[i]
				}
			}
			if bay == nil {
				continue
			}
			m.Routes = []mapmodel.Route{{Key: "r", Mode: mapmodel.ModeSea}}
			m.Activity.Routes = 1
			ships, street := 0, 0
			for _, v := range trafficFor(m, 3, bay.BayX, 160, 39) {
				if v.kind == vkRoute {
					if v.depth == dShip {
						ships++
					} else {
						street++
					}
				}
			}
			if ships == 0 || street != 0 {
				t.Errorf("%s seed %d: sea route drew %d ships and %d other vehicles", age, seed, ships, street)
			}
			return
		}
	}
	t.Skip("no fixture town with a bay")
}
