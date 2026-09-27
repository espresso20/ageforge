package config

// BuildingUpgradeDef defines an upgrade path from one building to another
type BuildingUpgradeDef struct {
	From      string  // old building key
	To        string  // new building key
	CostScale float64 // fraction of To's base cost (e.g. 0.25 = 25%)
	MinAge    string  // age key when upgrade becomes available
}

// BuildingUpgrades returns all building upgrade chain definitions
func BuildingUpgrades() []BuildingUpgradeDef {
	return []BuildingUpgradeDef{
		// Housing chain: hut → house → manor → apartment → skyscraper → neon_tower → orbital_habitat
		{From: "hut", To: "house", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "house", To: "manor", CostScale: 0.25, MinAge: "medieval_age"},
		{From: "manor", To: "tenement", CostScale: 0.25, MinAge: "industrial_age"},
		{From: "tenement", To: "tower_block", CostScale: 0.25, MinAge: "modern_age"},
		{From: "tower_block", To: "arcology_pod", CostScale: 0.25, MinAge: "cyberpunk_age"},
		{From: "arcology_pod", To: "orbital_habitat", CostScale: 0.25, MinAge: "space_age"},

		// No storage chain: storage never transforms (decision log). Upgrading a
		// stash spent a capped storage_pit slot on a copy the age lock never lets
		// you rebuild, which lowered the most you could ever store.

		// Knowledge chain: story_circle → scriptorium → library → university
		{From: "story_circle", To: "scriptorium", CostScale: 0.25, MinAge: "stone_age"},
		{From: "scriptorium", To: "library", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "library", To: "university", CostScale: 0.25, MinAge: "medieval_age"},

		// Resource production chains
		{From: "gathering_camp", To: "farm", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "woodcutter_camp", To: "lumber_mill", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "stone_pit", To: "quarry", CostScale: 0.25, MinAge: "bronze_age"},
	}
}

// UpgradesFromKey returns a map of fromKey -> BuildingUpgradeDef for quick lookup
func UpgradesFromKey() map[string]BuildingUpgradeDef {
	m := make(map[string]BuildingUpgradeDef)
	for _, u := range BuildingUpgrades() {
		m[u.From] = u
	}
	return m
}
