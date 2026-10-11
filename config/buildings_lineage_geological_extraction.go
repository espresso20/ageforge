package config

// buildingsLineageGeologicalExtraction returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// All fields are set inline; no separate buildingMeta() entries needed for these buildings.
func buildingsLineageGeologicalExtraction() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 4 — GEOLOGICAL EXTRACTION (lineageKey: "geological_extraction", domain: "masonry")
	// rate = 0.03 * 2^tier  CostScale: 1.30  Category: "production"
	// Output: stone(0-2) → dual marble+iron_ore(3-4) → iron_ore(5-8) → uranium(9-11)
	//         → titanium_ore(12-14) → dark_matter_crystals(15-17) → antimatter(18-20)
	// Dual-output tiers (3-4): two Effects each at half rate
	// =========================================================================

	// tier 0 — stone_age  rate=0.08  output=stone
	b = append(b, BuildingDef{
		Name: "Stone Camp", Key: "stone_camp", Category: "production",
		BaseCost:    map[string]float64{"wood": 45},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "stone", Value: 0.274}},
		BuildTicks:  80,
		RequiredAge: "stone_age",
		Description: "A basic camp for gathering stones.",
		LineageKey:  "geological_extraction", LineageTier: 0,
		WorkerDomain: "masonry", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 1 — stone_age  rate=0.16  output=stone
	b = append(b, BuildingDef{
		Name: "Stone Pit", Key: "stone_pit", Category: "production",
		BaseCost:    map[string]float64{"wood": 300, "stone": 180},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "stone", Value: 3.48}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A shallow pit dug for stone.",
		LineageKey:  "geological_extraction", LineageTier: 1,
		WorkerDomain: "masonry", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 2 — bronze_age  rate=0.32  output=stone
	b = append(b, BuildingDef{
		Name: "Quarry", Key: "quarry", Category: "production",
		BaseCost:    map[string]float64{"wood": 1800, "stone": 1200, "iron": 300},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "stone", Value: 4}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "Organized stone quarrying.",
		LineageKey:  "geological_extraction", LineageTier: 2,
		WorkerDomain: "masonry", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 3 — iron_age  rate=1.28  dual: marble(0.64) + iron_ore(0.64)
	b = append(b, BuildingDef{
		Name: "Marble Quarry", Key: "marble_quarry", Category: "production",
		BaseCost:  map[string]float64{"stone": 14000, "iron": 6000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 0.64},
			{Type: "production", Target: "iron_ore", Value: 0.64},
		},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Marble and iron ore dual extraction.",
		LineageKey:  "geological_extraction", LineageTier: 3,
		WorkerDomain: "masonry", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "stone",
	})
	// tier 4 — classical_age  rate=2.56  dual: marble(1.28) + iron_ore(1.28)
	b = append(b, BuildingDef{
		Name: "Marble Works", Key: "marble_works", Category: "production",
		BaseCost:  map[string]float64{"stone": 84000, "gold": 30000, "iron": 24000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 1.28},
			{Type: "production", Target: "iron_ore", Value: 1.28},
		},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Classical marble and ore works.",
		LineageKey:  "geological_extraction", LineageTier: 4,
		WorkerDomain: "masonry", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "stone",
	})
	// tier 5 — medieval_age  rate=2.56  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Stonemason's Guild", Key: "stonemasons_guild", Category: "production",
		BaseCost:    map[string]float64{"stone": 480000, "gold": 150000, "knowledge": 45000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron_ore", Value: 2.56}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Guild of stonemasons extracting iron ore.",
		LineageKey:  "geological_extraction", LineageTier: 5,
		WorkerDomain: "masonry", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "iron_ore",
	})
	// tier 6 — renaissance_age  rate=5.12  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Iron Mine", Key: "iron_mine", Category: "production",
		BaseCost:    map[string]float64{"gold": 1500000, "steel": 540000, "knowledge": 230000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron_ore", Value: 5.12}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Deep iron ore mining.",
		LineageKey:  "geological_extraction", LineageTier: 6,
		WorkerDomain: "masonry", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 7 — colonial_age  rate=10.24  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Deep Iron Mine", Key: "deep_iron_mine", Category: "production",
		BaseCost:    map[string]float64{"gold": 9000000, "steel": 5400000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron_ore", Value: 10.24}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Colonial deep iron ore extraction.",
		LineageKey:  "geological_extraction", LineageTier: 7,
		WorkerDomain: "masonry", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 8 — industrial_age  rate=20.48  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Steam Mine", Key: "steam_mine", Category: "production",
		BaseCost:    map[string]float64{"steel": 6.6e07, "coal": 2.4e07, "gold": 3.6e07},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron_ore", Value: 20.48}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Steam-powered industrial iron ore mine.",
		LineageKey:  "geological_extraction", LineageTier: 8,
		WorkerDomain: "masonry", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 9 — victorian_age  rate=40.96  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Mine", Key: "uranium_mine", Category: "production",
		BaseCost:    map[string]float64{"steel": 5.3e08, "iron": 2.4e08, "gold": 3e08},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "uranium", Value: 40.96}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "Early uranium ore extraction.",
		LineageKey:  "geological_extraction", LineageTier: 9,
		WorkerDomain: "masonry", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 10 — electric_age  rate=81.92  output=uranium
	b = append(b, BuildingDef{
		Name: "Nuclear Extraction Plant", Key: "nuclear_extraction_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 2.9e09, "electricity": 1.1e09},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "uranium", Value: 81.92}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "High-tech nuclear material extraction.",
		LineageKey:  "geological_extraction", LineageTier: 10,
		WorkerDomain: "masonry", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 11 — atomic_age  rate=163.84  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Processing Works", Key: "uranium_processing_works", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.5e10, "electricity": 6e09, "uranium": 1.2e09},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "uranium", Value: 171000}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Industrial uranium processing.",
		LineageKey:  "geological_extraction", LineageTier: 11,
		WorkerDomain: "masonry", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 12 — modern_age  rate=327.68  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Titanium Mine", Key: "titanium_mine", Category: "production",
		BaseCost:    map[string]float64{"steel": 8.4e10, "electricity": 3e10, "data": 2.4e09},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 327.68}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "High-precision titanium ore extraction.",
		LineageKey:  "geological_extraction", LineageTier: 12,
		WorkerDomain: "masonry", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 13 — information_age  rate=655.36  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Precision Mine", Key: "precision_mine", Category: "production",
		BaseCost:    map[string]float64{"electricity": 2.1e11, "data": 2.1e10, "steel": 3.9e11},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 655.36}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "AI-guided precision titanium mining.",
		LineageKey:  "geological_extraction", LineageTier: 13,
		WorkerDomain: "masonry", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 14 — digital_age  rate=1310.72  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Nano Drill Complex", Key: "nano_drill_complex", Category: "production",
		BaseCost:    map[string]float64{"electricity": 1.1e12, "data": 1.2e11, "steel": 1.6e12},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 1310.72}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Nanoscale drilling for titanium ore.",
		LineageKey:  "geological_extraction", LineageTier: 14,
		WorkerDomain: "masonry", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 15 — cyberpunk_age  rate=2621.44  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Dark Crystal Mine", Key: "dark_crystal_mine", Category: "production",
		BaseCost:    map[string]float64{"data": 5.1e11, "crypto": 3.3e12, "electricity": 6.6e12},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 2621.44}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Extraction of dark matter crystals.",
		LineageKey:  "geological_extraction", LineageTier: 15,
		WorkerDomain: "masonry", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 16 — fusion_age  rate=5242.88  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Exotic Mineral Extractor", Key: "exotic_mineral_extractor", Category: "production",
		BaseCost:    map[string]float64{"plasma": 1.4e13, "electricity": 3.9e13, "steel": 5.1e13},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 5242.88}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Plasma-assisted exotic mineral extraction.",
		LineageKey:  "geological_extraction", LineageTier: 16,
		WorkerDomain: "masonry", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 17 — space_age  rate=10485.76  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Asteroid Crystal Mine", Key: "asteroid_crystal_mine", Category: "production",
		BaseCost:    map[string]float64{"titanium": 2.1e14, "plasma": 9.6e13, "electricity": 2.6e14},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 10485.76}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Asteroid belt dark matter crystal mining.",
		LineageKey:  "geological_extraction", LineageTier: 17,
		WorkerDomain: "masonry", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 18 — interstellar_age  rate=20971.52  output=antimatter
	b = append(b, BuildingDef{
		Name: "Stellar Core Drill", Key: "stellar_core_drill", Category: "production",
		BaseCost:     map[string]float64{"dark_matter": 2.6e14, "titanium": 2.1e15, "plasma": 1.3e15},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "antimatter", Value: 20971.52}},
		BuildTicks:   3600,
		RequiredAge:  "interstellar_age",
		RequiredTech: "stellar_core_mining",
		Description:  "Drills into stellar cores for antimatter.",
		LineageKey:   "geological_extraction", LineageTier: 18,
		WorkerDomain: "masonry", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 19 — galactic_age  rate=41943.04  output=antimatter
	b = append(b, BuildingDef{
		Name: "Neutron Star Mine", Key: "neutron_star_mine", Category: "production",
		BaseCost:     map[string]float64{"antimatter": 5.1e14, "dark_matter": 2.6e15, "titanium": 1.3e16},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "antimatter", Value: 9.15e09}},
		BuildTicks:   3600,
		RequiredAge:  "galactic_age",
		RequiredTech: "neutron_mining",
		Description:  "Mining neutron stars for antimatter.",
		LineageKey:   "geological_extraction", LineageTier: 19,
		WorkerDomain: "masonry", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 20 — quantum_age  rate=83886.08  output=antimatter
	b = append(b, BuildingDef{
		Name: "Reality Excavator", Key: "reality_excavator", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 4.8e14, "antimatter": 1.4e17, "dark_matter": 1.1e17},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 9.19e12}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Excavates antimatter from the quantum foam.",
		LineageKey:  "geological_extraction", LineageTier: 20,
		WorkerDomain: "masonry", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})

	return b
}
