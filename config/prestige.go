package config

// PrestigeUpgradeDef defines a persistent upgrade purchasable with prestige points.
// Upgrades survive across full resets (DoPrestige/Succumb) — they are the primary
// permanent progression mechanism.
//
// EffectType semantics:
//
//	"rate_bonus"        — multiplier added to EffectKey rate (PerTier = fractional, e.g. 0.05 = +5%)
//	"flat_bonus"        — flat value added to EffectKey cap (PerTier = absolute units)
//	"starting_resource" — added to the player's starting amount of EffectKey resource on reset
type PrestigeUpgradeDef struct {
	Key         string
	Name        string
	Description string
	EffectKey   string  // engine bonus key (e.g. "gather_rate", "population") or resource key for starting_resource
	EffectType  string  // "rate_bonus", "flat_bonus", or "starting_resource"
	PerTier     float64 // bonus increment per tier (meaning depends on EffectType)
	MaxTier     int     // maximum purchasable tier; upgrade is "maxed" when Tier == MaxTier
	Costs       []int   // prestige point cost at each tier; len must equal MaxTier
}

// PrestigeUpgrades returns all prestige shop upgrades
func PrestigeUpgrades() []PrestigeUpgradeDef {
	return []PrestigeUpgradeDef{
		{
			Key: "gather_boost", Name: "Gather Boost",
			Description: "+5% gather rate per tier",
			EffectKey:   "gather_rate", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{2, 3, 4, 6, 8},
		},
		{
			Key: "storage_bonus", Name: "Storage Bonus",
			Description: "+20 all storage per tier",
			EffectKey:   "all", EffectType: "flat_bonus",
			PerTier: 20, MaxTier: 5,
			Costs: []int{2, 3, 4, 6, 8},
		},
		{
			Key: "research_speed", Name: "Research Speed",
			Description: "+5% knowledge rate per tier",
			EffectKey:   "knowledge_rate", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{2, 3, 5, 8, 10},
		},
		{
			Key: "military_power", Name: "Military Power",
			Description: "+5% military power per tier",
			EffectKey:   "military_power", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{2, 3, 5, 8, 10},
		},
		{
			Key: "starting_food", Name: "Starting Food",
			Description: "+25 starting food per tier",
			EffectKey:   "food", EffectType: "starting_resource",
			PerTier: 25, MaxTier: 5,
			Costs: []int{1, 2, 3, 4, 5},
		},
		{
			Key: "starting_wood", Name: "Starting Wood",
			Description: "+25 starting wood per tier",
			EffectKey:   "wood", EffectType: "starting_resource",
			PerTier: 25, MaxTier: 5,
			Costs: []int{1, 2, 3, 4, 5},
		},
		{
			Key: "population_cap", Name: "Population Cap",
			Description: "+2 population cap per tier",
			EffectKey:   "population", EffectType: "flat_bonus",
			PerTier: 2, MaxTier: 5,
			Costs: []int{2, 3, 5, 8, 10},
		},
		{
			Key: "expedition_loot", Name: "Expedition Loot",
			Description: "+5% expedition reward per tier",
			EffectKey:   "expedition_reward", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{2, 3, 5, 8, 10},
		},
		{
			Key: "tick_speed", Name: "Temporal Mastery",
			Description: "Game ticks 5% faster per tier",
			EffectKey:   "tick_speed", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{6, 10, 17, 23, 33},
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
