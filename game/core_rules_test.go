package game

import (
	"math/rand"

	"github.com/espresso20/ageforge/rules"
)

// These helpers read the core ruleset. The engine and its managers read
// their own set, so nothing outside the tests calls them; they keep the
// tests' wording as it was when these lookups were package-level.

// ageKeys is every age key of the core ruleset, in order.
func ageKeys() []string { return rules.Core().AgeKeys() }

// ageOrders maps each age key of the core ruleset to its place in the order.
func ageOrders() map[string]int { return rules.Core().Indexes() }

// newAgeSight is ageSightIn on the core ruleset.
func newAgeSight(current string, reached []string, highest string) AgeSight {
	return ageSightIn(rules.Core(), current, reached, highest)
}

// baseAgeTicks is age's pacing target in ticks, before any Era Mastery
// speed-up: expectedAgeTicks at k = 1.
func baseAgeTicks(age string) float64 { return rules.Core().TargetTicks(age) }

// baseEraTicks is epochKey's expected length at k = 1: expectedEraTicks
// with no mastery.
func baseEraTicks(epochKey string) float64 {
	total := 0.0
	era, _ := rules.Core().Era(epochKey)
	for _, a := range era.Ages {
		total += float64(baseAgeTicks(a))
	}
	return total
}

func eventDelay(rng *rand.Rand, age string) int { return eventDelayIn(rules.Core(), rng, age) }

func expeditionTicks(age string, ticks int) int { return rules.Core().StretchTicks(age, ticks) }

func autoExpeditionIntervalIn(age string, count int, fill float64) int {
	return autoExpeditionInterval(rules.Core(), age, count, fill)
}

func dealRefreshFor(age string) int { return dealRefreshIn(rules.Core(), age) }

func topReward(rewards map[string]float64) (string, float64) {
	return topRewardIn(rules.Core(), rewards)
}

func harbingerAppeaseAges(epochKey string) []string {
	return harbingerAppeaseAgesIn(rules.Core(), epochKey)
}

func threadAppeaseCost(h *HarbingerSave, level int) map[string]float64 {
	return threadAppeaseCostIn(rules.Core(), h, level)
}

func doomAppeaseCost(epochKey, age string, level int) map[string]float64 {
	return doomAppeaseCostIn(rules.Core(), epochKey, age, level)
}

func eraAppeaseCost(epochKey string, level int) map[string]float64 {
	return eraAppeaseCostIn(rules.Core(), epochKey, level)
}

func harbingerBraceBasis(epochKey string) map[string]float64 {
	return harbingerBraceBasisIn(rules.Core(), epochKey)
}

func harbingerBraceCost(epochKey string, level int) map[string]float64 {
	return harbingerBraceCostIn(rules.Core(), epochKey, level)
}

func loadPlan(saved []PlanItem) []PlanItem { return loadPlanIn(rules.Core(), saved) }

func cleanShares(in map[string]float64) map[string]float64 {
	return cleanSharesIn(rules.Core(), in)
}
