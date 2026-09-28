package config

import (
	"sort"
	"sync"
)

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
	if lv := exchangeLevels()[age][res]; lv > 0 {
		return lv
	}
	return flowDealLevels()[age][res]
}

// PricedResources lists the construction resources of age, sorted: the ones
// with a market price level.
func PricedResources(age string) []string {
	lv := exchangeLevels()[age]
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
	if def, ok := exchangeByKey()[from+":"+to]; ok {
		if ageIndex()[def.MinAge] > ageIndex()[age] {
			return 0, false
		}
	}
	return MarketRate(from, to, age)
}

var (
	flowDealOnce      sync.Once
	flowDealLevelsMap map[string]map[string]float64
)

// flowDealLevels caches the flow levels. Config is static; read-only.
func flowDealLevels() map[string]map[string]float64 {
	flowDealOnce.Do(func() {
		flowDealLevelsMap = computeFlowDealLevels(BaseBuildings())
	})
	return flowDealLevelsMap
}

func computeFlowDealLevels(defs []BuildingDef) map[string]map[string]float64 {
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
			v := float64(e.Value*PaybackTicks(d.RequiredAge)) / u
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
