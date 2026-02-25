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
		// === 0: PRIMITIVE AGE ===
		{
			Name: "Primitive Age", Key: "primitive_age", Order: 0,
			Description:     "Survival. Nothing but your hands and wits.",
			UnlockBuildings: []string{"hut", "stash", "altar", "sacred_grove"},
			UnlockResources: []string{"food", "wood", "knowledge"},
			UnlockVillagers: []string{"worker", "shaman"},
		},
		// === 1: STONE AGE ===
		{
			Name: "Stone Age", Key: "stone_age", Order: 1,
			Description:     "Tools of stone change everything.",
			ResourceReqs:    map[string]float64{"food": 8000, "wood": 5200, "knowledge": 1400},
			BuildingReqs:    map[string]int{"hut": 20, "altar": 10},
			UnlockBuildings: []string{"gathering_camp", "woodcutter_camp", "stone_pit", "firepit", "storage_pit", "great_monolith"},
			UnlockResources: []string{"stone"},
		},
		// === 2: BRONZE AGE ===
		{
			Name: "Bronze Age", Key: "bronze_age", Order: 2,
			Description:     "Discovery of metalworking changes everything.",
			ResourceReqs:    map[string]float64{"food": 15000, "wood": 8000, "stone": 4000, "knowledge": 5000},
			BuildingReqs:    map[string]int{"hut": 100, "stone_pit": 10, "firepit": 10},
			UnlockBuildings: []string{"farm", "lumber_mill", "quarry", "mine", "market", "library", "house", "warehouse", "stonehenge"},
			UnlockResources: []string{"iron", "gold"},
			UnlockVillagers: []string{"scholar"},
		},
		// === 3: IRON AGE ===
		{
			Name: "Iron Age", Key: "iron_age", Order: 3,
			Description:     "Iron tools and weapons transform society.",
			ResourceReqs:    map[string]float64{"food": 300000, "wood": 20000, "stone": 8000, "iron": 4000, "knowledge": 10000},
			BuildingReqs:    map[string]int{"mine": 10, "lumber_mill": 10, "library": 5},
			UnlockBuildings: []string{"coal_mine", "smithy", "barracks", "granary", "colosseum"},
			UnlockResources: []string{"coal"},
			UnlockVillagers: []string{"soldier"},
		},
		// === 4: CLASSICAL AGE ===
		{
			Name: "Classical Age", Key: "classical_age", Order: 4,
			Description:     "Great empires are built and philosophy flourishes.",
			ResourceReqs:    map[string]float64{"stone": 75000, "iron": 15000, "gold": 8000, "knowledge": 20000},
			BuildingReqs:    map[string]int{"barracks": 20, "library": 10, "market": 5},
			UnlockBuildings: []string{"forum", "aqueduct", "amphitheater", "classical_vault", "parthenon"},
		},
		// === 5: MEDIEVAL AGE ===
		{
			Name: "Medieval Age", Key: "medieval_age", Order: 5,
			Description:     "Kingdoms rise and feudalism takes hold.",
			ResourceReqs:    map[string]float64{"stone": 125000, "iron": 30000, "gold": 20000, "knowledge": 50000},
			BuildingReqs:    map[string]int{"forum": 5, "library": 30, "barracks": 50},
			UnlockBuildings: []string{"cathedral", "manor", "university", "castle", "keep", "great_library"},
			UnlockResources: []string{"steel", "faith"},
			UnlockVillagers: []string{"merchant"},
		},
		// === 6: RENAISSANCE AGE ===
		{
			Name: "Renaissance Age", Key: "renaissance_age", Order: 6,
			Description:     "Art, science, and exploration flourish.",
			ResourceReqs:    map[string]float64{"gold": 100000, "knowledge": 125000, "steel": 20500, "faith": 25000},
			BuildingReqs:    map[string]int{"university": 10, "market": 20, "castle": 5},
			UnlockBuildings: []string{"art_studio", "bank", "observatory", "renaissance_vault", "sistine_chapel"},
			UnlockResources: []string{"culture"},
		},
		// === 7: COLONIAL AGE ===
		{
			Name: "Colonial Age", Key: "colonial_age", Order: 7,
			Description:     "Exploration and trade span the globe.",
			ResourceReqs:    map[string]float64{"gold": 470000, "knowledge": 625000, "steel": 76500, "culture": 200000},
			BuildingReqs:    map[string]int{"bank": 10, "observatory": 3, "art_studio": 5},
			UnlockBuildings: []string{"colony", "port", "plantation", "colonial_warehouse", "grand_lighthouse"},
		},
		// === 8: INDUSTRIAL AGE ===
		{
			Name: "Industrial Age", Key: "industrial_age", Order: 8,
			Description:     "Machines revolutionize production.",
			ResourceReqs:    map[string]float64{"steel": 310000, "gold": 5340000, "knowledge": 4125000},
			BuildingReqs:    map[string]int{"colony": 5, "port": 10, "plantation": 15},
			UnlockBuildings: []string{"factory", "oil_well", "apartment", "industrial_depot", "crystal_palace"},
			UnlockResources: []string{"oil"},
			UnlockVillagers: []string{"engineer"},
		},
		// === 9: VICTORIAN AGE ===
		{
			Name: "Victorian Age", Key: "victorian_age", Order: 9,
			Description:     "Steam and innovation drive progress.",
			ResourceReqs:    map[string]float64{"steel": 1625000, "oil": 725000, "gold": 9687500},
			BuildingReqs:    map[string]int{"factory": 10, "oil_well": 15, "apartment": 50},
			UnlockBuildings: []string{"power_grid", "telegraph", "clocktower", "victorian_vault", "eiffel_tower"},
			UnlockResources: []string{"electricity"},
		},
		// === 10: ELECTRIC AGE ===
		{
			Name: "Electric Age", Key: "electric_age", Order: 10,
			Description:     "Electrification transforms daily life.",
			ResourceReqs:    map[string]float64{"steel": 9125000, "oil": 2625000, "electricity": 850000},
			BuildingReqs:    map[string]int{"power_grid": 100, "telegraph": 15, "factory": 50},
			UnlockBuildings: []string{"electric_mill", "telephone_exchange", "train_station", "electric_warehouse", "hoover_dam"},
		},
		// === 11: ATOMIC AGE ===
		{
			Name: "Atomic Age", Key: "atomic_age", Order: 11,
			Description:     "Nuclear power unleashes terrifying potential.",
			ResourceReqs:    map[string]float64{"steel": 85625000, "electricity": 9250000, "oil": 6125000},
			BuildingReqs:    map[string]int{"electric_mill": 50, "train_station": 50, "telephone_exchange": 50},
			UnlockBuildings: []string{"reactor", "bunker", "missile_silo", "atomic_vault", "particle_accelerator"},
			UnlockResources: []string{"uranium"},
		},
		// === 12: MODERN AGE ===
		{
			Name: "Modern Age", Key: "modern_age", Order: 12,
			Description:     "Technology and innovation define the era.",
			ResourceReqs:    map[string]float64{"electricity": 26250000, "uranium": 5500000, "steel": 378125000},
			BuildingReqs:    map[string]int{"reactor": 100, "bunker": 90, "missile_silo": 50},
			UnlockBuildings: []string{"power_plant", "research_lab", "skyscraper", "modern_depot", "space_program"},
			UnlockResources: []string{"data"},
		},
		// === 13: INFORMATION AGE ===
		{
			Name: "Information Age", Key: "information_age", Order: 13,
			Description:     "The Internet connects the world.",
			ResourceReqs:    map[string]float64{"electricity": 531250000, "data": 55000000, "gold": 1000000000},
			BuildingReqs:    map[string]int{"research_lab": 200, "skyscraper": 100, "power_plant": 300},
			UnlockBuildings: []string{"server_farm", "fiber_hub", "media_center", "info_vault", "global_network"},
			UnlockVillagers: []string{"hacker"},
		},
		// === 14: DIGITAL AGE ===
		{
			Name: "Digital Age", Key: "digital_age", Order: 14,
			Description:     "Full digitization of civilization.",
			ResourceReqs:    map[string]float64{"data": 2500000000, "electricity": 15625000000},
			BuildingReqs:    map[string]int{"server_farm": 100, "fiber_hub": 200, "media_center": 400},
			UnlockBuildings: []string{"data_center", "ai_lab", "smart_grid", "digital_archive", "world_simulation"},
		},
		// === 15: CYBERPUNK AGE ===
		{
			Name: "Cyberpunk Age", Key: "cyberpunk_age", Order: 15,
			Description:     "Neon lights and cybernetic augmentation.",
			ResourceReqs:    map[string]float64{"data": 125000000000, "electricity": 781250000000},
			BuildingReqs:    map[string]int{"ai_lab": 300, "data_center": 300, "smart_grid": 200},
			UnlockBuildings: []string{"augmentation_clinic", "neon_tower", "black_market", "cyber_vault", "neon_citadel"},
			UnlockResources: []string{"crypto"},
		},
		// === 16: FUSION AGE ===
		{
			Name: "Fusion Age", Key: "fusion_age", Order: 16,
			Description:     "Clean energy breakthrough changes everything.",
			ResourceReqs:    map[string]float64{"electricity": 390625000000, "crypto": 20000000000, "data": 62500000000},
			BuildingReqs:    map[string]int{"augmentation_clinic": 200, "neon_tower": 300, "black_market": 200},
			UnlockBuildings: []string{"fusion_reactor", "plasma_forge", "maglev_station", "fusion_vault", "stellar_cradle"},
			UnlockResources: []string{"plasma"},
		},
		// === 17: SPACE AGE ===
		{
			Name: "Space Age", Key: "space_age", Order: 17,
			Description:     "Orbital expansion begins.",
			ResourceReqs:    map[string]float64{"plasma": 50000000000, "electricity": 1953125000000, "data": 312500000000},
			BuildingReqs:    map[string]int{"fusion_reactor": 300, "plasma_forge": 200, "maglev_station": 200},
			UnlockBuildings: []string{"launch_pad", "space_station", "orbital_habitat", "orbital_depot", "dyson_scaffold"},
			UnlockResources: []string{"titanium"},
			UnlockVillagers: []string{"astronaut"},
		},
		// === 18: INTERSTELLAR AGE ===
		{
			Name: "Interstellar Age", Key: "interstellar_age", Order: 18,
			Description:     "Between the stars, new frontiers await.",
			ResourceReqs:    map[string]float64{"titanium": 100000000000, "plasma": 250000000000},
			BuildingReqs:    map[string]int{"launch_pad": 300, "space_station": 200, "orbital_habitat": 200},
			UnlockBuildings: []string{"warp_gate", "colony_ship", "star_forge", "stellar_vault", "warp_nexus"},
			UnlockResources: []string{"dark_matter"},
		},
		// === 19: GALACTIC AGE ===
		{
			Name: "Galactic Age", Key: "galactic_age", Order: 19,
			Description:     "Galactic civilization spans the cosmos.",
			ResourceReqs:    map[string]float64{"dark_matter": 200000000000, "titanium": 500000000000},
			BuildingReqs:    map[string]int{"warp_gate": 300, "colony_ship": 200, "star_forge": 200},
			UnlockBuildings: []string{"galactic_hub", "antimatter_plant", "megastructure", "galactic_vault", "cosmic_beacon"},
			UnlockResources: []string{"antimatter"},
		},
		// === 20: QUANTUM AGE ===
		{
			Name: "Quantum Age", Key: "quantum_age", Order: 20,
			Description:     "Reality bends to quantum mastery.",
			ResourceReqs:    map[string]float64{"antimatter": 5000000000000, "dark_matter": 10000000000000},
			BuildingReqs:    map[string]int{"galactic_hub": 500, "antimatter_plant": 800, "megastructure": 1000},
			UnlockBuildings: []string{"quantum_computer", "reality_engine", "transcendence_beacon", "quantum_vault", "reality_anchor"},
			UnlockResources: []string{"quantum_flux"},
		},
		// === 21: TRANSCENDENT AGE ===
		{
			Name: "Transcendent Age", Key: "transcendent_age", Order: 21,
			Description:     "Final ascension. The ultimate civilization.",
			ResourceReqs:    map[string]float64{"quantum_flux": 150000000000000, "antimatter": 250000000000000},
			BuildingReqs:    map[string]int{"quantum_computer": 3000, "reality_engine": 2000, "transcendence_beacon": 2000},
			UnlockBuildings: []string{"singularity_core"},
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
