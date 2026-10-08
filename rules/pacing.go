package rules

import (
	"maps"
	"slices"
	"time"

	"github.com/espresso20/ageforge/config"
)

// This file is the numbers worked out from the definitions: the pacing
// targets and the prices the market, the deals and the harbingers read.
// derive fills them once, at compile, with package config's own formulas,
// so a set and config can never disagree about a rate.

func (s *Set) derive() {
	s.priceLevels = config.PriceLevelsByAge(s.buildings)
	s.priced = make(map[string][]string, len(s.priceLevels))
	for age, lv := range s.priceLevels {
		s.priced[age] = config.PricedResourcesAt(lv)
	}
	s.flowLevels = config.FlowDealLevels(s.buildings, s.agePos)
	s.flowIncome = config.Incomes(s.buildings, s.techs, s.ageKeys, config.IsFlowResource)
	s.typIncome = config.Incomes(s.buildings, s.techs, s.ageKeys, config.AnyResource)
	s.flowBuilt = config.BuildingOutputs(s.buildings, s.ageKeys, config.IsFlowResource)
	s.built = config.BuildingOutputs(s.buildings, s.ageKeys, config.AnyResource)
}

// Target is the time a player should spend in age at 1x (0 for an age with
// no target).
func (s *Set) Target(age string) time.Duration { return s.targets[age] }

// TargetTicks is Target in ticks at 1x.
func (s *Set) TargetTicks(age string) float64 { return s.targetTicks[age] }

// StretchTicks re-times a clock typed in ticks for age: ticks times the
// factor the age's clocks run at (1 for an age the set has no factor for),
// rounded to the nearest tick.
func (s *Set) StretchTicks(age string, ticks int) int {
	factor, ok := s.stretch[age]
	if !ok {
		factor = 1
	}
	return config.StretchTicksBy(ticks, factor)
}

// SoftCap is the rule every production bonus pool follows (all production,
// and each resource's own production): a pool applies in full up to the
// knee, and a share of each point past it (config.SoftCap).
func (s *Set) SoftCap() config.SoftCap { return s.softCap }

// PriceLevels returns the median first-copy price of each construction
// resource in age (nil for an age with no buildings).
func (s *Set) PriceLevels(age string) map[string]float64 { return maps.Clone(s.priceLevels[age]) }

// PricedResources lists the construction resources of age, sorted.
func (s *Set) PricedResources(age string) []string { return slices.Clone(s.priced[age]) }

// DealPriceLevel is res's price level in age for faction deals: the market's
// level for a construction resource, the flow level for a flow resource the
// age produces, 0 for neither.
func (s *Set) DealPriceLevel(res, age string) float64 {
	return config.DealPriceLevelAt(res, s.priceLevels[age], s.flowLevels[age])
}

// FlowIncome is what a player who invests moderately in the flow resource
// res makes per tick in age at 1x (config.FlowIncome has the full rule).
func (s *Set) FlowIncome(res, age string) float64 { return s.flowIncome[age][res] }

// FlowBuildingOutput is the part of FlowIncome the moderate economy's own
// buildings make, before any bonus: config.FlowCopies fully staffed copies
// of every non-wonder producer of the flow resource res up to and including
// age, per tick (config.BuildingOutputs).
func (s *Set) FlowBuildingOutput(res, age string) float64 { return s.flowBuilt[age][res] }

// BuildingOutput is FlowBuildingOutput for any resource: what the moderate
// economy's own buildings make of res by age, before any bonus (0 for a
// resource only a wonder, a tech or the market supplies by then).
func (s *Set) BuildingOutput(res, age string) float64 { return s.built[age][res] }

// TypicalIncome is FlowIncome for any resource, construction ones included.
func (s *Set) TypicalIncome(res, age string) float64 { return s.typIncome[age][res] }

// ExchangeRate is the market rate from def.From to def.To in age.
func (s *Set) ExchangeRate(def config.ExchangeRateDef, age string) float64 {
	return config.ExchangeRateAt(def, s.priceLevels[age])
}

// MarketRate is the rate the market pays for one from in age, and whether
// the pair trades at all.
func (s *Set) MarketRate(from, to, age string) (float64, bool) {
	return config.MarketRateAt(from, to, s.exchangeBy, s.priceLevels[age])
}

// MarketOffers is MarketRate, except that a listed pair counts only once
// its MinAge is reached.
func (s *Set) MarketOffers(from, to, age string) (float64, bool) {
	return config.MarketOffersAt(from, to, age, s.exchangeBy, s.agePos, s.priceLevels[age])
}

// MarketPairs lists what the market offers in age, each pair with its rate
// for age as BaseRate, sorted by from, then to.
func (s *Set) MarketPairs(age string) []config.ExchangeRateDef {
	return config.MarketPairsAt(age, s.exchange, s.agePos, s.priceLevels[age])
}
