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
		BaseCost:    map[string]float64{"wood": 15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "stone", Value: 0.08}},
		BuildTicks:  80,
		RequiredAge: "stone_age",
		Description: "A basic camp for gathering stones. +0.08 stone/tick (3 workers).",
		LineageKey:  "geological_extraction", LineageTier: 0,
		WorkerDomain: "masonry", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 1 — stone_age  rate=0.16  output=stone
	b = append(b, BuildingDef{
		Name: "Stone Pit", Key: "stone_pit", Category: "production",
		BaseCost:    map[string]float64{"wood": 100, "stone": 60},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "stone", Value: 0.16}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A shallow pit dug for stone. +0.16 stone/tick (4 workers).",
		LineageKey:  "geological_extraction", LineageTier: 1,
		WorkerDomain: "masonry", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 2 — bronze_age  rate=0.32  output=stone
	b = append(b, BuildingDef{
		Name: "Quarry", Key: "quarry", Category: "production",
		BaseCost:    map[string]float64{"wood": 600, "stone": 400, "iron": 100},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "stone", Value: 0.32}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "Organised stone quarrying. +0.32 stone/tick (5 workers).",
		LineageKey:  "geological_extraction", LineageTier: 2,
		WorkerDomain: "masonry", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 3 — iron_age  rate=1.28  dual: marble(0.64) + iron_ore(0.64)
	b = append(b, BuildingDef{
		Name: "Marble Quarry", Key: "marble_quarry", Category: "production",
		BaseCost:  map[string]float64{"stone": 4500, "iron": 2000},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 0.64},
			{Type: "production", Target: "iron_ore", Value: 0.64},
		},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Marble and iron ore dual extraction. +0.64 marble, +0.64 iron_ore/tick (5 workers).",
		LineageKey:  "geological_extraction", LineageTier: 3,
		WorkerDomain: "masonry", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "stone",
	})
	// tier 4 — classical_age  rate=0.48  dual: marble(0.24) + iron_ore(0.24)
	b = append(b, BuildingDef{
		Name: "Marble Works", Key: "marble_works", Category: "production",
		BaseCost:  map[string]float64{"stone": 28000, "gold": 10000, "iron": 8000},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 0.24},
			{Type: "production", Target: "iron_ore", Value: 0.24},
		},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "Classical marble and ore works. +0.24 marble, +0.24 iron_ore/tick (6 workers).",
		LineageKey:  "geological_extraction", LineageTier: 4,
		WorkerDomain: "masonry", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "stone",
	})
	// tier 5 — medieval_age  rate=0.96  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Stonemason's Guild", Key: "stonemasons_guild", Category: "production",
		BaseCost:    map[string]float64{"stone": 160000, "gold": 50000, "knowledge": 15000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "iron_ore", Value: 0.96}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "Guild of stonemasons extracting iron ore. +0.96 iron_ore/tick (6 workers).",
		LineageKey:  "geological_extraction", LineageTier: 5,
		WorkerDomain: "masonry", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "iron_ore",
	})
	// tier 6 — renaissance_age  rate=1.92  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Iron Mine", Key: "iron_mine", Category: "production",
		BaseCost:    map[string]float64{"gold": 500000, "steel": 180000, "knowledge": 75000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "iron_ore", Value: 1.92}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "Deep iron ore mining. +1.92 iron_ore/tick (7 workers).",
		LineageKey:  "geological_extraction", LineageTier: 6,
		WorkerDomain: "masonry", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 7 — colonial_age  rate=3.84  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Deep Iron Mine", Key: "deep_iron_mine", Category: "production",
		BaseCost:    map[string]float64{"gold": 3e6, "steel": 1.8e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "iron_ore", Value: 3.84}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "Colonial deep iron ore extraction. +3.84 iron_ore/tick (8 workers).",
		LineageKey:  "geological_extraction", LineageTier: 7,
		WorkerDomain: "masonry", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 8 — industrial_age  rate=7.68  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Steam Mine", Key: "steam_mine", Category: "production",
		BaseCost:    map[string]float64{"steel": 22e6, "coal": 8e6, "gold": 12e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "iron_ore", Value: 7.68}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Steam-powered industrial iron ore mine. +7.68 iron_ore/tick (10 workers).",
		LineageKey:  "geological_extraction", LineageTier: 8,
		WorkerDomain: "masonry", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 9 — victorian_age  rate=15.36  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Mine", Key: "uranium_mine", Category: "production",
		BaseCost:    map[string]float64{"steel": 175e6, "iron": 80e6, "gold": 100e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "uranium", Value: 15.36}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Early uranium ore extraction. +15.36 uranium/tick (10 workers).",
		LineageKey:  "geological_extraction", LineageTier: 9,
		WorkerDomain: "masonry", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 10 — electric_age  rate=30.72  output=uranium
	b = append(b, BuildingDef{
		Name: "Nuclear Extraction Plant", Key: "nuclear_extraction_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 950e6, "electricity": 380e6, "uranium": 50e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "uranium", Value: 30.72}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "High-tech nuclear material extraction. +30.72 uranium/tick (12 workers).",
		LineageKey:  "geological_extraction", LineageTier: 10,
		WorkerDomain: "masonry", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 11 — atomic_age  rate=61.44  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Processing Works", Key: "uranium_processing_works", Category: "production",
		BaseCost:    map[string]float64{"steel": 5e9, "electricity": 2e9, "uranium": 400e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "uranium", Value: 61.44}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Industrial uranium processing. +61.44 uranium/tick (12 workers).",
		LineageKey:  "geological_extraction", LineageTier: 11,
		WorkerDomain: "masonry", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 12 — modern_age  rate=122.88  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Titanium Mine", Key: "titanium_mine", Category: "production",
		BaseCost:    map[string]float64{"steel": 28e9, "electricity": 10e9, "data": 800e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 122.88}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "High-precision titanium ore extraction. +122.88 titanium_ore/tick (14 workers).",
		LineageKey:  "geological_extraction", LineageTier: 12,
		WorkerDomain: "masonry", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 13 — information_age  rate=245.76  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Precision Mine", Key: "precision_mine", Category: "production",
		BaseCost:    map[string]float64{"electricity": 70e9, "data": 7e9, "steel": 130e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 245.76}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "AI-guided precision titanium mining. +245.76 titanium_ore/tick (15 workers).",
		LineageKey:  "geological_extraction", LineageTier: 13,
		WorkerDomain: "masonry", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 14 — digital_age  rate=491.52  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Nano Drill Complex", Key: "nano_drill_complex", Category: "production",
		BaseCost:    map[string]float64{"electricity": 360e9, "data": 40e9, "steel": 520e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "titanium_ore", Value: 491.52}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Nanoscale drilling for titanium ore. +491.52 titanium_ore/tick (16 workers).",
		LineageKey:  "geological_extraction", LineageTier: 14,
		WorkerDomain: "masonry", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 15 — cyberpunk_age  rate=983.04  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Dark Crystal Mine", Key: "dark_crystal_mine", Category: "production",
		BaseCost:    map[string]float64{"data": 170e9, "crypto": 1.1e12, "electricity": 2.2e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 983.04}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Extraction of dark matter crystals. +983.04 dark_matter_crystals/tick (18 workers).",
		LineageKey:  "geological_extraction", LineageTier: 15,
		WorkerDomain: "masonry", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 16 — fusion_age  rate=1966.08  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Exotic Mineral Extractor", Key: "exotic_mineral_extractor", Category: "production",
		BaseCost:    map[string]float64{"plasma": 4.5e12, "electricity": 13e12, "steel": 17e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 1966.08}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Plasma-assisted exotic mineral extraction. +1966.08 dark_matter_crystals/tick (20 workers).",
		LineageKey:  "geological_extraction", LineageTier: 16,
		WorkerDomain: "masonry", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 17 — space_age  rate=3932.16  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Asteroid Crystal Mine", Key: "asteroid_crystal_mine", Category: "production",
		BaseCost:    map[string]float64{"titanium": 70e12, "plasma": 32e12, "electricity": 85e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 3932.16}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Asteroid belt dark matter crystal mining. +3932.16 dark_matter_crystals/tick (20 workers).",
		LineageKey:  "geological_extraction", LineageTier: 17,
		WorkerDomain: "masonry", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 18 — interstellar_age  rate=7864.32  output=antimatter
	b = append(b, BuildingDef{
		Name: "Stellar Core Drill", Key: "stellar_core_drill", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 85e12, "titanium": 700e12, "plasma": 420e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 7864.32}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Drills into stellar cores for antimatter. +7864.32 antimatter/tick (25 workers).",
		LineageKey:  "geological_extraction", LineageTier: 18,
		WorkerDomain: "masonry", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 19 — galactic_age  rate=15728.64  output=antimatter
	b = append(b, BuildingDef{
		Name: "Neutron Star Mine", Key: "neutron_star_mine", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 170e12, "dark_matter": 850e12, "titanium": 4.2e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 15728.64}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Mining neutron stars for antimatter. +15728.64 antimatter/tick (25 workers).",
		LineageKey:  "geological_extraction", LineageTier: 19,
		WorkerDomain: "masonry", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 20 — quantum_age  rate=31457.28  output=antimatter
	b = append(b, BuildingDef{
		Name: "Reality Excavator", Key: "reality_excavator", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 160e12, "antimatter": 48e15, "dark_matter": 38e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 31457.28}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Excavates antimatter from the quantum foam. +31457.28 antimatter/tick (30 workers).",
		LineageKey:  "geological_extraction", LineageTier: 20,
		WorkerDomain: "masonry", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})

	return b
}
