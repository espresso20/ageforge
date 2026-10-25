package config

import (
	"maps"
	"slices"
)

// Price levels for faction trade deals (game/deals.go).
//
// A deal needs a value for both sides of it. For the construction resources
// of an age that is its price level (PriceLevels): the median
// first-copy price. Flow resources have no price level, because no building
// is priced in them, so a deal values them the other way round: by what a
// price unit invested in the age's producers of them returns over
// DealPaybackTicks[age] ticks (for a producer of one construction resource
// that gives back about that resource's price level),
// so the two kinds of level are in the same unit and a deal can trade food
// for iron.

// DealPriceLevel is res's price level in age for faction deals: the price
// level for a construction resource of age, and for a flow resource the
// median over age's producers of it of rate x DealPaybackTicks[age] /
// priceUnits(first copy). 0 when res has neither (a flow resource nothing
// in age produces, or a resource the age neither prices nor produces).
func DealPriceLevel(res, age string) float64 {
	defs := BaseBuildings()
	return DealPriceLevelAt(res, priceLevels(defs)[age], FlowDealLevels(defs, AgePositions(AgeOrder()))[age])
}

// DealPriceLevelAt is DealPriceLevel against an age's price levels and its
// flow levels (FlowDealLevels).
func DealPriceLevelAt(res string, lv, flow map[string]float64) float64 {
	if l := lv[res]; l > 0 {
		return l
	}
	return flow[res]
}

// PricedResources lists the construction resources of age, sorted: the ones
// with a price level.
func PricedResources(age string) []string {
	return PricedResourcesAt(PriceLevels(age))
}

// PricedResourcesAt is PricedResources against an age's price levels.
func PricedResourcesAt(lv map[string]float64) []string {
	return slices.Sorted(maps.Keys(lv))
}

// DealPaybackTicks is, for each age, how many ticks of a flow producer's
// output a deal counts as one price unit of that producer's cost. The numbers
// are written here as they are used; they are what the old payback rule gave
// each age when it still set building outputs, kept so that deal prices for
// food, faith, culture and soldiers do not move.
var DealPaybackTicks = map[string]float64{
	"primitive_age":    28.125,
	"stone_age":        109.30968650959576,
	"bronze_age":       764.3172991517536,
	"iron_age":         1364.5607501225809,
	"classical_age":    2194.6891749036686,
	"medieval_age":     3182.073766809561,
	"renaissance_age":  9434.442582123225,
	"colonial_age":     6050.82762760758,
	"industrial_age":   8664.788873733401,
	"victorian_age":    14891.66881178869,
	"electric_age":     20116.210779976245,
	"atomic_age":       24074.961022065374,
	"modern_age":       28867.708221475667,
	"information_age":  14935.330237167702 * LatePriceStep,
	"digital_age":      40917.155671143795 * LatePriceStep * LatePriceStep,
	"cyberpunk_age":    49441.02188451514 * LatePriceStep * LatePriceStep * LatePriceStep,
	"fusion_age":       129306.09003805669 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep,
	"space_age":        96507.82168363156 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep,
	"interstellar_age": 79929.9722418292 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep,
	"galactic_age":     169431.39649701896 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep,
	"quantum_age":      44778.06713018769 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep,
	"transcendent_age": 47224.742680648604 * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep * LatePriceStep,
}

// FlowDealLevels is every age's flow levels, as age -> flow resource ->
// level, with each age's position given (AgePositions). Pure.
func FlowDealLevels(defs []BuildingDef, pos map[string]int) map[string]map[string]float64 {
	levels := priceLevels(defs)
	vals := map[string]map[string][]float64{}
	for _, d := range defs {
		if d.RequiredAge == "" || d.Category == "wonder" {
			continue
		}
		u := priceUnits(d.BaseCost, levels[d.RequiredAge])
		if u <= 0 {
			continue
		}
		for _, e := range d.Effects {
			if e.Type != "production" || e.Value <= 0 || !flowResources[e.Target] {
				continue
			}
			if vals[d.RequiredAge] == nil {
				vals[d.RequiredAge] = map[string][]float64{}
			}
			v := float64(e.Value*DealPaybackTicks[d.RequiredAge]) / u
			vals[d.RequiredAge][e.Target] = append(vals[d.RequiredAge][e.Target], v)
		}
	}
	out := make(map[string]map[string]float64, len(vals))
	for age, byRes := range vals {
		out[age] = make(map[string]float64, len(byRes))
		for res, v := range byRes {
			out[age][res] = median(v)
		}
	}
	return out
}
