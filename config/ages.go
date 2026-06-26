// Package config provides all static game data: ages, buildings, techs,
// milestones, workers, epochs, events, expeditions, trade routes, and diplomacy.
// All functions are pure (no global state). Call them at startup or on-demand —
// they are cheap enough to call per-tick for config lookups.
package config

// AgeDef defines a single playable age. Ages are ordered by the Order field
// (0 = Primitive Age, 21 = Transcendent Age). Advancing requires meeting both
// ResourceReqs and BuildingReqs simultaneously.
type AgeDef struct {
	Name  string
	Key   string
	Order int // 0-indexed position in the age sequence; used for sorting and comparison
	// Requirements to advance TO this age (checked at the time of 'advance' command)
	ResourceReqs map[string]float64 // minimum resource amounts (float64, not capped to storage)
	BuildingReqs map[string]int     // minimum building counts
	// What this age unlocks when entered
	UnlockBuildings []string // building keys made available at the start of this age
	UnlockResources []string // resource keys made available at the start of this age
	UnlockVillagers []string // worker domain keys (legacy field — only used for primitive "worker")
	Description     string
	EpochKey string // which of the 7 epochs this age belongs to (e.g. "stone_era", "iron_era")
}

// Ages returns all 22 age definitions in ascending order (Primitive → Transcendent).
// Callers that need random access should use AgeByKey(). Callers that need the
// sorted key list should use AgeOrder().
func Ages() []AgeDef {
	return []AgeDef{
		// === 0: PRIMITIVE AGE (Stone Era) ===
		{
			Name: "Primitive Age", Key: "primitive_age", Order: 0,
			EpochKey:        "stone_era",
			Description:     "Survival. Nothing but your hands and wits.",
			UnlockBuildings: []string{"hut", "stash", "gathering_camp", "wood_camp", "story_circle", "shrine", "sacred_grove"},
			UnlockResources: []string{"food", "wood", "knowledge", "faith"},
			UnlockVillagers: []string{"worker"},
		},
		// === 1: STONE AGE (Stone Era) ===
		{
			Name: "Stone Age", Key: "stone_age", Order: 1,
			EpochKey:        "stone_era",
			Description:     "Tools of stone change everything.",
			ResourceReqs:    map[string]float64{"food": 8000, "wood": 5200, "knowledge": 1400},
			BuildingReqs:    map[string]int{"hut": 20, "story_circle": 5},
			UnlockBuildings: []string{"longhouse", "storage_pit", "forager_post", "woodcutter_camp", "stone_camp", "stone_pit", "elders_hall", "standing_stones", "war_camp", "great_monolith"},
			UnlockResources: []string{"stone"},
		},
		// === 2: BRONZE AGE (Stone Era) ===
		{
			Name: "Bronze Age", Key: "bronze_age", Order: 2,
			EpochKey:        "stone_era",
			Description:     "Discovery of metalworking changes everything.",
			ResourceReqs:    map[string]float64{"food": 15000, "wood": 8000, "stone": 4000, "knowledge": 5000},
			BuildingReqs:    map[string]int{"longhouse": 50, "stone_pit": 5, "elders_hall": 5},
			UnlockBuildings: []string{"house", "warehouse", "farm", "lumber_mill", "quarry", "scriptorium", "altar", "barracks", "market", "smithy", "stonehenge"},
			UnlockResources: []string{"iron", "gold"},
		},
		// === 3: IRON AGE (Iron Era) ===
		{
			Name: "Iron Age", Key: "iron_age", Order: 3,
			EpochKey:        "iron_era",
			Description:     "Iron tools and weapons transform society.",
			ResourceReqs:    map[string]float64{"food": 40000, "wood": 20000, "stone": 8000, "iron": 4000, "knowledge": 10000},
			BuildingReqs:    map[string]int{"lumber_mill": 8, "quarry": 8, "scriptorium": 3},
			UnlockBuildings: []string{"townhouse", "granary", "field_works", "timber_yard", "marble_quarry", "agora", "temple", "hunting_lodge", "legion_fort", "trading_post", "ironworks", "smelter", "colosseum"},
			UnlockResources: []string{"marble", "iron_ore", "soldiers"},
		},
		// === 4: CLASSICAL AGE (Iron Era) ===
		{
			Name: "Classical Age", Key: "classical_age", Order: 4,
			EpochKey:        "iron_era",
			Description:     "Great empires are built and philosophy flourishes.",
			ResourceReqs:    map[string]float64{"stone": 75000, "iron": 15000, "gold": 8000, "knowledge": 20000},
			BuildingReqs:    map[string]int{"barracks": 15, "agora": 8, "market": 5},
			UnlockBuildings: []string{"villa", "classical_vault", "estate_farm", "wood_workshop", "marble_works", "library", "oracle_house", "military_academy", "merchant_quarter", "aqueduct", "forge", "amphitheater", "parthenon"},
			UnlockResources: []string{"culture"},
		},
		// === 5: MEDIEVAL AGE (Iron Era) ===
		{
			Name: "Medieval Age", Key: "medieval_age", Order: 5,
			EpochKey:        "iron_era",
			Description:     "Kingdoms rise and feudalism takes hold.",
			ResourceReqs:    map[string]float64{"stone": 125000, "iron": 30000, "gold": 20000, "knowledge": 50000},
			BuildingReqs:    map[string]int{"merchant_quarter": 3, "library": 20, "barracks": 30},
			UnlockBuildings: []string{"manor", "keep", "demesne", "sawmill", "stonemasons_guild", "monastery_library", "cathedral", "castle_keep", "guildhall", "workshop", "ironmonger", "great_hall", "great_library"},
			UnlockResources: []string{"steel"},
		},
		// === 6: RENAISSANCE AGE (Steel Era) ===
		{
			Name: "Renaissance Age", Key: "renaissance_age", Order: 6,
			EpochKey:        "steel_era",
			Description:     "Art, science, and exploration flourish.",
			ResourceReqs:    map[string]float64{"gold": 100000, "knowledge": 125000, "steel": 2000, "faith": 25000},
			BuildingReqs:    map[string]int{"monastery_library": 5, "market": 15, "castle_keep": 3},
			UnlockBuildings: []string{"estate", "renaissance_vault", "market_garden", "coal_mine", "iron_mine", "university", "basilica", "fortress", "exchange", "mill", "foundry", "art_studio", "sistine_chapel"},
			UnlockResources: []string{"coal"},
		},
		// === 7: COLONIAL AGE (Steel Era) ===
		{
			Name: "Colonial Age", Key: "colonial_age", Order: 7,
			EpochKey:        "steel_era",
			Description:     "Exploration and trade span the globe.",
			ResourceReqs:    map[string]float64{"gold": 470000, "knowledge": 625000, "steel": 76500, "culture": 200000},
			BuildingReqs:    map[string]int{"exchange": 5, "university": 3, "art_studio": 5},
			UnlockBuildings: []string{"settlement_block", "colonial_warehouse", "plantation", "coal_works", "deep_iron_mine", "natural_philosophy_hall", "mission", "fort", "port", "dockyard", "iron_works", "concert_hall", "grand_lighthouse"},
		},
		// === 8: INDUSTRIAL AGE (Steel Era) ===
		{
			Name: "Industrial Age", Key: "industrial_age", Order: 8,
			EpochKey:        "steel_era",
			Description:     "Machines revolutionize production.",
			ResourceReqs:    map[string]float64{"steel": 310000, "gold": 2500000, "knowledge": 2000000},
			BuildingReqs:    map[string]int{"plantation": 5, "port": 8, "market_garden": 5},
			UnlockBuildings: []string{"tenement", "industrial_depot", "agricultural_works", "steam_coal_plant", "steam_mine", "research_institute", "church", "military_base", "stock_exchange", "iron_works_complex", "steel_mill", "coal_plant", "opera_house", "crystal_palace"},
			UnlockResources: []string{"oil"},
		},
		// === 9: VICTORIAN AGE (Electric Era) ===
		{
			Name: "Victorian Age", Key: "victorian_age", Order: 9,
			EpochKey:        "electric_era",
			Description:     "Steam and innovation drive progress.",
			ResourceReqs:    map[string]float64{"steel": 1625000, "oil": 725000, "gold": 9687500},
			BuildingReqs:    map[string]int{"steel_mill": 5, "iron_works_complex": 3, "tenement": 30},
			UnlockBuildings: []string{"row_house", "victorian_vault", "mechanized_farm", "oil_derrick", "uranium_mine", "academy", "grand_cathedral", "garrison", "bank", "steam_works", "bessemer_plant", "steam_turbine", "grand_museum", "eiffel_tower"},
			UnlockResources: []string{"electricity"},
		},
		// === 10: ELECTRIC AGE (Electric Era) ===
		{
			Name: "Electric Age", Key: "electric_age", Order: 10,
			EpochKey:        "electric_era",
			Description:     "Electrification transforms daily life.",
			ResourceReqs:    map[string]float64{"steel": 9125000, "oil": 2625000, "electricity": 850000},
			BuildingReqs:    map[string]int{"steam_turbine": 20, "academy": 10, "steel_mill": 15},
			UnlockBuildings: []string{"apartment_block", "electric_warehouse", "industrial_farm", "oil_field", "nuclear_extraction_plant", "physics_laboratory", "revival_hall", "command_post", "financial_district", "power_station", "electric_arc_furnace", "power_generator", "radio_station", "hoover_dam"},
		},
		// === 11: ATOMIC AGE (Electric Era) ===
		{
			Name: "Atomic Age", Key: "atomic_age", Order: 11,
			EpochKey:        "electric_era",
			Description:     "Nuclear power unleashes terrifying potential.",
			ResourceReqs:    map[string]float64{"steel": 85625000, "electricity": 9250000, "oil": 6125000},
			BuildingReqs:    map[string]int{"electric_arc_furnace": 20, "steam_works": 20, "physics_laboratory": 15},
			UnlockBuildings: []string{"housing_project", "atomic_vault", "agricultural_complex", "petroleum_refinery", "uranium_processing_works", "research_campus", "spiritual_center", "bunker_complex", "corporate_hq", "nuclear_plant", "advanced_alloy_plant", "nuclear_reactor", "cinema", "particle_accelerator"},
			UnlockResources: []string{"uranium"},
		},
		// === 12: MODERN AGE (Digital Era) ===
		{
			Name: "Modern Age", Key: "modern_age", Order: 12,
			EpochKey:        "digital_era",
			Description:     "Technology and innovation define the era.",
			ResourceReqs:    map[string]float64{"electricity": 26250000, "uranium": 5500000, "steel": 378125000},
			BuildingReqs:    map[string]int{"nuclear_reactor": 30, "bunker_complex": 30, "research_campus": 15},
			UnlockBuildings: []string{"tower_block", "modern_depot", "agri_complex", "oil_platform", "titanium_mine", "think_tank", "meditation_center", "special_ops_hq", "investment_firm", "power_grid_hub", "titanium_smelter", "oil_refinery", "tv_studio", "space_program"},
			UnlockResources: []string{"data", "nanobots"},
		},
		// === 13: INFORMATION AGE (Digital Era) ===
		{
			Name: "Information Age", Key: "information_age", Order: 13,
			EpochKey:        "digital_era",
			Description:     "The Internet connects the world.",
			ResourceReqs:    map[string]float64{"electricity": 531250000, "data": 55000000, "gold": 1000000000},
			BuildingReqs:    map[string]int{"think_tank": 50, "tower_block": 30, "oil_refinery": 60},
			UnlockBuildings: []string{"smart_complex", "info_vault", "smart_farm", "smart_refinery", "precision_mine", "innovation_hub", "digital_temple", "cyber_command", "venture_hub", "smart_grid_node", "aerospace_foundry", "smart_energy_grid", "server_farm", "media_center", "global_network"},
		},
		// === 14: DIGITAL AGE (Digital Era) ===
		{
			Name: "Digital Age", Key: "digital_age", Order: 14,
			EpochKey:        "digital_era",
			Description:     "Full digitization of civilization.",
			ResourceReqs:    map[string]float64{"data": 2500000000, "electricity": 15625000000},
			BuildingReqs:    map[string]int{"server_farm": 30, "media_center": 80, "innovation_hub": 30},
			UnlockBuildings: []string{"megaplex", "digital_archive", "nano_farm", "bio_fabrication_lab", "nano_drill_complex", "ai_research_lab", "cyber_shrine", "drone_warfare_center", "crypto_exchange", "neural_grid", "nano_alloy_plant", "quantum_battery_array", "data_center", "vr_studio", "world_simulation"},
		},
		// === 15: CYBERPUNK AGE (Neon Era) ===
		{
			Name: "Cyberpunk Age", Key: "cyberpunk_age", Order: 15,
			EpochKey:        "neon_era",
			Description:     "Neon lights and cybernetic augmentation.",
			ResourceReqs:    map[string]float64{"data": 125000000000, "electricity": 781250000000},
			BuildingReqs:    map[string]int{"ai_research_lab": 80, "data_center": 80, "neural_grid": 50},
			UnlockBuildings: []string{"arcology_pod", "cyber_vault", "vat_farm", "nanobot_vat", "dark_crystal_mine", "neuro_research_center", "neon_sanctuary", "combat_aug_center", "black_market", "augmentation_foundry", "dark_matter_refinery", "dark_energy_tap", "cyber_hub", "holographic_theater", "neon_citadel"},
			UnlockResources: []string{"crypto", "dark_matter_crystals"},
		},
		// === 16: FUSION AGE (Neon Era) ===
		{
			Name: "Fusion Age", Key: "fusion_age", Order: 16,
			EpochKey:        "neon_era",
			Description:     "Clean energy breakthrough changes everything.",
			ResourceReqs:    map[string]float64{"electricity": 390625000000, "crypto": 20000000000, "data": 62500000000},
			BuildingReqs:    map[string]int{"augmentation_foundry": 50, "arcology_pod": 80, "black_market": 50},
			UnlockBuildings: []string{"habitat_ring", "fusion_vault", "bio_reactor_farm", "molecular_synthesizer", "exotic_mineral_extractor", "theoretical_institute", "quantum_chapel", "plasma_command", "energy_exchange", "fusion_reactor", "exotic_matter_forge", "fusion_reactor_array", "quantum_server_farm", "neural_art_complex", "stellar_cradle"},
			UnlockResources: []string{"plasma"},
		},
		// === 17: SPACE AGE (Neon Era) ===
		{
			Name: "Space Age", Key: "space_age", Order: 17,
			EpochKey:        "neon_era",
			Description:     "Orbital expansion begins.",
			ResourceReqs:    map[string]float64{"plasma": 50000000000, "electricity": 1953125000000, "data": 312500000000},
			BuildingReqs:    map[string]int{"fusion_reactor": 80, "fusion_reactor_array": 60, "plasma_command": 50},
			UnlockBuildings: []string{"orbital_habitat", "orbital_depot", "hydroponic_bay", "quantum_organic_extractor", "asteroid_crystal_mine", "deep_space_observatory", "orbital_sanctuary", "space_force_base", "asteroid_market", "launch_complex", "orbital_refinery", "solar_collector_array", "orbital_data_relay", "zero_g_gallery", "dyson_scaffold"},
			UnlockResources: []string{"titanium", "titanium_ore"},
		},
		// === 18: INTERSTELLAR AGE (Cosmic Era) ===
		{
			Name: "Interstellar Age", Key: "interstellar_age", Order: 18,
			EpochKey:        "cosmic_era",
			Description:     "Between the stars, new frontiers await.",
			ResourceReqs:    map[string]float64{"titanium": 100000000000, "plasma": 250000000000},
			BuildingReqs:    map[string]int{"launch_complex": 80, "orbital_habitat": 60, "solar_collector_array": 50},
			UnlockBuildings: []string{"generation_ship", "stellar_vault", "protein_synthesizer", "reality_matter_weaver", "stellar_core_drill", "xenology_institute", "void_monastery", "fleet_command", "galactic_trade_hub", "warp_drive_plant", "antimatter_forge", "pulsar_tap", "galactic_network_node", "cultural_beacon", "warp_nexus"},
			UnlockResources: []string{"dark_matter"},
		},
		// === 19: GALACTIC AGE (Cosmic Era) ===
		{
			Name: "Galactic Age", Key: "galactic_age", Order: 19,
			EpochKey:        "cosmic_era",
			Description:     "Galactic civilization spans the cosmos.",
			ResourceReqs:    map[string]float64{"dark_matter": 200000000000, "titanium": 500000000000},
			BuildingReqs:    map[string]int{"warp_drive_plant": 80, "generation_ship": 60, "orbital_refinery": 50},
			UnlockBuildings: []string{"dyson_sphere_habitat", "galactic_vault", "matter_converter", "cosmic_organic_works", "neutron_star_mine", "cosmic_research_station", "stellar_shrine", "stellar_armada_hq", "stellar_exchange", "dyson_assembly", "stellar_metallurgy", "quasar_tap", "consciousness_upload_hub", "civilization_archive", "cosmic_beacon"},
			UnlockResources: []string{"antimatter"},
		},
		// === 20: QUANTUM AGE (Cosmic Era) ===
		{
			Name: "Quantum Age", Key: "quantum_age", Order: 20,
			EpochKey:        "cosmic_era",
			Description:     "Reality bends to quantum mastery.",
			ResourceReqs:    map[string]float64{"antimatter": 5000000000000, "dark_matter": 10000000000000},
			BuildingReqs:    map[string]int{"stellar_exchange": 80, "antimatter_forge": 100, "dyson_sphere_habitat": 120},
			UnlockBuildings: []string{"reality_fold", "quantum_vault", "quantum_cultivator", "reality_harvester", "reality_excavator", "reality_academy", "transcendence_hall", "probability_war_room", "probability_market", "reality_forge", "quantum_metal_works", "zero_point_generator", "reality_processor", "reality_art_engine", "reality_anchor"},
			UnlockResources: []string{"quantum_flux"},
		},
		// === 21: TRANSCENDENT AGE (Cosmic Era) ===
		{
			Name: "Transcendent Age", Key: "transcendent_age", Order: 21,
			EpochKey:        "cosmic_era",
			Description:     "Final ascension. The ultimate civilization.",
			ResourceReqs:    map[string]float64{"quantum_flux": 150000000000000, "antimatter": 250000000000000},
			BuildingReqs:    map[string]int{"reality_academy": 500, "reality_forge": 300, "probability_war_room": 200},
			UnlockBuildings: []string{"singularity_core", "transcendent_nexus", "omniversal_war_council", "omniversal_bazaar", "singularity_engine"},
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
