package config

import "testing"

// TestNoBridgePartIsPushedOverTheColdStartLimit: after the factor, no first-copy
// price part that the carried copies made in one target of the age before it
// is over what they make now. The parts that were held are the listed ones,
// each at the limit rounded down to three figures.
func TestNoBridgePartIsPushedOverTheColdStartLimit(t *testing.T) {
	defs, ages := BaseBuildings(), Ages()
	pos := AgePositions(AgeOrder())
	held := map[string]ColdStartCap{}
	for _, c := range ColdStartCaps() {
		held[c.Building+"/"+c.Resource] = c
	}
	if len(held) == 0 {
		t.Fatal("no part was held: the test below has nothing to prove")
	}
	for _, d := range defs {
		if d.RequiredAge == "" || d.Category == "wonder" || d.Category == "monument" {
			continue
		}
		at := pos[d.RequiredAge]
		limit := coldStartLimit(defs, ages, at)
		for res, now := range d.BaseCost {
			if IsFlowResource(res) || (handGatherable[res] && at <= pos["medieval_age"]) || limit[res] <= 0 {
				continue
			}
			c, wasHeld := held[d.Key+"/"+res]
			if wasHeld {
				if want := floorSig3(limit[res]); now != want || c.After != want || c.Before <= limit[res] {
					t.Errorf("%s %s: held to %s (recorded %s, before %s), want the limit %s rounded down", d.Key, res, GoLiteral(now), GoLiteral(c.After), GoLiteral(c.Before), GoLiteral(want))
				}
				continue
			}
			if now > limit[res] && now/NewCopyPriceFactor <= limit[res] {
				t.Errorf("%s %s costs %s, over the %s the carried copies make in one target, and the factor pushed it there", d.Key, res, GoLiteral(now), GoLiteral(limit[res]))
			}
		}
	}
}

// TestHeldMakerKeepsItsPayback: a maker whose price fell when a part was held
// is still on the standard (the maker-rule tests prove the tables; this names
// the one the cold-start rule held that is a maker).
func TestHeldMakerKeepsItsPayback(t *testing.T) {
	byKey := BuildingByKey()
	d := byKey["orbital_refinery"]
	if d.levelCost == nil {
		t.Skip("the Orbital Refinery was not held at this factor")
	}
	levels := priceLevels(BaseBuildings())[d.RequiredAge]
	price, out := makerUnits(d, levels, nil)
	if r := float64(price/out) / MakerPaybackTicks[d.RequiredAge]; r < 0.99 || r > 1.01 {
		t.Errorf("the held Orbital Refinery pays back in %.3f of payback(%s)", r, d.RequiredAge)
	}
}
