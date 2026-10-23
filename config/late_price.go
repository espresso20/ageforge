package config

import (
	"maps"
	"math"

	"github.com/espresso20/ageforge/detmath"
)

// LatePriceStep is the per-age price step from the Modern Age on. Prices in
// the Modern Age stay at their authored value; each following age costs this
// factor more. Outputs do not move with this curve.
const LatePriceStep = 1.4

// LatePriceStartAge is the first age on the late-price curve.
const LatePriceStartAge = "modern_age"

// NewCopyPriceFactor is what every new copy of a building costs, as a multiple
// of the price it was authored at. Upgrading an older building costs
// UpgradeShare of a new copy, so the factor is what keeps the upgrade road from
// shortening the game. Wonders and monuments are not scaled.
const NewCopyPriceFactor = 1.43

// WonderBuildFactor makes a wonder the long build that ends an age.
const WonderBuildFactor = 2

// LatePriceMultiplier is the price multiplier for age. It is one before and
// in the Modern Age, and LatePriceStep for every age after it.
func LatePriceMultiplier(age string) float64 {
	positions := AgePositions(AgeOrder())
	start, ok := positions[LatePriceStartAge]
	if !ok {
		return 1
	}
	pos, ok := positions[age]
	if !ok || pos <= start {
		return 1
	}
	return detmath.Pow(LatePriceStep, float64(pos-start))
}

// LatePriceTargetTicks is an analytical pacing target for a price-derived
// rule. It is not a game clock: it describes the longer stay created by the
// late-price curve while production remains unchanged.
func LatePriceTargetTicks(age string) float64 {
	return AgeTargetTicks(age) * LatePriceMultiplier(age)
}

func applyLatePriceClimb(defs []BuildingDef) {
	for i := range defs {
		// The authored tables retain their maps and slices between calls. Make
		// this runtime pass pure so callers never compound the age multiplier.
		defs[i].BaseCost = maps.Clone(defs[i].BaseCost)
		defs[i].Effects = append([]Effect(nil), defs[i].Effects...)
		c := LatePriceMultiplier(defs[i].RequiredAge)
		if c == 1 {
			continue
		}
		for resource, price := range defs[i].BaseCost {
			defs[i].BaseCost[resource] = price * c
		}
		if defs[i].Category == "wonder" {
			// A wonder keeps its authored share of its age after prices extend
			// the stay. Construction queues require whole ticks.
			defs[i].BuildTicks = int(math.Round(float64(defs[i].BuildTicks) * c))
		}
	}
}

// applyNewCopyPrice multiplies the price of every building that is not a
// wonder or a monument by NewCopyPriceFactor. It runs on the loaded definitions
// after the late-age climb, so every reader sees one price. The cost map is
// cloned first: the authored tables keep their maps between calls.
func applyNewCopyPrice(defs []BuildingDef) {
	for i := range defs {
		if defs[i].Category == "wonder" || defs[i].Category == "monument" {
			continue
		}
		defs[i].BaseCost = maps.Clone(defs[i].BaseCost)
		for resource, price := range defs[i].BaseCost {
			defs[i].BaseCost[resource] = price * NewCopyPriceFactor
		}
	}
}

// applyWonderBuildFactor lengthens every runtime wonder after the late-price
// curve has been applied. The authored building tables remain untouched.
func applyWonderBuildFactor(defs []BuildingDef) {
	for i := range defs {
		if defs[i].Category == "wonder" {
			defs[i].BuildTicks = int(math.Round(float64(defs[i].BuildTicks) * WonderBuildFactor))
		}
	}
}
