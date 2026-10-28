package game

import (
	"math"

	"github.com/espresso20/ageforge/config"
)

// The tech layer's hooks outside recalculateRates: the factors that cut a
// price or a time, and the mechanic numbers (config.Mechanics) the engine
// and its managers read through the techs' term.

// buildCostFactor is what every building price is multiplied by: the
// build_cost pool's factor (milestones; clamp(1 + pool, 0.10, 1.0)) times
// the techs' own factor, never under config.BuildCostFloor of the price.
// Caller holds the lock.
func (ge *GameEngine) buildCostFactor(pool float64) float64 {
	f := float64(poolFactor("build_cost", pool) * ge.Research.Bonus(config.EffectBuildCost, ""))
	return math.Max(f, config.BuildCostFloor)
}

// mechanic is the mechanic number key with the techs' term applied to base,
// its constant (config.Mechanics). Caller holds the lock.
func (ge *GameEngine) mechanic(key string, base float64) float64 {
	return ge.Research.MechanicValue(key, base)
}

// mechanicTicks is mechanic for a number of ticks: rounded down, one tick at
// least. A term of 1 returns ticks untouched. Caller holds the lock.
func (ge *GameEngine) mechanicTicks(key string, ticks int) int {
	return techTimeTicks(ticks, ge.Research.Mechanic(key))
}

// pushMechanics hands the managers the mechanic numbers they read on their
// own: a route's time and what it brings in (trade), a
// scouting expedition's time, a campaign's time and what it brings back
// (military), a gift's price and what it earns, a deal set's life and what
// an ally adds (diplomacy), what an upgrade costs (buildings). recalculateRates calls it, so
// they follow every change to what is researched, a load and a new run
// included. Caller holds the lock.
func (ge *GameEngine) pushMechanics() {
	ge.Trade.SetTechTerms(ge.Research.Mechanic(config.MechanicRouteTicks), ge.Research.Mechanic(config.MechanicRouteIncome))
	ge.Military.SetScoutTime(ge.Research.Mechanic(config.MechanicExpeditionTicks))
	ge.Military.SetCampaignTerms(ge.Research.Mechanic(config.MechanicCampaignTicks), ge.Research.Mechanic(config.MechanicCampaignReward))
	ge.Diplomacy.terms = diplomacyTerms{
		giftCost:    ge.Research.Mechanic(config.MechanicGiftCost),
		giftOpinion: ge.Research.Mechanic(config.MechanicGiftOpinion),
		dealRefresh: ge.Research.Mechanic(config.MechanicDealRefreshTicks),
		allyBonus:   ge.Research.Mechanic(config.MechanicAllianceBonus),
		allyCost:    ge.Research.Mechanic(config.MechanicAllianceCost),
	}
	ge.Buildings.SetUpgradeCostTerm(ge.Research.Mechanic(config.MechanicUpgradeCost))
}

// appeaseCut is amount, one resource of an Appease level's price, with the
// techs' cut (config.MechanicAppeaseCost): amount itself with none. Caller
// holds the lock.
func (ge *GameEngine) appeaseCut(amount float64) float64 {
	term := ge.Research.Mechanic(config.MechanicAppeaseCost)
	if term <= 0 || term >= 1 {
		return amount
	}
	return float64(amount * term)
}

// raidLossFactor is what the techs leave of a raid before the garrison meets
// it: 1 with none, 0.9 after one 10% cut. Caller holds the lock.
func (ge *GameEngine) raidLossFactor() float64 {
	return ge.Research.Mechanic(config.MechanicRaidLoss)
}

// gatherBonus is what the techs add to every hand gather. Caller holds the
// lock.
func (ge *GameEngine) gatherBonus() float64 {
	return ge.Research.Mechanic(config.MechanicGatherAmount)
}
