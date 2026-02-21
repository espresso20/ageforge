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
		{From: "manor", To: "apartment", CostScale: 0.25, MinAge: "industrial_age"},
		{From: "apartment", To: "skyscraper", CostScale: 0.25, MinAge: "modern_age"},
		{From: "skyscraper", To: "neon_tower", CostScale: 0.25, MinAge: "cyberpunk_age"},
		{From: "neon_tower", To: "orbital_habitat", CostScale: 0.25, MinAge: "space_age"},

		// Storage chain: stash → storage_pit → warehouse → classical_vault → ...
		{From: "stash", To: "storage_pit", CostScale: 0.25, MinAge: "stone_age"},
		{From: "storage_pit", To: "warehouse", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "warehouse", To: "classical_vault", CostScale: 0.25, MinAge: "classical_age"},
		{From: "classical_vault", To: "industrial_depot", CostScale: 0.25, MinAge: "industrial_age"},
		{From: "industrial_depot", To: "modern_depot", CostScale: 0.25, MinAge: "modern_age"},
		{From: "modern_depot", To: "info_vault", CostScale: 0.25, MinAge: "information_age"},
		{From: "info_vault", To: "digital_archive", CostScale: 0.25, MinAge: "digital_age"},
		{From: "digital_archive", To: "cyber_vault", CostScale: 0.25, MinAge: "cyberpunk_age"},
		{From: "cyber_vault", To: "fusion_vault", CostScale: 0.25, MinAge: "fusion_age"},
		{From: "fusion_vault", To: "orbital_depot", CostScale: 0.25, MinAge: "space_age"},
		{From: "orbital_depot", To: "stellar_vault", CostScale: 0.25, MinAge: "interstellar_age"},
		{From: "stellar_vault", To: "galactic_vault", CostScale: 0.25, MinAge: "galactic_age"},
		{From: "galactic_vault", To: "quantum_vault", CostScale: 0.25, MinAge: "quantum_age"},

		// Knowledge chain: altar → firepit → library → university
		{From: "altar", To: "firepit", CostScale: 0.25, MinAge: "stone_age"},
		{From: "firepit", To: "library", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "library", To: "university", CostScale: 0.25, MinAge: "medieval_age"},

		// Resource production chains
		{From: "gathering_camp", To: "farm", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "woodcutter_camp", To: "lumber_mill", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "stone_pit", To: "quarry", CostScale: 0.25, MinAge: "bronze_age"},
		{From: "mine", To: "smithy", CostScale: 0.25, MinAge: "iron_age"},
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
