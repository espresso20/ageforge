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

func eraBraceMaterials(epochKey string) []string {
	return eraBraceMaterialsIn(rules.Core(), epochKey)
}

// oldEraBrace is the price an ordinary doom's Brace had before it was priced
// on the warning: 12% of the most the era's remaining advances asked of each
// core resource, the same in every age of the era. It is written down, not
// worked out: the gates it was a share of have since been derived again
// (config/storage_rule.go), and 12% of today's gates is not the price that
// was.
var oldEraBrace = map[string]map[string]float64{
	"stone_era":    {"food": 9600, "wood": 4800},
	"iron_era":     {"gold": 21600, "iron": 6360, "stone": 26400},
	"steel_era":    {"gold": 1.8e+06, "steel": 288000},
	"electric_era": {"electricity": 3.96e+06, "oil": 924000, "steel": 5.64e+07},
	"digital_era":  {"data": 1.92e+10, "electricity": 1.176e+11, "gold": 1.56e+08},
	"neon_era":     {"crypto": 3e+09, "data": 4.68e+10, "electricity": 2.88e+11},
	"cosmic_era":   {"dark_matter": 1.56e+12, "titanium": 7.56e+10},
}

// eraBraceCost is that old price, doubled for level 2. Kept for the tests
// that show what the old price did.
func eraBraceCost(epochKey string, level int) map[string]float64 {
	cost := map[string]float64{}
	for k, v := range oldEraBrace[epochKey] {
		cost[k] = v * float64(level)
	}
	return cost
}

func doomBraceCost(epochKey, age string, level int) map[string]float64 {
	return doomBraceCostIn(rules.Core(), epochKey, age, level)
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
