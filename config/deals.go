package config

import "sort"

// Price levels for faction trade deals (game/deals.go).
//
// A deal needs a value for both sides of it. For the construction resources
// of an age that is the market's price level (PriceLevels): the median
// first-copy price. Flow resources have no price level, because no building
// is priced in them, so a deal values them the other way round: by what a
// price unit invested in the age's producers of them returns over
// PaybackTicks(age). That is the Payback Rule run backwards (for a producer
// of one construction resource it gives back that resource's price level),
// so the two kinds of level are in the same unit and a deal can trade food
// for iron.

// DealPriceLevel is res's price level in age for faction deals: the market's
// level for a construction resource of age, and for a flow resource the
// median over age's producers of it of rate x PaybackTicks(age) /
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
// with a market price level.
func PricedResources(age string) []string {
	return PricedResourcesAt(PriceLevels(age))
}

// PricedResourcesAt is PricedResources against an age's price levels.
func PricedResourcesAt(lv map[string]float64) []string {
	out := make([]string, 0, len(lv))
	for r := range lv {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}

// MarketOffers is the rate the market pays for one from in age, and whether
// it trades the pair there at all: MarketRate, except that a listed pair
// (BaseExchangeRates) counts only once its MinAge is reached, as in
// MarketPairs, which is what the trade panel and `trade` offer.
func MarketOffers(from, to, age string) (float64, bool) {
	return MarketOffersAt(from, to, age, ExchangeRateByKey(), AgePositions(AgeOrder()), PriceLevels(age))
}

// MarketOffersAt is MarketOffers against the listed pairs (keyed "from:to"),
// each age's position and age's price levels.
func MarketOffersAt(from, to, age string, listed map[string]ExchangeRateDef, pos map[string]int, lv map[string]float64) (float64, bool) {
	if def, ok := listed[from+":"+to]; ok {
		if pos[def.MinAge] > pos[age] {
			return 0, false
		}
	}
	return MarketRateAt(from, to, listed, lv)
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
			v := float64(e.Value*paybackTicks(d.RequiredAge, pos)) / u
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
