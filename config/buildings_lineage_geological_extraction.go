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
		LineageKey:  "geological_extraction", LineageTier: 0, LineKey: "geological_extraction_stone",
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
		LineageKey:  "geological_extraction", LineageTier: 1, LineKey: "geological_extraction_stone",
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
		LineageKey:  "geological_extraction", LineageTier: 2, LineKey: "geological_extraction_stone",
		WorkerDomain: "masonry", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 3 — iron_age  rate=1.28  dual: marble(0.64) + iron_ore(0.64)
	b = append(b, BuildingDef{
		Name: "Marble Quarry", Key: "marble_quarry", Category: "production",
		BaseCost:  map[string]float64{"stone": 14000, "iron": 6000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 4.51},
			{Type: "production", Target: "iron_ore", Value: 1.89},
			{Type: "production", Target: "stone", Value: 6.29},
			{Type: "production", Target: "iron", Value: 2.26},
		},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Marble, iron ore, stone and iron extraction.",
		LineageKey:  "geological_extraction", LineageTier: 3, LineKey: "geological_extraction_stone",
		WorkerDomain: "masonry", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "stone",
	})
	// tier 4 — classical_age  rate=2.56  dual: marble(1.28) + iron_ore(1.28)
	b = append(b, BuildingDef{
		Name: "Marble Works", Key: "marble_works", Category: "production",
		BaseCost:  map[string]float64{"stone": 84000, "gold": 30000, "iron": 24000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 28.4},
			{Type: "production", Target: "iron_ore", Value: 6.37},
			{Type: "production", Target: "stone", Value: 35.3},
			{Type: "production", Target: "iron", Value: 8.17},
		},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Classical marble, ore, stone and iron works.",
		LineageKey:  "geological_extraction", LineageTier: 4, LineKey: "geological_extraction_stone",
		WorkerDomain: "masonry", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "stone",
	})
	// tier 5 — medieval_age  rate=2.56  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Stonemason's Guild", Key: "stonemasons_guild", Category: "production",
		BaseCost:  map[string]float64{"stone": 759000, "gold": 237000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "iron_ore", Value: 10.1},
			{Type: "production", Target: "iron", Value: 14.9},
			{Type: "production", Target: "stone", Value: 339},
		},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Guild of stonemasons extracting iron ore, iron and stone.",
		LineageKey:  "geological_extraction", LineageTier: 5, LineKey: "geological_extraction_iron_ore",
		WorkerDomain: "masonry", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "iron_ore",
	})
	// tier 6 — renaissance_age  rate=5.12  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Iron Mine", Key: "iron_mine", Category: "production",
		BaseCost:  map[string]float64{"gold": 2.33 * Million, "steel": 839000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "iron_ore", Value: 5.27},
			{Type: "production", Target: "iron", Value: 9.02},
			{Type: "production", Target: "stone", Value: 207},
		},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Deep iron ore, iron and stone mining.",
		LineageKey:  "geological_extraction", LineageTier: 6, LineKey: "geological_extraction_iron_ore",
		WorkerDomain: "masonry", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 7 — colonial_age  rate=10.24  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Deep Iron Mine", Key: "deep_iron_mine", Category: "production",
		BaseCost:  map[string]float64{"gold": 9 * Million, "steel": 5.4 * Million},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "iron_ore", Value: 112},
			{Type: "production", Target: "iron", Value: 112},
			{Type: "production", Target: "stone", Value: 2090},
		},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Colonial deep iron ore, iron and stone extraction.",
		LineageKey:  "geological_extraction", LineageTier: 7, LineKey: "geological_extraction_iron_ore",
		WorkerDomain: "masonry", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 8 — industrial_age  rate=20.48  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Steam Mine", Key: "steam_mine", Category: "production",
		BaseCost:  map[string]float64{"steel": 66 * Million, "coal": 24 * Million, "gold": 36 * Million},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "iron_ore", Value: 161},
			{Type: "production", Target: "iron", Value: 2970},
			{Type: "production", Target: "stone", Value: 4450},
		},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Steam-powered industrial iron ore, iron and stone mine.",
		LineageKey:  "geological_extraction", LineageTier: 8, LineKey: "geological_extraction_iron_ore",
		WorkerDomain: "masonry", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 9 — victorian_age  rate=40.96  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Mine", Key: "uranium_mine", Category: "production",
		BaseCost:  map[string]float64{"steel": 530 * Million, "iron": 240 * Million, "gold": 300 * Million},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "uranium", Value: 40.96},
			{Type: "production", Target: "stone", Value: 46600},
		},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "Early uranium ore and stone extraction.",
		LineageKey:  "geological_extraction", LineageTier: 9, LineKey: "geological_extraction_uranium",
		WorkerDomain: "masonry", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 10 — electric_age  rate=81.92  output=uranium
	b = append(b, BuildingDef{
		Name: "Nuclear Extraction Plant", Key: "nuclear_extraction_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 2.9 * Billion, "electricity": 1.1 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "uranium", Value: 81.92}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "High-tech nuclear material extraction.",
		LineageKey:  "geological_extraction", LineageTier: 10, LineKey: "geological_extraction_uranium",
		WorkerDomain: "masonry", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 11 — atomic_age  rate=163.84  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Processing Works", Key: "uranium_processing_works", Category: "production",
		BaseCost:  map[string]float64{"steel": 24.6 * Billion, "electricity": 9.85 * Billion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "uranium", Value: 110000},
			{Type: "production", Target: "stone", Value: 1.04 * Million},
		},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Industrial uranium and stone processing.",
		LineageKey:  "geological_extraction", LineageTier: 11, LineKey: "geological_extraction_uranium",
		WorkerDomain: "masonry", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 12 — modern_age  rate=327.68  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Titanium Mine", Key: "titanium_mine", Category: "production",
		BaseCost:    map[string]float64{"steel": 84 * Billion, "electricity": 30 * Billion, "data": 2.4 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 9.17 * Million}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "High-precision titanium ore extraction.",
		LineageKey:  "geological_extraction", LineageTier: 12, LineKey: "geological_extraction_titanium_ore",
		WorkerDomain: "masonry", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 13 — information_age  rate=655.36  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Precision Mine", Key: "precision_mine", Category: "production",
		BaseCost:    map[string]float64{"electricity": 210 * Billion, "data": 21 * Billion, "steel": 390 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 72.9 * Million}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "AI-guided precision titanium mining.",
		LineageKey:  "geological_extraction", LineageTier: 13, LineKey: "geological_extraction_titanium_ore",
		WorkerDomain: "masonry", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 14 — digital_age  rate=1310.72  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Nano Drill Complex", Key: "nano_drill_complex", Category: "production",
		BaseCost:    map[string]float64{"electricity": 1.1 * Trillion, "data": 120 * Billion, "steel": 1.6 * Trillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 111 * Million}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Nanoscale drilling for titanium ore.",
		LineageKey:  "geological_extraction", LineageTier: 14, LineKey: "geological_extraction_titanium_ore",
		WorkerDomain: "masonry", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 15 — cyberpunk_age  rate=2621.44  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Dark Crystal Mine", Key: "dark_crystal_mine", Category: "production",
		BaseCost:    map[string]float64{"data": 510 * Billion, "crypto": 3.3 * Trillion, "electricity": 6.6 * Trillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 435 * Million}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Extraction of dark matter crystals.",
		LineageKey:  "geological_extraction", LineageTier: 15, LineKey: "geological_extraction_dark_matter_crystals",
		WorkerDomain: "masonry", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 16 — fusion_age  rate=5242.88  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Exotic Mineral Extractor", Key: "exotic_mineral_extractor", Category: "production",
		BaseCost:    map[string]float64{"plasma": 14 * Trillion, "electricity": 39 * Trillion, "steel": 51 * Trillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 922 * Million}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Plasma-assisted exotic mineral extraction.",
		LineageKey:  "geological_extraction", LineageTier: 16, LineKey: "geological_extraction_dark_matter_crystals",
		WorkerDomain: "masonry", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 17 — space_age  rate=10485.76  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Asteroid Crystal Mine", Key: "asteroid_crystal_mine", Category: "production",
		BaseCost:    map[string]float64{"titanium": 210 * Trillion, "plasma": 96 * Trillion, "electricity": 260 * Trillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 7.84 * Billion}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Asteroid belt dark matter crystal mining.",
		LineageKey:  "geological_extraction", LineageTier: 17, LineKey: "geological_extraction_dark_matter_crystals",
		WorkerDomain: "masonry", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 18 — interstellar_age  rate=20971.52  output=antimatter
	b = append(b, BuildingDef{
		Name: "Stellar Core Drill", Key: "stellar_core_drill", Category: "production",
		BaseCost:     map[string]float64{"dark_matter": 260 * Trillion, "titanium": 2.1 * Quadrillion, "plasma": 1.3 * Quadrillion},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "antimatter", Value: 20971.52}},
		BuildTicks:   3600,
		RequiredAge:  "interstellar_age",
		RequiredTech: "stellar_core_mining",
		Description:  "Drills into stellar cores for antimatter.",
		LineageKey:   "geological_extraction", LineageTier: 18, LineKey: "geological_extraction_antimatter",
		WorkerDomain: "masonry", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 19 — galactic_age  rate=41943.04  output=antimatter
	b = append(b, BuildingDef{
		Name: "Neutron Star Mine", Key: "neutron_star_mine", Category: "production",
		BaseCost:     map[string]float64{"antimatter": 510 * Trillion, "dark_matter": 2.6 * Quadrillion, "titanium": 13 * Quadrillion},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "antimatter", Value: 9.15 * Billion}},
		BuildTicks:   3600,
		RequiredAge:  "galactic_age",
		RequiredTech: "neutron_mining",
		Description:  "Mining neutron stars for antimatter.",
		LineageKey:   "geological_extraction", LineageTier: 19, LineKey: "geological_extraction_antimatter",
		WorkerDomain: "masonry", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 20 — quantum_age  rate=83886.08  output=antimatter
	b = append(b, BuildingDef{
		Name: "Reality Excavator", Key: "reality_excavator", Category: "production",
		BaseCost:  map[string]float64{"titanium": 5.71 * Quadrillion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "antimatter", Value: 7.04 * Trillion},
		},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Excavates antimatter from the quantum foam.",
		LineageKey:  "geological_extraction", LineageTier: 20, LineKey: "geological_extraction_antimatter",
		WorkerDomain: "masonry", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})

	return b
}
