package config

import (
	"maps"
	"math"
	"slices"

	"github.com/espresso20/ageforge/detmath"
)

// CarriedSupplyTargets is how many targets of the new age the carried required
// copies get to make a first copy's price in: one. A carried supply that
// cannot make the price in that time is a trickle against it, and does not
// count (the cold-start rule, smoke.coldStart.affords).
const CarriedSupplyTargets = 1.0

// handGatherable mirrors the gather command: food, wood and stone, by hand, up
// to and including the Medieval Age. Those parts need no carried supply.
var handGatherable = map[string]bool{"food": true, "wood": true, "stone": true}

// makerDrift is how far from the standard an output may sit before the hold is
// said to have moved it: half a percent, the tolerance the maker rule's test
// reads the tables to.
const makerDrift = 0.005

// ColdStartCap is one price part that capColdStartBridge held to the limit of
// the cold-start rule.
type ColdStartCap struct {
	Building, Resource string
	Before, After      float64
}

// coldStartLimit is what the carried required copies make of each resource in
// CarriedSupplyTargets targets of the age at index i, at listed output: the
// copies the gates so far asked for, and the wonder of the age before, making
// what the ages so far have unlocked. It is the rule's limit on a first-copy
// price part. It reads the outputs of earlier ages' buildings only.
func coldStartLimit(defs []BuildingDef, ages []AgeDef, i int) map[string]float64 {
	byKey := indexBy(defs, func(d BuildingDef) string { return d.Key })
	unlocked := map[string]bool{}
	copies := map[string]int{}
	for j := 0; j <= i; j++ {
		for _, r := range ages[j].UnlockResources {
			unlocked[r] = true
		}
		if j > 0 {
			for bld, n := range ages[j].BuildingReqs {
				copies[bld] = n
			}
			for _, k := range ages[j-1].UnlockBuildings {
				if d, ok := byKey[k]; ok && d.Category == "wonder" {
					copies[k] = 1
					break
				}
			}
		}
	}
	made := map[string]float64{}
	for _, k := range slices.Sorted(maps.Keys(copies)) {
		for _, e := range byKey[k].Effects {
			if e.Type == "production" && e.Value > 0 && unlocked[e.Target] {
				made[e.Target] += float64(float64(copies[k]) * e.Value)
			}
		}
	}
	lim := make(map[string]float64, len(made))
	for res, m := range made {
		lim[res] = float64(m * LatePriceTargetTicks(ages[i].Key) * CarriedSupplyTargets)
	}
	return lim
}

// floorSig3 rounds v down to three significant figures.
func floorSig3(v float64) float64 {
	if v <= 0 {
		return v
	}
	mag := detmath.Pow(10, math.Floor(detmath.Log10(v))-2)
	return math.Floor(v/mag+1e-9) * mag
}

// roundSig3 rounds v to three significant figures.
func roundSig3(v float64) float64 {
	if v <= 0 {
		return v
	}
	mag := detmath.Pow(10, math.Floor(detmath.Log10(v))-2)
	return math.Round(v/mag) * mag
}

// capColdStartBridge holds to the cold-start rule's limit every first-copy
// price part that NewCopyPriceFactor pushed over it. A part that the carried
// copies made in one target of the age before the factor (it was at or under
// the limit) and cannot after is a bridge a town arriving with nothing could
// no longer cross: its price part is set to the limit, rounded down to three
// significant figures. A part that was over the limit already is paid some
// other way (an age's own makers, a tech) and is left alone, as are flow
// resources and the hand-gathered ones in the ages that allow it. It runs
// after the factor and works for every age, in order, so it is derived, not
// typed. Two things follow from a held part, and both are done here:
//
//   - the price levels still read the price before the hold (levelCost): the
//     level is a median of what the age's buildings cost, and one held part
//     should not move the levels every other output is set against;
//   - a maker whose price fell keeps its payback: its outputs of priced
//     materials scale with its price in those units (the maker rule), to three
//     figures;
//   - a building the maker fills set (MakerFills) that no longer makes what
//     the standard gives, because the gate's asks it splits its output by
//     moved with a held price, is set to the standard (three figures).
//
// The parts it holds are returned in age order.
func capColdStartBridge(defs []BuildingDef) []ColdStartCap {
	ages := Ages()
	pos := AgePositions(AgeOrder())
	var caps []ColdStartCap
	for i, age := range ages {
		limit := coldStartLimit(defs, ages, i)
		var held []int
		for j := range defs {
			d := &defs[j]
			if d.RequiredAge != age.Key || d.Category == "wonder" || d.Category == "monument" {
				continue
			}
			resources := slices.Sorted(maps.Keys(d.BaseCost))
			before := maps.Clone(d.BaseCost)
			for _, res := range resources {
				if IsFlowResource(res) || (handGatherable[res] && i <= pos["medieval_age"]) {
					continue
				}
				lim, now := limit[res], d.BaseCost[res]
				if lim <= 0 || now <= lim || now/NewCopyPriceFactor > lim {
					continue
				}
				capped := floorSig3(lim)
				caps = append(caps, ColdStartCap{Building: d.Key, Resource: res, Before: now, After: capped})
				d.BaseCost[res] = capped
			}
			if !maps.Equal(before, d.BaseCost) {
				if d.levelCost == nil {
					d.levelCost = before
				}
				held = append(held, j)
			}
		}
		if len(held) == 0 {
			continue
		}
		levels := priceLevels(defs)[age.Key]
		for _, j := range held {
			d := &defs[j]
			was, now := priceUnits(d.levelCost, levels), priceUnits(d.BaseCost, levels)
			if was <= 0 || now >= was {
				continue
			}
			d.Effects = append([]Effect(nil), d.Effects...)
			for k, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 && levels[e.Target] > 0 && !IsFlowResource(e.Target) {
					d.Effects[k].Value = roundSig3(float64(e.Value * now / was))
				}
			}
		}
		// What the maker rule gives the age's filled makers now.
		std := MakerStandard(defs, ages, MakerPaybackTicks, MakerFills())
		for j := range defs {
			d := &defs[j]
			want, ok := std[d.Key]
			if !ok || d.RequiredAge != age.Key {
				continue
			}
			d.Effects = append([]Effect(nil), d.Effects...)
			for k, e := range d.Effects {
				if w, has := want[e.Target]; has && e.Type == "production" && e.Value > 0 && math.Abs(e.Value/w-1) > makerDrift {
					d.Effects[k].Value = roundSig3(w)
				}
			}
		}
	}
	return caps
}
