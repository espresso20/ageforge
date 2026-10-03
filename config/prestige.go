package config

// PrestigeUpgradeDef defines a persistent upgrade purchasable with prestige points.
// Upgrades survive across full resets (DoPrestige/Succumb) — they are the primary
// permanent progression mechanism. Keys are saved: never rename or remove one.
//
// EffectType semantics:
//
//	"rate_bonus"        — multiplier added to EffectKey rate (PerTier = fractional, e.g. 0.05 = +5%)
//	"flat_bonus"        — flat value added to EffectKey cap (PerTier = absolute units)
//	"starting_resource" — added to the player's starting amount of EffectKey resource on reset
//	"legacy"            — a legacy kit item: carries EffectKey's automation across a prestige
type PrestigeUpgradeDef struct {
	Key         string
	Name        string
	Description string
	EffectKey   string  // engine bonus key (e.g. "gather_rate", "population") or resource key for starting_resource
	EffectType  string  // "rate_bonus", "flat_bonus", "starting_resource" or "legacy"
	PerTier     float64 // bonus increment per tier (meaning depends on EffectType)
	MaxTier     int     // maximum purchasable tier; upgrade is "maxed" when Tier == MaxTier
	Costs       []int   // prestige point cost at each tier; len must equal MaxTier
	// Retired marks a perk of the first shop (shop version 1). It is hidden,
	// has no effect and can't be bought. Its Costs stay frozen: the refund
	// (ShopRefund) prices what a save spent on it with them.
	Retired bool
}

// PrestigeShopVersion is the shop's version. A save below it gets the
// one-time refund on load: version 1 is the first shop (nine perks bought
// in tiers); version 2 (Pacing v2) retired them for the legacy kit.
const PrestigeShopVersion = 2

// The legacy kit's keys (Pacing v2). Saved in the upgrades map: permanent.
const (
	LegacyPlan     = "legacy_plan"
	LegacyResearch = "legacy_research"
	LegacyWorkers  = "legacy_workers"
	LegacyFactions = "legacy_factions"
)

// LegacyKit lists the kit's keys in shop order (cheapest first).
func LegacyKit() []string {
	return []string{LegacyPlan, LegacyResearch, LegacyWorkers, LegacyFactions}
}

// PrestigeUpgrades returns every prestige shop upgrade, retired ones included.
func PrestigeUpgrades() []PrestigeUpgradeDef {
	return []PrestigeUpgradeDef{
		// The first shop's nine perks: retired by shop version 2 and
		// refunded. Kept for their keys and their frozen costs.
		{
			Key: "gather_boost", Name: "Gather Boost",
			Description: "+5% worker output per tier",
			EffectKey:   "gather_rate", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs:   []int{2, 3, 4, 6, 8},
			Retired: true,
		},
		{
			Key: "storage_bonus", Name: "Storage Bonus",
			Description: "+20 storage for every resource per tier",
			EffectKey:   "all", EffectType: "flat_bonus",
			PerTier: 20, MaxTier: 5,
			Costs:   []int{2, 3, 4, 6, 8},
			Retired: true,
		},
		{
			Key: "research_speed", Name: "Knowledge Production",
			Description: "+5% knowledge production per tier",
			EffectKey:   "knowledge_rate", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs:   []int{2, 3, 5, 8, 10},
			Retired: true,
		},
		{
			Key: "military_power", Name: "Military Power",
			Description: "+5% military power per tier",
			EffectKey:   "military_power", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs:   []int{2, 3, 5, 8, 10},
			Retired: true,
		},
		{
			Key: "starting_food", Name: "Starting Food",
			Description: "+25 starting food per tier",
			EffectKey:   "food", EffectType: "starting_resource",
			PerTier: 25, MaxTier: 5,
			Costs:   []int{1, 2, 3, 4, 5},
			Retired: true,
		},
		{
			Key: "starting_wood", Name: "Starting Wood",
			Description: "+25 starting wood per tier",
			EffectKey:   "wood", EffectType: "starting_resource",
			PerTier: 25, MaxTier: 5,
			Costs:   []int{1, 2, 3, 4, 5},
			Retired: true,
		},
		{
			Key: "population_cap", Name: "Housing Bonus",
			Description: "+2 housing per tier",
			EffectKey:   "population", EffectType: "flat_bonus",
			PerTier: 2, MaxTier: 5,
			Costs:   []int{2, 3, 5, 8, 10},
			Retired: true,
		},
		{
			Key: "expedition_loot", Name: "Expedition Loot",
			Description: "+5% expedition rewards per tier",
			EffectKey:   "expedition_reward", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs:   []int{2, 3, 5, 8, 10},
			Retired: true,
		},
		{
			Key: "tick_speed", Name: "Temporal Mastery",
			Description: "+5% game speed per tier",
			EffectKey:   "tick_speed", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs:   []int{6, 10, 17, 23, 33},
			Retired: true,
		},
		// The legacy kit (Pacing v2): four one-tier items that carry a run's
		// automation across a prestige. 117 points in all: a Medieval Age
		// prestige (9) buys the plan template, a first Modern Age run (120)
		// the rest.
		{
			Key: LegacyPlan, Name: "Plan Template",
			Description: "Your build plan carries over: each age's part of the plan you wrote is added again when you enter that age",
			EffectKey:   "plan", EffectType: "legacy",
			MaxTier: 1, Costs: []int{9},
		},
		{
			Key: LegacyResearch, Name: "Research Memory",
			Description: "While nothing is being researched, the next tech in your last run's research order starts by itself",
			EffectKey:   "research", EffectType: "legacy",
			MaxTier: 1, Costs: []int{18},
		},
		{
			Key: LegacyWorkers, Name: "Worker Shares",
			Description: "Your worker shares carry over to each new run",
			EffectKey:   "workers", EffectType: "legacy",
			MaxTier: 1, Costs: []int{36},
		},
		{
			Key: LegacyFactions, Name: "Old Friends",
			Description: "Civilizations you have met are met again as soon as your age reaches theirs, at neutral opinion",
			EffectKey:   "factions", EffectType: "legacy",
			MaxTier: 1, Costs: []int{54},
		},
	}
}

// PrestigeUpgradeByKey returns a map of key -> PrestigeUpgradeDef
func PrestigeUpgradeByKey() map[string]PrestigeUpgradeDef {
	m := make(map[string]PrestigeUpgradeDef)
	for _, u := range PrestigeUpgrades() {
		m[u.Key] = u
	}
	return m
}

// ActivePrestigeUpgrades returns the upgrades the shop sells (retired ones
// left out), in shop order.
func ActivePrestigeUpgrades() []PrestigeUpgradeDef {
	var out []PrestigeUpgradeDef
	for _, u := range PrestigeUpgrades() {
		if !u.Retired {
			out = append(out, u)
		}
	}
	return out
}

// ===== Depth points (Pacing v2) =====

// DepthWeight is what completing age pays at prestige: 3^epoch, counting
// the Stone Era as epoch 0. That is 1 for each Stone Era age, 3 Iron, 9
// Steel, 27 Electric, 81 Digital, 243 Neon and 729 Cosmic; 0 for an
// unknown age. The 22 ages weigh 4,008 in all.
func DepthWeight(age string) int {
	for _, ep := range Epochs() {
		for _, a := range ep.Ages {
			if a != age {
				continue
			}
			w := 1
			for i := 0; i < ep.Order; i++ {
				w *= 3
			}
			return w
		}
	}
	return 0
}

// DepthPoints is what a prestige from age pays: the sum of DepthWeight over
// every age before it (the ages the run completed), with no divisor. The
// Medieval Age pays 9, the Modern Age 120, the Information Age 201, the
// Cyberpunk Age (a run through the Digital Age) 363 and the Interstellar Age
// (through the Space Age) 1,092. 0 for an unknown age.
func DepthPoints(age string) int {
	total := 0
	for _, a := range AgeOrder() {
		if a == age {
			return total
		}
		total += DepthWeight(a)
	}
	return 0
}

// Shop refund (shop version 1 to 2). Each past prestige counts as one Modern
// Age run under the depth formula (RefundPerPrestige points), or the old
// points, spent and unspent, times RefundRateNum / RefundRateDen (120 new
// points for the 27 a first Modern Age run paid before), whichever is more.
const (
	RefundPerPrestige = 120
	RefundRateNum     = 120
	RefundRateDen     = 27
)

// ShopRefund is the refund for a save at shop version 1: the points it spent
// on the retired perks (at their frozen costs) plus the points it had left
// are its old points, converted at the new rate. Returns both.
func ShopRefund(level, available int, upgrades map[string]int) (old, refund int) {
	old = max(available, 0)
	for _, def := range PrestigeUpgrades() {
		if !def.Retired {
			continue
		}
		tier := min(max(upgrades[def.Key], 0), def.MaxTier)
		for t := 0; t < tier; t++ {
			old += def.Costs[t]
		}
	}
	refund = max(level, 0) * RefundPerPrestige
	if conv := old * RefundRateNum / RefundRateDen; conv > refund {
		refund = conv
	}
	return old, refund
}
