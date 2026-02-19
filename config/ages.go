package config

// AgeDef defines an age/era in the game
type AgeDef struct {
	Name  string
	Key   string
	Order int
	// Requirements to advance TO this age
	ResourceReqs map[string]float64
	BuildingReqs map[string]int
	// What this age unlocks
	UnlockBuildings []string
	UnlockResources []string
	UnlockVillagers []string
	Description     string
}

// Ages returns all ages in order
func Ages() []AgeDef {
	return []AgeDef{
		// ~30-60 min active play to advance
		{
			Name: "Primitive Age", Key: "primitive_age", Order: 0,
			Description:     "Survival. Nothing but your hands and wits.",
			UnlockBuildings: []string{"hut"},
			UnlockResources: []string{"food", "wood"},
			UnlockVillagers: []string{"worker"},
		},
		// ~2-4 days idle to advance
		{
			Name: "Stone Age", Key: "stone_age", Order: 1,
			Description:  "Tools of stone change everything.",
			ResourceReqs: map[string]float64{"food": 100, "wood": 75},
			BuildingReqs: map[string]int{"hut": 3},
			UnlockBuildings: []string{"gathering_camp", "woodcutter_camp", "stone_pit", "firepit", "storage_pit"},
			UnlockResources: []string{"stone", "knowledge"},
		},
		// ~1-2 weeks idle to advance
		{
			Name: "Bronze Age", Key: "bronze_age", Order: 2,
			Description:  "Discovery of metalworking changes everything.",
			ResourceReqs: map[string]float64{"food": 500, "stone": 300, "knowledge": 100},
			BuildingReqs: map[string]int{"hut": 6, "stone_pit": 3, "firepit": 2},
			UnlockBuildings: []string{"farm", "lumber_mill", "quarry", "mine", "market", "library", "house", "warehouse", "stonehenge"},
			UnlockResources: []string{"iron", "gold"},
			UnlockVillagers: []string{"scholar"},
		},
		// ~3-4 weeks idle to advance
		{
			Name: "Iron Age", Key: "iron_age", Order: 3,
			Description:  "Iron tools and weapons transform society.",
			ResourceReqs: map[string]float64{"stone": 2000, "iron": 500, "knowledge": 400},
			BuildingReqs: map[string]int{"mine": 4, "lumber_mill": 3, "library": 2},
			UnlockBuildings: []string{"coal_mine", "smithy", "barracks", "granary", "colosseum"},
			UnlockResources: []string{"coal"},
			UnlockVillagers: []string{"soldier"},
		},
		// ~1-2 months idle to advance
		{
			Name: "Medieval Age", Key: "medieval_age", Order: 4,
			Description:  "Kingdoms rise and feudalism takes hold.",
			ResourceReqs: map[string]float64{"stone": 5000, "iron": 2000, "gold": 800, "knowledge": 1500},
			BuildingReqs: map[string]int{"market": 3, "library": 4, "barracks": 2},
			UnlockBuildings: []string{"cathedral", "manor", "university", "castle", "great_library"},
			UnlockResources: []string{"steel", "faith"},
			UnlockVillagers: []string{"merchant"},
		},
		// ~2-3 months idle to advance
		{
			Name: "Renaissance Age", Key: "renaissance_age", Order: 5,
			Description:  "Art, science, and exploration flourish.",
			ResourceReqs: map[string]float64{"gold": 5000, "knowledge": 8000, "steel": 1000},
			BuildingReqs: map[string]int{"university": 3, "market": 5, "castle": 1},
			UnlockBuildings: []string{"art_studio", "bank", "observatory"},
			UnlockResources: []string{"culture"},
		},
		// ~3-4 months idle to advance
		{
			Name: "Industrial Age", Key: "industrial_age", Order: 6,
			Description:  "Machines revolutionize production.",
			ResourceReqs: map[string]float64{"steel": 5000, "gold": 15000, "knowledge": 20000},
			BuildingReqs: map[string]int{"smithy": 6, "bank": 3, "observatory": 2},
			UnlockBuildings: []string{"factory", "oil_well", "apartment"},
			UnlockResources: []string{"oil"},
		},
		// ~6+ months idle to reach
		{
			Name: "Modern Age", Key: "modern_age", Order: 7,
			Description:  "Technology and innovation define the era.",
			ResourceReqs: map[string]float64{"gold": 50000, "knowledge": 50000, "oil": 5000, "steel": 10000},
			BuildingReqs: map[string]int{"factory": 5, "oil_well": 4, "apartment": 3},
			UnlockBuildings: []string{"power_plant", "research_lab", "skyscraper", "space_program"},
			UnlockResources: []string{"electricity"},
		},
	}
}

// AgeByKey returns a map of key -> AgeDef
func AgeByKey() map[string]AgeDef {
	m := make(map[string]AgeDef)
	for _, a := range Ages() {
		m[a.Key] = a
	}
	return m
}

// AgeOrder returns an ordered list of age keys
func AgeOrder() []string {
	ages := Ages()
	keys := make([]string, len(ages))
	for i, a := range ages {
		keys[i] = a.Key
	}
	return keys
}
