package game

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// PrestigeManager manages the prestige meta-progression layer.
// Players prestige from the Medieval Age on (PrestigeMinAge), earning depth
// points: every age the run completed pays 3^epoch (config.DepthPoints).
// Points buy the legacy kit (legacy.go), which carries a run's automation
// across the reset; the first shop's nine perks are retired and refunded
// (refundShopLocked).
//
// It also holds Era Mastery (mastery.go): each age's mastery, the record
// (the deepest age ever entered) and this run's furthest age. The manager
// survives prestige and Succumb and is replaced only by Reset, so mastery
// lasts as long as the account's game does. The old passive (+2% production
// and +1% tick speed per level) retired into mastery.
type PrestigeManager struct {
	level       int
	totalEarned int
	available   int
	upgrades    map[string]int // upgrade key -> tier purchased (0 = not bought)

	// Era Mastery: age key -> mastery (ages at 0 left out), the record and
	// this run's furthest age ("" for none yet), and whether a save from
	// before Era Mastery has had its one-time seeding (seedMasteryLocked).
	// speeds is every age's k, rebuilt whenever mastery or the record
	// changes (rebuildSpeeds).
	mastery       map[string]int
	record        string
	runFurthest   string
	masterySeeded bool
	speeds        map[string]float64

	// shopVersion is the shop a save was written under (0 or 1: the first
	// shop; config.PrestigeShopVersion: the legacy kit). Below the current
	// version, a load refunds the retired perks once (refundShopLocked).
	shopVersion int

	// The legacy kit's memory (legacy.go): what the runs so far wrote and
	// did, kept whether or not the kit is bought, so an item bought after a
	// prestige works at once. legacyPlan is the plan template, by age;
	// legacyResearch the research order; legacyFactions the civilizations
	// met; legacyShares the worker shares (nil: every domain on auto).
	legacyPlan     []PlanTemplateItem
	legacyResearch []string
	legacyFactions []string
	legacyShares   map[string]float64

	// upgradeList / upgradeDefs are the static shop table, built once so
	// GetBonuses (hit every tick via the resolver) and Snapshot (every UI
	// refresh) don't rebuild config per call. Read-only.
	upgradeList []config.PrestigeUpgradeDef
	upgradeDefs map[string]config.PrestigeUpgradeDef
}

// NewPrestigeManager creates a new prestige manager
func NewPrestigeManager() *PrestigeManager {
	list := config.PrestigeUpgrades()
	defs := make(map[string]config.PrestigeUpgradeDef, len(list))
	for _, def := range list {
		defs[def.Key] = def
	}
	pm := &PrestigeManager{
		upgrades:    make(map[string]int),
		upgradeList: list,
		upgradeDefs: defs,
		mastery:     make(map[string]int),
		// A new game has no past prestiges to seed mastery from, and it
		// starts in the first age: its record and its run's furthest age.
		masterySeeded: true,
		record:        ageKeys()[0],
		runFurthest:   ageKeys()[0],
		// A new game starts on the current shop: nothing to refund.
		shopVersion: config.PrestigeShopVersion,
	}
	pm.rebuildSpeeds()
	return pm
}

// CalculatePoints is what a prestige from age pays: its depth points
// (config.DepthPoints), the sum of 3^epoch over every age the run
// completed. No divisor and no milestone, tech or building terms: a deeper
// run pays more, and a level costs nothing.
func (pm *PrestigeManager) CalculatePoints(age string) int {
	return config.DepthPoints(age)
}

// PrestigeMinAge is the age that opens prestige; every later age counts too.
// A prestige from here to the Atomic Age is an early taste: it pays little
// (the Medieval Age 9 points, the Modern Age 120).
const PrestigeMinAge = "medieval_age"

// PrestigeRunAge is the age from which a prestige counts as a full run
// rather than a taste, for the account's records (badges count only these).
const PrestigeRunAge = "modern_age"

// CanPrestige returns true if the player has reached PrestigeMinAge or later,
// comparing orders from the ageOrder map provided by ProgressManager.
func (pm *PrestigeManager) CanPrestige(age string, ageOrder map[string]int) bool {
	idx, ok := ageOrder[age]
	if !ok {
		return false
	}
	minIdx, ok := ageOrder[PrestigeMinAge]
	return ok && idx >= minIdx
}

// Prestige increments level and adds points
func (pm *PrestigeManager) Prestige(points int) {
	pm.level++
	pm.totalEarned += points
	pm.available += points
}

// BuyUpgrade purchases the next tier of an upgrade. Returns error if can't
// afford, maxed or retired.
func (pm *PrestigeManager) BuyUpgrade(key string) error {
	def, ok := pm.upgradeDefs[key]
	if !ok {
		active := map[string]bool{}
		for _, d := range pm.upgradeList {
			if !d.Retired {
				active[d.Key] = true
			}
		}
		return unknownKeyError("prestige upgrade", key, active, "Type prestige shop to see the upgrades.")
	}
	if def.Retired {
		return fmt.Errorf("%s was part of the old prestige shop and can't be bought any more; its points were refunded. Type prestige shop to see the legacy kit.", def.Name)
	}

	currentTier := pm.upgrades[key]
	if currentTier >= def.MaxTier {
		if def.MaxTier == 1 {
			return fmt.Errorf("You already own %s.", def.Name)
		}
		return fmt.Errorf("%s is already at its top tier (%d).", def.Name, def.MaxTier)
	}

	cost := def.Costs[currentTier]
	if pm.available < cost {
		if def.MaxTier == 1 {
			return fmt.Errorf("%s costs %s (you have %s). Prestige again to earn more.", def.Name, textfmt.Count(cost, "prestige point", "prestige points"), textfmt.Int(pm.available))
		}
		return fmt.Errorf("%s tier %d costs %s (you have %s). Prestige again to earn more.", def.Name, currentTier+1, textfmt.Count(cost, "prestige point", "prestige points"), textfmt.Int(pm.available))
	}

	pm.available -= cost
	pm.upgrades[key] = currentTier + 1
	return nil
}

// Owns reports whether the shop item key (a legacy kit item) is bought.
// Retired perks never count.
func (pm *PrestigeManager) Owns(key string) bool {
	def, ok := pm.upgradeDefs[key]
	return ok && !def.Retired && pm.upgrades[key] >= 1
}

// GetBonuses returns the bought upgrades' bonuses as a bonus map. The old
// passive (+2% production and +1% tick speed per level) retired into Era
// Mastery (mastery.go), which speeds up the ages a run completed instead.
func (pm *PrestigeManager) GetBonuses() map[string]float64 {
	bonuses := make(map[string]float64)

	// Upgrade bonuses (rate and flat bonuses, not starting resources)
	// List order, not map order: float sums must be the same every run.
	for _, def := range pm.upgradeList {
		tier := pm.upgrades[def.Key]
		if tier <= 0 || def.Retired {
			continue
		}
		if def.EffectType == "rate_bonus" || def.EffectType == "flat_bonus" {
			bonuses[def.EffectKey] += float64(def.PerTier * float64(tier))
		}
	}

	return bonuses
}

// Modifiers emits one OpAdd Modifier per (target, value) returned by GetBonuses,
// attributed to Source "prestige". Targets match the engine's current strings
// (e.g. "production_all", "tick_speed", "<res>_rate") — no renaming.
func (pm *PrestigeManager) Modifiers() []Modifier {
	bonuses := pm.GetBonuses()
	out := make([]Modifier, 0, len(bonuses))
	for t, v := range bonuses {
		out = append(out, Modifier{Source: "prestige", Target: t, Op: OpAdd, Value: v})
	}
	return out
}

// GetStartingResources returns bonus starting resources from prestige upgrades
func (pm *PrestigeManager) GetStartingResources() map[string]float64 {
	resources := make(map[string]float64)
	// List order, not map order: float sums must be the same every run.
	for _, def := range pm.upgradeList {
		tier := pm.upgrades[def.Key]
		if tier <= 0 || def.Retired {
			continue
		}
		if def.EffectType == "starting_resource" {
			resources[def.EffectKey] += float64(def.PerTier * float64(tier))
		}
	}
	return resources
}

// Snapshot returns a PrestigeState for UI consumption
func (pm *PrestigeManager) Snapshot() PrestigeState {
	upgrades := make(map[string]PrestigeUpgradeState)

	for _, def := range pm.upgradeList {
		tier := pm.upgrades[def.Key]
		nextCost := 0
		if tier < def.MaxTier {
			nextCost = def.Costs[tier]
		}
		if def.Retired {
			nextCost = 0
		}
		upgrades[def.Key] = PrestigeUpgradeState{
			Name:        def.Name,
			Description: def.Description,
			Tier:        tier,
			MaxTier:     def.MaxTier,
			NextCost:    nextCost,
			Effect:      formatPrestigeEffect(def, tier),
			Retired:     def.Retired,
			Kit:         def.EffectType == "legacy",
		}
	}
	return PrestigeState{
		Level:       pm.level,
		TotalEarned: pm.totalEarned,
		Available:   pm.available,
		Upgrades:    upgrades,
		ShopVersion: pm.shopVersion,
		Kit:         pm.kitState(),
	}
}

// LoadState restores prestige state from save data
func (pm *PrestigeManager) LoadState(level, totalEarned, available int, upgrades map[string]int) {
	pm.level = level
	pm.totalEarned = totalEarned
	pm.available = available
	if upgrades != nil {
		pm.upgrades = upgrades
	}
}

// GetLevel returns the current prestige level
func (pm *PrestigeManager) GetLevel() int {
	return pm.level
}

func formatPrestigeEffect(def config.PrestigeUpgradeDef, tier int) string {
	if def.Retired {
		return "Retired: refunded with the old shop"
	}
	if tier == 0 {
		return "Not purchased"
	}
	if def.EffectType == "legacy" {
		return "Owned"
	}
	v := def.PerTier * float64(tier)
	switch def.EffectType {
	case "rate_bonus":
		return textfmt.SignedPercent(v) + " " + EffectTargetName(def.EffectKey)
	case "flat_bonus":
		// A flat "all" bonus raises every resource's storage (see the
		// storage bonuses in engine.go), not production.
		if def.EffectKey == "all" {
			return textfmt.Signed(v) + " storage for every resource"
		}
		return textfmt.Signed(v) + " " + EffectTargetName(def.EffectKey)
	case "starting_resource":
		return textfmt.Signed(v) + " starting " + ResourceName(def.EffectKey)
	}
	return ""
}
