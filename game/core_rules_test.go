package game

import (
	"math/rand"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// plantResearchBonus sets what the researched techs add to one thing (a
// kind and its target), as if a tech with that effect had been researched.
func plantResearchBonus(ge *GameEngine, kind config.TechEffectKind, target string, v float64) {
	ge.Research.bonuses[config.TechEffectKey{Kind: kind, Target: target}] = v
	ge.Research.indexPools()
}

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

func threadAppeaseCost(h *HarbingerSave, level int) map[string]float64 {
	return threadAppeaseCostIn(rules.Core(), h, level)
}

func doomAppeaseCost(epochKey, age string, level int) map[string]float64 {
	return doomAppeaseCostIn(rules.Core(), epochKey, age, level)
}

func lastPassageAppeaseCost(epochKey, age string, level int) map[string]float64 {
	return lastPassageAppeaseCostIn(rules.Core(), epochKey, age, level)
}

func harbingerBraceBasis(epochKey string) map[string]float64 {
	return harbingerBraceBasisIn(rules.Core(), epochKey)
}

func harbingerBraceCost(epochKey string, level int) map[string]float64 {
	return harbingerBraceCostIn(rules.Core(), epochKey, level)
}

func lastPassageBraceCost(epochKey, age string, level int) map[string]float64 {
	return lastPassageBraceCostIn(rules.Core(), epochKey, age, level)
}

func threadBraceCost(h *HarbingerSave, level int) map[string]float64 {
	return threadBraceCostIn(rules.Core(), h, level)
}

func loadPlan(saved []PlanItem) []PlanItem { return loadPlanIn(rules.Core(), saved) }

func cleanShares(in map[string]float64) map[string]float64 {
	return cleanSharesIn(rules.Core(), in)
}
