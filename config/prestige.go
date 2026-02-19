package config

// PrestigeUpgradeDef defines a prestige shop upgrade
type PrestigeUpgradeDef struct {
	Key         string
	Name        string
	Description string
	EffectKey   string    // bonus key applied to engine (e.g., "gather_rate")
	EffectType  string    // "rate_bonus", "flat_bonus", "starting_resource"
	PerTier     float64   // value added per tier
	MaxTier     int
	Costs       []int     // cost at each tier (len == MaxTier)
}

// PrestigeUpgrades returns all prestige shop upgrades
func PrestigeUpgrades() []PrestigeUpgradeDef {
	return []PrestigeUpgradeDef{
		{
			Key: "gather_boost", Name: "Gather Boost",
			Description: "+5% gather rate per tier",
			EffectKey: "gather_rate", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{2, 4, 8, 15, 25},
		},
		{
			Key: "storage_bonus", Name: "Storage Bonus",
			Description: "+20 all storage per tier",
			EffectKey: "all", EffectType: "flat_bonus",
			PerTier: 20, MaxTier: 5,
			Costs: []int{2, 4, 8, 15, 25},
		},
		{
			Key: "research_speed", Name: "Research Speed",
			Description: "+5% knowledge rate per tier",
			EffectKey: "knowledge_rate", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{3, 6, 12, 20, 35},
		},
		{
			Key: "military_power", Name: "Military Power",
			Description: "+5% military power per tier",
			EffectKey: "military_power", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{3, 6, 12, 20, 35},
		},
		{
			Key: "starting_food", Name: "Starting Food",
			Description: "+25 starting food per tier",
			EffectKey: "food", EffectType: "starting_resource",
			PerTier: 25, MaxTier: 5,
			Costs: []int{1, 2, 4, 8, 15},
		},
		{
			Key: "starting_wood", Name: "Starting Wood",
			Description: "+25 starting wood per tier",
			EffectKey: "wood", EffectType: "starting_resource",
			PerTier: 25, MaxTier: 5,
			Costs: []int{1, 2, 4, 8, 15},
		},
		{
			Key: "population_cap", Name: "Population Cap",
			Description: "+2 population cap per tier",
			EffectKey: "population", EffectType: "flat_bonus",
			PerTier: 2, MaxTier: 5,
			Costs: []int{3, 6, 12, 20, 35},
		},
		{
			Key: "expedition_loot", Name: "Expedition Loot",
			Description: "+5% expedition reward per tier",
			EffectKey: "expedition_reward", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{3, 6, 12, 20, 35},
		},
		{
			Key: "tick_speed", Name: "Temporal Mastery",
			Description: "Game ticks 5% faster per tier",
			EffectKey: "tick_speed", EffectType: "rate_bonus",
			PerTier: 0.05, MaxTier: 5,
			Costs: []int{5, 10, 20, 35, 50},
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
