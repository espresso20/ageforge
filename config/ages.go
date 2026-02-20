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
			UnlockBuildings: []string{"hut", "stash", "altar"},
			UnlockResources: []string{"food", "wood", "knowledge"},
			UnlockVillagers: []string{"worker", "shaman"},
		},
		// === 1: STONE AGE ===
		{
			Name: "Stone Age", Key: "stone_age", Order: 1,
			Description:  "Tools of stone change everything.",
			ResourceReqs: map[string]float64{"food": 1500, "wood": 1200, "knowledge": 200},
			BuildingReqs: map[string]int{"hut": 10, "altar": 5},
			UnlockBuildings: []string{"gathering_camp", "woodcutter_camp", "stone_pit", "firepit", "storage_pit"},
			UnlockResources: []string{"stone"},
		},
		// === 2: BRONZE AGE ===
		{
			Name: "Bronze Age", Key: "bronze_age", Order: 2,
			Description:  "Discovery of metalworking changes everything.",
			ResourceReqs: map[string]float64{"food": 1250, "stone": 750, "knowledge": 250},
			BuildingReqs: map[string]int{"hut": 8, "stone_pit": 4, "firepit": 3},
			UnlockBuildings: []string{"farm", "lumber_mill", "quarry", "mine", "market", "library", "house", "warehouse", "stonehenge"},
			UnlockResources: []string{"iron", "gold"},
			UnlockVillagers: []string{"scholar"},
		},
		// === 3: IRON AGE ===
		{
			Name: "Iron Age", Key: "iron_age", Order: 3,
			Description:  "Iron tools and weapons transform society.",
			ResourceReqs: map[string]float64{"stone": 5000, "iron": 1250, "knowledge": 1000},
			BuildingReqs: map[string]int{"mine": 5, "lumber_mill": 4, "library": 3},
			UnlockBuildings: []string{"coal_mine", "smithy", "barracks", "granary", "colosseum"},
			UnlockResources: []string{"coal"},
			UnlockVillagers: []string{"soldier"},
		},
		// === 4: CLASSICAL AGE ===
		{
			Name: "Classical Age", Key: "classical_age", Order: 4,
			Description:  "Great empires rise and philosophy flourishes.",
			ResourceReqs: map[string]float64{"stone": 25000, "iron": 6000, "gold": 4000, "knowledge": 5000},
			BuildingReqs: map[string]int{"barracks": 3, "library": 4, "market": 4},
			UnlockBuildings: []string{"forum", "aqueduct", "amphitheater", "classical_vault", "parthenon"},
		},
		// === 5: MEDIEVAL AGE ===
		{
			Name: "Medieval Age", Key: "medieval_age", Order: 5,
			Description:  "Kingdoms rise and feudalism takes hold.",
			ResourceReqs: map[string]float64{"stone": 125000, "iron": 30000, "gold": 20000, "knowledge": 25000},
			BuildingReqs: map[string]int{"forum": 3, "library": 5, "barracks": 4},
			UnlockBuildings: []string{"cathedral", "manor", "university", "castle", "keep", "great_library"},
			UnlockResources: []string{"steel", "faith"},
			UnlockVillagers: []string{"merchant"},
		},
		// === 6: RENAISSANCE AGE ===
		{
			Name: "Renaissance Age", Key: "renaissance_age", Order: 6,
			Description:  "Art, science, and exploration flourish.",
			ResourceReqs: map[string]float64{"gold": 95000, "knowledge": 125000, "steel": 12500, "faith": 25000},
			BuildingReqs: map[string]int{"university": 4, "market": 6, "castle": 2},
			UnlockBuildings: []string{"art_studio", "bank", "observatory", "renaissance_vault"},
			UnlockResources: []string{"culture"},
		},
		// === 7: COLONIAL AGE ===
		{
			Name: "Colonial Age", Key: "colonial_age", Order: 7,
			Description:  "Exploration and trade span the globe.",
			ResourceReqs: map[string]float64{"gold": 470000, "knowledge": 625000, "steel": 62500, "culture": 37500},
			BuildingReqs: map[string]int{"bank": 4, "observatory": 3, "art_studio": 4},
			UnlockBuildings: []string{"colony", "port", "plantation", "colonial_warehouse"},
		},
		// === 8: INDUSTRIAL AGE ===
		{
			Name: "Industrial Age", Key: "industrial_age", Order: 8,
			Description:  "Machines revolutionize production.",
			ResourceReqs: map[string]float64{"steel": 310000, "gold": 2340000, "knowledge": 3125000},
			BuildingReqs: map[string]int{"colony": 3, "port": 4, "plantation": 4},
			UnlockBuildings: []string{"factory", "oil_well", "apartment", "industrial_depot"},
			UnlockResources: []string{"oil"},
			UnlockVillagers: []string{"engineer"},
		},
		// === 9: VICTORIAN AGE ===
		{
			Name: "Victorian Age", Key: "victorian_age", Order: 9,
			Description:  "Steam and innovation drive progress.",
			ResourceReqs: map[string]float64{"steel": 625000, "oil": 125000, "gold": 4687500},
			BuildingReqs: map[string]int{"factory": 4, "oil_well": 3, "apartment": 3},
			UnlockBuildings: []string{"power_grid", "telegraph", "clocktower", "victorian_vault"},
			UnlockResources: []string{"electricity"},
		},
		// === 10: ELECTRIC AGE ===
		{
			Name: "Electric Age", Key: "electric_age", Order: 10,
			Description:  "Electrification transforms daily life.",
			ResourceReqs: map[string]float64{"steel": 3125000, "oil": 625000, "electricity": 250000},
			BuildingReqs: map[string]int{"power_grid": 3, "telegraph": 2, "factory": 6},
			UnlockBuildings: []string{"electric_mill", "telephone_exchange", "train_station", "electric_warehouse"},
		},
		// === 11: ATOMIC AGE ===
		{
			Name: "Atomic Age", Key: "atomic_age", Order: 11,
			Description:  "Nuclear power unleashes terrifying potential.",
			ResourceReqs: map[string]float64{"steel": 15625000, "electricity": 1250000, "oil": 3125000},
			BuildingReqs: map[string]int{"electric_mill": 3, "train_station": 2, "telephone_exchange": 2},
			UnlockBuildings: []string{"reactor", "bunker", "missile_silo", "atomic_vault", "particle_accelerator"},
			UnlockResources: []string{"uranium"},
		},
		// === 12: MODERN AGE ===
		{
			Name: "Modern Age", Key: "modern_age", Order: 12,
			Description:  "Technology and innovation define the era.",
			ResourceReqs: map[string]float64{"electricity": 6250000, "uranium": 1500000, "steel": 78125000},
			BuildingReqs: map[string]int{"reactor": 3, "bunker": 2, "missile_silo": 1},
			UnlockBuildings: []string{"power_plant", "research_lab", "skyscraper", "modern_depot", "space_program"},
			UnlockResources: []string{"data"},
		},
		// === 13: INFORMATION AGE ===
		{
			Name: "Information Age", Key: "information_age", Order: 13,
			Description:  "The Internet connects the world.",
			ResourceReqs: map[string]float64{"electricity": 31250000, "data": 5000000, "gold": 100000000},
			BuildingReqs: map[string]int{"research_lab": 3, "skyscraper": 3, "power_plant": 3},
			UnlockBuildings: []string{"server_farm", "fiber_hub", "media_center", "info_vault"},
			UnlockVillagers: []string{"hacker"},
		},
		// === 14: DIGITAL AGE ===
		{
			Name: "Digital Age", Key: "digital_age", Order: 14,
			Description:  "Full digitization of civilization.",
			ResourceReqs: map[string]float64{"data": 25000000, "electricity": 156250000},
			BuildingReqs: map[string]int{"server_farm": 3, "fiber_hub": 2, "media_center": 2},
			UnlockBuildings: []string{"data_center", "ai_lab", "smart_grid", "digital_archive"},
		},
		// === 15: CYBERPUNK AGE ===
		{
			Name: "Cyberpunk Age", Key: "cyberpunk_age", Order: 15,
			Description:  "Neon lights and cybernetic augmentation.",
			ResourceReqs: map[string]float64{"data": 125000000, "electricity": 781250000},
			BuildingReqs: map[string]int{"ai_lab": 3, "data_center": 3, "smart_grid": 2},
			UnlockBuildings: []string{"augmentation_clinic", "neon_tower", "black_market", "cyber_vault"},
			UnlockResources: []string{"crypto"},
		},
		// === 16: FUSION AGE ===
		{
			Name: "Fusion Age", Key: "fusion_age", Order: 16,
			Description:  "Clean energy breakthrough changes everything.",
			ResourceReqs: map[string]float64{"electricity": 3906250000, "crypto": 200000000, "data": 625000000},
			BuildingReqs: map[string]int{"augmentation_clinic": 2, "neon_tower": 3, "black_market": 2},
			UnlockBuildings: []string{"fusion_reactor", "plasma_forge", "maglev_station", "fusion_vault"},
			UnlockResources: []string{"plasma"},
		},
		// === 17: SPACE AGE ===
		{
			Name: "Space Age", Key: "space_age", Order: 17,
			Description:  "Orbital expansion begins.",
			ResourceReqs: map[string]float64{"plasma": 500000000, "electricity": 19531250000, "data": 3125000000},
			BuildingReqs: map[string]int{"fusion_reactor": 3, "plasma_forge": 2, "maglev_station": 2},
			UnlockBuildings: []string{"launch_pad", "space_station", "orbital_habitat", "orbital_depot", "dyson_scaffold"},
			UnlockResources: []string{"titanium"},
			UnlockVillagers: []string{"astronaut"},
		},
		// === 18: INTERSTELLAR AGE ===
		{
			Name: "Interstellar Age", Key: "interstellar_age", Order: 18,
			Description:  "Between the stars, new frontiers await.",
			ResourceReqs: map[string]float64{"titanium": 1000000000, "plasma": 2500000000},
			BuildingReqs: map[string]int{"launch_pad": 3, "space_station": 2, "orbital_habitat": 2},
			UnlockBuildings: []string{"warp_gate", "colony_ship", "star_forge", "stellar_vault"},
			UnlockResources: []string{"dark_matter"},
		},
		// === 19: GALACTIC AGE ===
		{
			Name: "Galactic Age", Key: "galactic_age", Order: 19,
			Description:  "Galactic civilization spans the cosmos.",
			ResourceReqs: map[string]float64{"dark_matter": 2000000000, "titanium": 5000000000},
			BuildingReqs: map[string]int{"warp_gate": 3, "colony_ship": 2, "star_forge": 2},
			UnlockBuildings: []string{"galactic_hub", "antimatter_plant", "megastructure", "galactic_vault"},
			UnlockResources: []string{"antimatter"},
		},
		// === 20: QUANTUM AGE ===
		{
			Name: "Quantum Age", Key: "quantum_age", Order: 20,
			Description:  "Reality bends to quantum mastery.",
			ResourceReqs: map[string]float64{"antimatter": 5000000000, "dark_matter": 10000000000},
			BuildingReqs: map[string]int{"galactic_hub": 2, "antimatter_plant": 3, "megastructure": 1},
			UnlockBuildings: []string{"quantum_computer", "reality_engine", "transcendence_beacon", "quantum_vault"},
			UnlockResources: []string{"quantum_flux"},
		},
		// === 21: TRANSCENDENT AGE ===
		{
			Name: "Transcendent Age", Key: "transcendent_age", Order: 21,
			Description:  "Final ascension. The ultimate civilization.",
			ResourceReqs: map[string]float64{"quantum_flux": 15000000000, "antimatter": 25000000000},
			BuildingReqs: map[string]int{"quantum_computer": 3, "reality_engine": 2, "transcendence_beacon": 2},
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
