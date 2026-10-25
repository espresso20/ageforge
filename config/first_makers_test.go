package config

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/detmath"
)

// The market's exchange is going, so a resource must have a maker in the town
// that needs it. These tests hold the tables to the rules that gave each
// resource its first maker.

// TestFirstMakerIsNotPricedInWhatItMakes: a first maker is paid for with what a
// town already makes. The first building to make a resource (in age order,
// wonders aside) is not priced in that resource, or a town could never build
// the building that starts it. The hand-gathered raws of the early ages (wood,
// stone, steel) are not held to this; the rule starts with the Modern Age.
func TestFirstMakerIsNotPricedInWhatItMakes(t *testing.T) {
	made := map[string]bool{}
	defs := BaseBuildings()
	modern := AgeByKey()["modern_age"].Order
	for _, a := range Ages() {
		var started []string // resources first made in this age
		for _, d := range defs {
			if d.RequiredAge != a.Key || d.Category == "wonder" {
				continue
			}
			for _, e := range d.Effects {
				if e.Type != "production" || e.Value <= 0 || made[e.Target] {
					continue
				}
				started = append(started, e.Target)
				if _, has := d.BaseCost[e.Target]; has && a.Order >= modern {
					t.Errorf("%s is priced in %s, a resource it is a first maker of in %s", d.Key, e.Target, a.Key)
				}
			}
		}
		for _, res := range started {
			made[res] = true
		}
	}
}

// TestWonderPartsFitTheirAge: a flow part of a wonder's price (food, faith,
// culture) is covered by the age's income within its target, and the Space
// Program asks nothing its age does not price. It asked 600B knowledge, which
// the Modern Age does not price and no town made from buildings.
func TestWonderPartsFitTheirAge(t *testing.T) {
	for _, d := range BaseBuildings() {
		if d.Category != "wonder" {
			continue
		}
		for res, c := range d.BaseCost {
			if !IsFlowResource(res) || c <= 0 {
				continue
			}
			if made := FlowIncome(res, d.RequiredAge) * LatePriceTargetTicks(d.RequiredAge); c > made {
				t.Errorf("%s asks %s %s, more than the %s the age's income makes in its target", d.Key, GoLiteral(roundSignificant(c, 4)), res, GoLiteral(roundSignificant(made, 4)))
			}
		}
	}
	space := BuildingByKey()["space_program"]
	levels := PriceLevels(space.RequiredAge)
	for res := range space.BaseCost {
		if levels[res] <= 0 {
			t.Errorf("space_program asks %s, which %s does not price", res, space.RequiredAge)
		}
	}
}

// roundSignificant rounds v to `sig` significant figures for readability
// (e.g. 46.3→46, 644→640, 6420→6400, 90150→90000). A positive value is never
// rounded down to 0 — it floors at 1 so a real cost never becomes free.
func roundSignificant(v float64, sig int) float64 {
	if v <= 0 {
		return v
	}
	d := math.Ceil(detmath.Log10(v))
	power := float64(sig) - d
	var rounded float64
	if power >= 0 {
		mag := detmath.Pow(10, power)
		rounded = math.Round(v*mag) / mag
	} else {
		// Multiply by an exact power of ten rather than divide by an inexact
		// fraction: 10/1e-5 came out as 999999.9999999999, and a wonder bank
		// filled in whole units could then never reach its price.
		mag := detmath.Pow(10, -power)
		rounded = math.Round(v/mag) * mag
	}
	if rounded < 1 {
		return 1
	}
	return rounded
}
