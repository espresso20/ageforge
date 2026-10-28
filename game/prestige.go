package game

import (
	"fmt"
	"math"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// PrestigeManager manages the prestige meta-progression layer.
// Players prestige from the Medieval Age on (PrestigeMinAge), earning depth
// points: every age the run completed pays 3^epoch (rules.Set.DepthPoints).
// Points are a permanent production bonus (ProductionMultiplier). The legacy
// kit (legacy.go) carries a run's automation across the reset and comes with
// the first prestige.
//
// It also holds Era Mastery (mastery.go): each age's mastery, the record
// (the deepest age ever entered) and this run's furthest age. The manager
// survives prestige and Succumb and is replaced only by Reset, so mastery
// lasts as long as the account's game does. The old passive (+2% production
// and +1% tick speed per level) retired into mastery.
type PrestigeManager struct {
	rules       *rules.Set
	level       int
	totalEarned int
	available   int

	// Era Mastery: age key -> mastery (ages at 0 left out), the record and
	// this run's furthest age ("" for none yet).
	// speeds is every age's k, rebuilt whenever mastery or the record
	// changes (rebuildSpeeds).
	mastery     map[string]int
	record      string
	floor       string
	runFurthest string
	speeds      map[string]float64

	// The legacy kit's memory (legacy.go): what the runs so far wrote and
	// did, kept whether or not the kit is bought, so an item bought after a
	// prestige works at once. legacyPlan is the plan template, by age;
	// legacyFactions the civilizations met; legacyShares the worker shares
	// (nil: every domain on auto).
	legacyPlan     []PlanTemplateItem
	legacyFactions []string
	legacyShares   map[string]float64

	prestigeCoefficient float64
}

// NewPrestigeManagerWith creates a new prestige manager on set: its shop,
// its ages and its depth points.
func NewPrestigeManagerWith(set *rules.Set) *PrestigeManager {
	pm := &PrestigeManager{
		mastery: make(map[string]int),
		// A new game starts in the first age: its record and its run's furthest age.
		record:      set.AgeKeys()[0],
		runFurthest: set.AgeKeys()[0],
	}
	pm.Rebind(set)
	return pm
}

// Rebind moves the manager onto set and works
// every age's speed out again on set's ages. Levels, points, mastery and
// the legacy kit's memory stay.
func (pm *PrestigeManager) Rebind(set *rules.Set) {
	pm.rules = set
	pm.prestigeCoefficient = set.PrestigeBonusCoefficient()
	pm.rebuildSpeeds()
}

// CalculatePoints is what a prestige from age pays: its depth points
// (rules.Set.DepthPoints), the sum of 3^epoch over every age the run
// completed. No divisor and no milestone, tech or building terms: a deeper
// run pays more, and a level costs nothing.
func (pm *PrestigeManager) CalculatePoints(age string) int {
	return pm.rules.DepthPoints(age)
}

// PrestigeMinAge is the age that opens prestige; every later age counts too.
// A prestige from here to the Atomic Age is an early taste: it pays little
// (the Medieval Age 9 points, the Modern Age 120).
const PrestigeMinAge = "medieval_age"

// PrestigeRunAge is the age from which a prestige counts as a full run
// rather than a taste, for the account's records (badges count only these).
const PrestigeRunAge = "modern_age"

// countAges counts ages.
func countAges(n int) string {
	return textfmt.Count(n, vibe.Line("game.prestige.EarlyPrestigeLine.06", "age"), vibe.Line("game.prestige.EarlyPrestigeLine.07", "ages"))
}

// EarlyPrestigeLine is the plain line an early prestige carries: one from
// before PrestigeRunAge (an early taste, as the prestige panel calls it),
// which pays little. It says what the prestige pays
// and that going deeper pays far more, so nobody takes the early prestige
// for the game's real payout. done picks the tense: the confirm warning
// says what it will pay, the new run's log what it paid. "" for a prestige
// from PrestigeRunAge or later. sight keeps it from naming an age the player
// has not seen (the no-spoiler rule): it then counts the ages to go instead.
func EarlyPrestigeLine(sight AgeSight, from string, points int, done bool) string {
	set := orCore(sight.set)
	at, ok := set.Index(from)
	run, _ := set.Index(PrestigeRunAge)
	if !ok || at >= run {
		return ""
	}
	full := countPrestigePoints(set.DepthPoints(PrestigeRunAge))
	deeper := vibe.Line("game.prestige.EarlyPrestigeLine.03", "a run to the ") + set.Name(rules.KindAge, PrestigeRunAge) + vibe.Line("game.prestige.EarlyPrestigeLine.04", " pays ") + full
	if !sight.Age(PrestigeRunAge) {
		deeper = vibe.Line("game.prestige.EarlyPrestigeLine.05", "a full run, ") + countAges(run-at) + vibe.Line("game.prestige.EarlyPrestigeLine.08", " further on, pays ") + full
	}
	lead := vibe.Line("game.prestige.EarlyPrestigeLine.09", "This is an early taste: a prestige from the %s pays %s, for the %s this run completed.")
	if done {
		lead = vibe.Line("game.prestige.EarlyPrestigeLine.10", "That was an early taste: a prestige from the %s paid %s, for the %s the run completed.")
	}
	return fmt.Sprintf(lead+vibe.Line("game.prestige.EarlyPrestigeLine.11", " Going deeper pays far more: each era's ages are worth 3 times the era before, and %s."),
		set.Name(rules.KindAge, from), countPrestigePoints(points), countAges(at), deeper)
}

// CanPrestige returns true if the player has reached PrestigeMinAge or later,
// comparing orders from the ageOrder map provided by ProgressManager.
func (pm *PrestigeManager) CanPrestige(age string, ageOrder map[string]int) bool {
	idx, ok := ageOrder[age]
	if !ok {
		return false
	}
	minIdx, ok := ageOrder[PrestigeMinAge]
	if !ok || idx < minIdx {
		return false
	}
	if pm.floor != "" {
		floorIdx, found := ageOrder[pm.floor]
		if found {
			lastIdx := -1
			for _, order := range ageOrder {
				lastIdx = max(lastIdx, order)
			}
			if floorIdx == lastIdx { // A cleared game can start its next run over.
				return true
			}
		}
		return found && idx > floorIdx
	}
	return true
}

// Floor is the furthest age reached before the current run.
func (pm *PrestigeManager) Floor() string { return pm.floor }

// Prestige increments level and adds points
func (pm *PrestigeManager) Prestige(points int) {
	pm.level++
	pm.totalEarned += points
	pm.available += points
}

// ProductionMultiplier is the permanent all-production multiplier earned
// through every prestige point, spent or unspent.
func (pm *PrestigeManager) ProductionMultiplier() float64 {
	return 1 + float64(pm.prestigeCoefficient*math.Sqrt(float64(max(pm.totalEarned, 0))))
}

// NextProductionMultiplier is the multiplier a prestige from age would give.
func (pm *PrestigeManager) NextProductionMultiplier(age string) float64 {
	return 1 + float64(pm.prestigeCoefficient*math.Sqrt(float64(max(pm.totalEarned+pm.CalculatePoints(age), 0))))
}

// Owns reports whether the shop item key (a legacy kit item) is bought.
// Retired perks never count.
func (pm *PrestigeManager) Owns(key string) bool {
	return pm.level > 0 && (key == config.LegacyPlan || key == config.LegacyWorkers || key == config.LegacyFactions)
}

// Modifiers emits the permanent production multiplier every prestige point
// earns, attributed to Source "prestige".
func (pm *PrestigeManager) Modifiers() []Modifier {
	return []Modifier{{Source: "prestige", Target: "prestige_production", Op: OpMul, Value: pm.ProductionMultiplier()}}
}

// Snapshot returns a PrestigeState for UI consumption
func (pm *PrestigeManager) Snapshot() PrestigeState {
	return PrestigeState{
		Level:                pm.level,
		Floor:                pm.floor,
		TotalEarned:          pm.totalEarned,
		Available:            pm.available,
		Kit:                  pm.kitState(),
		ProductionMultiplier: pm.ProductionMultiplier(),
	}
}

// LoadState restores prestige state from save data
func (pm *PrestigeManager) LoadState(level, totalEarned, available int) {
	pm.level = level
	pm.totalEarned = totalEarned
	pm.available = available
}

// GetLevel returns the current prestige level
func (pm *PrestigeManager) GetLevel() int {
	return pm.level
}
