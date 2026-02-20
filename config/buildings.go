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
// Cost scaling: each age's buildings cost ~5x the previous age
func BaseBuildings() []BuildingDef {
	return []BuildingDef{
		// ===== PRIMITIVE AGE (costs: 30-100) =====
		{
			Name: "Hut", Key: "hut", Category: "housing",
			BaseCost:    map[string]float64{"wood": 30},
			CostScale:   1.3,
			Effects:     []Effect{{Type: "capacity", Target: "population", Value: 2}},
			BuildTicks:  3,
			RequiredAge: "primitive_age",
			Description: "A crude shelter of sticks and leaves. +2 pop cap.",
		},
		{
			Name: "Stash", Key: "stash", Category: "storage",
			BaseCost:    map[string]float64{"wood": 50},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 100}},
			BuildTicks:  3,
			RequiredAge: "primitive_age",
			Description: "A hidden pile of supplies. +100 storage.",
		},
		{
			Name: "Altar", Key: "altar", Category: "research",
			BaseCost:    map[string]float64{"wood": 300},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.01}},
			BuildTicks:  3,
			RequiredAge: "primitive_age",
			Description: "A sacred stone circle where shamans commune with spirits. +0.01 knowledge/tick.",
		},

		// ===== STONE AGE (costs: 200-1000) =====
		{
			Name: "Gathering Camp", Key: "gathering_camp", Category: "production",
			BaseCost:    map[string]float64{"wood": 200},
			CostScale:   1.25,
			Effects:     []Effect{{Type: "production", Target: "food", Value: 0.1}},
			BuildTicks:  5,
			RequiredAge: "stone_age",
			Description: "Foragers collect berries and roots. +0.1 food/tick.",
		},
		{
			Name: "Woodcutter's Camp", Key: "woodcutter_camp", Category: "production",
			BaseCost:    map[string]float64{"wood": 500, "stone": 300},
			CostScale:   1.25,
			Effects:     []Effect{{Type: "production", Target: "wood", Value: 0.08}},
			BuildTicks:  5,
			RequiredAge: "stone_age",
			Description: "Choppers fell trees with stone axes. +0.08 wood/tick.",
		},
		{
			Name: "Stone Pit", Key: "stone_pit", Category: "production",
			BaseCost:    map[string]float64{"wood": 800, "stone": 500},
			CostScale:   1.3,
			Effects:     []Effect{{Type: "production", Target: "stone", Value: 0.05}},
			BuildTicks:  5,
			RequiredAge: "stone_age",
			Description: "A shallow dig site for rocks. +0.05 stone/tick.",
		},
		{
			Name: "Firepit", Key: "firepit", Category: "research",
			BaseCost:    map[string]float64{"wood": 300, "stone": 200},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.02}},
			BuildTicks:  5,
			RequiredAge: "stone_age",
			Description: "Elders share stories by the fire. +0.02 knowledge/tick.",
		},
		{
			Name: "Storage Pit", Key: "storage_pit", Category: "storage",
			BaseCost:    map[string]float64{"wood": 1000, "stone": 800},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 500}},
			BuildTicks:  5,
			RequiredAge: "stone_age",
			Description: "A hole in the ground to stash things. +500 storage.",
		},

		// ===== BRONZE AGE (costs: 1500-5000) =====
		{
			Name: "Farm", Key: "farm", Category: "production",
			BaseCost:    map[string]float64{"wood": 1500, "stone": 900},
			CostScale:   1.3,
			Effects:     []Effect{{Type: "production", Target: "food", Value: 0.25}},
			BuildTicks:  8,
			RequiredAge: "bronze_age",
			Description: "Cultivated fields produce steady food. +0.25 food/tick.",
		},
		{
			Name: "Lumber Mill", Key: "lumber_mill", Category: "production",
			BaseCost:    map[string]float64{"wood": 2000, "stone": 1000, "iron": 300},
			CostScale:   1.3,
			Effects:     []Effect{{Type: "production", Target: "wood", Value: 0.2}},
			BuildTicks:  8,
			RequiredAge: "bronze_age",
			Description: "Bronze saws process wood efficiently. +0.2 wood/tick.",
		},
		{
			Name: "Quarry", Key: "quarry", Category: "production",
			BaseCost:    map[string]float64{"wood": 1500, "stone": 1200, "iron": 300},
			CostScale:   1.3,
			Effects:     []Effect{{Type: "production", Target: "stone", Value: 0.15}},
			BuildTicks:  8,
			RequiredAge: "bronze_age",
			Description: "Organized stone extraction. +0.15 stone/tick.",
		},
		{
			Name: "Mine", Key: "mine", Category: "production",
			BaseCost:    map[string]float64{"wood": 2000, "stone": 1500},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.12}},
			BuildTicks:  8,
			RequiredAge: "bronze_age",
			Description: "Digs deep for metal ore. +0.12 iron/tick.",
		},
		{
			Name: "Market", Key: "market", Category: "production",
			BaseCost:    map[string]float64{"wood": 2500, "stone": 2000, "iron": 600},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.1}},
			BuildTicks:  8,
			RequiredAge: "bronze_age",
			Description: "Trade goods for coin. +0.1 gold/tick.",
		},
		{
			Name: "Library", Key: "library", Category: "research",
			BaseCost:    map[string]float64{"wood": 2200, "stone": 1200, "gold": 300},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.06}},
			BuildTicks:  8,
			RequiredAge: "bronze_age",
			Description: "Scribes record and study. +0.06 knowledge/tick.",
		},
		{
			Name: "House", Key: "house", Category: "housing",
			BaseCost:    map[string]float64{"wood": 1500, "stone": 1200, "iron": 300},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "capacity", Target: "population", Value: 5}},
			BuildTicks:  8,
			RequiredAge: "bronze_age",
			Description: "Sturdy brick dwelling. +5 pop cap.",
		},
		{
			Name: "Warehouse", Key: "warehouse", Category: "storage",
			BaseCost:    map[string]float64{"wood": 2000, "stone": 1500, "iron": 300},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 3000}},
			BuildTicks:  8,
			RequiredAge: "bronze_age",
			Description: "Proper storage building. +3000 storage.",
		},

		// ===== IRON AGE (costs: 8k-25k) =====
		{
			Name: "Coal Mine", Key: "coal_mine", Category: "production",
			BaseCost:    map[string]float64{"wood": 10000, "stone": 8000, "iron": 3000},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "production", Target: "coal", Value: 0.15}},
			BuildTicks:  12,
			RequiredAge: "iron_age",
			Description: "Extracts coal. +0.15 coal/tick.",
		},
		{
			Name: "Smithy", Key: "smithy", Category: "production",
			BaseCost:    map[string]float64{"stone": 10000, "iron": 6000, "coal": 2000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "steel", Value: 0.1}},
			BuildTicks:  12,
			RequiredAge: "iron_age",
			Description: "Forges steel from iron and coal. +0.1 steel/tick.",
		},
		{
			Name: "Barracks", Key: "barracks", Category: "military",
			BaseCost:    map[string]float64{"wood": 12000, "stone": 10000, "iron": 5000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "capacity", Target: "military", Value: 10}},
			BuildTicks:  12,
			RequiredAge: "iron_age",
			Description: "Trains soldiers. +10 military cap.",
		},
		{
			Name: "Granary", Key: "granary", Category: "storage",
			BaseCost:    map[string]float64{"wood": 8000, "stone": 6000},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 12000}},
			BuildTicks:  12,
			RequiredAge: "iron_age",
			Description: "Organized supply storage. +12000 storage.",
		},

		// ===== CLASSICAL AGE (costs: 40k-120k) =====
		{
			Name: "Forum", Key: "forum", Category: "production",
			BaseCost:  map[string]float64{"stone": 60000, "gold": 20000, "iron": 15000},
			CostScale: 1.4,
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 0.25},
				{Type: "production", Target: "knowledge", Value: 0.06},
			},
			BuildTicks:  15,
			RequiredAge: "classical_age",
			Description: "Center of civic life. +0.25 gold, +0.06 knowledge/tick.",
		},
		{
			Name: "Aqueduct", Key: "aqueduct", Category: "production",
			BaseCost:    map[string]float64{"stone": 80000, "iron": 20000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "food", Value: 0.5}},
			BuildTicks:  15,
			RequiredAge: "classical_age",
			Description: "Water infrastructure boosts food. +0.5 food/tick.",
		},
		{
			Name: "Amphitheater", Key: "amphitheater", Category: "production",
			BaseCost:    map[string]float64{"stone": 70000, "gold": 15000, "wood": 30000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "culture", Value: 0.15}},
			BuildTicks:  15,
			RequiredAge: "classical_age",
			Description: "Drama and performance. +0.15 culture/tick.",
		},
		{
			Name: "Classical Vault", Key: "classical_vault", Category: "storage",
			BaseCost:    map[string]float64{"stone": 50000, "iron": 12000, "gold": 10000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 25000}},
			BuildTicks:  15,
			RequiredAge: "classical_age",
			Description: "Stone vault for valuables. +25000 storage.",
		},

		// ===== MEDIEVAL AGE (costs: 200k-600k) =====
		{
			Name: "Cathedral", Key: "cathedral", Category: "production",
			BaseCost:    map[string]float64{"stone": 300000, "gold": 90000, "iron": 60000},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.2}},
			BuildTicks:  20,
			RequiredAge: "medieval_age",
			Description: "Generates faith. +0.2 faith/tick.",
		},
		{
			Name: "Manor", Key: "manor", Category: "housing",
			BaseCost:    map[string]float64{"wood": 180000, "stone": 150000, "iron": 45000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "capacity", Target: "population", Value: 12}},
			BuildTicks:  20,
			RequiredAge: "medieval_age",
			Description: "Large estate. +12 pop cap.",
		},
		{
			Name: "University", Key: "university", Category: "research",
			BaseCost:    map[string]float64{"stone": 180000, "gold": 75000, "knowledge": 90000},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.3}},
			BuildTicks:  20,
			RequiredAge: "medieval_age",
			Description: "Advanced learning. +0.3 knowledge/tick.",
		},
		{
			Name: "Castle", Key: "castle", Category: "military",
			BaseCost:    map[string]float64{"stone": 360000, "iron": 120000, "gold": 60000},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "capacity", Target: "military", Value: 25}},
			BuildTicks:  20,
			RequiredAge: "medieval_age",
			MaxCount:    3,
			Description: "Stronghold. +25 military cap. Max 3.",
		},
		{
			Name: "Keep", Key: "keep", Category: "storage",
			BaseCost:    map[string]float64{"stone": 200000, "iron": 60000, "gold": 40000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 60000}},
			BuildTicks:  20,
			RequiredAge: "medieval_age",
			Description: "Fortified storehouse. +60000 storage.",
		},

		// ===== RENAISSANCE AGE (costs: 1M-3M) =====
		{
			Name: "Art Studio", Key: "art_studio", Category: "production",
			BaseCost:    map[string]float64{"wood": 400000, "gold": 200000, "knowledge": 100000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "culture", Value: 0.25}},
			BuildTicks:  25,
			RequiredAge: "renaissance_age",
			Description: "Creates cultural works. +0.25 culture/tick.",
		},
		{
			Name: "Bank", Key: "bank", Category: "production",
			BaseCost:    map[string]float64{"stone": 500000, "gold": 300000, "iron": 150000},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.5}},
			BuildTicks:  25,
			RequiredAge: "renaissance_age",
			Description: "Advanced finance. +0.5 gold/tick.",
		},
		{
			Name: "Observatory", Key: "observatory", Category: "research",
			BaseCost:    map[string]float64{"stone": 600000, "gold": 250000, "knowledge": 200000},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.4}},
			BuildTicks:  25,
			RequiredAge: "renaissance_age",
			MaxCount:    3,
			Description: "Studies the stars. +0.4 knowledge/tick. Max 3.",
		},
		{
			Name: "Renaissance Vault", Key: "renaissance_vault", Category: "storage",
			BaseCost:    map[string]float64{"stone": 450000, "gold": 200000, "iron": 100000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 500000}},
			BuildTicks:  25,
			RequiredAge: "renaissance_age",
			Description: "Ornate storage facility. +500000 storage.",
		},

		// ===== COLONIAL AGE (costs: 5M-15M) =====
		{
			Name: "Colony", Key: "colony", Category: "production",
			BaseCost:  map[string]float64{"wood": 2e6, "gold": 1.5e6, "steel": 500000},
			CostScale: 1.45,
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 1.0},
				{Type: "production", Target: "gold", Value: 0.75},
			},
			BuildTicks:  30,
			RequiredAge: "colonial_age",
			Description: "Overseas colony. +1.0 food, +0.75 gold/tick.",
		},
		{
			Name: "Port", Key: "port", Category: "production",
			BaseCost:    map[string]float64{"wood": 1.5e6, "stone": 1e6, "gold": 800000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "gold", Value: 1.0}},
			BuildTicks:  30,
			RequiredAge: "colonial_age",
			Description: "Maritime trade hub. +1.0 gold/tick.",
		},
		{
			Name: "Plantation", Key: "plantation", Category: "production",
			BaseCost:    map[string]float64{"wood": 1.2e6, "gold": 600000, "iron": 300000},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "production", Target: "food", Value: 1.5}},
			BuildTicks:  30,
			RequiredAge: "colonial_age",
			Description: "Large-scale farming. +1.5 food/tick.",
		},
		{
			Name: "Colonial Warehouse", Key: "colonial_warehouse", Category: "storage",
			BaseCost:    map[string]float64{"wood": 1.5e6, "stone": 1e6, "gold": 600000},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 10e6}},
			BuildTicks:  30,
			RequiredAge: "colonial_age",
			Description: "Trade goods warehouse. +10M storage.",
		},

		// ===== INDUSTRIAL AGE (costs: 25M-75M) =====
		{
			Name: "Factory", Key: "factory", Category: "production",
			BaseCost:  map[string]float64{"steel": 20e6, "coal": 15e6, "iron": 25e6},
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
			BaseCost:    map[string]float64{"steel": 15e6, "iron": 20e6, "gold": 25e6},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "oil", Value: 0.4}},
			RequiredAge: "industrial_age",
			Description: "Extracts oil. +0.4 oil/tick.",
		},
		{
			Name: "Apartment", Key: "apartment", Category: "housing",
			BaseCost:    map[string]float64{"steel": 12e6, "stone": 30e6, "iron": 15e6},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "capacity", Target: "population", Value: 25}},
			RequiredAge: "industrial_age",
			Description: "Dense housing. +25 pop cap.",
		},
		{
			Name: "Industrial Depot", Key: "industrial_depot", Category: "storage",
			BaseCost:    map[string]float64{"steel": 15e6, "iron": 20e6, "coal": 10e6},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 50e6}},
			RequiredAge: "industrial_age",
			Description: "Industrial-scale storage. +50M storage.",
		},

		// ===== VICTORIAN AGE (costs: 125M-375M) =====
		{
			Name: "Power Grid", Key: "power_grid", Category: "production",
			BaseCost:    map[string]float64{"steel": 150e6, "coal": 100e6, "gold": 125e6},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "production", Target: "electricity", Value: 0.8}},
			RequiredAge: "victorian_age",
			Description: "Steam-powered electrical generation. +0.8 electricity/tick.",
		},
		{
			Name: "Telegraph", Key: "telegraph", Category: "research",
			BaseCost:    map[string]float64{"steel": 100e6, "gold": 75e6, "iron": 60e6},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 1.2}},
			RequiredAge: "victorian_age",
			Description: "Long-distance communication. +1.2 knowledge/tick.",
		},
		{
			Name: "Clocktower", Key: "clocktower", Category: "production",
			BaseCost:  map[string]float64{"steel": 90e6, "gold": 110e6, "stone": 125e6},
			CostScale: 1.4,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.05},
			},
			RequiredAge: "victorian_age",
			MaxCount:    5,
			Description: "Precision timekeeping boosts efficiency. +5% all production. Max 5.",
		},
		{
			Name: "Victorian Vault", Key: "victorian_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 125e6, "gold": 100e6, "iron": 75e6},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 350e6}},
			RequiredAge: "victorian_age",
			Description: "Reinforced vault. +350M storage.",
		},

		// ===== ELECTRIC AGE (costs: 600M-2B) =====
		{
			Name: "Electric Mill", Key: "electric_mill", Category: "production",
			BaseCost:  map[string]float64{"steel": 1e9, "electricity": 250e6, "iron": 600e6},
			CostScale: 1.45,
			Effects: []Effect{
				{Type: "production", Target: "steel", Value: 1.5},
				{Type: "production", Target: "iron", Value: 3.0},
			},
			RequiredAge: "electric_age",
			Description: "Electric-powered manufacturing. +1.5 steel, +3.0 iron/tick.",
		},
		{
			Name: "Telephone Exchange", Key: "telephone_exchange", Category: "research",
			BaseCost:    map[string]float64{"steel": 750e6, "electricity": 200e6, "gold": 500e6},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 2.0}},
			RequiredAge: "electric_age",
			Description: "Connected communication network. +2.0 knowledge/tick.",
		},
		{
			Name: "Train Station", Key: "train_station", Category: "production",
			BaseCost:  map[string]float64{"steel": 900e6, "coal": 600e6, "gold": 450e6},
			CostScale: 1.4,
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 4.0},
				{Type: "storage", Target: "all", Value: 100e6},
			},
			RequiredAge: "electric_age",
			Description: "Rail transport hub. +4.0 gold/tick, +100M storage.",
		},
		{
			Name: "Electric Warehouse", Key: "electric_warehouse", Category: "storage",
			BaseCost:    map[string]float64{"steel": 750e6, "electricity": 125e6, "iron": 500e6},
			CostScale:   1.4,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 3.5e9}},
			RequiredAge: "electric_age",
			Description: "Climate-controlled storage. +3.5B storage.",
		},

		// ===== ATOMIC AGE (costs: 3B-10B) =====
		{
			Name: "Reactor", Key: "reactor", Category: "production",
			BaseCost:  map[string]float64{"steel": 7.5e9, "electricity": 2.5e9, "gold": 5e9},
			CostScale: 1.5,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 5.0},
				{Type: "production", Target: "uranium", Value: 0.3},
			},
			RequiredAge: "atomic_age",
			Description: "Nuclear reactor. +5.0 electricity, +0.3 uranium/tick.",
		},
		{
			Name: "Bunker", Key: "bunker", Category: "military",
			BaseCost:    map[string]float64{"steel": 6e9, "stone": 10e9, "iron": 4e9},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "capacity", Target: "military", Value: 50}},
			RequiredAge: "atomic_age",
			Description: "Fortified underground shelter. +50 military cap.",
		},
		{
			Name: "Missile Silo", Key: "missile_silo", Category: "military",
			BaseCost:    map[string]float64{"steel": 10e9, "uranium": 500e6, "gold": 7.5e9},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "bonus", Target: "military_power", Value: 0.3}},
			RequiredAge: "atomic_age",
			MaxCount:    5,
			Description: "Nuclear deterrent. +30% military power. Max 5.",
		},
		{
			Name: "Atomic Vault", Key: "atomic_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 5e9, "stone": 7.5e9, "iron": 3e9},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 15e9}},
			RequiredAge: "atomic_age",
			Description: "Radiation-shielded storage. +15B storage.",
		},

		// ===== MODERN AGE (costs: 15B-50B) =====
		{
			Name: "Power Plant", Key: "power_plant", Category: "production",
			BaseCost:    map[string]float64{"steel": 30e9, "oil": 15e9, "gold": 40e9},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "production", Target: "electricity", Value: 8.0}},
			RequiredAge: "modern_age",
			Description: "Advanced power generation. +8.0 electricity/tick.",
		},
		{
			Name: "Research Lab", Key: "research_lab", Category: "research",
			BaseCost:    map[string]float64{"steel": 25e9, "gold": 30e9, "electricity": 10e9},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 3.0}},
			RequiredAge: "modern_age",
			Description: "Cutting-edge research. +3.0 knowledge/tick.",
		},
		{
			Name: "Skyscraper", Key: "skyscraper", Category: "housing",
			BaseCost:    map[string]float64{"steel": 40e9, "gold": 25e9, "electricity": 8e9},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "capacity", Target: "population", Value: 50}},
			RequiredAge: "modern_age",
			Description: "Massive housing. +50 pop cap.",
		},
		{
			Name: "Modern Depot", Key: "modern_depot", Category: "storage",
			BaseCost:    map[string]float64{"steel": 35e9, "gold": 25e9, "electricity": 8e9},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 45e9}},
			RequiredAge: "modern_age",
			Description: "Automated logistics center. +45B storage.",
		},

		// ===== INFORMATION AGE (costs: 75B-250B) =====
		{
			Name: "Server Farm", Key: "server_farm", Category: "production",
			BaseCost:  map[string]float64{"steel": 125e9, "electricity": 60e9, "gold": 100e9},
			CostScale: 1.5,
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 2.0},
				{Type: "production", Target: "knowledge", Value: 2.0},
			},
			RequiredAge: "information_age",
			Description: "Data processing center. +2.0 data, +2.0 knowledge/tick.",
		},
		{
			Name: "Fiber Hub", Key: "fiber_hub", Category: "production",
			BaseCost:    map[string]float64{"steel": 100e9, "gold": 75e9, "electricity": 40e9},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "production", Target: "data", Value: 3.0}},
			RequiredAge: "information_age",
			Description: "High-speed network infrastructure. +3.0 data/tick.",
		},
		{
			Name: "Media Center", Key: "media_center", Category: "production",
			BaseCost:  map[string]float64{"steel": 90e9, "gold": 60e9, "data": 1.25e9},
			CostScale: 1.4,
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 3.0},
				{Type: "production", Target: "gold", Value: 5.0},
			},
			RequiredAge: "information_age",
			Description: "Digital entertainment. +3.0 culture, +5.0 gold/tick.",
		},
		{
			Name: "Info Vault", Key: "info_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 100e9, "electricity": 40e9, "data": 625e6},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 250e9}},
			RequiredAge: "information_age",
			Description: "Digital-physical storage hybrid. +250B storage.",
		},

		// ===== DIGITAL AGE (costs: 400B-1.2T) =====
		{
			Name: "Data Center", Key: "data_center", Category: "production",
			BaseCost:    map[string]float64{"steel": 750e9, "electricity": 400e9, "data": 10e9},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "production", Target: "data", Value: 8.0}},
			RequiredAge: "digital_age",
			Description: "Massive data processing. +8.0 data/tick.",
		},
		{
			Name: "AI Lab", Key: "ai_lab", Category: "research",
			BaseCost:  map[string]float64{"steel": 600e9, "data": 15e9, "electricity": 250e9},
			CostScale: 1.5,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 6.0},
				{Type: "production", Target: "data", Value: 3.0},
			},
			RequiredAge: "digital_age",
			Description: "Artificial intelligence research. +6.0 knowledge, +3.0 data/tick.",
		},
		{
			Name: "Smart Grid", Key: "smart_grid", Category: "production",
			BaseCost:  map[string]float64{"steel": 500e9, "electricity": 200e9, "data": 7.5e9},
			CostScale: 1.45,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 12.0},
				{Type: "bonus", Target: "production_all", Value: 0.03},
			},
			RequiredAge: "digital_age",
			Description: "AI-optimized power grid. +12.0 electricity/tick, +3% all production.",
		},
		{
			Name: "Digital Archive", Key: "digital_archive", Category: "storage",
			BaseCost:    map[string]float64{"steel": 500e9, "data": 10e9, "electricity": 150e9},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 1.5e12}},
			RequiredAge: "digital_age",
			Description: "Quantum-encrypted storage. +1.5T storage.",
		},

		// ===== CYBERPUNK AGE (costs: 2T-6T) =====
		{
			Name: "Augmentation Clinic", Key: "augmentation_clinic", Category: "production",
			BaseCost:  map[string]float64{"steel": 4e12, "data": 100e9, "gold": 3e12},
			CostScale: 1.5,
			Effects: []Effect{
				{Type: "bonus", Target: "gather_rate", Value: 0.1},
				{Type: "production", Target: "crypto", Value: 1.0},
			},
			RequiredAge: "cyberpunk_age",
			Description: "Cybernetic enhancements. +10% gather rate, +1.0 crypto/tick.",
		},
		{
			Name: "Neon Tower", Key: "neon_tower", Category: "housing",
			BaseCost:    map[string]float64{"steel": 3.6e12, "electricity": 2e12, "gold": 2.4e12},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "capacity", Target: "population", Value: 100}},
			RequiredAge: "cyberpunk_age",
			Description: "Towering arcology. +100 pop cap.",
		},
		{
			Name: "Black Market", Key: "black_market", Category: "production",
			BaseCost:  map[string]float64{"gold": 4e12, "data": 80e9, "electricity": 2e12},
			CostScale: 1.45,
			Effects: []Effect{
				{Type: "production", Target: "crypto", Value: 3.0},
				{Type: "production", Target: "gold", Value: 10.0},
			},
			RequiredAge: "cyberpunk_age",
			Description: "Underground economy. +3.0 crypto, +10.0 gold/tick.",
		},
		{
			Name: "Cyber Vault", Key: "cyber_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 3e12, "data": 60e9, "crypto": 10e9},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 5e12}},
			RequiredAge: "cyberpunk_age",
			Description: "Encrypted digital vault. +5T storage.",
		},

		// ===== FUSION AGE (costs: 10T-30T) =====
		{
			Name: "Fusion Reactor", Key: "fusion_reactor", Category: "production",
			BaseCost:  map[string]float64{"steel": 15e12, "electricity": 10e12, "uranium": 2.5e12},
			CostScale: 1.5,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 50.0},
				{Type: "production", Target: "plasma", Value: 2.0},
			},
			RequiredAge: "fusion_age",
			Description: "Clean fusion power. +50.0 electricity, +2.0 plasma/tick.",
		},
		{
			Name: "Plasma Forge", Key: "plasma_forge", Category: "production",
			BaseCost:  map[string]float64{"steel": 12.5e12, "electricity": 7.5e12, "uranium": 3e12},
			CostScale: 1.5,
			Effects: []Effect{
				{Type: "production", Target: "steel", Value: 10.0},
				{Type: "production", Target: "plasma", Value: 1.5},
			},
			RequiredAge: "fusion_age",
			Description: "Plasma-based manufacturing. +10.0 steel, +1.5 plasma/tick.",
		},
		{
			Name: "Maglev Station", Key: "maglev_station", Category: "production",
			BaseCost:  map[string]float64{"steel": 10e12, "electricity": 5e12, "gold": 7.5e12},
			CostScale: 1.45,
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 20.0},
				{Type: "bonus", Target: "production_all", Value: 0.05},
			},
			RequiredAge: "fusion_age",
			Description: "Magnetic levitation transport. +20.0 gold/tick, +5% all production.",
		},
		{
			Name: "Fusion Vault", Key: "fusion_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 10e12, "plasma": 500e9, "electricity": 5e12},
			CostScale:   1.45,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 30e12}},
			RequiredAge: "fusion_age",
			Description: "Plasma-shielded storage. +30T storage.",
		},

		// ===== SPACE AGE (costs: 50T-150T) =====
		{
			Name: "Launch Pad", Key: "launch_pad", Category: "production",
			BaseCost:  map[string]float64{"steel": 80e12, "plasma": 6e12, "electricity": 40e12},
			CostScale: 1.5,
			Effects: []Effect{
				{Type: "production", Target: "titanium", Value: 2.0},
				{Type: "production", Target: "knowledge", Value: 12.0},
			},
			RequiredAge: "space_age",
			Description: "Orbital launch facility. +2.0 titanium, +12.0 knowledge/tick.",
		},
		{
			Name: "Space Station", Key: "space_station", Category: "research",
			BaseCost:  map[string]float64{"titanium": 4e12, "plasma": 8e12, "electricity": 60e12},
			CostScale: 1.55,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 20.0},
				{Type: "production", Target: "data", Value: 20.0},
			},
			RequiredAge: "space_age",
			Description: "Orbital research platform. +20.0 knowledge, +20.0 data/tick.",
		},
		{
			Name: "Orbital Habitat", Key: "orbital_habitat", Category: "housing",
			BaseCost:    map[string]float64{"titanium": 6e12, "steel": 60e12, "plasma": 4e12},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "capacity", Target: "population", Value: 200}},
			RequiredAge: "space_age",
			Description: "Space habitat ring. +200 pop cap.",
		},
		{
			Name: "Orbital Depot", Key: "orbital_depot", Category: "storage",
			BaseCost:    map[string]float64{"steel": 50e12, "plasma": 6e12, "electricity": 30e12},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 200e12}},
			RequiredAge: "space_age",
			Description: "Zero-gravity storage facility. +200T storage.",
		},

		// ===== INTERSTELLAR AGE (costs: 250T-750T) =====
		{
			Name: "Warp Gate", Key: "warp_gate", Category: "production",
			BaseCost:  map[string]float64{"titanium": 100e12, "plasma": 80e12, "electricity": 200e12},
			CostScale: 1.55,
			Effects: []Effect{
				{Type: "production", Target: "dark_matter", Value: 2.0},
				{Type: "bonus", Target: "production_all", Value: 0.08},
			},
			RequiredAge: "interstellar_age",
			Description: "Faster-than-light gate. +2.0 dark matter/tick, +8% all production.",
		},
		{
			Name: "Colony Ship", Key: "colony_ship", Category: "production",
			BaseCost:  map[string]float64{"titanium": 80e12, "dark_matter": 5e12, "steel": 500e12},
			CostScale: 1.5,
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 50.0},
				{Type: "production", Target: "titanium", Value: 5.0},
			},
			RequiredAge: "interstellar_age",
			Description: "Interstellar colonization vessel. +50.0 food, +5.0 titanium/tick.",
		},
		{
			Name: "Star Forge", Key: "star_forge", Category: "production",
			BaseCost:  map[string]float64{"titanium": 120e12, "plasma": 100e12, "dark_matter": 8e12},
			CostScale: 1.55,
			Effects: []Effect{
				{Type: "production", Target: "steel", Value: 50.0},
				{Type: "production", Target: "titanium", Value: 8.0},
			},
			RequiredAge: "interstellar_age",
			Description: "Stellar-powered forge. +50.0 steel, +8.0 titanium/tick.",
		},
		{
			Name: "Stellar Vault", Key: "stellar_vault", Category: "storage",
			BaseCost:    map[string]float64{"titanium": 60e12, "plasma": 50e12, "electricity": 80e12},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 500e12}},
			RequiredAge: "interstellar_age",
			Description: "Pocket-dimension storage. +500T storage.",
		},

		// ===== GALACTIC AGE (costs: 1.25Q-3.75Q) =====
		{
			Name: "Galactic Hub", Key: "galactic_hub", Category: "production",
			BaseCost:  map[string]float64{"dark_matter": 250e12, "titanium": 1e15, "plasma": 500e12},
			CostScale: 1.55,
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 100.0},
				{Type: "production", Target: "knowledge", Value: 40.0},
				{Type: "bonus", Target: "production_all", Value: 0.1},
			},
			RequiredAge: "galactic_age",
			Description: "Galactic trade network. +100 gold, +40 knowledge/tick, +10% all.",
		},
		{
			Name: "Antimatter Plant", Key: "antimatter_plant", Category: "production",
			BaseCost:    map[string]float64{"dark_matter": 200e12, "plasma": 1e15, "electricity": 5e15},
			CostScale:   1.55,
			Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 3.0}},
			RequiredAge: "galactic_age",
			Description: "Produces antimatter from dark energy. +3.0 antimatter/tick.",
		},
		{
			Name: "Megastructure", Key: "megastructure", Category: "production",
			BaseCost:  map[string]float64{"titanium": 1.5e15, "dark_matter": 150e12, "antimatter": 50e12},
			CostScale: 1.6,
			Effects: []Effect{
				{Type: "capacity", Target: "population", Value: 500},
				{Type: "storage", Target: "all", Value: 500e12},
			},
			RequiredAge: "galactic_age",
			MaxCount:    5,
			Description: "Massive orbital structure. +500 pop cap, +500T storage. Max 5.",
		},
		{
			Name: "Galactic Vault", Key: "galactic_vault", Category: "storage",
			BaseCost:    map[string]float64{"dark_matter": 100e12, "titanium": 500e12, "plasma": 200e12},
			CostScale:   1.5,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 2e15}},
			RequiredAge: "galactic_age",
			Description: "Galaxy-spanning storage network. +2Q storage.",
		},

		// ===== QUANTUM AGE (costs: 6Q-20Q) =====
		{
			Name: "Quantum Computer", Key: "quantum_computer", Category: "research",
			BaseCost:  map[string]float64{"antimatter": 2.5e15, "dark_matter": 4e15, "titanium": 1e15},
			CostScale: 1.6,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 200.0},
				{Type: "production", Target: "quantum_flux", Value: 2.0},
			},
			RequiredAge: "quantum_age",
			Description: "Computes across realities. +200 knowledge, +2.0 quantum flux/tick.",
		},
		{
			Name: "Reality Engine", Key: "reality_engine", Category: "production",
			BaseCost:  map[string]float64{"quantum_flux": 500e12, "antimatter": 2e15, "dark_matter": 3e15},
			CostScale: 1.6,
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: 5.0},
				{Type: "bonus", Target: "production_all", Value: 0.15},
			},
			RequiredAge: "quantum_age",
			Description: "Manipulates reality itself. +5.0 quantum flux/tick, +15% all production.",
		},
		{
			Name: "Transcendence Beacon", Key: "transcendence_beacon", Category: "production",
			BaseCost:  map[string]float64{"quantum_flux": 750e12, "antimatter": 3e15, "dark_matter": 5e15},
			CostScale: 1.6,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.2},
				{Type: "production", Target: "quantum_flux", Value: 3.0},
			},
			RequiredAge: "quantum_age",
			Description: "Beacon to the next plane. +20% all production, +3.0 quantum flux/tick.",
		},
		{
			Name: "Quantum Vault", Key: "quantum_vault", Category: "storage",
			BaseCost:    map[string]float64{"antimatter": 1e15, "dark_matter": 2e15, "titanium": 500e12},
			CostScale:   1.55,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 5e15}},
			RequiredAge: "quantum_age",
			Description: "Stores matter in quantum superposition. +5Q storage.",
		},

		// ===== TRANSCENDENT AGE =====
		// (singularity_core is a wonder, listed below)

		// ===== WONDERS (scaled to their age) =====
		{
			Name: "Stonehenge", Key: "stonehenge", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 8000, "wood": 5000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 0.8},
				{Type: "production", Target: "faith", Value: 0.5},
			},
			RequiredAge: "bronze_age",
			MaxCount:    1,
			BuildTicks:  10,
			Description: "Ancient monument. +0.8 knowledge, +0.5 faith/tick.",
		},
		{
			Name: "Colosseum", Key: "colosseum", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 40000, "iron": 12000, "gold": 8000},
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
			Name: "Parthenon", Key: "parthenon", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 200000, "gold": 50000, "iron": 30000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 2.0},
				{Type: "production", Target: "knowledge", Value: 1.2},
			},
			RequiredAge: "classical_age",
			MaxCount:    1,
			BuildTicks:  18,
			Description: "Temple of wisdom. +2.0 culture, +1.2 knowledge/tick.",
		},
		{
			Name: "Great Library", Key: "great_library", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 400000, "gold": 150000, "knowledge": 120000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 2.0},
			},
			RequiredAge: "medieval_age",
			MaxCount:    1,
			BuildTicks:  20,
			Description: "Repository of all knowledge. +2.0 knowledge/tick.",
		},
		{
			Name: "Space Program", Key: "space_program", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 60e9, "gold": 90e9, "electricity": 30e9, "knowledge": 60e9},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 4.0},
				{Type: "production", Target: "culture", Value: 5.0},
			},
			RequiredAge: "modern_age",
			MaxCount:    1,
			BuildTicks:  30,
			Description: "The ultimate achievement. +4 knowledge, +5 culture/tick.",
		},
		{
			Name: "Particle Accelerator", Key: "particle_accelerator", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 15e9, "electricity": 10e9, "uranium": 1.5e9},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 10.0},
				{Type: "production", Target: "uranium", Value: 1.0},
			},
			RequiredAge: "atomic_age",
			MaxCount:    1,
			BuildTicks:  25,
			Description: "Smashes atoms for science. +10 knowledge, +1.0 uranium/tick.",
		},
		{
			Name: "Dyson Scaffold", Key: "dyson_scaffold", Category: "wonder",
			BaseCost:  map[string]float64{"titanium": 10e12, "plasma": 12e12, "steel": 100e12},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 200.0},
				{Type: "production", Target: "plasma", Value: 10.0},
			},
			RequiredAge: "space_age",
			MaxCount:    1,
			BuildTicks:  40,
			Description: "Framework for a Dyson sphere. +200 electricity, +10 plasma/tick.",
		},
		{
			Name: "Singularity Core", Key: "singularity_core", Category: "wonder",
			BaseCost:  map[string]float64{"quantum_flux": 2.5e15, "antimatter": 5e15, "dark_matter": 10e15},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 1.0},
				{Type: "production", Target: "quantum_flux", Value: 20.0},
			},
			RequiredAge: "transcendent_age",
			MaxCount:    1,
			BuildTicks:  50,
			Description: "The final wonder. +100% all production, +20 quantum flux/tick.",
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
