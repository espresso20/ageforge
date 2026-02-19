package config

// Effect represents a game effect from a building or tech
type Effect struct {
	Type   string  // "production", "capacity", "unlock", "bonus", "storage"
	Target string  // resource key, building key, etc.
	Value  float64 // amount per tick, capacity increase, multiplier, etc.
}

// BuildingDef defines a building type
type BuildingDef struct {
	Name         string
	Key          string
	Category     string // "production", "housing", "research", "military", "storage", "wonder"
	BaseCost     map[string]float64
	CostScale    float64 // each subsequent costs CostScale * previous
	Effects      []Effect
	BuildTicks   int    // 0 = instant
	RequiredAge  string // minimum age key
	RequiredTech string // required tech key (empty = none)
	MaxCount     int    // 0 = unlimited
	Description  string
}

// BaseBuildings returns all building definitions
func BaseBuildings() []BuildingDef {
	return []BuildingDef{
		// ===== STONE AGE =====
		{
			Name: "Hut", Key: "hut", Category: "housing",
			BaseCost:  map[string]float64{"wood": 10},
			CostScale: 1.3,
			Effects:   []Effect{{Type: "capacity", Target: "population", Value: 2}},
			RequiredAge: "stone_age",
			Description: "A crude shelter of sticks and leaves. +2 pop cap.",
		},
		{
			Name: "Gathering Camp", Key: "gathering_camp", Category: "production",
			BaseCost:  map[string]float64{"wood": 8},
			CostScale: 1.25,
			Effects:   []Effect{{Type: "production", Target: "food", Value: 0.2}},
			RequiredAge: "stone_age",
			Description: "Foragers collect berries and roots. +0.2 food/tick.",
		},
		{
			Name: "Woodcutter's Camp", Key: "woodcutter_camp", Category: "production",
			BaseCost:  map[string]float64{"wood": 5, "stone": 3},
			CostScale: 1.25,
			Effects:   []Effect{{Type: "production", Target: "wood", Value: 0.15}},
			RequiredAge: "stone_age",
			Description: "Choppers fell trees with stone axes. +0.15 wood/tick.",
		},
		{
			Name: "Stone Pit", Key: "stone_pit", Category: "production",
			BaseCost:  map[string]float64{"wood": 8, "stone": 5},
			CostScale: 1.3,
			Effects:   []Effect{{Type: "production", Target: "stone", Value: 0.1}},
			RequiredAge: "stone_age",
			Description: "A shallow dig site for rocks. +0.1 stone/tick.",
		},
		{
			Name: "Firepit", Key: "firepit", Category: "research",
			BaseCost:  map[string]float64{"wood": 10, "stone": 5},
			CostScale: 1.35,
			Effects:   []Effect{{Type: "production", Target: "knowledge", Value: 0.1}},
			RequiredAge: "stone_age",
			Description: "Elders share stories by the fire. +0.1 knowledge/tick.",
		},
		{
			Name: "Storage Pit", Key: "storage_pit", Category: "storage",
			BaseCost:  map[string]float64{"wood": 10, "stone": 8},
			CostScale: 1.35,
			Effects:   []Effect{{Type: "storage", Target: "all", Value: 50}},
			RequiredAge: "stone_age",
			Description: "A hole in the ground to stash things. +50 storage.",
		},

		// ===== BRONZE AGE =====
		{
			Name: "Farm", Key: "farm", Category: "production",
			BaseCost:  map[string]float64{"wood": 25, "stone": 15},
			CostScale: 1.3,
			Effects:   []Effect{{Type: "production", Target: "food", Value: 0.5}},
			RequiredAge: "bronze_age",
			Description: "Cultivated fields produce steady food. +0.5 food/tick.",
		},
		{
			Name: "Lumber Mill", Key: "lumber_mill", Category: "production",
			BaseCost:  map[string]float64{"wood": 30, "stone": 15, "iron": 5},
			CostScale: 1.3,
			Effects:   []Effect{{Type: "production", Target: "wood", Value: 0.4}},
			RequiredAge: "bronze_age",
			Description: "Bronze saws process wood efficiently. +0.4 wood/tick.",
		},
		{
			Name: "Quarry", Key: "quarry", Category: "production",
			BaseCost:  map[string]float64{"wood": 25, "stone": 20, "iron": 5},
			CostScale: 1.3,
			Effects:   []Effect{{Type: "production", Target: "stone", Value: 0.3}},
			RequiredAge: "bronze_age",
			Description: "Organized stone extraction. +0.3 stone/tick.",
		},
		{
			Name: "Mine", Key: "mine", Category: "production",
			BaseCost:  map[string]float64{"wood": 30, "stone": 25},
			CostScale: 1.35,
			Effects:   []Effect{{Type: "production", Target: "iron", Value: 0.25}},
			RequiredAge: "bronze_age",
			Description: "Digs deep for metal ore. +0.25 iron/tick.",
		},
		{
			Name: "Market", Key: "market", Category: "production",
			BaseCost:  map[string]float64{"wood": 40, "stone": 30, "iron": 10},
			CostScale: 1.4,
			Effects:   []Effect{{Type: "production", Target: "gold", Value: 0.2}},
			RequiredAge: "bronze_age",
			Description: "Trade goods for coin. +0.2 gold/tick.",
		},
		{
			Name: "Library", Key: "library", Category: "research",
			BaseCost:  map[string]float64{"wood": 35, "stone": 20, "gold": 5},
			CostScale: 1.35,
			Effects:   []Effect{{Type: "production", Target: "knowledge", Value: 0.3}},
			RequiredAge: "bronze_age",
			Description: "Scribes record and study. +0.3 knowledge/tick.",
		},
		{
			Name: "House", Key: "house", Category: "housing",
			BaseCost:  map[string]float64{"wood": 25, "stone": 20, "iron": 5},
			CostScale: 1.35,
			Effects:   []Effect{{Type: "capacity", Target: "population", Value: 5}},
			RequiredAge: "bronze_age",
			Description: "Sturdy brick dwelling. +5 pop cap.",
		},
		{
			Name: "Warehouse", Key: "warehouse", Category: "storage",
			BaseCost:  map[string]float64{"wood": 30, "stone": 25, "iron": 5},
			CostScale: 1.4,
			Effects:   []Effect{{Type: "storage", Target: "all", Value: 150}},
			RequiredAge: "bronze_age",
			Description: "Proper storage building. +150 storage.",
		},

		// ===== IRON AGE =====
		{
			Name: "Coal Mine", Key: "coal_mine", Category: "production",
			BaseCost:  map[string]float64{"wood": 40, "stone": 35, "iron": 15},
			CostScale: 1.35,
			Effects:   []Effect{{Type: "production", Target: "coal", Value: 0.3}},
			RequiredAge: "iron_age",
			Description: "Extracts coal. +0.3 coal/tick.",
		},
		{
			Name: "Smithy", Key: "smithy", Category: "production",
			BaseCost:  map[string]float64{"stone": 40, "iron": 25, "coal": 10},
			CostScale: 1.4,
			Effects:   []Effect{{Type: "production", Target: "steel", Value: 0.2}},
			RequiredAge: "iron_age",
			Description: "Forges steel from iron and coal. +0.2 steel/tick.",
		},
		{
			Name: "Barracks", Key: "barracks", Category: "military",
			BaseCost:  map[string]float64{"wood": 50, "stone": 40, "iron": 20},
			CostScale: 1.4,
			Effects:   []Effect{{Type: "capacity", Target: "military", Value: 10}},
			RequiredAge: "iron_age",
			Description: "Trains soldiers. +10 military cap.",
		},
		{
			Name: "Granary", Key: "granary", Category: "storage",
			BaseCost:  map[string]float64{"wood": 35, "stone": 25},
			CostScale: 1.35,
			Effects: []Effect{
				{Type: "storage", Target: "food", Value: 200},
			},
			RequiredAge: "iron_age",
			Description: "Stores extra food. +200 food storage.",
		},

		// ===== MEDIEVAL AGE =====
		{
			Name: "Cathedral", Key: "cathedral", Category: "production",
			BaseCost:  map[string]float64{"stone": 100, "gold": 30, "iron": 20},
			CostScale: 1.5,
			Effects:   []Effect{{Type: "production", Target: "faith", Value: 0.4}},
			RequiredAge: "medieval_age",
			Description: "Generates faith. +0.4 faith/tick.",
		},
		{
			Name: "Manor", Key: "manor", Category: "housing",
			BaseCost:  map[string]float64{"wood": 60, "stone": 50, "iron": 15},
			CostScale: 1.4,
			Effects:   []Effect{{Type: "capacity", Target: "population", Value: 12}},
			RequiredAge: "medieval_age",
			Description: "Large estate. +12 pop cap.",
		},
		{
			Name: "University", Key: "university", Category: "research",
			BaseCost:  map[string]float64{"stone": 60, "gold": 25, "knowledge": 30},
			CostScale: 1.45,
			Effects:   []Effect{{Type: "production", Target: "knowledge", Value: 1.5}},
			RequiredAge: "medieval_age",
			Description: "Advanced learning. +1.5 knowledge/tick.",
		},
		{
			Name: "Castle", Key: "castle", Category: "military",
			BaseCost:  map[string]float64{"stone": 120, "iron": 40, "gold": 20},
			CostScale: 1.5,
			Effects:   []Effect{{Type: "capacity", Target: "military", Value: 25}},
			RequiredAge: "medieval_age",
			MaxCount:    3,
			Description: "Stronghold. +25 military cap. Max 3.",
		},

		// ===== RENAISSANCE AGE =====
		{
			Name: "Art Studio", Key: "art_studio", Category: "production",
			BaseCost:  map[string]float64{"wood": 50, "gold": 40, "knowledge": 20},
			CostScale: 1.4,
			Effects:   []Effect{{Type: "production", Target: "culture", Value: 0.5}},
			RequiredAge: "renaissance_age",
			Description: "Creates cultural works. +0.5 culture/tick.",
		},
		{
			Name: "Bank", Key: "bank", Category: "production",
			BaseCost:  map[string]float64{"stone": 60, "gold": 50, "iron": 20},
			CostScale: 1.45,
			Effects:   []Effect{{Type: "production", Target: "gold", Value: 1.0}},
			RequiredAge: "renaissance_age",
			Description: "Advanced finance. +1.0 gold/tick.",
		},
		{
			Name: "Observatory", Key: "observatory", Category: "research",
			BaseCost:  map[string]float64{"stone": 70, "gold": 35, "knowledge": 40},
			CostScale: 1.5,
			Effects:   []Effect{{Type: "production", Target: "knowledge", Value: 2.0}},
			RequiredAge: "renaissance_age",
			MaxCount:    3,
			Description: "Studies the stars. +2.0 knowledge/tick. Max 3.",
		},

		// ===== INDUSTRIAL AGE =====
		{
			Name: "Factory", Key: "factory", Category: "production",
			BaseCost:  map[string]float64{"steel": 40, "coal": 30, "iron": 50},
			CostScale: 1.45,
			Effects: []Effect{
				{Type: "production", Target: "iron", Value: 2.0},
				{Type: "production", Target: "steel", Value: 0.5},
			},
			RequiredAge: "industrial_age",
			Description: "Mass production. +2.0 iron, +0.5 steel/tick.",
		},
		{
			Name: "Oil Well", Key: "oil_well", Category: "production",
			BaseCost:  map[string]float64{"steel": 30, "iron": 40, "gold": 50},
			CostScale: 1.4,
			Effects:   []Effect{{Type: "production", Target: "oil", Value: 0.4}},
			RequiredAge: "industrial_age",
			Description: "Extracts oil. +0.4 oil/tick.",
		},
		{
			Name: "Apartment", Key: "apartment", Category: "housing",
			BaseCost:  map[string]float64{"steel": 25, "stone": 60, "iron": 30},
			CostScale: 1.4,
			Effects:   []Effect{{Type: "capacity", Target: "population", Value: 25}},
			RequiredAge: "industrial_age",
			Description: "Dense housing. +25 pop cap.",
		},

		// ===== MODERN AGE =====
		{
			Name: "Power Plant", Key: "power_plant", Category: "production",
			BaseCost:  map[string]float64{"steel": 60, "oil": 30, "gold": 80},
			CostScale: 1.5,
			Effects:   []Effect{{Type: "production", Target: "electricity", Value: 1.0}},
			RequiredAge: "modern_age",
			Description: "Generates electricity. +1.0 electricity/tick.",
		},
		{
			Name: "Research Lab", Key: "research_lab", Category: "research",
			BaseCost:  map[string]float64{"steel": 50, "gold": 60, "electricity": 20},
			CostScale: 1.5,
			Effects:   []Effect{{Type: "production", Target: "knowledge", Value: 4.0}},
			RequiredAge: "modern_age",
			Description: "Cutting-edge research. +4.0 knowledge/tick.",
		},
		{
			Name: "Skyscraper", Key: "skyscraper", Category: "housing",
			BaseCost:  map[string]float64{"steel": 80, "gold": 50, "electricity": 15},
			CostScale: 1.5,
			Effects:   []Effect{{Type: "capacity", Target: "population", Value: 50}},
			RequiredAge: "modern_age",
			Description: "Massive housing. +50 pop cap.",
		},

		// ===== WONDERS (one per age, unique) =====
		{
			Name: "Stonehenge", Key: "stonehenge", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 150, "wood": 80},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 2.0},
				{Type: "production", Target: "faith", Value: 0.5},
			},
			RequiredAge: "bronze_age",
			MaxCount:    1,
			BuildTicks:  10,
			Description: "Ancient monument. +2.0 knowledge, +0.5 faith/tick.",
		},
		{
			Name: "Colosseum", Key: "colosseum", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 200, "iron": 60, "gold": 40},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "capacity", Target: "population", Value: 20},
				{Type: "production", Target: "culture", Value: 1.0},
			},
			RequiredAge: "iron_age",
			MaxCount:    1,
			BuildTicks:  15,
			Description: "Grand arena. +20 pop cap, +1.0 culture/tick.",
		},
		{
			Name: "Great Library", Key: "great_library", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 250, "gold": 100, "knowledge": 80},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 5.0},
			},
			RequiredAge: "medieval_age",
			MaxCount:    1,
			BuildTicks:  20,
			Description: "Repository of all knowledge. +5.0 knowledge/tick.",
		},
		{
			Name: "Space Program", Key: "space_program", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 200, "gold": 300, "electricity": 100, "knowledge": 200},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 10.0},
				{Type: "production", Target: "culture", Value: 5.0},
			},
			RequiredAge: "modern_age",
			MaxCount:    1,
			BuildTicks:  30,
			Description: "The ultimate achievement. +10 knowledge, +5 culture/tick.",
		},
	}
}

// BuildingByKey returns a map of key -> BuildingDef
func BuildingByKey() map[string]BuildingDef {
	m := make(map[string]BuildingDef)
	for _, b := range BaseBuildings() {
		m[b.Key] = b
	}
	return m
}
