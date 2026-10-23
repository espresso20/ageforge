package rules

import (
	"maps"
	"slices"
	"time"

	"github.com/espresso20/ageforge/config"
)

// This file is the numbers worked out from the definitions: the pacing
// targets and the prices the deals and the harbingers read.
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
	s.flowStaffed = config.StaffedOutputs(s.buildings, s.ageKeys, config.IsFlowResource)
	s.military = config.MilitaryScales(s.techs, s.ageKeys)
	s.typStore = config.TypicalStorages(s.buildings, s.resources, s.ageKeys)
	s.refSize, s.refStore = config.ReferenceStoreSizes(s.buildings, s.ages, config.ReferenceBaseStore())
	army := config.RefSoldiers(s.buildings, s.ages)
	s.refArmy = make(map[string]float64, len(s.ages))
	for i, a := range s.ages {
		s.refArmy[a.Key] = army[i]
	}
}

// TypicalStorage is what a moderate builder's store holds of res in age:
// its base storage and config.StorageRuleCopies copies of every storage building
// up to the age (config.TypicalStorages). 0 for an unknown age or resource.
func (s *Set) TypicalStorage(res, age string) float64 { return s.typStore[age][res] }

// ReferenceStorePerCopy is what one copy of the storage building key held of
// every resource, in the reference store (config.ReferenceStoreSizes): a
// storage building holds nothing now, and the few rules that still read a
// store size are measured on this. 0 for a building that is not a storage
// building. S3 of the no-caps work retires the readers.
func (s *Set) ReferenceStorePerCopy(key string) float64 { return s.refSize[key] }

// ReferenceStore is the reference store on leaving age: the base store and
// config.StorageRuleCopies copies of the storage building of every age up to
// and including it. 0 for an unknown age.
func (s *Set) ReferenceStore(age string) float64 { return s.refStore[age] }

// AgeHasStorageBuilding reports whether age has a storage building of its own.
func (s *Set) AgeHasStorageBuilding(age string) bool {
	return config.AgeHasStorageBuilding(s.Buildings(), age)
}

// MilitaryScale is the military yardstick of the age with the given order
// on this set's tree (config.MilitaryScales): what the typical player's
// military power was when the raid threat and the missions were measured,
// what it is now, and the two scales that put the difference back.
func (s *Set) MilitaryScale(order int) config.MilitaryScaleDef {
	return config.MilitaryScaleIn(s.military, order)
}

// MilitaryScaleAt is MilitaryScale for an order that is one of the set's ages,
// and false for any other (MilitaryScale clamps it instead).
func (s *Set) MilitaryScaleAt(order int) (config.MilitaryScaleDef, bool) {
	if order < 0 || order >= len(s.military) {
		return config.MilitaryScaleDef{}, false
	}
	return s.military[order], true
}

// AgeThreat is the raid threat of the age with the given order, on this
// set's military yardstick.
func (s *Set) AgeThreat(order int) float64 {
	return config.AgeThreatAt(order, s.MilitaryScale(order).Threat)
}

// MissionPower is military power as a mission's difficulty reads it in the
// age with the given order: power × the age's MilitaryScale.Mission.
func (s *Set) MissionPower(power float64, order int) float64 {
	return float64(power * s.MilitaryScale(order).Mission)
}

// Target is the time a player should spend in age at 1x (0 for an age with
// no target).
func (s *Set) Target(age string) time.Duration { return s.targets[age] }

// TargetTicks is Target in ticks at 1x.
func (s *Set) TargetTicks(age string) float64 { return s.targetTicks[age] }

// PriceTargetTicks is the target stay that price-derived rules use. The
// late-price curve lengthens the stay while leaving production unchanged.
func (s *Set) PriceTargetTicks(age string) float64 { return config.LatePriceTargetTicks(age) }

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

// TeachingBonus is the bonus to what workers make with every scholar
// teaching, in the age at 0-based position pos in the order of ages.
func (s *Set) TeachingBonus(pos int) float64 {
	bonus := 0.0
	for _, step := range s.teaching {
		if pos+1 >= step.From {
			bonus = step.Bonus
		}
	}
	return bonus
}

// PriceLevels returns the median first-copy price of each construction
// resource in age (nil for an age with no buildings).
func (s *Set) PriceLevels(age string) map[string]float64 { return maps.Clone(s.priceLevels[age]) }

// PricedResources lists the construction resources of age, sorted.
func (s *Set) PricedResources(age string) []string { return slices.Clone(s.priced[age]) }

// DealPriceLevel is res's price level in age for faction deals: its price
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

// FlowStaffedOutput is the part of FlowBuildingOutput that buildings with
// worker slots make (config.StaffedOutputs).
func (s *Set) FlowStaffedOutput(res, age string) float64 { return s.flowStaffed[age][res] }

// BuildingOutput is FlowBuildingOutput for any resource: what the moderate
// economy's own buildings make of res by age, before any bonus (0 for a
// resource only a wonder or a tech supplies by then).
func (s *Set) BuildingOutput(res, age string) float64 { return s.built[age][res] }

// TypicalIncome is FlowIncome for any resource, construction ones included.
func (s *Set) TypicalIncome(res, age string) float64 { return s.typIncome[age][res] }
