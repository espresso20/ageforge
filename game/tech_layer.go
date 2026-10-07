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
	f := float64(poolFactor("build_cost", pool, ge.rules.SoftCap()) * ge.Research.Bonus(config.EffectBuildCost, ""))
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

// marketFeeScale is what the techs' cut of the market's fee multiplies every
// market rate by: the market pays 1 - fee of parity, so a fee of 17 points
// where config.ExchangeFee is 20 pays 0.83 / 0.80 of the listed rate. 1 with
// no tech.
func marketFeeScale(feeShift float64) float64 {
	if feeShift == 0 {
		return 1
	}
	return (1 - (config.ExchangeFee + feeShift)) / (1 - config.ExchangeFee)
}

// pushMechanics hands the managers the mechanic numbers they read on their
// own: the market's fee and a route's time (trade), a scouting expedition's
// time (military), a gift's price and a deal set's life (diplomacy). recalculateRates calls it, so
// they follow every change to what is researched, a load and a new run
// included. Caller holds the lock.
func (ge *GameEngine) pushMechanics() {
	fee := marketFeeScale(ge.Research.Mechanic(config.MechanicMarketFee))
	ge.Trade.SetTechTerms(fee, ge.Research.Mechanic(config.MechanicRouteTicks))
	ge.Military.SetScoutTime(ge.Research.Mechanic(config.MechanicExpeditionTicks))
	ge.Diplomacy.SetTechTerms(ge.Research.Mechanic(config.MechanicGiftCost), ge.Research.Mechanic(config.MechanicDealRefreshTicks), fee)
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
